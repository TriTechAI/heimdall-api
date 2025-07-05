package utils

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

const (
	// AccessTokenExpiration 访问令牌有效期（2小时）
	AccessTokenExpiration = 2 * time.Hour
	// RefreshTokenExpiration 刷新令牌有效期（7天）
	RefreshTokenExpiration = 7 * 24 * time.Hour
)

var (
	// ErrInvalidToken 无效令牌错误
	ErrInvalidToken = errors.New("invalid token")
	// ErrTokenExpired 令牌过期错误
	ErrTokenExpired = errors.New("token expired")
	// ErrTokenNotYetValid 令牌尚未生效错误
	ErrTokenNotYetValid = errors.New("token not yet valid")
	// ErrMalformedToken 令牌格式错误
	ErrMalformedToken = errors.New("malformed token")
	// ErrUnknownClaims 未知声明错误
	ErrUnknownClaims = errors.New("unknown claims type")
)

// JWTClaims JWT声明结构，遵循安全设计规范
type JWTClaims struct {
	UserID   string `json:"sub"`      // 用户ID (Subject)
	Username string `json:"username"` // 用户名
	Role     string `json:"role"`     // 用户角色
	TokenID  string `json:"jti"`      // 令牌唯一标识 (JWT ID)
	jwt.RegisteredClaims
}

// TokenPair 令牌对结构
type TokenPair struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
	TokenType    string    `json:"tokenType"`
}

// JWTManager JWT管理器
type JWTManager struct {
	secretKey []byte
	issuer    string
}

// NewJWTManager 创建JWT管理器
func NewJWTManager(secretKey, issuer string) *JWTManager {
	return &JWTManager{
		secretKey: []byte(secretKey),
		issuer:    issuer,
	}
}

// GenerateToken 生成访问令牌
func (j *JWTManager) GenerateToken(userID, username, role string) (*TokenPair, error) {
	if userID == "" || username == "" || role == "" {
		return nil, errors.New("userID, username and role cannot be empty")
	}

	now := time.Now()
	tokenID := uuid.New().String()

	// 创建访问令牌
	accessClaims := &JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		TokenID:  tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenExpiration)),
			NotBefore: jwt.NewNumericDate(now),
			ID:        tokenID,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(j.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// 创建刷新令牌（有效期更长，但不包含详细用户信息）
	refreshTokenID := uuid.New().String()
	refreshClaims := &JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		TokenID:  refreshTokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(RefreshTokenExpiration)),
			NotBefore: jwt.NewNumericDate(now),
			ID:        refreshTokenID,
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(j.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    now.Add(AccessTokenExpiration),
		TokenType:    "Bearer",
	}, nil
}

// ValidateToken 验证令牌并返回声明
func (j *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	if tokenString == "" {
		return nil, ErrInvalidToken
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// 确保使用正确的签名方法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return j.secretKey, nil
		},
	)

	if err != nil {
		// 具体化错误类型
		if ve, ok := err.(*jwt.ValidationError); ok {
			switch {
			case ve.Errors&jwt.ValidationErrorMalformed != 0:
				return nil, ErrMalformedToken
			case ve.Errors&jwt.ValidationErrorExpired != 0:
				return nil, ErrTokenExpired
			case ve.Errors&jwt.ValidationErrorNotValidYet != 0:
				return nil, ErrTokenNotYetValid
			default:
				return nil, ErrInvalidToken
			}
		}
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrUnknownClaims
	}

	return claims, nil
}

// RefreshToken 刷新访问令牌
func (j *JWTManager) RefreshToken(refreshTokenString string) (*TokenPair, error) {
	claims, err := j.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// 生成新的令牌对
	return j.GenerateToken(claims.UserID, claims.Username, claims.Role)
}

// ExtractUserIDFromToken 从令牌中提取用户ID
func (j *JWTManager) ExtractUserIDFromToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

// ExtractUsernameFromToken 从令牌中提取用户名
func (j *JWTManager) ExtractUsernameFromToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.Username, nil
}

// ExtractRoleFromToken 从令牌中提取用户角色
func (j *JWTManager) ExtractRoleFromToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.Role, nil
}

// ExtractTokenIDFromToken 从令牌中提取令牌ID
func (j *JWTManager) ExtractTokenIDFromToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.TokenID, nil
}

// GetTokenExpirationTime 获取令牌过期时间
func (j *JWTManager) GetTokenExpirationTime(tokenString string) (time.Time, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return time.Time{}, err
	}
	return claims.ExpiresAt.Time, nil
}

// IsTokenExpired 检查令牌是否过期
func (j *JWTManager) IsTokenExpired(tokenString string) bool {
	expirationTime, err := j.GetTokenExpirationTime(tokenString)
	if err != nil {
		return true // 如果无法获取过期时间，认为已过期
	}
	return time.Now().After(expirationTime)
}

