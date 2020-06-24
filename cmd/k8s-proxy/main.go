package main

import (
	"context"
	"fmt"
	"runtime"

	"github.com/octohelm/k8s-proxy/pkg/secureproxy"
	"github.com/octohelm/k8s-proxy/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

func main() {
	printVersion()

	cc, err := secureproxy.ResolveKubeConfig()
	if err != nil {
		panic(err)
	}

	if err := validateConfig(cc); err != nil {
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

func validateConfig(c *rest.Config) error {
	clientset, err := kubernetes.NewForConfig(c)
	if err != nil {
		return err
	}

	if _, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{}); err != nil {
		return err
	}

	return nil
}
