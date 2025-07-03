package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidator(t *testing.T) {
	Convey("Test Validator Basic Functionality", t, func() {
		validator := NewValidator()

		Convey("Should create validator correctly", func() {
			So(validator, ShouldNotBeNil)
			So(validator.HasErrors(), ShouldBeFalse)
		})

		Convey("Should handle empty validation correctly", func() {
			hasErrors := validator.HasErrors()
			So(hasErrors, ShouldBeFalse)
		})
	})
}

func TestValidatorRequired(t *testing.T) {
	Convey("Test Required Validation", t, func() {
		Convey("Non-empty string should pass", func() {
			validator := NewValidator()
			validator.Required("test", "value")
			So(validator.HasErrors(), ShouldBeFalse)
		})

		Convey("Empty string should fail", func() {
			validator := NewValidator()
			validator.Required("test", "")
			So(validator.HasErrors(), ShouldBeTrue)
			So(validator.GetFirstError(), ShouldContainSubstring, "test is required")
		})

		Convey("Whitespace-only string should fail", func() {
			validator := NewValidator()
			validator.Required("test", "   ")
			So(validator.HasErrors(), ShouldBeTrue)
			So(validator.GetFirstError(), ShouldContainSubstring, "test is required")
		})

		Convey("Nil value should fail", func() {
			validator := NewValidator()
			validator.Required("test", nil)
			So(validator.HasErrors(), ShouldBeTrue)
			So(validator.GetFirstError(), ShouldContainSubstring, "test is required")
		})
	})
}

func TestValidatorEmail(t *testing.T) {
	Convey("Test Email Validation", t, func() {
		validEmails := []string{
			"test@example.com",
			"user.name@example.com",
			"user+tag@example.com",
			"user@example-site.com",
			"test123@example.org",
		}

		invalidEmails := []string{
			"invalid",
			"@example.com",
			"user@",
			"user..name@example.com",
		}

		Convey("Valid emails should pass", func() {
			for _, email := range validEmails {
				validator := NewValidator()
				validator.Email("email", email)
				So(validator.HasErrors(), ShouldBeFalse)
			}
		})

		Convey("Invalid emails should fail", func() {
			for _, email := range invalidEmails {
				validator := NewValidator()
				validator.Email("email", email)
				So(validator.HasErrors(), ShouldBeTrue)
			}
		})

		Convey("Empty email should pass (not required)", func() {
			validator := NewValidator()
			validator.Email("email", "")
			So(validator.HasErrors(), ShouldBeFalse)
		})
	})
}

func TestValidatorUsername(t *testing.T) {
	Convey("Test Username Validation", t, func() {
		validUsernames := []string{
			"user123",
			"test_user",
			"abc123def",
			"user_123",
		}

		invalidUsernames := []string{
			"us",                                // too short
			"123456789012345678901234567890123", // too long
			"user@name",                         // invalid character
			"user name",                         // space
			"user.name",                         // dot
			"123user",                           // starts with number
			"_username",                         // starts with underscore
		}

		Convey("Valid usernames should pass", func() {
			for _, username := range validUsernames {
				validator := NewValidator()
				validator.Username("username", username)
				So(validator.HasErrors(), ShouldBeFalse)
			}
		})

		Convey("Invalid usernames should fail", func() {
			for _, username := range invalidUsernames {
				validator := NewValidator()
				validator.Username("username", username)
				So(validator.HasErrors(), ShouldBeTrue)
			}
		})

		Convey("Empty username should pass (not required)", func() {
			validator := NewValidator()
			validator.Username("username", "")
			So(validator.HasErrors(), ShouldBeFalse)
		})
	})
}

func TestValidatorLength(t *testing.T) {
	Convey("Test Length Validation", t, func() {
		Convey("String in valid range should pass", func() {
			validator := NewValidator()
			validator.Length("test", "hello", 3, 10)
			So(validator.HasErrors(), ShouldBeFalse)
		})

		Convey("String at min boundary should pass", func() {
			validator := NewValidator()
			validator.Length("test", "hello", 5, 10)
			So(validator.HasErrors(), ShouldBeFalse)
		})

		Convey("String at max boundary should pass", func() {
			validator := NewValidator()
			validator.Length("test", "hello", 3, 5)
			So(validator.HasErrors(), ShouldBeFalse)
		})

		Convey("String shorter than min should fail", func() {
			validator := NewValidator()
			validator.Length("test", "hi", 5, 10)
			So(validator.HasErrors(), ShouldBeTrue)
			So(validator.GetFirstError(), ShouldContainSubstring, "at least 5 characters")
		})

		Convey("String longer than max should fail", func() {
			validator := NewValidator()
			validator.Length("test", "hello world", 3, 5)
			So(validator.HasErrors(), ShouldBeTrue)
			So(validator.GetFirstError(), ShouldContainSubstring, "no more than 5 characters")
		})
	})
}

