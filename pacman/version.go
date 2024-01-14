package pacman

import (
	"github.com/Jguer/go-alpm/v2"
)

func IsVersionNewer(oldVersion string, newVersion string) (bool, error) {
	result := alpm.VerCmp(oldVersion, newVersion)

	return result < 0, nil
}
