package logic

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/heimdall-api/common/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取用户列表
func NewGetUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserListLogic {
	return &GetUserListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserListLogic) GetUserList(req *types.UserListRequest) (resp *types.UserListResponse, err error) {
	// 1. 参数验证
	if err := l.validateRequest(req); err != nil {
		return nil, err
	}

	// 2. 构建查询过滤条件
	filter := l.buildFilter(req)

	// 3. 查询用户列表
	users, total, err := l.svcCtx.UserDAO.List(l.ctx, filter, req.Page, req.Limit)
	if err != nil {
		l.Logger.Errorf("查询用户列表失败: %v", err)
		return nil, errors.New("系统错误，请稍后重试")
	}

	// 4. 构建分页信息
	pagination := l.buildPagination(req.Page, req.Limit, total)

	// 5. 转换用户信息
	userList := l.convertUsersToUserInfo(users)

	// 6. 构造响应
	resp = &types.UserListResponse{
		Code:      200,
		Message:   "获取用户列表成功",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: types.UserListData{
			List:       userList,
			Pagination: pagination,
		},
	}

	return resp, nil
}

// validateRequest 验证请求参数
func (l *GetUserListLogic) validateRequest(req *types.UserListRequest) error {
	if req == nil {
		return errors.New("请求参数不能为空")
	}

	// 参数范围验证（goctl已经处理了基本验证，这里做补充验证）
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	return nil
}

// buildFilter 构建查询过滤条件
func (l *GetUserListLogic) buildFilter(req *types.UserListRequest) map[string]interface{} {
	filter := make(map[string]interface{})

	// 角色过滤
	if req.Role != "" {
		filter["role"] = req.Role
	}

	// 状态过滤
	if req.Status != "" {
		filter["status"] = req.Status
	}

	// 关键词搜索
	if req.Keyword != "" {
		filter["keyword"] = req.Keyword
	}

	// 排序设置
	filter["sortBy"] = req.SortBy
	filter["sortDesc"] = req.SortDesc

	return filter
}

// buildPagination 构建分页信息
func (l *GetUserListLogic) buildPagination(page, limit int, total int64) types.PaginationInfo {
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	return types.PaginationInfo{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// convertUsersToUserInfo 转换用户模型为响应格式
func (l *GetUserListLogic) convertUsersToUserInfo(users []*model.User) []types.UserInfo {
	if users == nil {
		return []types.UserInfo{}
	}

	userList := make([]types.UserInfo, 0, len(users))
	for _, user := range users {
		userInfo := l.buildUserInfo(user)
		userList = append(userList, userInfo)
	}

	return userList
}

// buildUserInfo 构建用户信息响应（复用ProfileLogic的逻辑）
func (l *GetUserListLogic) buildUserInfo(user *model.User) types.UserInfo {
	userInfo := types.UserInfo{
		ID:          user.ID.Hex(),
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        user.Role,
		Status:      user.Status,
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
	}

	// 设置可选字段
	if user.ProfileImage != "" {
		userInfo.ProfileImage = user.ProfileImage
	}
	if user.Bio != "" {
		userInfo.Bio = user.Bio
	}
	if user.Location != "" {
		userInfo.Location = user.Location
	}
	if user.Website != "" {
		userInfo.Website = user.Website
	}
	if user.Twitter != "" {
		userInfo.Twitter = user.Twitter
	}
	if user.Facebook != "" {
		userInfo.Facebook = user.Facebook
	}
	if user.LastLoginAt != nil {
		userInfo.LastLoginAt = user.LastLoginAt.Format(time.RFC3339)
	}

	return userInfo
}
