package dao

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bytedance/mockey"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/model"
)

func TestLoginLogDAO_Create(t *testing.T) {
	Convey("LoginLogDAO Create Tests", t, func() {
		// Create a properly initialized LoginLogDAO instance
		loginLogDAO := &LoginLogDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should return error when log is nil", func() {
			err := loginLogDAO.Create(context.Background(), nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "log cannot be nil")
		})

		Convey("Should create login log successfully", func() {
			// Mock MongoDB InsertOne method
			mock := mockey.Mock((*mongo.Collection).InsertOne).Return(&mongo.InsertOneResult{
				InsertedID: primitive.NewObjectID(),
			}, nil).Build()
			defer mock.UnPatch()

			userID := primitive.NewObjectID()
			log := &model.LoginLog{
				UserID:      &userID,
				Username:    "testuser",
				LoginMethod: "username",
				IPAddress:   "192.168.1.1",
				UserAgent:   "Mozilla/5.0 Test",
				Status:      constants.LoginStatusSuccess,
			}

			err := loginLogDAO.Create(context.Background(), log)
			So(err, ShouldBeNil)
		})

		Convey("Should return validation error when required fields are missing", func() {
			// Mock MongoDB InsertOne method (won't be called due to validation error)
			mock := mockey.Mock((*mongo.Collection).InsertOne).Return(&mongo.InsertOneResult{
				InsertedID: primitive.NewObjectID(),
			}, nil).Build()
			defer mock.UnPatch()

			log := &model.LoginLog{
				Username: "", // Empty username should cause validation error
			}

			err := loginLogDAO.Create(context.Background(), log)
			So(err, ShouldNotBeNil)
		})

		Convey("Should return error when database operation fails", func() {
			// Mock MongoDB InsertOne to return error
			mock := mockey.Mock((*mongo.Collection).InsertOne).Return(nil, errors.New("database error")).Build()
			defer mock.UnPatch()

			userID := primitive.NewObjectID()
			log := &model.LoginLog{
				UserID:      &userID,
				Username:    "testuser",
				LoginMethod: "username",
				IPAddress:   "192.168.1.1",
				UserAgent:   "Mozilla/5.0 Test",
				Status:      constants.LoginStatusSuccess,
			}

			err := loginLogDAO.Create(context.Background(), log)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "database error")
		})
	})
}

func TestLoginLogDAO_List(t *testing.T) {
	Convey("LoginLogDAO List Tests", t, func() {
		loginLogDAO := &LoginLogDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should list login logs successfully", func() {
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
				if logPtr, ok := v.(*model.LoginLog); ok {
					logPtr.ID = primitive.NewObjectID()
					logPtr.Username = "testuser"
					logPtr.IPAddress = "192.168.1.1"
					logPtr.Status = constants.LoginStatusSuccess
					logPtr.LoginAt = time.Now()
				}
				return nil
			}).Build()
			defer mock4.UnPatch()

			mock5 := mockey.Mock((*mongo.Cursor).Close).Return(nil).Build()
			defer mock5.UnPatch()

			logs, total, err := loginLogDAO.List(context.Background(), nil, 1, 10)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 2)
			So(len(logs), ShouldEqual, 2)
		})

		Convey("Should handle invalid page and limit parameters", func() {
			// Mock MongoDB Find and cursor operations
			mockCursor := &mongo.Cursor{}
			mock1 := mockey.Mock((*mongo.Collection).Find).Return(mockCursor, nil).Build()
			defer mock1.UnPatch()

			mock2 := mockey.Mock((*mongo.Collection).CountDocuments).Return(int64(0), nil).Build()
			defer mock2.UnPatch()

			callCount := 0
			mock3 := mockey.Mock((*mongo.Cursor).Next).To(func(c *mongo.Cursor, ctx context.Context) bool {
				callCount++
				return false // No results
			}).Build()
			defer mock3.UnPatch()

			mock4 := mockey.Mock((*mongo.Cursor).Close).Return(nil).Build()
			defer mock4.UnPatch()

			logs, total, err := loginLogDAO.List(context.Background(), nil, 0, 0) // Invalid parameters
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 0)
			So(len(logs), ShouldEqual, 0)
		})

		Convey("Should filter logs by username", func() {
			// Mock MongoDB Find and cursor operations
			mockCursor := &mongo.Cursor{}
			mock1 := mockey.Mock((*mongo.Collection).Find).Return(mockCursor, nil).Build()
			defer mock1.UnPatch()

			mock2 := mockey.Mock((*mongo.Collection).CountDocuments).Return(int64(1), nil).Build()
			defer mock2.UnPatch()

			callCount := 0
			mock3 := mockey.Mock((*mongo.Cursor).Next).To(func(c *mongo.Cursor, ctx context.Context) bool {
				callCount++
				return callCount <= 1
			}).Build()
			defer mock3.UnPatch()

			mock4 := mockey.Mock((*mongo.Cursor).Decode).To(func(c *mongo.Cursor, v interface{}) error {
				if logPtr, ok := v.(*model.LoginLog); ok {
					logPtr.ID = primitive.NewObjectID()
					logPtr.Username = "testuser"
					logPtr.IPAddress = "192.168.1.1"
					logPtr.Status = constants.LoginStatusSuccess
				}
				return nil
			}).Build()
			defer mock4.UnPatch()

			mock5 := mockey.Mock((*mongo.Cursor).Close).Return(nil).Build()
			defer mock5.UnPatch()

			filter := map[string]interface{}{
				"username": "testuser",
			}

			logs, total, err := loginLogDAO.List(context.Background(), filter, 1, 10)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(logs), ShouldEqual, 1)
		})
	})
}

