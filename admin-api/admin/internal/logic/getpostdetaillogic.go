package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/heimdall-api/common/model"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetPostDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取文章详情
func NewGetPostDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPostDetailLogic {
	return &GetPostDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPostDetailLogic) GetPostDetail(req *types.PostDetailRequest) (resp *types.PostDetailResponse, err error) {
	// 1. 验证文章ID格式
	if !primitive.IsValidObjectID(req.ID) {
		return nil, fmt.Errorf("无效的文章ID")
	}

	// 2. 查询文章详情
	post, err := l.svcCtx.PostDAO.GetByID(l.ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("文章不存在: %v", err)
	}

	// 3. 获取作者信息
	author, err := l.svcCtx.UserDAO.GetByID(l.ctx, post.AuthorID.Hex())
	if err != nil {
		return nil, fmt.Errorf("获取作者信息失败: %v", err)
	}

	// 4. 构建文章详情数据
	authorInfo := author.ToAuthorInfo()
	postDetailData := l.buildPostDetailData(post, authorInfo)

	// 5. 构建响应
	return &types.PostDetailResponse{
		Code:      200,
		Message:   "获取文章详情成功",
		Data:      *postDetailData,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// buildPostDetailData 构建文章详情数据
func (l *GetPostDetailLogic) buildPostDetailData(post *model.Post, author *model.AuthorInfo) *types.PostDetailData {
	// 转换标签
	tags := make([]types.TagInfo, len(post.Tags))
	for i, tag := range post.Tags {
		tags[i] = types.TagInfo{
			Name: tag.Name,
			Slug: tag.Slug,
		}
	}

	// 转换作者信息
	authorInfo := types.AuthorInfo{
		ID:           author.ID,
		Username:     author.Username,
		DisplayName:  author.DisplayName,
		ProfileImage: author.ProfileImage,
		Bio:          author.Bio,
	}

	// 格式化发布时间
	var publishedAt string
	if post.PublishedAt != nil {
		publishedAt = post.PublishedAt.Format(time.RFC3339)
	}

	return &types.PostDetailData{
		ID:              post.ID.Hex(),
		Title:           post.Title,
		Slug:            post.Slug,
		Excerpt:         post.Excerpt,
		Markdown:        post.Markdown,
		HTML:            post.HTML,
		FeaturedImage:   post.FeaturedImage,
		Type:            post.Type,
		Status:          post.Status,
		Visibility:      post.Visibility,
		Author:          authorInfo,
		Tags:            tags,
		MetaTitle:       post.MetaTitle,
		MetaDescription: post.MetaDescription,
		CanonicalURL:    post.CanonicalURL,
		ReadingTime:     post.ReadingTime,
		WordCount:       post.WordCount,
		ViewCount:       post.ViewCount,
		PublishedAt:     publishedAt,
		CreatedAt:       post.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       post.UpdatedAt.Format(time.RFC3339),
	}
}
