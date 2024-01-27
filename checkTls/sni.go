package checkTls

import (
	"encoding/binary"
	"fmt"
)

func ExtractSNI(data []byte) (string, error) {
	// Skip 43 bytes of the ClientHello header.
	data = data[43:]

	// Parse the session ID
	sessionIDLength := int(data[0])
	data = data[1+sessionIDLength:]

	// Skip over cipher suites
	cipherSuitesLength := int(binary.BigEndian.Uint16(data))
	data = data[2+cipherSuitesLength:]

	// Skip over compression methods
	compressionMethodsLength := int(data[0])
	data = data[1+compressionMethodsLength:]

	// Now we're at the extensions
	extensionsLength := int(binary.BigEndian.Uint16(data))
	data = data[2 : 2+extensionsLength]

	// Loop through the extensions to find the SNI.
	for len(data) != 0 {
		if len(data) < 4 {
			return "", fmt.Errorf("extension too short")
		}
		extensionType := binary.BigEndian.Uint16(data)
		extensionLength := int(binary.BigEndian.Uint16(data[2:]))

		if extensionType == 0 /* server_name */ {
			// Skip 5 bytes of the server_name extension header.
			sniData := data[7:]

			// The SNI value is a length-prefixed string.
			sniLength := int(binary.BigEndian.Uint16(sniData))
			sniData = sniData[2 : 2+sniLength]

			return string(sniData), nil
		}

		data = data[4+extensionLength:]
	}

	return "", fmt.Errorf("SNI not found")
}
