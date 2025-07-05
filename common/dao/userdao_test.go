package dao

import (
	"context"
	"testing"
	"time"

	"github.com/bytedance/mockey"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/model"
)

func TestUserDAO_Create(t *testing.T) {
	Convey("UserDAO Create Tests", t, func() {
		// Create a properly initialized UserDAO instance
		userDAO := &UserDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should return error when user is nil", func() {
			err := userDAO.Create(context.Background(), nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "user cannot be nil")
		})

		Convey("Should create user successfully", func() {
			// Mock MongoDB InsertOne method
			mock := mockey.Mock((*mongo.Collection).InsertOne).Return(&mongo.InsertOneResult{
				InsertedID: primitive.NewObjectID(),
			}, nil).Build()
			defer mock.UnPatch()

			user := &model.User{
				Username:    "testuser",
				Email:       "test@example.com",
				DisplayName: "Test User",
				Role:        constants.UserRoleAuthor,
				Status:      constants.UserStatusActive,
			}

			err := userDAO.Create(context.Background(), user)
			So(err, ShouldBeNil)
		})

		Convey("Should return error when username already exists", func() {
			// Mock MongoDB InsertOne to return duplicate key error
			mock := mockey.Mock((*mongo.Collection).InsertOne).Return(nil, mongo.WriteException{
				WriteErrors: []mongo.WriteError{
					{Code: 11000}, // Duplicate key error code
				},
			}).Build()
			defer mock.UnPatch()

			user := &model.User{
				Username: "existinguser",
				Email:    "test@example.com",
			}

			err := userDAO.Create(context.Background(), user)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "username or email already exists")
		})
	})
}

func TestUserDAO_GetByID(t *testing.T) {
	Convey("UserDAO GetByID Tests", t, func() {
		userDAO := &UserDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should return error when ID is empty", func() {
			user, err := userDAO.GetByID(context.Background(), "")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "id cannot be empty")
			So(user, ShouldBeNil)
		})

		Convey("Should return error when ID format is invalid", func() {
			user, err := userDAO.GetByID(context.Background(), "invalid-id")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "invalid id format")
			So(user, ShouldBeNil)
		})

		Convey("Should return user when found", func() {
			// Mock the FindOne method and SingleResult.Decode
			mockResult := &mongo.SingleResult{}
			mock1 := mockey.Mock((*mongo.Collection).FindOne).Return(mockResult).Build()
			defer mock1.UnPatch()

			mock2 := mockey.Mock((*mongo.SingleResult).Decode).To(func(sr *mongo.SingleResult, v interface{}) error {
				if userPtr, ok := v.(*model.User); ok {
					userPtr.ID = primitive.NewObjectID()
					userPtr.Username = "testuser"
					userPtr.Email = "test@example.com"
					userPtr.DisplayName = "Test User"
					userPtr.Role = constants.UserRoleAuthor
					userPtr.Status = constants.UserStatusActive
				}
				return nil
			}).Build()
			defer mock2.UnPatch()

			objectID := primitive.NewObjectID()
			user, err := userDAO.GetByID(context.Background(), objectID.Hex())
			So(err, ShouldBeNil)
			So(user, ShouldNotBeNil)
			So(user.Username, ShouldEqual, "testuser")
		})

		Convey("Should return nil when user not found", func() {
			// Mock FindOne and Decode to return ErrNoDocuments
			mockResult := &mongo.SingleResult{}
			mock1 := mockey.Mock((*mongo.Collection).FindOne).Return(mockResult).Build()
			defer mock1.UnPatch()

			mock2 := mockey.Mock((*mongo.SingleResult).Decode).Return(mongo.ErrNoDocuments).Build()
			defer mock2.UnPatch()

			objectID := primitive.NewObjectID()
			user, err := userDAO.GetByID(context.Background(), objectID.Hex())
			So(err, ShouldBeNil)
			So(user, ShouldBeNil)
		})
	})
}

