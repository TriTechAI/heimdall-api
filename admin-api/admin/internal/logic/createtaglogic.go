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

type CreateTagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTagLogic {
	return &CreateTagLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTagLogic) CreateTag(req *types.TagCreateRequest) (resp *types.TagCreateResponse, err error) {
	// 输入验证
	if err := l.validateCreateRequest(req); err != nil {
		l.Errorf("标签创建请求验证失败: %v", err)
		return nil, err
	}

	// 创建标签模型
	tag := l.buildTagModel(req)

	// 检查slug合法性
	if err := l.checkSlugAvailability(tag.Slug); err != nil {
		return nil, err
	}

	// 验证标签数据
	if err := tag.Validate(); err != nil {
		l.Errorf("标签数据验证失败: %v", err)
		return nil, fmt.Errorf("标签数据验证失败: %v", err)
	}

	// 创建标签
	if err := l.createTag(tag); err != nil {
		return nil, err
	}

	// 构建响应
	return l.buildCreateResponse(tag), nil
}

// buildTagModel 构建标签模型
func (l *CreateTagLogic) buildTagModel(req *types.TagCreateRequest) *model.TagModel {
	tag := &model.TagModel{
		Name:            strings.TrimSpace(req.Name),
		Description:     strings.TrimSpace(req.Description),
		Color:           strings.TrimSpace(req.Color),
		FeaturedImage:   strings.TrimSpace(req.FeaturedImage),
		MetaTitle:       strings.TrimSpace(req.MetaTitle),
		MetaDescription: strings.TrimSpace(req.MetaDescription),
		Visibility:      req.Visibility,
	}

	// 处理Slug
	if req.Slug != "" {
		tag.Slug = strings.TrimSpace(req.Slug)
	} else {
		tag.GenerateSlugFromName()
	}

	// 设置默认值
	tag.PrepareForCreation()

	return tag
}

// checkSlugAvailability 检查slug是否可用
func (l *CreateTagLogic) checkSlugAvailability(slug string) error {
	existingTag, err := l.svcCtx.TagDAO.GetBySlug(l.ctx, slug)
	if err == nil && existingTag != nil {
		l.Errorf("标签slug已存在: %s", slug)
		return fmt.Errorf("标签标识符 '%s' 已存在，请使用其他标识符", slug)
	}
	return nil
}

// createTag 创建标签
func (l *CreateTagLogic) createTag(tag *model.TagModel) error {
	if err := l.svcCtx.TagDAO.Create(l.ctx, tag); err != nil {
		l.Errorf("创建标签失败: %v", err)
		return fmt.Errorf("创建标签失败: %v", err)
	}
	l.Infof("标签创建成功: %s (ID: %s)", tag.Name, tag.ID.Hex())
	return nil
}

// buildCreateResponse 构建创建响应
func (l *CreateTagLogic) buildCreateResponse(tag *model.TagModel) *types.TagCreateResponse {
	return &types.TagCreateResponse{
		Code:      200,
		Message:   "标签创建成功",
		Timestamp: time.Now().Format(constants.DefaultTimeFormat),
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
			CreatedAt:       tag.CreatedAt.Format(constants.DefaultTimeFormat),
			UpdatedAt:       tag.UpdatedAt.Format(constants.DefaultTimeFormat),
		},
	}
}

// validateCreateRequest 验证创建请求
func (l *CreateTagLogic) validateCreateRequest(req *types.TagCreateRequest) error {
	if req == nil {
		return fmt.Errorf("请求不能为空")
	}

	// 标签名称验证
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("标签名称不能为空")
	}

	if len(strings.TrimSpace(req.Name)) > constants.TagNameMaxLength {
		return fmt.Errorf("标签名称不能超过 %d 个字符", constants.TagNameMaxLength)
	}

	// 可见性验证
	if req.Visibility != "" && !constants.IsValidTagVisibility(req.Visibility) {
		return fmt.Errorf("无效的可见性设置，支持的值: %v", constants.GetAllTagVisibilities())
	}

	// Slug验证（如果提供）
	if req.Slug != "" {
		if len(strings.TrimSpace(req.Slug)) > constants.TagSlugMaxLength {
			return fmt.Errorf("标签标识符不能超过 %d 个字符", constants.TagSlugMaxLength)
		}
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
