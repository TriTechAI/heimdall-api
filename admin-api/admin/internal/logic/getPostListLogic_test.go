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

func TestGetPostListLogic_GetPostList(t *testing.T) {
	Convey("测试文章列表查询功能", t, func() {
		// 准备测试数据
		ctx := context.Background()
		svcCtx := &svc.ServiceContext{
			PostDAO: &dao.PostDAO{},
			UserDAO: &dao.UserDAO{},
		}
		logic := NewGetPostListLogic(ctx, svcCtx)

		Convey("成功获取文章列表", func() {
			// 重置mock
			mockey.UnPatchAll()

			req := &types.PostListRequest{
				Page:     1,
				Limit:    10,
				SortBy:   "updatedAt",
				SortDesc: true,
			}

			// 准备模拟数据
			authorID := primitive.NewObjectID()
			mockPosts := []*model.Post{
				{
					ID:            primitive.NewObjectID(),
					Title:         "测试文章1",
					Slug:          "test-post-1",
					Excerpt:       "这是测试文章1的摘要",
					FeaturedImage: "https://example.com/image1.jpg",
					Type:          "post",
					Status:        "published",
					Visibility:    "public",
					AuthorID:      authorID,
					Tags:          []model.Tag{{Name: "Go", Slug: "go"}},
					ReadingTime:   5,
					ViewCount:     100,
					CreatedAt:     time.Now().Add(-2 * time.Hour),
					UpdatedAt:     time.Now().Add(-1 * time.Hour),
				},
				{
					ID:            primitive.NewObjectID(),
					Title:         "测试文章2",
					Slug:          "test-post-2",
					Excerpt:       "这是测试文章2的摘要",
					FeaturedImage: "https://example.com/image2.jpg",
					Type:          "post",
					Status:        "draft",
					Visibility:    "public",
					AuthorID:      authorID,
					Tags:          []model.Tag{{Name: "Test", Slug: "test"}},
					ReadingTime:   3,
					ViewCount:     50,
					CreatedAt:     time.Now().Add(-3 * time.Hour),
					UpdatedAt:     time.Now().Add(-30 * time.Minute),
				},
			}

			mockUser := &model.User{
				ID:          authorID,
				Username:    "testuser",
				DisplayName: "Test User",
				Email:       "test@example.com",
			}

			// Mock PostDAO.List
			mockey.Mock((*dao.PostDAO).List).To(func(postDAO *dao.PostDAO, ctx context.Context, filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
				return mockPosts, 2, nil
			}).Build()

			// Mock UserDAO.GetByID (获取作者信息)
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return mockUser, nil
			}).Build()

			// 执行测试
			resp, err := logic.GetPostList(req)

			// 验证结果
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Message, ShouldEqual, "获取文章列表成功")
			So(len(resp.Data.List), ShouldEqual, 2)
			So(resp.Data.Pagination.Page, ShouldEqual, 1)
			So(resp.Data.Pagination.Limit, ShouldEqual, 10)
			So(resp.Data.Pagination.Total, ShouldEqual, 2)
			So(resp.Data.List[0].Title, ShouldEqual, "测试文章1")
			So(resp.Data.List[0].Author.Username, ShouldEqual, "testuser")
		})

		Convey("支持状态过滤", func() {
			// 重置mock
			mockey.UnPatchAll()

			req := &types.PostListRequest{
				Page:     1,
				Limit:    10,
				Status:   "published",
				SortBy:   "updatedAt",
				SortDesc: true,
			}

			mockPosts := []*model.Post{
				{
					ID:        primitive.NewObjectID(),
					Title:     "已发布文章",
					Status:    "published",
					AuthorID:  primitive.NewObjectID(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			}

			mockUser := &model.User{
				ID:          primitive.NewObjectID(),
				Username:    "testuser",
				DisplayName: "Test User",
			}

			// Mock PostDAO.List - 验证过滤条件
			mockey.Mock((*dao.PostDAO).List).To(func(postDAO *dao.PostDAO, ctx context.Context, filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
				// 验证过滤条件
				So(filter.Status, ShouldEqual, "published")
				return mockPosts, 1, nil
			}).Build()

			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return mockUser, nil
			}).Build()

			resp, err := logic.GetPostList(req)

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp.Data.List), ShouldEqual, 1)
			So(resp.Data.List[0].Status, ShouldEqual, "published")
		})

		Convey("支持关键词搜索", func() {
			// 重置mock
			mockey.UnPatchAll()

			req := &types.PostListRequest{
				Page:     1,
				Limit:    10,
				Keyword:  "测试",
				SortBy:   "updatedAt",
				SortDesc: true,
			}

			mockPosts := []*model.Post{
				{
					ID:        primitive.NewObjectID(),
					Title:     "测试文章标题",
					Excerpt:   "包含测试关键词的摘要",
					AuthorID:  primitive.NewObjectID(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			}

			mockUser := &model.User{
				ID:          primitive.NewObjectID(),
				Username:    "testuser",
				DisplayName: "Test User",
			}

			// Mock PostDAO.List - 验证关键词过滤
			mockey.Mock((*dao.PostDAO).List).To(func(postDAO *dao.PostDAO, ctx context.Context, filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
				So(filter.Keyword, ShouldEqual, "测试")
				return mockPosts, 1, nil
			}).Build()

			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return mockUser, nil
			}).Build()

			resp, err := logic.GetPostList(req)

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp.Data.List), ShouldEqual, 1)
		})

		Convey("支持多维度过滤", func() {
			// 重置mock
			mockey.UnPatchAll()

			req := &types.PostListRequest{
				Page:       1,
				Limit:      10,
				Status:     "published",
				Type:       "post",
				Visibility: "public",
				AuthorID:   "507f1f77bcf86cd799439011",
				Tag:        "go",
				SortBy:     "viewCount",
				SortDesc:   true,
			}

			mockPosts := []*model.Post{
				{
					ID:         primitive.NewObjectID(),
					Title:      "Go语言文章",
					Status:     "published",
					Type:       "post",
					Visibility: "public",
					AuthorID:   primitive.NewObjectID(),
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				},
			}

			mockUser := &model.User{
				ID:          primitive.NewObjectID(),
				Username:    "testuser",
				DisplayName: "Test User",
			}

			// Mock PostDAO.List - 验证所有过滤条件
			mockey.Mock((*dao.PostDAO).List).To(func(postDAO *dao.PostDAO, ctx context.Context, filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
				So(filter.Status, ShouldEqual, "published")
				So(filter.Type, ShouldEqual, "post")
				So(filter.Visibility, ShouldEqual, "public")
				So(filter.AuthorID, ShouldEqual, "507f1f77bcf86cd799439011")
				So(filter.Tag, ShouldEqual, "go")
				So(filter.SortBy, ShouldEqual, "viewCount")
				So(filter.SortDesc, ShouldBeTrue)
				return mockPosts, 1, nil
			}).Build()

			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return mockUser, nil
			}).Build()

			resp, err := logic.GetPostList(req)

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp.Data.List), ShouldEqual, 1)
		})

		Convey("支持分页计算", func() {
			// 重置mock
			mockey.UnPatchAll()

			req := &types.PostListRequest{
				Page:     2,
				Limit:    5,
				SortBy:   "createdAt",
				SortDesc: false,
			}

			mockPosts := []*model.Post{
				{
					ID:        primitive.NewObjectID(),
					Title:     "第二页文章",
					AuthorID:  primitive.NewObjectID(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			}

			mockUser := &model.User{
				ID:          primitive.NewObjectID(),
				Username:    "testuser",
				DisplayName: "Test User",
			}

			// Mock PostDAO.List - 验证分页参数
			mockey.Mock((*dao.PostDAO).List).To(func(postDAO *dao.PostDAO, ctx context.Context, filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
				So(page, ShouldEqual, 2)
				So(limit, ShouldEqual, 5)
				So(filter.SortBy, ShouldEqual, "createdAt")
				So(filter.SortDesc, ShouldBeFalse)
				return mockPosts, 12, nil // 总共12条记录
			}).Build()

			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return mockUser, nil
			}).Build()

			resp, err := logic.GetPostList(req)

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Data.Pagination.Page, ShouldEqual, 2)
			So(resp.Data.Pagination.Limit, ShouldEqual, 5)
			So(resp.Data.Pagination.Total, ShouldEqual, 12)
			So(resp.Data.Pagination.TotalPages, ShouldEqual, 3)
			So(resp.Data.Pagination.HasPrev, ShouldBeTrue)
			So(resp.Data.Pagination.HasNext, ShouldBeTrue)
		})

		Convey("处理空结果", func() {
			// 重置mock
			mockey.UnPatchAll()

			req := &types.PostListRequest{
				Page:     1,
				Limit:    10,
				Status:   "archived",
				SortBy:   "updatedAt",
				SortDesc: true,
			}

			// Mock PostDAO.List - 返回空结果
			mockey.Mock((*dao.PostDAO).List).To(func(postDAO *dao.PostDAO, ctx context.Context, filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
				return []*model.Post{}, 0, nil
			}).Build()

			resp, err := logic.GetPostList(req)

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, 200)
			So(len(resp.Data.List), ShouldEqual, 0)
			So(resp.Data.Pagination.Total, ShouldEqual, 0)
			So(resp.Data.Pagination.TotalPages, ShouldEqual, 0)
		})

		Convey("处理数据库查询错误", func() {
			// 重置mock
			mockey.UnPatchAll()

			req := &types.PostListRequest{
				Page:     1,
				Limit:    10,
				SortBy:   "updatedAt",
				SortDesc: true,
			}

			// Mock PostDAO.List - 返回错误
			mockey.Mock((*dao.PostDAO).List).To(func(postDAO *dao.PostDAO, ctx context.Context, filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
				return nil, 0, errors.New("database connection error")
			}).Build()

			resp, err := logic.GetPostList(req)

			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "获取文章列表失败")
		})

		Convey("处理作者信息获取失败", func() {
			// 重置mock
			mockey.UnPatchAll()

			req := &types.PostListRequest{
				Page:     1,
				Limit:    10,
				SortBy:   "updatedAt",
				SortDesc: true,
			}

			mockPosts := []*model.Post{
				{
					ID:        primitive.NewObjectID(),
					Title:     "测试文章",
					AuthorID:  primitive.NewObjectID(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			}

			// Mock PostDAO.List - 成功返回文章
			mockey.Mock((*dao.PostDAO).List).To(func(postDAO *dao.PostDAO, ctx context.Context, filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
				return mockPosts, 1, nil
			}).Build()

			// Mock UserDAO.GetByID - 返回错误
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return nil, errors.New("user not found")
			}).Build()

			resp, err := logic.GetPostList(req)

			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "获取作者信息失败")
		})

		Reset(func() {
			mockey.UnPatchAll()
		})
	})
}