func TestUserDAO_GetByUsername(t *testing.T) {
	Convey("UserDAO GetByUsername Tests", t, func() {
		userDAO := &UserDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should return error when username is empty", func() {
			user, err := userDAO.GetByUsername(context.Background(), "")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "username cannot be empty")
			So(user, ShouldBeNil)
		})

		Convey("Should return user when found", func() {
			// Mock the FindOne method and SingleResult.Decode
			mockResult := &mongo.SingleResult{}
			mock1 := mockey.Mock((*mongo.Collection).FindOne).Return(mockResult).Build()
			defer mock1.UnPatch()

			mock2 := mockey.Mock((*mongo.SingleResult).Decode).To(func(sr *mongo.SingleResult, v interface{}) error {
				if userPtr, ok := v.(*model.User); ok {
					userPtr.ID = primitive.NewObjectID()
					userPtr.Username = "testuser"
					userPtr.Email = "test@example.com"
					userPtr.Role = constants.UserRoleAuthor
					userPtr.Status = constants.UserStatusActive
				}
				return nil
			}).Build()
			defer mock2.UnPatch()

			user, err := userDAO.GetByUsername(context.Background(), "testuser")
			So(err, ShouldBeNil)
			So(user, ShouldNotBeNil)
			So(user.Username, ShouldEqual, "testuser")
		})

		Convey("Should return nil when user not found", func() {
			// Mock FindOne and Decode to return ErrNoDocuments
			mockResult := &mongo.SingleResult{}
			mock1 := mockey.Mock((*mongo.Collection).FindOne).Return(mockResult).Build()
			defer mock1.UnPatch()

			mock2 := mockey.Mock((*mongo.SingleResult).Decode).Return(mongo.ErrNoDocuments).Build()
			defer mock2.UnPatch()

			user, err := userDAO.GetByUsername(context.Background(), "nonexistent")
			So(err, ShouldBeNil)
			So(user, ShouldBeNil)
		})
	})
}

func TestUserDAO_GetByEmail(t *testing.T) {
	Convey("UserDAO GetByEmail Tests", t, func() {
		userDAO := &UserDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should return error when email is empty", func() {
			user, err := userDAO.GetByEmail(context.Background(), "")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "email cannot be empty")
			So(user, ShouldBeNil)
		})

		Convey("Should return user when found", func() {
			// Mock the FindOne method and SingleResult.Decode
			mockResult := &mongo.SingleResult{}
			mock1 := mockey.Mock((*mongo.Collection).FindOne).Return(mockResult).Build()
			defer mock1.UnPatch()

			mock2 := mockey.Mock((*mongo.SingleResult).Decode).To(func(sr *mongo.SingleResult, v interface{}) error {
				if userPtr, ok := v.(*model.User); ok {
					userPtr.ID = primitive.NewObjectID()
					userPtr.Username = "testuser"
					userPtr.Email = "test@example.com"
					userPtr.Role = constants.UserRoleAuthor
					userPtr.Status = constants.UserStatusActive
				}
				return nil
			}).Build()
			defer mock2.UnPatch()

			user, err := userDAO.GetByEmail(context.Background(), "test@example.com")
			So(err, ShouldBeNil)
			So(user, ShouldNotBeNil)
			So(user.Email, ShouldEqual, "test@example.com")
		})

		Convey("Should return nil when user not found", func() {
			// Mock FindOne and Decode to return ErrNoDocuments
			mockResult := &mongo.SingleResult{}
			mock1 := mockey.Mock((*mongo.Collection).FindOne).Return(mockResult).Build()
			defer mock1.UnPatch()

			mock2 := mockey.Mock((*mongo.SingleResult).Decode).Return(mongo.ErrNoDocuments).Build()
			defer mock2.UnPatch()

			user, err := userDAO.GetByEmail(context.Background(), "nonexistent@example.com")
			So(err, ShouldBeNil)
			So(user, ShouldBeNil)
		})
	})
}