// GetTokenRemainingTime 获取令牌剩余有效时间
func (j *JWTManager) GetTokenRemainingTime(tokenString string) (time.Duration, error) {
	expirationTime, err := j.GetTokenExpirationTime(tokenString)
	if err != nil {
		return 0, err
	}

	remaining := time.Until(expirationTime)
	if remaining < 0 {
		return 0, ErrTokenExpired
	}

	return remaining, nil
}

// ParseTokenWithoutValidation 解析令牌但不验证（用于从过期令牌中提取信息）
func (j *JWTManager) ParseTokenWithoutValidation(tokenString string) (*JWTClaims, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())

	token, _, err := parser.ParseUnverified(tokenString, &JWTClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, ErrUnknownClaims
	}

	return claims, nil
}

// GenerateSessionKey 生成会话存储键
func GenerateSessionKey(userID, tokenID string) string {
	return fmt.Sprintf("session:%s:%s", userID, tokenID)
}

// GenerateBlacklistKey 生成黑名单存储键
func GenerateBlacklistKey(tokenID string) string {
	return fmt.Sprintf("blacklist:%s", tokenID)
}

// ParseAuthHeader 解析Authorization头
func ParseAuthHeader(authHeader string) (string, error) {
	const bearerPrefix = "Bearer "

	if authHeader == "" {
		return "", errors.New("authorization header is empty")
	}

	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", errors.New("invalid authorization header format")
	}

	token := authHeader[len(bearerPrefix):]
	if token == "" {
		return "", errors.New("token is empty")
	}

	return token, nil
}

// ValidateTokenFormat 验证令牌格式（不验证签名）
func ValidateTokenFormat(tokenString string) error {
	if tokenString == "" {
		return errors.New("token is empty")
	}

	// JWT应该包含3个部分，用.分隔
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return ErrMalformedToken
	}

	// 检查每个部分是否为空
	for i, part := range parts {
		if len(part) == 0 {
			return fmt.Errorf("token part %d is empty", i+1)
		}
	}

	return nil
}

// CreateCustomClaims 创建自定义声明
func CreateCustomClaims(userID, username, role string, customData map[string]interface{}) jwt.MapClaims {
	now := time.Now()
	tokenID := uuid.New().String()

	claims := jwt.MapClaims{
		"sub":      userID,
		"username": username,
		"role":     role,
		"jti":      tokenID,
		"iat":      now.Unix(),
		"exp":      now.Add(AccessTokenExpiration).Unix(),
		"nbf":      now.Unix(),
	}

	// 添加自定义数据
	for key, value := range customData {
		claims[key] = value
	}

	return claims
}

// GetTokenAge 获取令牌已存在时间
func (j *JWTManager) GetTokenAge(tokenString string) (time.Duration, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}

	issuedAt := claims.IssuedAt.Time
	return time.Since(issuedAt), nil
}

// IsTokenRecentlyIssued 检查令牌是否为最近签发（用于防止重放攻击）
func (j *JWTManager) IsTokenRecentlyIssued(tokenString string, threshold time.Duration) (bool, error) {
	age, err := j.GetTokenAge(tokenString)
	if err != nil {
		return false, err
	}

	return age <= threshold, nil
}

// ExtractTokenMetadata 提取令牌元数据
func (j *JWTManager) ExtractTokenMetadata(tokenString string) (map[string]interface{}, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	metadata := map[string]interface{}{
		"userID":    claims.UserID,
		"username":  claims.Username,
		"role":      claims.Role,
		"tokenID":   claims.TokenID,
		"issuer":    claims.Issuer,
		"issuedAt":  claims.IssuedAt.Time,
		"expiresAt": claims.ExpiresAt.Time,
		"notBefore": claims.NotBefore.Time,
	}

	return metadata, nil
}

// GenerateGoZeroCompatibleToken 生成与go-zero JWT中间件兼容的令牌
func (j *JWTManager) GenerateGoZeroCompatibleToken(userID, username, role string) (string, error) {
	if userID == "" || username == "" || role == "" {
		return "", errors.New("userID, username and role cannot be empty")
	}

	now := time.Now()

	// go-zero JWT中间件期望的标准claims格式
	// 参考go-zero源码，它期望包含标准的JWT claims
	claims := jwt.MapClaims{
		// 标准JWT claims
		"iss": j.issuer,                              // 发行者
		"sub": userID,                                // 主题（用户ID）- 这是关键字段
		"aud": "heimdall-admin",                      // 受众
		"exp": now.Add(AccessTokenExpiration).Unix(), // 过期时间
		"iat": now.Unix(),                            // 签发时间
		"nbf": now.Unix(),                            // 生效时间
		"jti": uuid.New().String(),                   // JWT ID

		// 自定义字段
		"username": username,
		"role":     role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign go-zero compatible token: %w", err)
	}

	return tokenString, nil
}

// ValidateGoZeroCompatibleToken 验证go-zero兼容的令牌
func (j *JWTManager) ValidateGoZeroCompatibleToken(tokenString string) (jwt.MapClaims, error) {
	if tokenString == "" {
		return nil, ErrInvalidToken
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse go-zero compatible token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
