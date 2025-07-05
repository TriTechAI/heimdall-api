package dao

import (
	"context"
	"testing"
	"time"

	"github.com/bytedance/mockey"
	"github.com/heimdall-api/common/model"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestPageDAO_Create(t *testing.T) {
	Convey("PageDAO Create Tests", t, func() {
		pageDAO := &PageDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should return error when page is nil", func() {
			err := pageDAO.Create(context.Background(), nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "page cannot be nil")
		})

		Convey("Should create page successfully", func() {
			// Mock MongoDB InsertOne method
			mock := mockey.Mock((*mongo.Collection).InsertOne).Return(&mongo.InsertOneResult{
				InsertedID: primitive.NewObjectID(),
			}, nil).Build()
			defer mock.UnPatch()

			authorID := primitive.NewObjectID()
			page := &model.Page{
				Title:    "测试页面",
				Content:  "测试内容",
				Status:   "draft",
				AuthorID: authorID,
			}

			err := pageDAO.Create(context.Background(), page)
			So(err, ShouldBeNil)
		})

		Convey("Should return error when slug already exists", func() {
			// Mock MongoDB InsertOne to return duplicate key error
			mock := mockey.Mock((*mongo.Collection).InsertOne).Return(nil, mongo.WriteException{
				WriteErrors: []mongo.WriteError{
					{Code: 11000}, // Duplicate key error code
				},
			}).Build()
			defer mock.UnPatch()

			authorID := primitive.NewObjectID()
			page := &model.Page{
				Title:    "测试页面",
				Content:  "测试内容",
				Status:   "draft",
				AuthorID: authorID,
			}

			err := pageDAO.Create(context.Background(), page)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "slug already exists")
		})

		Convey("Should return error when validation fails", func() {
			invalidPage := &model.Page{} // 缺少必填字段
			err := pageDAO.Create(context.Background(), invalidPage)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestPageDAO_GetByID(t *testing.T) {
	Convey("PageDAO GetByID Tests", t, func() {
		pageDAO := &PageDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should return error when ID is empty", func() {
			page, err := pageDAO.GetByID(context.Background(), "")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "id cannot be empty")
			So(page, ShouldBeNil)
		})

		Convey("Should return error when ID format is invalid", func() {
			page, err := pageDAO.GetByID(context.Background(), "invalid-id")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "invalid id format")
			So(page, ShouldBeNil)
		})

		Convey("Should return page when found", func() {
			// Mock the FindOne method and SingleResult.Decode
			mockResult := &mongo.SingleResult{}
			mock1 := mockey.Mock((*mongo.Collection).FindOne).Return(mockResult).Build()
			defer mock1.UnPatch()

			mock2 := mockey.Mock((*mongo.SingleResult).Decode).To(func(sr *mongo.SingleResult, v interface{}) error {
				if pagePtr, ok := v.(*model.Page); ok {
					pagePtr.ID = primitive.NewObjectID()
					pagePtr.Title = "测试页面"
					pagePtr.Slug = "test-page"
					pagePtr.Status = "published"
				}
				return nil
			}).Build()
			defer mock2.UnPatch()

			objectID := primitive.NewObjectID()
			page, err := pageDAO.GetByID(context.Background(), objectID.Hex())
			So(err, ShouldBeNil)
			So(page, ShouldNotBeNil)
			So(page.Title, ShouldEqual, "测试页面")
		})

		Convey("Should return nil when page not found", func() {
			// Mock FindOne and Decode to return ErrNoDocuments
			mockResult := &mongo.SingleResult{}
			mock1 := mockey.Mock((*mongo.Collection).FindOne).Return(mockResult).Build()
			defer mock1.UnPatch()

			mock2 := mockey.Mock((*mongo.SingleResult).Decode).Return(mongo.ErrNoDocuments).Build()
			defer mock2.UnPatch()

			objectID := primitive.NewObjectID()
			page, err := pageDAO.GetByID(context.Background(), objectID.Hex())
			So(err, ShouldBeNil)
			So(page, ShouldBeNil)
		})
	})
}

