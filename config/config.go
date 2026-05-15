package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Polling PollingConfig `yaml:"polling"`
	Logging LoggingConfig `yaml:"logging"`
	I2C     I2CConfig     `yaml:"i2c"`
	Status  StatusConfig  `yaml:"status"`
}

type PollingConfig struct {
	SoftwareRateMS           int  `yaml:"software_rate_ms"`
	NormalLoggingRateMS      int  `yaml:"normal_logging_rate_ms"`
	EnhancedLoggingRateMS    int  `yaml:"enhanced_logging_rate_ms"`
	EnableEnhancedOnAnomaly bool `yaml:"enable_enhanced_on_anomaly"`
}

type LoggingConfig struct {
	Directory  string `yaml:"directory"`
	MaxSizeMB  int    `yaml:"max_size_mb"`
	MaxBackups int    `yaml:"max_backups"`
}

type I2CConfig struct {
	Bus     int   `yaml:"bus"`
	Address uint8 `yaml:"address"`
}

type StatusConfig struct {
	ListenAddress string `yaml:"listen_address"`
}

// LoadConfig reads the YAML configuration file from the given path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
