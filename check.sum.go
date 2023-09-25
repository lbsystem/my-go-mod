package main

import (
	"fmt"
	"time"
)



func updateChecksums(data []byte) {
	
	ipHeader := data[:20]
	ipHeader[10], ipHeader[11] = 0, 0  // Clear existing checksum
	ipChecksum := checksum(ipHeader)
	ipHeader[10] = byte(ipChecksum >> 8)
	ipHeader[11] = byte(ipChecksum & 0xFF)

	
	udpHeader := data[20:28]
	udpLength := uint16(udpHeader[4])<<8 | uint16(udpHeader[5])
	udpData := data[20 : 20+int(udpLength)]
	udpHeader[6], udpHeader[7] = 0, 0  // Clear existing checksum
	pseudoHeader := createPseudoHeader(data[12:16], data[16:20], udpLength)
	udpChecksum := checksum(append(pseudoHeader, udpData...))
	udpHeader[6] = byte(udpChecksum >> 8)
	udpHeader[7] = byte(udpChecksum & 0xFF)
}

func checksum(data []byte) uint16 {
	var sum uint32
	for i := 0; i < len(data); i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}
	sum = (sum >> 16) + (sum & 0xffff)
	sum += sum >> 16
	return uint16(^sum)
}

func createPseudoHeader(srcIP, dstIP []byte, udpLength uint16) []byte {
	header := make([]byte, 12)
	copy(header[0:4], srcIP)
	copy(header[4:8], dstIP)
	header[9] = 17  // Protocol (UDP)
	header[10] = byte(udpLength >> 8)
	header[11] = byte(udpLength & 0xFF)
	return header
}

func main() {
	// Assume data contains the complete UDP packet
	  // IP header (20 bytes)
	  ipHeader := []byte{
        0x45, 0x00, 0x00, 0x2c,  // Version, IHL, Type of Service, Total Length
        0x00, 0x01, 0x00, 0x00,  // Identification, Flags, Fragment Offset
        0x40, 0x11, 0x00, 0x00,  // TTL, Protocol, Header Checksum (will be updated later)
        0xc0, 0xa8, 0x01, 0x01,  // Source IP: 192.168.1.1
        0xc0, 0xa8, 0x01, 0x02,  // Destination IP: 192.168.1.2
    }

    // UDP header (8 bytes)
    udpHeader := []byte{
        0x30, 0x39, 0x30, 0x39,  // Source Port: 12345, Destination Port: 12345
        0x00, 0x18, 0x00, 0x00,  // Length (will be updated later), Checksum (will be updated later)
    }

    // Payload (12 bytes): "Hello, World!"
    payload := []byte("Hello, World!")

    // Combine all parts together
    packet := append(ipHeader, udpHeader...)
    packet = append(packet, payload...)
	now:=time.Now()
    // Update the checksums and lengths in the IP and UDP headers
	for i:=0;i<10000000;i++{
		updateChecksums(packet)
	}
	end:=time.Since(now)
	fmt.Println(end.Milliseconds())
   
	
}