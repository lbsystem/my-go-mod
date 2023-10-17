package codec

import (
	"errors"
	"net"
	"unsafe"
)

type IPHeaderPtr struct {
	VersionIHL  byte
	TOS         byte
	TotalLen    uint16
	ID          uint16
	FlagsOffset uint16
	TTL         byte
	Protocol    byte
	Checksum    uint16
	Src         [4]byte
	Dst         [4]byte
}

func ByteToIPHeaderPtr(data []byte) *IPHeaderPtr {
	return (*IPHeaderPtr)(unsafe.Pointer(&data[0]))
}
func NewIPHeaderPtr(src, dst net.IP) (*IPHeaderPtr, error) {
	if len(src.To4()) != 4 || len(dst.To4()) != 4 {
		return nil, errors.New("invalid IPv4 address")
	}

	var header IPHeaderPtr
	header.VersionIHL = 0x45 // Version 4 and Header Length 5 (20 bytes)
	header.TTL = 64
	header.ID = 12322
	header.Checksum = 0
	copy(header.Src[:], src.To4())
	copy(header.Dst[:], dst.To4())

	// Set flags to 'Don't Fragment'
	header.FlagsOffset = 0b0000000001000000

	return &header, nil
}

func IPHeaderPtrToBytes(header *IPHeaderPtr) *[20]byte {
	return (*[20]byte)(unsafe.Pointer(header))
}
func Checksum(data [20]byte) uint16 {
	sum := uint32(0)
	for i := 0; i < len(data); i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}

	// 进位加回到尾部
	sum = (sum >> 16) + (sum & 0xFFFF)
	sum += (sum >> 16)

	// 取反
	return uint16(^sum)
}
func checksum11(data []byte) uint16 {
	var sum uint32
	length := len(data)
	for i := 0; i < length-1; i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}

	// Handle the case where the length of data is odd
	if length%2 == 1 {
		sum += uint32(data[length-1]) << 8
	}

	sum = (sum >> 16) + (sum & 0xffff)
	sum += sum >> 16
	sum2 := uint16(^sum)
	return (sum2<<8 | sum2>>8)
}

func (header *IPHeaderPtr) UpdateChecksum() {
	// 先设置校验和字段为0
	header.Checksum = 0
	bytes := IPHeaderPtrToBytes(header)

	header.Checksum = checksum11(bytes[:])
}
