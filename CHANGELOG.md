# Changelog

All notable changes to this project will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

## [v1.1.0] - 2026-05-26

### Added
- DS3231 RTC temperature logging for monitoring internal hardware health

## [v1.0.1] - 2026-05-16

### Changed
- Added `simulation_directory` config field so real and simulation log
  directories can be configured independently — no need to edit `config.yaml`
  when switching between `--simulate` and normal mode

## [v1.0.0] - 2026-05-16

### Added
- GNSS monitoring service for u-blox MAX-F10S receiver via I2C
- Decodes UBX `NAV-PVT` and `SEC-SIG` frames; logs position, accuracy,
  satellite count, PDOP, SBAS usage, and fix type in JSONL format
- Security monitoring for hardware jamming and spoofing detection flags
- Dynamic logging rate that switches to enhanced mode when anomalies are detected
- Size-based log rotation via lumberjack (configurable size, backup count)
- HTTP status endpoint for querying live service state and latest GNSS fix
- Simulation mode for testing without physical GNSS hardware
- Systemd service unit for running as a daemon on Raspberry Pi
- GitHub Actions release workflow publishing a pre-built ARM64 package
- `version` command to report the embedded build version
