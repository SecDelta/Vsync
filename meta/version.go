package meta

import (
	"fmt"
)

var (
	commit  = "notpassed"
	version = "dev"
)

// GetVersion exposes the binary version and commit hash
func GetVersion() string {
	return fmt.Sprintf("%s\ngit commit hash %s", version, commit)
}
