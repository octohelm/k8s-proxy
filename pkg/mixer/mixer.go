package mixer

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/http"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/transport"
)

type Mixer interface {
	WithMethod(method string) Mixer
	Method() string
	EncryptFor(writer io.Writer) (io.WriteCloser, error)
	DecryptFor(r io.Reader) (io.Reader, error)
}

func NewMixerMiddlare(mixer Mixer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Body != nil && req.ContentLength > 0 {
				ok, method, originContentType := isBodyEncrypted(req.Header.Get("Content-Type"))
				if ok {
					reader, err := DecryptAsReader(mixer, req.Body, method)
					if err != nil {
						e := apierrors.NewBadRequest(err.Error())
						rw.WriteHeader(int(e.ErrStatus.Code))
						_ = json.NewEncoder(rw).Encode(e.ErrStatus)
						return
					}

					if originContentType != "" {
						req.Header.Set("Content-Type", originContentType)
					}

					req.Body = reader
				}
			}

			mixerRw := MixerResponseWriter(rw, mixer)

			next.ServeHTTP(mixerRw, req)

			if closer, ok := mixerRw.(io.WriteCloser); ok {
				closer.Close()
			}
		})
	}
}

func MixerResponseWriter(rw http.ResponseWriter, mixer Mixer) http.ResponseWriter {
	h, hok := rw.(http.Hijacker)
	if !hok {
		h = nil
	}

	f, fok := rw.(http.Flusher)
	if !fok {
		f = nil
	}

	return &mixerWriterWrapper{
		mixer:          mixer,
		ResponseWriter: rw,
		Hijacker:       h,
		Flusher:        f,
	}
}

type mixerWriterWrapper struct {
	mixer Mixer
	io.WriteCloser
	http.ResponseWriter
	http.Hijacker
	http.Flusher
}

func (m *mixerWriterWrapper) Header() http.Header {
	return m.ResponseWriter.Header()
}

func (m *mixerWriterWrapper) WriteHeader(statusCode int) {
	writeToHeader(m.mixer, m.Header())
	m.ResponseWriter.WriteHeader(statusCode)
}

func (m *mixerWriterWrapper) Write(b []byte) (int, error) {
	if m.WriteCloser == nil {
		writeCloser, err := m.mixer.EncryptFor(m.ResponseWriter)
		if err != nil {
			return 0, err
		}
		m.WriteCloser = writeCloser
	}
	return m.WriteCloser.Write(b)
}

func NewMixerWrapperFunc(mixer Mixer) transport.WrapperFunc {
	return func(rt http.RoundTripper) http.RoundTripper {
		return &MixerRoundTripper{mixer: mixer, rt: rt}
	}
}

type MixerRoundTripper struct {
	mixer Mixer
	rt    http.RoundTripper
}

func (s *MixerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil && req.ContentLength > 0 {
		reader, err := EncryptAsReadCloser(s.mixer, req.Body)
		if err != nil {
			return nil, err
		}
		writeToHeader(s.mixer, req.Header)
		req.Body = reader
		req.GetBody = nil
	}

	resp, err := s.rt.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	ok, method, originContentType := isBodyEncrypted(resp.Header.Get("Content-Type"))
	if !ok {
		return resp, nil
	}

	if originContentType != "" {
		resp.Header.Set("Content-Type", originContentType)
	}

	resp.Body, err = DecryptAsReader(s.mixer, resp.Body, method)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func DecryptAsReader(mixer Mixer, body io.ReadCloser, method string) (io.ReadCloser, error) {
	if method != "" {
		mixer = mixer.WithMethod(method)
	}

	r, err := mixer.DecryptFor(body)
	if err != nil {
		return nil, err
	}

	return ToReadCloser(r, body), nil
}

func EncryptAsReadCloser(mixer Mixer, body io.ReadCloser) (io.ReadCloser, error) {
	buf := bytes.NewBuffer(nil)
	w, err := mixer.EncryptFor(buf)
	if err != nil {
		return nil, err
	}
	return ToReadCloser(CopyReader(body, ToReadWriter(buf, w)), body), nil
}

func writeToHeader(mixer Mixer, h http.Header) {
	params := map[string]string{
		"encrypted-by": mixer.Method(),
	}

	contentType := h.Get("Content-Type")
	if contentType != "" {
		params["origin-content-type"] = contentType
	}

	h.Set("Content-Type", mime.FormatMediaType("application/octet-stream", params))
}

func isBodyEncrypted(contentType string) (bool, string, string) {
	ct, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false, "", ""
	}

	if ct != "application/octet-stream" {
		return false, "", ""
	}

	m, ok := params["encrypted-by"]
	if !ok {
		return false, "", ""
	}

	return true, m, params["origin-content-type"]
}
