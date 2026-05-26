package i2c

import (
	"encoding/binary"
	"fmt"
	"time"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

// I2CBus defines the interface for I2C communication.
type I2CBus interface {
	Tx(addr uint16, w, r []byte) error
	Close() error
}

// Device provides a scoped handle for a specific address on an I2CBus.
type Device struct {
	Bus  I2CBus
	Addr uint16
}

func (d *Device) Tx(w, r []byte) error {
	return d.Bus.Tx(d.Addr, w, r)
}

// RealBus is a concrete implementation of I2CBus using periph.io.
type RealBus struct {
	bus i2c.BusCloser
}

func NewRealBus(busNum int) (*RealBus, error) {
	if _, err := host.Init(); err != nil {
		return nil, err
	}

	bus, err := i2creg.Open(fmt.Sprintf("/dev/i2c-%d", busNum))
	if err != nil {
		return nil, err
	}

	return &RealBus{bus: bus}, nil
}

func (b *RealBus) Tx(addr uint16, w, r []byte) error {
	return b.bus.Tx(addr, w, r)
}

func (b *RealBus) Close() error {
	return b.bus.Close()
}

// MockBus is a mock implementation for testing.
type MockBus struct {
	OnTx func(addr uint16, w, r []byte) error
}

func (m *MockBus) Tx(addr uint16, w, r []byte) error {
	if m.OnTx != nil {
		return m.OnTx(addr, w, r)
	}
	return nil
}

func (m *MockBus) Close() error {
	return nil
}

// SimulatedBus generates fake data for testing without hardware.
type SimulatedBus struct {
	ticker *time.Ticker
}

func NewSimulatedBus() *SimulatedBus {
	return &SimulatedBus{
		ticker: time.NewTicker(1 * time.Second),
	}
}

func (s *SimulatedBus) Tx(addr uint16, w, r []byte) error {
	// Handle RTC (0x68)
	if addr == 0x68 {
		if len(w) > 0 && w[0] == 0x11 && len(r) >= 2 {
			r[0] = 23   // 23 °C integer part
			r[1] = 0xC0 // 0.75 fractional part (0xC0 >> 6 = 3; 3 * 0.25 = 0.75)
		}
		return nil
	}

	// Handle GNSS (0x42)
	if addr == 0x42 {
		if len(w) > 0 && w[0] == 0xFD {
			if len(r) >= 2 {
				r[0] = 0x00
				r[1] = 100
			}
			return nil
		}

		if len(w) > 0 && w[0] == 0xFF {
			pvtPayload := make([]byte, 92)
			binary.LittleEndian.PutUint16(pvtPayload[4:6], 2026)
			pvtPayload[6] = 5
			pvtPayload[7] = 15
			pvtPayload[20] = 3
			lon := int32(-1234567)
			lat := int32(51507400)
			binary.LittleEndian.PutUint32(pvtPayload[24:28], uint32(lon))
			binary.LittleEndian.PutUint32(pvtPayload[28:32], uint32(lat))

			pvtFrame := wrapUBX(0x01, 0x07, pvtPayload)
			copy(r, pvtFrame)
			return nil
		}
	}

	return nil
}

func wrapUBX(class, id byte, payload []byte) []byte {
	length := uint16(len(payload))
	frame := make([]byte, 6+len(payload)+2)
	frame[0] = 0xB5
	frame[1] = 0x62
	frame[2] = class
	frame[3] = id
	binary.LittleEndian.PutUint16(frame[4:6], length)
	copy(frame[6:], payload)

	var ckA, ckB uint8
	for _, b := range frame[2 : 6+length] {
		ckA += b
		ckB += ckA
	}
	frame[6+length] = ckA
	frame[7+length] = ckB
	return frame
}

func (s *SimulatedBus) Close() error {
	s.ticker.Stop()
	return nil
}