func TestValidatorRange(t *testing.T) {
	Convey("Test Range Validation", t, func() {
		Convey("Value in range should pass", func() {
			validator := NewValidator()
			validator.Range("test", 5, 1, 10)
			So(validator.HasErrors(), ShouldBeFalse)
		})

		Convey("Value at min boundary should pass", func() {
			validator := NewValidator()
			validator.Range("test", 1, 1, 10)
			So(validator.HasErrors(), ShouldBeFalse)
		})

		Convey("Value at max boundary should pass", func() {
			validator := NewValidator()
			validator.Range("test", 10, 1, 10)
			So(validator.HasErrors(), ShouldBeFalse)
		})

		Convey("Value below min should fail", func() {
			validator := NewValidator()
			validator.Range("test", 0, 1, 10)
			So(validator.HasErrors(), ShouldBeTrue)
			So(validator.GetFirstError(), ShouldContainSubstring, "at least 1")
		})

		Convey("Value above max should fail", func() {
			validator := NewValidator()
			validator.Range("test", 11, 1, 10)
			So(validator.HasErrors(), ShouldBeTrue)
			So(validator.GetFirstError(), ShouldContainSubstring, "no more than 10")
		})
	})
}

func TestValidatorIn(t *testing.T) {
	Convey("Test In Validation", t, func() {
		allowed := []string{"admin", "user", "guest"}

		Convey("Value in allowed list should pass", func() {
			validator := NewValidator()
			validator.In("role", "admin", allowed)
			So(validator.HasErrors(), ShouldBeFalse)
		})

		Convey("Value not in allowed list should fail", func() {
			validator := NewValidator()
			validator.In("role", "invalid", allowed)
			So(validator.HasErrors(), ShouldBeTrue)
			So(validator.GetFirstError(), ShouldContainSubstring, "must be one of")
		})

		Convey("Empty value should pass (not required)", func() {
			validator := NewValidator()
			validator.In("role", "", allowed)
			So(validator.HasErrors(), ShouldBeFalse)
		})
	})
}

func TestValidatorCustom(t *testing.T) {
	Convey("Test Custom Validation", t, func() {
		Convey("Custom rule returning true should pass", func() {
			validator := NewValidator()
			validator.Custom("test", "value", func(v interface{}) bool { return true }, "Custom error")
			So(validator.HasErrors(), ShouldBeFalse)
		})

		Convey("Custom rule returning false should fail", func() {
			validator := NewValidator()
			validator.Custom("test", "value", func(v interface{}) bool { return false }, "Custom error")
			So(validator.HasErrors(), ShouldBeTrue)
			So(validator.GetFirstError(), ShouldContainSubstring, "Custom error")
		})
	})
}

func TestValidatorSlug(t *testing.T) {
	Convey("Test Slug Validation", t, func() {
		validSlugs := []string{
			"my-blog-post",
			"post-123",
			"hello-world",
			"a",
			"123",
		}

		invalidSlugs := []string{
			"My-Blog-Post",  // uppercase
			"my blog post",  // spaces
			"my_blog_post",  // underscores
			"my-blog-post!", // special characters
			"-my-blog",      // starts with hyphen
			"my-blog-",      // ends with hyphen
		}

		Convey("Valid slugs should pass", func() {
			for _, slug := range validSlugs {
				validator := NewValidator()
				validator.Slug("slug", slug)
				So(validator.HasErrors(), ShouldBeFalse)
			}
		})

		Convey("Invalid slugs should fail", func() {
			for _, slug := range invalidSlugs {
				validator := NewValidator()
				validator.Slug("slug", slug)
				So(validator.HasErrors(), ShouldBeTrue)
			}
		})

		Convey("Empty slug should pass (not required)", func() {
			validator := NewValidator()
			validator.Slug("slug", "")
			So(validator.HasErrors(), ShouldBeFalse)
		})
	})
}

