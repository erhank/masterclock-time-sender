package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Key for encryption, matching the C implementation
var key = []byte{0x74, 0x12, 0x02, 0xfb, 0xcc, 0x24, 0x5b, 0x82, 0x61, 0xe7, 0x3f, 0x9a, 0x26, 0x7c, 0xd3, 0xa0, 0x42}

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

	// Create a stop channel to signal termination
    stop := make(chan struct{})

    // Set up signal handling
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    // Goroutine to handle termination signals
    go func() {
        <-sigChan
        fmt.Println("\nTermination signal received. Sending final packet...")

        // Create final packet with CTRLCODE=0 and H=0, M=0, S=0
        finalPacket := TimePacket{
            HDR1:     HDR1,
            HDR2:     HDR2,
            RSRV1:    0,
            DEVICE:   0x1234, // Assuming a device ID; adjust if needed
            FAMILY:   FAMILY,
            RSRV2:    [3]byte{0, 0, 0},
            ZEROS:    0x00,
            RSRV3:    [24]byte{},
            CTRLCODE: 0x00, // Set to 0 to indicate session end
            H:        0x00,
            M:        0x00,
            S:        0x00,
        }

        // Convert to bytes
        data, err := packetToBytes(finalPacket)
        if err != nil {
            log.Printf("Error creating final packet: %v", err)
        } else {
            // Encrypt the packet
            crypt(data)
            // Send the encrypted final packet
            _, err = conn.Write(data)
            if err != nil {
                log.Printf("Error sending final packet: %v", err)
            } else {
                fmt.Println("Sent final packet with CTRLCODE=0")
            }
        }

        // Signal the main loop to stop
        close(stop)
    }()

	// Send packets every second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Main loop: send packets until stop is signaled
loop:
    for {
        select {
        case <-ticker.C:
            now := time.Now()
            // Create regular packet with current time
            packet := TimePacket{
                HDR1:     HDR1,
                HDR2:     HDR2,
                RSRV1:    0,
                DEVICE:   0x1234,
                FAMILY:   FAMILY,
                RSRV2:    [3]byte{0, 0, 0},
                ZEROS:    0x00,
                RSRV3:    [24]byte{},
                CTRLCODE: 0x02, // Regular control code
                H:        byte(now.Hour()),
                M:        byte(now.Minute()),
                S:        byte(now.Second()),
            }

            // Convert to bytes
            data, err := packetToBytes(packet)
            if err != nil {
                log.Printf("Error converting packet to bytes: %v", err)
                continue
            }

            // Encrypt the packet
            crypt(data)

            // Send the encrypted packet
            _, err = conn.Write(data)
            if err != nil {
                log.Printf("Error sending packet: %v", err)
                continue
            }

            fmt.Printf("Sent encrypted time packet: %02d:%02d:%02d (H:0x%02X M:0x%02X S:0x%02X)\n",
                now.Hour(), now.Minute(), now.Second(),
                packet.H, packet.M, packet.S)

        case <-stop:
            break loop
        }
    }

    fmt.Println("Program terminated gracefully.")
}


// packetToBytes converts the TimePacket struct to a byte slice
func packetToBytes(packet TimePacket) ([]byte, error) {
    buf := new(bytes.Buffer)
    err := binary.Write(buf, binary.BigEndian, &packet)
    if err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
}

// crypt encrypts the buffer in place using a simple XOR-based stream cipher
func crypt(buf []byte) {
    padcnt := byte(1)
    keycnt := 0
    for i := range buf {
        buf[i] ^= padcnt ^ key[keycnt]
        keycnt++
        if keycnt == len(key) {
            keycnt = 0
        }
        padcnt++
        if padcnt == 254 {
            padcnt = 1
        }
    }
}
