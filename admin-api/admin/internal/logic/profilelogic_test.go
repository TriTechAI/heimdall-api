package logic

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bytedance/mockey"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/heimdall-api/admin-api/admin/internal/config"
	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/dao"
	"github.com/heimdall-api/common/model"
)

func TestProfileLogic_Profile(t *testing.T) {
	Convey("ProfileLogic Profile Tests", t, func() {
		// 创建测试用的ServiceContext
		cfg := config.Config{
			Auth: struct {
				AccessSecret string
				AccessExpire int64
			}{
				AccessSecret: "test-secret",
				AccessExpire: 3600,
			},
			JWTBusiness: config.JWTBusinessConfig{
				RefreshExpire: 7200,
			},
		}

		svcCtx := &svc.ServiceContext{
			Config:  cfg,
			UserDAO: &dao.UserDAO{},
		}

		Convey("Should return error when user ID not found in context", func() {
			// 创建没有用户ID的context
			ctx := context.Background()
			profileLogic := NewProfileLogic(ctx, svcCtx)

			resp, err := profileLogic.Profile()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "用户未认证")
			So(resp, ShouldBeNil)
		})

		Convey("Should return error when user ID is empty", func() {
			// 创建包含空用户ID的context
			ctx := context.WithValue(context.Background(), "uid", "")
			profileLogic := NewProfileLogic(ctx, svcCtx)

			resp, err := profileLogic.Profile()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "用户ID无效")
			So(resp, ShouldBeNil)
		})

		Convey("Should return error when user ID is invalid", func() {
			// 创建包含无效用户ID的context
			ctx := context.WithValue(context.Background(), "uid", "invalid-object-id")
			profileLogic := NewProfileLogic(ctx, svcCtx)

			resp, err := profileLogic.Profile()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "用户ID格式无效")
			So(resp, ShouldBeNil)
		})

		Convey("Should return error when user not found", func() {
			userID := primitive.NewObjectID()
			ctx := context.WithValue(context.Background(), "uid", userID.Hex())
			profileLogic := NewProfileLogic(ctx, svcCtx)

			// Mock UserDAO.GetByID to return nil (user not found)
			mock := mockey.Mock((*dao.UserDAO).GetByID).Return(nil, nil).Build()
			defer mock.UnPatch()

			resp, err := profileLogic.Profile()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "用户不存在")
			So(resp, ShouldBeNil)
		})

		Convey("Should return error when database error occurs", func() {
			userID := primitive.NewObjectID()
			ctx := context.WithValue(context.Background(), "uid", userID.Hex())
			profileLogic := NewProfileLogic(ctx, svcCtx)

			// Mock UserDAO.GetByID to return database error
			mock := mockey.Mock((*dao.UserDAO).GetByID).Return(nil, errors.New("database connection failed")).Build()
			defer mock.UnPatch()

			resp, err := profileLogic.Profile()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "系统错误，请稍后重试")
			So(resp, ShouldBeNil)
		})

		Convey("Should return error when user status is inactive", func() {
			userID := primitive.NewObjectID()
			ctx := context.WithValue(context.Background(), "uid", userID.Hex())
			profileLogic := NewProfileLogic(ctx, svcCtx)

			// 创建非活跃用户
			testUser := &model.User{
				ID:          userID,
				Username:    "testuser",
				Email:       "test@example.com",
				DisplayName: "Test User",
				Role:        constants.UserRoleAuthor,
				Status:      constants.UserStatusInactive,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			// Mock UserDAO.GetByID to return inactive user
			mock := mockey.Mock((*dao.UserDAO).GetByID).Return(testUser, nil).Build()
			defer mock.UnPatch()

			resp, err := profileLogic.Profile()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "账户已被禁用")
			So(resp, ShouldBeNil)
		})

		Convey("Should return error when user is locked", func() {
			userID := primitive.NewObjectID()
			ctx := context.WithValue(context.Background(), "uid", userID.Hex())
			profileLogic := NewProfileLogic(ctx, svcCtx)

			// 创建被锁定的用户
			lockUntil := time.Now().Add(30 * time.Minute)
			testUser := &model.User{
				ID:          userID,
				Username:    "testuser",
				Email:       "test@example.com",
				DisplayName: "Test User",
				Role:        constants.UserRoleAuthor,
				Status:      constants.UserStatusActive,
				LockedUntil: &lockUntil,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			// Mock UserDAO.GetByID to return locked user
			mock := mockey.Mock((*dao.UserDAO).GetByID).Return(testUser, nil).Build()
			defer mock.UnPatch()

			resp, err := profileLogic.Profile()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContain, "账户已被锁定")
			So(resp, ShouldBeNil)
		})

		Convey("Should return user profile successfully", func() {
			userID := primitive.NewObjectID()
			ctx := context.WithValue(context.Background(), "uid", userID.Hex())
			profileLogic := NewProfileLogic(ctx, svcCtx)

			// 创建正常用户
			lastLoginAt := time.Now().Add(-1 * time.Hour)
			testUser := &model.User{
				ID:           userID,
				Username:     "testuser",
				Email:        "test@example.com",
				DisplayName:  "Test User",
				Role:         constants.UserRoleAuthor,
				Status:       constants.UserStatusActive,
				ProfileImage: "https://example.com/avatar.jpg",
				Bio:          "A test user",
				Location:     "Test City",
				Website:      "https://example.com",
				Twitter:      "@testuser",
				Facebook:     "testuser",
				LastLoginAt:  &lastLoginAt,
				CreatedAt:    time.Now().Add(-24 * time.Hour),
				UpdatedAt:    time.Now(),
			}

			// Mock UserDAO.GetByID to return valid user
			mock := mockey.Mock((*dao.UserDAO).GetByID).Return(testUser, nil).Build()
			defer mock.UnPatch()

			resp, err := profileLogic.Profile()
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Message, ShouldEqual, "获取用户信息成功")
			So(resp.Data, ShouldNotBeNil)
			So(resp.Data.ID, ShouldEqual, userID.Hex())
			So(resp.Data.Username, ShouldEqual, "testuser")
			So(resp.Data.Email, ShouldEqual, "test@example.com")
			So(resp.Data.DisplayName, ShouldEqual, "Test User")
			So(resp.Data.Role, ShouldEqual, constants.UserRoleAuthor)
			So(resp.Data.Status, ShouldEqual, constants.UserStatusActive)
			So(resp.Data.ProfileImage, ShouldEqual, "https://example.com/avatar.jpg")
			So(resp.Data.Bio, ShouldEqual, "A test user")
			So(resp.Data.Location, ShouldEqual, "Test City")
			So(resp.Data.Website, ShouldEqual, "https://example.com")
			So(resp.Data.Twitter, ShouldEqual, "@testuser")
			So(resp.Data.Facebook, ShouldEqual, "testuser")
			So(resp.Data.LastLoginAt, ShouldNotBeEmpty)
			So(resp.Data.CreatedAt, ShouldNotBeEmpty)
			So(resp.Data.UpdatedAt, ShouldNotBeEmpty)
		})
	})
}
