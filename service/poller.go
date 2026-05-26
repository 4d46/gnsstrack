package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"4d46.uk/gnsstrack/config"
	"4d46.uk/gnsstrack/i2c"
	"4d46.uk/gnsstrack/ubx"
)

type Poller struct {
	cfg     *config.Config
	dev     *i2c.Device
	lastLog time.Time
	logger  io.Writer

	mu    sync.RWMutex
	state ServiceStatus
}

type ServiceStatus struct {
	StartTime          time.Time   `json:"start_time"`
	UptimeSeconds      float64     `json:"uptime_seconds"`
	LastPoll           time.Time   `json:"last_poll"`
	LastLogTime        time.Time   `json:"last_log_time"`
	LoggingRate        string      `json:"current_logging_rate"`
	LatestGNSS         *GNSSStatus `json:"latest_gnss"`
	LogsWritten        int         `json:"logs_written"`
	LatestTemperatureC *float64    `json:"latest_temperature_c,omitempty"`
}

func NewPoller(cfg *config.Config, dev *i2c.Device, logger io.Writer) *Poller {
	return &Poller{
		cfg:    cfg,
		dev:    dev,
		logger: logger,
		state: ServiceStatus{
			StartTime:   time.Now(),
			LoggingRate: "Normal",
		},
	}
}

type GNSSStatus struct {
	Timestamp     time.Time `json:"timestamp"`
	Anomalies     bool      `json:"anomalies_detected"`
	JammingState  int       `json:"jamming_state"`
	SpoofingState int       `json:"spoofing_state"`
	Latitude      float64   `json:"lat,omitempty"`
	Longitude     float64   `json:"lon,omitempty"`
	AltitudeMSL   float64   `json:"alt_msl_m,omitempty"`
	HorizontalAcc float64   `json:"h_acc_m,omitempty"`
	VerticalAcc   float64   `json:"v_acc_m,omitempty"`
	Satellites    int       `json:"sats_used,omitempty"`
	PDOP          float64   `json:"pdop,omitempty"`
	FixType       int       `json:"fix_type"`
	SBASUsed      bool      `json:"sbas_used"`
}

func (p *Poller) SetLatestTemperature(t float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state.LatestTemperatureC = &t
}

func (p *Poller) GetStatus() ServiceStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()

	status := p.state
	status.UptimeSeconds = time.Since(status.StartTime).Seconds()
	return status
}

func (p *Poller) Run(stopCh <-chan struct{}) {
	ticker := time.NewTicker(time.Duration(p.cfg.Polling.SoftwareRateMS) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			p.tick()
		}
	}
}

func (p *Poller) tick() {
	// 1. Read data from I2C
	status, err := p.pollGNSS()

	p.mu.Lock()
	p.state.LastPoll = time.Now()
	if err == nil && status != nil {
		p.state.LatestGNSS = status
	}
	p.mu.Unlock()

	if err != nil {
		log.Printf("I2C Poll Error: %v", err)
		return
	}

	if status == nil {
		return
	}

	// 2. Decide if we should log to disk
	now := time.Now()
	rate := time.Duration(p.cfg.Polling.NormalLoggingRateMS) * time.Millisecond
	currentRateName := "Normal"

	if status.Anomalies && p.cfg.Polling.EnableEnhancedOnAnomaly {
		rate = time.Duration(p.cfg.Polling.EnhancedLoggingRateMS) * time.Millisecond
		currentRateName = "Enhanced"
	}

	p.mu.Lock()
	p.state.LoggingRate = currentRateName
	p.mu.Unlock()

	if now.Sub(p.lastLog) >= rate {
		p.logToDisk(status)

		p.mu.Lock()
		p.state.LastLogTime = now
		p.state.LogsWritten++
		p.mu.Unlock()

		p.lastLog = now
	}
}

func (p *Poller) pollGNSS() (*GNSSStatus, error) {
	// 1. Send active poll requests for NAV-PVT and SEC-SIG
	pvtPoll := ubx.EncodePoll(ubx.ClassNav, ubx.IDNavPVT)
	secPoll := ubx.EncodePoll(ubx.ClassSec, ubx.IDSecSig)

	if err := p.dev.Tx(pvtPoll, nil); err != nil {
		return nil, fmt.Errorf("failed to send NAV-PVT poll: %v", err)
	}
	if err := p.dev.Tx(secPoll, nil); err != nil {
		return nil, fmt.Errorf("failed to send SEC-SIG poll: %v", err)
	}

	// 2. Wait briefly for the chip to process
	time.Sleep(50 * time.Millisecond)

	// 3. Read available byte count
	lenBuf := make([]byte, 2)
	if err := p.dev.Tx([]byte{0xFD}, lenBuf); err != nil {
		return nil, err
	}

	avail := int(lenBuf[0])<<8 | int(lenBuf[1])
	if avail == 0 || avail == 0xFFFF {
		return nil, nil
	}

	// Limit read size to a sane maximum (u-blox buffer is typically < 4KB,
	// and i2c-dev often has a 4KB limit per message)
	if avail > 4096 {
		avail = 4096
	}

	// 4. Read available data
	data := make([]byte, avail)
	if err := p.dev.Tx([]byte{0xFF}, data); err != nil {
		return nil, err
	}

	status := &GNSSStatus{
		Timestamp: time.Now(),
	}

	// 5. Robust parsing
	found := false
	for i := 0; i < len(data)-8; i++ {
		if data[i] == ubx.Sync1 && data[i+1] == ubx.Sync2 {
			msg, err := ubx.Parse(data[i:])
			if err != nil {
				continue
			}

			found = true
			switch m := msg.(type) {
			case *ubx.NavPVT:
				status.Latitude = float64(m.Lat) / 1e7
				status.Longitude = float64(m.Lon) / 1e7
				status.AltitudeMSL = float64(m.HMSL) / 1000.0 // mm to m
				status.HorizontalAcc = float64(m.HAcc) / 1000.0
				status.VerticalAcc = float64(m.VAcc) / 1000.0
				status.Satellites = int(m.NumSV)
				status.PDOP = float64(m.PDOP) / 100.0
				status.FixType = int(m.FixType)
				status.SBASUsed = (m.Flags & 0x02) != 0 // diffSoln bit
				status.Anomalies = (m.FixType < 3)
			case *ubx.SecSig:
				status.JammingState = int(m.JammingState)
				status.SpoofingState = int(m.SpoofingState)
				if status.JammingState > 1 || status.SpoofingState > 1 {
					status.Anomalies = true
				}
			}
		}
	}

	if !found {
		return nil, nil
	}

	return status, nil
}

func (p *Poller) logToDisk(status *GNSSStatus) {
	if p.logger == nil {
		return
	}

	data, err := json.Marshal(status)
	if err != nil {
		log.Printf("Failed to marshal GNSS status: %v", err)
		return
	}

	_, err = p.logger.Write(append(data, '\n'))
	if err != nil {
		log.Printf("Failed to write to log file: %v", err)
	}
}
