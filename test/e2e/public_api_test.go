package e2e

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// TestPublicAPI 测试公开API访问的端到端流程
func TestPublicAPI(t *testing.T) {
	Convey("公开API访问端到端测试", t, func() {
		config := GetTestConfig()
		client := NewTestClient()

		// 准备测试数据：先在管理API创建一些已发布的文章
		var publishedPostID string
		var publishedPostSlug string

		Convey("准备测试数据", func() {
			// 登录管理系统
			adminClient := NewTestClient()
			token := authenticateUser(adminClient, config)
			adminClient.SetAuthToken(token)

			// 创建测试文章
			createReq := map[string]interface{}{
				"title":      "公开API测试文章",
				"slug":       "public-api-test-post",
				"markdown":   "# 公开API测试\n\n这是一篇用于测试公开API的文章。\n\n## 内容展示\n\n测试公开接口的文章展示功能。",
				"excerpt":    "这是一篇用于测试公开API的文章摘要。",
				"type":       "post",
				"status":     "published",
				"visibility": "public",
				"tags": []map[string]string{
					{"name": "公开", "slug": "public"},
					{"name": "测试", "slug": "test"},
				},
				"metaTitle":       "公开API测试文章 - SEO标题",
				"metaDescription": "这是一篇用于测试公开API的文章，验证前台展示功能。",
			}

			// 创建文章
			resp, err := adminClient.Post(config.AdminAPIURL+"/api/v1/admin/posts", createReq)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)

			var createResp map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&createResp)
			So(err, ShouldBeNil)

			data := createResp["data"].(map[string]interface{})
			publishedPostID = data["id"].(string)
			publishedPostSlug = data["slug"].(string)
		})

		Convey("获取公开文章列表", func() {
			// 获取公开文章列表
			resp, err := client.Get(config.PublicAPIURL + "/api/v1/public/posts?page=1&limit=10")
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)

			// 解析响应
			var listResp map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&listResp)
			So(err, ShouldBeNil)
			So(listResp["code"], ShouldEqual, 200)

			// 验证响应数据结构
			data := listResp["data"].(map[string]interface{})
			So(data["list"], ShouldNotBeNil)
			So(data["pagination"], ShouldNotBeNil)

			// 验证分页信息
			pagination := data["pagination"].(map[string]interface{})
			So(pagination["page"], ShouldEqual, 1)
			So(pagination["limit"], ShouldEqual, 10)

			// 验证文章列表
			list := data["list"].([]interface{})
			So(len(list), ShouldBeGreaterThan, 0)

			// 验证列表项数据结构（不应包含敏感信息）
			for _, item := range list {
				post := item.(map[string]interface{})
				So(post["id"], ShouldNotBeEmpty)
				So(post["title"], ShouldNotBeEmpty)
				So(post["slug"], ShouldNotBeEmpty)
				So(post["excerpt"], ShouldNotBeEmpty)
				So(post["status"], ShouldEqual, "published") // 只有已发布的文章
				So(post["visibility"], ShouldEqual, "public") // 只有公开的文章
				So(post["author"], ShouldNotBeNil)
				
				// 确保不包含敏感信息
				So(post["markdown"], ShouldBeNil) // 列表不应包含完整内容
			}

			Convey("分页功能测试", func() {
				// 测试第二页
				resp, err := client.Get(config.PublicAPIURL + "/api/v1/public/posts?page=2&limit=5")
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)

				var page2Resp map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&page2Resp)
				So(err, ShouldBeNil)
				So(page2Resp["code"], ShouldEqual, 200)

				data := page2Resp["data"].(map[string]interface{})
				pagination := data["pagination"].(map[string]interface{})
				So(pagination["page"], ShouldEqual, 2)
				So(pagination["limit"], ShouldEqual, 5)
			})

			Convey("排序功能测试", func() {
				// 测试按发布时间倒序
				resp, err := client.Get(config.PublicAPIURL + "/api/v1/public/posts?sortBy=publishedAt&sortDesc=true")
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)

				// 测试按标题正序
				resp, err = client.Get(config.PublicAPIURL + "/api/v1/public/posts?sortBy=title&sortDesc=false")
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)
			})

			Convey("过滤功能测试", func() {
				// 测试按标签过滤
				resp, err := client.Get(config.PublicAPIURL + "/api/v1/public/posts?tag=test")
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)

				// 测试关键词搜索
				resp, err = client.Get(config.PublicAPIURL + "/api/v1/public/posts?keyword=测试")
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)
			})
		})

		Convey("获取文章详情", func() {
			// 通过slug获取文章详情
			resp, err := client.Get(config.PublicAPIURL + "/api/v1/public/posts/" + publishedPostSlug)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)

			// 解析响应
			var detailResp map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&detailResp)
			So(err, ShouldBeNil)
			So(detailResp["code"], ShouldEqual, 200)

			// 验证文章详情数据
			data := detailResp["data"].(map[string]interface{})
			So(data["id"], ShouldEqual, publishedPostID)
			So(data["title"], ShouldNotBeEmpty)
			So(data["slug"], ShouldEqual, publishedPostSlug)
			So(data["markdown"], ShouldNotBeEmpty) // 详情页应包含完整内容
			So(data["html"], ShouldNotBeEmpty)     // 应包含HTML内容
			So(data["excerpt"], ShouldNotBeEmpty)
			So(data["status"], ShouldEqual, "published")
			So(data["visibility"], ShouldEqual, "public")
			So(data["author"], ShouldNotBeNil)
			So(data["tags"], ShouldNotBeNil)
			So(data["readingTime"], ShouldBeGreaterThan, 0)
			So(data["wordCount"], ShouldBeGreaterThan, 0)
			So(data["viewCount"], ShouldBeGreaterThanOrEqualTo, 0)
			So(data["publishedAt"], ShouldNotBeEmpty)
			So(data["createdAt"], ShouldNotBeEmpty)
			So(data["updatedAt"], ShouldNotBeEmpty)

			// 验证SEO元数据
			So(data["metaTitle"], ShouldNotBeEmpty)
			So(data["metaDescription"], ShouldNotBeEmpty)

			// 验证作者信息结构
			author := data["author"].(map[string]interface{})
			So(author["id"], ShouldNotBeEmpty)
			So(author["username"], ShouldNotBeEmpty)
			So(author["displayName"], ShouldNotBeEmpty)
			// 确保不包含敏感信息
			So(author["email"], ShouldBeNil)

			Convey("浏览计数功能", func() {
				// 记录初始浏览数
				initialViewCount := int64(data["viewCount"].(float64))

				// 再次访问文章详情
				resp, err := client.Get(config.PublicAPIURL + "/api/v1/public/posts/" + publishedPostSlug)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)

				var secondVisitResp map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&secondVisitResp)
				So(err, ShouldBeNil)

				data := secondVisitResp["data"].(map[string]interface{})
				newViewCount := int64(data["viewCount"].(float64))

				// 浏览计数应该增加
				So(newViewCount, ShouldBeGreaterThan, initialViewCount)
			})
		})

		Convey("访问控制测试", func() {
			Convey("访问不存在的文章", func() {
				resp, err := client.Get(config.PublicAPIURL + "/api/v1/public/posts/non-existent-slug")
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
			})

			Convey("访问私有文章应被拒绝", func() {
				// 首先在管理系统创建一篇私有文章
				adminClient := NewTestClient()
				token := authenticateUser(adminClient, config)
				adminClient.SetAuthToken(token)

				privateReq := map[string]interface{}{
					"title":      "私有文章",
					"slug":       "private-post",
					"markdown":   "这是私有内容",
					"type":       "post",
					"status":     "published",
					"visibility": "private",
				}

				resp, err := adminClient.Post(config.AdminAPIURL+"/api/v1/admin/posts", privateReq)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)

				// 尝试通过公开API访问私有文章
				resp, err = client.Get(config.PublicAPIURL + "/api/v1/public/posts/private-post")
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusNotFound) // 私有文章应返回404
			})

			Convey("访问草稿文章应被拒绝", func() {
				// 首先在管理系统创建一篇草稿文章
				adminClient := NewTestClient()
				token := authenticateUser(adminClient, config)
				adminClient.SetAuthToken(token)

				draftReq := map[string]interface{}{
					"title":      "草稿文章",
					"slug":       "draft-post",
					"markdown":   "这是草稿内容",
					"type":       "post",
					"status":     "draft",
					"visibility": "public",
				}

				resp, err := adminClient.Post(config.AdminAPIURL+"/api/v1/admin/posts", draftReq)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)

				// 尝试通过公开API访问草稿文章
				resp, err = client.Get(config.PublicAPIURL + "/api/v1/public/posts/draft-post")
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusNotFound) // 草稿文章应返回404
			})
		})

		Convey("API性能和安全测试", func() {
			Convey("响应时间测试", func() {
				// 测试文章列表响应时间
				start := time.Now()
				resp, err := client.Get(config.PublicAPIURL + "/api/v1/public/posts?limit=50")
				duration := time.Since(start)

				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)
				So(duration.Milliseconds(), ShouldBeLessThan, 2000) // 响应时间应小于2秒

				// 测试文章详情响应时间
				start = time.Now()
				resp, err = client.Get(config.PublicAPIURL + "/api/v1/public/posts/" + publishedPostSlug)
				duration = time.Since(start)

				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)
				So(duration.Milliseconds(), ShouldBeLessThan, 1000) // 响应时间应小于1秒
			})

			Convey("输入验证测试", func() {
				// 测试无效的分页参数
				testCases := []string{
					"/api/v1/public/posts?page=-1",
					"/api/v1/public/posts?limit=0",
					"/api/v1/public/posts?limit=1000", // 超过限制
					"/api/v1/public/posts?sortBy=invalid_field",
				}

				for _, testCase := range testCases {
					resp, err := client.Get(config.PublicAPIURL + testCase)
					So(err, ShouldBeNil)
					// 应该返回400错误或使用默认值并返回200
					So(resp.StatusCode, ShouldBeIn, []int{http.StatusOK, http.StatusBadRequest})
				}
			})

			Convey("XSS防护测试", func() {
				// 测试搜索关键词中的XSS
				xssPayload := "<script>alert('xss')</script>"
				resp, err := client.Get(config.PublicAPIURL + "/api/v1/public/posts?keyword=" + xssPayload)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)

				// 响应内容不应包含未转义的脚本
				body, err := client.GetResponseBody(resp)
				So(err, ShouldBeNil)
				So(body, ShouldNotContainSubstring, "<script>")
			})
		})

		// 清理测试数据
		Reset(func() {
			if publishedPostID != "" {
				adminClient := NewTestClient()
				token := authenticateUser(adminClient, config)
				adminClient.SetAuthToken(token)
				adminClient.Delete(config.AdminAPIURL + "/api/v1/admin/posts/" + publishedPostID)
			}
		})
	})
}