package hmac

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Keyed-Hash Message Authentication Code (HMAC)

// SHA256 is HMAC SHA 256
func SHA256(message []byte, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	sum := mac.Sum(nil)
	dst := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(dst, sum)
	return dst
}
