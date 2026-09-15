// Package utils 提供跨服务复用的通用工具函数。
package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword 使用 bcrypt 加密明文密码
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// CheckPassword 校验明文密码与 bcrypt 哈希是否匹配
func CheckPassword(password, hashed string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)) == nil
}
