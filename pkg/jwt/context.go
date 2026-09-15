package jwt

import "context"

// context key 类型，避免与其他包的 key 冲突
type ctxKey int

const (
	claimsKey ctxKey = iota
)

// ContextWithClaims 将 token 声明写入 context
func ContextWithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

// ClaimsFromContext 从 context 读取 token 声明，不存在时返回 nil
func ClaimsFromContext(ctx context.Context) *Claims {
	claims, _ := ctx.Value(claimsKey).(*Claims)
	return claims
}

// UserIdFromContext 从 context 读取用户 id，未登录返回 0
func UserIdFromContext(ctx context.Context) int64 {
	if claims := ClaimsFromContext(ctx); claims != nil {
		return claims.UserId
	}
	return 0
}
