package dao

import (
	"context"
	"testing"
	"time"

	"github.com/bytedance/mockey"
	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/model"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestPostDAO_Create(t *testing.T) {
	Convey("PostDAO Create Tests", t, func() {
		postDAO := &PostDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should return error when post is nil", func() {
			err := postDAO.Create(context.Background(), nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "post cannot be nil")
		})

		Convey("Should create post successfully", func() {
			// Mock MongoDB InsertOne method
			mock := mockey.Mock((*mongo.Collection).InsertOne).Return(&mongo.InsertOneResult{
				InsertedID: primitive.NewObjectID(),
			}, nil).Build()
			defer mock.UnPatch()

			authorID := primitive.NewObjectID()
			post := &model.Post{
				Title:      "测试文章",
				Markdown:   "测试内容",
				Type:       constants.PostTypePost,
				Status:     constants.PostStatusDraft,
				Visibility: constants.PostVisibilityPublic,
				AuthorID:   authorID,
			}

			err := postDAO.Create(context.Background(), post)
			So(err, ShouldBeNil)
			So(post.ID, ShouldNotEqual, primitive.NilObjectID)
			So(post.Slug, ShouldNotBeEmpty)
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
			post := &model.Post{
				Title:      "测试文章",
				Markdown:   "测试内容",
				Type:       constants.PostTypePost,
				Status:     constants.PostStatusDraft,
				Visibility: constants.PostVisibilityPublic,
				AuthorID:   authorID,
			}

			err := postDAO.Create(context.Background(), post)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "slug already exists")
		})

		Convey("Should return error when validation fails", func() {
			invalidPost := &model.Post{} // 缺少必填字段
			err := postDAO.Create(context.Background(), invalidPost)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestPostDAO_GetByID(t *testing.T) {
	Convey("PostDAO GetByID Tests", t, func() {
		postDAO := &PostDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should return error when ID is empty", func() {
			post, err := postDAO.GetByID(context.Background(), "")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "id cannot be empty")
			So(post, ShouldBeNil)
		})

		Convey("Should return error when ID format is invalid", func() {
			post, err := postDAO.GetByID(context.Background(), "invalid-id")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "invalid id format")
			So(post, ShouldBeNil)
		})

		Convey("Should return post when found", func() {
			// Mock the FindOne method and SingleResult.Decode
			mockResult := &mongo.SingleResult{}
			mock1 := mockey.Mock((*mongo.Collection).FindOne).Return(mockResult).Build()
			defer mock1.UnPatch()

			mock2 := mockey.Mock((*mongo.SingleResult).Decode).To(func(sr *mongo.SingleResult, v interface{}) error {
				if postPtr, ok := v.(*model.Post); ok {
					postPtr.ID = primitive.NewObjectID()
					postPtr.Title = "测试文章"
					postPtr.Slug = "test-post"
					postPtr.Status = constants.PostStatusPublished
				}
				return nil
			}).Build()
			defer mock2.UnPatch()

			objectID := primitive.NewObjectID()
			post, err := postDAO.GetByID(context.Background(), objectID.Hex())
			So(err, ShouldBeNil)
			So(post, ShouldNotBeNil)
			So(post.Title, ShouldEqual, "测试文章")
		})

		Convey("Should return nil when post not found", func() {
			// Mock FindOne and Decode to return ErrNoDocuments
			mockResult := &mongo.SingleResult{}
			mock1 := mockey.Mock((*mongo.Collection).FindOne).Return(mockResult).Build()
			defer mock1.UnPatch()

			mock2 := mockey.Mock((*mongo.SingleResult).Decode).Return(mongo.ErrNoDocuments).Build()
			defer mock2.UnPatch()

			objectID := primitive.NewObjectID()
			post, err := postDAO.GetByID(context.Background(), objectID.Hex())
			So(err, ShouldBeNil)
			So(post, ShouldBeNil)
		})
	})
}

