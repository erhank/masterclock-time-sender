package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	// Listen for multicast packets
	addr, err := net.ResolveUDPAddr("udp4", "239.252.0.0:6168")
	if err != nil {
		log.Fatal("Error resolving address:", err)
	}

	conn, err := net.ListenMulticastUDP("udp4", nil, addr)
	if err != nil {
		log.Fatal("Error listening on multicast:", err)
	}
	defer conn.Close()

	fmt.Println("Listening for time packets on 239.252.0.0:6168...")
	fmt.Println("Press Ctrl+C to stop")

	buffer := make([]byte, 1024)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading packet: %v", err)
			continue
		}

		if n >= 48 {
			analyzePacket(buffer[:n], clientAddr)
		} else {
			fmt.Printf("Received short packet (%d bytes) from %s\n", n, clientAddr)
		}
	}
}

func analyzePacket(data []byte, from net.Addr) {
	if len(data) < 48 {
		return
	}

	fmt.Printf("\n--- Packet from %s at %s ---\n", from, time.Now().Format("15:04:05"))
	fmt.Printf("Raw hex: %s\n", hex.EncodeToString(data[:48]))

	// Parse packet fields
	hdr1 := binary.BigEndian.Uint32(data[0:4])
	hdr2 := binary.BigEndian.Uint32(data[4:8])
	rsrv1 := binary.BigEndian.Uint16(data[8:10])
	device := binary.BigEndian.Uint16(data[10:12])
	family := binary.BigEndian.Uint32(data[12:16])
	zeros := data[19]
	ctrlcode := data[44]
	h := data[45]
	m := data[46]
	s := data[47]

	fmt.Printf("HDR1: 0x%08X (expected: 0x2381D765)\n", hdr1)
	fmt.Printf("HDR2: 0x%08X (expected: 0x10B32FE1)\n", hdr2)
	fmt.Printf("RSRV1: 0x%04X\n", rsrv1)
	fmt.Printf("DEVICE: 0x%04X\n", device)
	fmt.Printf("FAMILY: 0x%08X (expected: 0x00000080)\n", family)
	fmt.Printf("ZEROS: 0x%02X (0x00=ON, 0x01=OFF)\n", zeros)
	fmt.Printf("CTRLCODE: 0x%02X\n", ctrlcode)
	fmt.Printf("Time: %02d:%02d:%02d (H:0x%02X M:0x%02X S:0x%02X)\n", h, m, s, h, m, s)

	// Validate expected values
	valid := true
	if hdr1 != 0x2381D765 {
		fmt.Printf("WARNING: HDR1 mismatch!\n")
		valid = false
	}
	if hdr2 != 0x10B32FE1 {
		fmt.Printf("WARNING: HDR2 mismatch!\n")
		valid = false
	}
	if family != 0x00000080 {
		fmt.Printf("WARNING: FAMILY mismatch!\n")
		valid = false
	}
	if valid {
		fmt.Printf("✓ Packet format is correct!\n")
	}
}
