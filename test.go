package main

import (
	"encoding/hex"
	"fmt"
	"time"
)

func main() {
	// Test packet creation
	now := time.Now()
	packet := TimePacket{
		HDR1:     HDR1,
		HDR2:     HDR2,
		RSRV1:    0,
		DEVICE:   0x1234,
		FAMILY:   FAMILY,
		RSRV2:    [3]byte{0, 0, 0},
		ZEROS:    0x00,
		RSRV3:    [24]byte{},
		CTRLCODE: 0x02,
		H:        byte(now.Hour()),
		M:        byte(now.Minute()),
		S:        byte(now.Second()),
	}

	data, _ := packetToBytes(packet)
	fmt.Printf("Packet size: %d bytes\n", len(data))
	fmt.Printf("Packet hex: %s\n", hex.EncodeToString(data))
	fmt.Printf("Time: %02d:%02d:%02d\n", now.Hour(), now.Minute(), now.Second())
}
