package secureproxy

import (
	"encoding/base32"

	"github.com/octohelm/k8s-proxy/pkg/auth"
	"github.com/octohelm/k8s-proxy/pkg/auth/otp"
	"github.com/octohelm/k8s-proxy/pkg/httputil"
	"github.com/octohelm/k8s-proxy/pkg/mixer"
)

func NewSecureMiddleware(key []byte) (httputil.MiddlewareFunc, error) {
	o, err := otp.NewOTP(base32.StdEncoding.EncodeToString(key))
	if err != nil {
		return nil, err
	}

	m, err := mixer.NewAESMixer(key)
	if err != nil {
		return nil, err
	}

	return httputil.WithMiddlewares(auth.NewValidatorMiddleware(o), mixer.NewMixerMiddlare(m)), nil
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
