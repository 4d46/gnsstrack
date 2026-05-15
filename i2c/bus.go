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
	Tx(w, r []byte) error
	Close() error
}

// RealBus is a concrete implementation of I2CBus using periph.io.
type RealBus struct {
	closer func() error
	dev    *i2c.Dev
}

func NewRealBus(busNum int, addr uint16) (*RealBus, error) {
	if _, err := host.Init(); err != nil {
		return nil, err
	}

	bus, err := i2creg.Open(fmt.Sprintf("/dev/i2c-%d", busNum))
	if err != nil {
		return nil, err
	}

	return &RealBus{
		closer: bus.Close,
		dev:    &i2c.Dev{Bus: bus, Addr: addr},
	}, nil
}

func (b *RealBus) Tx(w, r []byte) error {
	return b.dev.Tx(w, r)
}

func (b *RealBus) Close() error {
	return b.closer()
}

// MockBus is a mock implementation for testing.
type MockBus struct {
	OnTx func(w, r []byte) error
}

func (m *MockBus) Tx(w, r []byte) error {
	if m.OnTx != nil {
		return m.OnTx(w, r)
	}
	return nil
}

func (m *MockBus) Close() error {
	return nil
}

// SimulatedBus generates fake u-blox DDC data for testing without hardware.
type SimulatedBus struct {
	ticker *time.Ticker
}

func NewSimulatedBus() *SimulatedBus {
	return &SimulatedBus{
		ticker: time.NewTicker(1 * time.Second),
	}
}

func (s *SimulatedBus) Tx(w, r []byte) error {
	if len(w) > 0 && w[0] == 0xFD {
		// We'll simulate a 100 byte buffer waiting (92 for PVT + 8 for SEC-SIG approx)
		if len(r) >= 2 {
			r[0] = 0x00
			r[1] = 100
		}
		return nil
	}

	if len(w) > 0 && w[0] == 0xFF {
		// Generate fake NAV-PVT (92 bytes payload + 8 header/checksum)
		pvtPayload := make([]byte, 92)
		binary.LittleEndian.PutUint16(pvtPayload[4:6], 2026) // Year
		pvtPayload[6] = 5                                    // Month
		pvtPayload[7] = 15                                   // Day
		pvtPayload[20] = 3                                   // FixType (3D)
		lon := int32(-1234567)
		lat := int32(51507400)
		binary.LittleEndian.PutUint32(pvtPayload[24:28], uint32(lon)) // Lon
		binary.LittleEndian.PutUint32(pvtPayload[28:32], uint32(lat)) // Lat

		pvtFrame := wrapUBX(0x01, 0x07, pvtPayload)
		copy(r, pvtFrame)
		return nil
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
