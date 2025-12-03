package common

import (
	"fmt"
)

func GetVersion() string {
	version, gitCommit, ok := GetModuleBuildInfo()
	if ok {
		return fmt.Sprintf("%s (git: %s)", version, gitCommit)
	}
	return "unknown"
}
