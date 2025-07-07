package logic

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bytedance/mockey"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/dao"
	"github.com/heimdall-api/common/model"
	"github.com/heimdall-api/common/utils"
	"github.com/heimdall-api/public-api/public/internal/config"
	"github.com/heimdall-api/public-api/public/internal/svc"
	"github.com/heimdall-api/public-api/public/internal/types"
)

func TestGetPublicTagListLogic(t *testing.T) {
	Convey("Test GetPublicTagList Logic", t, func() {
		// 创建模拟的ServiceContext
		tagDAO := &dao.TagDAO{}
		serviceCtx := &svc.ServiceContext{
			Config: config.Config{},
			TagDAO: tagDAO,
		}

		// 创建Logic实例
		l := NewGetPublicTagListLogic(context.Background(), serviceCtx)

		Convey("Success - Get public tag list with default parameters", func() {
			// 重置mock
			mockey.UnPatchAll()

			// 准备测试数据
			mockTags := []*model.TagModel{
				{
					ID:          primitive.NewObjectID(),
					Name:        "Go语言",
					Slug:        "golang",
					Description: "Go编程语言相关文章",
					Color:       "#00ADD8",
					PostCount:   15,
					Visibility:  constants.TagVisibilityPublic,
					CreatedAt:   time.Now(),
				},
				{
					ID:          primitive.NewObjectID(),
					Name:        "微服务",
					Slug:        "microservices",
					Description: "微服务架构设计",
					Color:       "#FF6B6B",
					PostCount:   8,
					Visibility:  constants.TagVisibilityPublic,
					CreatedAt:   time.Now(),
				},
			}

			// Mock TagDAO.GetPublishedList方法
			mockey.Mock((*dao.TagDAO).GetPublishedList).To(func(tagDAO *dao.TagDAO, ctx context.Context, filter dao.TagFilter, page, limit int) ([]*model.TagModel, int64, error) {
				return mockTags, 2, nil
			}).Build()

			// 准备请求
			req := &types.PublicTagListRequest{
				Page:     1,
				Limit:    20,
				SortBy:   "postCount",
				SortDesc: true,
			}

			// 执行测试
			resp, err := l.GetPublicTagList(req)

			// 验证结果
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, utils.StatusOK)
			So(resp.Message, ShouldEqual, "Success")
			So(resp.Data.List, ShouldHaveLength, 2)
			So(resp.Data.Pagination.Total, ShouldEqual, 2)
			So(resp.Data.Pagination.Page, ShouldEqual, 1)
			So(resp.Data.Pagination.Limit, ShouldEqual, 20)
			So(resp.Data.Pagination.TotalPages, ShouldEqual, 1)
			So(resp.Data.Pagination.HasNext, ShouldBeFalse)
			So(resp.Data.Pagination.HasPrev, ShouldBeFalse)

			// 验证第一个标签数据
			firstTag := resp.Data.List[0]
			So(firstTag.Name, ShouldEqual, "Go语言")
			So(firstTag.Slug, ShouldEqual, "golang")
			So(firstTag.Description, ShouldEqual, "Go编程语言相关文章")
			So(firstTag.Color, ShouldEqual, "#00ADD8")
			So(firstTag.PostCount, ShouldEqual, 15)
		})

		Convey("Success - Get empty tag list", func() {
			mockey.UnPatchAll()
			
			// Mock返回空结果
			mockey.Mock((*dao.TagDAO).GetPublishedList).To(func(tagDAO *dao.TagDAO, ctx context.Context, filter dao.TagFilter, page, limit int) ([]*model.TagModel, int64, error) {
				return []*model.TagModel{}, 0, nil
			}).Build()

			req := &types.PublicTagListRequest{
				Page:     1,
				Limit:    20,
				SortBy:   "name",
				SortDesc: false,
			}

			resp, err := l.GetPublicTagList(req)

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, utils.StatusOK)
			So(resp.Data.List, ShouldHaveLength, 0)
			So(resp.Data.Pagination.Total, ShouldEqual, 0)
		})

		Convey("Success - Test pagination with large dataset", func() {
			mockey.UnPatchAll()
			
			// Mock返回大量数据的第二页
			mockTags := make([]*model.TagModel, 20)
			for i := 0; i < 20; i++ {
				mockTags[i] = &model.TagModel{
					ID:          primitive.NewObjectID(),
					Name:        "Tag " + string(rune(65+i)),
					Slug:        "tag-" + string(rune(97+i)),
					PostCount:   i + 1,
					Visibility:  constants.TagVisibilityPublic,
					CreatedAt:   time.Now(),
				}
			}

			mockey.Mock((*dao.TagDAO).GetPublishedList).To(func(tagDAO *dao.TagDAO, ctx context.Context, filter dao.TagFilter, page, limit int) ([]*model.TagModel, int64, error) {
				return mockTags, 55, nil // 总共55个标签
			}).Build()

			req := &types.PublicTagListRequest{
				Page:     2,
				Limit:    20,
				SortBy:   "createdAt",
				SortDesc: true,
			}

			resp, err := l.GetPublicTagList(req)

			So(err, ShouldBeNil)
			So(resp.Data.List, ShouldHaveLength, 20)
			So(resp.Data.Pagination.Total, ShouldEqual, 55)
			So(resp.Data.Pagination.Page, ShouldEqual, 2)
			So(resp.Data.Pagination.TotalPages, ShouldEqual, 3)
			So(resp.Data.Pagination.HasNext, ShouldBeTrue)
			So(resp.Data.Pagination.HasPrev, ShouldBeTrue)
		})

		Convey("Error - Invalid sort field", func() {
			mockey.UnPatchAll()
			
			req := &types.PublicTagListRequest{
				Page:     1,
				Limit:    20,
				SortBy:   "invalidField",
				SortDesc: true,
			}

			resp, err := l.GetPublicTagList(req)

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, utils.StatusBadRequest)
			So(resp.Message, ShouldEqual, "invalid sort field")
			So(resp.Data.List, ShouldHaveLength, 0)
		})

		Convey("Error - Database error", func() {
			mockey.UnPatchAll()
			
			// Mock数据库错误
			mockey.Mock((*dao.TagDAO).GetPublishedList).To(func(tagDAO *dao.TagDAO, ctx context.Context, filter dao.TagFilter, page, limit int) ([]*model.TagModel, int64, error) {
				return nil, 0, errors.New("database connection failed")
			}).Build()

			req := &types.PublicTagListRequest{
				Page:     1,
				Limit:    20,
				SortBy:   "postCount",
				SortDesc: true,
			}

			resp, err := l.GetPublicTagList(req)

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, utils.StatusInternalServerError)
			So(resp.Message, ShouldEqual, "Failed to get tag list")
			So(resp.Data.List, ShouldHaveLength, 0)
		})
	})
}

