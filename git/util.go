package git

import "strings"

func CleanTagName(branch string) string {
	return strings.Replace(branch, ":", "-", -1)
}
