package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"despaddosy/pkg/config"
	"despaddosy/pkg/features"
	"despaddosy/pkg/logging"
	"despaddosy/pkg/metrics"
	"despaddosy/pkg/policy"
	"despaddosy/pkg/ratelimit"
	"despaddosy/pkg/store"
	"despaddosy/pkg/tracing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type securityEvent struct {
	Timestamp       time.Time `json:"timestamp"`
	RequestID       string    `json:"request_id"`
	RouteID         string    `json:"route_id"`
	Method          string    `json:"method"`
	Decision        string    `json:"decision"`
	Score           int       `json:"score"`
	Reasons         []string  `json:"reasons"`
	IPHash          string    `json:"ip_hash"`
	UserID          string    `json:"user_id,omitempty"`
	UserAgentFamily string    `json:"user_agent_family"`
	RateSnapshot    int       `json:"rate_snapshot"`
}

type riskResponse struct {
	Decision   string   `json:"decision"`
	Score      int      `json:"score"`
	Reasons    []string `json:"reasons"`
	TTLSeconds int      `json:"ttl_seconds"`
}

type claims struct {
	UserID string   `json:"user_id"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

func main() {
	cfgPath := os.Getenv("GATEWAY_CONFIG")
	if cfgPath == "" {
		cfgPath = "configs/gateway.yaml"
	}
	cfg, err := config.LoadGatewayConfig(cfgPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load gateway config")
	}
	logger := logging.Init(cfg.ServiceName)
	_ = logger

	ctx := context.Background()
	shutdown, err := tracing.Init(ctx, cfg.ServiceName, cfg.Observability.OTLPEndpoint)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init tracing")
	}
	defer func() {
		_ = shutdown(context.Background())
	}()

	redisClient := store.NewRedis(cfg.Redis.Addr)
	limiter := ratelimit.NewRedisLimiter(redisClient)
	matcher := policy.NewMatcher(cfg.Routes)
	gatewayMetrics := metrics.NewGatewayMetrics()

	proxies := buildProxies(cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})
	if cfg.Observability.MetricsAddr != "" {
		go func() {
			metricsMux := http.NewServeMux()
			metricsMux.Handle("/metrics", metrics.Handler())
			if err := http.ListenAndServe(cfg.Observability.MetricsAddr, metricsMux); err != nil {
				log.Fatal().Err(err).Msg("metrics server failed")
			}
		}()
	}

	mux.Handle("/", gatewayHandler(cfg, matcher, limiter, redisClient, gatewayMetrics, proxies))

	handler := logging.RequestID(logging.AccessLog(mux))

	srv := &http.Server{
		Addr:              cfg.Listener.Address,
		Handler:           handler,
		ReadHeaderTimeout: time.Duration(cfg.Listener.Timeouts.ReadHeaderSeconds) * time.Second,
		ReadTimeout:       time.Duration(cfg.Listener.Timeouts.ReadSeconds) * time.Second,
		WriteTimeout:      time.Duration(cfg.Listener.Timeouts.WriteSeconds) * time.Second,
	}

	log.Info().Str("addr", cfg.Listener.Address).Msg("gateway listening")
	if cfg.Listener.TLS.Enabled {
		log.Info().Msg("TLS enabled")
		if err := srv.ListenAndServeTLS(cfg.Listener.TLS.CertFile, cfg.Listener.TLS.KeyFile); err != nil {
			log.Fatal().Err(err).Msg("gateway stopped")
		}
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("gateway stopped")
	}
}

func gatewayHandler(cfg config.GatewayConfig, matcher *policy.Matcher, limiter ratelimit.Limiter, redisClient *redis.Client, metrics *metrics.GatewayMetrics, proxies map[string]*httputil.ReverseProxy) http.Handler {
	tracer := otel.Tracer("gateway")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		route, ok := matcher.Match(r)
		if !ok {
			http.NotFound(w, r)
			return
		}
		if !policy.MethodAllowed(route, r.Method) {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		cleanPath := path.Clean(r.URL.Path)
		if cleanPath != r.URL.Path {
			r.URL.Path = cleanPath
		}
		if route.MaxBodyBytes > 0 {
			r.Body = http.MaxBytesReader(w, r.Body, route.MaxBodyBytes)
		}
		if len(route.AllowedContentType) > 0 && r.Header.Get("Content-Type") != "" {
			ct := strings.Split(r.Header.Get("Content-Type"), ";")[0]
			if !contentTypeAllowed(route.AllowedContentType, strings.TrimSpace(ct)) {
				w.WriteHeader(http.StatusUnsupportedMediaType)
				return
			}
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		r = r.WithContext(ctx)

		userID, roles, hasAuth := authenticate(r, cfg.JWT.Secret)
		if route.AuthRequired && !hasAuth {
			respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "auth required"})
			return
		}
		if len(route.RolesAllowed) > 0 && !hasRole(route.RolesAllowed, roles) {
			respondJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
			return
		}

		ipKey := fmt.Sprintf("ip:%s:route:%s", clientIP(r.RemoteAddr), route.ID)
		allowed, snapshot, err := limiter.Allow(r.Context(), ipKey, route.RateLimits.PerIP.RPS, route.RateLimits.PerIP.Burst)
		if err != nil {
			log.Error().Err(err).Msg("ratelimit error")
		}
		if !allowed {
			metrics.RateLimit.WithLabelValues(route.ID, "ip").Inc()
			respondJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit exceeded"})
			return
		}
		if userID != "" {
			userKey := fmt.Sprintf("user:%s:route:%s", userID, route.ID)
			allowed, _, err = limiter.Allow(r.Context(), userKey, route.RateLimits.PerUser.RPS, route.RateLimits.PerUser.Burst)
			if err != nil {
				log.Error().Err(err).Msg("ratelimit error")
			}
			if !allowed {
				metrics.RateLimit.WithLabelValues(route.ID, "user").Inc()
				respondJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit exceeded"})
				return
			}
		}
		routeKey := fmt.Sprintf("route:%s", route.ID)
		allowed, _, err = limiter.Allow(r.Context(), routeKey, route.RateLimits.PerRoute.RPS, route.RateLimits.PerRoute.Burst)
		if err != nil {
			log.Error().Err(err).Msg("ratelimit error")
		}
		if !allowed {
			metrics.RateLimit.WithLabelValues(route.ID, "route").Inc()
			respondJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit exceeded"})
			return
		}

		ipAddr := clientIP(r.RemoteAddr)
		if isBlocked(r.Context(), redisClient, route.ID, userID, ipAddr) {
			respondBlocked(w, cfg.BlockedResponse.RetryAfterSeconds, cfg.BlockedResponse.Message, logging.GetRequestID(r.Context()), http.StatusTooManyRequests)
			return
		}

		bodyLen := r.ContentLength
		featuresVector := features.Extract(r, route.ID, bodyLen, snapshot, hasAuth, userID, cfg.Riskd.Salt)

		decision := "allow"
		score := 0
		reasons := []string{}
		if route.Risk.Enabled {
			ctxRisk, cancelRisk := context.WithTimeout(r.Context(), time.Duration(cfg.Riskd.TimeoutMs)*time.Millisecond)
			defer cancelRisk()
			resp, err := callRiskd(ctxRisk, cfg.Riskd.Endpoint, featuresVector, route)
			if err != nil {
				metrics.RiskdErr.Inc()
				log.Error().Err(err).Msg("riskd error")
			} else {
				decision = resp.Decision
				score = resp.Score
				reasons = resp.Reasons
				if resp.Decision == "challenge" {
					metrics.Challenges.WithLabelValues(route.ID, "risk").Inc()
					setBlock(r.Context(), redisClient, route.ID, userID, ipAddr, resp.TTLSeconds)
					publishEvent(r.Context(), redisClient, route.ID, r.Method, decision, score, reasons, featuresVector, userID, snapshot)
					respondBlocked(w, resp.TTLSeconds, "suspicious activity detected", logging.GetRequestID(r.Context()), http.StatusTooManyRequests)
					return
				}
				if resp.Decision == "block" {
					metrics.Blocks.WithLabelValues(route.ID, "risk", "block").Inc()
					setBlock(r.Context(), redisClient, route.ID, userID, ipAddr, resp.TTLSeconds)
					publishEvent(r.Context(), redisClient, route.ID, r.Method, decision, score, reasons, featuresVector, userID, snapshot)
					respondBlocked(w, resp.TTLSeconds, "access temporarily restricted", logging.GetRequestID(r.Context()), http.StatusForbidden)
					return
				}
				if resp.Decision == "monitor" {
					publishEvent(r.Context(), redisClient, route.ID, r.Method, decision, score, reasons, featuresVector, userID, snapshot)
				}
			}
		}

		ctx, span := tracer.Start(r.Context(), "gateway.request")
		span.SetAttributes(attribute.String("route", route.ID), attribute.String("decision", decision), attribute.Int("score", score))
		defer span.End()
		r = r.WithContext(ctx)
		proxy := proxies[route.ID]
		if proxy == nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		proxy.ServeHTTP(w, r)

		metrics.Requests.WithLabelValues(route.ID, r.Method, fmt.Sprintf("%d", http.StatusOK)).Inc()
		metrics.Latency.WithLabelValues(route.ID).Observe(time.Since(start).Seconds())
	})
}

func callRiskd(ctx context.Context, endpoint string, vector features.Vector, route config.RoutePolicy) (riskResponse, error) {
	var resp riskResponse
	if endpoint == "" {
		return resp, errors.New("riskd endpoint missing")
	}
	payload := map[string]any{
		"features":      vector,
		"auth_hint":     route.Risk.AuthHint,
		"max_body_hint": route.Risk.MaxBodyHint,
		"thresholds":    route.Risk.Thresholds,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return resp, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/v1/score", bytes.NewReader(data))
	if err != nil {
		return resp, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 2 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return resp, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("riskd status %d", res.StatusCode)
	}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func authenticate(r *http.Request, secret string) (string, []string, bool) {
	auth := r.Header.Get("Authorization")
	var tokenStr string
	if auth != "" && strings.HasPrefix(auth, "Bearer ") {
		tokenStr = strings.TrimPrefix(auth, "Bearer ")
	} else if cookie, err := r.Cookie("access_token"); err == nil {
		tokenStr = cookie.Value
	}
	if tokenStr == "" {
		return "", nil, false
	}
	parsed, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", nil, false
	}
	if cl, ok := parsed.Claims.(*claims); ok && parsed.Valid {
		return cl.UserID, cl.Roles, true
	}
	return "", nil, false
}

func hasRole(allowed, roles []string) bool {
	for _, role := range roles {
		for _, allowedRole := range allowed {
			if role == allowedRole {
				return true
			}
		}
	}
	return false
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondBlocked(w http.ResponseWriter, retryAfter int, message, requestID string, status int) {
	w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
	respondJSON(w, status, map[string]string{"error": message, "request_id": requestID})
}

func isBlocked(ctx context.Context, client *redis.Client, routeID, userID, ip string) bool {
	if userID != "" {
		if val, _ := client.Get(ctx, fmt.Sprintf("block:user:%s:route:%s", userID, routeID)).Result(); val != "" {
			return true
		}
	}
	if val, _ := client.Get(ctx, fmt.Sprintf("block:ip:%s:route:%s", ip, routeID)).Result(); val != "" {
		return true
	}
	return false
}

func setBlock(ctx context.Context, client *redis.Client, routeID, userID, ip string, ttl int) {
	if userID != "" {
		_ = client.Set(ctx, fmt.Sprintf("block:user:%s:route:%s", userID, routeID), "1", time.Duration(ttl)*time.Second).Err()
	}
	_ = client.Set(ctx, fmt.Sprintf("block:ip:%s:route:%s", ip, routeID), "1", time.Duration(ttl)*time.Second).Err()
}

func publishEvent(ctx context.Context, client *redis.Client, routeID, method, decision string, score int, reasons []string, vector features.Vector, userID string, snapshot int) {
	event := securityEvent{
		Timestamp:       time.Now().UTC(),
		RequestID:       logging.GetRequestID(ctx),
		RouteID:         routeID,
		Method:          method,
		Decision:        decision,
		Score:           score,
		Reasons:         reasons,
		IPHash:          vector.IPHash,
		UserID:          userID,
		UserAgentFamily: vector.UserAgentFamily,
		RateSnapshot:    snapshot,
	}
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	_ = client.LPush(ctx, "secevents", data).Err()
	_ = client.LTrim(ctx, "secevents", 0, 5000).Err()
}

func contentTypeAllowed(allowed []string, contentType string) bool {
	for _, entry := range allowed {
		if strings.EqualFold(entry, contentType) {
			return true
		}
	}
	return false
}

func buildProxies(cfg config.GatewayConfig) map[string]*httputil.ReverseProxy {
	proxies := make(map[string]*httputil.ReverseProxy)
	for _, route := range cfg.Routes {
		target := route.UpstreamURL
		if target == "" {
			target = "http://sociald:8081"
		}
		upstream, err := url.Parse(target)
		if err != nil {
			continue
		}
		proxy := httputil.NewSingleHostReverseProxy(upstream)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Error().Err(err).Msg("proxy error")
			w.WriteHeader(http.StatusBadGateway)
		}
		if cfg.TLSUpstream.Enabled {
			proxy.Transport = &http.Transport{TLSClientConfig: upstreamTLS(cfg)}
		}
		proxies[route.ID] = proxy
	}
	return proxies
}

func upstreamTLS(cfg config.GatewayConfig) *tls.Config {
	caCert, err := os.ReadFile(cfg.TLSUpstream.CAFile)
	if err != nil {
		return &tls.Config{}
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCert)
	cert, err := tls.LoadX509KeyPair(cfg.TLSUpstream.CertFile, cfg.TLSUpstream.KeyFile)
	if err != nil {
		return &tls.Config{}
	}
	return &tls.Config{
		RootCAs:      pool,
		Certificates: []tls.Certificate{cert},
		ServerName:   "sociald",
	}
}

func clientIP(remoteAddr string) string {
	parts := strings.Split(remoteAddr, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return remoteAddr
}
