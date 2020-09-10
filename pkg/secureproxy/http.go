package secureproxy

import (
	"context"
	"encoding/base32"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/octohelm/k8s-proxy/pkg/auth"
	"github.com/octohelm/k8s-proxy/pkg/auth/otp"
	"github.com/octohelm/k8s-proxy/pkg/httputil"
	"github.com/octohelm/k8s-proxy/pkg/mixer"
)

func NewSecureMiddleware(ctx context.Context, key []byte) (httputil.MiddlewareFunc, error) {
	o, err := otp.NewOTP(base32.StdEncoding.EncodeToString(key))
	if err != nil {
		return nil, err
	}

	m, err := mixer.NewAESMixer(key)
	if err != nil {
		return nil, err
	}

	return httputil.WithMiddlewares(auth.NewValidatorMiddleware(o), mixer.NewMixerMiddlare(m), logger(log.FromContext(ctx))), nil
}

func NewSecureWrapperFunc(key []byte) (httputil.Transport, error) {
	o, err := otp.NewOTP(base32.StdEncoding.EncodeToString(key))
	if err != nil {
		return nil, err
	}

	m, err := mixer.NewAESMixer(key)
	if err != nil {
		return nil, err
	}

	return httputil.WithTransports(auth.NewSignerWrapperFunc(o), mixer.NewMixerWrapperFunc(m)), nil
}

func logger(log logr.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			startedAt := time.Now()
			defer func() {
				log.Info(
					"",
					"method", req.Method,
					"uri", req.URL,
					"costs", time.Since(startedAt),
				)
			}()

			next.ServeHTTP(rw, req)
		})
	}
}
