package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	content := `
polling:
  software_rate_ms: 500
  normal_logging_rate_ms: 10000
  enhanced_logging_rate_ms: 500
  enable_enhanced_on_anomaly: true

logging:
  directory: "/tmp/gnsstrack"
  simulation_directory: "/tmp/gnsstrack/sim"
  max_size_mb: 10
  max_backups: 5

i2c:
  bus: 0
  address: 0x43

status:
  listen_address: "127.0.0.1:9090"
`
	tmpfile, err := os.CreateTemp("", "config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Polling.SoftwareRateMS != 500 {
		t.Errorf("expected SoftwareRateMS 500, got %d", cfg.Polling.SoftwareRateMS)
	}
	if cfg.Logging.Directory != "/tmp/gnsstrack" {
		t.Errorf("expected Directory /tmp/gnsstrack, got %s", cfg.Logging.Directory)
	}
	if cfg.Logging.SimulationDirectory != "/tmp/gnsstrack/sim" {
		t.Errorf("expected SimulationDirectory /tmp/gnsstrack/sim, got %s", cfg.Logging.SimulationDirectory)
	}
	if cfg.I2C.Address != 0x43 {
		t.Errorf("expected Address 0x43, got 0x%x", cfg.I2C.Address)
	}
	if cfg.Status.ListenAddress != "127.0.0.1:9090" {
		t.Errorf("expected ListenAddress 127.0.0.1:9090, got %s", cfg.Status.ListenAddress)
	}
}
