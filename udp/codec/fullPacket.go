package codec

import (
	"unsafe"
)

type FullPacket struct {
	IP              *IPHeaderPtr
	UDP             *UDPHeaderPtr
	PseudoHeader    *PseudoHeaderIPv4
	Gtemp           []byte
	IpID            uint16
	tmpPseudoHeader []byte
}

func (P *FullPacket) AddPayload(payload []byte) []byte {
	ipID++
	packet := P.Gtemp[:0]
	P.UDP.Length = htons(uint16(8 + len(payload))) // 8 bytes for UDP header
	P.UDP.Checksum = 0                             // Reset before calculating
	P.IP.Protocol = 17
	P.IP.ID = ipID
	P.IP.Checksum = 0

	P.IP.TotalLen = htons(uint16(28 + len(payload))) // 20 bytes for IP header, 8 for UDP
	P.IP.UpdateChecksum()
	P.UDP.Checksum = UDPChecksum(P.IP, P.UDP, payload)
	ipBytes := IPHeaderPtrToBytes(P.IP)[:]
	udpBytes := UDPHeaderPtrToBytes(P.UDP)[:]
	packet = append(packet, ipBytes...)
	// // Construct the full packet
	packet = append(packet, udpBytes...)
	packet = append(packet, payload...)
	return packet
}

func (P *FullPacket) UDPChecksum(payload []byte) uint16 {

	if P.PseudoHeader == nil {
		udpLen := 8 + len(payload)
		udpLenHigh := byte(udpLen >> 8)
		udpLenLow := byte(udpLen & 0x00FF)
		P.tmpPseudoHeader = []byte{
			P.IP.Src[0], P.IP.Src[1], P.IP.Src[2], P.IP.Src[3], // Source IP
			P.IP.Dst[0], P.IP.Dst[1], P.IP.Dst[2], P.IP.Dst[3], // Destination IP
			0,             // Zero
			P.IP.Protocol, // Protocol
			udpLenHigh,    // UDP length high byte
			udpLenLow,     // UDP length low byte
		}
		P.PseudoHeader = (*PseudoHeaderIPv4)(unsafe.Pointer(&P.tmpPseudoHeader[0]))
	}
	P.PseudoHeader.TCPUDPLength = P.UDP.Length
	P.tmpPseudoHeader = P.PseudoHeader.ToBytes()
	// Create a buffer of P.tmpPseudoHeader, UDP header, and payload
	buf := Gtemp[:0]
	buf = append(buf, P.tmpPseudoHeader...)
	buf = append(buf, UDPHeaderPtrToBytes(P.UDP)[:]...)
	buf = append(buf, payload...)

	return checksum11(buf)
}

func (P *FullPacket) UDPHeaderPtrToBytes(udp *UDPHeaderPtr) *[8]byte {
	return (*[8]byte)(unsafe.Pointer(udp))
}
