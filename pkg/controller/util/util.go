package util

import (
	"crypto/md5"
	"encoding/hex"
)

// GetMD5Hash takes a string and returns its MD5 hash
func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
