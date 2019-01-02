package hash

import (
	"crypto/sha1"
	"encoding/hex"
)

func Sha1s(s string) string {
	r := sha1.Sum([]byte(s))
	return hex.EncodeToString(r[:])
}
