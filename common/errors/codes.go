package errors

// ====================
// 错误码定义
// ====================

const (
	// 通用错误码 (1000-1999)
	CodeSuccess           = "0"      // 成功
	CodeInternalError     = "1000"   // 内部服务器错误
	CodeInvalidRequest    = "1001"   // 请求参数无效
	CodeUnauthorized      = "1002"   // 未授权
	CodeForbidden         = "1003"   // 禁止访问
	CodeNotFound          = "1004"   // 资源不存在
	CodeMethodNotAllowed  = "1005"   // 方法不允许
	CodeTooManyRequests   = "1006"   // 请求过于频繁
	CodeServiceUnavailable = "1007"  // 服务不可用
	CodeTimeout           = "1008"   // 请求超时
	CodeConflict          = "1009"   // 资源冲突
	CodeValidationFailed  = "1010"   // 数据验证失败

	// 认证相关错误码 (2000-2999)
	CodeAuthFailed        = "2000"   // 认证失败
	CodeTokenInvalid      = "2001"   // Token无效
	CodeTokenExpired      = "2002"   // Token已过期
	CodeLoginFailed       = "2003"   // 登录失败
	CodePasswordIncorrect = "2004"   // 密码错误
	CodeUserNotFound      = "2005"   // 用户不存在
	CodeUserDisabled      = "2006"   // 用户已禁用
	CodeUserLocked        = "2007"   // 用户已锁定
	CodeTooManyAttempts   = "2008"   // 登录尝试次数过多
	CodeSessionExpired    = "2009"   // 会话已过期
	CodePermissionDenied  = "2010"   // 权限不足
	CodeRefreshTokenInvalid = "2011" // 刷新令牌无效

	// 用户管理错误码 (3000-3999)
	CodeUserExists        = "3000"   // 用户已存在
	CodeUsernameExists    = "3001"   // 用户名已存在
	CodeEmailExists       = "3002"   // 邮箱已存在
	CodeUserCreateFailed  = "3003"   // 用户创建失败
	CodeUserUpdateFailed  = "3004"   // 用户更新失败
	CodeUserDeleteFailed  = "3005"   // 用户删除失败
	CodePasswordTooWeak   = "3006"   // 密码强度不足
	CodePasswordSame      = "3007"   // 新密码与旧密码相同
	CodeUserInfoIncomplete = "3008"  // 用户信息不完整
	CodeProfileUpdateFailed = "3009" // 个人资料更新失败

	// 内容管理错误码 (4000-4999)
	CodePostNotFound      = "4000"   // 文章不存在
	CodePostCreateFailed  = "4001"   // 文章创建失败
	CodePostUpdateFailed  = "4002"   // 文章更新失败
	CodePostDeleteFailed  = "4003"   // 文章删除失败
	CodePostPublishFailed = "4004"   // 文章发布失败
	CodeSlugExists        = "4005"   // Slug已存在
	CodeContentTooLong    = "4006"   // 内容过长
	CodeTitleRequired     = "4007"   // 标题必填
	CodeContentRequired   = "4008"   // 内容必填
	CodeInvalidStatus     = "4009"   // 无效状态
	CodePostNotPublished  = "4010"   // 文章未发布

	// 页面管理错误码 (4100-4199)
	CodePageNotFound      = "4100"   // 页面不存在
	CodePageCreateFailed  = "4101"   // 页面创建失败
	CodePageUpdateFailed  = "4102"   // 页面更新失败
	CodePageDeleteFailed  = "4103"   // 页面删除失败
	CodePagePublishFailed = "4104"   // 页面发布失败

	// 评论管理错误码 (4200-4299)
	CodeCommentNotFound     = "4200" // 评论不存在
	CodeCommentCreateFailed = "4201" // 评论创建失败
	CodeCommentUpdateFailed = "4202" // 评论更新失败
	CodeCommentDeleteFailed = "4203" // 评论删除失败
	CodeCommentTooLong      = "4204" // 评论过长
	CodeCommentSpam         = "4205" // 垃圾评论
	CodeCommentPending      = "4206" // 评论待审核

	// 标签管理错误码 (4300-4399)
	CodeTagNotFound      = "4300"    // 标签不存在
	CodeTagCreateFailed  = "4301"    // 标签创建失败
	CodeTagUpdateFailed  = "4302"    // 标签更新失败
	CodeTagDeleteFailed  = "4303"    // 标签删除失败
	CodeTagExists        = "4304"    // 标签已存在
	CodeTooManyTags      = "4305"    // 标签过多

	// 媒体管理错误码 (4400-4499)
	CodeMediaNotFound     = "4400"   // 媒体文件不存在
	CodeMediaUploadFailed = "4401"   // 媒体上传失败
	CodeMediaDeleteFailed = "4402"   // 媒体删除失败
	CodeFileTypeNotAllowed = "4403"  // 文件类型不允许
	CodeFileTooLarge      = "4404"   // 文件过大
	CodeStorageQuotaExceeded = "4405" // 存储配额超限

	// 系统设置错误码 (5000-5999)
	CodeSettingNotFound    = "5000"  // 设置项不存在
	CodeSettingUpdateFailed = "5001" // 设置更新失败
	CodeInvalidSetting     = "5002"  // 无效设置
	CodeSettingReadOnly    = "5003"  // 设置只读

	// 数据库错误码 (6000-6999)
	CodeDatabaseError     = "6000"   // 数据库错误
	CodeConnectionFailed  = "6001"   // 数据库连接失败
	CodeQueryFailed       = "6002"   // 查询失败
	CodeInsertFailed      = "6003"   // 插入失败
	CodeUpdateFailed      = "6004"   // 更新失败
	CodeDeleteFailed      = "6005"   // 删除失败
	CodeTransactionFailed = "6006"   // 事务失败
	CodeIndexError        = "6007"   // 索引错误
	CodeConstraintViolation = "6008" // 约束违反

	// 缓存错误码 (7000-7999)
	CodeCacheError       = "7000"    // 缓存错误
	CodeCacheNotFound    = "7001"    // 缓存不存在
	CodeCacheExpired     = "7002"    // 缓存已过期
	CodeCacheSetFailed   = "7003"    // 缓存设置失败
	CodeCacheGetFailed   = "7004"    // 缓存获取失败
	CodeCacheDeleteFailed = "7005"   // 缓存删除失败

	// 第三方服务错误码 (8000-8999)
	CodeExternalServiceError = "8000" // 外部服务错误
	CodeEmailSendFailed     = "8001"  // 邮件发送失败
	CodeSMSSendFailed       = "8002"  // 短信发送失败
	CodePaymentFailed       = "8003"  // 支付失败
	CodeStorageServiceError = "8004"  // 存储服务错误
	CodeSearchServiceError  = "8005"  // 搜索服务错误

	// 业务逻辑错误码 (9000-9999)
	CodeBusinessLogicError = "9000"  // 业务逻辑错误
	CodeWorkflowError      = "9001"  // 工作流错误
	CodeStateError         = "9002"  // 状态错误
	CodeDependencyError    = "9003"  // 依赖错误
	CodeQuotaExceeded      = "9004"  // 配额超限
	CodeOperationNotAllowed = "9005" // 操作不允许
)

