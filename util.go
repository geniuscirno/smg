package smg

import (
	"strings"
)

func split2(s, sep string) (string, string, bool) {
	spl := strings.SplitN(s, sep, 2)
	if len(spl) < 2 {
		return spl[0], "", true
	}
	return spl[0], spl[1], true
}
