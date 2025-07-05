package logic

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/bytedance/mockey"
	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/dao"
	"github.com/heimdall-api/common/model"
	"github.com/heimdall-api/public-api/public/internal/svc"
	"github.com/heimdall-api/public-api/public/internal/types"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGetPublicPageDetailLogic_GetPublicPageDetail(t *testing.T) {
	mockey.PatchConvey("GetPublicPageDetail", t, func() {
		// 准备测试数据
		pageID := primitive.NewObjectID()
		authorID := primitive.NewObjectID()
		now := time.Now()
		publishedAt := now.Add(-time.Hour)

		mockPage := &model.Page{
			ID:              pageID,
			Title:           "测试页面",
			Slug:            "test-page",
			Content:         "这是测试页面内容",
			HTML:            "<p>这是测试页面内容</p>",
			AuthorID:        authorID,
			Status:          constants.PostStatusPublished,
			Template:        "default",
			MetaTitle:       "测试页面 - SEO标题",
			MetaDescription: "测试页面的SEO描述",
			FeaturedImage:   "https://example.com/image.jpg",
			CanonicalURL:    "https://example.com/pages/test-page",
			PublishedAt:     &publishedAt,
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		mockAuthor := &model.User{
			ID:           authorID,
			Username:     "testuser",
			DisplayName:  "Test User",
			ProfileImage: "https://example.com/avatar.jpg",
			Bio:          "测试用户",
		}

		// 创建ServiceContext
		svcCtx := &svc.ServiceContext{
			PageDAO: &dao.PageDAO{},
			UserDAO: &dao.UserDAO{},
		}

		// 创建Logic实例
		logic := NewGetPublicPageDetailLogic(context.Background(), svcCtx)

		Convey("正常场景", func() {
			Convey("获取已发布页面详情应该成功", func() {
				// Mock PageDAO.GetBySlug
				mockey.Mock((*dao.PageDAO).GetBySlug).Return(mockPage, nil).Build()
				// Mock UserDAO.GetByID
				mockey.Mock((*dao.UserDAO).GetByID).Return(mockAuthor, nil).Build()

				req := &types.PublicPageDetailRequest{
					Slug: "test-page",
				}

				resp, err := logic.GetPublicPageDetail(req)

				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Code, ShouldEqual, 200)
				So(resp.Message, ShouldEqual, "success")
				So(resp.Data.Title, ShouldEqual, "测试页面")
				So(resp.Data.Slug, ShouldEqual, "test-page")
				So(resp.Data.HTML, ShouldEqual, "<p>这是测试页面内容</p>")
				So(resp.Data.Template, ShouldEqual, "default")
				So(resp.Data.Author.Username, ShouldEqual, "testuser")
				So(resp.Data.Author.DisplayName, ShouldEqual, "Test User")
				So(resp.Data.MetaTitle, ShouldEqual, "测试页面 - SEO标题")
				So(resp.Data.MetaDescription, ShouldEqual, "测试页面的SEO描述")
				So(resp.Data.FeaturedImage, ShouldEqual, "https://example.com/image.jpg")
				So(resp.Data.CanonicalURL, ShouldEqual, "/pages/test-page")
				So(resp.Data.PublishedAt, ShouldEqual, publishedAt.Format(time.RFC3339))
				So(resp.Data.UpdatedAt, ShouldEqual, now.Format(time.RFC3339))
			})

			Convey("获取页面详情时PublishedAt为nil应该返回空字符串", func() {
				mockPageNoPubTime := *mockPage
				mockPageNoPubTime.PublishedAt = nil

				mockey.Mock((*dao.PageDAO).GetBySlug).Return(&mockPageNoPubTime, nil).Build()
				mockey.Mock((*dao.UserDAO).GetByID).Return(mockAuthor, nil).Build()

				req := &types.PublicPageDetailRequest{
					Slug: "test-page",
				}

				resp, err := logic.GetPublicPageDetail(req)

				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Data.PublishedAt, ShouldEqual, "")
			})
		})

		Convey("异常场景", func() {
			Convey("slug为空应该返回错误", func() {
				req := &types.PublicPageDetailRequest{
					Slug: "",
				}

				resp, err := logic.GetPublicPageDetail(req)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "slug不能为空")
				So(resp, ShouldBeNil)
			})

			Convey("slug为空白字符应该返回错误", func() {
				req := &types.PublicPageDetailRequest{
					Slug: "   ",
				}

				resp, err := logic.GetPublicPageDetail(req)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "slug不能为空")
				So(resp, ShouldBeNil)
			})

			Convey("页面不存在应该返回错误", func() {
				mockey.Mock((*dao.PageDAO).GetBySlug).Return(nil, fmt.Errorf("页面不存在")).Build()

				req := &types.PublicPageDetailRequest{
					Slug: "non-existent-page",
				}

				resp, err := logic.GetPublicPageDetail(req)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "获取页面失败")
				So(resp, ShouldBeNil)
			})

			Convey("页面状态为草稿应该返回错误", func() {
				draftPage := *mockPage
				draftPage.Status = constants.PostStatusDraft

				mockey.Mock((*dao.PageDAO).GetBySlug).Return(&draftPage, nil).Build()

				req := &types.PublicPageDetailRequest{
					Slug: "test-page",
				}

				resp, err := logic.GetPublicPageDetail(req)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "页面未发布")
				So(resp, ShouldBeNil)
			})

			Convey("页面状态为定时发布应该返回错误", func() {
				scheduledPage := *mockPage
				scheduledPage.Status = constants.PostStatusScheduled

				mockey.Mock((*dao.PageDAO).GetBySlug).Return(&scheduledPage, nil).Build()

				req := &types.PublicPageDetailRequest{
					Slug: "test-page",
				}

				resp, err := logic.GetPublicPageDetail(req)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "页面未发布")
				So(resp, ShouldBeNil)
			})

			Convey("获取作者信息失败应该返回错误", func() {
				mockey.Mock((*dao.PageDAO).GetBySlug).Return(mockPage, nil).Build()
				mockey.Mock((*dao.UserDAO).GetByID).Return(nil, fmt.Errorf("用户不存在")).Build()

				req := &types.PublicPageDetailRequest{
					Slug: "test-page",
				}

				resp, err := logic.GetPublicPageDetail(req)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "获取作者信息失败")
				So(resp, ShouldBeNil)
			})
		})

		Convey("边界场景", func() {
			Convey("页面字段为空值应该正常处理", func() {
				emptyPage := &model.Page{
					ID:          pageID,
					Title:       "测试页面",
					Slug:        "test-page",
					Content:     "内容",
					HTML:        "<p>内容</p>",
					AuthorID:    authorID,
					Status:      constants.PostStatusPublished,
					Template:    "default",
					PublishedAt: &publishedAt,
					CreatedAt:   now,
					UpdatedAt:   now,
					// 其他字段为空
				}

				mockey.Mock((*dao.PageDAO).GetBySlug).Return(emptyPage, nil).Build()
				mockey.Mock((*dao.UserDAO).GetByID).Return(mockAuthor, nil).Build()

				req := &types.PublicPageDetailRequest{
					Slug: "test-page",
				}

				resp, err := logic.GetPublicPageDetail(req)

				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Data.MetaTitle, ShouldEqual, "")
				So(resp.Data.MetaDescription, ShouldEqual, "")
				So(resp.Data.FeaturedImage, ShouldEqual, "")
				So(resp.Data.CanonicalURL, ShouldEqual, "/pages/test-page")
			})

			Convey("作者字段为空值应该正常处理", func() {
				emptyAuthor := &model.User{
					ID:       authorID,
					Username: "testuser",
					// 其他字段为空
				}

				mockey.Mock((*dao.PageDAO).GetBySlug).Return(mockPage, nil).Build()
				mockey.Mock((*dao.UserDAO).GetByID).Return(emptyAuthor, nil).Build()

				req := &types.PublicPageDetailRequest{
					Slug: "test-page",
				}

				resp, err := logic.GetPublicPageDetail(req)

				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Data.Author.Username, ShouldEqual, "testuser")
				So(resp.Data.Author.DisplayName, ShouldEqual, "")
				So(resp.Data.Author.ProfileImage, ShouldEqual, "")
				So(resp.Data.Author.Bio, ShouldEqual, "")
			})
		})
	})
}

