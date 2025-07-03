package constants

// 错误码结构: EHHCCCC
// E: 固定前缀
// HH: 服务代码 (01=Admin, 02=Public, 99=Common)
// CCCC: 错误序号

// ====================
// 通用错误码 (E99xxxx)
// ====================

const (
	// 系统级错误
	ErrInternalServer  = "E990001" // 内部服务器错误
	ErrDatabaseError   = "E990002" // 数据库错误
	ErrRedisError      = "E990003" // Redis错误
	ErrFileSystemError = "E990004" // 文件系统错误
	ErrNetworkError    = "E990005" // 网络错误

	// 请求相关错误
	ErrInvalidParams = "E990101" // 请求参数无效
	ErrInvalidJSON   = "E990102" // JSON格式错误
	ErrMissingParams = "E990103" // 缺少必需参数
	ErrParamTooLong  = "E990104" // 参数过长
	ErrParamTooShort = "E990105" // 参数过短
	ErrInvalidFormat = "E990106" // 格式错误
	ErrInvalidEmail  = "E990107" // 邮箱格式错误
	ErrInvalidURL    = "E990108" // URL格式错误

	// 认证授权错误
	ErrUnauthorized     = "E990201" // 未认证
	ErrForbidden        = "E990202" // 权限不足
	ErrTokenExpired     = "E990203" // 令牌过期
	ErrTokenInvalid     = "E990204" // 令牌无效
	ErrTokenBlacklisted = "E990205" // 令牌已被拉黑

	// 资源相关错误
	ErrNotFound        = "E990301" // 资源不存在
	ErrConflict        = "E990302" // 资源冲突
	ErrGone            = "E990303" // 资源已删除
	ErrLocked          = "E990304" // 资源被锁定
	ErrTooManyRequests = "E990305" // 请求过于频繁

	// 限流相关错误
	ErrRateLimit       = "E990401" // 达到频率限制
	ErrIPBlocked       = "E990402" // IP被封禁
	ErrTooManyFailures = "E990403" // 失败次数过多
	ErrConcurrentLimit = "E990404" // 并发限制
)

// ====================
// Admin API错误码 (E01xxxx)
// ====================

const (
	// 用户相关错误
	ErrUserNotFound           = "E010001" // 用户不存在
	ErrInvalidPassword        = "E010002" // 密码错误
	ErrUserLocked             = "E010003" // 用户被锁定
	ErrUserInactive           = "E010004" // 用户未激活
	ErrUserSuspended          = "E010005" // 用户被暂停
	ErrUsernameExists         = "E010006" // 用户名已存在
	ErrEmailExists            = "E010007" // 邮箱已存在
	ErrInvalidRole            = "E010008" // 无效的用户角色
	ErrPasswordTooWeak        = "E010009" // 密码强度不足
	ErrPasswordExpired        = "E010010" // 密码已过期
	ErrPasswordReused         = "E010011" // 密码不能重复使用
	ErrInsufficientPermission = "E010012" // 权限不足
	ErrCannotDeleteSelf       = "E010013" // 不能删除自己
	ErrLastOwner              = "E010014" // 不能删除最后一个所有者

	// 文章相关错误
	ErrPostNotFound         = "E010101" // 文章不存在
	ErrPostSlugExists       = "E010102" // 文章Slug已存在
	ErrInvalidPostStatus    = "E010103" // 无效的文章状态
	ErrInvalidPostType      = "E010104" // 无效的文章类型
	ErrPostTitleEmpty       = "E010105" // 文章标题为空
	ErrPostContentEmpty     = "E010106" // 文章内容为空
	ErrPostSlugEmpty        = "E010107" // 文章Slug为空
	ErrPostSlugInvalid      = "E010108" // 文章Slug格式无效
	ErrPostNotAuthor        = "E010109" // 不是文章作者
	ErrPostCannotPublish    = "E010110" // 文章无法发布
	ErrPostAlreadyPublished = "E010111" // 文章已发布
	ErrTooManyTags          = "E010112" // 标签数量过多

	// 评论相关错误
	ErrCommentNotFound      = "E010201" // 评论不存在
	ErrCommentClosed        = "E010202" // 评论已关闭
	ErrCommentApproved      = "E010203" // 评论已审核
	ErrCommentRejected      = "E010204" // 评论已拒绝
	ErrCommentSpam          = "E010205" // 评论是垃圾信息
	ErrCannotReplyToComment = "E010206" // 无法回复此评论

	// 媒体文件相关错误
	ErrMediaNotFound        = "E010301" // 媒体文件不存在
	ErrFileTypeNotAllowed   = "E010302" // 文件类型不支持
	ErrFileTooLarge         = "E010303" // 文件过大
	ErrFileUploadFailed     = "E010304" // 文件上传失败
	ErrImageProcessFailed   = "E010305" // 图片处理失败
	ErrStorageQuotaExceeded = "E010306" // 存储配额已满

	// 设置相关错误
	ErrSettingNotFound     = "E010401" // 设置项不存在
	ErrInvalidSettingValue = "E010402" // 设置值无效
	ErrSettingReadOnly     = "E010403" // 设置项只读

	// 标签相关错误
	ErrTagNotFound   = "E010501" // 标签不存在
	ErrTagSlugExists = "E010502" // 标签Slug已存在
	ErrTagNameExists = "E010503" // 标签名已存在
	ErrTagInUse      = "E010504" // 标签正在使用中
)