func TestPostDAO_GetBySlug(t *testing.T) {
	Convey("PostDAO GetBySlug Tests", t, func() {
		postDAO := &PostDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should return error when slug is empty", func() {
			post, err := postDAO.GetBySlug(context.Background(), "")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "slug cannot be empty")
			So(post, ShouldBeNil)
		})

		Convey("Should return post when found", func() {
			// Mock the FindOne method and SingleResult.Decode
			mockResult := &mongo.SingleResult{}
			mock1 := mockey.Mock((*mongo.Collection).FindOne).Return(mockResult).Build()
			defer mock1.UnPatch()

			mock2 := mockey.Mock((*mongo.SingleResult).Decode).To(func(sr *mongo.SingleResult, v interface{}) error {
				if postPtr, ok := v.(*model.Post); ok {
					postPtr.ID = primitive.NewObjectID()
					postPtr.Title = "测试文章"
					postPtr.Slug = "test-post"
				}
				return nil
			}).Build()
			defer mock2.UnPatch()

			post, err := postDAO.GetBySlug(context.Background(), "test-post")
			So(err, ShouldBeNil)
			So(post, ShouldNotBeNil)
			So(post.Slug, ShouldEqual, "test-post")
		})

		Convey("Should return nil when post not found", func() {
			// Mock FindOne and Decode to return ErrNoDocuments
			mockResult := &mongo.SingleResult{}
			mock1 := mockey.Mock((*mongo.Collection).FindOne).Return(mockResult).Build()
			defer mock1.UnPatch()

			mock2 := mockey.Mock((*mongo.SingleResult).Decode).Return(mongo.ErrNoDocuments).Build()
			defer mock2.UnPatch()

			post, err := postDAO.GetBySlug(context.Background(), "nonexistent")
			So(err, ShouldBeNil)
			So(post, ShouldBeNil)
		})
	})
}

func TestPostDAO_Update(t *testing.T) {
	Convey("PostDAO Update Tests", t, func() {
		postDAO := &PostDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should return error when ID is empty", func() {
			err := postDAO.Update(context.Background(), "", map[string]interface{}{"title": "test"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "id cannot be empty")
		})

		Convey("Should return error when updates is empty", func() {
			objectID := primitive.NewObjectID()
			err := postDAO.Update(context.Background(), objectID.Hex(), nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "updates cannot be empty")
		})

		Convey("Should return error when ID format is invalid", func() {
			err := postDAO.Update(context.Background(), "invalid-id", map[string]interface{}{"title": "test"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "invalid id format")
		})

		Convey("Should update post successfully", func() {
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

			err := postDAO.Update(context.Background(), objectID.Hex(), updates)
			So(err, ShouldBeNil)
		})

		Convey("Should return error when post not found", func() {
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

			err := postDAO.Update(context.Background(), objectID.Hex(), updates)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "post not found")
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

			err := postDAO.Update(context.Background(), objectID.Hex(), updates)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "slug already exists")
		})
	})
}

func TestPostDAO_Delete(t *testing.T) {
	Convey("PostDAO Delete Tests", t, func() {
		postDAO := &PostDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should return error when ID is empty", func() {
			err := postDAO.Delete(context.Background(), "")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "id cannot be empty")
		})

		Convey("Should return error when ID format is invalid", func() {
			err := postDAO.Delete(context.Background(), "invalid-id")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "invalid id format")
		})

		Convey("Should delete post successfully (soft delete)", func() {
			// Mock MongoDB UpdateOne method for soft delete
			mock := mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
			}, nil).Build()
			defer mock.UnPatch()

			objectID := primitive.NewObjectID()
			err := postDAO.Delete(context.Background(), objectID.Hex())
			So(err, ShouldBeNil)
		})

		Convey("Should return error when post not found", func() {
			// Mock MongoDB UpdateOne to return no matched documents
			mock := mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
				MatchedCount:  0,
				ModifiedCount: 0,
			}, nil).Build()
			defer mock.UnPatch()

			objectID := primitive.NewObjectID()
			err := postDAO.Delete(context.Background(), objectID.Hex())
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "post not found")
		})
	})
}

