package logic

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bytedance/mockey"
	"github.com/go-redis/redis/v8"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/heimdall-api/admin-api/admin/internal/config"
	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/dao"
	"github.com/heimdall-api/common/model"
	"github.com/heimdall-api/common/utils"
)

func TestLoginLogic_Login(t *testing.T) {
	Convey("LoginLogic Login Tests", t, func() {
		// 创建测试用的ServiceContext
		cfg := config.Config{
			Auth: config.AuthConfig{
				AccessSecret:  "test-secret",
				AccessExpire:  3600,
				RefreshExpire: 7200,
			},
			Security: config.SecurityConfig{
				MaxLoginAttempts:     5,
				LoginLockoutDuration: 1800,
			},
			Cache: config.CacheConfig{
				LoginAttempts: config.CacheItem{
					Prefix: "login_attempts:",
					TTL:    1800,
				},
			},
		}

		svcCtx := &svc.ServiceContext{
			Config:      cfg,
			UserDAO:     &dao.UserDAO{},
			LoginLogDAO: &dao.LoginLogDAO{},
			Redis:       &redis.Client{},
		}

		loginLogic := NewLoginLogic(context.Background(), svcCtx)

		Convey("Should return error when request is nil", func() {
			resp, err := loginLogic.Login(nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "登录请求不能为空")
			So(resp, ShouldBeNil)
		})

		Convey("Should return error when username is empty", func() {
			req := &types.LoginRequest{
				Username: "",
				Password: "password123",
			}

			resp, err := loginLogic.Login(req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "用户名不能为空")
			So(resp, ShouldBeNil)
		})

		Convey("Should return error when password is empty", func() {
			req := &types.LoginRequest{
				Username: "testuser",
				Password: "",
			}

			resp, err := loginLogic.Login(req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "密码不能为空")
			So(resp, ShouldBeNil)
		})

		Convey("Should return error when user not found", func() {
			req := &types.LoginRequest{
				Username: "nonexistent",
				Password: "password123",
			}

			// Mock UserDAO.GetByUsername to return nil (user not found)
			mock := mockey.Mock((*dao.UserDAO).GetByUsername).Return(nil, nil).Build()
			defer mock.UnPatch()

			resp, err := loginLogic.Login(req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "用户名或密码错误")
			So(resp, ShouldBeNil)
		})

		Convey("Should return error when password is incorrect", func() {
			req := &types.LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			}

			// 创建测试用户
			testUser := &model.User{
				ID:             primitive.NewObjectID(),
				Username:       "testuser",
				PasswordHash:   "$2a$12$valid.hashed.password.here", // 假设的有效哈希密码
				Status:         constants.UserStatusActive,
				Role:           constants.UserRoleAuthor,
				LoginFailCount: 2,
			}

			// Mock UserDAO.GetByUsername to return test user
			mock1 := mockey.Mock((*dao.UserDAO).GetByUsername).Return(testUser, nil).Build()
			defer mock1.UnPatch()

			// Mock password verification to return false
			mock2 := mockey.Mock(utils.VerifyPassword).Return(false).Build()
			defer mock2.UnPatch()

			// Mock Redis operations for login attempts tracking
			mock3 := mockey.Mock((*redis.Client).Get).Return(redis.NewStringResult("2", nil)).Build()
			defer mock3.UnPatch()

			mock4 := mockey.Mock((*redis.Client).Incr).Return(redis.NewIntResult(3, nil)).Build()
			defer mock4.UnPatch()

			mock5 := mockey.Mock((*redis.Client).Expire).Return(redis.NewBoolResult(true, nil)).Build()
			defer mock5.UnPatch()

			// Mock UserDAO.IncrementLoginFailCount
			mock6 := mockey.Mock((*dao.UserDAO).IncrementLoginFailCount).Return(nil).Build()
			defer mock6.UnPatch()

			// Mock LoginLogDAO.Create for failed login log
			mock7 := mockey.Mock((*dao.LoginLogDAO).Create).Return(nil).Build()
			defer mock7.UnPatch()

			resp, err := loginLogic.Login(req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "用户名或密码错误")
			So(resp, ShouldBeNil)
		})

		Convey("Should lock account when max login attempts exceeded", func() {
			req := &types.LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			}

			// 创建测试用户，已经有4次失败
			testUser := &model.User{
				ID:             primitive.NewObjectID(),
				Username:       "testuser",
				PasswordHash:   "$2a$12$valid.hashed.password.here",
				Status:         constants.UserStatusActive,
				Role:           constants.UserRoleAuthor,
				LoginFailCount: 4,
			}

			// Mock UserDAO.GetByUsername
			mock1 := mockey.Mock((*dao.UserDAO).GetByUsername).Return(testUser, nil).Build()
			defer mock1.UnPatch()

			// Mock password verification to return false
			mock2 := mockey.Mock(utils.VerifyPassword).Return(false).Build()
			defer mock2.UnPatch()

			// Mock Redis operations - 这是第5次尝试
			mock3 := mockey.Mock((*redis.Client).Get).Return(redis.NewStringResult("4", nil)).Build()
			defer mock3.UnPatch()

			mock4 := mockey.Mock((*redis.Client).Incr).Return(redis.NewIntResult(5, nil)).Build()
			defer mock4.UnPatch()

			mock5 := mockey.Mock((*redis.Client).Expire).Return(redis.NewBoolResult(true, nil)).Build()
			defer mock5.UnPatch()

			// Mock UserDAO methods for account locking
			mock6 := mockey.Mock((*dao.UserDAO).IncrementLoginFailCount).Return(nil).Build()
			defer mock6.UnPatch()

			mock7 := mockey.Mock((*dao.UserDAO).LockUser).Return(nil).Build()
			defer mock7.UnPatch()

			// Mock LoginLogDAO.Create
			mock8 := mockey.Mock((*dao.LoginLogDAO).Create).Return(nil).Build()
			defer mock8.UnPatch()

			resp, err := loginLogic.Login(req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "登录失败次数过多，账户已被锁定30分钟")
			So(resp, ShouldBeNil)
		})

		Convey("Should return error when user account is locked", func() {
			req := &types.LoginRequest{
				Username: "lockeduser",
				Password: "password123",
			}

			// 创建已锁定的用户
			lockUntil := time.Now().Add(20 * time.Minute) // 还有20分钟解锁
			testUser := &model.User{
				ID:           primitive.NewObjectID(),
				Username:     "lockeduser",
				PasswordHash: "$2a$12$valid.hashed.password.here",
				Status:       constants.UserStatusActive,
				Role:         constants.UserRoleAuthor,
				LockedUntil:  &lockUntil,
			}

			// Mock UserDAO.GetByUsername
			mock := mockey.Mock((*dao.UserDAO).GetByUsername).Return(testUser, nil).Build()
			defer mock.UnPatch()

			resp, err := loginLogic.Login(req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContain, "账户已被锁定")
			So(resp, ShouldBeNil)
		})

		Convey("Should return error when user status is inactive", func() {
			req := &types.LoginRequest{
				Username: "inactiveuser",
				Password: "password123",
			}

			// 创建非活跃用户
			testUser := &model.User{
				ID:           primitive.NewObjectID(),
				Username:     "inactiveuser",
				PasswordHash: "$2a$12$valid.hashed.password.here",
				Status:       constants.UserStatusInactive,
				Role:         constants.UserRoleAuthor,
			}

			// Mock UserDAO.GetByUsername
			mock := mockey.Mock((*dao.UserDAO).GetByUsername).Return(testUser, nil).Build()
			defer mock.UnPatch()

			resp, err := loginLogic.Login(req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "账户已被禁用")
			So(resp, ShouldBeNil)
		})

		Convey("Should login successfully with correct credentials", func() {
			req := &types.LoginRequest{
				Username:   "testuser",
				Password:   "password123",
				RememberMe: true,
			}

			// 创建有效用户
			testUser := &model.User{
				ID:             primitive.NewObjectID(),
				Username:       "testuser",
				Email:          "test@example.com",
				DisplayName:    "Test User",
				PasswordHash:   "$2a$12$valid.hashed.password.here",
				Status:         constants.UserStatusActive,
				Role:           constants.UserRoleAdmin,
				LoginFailCount: 2, // 之前有失败记录，成功后应该重置
			}

			// Mock UserDAO.GetByUsername
			mock1 := mockey.Mock((*dao.UserDAO).GetByUsername).Return(testUser, nil).Build()
			defer mock1.UnPatch()

			// Mock password verification to return true
			mock2 := mockey.Mock(utils.VerifyPassword).Return(true).Build()
			defer mock2.UnPatch()

			// Mock JWT generation
			tokenPair := &utils.TokenPair{
				AccessToken:  "jwt.token.here",
				RefreshToken: "refresh.token.here",
				ExpiresAt:    time.Now().Add(time.Hour),
				TokenType:    "Bearer",
			}
			mock3 := mockey.Mock((*utils.JWTManager).GenerateToken).Return(tokenPair, nil).Build()
			defer mock3.UnPatch()

			// Mock Redis operations to clear login attempts
			mock4 := mockey.Mock((*redis.Client).Del).Return(redis.NewIntResult(1, nil)).Build()
			defer mock4.UnPatch()

			// Mock UserDAO operations for successful login
			mock5 := mockey.Mock((*dao.UserDAO).UpdateLoginInfo).Return(nil).Build()
			defer mock5.UnPatch()

			// Mock LoginLogDAO.Create for successful login log
			mock6 := mockey.Mock((*dao.LoginLogDAO).Create).Return(nil).Build()
			defer mock6.UnPatch()

			resp, err := loginLogic.Login(req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Message, ShouldEqual, "登录成功")
			So(resp.Data, ShouldNotBeNil)
			So(resp.Data.Token, ShouldEqual, "jwt.token.here")
			So(resp.Data.User, ShouldNotBeNil)
			So(resp.Data.User.Username, ShouldEqual, "testuser")
			So(resp.Data.User.Email, ShouldEqual, "test@example.com")
			So(resp.Data.User.Role, ShouldEqual, constants.UserRoleAdmin)
		})

		Convey("Should handle database errors gracefully", func() {
			req := &types.LoginRequest{
				Username: "testuser",
				Password: "password123",
			}

			// Mock UserDAO.GetByUsername to return database error
			mock := mockey.Mock((*dao.UserDAO).GetByUsername).Return(nil, errors.New("database connection failed")).Build()
			defer mock.UnPatch()

			resp, err := loginLogic.Login(req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "系统错误，请稍后重试")
			So(resp, ShouldBeNil)
		})

		Convey("Should handle JWT generation errors", func() {
			req := &types.LoginRequest{
				Username: "testuser",
				Password: "password123",
			}

			// 创建有效用户
			testUser := &model.User{
				ID:           primitive.NewObjectID(),
				Username:     "testuser",
				PasswordHash: "$2a$12$valid.hashed.password.here",
				Status:       constants.UserStatusActive,
				Role:         constants.UserRoleAuthor,
			}

			// Mock UserDAO.GetByUsername
			mock1 := mockey.Mock((*dao.UserDAO).GetByUsername).Return(testUser, nil).Build()
			defer mock1.UnPatch()

			// Mock password verification to return true
			mock2 := mockey.Mock(utils.VerifyPassword).Return(true).Build()
			defer mock2.UnPatch()

			// Mock JWT generation to return error
			mock3 := mockey.Mock((*utils.JWTManager).GenerateToken).Return(nil, errors.New("JWT generation failed")).Build()
			defer mock3.UnPatch()

			resp, err := loginLogic.Login(req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "系统错误，请稍后重试")
			So(resp, ShouldBeNil)
		})
	})
}