func TestPageDAO_GetBySlug(t *testing.T) {
	Convey("PageDAO GetBySlug Tests", t, func() {
		pageDAO := &PageDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should return error when slug is empty", func() {
			page, err := pageDAO.GetBySlug(context.Background(), "")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "slug cannot be empty")
			So(page, ShouldBeNil)
		})

		Convey("Should return page when found", func() {
			// Mock the FindOne method and SingleResult.Decode
			mockResult := &mongo.SingleResult{}
			mock1 := mockey.Mock((*mongo.Collection).FindOne).Return(mockResult).Build()
			defer mock1.UnPatch()

			mock2 := mockey.Mock((*mongo.SingleResult).Decode).To(func(sr *mongo.SingleResult, v interface{}) error {
				if pagePtr, ok := v.(*model.Page); ok {
					pagePtr.ID = primitive.NewObjectID()
					pagePtr.Title = "测试页面"
					pagePtr.Slug = "test-page"
				}
				return nil
			}).Build()
			defer mock2.UnPatch()

			page, err := pageDAO.GetBySlug(context.Background(), "test-page")
			So(err, ShouldBeNil)
			So(page, ShouldNotBeNil)
			So(page.Slug, ShouldEqual, "test-page")
		})

		Convey("Should return nil when page not found", func() {
			// Mock FindOne and Decode to return ErrNoDocuments
			mockResult := &mongo.SingleResult{}
			mock1 := mockey.Mock((*mongo.Collection).FindOne).Return(mockResult).Build()
			defer mock1.UnPatch()

			mock2 := mockey.Mock((*mongo.SingleResult).Decode).Return(mongo.ErrNoDocuments).Build()
			defer mock2.UnPatch()

			page, err := pageDAO.GetBySlug(context.Background(), "nonexistent")
			So(err, ShouldBeNil)
			So(page, ShouldBeNil)
		})
	})
}

func TestPageDAO_Update(t *testing.T) {
	Convey("PageDAO Update Tests", t, func() {
		pageDAO := &PageDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should return error when ID is empty", func() {
			err := pageDAO.Update(context.Background(), "", map[string]interface{}{"title": "test"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "id cannot be empty")
		})

		Convey("Should return error when updates is empty", func() {
			objectID := primitive.NewObjectID()
			err := pageDAO.Update(context.Background(), objectID.Hex(), nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "updates cannot be empty")
		})

		Convey("Should return error when ID format is invalid", func() {
			err := pageDAO.Update(context.Background(), "invalid-id", map[string]interface{}{"title": "test"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "invalid id format")
		})

		Convey("Should update page successfully", func() {
			// Mock MongoDB UpdateOne method
			mock := mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
			}, nil).Build()
			defer mock.UnPatch()

			objectID := primitive.NewObjectID()
			updates := map[string]interface{}{
				"title": "更新后的标题",
			}

			err := pageDAO.Update(context.Background(), objectID.Hex(), updates)
			So(err, ShouldBeNil)
		})

		Convey("Should return error when page not found", func() {
			// Mock MongoDB UpdateOne to return no matched documents
			mock := mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
				MatchedCount:  0,
				ModifiedCount: 0,
			}, nil).Build()
			defer mock.UnPatch()

			objectID := primitive.NewObjectID()
			updates := map[string]interface{}{
				"title": "更新后的标题",
			}

			err := pageDAO.Update(context.Background(), objectID.Hex(), updates)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "page not found")
		})

		Convey("Should return error when slug already exists", func() {
			// Mock MongoDB UpdateOne to return duplicate key error
			mock := mockey.Mock((*mongo.Collection).UpdateOne).Return(nil, mongo.WriteException{
				WriteErrors: []mongo.WriteError{
					{Code: 11000}, // Duplicate key error code
				},
			}).Build()
			defer mock.UnPatch()

			objectID := primitive.NewObjectID()
			updates := map[string]interface{}{
				"slug": "existing-slug",
			}

			err := pageDAO.Update(context.Background(), objectID.Hex(), updates)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "slug already exists")
		})
	})
}

