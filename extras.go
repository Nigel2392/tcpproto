package tcpproto

import (
	"strings"
)

func TrimExtraSpaces(s string) string {
	var res string
	var lastChar rune
	for _, char := range s {
		if char != ' ' || lastChar != ' ' {
			res += string(char)
		}
		lastChar = char
	}
	res = strings.TrimPrefix(res, " ")
	res = strings.TrimSuffix(res, " ")
	return res
}
