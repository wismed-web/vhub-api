package admin

import (
	"regexp"
	"strings"
)

// wildcard to regular expression
func wc2re(wc string) *regexp.Regexp {
	if len(wc) == 0 {
		return nil
	}
	sb := strings.Builder{}
	for _, c := range wc {
		if c == '*' {
			sb.WriteString("[\\w\\d\u4e00-\u9fa5]*")
		} else {
			sb.WriteRune(c)
		}
	}
	return regexp.MustCompile("^" + sb.String() + "$")
}
