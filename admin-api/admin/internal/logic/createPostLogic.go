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
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreatePostLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建文章
func NewCreatePostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePostLogic {
	return &CreatePostLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreatePostLogic) CreatePost(req *types.PostCreateRequest) (resp *types.PostCreateResponse, err error) {
	// 1. 参数验证
	if err := l.validateRequest(req); err != nil {
		return nil, err
	}

	// 2. 获取当前用户信息
	userID := l.ctx.Value("userId")
	if userID == nil {
		return nil, fmt.Errorf("用户未登录")
	}

	authorID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		return nil, fmt.Errorf("无效的用户ID")
	}

	// 验证用户是否存在
	author, err := l.svcCtx.UserDAO.GetByID(l.ctx, authorID.Hex())
	if err != nil {
		return nil, fmt.Errorf("作者不存在: %v", err)
	}

	// 3. 处理slug
	slug := req.Slug
	if slug == "" {
		slug = l.generateSlugFromTitle(req.Title)
	}

	// 检查slug重复并生成唯一slug
	uniqueSlug, err := l.generateUniqueSlug(slug)
	if err != nil {
		return nil, fmt.Errorf("slug生成失败: %v", err)
	}

	// 4. 创建文章模型
	post := l.buildPostFromRequest(req, authorID, uniqueSlug)

	// 5. 保存到数据库
	if err := l.svcCtx.PostDAO.Create(l.ctx, post); err != nil {
		return nil, fmt.Errorf("文章创建失败: %v", err)
	}

	// 6. 获取创建后的文章（包含生成的ID）
	createdPost, err := l.svcCtx.PostDAO.GetByID(l.ctx, post.ID.Hex())
	if err != nil {
		return nil, fmt.Errorf("获取创建的文章失败: %v", err)
	}

	// 7. 构建响应
	authorInfo := author.ToAuthorInfo()
	postDetailData := l.buildPostDetailData(createdPost, authorInfo)

	return &types.PostCreateResponse{
		Code:      200,
		Message:   "文章创建成功",
		Data:      *postDetailData,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// validateRequest 验证请求参数
func (l *CreatePostLogic) validateRequest(req *types.PostCreateRequest) error {
	if req.Title == "" {
		return fmt.Errorf("标题不能为空")
	}
	if req.Markdown == "" {
		return fmt.Errorf("内容不能为空")
	}
	if req.Type == "" {
		return fmt.Errorf("文章类型不能为空")
	}
	if !constants.IsValidPostType(req.Type) {
		return fmt.Errorf("无效的文章类型")
	}
	if req.Status == "" {
		return fmt.Errorf("文章状态不能为空")
	}
	if !constants.IsValidPostStatus(req.Status) {
		return fmt.Errorf("无效的文章状态")
	}
	if req.Visibility == "" {
		return fmt.Errorf("文章可见性不能为空")
	}
	if !constants.IsValidPostVisibility(req.Visibility) {
		return fmt.Errorf("无效的文章可见性")
	}

	return nil
}

// generateSlugFromTitle 从标题生成slug
func (l *CreatePostLogic) generateSlugFromTitle(title string) string {
	return model.GenerateSlugFromText(title)
}

// generateUniqueSlug 生成唯一的slug
func (l *CreatePostLogic) generateUniqueSlug(baseSlug string) (string, error) {
	// 检查原始slug是否可用
	_, err := l.svcCtx.PostDAO.GetBySlug(l.ctx, baseSlug)
	if err != nil {
		// 如果查询出错（通常是不存在），说明slug可用
		return baseSlug, nil
	}

	// 如果slug已存在，尝试添加数字后缀
	for i := 1; i <= 100; i++ {
		newSlug := fmt.Sprintf("%s-%d", baseSlug, i)
		_, err := l.svcCtx.PostDAO.GetBySlug(l.ctx, newSlug)
		if err != nil {
			// 找到可用的slug
			return newSlug, nil
		}
	}

	return "", fmt.Errorf("无法生成唯一的slug")
}

// buildPostFromRequest 从请求构建文章模型
func (l *CreatePostLogic) buildPostFromRequest(req *types.PostCreateRequest, authorID primitive.ObjectID, slug string) *model.Post {
	now := time.Now()

	// 转换标签
	tags := make([]model.Tag, len(req.Tags))
	for i, tag := range req.Tags {
		tagSlug := tag.Slug
		if tagSlug == "" {
			tagSlug = model.GenerateSlugFromText(tag.Name)
		}
		tags[i] = model.Tag{
			Name: tag.Name,
			Slug: tagSlug,
		}
	}

	// 处理发布时间
	var publishedAt *time.Time
	if req.PublishedAt != "" {
		if parsedTime, err := time.Parse(time.RFC3339, req.PublishedAt); err == nil {
			publishedAt = &parsedTime
		}
	}

	post := &model.Post{
		ID:              primitive.NewObjectID(),
		Title:           req.Title,
		Slug:            slug,
		Excerpt:         req.Excerpt,
		Markdown:        req.Markdown,
		HTML:            l.convertMarkdownToHTML(req.Markdown),
		FeaturedImage:   req.FeaturedImage,
		Type:            req.Type,
		Status:          req.Status,
		Visibility:      req.Visibility,
		AuthorID:        authorID,
		Tags:            tags,
		MetaTitle:       req.MetaTitle,
		MetaDescription: req.MetaDescription,
		CanonicalURL:    req.CanonicalURL,
		ViewCount:       0,
		PublishedAt:     publishedAt,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// 自动生成摘要（如果没有提供）
	if post.Excerpt == "" {
		post.Excerpt = post.GenerateExcerpt(200)
	}

	// 计算内容指标
	post.UpdateContentMetrics()

	return post
}

// convertMarkdownToHTML 将Markdown转换为HTML（简单实现）
func (l *CreatePostLogic) convertMarkdownToHTML(markdown string) string {
	// 这里实现简单的Markdown到HTML转换
	// 在实际项目中，应该使用专业的Markdown库如goldmark
	html := strings.ReplaceAll(markdown, "\n", "<br/>")

	// 处理标题
	html = strings.ReplaceAll(html, "# ", "<h1>")
	html = strings.ReplaceAll(html, "## ", "<h2>")
	html = strings.ReplaceAll(html, "### ", "<h3>")

	// 简单的段落处理
	if !strings.Contains(html, "<h") {
		html = "<p>" + html + "</p>"
	}

	return html
}

// buildPostDetailData 构建文章详情数据
func (l *CreatePostLogic) buildPostDetailData(post *model.Post, author *model.AuthorInfo) *types.PostDetailData {
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

	// 格式化时间
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
