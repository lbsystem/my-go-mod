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

type EthHander struct {
	DestinationMAC [6]byte
	SourceMAC      [6]byte

	Proto [2]byte
}

type EthVLAN struct {
	DestinationMAC [6]byte
	SourceMAC      [6]byte
	TPID           [2]byte // Tag Protocol Identifier, often 0x8100 for VLAN
	TCI            [2]byte // Tag Control Information
	Proto          [2]byte // EtherType for the encapsulated frame
}

func (e *EthVLAN) VlanID() uint16 {
	highBits := uint16(e.TCI[0]) << 8
	lowBits := uint16(e.TCI[1])
	return (highBits | lowBits) & 0x0FFF
}

type UDPHeader struct {
	SourcePort      uint16
	DestinationPort uint16
	Length          uint16
	Checksum        uint16
}

type IPHeader struct {
	VersionIHL         uint8
	TypeOfService      uint8
	TotalLength        uint16
	Identification     uint16
	FlagsFragOffset    uint16
	TimeToLive         uint8
	Protocol           uint8
	HeaderChecksum     uint16
	SourceAddress      [4]byte
	DestinationAddress [4]byte
}

func (T *IPHeader) HeaderLen() uint8 {
	return (T.VersionIHL & 0b00001111) * 4
}
func (T *IPHeader) Version() uint8 {
	return (T.VersionIHL >> 4)
}
func (T *IPHeader) Flags() uint16 {
	return T.FlagsFragOffset >> 12
}
func (T *IPHeader) FragOffset() uint16 {
	return T.FlagsFragOffset & 0x0fff
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

func IPandUDPChecksums(data []byte) {
	// Recalculate IP checksum
	ipHeader := data[:20]
	ipHeader[10], ipHeader[11] = 0, 0 // Clear existing checksum
	ipChecksum := checksum1(ipHeader)
	ipHeader[10] = byte(ipChecksum >> 8)
	ipHeader[11] = byte(ipChecksum & 0xFF)

	// Recalculate UDP checksum
	udpHeader := data[20:28]
	udpLength := uint16(udpHeader[4])<<8 | uint16(udpHeader[5])
	udpData := data[20 : 20+int(udpLength)]
	udpHeader[6], udpHeader[7] = 0, 0 // Clear existing checksum
	pseudoHeader := createPseudoHeader(data[12:16], data[16:20], udpLength)

	// Handle odd length by adding a zero byte at the end
	if len(udpData)%2 != 0 {
		udpData = append(udpData, 0)
	}
	udpChecksum := checksum1(append(pseudoHeader, udpData...))
	udpHeader[6] = byte(udpChecksum >> 8)
	udpHeader[7] = byte(udpChecksum & 0xFF)
}

func checksum1(data []byte) uint16 {
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
	header[9] = 17 // Protocol (UDP)
	header[10] = byte(udpLength >> 8)
	header[11] = byte(udpLength & 0xFF)
	return header
}
