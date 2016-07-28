package commands

import "strings"

func escape(s string) string {
	for _, c := range []string{"_", "*", "`", "~"} {
		s = strings.Replace(s, c, "\\"+c, -1)
	}
	return s
}
