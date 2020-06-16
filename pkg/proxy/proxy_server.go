package proxy

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/octohelm/k8s-proxy/pkg/httputil"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport"
	"k8s.io/klog"
)

func NewServer(cfg *rest.Config, keepalive time.Duration, middlewares ...httputil.MiddlewareFunc) (*Server, error) {
	host := cfg.Host
	if !strings.HasSuffix(host, "/") {
		host = host + "/"
	}
	target, err := url.Parse(host)

	if err != nil {
		return nil, err
	}

	responder := &responder{}

	t, err := rest.TransportFor(cfg)
	if err != nil {
		return nil, err
	}

	upgradeTransport, err := makeUpgradeTransport(cfg, keepalive)
	if err != nil {
		return nil, err
	}

	p := proxy.NewUpgradeAwareHandler(target, t, false, false, responder)
	p.UpgradeTransport = upgradeTransport
	p.UseRequestLocation = true

	proxyServer := http.Handler(p)

	if !strings.HasPrefix("/", "/api") {
		proxyServer = stripLeaveSlash("/", proxyServer)
	}

	mux := http.NewServeMux()
	mux.Handle("/", httputil.WithMiddlewares(middlewares...)(proxyServer))

	return &Server{handler: mux}, nil
}

// Server is a http.Handler which proxies Kubernetes APIs to remote API server.
type Server struct {
	handler http.Handler
}

// Serve loops forever.
func (s *Server) Serve() error {
	srv := http.Server{}

	srv.Handler = s.handler
	srv.Addr = ":80"

	go func() {
		klog.Infof("proxy listen on %s", srv.Addr)

		if err := srv.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				klog.Error(err)
			} else {
				klog.Fatal(err)
			}
		}
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)
	<-stopCh

	timeout := 10 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	klog.Infof("shutdowning in %s", timeout)

	return srv.Shutdown(ctx)
}

// like http.StripPrefix, but always leaves an initial slash. (so that our
// regexps will work.)
func stripLeaveSlash(prefix string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		p := strings.TrimPrefix(req.URL.Path, prefix)
		if len(p) >= len(req.URL.Path) {
			http.NotFound(w, req)
			return
		}
		if len(p) > 0 && p[:1] != "/" {
			p = "/" + p
		}
		req.URL.Path = p
		h.ServeHTTP(w, req)
	})
}

type responder struct{}

func (r *responder) Error(w http.ResponseWriter, req *http.Request, err error) {
	klog.Errorf("Error while proxying request: %v", err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// makeUpgradeTransport creates a transport that explicitly bypasses HTTP2 support
// for proxy connections that must upgrade.
func makeUpgradeTransport(config *rest.Config, keepalive time.Duration) (proxy.UpgradeRequestRoundTripper, error) {
	transportConfig, err := config.TransportConfig()
	if err != nil {
		return nil, err
	}
	tlsConfig, err := transport.TLSConfigFor(transportConfig)
	if err != nil {
		return nil, err
	}
	rt := utilnet.SetOldTransportDefaults(&http.Transport{
		TLSClientConfig: tlsConfig,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: keepalive,
		}).DialContext,
	})

	upgrader, err := transport.HTTPWrappersForConfig(transportConfig, proxy.MirrorRequest)
	if err != nil {
		return nil, err
	}
	return proxy.NewUpgradeRequestRoundTripper(rt, upgrader), nil
}
