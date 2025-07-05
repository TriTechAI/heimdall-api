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

func TestGetPublicPostListLogic_GetPublicPostList(t *testing.T) {
	Convey("测试获取公开文章列表功能", t, func() {
		// 准备测试数据
		ctx := context.Background()
		svcCtx := &svc.ServiceContext{
			PostDAO: &dao.PostDAO{},
			UserDAO: &dao.UserDAO{},
		}
		logic := NewGetPublicPostListLogic(ctx, svcCtx)

		Convey("成功获取文章列表", func() {
			// 重置mock
			mockey.UnPatchAll()

			// 准备测试文章数据
			authorID := primitive.NewObjectID()
			now := time.Now()
			mockPosts := []*model.Post{
				{
					ID:            primitive.NewObjectID(),
					Title:         "测试文章1",
					Slug:          "test-post-1",
					Excerpt:       "这是第一篇测试文章",
					HTML:          "<p>这是第一篇测试文章的内容</p>",
					FeaturedImage: "https://example.com/image1.jpg",
					AuthorID:      authorID,
					Tags: []model.Tag{
						{Name: "技术", Slug: "tech"},
						{Name: "Go语言", Slug: "golang"},
					},
					ReadingTime: 5,
					ViewCount:   100,
					Status:      constants.PostStatusPublished,
					Visibility:  constants.PostVisibilityPublic,
					PublishedAt: &now,
					CreatedAt:   now.Add(-2 * time.Hour),
					UpdatedAt:   now.Add(-1 * time.Hour),
				},
				{
					ID:            primitive.NewObjectID(),
					Title:         "测试文章2",
					Slug:          "test-post-2",
					Excerpt:       "这是第二篇测试文章",
					HTML:          "<p>这是第二篇测试文章的内容</p>",
					FeaturedImage: "https://example.com/image2.jpg",
					AuthorID:      authorID,
					Tags: []model.Tag{
						{Name: "技术", Slug: "tech"},
					},
					ReadingTime: 3,
					ViewCount:   50,
					Status:      constants.PostStatusPublished,
					Visibility:  constants.PostVisibilityPublic,
					PublishedAt: &now,
					CreatedAt:   now.Add(-3 * time.Hour),
					UpdatedAt:   now.Add(-2 * time.Hour),
				},
			}

			// 准备用户数据
			mockUser := &model.User{
				ID:          authorID,
				Username:    "testuser",
				DisplayName: "Test User",
				Bio:         "This is a test user",
			}

			// Mock PostDAO.GetPublishedList
			mockey.Mock((*dao.PostDAO).GetPublishedList).To(func(postDAO *dao.PostDAO, ctx context.Context, filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
				So(filter.Status, ShouldEqual, constants.PostStatusPublished)
				So(filter.Visibility, ShouldEqual, constants.PostVisibilityPublic)
				So(page, ShouldEqual, 1)
				So(limit, ShouldEqual, 10)
				return mockPosts, 2, nil
			}).Build()

			// Mock UserDAO.GetByID
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				So(id, ShouldEqual, authorID.Hex())
				return mockUser, nil
			}).Build()

			// 准备请求
			req := &types.PublicPostListRequest{
				Page:     1,
				Limit:    10,
				SortBy:   "publishedAt",
				SortDesc: true,
			}

			// 执行测试
			resp, err := logic.GetPublicPostList(req)

			// 验证结果
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Message, ShouldEqual, "success")
			So(resp.Data, ShouldNotBeNil)
			So(resp.Data.List, ShouldHaveLength, 2)
			So(resp.Data.Pagination.Total, ShouldEqual, 2)
			So(resp.Data.Pagination.Page, ShouldEqual, 1)
			So(resp.Data.Pagination.Limit, ShouldEqual, 10)

			// 验证文章数据
			firstPost := resp.Data.List[0]
			So(firstPost.Title, ShouldEqual, "测试文章1")
			So(firstPost.Slug, ShouldEqual, "test-post-1")
			So(firstPost.Author.Username, ShouldEqual, "testuser")
			So(firstPost.Tags, ShouldHaveLength, 2)
			So(firstPost.ViewCount, ShouldEqual, 100)
		})

		Convey("带标签过滤的文章列表", func() {
			// 重置mock
			mockey.UnPatchAll()

			authorID := primitive.NewObjectID()
			now := time.Now()
			mockPosts := []*model.Post{
				{
					ID:       primitive.NewObjectID(),
					Title:    "Go语言文章",
					Slug:     "golang-post",
					Excerpt:  "关于Go语言的文章",
					AuthorID: authorID,
					Tags: []model.Tag{
						{Name: "Go语言", Slug: "golang"},
					},
					Status:      constants.PostStatusPublished,
					Visibility:  constants.PostVisibilityPublic,
					PublishedAt: &now,
					CreatedAt:   now,
					UpdatedAt:   now,
				},
			}

			mockUser := &model.User{
				ID:          authorID,
				Username:    "testuser",
				DisplayName: "Test User",
			}

			// Mock PostDAO.GetPublishedList with tag filter
			mockey.Mock((*dao.PostDAO).GetPublishedList).To(func(postDAO *dao.PostDAO, ctx context.Context, filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
				So(filter.Tag, ShouldEqual, "golang")
				return mockPosts, 1, nil
			}).Build()

			// Mock UserDAO.GetByID
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return mockUser, nil
			}).Build()

			// 准备请求
			req := &types.PublicPostListRequest{
				Page:   1,
				Limit:  10,
				Tag:    "golang",
				SortBy: "publishedAt",
			}

			// 执行测试
			resp, err := logic.GetPublicPostList(req)

			// 验证结果
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Data.List, ShouldHaveLength, 1)
			So(resp.Data.List[0].Title, ShouldEqual, "Go语言文章")
		})

		Convey("带作者过滤的文章列表", func() {
			// 重置mock
			mockey.UnPatchAll()

			authorID := primitive.NewObjectID()
			now := time.Now()
			mockPosts := []*model.Post{
				{
					ID:          primitive.NewObjectID(),
					Title:       "作者文章",
					Slug:        "author-post",
					Excerpt:     "特定作者的文章",
					AuthorID:    authorID,
					Status:      constants.PostStatusPublished,
					Visibility:  constants.PostVisibilityPublic,
					PublishedAt: &now,
					CreatedAt:   now,
					UpdatedAt:   now,
				},
			}

			mockUser := &model.User{
				ID:          authorID,
				Username:    "testauthor",
				DisplayName: "Test Author",
			}

			// Mock UserDAO.GetByUsername to get author ID
			mockey.Mock((*dao.UserDAO).GetByUsername).To(func(userDAO *dao.UserDAO, ctx context.Context, username string) (*model.User, error) {
				So(username, ShouldEqual, "testauthor")
				return mockUser, nil
			}).Build()

			// Mock PostDAO.GetPublishedList with author filter
			mockey.Mock((*dao.PostDAO).GetPublishedList).To(func(postDAO *dao.PostDAO, ctx context.Context, filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
				So(filter.AuthorID, ShouldEqual, authorID.Hex())
				return mockPosts, 1, nil
			}).Build()

			// Mock UserDAO.GetByID
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return mockUser, nil
			}).Build()

			// 准备请求
			req := &types.PublicPostListRequest{
				Page:   1,
				Limit:  10,
				Author: "testauthor",
				SortBy: "publishedAt",
			}

			// 执行测试
			resp, err := logic.GetPublicPostList(req)

			// 验证结果
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Data.List, ShouldHaveLength, 1)
			So(resp.Data.List[0].Author.Username, ShouldEqual, "testauthor")
		})

		Convey("处理空结果", func() {
			// 重置mock
			mockey.UnPatchAll()

			// Mock PostDAO.GetPublishedList - 返回空结果
			mockey.Mock((*dao.PostDAO).GetPublishedList).To(func(postDAO *dao.PostDAO, ctx context.Context, filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
				return []*model.Post{}, 0, nil
			}).Build()

			// 准备请求
			req := &types.PublicPostListRequest{
				Page:   1,
				Limit:  10,
				SortBy: "publishedAt",
			}

			// 执行测试
			resp, err := logic.GetPublicPostList(req)

			// 验证结果
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Data.List, ShouldHaveLength, 0)
			So(resp.Data.Pagination.Total, ShouldEqual, 0)
		})

		Convey("处理数据库查询错误", func() {
			// 重置mock
			mockey.UnPatchAll()

			// Mock PostDAO.GetPublishedList - 返回错误
			mockey.Mock((*dao.PostDAO).GetPublishedList).To(func(postDAO *dao.PostDAO, ctx context.Context, filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
				return nil, 0, errors.New("数据库连接失败")
			}).Build()

			// 准备请求
			req := &types.PublicPostListRequest{
				Page:   1,
				Limit:  10,
				SortBy: "publishedAt",
			}

			// 执行测试
			resp, err := logic.GetPublicPostList(req)

			// 验证结果
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "获取文章列表失败")
		})

		Convey("处理无效的作者用户名", func() {
			// 重置mock
			mockey.UnPatchAll()

			// Mock UserDAO.GetByUsername - 返回用户不存在
			mockey.Mock((*dao.UserDAO).GetByUsername).To(func(userDAO *dao.UserDAO, ctx context.Context, username string) (*model.User, error) {
				return nil, errors.New("用户不存在")
			}).Build()

			// 准备请求
			req := &types.PublicPostListRequest{
				Page:   1,
				Limit:  10,
				Author: "nonexistentuser",
				SortBy: "publishedAt",
			}

			// 执行测试
			resp, err := logic.GetPublicPostList(req)

			// 验证结果
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "作者不存在")
		})

		Convey("处理用户信息获取失败", func() {
			// 重置mock
			mockey.UnPatchAll()

			authorID := primitive.NewObjectID()
			now := time.Now()
			mockPosts := []*model.Post{
				{
					ID:          primitive.NewObjectID(),
					Title:       "测试文章",
					Slug:        "test-post",
					AuthorID:    authorID,
					Status:      constants.PostStatusPublished,
					Visibility:  constants.PostVisibilityPublic,
					PublishedAt: &now,
					CreatedAt:   now,
					UpdatedAt:   now,
				},
			}

			// Mock PostDAO.GetPublishedList
			mockey.Mock((*dao.PostDAO).GetPublishedList).To(func(postDAO *dao.PostDAO, ctx context.Context, filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
				return mockPosts, 1, nil
			}).Build()

			// Mock UserDAO.GetByID - 返回错误
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return nil, errors.New("用户查询失败")
			}).Build()

			// 准备请求
			req := &types.PublicPostListRequest{
				Page:   1,
				Limit:  10,
				SortBy: "publishedAt",
			}

			// 执行测试
			resp, err := logic.GetPublicPostList(req)

			// 验证结果
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "获取作者信息失败")
		})
	})
}
