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

func TestGetPublicTagDetailLogic(t *testing.T) {
	Convey("Test GetPublicTagDetail Logic", t, func() {
		// 创建模拟的ServiceContext
		tagDAO := &dao.TagDAO{}
		postDAO := &dao.PostDAO{}
		userDAO := &dao.UserDAO{}
		serviceCtx := &svc.ServiceContext{
			Config:  config.Config{},
			TagDAO:  tagDAO,
			PostDAO: postDAO,
			UserDAO: userDAO,
		}

		// 创建Logic实例
		l := NewGetPublicTagDetailLogic(context.Background(), serviceCtx)

		Convey("Success - Get public tag detail with recent posts", func() {
			mockey.UnPatchAll()

			// 准备标签数据
			tagID := primitive.NewObjectID()
			mockTag := &model.TagModel{
				ID:          tagID,
				Name:        "Go语言",
				Slug:        "golang",
				Description: "Go编程语言相关技术文章",
				Color:       "#00ADD8",
				PostCount:   15,
				Visibility:  constants.TagVisibilityPublic,
				CreatedAt:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			}

			// 准备文章数据
			authorID := primitive.NewObjectID()
			publishedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
			mockPosts := []*model.Post{
				{
					ID:            primitive.NewObjectID(),
					Title:         "Go语言并发编程",
					Slug:          "go-concurrency",
					Excerpt:       "深入理解Go语言的并发特性",
					FeaturedImage: "https://example.com/image1.jpg",
					AuthorID:      authorID,
					Tags: []model.Tag{
						{Name: "Go语言", Slug: "golang"},
						{Name: "并发", Slug: "concurrency"},
					},
					ReadingTime: 10,
					ViewCount:   1200,
					Status:      constants.PostStatusPublished,
					PublishedAt: &publishedTime,
					UpdatedAt:   time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
				},
			}

			// 准备作者数据
			mockUser := &model.User{
				ID:           authorID,
				Username:     "gopher",
				DisplayName:  "Go开发者",
				ProfileImage: "https://example.com/avatar.jpg",
				Bio:          "资深Go语言开发工程师",
			}

			// Mock TagDAO.GetBySlug方法
			mockey.Mock((*dao.TagDAO).GetBySlug).To(func(tagDAO *dao.TagDAO, ctx context.Context, slug string) (*model.TagModel, error) {
				return mockTag, nil
			}).Build()

			// Mock PostDAO.GetPublishedList方法
			mockey.Mock((*dao.PostDAO).GetPublishedList).To(func(postDAO *dao.PostDAO, ctx context.Context, filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
				return mockPosts, 1, nil
			}).Build()

			// Mock UserDAO.GetByID方法
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return mockUser, nil
			}).Build()

			// 准备请求
			req := &types.PublicTagDetailRequest{
				Slug: "golang",
			}

			// 执行测试
			resp, err := l.GetPublicTagDetail(req)

			// 验证结果
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, utils.StatusOK)
			So(resp.Message, ShouldEqual, "Success")

			// 验证标签详情
			tagDetail := resp.Data
			So(tagDetail.ID, ShouldEqual, tagID.Hex())
			So(tagDetail.Name, ShouldEqual, "Go语言")
			So(tagDetail.Slug, ShouldEqual, "golang")
			So(tagDetail.Description, ShouldEqual, "Go编程语言相关技术文章")
			So(tagDetail.Color, ShouldEqual, "#00ADD8")
			So(tagDetail.PostCount, ShouldEqual, 15)
			So(tagDetail.CreatedAt, ShouldEqual, "2024-01-01T12:00:00Z")

			// 验证相关文章
			So(tagDetail.RecentPosts, ShouldHaveLength, 1)
			post := tagDetail.RecentPosts[0]
			So(post.Title, ShouldEqual, "Go语言并发编程")
			So(post.Slug, ShouldEqual, "go-concurrency")
			So(post.Author.Username, ShouldEqual, "gopher")
			So(post.Author.DisplayName, ShouldEqual, "Go开发者")
			So(post.Tags, ShouldHaveLength, 2)
			So(post.ViewCount, ShouldEqual, 1200)
		})

		Convey("Success - Get tag detail without recent posts", func() {
			mockey.UnPatchAll()

			mockTag := &model.TagModel{
				ID:          primitive.NewObjectID(),
				Name:        "新标签",
				Slug:        "new-tag",
				Description: "这是一个新标签",
				Color:       "#FF0000",
				PostCount:   0,
				Visibility:  constants.TagVisibilityPublic,
				CreatedAt:   time.Now(),
			}

			// Mock标签存在但没有文章
			mockey.Mock((*dao.TagDAO).GetBySlug).To(func(tagDAO *dao.TagDAO, ctx context.Context, slug string) (*model.TagModel, error) {
				return mockTag, nil
			}).Build()

			mockey.Mock((*dao.PostDAO).GetPublishedList).To(func(postDAO *dao.PostDAO, ctx context.Context, filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
				return []*model.Post{}, 0, nil
			}).Build()

			req := &types.PublicTagDetailRequest{Slug: "new-tag"}
			resp, err := l.GetPublicTagDetail(req)

			So(err, ShouldBeNil)
			So(resp.Code, ShouldEqual, utils.StatusOK)
			So(resp.Data.Name, ShouldEqual, "新标签")
			So(resp.Data.PostCount, ShouldEqual, 0)
			So(resp.Data.RecentPosts, ShouldHaveLength, 0)
		})

		Convey("Success - Get tag detail but posts query fails", func() {
			mockey.UnPatchAll()

			mockTag := &model.TagModel{
				ID:         primitive.NewObjectID(),
				Name:       "测试标签",
				Slug:       "test-tag",
				Visibility: constants.TagVisibilityPublic,
				CreatedAt:  time.Now(),
			}

			mockey.Mock((*dao.TagDAO).GetBySlug).To(func(tagDAO *dao.TagDAO, ctx context.Context, slug string) (*model.TagModel, error) {
				return mockTag, nil
			}).Build()

			// Mock文章查询失败
			mockey.Mock((*dao.PostDAO).GetPublishedList).To(func(postDAO *dao.PostDAO, ctx context.Context, filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
				return nil, 0, errors.New("database error")
			}).Build()

			req := &types.PublicTagDetailRequest{Slug: "test-tag"}
			resp, err := l.GetPublicTagDetail(req)

			// 即使文章查询失败，标签详情仍应该成功返回
			So(err, ShouldBeNil)
			So(resp.Code, ShouldEqual, utils.StatusOK)
			So(resp.Data.Name, ShouldEqual, "测试标签")
			So(resp.Data.RecentPosts, ShouldHaveLength, 0)
		})

		Convey("Error - Empty slug", func() {
			mockey.UnPatchAll()

			req := &types.PublicTagDetailRequest{Slug: ""}
			resp, err := l.GetPublicTagDetail(req)

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, utils.StatusBadRequest)
			So(resp.Message, ShouldEqual, "tag slug cannot be empty")
		})

		Convey("Error - Tag not found", func() {
			mockey.UnPatchAll()

			// Mock标签不存在
			mockey.Mock((*dao.TagDAO).GetBySlug).To(func(tagDAO *dao.TagDAO, ctx context.Context, slug string) (*model.TagModel, error) {
				return nil, errors.New("tag not found")
			}).Build()

			req := &types.PublicTagDetailRequest{Slug: "non-existent"}
			resp, err := l.GetPublicTagDetail(req)

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, utils.StatusNotFound)
			So(resp.Message, ShouldEqual, "Tag not found")
		})

		Convey("Error - Database error when getting tag", func() {
			mockey.UnPatchAll()

			// Mock数据库错误
			mockey.Mock((*dao.TagDAO).GetBySlug).To(func(tagDAO *dao.TagDAO, ctx context.Context, slug string) (*model.TagModel, error) {
				return nil, errors.New("database connection failed")
			}).Build()

			req := &types.PublicTagDetailRequest{Slug: "golang"}
			resp, err := l.GetPublicTagDetail(req)

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, utils.StatusNotFound)
			So(resp.Message, ShouldEqual, "Tag not found")
		})

		Convey("Error - Tag is not public", func() {
			mockey.UnPatchAll()

			// 准备私有标签
			mockTag := &model.TagModel{
				ID:         primitive.NewObjectID(),
				Name:       "私有标签",
				Slug:       "private-tag",
				Visibility: constants.TagVisibilityInternal,
				CreatedAt:  time.Now(),
			}

			mockey.Mock((*dao.TagDAO).GetBySlug).To(func(tagDAO *dao.TagDAO, ctx context.Context, slug string) (*model.TagModel, error) {
				return mockTag, nil
			}).Build()

			req := &types.PublicTagDetailRequest{Slug: "private-tag"}
			resp, err := l.GetPublicTagDetail(req)

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, utils.StatusNotFound)
			So(resp.Message, ShouldEqual, "Tag not found")
		})
	})
}

