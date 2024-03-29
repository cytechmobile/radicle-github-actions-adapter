package version

import (
	"fmt"
	"runtime/debug"
)

var Version = "development"
var BuildTime = ""

func GetRevision() string {
	var revision string
	var modified bool
	bi, ok := debug.ReadBuildInfo()
	if ok {
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				revision = s.Value
			case "vcs.modified":
				if s.Value == "true" {
					modified = true
				}
			}
		}
	}
	if revision == "" {
		return "unavailable"
	}
	if modified {
		return fmt.Sprintf("%s-dirty", revision)
	}
	return revision
}

func GetVersion() string {
	return Version
}

func GetBuildTime() string {
	return BuildTime
}
