package logic

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bytedance/mockey"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/dao"
	"github.com/heimdall-api/common/model"
	"github.com/heimdall-api/public-api/public/internal/svc"
	"github.com/heimdall-api/public-api/public/internal/types"
)

func TestGetPublicPostDetailLogic_GetPublicPostDetail(t *testing.T) {
	Convey("测试获取公开文章详情功能", t, func() {
		// 准备测试数据
		ctx := context.Background()
		svcCtx := &svc.ServiceContext{
			PostDAO: &dao.PostDAO{},
			UserDAO: &dao.UserDAO{},
		}
		logic := NewGetPublicPostDetailLogic(ctx, svcCtx)

		Convey("成功获取文章详情", func() {
			// 重置mock
			mockey.UnPatchAll()

			// 准备测试数据
			authorID := primitive.NewObjectID()
			postID := primitive.NewObjectID()
			now := time.Now()
			mockPost := &model.Post{
				ID:            postID,
				Title:         "测试文章标题",
				Slug:          "test-post-slug",
				Excerpt:       "这是一篇测试文章的摘要",
				HTML:          "<p>这是文章的完整HTML内容</p>",
				Markdown:      "这是文章的Markdown内容",
				FeaturedImage: "https://example.com/featured-image.jpg",
				AuthorID:      authorID,
				Tags: []model.Tag{
					{Name: "技术", Slug: "tech"},
					{Name: "Go语言", Slug: "golang"},
				},
				MetaTitle:       "SEO标题",
				MetaDescription: "SEO描述",
				ReadingTime:     5,
				ViewCount:       100,
				Status:          constants.PostStatusPublished,
				Visibility:      constants.PostVisibilityPublic,
				PublishedAt:     &now,
				CreatedAt:       now.Add(-2 * time.Hour),
				UpdatedAt:       now.Add(-1 * time.Hour),
			}

			mockUser := &model.User{
				ID:           authorID,
				Username:     "testuser",
				DisplayName:  "Test User",
				Bio:          "This is a test user bio",
				ProfileImage: "https://example.com/avatar.jpg",
				Website:      "https://testuser.com",
				Location:     "Test City",
			}

			// Mock PostDAO.GetBySlug
			mockey.Mock((*dao.PostDAO).GetBySlug).To(func(postDAO *dao.PostDAO, ctx context.Context, slug string) (*model.Post, error) {
				So(slug, ShouldEqual, "test-post-slug")
				return mockPost, nil
			}).Build()

			// Mock UserDAO.GetByID
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				So(id, ShouldEqual, authorID.Hex())
				return mockUser, nil
			}).Build()

			// Mock PostDAO.IncrementViewCount
			mockey.Mock((*dao.PostDAO).IncrementViewCount).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) error {
				So(id, ShouldEqual, postID.Hex())
				return nil
			}).Build()

			// 准备请求
			req := &types.PublicPostDetailRequest{
				Slug: "test-post-slug",
			}

			// 执行测试
			resp, err := logic.GetPublicPostDetail(req)

			// 验证结果
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Message, ShouldEqual, "success")
			So(resp.Data, ShouldNotBeNil)

			// 验证文章数据
			post := resp.Data
			So(post.Title, ShouldEqual, "测试文章标题")
			So(post.Slug, ShouldEqual, "test-post-slug")
			So(post.Excerpt, ShouldEqual, "这是一篇测试文章的摘要")
			So(post.HTML, ShouldEqual, "<p>这是文章的完整HTML内容</p>")
			So(post.FeaturedImage, ShouldEqual, "https://example.com/featured-image.jpg")
			So(post.MetaTitle, ShouldEqual, "SEO标题")
			So(post.MetaDescription, ShouldEqual, "SEO描述")
			So(post.ReadingTime, ShouldEqual, 5)
			So(post.ViewCount, ShouldEqual, 100)
			So(post.Tags, ShouldHaveLength, 2)
			So(post.Tags[0].Name, ShouldEqual, "技术")
			So(post.Tags[1].Name, ShouldEqual, "Go语言")

			// 验证作者信息
			So(post.Author.Username, ShouldEqual, "testuser")
			So(post.Author.DisplayName, ShouldEqual, "Test User")
			So(post.Author.Bio, ShouldEqual, "This is a test user bio")
			So(post.Author.ProfileImage, ShouldEqual, "https://example.com/avatar.jpg")

			// 验证时间字段
			So(post.PublishedAt, ShouldNotBeNil)
			So(post.UpdatedAt, ShouldNotBeNil)
		})

		Convey("文章不存在", func() {
			// 重置mock
			mockey.UnPatchAll()

			// Mock PostDAO.GetBySlug - 返回文章不存在
			mockey.Mock((*dao.PostDAO).GetBySlug).To(func(postDAO *dao.PostDAO, ctx context.Context, slug string) (*model.Post, error) {
				So(slug, ShouldEqual, "nonexistent-post")
				return nil, errors.New("文章不存在")
			}).Build()

			// 准备请求
			req := &types.PublicPostDetailRequest{
				Slug: "nonexistent-post",
			}

			// 执行测试
			resp, err := logic.GetPublicPostDetail(req)

			// 验证结果
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "文章不存在")
		})

		Convey("文章未发布", func() {
			// 重置mock
			mockey.UnPatchAll()

			// 准备草稿文章
			authorID := primitive.NewObjectID()
			mockPost := &model.Post{
				ID:         primitive.NewObjectID(),
				Title:      "草稿文章",
				Slug:       "draft-post",
				AuthorID:   authorID,
				Status:     constants.PostStatusDraft,
				Visibility: constants.PostVisibilityPublic,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			// Mock PostDAO.GetBySlug
			mockey.Mock((*dao.PostDAO).GetBySlug).To(func(postDAO *dao.PostDAO, ctx context.Context, slug string) (*model.Post, error) {
				return mockPost, nil
			}).Build()

			// 准备请求
			req := &types.PublicPostDetailRequest{
				Slug: "draft-post",
			}

			// 执行测试
			resp, err := logic.GetPublicPostDetail(req)

			// 验证结果
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "文章未发布")
		})

		Convey("文章不可见", func() {
			// 重置mock
			mockey.UnPatchAll()

			// 准备私有文章
			authorID := primitive.NewObjectID()
			now := time.Now()
			mockPost := &model.Post{
				ID:          primitive.NewObjectID(),
				Title:       "私有文章",
				Slug:        "private-post",
				AuthorID:    authorID,
				Status:      constants.PostStatusPublished,
				Visibility:  constants.PostVisibilityPrivate,
				PublishedAt: &now,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			// Mock PostDAO.GetBySlug
			mockey.Mock((*dao.PostDAO).GetBySlug).To(func(postDAO *dao.PostDAO, ctx context.Context, slug string) (*model.Post, error) {
				return mockPost, nil
			}).Build()

			// 准备请求
			req := &types.PublicPostDetailRequest{
				Slug: "private-post",
			}

			// 执行测试
			resp, err := logic.GetPublicPostDetail(req)

			// 验证结果
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "文章不可见")
		})

		Convey("作者信息获取失败", func() {
			// 重置mock
			mockey.UnPatchAll()

			// 准备文章数据
			authorID := primitive.NewObjectID()
			postID := primitive.NewObjectID()
			now := time.Now()
			mockPost := &model.Post{
				ID:          postID,
				Title:       "测试文章",
				Slug:        "test-post",
				AuthorID:    authorID,
				Status:      constants.PostStatusPublished,
				Visibility:  constants.PostVisibilityPublic,
				PublishedAt: &now,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			// Mock PostDAO.GetBySlug
			mockey.Mock((*dao.PostDAO).GetBySlug).To(func(postDAO *dao.PostDAO, ctx context.Context, slug string) (*model.Post, error) {
				return mockPost, nil
			}).Build()

			// Mock UserDAO.GetByID - 返回错误
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return nil, errors.New("用户查询失败")
			}).Build()

			// 准备请求
			req := &types.PublicPostDetailRequest{
				Slug: "test-post",
			}

			// 执行测试
			resp, err := logic.GetPublicPostDetail(req)

			// 验证结果
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "获取作者信息失败")
		})

		Convey("浏览计数更新失败", func() {
			// 重置mock
			mockey.UnPatchAll()

			// 准备测试数据
			authorID := primitive.NewObjectID()
			postID := primitive.NewObjectID()
			now := time.Now()
			mockPost := &model.Post{
				ID:          postID,
				Title:       "测试文章",
				Slug:        "test-post",
				AuthorID:    authorID,
				Status:      constants.PostStatusPublished,
				Visibility:  constants.PostVisibilityPublic,
				PublishedAt: &now,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			mockUser := &model.User{
				ID:          authorID,
				Username:    "testuser",
				DisplayName: "Test User",
			}

			// Mock PostDAO.GetBySlug
			mockey.Mock((*dao.PostDAO).GetBySlug).To(func(postDAO *dao.PostDAO, ctx context.Context, slug string) (*model.Post, error) {
				return mockPost, nil
			}).Build()

			// Mock UserDAO.GetByID
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return mockUser, nil
			}).Build()

			// Mock PostDAO.IncrementViewCount - 返回错误
			mockey.Mock((*dao.PostDAO).IncrementViewCount).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) error {
				return errors.New("浏览计数更新失败")
			}).Build()

			// 准备请求
			req := &types.PublicPostDetailRequest{
				Slug: "test-post",
			}

			// 执行测试
			resp, err := logic.GetPublicPostDetail(req)

			// 验证结果 - 浏览计数更新失败不应该影响文章详情获取
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Data.Title, ShouldEqual, "测试文章")
		})

		Convey("处理空slug", func() {
			// 重置mock
			mockey.UnPatchAll()

			// 准备请求
			req := &types.PublicPostDetailRequest{
				Slug: "",
			}

			// 执行测试
			resp, err := logic.GetPublicPostDetail(req)

			// 验证结果
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "slug不能为空")
		})

		Convey("处理数据库查询错误", func() {
			// 重置mock
			mockey.UnPatchAll()

			// Mock PostDAO.GetBySlug - 返回数据库错误
			mockey.Mock((*dao.PostDAO).GetBySlug).To(func(postDAO *dao.PostDAO, ctx context.Context, slug string) (*model.Post, error) {
				return nil, errors.New("数据库连接失败")
			}).Build()

			// 准备请求
			req := &types.PublicPostDetailRequest{
				Slug: "test-post",
			}

			// 执行测试
			resp, err := logic.GetPublicPostDetail(req)

			// 验证结果
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "获取文章失败")
		})
	})
}
