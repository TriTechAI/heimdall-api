package constants

// 评论状态常量
const (
	CommentStatusPending  = "pending"  // 待审核
	CommentStatusApproved = "approved" // 已通过
	CommentStatusRejected = "rejected" // 已拒绝
	CommentStatusSpam     = "spam"     // 垃圾评论
	CommentStatusDeleted  = "deleted"  // 已删除（软删除）
)

// 评论可见性常量
const (
	CommentVisibilityPublic  = "public"  // 公开可见
	CommentVisibilityPrivate = "private" // 私有（仅作者和管理员可见）
)

// 评论类型常量
const (
	CommentTypeComment = "comment" // 普通评论
	CommentTypeReply   = "reply"   // 回复评论
)

// 评论字段长度限制
const (
	CommentContentMaxLength       = 1000 // 评论内容最大长度
	CommentAuthorNameMaxLength    = 100  // 作者姓名最大长度
	CommentAuthorEmailMaxLength   = 255  // 作者邮箱最大长度
	CommentAuthorWebsiteMaxLength = 255  // 作者网站最大长度
	CommentIPMaxLength            = 45   // IP地址最大长度（IPv6）
	CommentUserAgentMaxLength     = 500  // User Agent最大长度
)

// 评论业务限制
const (
	CommentMaxNestingLevel = 3   // 最大嵌套层级
	CommentMinInterval     = 60  // 最小评论间隔（秒）
	CommentDefaultPageSize = 20  // 默认分页大小
	CommentMaxPageSize     = 100 // 最大分页大小
)

// 评论状态验证
func IsValidCommentStatus(status string) bool {
	switch status {
	case CommentStatusPending, CommentStatusApproved, CommentStatusRejected, CommentStatusSpam, CommentStatusDeleted:
		return true
	default:
		return false
	}
}

// 评论可见性验证
func IsValidCommentVisibility(visibility string) bool {
	switch visibility {
	case CommentVisibilityPublic, CommentVisibilityPrivate:
		return true
	default:
		return false
	}
}

// 评论类型验证
func IsValidCommentType(commentType string) bool {
	switch commentType {
	case CommentTypeComment, CommentTypeReply:
		return true
	default:
		return false
	}
}

// 获取所有有效的评论状态
func GetAllCommentStatuses() []string {
	return []string{
		CommentStatusPending,
		CommentStatusApproved,
		CommentStatusRejected,
		CommentStatusSpam,
		CommentStatusDeleted,
	}
}

// 获取所有有效的评论可见性
func GetAllCommentVisibilities() []string {
	return []string{
		CommentVisibilityPublic,
		CommentVisibilityPrivate,
	}
}

// 获取所有有效的评论类型
func GetAllCommentTypes() []string {
	return []string{
		CommentTypeComment,
		CommentTypeReply,
	}
}
