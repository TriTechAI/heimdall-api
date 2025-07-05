package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/model"
	"github.com/heimdall-api/public-api/public/internal/svc"
	"github.com/heimdall-api/public-api/public/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPublicPageDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 根据slug获取公开页面详情
func NewGetPublicPageDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPublicPageDetailLogic {
	return &GetPublicPageDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPublicPageDetailLogic) GetPublicPageDetail(req *types.PublicPageDetailRequest) (resp *types.PublicPageDetailResponse, err error) {
	// 1. 验证请求参数
	if err := l.validateRequest(req); err != nil {
		return nil, err
	}

	// 2. 获取页面
	page, err := l.getPageBySlug(req.Slug)
	if err != nil {
		return nil, fmt.Errorf("获取页面失败: %w", err)
	}

	// 3. 验证页面可见性
	if err := l.validatePageVisibility(page); err != nil {
		return nil, err
	}

	// 4. 获取作者信息
	author, err := l.getAuthorInfo(page.AuthorID.Hex())
	if err != nil {
		return nil, fmt.Errorf("获取作者信息失败: %w", err)
	}

	// 5. 构建响应数据
	pageDetail := l.buildPageDetail(page, author)

	// 6. 构建响应
	return l.buildResponse(pageDetail), nil
}

// validateRequest 验证请求参数
func (l *GetPublicPageDetailLogic) validateRequest(req *types.PublicPageDetailRequest) error {
	if strings.TrimSpace(req.Slug) == "" {
		return fmt.Errorf("slug不能为空")
	}
	return nil
}

// getPageBySlug 根据slug获取页面
func (l *GetPublicPageDetailLogic) getPageBySlug(slug string) (*model.Page, error) {
	page, err := l.svcCtx.PageDAO.GetBySlug(l.ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("页面不存在: %w", err)
	}
	return page, nil
}

// validatePageVisibility 验证页面可见性
func (l *GetPublicPageDetailLogic) validatePageVisibility(page *model.Page) error {
	// 检查页面状态
	if page.Status != constants.PostStatusPublished {
		return fmt.Errorf("页面未发布")
	}

	return nil
}

// getAuthorInfo 获取作者信息
func (l *GetPublicPageDetailLogic) getAuthorInfo(authorID string) (*model.User, error) {
	return l.svcCtx.UserDAO.GetByID(l.ctx, authorID)
}

// buildPageDetail 构建页面详情
func (l *GetPublicPageDetailLogic) buildPageDetail(page *model.Page, author *model.User) types.PublicPageDetailData {
	// 构建作者信息
	authorInfo := l.buildAuthorInfo(author)

	// 格式化时间
	publishedAt := ""
	if page.PublishedAt != nil {
		publishedAt = page.PublishedAt.Format(time.RFC3339)
	}

	// 构建canonical URL
	canonicalURL := l.buildCanonicalURL(page.Slug)

	return types.PublicPageDetailData{
		Title:           page.Title,
		Slug:            page.Slug,
		HTML:            page.HTML,
		Template:        page.Template,
		Author:          authorInfo,
		MetaTitle:       page.MetaTitle,
		MetaDescription: page.MetaDescription,
		FeaturedImage:   page.FeaturedImage,
		CanonicalURL:    canonicalURL,
		PublishedAt:     publishedAt,
		UpdatedAt:       page.UpdatedAt.Format(time.RFC3339),
	}
}

// buildAuthorInfo 构建作者信息
func (l *GetPublicPageDetailLogic) buildAuthorInfo(author *model.User) types.PublicAuthorInfo {
	return types.PublicAuthorInfo{
		Username:     author.Username,
		DisplayName:  author.DisplayName,
		ProfileImage: author.ProfileImage,
		Bio:          author.Bio,
	}
}

// buildCanonicalURL 构建canonical URL
func (l *GetPublicPageDetailLogic) buildCanonicalURL(slug string) string {
	// 这里可以从配置中读取域名，暂时使用相对路径
	return fmt.Sprintf("/pages/%s", slug)
}

// buildResponse 构建响应
func (l *GetPublicPageDetailLogic) buildResponse(pageDetail types.PublicPageDetailData) *types.PublicPageDetailResponse {
	return &types.PublicPageDetailResponse{
		Code:      200,
		Message:   "success",
		Data:      pageDetail,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}
