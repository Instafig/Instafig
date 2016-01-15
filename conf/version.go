package conf

import (
	"fmt"
)

const (
	VERSION_MAJOR = 0
	VERSION_MINOR = 1
	VERSION_PATCH = 0
	VERSION_STAGE = "dev" // dev/alpha/beta/rc<N>/release
)

//go:generate ../scripts/make-build-info

func VersionString() string {
	return fmt.Sprintf("%d.%d.%d-%s+%s",
		VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH,
		VERSION_STAGE, VERSION_BUILD_INFO)
}
