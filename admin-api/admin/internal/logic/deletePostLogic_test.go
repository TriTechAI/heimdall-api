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

func TestDeletePostLogic_DeletePost(t *testing.T) {
	Convey("测试文章删除功能", t, func() {
		// 准备测试数据
		ctx := context.Background()
		svcCtx := &svc.ServiceContext{
			PostDAO: &dao.PostDAO{},
			UserDAO: &dao.UserDAO{},
		}
		logic := NewDeletePostLogic(ctx, svcCtx)

		Convey("成功删除文章", func() {
			// 重置mock
			mockey.UnPatchAll()

			postID := primitive.NewObjectID()
			authorID := primitive.NewObjectID()
			userID := authorID // 设置为同一个用户

			// 设置用户ID到context（模拟JWT中间件）
			ctxWithUser := context.WithValue(ctx, "uid", userID.Hex())
			logic.ctx = ctxWithUser

			// 准备现有文章
			existingPost := &model.Post{
				ID:         postID,
				Title:      "测试文章",
				Slug:       "test-article",
				Excerpt:    "这是一篇测试文章",
				Markdown:   "# 测试文章\n\n这是内容",
				HTML:       "<h1>测试文章</h1><p>这是内容</p>",
				AuthorID:   authorID,
				Status:     constants.PostStatusDraft,
				Type:       constants.PostTypePost,
				Visibility: constants.PostVisibilityPublic,
				CreatedAt:  time.Now().Add(-1 * time.Hour),
				UpdatedAt:  time.Now().Add(-30 * time.Minute),
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

			// Mock PostDAO.Delete - 软删除文章
			mockey.Mock((*dao.PostDAO).Delete).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) error {
				So(id, ShouldEqual, postID.Hex())
				return nil
			}).Build()

			// 准备请求
			req := &types.PostDeleteRequest{
				ID: postID.Hex(),
			}

			// 执行测试
			resp, err := logic.DeletePost(req)

			// 验证结果
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Message, ShouldEqual, "文章删除成功")
			So(resp.Timestamp, ShouldNotBeEmpty)
		})

		Convey("处理无效的文章ID", func() {
			// 重置mock
			mockey.UnPatchAll()

			userID := primitive.NewObjectID()
			ctxWithUser := context.WithValue(ctx, "uid", userID.Hex())
			logic.ctx = ctxWithUser

			// 准备请求
			req := &types.PostDeleteRequest{
				ID: "invalid-id",
			}

			// 执行测试
			resp, err := logic.DeletePost(req)

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
			req := &types.PostDeleteRequest{
				ID: postID.Hex(),
			}

			// 执行测试
			resp, err := logic.DeletePost(req)

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

			existingPost := &model.Post{
				ID:       postID,
				AuthorID: authorID,
				Status:   constants.PostStatusDraft,
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
			req := &types.PostDeleteRequest{
				ID: postID.Hex(),
			}

			// 执行测试
			resp, err := logic.DeletePost(req)

			// 验证结果
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "无权限删除此文章")
		})

		Convey("处理用户认证失败", func() {
			// 重置mock
			mockey.UnPatchAll()

			postID := primitive.NewObjectID()

			// 没有设置用户ID到context
			logic.ctx = ctx

			// 准备请求
			req := &types.PostDeleteRequest{
				ID: postID.Hex(),
			}

			// 执行测试
			resp, err := logic.DeletePost(req)

			// 验证结果
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "用户认证失败")
		})

		Convey("处理用户不存在", func() {
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
				Status:   constants.PostStatusDraft,
			}

			// Mock PostDAO.GetByID
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				return existingPost, nil
			}).Build()

			// Mock UserDAO.GetByID - 返回用户不存在
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return nil, errors.New("用户不存在")
			}).Build()

			// 准备请求
			req := &types.PostDeleteRequest{
				ID: postID.Hex(),
			}

			// 执行测试
			resp, err := logic.DeletePost(req)

			// 验证结果
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "获取用户信息失败")
		})

		Convey("处理数据库删除失败", func() {
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
				Status:   constants.PostStatusDraft,
			}

			// Mock PostDAO.GetByID
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				return existingPost, nil
			}).Build()

			// Mock UserDAO.GetByID
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return &model.User{ID: userID}, nil
			}).Build()

			// Mock PostDAO.Delete - 返回错误
			mockey.Mock((*dao.PostDAO).Delete).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) error {
				return errors.New("数据库连接失败")
			}).Build()

			// 准备请求
			req := &types.PostDeleteRequest{
				ID: postID.Hex(),
			}

			// 执行测试
			resp, err := logic.DeletePost(req)

			// 验证结果
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "删除文章失败")
		})

		Convey("处理已删除的文章", func() {
			// 重置mock
			mockey.UnPatchAll()

			postID := primitive.NewObjectID()
			authorID := primitive.NewObjectID()
			userID := authorID

			ctxWithUser := context.WithValue(ctx, "uid", userID.Hex())
			logic.ctx = ctxWithUser

			// 准备已删除的文章（状态为archived）
			existingPost := &model.Post{
				ID:       postID,
				AuthorID: authorID,
				Status:   constants.PostStatusArchived, // 已删除状态
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
			req := &types.PostDeleteRequest{
				ID: postID.Hex(),
			}

			// 执行测试
			resp, err := logic.DeletePost(req)

			// 验证结果
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "文章已被删除")
		})
	})
}