func TestGetPublicTagListLogic_validateRequest(t *testing.T) {
	Convey("Test validateRequest", t, func() {
		l := &GetPublicTagListLogic{}

		Convey("Valid requests", func() {
			validRequests := []*types.PublicTagListRequest{
				{SortBy: "name"},
				{SortBy: "postCount"},
				{SortBy: "createdAt"},
			}

			for _, req := range validRequests {
				err := l.validateRequest(req)
				So(err, ShouldBeNil)
			}
		})

		Convey("Invalid sort field", func() {
			req := &types.PublicTagListRequest{SortBy: "invalidField"}
			err := l.validateRequest(req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "invalid sort field")
		})
	})
}

func TestGetPublicTagListLogic_buildTagFilter(t *testing.T) {
	Convey("Test buildTagFilter", t, func() {
		l := &GetPublicTagListLogic{}

		Convey("Build filter with descending sort", func() {
			req := &types.PublicTagListRequest{
				SortBy:   "postCount",
				SortDesc: true,
			}

			filter := l.buildTagFilter(req)

			So(filter.Visibility, ShouldEqual, constants.TagVisibilityPublic)
			So(filter.SortBy, ShouldEqual, "postCount")
			So(filter.SortOrder, ShouldEqual, "desc")
		})

		Convey("Build filter with ascending sort", func() {
			req := &types.PublicTagListRequest{
				SortBy:   "name",
				SortDesc: false,
			}

			filter := l.buildTagFilter(req)

			So(filter.Visibility, ShouldEqual, constants.TagVisibilityPublic)
			So(filter.SortBy, ShouldEqual, "name")
			So(filter.SortOrder, ShouldEqual, "asc")
		})
	})
}

