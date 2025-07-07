package e2e

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// TestPostCRUD 测试文章CRUD操作的端到端流程
func TestPostCRUD(t *testing.T) {
	Convey("文章CRUD操作端到端测试", t, func() {
		config := GetTestConfig()
		client := NewTestClient()

		// 首先登录获取Token
		token := authenticateUser(client, config)
		client.SetAuthToken(token)

		var createdPostID string

		Convey("创建文章", func() {
			// 准备文章创建数据
			createReq := map[string]interface{}{
				"title":      "E2E测试文章标题",
				"markdown":   "# E2E测试文章\n\n这是一篇用于端到端测试的文章内容。\n\n## 功能测试\n\n测试文章的创建、读取、更新和删除功能。",
				"type":       "post",
				"status":     "draft",
				"visibility": "public",
				"tags": []map[string]string{
					{"name": "测试", "slug": "test"},
					{"name": "E2E", "slug": "e2e"},
				},
				"metaTitle":       "E2E测试文章 - SEO标题",
				"metaDescription": "这是一篇用于端到端测试的文章，测试文章管理功能。",
			}

			// 发送创建请求
			resp, err := client.Post(config.AdminAPIURL+"/api/v1/admin/posts", createReq)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)

			// 解析响应
			var createResp map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&createResp)
			So(err, ShouldBeNil)
			So(createResp["code"], ShouldEqual, 200)

			// 验证创建的文章数据
			data := createResp["data"].(map[string]interface{})
			So(data["id"], ShouldNotBeEmpty)
			So(data["title"], ShouldEqual, createReq["title"])
			So(data["status"], ShouldEqual, "draft")
			So(data["slug"], ShouldNotBeEmpty)

			// 保存文章ID用于后续测试
			createdPostID = data["id"].(string)

			Convey("获取文章详情", func() {
				// 获取刚创建的文章详情
				resp, err := client.Get(config.AdminAPIURL + "/api/v1/admin/posts/" + createdPostID)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)

				// 解析响应
				var detailResp map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&detailResp)
				So(err, ShouldBeNil)
				So(detailResp["code"], ShouldEqual, 200)

				// 验证文章详情数据
				data := detailResp["data"].(map[string]interface{})
				So(data["id"], ShouldEqual, createdPostID)
				So(data["title"], ShouldEqual, createReq["title"])
				So(data["markdown"], ShouldEqual, createReq["markdown"])
				So(data["wordCount"], ShouldBeGreaterThan, 0)
				So(data["readingTime"], ShouldBeGreaterThan, 0)
			})

			Convey("获取文章列表", func() {
				// 获取文章列表
				resp, err := client.Get(config.AdminAPIURL + "/api/v1/admin/posts?page=1&limit=10")
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)

				// 解析响应
				var listResp map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&listResp)
				So(err, ShouldBeNil)
				So(listResp["code"], ShouldEqual, 200)

				// 验证分页数据
				data := listResp["data"].(map[string]interface{})
				pagination := data["pagination"].(map[string]interface{})
				So(pagination["page"], ShouldEqual, 1)
				So(pagination["limit"], ShouldEqual, 10)

				// 验证列表包含刚创建的文章
				list := data["list"].([]interface{})
				So(len(list), ShouldBeGreaterThan, 0)

				// 查找创建的文章
				found := false
				for _, item := range list {
					post := item.(map[string]interface{})
					if post["id"].(string) == createdPostID {
						found = true
						So(post["title"], ShouldEqual, createReq["title"])
						break
					}
				}
				So(found, ShouldBeTrue)
			})

			Convey("更新文章", func() {
				// 准备更新数据
				updateReq := map[string]interface{}{
					"title":    "更新后的E2E测试文章标题",
					"markdown": "# 更新后的E2E测试文章\n\n这是更新后的文章内容。\n\n## 更新测试\n\n验证文章更新功能正常工作。",
					"status":   "published",
					"tags": []map[string]string{
						{"name": "测试", "slug": "test"},
						{"name": "更新", "slug": "update"},
						{"name": "E2E", "slug": "e2e"},
					},
				}

				// 发送更新请求
				resp, err := client.Put(config.AdminAPIURL+"/api/v1/admin/posts/"+createdPostID, updateReq)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)

				// 解析响应
				var updateResp map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&updateResp)
				So(err, ShouldBeNil)
				So(updateResp["code"], ShouldEqual, 200)

				// 验证更新结果
				data := updateResp["data"].(map[string]interface{})
				So(data["title"], ShouldEqual, updateReq["title"])
				So(data["status"], ShouldEqual, "published")
				So(data["publishedAt"], ShouldNotBeEmpty)

				Convey("发布状态管理", func() {
					// 测试取消发布
					resp, err := client.Post(config.AdminAPIURL+"/api/v1/admin/posts/"+createdPostID+"/unpublish", map[string]interface{}{})
					So(err, ShouldBeNil)
					So(resp.StatusCode, ShouldEqual, http.StatusOK)

					// 验证状态变更
					var unpublishResp map[string]interface{}
					err = json.NewDecoder(resp.Body).Decode(&unpublishResp)
					So(err, ShouldBeNil)
					So(unpublishResp["code"], ShouldEqual, 200)

					data := unpublishResp["data"].(map[string]interface{})
					So(data["status"], ShouldEqual, "draft")

					// 测试重新发布
					publishReq := map[string]interface{}{
						"publishedAt": time.Now().Format(time.RFC3339),
					}
					resp, err = client.Post(config.AdminAPIURL+"/api/v1/admin/posts/"+createdPostID+"/publish", publishReq)
					So(err, ShouldBeNil)
					So(resp.StatusCode, ShouldEqual, http.StatusOK)

					// 验证重新发布
					var publishResp map[string]interface{}
					err = json.NewDecoder(resp.Body).Decode(&publishResp)
					So(err, ShouldBeNil)
					So(publishResp["code"], ShouldEqual, 200)

					data = publishResp["data"].(map[string]interface{})
					So(data["status"], ShouldEqual, "published")
					So(data["publishedAt"], ShouldNotBeEmpty)
				})
			})

			Convey("文章搜索和过滤", func() {
				// 测试按标题搜索
				resp, err := client.Get(config.AdminAPIURL + "/api/v1/admin/posts?keyword=E2E&status=published")
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)

				var searchResp map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&searchResp)
				So(err, ShouldBeNil)
				So(searchResp["code"], ShouldEqual, 200)

				// 验证搜索结果
				data := searchResp["data"].(map[string]interface{})
				list := data["list"].([]interface{})
				
				// 应该能找到我们的测试文章
				found := false
				for _, item := range list {
					post := item.(map[string]interface{})
					if post["id"].(string) == createdPostID {
						found = true
						break
					}
				}
				So(found, ShouldBeTrue)

				// 测试按状态过滤
				resp, err = client.Get(config.AdminAPIURL + "/api/v1/admin/posts?status=draft")
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)

				// 测试排序
				resp, err = client.Get(config.AdminAPIURL + "/api/v1/admin/posts?sortBy=title&sortDesc=false")
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)
			})

			Convey("删除文章", func() {
				// 删除文章
				resp, err := client.Delete(config.AdminAPIURL + "/api/v1/admin/posts/" + createdPostID)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)

				// 解析删除响应
				var deleteResp map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&deleteResp)
				So(err, ShouldBeNil)
				So(deleteResp["code"], ShouldEqual, 200)

				// 验证文章已被删除（软删除，状态变为archived）
				resp, err = client.Get(config.AdminAPIURL + "/api/v1/admin/posts/" + createdPostID)
				So(err, ShouldBeNil)
				
				// 根据具体实现，删除的文章可能返回404或状态为archived
				if resp.StatusCode == http.StatusOK {
					var detailResp map[string]interface{}
					err = json.NewDecoder(resp.Body).Decode(&detailResp)
					So(err, ShouldBeNil)
					
					if detailResp["code"].(float64) == 200 {
						data := detailResp["data"].(map[string]interface{})
						// 软删除的文章状态应该是archived
						So(data["status"], ShouldEqual, "archived")
					}
				}
			})
		})

		Convey("文章权限验证", func() {
			// 清除Token测试无权限访问
			client.SetAuthToken("")

			// 尝试创建文章（应该失败）
			createReq := map[string]interface{}{
				"title":    "未授权测试文章",
				"markdown": "这应该创建失败",
				"type":     "post",
				"status":   "draft",
			}

			resp, err := client.Post(config.AdminAPIURL+"/api/v1/admin/posts", createReq)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusUnauthorized)

			// 尝试获取文章列表（应该失败）
			resp, err = client.Get(config.AdminAPIURL + "/api/v1/admin/posts")
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusUnauthorized)
		})
	})
}

