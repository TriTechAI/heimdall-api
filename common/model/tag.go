package model

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/heimdall-api/common/constants"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TagModel 独立标签模型 - 完整的标签管理
type TagModel struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name            string             `bson:"name" json:"name"`                       // 标签名称
	Slug            string             `bson:"slug" json:"slug"`                       // URL友好标识，唯一
	Description     string             `bson:"description" json:"description"`         // 标签描述
	Color           string             `bson:"color" json:"color"`                     // 标签颜色
	FeaturedImage   string             `bson:"featuredImage" json:"featuredImage"`     // 特色图片
	MetaTitle       string             `bson:"metaTitle" json:"metaTitle"`             // SEO标题
	MetaDescription string             `bson:"metaDescription" json:"metaDescription"` // SEO描述
	PostCount       int                `bson:"postCount" json:"postCount"`             // 关联文章数量
	Visibility      string             `bson:"visibility" json:"visibility"`           // 可见性
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`             // 创建时间
	UpdatedAt       time.Time          `bson:"updatedAt" json:"updatedAt"`             // 更新时间
}

// tagSlugRegex 标签slug验证正则表达式
var tagSlugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// tagColorRegex 颜色值验证正则表达式（支持hex格式，大小写不敏感）
var tagColorRegex = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// Validate 验证标签数据有效性
func (t *TagModel) Validate() error {
	// 验证标签名称
	if err := t.validateName(); err != nil {
		return err
	}

	// 验证slug
	if err := t.validateSlug(); err != nil {
		return err
	}

	// 验证描述
	if err := t.validateDescription(); err != nil {
		return err
	}

	// 验证颜色
	if err := t.validateColor(); err != nil {
		return err
	}

	// 验证特色图片
	if err := t.validateFeaturedImage(); err != nil {
		return err
	}

	// 验证SEO字段
	if err := t.validateSEOFields(); err != nil {
		return err
	}

	// 验证可见性
	if err := t.validateVisibility(); err != nil {
		return err
	}

	return nil
}

// validateName 验证标签名称
func (t *TagModel) validateName() error {
	if strings.TrimSpace(t.Name) == "" {
		return fmt.Errorf("tag name cannot be empty")
	}

	nameLength := utf8.RuneCountInString(t.Name)
	if nameLength < constants.TagNameMinLength {
		return fmt.Errorf("tag name must be at least %d characters", constants.TagNameMinLength)
	}
	if nameLength > constants.TagNameMaxLength {
		return fmt.Errorf("tag name cannot exceed %d characters", constants.TagNameMaxLength)
	}

	return nil
}

// validateSlug 验证标签slug
func (t *TagModel) validateSlug() error {
	if strings.TrimSpace(t.Slug) == "" {
		return fmt.Errorf("tag slug cannot be empty")
	}

	slugLength := len(t.Slug)
	if slugLength < constants.TagSlugMinLength {
		return fmt.Errorf("tag slug must be at least %d characters", constants.TagSlugMinLength)
	}
	if slugLength > constants.TagSlugMaxLength {
		return fmt.Errorf("tag slug cannot exceed %d characters", constants.TagSlugMaxLength)
	}

	if !tagSlugRegex.MatchString(t.Slug) {
		return fmt.Errorf("tag slug must contain only lowercase letters, numbers, and hyphens")
	}

	return nil
}

// validateDescription 验证标签描述
func (t *TagModel) validateDescription() error {
	if t.Description != "" {
		descLength := utf8.RuneCountInString(t.Description)
		if descLength > constants.TagDescriptionMaxLength {
			return fmt.Errorf("tag description cannot exceed %d characters", constants.TagDescriptionMaxLength)
		}
	}
	return nil
}

// validateColor 验证标签颜色
func (t *TagModel) validateColor() error {
	if t.Color != "" {
		if !tagColorRegex.MatchString(t.Color) {
			return fmt.Errorf("tag color must be a valid hex color (e.g., #FF5733)")
		}
	}
	return nil
}

// validateFeaturedImage 验证特色图片
func (t *TagModel) validateFeaturedImage() error {
	if t.FeaturedImage != "" {
		imageLength := len(t.FeaturedImage)
		if imageLength > constants.TagFeaturedImageMaxLength {
			return fmt.Errorf("featured image URL cannot exceed %d characters", constants.TagFeaturedImageMaxLength)
		}
	}
	return nil
}

// validateSEOFields 验证SEO相关字段
func (t *TagModel) validateSEOFields() error {
	// 验证MetaTitle
	if t.MetaTitle != "" {
		metaTitleLength := utf8.RuneCountInString(t.MetaTitle)
		if metaTitleLength > constants.TagMetaTitleMaxLength {
			return fmt.Errorf("meta title cannot exceed %d characters", constants.TagMetaTitleMaxLength)
		}
	}

	// 验证MetaDescription
	if t.MetaDescription != "" {
		metaDescLength := utf8.RuneCountInString(t.MetaDescription)
		if metaDescLength > constants.TagMetaDescMaxLength {
			return fmt.Errorf("meta description cannot exceed %d characters", constants.TagMetaDescMaxLength)
		}
	}

	return nil
}

// validateVisibility 验证可见性
func (t *TagModel) validateVisibility() error {
	if t.Visibility != "" && !constants.IsValidTagVisibility(t.Visibility) {
		return fmt.Errorf("invalid visibility: %s", t.Visibility)
	}
	return nil
}

// IsPublic 检查标签是否为公开可见
func (t *TagModel) IsPublic() bool {
	return t.Visibility == constants.TagVisibilityPublic
}

// IsInternal 检查标签是否为内部可见
func (t *TagModel) IsInternal() bool {
	return t.Visibility == constants.TagVisibilityInternal
}

// SetDefaultVisibility 设置默认可见性
func (t *TagModel) SetDefaultVisibility() {
	if t.Visibility == "" {
		t.Visibility = constants.TagVisibilityPublic
	}
}

// SetDefaultColor 设置默认颜色
func (t *TagModel) SetDefaultColor() {
	if t.Color == "" {
		t.Color = constants.TagDefaultColor
	}
}

// PrepareForCreation 为创建准备数据
func (t *TagModel) PrepareForCreation() {
	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now
	t.PostCount = 0 // 新标签初始文章数为0
	t.SetDefaultVisibility()
	t.SetDefaultColor()
}

// PrepareForUpdate 为更新准备数据
func (t *TagModel) PrepareForUpdate() {
	t.UpdatedAt = time.Now()
}

// GenerateSlugFromName 从名称生成slug
func (t *TagModel) GenerateSlugFromName() {
	if t.Name != "" && t.Slug == "" {
		t.Slug = generateSlugFromText(t.Name)
	}
}

// generateSlugFromText 从文本生成URL友好的slug
func generateSlugFromText(text string) string {
	// 转换为小写
	slug := strings.ToLower(text)

	// 将下划线替换为连字符
	slug = strings.ReplaceAll(slug, "_", "-")

	// 移除特殊字符，保留字母、数字、空格和连字符
	reg := regexp.MustCompile(`[^a-z0-9\s-]+`)
	slug = reg.ReplaceAllString(slug, "")

	// 将空格替换为连字符
	slug = regexp.MustCompile(`\s+`).ReplaceAllString(slug, "-")

	// 合并多个连续连字符为单个连字符
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")

	// 移除开头和结尾的连字符
	slug = strings.Trim(slug, "-")

	// 限制长度
	if len(slug) > constants.TagSlugMaxLength {
		slug = slug[:constants.TagSlugMaxLength]
		slug = strings.TrimRight(slug, "-")
	}

	return slug
}

// ToPublicResponse 转换为公开响应格式（过滤敏感信息）
func (t *TagModel) ToPublicResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":          t.ID.Hex(),
		"name":        t.Name,
		"slug":        t.Slug,
		"description": t.Description,
		"color":       t.Color,
		"postCount":   t.PostCount,
		"createdAt":   t.CreatedAt,
	}
}

// ToAdminResponse 转换为管理员响应格式（包含所有信息）
func (t *TagModel) ToAdminResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":              t.ID.Hex(),
		"name":            t.Name,
		"slug":            t.Slug,
		"description":     t.Description,
		"color":           t.Color,
		"featuredImage":   t.FeaturedImage,
		"metaTitle":       t.MetaTitle,
		"metaDescription": t.MetaDescription,
		"postCount":       t.PostCount,
		"visibility":      t.Visibility,
		"createdAt":       t.CreatedAt,
		"updatedAt":       t.UpdatedAt,
	}
}
