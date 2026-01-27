package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"despaddosy/pkg/config"
	"despaddosy/pkg/features"
	"despaddosy/pkg/logging"
	"despaddosy/pkg/metrics"
	"despaddosy/pkg/store"
	"despaddosy/pkg/tracing"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type scoreRequest struct {
	Features    features.Vector       `json:"features"`
	AuthHint    bool                  `json:"auth_hint"`
	MaxBodyHint int64                 `json:"max_body_hint"`
	Thresholds  config.RiskThresholds `json:"thresholds"`
}

type scoreResponse struct {
	Decision   string   `json:"decision"`
	Score      int      `json:"score"`
	Reasons    []string `json:"reasons"`
	TTLSeconds int      `json:"ttl_seconds"`
}

func main() {
	cfgPath := os.Getenv("RISKD_CONFIG")
	if cfgPath == "" {
		cfgPath = "configs/riskd.yaml"
	}
	cfg, err := config.LoadRiskdConfig(cfgPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load riskd config")
	}
	logging.Init(cfg.ServiceName)

	ctx := context.Background()
	shutdown, err := tracing.Init(ctx, cfg.ServiceName, cfg.Observability.OTLPEndpoint)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init tracing")
	}
	defer func() { _ = shutdown(context.Background()) }()

	riskMetrics := metrics.NewRiskMetrics()
	redisClient := store.NewRedis(cfg.Redis.Addr)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("/metrics", metrics.Handler())
	mux.HandleFunc("/v1/score", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		var req scoreRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		resp := score(ctx, redisClient, riskMetrics, cfg, req)
		riskMetrics.Latency.Observe(time.Since(start).Seconds())
		riskMetrics.Scores.WithLabelValues(resp.Decision).Inc()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})

	srv := &http.Server{Addr: cfg.Address, Handler: logging.RequestID(logging.AccessLog(mux))}
	log.Info().Str("addr", cfg.Address).Msg("riskd listening")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("riskd stopped")
	}
}

func score(ctx context.Context, client *redis.Client, metrics *metrics.RiskMetrics, cfg config.RiskdConfig, req scoreRequest) scoreResponse {
	score := 0
	reasons := []string{}
	if req.Features.RequestRateSnapshot > 5 {
		score += 20
		reasons = append(reasons, "high request rate")
	}
	if req.MaxBodyHint > 0 && req.Features.BodyLen > req.MaxBodyHint {
		score += 20
		reasons = append(reasons, "body too large")
	}
	if req.Features.SuspiciousTokenFlag {
		score += 25
		reasons = append(reasons, "suspicious token")
	}
	if req.Features.HighEntropyQueryFlag {
		score += 15
		reasons = append(reasons, "high entropy query")
	}
	if req.AuthHint && !req.Features.HasAuth {
		score += 20
		reasons = append(reasons, "auth missing")
	}
	if req.Features.UserAgentLen == 0 || req.Features.UserAgentLen > 512 {
		score += 10
		reasons = append(reasons, "unusual user-agent")
	}

	repDelta := getReputation(ctx, client, metrics, req.Features)
	score -= repDelta
	if repDelta > 0 {
		reasons = append(reasons, "positive reputation")
	}

	decision := "allow"
	ttl := 0
	thresholds := cfg.Thresholds
	if req.Thresholds.Block > 0 {
		thresholds = req.Thresholds
	}
	if score >= thresholds.Block {
		decision = "block"
		ttl = 120
	} else if score >= thresholds.Challenge {
		decision = "challenge"
		ttl = 120
	} else if score >= thresholds.Monitor {
		decision = "monitor"
	}

	if decision == "allow" && req.Features.IPHash != "" {
		_ = client.Incr(ctx, "rep:iphash:"+req.Features.IPHash).Err()
	}

	return scoreResponse{Decision: decision, Score: score, Reasons: reasons, TTLSeconds: ttl}
}

func getReputation(ctx context.Context, client *redis.Client, metrics *metrics.RiskMetrics, vector features.Vector) int {
	if vector.IPHash == "" {
		return 0
	}
	val, err := client.Get(ctx, "rep:iphash:"+vector.IPHash).Int()
	if err == nil {
		return val / 10
	}
	if err != nil && err != redis.Nil {
		metrics.RedisErr.Inc()
	}
	return 0
}
