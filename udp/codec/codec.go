package codec

import (
	"encoding/binary"
	"golang.org/x/net/ipv4"
	"math/rand"
	"net"
	"time"
)

var source = rand.NewSource(time.Now().UnixNano())
var localRand = rand.New(source)

type UDPHeader struct {
	SourcePort      uint16
	DestinationPort uint16
	Length          uint16
	Checksum        uint16
}

func checksum(data []byte) uint16 {
	var sum uint32
	for i := 0; i < len(data)-1; i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}
	if len(data)%2 == 1 {
		sum += uint32(data[len(data)-1])
	}
	for sum>>16 > 0 {
		sum = (sum & 0xFFFF) + (sum >> 16)
	}
	return uint16(^sum)
}

func BuildIPandUDPPacket(srcIP, dstIP *net.UDPAddr, ipID int, payload []byte) (*ipv4.Header, []byte) {
	ipHeader := &ipv4.Header{
		Version:  4,
		Len:      20,
		TotalLen: 20 + 8 + len(payload),
		ID:       ipID,
		Flags:    ipv4.DontFragment,
		TTL:      64,
		Protocol: 17,
		Src:      srcIP.IP.To4(),
		Dst:      dstIP.IP.To4(),
	}
	udpLen := uint16(8 + len(payload))
	udpHeader := make([]byte, 8)
	udpHeader[0] = byte(srcIP.Port >> 8)
	udpHeader[1] = byte(srcIP.Port & 0xFF)
	udpHeader[2] = byte(dstIP.Port >> 8)
	udpHeader[3] = byte(dstIP.Port & 0xFF)
	udpHeader[4] = byte(udpLen >> 8)
	udpHeader[5] = byte(udpLen & 0xFF)
	udpHeader[6] = 0
	udpHeader[7] = 0

	pseudoHeader := make([]byte, 12)
	copy(pseudoHeader[0:4], ipHeader.Src.To4())
	copy(pseudoHeader[4:8], ipHeader.Dst.To4())
	pseudoHeader[9] = 17
	pseudoHeader[10] = udpHeader[4]
	pseudoHeader[11] = udpHeader[5]

	checksumData := append(pseudoHeader, udpHeader...)
	checksumData = append(checksumData, payload...)
	if len(checksumData)%2 == 1 {
		checksumData = append(checksumData, 0)
	}
	checksumValue := checksum(checksumData)
	udpHeader[6] = byte(checksumValue >> 8)
	udpHeader[7] = byte(checksumValue & 0xFF)
	return ipHeader, append(udpHeader, payload...)
}
func BuildUDPPacket(srcIP, dstIP *net.UDPAddr, payload []byte) []byte {
	udpLen := uint16(8 + len(payload))
	udpHeader := make([]byte, 8)
	udpHeader[0] = byte(srcIP.Port >> 8)
	udpHeader[1] = byte(srcIP.Port & 0xFF)
	udpHeader[2] = byte(dstIP.Port >> 8)
	udpHeader[3] = byte(dstIP.Port & 0xFF)
	udpHeader[4] = byte(udpLen >> 8)
	udpHeader[5] = byte(udpLen & 0xFF)
	udpHeader[6] = 0
	udpHeader[7] = 0
	pseudoHeader := make([]byte, 12)
	copy(pseudoHeader[0:4], srcIP.IP.To4())
	copy(pseudoHeader[4:8], dstIP.IP.To4())
	pseudoHeader[9] = 17
	pseudoHeader[10] = udpHeader[4]
	pseudoHeader[11] = udpHeader[5]
	checksumData := append(pseudoHeader, udpHeader...)
	checksumData = append(checksumData, payload...)
	if len(checksumData)%2 == 1 {
		checksumData = append(checksumData, 0)
	}
	checksumValue := checksum(checksumData)

	udpHeader[6] = byte(checksumValue >> 8)
	udpHeader[7] = byte(checksumValue & 0xFF)
	return append(udpHeader, payload...)
}

func ParseUDPHeader(data []byte) (UDPHeader, []byte) {
	header := UDPHeader{
		SourcePort:      binary.BigEndian.Uint16(data[0:2]),
		DestinationPort: binary.BigEndian.Uint16(data[2:4]),
		Length:          binary.BigEndian.Uint16(data[4:6]),
		Checksum:        binary.BigEndian.Uint16(data[6:8]),
	}
	return header, data[8:]
}

func GenerateRandomPort() int {
	return localRand.Intn(65535-1) + 1
}
