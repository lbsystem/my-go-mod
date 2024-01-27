package checkTls

import "regexp"

func SSHCheck(header []byte) bool {

	flags := string(header[:5])
	match, _ := regexp.MatchString("^SSH-[123]", flags)
	if match {
		return true
	} else {
		return false
	}
}
