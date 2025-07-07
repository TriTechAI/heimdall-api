package e2e

import (
	"os"
	"strconv"
)

// TestConfig 测试配置
type TestConfig struct {
	AdminAPIURL  string
	PublicAPIURL string
	TestUser     TestUser
	Database     DatabaseConfig
}

// TestUser 测试用户配置
type TestUser struct {
	Username string
	Password string
	Email    string
	Role     string
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	MongoURL string
	RedisURL string
	TestDB   string
}

// GetTestConfig 获取测试配置
func GetTestConfig() *TestConfig {
	return &TestConfig{
		AdminAPIURL:  getEnv("TEST_ADMIN_API_URL", "http://localhost:8080"),
		PublicAPIURL: getEnv("TEST_PUBLIC_API_URL", "http://localhost:8081"),
		TestUser: TestUser{
			Username: getEnv("TEST_USER_USERNAME", "testuser"),
			Password: getEnv("TEST_USER_PASSWORD", "testpass123"),
			Email:    getEnv("TEST_USER_EMAIL", "test@example.com"),
			Role:     getEnv("TEST_USER_ROLE", "admin"),
		},
		Database: DatabaseConfig{
			MongoURL: getEnv("TEST_MONGO_URL", "mongodb://localhost:27017"),
			RedisURL: getEnv("TEST_REDIS_URL", "redis://localhost:6379"),
			TestDB:   getEnv("TEST_DATABASE", "heimdall_test"),
		},
	}
}

// getEnv 获取环境变量，如果不存在则使用默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取整数型环境变量
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool 获取布尔型环境变量
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}