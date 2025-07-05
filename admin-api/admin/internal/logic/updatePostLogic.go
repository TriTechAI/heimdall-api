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

type UpdatePostLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新文章
func NewUpdatePostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePostLogic {
	return &UpdatePostLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdatePostLogic) UpdatePost(req *types.PostUpdateRequest) (resp *types.PostUpdateResponse, err error) {
	// 1. 验证文章ID格式
	if !primitive.IsValidObjectID(req.ID) {
		return nil, fmt.Errorf("无效的文章ID格式")
	}

	// 2. 获取当前用户ID
	userID, err := l.getCurrentUserID()
	if err != nil {
		return nil, err
	}

	// 3. 获取现有文章
	existingPost, err := l.svcCtx.PostDAO.GetByID(l.ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("文章不存在: %w", err)
	}

	// 4. 检查权限
	if err := l.checkPermission(userID, existingPost); err != nil {
		return nil, err
	}

	// 5. 验证并处理slug
	if err := l.validateSlug(req, existingPost); err != nil {
		return nil, err
	}

	// 6. 构建更新数据
	updates, err := l.buildUpdateData(req, existingPost)
	if err != nil {
		return nil, err
	}

	// 7. 执行更新
	if err := l.svcCtx.PostDAO.Update(l.ctx, req.ID, updates); err != nil {
		return nil, fmt.Errorf("更新文章失败: %w", err)
	}

	// 8. 获取更新后的文章并构建响应
	return l.buildUpdateResponse(req.ID)
}

// getCurrentUserID 获取当前用户ID
func (l *UpdatePostLogic) getCurrentUserID() (string, error) {
	userID, ok := l.ctx.Value("uid").(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("用户未认证")
	}
	return userID, nil
}

// checkPermission 检查用户权限
func (l *UpdatePostLogic) checkPermission(userID string, post *model.Post) error {
	if post.AuthorID.Hex() != userID {
		return fmt.Errorf("无权限修改此文章")
	}
	return nil
}

// validateSlug 验证并处理slug
func (l *UpdatePostLogic) validateSlug(req *types.PostUpdateRequest, existingPost *model.Post) error {
	// 如果没有提供slug，跳过验证
	if req.Slug == "" {
		return nil
	}

	// 如果slug没有变化，跳过验证
	if req.Slug == existingPost.Slug {
		return nil
	}

	// 检查新slug是否已被其他文章使用
	existingSlugPost, err := l.svcCtx.PostDAO.GetBySlug(l.ctx, req.Slug)
	if err == nil && existingSlugPost != nil && existingSlugPost.ID != existingPost.ID {
		return fmt.Errorf("slug已被使用: %s", req.Slug)
	}

	return nil
}

// buildUpdateData 构建更新数据
func (l *UpdatePostLogic) buildUpdateData(req *types.PostUpdateRequest, existingPost *model.Post) (map[string]interface{}, error) {
	updates := make(map[string]interface{})

	// 只更新非空字段
	if req.Title != "" {
		updates["title"] = req.Title
	}

	if req.Slug != "" {
		updates["slug"] = req.Slug
	}

	if req.Excerpt != "" {
		updates["excerpt"] = req.Excerpt
	}

	if req.Markdown != "" {
		// 处理Markdown内容
		updates["markdown"] = req.Markdown
		updates["html"] = l.convertMarkdownToHTML(req.Markdown)

		// 重新计算内容指标
		wordCount := l.calculateWordCount(req.Markdown)
		readingTime := l.calculateReadingTime(wordCount)
		updates["wordCount"] = wordCount
		updates["readingTime"] = readingTime
	}

	if req.FeaturedImage != "" {
		updates["featuredImage"] = req.FeaturedImage
	}

	if req.Type != "" {
		if err := l.validateType(req.Type); err != nil {
			return nil, err
		}
		updates["type"] = req.Type
	}

	if req.Status != "" {
		if err := l.validateStatus(req.Status); err != nil {
			return nil, err
		}
		updates["status"] = req.Status
	}

	if req.Visibility != "" {
		if err := l.validateVisibility(req.Visibility); err != nil {
			return nil, err
		}
		updates["visibility"] = req.Visibility
	}

	if req.Tags != nil {
		tags := make([]model.Tag, len(req.Tags))
		for i, tag := range req.Tags {
			tags[i] = model.Tag{
				Name: tag.Name,
				Slug: tag.Slug,
			}
		}
		updates["tags"] = tags
	}

	if req.MetaTitle != "" {
		updates["metaTitle"] = req.MetaTitle
	}

	if req.MetaDescription != "" {
		updates["metaDescription"] = req.MetaDescription
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

// validateType 验证文章类型
func (l *UpdatePostLogic) validateType(postType string) error {
	validTypes := []string{constants.PostTypePost, constants.PostTypePage}
	for _, validType := range validTypes {
		if postType == validType {
			return nil
		}
	}
	return fmt.Errorf("无效的文章类型: %s", postType)
}

// validateStatus 验证文章状态
func (l *UpdatePostLogic) validateStatus(status string) error {
	validStatuses := []string{
		constants.PostStatusDraft,
		constants.PostStatusPublished,
		constants.PostStatusScheduled,
		constants.PostStatusArchived,
	}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return nil
		}
	}
	return fmt.Errorf("无效的文章状态: %s", status)
}

