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

func TestUpdatePostLogic_UpdatePost(t *testing.T) {
	Convey("测试文章更新功能", t, func() {
		// 准备测试数据
		ctx := context.Background()
		svcCtx := &svc.ServiceContext{
			PostDAO: &dao.PostDAO{},
			UserDAO: &dao.UserDAO{},
		}
		logic := NewUpdatePostLogic(ctx, svcCtx)

		Convey("成功更新文章基础信息", func() {
			// 重置mock
			mockey.UnPatchAll()

			postID := primitive.NewObjectID()
			authorID := primitive.NewObjectID()
			userID := authorID // 设置为同一个用户

			// 设置用户ID到context（模拟JWT中间件）
			ctxWithUser := context.WithValue(ctx, "uid", userID.Hex())
			logic = NewUpdatePostLogic(ctxWithUser, svcCtx)

			req := &types.PostUpdateRequest{
				ID:              postID.Hex(),
				Title:           "更新后的标题",
				Excerpt:         "更新后的摘要",
				Markdown:        "# 更新后的内容\n\n这是更新后的文章内容",
				FeaturedImage:   "https://example.com/new-image.jpg",
				MetaTitle:       "更新后的SEO标题",
				MetaDescription: "更新后的SEO描述",
			}

			// 准备现有文章数据
			existingPost := &model.Post{
				ID:              postID,
				Title:           "原始标题",
				Slug:            "original-title",
				Excerpt:         "原始摘要",
				Markdown:        "# 原始内容",
				HTML:            "<h1>原始内容</h1>",
				FeaturedImage:   "https://example.com/old-image.jpg",
				Type:            "post",
				Status:          "draft",
				Visibility:      "public",
				AuthorID:        authorID,
				Tags:            []model.Tag{{Name: "Go", Slug: "go"}},
				MetaTitle:       "原始SEO标题",
				MetaDescription: "原始SEO描述",
				ReadingTime:     5,
				WordCount:       100,
				ViewCount:       50,
				CreatedAt:       time.Now().Add(-2 * time.Hour),
				UpdatedAt:       time.Now().Add(-1 * time.Hour),
			}

			updatedPost := &model.Post{
				ID:              postID,
				Title:           "更新后的标题",
				Slug:            "original-title", // slug不变
				Excerpt:         "更新后的摘要",
				Markdown:        "# 更新后的内容\n\n这是更新后的文章内容",
				HTML:            "<h1>更新后的内容</h1><p>这是更新后的文章内容</p>",
				FeaturedImage:   "https://example.com/new-image.jpg",
				Type:            "post",
				Status:          "draft",
				Visibility:      "public",
				AuthorID:        authorID,
				Tags:            []model.Tag{{Name: "Go", Slug: "go"}},
				MetaTitle:       "更新后的SEO标题",
				MetaDescription: "更新后的SEO描述",
				ReadingTime:     8,
				WordCount:       200,
				ViewCount:       50,
				CreatedAt:       existingPost.CreatedAt,
				UpdatedAt:       time.Now(),
			}

			mockUser := &model.User{
				ID:          authorID,
				Username:    "testuser",
				DisplayName: "Test User",
				Email:       "test@example.com",
			}

			// Mock PostDAO.GetByID - 处理两次调用：获取现有文章和获取更新后的文章
			callCount := 0
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				So(id, ShouldEqual, postID.Hex())
				callCount++
				if callCount == 1 {
					// 第一次调用：返回现有文章
					return existingPost, nil
				} else {
					// 第二次调用：返回更新后的文章
					return updatedPost, nil
				}
			}).Build()

			// Mock PostDAO.Update - 更新文章
			mockey.Mock((*dao.PostDAO).Update).To(func(postDAO *dao.PostDAO, ctx context.Context, id string, updates map[string]interface{}) error {
				So(id, ShouldEqual, postID.Hex())
				So(updates["title"], ShouldEqual, "更新后的标题")
				So(updates["excerpt"], ShouldEqual, "更新后的摘要")
				So(updates["markdown"], ShouldEqual, "# 更新后的内容\n\n这是更新后的文章内容")
				So(updates["featuredImage"], ShouldEqual, "https://example.com/new-image.jpg")
				So(updates["metaTitle"], ShouldEqual, "更新后的SEO标题")
				So(updates["metaDescription"], ShouldEqual, "更新后的SEO描述")
				return nil
			}).Build()

			// Mock UserDAO.GetByID - 获取作者信息
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				So(id, ShouldEqual, authorID.Hex())
				return mockUser, nil
			}).Build()

			// 执行测试
			resp, err := logic.UpdatePost(req)

			// 验证结果
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Message, ShouldEqual, "文章更新成功")
			So(resp.Data.ID, ShouldEqual, postID.Hex())
			So(resp.Data.Title, ShouldEqual, "更新后的标题")
			So(resp.Data.Excerpt, ShouldEqual, "更新后的摘要")
			So(resp.Data.FeaturedImage, ShouldEqual, "https://example.com/new-image.jpg")
			So(resp.Data.MetaTitle, ShouldEqual, "更新后的SEO标题")
		})

		Convey("支持部分更新", func() {
			// 重置mock
			mockey.UnPatchAll()

			postID := primitive.NewObjectID()
			authorID := primitive.NewObjectID()
			userID := authorID // 设置为同一个用户

			ctxWithUser := context.WithValue(ctx, "uid", userID.Hex())
			logic = NewUpdatePostLogic(ctxWithUser, svcCtx)

			// 只更新标题和摘要
			req := &types.PostUpdateRequest{
				ID:      postID.Hex(),
				Title:   "仅更新标题",
				Excerpt: "仅更新摘要",
			}

			existingPost := &model.Post{
				ID:        postID,
				Title:     "原始标题",
				Excerpt:   "原始摘要",
				Markdown:  "# 原始内容",
				AuthorID:  authorID,
				CreatedAt: time.Now().Add(-1 * time.Hour),
				UpdatedAt: time.Now().Add(-30 * time.Minute),
			}

			mockUser := &model.User{
				ID:          authorID,
				Username:    "testuser",
				DisplayName: "Test User",
				Email:       "test@example.com",
			}

			// Mock 获取更新后的文章
			updatedPost := *existingPost
			updatedPost.Title = "仅更新标题"
			updatedPost.Excerpt = "仅更新摘要"
			updatedPost.UpdatedAt = time.Now()

			// Mock PostDAO.GetByID - 处理两次调用
			callCount := 0
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				callCount++
				if callCount == 1 {
					return existingPost, nil
				} else {
					return &updatedPost, nil
				}
			}).Build()

			// Mock PostDAO.Update - 验证只更新指定字段
			mockey.Mock((*dao.PostDAO).Update).To(func(postDAO *dao.PostDAO, ctx context.Context, id string, updates map[string]interface{}) error {
				So(updates["title"], ShouldEqual, "仅更新标题")
				So(updates["excerpt"], ShouldEqual, "仅更新摘要")
				// 其他字段不应该在updates中
				_, hasMarkdown := updates["markdown"]
				So(hasMarkdown, ShouldBeFalse)
				return nil
			}).Build()

			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return mockUser, nil
			}).Build()

			resp, err := logic.UpdatePost(req)

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Data.Title, ShouldEqual, "仅更新标题")
			So(resp.Data.Excerpt, ShouldEqual, "仅更新摘要")
		})

		Convey("处理无效的文章ID", func() {
			// 重置mock
			mockey.UnPatchAll()

			userID := primitive.NewObjectID()
			ctxWithUser := context.WithValue(ctx, "uid", userID.Hex())
			logic = NewUpdatePostLogic(ctxWithUser, svcCtx)

			req := &types.PostUpdateRequest{
				ID:    "invalid-id",
				Title: "测试标题",
			}

			resp, err := logic.UpdatePost(req)

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
			logic = NewUpdatePostLogic(ctxWithUser, svcCtx)

			req := &types.PostUpdateRequest{
				ID:    postID.Hex(),
				Title: "测试标题",
			}

			// Mock PostDAO.GetByID - 返回文章不存在
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				return nil, errors.New("post not found")
			}).Build()

			resp, err := logic.UpdatePost(req)

			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "文章不存在")
		})

		Convey("处理权限验证失败", func() {
			// 重置mock
			mockey.UnPatchAll()

			postID := primitive.NewObjectID()
			authorID := primitive.NewObjectID()
			userID := primitive.NewObjectID() // 不同的用户ID

			ctxWithUser := context.WithValue(ctx, "uid", userID.Hex())
			logic = NewUpdatePostLogic(ctxWithUser, svcCtx)

			req := &types.PostUpdateRequest{
				ID:    postID.Hex(),
				Title: "测试标题",
			}

			existingPost := &model.Post{
				ID:       postID,
				Title:    "原始标题",
				AuthorID: authorID, // 不同的作者
			}

			// Mock PostDAO.GetByID
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				return existingPost, nil
			}).Build()

			resp, err := logic.UpdatePost(req)

			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "无权限修改此文章")
		})

		Convey("处理slug重复", func() {
			// 重置mock
			mockey.UnPatchAll()

			postID := primitive.NewObjectID()
			authorID := primitive.NewObjectID()
			userID := authorID // 同一个用户

			ctxWithUser := context.WithValue(ctx, "uid", userID.Hex())
			logic = NewUpdatePostLogic(ctxWithUser, svcCtx)

			req := &types.PostUpdateRequest{
				ID:   postID.Hex(),
				Slug: "existing-slug",
			}

			existingPost := &model.Post{
				ID:       postID,
				Title:    "原始标题",
				Slug:     "original-slug",
				AuthorID: authorID,
			}

			// Mock PostDAO.GetByID
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				return existingPost, nil
			}).Build()

			// Mock PostDAO.GetBySlug - 返回slug已存在
			mockey.Mock((*dao.PostDAO).GetBySlug).To(func(postDAO *dao.PostDAO, ctx context.Context, slug string) (*model.Post, error) {
				if slug == "existing-slug" {
					// 返回另一个文章，表示slug已被占用
					return &model.Post{
						ID:   primitive.NewObjectID(),
						Slug: "existing-slug",
					}, nil
				}
				return nil, errors.New("not found")
			}).Build()

			resp, err := logic.UpdatePost(req)

			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "slug已被使用")
		})

		Convey("处理数据库更新错误", func() {
			// 重置mock
			mockey.UnPatchAll()

			postID := primitive.NewObjectID()
			authorID := primitive.NewObjectID()
			userID := authorID

			ctxWithUser := context.WithValue(ctx, "uid", userID.Hex())
			logic = NewUpdatePostLogic(ctxWithUser, svcCtx)

			req := &types.PostUpdateRequest{
				ID:    postID.Hex(),
				Title: "测试标题",
			}

			existingPost := &model.Post{
				ID:       postID,
				Title:    "原始标题",
				AuthorID: authorID,
			}

			// Mock PostDAO.GetByID
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				return existingPost, nil
			}).Build()

			// Mock PostDAO.Update - 返回数据库错误
			mockey.Mock((*dao.PostDAO).Update).To(func(postDAO *dao.PostDAO, ctx context.Context, id string, updates map[string]interface{}) error {
				return errors.New("database connection error")
			}).Build()

			resp, err := logic.UpdatePost(req)

			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "更新文章失败")
		})

		Convey("处理用户未认证", func() {
			// 重置mock
			mockey.UnPatchAll()

			// 没有设置用户ID的context
			logic = NewUpdatePostLogic(ctx, svcCtx)

			postID := primitive.NewObjectID()
			req := &types.PostUpdateRequest{
				ID:    postID.Hex(),
				Title: "测试标题",
			}

			resp, err := logic.UpdatePost(req)

			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "用户未认证")
		})

		Reset(func() {
			mockey.UnPatchAll()
		})
	})
}
