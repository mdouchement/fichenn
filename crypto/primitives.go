package crypto

import (
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

// StreamNonceSize is the size of the nonce used during streaming.
const StreamNonceSize = 16

func streamKey(key, nonce []byte) []byte {
	h := hkdf.New(sha256.New, key, nonce, []byte("payload"))
	streamKey := make([]byte, chacha20poly1305.KeySize)
	if _, err := io.ReadFull(h, streamKey); err != nil {
		panic("internal error: failed to read from HKDF: " + err.Error())
	}
	return streamKey
}