// TestPostValidation 测试文章数据验证
func TestPostValidation(t *testing.T) {
	Convey("文章数据验证测试", t, func() {
		config := GetTestConfig()
		client := NewTestClient()

		// 登录获取Token
		token := authenticateUser(client, config)
		client.SetAuthToken(token)

		Convey("无效数据应该被拒绝", func() {
			testCases := []struct {
				name     string
				data     map[string]interface{}
				expected int
			}{
				{
					"缺少标题",
					map[string]interface{}{
						"markdown": "内容",
						"type":     "post",
						"status":   "draft",
					},
					http.StatusBadRequest,
				},
				{
					"缺少内容",
					map[string]interface{}{
						"title":  "标题",
						"type":   "post",
						"status": "draft",
					},
					http.StatusBadRequest,
				},
				{
					"无效状态",
					map[string]interface{}{
						"title":    "标题",
						"markdown": "内容",
						"type":     "post",
						"status":   "invalid_status",
					},
					http.StatusBadRequest,
				},
				{
					"无效类型",
					map[string]interface{}{
						"title":    "标题",
						"markdown": "内容",
						"type":     "invalid_type",
						"status":   "draft",
					},
					http.StatusBadRequest,
				},
			}

			for _, tc := range testCases {
				Convey(tc.name, func() {
					resp, err := client.Post(config.AdminAPIURL+"/api/v1/admin/posts", tc.data)
					So(err, ShouldBeNil)
					So(resp.StatusCode, ShouldEqual, tc.expected)
				})
			}
		})
	})
}

// authenticateUser 辅助函数：用户认证获取Token
func authenticateUser(client *TestClient, config *TestConfig) string {
	loginReq := map[string]interface{}{
		"username": config.TestUser.Username,
		"password": config.TestUser.Password,
	}

	resp, err := client.Post(config.AdminAPIURL+"/api/v1/admin/auth/login", loginReq)
	if err != nil || resp.StatusCode != http.StatusOK {
		panic("Failed to authenticate test user")
	}

	var loginResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		panic("Failed to parse login response")
	}

	data := loginResp["data"].(map[string]interface{})
	return data["token"].(string)
}