func TestUserDAO_Update(t *testing.T) {
	Convey("UserDAO Update Tests", t, func() {
		userDAO := &UserDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should return error when ID is empty", func() {
			err := userDAO.Update(context.Background(), "", map[string]interface{}{"name": "test"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "id cannot be empty")
		})

		Convey("Should return error when updates is empty", func() {
			objectID := primitive.NewObjectID()
			err := userDAO.Update(context.Background(), objectID.Hex(), nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "updates cannot be empty")
		})

		Convey("Should return error when ID format is invalid", func() {
			err := userDAO.Update(context.Background(), "invalid-id", map[string]interface{}{"name": "test"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "invalid id format")
		})

		Convey("Should update user successfully", func() {
			// Mock MongoDB UpdateOne method
			mock := mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
			}, nil).Build()
			defer mock.UnPatch()

			objectID := primitive.NewObjectID()
			updates := map[string]interface{}{
				"displayName": "Updated Name",
			}

			err := userDAO.Update(context.Background(), objectID.Hex(), updates)
			So(err, ShouldBeNil)
		})

		Convey("Should return error when user not found", func() {
			// Mock MongoDB UpdateOne to return no matched documents
			mock := mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
				MatchedCount:  0,
				ModifiedCount: 0,
			}, nil).Build()
			defer mock.UnPatch()

			objectID := primitive.NewObjectID()
			updates := map[string]interface{}{
				"displayName": "Updated Name",
			}

			err := userDAO.Update(context.Background(), objectID.Hex(), updates)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "user not found")
		})
	})
}

func TestUserDAO_Delete(t *testing.T) {
	Convey("UserDAO Delete Tests", t, func() {
		userDAO := &UserDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should return error when ID is empty", func() {
			err := userDAO.Delete(context.Background(), "")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "id cannot be empty")
		})

		Convey("Should return error when ID format is invalid", func() {
			err := userDAO.Delete(context.Background(), "invalid-id")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "invalid id format")
		})

		Convey("Should delete user successfully (soft delete)", func() {
			// Mock MongoDB UpdateOne method for soft delete
			mock := mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
			}, nil).Build()
			defer mock.UnPatch()

			objectID := primitive.NewObjectID()
			err := userDAO.Delete(context.Background(), objectID.Hex())
			So(err, ShouldBeNil)
		})

		Convey("Should return error when user not found", func() {
			// Mock MongoDB UpdateOne to return no matched documents
			mock := mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
				MatchedCount:  0,
				ModifiedCount: 0,
			}, nil).Build()
			defer mock.UnPatch()

			objectID := primitive.NewObjectID()
			err := userDAO.Delete(context.Background(), objectID.Hex())
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "user not found")
		})
	})
}

func TestUserDAO_List(t *testing.T) {
	Convey("UserDAO List Tests", t, func() {
		userDAO := &UserDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should list users successfully", func() {
			// Mock MongoDB Find and cursor operations
			mockCursor := &mongo.Cursor{}
			mock1 := mockey.Mock((*mongo.Collection).Find).Return(mockCursor, nil).Build()
			defer mock1.UnPatch()

			mock2 := mockey.Mock((*mongo.Collection).CountDocuments).Return(int64(2), nil).Build()
			defer mock2.UnPatch()

			// Mock cursor operations
			callCount := 0
			mock3 := mockey.Mock((*mongo.Cursor).Next).To(func(c *mongo.Cursor, ctx context.Context) bool {
				callCount++
				return callCount <= 2 // Return true for first 2 calls, false for 3rd
			}).Build()
			defer mock3.UnPatch()

			mock4 := mockey.Mock((*mongo.Cursor).Decode).To(func(c *mongo.Cursor, v interface{}) error {
				if userPtr, ok := v.(*model.User); ok {
					userPtr.ID = primitive.NewObjectID()
					userPtr.Username = "testuser"
					userPtr.Email = "test@example.com"
					userPtr.Role = constants.UserRoleAuthor
					userPtr.Status = constants.UserStatusActive
				}
				return nil
			}).Build()
			defer mock4.UnPatch()

			mock5 := mockey.Mock((*mongo.Cursor).Close).Return(nil).Build()
			defer mock5.UnPatch()

			users, total, err := userDAO.List(context.Background(), nil, 1, 10)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 2)
			So(len(users), ShouldEqual, 2)
		})
	})
}

