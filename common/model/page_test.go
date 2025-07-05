package model

import (
	"testing"
	"time"

	"github.com/heimdall-api/common/constants"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPageModel(t *testing.T) {
	Convey("页面模型测试", t, func() {
		authorID := primitive.NewObjectID()

		Convey("页面创建", func() {
			Convey("使用NewPage工厂方法", func() {
				page := NewPage("测试页面", "测试内容", constants.PostStatusDraft, authorID)

				So(page, ShouldNotBeNil)
				So(page.Title, ShouldEqual, "测试页面")
				So(page.Content, ShouldEqual, "测试内容")
				So(page.Status, ShouldEqual, constants.PostStatusDraft)
				So(page.AuthorID, ShouldEqual, authorID)
				So(page.Template, ShouldEqual, "default")
				So(page.Slug, ShouldNotBeEmpty)
				So(page.CreatedAt, ShouldNotBeZeroValue)
				So(page.UpdatedAt, ShouldNotBeZeroValue)
			})

			Convey("使用NewPageFromCreateRequest工厂方法", func() {
				req := &PageCreateRequest{
					Title:           "测试页面",
					Slug:            "test-page",
					Content:         "测试内容",
					Status:          constants.PostStatusDraft,
					Template:        "custom",
					MetaTitle:       "SEO标题",
					MetaDescription: "SEO描述",
				}

				page := NewPageFromCreateRequest(req, authorID)

				So(page, ShouldNotBeNil)
				So(page.Title, ShouldEqual, req.Title)
				So(page.Slug, ShouldEqual, req.Slug)
				So(page.Content, ShouldEqual, req.Content)
				So(page.Status, ShouldEqual, req.Status)
				So(page.Template, ShouldEqual, req.Template)
				So(page.MetaTitle, ShouldEqual, req.MetaTitle)
				So(page.MetaDescription, ShouldEqual, req.MetaDescription)
				So(page.AuthorID, ShouldEqual, authorID)
			})

			Convey("从创建请求创建页面时设置默认模板", func() {
				req := &PageCreateRequest{
					Title:   "测试页面",
					Content: "测试内容",
					Status:  constants.PostStatusDraft,
				}

				page := NewPageFromCreateRequest(req, authorID)

				So(page.Template, ShouldEqual, "default")
			})
		})

		Convey("页面验证", func() {
			Convey("创建验证", func() {
				Convey("有效的页面数据", func() {
					page := &Page{
						Title:    "测试页面",
						Content:  "测试内容",
						Status:   constants.PostStatusDraft,
						AuthorID: authorID,
					}

					err := page.ValidateForCreate()
					So(err, ShouldBeNil)
				})

				Convey("缺少必填字段", func() {
					testCases := []struct {
						name     string
						page     *Page
						expected string
					}{
						{
							name: "缺少标题",
							page: &Page{
								Content:  "测试内容",
								Status:   constants.PostStatusDraft,
								AuthorID: authorID,
							},
							expected: "页面标题不能为空",
						},
						{
							name: "缺少内容",
							page: &Page{
								Title:    "测试页面",
								Status:   constants.PostStatusDraft,
								AuthorID: authorID,
							},
							expected: "页面内容不能为空",
						},
						{
							name: "缺少状态",
							page: &Page{
								Title:    "测试页面",
								Content:  "测试内容",
								AuthorID: authorID,
							},
							expected: "页面状态不能为空",
						},
						{
							name: "缺少作者ID",
							page: &Page{
								Title:   "测试页面",
								Content: "测试内容",
								Status:  constants.PostStatusDraft,
							},
							expected: "作者ID不能为空",
						},
					}

					for _, tc := range testCases {
						Convey(tc.name, func() {
							err := tc.page.ValidateForCreate()
							So(err, ShouldNotBeNil)
							So(err.Error(), ShouldContainSubstring, tc.expected)
						})
					}
				})

				Convey("字段长度验证", func() {
					longTitle := string(make([]byte, constants.PostTitleMaxLength+1))
					longSlug := string(make([]byte, constants.PostSlugMaxLength+1))
					longContent := string(make([]byte, constants.PostContentMaxLength+1))
					longTemplate := string(make([]byte, 101))
					longMetaTitle := string(make([]byte, constants.PostMetaTitleMaxLength+1))
					longMetaDesc := string(make([]byte, constants.PostMetaDescMaxLength+1))
					longCanonicalURL := string(make([]byte, constants.PostCanonicalUrlMaxLength+1))

					testCases := []struct {
						name     string
						page     *Page
						expected string
					}{
						{
							name: "标题过长",
							page: &Page{
								Title:    longTitle,
								Content:  "测试内容",
								Status:   constants.PostStatusDraft,
								AuthorID: authorID,
							},
							expected: "页面标题长度不能超过255字符",
						},
						{
							name: "Slug过长",
							page: &Page{
								Title:    "测试页面",
								Slug:     longSlug,
								Content:  "测试内容",
								Status:   constants.PostStatusDraft,
								AuthorID: authorID,
							},
							expected: "页面slug长度不能超过255字符",
						},
						{
							name: "内容过长",
							page: &Page{
								Title:    "测试页面",
								Content:  longContent,
								Status:   constants.PostStatusDraft,
								AuthorID: authorID,
							},
							expected: "页面内容长度不能超过1MB",
						},
						{
							name: "模板名过长",
							page: &Page{
								Title:    "测试页面",
								Content:  "测试内容",
								Template: longTemplate,
								Status:   constants.PostStatusDraft,
								AuthorID: authorID,
							},
							expected: "模板名称长度不能超过100字符",
						},
						{
							name: "SEO标题过长",
							page: &Page{
								Title:     "测试页面",
								Content:   "测试内容",
								MetaTitle: longMetaTitle,
								Status:    constants.PostStatusDraft,
								AuthorID:  authorID,
							},
							expected: "SEO标题长度不能超过70字符",
						},
						{
							name: "SEO描述过长",
							page: &Page{
								Title:           "测试页面",
								Content:         "测试内容",
								MetaDescription: longMetaDesc,
								Status:          constants.PostStatusDraft,
								AuthorID:        authorID,
							},
							expected: "SEO描述长度不能超过160字符",
						},
						{
							name: "规范化URL过长",
							page: &Page{
								Title:        "测试页面",
								Content:      "测试内容",
								CanonicalURL: longCanonicalURL,
								Status:       constants.PostStatusDraft,
								AuthorID:     authorID,
							},
							expected: "规范化URL长度不能超过255字符",
						},
					}

					for _, tc := range testCases {
						Convey(tc.name, func() {
							err := tc.page.ValidateForCreate()
							So(err, ShouldNotBeNil)
							So(err.Error(), ShouldContainSubstring, tc.expected)
						})
					}
				})

				Convey("无效的枚举值", func() {
					page := &Page{
						Title:    "测试页面",
						Content:  "测试内容",
						Status:   "invalid_status",
						AuthorID: authorID,
					}

					err := page.ValidateForCreate()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "无效的页面状态")
				})

				Convey("无效的slug格式", func() {
					page := &Page{
						Title:    "测试页面",
						Slug:     "Invalid Slug!",
						Content:  "测试内容",
						Status:   constants.PostStatusDraft,
						AuthorID: authorID,
					}

					err := page.ValidateForCreate()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "slug格式无效")
				})
			})

			Convey("更新验证", func() {
				Convey("有效的更新数据", func() {
					page := &Page{
						Title:   "更新标题",
						Content: "更新内容",
						Status:  constants.PostStatusPublished,
					}

					err := page.ValidateForUpdate()
					So(err, ShouldBeNil)
				})

				Convey("更新时字段长度验证", func() {
					longTitle := string(make([]byte, constants.PostTitleMaxLength+1))

					page := &Page{
						Title: longTitle,
					}

					err := page.ValidateForUpdate()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "页面标题长度不能超过255字符")
				})
			})
		})

		Convey("状态检查方法", func() {
			testCases := []struct {
				status      string
				isPublished bool
				isDraft     bool
				isScheduled bool
				canPublish  bool
			}{
				{constants.PostStatusDraft, false, true, false, true},
				{constants.PostStatusPublished, true, false, false, false},
				{constants.PostStatusScheduled, false, false, true, true},
			}

			for _, tc := range testCases {
				Convey("状态: "+tc.status, func() {
					page := &Page{Status: tc.status}

					So(page.IsPublished(), ShouldEqual, tc.isPublished)
					So(page.IsDraft(), ShouldEqual, tc.isDraft)
					So(page.IsScheduled(), ShouldEqual, tc.isScheduled)
					So(page.CanBePublished(), ShouldEqual, tc.canPublish)
				})
			}

			Convey("定时发布检查", func() {
				pastTime := time.Now().Add(-1 * time.Hour)
				futureTime := time.Now().Add(1 * time.Hour)

				Convey("应该现在发布", func() {
					page := &Page{
						Status:      constants.PostStatusScheduled,
						PublishedAt: &pastTime,
					}

					So(page.ShouldBePublishedNow(), ShouldBeTrue)
				})

				Convey("不应该现在发布", func() {
					page := &Page{
						Status:      constants.PostStatusScheduled,
						PublishedAt: &futureTime,
					}

					So(page.ShouldBePublishedNow(), ShouldBeFalse)
				})

				Convey("非定时发布状态", func() {
					page := &Page{
						Status:      constants.PostStatusDraft,
						PublishedAt: &pastTime,
					}

					So(page.ShouldBePublishedNow(), ShouldBeFalse)
				})
			})
		})

		Convey("Slug处理", func() {
			Convey("自动生成slug", func() {
				page := &Page{Title: "测试页面标题"}
				slug := page.GenerateSlug()

				So(slug, ShouldNotBeEmpty)
				So(slug, ShouldNotContainSubstring, " ")
			})

			Convey("确保有slug", func() {
				Convey("没有slug时自动生成", func() {
					page := &Page{Title: "测试页面"}
					page.EnsureSlug()

					So(page.Slug, ShouldNotBeEmpty)
				})

				Convey("已有slug时不覆盖", func() {
					existingSlug := "existing-slug"
					page := &Page{
						Title: "测试页面",
						Slug:  existingSlug,
					}
					page.EnsureSlug()

					So(page.Slug, ShouldEqual, existingSlug)
				})
			})
		})

		Convey("转换方法", func() {
			author := &AuthorInfo{
				ID:          authorID.Hex(),
				Username:    "testuser",
				DisplayName: "Test User",
			}

			page := &Page{
				ID:              primitive.NewObjectID(),
				Title:           "测试页面",
				Slug:            "test-page",
				Content:         "测试内容",
				HTML:            "<p>测试内容</p>",
				AuthorID:        authorID,
				Status:          constants.PostStatusPublished,
				Template:        "custom",
				MetaTitle:       "SEO标题",
				MetaDescription: "SEO描述",
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}

			Convey("转换为详情响应", func() {
				response := page.ToDetailResponse(author)

				So(response, ShouldNotBeNil)
				So(response.ID, ShouldEqual, page.ID.Hex())
				So(response.Title, ShouldEqual, page.Title)
				So(response.Slug, ShouldEqual, page.Slug)
				So(response.Content, ShouldEqual, page.Content)
				So(response.HTML, ShouldEqual, page.HTML)
				So(response.Author, ShouldResemble, *author)
				So(response.Status, ShouldEqual, page.Status)
				So(response.Template, ShouldEqual, page.Template)
				So(response.MetaTitle, ShouldEqual, page.MetaTitle)
				So(response.MetaDescription, ShouldEqual, page.MetaDescription)
			})

			Convey("转换为列表项", func() {
				listItem := page.ToListItem(author)

				So(listItem, ShouldNotBeNil)
				So(listItem.ID, ShouldEqual, page.ID.Hex())
				So(listItem.Title, ShouldEqual, page.Title)
				So(listItem.Slug, ShouldEqual, page.Slug)
				So(listItem.Author, ShouldResemble, *author)
				So(listItem.Status, ShouldEqual, page.Status)
				So(listItem.Template, ShouldEqual, page.Template)
			})
		})

		Convey("准备方法", func() {
			Convey("准备插入", func() {
				page := &Page{
					Title:   "测试页面",
					Content: "测试内容",
					Status:  constants.PostStatusDraft,
				}

				page.PrepareForInsert()

				So(page.ID, ShouldNotBeZeroValue)
				So(page.CreatedAt, ShouldNotBeZeroValue)
				So(page.UpdatedAt, ShouldNotBeZeroValue)
				So(page.Template, ShouldEqual, "default")
				So(page.Slug, ShouldNotBeEmpty)
			})

			Convey("准备更新", func() {
				page := &Page{
					UpdatedAt: time.Now().Add(-1 * time.Hour),
				}
				oldUpdateTime := page.UpdatedAt

				page.PrepareForUpdate()

				So(page.UpdatedAt, ShouldHappenAfter, oldUpdateTime)
			})
		})

		Convey("发布管理", func() {
			Convey("发布页面", func() {
				page := &Page{
					Status: constants.PostStatusDraft,
				}

				page.Publish()

				So(page.Status, ShouldEqual, constants.PostStatusPublished)
				So(page.PublishedAt, ShouldNotBeNil)
			})

			Convey("发布已有发布时间的页面", func() {
				publishTime := time.Now().Add(-1 * time.Hour)
				page := &Page{
					Status:      constants.PostStatusDraft,
					PublishedAt: &publishTime,
				}

				page.Publish()

				So(page.Status, ShouldEqual, constants.PostStatusPublished)
				So(page.PublishedAt, ShouldEqual, &publishTime)
			})

			Convey("取消发布页面", func() {
				page := &Page{
					Status: constants.PostStatusPublished,
				}

				page.Unpublish()

				So(page.Status, ShouldEqual, constants.PostStatusDraft)
			})
		})

		Convey("错误类型", func() {
			Convey("页面验证错误", func() {
				err := NewPageValidationError("title", "标题错误")

				So(err, ShouldNotBeNil)
				So(err.Field, ShouldEqual, "title")
				So(err.Message, ShouldEqual, "标题错误")
				So(err.Error(), ShouldContainSubstring, "页面验证错误")
				So(err.Error(), ShouldContainSubstring, "title")
				So(err.Error(), ShouldContainSubstring, "标题错误")
			})
		})
	})
}
