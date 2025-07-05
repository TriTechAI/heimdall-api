package model

import (
	"fmt"
	"time"

	"github.com/heimdall-api/common/constants"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Page 页面模型
type Page struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title           string             `bson:"title" json:"title"`
	Slug            string             `bson:"slug" json:"slug"`
	Content         string             `bson:"content" json:"content"`
	HTML            string             `bson:"html" json:"html"`
	AuthorID        primitive.ObjectID `bson:"authorId" json:"authorId"`
	Status          string             `bson:"status" json:"status"`
	Template        string             `bson:"template" json:"template"`
	MetaTitle       string             `bson:"metaTitle" json:"metaTitle"`
	MetaDescription string             `bson:"metaDescription" json:"metaDescription"`
	FeaturedImage   string             `bson:"featuredImage" json:"featuredImage"`
	CanonicalURL    string             `bson:"canonicalUrl" json:"canonicalUrl"`
	PublishedAt     *time.Time         `bson:"publishedAt,omitempty" json:"publishedAt,omitempty"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// PageCreateRequest 页面创建请求
type PageCreateRequest struct {
	Title           string     `json:"title" validate:"required,min=1,max=255"`
	Slug            string     `json:"slug" validate:"omitempty,min=1,max=255"`
	Content         string     `json:"content" validate:"required"`
	Template        string     `json:"template" validate:"omitempty,max=100"`
	Status          string     `json:"status" validate:"required,oneof=draft published scheduled"`
	MetaTitle       string     `json:"metaTitle" validate:"max=70"`
	MetaDescription string     `json:"metaDescription" validate:"max=160"`
	FeaturedImage   string     `json:"featuredImage" validate:"omitempty,url"`
	CanonicalURL    string     `json:"canonicalUrl" validate:"omitempty,url,max=255"`
	PublishedAt     *time.Time `json:"publishedAt,omitempty"`
}

// PageUpdateRequest 页面更新请求
type PageUpdateRequest struct {
	Title           string     `json:"title" validate:"omitempty,min=1,max=255"`
	Slug            string     `json:"slug" validate:"omitempty,min=1,max=255"`
	Content         string     `json:"content" validate:"omitempty"`
	Template        string     `json:"template" validate:"omitempty,max=100"`
	Status          string     `json:"status" validate:"omitempty,oneof=draft published scheduled"`
	MetaTitle       string     `json:"metaTitle" validate:"max=70"`
	MetaDescription string     `json:"metaDescription" validate:"max=160"`
	FeaturedImage   string     `json:"featuredImage" validate:"omitempty,url"`
	CanonicalURL    string     `json:"canonicalUrl" validate:"omitempty,url,max=255"`
	PublishedAt     *time.Time `json:"publishedAt,omitempty"`
}

// PageFilter 页面过滤器
type PageFilter struct {
	Status   string `json:"status"`
	Template string `json:"template"`
	AuthorID string `json:"authorId"`
	Keyword  string `json:"keyword"` // 搜索关键词
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
	SortBy   string `json:"sortBy"`   // created_at, updated_at, published_at, title
	SortDesc bool   `json:"sortDesc"` // 是否降序
}

// PageDetailResponse 页面详情响应
type PageDetailResponse struct {
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	Slug            string     `json:"slug"`
	Content         string     `json:"content"`
	HTML            string     `json:"html"`
	Author          AuthorInfo `json:"author"`
	Status          string     `json:"status"`
	Template        string     `json:"template"`
	MetaTitle       string     `json:"metaTitle"`
	MetaDescription string     `json:"metaDescription"`
	FeaturedImage   string     `json:"featuredImage"`
	CanonicalURL    string     `json:"canonicalUrl"`
	PublishedAt     *time.Time `json:"publishedAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// PageListItem 页面列表项
type PageListItem struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	Slug          string     `json:"slug"`
	Author        AuthorInfo `json:"author"`
	Status        string     `json:"status"`
	Template      string     `json:"template"`
	FeaturedImage string     `json:"featuredImage"`
	PublishedAt   *time.Time `json:"publishedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// ===============================
// 验证方法
// ===============================

// ValidateForCreate 验证页面创建数据
func (p *Page) ValidateForCreate() error {
	// 验证必填字段
	if p.Title == "" {
		return NewPageValidationError("title", "页面标题不能为空")
	}
	if p.Content == "" {
		return NewPageValidationError("content", "页面内容不能为空")
	}
	if p.Status == "" {
		return NewPageValidationError("status", "页面状态不能为空")
	}
	if p.AuthorID.IsZero() {
		return NewPageValidationError("authorId", "作者ID不能为空")
	}

	// 验证字段长度
	if len(p.Title) > constants.PostTitleMaxLength {
		return NewPageValidationError("title", "页面标题长度不能超过255字符")
	}
	if len(p.Slug) > constants.PostSlugMaxLength {
		return NewPageValidationError("slug", "页面slug长度不能超过255字符")
	}
	if len(p.Content) > constants.PostContentMaxLength {
		return NewPageValidationError("content", "页面内容长度不能超过1MB")
	}
	if len(p.Template) > 100 {
		return NewPageValidationError("template", "模板名称长度不能超过100字符")
	}
	if len(p.MetaTitle) > constants.PostMetaTitleMaxLength {
		return NewPageValidationError("metaTitle", "SEO标题长度不能超过70字符")
	}
	if len(p.MetaDescription) > constants.PostMetaDescMaxLength {
		return NewPageValidationError("metaDescription", "SEO描述长度不能超过160字符")
	}
	if len(p.CanonicalURL) > constants.PostCanonicalUrlMaxLength {
		return NewPageValidationError("canonicalUrl", "规范化URL长度不能超过255字符")
	}

	// 验证枚举值
	if !constants.IsValidPostStatus(p.Status) {
		return NewPageValidationError("status", "无效的页面状态")
	}

	// 验证slug格式
	if p.Slug != "" && !IsValidSlug(p.Slug) {
		return NewPageValidationError("slug", "slug格式无效，只能包含小写字母、数字和连字符")
	}

	return nil
}

// ValidateForUpdate 验证页面更新数据
func (p *Page) ValidateForUpdate() error {
	// 对于更新，只验证非空字段
	if p.Title != "" && len(p.Title) > constants.PostTitleMaxLength {
		return NewPageValidationError("title", "页面标题长度不能超过255字符")
	}
	if p.Slug != "" {
		if len(p.Slug) > constants.PostSlugMaxLength {
			return NewPageValidationError("slug", "页面slug长度不能超过255字符")
		}
		if !IsValidSlug(p.Slug) {
			return NewPageValidationError("slug", "slug格式无效，只能包含小写字母、数字和连字符")
		}
	}
	if p.Content != "" && len(p.Content) > constants.PostContentMaxLength {
		return NewPageValidationError("content", "页面内容长度不能超过1MB")
	}
	if len(p.Template) > 100 {
		return NewPageValidationError("template", "模板名称长度不能超过100字符")
	}
	if len(p.MetaTitle) > constants.PostMetaTitleMaxLength {
		return NewPageValidationError("metaTitle", "SEO标题长度不能超过70字符")
	}
	if len(p.MetaDescription) > constants.PostMetaDescMaxLength {
		return NewPageValidationError("metaDescription", "SEO描述长度不能超过160字符")
	}
	if len(p.CanonicalURL) > constants.PostCanonicalUrlMaxLength {
		return NewPageValidationError("canonicalUrl", "规范化URL长度不能超过255字符")
	}

	// 验证枚举值
	if p.Status != "" && !constants.IsValidPostStatus(p.Status) {
		return NewPageValidationError("status", "无效的页面状态")
	}

	return nil
}

// ===============================
// 状态检查方法
// ===============================

// IsPublished 检查页面是否已发布
func (p *Page) IsPublished() bool {
	return p.Status == constants.PostStatusPublished
}

// IsDraft 检查页面是否为草稿
func (p *Page) IsDraft() bool {
	return p.Status == constants.PostStatusDraft
}

// IsScheduled 检查页面是否为定时发布
func (p *Page) IsScheduled() bool {
	return p.Status == constants.PostStatusScheduled
}

// CanBePublished 检查页面是否可以发布
func (p *Page) CanBePublished() bool {
	return p.Status == constants.PostStatusDraft || p.Status == constants.PostStatusScheduled
}

// ShouldBePublishedNow 检查定时发布的页面是否应该现在发布
func (p *Page) ShouldBePublishedNow() bool {
	return p.IsScheduled() && p.PublishedAt != nil && p.PublishedAt.Before(time.Now())
}

// ===============================
// Slug处理方法
// ===============================

// GenerateSlug 自动生成slug
func (p *Page) GenerateSlug() string {
	if p.Title == "" {
		return ""
	}
	return GenerateSlugFromText(p.Title)
}

// EnsureSlug 确保页面有有效的slug
func (p *Page) EnsureSlug() {
	if p.Slug == "" {
		p.Slug = p.GenerateSlug()
	}
}

// ===============================
// 转换方法
// ===============================

// ToDetailResponse 转换为页面详情响应
func (p *Page) ToDetailResponse(author *AuthorInfo) *PageDetailResponse {
	return &PageDetailResponse{
		ID:              p.ID.Hex(),
		Title:           p.Title,
		Slug:            p.Slug,
		Content:         p.Content,
		HTML:            p.HTML,
		Author:          *author,
		Status:          p.Status,
		Template:        p.Template,
		MetaTitle:       p.MetaTitle,
		MetaDescription: p.MetaDescription,
		FeaturedImage:   p.FeaturedImage,
		CanonicalURL:    p.CanonicalURL,
		PublishedAt:     p.PublishedAt,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

// ToListItem 转换为页面列表项
func (p *Page) ToListItem(author *AuthorInfo) *PageListItem {
	return &PageListItem{
		ID:            p.ID.Hex(),
		Title:         p.Title,
		Slug:          p.Slug,
		Author:        *author,
		Status:        p.Status,
		Template:      p.Template,
		FeaturedImage: p.FeaturedImage,
		PublishedAt:   p.PublishedAt,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

// ===============================
// 工厂方法
// ===============================

// NewPage 创建新页面
func NewPage(title, content, status string, authorID primitive.ObjectID) *Page {
	now := time.Now()

	page := &Page{
		ID:        primitive.NewObjectID(),
		Title:     title,
		Content:   content,
		Status:    status,
		AuthorID:  authorID,
		Template:  "default", // 默认模板
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 自动生成slug
	page.EnsureSlug()

	return page
}

// NewPageFromCreateRequest 从创建请求创建页面
func NewPageFromCreateRequest(req *PageCreateRequest, authorID primitive.ObjectID) *Page {
	now := time.Now()

	page := &Page{
		ID:              primitive.NewObjectID(),
		Title:           req.Title,
		Slug:            req.Slug,
		Content:         req.Content,
		Status:          req.Status,
		Template:        req.Template,
		AuthorID:        authorID,
		MetaTitle:       req.MetaTitle,
		MetaDescription: req.MetaDescription,
		FeaturedImage:   req.FeaturedImage,
		CanonicalURL:    req.CanonicalURL,
		PublishedAt:     req.PublishedAt,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// 设置默认模板
	if page.Template == "" {
		page.Template = "default"
	}

	// 自动生成slug
	page.EnsureSlug()

	return page
}

// ===============================
// 准备方法
// ===============================

// PrepareForInsert 准备插入数据库
func (p *Page) PrepareForInsert() {
	now := time.Now()
	if p.ID.IsZero() {
		p.ID = primitive.NewObjectID()
	}
	p.CreatedAt = now
	p.UpdatedAt = now

	// 确保有默认模板
	if p.Template == "" {
		p.Template = "default"
	}

	// 确保有slug
	p.EnsureSlug()
}

// PrepareForUpdate 准备更新数据库
func (p *Page) PrepareForUpdate() {
	p.UpdatedAt = time.Now()
}

// ===============================
// 发布管理方法
// ===============================

// Publish 发布页面
func (p *Page) Publish() {
	p.Status = constants.PostStatusPublished
	if p.PublishedAt == nil {
		now := time.Now()
		p.PublishedAt = &now
	}
	p.UpdatedAt = time.Now()
}

// Unpublish 取消发布页面
func (p *Page) Unpublish() {
	p.Status = constants.PostStatusDraft
	p.UpdatedAt = time.Now()
}

// ===============================
// 错误类型
// ===============================

// PageValidationError 页面验证错误
type PageValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error 实现error接口
func (e *PageValidationError) Error() string {
	return fmt.Sprintf("页面验证错误 - %s: %s", e.Field, e.Message)
}

// NewPageValidationError 创建页面验证错误
func NewPageValidationError(field, message string) *PageValidationError {
	return &PageValidationError{
		Field:   field,
		Message: message,
	}
}
