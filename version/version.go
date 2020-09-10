package version

import (
	"fmt"
	"runtime"
)

var (
	Name    = "k8s-proxy"
	Version = "0.0.1"
)

func GetVersion() string {
	return fmt.Sprintf("%s (%s, %s)", Version, runtime.GOOS, runtime.GOARCH)
}
