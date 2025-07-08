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

type DeletePageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除页面
func NewDeletePageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePageLogic {
	return &DeletePageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeletePageLogic) DeletePage(req *types.PageDeleteRequest) (resp *types.PageDeleteResponse, err error) {
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
	page, err := l.getPageByID(req.ID)
	if err != nil {
		return nil, err
	}

	// 4. 检查页面是否已被删除
	if err := l.validateDeleteStatus(page); err != nil {
		return nil, err
	}

	// 5. 验证用户权限
	if err := l.checkPermission(userID, page.AuthorID.Hex()); err != nil {
		return nil, err
	}

	// 6. 执行软删除
	if err := l.executeDelete(req.ID); err != nil {
		return nil, err
	}

	// 7. 构建删除响应
	return l.buildDeleteResponse(), nil
}

// validatePageID 验证页面ID格式
func (l *DeletePageLogic) validatePageID(id string) error {
	if id == "" {
		return fmt.Errorf("页面ID不能为空")
	}

	if !primitive.IsValidObjectID(id) {
		return fmt.Errorf("无效的页面ID格式")
	}

	return nil
}

// getCurrentUserID 从context中获取当前用户ID
func (l *DeletePageLogic) getCurrentUserID() (string, error) {
	userID, ok := l.ctx.Value("userId").(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("用户认证失败")
	}

	return userID, nil
}

// getPageByID 根据ID获取页面信息
func (l *DeletePageLogic) getPageByID(id string) (*model.Page, error) {
	page, err := l.svcCtx.PageDAO.GetByID(l.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取页面信息失败: %v", err)
	}

	if page == nil {
		return nil, fmt.Errorf("页面不存在")
	}

	return page, nil
}

// validateDeleteStatus 验证页面是否已被删除
func (l *DeletePageLogic) validateDeleteStatus(page *model.Page) error {
	if page.Status == constants.PostStatusArchived {
		return fmt.Errorf("页面已被删除")
	}

	return nil
}

// checkPermission 检查用户权限
func (l *DeletePageLogic) checkPermission(userID, authorID string) error {
	// 获取用户信息以验证权限
	user, err := l.svcCtx.UserDAO.GetByID(l.ctx, userID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %v", err)
	}

	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 检查是否为页面作者
	if userID != authorID {
		return fmt.Errorf("无权限删除此页面")
	}

	return nil
}

// executeDelete 执行软删除操作
func (l *DeletePageLogic) executeDelete(id string) error {
	err := l.svcCtx.PageDAO.Delete(l.ctx, id)
	if err != nil {
		return fmt.Errorf("删除页面失败: %v", err)
	}

	return nil
}

// buildDeleteResponse 构建删除响应
func (l *DeletePageLogic) buildDeleteResponse() *types.PageDeleteResponse {
	return &types.PageDeleteResponse{
		Code:      200,
		Message:   "页面删除成功",
		Timestamp: time.Now().Format(time.RFC3339),
	}
}
