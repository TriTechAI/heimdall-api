package model

import (
	"testing"
	"time"

	"github.com/heimdall-api/common/constants"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPostModel(t *testing.T) {
	Convey("文章模型测试", t, func() {

		Convey("文章创建验证", func() {
			authorID := primitive.NewObjectID()

			Convey("有效的文章数据应该通过验证", func() {
				post := &Post{
					Title:      "测试文章标题",
					Markdown:   "# 测试文章内容\n\n这是一篇测试文章。",
					Type:       constants.PostTypePost,
					Status:     constants.PostStatusDraft,
					Visibility: constants.PostVisibilityPublic,
					AuthorID:   authorID,
					Tags:       []Tag{{Name: "测试", Slug: "test"}},
				}

				err := post.ValidateForCreate()
				So(err, ShouldBeNil)
			})

			Convey("缺少必填字段应该验证失败", func() {
				post := &Post{}

				err := post.ValidateForCreate()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "文章标题不能为空")
			})

			Convey("标题长度超限应该验证失败", func() {
				longTitle := make([]rune, constants.PostTitleMaxLength+1)
				for i := range longTitle {
					longTitle[i] = 'a'
				}

				post := &Post{
					Title:      string(longTitle),
					Markdown:   "内容",
					Type:       constants.PostTypePost,
					Status:     constants.PostStatusDraft,
					Visibility: constants.PostVisibilityPublic,
					AuthorID:   authorID,
				}

				err := post.ValidateForCreate()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "文章标题长度不能超过255字符")
			})

			Convey("无效的文章类型应该验证失败", func() {
				post := &Post{
					Title:      "测试标题",
					Markdown:   "内容",
					Type:       "invalid_type",
					Status:     constants.PostStatusDraft,
					Visibility: constants.PostVisibilityPublic,
					AuthorID:   authorID,
				}

				err := post.ValidateForCreate()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "无效的文章类型")
			})

			Convey("标签数量超限应该验证失败", func() {
				tags := make([]Tag, constants.PostTagMaxCount+1)
				for i := range tags {
					tags[i] = Tag{Name: "标签" + string(rune(i)), Slug: "tag" + string(rune(i))}
				}

				post := &Post{
					Title:      "测试标题",
					Markdown:   "内容",
					Type:       constants.PostTypePost,
					Status:     constants.PostStatusDraft,
					Visibility: constants.PostVisibilityPublic,
					AuthorID:   authorID,
					Tags:       tags,
				}

				err := post.ValidateForCreate()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "标签数量不能超过20个")
			})

			Convey("无效的slug格式应该验证失败", func() {
				post := &Post{
					Title:      "测试标题",
					Slug:       "Invalid Slug!",
					Markdown:   "内容",
					Type:       constants.PostTypePost,
					Status:     constants.PostStatusDraft,
					Visibility: constants.PostVisibilityPublic,
					AuthorID:   authorID,
				}

				err := post.ValidateForCreate()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "slug格式无效")
			})
		})

		Convey("文章更新验证", func() {
			Convey("有效的更新数据应该通过验证", func() {
				post := &Post{
					Title:   "更新后的标题",
					Excerpt: "更新后的摘要",
				}

				err := post.ValidateForUpdate()
				So(err, ShouldBeNil)
			})

			Convey("更新数据长度超限应该验证失败", func() {
				longTitle := make([]rune, constants.PostTitleMaxLength+1)
				for i := range longTitle {
					longTitle[i] = 'a'
				}

				post := &Post{
					Title: string(longTitle),
				}

				err := post.ValidateForUpdate()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "文章标题长度不能超过255字符")
			})
		})

		Convey("文章状态检查", func() {
			post := &Post{}

			Convey("已发布状态检查", func() {
				post.Status = constants.PostStatusPublished
				So(post.IsPublished(), ShouldBeTrue)
				So(post.IsDraft(), ShouldBeFalse)
				So(post.IsScheduled(), ShouldBeFalse)
			})

			Convey("草稿状态检查", func() {
				post.Status = constants.PostStatusDraft
				So(post.IsDraft(), ShouldBeTrue)
				So(post.IsPublished(), ShouldBeFalse)
				So(post.CanBePublished(), ShouldBeTrue)
			})

			Convey("定时发布状态检查", func() {
				post.Status = constants.PostStatusScheduled
				futureTime := time.Now().Add(time.Hour)
				post.PublishedAt = &futureTime

				So(post.IsScheduled(), ShouldBeTrue)
				So(post.ShouldBePublishedNow(), ShouldBeFalse)

				pastTime := time.Now().Add(-time.Hour)
				post.PublishedAt = &pastTime
				So(post.ShouldBePublishedNow(), ShouldBeTrue)
			})

			Convey("公开可见性检查", func() {
				post.Visibility = constants.PostVisibilityPublic
				So(post.IsPublic(), ShouldBeTrue)

				post.Visibility = constants.PostVisibilityPrivate
				So(post.IsPublic(), ShouldBeFalse)
			})
		})

		Convey("Slug处理", func() {
			post := &Post{}

			Convey("从标题生成slug", func() {
				post.Title = "这是一个测试标题 With English"
				slug := post.GenerateSlug()
				So(slug, ShouldNotBeEmpty)
				So(IsValidSlug(slug), ShouldBeTrue)
			})

			Convey("确保文章有slug", func() {
				post.Title = "测试标题"
				post.EnsureSlug()
				So(post.Slug, ShouldNotBeEmpty)
				So(IsValidSlug(post.Slug), ShouldBeTrue)
			})

			Convey("slug格式验证", func() {
				So(IsValidSlug("valid-slug-123"), ShouldBeTrue)
				So(IsValidSlug("invalid_slug"), ShouldBeFalse)
				So(IsValidSlug("Invalid-Slug"), ShouldBeFalse)
				So(IsValidSlug("-invalid"), ShouldBeFalse)
				So(IsValidSlug("invalid-"), ShouldBeFalse)
				So(IsValidSlug(""), ShouldBeFalse)
			})
		})

		Convey("内容处理", func() {
			post := &Post{}

			Convey("字数统计", func() {
				post.Markdown = "# 标题\n\n这是一个 **粗体** 和 *斜体* 的测试文章。[链接](http://example.com)"
				wordCount := post.CalculateWordCount()
				So(wordCount, ShouldBeGreaterThan, 0)
			})

			Convey("阅读时间计算", func() {
				post.WordCount = 400
				readingTime := post.CalculateReadingTime()
				So(readingTime, ShouldEqual, 2) // 400字 / 200字每分钟 = 2分钟
			})

			Convey("更新内容指标", func() {
				post.Markdown = "这是一篇测试文章，包含多个单词用于测试字数统计功能。"
				post.UpdateContentMetrics()
				So(post.WordCount, ShouldBeGreaterThan, 0)
				So(post.ReadingTime, ShouldBeGreaterThan, 0)
			})

			Convey("自动生成摘要", func() {
				post.Markdown = "# 标题\n\n这是一篇很长的文章内容，用于测试自动生成摘要的功能。文章包含多个段落和各种markdown格式。"
				excerpt := post.GenerateExcerpt(50)
				So(excerpt, ShouldNotBeEmpty)
				So(len([]rune(excerpt)), ShouldBeLessThanOrEqualTo, 53) // 50 + "..."
			})

			Convey("确保文章有摘要", func() {
				post.Markdown = "这是文章内容"
				post.EnsureExcerpt()
				So(post.Excerpt, ShouldNotBeEmpty)
			})
		})

		Convey("文章转换方法", func() {
			authorID := primitive.NewObjectID()
			author := &AuthorInfo{
				ID:          authorID.Hex(),
				Username:    "testuser",
				DisplayName: "Test User",
			}

			post := &Post{
				ID:          primitive.NewObjectID(),
				Title:       "测试文章",
				Slug:        "test-post",
				Excerpt:     "测试摘要",
				Type:        constants.PostTypePost,
				Status:      constants.PostStatusPublished,
				Visibility:  constants.PostVisibilityPublic,
				AuthorID:    authorID,
				Tags:        []Tag{{Name: "测试", Slug: "test"}},
				ReadingTime: 5,
				ViewCount:   100,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			Convey("转换为详情响应", func() {
				response := post.ToDetailResponse(author)
				So(response, ShouldNotBeNil)
				So(response.ID, ShouldEqual, post.ID.Hex())
				So(response.Title, ShouldEqual, post.Title)
				So(response.Author.Username, ShouldEqual, author.Username)
			})

			Convey("转换为列表项", func() {
				listItem := post.ToListItem(author)
				So(listItem, ShouldNotBeNil)
				So(listItem.ID, ShouldEqual, post.ID.Hex())
				So(listItem.Title, ShouldEqual, post.Title)
				So(listItem.Author.Username, ShouldEqual, author.Username)
			})
		})

		Convey("文章工厂方法", func() {
			authorID := primitive.NewObjectID()

			Convey("创建新文章", func() {
				post := NewPost("测试标题", "测试内容", constants.PostTypePost, constants.PostStatusDraft, constants.PostVisibilityPublic, authorID)
				So(post, ShouldNotBeNil)
				So(post.ID, ShouldNotEqual, primitive.NilObjectID)
				So(post.Title, ShouldEqual, "测试标题")
				So(post.Slug, ShouldNotBeEmpty)
				So(post.Excerpt, ShouldNotBeEmpty)
				So(post.WordCount, ShouldBeGreaterThan, 0)
				So(post.ReadingTime, ShouldBeGreaterThan, 0)
			})

			Convey("从创建请求创建文章", func() {
				req := &PostCreateRequest{
					Title:      "测试文章",
					Markdown:   "测试内容",
					Type:       constants.PostTypePost,
					Status:     constants.PostStatusDraft,
					Visibility: constants.PostVisibilityPublic,
					Tags: []TagInfo{
						{Name: "Go", Slug: "go"},
						{Name: "测试"},
					},
				}

				post := NewPostFromCreateRequest(req, authorID)
				So(post, ShouldNotBeNil)
				So(post.Title, ShouldEqual, req.Title)
				So(len(post.Tags), ShouldEqual, 2)
				So(post.Tags[0].Slug, ShouldEqual, "go")
				So(post.Tags[1].Slug, ShouldNotBeEmpty) // 自动生成
			})
		})

		Convey("数据库操作辅助方法", func() {
			post := &Post{
				Title:    "测试文章",
				Markdown: "测试内容",
			}

			Convey("准备插入数据库", func() {
				post.PrepareForInsert()
				So(post.ID, ShouldNotEqual, primitive.NilObjectID)
				So(post.CreatedAt, ShouldNotBeZeroValue)
				So(post.UpdatedAt, ShouldNotBeZeroValue)
				So(post.Type, ShouldEqual, constants.PostTypePost)
				So(post.Status, ShouldEqual, constants.PostStatusDraft)
				So(post.Visibility, ShouldEqual, constants.PostVisibilityPublic)
				So(post.Tags, ShouldNotBeNil)
			})

			Convey("准备更新数据库", func() {
				oldTime := post.UpdatedAt
				time.Sleep(time.Millisecond) // 确保时间差异
				post.PrepareForUpdate()
				So(post.UpdatedAt, ShouldHappenAfter, oldTime)
			})

			Convey("增加浏览量", func() {
				oldCount := post.ViewCount
				post.IncrementViewCount()
				So(post.ViewCount, ShouldEqual, oldCount+1)
			})

			Convey("发布文章", func() {
				post.Status = constants.PostStatusDraft
				post.Publish()
				So(post.Status, ShouldEqual, constants.PostStatusPublished)
				So(post.PublishedAt, ShouldNotBeNil)
			})

			Convey("取消发布", func() {
				post.Status = constants.PostStatusPublished
				post.Unpublish()
				So(post.Status, ShouldEqual, constants.PostStatusDraft)
			})
		})

		Convey("工具函数测试", func() {
			Convey("从文本生成slug", func() {
				testCases := []struct {
					input    string
					expected string
				}{
					{"Hello World", "hello-world"},
					{"Go 语言编程", "go"},
					{"Test123", "test123"},
					{"", ""},
				}

				for _, tc := range testCases {
					result := GenerateSlugFromText(tc.input)
					if tc.expected == "" {
						So(result, ShouldStartWith, "post-")
					} else {
						So(result, ShouldEqual, tc.expected)
					}
				}
			})
		})

		Convey("验证错误", func() {
			Convey("创建验证错误", func() {
				err := NewPostValidationError("title", "标题不能为空")
				So(err, ShouldNotBeNil)
				So(err.Field, ShouldEqual, "title")
				So(err.Message, ShouldEqual, "标题不能为空")
				So(err.Error(), ShouldEqual, "标题不能为空")
			})
		})
	})
}
