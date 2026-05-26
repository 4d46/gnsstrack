package rtc

import "4d46.uk/gnsstrack/i2c"

// DS3231 reads from a DS3231 RTC over I2C.
// Temperature registers: 0x11 (MSB, signed integer °C), 0x12 (LSB, bits 7:6 = 0.25°C each).
type DS3231 struct {
	bus i2c.I2CBus
}

func New(bus i2c.I2CBus) *DS3231 {
	return &DS3231{bus: bus}
}

func (d *DS3231) ReadTemperature() (float64, error) {
	buf := make([]byte, 2)
	if err := d.bus.Tx([]byte{0x11}, buf); err != nil {
		return 0, err
	}
	return float64(int8(buf[0])) + float64(buf[1]>>6)*0.25, nil
}
