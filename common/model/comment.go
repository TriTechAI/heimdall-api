package model

import (
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/heimdall-api/common/constants"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Comment 评论模型
type Comment struct {
	ID            primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	PostID        primitive.ObjectID  `bson:"postId" json:"postId"`                         // 所属文章ID
	ParentID      *primitive.ObjectID `bson:"parentId,omitempty" json:"parentId"`           // 父评论ID（支持嵌套）
	Content       string              `bson:"content" json:"content"`                       // 评论内容
	AuthorName    string              `bson:"authorName" json:"authorName"`                 // 作者姓名
	AuthorEmail   string              `bson:"authorEmail" json:"authorEmail"`               // 作者邮箱
	AuthorWebsite string              `bson:"authorWebsite,omitempty" json:"authorWebsite"` // 作者网站
	AuthorIP      string              `bson:"authorIP" json:"authorIP"`                     // 作者IP地址
	UserAgent     string              `bson:"userAgent,omitempty" json:"userAgent"`         // 用户代理
	Status        string              `bson:"status" json:"status"`                         // 评论状态
	Visibility    string              `bson:"visibility" json:"visibility"`                 // 可见性
	Type          string              `bson:"type" json:"type"`                             // 评论类型
	Level         int                 `bson:"level" json:"level"`                           // 嵌套层级
	ReplyCount    int                 `bson:"replyCount" json:"replyCount"`                 // 回复数量
	LikeCount     int                 `bson:"likeCount" json:"likeCount"`                   // 点赞数量
	CreatedAt     time.Time           `bson:"createdAt" json:"createdAt"`                   // 创建时间
	UpdatedAt     time.Time           `bson:"updatedAt" json:"updatedAt"`                   // 更新时间
	ApprovedAt    *time.Time          `bson:"approvedAt,omitempty" json:"approvedAt"`       // 审核通过时间
}

// CommentCreateRequest 评论创建请求
type CommentCreateRequest struct {
	PostID        string `json:"postId" validate:"required"`
	ParentID      string `json:"parentId,omitempty"`
	Content       string `json:"content" validate:"required,min=1,max=1000"`
	AuthorName    string `json:"authorName" validate:"required,min=1,max=100"`
	AuthorEmail   string `json:"authorEmail" validate:"required,email,max=255"`
	AuthorWebsite string `json:"authorWebsite,omitempty,url,max=255"`
}

// CommentUpdateRequest 评论更新请求
type CommentUpdateRequest struct {
	Content       string `json:"content,omitempty" validate:"omitempty,min=1,max=1000"`
	AuthorName    string `json:"authorName,omitempty" validate:"omitempty,min=1,max=100"`
	AuthorEmail   string `json:"authorEmail,omitempty" validate:"omitempty,email,max=255"`
	AuthorWebsite string `json:"authorWebsite,omitempty,url,max=255"`
	Status        string `json:"status,omitempty"`
	Visibility    string `json:"visibility,omitempty"`
}

// CommentFilter 评论过滤器
type CommentFilter struct {
	PostID      string    `json:"postId,omitempty"`
	ParentID    string    `json:"parentId,omitempty"`
	AuthorEmail string    `json:"authorEmail,omitempty"`
	AuthorIP    string    `json:"authorIP,omitempty"`
	Status      string    `json:"status,omitempty"`
	Visibility  string    `json:"visibility,omitempty"`
	Type        string    `json:"type,omitempty"`
	Level       int       `json:"level,omitempty"`
	Keyword     string    `json:"keyword,omitempty"`
	StartTime   time.Time `json:"startTime,omitempty"`
	EndTime     time.Time `json:"endTime,omitempty"`
	SortBy      string    `json:"sortBy,omitempty"`
	SortDesc    bool      `json:"sortDesc,omitempty"`
}

// CommentListItem 评论列表项
type CommentListItem struct {
	ID            string             `json:"id"`
	PostID        string             `json:"postId"`
	ParentID      string             `json:"parentId,omitempty"`
	Content       string             `json:"content"`
	AuthorName    string             `json:"authorName"`
	AuthorEmail   string             `json:"authorEmail"`
	AuthorWebsite string             `json:"authorWebsite,omitempty"`
	Status        string             `json:"status"`
	Type          string             `json:"type"`
	Level         int                `json:"level"`
	ReplyCount    int                `json:"replyCount"`
	LikeCount     int                `json:"likeCount"`
	CreatedAt     time.Time          `json:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt"`
	ApprovedAt    *time.Time         `json:"approvedAt,omitempty"`
	Replies       []*CommentListItem `json:"replies,omitempty"` // 子评论
}

// CommentDetailResponse 评论详情响应
type CommentDetailResponse struct {
	ID            string     `json:"id"`
	PostID        string     `json:"postId"`
	ParentID      string     `json:"parentId,omitempty"`
	Content       string     `json:"content"`
	AuthorName    string     `json:"authorName"`
	AuthorEmail   string     `json:"authorEmail"`
	AuthorWebsite string     `json:"authorWebsite,omitempty"`
	Status        string     `json:"status"`
	Visibility    string     `json:"visibility"`
	Type          string     `json:"type"`
	Level         int        `json:"level"`
	ReplyCount    int        `json:"replyCount"`
	LikeCount     int        `json:"likeCount"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	ApprovedAt    *time.Time `json:"approvedAt,omitempty"`
}

// ===============================
// 验证方法
// ===============================

// ValidateForCreate 验证评论创建数据
func (c *Comment) ValidateForCreate() error {
	// 验证必填字段
	if c.PostID.IsZero() {
		return NewCommentValidationError("postId", "文章ID不能为空")
	}
	if strings.TrimSpace(c.Content) == "" {
		return NewCommentValidationError("content", "评论内容不能为空")
	}
	if strings.TrimSpace(c.AuthorName) == "" {
		return NewCommentValidationError("authorName", "作者姓名不能为空")
	}
	if strings.TrimSpace(c.AuthorEmail) == "" {
		return NewCommentValidationError("authorEmail", "作者邮箱不能为空")
	}

	// 验证字段长度
	if len(c.Content) > constants.CommentContentMaxLength {
		return NewCommentValidationError("content", "评论内容长度不能超过1000字符")
	}
	if len(c.AuthorName) > constants.CommentAuthorNameMaxLength {
		return NewCommentValidationError("authorName", "作者姓名长度不能超过100字符")
	}
	if len(c.AuthorEmail) > constants.CommentAuthorEmailMaxLength {
		return NewCommentValidationError("authorEmail", "作者邮箱长度不能超过255字符")
	}
	if len(c.AuthorWebsite) > constants.CommentAuthorWebsiteMaxLength {
		return NewCommentValidationError("authorWebsite", "作者网站长度不能超过255字符")
	}

	// 验证邮箱格式
	if !isValidEmail(c.AuthorEmail) {
		return NewCommentValidationError("authorEmail", "邮箱格式无效")
	}

	// 验证网站URL格式
	if c.AuthorWebsite != "" && !isValidURL(c.AuthorWebsite) {
		return NewCommentValidationError("authorWebsite", "网站URL格式无效")
	}

	// 验证IP地址格式
	if c.AuthorIP != "" && net.ParseIP(c.AuthorIP) == nil {
		return NewCommentValidationError("authorIP", "IP地址格式无效")
	}

	// 验证枚举值
	if c.Status != "" && !constants.IsValidCommentStatus(c.Status) {
		return NewCommentValidationError("status", "无效的评论状态")
	}
	if c.Visibility != "" && !constants.IsValidCommentVisibility(c.Visibility) {
		return NewCommentValidationError("visibility", "无效的评论可见性")
	}
	if c.Type != "" && !constants.IsValidCommentType(c.Type) {
		return NewCommentValidationError("type", "无效的评论类型")
	}

	// 验证嵌套层级
	if c.Level > constants.CommentMaxNestingLevel {
		return NewCommentValidationError("level", "评论嵌套层级不能超过3层")
	}

	return nil
}

// ValidateForUpdate 验证评论更新数据
func (c *Comment) ValidateForUpdate() error {
	// 对于更新，只验证非空字段
	if c.Content != "" && len(c.Content) > constants.CommentContentMaxLength {
		return NewCommentValidationError("content", "评论内容长度不能超过1000字符")
	}
	if c.AuthorName != "" && len(c.AuthorName) > constants.CommentAuthorNameMaxLength {
		return NewCommentValidationError("authorName", "作者姓名长度不能超过100字符")
	}
	if c.AuthorEmail != "" {
		if len(c.AuthorEmail) > constants.CommentAuthorEmailMaxLength {
			return NewCommentValidationError("authorEmail", "作者邮箱长度不能超过255字符")
		}
		if !isValidEmail(c.AuthorEmail) {
			return NewCommentValidationError("authorEmail", "邮箱格式无效")
		}
	}
	if c.AuthorWebsite != "" {
		if len(c.AuthorWebsite) > constants.CommentAuthorWebsiteMaxLength {
			return NewCommentValidationError("authorWebsite", "作者网站长度不能超过255字符")
		}
		if !isValidURL(c.AuthorWebsite) {
			return NewCommentValidationError("authorWebsite", "网站URL格式无效")
		}
	}

	// 验证枚举值
	if c.Status != "" && !constants.IsValidCommentStatus(c.Status) {
		return NewCommentValidationError("status", "无效的评论状态")
	}
	if c.Visibility != "" && !constants.IsValidCommentVisibility(c.Visibility) {
		return NewCommentValidationError("visibility", "无效的评论可见性")
	}

	return nil
}

// ===============================
// 状态检查方法
// ===============================

// IsPending 检查评论是否待审核
func (c *Comment) IsPending() bool {
	return c.Status == constants.CommentStatusPending
}

// IsApproved 检查评论是否已通过审核
func (c *Comment) IsApproved() bool {
	return c.Status == constants.CommentStatusApproved
}

// IsRejected 检查评论是否已被拒绝
func (c *Comment) IsRejected() bool {
	return c.Status == constants.CommentStatusRejected
}

// IsSpam 检查评论是否为垃圾评论
func (c *Comment) IsSpam() bool {
	return c.Status == constants.CommentStatusSpam
}

// IsPublic 检查评论是否公开可见
func (c *Comment) IsPublic() bool {
	return c.Visibility == constants.CommentVisibilityPublic
}

// IsReply 检查是否为回复评论
func (c *Comment) IsReply() bool {
	return c.ParentID != nil && !c.ParentID.IsZero()
}

// CanBeApproved 检查评论是否可以被审核通过
func (c *Comment) CanBeApproved() bool {
	return c.Status == constants.CommentStatusPending
}

// ===============================
// 业务方法
// ===============================

// Approve 通过审核
func (c *Comment) Approve() {
	now := time.Now()
	c.Status = constants.CommentStatusApproved
	c.UpdatedAt = now
	c.ApprovedAt = &now
}

// Reject 拒绝审核
func (c *Comment) Reject() {
	c.Status = constants.CommentStatusRejected
	c.UpdatedAt = time.Now()
	c.ApprovedAt = nil
}

// MarkAsSpam 标记为垃圾评论
func (c *Comment) MarkAsSpam() {
	c.Status = constants.CommentStatusSpam
	c.UpdatedAt = time.Now()
	c.ApprovedAt = nil
}

// IncrementReplyCount 增加回复数量
func (c *Comment) IncrementReplyCount() {
	c.ReplyCount++
	c.UpdatedAt = time.Now()
}

// IncrementLikeCount 增加点赞数量
func (c *Comment) IncrementLikeCount() {
	c.LikeCount++
	c.UpdatedAt = time.Now()
}

// ===============================
// 转换方法
// ===============================

// ToListItem 转换为列表项
func (c *Comment) ToListItem() *CommentListItem {
	item := &CommentListItem{
		ID:            c.ID.Hex(),
		PostID:        c.PostID.Hex(),
		Content:       c.Content,
		AuthorName:    c.AuthorName,
		AuthorEmail:   maskEmail(c.AuthorEmail), // 脱敏处理
		AuthorWebsite: c.AuthorWebsite,
		Status:        c.Status,
		Type:          c.Type,
		Level:         c.Level,
		ReplyCount:    c.ReplyCount,
		LikeCount:     c.LikeCount,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
		ApprovedAt:    c.ApprovedAt,
	}

	if c.ParentID != nil {
		item.ParentID = c.ParentID.Hex()
	}

	return item
}

// ToDetailResponse 转换为详情响应
func (c *Comment) ToDetailResponse() *CommentDetailResponse {
	response := &CommentDetailResponse{
		ID:            c.ID.Hex(),
		PostID:        c.PostID.Hex(),
		Content:       c.Content,
		AuthorName:    c.AuthorName,
		AuthorEmail:   maskEmail(c.AuthorEmail), // 脱敏处理
		AuthorWebsite: c.AuthorWebsite,
		Status:        c.Status,
		Visibility:    c.Visibility,
		Type:          c.Type,
		Level:         c.Level,
		ReplyCount:    c.ReplyCount,
		LikeCount:     c.LikeCount,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
		ApprovedAt:    c.ApprovedAt,
	}

	if c.ParentID != nil {
		response.ParentID = c.ParentID.Hex()
	}

	return response
}

// ===============================
// 工厂方法
// ===============================

// NewComment 创建新评论
func NewComment(postID primitive.ObjectID, content, authorName, authorEmail, authorIP string) *Comment {
	now := time.Now()

	return &Comment{
		ID:          primitive.NewObjectID(),
		PostID:      postID,
		Content:     content,
		AuthorName:  authorName,
		AuthorEmail: authorEmail,
		AuthorIP:    authorIP,
		Status:      constants.CommentStatusPending, // 默认待审核
		Visibility:  constants.CommentVisibilityPublic,
		Type:        constants.CommentTypeComment,
		Level:       0,
		ReplyCount:  0,
		LikeCount:   0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewReplyComment 创建回复评论
func NewReplyComment(postID, parentID primitive.ObjectID, content, authorName, authorEmail, authorIP string, level int) *Comment {
	now := time.Now()

	return &Comment{
		ID:          primitive.NewObjectID(),
		PostID:      postID,
		ParentID:    &parentID,
		Content:     content,
		AuthorName:  authorName,
		AuthorEmail: authorEmail,
		AuthorIP:    authorIP,
		Status:      constants.CommentStatusPending, // 默认待审核
		Visibility:  constants.CommentVisibilityPublic,
		Type:        constants.CommentTypeReply,
		Level:       level,
		ReplyCount:  0,
		LikeCount:   0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewCommentFromCreateRequest 从创建请求创建评论
func NewCommentFromCreateRequest(req *CommentCreateRequest, authorIP string) (*Comment, error) {
	// 解析PostID
	postObjID, err := primitive.ObjectIDFromHex(req.PostID)
	if err != nil {
		return nil, NewCommentValidationError("postId", "无效的文章ID格式")
	}

	var comment *Comment

	// 处理回复评论
	if req.ParentID != "" {
		parentObjID, err := primitive.ObjectIDFromHex(req.ParentID)
		if err != nil {
			return nil, NewCommentValidationError("parentId", "无效的父评论ID格式")
		}
		// 这里应该查询父评论来确定层级，暂时设为1
		comment = NewReplyComment(postObjID, parentObjID, req.Content, req.AuthorName, req.AuthorEmail, authorIP, 1)
	} else {
		comment = NewComment(postObjID, req.Content, req.AuthorName, req.AuthorEmail, authorIP)
	}

	// 设置可选字段
	if req.AuthorWebsite != "" {
		comment.AuthorWebsite = req.AuthorWebsite
	}

	return comment, nil
}

// ===============================
// 数据库操作辅助方法
// ===============================

// PrepareForInsert 准备插入数据库
func (c *Comment) PrepareForInsert() {
	now := time.Now()
	if c.ID.IsZero() {
		c.ID = primitive.NewObjectID()
	}
	c.CreatedAt = now
	c.UpdatedAt = now

	// 设置默认值
	if c.Status == "" {
		c.Status = constants.CommentStatusPending
	}
	if c.Visibility == "" {
		c.Visibility = constants.CommentVisibilityPublic
	}
	if c.Type == "" {
		if c.ParentID != nil && !c.ParentID.IsZero() {
			c.Type = constants.CommentTypeReply
		} else {
			c.Type = constants.CommentTypeComment
		}
	}
}

// PrepareForUpdate 准备更新数据库
func (c *Comment) PrepareForUpdate() {
	c.UpdatedAt = time.Now()
}

// ===============================
// 工具函数
// ===============================

// isValidEmail 验证邮箱格式
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// isValidURL 验证URL格式
func isValidURL(url string) bool {
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	return urlRegex.MatchString(url)
}

// maskEmail 邮箱脱敏处理
func maskEmail(email string) string {
	if email == "" {
		return ""
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	username := parts[0]
	domain := parts[1]

	if len(username) <= 2 {
		return email // 太短不处理
	}

	maskedUsername := string(username[0]) + "***" + string(username[len(username)-1])
	return maskedUsername + "@" + domain
}

// ===============================
// 验证错误类型
// ===============================

// CommentValidationError 评论验证错误
type CommentValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error 实现error接口
func (e *CommentValidationError) Error() string {
	return e.Message
}

// NewCommentValidationError 创建评论验证错误
func NewCommentValidationError(field, message string) *CommentValidationError {
	return &CommentValidationError{
		Field:   field,
		Message: message,
	}
}