func TestPostDAO_List(t *testing.T) {
	Convey("PostDAO List Tests", t, func() {
		postDAO := &PostDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should list posts successfully", func() {
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
				if postPtr, ok := v.(*model.Post); ok {
					postPtr.ID = primitive.NewObjectID()
					postPtr.Title = "测试文章"
					postPtr.Status = constants.PostStatusPublished
				}
				return nil
			}).Build()
			defer mock4.UnPatch()

			mock5 := mockey.Mock((*mongo.Cursor).Close).Return(nil).Build()
			defer mock5.UnPatch()

			filter := model.PostFilter{
				Status: constants.PostStatusPublished,
			}

			posts, total, err := postDAO.List(context.Background(), filter, 1, 10)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 2)
			So(len(posts), ShouldEqual, 2)
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

			filter := model.PostFilter{}

			// 测试负数页码
			posts, total, err := postDAO.List(context.Background(), filter, -1, 10)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 0)
			So(len(posts), ShouldEqual, 0)

			// 测试超大limit
			posts, total, err = postDAO.List(context.Background(), filter, 1, 200)
			So(err, ShouldBeNil)
			// limit应该被限制为100
		})
	})
}

func TestPostDAO_IncrementViewCount(t *testing.T) {
	Convey("PostDAO IncrementViewCount Tests", t, func() {
		postDAO := &PostDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should return error when ID is empty", func() {
			err := postDAO.IncrementViewCount(context.Background(), "")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "id cannot be empty")
		})

		Convey("Should increment view count successfully", func() {
			// Mock MongoDB UpdateOne method
			mock := mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
			}, nil).Build()
			defer mock.UnPatch()

			objectID := primitive.NewObjectID()
			err := postDAO.IncrementViewCount(context.Background(), objectID.Hex())
			So(err, ShouldBeNil)
		})

		Convey("Should return error when post not found", func() {
			// Mock MongoDB UpdateOne to return no matched documents
			mock := mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
				MatchedCount:  0,
				ModifiedCount: 0,
			}, nil).Build()
			defer mock.UnPatch()

			objectID := primitive.NewObjectID()
			err := postDAO.IncrementViewCount(context.Background(), objectID.Hex())
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "post not found")
		})
	})
}

func TestPostDAO_PublishMethods(t *testing.T) {
	Convey("PostDAO Publish Methods Tests", t, func() {
		postDAO := &PostDAO{
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
			err := postDAO.Publish(context.Background(), objectID.Hex())
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
			err := postDAO.Unpublish(context.Background(), objectID.Hex())
			So(err, ShouldBeNil)
		})

		Convey("Should return error when post not found", func() {
			// Mock MongoDB UpdateOne to return no matched documents
			mock := mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
				MatchedCount:  0,
				ModifiedCount: 0,
			}, nil).Build()
			defer mock.UnPatch()

			objectID := primitive.NewObjectID()
			err := postDAO.Publish(context.Background(), objectID.Hex())
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "post not found")
		})
	})
}

