package crypto

import (
	"io"

	"github.com/pkg/errors"
)

// Decrypt takes the given source and master key and return a decrypter reader.
func Decrypt(src io.Reader, mk []byte) (io.Reader, error) {
	nonce := make([]byte, StreamNonceSize)
	if _, err := io.ReadFull(src, nonce); err != nil {
		return nil, errors.Wrap(err, "failed to read nonce")
	}

	return NewReader(streamKey(mk, nonce), src)
}
