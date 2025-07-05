package model

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/heimdall-api/common/constants"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Post 文章模型
type Post struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title           string             `bson:"title" json:"title"`
	Slug            string             `bson:"slug" json:"slug"`
	Excerpt         string             `bson:"excerpt" json:"excerpt"`
	Markdown        string             `bson:"markdown" json:"markdown"`
	HTML            string             `bson:"html" json:"html"`
	FeaturedImage   string             `bson:"featuredImage" json:"featuredImage"`
	Type            string             `bson:"type" json:"type"`
	Status          string             `bson:"status" json:"status"`
	Visibility      string             `bson:"visibility" json:"visibility"`
	AuthorID        primitive.ObjectID `bson:"authorId" json:"authorId"`
	Tags            []Tag              `bson:"tags" json:"tags"`
	MetaTitle       string             `bson:"metaTitle" json:"metaTitle"`
	MetaDescription string             `bson:"metaDescription" json:"metaDescription"`
	CanonicalURL    string             `bson:"canonicalUrl" json:"canonicalUrl"`
	ReadingTime     int                `bson:"readingTime" json:"readingTime"`
	WordCount       int                `bson:"wordCount" json:"wordCount"`
	ViewCount       int64              `bson:"viewCount" json:"viewCount"`
	PublishedAt     *time.Time         `bson:"publishedAt,omitempty" json:"publishedAt,omitempty"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// Tag 内嵌标签结构
type Tag struct {
	Name string `bson:"name" json:"name"`
	Slug string `bson:"slug" json:"slug"`
}

// PostCreateRequest 文章创建请求
type PostCreateRequest struct {
	Title           string     `json:"title" validate:"required,min=1,max=255"`
	Slug            string     `json:"slug" validate:"omitempty,min=1,max=255"`
	Excerpt         string     `json:"excerpt" validate:"max=500"`
	Markdown        string     `json:"markdown" validate:"required"`
	FeaturedImage   string     `json:"featuredImage" validate:"omitempty,url"`
	Type            string     `json:"type" validate:"required,oneof=post page"`
	Status          string     `json:"status" validate:"required,oneof=draft published scheduled archived"`
	Visibility      string     `json:"visibility" validate:"required,oneof=public members_only private"`
	Tags            []TagInfo  `json:"tags" validate:"max=20"`
	MetaTitle       string     `json:"metaTitle" validate:"max=70"`
	MetaDescription string     `json:"metaDescription" validate:"max=160"`
	CanonicalURL    string     `json:"canonicalUrl" validate:"omitempty,url,max=255"`
	PublishedAt     *time.Time `json:"publishedAt,omitempty"`
}

// PostUpdateRequest 文章更新请求
type PostUpdateRequest struct {
	Title           string     `json:"title" validate:"omitempty,min=1,max=255"`
	Slug            string     `json:"slug" validate:"omitempty,min=1,max=255"`
	Excerpt         string     `json:"excerpt" validate:"max=500"`
	Markdown        string     `json:"markdown" validate:"omitempty"`
	FeaturedImage   string     `json:"featuredImage" validate:"omitempty,url"`
	Type            string     `json:"type" validate:"omitempty,oneof=post page"`
	Status          string     `json:"status" validate:"omitempty,oneof=draft published scheduled archived"`
	Visibility      string     `json:"visibility" validate:"omitempty,oneof=public members_only private"`
	Tags            []TagInfo  `json:"tags" validate:"max=20"`
	MetaTitle       string     `json:"metaTitle" validate:"max=70"`
	MetaDescription string     `json:"metaDescription" validate:"max=160"`
	CanonicalURL    string     `json:"canonicalUrl" validate:"omitempty,url,max=255"`
	PublishedAt     *time.Time `json:"publishedAt,omitempty"`
}

// TagInfo 标签信息
type TagInfo struct {
	Name string `json:"name" validate:"required,max=50"`
	Slug string `json:"slug" validate:"omitempty,max=50"`
}

// PostFilter 文章过滤器
type PostFilter struct {
	Status     string `json:"status"`
	Type       string `json:"type"`
	Visibility string `json:"visibility"`
	AuthorID   string `json:"authorId"`
	Tag        string `json:"tag"`     // 标签slug
	Keyword    string `json:"keyword"` // 搜索关键词
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	SortBy     string `json:"sortBy"`   // created_at, updated_at, published_at, view_count, title
	SortDesc   bool   `json:"sortDesc"` // 是否降序
}

// PostDetailResponse 文章详情响应
type PostDetailResponse struct {
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	Slug            string     `json:"slug"`
	Excerpt         string     `json:"excerpt"`
	Markdown        string     `json:"markdown"`
	HTML            string     `json:"html"`
	FeaturedImage   string     `json:"featuredImage"`
	Type            string     `json:"type"`
	Status          string     `json:"status"`
	Visibility      string     `json:"visibility"`
	Author          AuthorInfo `json:"author"`
	Tags            []Tag      `json:"tags"`
	MetaTitle       string     `json:"metaTitle"`
	MetaDescription string     `json:"metaDescription"`
	CanonicalURL    string     `json:"canonicalUrl"`
	ReadingTime     int        `json:"readingTime"`
	WordCount       int        `json:"wordCount"`
	ViewCount       int64      `json:"viewCount"`
	PublishedAt     *time.Time `json:"publishedAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// PostListItem 文章列表项
