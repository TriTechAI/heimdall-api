package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// HealthStatus 健康状态枚举
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusDegraded  HealthStatus = "degraded"
)

// ComponentHealth 组件健康信息
type ComponentHealth struct {
	Status  HealthStatus           `json:"status"`
	Message string                 `json:"message,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
	Latency string                 `json:"latency,omitempty"`
}

// OverallHealth 总体健康状态
type OverallHealth struct {
	Status     HealthStatus               `json:"status"`
	Timestamp  time.Time                  `json:"timestamp"`
	Version    string                     `json:"version,omitempty"`
	Components map[string]ComponentHealth `json:"components"`
	Summary    string                     `json:"summary,omitempty"`
}

// HealthChecker 健康检查器
type HealthChecker struct {
	mongoClient *MongoClient
	redisClient *RedisClient
	version     string
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(mongoClient *MongoClient, redisClient *RedisClient, version string) *HealthChecker {
	return &HealthChecker{
		mongoClient: mongoClient,
		redisClient: redisClient,
		version:     version,
	}
}

// CheckHealth 执行健康检查
func (hc *HealthChecker) CheckHealth(ctx context.Context) *OverallHealth {
	health := &OverallHealth{
		Timestamp:  time.Now(),
		Version:    hc.version,
		Components: make(map[string]ComponentHealth),
	}

	// 检查各组件健康状态
	var healthyCount, totalCount int

	// 检查MongoDB
	if hc.mongoClient != nil {
		mongoHealth := hc.checkMongoHealth(ctx)
		health.Components["mongodb"] = mongoHealth
		totalCount++
		if mongoHealth.Status == StatusHealthy {
			healthyCount++
		}
	}

	// 检查Redis
	if hc.redisClient != nil {
		redisHealth := hc.checkRedisHealth(ctx)
		health.Components["redis"] = redisHealth
		totalCount++
		if redisHealth.Status == StatusHealthy {
			healthyCount++
		}
	}

	// 确定总体状态
	health.Status, health.Summary = hc.determineOverallStatus(healthyCount, totalCount)

	return health
}

// checkMongoHealth 检查MongoDB健康状态
func (hc *HealthChecker) checkMongoHealth(ctx context.Context) ComponentHealth {
	start := time.Now()

	if !hc.mongoClient.IsHealthy(ctx) {
		return ComponentHealth{
			Status:  StatusUnhealthy,
			Message: "MongoDB connection failed",
			Latency: time.Since(start).String(),
		}
	}

	// 获取统计信息
	stats, err := hc.mongoClient.GetStats(ctx)
	if err != nil {
		return ComponentHealth{
			Status:  StatusDegraded,
			Message: fmt.Sprintf("MongoDB stats unavailable: %v", err),
			Latency: time.Since(start).String(),
		}
	}

	return ComponentHealth{
		Status:  StatusHealthy,
		Message: "MongoDB is operational",
		Details: stats,
		Latency: time.Since(start).String(),
	}
}

// checkRedisHealth 检查Redis健康状态
func (hc *HealthChecker) checkRedisHealth(ctx context.Context) ComponentHealth {
	start := time.Now()

	if !hc.redisClient.IsHealthy(ctx) {
		return ComponentHealth{
			Status:  StatusUnhealthy,
			Message: "Redis connection failed",
			Latency: time.Since(start).String(),
		}
	}

	// 获取统计信息
	stats, err := hc.redisClient.GetStats(ctx)
	if err != nil {
		return ComponentHealth{
			Status:  StatusDegraded,
			Message: fmt.Sprintf("Redis stats unavailable: %v", err),
			Latency: time.Since(start).String(),
		}
	}

	return ComponentHealth{
		Status:  StatusHealthy,
		Message: "Redis is operational",
		Details: stats,
		Latency: time.Since(start).String(),
	}
}

// determineOverallStatus 确定总体健康状态
func (hc *HealthChecker) determineOverallStatus(healthyCount, totalCount int) (HealthStatus, string) {
	if totalCount == 0 {
		return StatusHealthy, "No components to check"
	}

	if healthyCount == totalCount {
		return StatusHealthy, "All components are healthy"
	}

	if healthyCount > 0 {
		return StatusDegraded, fmt.Sprintf("%d of %d components are healthy", healthyCount, totalCount)
	}

	return StatusUnhealthy, "All components are unhealthy"
}

// CheckReadiness 检查服务准备状态(用于Kubernetes就绪探针)
func (hc *HealthChecker) CheckReadiness(ctx context.Context) bool {
	health := hc.CheckHealth(ctx)
	return health.Status == StatusHealthy || health.Status == StatusDegraded
}

// CheckLiveness 检查服务存活状态(用于Kubernetes存活探针)
func (hc *HealthChecker) CheckLiveness(ctx context.Context) bool {
	// 存活检查通常比就绪检查更宽松
	// 只要有任何一个关键组件正常即可
	if hc.mongoClient != nil {
		mongoCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		if hc.mongoClient.IsHealthy(mongoCtx) {
			return true
		}
	}

	if hc.redisClient != nil {
		redisCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		if hc.redisClient.IsHealthy(redisCtx) {
			return true
		}
	}

	return false
}

// ToJSON 将健康状态转换为JSON字符串
func (oh *OverallHealth) ToJSON() (string, error) {
	data, err := json.MarshalIndent(oh, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToHTTPStatus 将健康状态转换为HTTP状态码
func (oh *OverallHealth) ToHTTPStatus() int {
	switch oh.Status {
	case StatusHealthy:
		return 200
	case StatusDegraded:
		return 200 // 降级状态仍然返回200，但在响应中标明状态
	case StatusUnhealthy:
		return 503
	default:
		return 500
	}
}