func TestPageDAO_Delete(t *testing.T) {
	Convey("PageDAO Delete Tests", t, func() {
		pageDAO := &PageDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should return error when ID is empty", func() {
			err := pageDAO.Delete(context.Background(), "")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "id cannot be empty")
		})

		Convey("Should return error when ID format is invalid", func() {
			err := pageDAO.Delete(context.Background(), "invalid-id")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "invalid id format")
		})

		Convey("Should delete page successfully (soft delete)", func() {
			// Mock MongoDB UpdateOne method for soft delete
			mock := mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
			}, nil).Build()
			defer mock.UnPatch()

			objectID := primitive.NewObjectID()
			err := pageDAO.Delete(context.Background(), objectID.Hex())
			So(err, ShouldBeNil)
		})

		Convey("Should return error when page not found", func() {
			// Mock MongoDB UpdateOne to return no matched documents
			mock := mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
				MatchedCount:  0,
				ModifiedCount: 0,
			}, nil).Build()
			defer mock.UnPatch()

			objectID := primitive.NewObjectID()
			err := pageDAO.Delete(context.Background(), objectID.Hex())
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "page not found")
		})
	})
}

func TestPageDAO_List(t *testing.T) {
	Convey("PageDAO List Tests", t, func() {
		pageDAO := &PageDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should list pages successfully", func() {
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
				if pagePtr, ok := v.(*model.Page); ok {
					pagePtr.ID = primitive.NewObjectID()
					pagePtr.Title = "测试页面"
					pagePtr.Status = "published"
				}
				return nil
			}).Build()
			defer mock4.UnPatch()

			mock5 := mockey.Mock((*mongo.Cursor).Close).Return(nil).Build()
			defer mock5.UnPatch()

			filter := model.PageFilter{
				Status: "published",
			}

			pages, total, err := pageDAO.List(context.Background(), filter, 1, 10)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 2)
			So(len(pages), ShouldEqual, 2)
		})

		Convey("Should validate pagination parameters", func() {
			// Mock MongoDB operations
			mockCursor := &mongo.Cursor{}
			mock1 := mockey.Mock((*mongo.Collection).Find).Return(mockCursor, nil).Build()
			defer mock1.UnPatch()

			mock2 := mockey.Mock((*mongo.Collection).CountDocuments).Return(int64(0), nil).Build()
			defer mock2.UnPatch()

			mock3 := mockey.Mock((*mongo.Cursor).Next).Return(false).Build()
			defer mock3.UnPatch()

			mock4 := mockey.Mock((*mongo.Cursor).Close).Return(nil).Build()
			defer mock4.UnPatch()

			filter := model.PageFilter{}

			// 测试负数页码
			pages, total, err := pageDAO.List(context.Background(), filter, -1, 10)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 0)
			So(len(pages), ShouldEqual, 0)

			// 测试超大limit
			pages, total, err = pageDAO.List(context.Background(), filter, 1, 200)
			So(err, ShouldBeNil)
			// limit应该被限制为100
		})
	})
}

func TestPageDAO_GetByAuthor(t *testing.T) {
	Convey("PageDAO GetByAuthor Tests", t, func() {
		pageDAO := &PageDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should return error when authorID is empty", func() {
			filter := model.PageFilter{}
			pages, total, err := pageDAO.GetByAuthor(context.Background(), "", filter, 1, 10)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "authorID cannot be empty")
			So(pages, ShouldBeNil)
			So(total, ShouldEqual, 0)
		})

		Convey("Should get pages by author successfully", func() {
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
				if pagePtr, ok := v.(*model.Page); ok {
					pagePtr.ID = primitive.NewObjectID()
					pagePtr.Title = "作者页面"
					pagePtr.AuthorID = primitive.NewObjectID()
				}
				return nil
			}).Build()
			defer mock4.UnPatch()

			mock5 := mockey.Mock((*mongo.Cursor).Close).Return(nil).Build()
			defer mock5.UnPatch()

			authorID := primitive.NewObjectID()
			filter := model.PageFilter{}
			pages, total, err := pageDAO.GetByAuthor(context.Background(), authorID.Hex(), filter, 1, 10)

			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(pages), ShouldEqual, 1)
		})
	})
}

