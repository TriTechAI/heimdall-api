package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// TestAuthFlow 测试完整的用户认证流程
func TestAuthFlow(t *testing.T) {
	Convey("用户认证流程端到端测试", t, func() {
		// 获取测试配置
		config := GetTestConfig()
		client := NewTestClient()

		Convey("用户登录流程", func() {
			// 准备登录请求
			loginReq := map[string]interface{}{
				"username":   config.TestUser.Username,
				"password":   config.TestUser.Password,
				"rememberMe": false,
			}

			// 执行登录请求
			resp, err := client.Post(config.AdminAPIURL+"/api/v1/admin/auth/login", loginReq)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)

			// 解析登录响应
			var loginResp map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&loginResp)
			So(err, ShouldBeNil)
			So(loginResp["code"], ShouldEqual, 200)

			// 验证响应数据结构
			data := loginResp["data"].(map[string]interface{})
			So(data["token"], ShouldNotBeEmpty)
			So(data["user"], ShouldNotBeNil)

			// 保存Token用于后续测试
			token := data["token"].(string)
			client.SetAuthToken(token)

			Convey("获取用户信息", func() {
				// 使用Token获取用户信息
				resp, err := client.Get(config.AdminAPIURL + "/api/v1/admin/auth/profile")
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)

				// 解析用户信息响应
				var profileResp map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&profileResp)
				So(err, ShouldBeNil)
				So(profileResp["code"], ShouldEqual, 200)

				// 验证用户信息
				userData := profileResp["data"].(map[string]interface{})
				So(userData["username"], ShouldEqual, config.TestUser.Username)
				So(userData["email"], ShouldNotBeEmpty)
				So(userData["role"], ShouldNotBeEmpty)
			})

			Convey("Token失效测试", func() {
				// 登出操作
				logoutReq := map[string]interface{}{}
				resp, err := client.Post(config.AdminAPIURL+"/api/v1/admin/auth/logout", logoutReq)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)

				// 验证Token失效后无法访问受保护资源
				resp, err = client.Get(config.AdminAPIURL + "/api/v1/admin/auth/profile")
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusUnauthorized)
			})
		})

		Convey("登录失败场景测试", func() {
			testCases := []struct {
				name     string
				username string
				password string
				expected int
			}{
				{"错误用户名", "wronguser", config.TestUser.Password, http.StatusUnauthorized},
				{"错误密码", config.TestUser.Username, "wrongpassword", http.StatusUnauthorized},
				{"空用户名", "", config.TestUser.Password, http.StatusBadRequest},
				{"空密码", config.TestUser.Username, "", http.StatusBadRequest},
			}

			for _, tc := range testCases {
				Convey(tc.name, func() {
					loginReq := map[string]interface{}{
						"username": tc.username,
						"password": tc.password,
					}

					resp, err := client.Post(config.AdminAPIURL+"/api/v1/admin/auth/login", loginReq)
					So(err, ShouldBeNil)
					So(resp.StatusCode, ShouldEqual, tc.expected)
				})
			}
		})

		Convey("认证安全性测试", func() {
			Convey("无Token访问受保护资源", func() {
				// 清除Token
				client.SetAuthToken("")

				// 尝试访问需要认证的接口
				protectedEndpoints := []string{
					"/api/v1/admin/auth/profile",
					"/api/v1/admin/users",
					"/api/v1/admin/posts",
					"/api/v1/admin/pages",
				}

				for _, endpoint := range protectedEndpoints {
					resp, err := client.Get(config.AdminAPIURL + endpoint)
					So(err, ShouldBeNil)
					So(resp.StatusCode, ShouldEqual, http.StatusUnauthorized)
				}
			})

			Convey("无效Token测试", func() {
				// 设置无效Token
				invalidTokens := []string{
					"invalid.token.here",
					"Bearer invalid",
					"malformed_token",
					"",
				}

				for _, token := range invalidTokens {
					client.SetAuthToken(token)
					resp, err := client.Get(config.AdminAPIURL + "/api/v1/admin/auth/profile")
					So(err, ShouldBeNil)
					So(resp.StatusCode, ShouldEqual, http.StatusUnauthorized)
				}
			})
		})
	})
}

// TestLoginAttemptLimiting 测试登录尝试限制
func TestLoginAttemptLimiting(t *testing.T) {
	Convey("登录尝试限制测试", t, func() {
		config := GetTestConfig()
		client := NewTestClient()

		Convey("频繁失败登录应触发限制", func() {
			// 多次使用错误密码登录
			loginReq := map[string]interface{}{
				"username": config.TestUser.Username,
				"password": "wrongpassword",
			}

			// 连续失败登录
			for i := 0; i < 6; i++ { // 假设限制为5次
				resp, err := client.Post(config.AdminAPIURL+"/api/v1/admin/auth/login", loginReq)
				So(err, ShouldBeNil)
				
				if i < 5 {
					So(resp.StatusCode, ShouldEqual, http.StatusUnauthorized)
				} else {
					// 第6次应该被限制
					So(resp.StatusCode, ShouldEqual, http.StatusTooManyRequests)
				}
				
				// 短暂等待
				time.Sleep(100 * time.Millisecond)
			}

			// 即使使用正确密码也应该被限制
			correctLoginReq := map[string]interface{}{
				"username": config.TestUser.Username,
				"password": config.TestUser.Password,
			}

			resp, err := client.Post(config.AdminAPIURL+"/api/v1/admin/auth/login", correctLoginReq)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusTooManyRequests)
		})
	})
}