// ====================
// Public API错误码 (E02xxxx)
// ====================

const (
	// 内容访问错误
	ErrPostNotPublished = "E020001" // 文章未发布
	ErrPostPrivate      = "E020002" // 文章为私有
	ErrPostMembersOnly  = "E020003" // 文章仅会员可见
	ErrPageNotFound     = "E020004" // 页面不存在
	ErrContentNotFound  = "E020005" // 内容不存在

	// 评论相关错误
	ErrCommentNotAllowed   = "E020101" // 不允许评论
	ErrCommentTooLong      = "E020102" // 评论内容过长
	ErrCommentTooShort     = "E020103" // 评论内容过短
	ErrCommentSpamDetected = "E020104" // 检测到垃圾评论
	ErrAuthorNameRequired  = "E020105" // 评论者姓名必填
	ErrAuthorEmailRequired = "E020106" // 评论者邮箱必填
	ErrInvalidAuthorEmail  = "E020107" // 评论者邮箱格式错误

	// 搜索相关错误
	ErrSearchTimeout            = "E020201" // 搜索超时
	ErrSearchQueryEmpty         = "E020202" // 搜索关键词为空
	ErrSearchQueryTooShort      = "E020203" // 搜索关键词过短
	ErrSearchQueryTooLong       = "E020204" // 搜索关键词过长
	ErrSearchServiceUnavailable = "E020205" // 搜索服务不可用

	// 订阅相关错误
	ErrSubscriptionNotFound = "E020301" // 订阅不存在
	ErrAlreadySubscribed    = "E020302" // 已经订阅
	ErrInvalidEmailAddress  = "E020303" // 邮箱地址无效

	// 访问统计错误
	ErrViewCountFailed       = "E020401" // 浏览计数失败
	ErrStatisticsUnavailable = "E020402" // 统计服务不可用
)

// ====================
// 错误码映射
// ====================

// ErrorCodeToHTTPStatus 错误码到HTTP状态码的映射
var ErrorCodeToHTTPStatus = map[string]int{
	// 通用错误 5xx
	ErrInternalServer:  500,
	ErrDatabaseError:   500,
	ErrRedisError:      500,
	ErrFileSystemError: 500,
	ErrNetworkError:    500,

	// 请求错误 4xx
	ErrInvalidParams: 400,
	ErrInvalidJSON:   400,
	ErrMissingParams: 400,
	ErrParamTooLong:  400,
	ErrParamTooShort: 400,
	ErrInvalidFormat: 400,
	ErrInvalidEmail:  400,
	ErrInvalidURL:    400,

	// 认证错误 401/403
	ErrUnauthorized:     401,
	ErrTokenExpired:     401,
	ErrTokenInvalid:     401,
	ErrTokenBlacklisted: 401,
	ErrForbidden:        403,

	// 资源错误 404/409/429
	ErrNotFound:        404,
	ErrConflict:        409,
	ErrGone:            410,
	ErrLocked:          423,
	ErrRateLimit:       429,
	ErrTooManyRequests: 429,

	// Admin API错误
	ErrUserNotFound:    404,
	ErrInvalidPassword: 401,
	ErrUserLocked:      423,
	ErrUsernameExists:  409,
	ErrEmailExists:     409,
	ErrPostNotFound:    404,
	ErrPostSlugExists:  409,
	ErrCommentNotFound: 404,
	ErrMediaNotFound:   404,
	ErrFileTooLarge:    413,

	// Public API错误
	ErrPostNotPublished:  404,
	ErrPostPrivate:       403,
	ErrPostMembersOnly:   403,
	ErrCommentNotAllowed: 403,
	ErrSearchTimeout:     408,
}

// GetHTTPStatusCode 根据错误码获取HTTP状态码
func GetHTTPStatusCode(errorCode string) int {
	if status, exists := ErrorCodeToHTTPStatus[errorCode]; exists {
		return status
	}
	return 500 // 默认返回500
}

// IsClientError 判断是否为客户端错误（4xx）
func IsClientError(errorCode string) bool {
	status := GetHTTPStatusCode(errorCode)
	return status >= 400 && status < 500
}

// IsServerError 判断是否为服务器错误（5xx）
func IsServerError(errorCode string) bool {
	status := GetHTTPStatusCode(errorCode)
	return status >= 500
}

// IsAuthError 判断是否为认证错误
func IsAuthError(errorCode string) bool {
	authErrors := []string{
		ErrUnauthorized,
		ErrTokenExpired,
		ErrTokenInvalid,
		ErrTokenBlacklisted,
		ErrInvalidPassword,
	}

	for _, authError := range authErrors {
		if errorCode == authError {
			return true
		}
	}
	return false
}

// IsPermissionError 判断是否为权限错误
func IsPermissionError(errorCode string) bool {
	permissionErrors := []string{
		ErrForbidden,
		ErrInsufficientPermission,
		ErrPostPrivate,
		ErrPostMembersOnly,
		ErrCommentNotAllowed,
	}

	for _, permError := range permissionErrors {
		if errorCode == permError {
			return true
		}
	}
	return false
}
