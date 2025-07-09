package dao

import (
	"context"
	"testing"

	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/model"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestTagDAO(t *testing.T) {
	Convey("TagDAO测试", t, func() {

		var tagDAO *TagDAO
		var ctx context.Context

		Reset(func() {
			ctx = context.Background()
			// 使用空的collection进行基本测试
			tagDAO = &TagDAO{collection: &mongo.Collection{}}
		})

		Convey("NewTagDAO", func() {
			Convey("应该正确创建TagDAO实例", func() {
				// 仅测试构造函数不为nil
				So(NewTagDAO, ShouldNotBeNil)
			})
		})

		Convey("参数验证测试", func() {

			Convey("Create方法参数验证", func() {
				Convey("空标签应该失败", func() {
					err := tagDAO.Create(ctx, nil)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "tag cannot be nil")
				})

				Convey("标签验证失败应该返回错误", func() {
					invalidTag := &model.TagModel{
						Name: "", // 空名称
						Slug: "test",
					}

					err := tagDAO.Create(ctx, invalidTag)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "tag name cannot be empty")
				})
			})

			Convey("GetByID方法参数验证", func() {
				Convey("空ID应该失败", func() {
					tag, err := tagDAO.GetByID(ctx, "")
					So(err, ShouldNotBeNil)
					So(tag, ShouldBeNil)
					So(err.Error(), ShouldContainSubstring, "tag id cannot be empty")
				})

				Convey("无效ID格式应该失败", func() {
					tag, err := tagDAO.GetByID(ctx, "invalid-id")
					So(err, ShouldNotBeNil)
					So(tag, ShouldBeNil)
					So(err.Error(), ShouldContainSubstring, "invalid tag id format")
				})
			})

			Convey("GetBySlug方法参数验证", func() {
				Convey("空slug应该失败", func() {
					tag, err := tagDAO.GetBySlug(ctx, "")
					So(err, ShouldNotBeNil)
					So(tag, ShouldBeNil)
					So(err.Error(), ShouldContainSubstring, "tag slug cannot be empty")
				})
			})

			Convey("Update方法参数验证", func() {
				Convey("空ID应该失败", func() {
					updates := map[string]interface{}{"name": "test"}
					err := tagDAO.Update(ctx, "", updates)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "tag id cannot be empty")
				})

				Convey("空更新内容应该失败", func() {
					tagID := primitive.NewObjectID()
					err := tagDAO.Update(ctx, tagID.Hex(), nil)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "no updates provided")
				})

				Convey("无效ID格式应该失败", func() {
					updates := map[string]interface{}{"name": "test"}
					err := tagDAO.Update(ctx, "invalid-id", updates)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "invalid tag id format")
				})
			})

			Convey("Delete方法参数验证", func() {
				Convey("空ID应该失败", func() {
					err := tagDAO.Delete(ctx, "")
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "tag id cannot be empty")
				})

				Convey("无效ID格式应该失败", func() {
					err := tagDAO.Delete(ctx, "invalid-id")
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "invalid tag id format")
				})
			})

			Convey("IncrementPostCount方法参数验证", func() {
				Convey("空ID应该失败", func() {
					err := tagDAO.IncrementPostCount(ctx, "", 1)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "tag id cannot be empty")
				})

				Convey("无效ID格式应该失败", func() {
					err := tagDAO.IncrementPostCount(ctx, "invalid-id", 1)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "invalid tag id format")
				})
			})

			Convey("UpdatePostCount方法参数验证", func() {
				Convey("空ID应该失败", func() {
					err := tagDAO.UpdatePostCount(ctx, "", 25)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "tag id cannot be empty")
				})

				Convey("无效ID格式应该失败", func() {
					err := tagDAO.UpdatePostCount(ctx, "invalid-id", 25)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "invalid tag id format")
				})
			})
		})

		Convey("查询构建方法测试", func() {

			Convey("buildQuery方法", func() {

				Convey("空过滤器应该只排除已删除的标签", func() {
					filter := TagFilter{}
					query := tagDAO.buildQuery(filter)

					So(query["deletedAt"], ShouldResemble, bson.M{"$exists": false})
					So(len(query), ShouldEqual, 1)
				})

				Convey("名称过滤应该使用正则表达式", func() {
					filter := TagFilter{Name: "Go"}
					query := tagDAO.buildQuery(filter)

					So(query["name"], ShouldResemble, bson.M{"$regex": "Go", "$options": "i"})
					So(query["deletedAt"], ShouldResemble, bson.M{"$exists": false})
				})

				Convey("可见性过滤应该设置精确匹配", func() {
					filter := TagFilter{Visibility: constants.TagVisibilityPublic}
					query := tagDAO.buildQuery(filter)

					So(query["visibility"], ShouldEqual, constants.TagVisibilityPublic)
					So(query["deletedAt"], ShouldResemble, bson.M{"$exists": false})
				})

				Convey("无效可见性应该被忽略", func() {
					filter := TagFilter{Visibility: "invalid"}
					query := tagDAO.buildQuery(filter)

					So(query["visibility"], ShouldBeNil)
					So(query["deletedAt"], ShouldResemble, bson.M{"$exists": false})
				})

				Convey("组合过滤条件", func() {
					filter := TagFilter{
						Name:       "Go",
						Visibility: constants.TagVisibilityPublic,
					}
					query := tagDAO.buildQuery(filter)

					So(query["name"], ShouldResemble, bson.M{"$regex": "Go", "$options": "i"})
					So(query["visibility"], ShouldEqual, constants.TagVisibilityPublic)
					So(query["deletedAt"], ShouldResemble, bson.M{"$exists": false})
					So(len(query), ShouldEqual, 3)
				})
			})

			Convey("buildSort方法", func() {

				Convey("默认排序应该按创建时间降序", func() {
					sort := tagDAO.buildSort("", "")
					So(sort, ShouldResemble, bson.D{{"createdAt", -1}})
				})

				Convey("按名称升序排序", func() {
					sort := tagDAO.buildSort(constants.TagSortByName, "asc")
					So(sort, ShouldResemble, bson.D{{"name", 1}})
				})

				Convey("按名称降序排序", func() {
					sort := tagDAO.buildSort(constants.TagSortByName, "desc")
					So(sort, ShouldResemble, bson.D{{"name", -1}})
				})

				Convey("按slug升序排序", func() {
					sort := tagDAO.buildSort(constants.TagSortBySlug, "asc")
					So(sort, ShouldResemble, bson.D{{"slug", 1}})
				})

				Convey("按文章数量排序应该包含二级排序", func() {
					sort := tagDAO.buildSort(constants.TagSortByPostCount, "desc")
					So(sort, ShouldResemble, bson.D{{"postCount", -1}, {"createdAt", -1}})
				})

				Convey("按创建时间升序排序", func() {
					sort := tagDAO.buildSort(constants.TagSortByCreatedAt, "asc")
					So(sort, ShouldResemble, bson.D{{"createdAt", 1}})
				})

				Convey("按更新时间降序排序", func() {
					sort := tagDAO.buildSort(constants.TagSortByUpdatedAt, "desc")
					So(sort, ShouldResemble, bson.D{{"updatedAt", -1}})
				})

				Convey("无效排序字段应该使用默认排序", func() {
					sort := tagDAO.buildSort("invalid", "asc")
					So(sort, ShouldResemble, bson.D{{"createdAt", -1}})
				})

				Convey("默认排序方向应该是降序", func() {
					sort := tagDAO.buildSort(constants.TagSortByName, "")
					So(sort, ShouldResemble, bson.D{{"name", -1}})
				})
			})
		})

		Convey("数据验证方法测试", func() {

			Convey("validateUpdateFields方法", func() {

				Convey("禁止更新的字段应该失败", func() {
					forbiddenFields := []string{"_id", "createdAt", "deletedAt", "postCount"}

					for _, field := range forbiddenFields {
						updates := map[string]interface{}{field: "test"}
						err := tagDAO.validateUpdateFields(updates)
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldContainSubstring, "cannot update field: "+field)
					}
				})

				Convey("有效字段应该通过验证", func() {
					updates := map[string]interface{}{
						"name":            "Valid Name",
						"slug":            "valid-slug",
						"description":     "Valid description",
						"color":           "#FF5733",
						"visibility":      constants.TagVisibilityPublic,
						"featuredImage":   "https://example.com/image.jpg",
						"metaTitle":       "Valid Meta Title",
						"metaDescription": "Valid meta description",
					}

					err := tagDAO.validateUpdateFields(updates)
					So(err, ShouldBeNil)
				})

				Convey("无效的name字段应该失败", func() {
					invalidNames := []interface{}{"", 123, true, nil}

					for _, invalidName := range invalidNames {
						updates := map[string]interface{}{"name": invalidName}
						err := tagDAO.validateUpdateFields(updates)
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldContainSubstring, "invalid name field")
					}
				})

				Convey("无效的slug字段应该失败", func() {
					invalidSlugs := []interface{}{"", 123, true, nil}

					for _, invalidSlug := range invalidSlugs {
						updates := map[string]interface{}{"slug": invalidSlug}
						err := tagDAO.validateUpdateFields(updates)
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldContainSubstring, "invalid slug field")
					}
				})

				Convey("无效的visibility字段应该失败", func() {
					invalidVisibilities := []interface{}{"invalid", 123, true, nil}

					for _, invalidVisibility := range invalidVisibilities {
						updates := map[string]interface{}{"visibility": invalidVisibility}
						err := tagDAO.validateUpdateFields(updates)
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldContainSubstring, "invalid visibility field")
					}
				})

				Convey("有效的visibility字段应该通过", func() {
					validVisibilities := []string{
						constants.TagVisibilityPublic,
						constants.TagVisibilityInternal,
					}

					for _, validVisibility := range validVisibilities {
						updates := map[string]interface{}{"visibility": validVisibility}
						err := tagDAO.validateUpdateFields(updates)
						So(err, ShouldBeNil)
					}
				})

				Convey("空更新字段对象应该通过", func() {
					updates := map[string]interface{}{}
					err := tagDAO.validateUpdateFields(updates)
					So(err, ShouldBeNil)
				})
			})
		})

		Convey("List方法分页参数处理", func() {
			Convey("方法存在性验证", func() {
				// 仅验证方法存在
				So(tagDAO.List, ShouldNotBeNil)
			})
		})

		Convey("GetPublishedList方法", func() {
			Convey("方法存在性验证", func() {
				// 仅验证方法存在
				So(tagDAO.GetPublishedList, ShouldNotBeNil)
			})
		})

		Convey("GetPopularTags方法", func() {
			Convey("方法存在性验证", func() {
				// 仅验证方法存在
				So(tagDAO.GetPopularTags, ShouldNotBeNil)
			})
		})
	})
}

func TestTagFilter(t *testing.T) {
	Convey("TagFilter结构体测试", t, func() {

		Convey("TagFilter字段设置", func() {
			filter := TagFilter{
				Name:       "Go语言",
				Visibility: constants.TagVisibilityPublic,
				SortBy:     constants.TagSortByName,
				SortOrder:  "asc",
			}

			So(filter.Name, ShouldEqual, "Go语言")
			So(filter.Visibility, ShouldEqual, constants.TagVisibilityPublic)
			So(filter.SortBy, ShouldEqual, constants.TagSortByName)
			So(filter.SortOrder, ShouldEqual, "asc")
		})

		Convey("空TagFilter应该有零值", func() {
			filter := TagFilter{}

			So(filter.Name, ShouldEqual, "")
			So(filter.Visibility, ShouldEqual, "")
			So(filter.SortBy, ShouldEqual, "")
			So(filter.SortOrder, ShouldEqual, "")
		})
	})
}
