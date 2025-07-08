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
	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/dao"
	"github.com/heimdall-api/common/model"
)

func TestUnpublishPostLogic_UnpublishPost(t *testing.T) {
	Convey("测试文章取消发布功能", t, func() {
		// 准备测试数据
		ctx := context.Background()
		svcCtx := &svc.ServiceContext{
			PostDAO: &dao.PostDAO{},
			UserDAO: &dao.UserDAO{},
		}
		logic := NewUnpublishPostLogic(ctx, svcCtx)

		Convey("成功取消发布文章", func() {
			// 重置mock
			mockey.UnPatchAll()

			postID := primitive.NewObjectID()
			authorID := primitive.NewObjectID()
			userID := authorID // 设置为同一个用户

			// 设置用户ID到context（模拟JWT中间件）
			ctxWithUser := context.WithValue(ctx, "uid", userID.Hex())
			logic.ctx = ctxWithUser

			// 准备现有文章（已发布状态）
			publishedTime := time.Now().Add(-1 * time.Hour)
			existingPost := &model.Post{
				ID:          postID,
				Title:       "测试文章",
				Slug:        "test-article",
				Excerpt:     "这是一篇测试文章",
				Markdown:    "# 测试文章\n\n这是内容",
				HTML:        "<h1>测试文章</h1><p>这是内容</p>",
				AuthorID:    authorID,
				Status:      constants.PostStatusPublished,
				Type:        constants.PostTypePost,
				Visibility:  constants.PostVisibilityPublic,
				PublishedAt: &publishedTime,
				CreatedAt:   time.Now().Add(-2 * time.Hour),
				UpdatedAt:   time.Now().Add(-1 * time.Hour),
			}

			// 准备用户信息
			mockUser := &model.User{
				ID:          authorID,
				Username:    "testuser",
				DisplayName: "Test User",
				Email:       "test@example.com",
			}

			// Mock PostDAO.GetByID - 获取现有文章
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				So(id, ShouldEqual, postID.Hex())
				return existingPost, nil
			}).Build()

			// Mock UserDAO.GetByID - 获取用户信息
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				So(id, ShouldEqual, userID.Hex())
				return mockUser, nil
			}).Build()

			// Mock PostDAO.Unpublish - 取消发布文章
			mockey.Mock((*dao.PostDAO).Unpublish).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) error {
				So(id, ShouldEqual, postID.Hex())
				return nil
			}).Build()

			// Mock 获取取消发布后的文章
			unpublishedPost := *existingPost
			unpublishedPost.Status = constants.PostStatusDraft
			unpublishedPost.PublishedAt = nil
			unpublishedPost.UpdatedAt = time.Now()

			// 计数器来区分两次GetByID调用
			getByIDCallCount := 0
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				getByIDCallCount++
				if getByIDCallCount == 1 {
					return existingPost, nil
				}
				return &unpublishedPost, nil
			}).Build()

			// 准备请求
			req := &types.PostUnpublishRequest{
				ID: postID.Hex(),
			}

			// 执行测试
			resp, err := logic.UnpublishPost(req)

			// 验证结果
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Message, ShouldEqual, "文章取消发布成功")
			So(resp.Data.ID, ShouldEqual, postID.Hex())
			So(resp.Data.Status, ShouldEqual, constants.PostStatusDraft)
			So(resp.Data.PublishedAt, ShouldBeEmpty)
			So(resp.Data.Author.ID, ShouldEqual, authorID.Hex())
		})

		Convey("处理无效的文章ID", func() {
			// 重置mock
			mockey.UnPatchAll()

			userID := primitive.NewObjectID()
			ctxWithUser := context.WithValue(ctx, "uid", userID.Hex())
			logic.ctx = ctxWithUser

			// 准备请求
			req := &types.PostUnpublishRequest{
				ID: "invalid-id",
			}

			// 执行测试
			resp, err := logic.UnpublishPost(req)

			// 验证结果
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "无效的文章ID")
		})

		Convey("处理文章不存在", func() {
			// 重置mock
			mockey.UnPatchAll()

			postID := primitive.NewObjectID()
			userID := primitive.NewObjectID()
			ctxWithUser := context.WithValue(ctx, "uid", userID.Hex())
			logic.ctx = ctxWithUser

			// Mock PostDAO.GetByID - 返回文章不存在
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				return nil, errors.New("文章不存在")
			}).Build()

			// 准备请求
			req := &types.PostUnpublishRequest{
				ID: postID.Hex(),
			}

			// 执行测试
			resp, err := logic.UnpublishPost(req)

			// 验证结果
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "获取文章信息失败")
		})

		Convey("处理权限不足", func() {
			// 重置mock
			mockey.UnPatchAll()

			postID := primitive.NewObjectID()
			authorID := primitive.NewObjectID()
			userID := primitive.NewObjectID() // 不同的用户ID

			ctxWithUser := context.WithValue(ctx, "uid", userID.Hex())
			logic.ctx = ctxWithUser

			publishedTime := time.Now().Add(-1 * time.Hour)
			existingPost := &model.Post{
				ID:          postID,
				AuthorID:    authorID,
				Status:      constants.PostStatusPublished,
				PublishedAt: &publishedTime,
			}

			// Mock PostDAO.GetByID
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				return existingPost, nil
			}).Build()

			// Mock UserDAO.GetByID
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return &model.User{ID: userID}, nil
			}).Build()

			// 准备请求
			req := &types.PostUnpublishRequest{
				ID: postID.Hex(),
			}

			// 执行测试
			resp, err := logic.UnpublishPost(req)

			// 验证结果
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "无权限取消发布此文章")
		})

		Convey("处理未发布的文章", func() {
			// 重置mock
			mockey.UnPatchAll()

			postID := primitive.NewObjectID()
			authorID := primitive.NewObjectID()
			userID := authorID

			ctxWithUser := context.WithValue(ctx, "uid", userID.Hex())
			logic.ctx = ctxWithUser

			existingPost := &model.Post{
				ID:       postID,
				AuthorID: authorID,
				Status:   constants.PostStatusDraft, // 草稿状态
			}

			// Mock PostDAO.GetByID
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				return existingPost, nil
			}).Build()

			// Mock UserDAO.GetByID
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return &model.User{ID: userID}, nil
			}).Build()

			// 准备请求
			req := &types.PostUnpublishRequest{
				ID: postID.Hex(),
			}

			// 执行测试
			resp, err := logic.UnpublishPost(req)

			// 验证结果
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "文章未发布")
		})

		Convey("处理数据库取消发布失败", func() {
			// 重置mock
			mockey.UnPatchAll()

			postID := primitive.NewObjectID()
			authorID := primitive.NewObjectID()
			userID := authorID

			ctxWithUser := context.WithValue(ctx, "uid", userID.Hex())
			logic.ctx = ctxWithUser

			publishedTime := time.Now().Add(-1 * time.Hour)
			existingPost := &model.Post{
				ID:          postID,
				AuthorID:    authorID,
				Status:      constants.PostStatusPublished,
				PublishedAt: &publishedTime,
			}

			// Mock PostDAO.GetByID
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				return existingPost, nil
			}).Build()

			// Mock UserDAO.GetByID
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return &model.User{ID: userID}, nil
			}).Build()

			// Mock PostDAO.Unpublish - 返回错误
			mockey.Mock((*dao.PostDAO).Unpublish).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) error {
				return errors.New("数据库连接失败")
			}).Build()

			// 准备请求
			req := &types.PostUnpublishRequest{
				ID: postID.Hex(),
			}

			// 执行测试
			resp, err := logic.UnpublishPost(req)

			// 验证结果
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "取消发布文章失败")
		})
	})
}