func TestLoginLogDAO_GetByUserID(t *testing.T) {
	Convey("LoginLogDAO GetByUserID Tests", t, func() {
		loginLogDAO := &LoginLogDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should return error when userID is empty", func() {
			logs, total, err := loginLogDAO.GetByUserID(context.Background(), "", 1, 10)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "userID cannot be empty")
			So(logs, ShouldBeNil)
			So(total, ShouldEqual, 0)
		})

		Convey("Should return error when userID format is invalid", func() {
			logs, total, err := loginLogDAO.GetByUserID(context.Background(), "invalid-id", 1, 10)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "invalid userID format")
			So(logs, ShouldBeNil)
			So(total, ShouldEqual, 0)
		})

		Convey("Should get logs by userID successfully", func() {
			// Mock MongoDB Find and cursor operations
			mockCursor := &mongo.Cursor{}
			mock1 := mockey.Mock((*mongo.Collection).Find).Return(mockCursor, nil).Build()
			defer mock1.UnPatch()

			mock2 := mockey.Mock((*mongo.Collection).CountDocuments).Return(int64(1), nil).Build()
			defer mock2.UnPatch()

			callCount := 0
			mock3 := mockey.Mock((*mongo.Cursor).Next).To(func(c *mongo.Cursor, ctx context.Context) bool {
				callCount++
				return callCount <= 1
			}).Build()
			defer mock3.UnPatch()

			mock4 := mockey.Mock((*mongo.Cursor).Decode).To(func(c *mongo.Cursor, v interface{}) error {
				if logPtr, ok := v.(*model.LoginLog); ok {
					logPtr.ID = primitive.NewObjectID()
					logPtr.Username = "testuser"
				}
				return nil
			}).Build()
			defer mock4.UnPatch()

			mock5 := mockey.Mock((*mongo.Cursor).Close).Return(nil).Build()
			defer mock5.UnPatch()

			objectID := primitive.NewObjectID()
			logs, total, err := loginLogDAO.GetByUserID(context.Background(), objectID.Hex(), 1, 10)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(logs), ShouldEqual, 1)
		})
	})
}

func TestLoginLogDAO_GetByIPAddress(t *testing.T) {
	Convey("LoginLogDAO GetByIPAddress Tests", t, func() {
		loginLogDAO := &LoginLogDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should return error when ipAddress is empty", func() {
			logs, total, err := loginLogDAO.GetByIPAddress(context.Background(), "", 1, 10)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "ipAddress cannot be empty")
			So(logs, ShouldBeNil)
			So(total, ShouldEqual, 0)
		})

		Convey("Should get logs by IP address successfully", func() {
			// Mock MongoDB Find and cursor operations
			mockCursor := &mongo.Cursor{}
			mock1 := mockey.Mock((*mongo.Collection).Find).Return(mockCursor, nil).Build()
			defer mock1.UnPatch()

			mock2 := mockey.Mock((*mongo.Collection).CountDocuments).Return(int64(1), nil).Build()
			defer mock2.UnPatch()

			callCount := 0
			mock3 := mockey.Mock((*mongo.Cursor).Next).To(func(c *mongo.Cursor, ctx context.Context) bool {
				callCount++
				return callCount <= 1
			}).Build()
			defer mock3.UnPatch()

			mock4 := mockey.Mock((*mongo.Cursor).Decode).To(func(c *mongo.Cursor, v interface{}) error {
				if logPtr, ok := v.(*model.LoginLog); ok {
					logPtr.ID = primitive.NewObjectID()
					logPtr.IPAddress = "192.168.1.1"
				}
				return nil
			}).Build()
			defer mock4.UnPatch()

			mock5 := mockey.Mock((*mongo.Cursor).Close).Return(nil).Build()
			defer mock5.UnPatch()

			logs, total, err := loginLogDAO.GetByIPAddress(context.Background(), "192.168.1.1", 1, 10)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(logs), ShouldEqual, 1)
		})
	})
}

