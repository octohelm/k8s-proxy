package mixer

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"strings"
)

type AESMethod string

const (
	AESMethodCFB AESMethod = "cfb"
	AESMethodOFB AESMethod = "ofb"
)

type AESMixerOption func(opt *AESMixer) error

func WithAESMethod(m AESMethod) AESMixerOption {
	return func(opt *AESMixer) error {
		if !(m == AESMethodOFB || m == AESMethodCFB) {
			return fmt.Errorf("invalid aes method %s", m)
		}
		opt.method = m
		return nil
	}
}

func NewAESMixer(key []byte, opts ...AESMixerOption) (Mixer, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	m := &AESMixer{
		block: block,
	}

	switch len(key) {
	case 16:
		m.size = 128
	case 24:
		m.size = 192
	case 32:
		m.size = 256
	}

	defaults := []AESMixerOption{
		WithAESMethod(AESMethodCFB),
	}

	for _, optFn := range append(defaults, opts...) {
		if err := optFn(m); err != nil {
			return nil, err
		}
	}

	return m, nil
}

type AESMixer struct {
	block  cipher.Block
	method AESMethod
	size   int
}

func (m AESMixer) WithMethod(method string) Mixer {
	parts := strings.Split(method, "/")
	if len(parts) == 2 {
		m.method = AESMethod(strings.ToLower(parts[1]))
	}
	return &m
}

func (m *AESMixer) Method() string {
	return "aes/" + string(m.method)
}

func (m *AESMixer) EncryptFor(writer io.Writer) (io.WriteCloser, error) {
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	var iv [aes.BlockSize]byte
	var stream cipher.Stream

	switch m.method {
	case AESMethodCFB:
		stream = cipher.NewCFBEncrypter(m.block, iv[:])
	case AESMethodOFB:
		stream = cipher.NewOFB(m.block, iv[:])
	default:
		return nil, fmt.Errorf("invalid aes method %s", m.method)
	}

	return &cipher.StreamWriter{S: stream, W: writer}, nil
}

func (m *AESMixer) DecryptFor(reader io.Reader) (io.Reader, error) {
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	var iv [aes.BlockSize]byte
	var stream cipher.Stream

	switch m.method {
	case AESMethodCFB:
		stream = cipher.NewCFBDecrypter(m.block, iv[:])
	case AESMethodOFB:
		stream = cipher.NewOFB(m.block, iv[:])
	default:
		return nil, fmt.Errorf("invalid aes method %s", m.method)
	}

	return &cipher.StreamReader{S: stream, R: reader}, nil
}
