package crypto

import (
	"fmt"
	"io"

	argon2 "github.com/mdouchement/simple-argon2"
)

// Encrypt takes the given destination and master key and return a encrypter writer.
func Encrypt(dst io.WriteCloser, mk []byte) (io.WriteCloser, error) {
	nonce, err := argon2.GenerateRandomBytes(StreamNonceSize)
	if err != nil {
		return nil, err
	}

	if _, err := dst.Write(nonce); err != nil {
		return nil, fmt.Errorf("failed to write nonce: %v", err)
	}

	return NewWriter(streamKey(mk, nonce), dst)
}
