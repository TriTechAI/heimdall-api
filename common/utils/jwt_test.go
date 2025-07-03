package utils

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestJWTManager(t *testing.T) {
	secretKey := "test-secret-key"
	issuer := "test-issuer"
	jwtManager := NewJWTManager(secretKey, issuer)

	Convey("Test JWT Manager Creation", t, func() {
		Convey("Should create JWT manager successfully", func() {
			So(jwtManager, ShouldNotBeNil)
			So(jwtManager.secretKey, ShouldResemble, []byte(secretKey))
			So(jwtManager.issuer, ShouldEqual, issuer)
		})
	})
}

func TestGenerateToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret", "test-issuer")

	Convey("Test GenerateToken", t, func() {
		userID := "user123"
		username := "testuser"
		role := "admin"

		Convey("Valid parameters should generate token successfully", func() {
			tokenPair, err := jwtManager.GenerateToken(userID, username, role)

			So(err, ShouldBeNil)
			So(tokenPair, ShouldNotBeNil)
			So(tokenPair.AccessToken, ShouldNotBeEmpty)
			So(tokenPair.RefreshToken, ShouldNotBeEmpty)
			So(tokenPair.TokenType, ShouldEqual, "Bearer")
			So(tokenPair.ExpiresAt, ShouldHappenAfter, time.Now())
		})

		Convey("Empty userID should return error", func() {
			tokenPair, err := jwtManager.GenerateToken("", username, role)

			So(err, ShouldNotBeNil)
			So(tokenPair, ShouldBeNil)
		})

		Convey("Empty username should return error", func() {
			tokenPair, err := jwtManager.GenerateToken(userID, "", role)

			So(err, ShouldNotBeNil)
			So(tokenPair, ShouldBeNil)
		})

		Convey("Empty role should return error", func() {
			tokenPair, err := jwtManager.GenerateToken(userID, username, "")

			So(err, ShouldNotBeNil)
			So(tokenPair, ShouldBeNil)
		})
	})
}

func TestValidateToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret", "test-issuer")

	Convey("Test ValidateToken", t, func() {
		userID := "user123"
		username := "testuser"
		role := "admin"

		Convey("Valid token should be validated successfully", func() {
			tokenPair, _ := jwtManager.GenerateToken(userID, username, role)
			claims, err := jwtManager.ValidateToken(tokenPair.AccessToken)

			So(err, ShouldBeNil)
			So(claims, ShouldNotBeNil)
			So(claims.UserID, ShouldEqual, userID)
			So(claims.Username, ShouldEqual, username)
			So(claims.Role, ShouldEqual, role)
		})

		Convey("Empty token should return error", func() {
			claims, err := jwtManager.ValidateToken("")

			So(err, ShouldEqual, ErrInvalidToken)
			So(claims, ShouldBeNil)
		})

		Convey("Invalid token should return error", func() {
			claims, err := jwtManager.ValidateToken("invalid.token.here")

			So(err, ShouldNotBeNil)
			So(claims, ShouldBeNil)
		})

		Convey("Token with wrong secret should return error", func() {
			wrongManager := NewJWTManager("wrong-secret", "test-issuer")
			tokenPair, _ := jwtManager.GenerateToken(userID, username, role)
			claims, err := wrongManager.ValidateToken(tokenPair.AccessToken)

			So(err, ShouldNotBeNil)
			So(claims, ShouldBeNil)
		})

		Convey("Malformed token should return specific error", func() {
			claims, err := jwtManager.ValidateToken("malformed")

			So(err, ShouldEqual, ErrMalformedToken)
			So(claims, ShouldBeNil)
		})
	})
}

func TestRefreshToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret", "test-issuer")

	Convey("Test RefreshToken", t, func() {
		userID := "user123"
		username := "testuser"
		role := "admin"

		Convey("Valid refresh token should generate new token pair", func() {
			originalPair, _ := jwtManager.GenerateToken(userID, username, role)
			newPair, err := jwtManager.RefreshToken(originalPair.RefreshToken)

			So(err, ShouldBeNil)
			So(newPair, ShouldNotBeNil)
			So(newPair.AccessToken, ShouldNotBeEmpty)
			So(newPair.AccessToken, ShouldNotEqual, originalPair.AccessToken)
		})

		Convey("Invalid refresh token should return error", func() {
			newPair, err := jwtManager.RefreshToken("invalid-token")

			So(err, ShouldNotBeNil)
			So(newPair, ShouldBeNil)
		})
	})
}

func TestTokenExtractionMethods(t *testing.T) {
	jwtManager := NewJWTManager("test-secret", "test-issuer")
	userID := "user123"
	username := "testuser"
	role := "admin"

	Convey("Test Token Extraction Methods", t, func() {
		tokenPair, _ := jwtManager.GenerateToken(userID, username, role)
		token := tokenPair.AccessToken

		Convey("ExtractUserIDFromToken should work correctly", func() {
			extractedUserID, err := jwtManager.ExtractUserIDFromToken(token)

			So(err, ShouldBeNil)
			So(extractedUserID, ShouldEqual, userID)
		})

		Convey("ExtractUsernameFromToken should work correctly", func() {
			extractedUsername, err := jwtManager.ExtractUsernameFromToken(token)

			So(err, ShouldBeNil)
			So(extractedUsername, ShouldEqual, username)
		})

		Convey("ExtractRoleFromToken should work correctly", func() {
			extractedRole, err := jwtManager.ExtractRoleFromToken(token)

			So(err, ShouldBeNil)
			So(extractedRole, ShouldEqual, role)
		})

		Convey("ExtractTokenIDFromToken should work correctly", func() {
			tokenID, err := jwtManager.ExtractTokenIDFromToken(token)

			So(err, ShouldBeNil)
			So(tokenID, ShouldNotBeEmpty)
		})
	})
}

func TestTokenTimeOperations(t *testing.T) {
	jwtManager := NewJWTManager("test-secret", "test-issuer")

	Convey("Test Token Time Operations", t, func() {
		tokenPair, _ := jwtManager.GenerateToken("user123", "testuser", "admin")
		token := tokenPair.AccessToken

		Convey("GetTokenExpirationTime should return correct time", func() {
			expirationTime, err := jwtManager.GetTokenExpirationTime(token)

			So(err, ShouldBeNil)
			So(expirationTime, ShouldHappenAfter, time.Now())
			So(expirationTime, ShouldHappenBefore, time.Now().Add(AccessTokenExpiration+time.Minute))
		})

		Convey("IsTokenExpired should return false for valid token", func() {
			isExpired := jwtManager.IsTokenExpired(token)
			So(isExpired, ShouldBeFalse)
		})

		Convey("GetTokenRemainingTime should return positive duration", func() {
			remaining, err := jwtManager.GetTokenRemainingTime(token)

			So(err, ShouldBeNil)
			So(remaining, ShouldBeGreaterThan, 0)
			So(remaining, ShouldBeLessThanOrEqualTo, AccessTokenExpiration)
		})

		Convey("GetTokenAge should return reasonable age", func() {
			age, err := jwtManager.GetTokenAge(token)

			So(err, ShouldBeNil)
			So(age, ShouldBeGreaterThan, 0)
			So(age, ShouldBeLessThan, time.Minute) // 刚生成的token
		})

		Convey("IsTokenRecentlyIssued should work correctly", func() {
			isRecent, err := jwtManager.IsTokenRecentlyIssued(token, time.Minute)

			So(err, ShouldBeNil)
			So(isRecent, ShouldBeTrue)

			isRecent, err = jwtManager.IsTokenRecentlyIssued(token, time.Nanosecond)
			So(err, ShouldBeNil)
			So(isRecent, ShouldBeFalse)
		})
	})
}

