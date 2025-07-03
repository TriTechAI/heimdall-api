package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const (
	// BCryptCost bcrypt加密成本因子，设置为12（推荐值）
	BCryptCost = 12
	// MinPasswordLength 密码最小长度
	MinPasswordLength = 8
	// MaxPasswordLength 密码最大长度
	MaxPasswordLength = 128
)

var (
	// ErrWeakPassword 弱密码错误
	ErrWeakPassword = errors.New("password does not meet strength requirements")
	// ErrPasswordTooShort 密码太短错误
	ErrPasswordTooShort = fmt.Errorf("password must be at least %d characters long", MinPasswordLength)
	// ErrPasswordTooLong 密码太长错误
	ErrPasswordTooLong = fmt.Errorf("password must be no more than %d characters long", MaxPasswordLength)
	// ErrPasswordInvalidChars 密码包含无效字符错误
	ErrPasswordInvalidChars = errors.New("password contains invalid characters")
	// ErrCommonPassword 常见弱密码错误
	ErrCommonPassword = errors.New("password is too common")
)

// 常见弱密码黑名单
var commonPasswords = map[string]bool{
	"password":    true,
	"123456":      true,
	"12345678":    true,
	"qwerty":      true,
	"abc123":      true,
	"password123": true,
	"admin":       true,
	"root":        true,
	"guest":       true,
	"test":        true,
	"user":        true,
	"123123":      true,
	"000000":      true,
	"111111":      true,
	"888888":      true,
	"666666":      true,
}

// PasswordStrengthConfig 密码强度配置
type PasswordStrengthConfig struct {
	MinLength      int
	MaxLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
	MinTypes       int // 至少需要包含的字符类型数量
}

// DefaultPasswordConfig 默认密码强度配置
var DefaultPasswordConfig = PasswordStrengthConfig{
	MinLength:      MinPasswordLength,
	MaxLength:      MaxPasswordLength,
	RequireUpper:   false, // 不强制要求，但通过MinTypes控制
	RequireLower:   false, // 不强制要求，但通过MinTypes控制
	RequireNumber:  false, // 不强制要求，但通过MinTypes控制
	RequireSpecial: false, // 不强制要求，但通过MinTypes控制
	MinTypes:       3,     // 至少包含3种类型
}

// HashPassword 对密码进行bcrypt加密
func HashPassword(password string) (string, error) {
	if err := ValidatePasswordStrength(password, DefaultPasswordConfig); err != nil {
		return "", err
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BCryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// VerifyPassword 验证明文密码与哈希密码是否匹配
func VerifyPassword(plainPassword, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}

// ValidatePasswordStrength 验证密码强度
func ValidatePasswordStrength(password string, config PasswordStrengthConfig) error {
	// 检查长度
	if len(password) < config.MinLength {
		return ErrPasswordTooShort
	}
	if len(password) > config.MaxLength {
		return ErrPasswordTooLong
	}

	// 检查是否为常见弱密码
	if commonPasswords[strings.ToLower(password)] {
		return ErrCommonPassword
	}

	// 统计字符类型
	var hasUpper, hasLower, hasNumber, hasSpecial bool
	var typeCount int

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			if !hasUpper {
				hasUpper = true
				typeCount++
			}
		case unicode.IsLower(char):
			if !hasLower {
				hasLower = true
				typeCount++
			}
		case unicode.IsDigit(char):
			if !hasNumber {
				hasNumber = true
				typeCount++
			}
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			if !hasSpecial {
				hasSpecial = true
				typeCount++
			}
		case !unicode.IsPrint(char):
			// 不允许不可打印字符
			return ErrPasswordInvalidChars
		}
	}

	// 检查强制要求的字符类型
	if config.RequireUpper && !hasUpper {
		return ErrWeakPassword
	}
	if config.RequireLower && !hasLower {
		return ErrWeakPassword
	}
	if config.RequireNumber && !hasNumber {
		return ErrWeakPassword
	}
	if config.RequireSpecial && !hasSpecial {
		return ErrWeakPassword
	}

	// 检查最少字符类型数量
	if typeCount < config.MinTypes {
		return ErrWeakPassword
	}

	return nil
}

