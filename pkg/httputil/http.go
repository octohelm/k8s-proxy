package httputil

import (
	"net/http"

	"k8s.io/client-go/transport"
)

type MiddlewareFunc = func(next http.Handler) http.Handler

func WithMiddlewares(fns ...MiddlewareFunc) MiddlewareFunc {
	return func(handler http.Handler) http.Handler {
		base := handler
		for i := range fns {
			fn := fns[len(fns)-1-i]
			if fn != nil {
				base = fn(base)
			}
		}
		return base
	}
}

type Transport = transport.WrapperFunc

func WithTransports(fns ...Transport) Transport {
	return func(rt http.RoundTripper) http.RoundTripper {
		base := rt
		for i := range fns {
			fn := fns[len(fns)-1-i]
			if fn != nil {
				base = fn(base)
			}
		}
		return base
	}
}
