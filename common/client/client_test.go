package client

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMongoClient(t *testing.T) {
	Convey("Test MongoDB Client", t, func() {
		config := MongoConfig{
			Host:     "localhost",
			Port:     27017,
			Database: "heimdall_test",
			Timeout:  5,
		}

		Convey("Should create MongoDB client successfully", func() {
			client, err := NewMongoClient(config)
			So(err, ShouldBeNil)
			So(client, ShouldNotBeNil)

			// 测试连接
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err = client.Ping(ctx)
			So(err, ShouldBeNil)

			// 测试健康检查
			isHealthy := client.IsHealthy(ctx)
			So(isHealthy, ShouldBeTrue)

			// 获取统计信息
			stats, err := client.GetStats(ctx)
			So(err, ShouldBeNil)
			So(stats, ShouldNotBeNil)

			// 清理
			client.Close(ctx)
		})
	})
}

func TestRedisClient(t *testing.T) {
	Convey("Test Redis Client", t, func() {
		config := RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Database: 1, // 使用测试数据库
			Timeout:  5,
		}

		Convey("Should create Redis client successfully", func() {
			client, err := NewRedisClient(config)
			So(err, ShouldBeNil)
			So(client, ShouldNotBeNil)

			// 测试连接
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err = client.Ping(ctx)
			So(err, ShouldBeNil)

			// 测试健康检查
			isHealthy := client.IsHealthy(ctx)
			So(isHealthy, ShouldBeTrue)

			// 测试基本操作
			err = client.Set(ctx, "test_key", "test_value", time.Minute)
			So(err, ShouldBeNil)

			value, err := client.Get(ctx, "test_key")
			So(err, ShouldBeNil)
			So(value, ShouldEqual, "test_value")

			// 获取统计信息
			stats, err := client.GetStats(ctx)
			So(err, ShouldBeNil)
			So(stats, ShouldNotBeNil)

			// 清理测试数据
			client.Del(ctx, "test_key")

			// 关闭连接
			client.Close()
		})
	})
}

func TestHealthChecker(t *testing.T) {
	Convey("Test Health Checker", t, func() {
		// 创建MongoDB客户端
		mongoConfig := MongoConfig{
			Host:     "localhost",
			Port:     27017,
			Database: "heimdall_test",
			Timeout:  5,
		}
		mongoClient, err := NewMongoClient(mongoConfig)
		So(err, ShouldBeNil)

		// 创建Redis客户端
		redisConfig := RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Database: 1,
			Timeout:  5,
		}
		redisClient, err := NewRedisClient(redisConfig)
		So(err, ShouldBeNil)

		// 创建健康检查器
		healthChecker := NewHealthChecker(mongoClient, redisClient, "test-1.0.0")

		Convey("Should check health successfully", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			health := healthChecker.CheckHealth(ctx)
			So(health, ShouldNotBeNil)
			So(health.Status, ShouldBeIn, StatusHealthy, StatusDegraded)
			So(health.Components, ShouldContainKey, "mongodb")
			So(health.Components, ShouldContainKey, "redis")

			// 测试JSON序列化
			jsonStr, err := health.ToJSON()
			So(err, ShouldBeNil)
			So(jsonStr, ShouldNotBeEmpty)

			// 测试HTTP状态码
			httpStatus := health.ToHTTPStatus()
			So(httpStatus, ShouldBeIn, 200, 503)
		})

		Convey("Should check readiness", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			isReady := healthChecker.CheckReadiness(ctx)
			So(isReady, ShouldBeTrue)
		})

		Convey("Should check liveness", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			isAlive := healthChecker.CheckLiveness(ctx)
			So(isAlive, ShouldBeTrue)
		})

		// 清理
		Reset(func() {
			mongoClient.Close(context.Background())
			redisClient.Close()
		})
	})
}
