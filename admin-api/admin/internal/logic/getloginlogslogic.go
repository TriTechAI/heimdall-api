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

type GetLoginLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取登录日志列表
func NewGetLoginLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLoginLogsLogic {
	return &GetLoginLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLoginLogsLogic) GetLoginLogs(req *types.LoginLogsRequest) (resp *types.LoginLogsResponse, err error) {
	// 1. 参数验证
	if err := l.validateRequest(req); err != nil {
		return nil, err
	}

	// 2. 构建查询过滤条件
	filter, err := l.buildFilter(req)
	if err != nil {
		return nil, err
	}

	// 3. 查询登录日志列表
	logs, total, err := l.svcCtx.LoginLogDAO.List(l.ctx, filter, req.Page, req.Limit)
	if err != nil {
		l.Logger.Errorf("查询登录日志列表失败: %v", err)
		return nil, errors.New("系统错误，请稍后重试")
	}

	// 4. 构建分页信息
	pagination := l.buildPagination(req.Page, req.Limit, total)

	// 5. 转换登录日志信息
	logList := l.convertLogsToLogInfo(logs)

	// 6. 构造响应
	resp = &types.LoginLogsResponse{
		Code:      200,
		Message:   "获取登录日志列表成功",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: types.LoginLogsData{
			List:       logList,
			Pagination: pagination,
		},
	}

	return resp, nil
}

// validateRequest 验证请求参数
func (l *GetLoginLogsLogic) validateRequest(req *types.LoginLogsRequest) error {
	if req == nil {
		return errors.New("请求参数不能为空")
	}

	// 参数范围验证（goctl已经处理了基本验证，这里做补充验证）
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// 验证时间格式
	if req.StartTime != "" {
		if _, err := time.Parse(time.RFC3339, req.StartTime); err != nil {
			return errors.New("开始时间格式错误，请使用RFC3339格式")
		}
	}
	if req.EndTime != "" {
		if _, err := time.Parse(time.RFC3339, req.EndTime); err != nil {
			return errors.New("结束时间格式错误，请使用RFC3339格式")
		}
	}

	// 验证时间范围
	if req.StartTime != "" && req.EndTime != "" {
		startTime, _ := time.Parse(time.RFC3339, req.StartTime)
		endTime, _ := time.Parse(time.RFC3339, req.EndTime)
		if startTime.After(endTime) {
			return errors.New("开始时间不能晚于结束时间")
		}
	}

	return nil
}

// buildFilter 构建查询过滤条件
func (l *GetLoginLogsLogic) buildFilter(req *types.LoginLogsRequest) (map[string]interface{}, error) {
	filter := make(map[string]interface{})

	// 用户ID过滤
	if req.UserID != "" {
		filter["userId"] = req.UserID
	}

	// 用户名过滤（模糊搜索）
	if req.Username != "" {
		filter["username"] = req.Username
	}

	// 状态过滤
	if req.Status != "" {
		filter["status"] = req.Status
	}

	// IP地址过滤
	if req.IPAddress != "" {
		filter["ipAddress"] = req.IPAddress
	}

	// 时间范围过滤
	if req.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			return nil, errors.New("开始时间格式错误")
		}
		filter["startTime"] = startTime
	}

	if req.EndTime != "" {
		endTime, err := time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			return nil, errors.New("结束时间格式错误")
		}
		filter["endTime"] = endTime
	}

	// 地理位置过滤
	if req.Country != "" {
		filter["country"] = req.Country
	}

	// 设备信息过滤
	if req.DeviceType != "" {
		filter["deviceType"] = req.DeviceType
	}

	if req.Browser != "" {
		filter["browser"] = req.Browser
	}

	// 排序设置
	filter["sortBy"] = req.SortBy
	filter["sortDesc"] = req.SortDesc

	return filter, nil
}

// buildPagination 构建分页信息
func (l *GetLoginLogsLogic) buildPagination(page, limit int, total int64) types.PaginationInfo {
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

// convertLogsToLogInfo 转换登录日志模型为响应格式
func (l *GetLoginLogsLogic) convertLogsToLogInfo(logs []*model.LoginLog) []types.LoginLogInfo {
	if logs == nil {
		return []types.LoginLogInfo{}
	}

	logList := make([]types.LoginLogInfo, 0, len(logs))
	for _, log := range logs {
		logInfo := l.buildLoginLogInfo(log)
		logList = append(logList, logInfo)
	}

	return logList
}

// buildLoginLogInfo 构建登录日志信息响应
func (l *GetLoginLogsLogic) buildLoginLogInfo(log *model.LoginLog) types.LoginLogInfo {
	logInfo := types.LoginLogInfo{
		ID:          log.ID.Hex(),
		Username:    log.Username,
		LoginMethod: log.LoginMethod,
		IPAddress:   log.IPAddress,
		UserAgent:   log.UserAgent,
		Status:      log.Status,
		LoginAt:     log.LoginAt.Format(time.RFC3339),
	}

	// 设置可选字段
	if log.UserID != nil {
		logInfo.UserID = log.UserID.Hex()
	}

	if log.FailReason != "" {
		logInfo.FailReason = log.FailReason
	}

	if log.SessionID != "" {
		logInfo.SessionID = log.SessionID
	}

	if log.Country != "" {
		logInfo.Country = log.Country
	}

	if log.Region != "" {
		logInfo.Region = log.Region
	}

	if log.City != "" {
		logInfo.City = log.City
	}

	if log.DeviceType != "" {
		logInfo.DeviceType = log.DeviceType
	}

	if log.Browser != "" {
		logInfo.Browser = log.Browser
	}

	if log.OS != "" {
		logInfo.OS = log.OS
	}

	if log.LogoutAt != nil {
		logInfo.LogoutAt = log.LogoutAt.Format(time.RFC3339)
	}

	if log.Duration != nil {
		logInfo.Duration = *log.Duration
	}

	return logInfo
}
