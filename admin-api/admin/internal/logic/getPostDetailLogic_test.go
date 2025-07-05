package logic

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bytedance/mockey"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/heimdall-api/common/dao"
	"github.com/heimdall-api/common/model"
)

func TestGetPostDetailLogic_GetPostDetail(t *testing.T) {
	Convey("测试文章详情查询功能", t, func() {
		// 准备测试数据
		ctx := context.Background()
		svcCtx := &svc.ServiceContext{
			PostDAO: &dao.PostDAO{},
			UserDAO: &dao.UserDAO{},
		}
		logic := NewGetPostDetailLogic(ctx, svcCtx)

		Convey("成功获取文章详情", func() {
			// 重置mock
			mockey.UnPatchAll()

			postID := primitive.NewObjectID()
			authorID := primitive.NewObjectID()

			req := &types.PostDetailRequest{
				ID: postID.Hex(),
			}

			// 准备模拟数据
			mockPost := &model.Post{
				ID:              postID,
				Title:           "测试文章标题",
				Slug:            "test-post-slug",
				Excerpt:         "这是文章摘要",
				Markdown:        "# 标题\n\n这是文章内容",
				HTML:            "<h1>标题</h1><p>这是文章内容</p>",
				FeaturedImage:   "https://example.com/image.jpg",
				Type:            "post",
				Status:          "published",
				Visibility:      "public",
				AuthorID:        authorID,
				Tags:            []model.Tag{{Name: "Go", Slug: "go"}},
				MetaTitle:       "SEO标题",
				MetaDescription: "SEO描述",
				CanonicalURL:    "https://example.com/canonical",
				ReadingTime:     5,
				WordCount:       1000,
				ViewCount:       150,
				PublishedAt:     &time.Time{},
				CreatedAt:       time.Now().Add(-2 * time.Hour),
				UpdatedAt:       time.Now().Add(-1 * time.Hour),
			}

			mockUser := &model.User{
				ID:          authorID,
				Username:    "testuser",
				DisplayName: "Test User",
				Email:       "test@example.com",
				Bio:         "这是作者简介",
			}

			// Mock PostDAO.GetByID
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				So(id, ShouldEqual, postID.Hex())
				return mockPost, nil
			}).Build()

			// Mock UserDAO.GetByID (获取作者信息)
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				So(id, ShouldEqual, authorID.Hex())
				return mockUser, nil
			}).Build()

			// 执行测试
			resp, err := logic.GetPostDetail(req)

			// 验证结果
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Message, ShouldEqual, "获取文章详情成功")
			So(resp.Data.ID, ShouldEqual, postID.Hex())
			So(resp.Data.Title, ShouldEqual, "测试文章标题")
			So(resp.Data.Slug, ShouldEqual, "test-post-slug")
			So(resp.Data.Markdown, ShouldEqual, "# 标题\n\n这是文章内容")
			So(resp.Data.HTML, ShouldEqual, "<h1>标题</h1><p>这是文章内容</p>")
			So(resp.Data.Author.Username, ShouldEqual, "testuser")
			So(resp.Data.Author.DisplayName, ShouldEqual, "Test User")
			So(len(resp.Data.Tags), ShouldEqual, 1)
			So(resp.Data.Tags[0].Name, ShouldEqual, "Go")
		})

		Convey("处理无效的文章ID", func() {
			// 重置mock
			mockey.UnPatchAll()

			req := &types.PostDetailRequest{
				ID: "invalid-id",
			}

			resp, err := logic.GetPostDetail(req)

			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "无效的文章ID")
		})

		Convey("处理文章不存在", func() {
			// 重置mock
			mockey.UnPatchAll()

			postID := primitive.NewObjectID()
			req := &types.PostDetailRequest{
				ID: postID.Hex(),
			}

			// Mock PostDAO.GetByID - 返回文章不存在错误
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				return nil, errors.New("post not found")
			}).Build()

			resp, err := logic.GetPostDetail(req)

			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "文章不存在")
		})

		Convey("处理作者信息获取失败", func() {
			// 重置mock
			mockey.UnPatchAll()

			postID := primitive.NewObjectID()
			authorID := primitive.NewObjectID()

			req := &types.PostDetailRequest{
				ID: postID.Hex(),
			}

			mockPost := &model.Post{
				ID:        postID,
				Title:     "测试文章",
				AuthorID:  authorID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Mock PostDAO.GetByID - 成功返回文章
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				return mockPost, nil
			}).Build()

			// Mock UserDAO.GetByID - 返回错误
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return nil, errors.New("user not found")
			}).Build()

			resp, err := logic.GetPostDetail(req)

			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "获取作者信息失败")
		})

		Convey("处理数据库查询错误", func() {
			// 重置mock
			mockey.UnPatchAll()

			postID := primitive.NewObjectID()
			req := &types.PostDetailRequest{
				ID: postID.Hex(),
			}

			// Mock PostDAO.GetByID - 返回数据库错误
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				return nil, errors.New("database connection error")
			}).Build()

			resp, err := logic.GetPostDetail(req)

			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "文章不存在")
		})

		Convey("验证完整的响应数据结构", func() {
			// 重置mock
			mockey.UnPatchAll()

			postID := primitive.NewObjectID()
			authorID := primitive.NewObjectID()
			publishedAt := time.Now().Add(-1 * time.Hour)

			req := &types.PostDetailRequest{
				ID: postID.Hex(),
			}

			mockPost := &model.Post{
				ID:              postID,
				Title:           "完整测试文章",
				Slug:            "complete-test-post",
				Excerpt:         "完整测试摘要",
				Markdown:        "# 完整测试\n\n这是完整的文章内容",
				HTML:            "<h1>完整测试</h1><p>这是完整的文章内容</p>",
				FeaturedImage:   "https://example.com/featured.jpg",
				Type:            "post",
				Status:          "published",
				Visibility:      "public",
				AuthorID:        authorID,
				Tags:            []model.Tag{{Name: "Test", Slug: "test"}, {Name: "Go", Slug: "go"}},
				MetaTitle:       "完整SEO标题",
				MetaDescription: "完整SEO描述",
				CanonicalURL:    "https://example.com/complete-canonical",
				ReadingTime:     8,
				WordCount:       1500,
				ViewCount:       300,
				PublishedAt:     &publishedAt,
				CreatedAt:       time.Now().Add(-3 * time.Hour),
				UpdatedAt:       time.Now().Add(-30 * time.Minute),
			}

			mockUser := &model.User{
				ID:           authorID,
				Username:     "completeuser",
				DisplayName:  "Complete User",
				Email:        "complete@example.com",
				Bio:          "完整的作者简介",
				ProfileImage: "https://example.com/avatar.jpg",
			}

			// Mock PostDAO.GetByID
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				return mockPost, nil
			}).Build()

			// Mock UserDAO.GetByID
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return mockUser, nil
			}).Build()

			// 执行测试
			resp, err := logic.GetPostDetail(req)

			// 验证完整响应结构
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Message, ShouldEqual, "获取文章详情成功")
			So(resp.Timestamp, ShouldNotBeEmpty)

			// 验证文章数据
			data := resp.Data
			So(data.ID, ShouldEqual, postID.Hex())
			So(data.Title, ShouldEqual, "完整测试文章")
			So(data.Slug, ShouldEqual, "complete-test-post")
			So(data.Excerpt, ShouldEqual, "完整测试摘要")
			So(data.Markdown, ShouldEqual, "# 完整测试\n\n这是完整的文章内容")
			So(data.HTML, ShouldEqual, "<h1>完整测试</h1><p>这是完整的文章内容</p>")
			So(data.FeaturedImage, ShouldEqual, "https://example.com/featured.jpg")
			So(data.Type, ShouldEqual, "post")
			So(data.Status, ShouldEqual, "published")
			So(data.Visibility, ShouldEqual, "public")
			So(data.MetaTitle, ShouldEqual, "完整SEO标题")
			So(data.MetaDescription, ShouldEqual, "完整SEO描述")
			So(data.CanonicalURL, ShouldEqual, "https://example.com/complete-canonical")
			So(data.ReadingTime, ShouldEqual, 8)
			So(data.WordCount, ShouldEqual, 1500)
			So(data.ViewCount, ShouldEqual, 300)
			So(data.PublishedAt, ShouldNotBeEmpty)
			So(data.CreatedAt, ShouldNotBeEmpty)
			So(data.UpdatedAt, ShouldNotBeEmpty)

			// 验证作者信息
			author := data.Author
			So(author.ID, ShouldEqual, authorID.Hex())
			So(author.Username, ShouldEqual, "completeuser")
			So(author.DisplayName, ShouldEqual, "Complete User")
			So(author.Bio, ShouldEqual, "完整的作者简介")
			So(author.ProfileImage, ShouldEqual, "https://example.com/avatar.jpg")

			// 验证标签信息
			So(len(data.Tags), ShouldEqual, 2)
			So(data.Tags[0].Name, ShouldEqual, "Test")
			So(data.Tags[0].Slug, ShouldEqual, "test")
			So(data.Tags[1].Name, ShouldEqual, "Go")
			So(data.Tags[1].Slug, ShouldEqual, "go")
		})

		Reset(func() {
			mockey.UnPatchAll()
		})
	})
}
