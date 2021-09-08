package stream

import (
	"io"

	"filippo.io/age"
	"github.com/klauspost/compress/zstd"
)

type stream struct {
	close func() error
	r     io.Reader
	w     io.Writer
}

// NewReader returns an io.Reader that uses ZStandard compression and the STREAM chunked encryption scheme
func NewReader(password string, r io.ReadCloser) (io.ReadCloser, error) {
	identity, err := age.NewScryptIdentity(password)
	if err != nil {
		return nil, err
	}

	rE, err := age.Decrypt(r, identity)
	if err != nil {
		return nil, err
	}

	rC, err := zstd.NewReader(rE)
	if err != nil {
		return nil, err
	}

	return &stream{
		r: rC,
		close: func() error {
			rC.Close()
			return nil
		},
	}, nil
}

// NewWriter returns an io.WriteCloser that uses ZStandard compression and the STREAM chunked encryption scheme
func NewWriter(password string, w io.WriteCloser) (io.WriteCloser, error) {
	recipient, err := age.NewScryptRecipient(password)
	if err != nil {
		return nil, err
	}

	wE, err := age.Encrypt(w, recipient)
	if err != nil {
		return nil, err
	}

	wC, err := zstd.NewWriter(wE)
	if err != nil {
		return nil, err
	}

	return &stream{
		w: wC,
		close: func() error {
			defer wE.Close()
			defer wC.Close()

			if err := wC.Close(); err != nil {
				return err
			}
			return wE.Close()
		},
	}, nil
}

func (s *stream) Read(p []byte) (int, error) {
	return s.r.Read(p)
}

func (s *stream) Write(p []byte) (int, error) {
	return s.w.Write(p)
}

func (s *stream) Close() error {
	return s.close()
}
