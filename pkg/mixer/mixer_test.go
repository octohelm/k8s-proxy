package mixer

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"
	"testing"

	otp2 "github.com/go-courier/otp"
)

func Test(t *testing.T) {
	m, _ := NewAESMixer([]byte(otp2.RandomSecret(32)))

	svc := http.Server{}
	svc.Addr = ":999"

	svc.Handler = NewMixerMiddlare(m)(
		http.HandlerFunc(func(rw http.ResponseWriter, request *http.Request) {
			d, _ := httputil.DumpRequest(request, true)
			fmt.Println(string(d))

			data := []byte(`{"status": "ok"}`)

			rw.Header().Set("Content-Type", "application/json; charset=utf-8")
			rw.Header().Set("Content-Length", strconv.FormatInt(int64(len(data)), 10))
			rw.WriteHeader(http.StatusOK)
			_, _ = rw.Write(data)
		}),
	)

	defer func() {
		_ = svc.Shutdown(context.Background())
	}()

	go func() {
		_ = svc.ListenAndServe()
	}()

	t.Run("GET", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "http://localhost"+svc.Addr, nil)

		t.Run("request", func(t *testing.T) {
			c := http.Client{}

			resp, _ := c.Do(req)
			data, _ := httputil.DumpResponse(resp, true)

			t.Log(string(data))
		})

		t.Run("request & decrypt", func(t *testing.T) {
			c := http.Client{}
			c.Transport = NewMixerWrapperFunc(m.WithMethod("aes/" + string(AESMethodOFB)))(http.DefaultTransport)

			resp, err := c.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			data, _ := httputil.DumpResponse(resp, true)
			t.Log(string(data))
		})
	})

	t.Run("POST", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "http://localhost"+svc.Addr, bytes.NewBufferString("hello"))
		req.Header.Set("Content-Type", "text/plain; charset=utf-8")

		t.Run("request & decrypt", func(t *testing.T) {
			c := http.Client{}
			c.Transport = NewMixerWrapperFunc(m)(http.DefaultTransport)

			resp, err := c.Do(req)
			if err != nil {
				return
			}

			defer resp.Body.Close()

			data, _ := httputil.DumpResponse(resp, true)
			t.Log(string(data))
		})
	})
}
