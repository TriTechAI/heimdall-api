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

type GetPageDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取页面详情
func NewGetPageDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPageDetailLogic {
	return &GetPageDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPageDetailLogic) GetPageDetail(req *types.PageDetailRequest) (resp *types.PageDetailResponse, err error) {
	// 1. 验证页面ID格式
	if !primitive.IsValidObjectID(req.ID) {
		return nil, fmt.Errorf("无效的页面ID")
	}

	// 2. 查询页面详情
	page, err := l.svcCtx.PageDAO.GetByID(l.ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("页面不存在: %v", err)
	}

	// 3. 获取作者信息
	author, err := l.svcCtx.UserDAO.GetByID(l.ctx, page.AuthorID.Hex())
	if err != nil {
		return nil, fmt.Errorf("获取作者信息失败: %v", err)
	}

	// 4. 构建页面详情数据
	authorInfo := author.ToAuthorInfo()
	pageDetailData := l.buildPageDetailData(page, authorInfo)

	// 5. 构建响应
	return &types.PageDetailResponse{
		Code:      200,
		Message:   "获取页面详情成功",
		Data:      *pageDetailData,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// buildPageDetailData 构建页面详情数据
func (l *GetPageDetailLogic) buildPageDetailData(page *model.Page, author *model.AuthorInfo) *types.PageDetailData {
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
	if page.PublishedAt != nil {
		publishedAt = page.PublishedAt.Format(time.RFC3339)
	}

	return &types.PageDetailData{
		ID:              page.ID.Hex(),
		Title:           page.Title,
		Slug:            page.Slug,
		Content:         page.Content,
		HTML:            page.HTML,
		Author:          authorInfo,
		Status:          page.Status,
		Template:        page.Template,
		MetaTitle:       page.MetaTitle,
		MetaDescription: page.MetaDescription,
		FeaturedImage:   page.FeaturedImage,
		CanonicalURL:    page.CanonicalURL,
		PublishedAt:     publishedAt,
		CreatedAt:       page.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       page.UpdatedAt.Format(time.RFC3339),
	}
}
