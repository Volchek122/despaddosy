package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type GatewayMetrics struct {
	Requests   *prometheus.CounterVec
	Latency    *prometheus.HistogramVec
	Blocks     *prometheus.CounterVec
	Challenges *prometheus.CounterVec
	RiskdErr   prometheus.Counter
	RateLimit  *prometheus.CounterVec
}

type RiskMetrics struct {
	Scores   *prometheus.CounterVec
	Latency  prometheus.Histogram
	RedisErr prometheus.Counter
}

func NewGatewayMetrics() *GatewayMetrics {
	gm := &GatewayMetrics{
		Requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gateway_requests_total",
			Help: "Total requests through gateway.",
		}, []string{"route", "method", "status"}),
		Latency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "gateway_latency_seconds",
			Help:    "Gateway latency by route.",
			Buckets: prometheus.DefBuckets,
		}, []string{"route"}),
		Blocks: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gateway_blocks_total",
			Help: "Total blocks by route.",
		}, []string{"route", "reason", "action"}),
		Challenges: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gateway_challenges_total",
			Help: "Total challenges by route.",
		}, []string{"route", "reason"}),
		RiskdErr: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "gateway_riskd_errors_total",
			Help: "Total riskd errors.",
		}),
		RateLimit: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gateway_ratelimit_total",
			Help: "Total rate limit triggers.",
		}, []string{"route", "keytype"}),
	}
	prometheus.MustRegister(gm.Requests, gm.Latency, gm.Blocks, gm.Challenges, gm.RiskdErr, gm.RateLimit)
	return gm
}

func NewRiskMetrics() *RiskMetrics {
	rm := &RiskMetrics{
		Scores: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "riskd_score_requests_total",
			Help: "Total risk decisions.",
		}, []string{"decision"}),
		Latency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "riskd_latency_seconds",
			Help:    "Riskd latency.",
			Buckets: prometheus.DefBuckets,
		}),
		RedisErr: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "riskd_redis_errors_total",
			Help: "Redis errors in riskd.",
		}),
	}
	prometheus.MustRegister(rm.Scores, rm.Latency, rm.RedisErr)
	return rm
}

func Handler() http.Handler {
	return promhttp.Handler()
}
