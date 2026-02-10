package main

import (
	"context"
	"testing"

	"despaddosy/pkg/config"
	"despaddosy/pkg/features"
	"despaddosy/pkg/metrics"

	"github.com/redis/go-redis/v9"
)

func TestScoreDecision(t *testing.T) {
	cfg := config.RiskdConfig{}
	cfg.Thresholds.Block = 80
	cfg.Thresholds.Challenge = 60
	cfg.Thresholds.Monitor = 30
	req := scoreRequest{
		Features:   features.Vector{RequestRateSnapshot: 10, SuspiciousTokenFlag: true, UserAgentLen: 0, HighEntropyQueryFlag: true},
		Thresholds: config.RiskThresholds{Block: 80, Challenge: 60, Monitor: 30},
	}
	resp := score(context.Background(), redis.NewClient(&redis.Options{Addr: "localhost:0"}), metrics.NewRiskMetrics(), cfg, req)
	if resp.Decision != "challenge" && resp.Decision != "block" {
		t.Fatalf("expected challenge or block, got %s", resp.Decision)
	}
}
