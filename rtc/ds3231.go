package rtc

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Txer interface {
	Tx(w, r []byte) error
}

// DS3231 reads from a DS3231 RTC.
type DS3231 struct {
	dev Txer
}

func New(dev Txer) *DS3231 {
	return &DS3231{dev: dev}
}

func (d *DS3231) ReadTemperature() (float64, error) {
	// 1. Try I2C first
	if d.dev != nil {
		buf := make([]byte, 2)
		if err := d.dev.Tx([]byte{0x11}, buf); err == nil {
			return float64(int8(buf[0])) + float64(buf[1]>>6)*0.25, nil
		}
	}

	// 2. Fallback to sysfs (if kernel driver has claimed the device)
	return ReadTemperatureFromSysfs()
}

func ReadTemperatureFromSysfs() (float64, error) {
	// Look for ds3231 or rtc-ds1307 in /sys/class/hwmon
	hwmons, _ := filepath.Glob("/sys/class/hwmon/hwmon*/name")
	for _, namePath := range hwmons {
		name, err := os.ReadFile(namePath)
		if err != nil {
			continue
		}
		n := strings.TrimSpace(string(name))
		if n == "ds3231" || n == "rtc-ds1307" {
			tempPath := filepath.Join(filepath.Dir(namePath), "temp1_input")
			tempStr, err := os.ReadFile(tempPath)
			if err != nil {
				continue
			}
			tempVal, err := strconv.ParseFloat(strings.TrimSpace(string(tempStr)), 64)
			if err != nil {
				continue
			}
			return tempVal / 1000.0, nil
		}
	}

	return 0, fmt.Errorf("no RTC temperature sensor found in sysfs or I2C")
}
