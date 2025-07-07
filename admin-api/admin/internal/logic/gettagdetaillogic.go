package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTagDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTagDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTagDetailLogic {
	return &GetTagDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTagDetailLogic) GetTagDetail(req *types.TagDetailRequest) (resp *types.TagDetailResponse, err error) {
	// 输入验证
	if err := l.validateDetailRequest(req); err != nil {
		l.Errorf("标签详情请求验证失败: %v", err)
		return nil, err
	}

	// 获取标签详情
	tag, err := l.svcCtx.TagDAO.GetByID(l.ctx, req.ID)
	if err != nil {
		l.Errorf("获取标签详情失败: %v", err)
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("标签不存在")
		}
		return nil, fmt.Errorf("获取标签详情失败: %v", err)
	}

	// 构建响应
	resp = &types.TagDetailResponse{
		Code:      200,
		Message:   "获取标签详情成功",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: types.TagDetailInfo{
			ID:              tag.ID.Hex(),
			Name:            tag.Name,
			Slug:            tag.Slug,
			Description:     tag.Description,
			Color:           tag.Color,
			FeaturedImage:   tag.FeaturedImage,
			MetaTitle:       tag.MetaTitle,
			MetaDescription: tag.MetaDescription,
			PostCount:       tag.PostCount,
			Visibility:      tag.Visibility,
			CreatedAt:       tag.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       tag.UpdatedAt.Format(time.RFC3339),
		},
	}

	l.Infof("获取标签详情成功: %s (ID: %s)", tag.Name, tag.ID.Hex())
	return resp, nil
}

// validateDetailRequest 验证详情请求
func (l *GetTagDetailLogic) validateDetailRequest(req *types.TagDetailRequest) error {
	if req == nil {
		return fmt.Errorf("请求不能为空")
	}

	if strings.TrimSpace(req.ID) == "" {
		return fmt.Errorf("标签ID不能为空")
	}

	return nil
}
