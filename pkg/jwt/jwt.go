// Package jwt 提供 Access/Refresh 双 token 的签发与解析。
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenInvalid = errors.New("token 无效")
	ErrTokenExpired = errors.New("token 已过期")
)

// Claims 自定义声明，同时用于 Access 与 Refresh token。
type Claims struct {
	UserId   int64  `json:"user_id"`             // 用户 id
	Username string `json:"username"`            // 用户名
	TokenUse string `json:"token_use"`           // token 用途：access / refresh
	jwt.RegisteredClaims
}

// TokenPair 一次登录签发的 token 对。
type TokenPair struct {
	AccessToken  string // 访问令牌，短期
	RefreshToken string // 刷新令牌，长期
}

// Signer token 签发器。
type Signer struct {
	secret          []byte        // 签名密钥
	accessDuration  time.Duration // access token 有效期
	refreshDuration time.Duration // refresh token 有效期
}

// NewSigner 创建签发器。
func NewSigner(secret string, accessDuration, refreshDuration time.Duration) *Signer {
	return &Signer{
		secret:          []byte(secret),
		accessDuration:  accessDuration,
		refreshDuration: refreshDuration,
	}
}

// SignPair 为用户签发 Access/Refresh token 对。
func (s *Signer) SignPair(userId int64, username string) (*TokenPair, error) {
	access, err := s.sign(userId, username, "access", s.accessDuration)
	if err != nil {
		return nil, err
	}
	refresh, err := s.sign(userId, username, "refresh", s.refreshDuration)
	if err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

func (s *Signer) sign(userId int64, username, use string, duration time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserId:   userId,
		Username: username,
		TokenUse: use,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   username,
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

// Parse 校验并解析 token，返回声明。token 无效或过期返回相应错误。
func (s *Signer) Parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		return s.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}

// ParseAccess 校验并解析 access token（拒绝其他用途的 token）。
func (s *Signer) ParseAccess(tokenStr string) (*Claims, error) {
	claims, err := s.Parse(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.TokenUse != "access" {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}