func TestGetPublicPageDetailLogic_validateRequest(t *testing.T) {
	mockey.PatchConvey("validateRequest", t, func() {
		logic := &GetPublicPageDetailLogic{}

		Convey("有效的slug应该通过验证", func() {
			req := &types.PublicPageDetailRequest{
				Slug: "valid-slug",
			}

			err := logic.validateRequest(req)

			So(err, ShouldBeNil)
		})

		Convey("空slug应该返回错误", func() {
			req := &types.PublicPageDetailRequest{
				Slug: "",
			}

			err := logic.validateRequest(req)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "slug不能为空")
		})

		Convey("空白slug应该返回错误", func() {
			req := &types.PublicPageDetailRequest{
				Slug: "   ",
			}

			err := logic.validateRequest(req)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "slug不能为空")
		})
	})
}

func TestGetPublicPageDetailLogic_validatePageVisibility(t *testing.T) {
	mockey.PatchConvey("validatePageVisibility", t, func() {
		logic := &GetPublicPageDetailLogic{}

		Convey("已发布页面应该通过验证", func() {
			page := &model.Page{
				Status: constants.PostStatusPublished,
			}

			err := logic.validatePageVisibility(page)

			So(err, ShouldBeNil)
		})

		Convey("草稿页面应该返回错误", func() {
			page := &model.Page{
				Status: constants.PostStatusDraft,
			}

			err := logic.validatePageVisibility(page)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "页面未发布")
		})

		Convey("定时发布页面应该返回错误", func() {
			page := &model.Page{
				Status: constants.PostStatusScheduled,
			}

			err := logic.validatePageVisibility(page)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "页面未发布")
		})
	})
}

