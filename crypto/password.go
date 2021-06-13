package crypto

import (
	"crypto/rand"
	"math/big"
	mrand "math/rand"
	"time"
)

const (
	digits   = "0123456789"
	letters  = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	specials = "~=%/()[]#|" // Subset supported by Bash and ZSH
)

// NewPassword generates a new password for the given length.
func NewPassword(length int) string {
	pass := make([]byte, length)
	chars := []byte(letters + digits + specials)
	mrand.New(mrand.NewSource(time.Now().UnixNano())).Shuffle(len(chars), func(i, j int) {
		chars[i], chars[j] = chars[j], chars[i]
	})
	max := big.NewInt(int64(len(chars)))

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic(err) // should never occured because max >= 0
		}
		pass[i] = chars[int(n.Int64())]
	}
	return string(pass)
}
