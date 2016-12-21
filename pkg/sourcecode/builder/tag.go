// Author: lipixun
// Created Time : 四 10/20 19:21:00 2016
//
// File Name: tag.go
// Description:
//
package builder

import (
	"crypto/rand"
	"encoding/hex"
)

func NewTag() (string, error) {
	rands := make([]byte, 8)
	if _, err := rand.Read(rands); err != nil {
		return "", err
	}
	return hex.EncodeToString(rands), nil
}
