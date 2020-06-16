package secureproxy

import (
	"math/rand"
	"time"
)

var letterRunes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

func RandomKey(length int) []byte {
	rand.Seed(time.Now().UnixNano())

	bytes := make([]byte, length)

	for i := range bytes {
		bytes[i] = byte(letterRunes[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(letterRunes))])
	}

	return bytes
}
