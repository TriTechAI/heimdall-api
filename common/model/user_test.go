package model

import (
	"testing"
	"time"

	"github.com/heimdall-api/common/constants"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUserModel(t *testing.T) {
	Convey("用户模型测试", t, func() {

		Convey("用户创建验证", func() {
			Convey("有效的用户数据应该通过验证", func() {
				user := &User{
					Username:    "testuser",
					Email:       "test@example.com",
					DisplayName: "Test User",
					Role:        constants.UserRoleAuthor,
					Status:      constants.UserStatusActive,
				}

				err := user.ValidateForCreate()
				So(err, ShouldBeNil)
			})

			Convey("缺少必填字段应该验证失败", func() {
				user := &User{}

				err := user.ValidateForCreate()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "用户名不能为空")
			})

			Convey("用户名长度不足应该验证失败", func() {
				user := &User{
					Username:    "ab", // 少于3字符
					Email:       "test@example.com",
					DisplayName: "Test User",
					Role:        constants.UserRoleAuthor,
				}

				err := user.ValidateForCreate()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "用户名长度不能少于3字符")
			})

			Convey("用户名过长应该验证失败", func() {
				longUsername := "a"
				for i := 0; i < 35; i++ {
					longUsername += "a"
				}

				user := &User{
					Username:    longUsername,
					Email:       "test@example.com",
					DisplayName: "Test User",
					Role:        constants.UserRoleAuthor,
				}

				err := user.ValidateForCreate()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "用户名长度不能超过32字符")
			})

			Convey("无效的用户角色应该验证失败", func() {
				user := &User{
					Username:    "testuser",
					Email:       "test@example.com",
					DisplayName: "Test User",
					Role:        "invalid_role",
				}

				err := user.ValidateForCreate()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "无效的用户角色")
			})

			Convey("其他字段长度超限应该验证失败", func() {
				user := &User{
					Username:    "testuser",
					Email:       "test@example.com",
					DisplayName: "Test User",
					Role:        constants.UserRoleAuthor,
					Bio:         getLongString(501), // 超过500字符
				}

				err := user.ValidateForCreate()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "简介长度不能超过500字符")
			})
		})

		Convey("用户更新验证", func() {
			Convey("有效的更新数据应该通过验证", func() {
				user := &User{
					DisplayName: "Updated Name",
					Bio:         "Updated bio",
				}

				err := user.ValidateForUpdate()
				So(err, ShouldBeNil)
			})

			Convey("更新数据过长应该验证失败", func() {
				user := &User{
					DisplayName: getLongString(65), // 超过64字符
				}

				err := user.ValidateForUpdate()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "显示名长度不能超过64字符")
			})
		})

		Convey("用户状态检查", func() {
			user := &User{
				Status: constants.UserStatusActive,
				Role:   constants.UserRoleEditor,
			}

			Convey("活跃用户状态检查", func() {
				So(user.IsActive(), ShouldBeTrue)
				So(user.CanLogin(), ShouldBeTrue)
			})

			Convey("锁定用户状态检查", func() {
				user.Status = constants.UserStatusLocked
				So(user.IsLocked(), ShouldBeTrue)
				So(user.CanLogin(), ShouldBeFalse)
			})

			Convey("临时锁定检查", func() {
				user.Status = constants.UserStatusActive
				future := time.Now().Add(time.Hour)
				user.LockedUntil = &future

				So(user.IsLocked(), ShouldBeTrue)
				So(user.CanLogin(), ShouldBeFalse)
			})

			Convey("过期锁定检查", func() {
				user.Status = constants.UserStatusActive
				past := time.Now().Add(-time.Hour)
				user.LockedUntil = &past

				So(user.IsLocked(), ShouldBeFalse)
				So(user.CanLogin(), ShouldBeTrue)
			})
		})

		Convey("用户权限检查", func() {
			Convey("所有者权限检查", func() {
				user := &User{Role: constants.UserRoleOwner}
				So(user.IsOwner(), ShouldBeTrue)
				So(user.IsAdmin(), ShouldBeTrue)
				So(user.IsEditor(), ShouldBeTrue)
				So(user.CanManageUser(), ShouldBeTrue)
				So(user.CanManageAllPosts(), ShouldBeTrue)
				So(user.CanManageComments(), ShouldBeTrue)
			})

			Convey("管理员权限检查", func() {
				user := &User{Role: constants.UserRoleAdmin}
				So(user.IsOwner(), ShouldBeFalse)
				So(user.IsAdmin(), ShouldBeTrue)
				So(user.IsEditor(), ShouldBeTrue)
				So(user.CanManageUser(), ShouldBeTrue)
				So(user.CanManageAllPosts(), ShouldBeTrue)
				So(user.CanManageComments(), ShouldBeTrue)
			})

			Convey("编辑权限检查", func() {
				user := &User{Role: constants.UserRoleEditor}
				So(user.IsOwner(), ShouldBeFalse)
				So(user.IsAdmin(), ShouldBeFalse)
				So(user.IsEditor(), ShouldBeTrue)
				So(user.CanManageUser(), ShouldBeFalse)
				So(user.CanManageAllPosts(), ShouldBeTrue)
				So(user.CanManageComments(), ShouldBeTrue)
			})

			Convey("作者权限检查", func() {
				user := &User{Role: constants.UserRoleAuthor}
				So(user.IsOwner(), ShouldBeFalse)
				So(user.IsAdmin(), ShouldBeFalse)
				So(user.IsEditor(), ShouldBeFalse)
				So(user.CanManageUser(), ShouldBeFalse)
				So(user.CanManageAllPosts(), ShouldBeFalse)
				So(user.CanManageComments(), ShouldBeFalse)
			})
		})

		Convey("用户转换方法", func() {
			user := createTestUser()

			Convey("转换为用户档案响应", func() {
				profile := user.ToProfileResponse()
				So(profile, ShouldNotBeNil)
				So(profile.ID, ShouldEqual, user.ID.Hex())
				So(profile.Username, ShouldEqual, user.Username)
				So(profile.Email, ShouldEqual, user.Email)
				So(profile.DisplayName, ShouldEqual, user.DisplayName)
				So(profile.Role, ShouldEqual, user.Role)
			})

			Convey("转换为用户列表项", func() {
				listItem := user.ToListItem()
				So(listItem, ShouldNotBeNil)
				So(listItem.ID, ShouldEqual, user.ID.Hex())
				So(listItem.Username, ShouldEqual, user.Username)
				So(listItem.Role, ShouldEqual, user.Role)
			})

			Convey("转换为作者信息", func() {
				authorInfo := user.ToAuthorInfo()
				So(authorInfo, ShouldNotBeNil)
				So(authorInfo.ID, ShouldEqual, user.ID.Hex())
				So(authorInfo.Username, ShouldEqual, user.Username)
				So(authorInfo.DisplayName, ShouldEqual, user.DisplayName)
			})
		})

		Convey("用户工厂方法", func() {
			Convey("创建新用户", func() {
				user := NewUser("testuser", "test@example.com", "hashedpassword", "Test User", constants.UserRoleAuthor)
				So(user, ShouldNotBeNil)
				So(user.Username, ShouldEqual, "testuser")
				So(user.Email, ShouldEqual, "test@example.com")
				So(user.PasswordHash, ShouldEqual, "hashedpassword")
				So(user.DisplayName, ShouldEqual, "Test User")
				So(user.Role, ShouldEqual, constants.UserRoleAuthor)
				So(user.Status, ShouldEqual, constants.UserStatusActive)
				So(user.LoginFailCount, ShouldEqual, 0)
				So(user.ID.IsZero(), ShouldBeFalse)
			})

			Convey("从创建请求创建用户", func() {
				req := &UserCreateRequest{
					Username:    "testuser",
					Email:       "test@example.com",
					Password:    "password123",
					DisplayName: "Test User",
					Role:        constants.UserRoleAuthor,
					Bio:         "Test bio",
					Location:    "Test location",
				}

				user := NewUserFromCreateRequest(req, "hashedpassword")
				So(user, ShouldNotBeNil)
				So(user.Username, ShouldEqual, req.Username)
				So(user.Email, ShouldEqual, req.Email)
				So(user.DisplayName, ShouldEqual, req.DisplayName)
				So(user.Role, ShouldEqual, req.Role)
				So(user.Bio, ShouldEqual, req.Bio)
				So(user.Location, ShouldEqual, req.Location)
			})
		})

		Convey("数据库操作辅助方法", func() {
			user := createTestUser()

			Convey("准备插入数据库", func() {
				user.ID = primitive.ObjectID{} // 清空ID
				originalTime := user.CreatedAt

				user.PrepareForInsert()

				So(user.ID.IsZero(), ShouldBeFalse)
				So(user.CreatedAt.After(originalTime), ShouldBeTrue)
				So(user.UpdatedAt.After(originalTime), ShouldBeTrue)
				So(user.Status, ShouldEqual, constants.UserStatusActive)
			})

			Convey("准备更新数据库", func() {
				originalTime := user.UpdatedAt
				time.Sleep(time.Millisecond) // 确保时间差异

				user.PrepareForUpdate()

				So(user.UpdatedAt.After(originalTime), ShouldBeTrue)
			})

			Convey("增加登录失败次数", func() {
				originalCount := user.LoginFailCount
				originalTime := user.UpdatedAt

				user.IncrementLoginFailCount()

				So(user.LoginFailCount, ShouldEqual, originalCount+1)
				So(user.UpdatedAt.After(originalTime), ShouldBeTrue)
			})

			Convey("多次失败导致锁定", func() {
				user.LoginFailCount = 2 // 设置为2次失败

				user.IncrementLoginFailCount() // 第3次失败

				So(user.LoginFailCount, ShouldEqual, 3)
				So(user.LockedUntil, ShouldNotBeNil)
				So(user.Status, ShouldEqual, constants.UserStatusLocked)
			})

			Convey("重置登录失败次数", func() {
				user.LoginFailCount = 5
				user.Status = constants.UserStatusLocked
				future := time.Now().Add(time.Hour)
				user.LockedUntil = &future

				user.ResetLoginFailCount()

				So(user.LoginFailCount, ShouldEqual, 0)
				So(user.LockedUntil, ShouldBeNil)
				So(user.Status, ShouldEqual, constants.UserStatusActive)
			})

			Convey("更新最后登录信息", func() {
				user.LoginFailCount = 3 // 设置一些失败次数
				testIP := "192.168.1.1"

				user.UpdateLastLogin(testIP)

				So(user.LastLoginAt, ShouldNotBeNil)
				So(user.LastLoginIP, ShouldEqual, testIP)
				So(user.LoginFailCount, ShouldEqual, 0) // 应该被重置
			})
		})

		Convey("验证错误", func() {
			Convey("创建验证错误", func() {
				err := NewValidationError("field", "message")
				So(err, ShouldNotBeNil)
				So(err.Field, ShouldEqual, "field")
				So(err.Message, ShouldEqual, "message")
				So(err.Error(), ShouldEqual, "message")
			})
		})
	})
}

// 创建测试用户
func createTestUser() *User {
	return &User{
		ID:           primitive.NewObjectID(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		DisplayName:  "Test User",
		Role:         constants.UserRoleEditor,
		Status:       constants.UserStatusActive,
		ProfileImage: "https://example.com/profile.jpg",
		CoverImage:   "https://example.com/cover.jpg",
		Bio:          "Test user bio",
		Location:     "Test City",
		Website:      "https://example.com",
		Twitter:      "@testuser",
		Facebook:     "testuser",
		CreatedAt:    time.Now().Add(-time.Hour),
		UpdatedAt:    time.Now().Add(-time.Minute),
	}
}

// 生成指定长度的字符串
func getLongString(length int) string {
	result := ""
	for i := 0; i < length; i++ {
		result += "a"
	}
	return result
}