func TestValidatorChaining(t *testing.T) {
	Convey("Test Validator Chaining", t, func() {
		Convey("Multiple valid rules should pass", func() {
			validator := NewValidator()
			validator.Required("email", "test@example.com").
				Email("email", "test@example.com").
				Length("email", "test@example.com", 5, 50)

			So(validator.HasErrors(), ShouldBeFalse)
		})

		Convey("First invalid rule should fail", func() {
			validator := NewValidator()
			validator.Required("email", "").
				Email("email", "").
				Length("email", "", 5, 50)

			So(validator.HasErrors(), ShouldBeTrue)
			So(validator.GetFirstError(), ShouldContainSubstring, "email is required")
		})

		Convey("Multiple invalid rules should return first error", func() {
			validator := NewValidator()
			validator.Required("email", "invalid").
				Email("email", "invalid")

			So(validator.HasErrors(), ShouldBeTrue)
			// Should return first error (email format)
			So(validator.GetFirstError(), ShouldContainSubstring, "valid email address")
		})
	})
}

func TestValidatorUtilities(t *testing.T) {
	Convey("Test Validator Utility Functions", t, func() {
		Convey("SanitizeString should clean input", func() {
			result := SanitizeString("  <script>alert('xss')</script>  ")
			So(result, ShouldEqual, "&lt;script&gt;alert(&#x27;xss&#x27;)&lt;/script&gt;")
		})

		Convey("ValidatePageParams should work with valid params", func() {
			page, size, err := ValidatePageParams(1, 10)
			So(err, ShouldBeNil)
			So(page, ShouldEqual, 1)
			So(size, ShouldEqual, 10)
		})

		Convey("ValidatePageParams should handle invalid page", func() {
			page, size, err := ValidatePageParams(0, 10)
			So(err, ShouldBeNil)
			So(page, ShouldEqual, 1) // default
			So(size, ShouldEqual, 10)
		})

		Convey("ValidatePageParams should handle invalid size", func() {
			page, size, err := ValidatePageParams(1, 0)
			So(err, ShouldBeNil)
			So(page, ShouldEqual, 1)
			So(size, ShouldEqual, 10) // default
		})

		Convey("ValidatePageParams should limit large size", func() {
			page, size, err := ValidatePageParams(1, 1000)
			So(err, ShouldBeNil)
			So(page, ShouldEqual, 1)
			So(size, ShouldEqual, 100) // max
		})

		Convey("ParseAndValidateID should work with valid ObjectID", func() {
			err := ParseAndValidateID("507f1f77bcf86cd799439011")
			So(err, ShouldBeNil)
		})

		Convey("ParseAndValidateID should work with valid UUID", func() {
			err := ParseAndValidateID("550e8400-e29b-41d4-a716-446655440000")
			So(err, ShouldBeNil)
		})

		Convey("ParseAndValidateID should fail with invalid ID", func() {
			err := ParseAndValidateID("invalid")
			So(err, ShouldNotBeNil)
		})

		Convey("ParseAndValidateID should fail with empty ID", func() {
			err := ParseAndValidateID("")
			So(err, ShouldNotBeNil)
		})

		Convey("ValidateSlug should work with valid slug", func() {
			err := ValidateSlug("my-blog-post")
			So(err, ShouldBeNil)
		})

		Convey("ValidateSlug should work with numbers", func() {
			err := ValidateSlug("post-123")
			So(err, ShouldBeNil)
		})

		Convey("ValidateSlug should fail with invalid characters", func() {
			err := ValidateSlug("my blog post")
			So(err, ShouldNotBeNil)
		})

		Convey("ValidateSlug should fail with uppercase", func() {
			err := ValidateSlug("My-Blog-Post")
			So(err, ShouldNotBeNil)
		})

		Convey("ValidateSlug should fail with special characters", func() {
			err := ValidateSlug("my-blog-post!")
			So(err, ShouldNotBeNil)
		})
	})
}

