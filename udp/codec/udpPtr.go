package codec

import (
	"unsafe"
)

// UDPHeaderPtr 结构体定义
type UDPHeaderPtr struct {
	SrcPort  uint16
	DstPort  uint16
	Length   uint16
	Checksum uint16
}

// htons 将一个 uint16 从主机字节顺序转换为网络字节顺序
func htons(n uint16) uint16 {
	return (n&0xff)<<8 | n>>8
}

// NewUDPHeaderPtr 创建一个新的 UDPHeaderPtr 结构体并返回其指针
func NewUDPHeaderPtr(srcPort, dstPort uint16) *UDPHeaderPtr {
	return &UDPHeaderPtr{
		SrcPort:  htons(srcPort),
		DstPort:  htons(dstPort),
		Length:   0, // 会在后续设置
		Checksum: 0, // 会在后续设置
	}
}

type PseudoHeaderIPv4 struct {
	SourceIP      [4]byte
	DestinationIP [4]byte
	Zero          byte
	Protocol      byte
	TCPUDPLength  uint16
}

func (ph *PseudoHeaderIPv4) ToBytes() []byte {

	return (*[12]byte)(unsafe.Pointer(ph))[:]
}

var Gtemp = make([]byte, 0, 3500)
var ipID uint16 = 1

func AddPayload(ip *IPHeaderPtr, udp *UDPHeaderPtr, payload []byte) []byte {
	ipID++
	packet := Gtemp[:0]
	udp.Length = htons(uint16(8 + len(payload))) // 8 bytes for UDP header
	udp.Checksum = 0                             // Reset before calculating
	ip.Protocol = 17
	ip.ID = ipID
	ip.Checksum = 0
	ip.UpdateChecksum()
	// udp.Checksum = UDPChecksum(ip, udp, payload)

	ip.TotalLen = htons(uint16(28 + len(payload))) // 20 bytes for IP header, 8 for UDP

	ipBytes := IPHeaderPtrToBytes(ip)[:]
	udpBytes := UDPHeaderPtrToBytes(udp)[:]
	packet = append(packet, ipBytes...)
	// // Construct the full packet
	packet = append(packet, udpBytes...)
	packet = append(packet, payload...)
	return packet
}

var Gbuf = make([]byte, 0, 3500)
var GPseudoHeader *PseudoHeaderIPv4
var pseudoHeader []byte

func UDPChecksum(ip *IPHeaderPtr, udp *UDPHeaderPtr, payload []byte) uint16 {

	if GPseudoHeader == nil {
		udpLen := 8 + len(payload)
		udpLenHigh := byte(udpLen >> 8)
		udpLenLow := byte(udpLen & 0x00FF)
		pseudoHeader = []byte{
			ip.Src[0], ip.Src[1], ip.Src[2], ip.Src[3], // Source IP
			ip.Dst[0], ip.Dst[1], ip.Dst[2], ip.Dst[3], // Destination IP
			0,           // Zero
			ip.Protocol, // Protocol
			udpLenHigh,  // UDP length high byte
			udpLenLow,   // UDP length low byte
		}
		GPseudoHeader = (*PseudoHeaderIPv4)(unsafe.Pointer(&pseudoHeader[0]))
	}
	GPseudoHeader.TCPUDPLength = udp.Length
	pseudoHeader = GPseudoHeader.ToBytes()
	// Create a buffer of pseudoHeader, UDP header, and payload
	buf := Gbuf[:0]
	buf = append(buf, pseudoHeader...)
	buf = append(buf, UDPHeaderPtrToBytes(udp)[:]...)
	buf = append(buf, payload...)

	return checksum11(buf)
}

func UDPHeaderPtrToBytes(udp *UDPHeaderPtr) *[8]byte {
	return (*[8]byte)(unsafe.Pointer(udp))
}
