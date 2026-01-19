package shortener

import (
	"crypto/rand"
	"math/big"
)

const validChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateCode(n int) (string, error) {
	code := make([]byte, n)
	maxValue := big.NewInt(int64(len(validChars)))

	for i := range code {
		num, err := rand.Int(rand.Reader, maxValue)
		if err != nil {
			return "", ErrGenCode
		}
		code[i] = validChars[num.Int64()]
	}
	return string(code), nil
}
