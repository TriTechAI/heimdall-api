package constants

// TagVisibility 标签可见性常量
const (
	TagVisibilityPublic   = "public"   // 公开
	TagVisibilityInternal = "internal" // 内部
)

// TagValidation 标签验证相关常量
const (
	TagNameMinLength          = 1   // 标签名最小长度
	TagNameMaxLength          = 50  // 标签名最大长度
	TagSlugMinLength          = 1   // Slug最小长度
	TagSlugMaxLength          = 50  // Slug最大长度
	TagDescriptionMaxLength   = 300 // 描述最大长度
	TagMetaTitleMaxLength     = 70  // SEO标题最大长度
	TagMetaDescMaxLength      = 160 // SEO描述最大长度
	TagFeaturedImageMaxLength = 500 // 特色图片URL最大长度
)

// TagDefaults 标签默认值常量
const (
	TagDefaultColor = "#6B7280" // 默认标签颜色（灰色）
)

// TagLimits 标签数量限制常量
const (
	TagsPerPageDefault = 20  // 默认每页标签数
	TagsPerPageMin     = 1   // 最小每页标签数
	TagsPerPageMax     = 100 // 最大每页标签数

	PopularTagsCount = 10 // 热门标签数量
	RecentTagsCount  = 5  // 最新标签数量
)

// TagSortOrder 标签排序方式常量
const (
	TagSortByName      = "name"       // 按名称排序
	TagSortBySlug      = "slug"       // 按Slug排序
	TagSortByPostCount = "post_count" // 按文章数排序
	TagSortByCreatedAt = "created_at" // 按创建时间排序
	TagSortByUpdatedAt = "updated_at" // 按更新时间排序
)

// GetAllTagVisibilities 返回所有标签可见性选项
func GetAllTagVisibilities() []string {
	return []string{
		TagVisibilityPublic,
		TagVisibilityInternal,
	}
}

// GetAllTagSortOrders 返回所有标签排序方式
func GetAllTagSortOrders() []string {
	return []string{
		TagSortByName,
		TagSortBySlug,
		TagSortByPostCount,
		TagSortByCreatedAt,
		TagSortByUpdatedAt,
	}
}

// IsValidTagVisibility 验证标签可见性是否有效
func IsValidTagVisibility(visibility string) bool {
	validVisibilities := GetAllTagVisibilities()
	for _, validVisibility := range validVisibilities {
		if visibility == validVisibility {
			return true
		}
	}
	return false
}

// IsValidTagSortOrder 验证标签排序方式是否有效
func IsValidTagSortOrder(sortOrder string) bool {
	validOrders := GetAllTagSortOrders()
	for _, validOrder := range validOrders {
		if sortOrder == validOrder {
			return true
		}
	}
	return false
}

// IsPublicTagVisibility 检查是否为公开可见
func IsPublicTagVisibility(visibility string) bool {
	return visibility == TagVisibilityPublic
}

// IsInternalTagVisibility 检查是否为内部可见
func IsInternalTagVisibility(visibility string) bool {
	return visibility == TagVisibilityInternal
}
