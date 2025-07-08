package logic

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/heimdall-api/common/dao"
	"github.com/heimdall-api/common/model"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/bytedance/mockey"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreatePageLogic_CreatePage(t *testing.T) {
	Convey("测试页面创建功能", t, func() {
		// 准备测试数据
		ctx := context.WithValue(context.Background(), "userId", "507f1f77bcf86cd799439011")
		svcCtx := &svc.ServiceContext{
			UserDAO: &dao.UserDAO{},
			PageDAO: &dao.PageDAO{},
		}
		logic := NewCreatePageLogic(ctx, svcCtx)

		// 准备测试用户
		testUser := &model.User{
			ID:          primitive.ObjectID{},
			Username:    "testuser",
			DisplayName: "Test User",
			Email:       "test@example.com",
		}

		// 准备测试页面
		testPage := &model.Page{
			ID:       primitive.NewObjectID(),
			Title:    "Test Page",
			Slug:     "test-page",
			Content:  "Test content",
			AuthorID: primitive.ObjectID{},
			Status:   "draft",
			Template: "default",
		}

		Convey("成功创建页面", func() {
			req := &types.PageCreateRequest{
				Title:    "Test Page",
				Content:  "Test content",
				Status:   "draft",
				Template: "default",
			}

			mockey.PatchConvey("Mock UserDAO.GetByID", func() {
				mockey.Mock((*dao.UserDAO).GetByID).Return(testUser, nil).Build()
				mockey.Mock((*dao.PageDAO).GetBySlug).Return(nil, ErrNotFound).Build()
				mockey.Mock((*dao.PageDAO).Create).Return(nil).Build()
				mockey.Mock((*dao.PageDAO).GetByID).Return(testPage, nil).Build()

				resp, err := logic.CreatePage(req)

				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Code, ShouldEqual, 200)
				So(resp.Message, ShouldEqual, "页面创建成功")
				So(resp.Data.Title, ShouldEqual, "Test Page")
			})
		})

		Convey("参数验证失败", func() {
			Convey("标题为空", func() {
				req := &types.PageCreateRequest{
					Title:   "",
					Content: "Test content",
					Status:  "draft",
				}

				resp, err := logic.CreatePage(req)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "标题不能为空")
				So(resp, ShouldBeNil)
			})

			Convey("内容为空", func() {
				req := &types.PageCreateRequest{
					Title:   "Test Page",
					Content: "",
					Status:  "draft",
				}

				resp, err := logic.CreatePage(req)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "内容不能为空")
				So(resp, ShouldBeNil)
			})

			Convey("状态为空", func() {
				req := &types.PageCreateRequest{
					Title:   "Test Page",
					Content: "Test content",
					Status:  "",
				}

				resp, err := logic.CreatePage(req)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "页面状态不能为空")
				So(resp, ShouldBeNil)
			})

			Convey("无效状态", func() {
				req := &types.PageCreateRequest{
					Title:   "Test Page",
					Content: "Test content",
					Status:  "invalid",
				}

				resp, err := logic.CreatePage(req)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "无效的页面状态")
				So(resp, ShouldBeNil)
			})
		})

		Convey("用户认证失败", func() {
			// 创建没有用户ID的context
			ctxNoUser := context.Background()
			logicNoUser := NewCreatePageLogic(ctxNoUser, svcCtx)

			req := &types.PageCreateRequest{
				Title:   "Test Page",
				Content: "Test content",
				Status:  "draft",
			}

			resp, err := logicNoUser.CreatePage(req)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "用户未登录")
			So(resp, ShouldBeNil)
		})

		Convey("用户不存在", func() {
			req := &types.PageCreateRequest{
				Title:   "Test Page",
				Content: "Test content",
				Status:  "draft",
			}

			mockey.PatchConvey("Mock UserDAO.GetByID 返回错误", func() {
				mockey.Mock((*dao.UserDAO).GetByID).Return(nil, ErrUserNotFound).Build()

				resp, err := logic.CreatePage(req)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "作者不存在")
				So(resp, ShouldBeNil)
			})
		})

		Convey("Slug重复检查", func() {
			req := &types.PageCreateRequest{
				Title:   "Test Page",
				Slug:    "existing-slug",
				Content: "Test content",
				Status:  "draft",
			}

			mockey.PatchConvey("Mock Slug已存在", func() {
				mockey.Mock((*dao.UserDAO).GetByID).Return(testUser, nil).Build()
				mockey.Mock((*dao.PageDAO).GetBySlug).Return(testPage, nil).Build()

				resp, err := logic.CreatePage(req)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "slug生成失败")
				So(resp, ShouldBeNil)
			})
		})

		Convey("数据库创建失败", func() {
			req := &types.PageCreateRequest{
				Title:   "Test Page",
				Content: "Test content",
				Status:  "draft",
			}

			mockey.PatchConvey("Mock PageDAO.Create 返回错误", func() {
				mockey.Mock((*dao.UserDAO).GetByID).Return(testUser, nil).Build()
				mockey.Mock((*dao.PageDAO).GetBySlug).Return(nil, ErrNotFound).Build()
				mockey.Mock((*dao.PageDAO).Create).Return(ErrDatabaseError).Build()

				resp, err := logic.CreatePage(req)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "页面创建失败")
				So(resp, ShouldBeNil)
			})
		})

		Convey("自动生成slug", func() {
			req := &types.PageCreateRequest{
				Title:   "Test Page Title",
				Content: "Test content",
				Status:  "draft",
			}

			mockey.PatchConvey("Mock 自动生成slug", func() {
				mockey.Mock((*dao.UserDAO).GetByID).Return(testUser, nil).Build()
				mockey.Mock((*dao.PageDAO).GetBySlug).Return(nil, ErrNotFound).Build()
				mockey.Mock((*dao.PageDAO).Create).Return(nil).Build()
				mockey.Mock((*dao.PageDAO).GetByID).Return(testPage, nil).Build()

				resp, err := logic.CreatePage(req)

				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Code, ShouldEqual, 200)
			})
		})

		Convey("自定义发布时间", func() {
			publishTime := time.Now().Add(24 * time.Hour)
			req := &types.PageCreateRequest{
				Title:       "Test Page",
				Content:     "Test content",
				Status:      "scheduled",
				PublishedAt: publishTime.Format(time.RFC3339),
			}

			mockey.PatchConvey("Mock 带发布时间的创建", func() {
				mockey.Mock((*dao.UserDAO).GetByID).Return(testUser, nil).Build()
				mockey.Mock((*dao.PageDAO).GetBySlug).Return(nil, ErrNotFound).Build()
				mockey.Mock((*dao.PageDAO).Create).Return(nil).Build()
				mockey.Mock((*dao.PageDAO).GetByID).Return(testPage, nil).Build()

				resp, err := logic.CreatePage(req)

				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Code, ShouldEqual, 200)
			})
		})
	})
}

// 定义测试用的错误类型
var (
	ErrNotFound      = fmt.Errorf("not found")
	ErrUserNotFound  = fmt.Errorf("user not found")
	ErrDatabaseError = fmt.Errorf("database error")
)
