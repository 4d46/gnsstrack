package ubx

import (
	"encoding/binary"
	"fmt"
)

const (
	Sync1 = 0xB5
	Sync2 = 0x62

	ClassNav = 0x01
	IDNavPVT = 0x07

	ClassSec = 0x27
	IDSecSig = 0x01
)

type NavPVT struct {
	Year      uint16
	Month     uint8
	Day       uint8
	Hour      uint8
	Min       uint8
	Sec       uint8
	FixType   uint8
	Flags     uint8
	NumSV     uint8
	Lon       int32 // 1e-7 deg
	Lat       int32 // 1e-7 deg
	Height    int32 // mm
	HMSL      int32 // mm
	HAcc      uint32
	VAcc      uint32
	GSpeed    int32 // mm/s
	HeadMot   int32 // 1e-5 deg
	PDOP      uint16
}

type SecSig struct {
	JamDetEnabled bool
	JammingState  uint8
	SpfDetEnabled bool
	SpoofingState uint8
}

// CalculateChecksum computes the 8-bit Fletcher checksum.
func CalculateChecksum(payload []byte) (uint8, uint8) {
	var ckA, ckB uint8
	for _, b := range payload {
		ckA += b
		ckB += ckA
	}
	return ckA, ckB
}

// Parse attempts to decode a single UBX frame from the buffer.
func Parse(data []byte) (interface{}, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("buffer too short")
	}

	if data[0] != Sync1 || data[1] != Sync2 {
		return nil, fmt.Errorf("invalid sync chars")
	}

	class := data[2]
	id := data[3]
	length := binary.LittleEndian.Uint16(data[4:6])

	if len(data) < int(length)+8 {
		return nil, fmt.Errorf("buffer too short for payload")
	}

	payload := data[6 : 6+length]
	ckA_recv := data[6+length]
	ckB_recv := data[7+length]

	ckA, ckB := CalculateChecksum(data[2 : 6+length])
	if ckA != ckA_recv || ckB != ckB_recv {
		return nil, fmt.Errorf("checksum mismatch")
	}

	switch {
	case class == ClassNav && id == IDNavPVT:
		return decodeNavPVT(payload)
	case class == ClassSec && id == IDSecSig:
		return decodeSecSig(payload)
	}

	return nil, fmt.Errorf("unsupported message class 0x%x ID 0x%x", class, id)
}

func decodeNavPVT(p []byte) (*NavPVT, error) {
	if len(p) < 92 {
		return nil, fmt.Errorf("NavPVT payload too short")
	}
	return &NavPVT{
		Year:    binary.LittleEndian.Uint16(p[4:6]),
		Month:   p[6],
		Day:     p[7],
		Hour:    p[8],
		Min:     p[9],
		Sec:     p[10],
		FixType: p[20],
		Flags:   p[21],
		NumSV:   p[23],
		Lon:     int32(binary.LittleEndian.Uint32(p[24:28])),
		Lat:     int32(binary.LittleEndian.Uint32(p[28:32])),
		Height:  int32(binary.LittleEndian.Uint32(p[32:36])),
		HMSL:    int32(binary.LittleEndian.Uint32(p[36:40])),
		HAcc:    binary.LittleEndian.Uint32(p[40:44]),
		VAcc:    binary.LittleEndian.Uint32(p[44:48]),
		GSpeed:  int32(binary.LittleEndian.Uint32(p[60:64])),
		HeadMot: int32(binary.LittleEndian.Uint32(p[64:68])),
		PDOP:    binary.LittleEndian.Uint16(p[76:78]),
	}, nil
}

func decodeSecSig(p []byte) (*SecSig, error) {
	if len(p) < 8 {
		return nil, fmt.Errorf("SecSig payload too short")
	}
	return &SecSig{
		JamDetEnabled: (p[1] & 0x01) != 0,
		JammingState:  p[2],
		SpfDetEnabled: (p[3] & 0x01) != 0,
		SpoofingState: p[4],
	}, nil
}

// EncodePoll creates an 8-byte UBX poll request for a given class and ID.
func EncodePoll(class, id byte) []byte {
	frame := make([]byte, 8)
	frame[0] = Sync1
	frame[1] = Sync2
	frame[2] = class
	frame[3] = id
	frame[4] = 0 // Length LS
	frame[5] = 0 // Length MS
	
	ckA, ckB := CalculateChecksum(frame[2:6])
	frame[6] = ckA
	frame[7] = ckB
	return frame
}
