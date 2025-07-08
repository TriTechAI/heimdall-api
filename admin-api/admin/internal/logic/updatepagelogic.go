package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/model"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UpdatePageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新页面
func NewUpdatePageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePageLogic {
	return &UpdatePageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdatePageLogic) UpdatePage(req *types.PageUpdateRequest) (resp *types.PageUpdateResponse, err error) {
	// 1. 验证页面ID格式
	if !primitive.IsValidObjectID(req.ID) {
		return nil, fmt.Errorf("无效的页面ID格式")
	}

	// 2. 获取当前用户ID
	userID, err := l.getCurrentUserID()
	if err != nil {
		return nil, err
	}

	// 3. 获取现有页面
	existingPage, err := l.svcCtx.PageDAO.GetByID(l.ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("页面不存在: %w", err)
	}

	// 4. 检查权限
	if err := l.checkPermission(userID, existingPage); err != nil {
		return nil, err
	}

	// 5. 验证并处理slug
	if err := l.validateSlug(req, existingPage); err != nil {
		return nil, err
	}

	// 6. 构建更新数据
	updates, err := l.buildUpdateData(req, existingPage)
	if err != nil {
		return nil, err
	}

	// 7. 执行更新
	if err := l.svcCtx.PageDAO.Update(l.ctx, req.ID, updates); err != nil {
		return nil, fmt.Errorf("更新页面失败: %w", err)
	}

	// 8. 获取更新后的页面并构建响应
	return l.buildUpdateResponse(req.ID)
}

// getCurrentUserID 获取当前用户ID
func (l *UpdatePageLogic) getCurrentUserID() (string, error) {
	userID, ok := l.ctx.Value("userId").(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("用户未认证")
	}
	return userID, nil
}

// checkPermission 检查用户权限
func (l *UpdatePageLogic) checkPermission(userID string, page *model.Page) error {
	if page.AuthorID.Hex() != userID {
		return fmt.Errorf("无权限修改此页面")
	}
	return nil
}

// validateSlug 验证并处理slug
func (l *UpdatePageLogic) validateSlug(req *types.PageUpdateRequest, existingPage *model.Page) error {
	// 如果没有提供slug，跳过验证
	if req.Slug == "" {
		return nil
	}

	// 如果slug没有变化，跳过验证
	if req.Slug == existingPage.Slug {
		return nil
	}

	// 检查新slug是否已被其他页面使用
	existingSlugPage, err := l.svcCtx.PageDAO.GetBySlug(l.ctx, req.Slug)
	if err == nil && existingSlugPage != nil && existingSlugPage.ID != existingPage.ID {
		return fmt.Errorf("slug已被使用: %s", req.Slug)
	}

	return nil
}

// buildUpdateData 构建更新数据
func (l *UpdatePageLogic) buildUpdateData(req *types.PageUpdateRequest, existingPage *model.Page) (map[string]interface{}, error) {
	updates := make(map[string]interface{})

	// 只更新非空字段
	if req.Title != "" {
		updates["title"] = req.Title
	}

	if req.Slug != "" {
		updates["slug"] = req.Slug
	}

	if req.Content != "" {
		// 处理内容
		updates["content"] = req.Content
		updates["html"] = l.convertContentToHTML(req.Content)
	}

	if req.Template != "" {
		updates["template"] = req.Template
	}

	if req.Status != "" {
		if err := l.validateStatus(req.Status); err != nil {
			return nil, err
		}
		updates["status"] = req.Status
	}

	if req.MetaTitle != "" {
		updates["metaTitle"] = req.MetaTitle
	}

	if req.MetaDescription != "" {
		updates["metaDescription"] = req.MetaDescription
	}

	if req.FeaturedImage != "" {
		updates["featuredImage"] = req.FeaturedImage
	}

	if req.CanonicalURL != "" {
		updates["canonicalUrl"] = req.CanonicalURL
	}

	if req.PublishedAt != "" {
		publishedAt, err := time.Parse(time.RFC3339, req.PublishedAt)
		if err != nil {
			return nil, fmt.Errorf("无效的发布时间格式: %w", err)
		}
		updates["publishedAt"] = publishedAt
	}

	// 更新修改时间
	updates["updatedAt"] = time.Now()

	return updates, nil
}

// validateStatus 验证页面状态
func (l *UpdatePageLogic) validateStatus(status string) error {
	if !constants.IsValidPostStatus(status) {
		return fmt.Errorf("无效的页面状态: %s", status)
	}
	return nil
}

// convertContentToHTML 简单的内容转HTML
func (l *UpdatePageLogic) convertContentToHTML(content string) string {
	html := strings.ReplaceAll(content, "\n", "<br/>")

	// 简单的标题处理
	lines := strings.Split(html, "<br/>")
	for i, line := range lines {
		if strings.HasPrefix(line, "# ") {
			lines[i] = "<h1>" + strings.TrimPrefix(line, "# ") + "</h1>"
		} else if strings.HasPrefix(line, "## ") {
			lines[i] = "<h2>" + strings.TrimPrefix(line, "## ") + "</h2>"
		} else if strings.HasPrefix(line, "### ") {
			lines[i] = "<h3>" + strings.TrimPrefix(line, "### ") + "</h3>"
		} else if line != "" && !strings.HasPrefix(line, "<h") && !strings.HasPrefix(line, "<p") {
			lines[i] = "<p>" + line + "</p>"
		}
	}

	return strings.Join(lines, "")
}

// buildUpdateResponse 构建更新响应
func (l *UpdatePageLogic) buildUpdateResponse(pageID string) (*types.PageUpdateResponse, error) {
	// 获取更新后的页面
	updatedPage, err := l.svcCtx.PageDAO.GetByID(l.ctx, pageID)
	if err != nil {
		return nil, fmt.Errorf("获取更新后的页面失败: %w", err)
	}

	// 获取作者信息
	author, err := l.svcCtx.UserDAO.GetByID(l.ctx, updatedPage.AuthorID.Hex())
	if err != nil {
		return nil, fmt.Errorf("获取作者信息失败: %w", err)
	}
	if author == nil {
		return nil, fmt.Errorf("作者信息不存在")
	}

	// 构建响应数据
	pageDetailData := l.buildPageDetailData(updatedPage, author)

	return &types.PageUpdateResponse{
		Code:      200,
		Message:   "页面更新成功",
		Data:      pageDetailData,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// buildPageDetailData 构建页面详情数据
func (l *UpdatePageLogic) buildPageDetailData(page *model.Page, author *model.User) types.PageDetailData {
	// 构建作者信息
	var authorInfo types.AuthorInfo
	if author != nil {
		authorInfo = types.AuthorInfo{
			ID:           author.ID.Hex(),
			Username:     author.Username,
			DisplayName:  author.DisplayName,
			ProfileImage: author.ProfileImage,
			Bio:          author.Bio,
		}
	}

	// 处理发布时间
	var publishedAt string
	if page.PublishedAt != nil {
		publishedAt = page.PublishedAt.Format(time.RFC3339)
	}

	return types.PageDetailData{
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