func TestStandaloneValidators(t *testing.T) {
	Convey("Test Standalone Validation Functions", t, func() {
		Convey("ValidateRequired should work correctly", func() {
			err := ValidateRequired("field", "value")
			So(err, ShouldBeNil)

			err = ValidateRequired("field", "")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "field is required")
		})

		Convey("ValidateEmail should work correctly", func() {
			err := ValidateEmail("test@example.com")
			So(err, ShouldBeNil)

			err = ValidateEmail("invalid")
			So(err, ShouldNotBeNil)
		})

		Convey("ValidateUsername should work correctly", func() {
			err := ValidateUsername("testuser123")
			So(err, ShouldBeNil)

			err = ValidateUsername("123user")
			So(err, ShouldNotBeNil)
		})

		Convey("ValidateStringLength should work correctly", func() {
			err := ValidateStringLength("hello", 3, 10)
			So(err, ShouldBeNil)

			err = ValidateStringLength("hi", 5, 10)
			So(err, ShouldNotBeNil)
		})

		Convey("ValidateIntRange should work correctly", func() {
			err := ValidateIntRange(5, 1, 10)
			So(err, ShouldBeNil)

			err = ValidateIntRange(15, 1, 10)
			So(err, ShouldNotBeNil)
		})

		Convey("ValidateEnum should work correctly", func() {
			allowed := []string{"admin", "user", "guest"}

			err := ValidateEnum("admin", allowed)
			So(err, ShouldBeNil)

			err = ValidateEnum("invalid", allowed)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestValidatorEdgeCases(t *testing.T) {
	Convey("Test Validator Edge Cases", t, func() {
		Convey("Unicode characters should be handled correctly", func() {
			validator := NewValidator()
			validator.Required("test", "测试")
			validator.Length("test", "测试", 2, 10)
			So(validator.HasErrors(), ShouldBeFalse)
		})

		Convey("Very long strings should be handled", func() {
			longString := string(make([]byte, 10000))
			for i := range longString {
				longString = longString[:i] + "a" + longString[i+1:]
			}

			validator := NewValidator()
			validator.Length("test", longString, 0, 5000)
			So(validator.HasErrors(), ShouldBeTrue)
		})

		Convey("Empty validator should not have errors", func() {
			validator := NewValidator()
			So(validator.HasErrors(), ShouldBeFalse)
		})

		Convey("Multiple errors should return first", func() {
			validator := NewValidator()
			validator.Required("field1", "")
			validator.Required("field2", "")
			So(validator.HasErrors(), ShouldBeTrue)
			// Should contain first error
			So(validator.GetFirstError(), ShouldContainSubstring, "field1 is required")
		})
	})
}

func TestValidatorURL(t *testing.T) {
	Convey("Test URL Validation", t, func() {
		validURLs := []string{
			"http://example.com",
			"https://example.com",
			"https://www.example.com/path",
			"http://localhost:8080",
		}

		invalidURLs := []string{
			"not-a-url",
			"ftp://example.com",
			"example.com",
			"//example.com",
		}

		Convey("Valid URLs should pass", func() {
			for _, url := range validURLs {
				validator := NewValidator()
				validator.URL("url", url)
				So(validator.HasErrors(), ShouldBeFalse)
			}
		})

		Convey("Invalid URLs should fail", func() {
			for _, url := range invalidURLs {
				validator := NewValidator()
				validator.URL("url", url)
				So(validator.HasErrors(), ShouldBeTrue)
			}
		})
	})
}

func TestValidatorPhone(t *testing.T) {
	Convey("Test Phone Validation", t, func() {
		validPhones := []string{
			"13888888888",
			"15999999999",
			"18000000000",
		}

		invalidPhones := []string{
			"12888888888",  // starts with 12
			"1388888888",   // too short
			"138888888888", // too long
			"23888888888",  // starts with 2
			"phone",        // not numeric
		}

		Convey("Valid phones should pass", func() {
			for _, phone := range validPhones {
				validator := NewValidator()
				validator.Phone("phone", phone)
				So(validator.HasErrors(), ShouldBeFalse)
			}
		})

		Convey("Invalid phones should fail", func() {
			for _, phone := range invalidPhones {
				validator := NewValidator()
				validator.Phone("phone", phone)
				So(validator.HasErrors(), ShouldBeTrue)
			}
		})
	})
}

func BenchmarkValidator(b *testing.B) {
	for i := 0; i < b.N; i++ {
		validator := NewValidator()
		validator.Required("email", "test@example.com").
			Email("email", "test@example.com").
			Length("email", "test@example.com", 5, 50)
		validator.HasErrors()
	}
}

func BenchmarkEmailValidation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		validator := NewValidator()
		validator.Email("email", "test@example.com")
		validator.HasErrors()
	}
}

func BenchmarkUsernameValidation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		validator := NewValidator()
		validator.Username("username", "testuser123")
		validator.HasErrors()
	}
}
