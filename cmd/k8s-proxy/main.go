package main

import (
	"fmt"
	"runtime"

	"github.com/octohelm/k8s-proxy/pkg/secureproxy"
	"github.com/octohelm/k8s-proxy/version"
	"k8s.io/klog"
)

func main() {
	printVersion()

	cc, err := secureproxy.ResolveKubeConfig()
	if err != nil {
		panic(err)
	}

	s, err := secureproxy.NewServer(cc, 0)
	if err != nil {
		panic(err)
	}

	_ = s.Serve()
}

func printVersion() {
	klog.Info(fmt.Sprintf("%s Version: %s", version.Name, version.Version))
	klog.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	klog.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
}
