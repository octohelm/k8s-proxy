package otp

import (
	"testing"

	"github.com/go-courier/otp"
	. "github.com/onsi/gomega"
)

func TestOTPAuth(t *testing.T) {
	secret := otp.RandomSecret(16)

	otpAuth, err := NewOTP(secret)
	NewWithT(t).Expect(err).To(BeNil())

	token, err := otpAuth.Sign("test")
	NewWithT(t).Expect(err).To(BeNil())

	ok, err := otpAuth.Validate(token)
	NewWithT(t).Expect(err).To(BeNil())
	NewWithT(t).Expect(ok).To(BeTrue())
}
