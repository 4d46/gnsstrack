package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"4d46.uk/gnsstrack/config"
	"4d46.uk/gnsstrack/i2c"
	"4d46.uk/gnsstrack/service"
	"gopkg.in/natefinch/lumberjack.v2"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "service":
		runService(os.Args[2:])
	case "status":
		runStatus(os.Args[2:])
	case "version":
		fmt.Printf("gnsstrack %s\n", version)
	case "help":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("GNSS Tracking Service (gnsstrack)")
	fmt.Println("\nUsage:")
	fmt.Println("  gnsstrack <command> [arguments]")
	fmt.Println("\nCommands:")
	fmt.Println("  service   Run the GNSS tracking daemon")
	fmt.Println("  status    Check the status of the running service")
	fmt.Println("  version   Print version information")
	fmt.Println("  help      Print this help message")
}

func runService(args []string) {
	fs := flag.NewFlagSet("service", flag.ExitOnError)
	configPath := fs.String("config", "config.yaml", "Path to config file")
	simulate := fs.Bool("simulate", false, "Simulate GNSS data instead of using real I2C")
	fs.Parse(args)

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Logger
	logFilename := "gnss_history.log"
	logDir := cfg.Logging.Directory
	if *simulate {
		logFilename = "simulated_gnss_history.log"
		if cfg.Logging.SimulationDirectory != "" {
			logDir = cfg.Logging.SimulationDirectory
		}
	}

	diskLogger := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, logFilename),
		MaxSize:    cfg.Logging.MaxSizeMB,
		MaxBackups: cfg.Logging.MaxBackups,
		LocalTime:  true,
		Compress:   true,
	}

	var bus i2c.I2CBus
	if *simulate {
		log.Printf("Starting in SIMULATION MODE (Logging to %s)", filepath.Join(cfg.Logging.Directory, logFilename))
		bus = i2c.NewSimulatedBus()
	} else {
		log.Printf("Starting gnsstrack service (Logging to %s)", filepath.Join(cfg.Logging.Directory, logFilename))
		bus, err = i2c.NewRealBus(cfg.I2C.Bus, uint16(cfg.I2C.Address))
		if err != nil {
			log.Fatalf("Failed to open I2C bus: %v", err)
		}
	}
	defer bus.Close()

	poller := service.NewPoller(cfg, bus, diskLogger)
	stopCh := make(chan struct{})

	// Start status HTTP server in a goroutine
	go poller.StartStatusServer(cfg.Status.ListenAddress)

	poller.Run(stopCh)
}

func runStatus(args []string) {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	configPath := fs.String("config", "config.yaml", "Path to config file")
	fs.Parse(args)

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config (needed for status address): %v", err)
	}

	resp, err := http.Get(fmt.Sprintf("http://%s/status", cfg.Status.ListenAddress))
	if err != nil {
		fmt.Printf("Error: Could not connect to service at %s. Is it running?\n", cfg.Status.ListenAddress)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	var status service.ServiceStatus
	if err := json.Unmarshal(body, &status); err != nil {
		log.Fatalf("Failed to parse status JSON: %v", err)
	}

	fmt.Printf("GNSSTRACK SERVICE STATUS\n")
	fmt.Printf("========================\n")
	fmt.Printf("Uptime:         %.1f seconds\n", status.UptimeSeconds)
	fmt.Printf("Logging Rate:   %s\n", status.LoggingRate)
	fmt.Printf("Logs Written:   %d\n", status.LogsWritten)
	fmt.Printf("Last Poll:      %v\n", status.LastPoll.Format("15:04:05"))
	if status.LatestGNSS != nil {
		fmt.Printf("Anomalies:      %v\n", status.LatestGNSS.Anomalies)
		fmt.Printf("SBAS Used:      %v\n", status.LatestGNSS.SBASUsed)
		fmt.Printf("Jamming:        %d\n", status.LatestGNSS.JammingState)
		fmt.Printf("Spoofing:       %d\n", status.LatestGNSS.SpoofingState)
	} else {
		fmt.Printf("GNSS Data:      No data yet\n")
	}
}
