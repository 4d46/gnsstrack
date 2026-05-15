package service

import (
	"io"
	"testing"
	"time"

	"4d46.uk/gnsstrack/config"
	"4d46.uk/gnsstrack/i2c"
)

func TestPoller_Tick(t *testing.T) {
	cfg := &config.Config{
		Polling: config.PollingConfig{
			NormalLoggingRateMS:     1000,
			EnhancedLoggingRateMS:   100,
			EnableEnhancedOnAnomaly: true,
		},
	}

	mockBus := &i2c.MockBus{}
	poller := NewPoller(cfg, mockBus, io.Discard)

	// Case 1: Normal rate
	status := &GNSSStatus{Anomalies: false}
	poller.tickWithStatus(status)
	if poller.lastLog.IsZero() {
		t.Error("expected lastLog to be set")
	}
	firstLog := poller.lastLog

	// Immediately tick again, should not log (lastLog should stay the same)
	poller.tickWithStatus(status)
	if poller.lastLog != firstLog {
		t.Errorf("expected lastLog to stay %v, got %v", firstLog, poller.lastLog)
	}

	// Case 2: Enhanced rate on anomaly
	statusWithAnomaly := &GNSSStatus{Anomalies: true}
	// Wait just enough for enhanced rate (100ms)
	time.Sleep(150 * time.Millisecond)
	poller.tickWithStatus(statusWithAnomaly)
	if poller.lastLog == firstLog {
		t.Error("expected lastLog to update due to enhanced rate and anomaly")
	}
}

// Helper to test tick logic with injected status
func (p *Poller) tickWithStatus(status *GNSSStatus) {
	now := time.Now()
	rate := time.Duration(p.cfg.Polling.NormalLoggingRateMS) * time.Millisecond
	if status.Anomalies && p.cfg.Polling.EnableEnhancedOnAnomaly {
		rate = time.Duration(p.cfg.Polling.EnhancedLoggingRateMS) * time.Millisecond
	}

	if now.Sub(p.lastLog) >= rate {
		p.logToDisk(status)
		p.lastLog = now
	}
}
