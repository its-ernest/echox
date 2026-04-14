package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// GenerateOTP generates a 6-digit numeric OTP as a string
func GenerateOTP() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func FormatPhone(phone string) string {
	// add + if missing
	if len(phone) > 0 && phone[0] != '+' {
		return "+" + phone
	}
	return phone
}
