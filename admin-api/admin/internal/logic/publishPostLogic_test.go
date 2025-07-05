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

func TestPublishPostLogic_PublishPost(t *testing.T) {
	Convey("测试文章发布功能", t, func() {
		// 准备测试数据
		ctx := context.Background()
		svcCtx := &svc.ServiceContext{
			PostDAO: &dao.PostDAO{},
			UserDAO: &dao.UserDAO{},
		}
		logic := NewPublishPostLogic(ctx, svcCtx)

		Convey("成功发布草稿文章", func() {
			// 重置mock
			mockey.UnPatchAll()

			postID := primitive.NewObjectID()
			authorID := primitive.NewObjectID()
			userID := authorID // 设置为同一个用户

			// 设置用户ID到context（模拟JWT中间件）
			ctxWithUser := context.WithValue(ctx, "uid", userID.Hex())
			logic.ctx = ctxWithUser

			// 准备现有文章（草稿状态）
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

			// Mock PostDAO.Update - 更新发布时间
			mockey.Mock((*dao.PostDAO).Update).To(func(postDAO *dao.PostDAO, ctx context.Context, id string, updates map[string]interface{}) error {
				So(id, ShouldEqual, postID.Hex())
				So(updates["publishedAt"], ShouldNotBeNil)
				return nil
			}).Build()

			// Mock PostDAO.Publish - 发布文章
			mockey.Mock((*dao.PostDAO).Publish).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) error {
				So(id, ShouldEqual, postID.Hex())
				return nil
			}).Build()

			// Mock 获取发布后的文章
			publishedPost := *existingPost
			publishedPost.Status = constants.PostStatusPublished
			now := time.Now()
			publishedPost.PublishedAt = &now
			publishedPost.UpdatedAt = now

			// 计数器来区分两次GetByID调用
			getByIDCallCount := 0
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				getByIDCallCount++
				if getByIDCallCount == 1 {
					return existingPost, nil
				}
				return &publishedPost, nil
			}).Build()

			// 准备请求
			req := &types.PostPublishRequest{
				ID: postID.Hex(),
			}

			// 执行测试
			resp, err := logic.PublishPost(req)

			// 验证结果
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Message, ShouldEqual, "文章发布成功")
			So(resp.Data.ID, ShouldEqual, postID.Hex())
			So(resp.Data.Status, ShouldEqual, constants.PostStatusPublished)
			So(resp.Data.PublishedAt, ShouldNotBeEmpty)
			So(resp.Data.Author.ID, ShouldEqual, authorID.Hex())
		})

		Convey("成功发布文章并指定发布时间", func() {
			// 重置mock
			mockey.UnPatchAll()

			postID := primitive.NewObjectID()
			authorID := primitive.NewObjectID()
			userID := authorID

			ctxWithUser := context.WithValue(ctx, "uid", userID.Hex())
			logic.ctx = ctxWithUser

			existingPost := &model.Post{
				ID:         postID,
				Title:      "测试文章",
				Slug:       "test-article",
				AuthorID:   authorID,
				Status:     constants.PostStatusDraft,
				Type:       constants.PostTypePost,
				Visibility: constants.PostVisibilityPublic,
				CreatedAt:  time.Now().Add(-1 * time.Hour),
				UpdatedAt:  time.Now().Add(-30 * time.Minute),
			}

			mockUser := &model.User{
				ID:          authorID,
				Username:    "testuser",
				DisplayName: "Test User",
				Email:       "test@example.com",
			}

			// 指定的发布时间
			publishTime := time.Now().Add(1 * time.Hour)
			publishTimeStr := publishTime.Format(time.RFC3339)

			// Mock PostDAO.GetByID
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				return existingPost, nil
			}).Build()

			// Mock UserDAO.GetByID
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return mockUser, nil
			}).Build()

			// Mock PostDAO.Update - 验证发布时间
			mockey.Mock((*dao.PostDAO).Update).To(func(postDAO *dao.PostDAO, ctx context.Context, id string, updates map[string]interface{}) error {
				So(id, ShouldEqual, postID.Hex())
				publishedAtValue := updates["publishedAt"].(*time.Time)
				So(publishedAtValue, ShouldNotBeNil)
				So(publishedAtValue.Format(time.RFC3339), ShouldEqual, publishTimeStr)
				return nil
			}).Build()

			// Mock PostDAO.Publish - 发布文章
			mockey.Mock((*dao.PostDAO).Publish).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) error {
				So(id, ShouldEqual, postID.Hex())
				return nil
			}).Build()

			// Mock 获取发布后的文章
			publishedPost := *existingPost
			publishedPost.Status = constants.PostStatusPublished
			publishedPost.PublishedAt = &publishTime
			publishedPost.UpdatedAt = time.Now()

			getByIDCallCount := 0
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				getByIDCallCount++
				if getByIDCallCount == 1 {
					return existingPost, nil
				}
				return &publishedPost, nil
			}).Build()

			// 准备请求
			req := &types.PostPublishRequest{
				ID:          postID.Hex(),
				PublishedAt: publishTimeStr,
			}

			// 执行测试
			resp, err := logic.PublishPost(req)

			// 验证结果
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Message, ShouldEqual, "文章发布成功")
			So(resp.Data.PublishedAt, ShouldEqual, publishTimeStr)
		})

		Convey("处理无效的文章ID", func() {
			// 重置mock
			mockey.UnPatchAll()

			userID := primitive.NewObjectID()
			ctxWithUser := context.WithValue(ctx, "uid", userID.Hex())
			logic.ctx = ctxWithUser

			// 准备请求
			req := &types.PostPublishRequest{
				ID: "invalid-id",
			}

			// 执行测试
			resp, err := logic.PublishPost(req)

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
			req := &types.PostPublishRequest{
				ID: postID.Hex(),
			}

			// 执行测试
			resp, err := logic.PublishPost(req)

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
			req := &types.PostPublishRequest{
				ID: postID.Hex(),
			}

			// 执行测试
			resp, err := logic.PublishPost(req)

			// 验证结果
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "无权限发布此文章")
		})

		Convey("处理已发布的文章", func() {
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
				Status:      constants.PostStatusPublished, // 已发布状态
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
			req := &types.PostPublishRequest{
				ID: postID.Hex(),
			}

			// 执行测试
			resp, err := logic.PublishPost(req)

			// 验证结果
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "文章已经发布")
		})

		Convey("处理发布时间解析错误", func() {
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

			// 准备请求 - 无效的时间格式
			req := &types.PostPublishRequest{
				ID:          postID.Hex(),
				PublishedAt: "invalid-time-format",
			}

			// 执行测试
			resp, err := logic.PublishPost(req)

			// 验证结果
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "无效的发布时间格式")
		})

		Convey("处理数据库发布失败", func() {
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

			// Mock PostDAO.Update - 更新发布时间
			mockey.Mock((*dao.PostDAO).Update).To(func(postDAO *dao.PostDAO, ctx context.Context, id string, updates map[string]interface{}) error {
				return nil
			}).Build()

			// Mock PostDAO.Publish - 返回错误
			mockey.Mock((*dao.PostDAO).Publish).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) error {
				return errors.New("数据库连接失败")
			}).Build()

			// 准备请求
			req := &types.PostPublishRequest{
				ID: postID.Hex(),
			}

			// 执行测试
			resp, err := logic.PublishPost(req)

			// 验证结果
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "发布文章失败")
		})
	})
}
