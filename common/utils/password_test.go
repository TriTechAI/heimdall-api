package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	Convey("Test HashPassword", t, func() {
		Convey("Valid password should be hashed successfully", func() {
			password := "TestPass123!"
			hash, err := HashPassword(password)

			So(err, ShouldBeNil)
			So(hash, ShouldNotBeEmpty)
			So(hash, ShouldNotEqual, password)

			// 验证哈希结果可以正常验证
			err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
			So(err, ShouldBeNil)
		})

		Convey("Weak password should return error", func() {
			password := "123"
			hash, err := HashPassword(password)

			So(err, ShouldNotBeNil)
			So(hash, ShouldBeEmpty)
		})

		Convey("Common password should return error", func() {
			password := "password"
			hash, err := HashPassword(password)

			So(err, ShouldEqual, ErrCommonPassword)
			So(hash, ShouldBeEmpty)
		})
	})
}

func TestVerifyPassword(t *testing.T) {
	Convey("Test VerifyPassword", t, func() {
		password := "TestPass123!"
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), BCryptCost)
		hashString := string(hash)

		Convey("Correct password should verify successfully", func() {
			err := VerifyPassword(password, hashString)
			So(err, ShouldBeNil)
		})

		Convey("Wrong password should return error", func() {
			err := VerifyPassword("wrongpassword", hashString)
			So(err, ShouldNotBeNil)
		})

		Convey("Empty password should return error", func() {
			err := VerifyPassword("", hashString)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestValidatePasswordStrength(t *testing.T) {
	Convey("Test ValidatePasswordStrength", t, func() {
		config := DefaultPasswordConfig

		Convey("Strong password should pass validation", func() {
			passwords := []string{
				"TestPass123!",
				"MySecure2023@",
				"StrongP@ss1",
				"Complex123$",
			}

			for _, password := range passwords {
				err := ValidatePasswordStrength(password, config)
				So(err, ShouldBeNil)
			}
		})

		Convey("Too short password should fail", func() {
			err := ValidatePasswordStrength("Ab1!", config)
			So(err, ShouldEqual, ErrPasswordTooShort)
		})

		Convey("Too long password should fail", func() {
			longPassword := make([]byte, MaxPasswordLength+1)
			for i := range longPassword {
				longPassword[i] = 'a'
			}
			err := ValidatePasswordStrength(string(longPassword), config)
			So(err, ShouldEqual, ErrPasswordTooLong)
		})

		Convey("Common password should fail", func() {
			err := ValidatePasswordStrength("password", config)
			So(err, ShouldEqual, ErrCommonPassword)
		})

		Convey("Insufficient character types should fail", func() {
			passwords := []string{
				"alllowercase", // 只有小写字母
				"ALLUPPERCASE", // 只有大写字母
				"123456789",    // 只有数字
				"testtest123",  // 只有小写字母和数字
			}

			for _, password := range passwords {
				err := ValidatePasswordStrength(password, config)
				So(err, ShouldEqual, ErrWeakPassword)
			}
		})

		Convey("Invalid characters should fail", func() {
			password := "Test123!\x00" // 包含null字符
			err := ValidatePasswordStrength(password, config)
			So(err, ShouldEqual, ErrPasswordInvalidChars)
		})
	})
}

func TestValidatePasswordForUser(t *testing.T) {
	Convey("Test ValidatePasswordForUser", t, func() {
		Convey("Password not containing user info should pass", func() {
			err := ValidatePasswordForUser("TestPass123!", "john", "john@example.com")
			So(err, ShouldBeNil)
		})

		Convey("Password containing username should fail", func() {
			err := ValidatePasswordForUser("johnabc123", "john", "john@example.com")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "username")
		})

		Convey("Password containing email local part should fail", func() {
			err := ValidatePasswordForUser("johnabc123", "user", "john@example.com")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "email")
		})

		Convey("Empty username and email should pass", func() {
			err := ValidatePasswordForUser("TestPass123!", "", "")
			So(err, ShouldBeNil)
		})
	})
}

func TestGetPasswordStrengthScore(t *testing.T) {
	Convey("Test GetPasswordStrengthScore", t, func() {
		Convey("Too short password should get 0 score", func() {
			score := GetPasswordStrengthScore("123")
			So(score, ShouldEqual, 0)
		})

		Convey("Strong password should get high score", func() {
			score := GetPasswordStrengthScore("TestPass123!")
			So(score, ShouldBeGreaterThan, 70)
		})

		Convey("Weak password should get low score", func() {
			score := GetPasswordStrengthScore("password")
			So(score, ShouldBeLessThan, 50)
		})

		Convey("Very long complex password should get high score", func() {
			score := GetPasswordStrengthScore("VeryLongAndComplexPassword123!@#")
			So(score, ShouldBeGreaterThan, 80)
		})

		Convey("Score should be in range 0-100", func() {
			testPasswords := []string{
				"a",
				"abc123",
				"TestPass123!",
				"VeryLongAndComplexPassword123!@#$%^&*()",
			}

			for _, password := range testPasswords {
				score := GetPasswordStrengthScore(password)
				So(score, ShouldBeGreaterThanOrEqualTo, 0)
				So(score, ShouldBeLessThanOrEqualTo, 100)
			}
		})
	})
}

func TestPasswordPatternValidation(t *testing.T) {
	Convey("Test Pattern Validation Functions", t, func() {
		Convey("isRepeatingPattern should detect repeating characters", func() {
			So(isRepeatingPattern("aaa"), ShouldBeTrue)
			So(isRepeatingPattern("111"), ShouldBeTrue)
			So(isRepeatingPattern("aaaa"), ShouldBeTrue)
			So(isRepeatingPattern("abc"), ShouldBeFalse)
			So(isRepeatingPattern("ab"), ShouldBeFalse)
		})

		Convey("isSequentialPattern should detect sequential characters", func() {
			So(isSequentialPattern("abc"), ShouldBeTrue)
			So(isSequentialPattern("123"), ShouldBeTrue)
			So(isSequentialPattern("cba"), ShouldBeTrue)
			So(isSequentialPattern("321"), ShouldBeTrue)
			So(isSequentialPattern("ace"), ShouldBeFalse)
			So(isSequentialPattern("ab"), ShouldBeFalse)
		})

		Convey("isCommonPattern should detect common patterns", func() {
			So(isCommonPattern("123456"), ShouldBeTrue)
			So(isCommonPattern("abcdef"), ShouldBeTrue)
			So(isCommonPattern("qwerty"), ShouldBeTrue)
			So(isCommonPattern("TestPass123!"), ShouldBeFalse)
		})

		Convey("isPatternComplex should validate complexity", func() {
			So(isPatternComplex("aaa"), ShouldBeFalse)         // 重复
			So(isPatternComplex("123"), ShouldBeFalse)         // 连续
			So(isPatternComplex("TestPass12A!"), ShouldBeTrue) // 复杂（没有连续的123）
		})
	})
}

func TestPasswordStrengthConfig(t *testing.T) {
	Convey("Test Custom Password Strength Config", t, func() {
		Convey("Config with all requirements should work", func() {
			config := PasswordStrengthConfig{
				MinLength:      10,
				MaxLength:      50,
				RequireUpper:   true,
				RequireLower:   true,
				RequireNumber:  true,
				RequireSpecial: true,
				MinTypes:       4,
			}

			// 满足所有要求的密码
			err := ValidatePasswordStrength("TestPass123!", config)
			So(err, ShouldBeNil)

			// 缺少大写字母
			err = ValidatePasswordStrength("testpass123!", config)
			So(err, ShouldEqual, ErrWeakPassword)

			// 长度不足
			err = ValidatePasswordStrength("Test123!", config)
			So(err, ShouldEqual, ErrPasswordTooShort)
		})

		Convey("Config with relaxed requirements should work", func() {
			config := PasswordStrengthConfig{
				MinLength:      6,
				MaxLength:      20,
				RequireUpper:   false,
				RequireLower:   false,
				RequireNumber:  false,
				RequireSpecial: false,
				MinTypes:       2,
			}

			// 只有字母和数字
			err := ValidatePasswordStrength("test123", config)
			So(err, ShouldBeNil)

			// 只有一种类型
			err = ValidatePasswordStrength("testtest", config)
			So(err, ShouldEqual, ErrWeakPassword)
		})
	})
}

func TestPasswordConstants(t *testing.T) {
	Convey("Test Password Constants", t, func() {
		Convey("BCrypt cost should be reasonable", func() {
			So(BCryptCost, ShouldEqual, 12)
			So(BCryptCost, ShouldBeGreaterThanOrEqualTo, 10)
			So(BCryptCost, ShouldBeLessThanOrEqualTo, 15)
		})

		Convey("Password length limits should be reasonable", func() {
			So(MinPasswordLength, ShouldEqual, 8)
			So(MaxPasswordLength, ShouldEqual, 128)
			So(MinPasswordLength, ShouldBeLessThan, MaxPasswordLength)
		})

		Convey("Default config should be reasonable", func() {
			config := DefaultPasswordConfig
			So(config.MinLength, ShouldEqual, MinPasswordLength)
			So(config.MaxLength, ShouldEqual, MaxPasswordLength)
			So(config.MinTypes, ShouldEqual, 3)
		})
	})
}

func TestPasswordUtilityFunctions(t *testing.T) {
	Convey("Test Utility Functions", t, func() {
		Convey("min function should work correctly", func() {
			So(min(5, 3), ShouldEqual, 3)
			So(min(3, 5), ShouldEqual, 3)
			So(min(5, 5), ShouldEqual, 5)
		})
	})
}

func TestPasswordErrorTypes(t *testing.T) {
	Convey("Test Password Error Types", t, func() {
		Convey("All error types should be defined", func() {
			So(ErrWeakPassword, ShouldNotBeNil)
			So(ErrPasswordTooShort, ShouldNotBeNil)
			So(ErrPasswordTooLong, ShouldNotBeNil)
			So(ErrPasswordInvalidChars, ShouldNotBeNil)
			So(ErrCommonPassword, ShouldNotBeNil)
		})

		Convey("Error messages should be meaningful", func() {
			So(ErrWeakPassword.Error(), ShouldContainSubstring, "strength")
			So(ErrPasswordTooShort.Error(), ShouldContainSubstring, "8")
			So(ErrPasswordTooLong.Error(), ShouldContainSubstring, "128")
			So(ErrCommonPassword.Error(), ShouldContainSubstring, "common")
		})
	})
}

func BenchmarkHashPassword(b *testing.B) {
	password := "TestPass123!"
	for i := 0; i < b.N; i++ {
		HashPassword(password)
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	password := "TestPass123!"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), BCryptCost)
	hashString := string(hash)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifyPassword(password, hashString)
	}
}

func BenchmarkGetPasswordStrengthScore(b *testing.B) {
	password := "TestPass123!"
	for i := 0; i < b.N; i++ {
		GetPasswordStrengthScore(password)
	}
}
