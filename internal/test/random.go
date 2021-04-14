package test

import (
	"crypto/rand"
	"math/big"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyz")

func GenerateName() string {
	b := make([]rune, 5)
	for i := range b {
		r, _ := rand.Int(rand.Reader, new(big.Int).SetInt64(int64(len(letters))))
		b[i] = letters[r.Int64()]
	}
	return string(b)
}