func TestPageDAO_PublishMethods(t *testing.T) {
	Convey("PageDAO Publish Methods Tests", t, func() {
		pageDAO := &PageDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Publish should work correctly", func() {
			// Mock MongoDB UpdateOne method
			mock := mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
			}, nil).Build()
			defer mock.UnPatch()

			objectID := primitive.NewObjectID()
			err := pageDAO.Publish(context.Background(), objectID.Hex())
			So(err, ShouldBeNil)
		})

		Convey("Unpublish should work correctly", func() {
			// Mock MongoDB UpdateOne method
			mock := mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
			}, nil).Build()
			defer mock.UnPatch()

			objectID := primitive.NewObjectID()
			err := pageDAO.Unpublish(context.Background(), objectID.Hex())
			So(err, ShouldBeNil)
		})

		Convey("Should return error when page not found", func() {
			// Mock MongoDB UpdateOne to return no matched documents
			mock := mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
				MatchedCount:  0,
				ModifiedCount: 0,
			}, nil).Build()
			defer mock.UnPatch()

			objectID := primitive.NewObjectID()
			err := pageDAO.Publish(context.Background(), objectID.Hex())
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "page not found")
		})
	})
}

func TestPageDAO_SpecialQueries(t *testing.T) {
	Convey("PageDAO Special Queries Tests", t, func() {
		pageDAO := &PageDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("GetScheduledPages should work correctly", func() {
			// Mock MongoDB Find and cursor operations
			mockCursor := &mongo.Cursor{}
			mock1 := mockey.Mock((*mongo.Collection).Find).Return(mockCursor, nil).Build()
			defer mock1.UnPatch()

			callCount := 0
			mock2 := mockey.Mock((*mongo.Cursor).Next).To(func(c *mongo.Cursor, ctx context.Context) bool {
				callCount++
				return callCount <= 1
			}).Build()
			defer mock2.UnPatch()

			mock3 := mockey.Mock((*mongo.Cursor).Decode).To(func(c *mongo.Cursor, v interface{}) error {
				if pagePtr, ok := v.(*model.Page); ok {
					pagePtr.ID = primitive.NewObjectID()
					pagePtr.Title = "定时页面"
					pagePtr.Status = "scheduled"
					publishedAt := time.Now().Add(-1 * time.Hour)
					pagePtr.PublishedAt = &publishedAt
				}
				return nil
			}).Build()
			defer mock3.UnPatch()

			mock4 := mockey.Mock((*mongo.Cursor).Close).Return(nil).Build()
			defer mock4.UnPatch()

			pages, err := pageDAO.GetScheduledPages(context.Background())

			So(err, ShouldBeNil)
			So(len(pages), ShouldEqual, 1)
		})

		Convey("GetRecentPages should work correctly", func() {
			// Mock MongoDB Find and cursor operations
			mockCursor := &mongo.Cursor{}
			mock1 := mockey.Mock((*mongo.Collection).Find).Return(mockCursor, nil).Build()
			defer mock1.UnPatch()

			callCount := 0
			mock2 := mockey.Mock((*mongo.Cursor).Next).To(func(c *mongo.Cursor, ctx context.Context) bool {
				callCount++
				return callCount <= 1
			}).Build()
			defer mock2.UnPatch()

			mock3 := mockey.Mock((*mongo.Cursor).Decode).To(func(c *mongo.Cursor, v interface{}) error {
				if pagePtr, ok := v.(*model.Page); ok {
					pagePtr.ID = primitive.NewObjectID()
					pagePtr.Title = "最新页面"
					pagePtr.Status = "published"
				}
				return nil
			}).Build()
			defer mock3.UnPatch()

			mock4 := mockey.Mock((*mongo.Cursor).Close).Return(nil).Build()
			defer mock4.UnPatch()

			pages, err := pageDAO.GetRecentPages(context.Background(), 10)

			So(err, ShouldBeNil)
			So(len(pages), ShouldEqual, 1)
		})
	})
}