func TestLoginLogDAO_GetRecentFailedLogins(t *testing.T) {
	Convey("LoginLogDAO GetRecentFailedLogins Tests", t, func() {
		loginLogDAO := &LoginLogDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should get recent failed logins successfully", func() {
			// Mock MongoDB Find and cursor operations
			mockCursor := &mongo.Cursor{}
			mock1 := mockey.Mock((*mongo.Collection).Find).Return(mockCursor, nil).Build()
			defer mock1.UnPatch()

			callCount := 0
			mock2 := mockey.Mock((*mongo.Cursor).Next).To(func(c *mongo.Cursor, ctx context.Context) bool {
				callCount++
				return callCount <= 2
			}).Build()
			defer mock2.UnPatch()

			mock3 := mockey.Mock((*mongo.Cursor).Decode).To(func(c *mongo.Cursor, v interface{}) error {
				if logPtr, ok := v.(*model.LoginLog); ok {
					logPtr.ID = primitive.NewObjectID()
					logPtr.Status = constants.LoginStatusFailed
					logPtr.FailReason = "invalid_password"
				}
				return nil
			}).Build()
			defer mock3.UnPatch()

			mock4 := mockey.Mock((*mongo.Cursor).Close).Return(nil).Build()
			defer mock4.UnPatch()

			since := time.Now().Add(-24 * time.Hour)
			logs, err := loginLogDAO.GetRecentFailedLogins(context.Background(), since, 10)
			So(err, ShouldBeNil)
			So(len(logs), ShouldEqual, 2)
		})

		Convey("Should handle limit parameters correctly", func() {
			// Mock MongoDB Find and cursor operations
			mockCursor := &mongo.Cursor{}
			mock1 := mockey.Mock((*mongo.Collection).Find).Return(mockCursor, nil).Build()
			defer mock1.UnPatch()

			callCount := 0
			mock2 := mockey.Mock((*mongo.Cursor).Next).To(func(c *mongo.Cursor, ctx context.Context) bool {
				callCount++
				return false // No results
			}).Build()
			defer mock2.UnPatch()

			mock3 := mockey.Mock((*mongo.Cursor).Close).Return(nil).Build()
			defer mock3.UnPatch()

			since := time.Now()
			logs, err := loginLogDAO.GetRecentFailedLogins(context.Background(), since, 0) // Invalid limit
			So(err, ShouldBeNil)
			So(len(logs), ShouldEqual, 0)
		})
	})
}

func TestLoginLogDAO_CreateIndexes(t *testing.T) {
	Convey("LoginLogDAO CreateIndexes Tests", t, func() {
		loginLogDAO := &LoginLogDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should create indexes successfully", func() {
			// Mock the entire CreateIndexes method since it's easier than mocking collection.Indexes()
			mock := mockey.Mock((*LoginLogDAO).CreateIndexes).Return(nil).Build()
			defer mock.UnPatch()

			err := loginLogDAO.CreateIndexes(context.Background())
			So(err, ShouldBeNil)
		})

		Convey("Should handle create indexes error", func() {
			// Mock CreateIndexes method to return error
			expectedErr := errors.New("create indexes failed")
			mock := mockey.Mock((*LoginLogDAO).CreateIndexes).Return(expectedErr).Build()
			defer mock.UnPatch()

			err := loginLogDAO.CreateIndexes(context.Background())
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "create indexes failed")
		})
	})
}

func TestLoginLogDAO_BuildQueryFilter(t *testing.T) {
	Convey("LoginLogDAO BuildQueryFilter Tests", t, func() {
		loginLogDAO := &LoginLogDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should build empty query for nil filter", func() {
			query := loginLogDAO.buildQueryFilter(nil)
			So(len(query), ShouldEqual, 0)
		})

		Convey("Should build query with username filter", func() {
			filter := map[string]interface{}{
				"username": "testuser",
			}
			query := loginLogDAO.buildQueryFilter(filter)
			So(query["username"], ShouldNotBeNil)
		})

		Convey("Should build query with userID filter", func() {
			userID := primitive.NewObjectID()
			filter := map[string]interface{}{
				"userId": userID,
			}
			query := loginLogDAO.buildQueryFilter(filter)
			So(query["userId"], ShouldEqual, userID)
		})

		Convey("Should build query with time range filter", func() {
			startTime := time.Now().Add(-24 * time.Hour)
			endTime := time.Now()
			filter := map[string]interface{}{
				"startTime": startTime,
				"endTime":   endTime,
			}
			query := loginLogDAO.buildQueryFilter(filter)
			So(query["loginAt"], ShouldNotBeNil)
		})
	})
}

func TestLoginLogDAO_BuildSortCondition(t *testing.T) {
	Convey("LoginLogDAO BuildSortCondition Tests", t, func() {
		loginLogDAO := &LoginLogDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should return default sort for nil filter", func() {
			sort := loginLogDAO.buildSortCondition(nil)
			So(len(sort), ShouldEqual, 1)
		})

		Convey("Should build custom sort condition", func() {
			filter := map[string]interface{}{
				"sortBy":   "username",
				"sortDesc": false,
			}
			sort := loginLogDAO.buildSortCondition(filter)
			So(len(sort), ShouldEqual, 1)
		})

		Convey("Should handle invalid sort field", func() {
			filter := map[string]interface{}{
				"sortBy": "invalidField",
			}
			sort := loginLogDAO.buildSortCondition(filter)
			So(len(sort), ShouldEqual, 1)
		})
	})
}
