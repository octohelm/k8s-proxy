package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/octohelm/k8s-proxy/pkg/httputil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

var (
	ErrInvalidToken = errors.New("invalid token or token expired")
)

type Validator interface {
	Type() string
	Validate(token string) (bool, error)
}

type Signer interface {
	Type() string
	Sign(id string) (string, error)
}

func NewSignerWrapperFunc(signer Signer) httputil.Transport {
	return func(rt http.RoundTripper) http.RoundTripper {
		return &SignerRoundTripper{signer: signer, rt: rt}
	}
}

type SignerRoundTripper struct {
	signer Signer
	rt     http.RoundTripper
}

func (s *SignerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := s.signer.Sign("")
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("%s %s", s.signer.Type(), token))

	return s.rt.RoundTrip(req)
}

func NewValidatorMiddleware(validator Validator) httputil.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			token := req.Header.Get("Authorization")

			writeErr := func(err error) {
				e := apierrors.NewUnauthorized(err.Error())

				rw.Header().Set("Content-Type", "application/json; charset=utf-8")
				rw.WriteHeader(int(e.ErrStatus.Code))

				_ = json.NewEncoder(rw).Encode(e.ErrStatus)
			}

			if token != "" {
				parts := strings.Split(token, " ")
				if len(parts) == 2 {
					tpe := parts[0]

					if strings.EqualFold(tpe, validator.Type()) {
						ok, err := validator.Validate(parts[1])
						if err != nil {
							writeErr(err)
							return
						}

						if !ok {
							writeErr(ErrInvalidToken)
							return
						}

						next.ServeHTTP(rw, req)
						return
					}
				}
			}

			writeErr(ErrInvalidToken)
		})
	}
}