// ====================
// 错误消息映射
// ====================

var ErrorMessages = map[string]string{
	// 通用错误消息
	CodeSuccess:           "操作成功",
	CodeInternalError:     "内部服务器错误",
	CodeInvalidRequest:    "请求参数无效",
	CodeUnauthorized:      "未授权访问",
	CodeForbidden:         "禁止访问",
	CodeNotFound:          "资源不存在",
	CodeMethodNotAllowed:  "请求方法不允许",
	CodeTooManyRequests:   "请求过于频繁，请稍后重试",
	CodeServiceUnavailable: "服务暂时不可用",
	CodeTimeout:           "请求超时",
	CodeConflict:          "资源冲突",
	CodeValidationFailed:  "数据验证失败",

	// 认证相关错误消息
	CodeAuthFailed:        "认证失败",
	CodeTokenInvalid:      "访问令牌无效",
	CodeTokenExpired:      "访问令牌已过期",
	CodeLoginFailed:       "登录失败",
	CodePasswordIncorrect: "用户名或密码错误",
	CodeUserNotFound:      "用户不存在",
	CodeUserDisabled:      "用户账户已被禁用",
	CodeUserLocked:        "用户账户已被锁定",
	CodeTooManyAttempts:   "登录尝试次数过多，请稍后重试",
	CodeSessionExpired:    "会话已过期，请重新登录",
	CodePermissionDenied:  "权限不足",
	CodeRefreshTokenInvalid: "刷新令牌无效",

	// 用户管理错误消息
	CodeUserExists:        "用户已存在",
	CodeUsernameExists:    "用户名已存在",
	CodeEmailExists:       "邮箱地址已存在",
	CodeUserCreateFailed:  "用户创建失败",
	CodeUserUpdateFailed:  "用户信息更新失败",
	CodeUserDeleteFailed:  "用户删除失败",
	CodePasswordTooWeak:   "密码强度不足",
	CodePasswordSame:      "新密码不能与当前密码相同",
	CodeUserInfoIncomplete: "用户信息不完整",
	CodeProfileUpdateFailed: "个人资料更新失败",

	// 内容管理错误消息
	CodePostNotFound:      "文章不存在",
	CodePostCreateFailed:  "文章创建失败",
	CodePostUpdateFailed:  "文章更新失败",
	CodePostDeleteFailed:  "文章删除失败",
	CodePostPublishFailed: "文章发布失败",
	CodeSlugExists:        "URL别名已存在",
	CodeContentTooLong:    "内容长度超出限制",
	CodeTitleRequired:     "文章标题不能为空",
	CodeContentRequired:   "文章内容不能为空",
	CodeInvalidStatus:     "无效的状态值",
	CodePostNotPublished:  "文章尚未发布",

	// 页面管理错误消息
	CodePageNotFound:      "页面不存在",
	CodePageCreateFailed:  "页面创建失败",
	CodePageUpdateFailed:  "页面更新失败",
	CodePageDeleteFailed:  "页面删除失败",
	CodePagePublishFailed: "页面发布失败",

	// 评论管理错误消息
	CodeCommentNotFound:     "评论不存在",
	CodeCommentCreateFailed: "评论创建失败",
	CodeCommentUpdateFailed: "评论更新失败",
	CodeCommentDeleteFailed: "评论删除失败",
	CodeCommentTooLong:      "评论内容过长",
	CodeCommentSpam:         "评论被识别为垃圾内容",
	CodeCommentPending:      "评论正在审核中",

	// 标签管理错误消息
	CodeTagNotFound:      "标签不存在",
	CodeTagCreateFailed:  "标签创建失败",
	CodeTagUpdateFailed:  "标签更新失败",
	CodeTagDeleteFailed:  "标签删除失败",
	CodeTagExists:        "标签已存在",
	CodeTooManyTags:      "标签数量超出限制",

	// 媒体管理错误消息
	CodeMediaNotFound:     "媒体文件不存在",
	CodeMediaUploadFailed: "媒体文件上传失败",
	CodeMediaDeleteFailed: "媒体文件删除失败",
	CodeFileTypeNotAllowed: "不支持的文件类型",
	CodeFileTooLarge:      "文件大小超出限制",
	CodeStorageQuotaExceeded: "存储空间不足",

	// 系统设置错误消息
	CodeSettingNotFound:    "设置项不存在",
	CodeSettingUpdateFailed: "设置更新失败",
	CodeInvalidSetting:     "无效的设置值",
	CodeSettingReadOnly:    "该设置项为只读",

	// 数据库错误消息
	CodeDatabaseError:     "数据库操作失败",
	CodeConnectionFailed:  "数据库连接失败",
	CodeQueryFailed:       "数据查询失败",
	CodeInsertFailed:      "数据插入失败",
	CodeUpdateFailed:      "数据更新失败",
	CodeDeleteFailed:      "数据删除失败",
	CodeTransactionFailed: "事务执行失败",
	CodeIndexError:        "索引操作失败",
	CodeConstraintViolation: "数据约束违反",

	// 缓存错误消息
	CodeCacheError:       "缓存操作失败",
	CodeCacheNotFound:    "缓存数据不存在",
	CodeCacheExpired:     "缓存数据已过期",
	CodeCacheSetFailed:   "缓存设置失败",
	CodeCacheGetFailed:   "缓存获取失败",
	CodeCacheDeleteFailed: "缓存删除失败",

	// 第三方服务错误消息
	CodeExternalServiceError: "外部服务调用失败",
	CodeEmailSendFailed:     "邮件发送失败",
	CodeSMSSendFailed:       "短信发送失败",
	CodePaymentFailed:       "支付处理失败",
	CodeStorageServiceError: "存储服务异常",
	CodeSearchServiceError:  "搜索服务异常",

	// 业务逻辑错误消息
	CodeBusinessLogicError: "业务逻辑错误",
	CodeWorkflowError:      "工作流程错误",
	CodeStateError:         "状态异常",
	CodeDependencyError:    "依赖关系错误",
	CodeQuotaExceeded:      "使用配额已超限",
	CodeOperationNotAllowed: "当前操作不被允许",
}

// GetErrorMessage 获取错误消息
func GetErrorMessage(code string) string {
	if msg, exists := ErrorMessages[code]; exists {
		return msg
	}
	return ErrorMessages[CodeInternalError]
}

// IsValidErrorCode 检查是否为有效的错误码
func IsValidErrorCode(code string) bool {
	_, exists := ErrorMessages[code]
	return exists
}