func TestGetPublicTagDetailLogic_validateRequest(t *testing.T) {
	Convey("Test validateRequest", t, func() {
		l := &GetPublicTagDetailLogic{}

		Convey("Valid request", func() {
			req := &types.PublicTagDetailRequest{Slug: "golang"}
			err := l.validateRequest(req)
			So(err, ShouldBeNil)
		})

		Convey("Empty slug", func() {
			req := &types.PublicTagDetailRequest{Slug: ""}
			err := l.validateRequest(req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "tag slug cannot be empty")
		})
	})
}

func TestGetPublicTagDetailLogic_getRecentPostsByTag(t *testing.T) {
	Convey("Test getRecentPostsByTag", t, func() {
		postDAO := &dao.PostDAO{}
		userDAO := &dao.UserDAO{}
		serviceCtx := &svc.ServiceContext{
			PostDAO: postDAO,
			UserDAO: userDAO,
		}
		l := NewGetPublicTagDetailLogic(context.Background(), serviceCtx)

		Convey("Success - Get recent posts", func() {
			mockey.UnPatchAll()

			publishedTime := time.Now()
			mockPosts := []*model.Post{
				{
					ID:          primitive.NewObjectID(),
					Title:       "测试文章1",
					Slug:        "test-post-1",
					Excerpt:     "这是测试文章1",
					Status:      constants.PostStatusPublished,
					PublishedAt: &publishedTime,
				},
			}

			mockey.Mock((*dao.PostDAO).GetPublishedList).To(func(postDAO *dao.PostDAO, ctx context.Context, filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
				return mockPosts, 1, nil
			}).Build()

			// Mock UserDAO as well since buildPostListItems will call it
			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return &model.User{Username: "testuser", DisplayName: "Test User"}, nil
			}).Build()

			posts, err := l.getRecentPostsByTag("golang")

			So(err, ShouldBeNil)
			So(posts, ShouldHaveLength, 1)
		})

		Convey("Error - Database error", func() {
			mockey.UnPatchAll()

			mockey.Mock((*dao.PostDAO).GetPublishedList).To(func(postDAO *dao.PostDAO, ctx context.Context, filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
				return nil, 0, errors.New("database error")
			}).Build()

			posts, err := l.getRecentPostsByTag("golang")

			So(err, ShouldNotBeNil)
			So(posts, ShouldBeNil)
		})
	})
}

