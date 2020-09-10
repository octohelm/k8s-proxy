package main

import (
	"context"

	"github.com/octohelm/k8s-proxy/pkg/secureproxy"
	"github.com/octohelm/k8s-proxy/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	log.SetLogger(zap.New(zap.UseDevMode(true)))

	l := log.Log.WithValues(version.Name, version.GetVersion())

	ctx := log.IntoContext(context.Background(), l)

	cc, err := secureproxy.ResolveKubeConfig()
	if err != nil {
		panic(err)
	}

	if err := validateConfig(cc); err != nil {
		panic(err)
	}

	s, err := secureproxy.NewServer(ctx, cc, 0)
	if err != nil {
		panic(err)
	}

	_ = s.Serve()
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