// ValidatePasswordForUser 验证密码是否包含用户相关信息
func ValidatePasswordForUser(password, username, email string) error {
	passwordLower := strings.ToLower(password)

	// 检查是否包含用户名
	if username != "" && strings.Contains(passwordLower, strings.ToLower(username)) {
		return errors.New("password cannot contain username")
	}

	// 检查是否包含邮箱地址的本地部分
	if email != "" {
		if atIndex := strings.Index(email, "@"); atIndex > 0 {
			localPart := strings.ToLower(email[:atIndex])
			if strings.Contains(passwordLower, localPart) {
				return errors.New("password cannot contain email address")
			}
		}
	}

	return nil
}

// GetPasswordStrengthScore 获取密码强度评分 (0-100)
func GetPasswordStrengthScore(password string) int {
	if len(password) < MinPasswordLength {
		return 0
	}

	score := 0

	// 长度评分 (最多30分)
	lengthScore := min(len(password)*2, 30)
	score += lengthScore

	// 字符类型评分 (每种类型15分，最多60分)
	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char) && !hasUpper:
			hasUpper = true
			score += 15
		case unicode.IsLower(char) && !hasLower:
			hasLower = true
			score += 15
		case unicode.IsDigit(char) && !hasNumber:
			hasNumber = true
			score += 15
		case (unicode.IsPunct(char) || unicode.IsSymbol(char)) && !hasSpecial:
			hasSpecial = true
			score += 15
		}
	}

	// 复杂度评分 (最多10分)
	if len(password) >= 12 {
		score += 5
	}
	if isPatternComplex(password) {
		score += 5
	}

	// 扣分项：常见密码模式
	if isCommonPattern(password) {
		score -= 20
	}

	// 确保评分在0-100范围内
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// isPatternComplex 检查密码是否有复杂的模式
func isPatternComplex(password string) bool {
	// 检查是否不是简单的重复模式
	if isRepeatingPattern(password) {
		return false
	}

	// 检查是否不是简单的递增/递减模式
	if isSequentialPattern(password) {
		return false
	}

	return true
}

// isCommonPattern 检查是否是常见密码模式
func isCommonPattern(password string) bool {
	commonPatterns := []string{
		`^(\w)\1+$`,   // 重复字符 (aaa, 111)
		`^\d+$`,       // 纯数字
		`^[a-zA-Z]+$`, // 纯字母
		`^123+`,       // 以123开头
		`^abc+`,       // 以abc开头
		`qwerty`,      // 键盘序列
		`asdf`,        // 键盘序列
	}

	for _, pattern := range commonPatterns {
		if matched, _ := regexp.MatchString(pattern, strings.ToLower(password)); matched {
			return true
		}
	}

	return false
}

// isRepeatingPattern 检查是否是重复模式
func isRepeatingPattern(password string) bool {
	if len(password) < 3 {
		return false
	}

	// 检查连续3个或以上相同字符
	count := 1
	for i := 1; i < len(password); i++ {
		if password[i] == password[i-1] {
			count++
			if count >= 3 {
				return true
			}
		} else {
			count = 1
		}
	}

	return false
}

// isSequentialPattern 检查是否是连续模式
func isSequentialPattern(password string) bool {
	if len(password) < 3 {
		return false
	}

	// 检查连续3个或以上递增字符
	ascending := 0
	descending := 0

	for i := 1; i < len(password); i++ {
		if password[i] == password[i-1]+1 {
			ascending++
			descending = 0
		} else if password[i] == password[i-1]-1 {
			descending++
			ascending = 0
		} else {
			ascending = 0
			descending = 0
		}

		if ascending >= 2 || descending >= 2 {
			return true
		}
	}

	return false
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