func TestPostDAO_SpecialQueries(t *testing.T) {
	Convey("PostDAO Special Queries Tests", t, func() {
		postDAO := &PostDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("GetScheduledPosts should work correctly", func() {
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
				if postPtr, ok := v.(*model.Post); ok {
					postPtr.ID = primitive.NewObjectID()
					postPtr.Title = "定时文章"
					postPtr.Status = constants.PostStatusScheduled
					pastTime := time.Now().Add(-time.Hour)
					postPtr.PublishedAt = &pastTime
				}
				return nil
			}).Build()
			defer mock3.UnPatch()

			mock4 := mockey.Mock((*mongo.Cursor).Close).Return(nil).Build()
			defer mock4.UnPatch()

			posts, err := postDAO.GetScheduledPosts(context.Background())
			So(err, ShouldBeNil)
			So(len(posts), ShouldEqual, 1)
		})

		Convey("GetPopularPosts should work correctly", func() {
			// Mock MongoDB Find and cursor operations
			mockCursor := &mongo.Cursor{}
			mock1 := mockey.Mock((*mongo.Collection).Find).Return(mockCursor, nil).Build()
			defer mock1.UnPatch()

			// Mock cursor operations
			callCount := 0
			mock2 := mockey.Mock((*mongo.Cursor).Next).To(func(c *mongo.Cursor, ctx context.Context) bool {
				callCount++
				return callCount <= 2 // Return true for first 2 calls, false for 3rd
			}).Build()
			defer mock2.UnPatch()

			mock3 := mockey.Mock((*mongo.Cursor).Decode).To(func(c *mongo.Cursor, v interface{}) error {
				if postPtr, ok := v.(*model.Post); ok {
					postPtr.ID = primitive.NewObjectID()
					postPtr.Title = "热门文章"
					postPtr.Status = constants.PostStatusPublished
					postPtr.ViewCount = 1000
				}
				return nil
			}).Build()
			defer mock3.UnPatch()

			mock4 := mockey.Mock((*mongo.Cursor).Close).Return(nil).Build()
			defer mock4.UnPatch()

			posts, err := postDAO.GetPopularPosts(context.Background(), 10, 30)
			So(err, ShouldBeNil)
			So(len(posts), ShouldEqual, 2)
		})

		Convey("GetRecentPosts should work correctly", func() {
			// Mock MongoDB Find and cursor operations
			mockCursor := &mongo.Cursor{}
			mock1 := mockey.Mock((*mongo.Collection).Find).Return(mockCursor, nil).Build()
			defer mock1.UnPatch()

			// Mock cursor operations
			callCount := 0
			mock2 := mockey.Mock((*mongo.Cursor).Next).To(func(c *mongo.Cursor, ctx context.Context) bool {
				callCount++
				return callCount <= 2 // Return true for first 2 calls, false for 3rd
			}).Build()
			defer mock2.UnPatch()

			mock3 := mockey.Mock((*mongo.Cursor).Decode).To(func(c *mongo.Cursor, v interface{}) error {
				if postPtr, ok := v.(*model.Post); ok {
					postPtr.ID = primitive.NewObjectID()
					postPtr.Title = "最新文章"
					postPtr.Status = constants.PostStatusPublished
					now := time.Now()
					postPtr.PublishedAt = &now
				}
				return nil
			}).Build()
			defer mock3.UnPatch()

			mock4 := mockey.Mock((*mongo.Cursor).Close).Return(nil).Build()
			defer mock4.UnPatch()

			posts, err := postDAO.GetRecentPosts(context.Background(), 10)
			So(err, ShouldBeNil)
			So(len(posts), ShouldEqual, 2)
		})
	})
}

func TestPostDAO_HelperMethods(t *testing.T) {
	Convey("PostDAO Helper Methods Tests", t, func() {
		postDAO := &PostDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("GetByAuthor should work correctly", func() {
			// Mock the List method since GetByAuthor calls it
			mock := mockey.Mock((*PostDAO).List).Return([]*model.Post{
				{
					ID:       primitive.NewObjectID(),
					Title:    "作者文章",
					AuthorID: primitive.NewObjectID(),
				},
			}, int64(1), nil).Build()
			defer mock.UnPatch()

			authorID := primitive.NewObjectID()
			posts, total, err := postDAO.GetByAuthor(context.Background(), authorID.Hex(), model.PostFilter{}, 1, 10)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(posts), ShouldEqual, 1)
		})

		Convey("GetByAuthor should return error when authorID is empty", func() {
			posts, total, err := postDAO.GetByAuthor(context.Background(), "", model.PostFilter{}, 1, 10)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(posts, ShouldBeNil)
			So(err.Error(), ShouldEqual, "authorID cannot be empty")
		})

		Convey("GetByTag should work correctly", func() {
			// Mock the List method since GetByTag calls it
			mock := mockey.Mock((*PostDAO).List).Return([]*model.Post{
				{
					ID:    primitive.NewObjectID(),
					Title: "标签文章",
					Tags:  []model.Tag{{Name: "Go", Slug: "golang"}},
				},
			}, int64(1), nil).Build()
			defer mock.UnPatch()

			posts, total, err := postDAO.GetByTag(context.Background(), "golang", model.PostFilter{}, 1, 10)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(posts), ShouldEqual, 1)
		})

		Convey("GetByTag should return error when tagSlug is empty", func() {
			posts, total, err := postDAO.GetByTag(context.Background(), "", model.PostFilter{}, 1, 10)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(posts, ShouldBeNil)
			So(err.Error(), ShouldEqual, "tagSlug cannot be empty")
		})

		Convey("GetPublishedList should work correctly", func() {
			// Mock the List method since GetPublishedList calls it
			mock := mockey.Mock((*PostDAO).List).Return([]*model.Post{
				{
					ID:         primitive.NewObjectID(),
					Title:      "已发布文章",
					Status:     constants.PostStatusPublished,
					Visibility: constants.PostVisibilityPublic,
				},
			}, int64(1), nil).Build()
			defer mock.UnPatch()

			posts, total, err := postDAO.GetPublishedList(context.Background(), model.PostFilter{}, 1, 10)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(posts), ShouldEqual, 1)
		})
	})
}

