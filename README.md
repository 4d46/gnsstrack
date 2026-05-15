# gnsstrack

A lightweight Go-based service for monitoring and logging GNSS data from a u-blox MAX-F10S receiver via I2C. Designed for high-accuracy timing servers (like the Timebeat Timecard Mini) on Raspberry Pi CM4.

## Features

- **Active I2C Polling**: Explicitly queries the GNSS module using UBX poll requests, working even if automatic streaming is disabled.
- **Atomic Transactions**: Uses `I2C_RDWR` to safely share the bus with other processes (like Timebeat or CLI tools).
- **Robust UBX Parsing**: Decodes binary `NAV-PVT` and `SEC-SIG` frames.
- **Rich Data Collection**: Logs 3D position (Lat/Lon/Alt), accuracy estimates (H/V), satellite count, PDOP, and SBAS usage status.
- **Security Monitoring**: Tracks u-blox hardware jamming and spoofing detection flags.
- **Dynamic Logging Rates**: Automatically switches to a high-resolution "enhanced" logging rate when anomalies are detected.
- **JSONL Persistence**: Logs data in JSON Lines format for easy analysis with `jq`, Python, or ELK.
- **Size-Based Rotation**: Built-in log rotation ensures you keep a long history without filling your disk.
- **Status Monitoring**: Query the running service state and latest GNSS fix via the CLI.
- **Simulation Mode**: Built-in hardware simulation for testing without physical GNSS hardware.

## Installation

### 1. Build the Binary
You can build the binary directly on your Mac for the Raspberry Pi using the provided Makefile:
```bash
# To cross-compile for Raspberry Pi (ARM64)
make build-linux-arm64
```
The binary will be created at `bin/gnsstrack-linux-arm64`.

### 2. Configure the Service
Edit `config.yaml` to match your hardware setup. The default u-blox I2C address is `0x42`.
```yaml
i2c:
  bus: 1
  address: 0x42
logging:
  directory: "/var/log/gnsstrack"
```
*Note: Ensure the logging directory exists and the service has write permissions.*

### 3. Install the Systemd Service
1. Copy the binary to your Raspberry Pi:
   ```bash
   scp bin/gnsstrack-linux-arm64 pi@your-pi-ip:/tmp/gnsstrack
   ```
2. On the Raspberry Pi, move it to a standard location:
   ```bash
   sudo mv /tmp/gnsstrack /usr/local/bin/gnsstrack
   sudo chmod +x /usr/local/bin/gnsstrack
   ```
3. Create the configuration directory and copy the config:
   ```bash
   sudo mkdir -p /etc/gnsstrack
   sudo cp config.yaml /etc/gnsstrack/
   ```
4. Install and start the service:
   ```bash
   sudo cp gnsstrack.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable gnsstrack --now
   ```

## Usage

### Check Service Status
Query the background daemon for real-time state:
```bash
gnsstrack status
```

### Manual Run (Service Mode)
Run the daemon in the foreground for debugging:
```bash
gnsstrack service --config config.yaml
```

### Simulation Mode
Run the service with fake GNSS data generation:
```bash
gnsstrack service --simulate --config config.yaml
```
In this mode, logs will be written to `simulated_gnss_history.log`.

## Logging
Logs are written to the directory specified in `config.yaml`.
- **Production Logs**: `gnss_history.log`
- **Simulation Logs**: `simulated_gnss_history.log`

Example log entry (JSONL):
```json
{
  "timestamp": "2026-05-15T19:34:49Z",
  "anomalies_detected": false,
  "jamming_state": 0,
  "spoofing_state": 0,
  "lat": 51.5074,
  "lon": -0.1234,
  "alt_msl_m": 45.2,
  "h_acc_m": 1.2,
  "v_acc_m": 2.5,
  "sats_used": 12,
  "pdop": 1.1,
  "fix_type": 3,
  "sbas_used": true
}
```

## Hardware Compatibility
Tested on **Raspberry Pi CM4** with **Timebeat Timecard Mini (u-blox MAX-F10S)**. Requires I2C to be enabled in `/boot/config.txt`:
```text
dtparam=i2c_arm=on
```
