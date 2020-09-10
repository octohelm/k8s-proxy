package secureproxy_test

import (
	"context"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/octohelm/k8s-proxy/pkg/secureproxy"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func Test(t *testing.T) {
	cc, err := secureproxy.ResolveKubeConfig()
	if err != nil {
		panic(err)
	}

	ctx := log.IntoContext(context.Background(), log.Log)

	svc, err := secureproxy.NewServer(ctx, cc, 0)
	if err != nil {
		panic(err)
	}

	go func() {
		_ = svc.Serve()
	}()

	c, err := secureproxy.ProxyConfig("http://localhost", secureproxy.ResolveKubeProxySecret())
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(c)
	if err != nil {
		panic(err)
	}

	list, err := clientset.CoreV1().Pods("aisys").List(context.Background(), v1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	spew.Dump(list)
}
