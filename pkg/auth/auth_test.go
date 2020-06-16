package auth_test

import (
	"context"
	"net/http"
	nethttputil "net/http/httputil"
	"testing"

	otp2 "github.com/go-courier/otp"
	"github.com/octohelm/k8s-proxy/pkg/auth"
	"github.com/octohelm/k8s-proxy/pkg/auth/otp"
	"github.com/octohelm/k8s-proxy/pkg/httputil"
)

func Test(t *testing.T) {
	o, _ := otp.NewOTP(otp2.RandomSecret(32))

	svc := http.Server{}
	svc.Addr = ":999"

	svc.Handler = httputil.WithMiddlewares(auth.NewValidatorMiddleware(o))(http.HandlerFunc(func(rw http.ResponseWriter, request *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("hello"))
	}))

	defer func() {
		_ = svc.Shutdown(context.Background())
	}()

	go func() {
		_ = svc.ListenAndServe()
	}()

	req, _ := http.NewRequest(http.MethodGet, "http://localhost"+svc.Addr, nil)

	c := http.Client{}

	c.Transport = httputil.WithTransports(auth.NewSignerWrapperFunc(o))(http.DefaultTransport)

	resp, _ := c.Do(req)
	data, _ := nethttputil.DumpResponse(resp, true)
	t.Log(string(data))
}
