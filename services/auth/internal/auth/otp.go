package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"math/big"
)

const otpDigits = 6

func generateOTP() (string, error) {
	const digits = "0123456789"

	code := make([]byte, otpDigits)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", fmt.Errorf("generate otp: %w", err)
		}
		code[i] = digits[n.Int64()]
	}
	return string(code), nil
}

func hashOTP(otp string) string {
	sum := sha256.Sum256([]byte(otp))
	return hex.EncodeToString(sum[:])
}

func otpMatches(otp, want string) bool {
	got := hashOTP(otp)
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}
