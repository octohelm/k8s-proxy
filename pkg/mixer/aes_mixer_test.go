package mixer

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

func TestAESMixer(t *testing.T) {
	p, err := NewAESMixer([]byte(RandomKey(32)))
	NewWithT(t).Expect(err).To(BeNil())

	t.Run("encrypt", func(t *testing.T) {
		data := bytes.NewBuffer(nil)

		w, err := p.EncryptFor(data)
		if err != nil {
			panic(err)
		}
		defer w.Close()

		_, _ = io.Copy(w, bytes.NewBufferString(strings.Repeat("hello\n", 10)))

		t.Log("Encrypted:", data.String())

		t.Run("decrypted", func(t *testing.T) {
			decrypted, err := p.DecryptFor(data)
			if err != nil {
				panic(err)
			}
			_, _ = io.Copy(os.Stdout, decrypted)
		})
	})

	t.Run("encrypt as reader", func(t *testing.T) {
		data := bytes.NewBufferString("hello hello\n")

		rw, err := BufferReadWriter(func(w io.Writer) (io.Writer, error) {
			return p.EncryptFor(w)
		})

		if err != nil {
			panic(err)
		}

		r := CopyReader(data, rw)

		t.Log(io.Copy(os.Stdout, r))
	})
}
