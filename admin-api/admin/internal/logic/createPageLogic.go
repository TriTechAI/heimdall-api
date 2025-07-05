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

type CreatePageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建页面
func NewCreatePageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePageLogic {
	return &CreatePageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreatePageLogic) CreatePage(req *types.PageCreateRequest) (resp *types.PageCreateResponse, err error) {
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

	// 4. 创建页面模型
	page := l.buildPageFromRequest(req, authorID, uniqueSlug)

	// 5. 保存到数据库
	if err := l.svcCtx.PageDAO.Create(l.ctx, page); err != nil {
		return nil, fmt.Errorf("页面创建失败: %v", err)
	}

	// 6. 获取创建后的页面（包含生成的ID）
	createdPage, err := l.svcCtx.PageDAO.GetByID(l.ctx, page.ID.Hex())
	if err != nil {
		return nil, fmt.Errorf("获取创建的页面失败: %v", err)
	}

	// 7. 构建响应
	authorInfo := author.ToAuthorInfo()
	pageDetailData := l.buildPageDetailData(createdPage, authorInfo)

	return &types.PageCreateResponse{
		Code:      200,
		Message:   "页面创建成功",
		Data:      *pageDetailData,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// validateRequest 验证请求参数
func (l *CreatePageLogic) validateRequest(req *types.PageCreateRequest) error {
	if req.Title == "" {
		return fmt.Errorf("标题不能为空")
	}
	if req.Content == "" {
		return fmt.Errorf("内容不能为空")
	}
	if req.Status == "" {
		return fmt.Errorf("页面状态不能为空")
	}
	if !constants.IsValidPostStatus(req.Status) {
		return fmt.Errorf("无效的页面状态")
	}

	return nil
}

// generateSlugFromTitle 从标题生成slug
func (l *CreatePageLogic) generateSlugFromTitle(title string) string {
	return model.GenerateSlugFromText(title)
}

// generateUniqueSlug 生成唯一的slug
func (l *CreatePageLogic) generateUniqueSlug(baseSlug string) (string, error) {
	// 检查原始slug是否可用
	_, err := l.svcCtx.PageDAO.GetBySlug(l.ctx, baseSlug)
	if err != nil {
		// 如果查询出错（通常是不存在），说明slug可用
		return baseSlug, nil
	}

	// 如果slug已存在，尝试添加数字后缀
	for i := 1; i <= 100; i++ {
		newSlug := fmt.Sprintf("%s-%d", baseSlug, i)
		_, err := l.svcCtx.PageDAO.GetBySlug(l.ctx, newSlug)
		if err != nil {
			// 找到可用的slug
			return newSlug, nil
		}
	}

	return "", fmt.Errorf("无法生成唯一的slug")
}

// buildPageFromRequest 从请求构建页面模型
func (l *CreatePageLogic) buildPageFromRequest(req *types.PageCreateRequest, authorID primitive.ObjectID, slug string) *model.Page {
	now := time.Now()

	// 处理发布时间
	var publishedAt *time.Time
	if req.PublishedAt != "" {
		if parsedTime, err := time.Parse(time.RFC3339, req.PublishedAt); err == nil {
			publishedAt = &parsedTime
		}
	}

	// 设置默认模板
	template := req.Template
	if template == "" {
		template = "default"
	}

	page := &model.Page{
		ID:              primitive.NewObjectID(),
		Title:           req.Title,
		Slug:            slug,
		Content:         req.Content,
		HTML:            l.convertContentToHTML(req.Content),
		AuthorID:        authorID,
		Status:          req.Status,
		Template:        template,
		MetaTitle:       req.MetaTitle,
		MetaDescription: req.MetaDescription,
		FeaturedImage:   req.FeaturedImage,
		CanonicalURL:    req.CanonicalURL,
		PublishedAt:     publishedAt,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	return page
}

// convertContentToHTML 将内容转换为HTML
func (l *CreatePageLogic) convertContentToHTML(content string) string {
	// 简单的内容到HTML转换
	// 在实际项目中，应该使用专业的Markdown库或富文本编辑器
	html := strings.ReplaceAll(content, "\n", "<br/>")

	// 处理标题
	html = strings.ReplaceAll(html, "# ", "<h1>")
	html = strings.ReplaceAll(html, "## ", "<h2>")
	html = strings.ReplaceAll(html, "### ", "<h3>")

	// 简单的段落处理
	if !strings.Contains(html, "<h") && !strings.Contains(html, "<p") {
		html = "<p>" + html + "</p>"
	}

	return html
}

// buildPageDetailData 构建页面详情数据
func (l *CreatePageLogic) buildPageDetailData(page *model.Page, author *model.AuthorInfo) *types.PageDetailData {
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
