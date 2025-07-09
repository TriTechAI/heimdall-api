package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/model"

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
	existingTag, err := l.getExistingTag(req.ID)
	if err != nil {
		return nil, err
	}

	// 构建更新数据
	updates, err := l.buildUpdateData(req, existingTag)
	if err != nil {
		return nil, err
	}

	// 执行更新
	if err := l.performUpdate(req.ID, updates); err != nil {
		return nil, err
	}

	// 获取更新后的标签并构建响应
	return l.buildUpdateResponse(req.ID)
}

// getExistingTag 获取现有标签
func (l *UpdateTagLogic) getExistingTag(tagID string) (*model.TagModel, error) {
	existingTag, err := l.svcCtx.TagDAO.GetByID(l.ctx, tagID)
	if err != nil {
		l.Errorf("获取标签失败: %v", err)
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("标签不存在")
		}
		return nil, fmt.Errorf("获取标签失败: %v", err)
	}
	return existingTag, nil
}

// buildUpdateData 构建更新数据
func (l *UpdateTagLogic) buildUpdateData(req *types.TagUpdateRequest, existingTag *model.TagModel) (map[string]interface{}, error) {
	updates := make(map[string]interface{})

	// 处理各个字段的更新
	if req.Name != "" {
		updates["name"] = strings.TrimSpace(req.Name)
	}

	// 处理slug更新
	if req.Slug != "" {
		if err := l.handleSlugUpdate(req.Slug, existingTag, updates); err != nil {
			return nil, err
		}
	}

	// 处理其他字段
	l.addStringUpdate(updates, "description", req.Description)
	l.addStringUpdate(updates, "color", req.Color)
	l.addStringUpdate(updates, "featuredImage", req.FeaturedImage)
	l.addStringUpdate(updates, "metaTitle", req.MetaTitle)
	l.addStringUpdate(updates, "metaDescription", req.MetaDescription)

	if req.Visibility != "" {
		updates["visibility"] = req.Visibility
	}

	// 检查是否有更新
	if len(updates) == 0 {
		return nil, fmt.Errorf("未提供任何更新字段")
	}

	return updates, nil
}

// handleSlugUpdate 处理slug更新，检查冲突
func (l *UpdateTagLogic) handleSlugUpdate(newSlug string, existingTag *model.TagModel, updates map[string]interface{}) error {
	newSlug = strings.TrimSpace(newSlug)
	if newSlug != existingTag.Slug {
		conflictTag, err := l.svcCtx.TagDAO.GetBySlug(l.ctx, newSlug)
		if err == nil && conflictTag != nil && conflictTag.ID != existingTag.ID {
			return fmt.Errorf("标签标识符 '%s' 已存在，请使用其他标识符", newSlug)
		}
		updates["slug"] = newSlug
	}
	return nil
}

// addStringUpdate 添加字符串类型的更新
func (l *UpdateTagLogic) addStringUpdate(updates map[string]interface{}, field, value string) {
	if value != "" {
		updates[field] = strings.TrimSpace(value)
	}
}

// performUpdate 执行更新操作
func (l *UpdateTagLogic) performUpdate(tagID string, updates map[string]interface{}) error {
	if err := l.svcCtx.TagDAO.Update(l.ctx, tagID, updates); err != nil {
		l.Errorf("更新标签失败: %v", err)
		return fmt.Errorf("更新标签失败: %v", err)
	}
	return nil
}

// buildUpdateResponse 构建更新响应
func (l *UpdateTagLogic) buildUpdateResponse(tagID string) (*types.TagUpdateResponse, error) {
	// 获取更新后的标签信息
	updatedTag, err := l.svcCtx.TagDAO.GetByID(l.ctx, tagID)
	if err != nil {
		l.Errorf("获取更新后的标签失败: %v", err)
		return nil, fmt.Errorf("获取更新后的标签失败: %v", err)
	}

	l.Infof("标签更新成功: %s (ID: %s)", updatedTag.Name, updatedTag.ID.Hex())

	// 构建响应
	return &types.TagUpdateResponse{
		Code:      200,
		Message:   "标签更新成功",
		Timestamp: time.Now().Format(constants.DefaultTimeFormat),
		Data:      l.buildTagDetailInfo(updatedTag),
	}, nil
}

// buildTagDetailInfo 构建标签详情信息
func (l *UpdateTagLogic) buildTagDetailInfo(tag *model.TagModel) types.TagDetailInfo {
	return types.TagDetailInfo{
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
		CreatedAt:       tag.CreatedAt.Format(constants.DefaultTimeFormat),
		UpdatedAt:       tag.UpdatedAt.Format(constants.DefaultTimeFormat),
	}
}

// validateUpdateRequest 验证更新请求
func (l *UpdateTagLogic) validateUpdateRequest(req *types.TagUpdateRequest) error {
	if req == nil {
		return fmt.Errorf("请求不能为空")
	}

	if strings.TrimSpace(req.ID) == "" {
		return fmt.Errorf("标签ID不能为空")
	}

	// 验证各个字段
	if err := l.validateTagName(req.Name); err != nil {
		return err
	}

	if err := l.validateTagSlug(req.Slug); err != nil {
		return err
	}

	if err := l.validateTagMetaData(req); err != nil {
		return err
	}

	if err := l.validateTagVisibility(req.Visibility); err != nil {
		return err
	}

	return nil
}

// validateTagName 验证标签名称
func (l *UpdateTagLogic) validateTagName(name string) error {
	if name != "" {
		name = strings.TrimSpace(name)
		if len(name) == 0 {
			return fmt.Errorf("标签名称不能为空")
		}
		if len(name) > constants.TagNameMaxLength {
			return fmt.Errorf("标签名称不能超过 %d 个字符", constants.TagNameMaxLength)
		}
	}
	return nil
}

// validateTagSlug 验证标签标识符
func (l *UpdateTagLogic) validateTagSlug(slug string) error {
	if slug != "" {
		slug = strings.TrimSpace(slug)
		if len(slug) == 0 {
			return fmt.Errorf("标签标识符不能为空")
		}
		if len(slug) > constants.TagSlugMaxLength {
			return fmt.Errorf("标签标识符不能超过 %d 个字符", constants.TagSlugMaxLength)
		}
	}
	return nil
}

// validateTagMetaData 验证标签元数据
func (l *UpdateTagLogic) validateTagMetaData(req *types.TagUpdateRequest) error {
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

// validateTagVisibility 验证标签可见性
func (l *UpdateTagLogic) validateTagVisibility(visibility string) error {
	if visibility != "" && !constants.IsValidTagVisibility(visibility) {
		return fmt.Errorf("无效的可见性设置，支持的值: %v", constants.GetAllTagVisibilities())
	}
	return nil
}
