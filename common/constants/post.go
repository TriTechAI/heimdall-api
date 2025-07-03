package constants

// PostStatus 文章状态常量
const (
	PostStatusDraft     = "draft"     // 草稿
	PostStatusPublished = "published" // 已发布
	PostStatusScheduled = "scheduled" // 定时发布
	PostStatusArchived  = "archived"  // 已归档
	PostStatusTrash     = "trash"     // 回收站
)

// PostType 文章类型常量
const (
	PostTypePost = "post" // 文章
	PostTypePage = "page" // 页面
)

// PostVisibility 文章可见性常量
const (
	PostVisibilityPublic      = "public"       // 公开
	PostVisibilityMembersOnly = "members_only" // 仅会员可见
	PostVisibilityPrivate     = "private"      // 私有
)

// PostSortOrder 文章排序方式常量
const (
	PostSortByCreatedAt    = "created_at"    // 按创建时间排序
	PostSortByUpdatedAt    = "updated_at"    // 按更新时间排序
	PostSortByPublishedAt  = "published_at"  // 按发布时间排序
	PostSortByViewCount    = "view_count"    // 按浏览量排序
	PostSortByCommentCount = "comment_count" // 按评论数排序
	PostSortByTitle        = "title"         // 按标题排序
)

// PostValidation 文章验证相关常量
const (
	PostTitleMinLength        = 1       // 标题最小长度
	PostTitleMaxLength        = 255     // 标题最大长度
	PostSlugMinLength         = 1       // Slug最小长度
	PostSlugMaxLength         = 255     // Slug最大长度
	PostExcerptMaxLength      = 500     // 摘要最大长度
	PostContentMaxLength      = 1000000 // 内容最大长度（1MB）
	PostMetaTitleMaxLength    = 70      // SEO标题最大长度
	PostMetaDescMaxLength     = 160     // SEO描述最大长度
	PostCanonicalUrlMaxLength = 255     // 规范化URL最大长度
	PostTagMaxCount           = 20      // 最大标签数量
	PostTagNameMaxLength      = 50      // 标签名最大长度
)

// ReadingTime 阅读时间相关常量
const (
	ReadingSpeedWordsPerMinute = 200 // 平均阅读速度（每分钟字数）
	ReadingTimeMin             = 1   // 最小阅读时间（分钟）
	ReadingTimeMax             = 999 // 最大阅读时间（分钟）
)

// PostLimits 文章数量限制常量
const (
	PostsPerPageDefault = 10  // 默认每页文章数
	PostsPerPageMin     = 1   // 最小每页文章数
	PostsPerPageMax     = 100 // 最大每页文章数

	RelatedPostsCount = 5  // 相关文章数量
	PopularPostsCount = 10 // 热门文章数量
	RecentPostsCount  = 5  // 最新文章数量
)

// SearchLimits 搜索相关常量
const (
	SearchQueryMinLength = 2   // 搜索关键词最小长度
	SearchQueryMaxLength = 100 // 搜索关键词最大长度
	SearchResultsMax     = 50  // 最大搜索结果数
)

// GetAllPostStatuses 返回所有文章状态
func GetAllPostStatuses() []string {
	return []string{
		PostStatusDraft,
		PostStatusPublished,
		PostStatusScheduled,
		PostStatusArchived,
		PostStatusTrash,
	}
}

// GetAllPostTypes 返回所有文章类型
func GetAllPostTypes() []string {
	return []string{
		PostTypePost,
		PostTypePage,
	}
}

// GetAllPostVisibilities 返回所有可见性选项
func GetAllPostVisibilities() []string {
	return []string{
		PostVisibilityPublic,
		PostVisibilityMembersOnly,
		PostVisibilityPrivate,
	}
}

// GetAllPostSortOrders 返回所有排序方式
func GetAllPostSortOrders() []string {
	return []string{
		PostSortByCreatedAt,
		PostSortByUpdatedAt,
		PostSortByPublishedAt,
		PostSortByViewCount,
		PostSortByCommentCount,
		PostSortByTitle,
	}
}

// IsValidPostStatus 验证文章状态是否有效
func IsValidPostStatus(status string) bool {
	validStatuses := GetAllPostStatuses()
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// IsValidPostType 验证文章类型是否有效
func IsValidPostType(postType string) bool {
	validTypes := GetAllPostTypes()
	for _, validType := range validTypes {
		if postType == validType {
			return true
		}
	}
	return false
}

// IsValidPostVisibility 验证可见性是否有效
func IsValidPostVisibility(visibility string) bool {
	validVisibilities := GetAllPostVisibilities()
	for _, validVisibility := range validVisibilities {
		if visibility == validVisibility {
			return true
		}
	}
	return false
}

// IsValidPostSortOrder 验证排序方式是否有效
func IsValidPostSortOrder(sortOrder string) bool {
	validOrders := GetAllPostSortOrders()
	for _, validOrder := range validOrders {
		if sortOrder == validOrder {
			return true
		}
	}
	return false
}

// IsPublishedStatus 检查是否为已发布状态
func IsPublishedStatus(status string) bool {
	return status == PostStatusPublished
}

// IsPublicVisible 检查是否为公开可见
func IsPublicVisible(visibility string) bool {
	return visibility == PostVisibilityPublic
}

// CalculateReadingTime 计算阅读时间（分钟）
func CalculateReadingTime(wordCount int) int {
	if wordCount <= 0 {
		return ReadingTimeMin
	}

	readingTime := wordCount / ReadingSpeedWordsPerMinute
	if readingTime < ReadingTimeMin {
		return ReadingTimeMin
	}
	if readingTime > ReadingTimeMax {
		return ReadingTimeMax
	}

	return readingTime
}
