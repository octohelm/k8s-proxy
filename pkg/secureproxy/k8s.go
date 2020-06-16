package secureproxy

import (
	"os"
	"time"

	"github.com/octohelm/k8s-proxy/pkg/proxy"
	"k8s.io/client-go/rest"
)

func ResolveKubeProxySecret() []byte {
	v := os.Getenv("KUBE_PROXY_SECRET")
	if v != "" {
		return []byte(v)
	}
	return []byte("FxsZE3Mpiy0rMUVqIzNkxM4GuOVgalOZ")
}

func NewServer(cfg *rest.Config, keepalive time.Duration) (*proxy.Server, error) {
	m, err := NewSecureMiddleware(ResolveKubeProxySecret())
	if err != nil {
		return nil, err
	}
	return proxy.NewServer(cfg, keepalive, m)
}

func ProxyConfig(host string, key []byte) (*rest.Config, error) {
	c := &rest.Config{
		Host: host,
	}

	wrapperFunc, err := NewSecureWrapperFunc(key)
	if err != nil {
		return nil, err
	}

	c.Wrap(wrapperFunc)

	return c, nil
}