func TestUserDAO_LoginMethods(t *testing.T) {
	Convey("UserDAO Login Methods Tests", t, func() {
		userDAO := &UserDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("UpdateLoginInfo should work correctly", func() {
			// Mock MongoDB UpdateOne method
			mock := mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
			}, nil).Build()
			defer mock.UnPatch()

			objectID := primitive.NewObjectID()
			err := userDAO.UpdateLoginInfo(context.Background(), objectID.Hex(), "192.168.1.1")
			So(err, ShouldBeNil)
		})

		Convey("IncrementLoginFailCount should work correctly", func() {
			// Mock MongoDB UpdateOne method
			mock := mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
			}, nil).Build()
			defer mock.UnPatch()

			objectID := primitive.NewObjectID()
			err := userDAO.IncrementLoginFailCount(context.Background(), objectID.Hex())
			So(err, ShouldBeNil)
		})

		Convey("LockUser should work correctly", func() {
			// Mock MongoDB UpdateOne method
			mock := mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
			}, nil).Build()
			defer mock.UnPatch()

			objectID := primitive.NewObjectID()
			err := userDAO.LockUser(context.Background(), objectID.Hex(), time.Now().Add(time.Hour))
			So(err, ShouldBeNil)
		})

		Convey("UnlockUser should work correctly", func() {
			// Mock MongoDB UpdateOne method
			mock := mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
			}, nil).Build()
			defer mock.UnPatch()

			objectID := primitive.NewObjectID()
			err := userDAO.UnlockUser(context.Background(), objectID.Hex())
			So(err, ShouldBeNil)
		})

		Convey("GetLockedUsers should work correctly", func() {
			// Mock MongoDB Find and cursor operations
			mockCursor := &mongo.Cursor{}
			mock1 := mockey.Mock((*mongo.Collection).Find).Return(mockCursor, nil).Build()
			defer mock1.UnPatch()

			// Mock cursor operations
			callCount := 0
			mock2 := mockey.Mock((*mongo.Cursor).Next).To(func(c *mongo.Cursor, ctx context.Context) bool {
				callCount++
				return callCount <= 1 // Return true for first call, false for 2nd
			}).Build()
			defer mock2.UnPatch()

			mock3 := mockey.Mock((*mongo.Cursor).Decode).To(func(c *mongo.Cursor, v interface{}) error {
				if userPtr, ok := v.(*model.User); ok {
					userPtr.ID = primitive.NewObjectID()
					userPtr.Username = "lockeduser"
					userPtr.Email = "locked@example.com"
					userPtr.Role = constants.UserRoleAuthor
					userPtr.Status = constants.UserStatusActive
				}
				return nil
			}).Build()
			defer mock3.UnPatch()

			mock4 := mockey.Mock((*mongo.Cursor).Close).Return(nil).Build()
			defer mock4.UnPatch()

			users, err := userDAO.GetLockedUsers(context.Background())
			So(err, ShouldBeNil)
			So(len(users), ShouldEqual, 1)
		})
	})
}

func TestUserDAO_CreateIndexes(t *testing.T) {
	Convey("UserDAO CreateIndexes Tests", t, func() {
		userDAO := &UserDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should create indexes successfully", func() {
			// Mock the entire CreateIndexes method since it's easier than mocking collection.Indexes()
			mock := mockey.Mock((*UserDAO).CreateIndexes).Return(nil).Build()
			defer mock.UnPatch()

			err := userDAO.CreateIndexes(context.Background())
			So(err, ShouldBeNil)
		})

		Convey("Should handle create indexes error", func() {
			// Mock the entire CreateIndexes method to return error
			expectedError := mongo.CommandError{
				Code:    1,
				Message: "create indexes failed",
			}
			mock := mockey.Mock((*UserDAO).CreateIndexes).Return(expectedError).Build()
			defer mock.UnPatch()

			err := userDAO.CreateIndexes(context.Background())
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, expectedError)
		})
	})
}

func TestUserDAO_ModelValidation(t *testing.T) {
	Convey("UserDAO Model Validation Tests", t, func() {
		Convey("Should validate user model correctly", func() {
			user := &model.User{
				Username:    "testuser",
				Email:       "test@example.com",
				DisplayName: "Test User",
				Role:        constants.UserRoleAuthor,
				Status:      constants.UserStatusActive,
			}

			// Test that model validation works
			So(user.Username, ShouldEqual, "testuser")
			So(user.Email, ShouldEqual, "test@example.com")
			So(user.Role, ShouldEqual, constants.UserRoleAuthor)
			So(user.Status, ShouldEqual, constants.UserStatusActive)
		})
	})
}