func TestParseTokenWithoutValidation(t *testing.T) {
	jwtManager := NewJWTManager("test-secret", "test-issuer")

	Convey("Test ParseTokenWithoutValidation", t, func() {
		tokenPair, _ := jwtManager.GenerateToken("user123", "testuser", "admin")
		token := tokenPair.AccessToken

		Convey("Valid token should be parsed without validation", func() {
			claims, err := jwtManager.ParseTokenWithoutValidation(token)

			So(err, ShouldBeNil)
			So(claims, ShouldNotBeNil)
			So(claims.UserID, ShouldEqual, "user123")
		})

		Convey("Invalid token should return error", func() {
			claims, err := jwtManager.ParseTokenWithoutValidation("invalid-token")

			So(err, ShouldNotBeNil)
			So(claims, ShouldBeNil)
		})
	})
}

func TestExtractTokenMetadata(t *testing.T) {
	jwtManager := NewJWTManager("test-secret", "test-issuer")

	Convey("Test ExtractTokenMetadata", t, func() {
		tokenPair, _ := jwtManager.GenerateToken("user123", "testuser", "admin")
		token := tokenPair.AccessToken

		Convey("Should extract all metadata correctly", func() {
			metadata, err := jwtManager.ExtractTokenMetadata(token)

			So(err, ShouldBeNil)
			So(metadata, ShouldNotBeNil)
			So(metadata["userID"], ShouldEqual, "user123")
			So(metadata["username"], ShouldEqual, "testuser")
			So(metadata["role"], ShouldEqual, "admin")
			So(metadata["tokenID"], ShouldNotBeEmpty)
			So(metadata["issuer"], ShouldEqual, "test-issuer")
			So(metadata["issuedAt"], ShouldNotBeNil)
			So(metadata["expiresAt"], ShouldNotBeNil)
			So(metadata["notBefore"], ShouldNotBeNil)
		})
	})
}

