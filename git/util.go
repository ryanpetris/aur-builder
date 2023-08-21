package git

import "strings"

func CleanBranchName(branch string) string {
	return strings.Replace(branch, ":", "-", -1)
}