func TestPostDAO_QueryBuilders(t *testing.T) {
	Convey("PostDAO Query Builders Tests", t, func() {
		postDAO := &PostDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("buildQuery should work correctly", func() {
			filter := model.PostFilter{
				Status:     constants.PostStatusPublished,
				Type:       constants.PostTypePost,
				Visibility: constants.PostVisibilityPublic,
				AuthorID:   primitive.NewObjectID().Hex(),
				Tag:        "golang",
				Keyword:    "测试",
			}

			query := postDAO.buildQuery(filter)
			So(query["status"], ShouldEqual, constants.PostStatusPublished)
			So(query["type"], ShouldEqual, constants.PostTypePost)
			So(query["visibility"], ShouldEqual, constants.PostVisibilityPublic)
			So(query["tags.slug"], ShouldEqual, "golang")
			So(query["$or"], ShouldNotBeNil)
		})

		Convey("buildSort should work correctly", func() {
			// 测试默认排序
			filter := model.PostFilter{}
			sort := postDAO.buildSort(filter)
			So(len(sort), ShouldEqual, 1)
			So(sort[0].Key, ShouldEqual, "createdAt")
			So(sort[0].Value, ShouldEqual, -1)

			// 测试自定义排序（降序）
			filter = model.PostFilter{
				SortBy:   "view_count",
				SortDesc: true,
			}
			sort = postDAO.buildSort(filter)
			So(sort[0].Key, ShouldEqual, "viewCount")
			So(sort[0].Value, ShouldEqual, -1)

			// 测试升序排序
			filter = model.PostFilter{
				SortBy:   "title",
				SortDesc: false,
			}
			sort = postDAO.buildSort(filter)
			So(sort[0].Key, ShouldEqual, "title")
			So(sort[0].Value, ShouldEqual, 1)
		})
	})
}

func TestPostDAO_CreateIndexes(t *testing.T) {
	Convey("PostDAO CreateIndexes Tests", t, func() {
		postDAO := &PostDAO{
			collection: &mongo.Collection{}, // Mock collection
		}

		Convey("Should create indexes successfully", func() {
			// Mock the entire CreateIndexes method since it's easier than mocking collection.Indexes()
			mock := mockey.Mock((*PostDAO).CreateIndexes).Return(nil).Build()
			defer mock.UnPatch()

			err := postDAO.CreateIndexes(context.Background())
			So(err, ShouldBeNil)
		})

		Convey("Should handle index creation errors", func() {
			// Mock CreateIndexes to return error
			mock := mockey.Mock((*PostDAO).CreateIndexes).Return(mongo.CommandError{
				Code:    1,
				Message: "index creation failed",
			}).Build()
			defer mock.UnPatch()

			err := postDAO.CreateIndexes(context.Background())
			So(err, ShouldNotBeNil)
		})
	})
}