func TestPageDAO_HelperMethods(t *testing.T) {
	Convey("PageDAO Helper Methods Tests", t, func() {
		pageDAO := &PageDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("GetByTemplate should work correctly", func() {
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
				if pagePtr, ok := v.(*model.Page); ok {
					pagePtr.ID = primitive.NewObjectID()
					pagePtr.Title = "自定义模板页面"
					pagePtr.Template = "custom"
				}
				return nil
			}).Build()
			defer mock4.UnPatch()

			mock5 := mockey.Mock((*mongo.Cursor).Close).Return(nil).Build()
			defer mock5.UnPatch()

			filter := model.PageFilter{}
			pages, total, err := pageDAO.GetByTemplate(context.Background(), "custom", filter, 1, 10)

			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(pages), ShouldEqual, 1)
		})

		Convey("GetByTemplate should return error when template is empty", func() {
			filter := model.PageFilter{}
			pages, total, err := pageDAO.GetByTemplate(context.Background(), "", filter, 1, 10)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "template cannot be empty")
			So(pages, ShouldBeNil)
			So(total, ShouldEqual, 0)
		})

		Convey("GetPublishedList should work correctly", func() {
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
				if pagePtr, ok := v.(*model.Page); ok {
					pagePtr.ID = primitive.NewObjectID()
					pagePtr.Title = "已发布页面"
					pagePtr.Status = "published"
				}
				return nil
			}).Build()
			defer mock4.UnPatch()

			mock5 := mockey.Mock((*mongo.Cursor).Close).Return(nil).Build()
			defer mock5.UnPatch()

			filter := model.PageFilter{}
			pages, total, err := pageDAO.GetPublishedList(context.Background(), filter, 1, 10)

			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(pages), ShouldEqual, 1)
		})
	})
}

func TestPageDAO_QueryBuilders(t *testing.T) {
	Convey("PageDAO Query Builders Tests", t, func() {
		pageDAO := NewPageDAO(nil)

		Convey("buildQuery should work correctly", func() {
			filter := model.PageFilter{
				Status:   "published",
				Template: "custom",
				AuthorID: primitive.NewObjectID().Hex(),
				Keyword:  "测试",
			}

			query := pageDAO.buildQuery(filter)

			So(query["status"], ShouldEqual, "published")
			So(query["template"], ShouldEqual, "custom")
			So(query["authorId"], ShouldNotBeNil)
			So(query["$or"], ShouldNotBeNil)
		})

		Convey("buildSort should work correctly", func() {
			filter := model.PageFilter{
				SortBy:   "title",
				SortDesc: true,
			}

			sort := pageDAO.buildSort(filter)

			So(len(sort), ShouldEqual, 1)
			So(sort[0].Key, ShouldEqual, "title")
			So(sort[0].Value, ShouldEqual, -1)
		})

		Convey("buildSort should use default sort", func() {
			filter := model.PageFilter{}

			sort := pageDAO.buildSort(filter)

			So(len(sort), ShouldEqual, 1)
			So(sort[0].Key, ShouldEqual, "createdAt")
			So(sort[0].Value, ShouldEqual, -1)
		})
	})
}

func TestPageDAO_CreateIndexes(t *testing.T) {
	Convey("PageDAO CreateIndexes Tests", t, func() {
		pageDAO := &PageDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should create indexes successfully", func() {
			// Mock MongoDB CreateMany method
			mock := mockey.Mock((*mongo.IndexView).CreateMany).Return([]string{"index1", "index2"}, nil).Build()
			defer mock.UnPatch()

			err := pageDAO.CreateIndexes(context.Background())

			So(err, ShouldBeNil)
		})

		Convey("Should handle index creation errors", func() {
			// Mock MongoDB CreateMany to return error
			mock := mockey.Mock((*mongo.IndexView).CreateMany).Return(nil, mongo.CommandError{
				Code:    1,
				Message: "index creation failed",
			}).Build()
			defer mock.UnPatch()

			err := pageDAO.CreateIndexes(context.Background())

			So(err, ShouldNotBeNil)
		})
	})
}