func TestUtilityFunctions(t *testing.T) {
	Convey("Test Utility Functions", t, func() {
		Convey("GenerateSessionKey should work correctly", func() {
			key := GenerateSessionKey("user123", "token456")
			So(key, ShouldEqual, "session:user123:token456")
		})

		Convey("GenerateBlacklistKey should work correctly", func() {
			key := GenerateBlacklistKey("token456")
			So(key, ShouldEqual, "blacklist:token456")
		})

		Convey("ParseAuthHeader should work correctly", func() {
			token, err := ParseAuthHeader("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")
			So(err, ShouldBeNil)
			So(token, ShouldEqual, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")

			_, err = ParseAuthHeader("")
			So(err, ShouldNotBeNil)

			_, err = ParseAuthHeader("Invalid header")
			So(err, ShouldNotBeNil)

			_, err = ParseAuthHeader("Bearer ")
			So(err, ShouldNotBeNil)
		})

		Convey("ValidateTokenFormat should work correctly", func() {
			err := ValidateTokenFormat("part1.part2.part3")
			So(err, ShouldBeNil)

			err = ValidateTokenFormat("")
			So(err, ShouldNotBeNil)

			err = ValidateTokenFormat("invalid")
			So(err, ShouldEqual, ErrMalformedToken)

			err = ValidateTokenFormat("part1..part3")
			So(err, ShouldNotBeNil)
		})
	})
}

func TestCreateCustomClaims(t *testing.T) {
	Convey("Test CreateCustomClaims", t, func() {
		customData := map[string]interface{}{
			"department": "engineering",
			"level":      5,
		}

		Convey("Should create custom claims correctly", func() {
			claims := CreateCustomClaims("user123", "testuser", "admin", customData)

			So(claims["sub"], ShouldEqual, "user123")
			So(claims["username"], ShouldEqual, "testuser")
			So(claims["role"], ShouldEqual, "admin")
			So(claims["department"], ShouldEqual, "engineering")
			So(claims["level"], ShouldEqual, 5)
			So(claims["jti"], ShouldNotBeEmpty)
			So(claims["iat"], ShouldNotBeNil)
			So(claims["exp"], ShouldNotBeNil)
			So(claims["nbf"], ShouldNotBeNil)
		})
	})
}

func TestTokenConstants(t *testing.T) {
	Convey("Test Token Constants", t, func() {
		Convey("Token expiration constants should be reasonable", func() {
			So(AccessTokenExpiration, ShouldEqual, 2*time.Hour)
			So(RefreshTokenExpiration, ShouldEqual, 7*24*time.Hour)
		})

		Convey("Error constants should be defined", func() {
			So(ErrInvalidToken, ShouldNotBeNil)
			So(ErrTokenExpired, ShouldNotBeNil)
			So(ErrTokenNotYetValid, ShouldNotBeNil)
			So(ErrMalformedToken, ShouldNotBeNil)
			So(ErrUnknownClaims, ShouldNotBeNil)
		})
	})
}

func TestJWTClaims(t *testing.T) {
	Convey("Test JWT Claims Structure", t, func() {
		claims := &JWTClaims{
			UserID:   "user123",
			Username: "testuser",
			Role:     "admin",
			TokenID:  "token456",
		}

		Convey("Claims structure should be correct", func() {
			So(claims.UserID, ShouldEqual, "user123")
			So(claims.Username, ShouldEqual, "testuser")
			So(claims.Role, ShouldEqual, "admin")
			So(claims.TokenID, ShouldEqual, "token456")
		})
	})
}

func TestTokenPair(t *testing.T) {
	Convey("Test Token Pair Structure", t, func() {
		tokenPair := &TokenPair{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiresAt:    time.Now().Add(AccessTokenExpiration),
			TokenType:    "Bearer",
		}

		Convey("Token pair structure should be correct", func() {
			So(tokenPair.AccessToken, ShouldEqual, "access-token")
			So(tokenPair.RefreshToken, ShouldEqual, "refresh-token")
			So(tokenPair.TokenType, ShouldEqual, "Bearer")
			So(tokenPair.ExpiresAt, ShouldHappenAfter, time.Now())
		})
	})
}

func TestEdgeCases(t *testing.T) {
	jwtManager := NewJWTManager("test-secret", "test-issuer")

	Convey("Test Edge Cases", t, func() {
		Convey("Very long values should work", func() {
			longUserID := string(make([]byte, 1000))
			for i := range longUserID {
				longUserID = longUserID[:i] + "a" + longUserID[i+1:]
			}

			tokenPair, err := jwtManager.GenerateToken(longUserID, "user", "role")
			So(err, ShouldBeNil)
			So(tokenPair, ShouldNotBeNil)
		})

		Convey("Special characters in values should work", func() {
			tokenPair, err := jwtManager.GenerateToken("user@123", "user-name_123", "admin-role")
			So(err, ShouldBeNil)
			So(tokenPair, ShouldNotBeNil)

			claims, err := jwtManager.ValidateToken(tokenPair.AccessToken)
			So(err, ShouldBeNil)
			So(claims.UserID, ShouldEqual, "user@123")
			So(claims.Username, ShouldEqual, "user-name_123")
			So(claims.Role, ShouldEqual, "admin-role")
		})

		Convey("Invalid signing method should be rejected", func() {
			// 测试ValidateToken会拒绝错误的签名方法
			// 使用一个伪造的RS256令牌字符串
			tokenString := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMTIzIn0.invalid"

			claims, err := jwtManager.ValidateToken(tokenString)
			So(err, ShouldNotBeNil)
			So(claims, ShouldBeNil)
		})
	})
}

func BenchmarkGenerateToken(b *testing.B) {
	jwtManager := NewJWTManager("test-secret", "test-issuer")
	for i := 0; i < b.N; i++ {
		jwtManager.GenerateToken("user123", "testuser", "admin")
	}
}

func BenchmarkValidateToken(b *testing.B) {
	jwtManager := NewJWTManager("test-secret", "test-issuer")
	tokenPair, _ := jwtManager.GenerateToken("user123", "testuser", "admin")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jwtManager.ValidateToken(tokenPair.AccessToken)
	}
}
