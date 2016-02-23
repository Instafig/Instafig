package conf

import (
	"fmt"
	"runtime"
)

var VERSION_BUILD_INFO = "build.dev-darwin-x86_64"

func init() {
	VERSION_BUILD_INFO = fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
}
