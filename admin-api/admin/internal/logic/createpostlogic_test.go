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

func TestCreatePostLogic_CreatePost(t *testing.T) {
	Convey("测试文章创建功能", t, func() {
		// 准备测试数据
		ctx := context.Background()
		svcCtx := &svc.ServiceContext{
			PostDAO: &dao.PostDAO{},
			UserDAO: &dao.UserDAO{},
		}
		logic := NewCreatePostLogic(ctx, svcCtx)

		// 模拟用户信息（从JWT中获取）
		authorID := primitive.NewObjectID()

		Convey("成功创建文章", func() {
			// 重置mock
			mockey.UnPatchAll()

			// 设置用户上下文
			logic.ctx = context.WithValue(logic.ctx, "userId", authorID.Hex())

			// 准备请求数据
			req := &types.PostCreateRequest{
				Title:           "测试文章标题",
				Slug:            "",
				Excerpt:         "这是文章摘要",
				Markdown:        "# 标题\n\n这是文章内容",
				FeaturedImage:   "https://example.com/image.jpg",
				Type:            "post",
				Status:          "draft",
				Visibility:      "public",
				Tags:            []types.TagInfo{{Name: "Go", Slug: "go"}},
				MetaTitle:       "SEO标题",
				MetaDescription: "SEO描述",
				CanonicalURL:    "https://example.com/canonical",
			}

			// 准备模拟数据
			mockUser := &model.User{
				ID:          authorID,
				Username:    "testuser",
				DisplayName: "Test User",
				Email:       "test@example.com",
			}

			mockPost := &model.Post{
				ID:              primitive.NewObjectID(),
				Title:           req.Title,
				Slug:            "测试文章标题",
				Excerpt:         req.Excerpt,
				Markdown:        req.Markdown,
				HTML:            "<h1>标题</h1><p>这是文章内容</p>",
				FeaturedImage:   req.FeaturedImage,
				Type:            req.Type,
				Status:          req.Status,
				Visibility:      req.Visibility,
				AuthorID:        authorID,
				Tags:            []model.Tag{{Name: "Go", Slug: "go"}},
				MetaTitle:       req.MetaTitle,
				MetaDescription: req.MetaDescription,
				CanonicalURL:    req.CanonicalURL,
				ReadingTime:     1,
				WordCount:       10,
				ViewCount:       0,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}

			// Mock UserDAO.GetByID
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return mockUser, nil
			}).Build()

			// Mock PostDAO.GetBySlug (检查slug重复)
			mockey.Mock((*dao.PostDAO).GetBySlug).To(func(postDAO *dao.PostDAO, ctx context.Context, slug string) (*model.Post, error) {
				return nil, errors.New("post not found") // 表示slug不重复
			}).Build()

			// Mock PostDAO.Create
			mockey.Mock((*dao.PostDAO).Create).To(func(postDAO *dao.PostDAO, ctx context.Context, post *model.Post) error {
				post.ID = mockPost.ID
				return nil
			}).Build()

			// Mock PostDAO.GetByID (获取创建后的文章)
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				return mockPost, nil
			}).Build()

			// 执行测试
			resp, err := logic.CreatePost(req)

			// 验证结果
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Message, ShouldEqual, "文章创建成功")
			So(resp.Data.Title, ShouldEqual, req.Title)
			So(resp.Data.Status, ShouldEqual, req.Status)
			So(resp.Data.Author.Username, ShouldEqual, mockUser.Username)
		})

		Convey("创建文章时自动生成slug", func() {
			// 重置mock
			mockey.UnPatchAll()

			// 设置用户上下文
			logic.ctx = context.WithValue(logic.ctx, "userId", authorID.Hex())

			req := &types.PostCreateRequest{
				Title:      "测试文章标题",
				Slug:       "", // 空slug，应该自动生成
				Markdown:   "# 标题\n\n内容",
				Type:       "post",
				Status:     "draft",
				Visibility: "public",
			}

			mockUser := &model.User{
				ID:          authorID,
				Username:    "testuser",
				DisplayName: "Test User",
			}

			// Mock UserDAO.GetByID
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return mockUser, nil
			}).Build()

			// Mock PostDAO.GetBySlug (检查slug重复)
			mockey.Mock((*dao.PostDAO).GetBySlug).To(func(postDAO *dao.PostDAO, ctx context.Context, slug string) (*model.Post, error) {
				return nil, errors.New("post not found")
			}).Build()

			// Mock PostDAO.Create
			mockey.Mock((*dao.PostDAO).Create).To(func(postDAO *dao.PostDAO, ctx context.Context, post *model.Post) error {
				// 验证slug已经生成
				So(post.Slug, ShouldNotBeEmpty)
				return nil
			}).Build()

			// Mock PostDAO.GetByID
			mockPost := &model.Post{
				ID:    primitive.NewObjectID(),
				Title: req.Title,
				Slug:  "测试文章标题",
			}
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				return mockPost, nil
			}).Build()

			// 执行测试
			resp, err := logic.CreatePost(req)

			// 验证结果
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Data.Slug, ShouldNotBeEmpty)
		})

		Convey("slug重复时自动生成新的slug", func() {
			// 重置mock
			mockey.UnPatchAll()

			// 设置用户上下文
			logic.ctx = context.WithValue(logic.ctx, "userId", authorID.Hex())

			req := &types.PostCreateRequest{
				Title:      "重复标题",
				Slug:       "重复标题",
				Markdown:   "# 内容",
				Type:       "post",
				Status:     "draft",
				Visibility: "public",
			}

			mockUser := &model.User{
				ID:          authorID,
				Username:    "testuser",
				DisplayName: "Test User",
			}

			existingPost := &model.Post{
				ID:   primitive.NewObjectID(),
				Slug: "重复标题",
			}

			// Mock UserDAO.GetByID
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return mockUser, nil
			}).Build()

			// Mock PostDAO.GetBySlug (第一次返回已存在的文章，第二次返回不存在)
			callCount := 0
			mockey.Mock((*dao.PostDAO).GetBySlug).To(func(postDAO *dao.PostDAO, ctx context.Context, slug string) (*model.Post, error) {
				callCount++
				if callCount == 1 {
					return existingPost, nil // 第一次检查，slug已存在
				}
				return nil, errors.New("post not found") // 第二次检查，新slug不存在
			}).Build()

			// Mock PostDAO.Create
			mockey.Mock((*dao.PostDAO).Create).To(func(postDAO *dao.PostDAO, ctx context.Context, post *model.Post) error {
				// 验证slug已经修改
				So(post.Slug, ShouldNotEqual, req.Slug)
				So(post.Slug, ShouldContainSubstring, req.Slug)
				return nil
			}).Build()

			// Mock PostDAO.GetByID
			mockPost := &model.Post{
				ID:    primitive.NewObjectID(),
				Title: req.Title,
				Slug:  "重复标题-1",
			}
			mockey.Mock((*dao.PostDAO).GetByID).To(func(postDAO *dao.PostDAO, ctx context.Context, id string) (*model.Post, error) {
				return mockPost, nil
			}).Build()

			// 执行测试
			resp, err := logic.CreatePost(req)

			// 验证结果
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Data.Slug, ShouldNotEqual, req.Slug)
		})

		Convey("参数验证失败", func() {
			Convey("标题为空", func() {
				req := &types.PostCreateRequest{
					Title:      "", // 空标题
					Markdown:   "内容",
					Type:       "post",
					Status:     "draft",
					Visibility: "public",
				}

				resp, err := logic.CreatePost(req)

				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "标题不能为空")
			})

			Convey("内容为空", func() {
				req := &types.PostCreateRequest{
					Title:      "标题",
					Markdown:   "", // 空内容
					Type:       "post",
					Status:     "draft",
					Visibility: "public",
				}

				resp, err := logic.CreatePost(req)

				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "内容不能为空")
			})

			Convey("无效的文章类型", func() {
				req := &types.PostCreateRequest{
					Title:      "标题",
					Markdown:   "内容",
					Type:       "invalid", // 无效类型
					Status:     "draft",
					Visibility: "public",
				}

				resp, err := logic.CreatePost(req)

				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "无效的文章类型")
			})
		})

		Convey("作者不存在", func() {
			// 设置用户上下文
			logic.ctx = context.WithValue(logic.ctx, "userId", authorID.Hex())

			req := &types.PostCreateRequest{
				Title:      "标题",
				Markdown:   "内容",
				Type:       "post",
				Status:     "draft",
				Visibility: "public",
			}

			// 重置之前的mock，避免冲突
			mockey.UnPatchAll()

			// Mock UserDAO.GetByID 返回错误
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return nil, errors.New("user not found")
			}).Build()

			resp, err := logic.CreatePost(req)

			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "作者不存在")
		})

		Convey("数据库创建失败", func() {
			// 设置用户上下文
			logic.ctx = context.WithValue(logic.ctx, "userId", authorID.Hex())

			req := &types.PostCreateRequest{
				Title:      "标题",
				Markdown:   "内容",
				Type:       "post",
				Status:     "draft",
				Visibility: "public",
			}

			mockUser := &model.User{
				ID:          authorID,
				Username:    "testuser",
				DisplayName: "Test User",
			}

			// 重置之前的mock，避免冲突
			mockey.UnPatchAll()

			// Mock UserDAO.GetByID - 成功返回用户
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return mockUser, nil
			}).Build()

			// Mock PostDAO.GetBySlug - slug不存在
			mockey.Mock((*dao.PostDAO).GetBySlug).To(func(postDAO *dao.PostDAO, ctx context.Context, slug string) (*model.Post, error) {
				return nil, errors.New("post not found")
			}).Build()

			// Mock PostDAO.Create - 返回数据库错误
			mockey.Mock((*dao.PostDAO).Create).To(func(postDAO *dao.PostDAO, ctx context.Context, post *model.Post) error {
				return errors.New("database error")
			}).Build()

			resp, err := logic.CreatePost(req)

			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "文章创建失败")
		})

		Reset(func() {
			mockey.UnPatchAll()
		})
	})
}
