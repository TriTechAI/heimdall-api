package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/heimdall-api/common/constants"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTagLogic {
	return &UpdateTagLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTagLogic) UpdateTag(req *types.TagUpdateRequest) (resp *types.TagUpdateResponse, err error) {
	// 输入验证
	if err := l.validateUpdateRequest(req); err != nil {
		l.Errorf("标签更新请求验证失败: %v", err)
		return nil, err
	}

	// 检查标签是否存在
	existingTag, err := l.svcCtx.TagDAO.GetByID(l.ctx, req.ID)
	if err != nil {
		l.Errorf("获取标签失败: %v", err)
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("标签不存在")
		}
		return nil, fmt.Errorf("获取标签失败: %v", err)
	}

	// 构建更新数据
	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = strings.TrimSpace(req.Name)
	}

	if req.Slug != "" {
		// 如果要更新slug，需要检查是否与其他标签冲突
		newSlug := strings.TrimSpace(req.Slug)
		if newSlug != existingTag.Slug {
			conflictTag, err := l.svcCtx.TagDAO.GetBySlug(l.ctx, newSlug)
			if err == nil && conflictTag != nil && conflictTag.ID != existingTag.ID {
				return nil, fmt.Errorf("标签标识符 '%s' 已存在，请使用其他标识符", newSlug)
			}
			updates["slug"] = newSlug
		}
	}

	if req.Description != "" {
		updates["description"] = strings.TrimSpace(req.Description)
	}

	if req.Color != "" {
		updates["color"] = strings.TrimSpace(req.Color)
	}

	if req.FeaturedImage != "" {
		updates["featuredImage"] = strings.TrimSpace(req.FeaturedImage)
	}

	if req.MetaTitle != "" {
		updates["metaTitle"] = strings.TrimSpace(req.MetaTitle)
	}

	if req.MetaDescription != "" {
		updates["metaDescription"] = strings.TrimSpace(req.MetaDescription)
	}

	if req.Visibility != "" {
		updates["visibility"] = req.Visibility
	}

	// 如果没有任何更新字段，返回错误
	if len(updates) == 0 {
		return nil, fmt.Errorf("未提供任何更新字段")
	}

	// 执行更新
	if err := l.svcCtx.TagDAO.Update(l.ctx, req.ID, updates); err != nil {
		l.Errorf("更新标签失败: %v", err)
		return nil, fmt.Errorf("更新标签失败: %v", err)
	}

	// 获取更新后的标签信息
	updatedTag, err := l.svcCtx.TagDAO.GetByID(l.ctx, req.ID)
	if err != nil {
		l.Errorf("获取更新后的标签失败: %v", err)
		return nil, fmt.Errorf("获取更新后的标签失败: %v", err)
	}

	l.Infof("标签更新成功: %s (ID: %s)", updatedTag.Name, updatedTag.ID.Hex())

	// 构建响应
	resp = &types.TagUpdateResponse{
		Code:      200,
		Message:   "标签更新成功",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: types.TagDetailInfo{
			ID:              updatedTag.ID.Hex(),
			Name:            updatedTag.Name,
			Slug:            updatedTag.Slug,
			Description:     updatedTag.Description,
			Color:           updatedTag.Color,
			FeaturedImage:   updatedTag.FeaturedImage,
			MetaTitle:       updatedTag.MetaTitle,
			MetaDescription: updatedTag.MetaDescription,
			PostCount:       updatedTag.PostCount,
			Visibility:      updatedTag.Visibility,
			CreatedAt:       updatedTag.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       updatedTag.UpdatedAt.Format(time.RFC3339),
		},
	}

	return resp, nil
}

// validateUpdateRequest 验证更新请求
func (l *UpdateTagLogic) validateUpdateRequest(req *types.TagUpdateRequest) error {
	if req == nil {
		return fmt.Errorf("请求不能为空")
	}

	if strings.TrimSpace(req.ID) == "" {
		return fmt.Errorf("标签ID不能为空")
	}

	// 标签名称验证
	if req.Name != "" {
		name := strings.TrimSpace(req.Name)
		if len(name) == 0 {
			return fmt.Errorf("标签名称不能为空")
		}
		if len(name) > constants.TagNameMaxLength {
			return fmt.Errorf("标签名称不能超过 %d 个字符", constants.TagNameMaxLength)
		}
	}

	// Slug验证
	if req.Slug != "" {
		slug := strings.TrimSpace(req.Slug)
		if len(slug) == 0 {
			return fmt.Errorf("标签标识符不能为空")
		}
		if len(slug) > constants.TagSlugMaxLength {
			return fmt.Errorf("标签标识符不能超过 %d 个字符", constants.TagSlugMaxLength)
		}
	}

	// 可见性验证
	if req.Visibility != "" && !constants.IsValidTagVisibility(req.Visibility) {
		return fmt.Errorf("无效的可见性设置，支持的值: %v", constants.GetAllTagVisibilities())
	}

	// 描述验证
	if len(req.Description) > constants.TagDescriptionMaxLength {
		return fmt.Errorf("标签描述不能超过 %d 个字符", constants.TagDescriptionMaxLength)
	}

	// SEO字段验证
	if len(req.MetaTitle) > constants.TagMetaTitleMaxLength {
		return fmt.Errorf("SEO标题不能超过 %d 个字符", constants.TagMetaTitleMaxLength)
	}

	if len(req.MetaDescription) > constants.TagMetaDescMaxLength {
		return fmt.Errorf("SEO描述不能超过 %d 个字符", constants.TagMetaDescMaxLength)
	}

	// 特色图片URL验证
	if len(req.FeaturedImage) > constants.TagFeaturedImageMaxLength {
		return fmt.Errorf("特色图片URL不能超过 %d 个字符", constants.TagFeaturedImageMaxLength)
	}

	return nil
}