func TestGetPublicTagListLogic_buildTagInfos(t *testing.T) {
	Convey("Test buildTagInfos", t, func() {
		l := &GetPublicTagListLogic{}

		Convey("Build tag infos from non-empty list", func() {
			mockTags := []*model.TagModel{
				{
					ID:          primitive.NewObjectID(),
					Name:        "测试标签",
					Slug:        "test-tag",
					Description: "这是一个测试标签",
					Color:       "#FF0000",
					PostCount:   5,
					CreatedAt:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				},
			}

			result := l.buildTagInfos(mockTags)

			So(result, ShouldHaveLength, 1)
			So(result[0].Name, ShouldEqual, "测试标签")
			So(result[0].Slug, ShouldEqual, "test-tag")
			So(result[0].Description, ShouldEqual, "这是一个测试标签")
			So(result[0].Color, ShouldEqual, "#FF0000")
			So(result[0].PostCount, ShouldEqual, 5)
			So(result[0].ID, ShouldNotBeEmpty)
			So(result[0].CreatedAt, ShouldEqual, "2024-01-01T12:00:00Z")
		})

		Convey("Build tag infos from empty list", func() {
			result := l.buildTagInfos([]*model.TagModel{})
			So(result, ShouldHaveLength, 0)
		})

		Convey("Build tag infos from nil list", func() {
			result := l.buildTagInfos(nil)
			So(result, ShouldHaveLength, 0)
		})
	})
}

func TestGetPublicTagListLogic_buildPagination(t *testing.T) {
	Convey("Test buildPagination", t, func() {
		l := &GetPublicTagListLogic{}

		Convey("First page with next", func() {
			pagination := l.buildPagination(1, 10, 25)

			So(pagination.Page, ShouldEqual, 1)
			So(pagination.Limit, ShouldEqual, 10)
			So(pagination.Total, ShouldEqual, 25)
			So(pagination.TotalPages, ShouldEqual, 3)
			So(pagination.HasNext, ShouldBeTrue)
			So(pagination.HasPrev, ShouldBeFalse)
		})

		Convey("Middle page", func() {
			pagination := l.buildPagination(2, 10, 25)

			So(pagination.Page, ShouldEqual, 2)
			So(pagination.TotalPages, ShouldEqual, 3)
			So(pagination.HasNext, ShouldBeTrue)
			So(pagination.HasPrev, ShouldBeTrue)
		})

		Convey("Last page", func() {
			pagination := l.buildPagination(3, 10, 25)

			So(pagination.Page, ShouldEqual, 3)
			So(pagination.TotalPages, ShouldEqual, 3)
			So(pagination.HasNext, ShouldBeFalse)
			So(pagination.HasPrev, ShouldBeTrue)
		})

		Convey("Single page", func() {
			pagination := l.buildPagination(1, 10, 5)

			So(pagination.Page, ShouldEqual, 1)
			So(pagination.TotalPages, ShouldEqual, 1)
			So(pagination.HasNext, ShouldBeFalse)
			So(pagination.HasPrev, ShouldBeFalse)
		})

		Convey("Empty result", func() {
			pagination := l.buildPagination(1, 10, 0)

			So(pagination.Page, ShouldEqual, 1)
			So(pagination.Total, ShouldEqual, 0)
			So(pagination.TotalPages, ShouldEqual, 0)
			So(pagination.HasNext, ShouldBeFalse)
			So(pagination.HasPrev, ShouldBeFalse)
		})
	})
}

func TestGetPublicTagListLogic_isValidSortBy(t *testing.T) {
	Convey("Test isValidSortBy", t, func() {
		l := &GetPublicTagListLogic{}

		Convey("Valid sort fields", func() {
			validFields := []string{"name", "postCount", "createdAt"}
			for _, field := range validFields {
				result := l.isValidSortBy(field)
				So(result, ShouldBeTrue)
			}
		})

		Convey("Invalid sort fields", func() {
			invalidFields := []string{"", "invalid", "updatedAt", "id", "slug"}
			for _, field := range invalidFields {
				result := l.isValidSortBy(field)
				So(result, ShouldBeFalse)
			}
		})
	})
}