func TestGetPublicTagDetailLogic_buildTagDetail(t *testing.T) {
	Convey("Test buildTagDetail", t, func() {
		l := &GetPublicTagDetailLogic{}

		Convey("Build tag detail with posts", func() {
			tag := &model.TagModel{
				ID:          primitive.NewObjectID(),
				Name:        "测试标签",
				Slug:        "test-tag",
				Description: "这是一个测试标签",
				Color:       "#FF0000",
				PostCount:   5,
				CreatedAt:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			}

			recentPosts := []types.PublicPostListItem{
				{
					Title: "测试文章",
					Slug:  "test-post",
				},
			}

			result := l.buildTagDetail(tag, recentPosts)

			So(result.Name, ShouldEqual, "测试标签")
			So(result.Slug, ShouldEqual, "test-tag")
			So(result.Description, ShouldEqual, "这是一个测试标签")
			So(result.Color, ShouldEqual, "#FF0000")
			So(result.PostCount, ShouldEqual, 5)
			So(result.CreatedAt, ShouldEqual, "2024-01-01T12:00:00Z")
			So(result.RecentPosts, ShouldHaveLength, 1)
		})

		Convey("Build tag detail without posts", func() {
			tag := &model.TagModel{
				ID:        primitive.NewObjectID(),
				Name:      "空标签",
				Slug:      "empty-tag",
				PostCount: 0,
				CreatedAt: time.Now(),
			}

			result := l.buildTagDetail(tag, []types.PublicPostListItem{})

			So(result.Name, ShouldEqual, "空标签")
			So(result.PostCount, ShouldEqual, 0)
			So(result.RecentPosts, ShouldHaveLength, 0)
		})
	})
}