type PostListItem struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	Slug          string     `json:"slug"`
	Excerpt       string     `json:"excerpt"`
	FeaturedImage string     `json:"featuredImage"`
	Type          string     `json:"type"`
	Status        string     `json:"status"`
	Visibility    string     `json:"visibility"`
	Author        AuthorInfo `json:"author"`
	Tags          []Tag      `json:"tags"`
	ReadingTime   int        `json:"readingTime"`
	ViewCount     int64      `json:"viewCount"`
	PublishedAt   *time.Time `json:"publishedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// ===============================
// 验证方法
// ===============================

// ValidateForCreate 验证文章创建数据
func (p *Post) ValidateForCreate() error {
	// 验证必填字段
	if p.Title == "" {
		return NewPostValidationError("title", "文章标题不能为空")
	}
	if p.Markdown == "" {
		return NewPostValidationError("markdown", "文章内容不能为空")
	}
	if p.Type == "" {
		return NewPostValidationError("type", "文章类型不能为空")
	}
	if p.Status == "" {
		return NewPostValidationError("status", "文章状态不能为空")
	}
	if p.Visibility == "" {
		return NewPostValidationError("visibility", "文章可见性不能为空")
	}
	if p.AuthorID.IsZero() {
		return NewPostValidationError("authorId", "作者ID不能为空")
	}

	// 验证字段长度
	if len(p.Title) > constants.PostTitleMaxLength {
		return NewPostValidationError("title", "文章标题长度不能超过255字符")
	}
	if len(p.Slug) > constants.PostSlugMaxLength {
		return NewPostValidationError("slug", "文章slug长度不能超过255字符")
	}
	if len(p.Excerpt) > constants.PostExcerptMaxLength {
		return NewPostValidationError("excerpt", "文章摘要长度不能超过500字符")
	}
	if len(p.Markdown) > constants.PostContentMaxLength {
		return NewPostValidationError("markdown", "文章内容长度不能超过1MB")
	}
	if len(p.MetaTitle) > constants.PostMetaTitleMaxLength {
		return NewPostValidationError("metaTitle", "SEO标题长度不能超过70字符")
	}
	if len(p.MetaDescription) > constants.PostMetaDescMaxLength {
		return NewPostValidationError("metaDescription", "SEO描述长度不能超过160字符")
	}
	if len(p.CanonicalURL) > constants.PostCanonicalUrlMaxLength {
		return NewPostValidationError("canonicalUrl", "规范化URL长度不能超过255字符")
	}

	// 验证标签数量
	if len(p.Tags) > constants.PostTagMaxCount {
		return NewPostValidationError("tags", "标签数量不能超过20个")
	}

	// 验证标签内容
	for i, tag := range p.Tags {
		if tag.Name == "" {
			return NewPostValidationError("tags", fmt.Sprintf("第%d个标签名称不能为空", i+1))
		}
		if len(tag.Name) > constants.PostTagNameMaxLength {
			return NewPostValidationError("tags", fmt.Sprintf("第%d个标签名称长度不能超过50字符", i+1))
		}
	}

	// 验证枚举值
	if !constants.IsValidPostType(p.Type) {
		return NewPostValidationError("type", "无效的文章类型")
	}
	if !constants.IsValidPostStatus(p.Status) {
		return NewPostValidationError("status", "无效的文章状态")
	}
	if !constants.IsValidPostVisibility(p.Visibility) {
		return NewPostValidationError("visibility", "无效的文章可见性")
	}

	// 验证slug格式
	if p.Slug != "" && !IsValidSlug(p.Slug) {
		return NewPostValidationError("slug", "slug格式无效，只能包含小写字母、数字和连字符")
	}

	return nil
}

// ValidateForUpdate 验证文章更新数据
func (p *Post) ValidateForUpdate() error {
	// 对于更新，只验证非空字段
	if p.Title != "" && len(p.Title) > constants.PostTitleMaxLength {
		return NewPostValidationError("title", "文章标题长度不能超过255字符")
	}
	if p.Slug != "" {
		if len(p.Slug) > constants.PostSlugMaxLength {
			return NewPostValidationError("slug", "文章slug长度不能超过255字符")
		}
		if !IsValidSlug(p.Slug) {
			return NewPostValidationError("slug", "slug格式无效，只能包含小写字母、数字和连字符")
		}
	}
	if len(p.Excerpt) > constants.PostExcerptMaxLength {
		return NewPostValidationError("excerpt", "文章摘要长度不能超过500字符")
	}
	if p.Markdown != "" && len(p.Markdown) > constants.PostContentMaxLength {
		return NewPostValidationError("markdown", "文章内容长度不能超过1MB")
	}
	if len(p.MetaTitle) > constants.PostMetaTitleMaxLength {
		return NewPostValidationError("metaTitle", "SEO标题长度不能超过70字符")
	}
	if len(p.MetaDescription) > constants.PostMetaDescMaxLength {
		return NewPostValidationError("metaDescription", "SEO描述长度不能超过160字符")
	}
	if len(p.CanonicalURL) > constants.PostCanonicalUrlMaxLength {
		return NewPostValidationError("canonicalUrl", "规范化URL长度不能超过255字符")
	}

	// 验证标签数量
	if len(p.Tags) > constants.PostTagMaxCount {
		return NewPostValidationError("tags", "标签数量不能超过20个")
	}

	// 验证标签内容
	for i, tag := range p.Tags {
		if tag.Name == "" {
			return NewPostValidationError("tags", fmt.Sprintf("第%d个标签名称不能为空", i+1))
		}
		if len(tag.Name) > constants.PostTagNameMaxLength {
			return NewPostValidationError("tags", fmt.Sprintf("第%d个标签名称长度不能超过50字符", i+1))
		}
	}

	// 验证枚举值
	if p.Type != "" && !constants.IsValidPostType(p.Type) {
		return NewPostValidationError("type", "无效的文章类型")
	}
	if p.Status != "" && !constants.IsValidPostStatus(p.Status) {
		return NewPostValidationError("status", "无效的文章状态")
	}
	if p.Visibility != "" && !constants.IsValidPostVisibility(p.Visibility) {
		return NewPostValidationError("visibility", "无效的文章可见性")
	}

	return nil
}

// ===============================
// 状态检查方法
// ===============================

// IsPublished 检查文章是否已发布
func (p *Post) IsPublished() bool {
	return p.Status == constants.PostStatusPublished
}

// IsDraft 检查文章是否为草稿
func (p *Post) IsDraft() bool {
	return p.Status == constants.PostStatusDraft
}

// IsScheduled 检查文章是否为定时发布
func (p *Post) IsScheduled() bool {
	return p.Status == constants.PostStatusScheduled
}

// IsPublic 检查文章是否公开可见
func (p *Post) IsPublic() bool {
	return p.Visibility == constants.PostVisibilityPublic
}

// CanBePublished 检查文章是否可以发布
func (p *Post) CanBePublished() bool {
	return p.Status == constants.PostStatusDraft || p.Status == constants.PostStatusScheduled
}

// ShouldBePublishedNow 检查定时发布的文章是否应该现在发布
func (p *Post) ShouldBePublishedNow() bool {
	return p.IsScheduled() && p.PublishedAt != nil && p.PublishedAt.Before(time.Now())
}

// ===============================
// Slug处理方法
// ===============================

// GenerateSlug 自动生成slug
func (p *Post) GenerateSlug() string {
	if p.Title == "" {
		return ""
	}

	slug := GenerateSlugFromText(p.Title)
	return slug
}

// EnsureSlug 确保文章有有效的slug
func (p *Post) EnsureSlug() {
	if p.Slug == "" {
		p.Slug = p.GenerateSlug()
	}
}

// ===============================
// 内容处理方法
// ===============================

// CalculateWordCount 计算字数
func (p *Post) CalculateWordCount() int {
	if p.Markdown == "" {
		return 0
	}

	// 简单的字数统计（按空格分割）
	text := strings.TrimSpace(p.Markdown)
	if text == "" {
		return 0
	}

	// 移除markdown语法标记的简单处理
	text = regexp.MustCompile(`#+ `).ReplaceAllString(text, "")
	text = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(text, "$1")
	text = regexp.MustCompile(`\*([^*]+)\*`).ReplaceAllString(text, "$1")
	text = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`).ReplaceAllString(text, "$1")

	// 按空格分割计算单词数
	words := strings.Fields(text)
	return len(words)
}

// CalculateReadingTime 计算阅读时间
func (p *Post) CalculateReadingTime() int {
	return constants.CalculateReadingTime(p.WordCount)
}

// UpdateContentMetrics 更新内容指标
func (p *Post) UpdateContentMetrics() {
	p.WordCount = p.CalculateWordCount()
	p.ReadingTime = p.CalculateReadingTime()
}

// GenerateExcerpt 自动生成摘要
func (p *Post) GenerateExcerpt(maxLength int) string {
	if p.Markdown == "" {
		return ""
	}

	// 移除markdown标记
	text := regexp.MustCompile(`#+ `).ReplaceAllString(p.Markdown, "")
	text = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(text, "$1")
	text = regexp.MustCompile(`\*([^*]+)\*`).ReplaceAllString(text, "$1")
	text = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`).ReplaceAllString(text, "$1")

	// 移除多余的空白字符
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	// 截取指定长度
	if maxLength <= 0 {
		maxLength = 200
	}

	if utf8.RuneCountInString(text) <= maxLength {
		return text
	}

	runes := []rune(text)
	if len(runes) > maxLength {
		return string(runes[:maxLength]) + "..."
	}

	return text
}

// EnsureExcerpt 确保文章有摘要
func (p *Post) EnsureExcerpt() {
	if p.Excerpt == "" {
		p.Excerpt = p.GenerateExcerpt(200)
	}
}

// ===============================
// 转换方法
// ===============================

// ToDetailResponse 转换为文章详情响应
func (p *Post) ToDetailResponse(author *AuthorInfo) *PostDetailResponse {
	return &PostDetailResponse{
		ID:              p.ID.Hex(),
		Title:           p.Title,
		Slug:            p.Slug,
		Excerpt:         p.Excerpt,
		Markdown:        p.Markdown,
		HTML:            p.HTML,
		FeaturedImage:   p.FeaturedImage,
		Type:            p.Type,
		Status:          p.Status,
		Visibility:      p.Visibility,
		Author:          *author,
		Tags:            p.Tags,
		MetaTitle:       p.MetaTitle,
		MetaDescription: p.MetaDescription,
		CanonicalURL:    p.CanonicalURL,
		ReadingTime:     p.ReadingTime,
		WordCount:       p.WordCount,
		ViewCount:       p.ViewCount,
		PublishedAt:     p.PublishedAt,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

// ToListItem 转换为文章列表项
func (p *Post) ToListItem(author *AuthorInfo) *PostListItem {
	return &PostListItem{
		ID:            p.ID.Hex(),
		Title:         p.Title,
		Slug:          p.Slug,
		Excerpt:       p.Excerpt,
		FeaturedImage: p.FeaturedImage,
		Type:          p.Type,
		Status:        p.Status,
		Visibility:    p.Visibility,
		Author:        *author,
		Tags:          p.Tags,
		ReadingTime:   p.ReadingTime,
		ViewCount:     p.ViewCount,
		PublishedAt:   p.PublishedAt,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

// ===============================
// 工厂方法
// ===============================

// NewPost 创建新文章
func NewPost(title, markdown, postType, status, visibility string, authorID primitive.ObjectID) *Post {
	now := time.Now()

	post := &Post{
		ID:         primitive.NewObjectID(),
		Title:      title,
		Markdown:   markdown,
		Type:       postType,
		Status:     status,
		Visibility: visibility,
		AuthorID:   authorID,
		Tags:       []Tag{},
		ViewCount:  0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// 自动生成slug和摘要
	post.EnsureSlug()
	post.EnsureExcerpt()
	post.UpdateContentMetrics()

	return post
}

// NewPostFromCreateRequest 从创建请求创建文章
func NewPostFromCreateRequest(req *PostCreateRequest, authorID primitive.ObjectID) *Post {
	post := NewPost(req.Title, req.Markdown, req.Type, req.Status, req.Visibility, authorID)

	// 设置可选字段
	if req.Slug != "" {
		post.Slug = req.Slug
	}
	post.Excerpt = req.Excerpt
	post.FeaturedImage = req.FeaturedImage
	post.MetaTitle = req.MetaTitle
	post.MetaDescription = req.MetaDescription
	post.CanonicalURL = req.CanonicalURL
	post.PublishedAt = req.PublishedAt

	// 转换标签
	post.Tags = make([]Tag, len(req.Tags))
	for i, tagInfo := range req.Tags {
		post.Tags[i] = Tag{
			Name: tagInfo.Name,
			Slug: tagInfo.Slug,
		}
		// 如果没有提供slug，自动生成
		if post.Tags[i].Slug == "" {
			post.Tags[i].Slug = GenerateSlugFromText(tagInfo.Name)
		}
	}

	// 确保摘要存在
	if post.Excerpt == "" {
		post.EnsureExcerpt()
	}

	return post
}

// ===============================
// 数据库操作辅助方法
// ===============================

// PrepareForInsert 准备插入数据库
func (p *Post) PrepareForInsert() {
	now := time.Now()
	if p.ID.IsZero() {
		p.ID = primitive.NewObjectID()
	}
	p.CreatedAt = now
	p.UpdatedAt = now

	// 确保必要字段有值
	p.EnsureSlug()
	p.EnsureExcerpt()
	p.UpdateContentMetrics()

	// 设置默认值
	if p.Type == "" {
		p.Type = constants.PostTypePost
	}
	if p.Status == "" {
		p.Status = constants.PostStatusDraft
	}
	if p.Visibility == "" {
		p.Visibility = constants.PostVisibilityPublic
	}
	if p.Tags == nil {
		p.Tags = []Tag{}
	}
}

// PrepareForUpdate 准备更新数据库
func (p *Post) PrepareForUpdate() {
	p.UpdatedAt = time.Now()
	p.UpdateContentMetrics()
}

// IncrementViewCount 增加浏览量
func (p *Post) IncrementViewCount() {
	p.ViewCount++
	p.UpdatedAt = time.Now()
}

// Publish 发布文章
func (p *Post) Publish() {
	now := time.Now()
	p.Status = constants.PostStatusPublished
	if p.PublishedAt == nil {
		p.PublishedAt = &now
	}
	p.UpdatedAt = now
}

// Unpublish 取消发布
func (p *Post) Unpublish() {
	p.Status = constants.PostStatusDraft
	p.UpdatedAt = time.Now()
}

// ===============================
// 工具函数
// ===============================

// IsValidSlug 验证slug格式
func IsValidSlug(slug string) bool {
	if slug == "" {
		return false
	}

	// slug只能包含小写字母、数字和连字符，不能以连字符开头或结尾
	matched, _ := regexp.MatchString(`^[a-z0-9]+(-[a-z0-9]+)*$`, slug)
	return matched
}

// GenerateSlugFromText 从文本生成slug
func GenerateSlugFromText(text string) string {
	if text == "" {
		return "post-" + strconv.FormatInt(time.Now().Unix(), 10)
	}

	// 转换为小写
	slug := strings.ToLower(text)

	// 替换空格和特殊字符为连字符
	slug = regexp.MustCompile(`[^a-z0-9\s-]`).ReplaceAllString(slug, "")
	slug = regexp.MustCompile(`\s+`).ReplaceAllString(slug, "-")
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")

	// 移除首尾的连字符
	slug = strings.Trim(slug, "-")

	// 限制长度
	if len(slug) > 100 {
		slug = slug[:100]
		slug = strings.Trim(slug, "-")
	}

	// 如果结果为空，生成一个默认值
	if slug == "" {
		slug = "post-" + strconv.FormatInt(time.Now().Unix(), 10)
	}

	return slug
}

// ===============================
// 验证错误类型
// ===============================

// PostValidationError 文章验证错误
type PostValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error 实现error接口
func (e *PostValidationError) Error() string {
	return e.Message
}

// NewPostValidationError 创建文章验证错误
func NewPostValidationError(field, message string) *PostValidationError {
	return &PostValidationError{
		Field:   field,
		Message: message,
	}
}
