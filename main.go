package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

const (
	// Multicast configuration
	multicastAddr = "239.252.0.0:6168"
	
	// Packet header constants (in hex)
	HDR1   = 0x2381D765  // HDR1 4 bytes
	HDR2   = 0x10B32FE1  // HDR2 4 bytes
	FAMILY = 0x00000080  // FAMILY 4 bytes
)

// TimePacket represents the UDP packet structure
type TimePacket struct {
	HDR1     uint32    // 4 bytes - Header 1
	HDR2     uint32    // 4 bytes - Header 2
	RSRV1    uint16    // 2 bytes - Reserved (zero-filled)
	DEVICE   uint16    // 2 bytes - Device MAC or Control Source ID
	FAMILY   uint32    // 4 bytes - Family identifier
	RSRV2    [3]byte   // 3 bytes - Reserved (zero-filled)
	ZEROS    byte      // 1 byte - Leading zeros setting
	RSRV3    [24]byte  // 24 bytes - Reserved (zero-filled)
	CTRLCODE byte      // 1 byte - Control code
	H        byte      // 1 byte - Hours (hex)
	M        byte      // 1 byte - Minutes (hex)
	S        byte      // 1 byte - Seconds (hex)
}

func main() {
	// Resolve multicast address
	addr, err := net.ResolveUDPAddr("udp4", multicastAddr)
	if err != nil {
		log.Fatal("Error resolving multicast address:", err)
	}

	// Create UDP connection
	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		log.Fatal("Error creating UDP connection:", err)
	}
	defer conn.Close()

	fmt.Printf("Sending time packets to multicast address %s\n", multicastAddr)
	fmt.Println("Press Ctrl+C to stop...")

	// Send packets every second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		
		// Create packet with current time
		packet := TimePacket{
			HDR1:     HDR1,
			HDR2:     HDR2,
			RSRV1:    0,              // Zero-filled
			DEVICE:   0x1234,         // Example device ID (you can modify this)
			FAMILY:   FAMILY,
			RSRV2:    [3]byte{0, 0, 0}, // Zero-filled
			ZEROS:    0x00,           // Leading zeros ON
			RSRV3:    [24]byte{},     // Zero-filled by default
			CTRLCODE: 0x02,           // Clock displays H:M:S values
			H:        byte(now.Hour()),
			M:        byte(now.Minute()),
			S:        byte(now.Second()),
		}

		// Convert packet to bytes
		data, err := packetToBytes(packet)
		if err != nil {
			log.Printf("Error converting packet to bytes: %v", err)
			continue
		}

		// Send packet
		_, err = conn.Write(data)
		if err != nil {
			log.Printf("Error sending packet: %v", err)
			continue
		}

		fmt.Printf("Sent time packet: %02d:%02d:%02d (H:0x%02X M:0x%02X S:0x%02X)\n", 
			now.Hour(), now.Minute(), now.Second(),
			packet.H, packet.M, packet.S)
	}
}

// packetToBytes converts the TimePacket struct to a byte slice
func packetToBytes(packet TimePacket) ([]byte, error) {
	// Calculate total packet size
	// 4+4+2+2+4+3+1+24+1+1+1+1 = 48 bytes
	data := make([]byte, 48)
	offset := 0

	// HDR1 (4 bytes, big-endian)
	binary.BigEndian.PutUint32(data[offset:], packet.HDR1)
	offset += 4

	// HDR2 (4 bytes, big-endian)
	binary.BigEndian.PutUint32(data[offset:], packet.HDR2)
	offset += 4

	// RSRV1 (2 bytes, big-endian)
	binary.BigEndian.PutUint16(data[offset:], packet.RSRV1)
	offset += 2

	// DEVICE (2 bytes, big-endian)
	binary.BigEndian.PutUint16(data[offset:], packet.DEVICE)
	offset += 2

	// FAMILY (4 bytes, big-endian)
	binary.BigEndian.PutUint32(data[offset:], packet.FAMILY)
	offset += 4

	// RSRV2 (3 bytes)
	copy(data[offset:offset+3], packet.RSRV2[:])
	offset += 3

	// ZEROS (1 byte)
	data[offset] = packet.ZEROS
	offset += 1

	// RSRV3 (24 bytes)
	copy(data[offset:offset+24], packet.RSRV3[:])
	offset += 24

	// CTRLCODE (1 byte)
	data[offset] = packet.CTRLCODE
	offset += 1

	// H (1 byte)
	data[offset] = packet.H
	offset += 1

	// M (1 byte)
	data[offset] = packet.M
	offset += 1

	// S (1 byte)
	data[offset] = packet.S

	return data, nil
}
