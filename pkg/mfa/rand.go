package mfa

import "crypto/rand"

// randomByte 密码学安全随机字节
func randomByte() byte {
	b := make([]byte, 1)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand 不可用: " + err.Error())
	}
	return b[0]
}
