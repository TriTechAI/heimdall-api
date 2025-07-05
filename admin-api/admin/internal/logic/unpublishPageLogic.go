package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/model"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UnpublishPageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 取消发布页面
func NewUnpublishPageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnpublishPageLogic {
	return &UnpublishPageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnpublishPageLogic) UnpublishPage(req *types.PageUnpublishRequest) (resp *types.PageUnpublishResponse, err error) {
	// 1. 验证页面ID
	if err := l.validatePageID(req.ID); err != nil {
		return nil, err
	}

	// 2. 获取当前用户ID
	userID, err := l.getCurrentUserID()
	if err != nil {
		return nil, err
	}

	// 3. 获取页面信息
	page, err := l.svcCtx.PageDAO.GetByID(l.ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("获取页面信息失败: %w", err)
	}

	// 4. 检查权限
	if err := l.checkPermission(page, userID); err != nil {
		return nil, err
	}

	// 5. 验证发布状态
	if err := l.validateUnpublishStatus(page); err != nil {
		return nil, err
	}

	// 6. 执行取消发布操作
	if err := l.svcCtx.PageDAO.Unpublish(l.ctx, req.ID); err != nil {
		return nil, fmt.Errorf("取消发布页面失败: %w", err)
	}

	// 7. 构建响应
	return l.buildUnpublishResponse(req.ID)
}

// validatePageID 验证页面ID格式
func (l *UnpublishPageLogic) validatePageID(id string) error {
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		return fmt.Errorf("无效的页面ID格式")
	}
	return nil
}

// getCurrentUserID 获取当前用户ID
func (l *UnpublishPageLogic) getCurrentUserID() (string, error) {
	userID, ok := l.ctx.Value("userId").(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("用户认证失败")
	}
	return userID, nil
}

// checkPermission 检查用户权限
func (l *UnpublishPageLogic) checkPermission(page *model.Page, userID string) error {
	// 验证用户是否存在
	user, err := l.svcCtx.UserDAO.GetByID(l.ctx, userID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}
	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 检查是否为页面作者
	if page.AuthorID.Hex() != userID {
		return fmt.Errorf("无权限取消发布此页面")
	}

	return nil
}

// validateUnpublishStatus 验证取消发布状态
func (l *UnpublishPageLogic) validateUnpublishStatus(page *model.Page) error {
	if page.Status != constants.PostStatusPublished {
		return fmt.Errorf("页面未发布")
	}
	return nil
}

// buildUnpublishResponse 构建取消发布响应
func (l *UnpublishPageLogic) buildUnpublishResponse(pageID string) (*types.PageUnpublishResponse, error) {
	// 获取取消发布后的页面信息
	page, err := l.svcCtx.PageDAO.GetByID(l.ctx, pageID)
	if err != nil {
		return nil, fmt.Errorf("获取取消发布后页面信息失败: %w", err)
	}

	// 获取作者信息
	author, err := l.svcCtx.UserDAO.GetByID(l.ctx, page.AuthorID.Hex())
	if err != nil {
		return nil, fmt.Errorf("获取作者信息失败: %w", err)
	}

	// 构建页面详情数据
	data := l.buildPageDetailData(page, author)

	return &types.PageUnpublishResponse{
		Code:      200,
		Message:   "页面取消发布成功",
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// buildPageDetailData 构建页面详情数据
func (l *UnpublishPageLogic) buildPageDetailData(page *model.Page, author *model.User) types.PageDetailData {
	// 构建作者信息
	var authorInfo types.AuthorInfo
	if author != nil {
		authorInfo = types.AuthorInfo{
			ID:           author.ID.Hex(),
			Username:     author.Username,
			DisplayName:  author.DisplayName,
			ProfileImage: author.ProfileImage,
			Bio:          author.Bio,
		}
	}

	// 格式化发布时间
	var publishedAt string
	if page.PublishedAt != nil {
		publishedAt = page.PublishedAt.Format(time.RFC3339)
	}

	return types.PageDetailData{
		ID:              page.ID.Hex(),
		Title:           page.Title,
		Slug:            page.Slug,
		Content:         page.Content,
		HTML:            page.HTML,
		Author:          authorInfo,
		Status:          page.Status,
		Template:        page.Template,
		MetaTitle:       page.MetaTitle,
		MetaDescription: page.MetaDescription,
		FeaturedImage:   page.FeaturedImage,
		CanonicalURL:    page.CanonicalURL,
		PublishedAt:     publishedAt,
		CreatedAt:       page.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       page.UpdatedAt.Format(time.RFC3339),
	}
}
