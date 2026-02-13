package helpers

import (
	"crypto/rand"
	"math/big"
	"time"
)

func GenerateID() int64 {
	now := time.Now()
	year := now.Year() % 100
	month := int(now.Month())

	nBig, _ := rand.Int(rand.Reader, big.NewInt(9000))
	randomNumber := int(nBig.Int64()) + 1000

	clientID := int64(year)*1000000 +
		int64(month)*10000 +
		int64(randomNumber)

	return clientID
}