// validateVisibility 验证可见性
func (l *UpdatePostLogic) validateVisibility(visibility string) error {
	validVisibilities := []string{
		constants.PostVisibilityPublic,
		constants.PostVisibilityMembersOnly,
		constants.PostVisibilityPrivate,
	}
	for _, validVisibility := range validVisibilities {
		if visibility == validVisibility {
			return nil
		}
	}
	return fmt.Errorf("无效的可见性设置: %s", visibility)
}

// convertMarkdownToHTML 简单的Markdown转HTML
func (l *UpdatePostLogic) convertMarkdownToHTML(markdown string) string {
	html := strings.ReplaceAll(markdown, "\n", "<br/>")

	// 简单的标题处理
	lines := strings.Split(html, "<br/>")
	for i, line := range lines {
		if strings.HasPrefix(line, "# ") {
			lines[i] = "<h1>" + strings.TrimPrefix(line, "# ") + "</h1>"
		} else if strings.HasPrefix(line, "## ") {
			lines[i] = "<h2>" + strings.TrimPrefix(line, "## ") + "</h2>"
		} else if strings.HasPrefix(line, "### ") {
			lines[i] = "<h3>" + strings.TrimPrefix(line, "### ") + "</h3>"
		} else if line != "" && !strings.HasPrefix(line, "<h") {
			lines[i] = "<p>" + line + "</p>"
		}
	}

	return strings.Join(lines, "")
}

// calculateWordCount 计算字数
func (l *UpdatePostLogic) calculateWordCount(content string) int {
	words := strings.Fields(content)
	return len(words)
}

// calculateReadingTime 计算阅读时间（分钟）
func (l *UpdatePostLogic) calculateReadingTime(wordCount int) int {
	// 假设每分钟阅读200字
	readingTime := wordCount / 200
	if readingTime < 1 {
		return 1
	}
	return readingTime
}

// buildUpdateResponse 构建更新响应
func (l *UpdatePostLogic) buildUpdateResponse(postID string) (*types.PostUpdateResponse, error) {
	// 获取更新后的文章
	updatedPost, err := l.svcCtx.PostDAO.GetByID(l.ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("获取更新后的文章失败: %w", err)
	}

	// 获取作者信息
	author, err := l.svcCtx.UserDAO.GetByID(l.ctx, updatedPost.AuthorID.Hex())
	if err != nil {
		return nil, fmt.Errorf("获取作者信息失败: %w", err)
	}
	if author == nil {
		return nil, fmt.Errorf("作者信息不存在")
	}

	// 构建响应数据
	postDetailData := l.buildPostDetailData(updatedPost, author)

	return &types.PostUpdateResponse{
		Code:      200,
		Message:   "文章更新成功",
		Data:      postDetailData,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// buildPostDetailData 构建文章详情数据
func (l *UpdatePostLogic) buildPostDetailData(post *model.Post, author *model.User) types.PostDetailData {
	// 转换标签
	tags := make([]types.TagInfo, len(post.Tags))
	for i, tag := range post.Tags {
		tags[i] = types.TagInfo{
			Name: tag.Name,
			Slug: tag.Slug,
		}
	}

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
	if !post.PublishedAt.IsZero() {
		publishedAt = post.PublishedAt.Format(time.RFC3339)
	}

	return types.PostDetailData{
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
