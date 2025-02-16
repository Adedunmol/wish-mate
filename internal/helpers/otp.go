package helpers

import (
	"crypto/rand"
	"math/big"
)

func GenerateSecureOTP(length int) (string, error) {
	otp := ""
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10)) // Random digit (0-9)
		if err != nil {
			return "", err
		}
		otp += n.String()
	}
	return otp, nil
}
