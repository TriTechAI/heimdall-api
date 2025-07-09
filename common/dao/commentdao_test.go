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

func TestCommentDAO(t *testing.T) {
	mockey.PatchConvey("CommentDAO单元测试", t, func() {
		ctx := context.Background()

		// Mock the collection to avoid nil pointer errors
		mockCollection := &mongo.Collection{}
		dao := &CommentDAO{
			collection: mockCollection,
		}

		Convey("Create方法测试", func() {
			postID := primitive.NewObjectID()
			comment := &model.Comment{
				PostID:      postID,
				Content:     "测试评论内容",
				AuthorName:  "测试作者",
				AuthorEmail: "test@example.com",
				AuthorIP:    "192.168.1.1",
			}

			Convey("正常创建评论", func() {
				mockey.Mock((*mongo.Collection).InsertOne).Return(&mongo.InsertOneResult{
					InsertedID: primitive.NewObjectID(),
				}, nil).Build()

				err := dao.Create(ctx, comment)
				So(err, ShouldBeNil)
			})

			Convey("空评论创建失败", func() {
				err := dao.Create(ctx, nil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "comment cannot be nil")
			})

			Convey("数据库插入失败", func() {
				mockey.Mock((*mongo.Collection).InsertOne).Return(nil, mongo.WriteException{}).Build()

				err := dao.Create(ctx, comment)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to create comment")
			})
		})

		Convey("GetByID方法测试", func() {
			validID := primitive.NewObjectID()
			validIDStr := validID.Hex()

			Convey("正常获取评论", func() {
				expectedComment := &model.Comment{
					ID:          validID,
					Content:     "测试评论",
					AuthorName:  "测试作者",
					AuthorEmail: "test@example.com",
				}

				// Mock FindOne
				mockSingleResult := &mongo.SingleResult{}
				mockey.Mock((*mongo.Collection).FindOne).Return(mockSingleResult).Build()
				mockey.Mock((*mongo.SingleResult).Decode).To(func(result interface{}) error {
					comment := result.(*model.Comment)
					*comment = *expectedComment
					return nil
				}).Build()

				comment, err := dao.GetByID(ctx, validIDStr)
				So(err, ShouldBeNil)
				So(comment, ShouldNotBeNil)
				So(comment.ID, ShouldEqual, validID)
			})

			Convey("无效ID格式", func() {
				comment, err := dao.GetByID(ctx, "invalid-id")
				So(err, ShouldNotBeNil)
				So(comment, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid comment ID")
			})

			Convey("评论不存在", func() {
				mockSingleResult := &mongo.SingleResult{}
				mockey.Mock((*mongo.Collection).FindOne).Return(mockSingleResult).Build()
				mockey.Mock((*mongo.SingleResult).Decode).Return(mongo.ErrNoDocuments).Build()

				comment, err := dao.GetByID(ctx, validIDStr)
				So(err, ShouldNotBeNil)
				So(comment, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "comment not found")
			})
		})

		Convey("Update方法测试", func() {
			validID := primitive.NewObjectID()
			validIDStr := validID.Hex()

			Convey("正常更新评论", func() {
				updates := map[string]interface{}{
					"content": "更新后的内容",
					"status":  constants.CommentStatusApproved,
				}

				mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
					MatchedCount:  1,
					ModifiedCount: 1,
				}, nil).Build()

				err := dao.Update(ctx, validIDStr, updates)
				So(err, ShouldBeNil)
			})

			Convey("无效ID格式", func() {
				updates := map[string]interface{}{"content": "test"}
				err := dao.Update(ctx, "invalid-id", updates)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid comment ID")
			})

			Convey("空更新数据", func() {
				err := dao.Update(ctx, validIDStr, map[string]interface{}{})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "no updates provided")
			})

			Convey("评论不存在", func() {
				updates := map[string]interface{}{"content": "test"}
				mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
					MatchedCount: 0,
				}, nil).Build()

				err := dao.Update(ctx, validIDStr, updates)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "comment not found")
			})
		})

		Convey("Delete方法测试", func() {
			validID := primitive.NewObjectID()
			validIDStr := validID.Hex()

			Convey("正常删除评论", func() {
				mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
					MatchedCount:  1,
					ModifiedCount: 1,
				}, nil).Build()

				err := dao.Delete(ctx, validIDStr)
				So(err, ShouldBeNil)
			})

			Convey("无效ID格式", func() {
				err := dao.Delete(ctx, "invalid-id")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid comment ID")
			})

			Convey("评论不存在", func() {
				mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
					MatchedCount: 0,
				}, nil).Build()

				err := dao.Delete(ctx, validIDStr)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "comment not found")
			})
		})

		Convey("List方法测试", func() {
			filter := model.CommentFilter{
				Status: constants.CommentStatusApproved,
			}

			Convey("正常获取评论列表", func() {
				mockey.Mock((*mongo.Collection).CountDocuments).Return(int64(10), nil).Build()

				mockCursor := &mongo.Cursor{}
				mockey.Mock((*mongo.Collection).Find).Return(mockCursor, nil).Build()
				mockey.Mock((*mongo.Cursor).All).To(func(ctx context.Context, results interface{}) error {
					comments := results.(*[]*model.Comment)
					*comments = []*model.Comment{
						{
							ID:          primitive.NewObjectID(),
							Content:     "测试评论1",
							AuthorName:  "作者1",
							AuthorEmail: "test1@example.com",
							Status:      constants.CommentStatusApproved,
						},
						{
							ID:          primitive.NewObjectID(),
							Content:     "测试评论2",
							AuthorName:  "作者2",
							AuthorEmail: "test2@example.com",
							Status:      constants.CommentStatusApproved,
						},
					}
					return nil
				}).Build()
				mockey.Mock((*mongo.Cursor).Close).Return(nil).Build()

				comments, total, err := dao.List(ctx, filter, 1, 10)
				So(err, ShouldBeNil)
				So(total, ShouldEqual, 10)
				So(len(comments), ShouldEqual, 2)
			})

			Convey("空结果列表", func() {
				mockey.Mock((*mongo.Collection).CountDocuments).Return(int64(0), nil).Build()

				comments, total, err := dao.List(ctx, filter, 1, 10)
				So(err, ShouldBeNil)
				So(total, ShouldEqual, 0)
				So(len(comments), ShouldEqual, 0)
			})

			Convey("计数查询失败", func() {
				mockey.Mock((*mongo.Collection).CountDocuments).Return(int64(0), mongo.WriteException{}).Build()

				comments, total, err := dao.List(ctx, filter, 1, 10)
				So(err, ShouldNotBeNil)
				So(total, ShouldEqual, 0)
				So(comments, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to count comments")
			})
		})

		Convey("GetByPostID方法测试", func() {
			postID := primitive.NewObjectID()
			postIDStr := postID.Hex()

			Convey("正常获取文章评论", func() {
				mockey.Mock((*mongo.Collection).CountDocuments).Return(int64(5), nil).Build()

				mockCursor := &mongo.Cursor{}
				mockey.Mock((*mongo.Collection).Find).Return(mockCursor, nil).Build()
				mockey.Mock((*mongo.Cursor).All).To(func(ctx context.Context, results interface{}) error {
					comments := results.(*[]*model.Comment)
					*comments = []*model.Comment{
						{
							ID:          primitive.NewObjectID(),
							PostID:      postID,
							Content:     "文章评论1",
							AuthorName:  "作者1",
							AuthorEmail: "test1@example.com",
							Status:      constants.CommentStatusApproved,
						},
					}
					return nil
				}).Build()
				mockey.Mock((*mongo.Cursor).Close).Return(nil).Build()

				comments, total, err := dao.GetByPostID(ctx, postIDStr, 1, 10)
				So(err, ShouldBeNil)
				So(total, ShouldEqual, 5)
				So(len(comments), ShouldEqual, 1)
				So(comments[0].PostID, ShouldEqual, postID)
			})

			Convey("无效文章ID", func() {
				comments, total, err := dao.GetByPostID(ctx, "invalid-id", 1, 10)
				So(err, ShouldNotBeNil)
				So(total, ShouldEqual, 0)
				So(comments, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid post ID")
			})
		})

		Convey("ApproveComment方法测试", func() {
			validID := primitive.NewObjectID()
			validIDStr := validID.Hex()

			Convey("正常审核通过", func() {
				mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
					MatchedCount:  1,
					ModifiedCount: 1,
				}, nil).Build()

				err := dao.ApproveComment(ctx, validIDStr)
				So(err, ShouldBeNil)
			})

			Convey("无效评论ID", func() {
				err := dao.ApproveComment(ctx, "invalid-id")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid comment ID")
			})

			Convey("评论不存在", func() {
				mockey.Mock((*mongo.Collection).UpdateOne).Return(&mongo.UpdateResult{
					MatchedCount: 0,
				}, nil).Build()

				err := dao.ApproveComment(ctx, validIDStr)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "comment not found")
			})
		})

		Convey("BatchApprove方法测试", func() {
			validIDs := []string{
				primitive.NewObjectID().Hex(),
				primitive.NewObjectID().Hex(),
			}

			Convey("正常批量审核", func() {
				mockey.Mock((*mongo.Collection).UpdateMany).Return(&mongo.UpdateResult{
					MatchedCount:  2,
					ModifiedCount: 2,
				}, nil).Build()

				err := dao.BatchApprove(ctx, validIDs)
				So(err, ShouldBeNil)
			})

			Convey("空ID列表", func() {
				err := dao.BatchApprove(ctx, []string{})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "no comment IDs provided")
			})

			Convey("包含无效ID", func() {
				invalidIDs := []string{"invalid-id", validIDs[0]}
				err := dao.BatchApprove(ctx, invalidIDs)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid comment ID")
			})

			Convey("没有找到评论", func() {
				mockey.Mock((*mongo.Collection).UpdateMany).Return(&mongo.UpdateResult{
					MatchedCount: 0,
				}, nil).Build()

				err := dao.BatchApprove(ctx, validIDs)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "no comments found")
			})
		})

		Convey("buildQuery方法测试", func() {
			Convey("空过滤器", func() {
				filter := model.CommentFilter{}
				query := dao.buildQuery(filter)
				So(query, ShouldNotBeNil)
				So(query["$or"], ShouldNotBeNil) // 软删除过滤条件
			})

			Convey("多条件过滤器", func() {
				postID := primitive.NewObjectID()
				filter := model.CommentFilter{
					PostID:      postID.Hex(),
					Status:      constants.CommentStatusApproved,
					AuthorEmail: "test@example.com",
					Keyword:     "测试",
					StartTime:   time.Now().Add(-24 * time.Hour),
					EndTime:     time.Now(),
				}
				query := dao.buildQuery(filter)
				So(query, ShouldNotBeNil)
				So(query["postId"], ShouldEqual, postID)
				So(query["status"], ShouldEqual, constants.CommentStatusApproved)
				So(query["authorEmail"], ShouldEqual, "test@example.com")
				So(query["createdAt"], ShouldNotBeNil)
			})
		})

		Convey("buildSort方法测试", func() {
			Convey("默认排序", func() {
				filter := model.CommentFilter{}
				sort := dao.buildSort(filter)
				So(sort, ShouldNotBeNil)
				So(sort["createdAt"], ShouldEqual, -1)
			})

			Convey("指定排序字段", func() {
				filter := model.CommentFilter{
					SortBy:   "replyCount",
					SortDesc: true,
				}
				sort := dao.buildSort(filter)
				So(sort, ShouldNotBeNil)
				So(sort["replyCount"], ShouldEqual, -1)
			})

			Convey("无效排序字段", func() {
				filter := model.CommentFilter{
					SortBy: "invalid_field",
				}
				sort := dao.buildSort(filter)
				So(sort, ShouldNotBeNil)
				So(sort["createdAt"], ShouldEqual, -1) // 回退到默认排序
			})
		})
	})
}
