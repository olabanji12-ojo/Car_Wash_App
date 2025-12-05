package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GenerateNumericCode generates a random numeric code of specified length
// For example, GenerateNumericCode(6) will generate a 6-digit code like "123456"
func GenerateNumericCode(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be greater than 0")
	}

	// Generate random digits
	code := ""
	for i := 0; i < length; i++ {
		// Generate a random number between 0 and 9
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %v", err)
		}
		code += fmt.Sprintf("%d", num.Int64())
	}

	return code, nil
}
