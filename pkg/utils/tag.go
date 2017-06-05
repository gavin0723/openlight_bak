// Author: lipixun
// File Name: tag.go
// Description:

package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// NewTag creates a new tag string
func NewTag() (string, error) {
	var dataBytes = make([]byte, 16, 16)
	if _, err := rand.Read(dataBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(dataBytes), nil
}
