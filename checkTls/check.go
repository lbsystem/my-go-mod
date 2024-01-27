package checkTls

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func Check(header []byte) (bool, int) {
	var packageLen uint16
	err := binary.Read(bytes.NewReader(header[3:5]), binary.BigEndian, &packageLen)
	fmt.Println(packageLen)
	if err != nil {
		fmt.Println(err.Error())
	}
	if header[0] == 0x16 && header[1] == 0x03 && (header[2] == 0x01 || header[2] == 0x03 || header[2] == 0x04) && packageLen > 128 && packageLen < 2048 {
		return true, int(packageLen)
	} else {
		return false, 0
	}
}