func TestGetPublicPageDetailLogic_buildCanonicalURL(t *testing.T) {
	mockey.PatchConvey("buildCanonicalURL", t, func() {
		logic := &GetPublicPageDetailLogic{}

		Convey("应该正确构建canonical URL", func() {
			slug := "test-page"
			expected := "/pages/test-page"

			result := logic.buildCanonicalURL(slug)

			So(result, ShouldEqual, expected)
		})

		Convey("应该正确处理特殊字符的slug", func() {
			slug := "test-page-with-numbers-123"
			expected := "/pages/test-page-with-numbers-123"

			result := logic.buildCanonicalURL(slug)

			So(result, ShouldEqual, expected)
		})
	})
}

func TestGetPublicPageDetailLogic_buildAuthorInfo(t *testing.T) {
	mockey.PatchConvey("buildAuthorInfo", t, func() {
		logic := &GetPublicPageDetailLogic{}

		Convey("应该正确构建作者信息", func() {
			author := &model.User{
				Username:     "testuser",
				DisplayName:  "Test User",
				ProfileImage: "https://example.com/avatar.jpg",
				Bio:          "测试用户",
			}

			result := logic.buildAuthorInfo(author)

			So(result.Username, ShouldEqual, "testuser")
			So(result.DisplayName, ShouldEqual, "Test User")
			So(result.ProfileImage, ShouldEqual, "https://example.com/avatar.jpg")
			So(result.Bio, ShouldEqual, "测试用户")
		})

		Convey("应该正确处理空字段", func() {
			author := &model.User{
				Username: "testuser",
			}

			result := logic.buildAuthorInfo(author)

			So(result.Username, ShouldEqual, "testuser")
			So(result.DisplayName, ShouldEqual, "")
			So(result.ProfileImage, ShouldEqual, "")
			So(result.Bio, ShouldEqual, "")
		})
	})
}