func TestGetPublicTagDetailLogic_buildPostListItems(t *testing.T) {
	Convey("Test buildPostListItems", t, func() {
		userDAO := &dao.UserDAO{}
		serviceCtx := &svc.ServiceContext{UserDAO: userDAO}
		l := NewGetPublicTagDetailLogic(context.Background(), serviceCtx)

		Convey("Build post list items successfully", func() {
			mockey.UnPatchAll()

			authorID := primitive.NewObjectID()
			publishedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
			mockPosts := []*model.Post{
				{
					ID:          primitive.NewObjectID(),
					Title:       "测试文章",
					Slug:        "test-post",
					Excerpt:     "这是一个测试文章",
					AuthorID:    authorID,
					Tags:        []model.Tag{{Name: "Go", Slug: "go"}},
					ReadingTime: 5,
					ViewCount:   100,
					PublishedAt: &publishedTime,
					UpdatedAt:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				},
			}

			mockUser := &model.User{
				ID:          authorID,
				Username:    "testuser",
				DisplayName: "测试用户",
			}

			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return mockUser, nil
			}).Build()

			result := l.buildPostListItems(mockPosts)

			So(result, ShouldHaveLength, 1)
			So(result[0].Title, ShouldEqual, "测试文章")
			So(result[0].Slug, ShouldEqual, "test-post")
			So(result[0].Author.Username, ShouldEqual, "testuser")
			So(result[0].Tags, ShouldHaveLength, 1)
			So(result[0].ViewCount, ShouldEqual, 100)
		})

		Convey("Handle empty post list", func() {
			result := l.buildPostListItems([]*model.Post{})
			So(result, ShouldHaveLength, 0)
		})

		Convey("Handle nil post list", func() {
			result := l.buildPostListItems(nil)
			So(result, ShouldHaveLength, 0)
		})
	})
}

func TestGetPublicTagDetailLogic_getAuthorInfo(t *testing.T) {
	Convey("Test getAuthorInfo", t, func() {
		userDAO := &dao.UserDAO{}
		serviceCtx := &svc.ServiceContext{UserDAO: userDAO}
		l := NewGetPublicTagDetailLogic(context.Background(), serviceCtx)

		Convey("Success - Get author info", func() {
			mockey.UnPatchAll()

			mockUser := &model.User{
				Username:     "gopher",
				DisplayName:  "Go开发者",
				ProfileImage: "https://example.com/avatar.jpg",
				Bio:          "专业Go开发工程师",
			}

			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return mockUser, nil
			}).Build()

			result := l.getAuthorInfo("123")

			So(result.Username, ShouldEqual, "gopher")
			So(result.DisplayName, ShouldEqual, "Go开发者")
			So(result.ProfileImage, ShouldEqual, "https://example.com/avatar.jpg")
			So(result.Bio, ShouldEqual, "专业Go开发工程师")
		})

		Convey("Error - Author not found", func() {
			mockey.UnPatchAll()

			mockey.Mock((*dao.UserDAO).GetByID).To(func(userDAO *dao.UserDAO, ctx context.Context, id string) (*model.User, error) {
				return nil, errors.New("user not found")
			}).Build()

			result := l.getAuthorInfo("nonexistent")

			So(result.Username, ShouldEqual, "unknown")
			So(result.DisplayName, ShouldEqual, "Unknown Author")
		})
	})
}

func TestGetPublicTagDetailLogic_buildTagInfosFromPost(t *testing.T) {
	Convey("Test buildTagInfosFromPost", t, func() {
		l := &GetPublicTagDetailLogic{}

		Convey("Build tag infos from post tags", func() {
			postTags := []model.Tag{
				{Name: "Go语言", Slug: "golang"},
				{Name: "微服务", Slug: "microservices"},
			}

			result := l.buildTagInfosFromPost(postTags)

			So(result, ShouldHaveLength, 2)
			So(result[0].Name, ShouldEqual, "Go语言")
			So(result[0].Slug, ShouldEqual, "golang")
			So(result[1].Name, ShouldEqual, "微服务")
			So(result[1].Slug, ShouldEqual, "microservices")
		})

		Convey("Handle empty tag list", func() {
			result := l.buildTagInfosFromPost([]model.Tag{})
			So(result, ShouldHaveLength, 0)
		})

		Convey("Handle nil tag list", func() {
			result := l.buildTagInfosFromPost(nil)
			So(result, ShouldHaveLength, 0)
		})
	})
}