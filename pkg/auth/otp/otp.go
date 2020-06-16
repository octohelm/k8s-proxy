package otp

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/go-courier/otp"
	"github.com/octohelm/k8s-proxy/pkg/auth"
)

func NewOTP(secret string, opts ...OTPOption) (*OTP, error) {
	s := OTP{
		otp: otp.NewOTP(secret, 9, nil),
	}

	defaults := make([]OTPOption, 0)

	for _, optFn := range append(defaults, opts...) {
		if err := optFn(&s); err != nil {
			return nil, err
		}
	}

	return &s, nil
}

type OTPOption func(opt *OTP) error

// OTP base64(code@timestamp)
type OTP struct {
	otp *otp.OTP
}

func (OTP) Type() string {
	return "OTP"
}

func (s *OTP) Sign(aud string) (string, error) {
	t := time.Now().Unix()
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s@%d", s.otp.GenerateOTP(int(t)), t))), nil
}

func (s *OTP) Validate(token string) (bool, error) {
	d, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return false, err
	}

	parts := bytes.Split(d, []byte("@"))

	if len(parts) != 2 {
		return false, auth.ErrInvalidToken
	}

	code := string(parts[0])

	timestamp, err := strconv.ParseUint(string(parts[1]), 10, 64)
	if err != nil {
		return false, auth.ErrInvalidToken
	}

	delta := time.Since(time.Unix(int64(timestamp), 0))
	if delta < 0 {
		delta = -delta
	}

	if delta > 30*time.Second {
		return false, auth.ErrInvalidToken
	}

	return s.otp.GenerateOTP(int(timestamp)) == code, nil
}
