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

type DeletePostLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除文章
func NewDeletePostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePostLogic {
	return &DeletePostLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeletePostLogic) DeletePost(req *types.PostDeleteRequest) (resp *types.PostDeleteResponse, err error) {
	// 1. 验证文章ID
	if err := l.validatePostID(req.ID); err != nil {
		return nil, err
	}

	// 2. 获取当前用户ID
	userID, err := l.getCurrentUserID()
	if err != nil {
		return nil, err
	}

	// 3. 获取文章信息
	post, err := l.getPostByID(req.ID)
	if err != nil {
		return nil, err
	}

	// 4. 检查文章是否已被删除
	if err := l.validateDeleteStatus(post); err != nil {
		return nil, err
	}

	// 5. 验证用户权限
	if err := l.checkPermission(userID, post.AuthorID.Hex()); err != nil {
		return nil, err
	}

	// 6. 执行软删除
	if err := l.executeDelete(req.ID); err != nil {
		return nil, err
	}

	// 7. 构建删除响应
	return l.buildDeleteResponse(), nil
}

// validatePostID 验证文章ID格式
func (l *DeletePostLogic) validatePostID(id string) error {
	if id == "" {
		return fmt.Errorf("文章ID不能为空")
	}

	if !primitive.IsValidObjectID(id) {
		return fmt.Errorf("无效的文章ID格式")
	}

	return nil
}

// getCurrentUserID 从context中获取当前用户ID
func (l *DeletePostLogic) getCurrentUserID() (string, error) {
	userID, ok := l.ctx.Value("uid").(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("用户认证失败")
	}

	return userID, nil
}

// getPostByID 根据ID获取文章信息
func (l *DeletePostLogic) getPostByID(id string) (*model.Post, error) {
	post, err := l.svcCtx.PostDAO.GetByID(l.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取文章信息失败: %v", err)
	}

	if post == nil {
		return nil, fmt.Errorf("文章不存在")
	}

	return post, nil
}

// validateDeleteStatus 验证文章是否已被删除
func (l *DeletePostLogic) validateDeleteStatus(post *model.Post) error {
	if post.Status == constants.PostStatusArchived {
		return fmt.Errorf("文章已被删除")
	}

	return nil
}

// checkPermission 检查用户权限
func (l *DeletePostLogic) checkPermission(userID, authorID string) error {
	// 获取用户信息以验证权限
	user, err := l.svcCtx.UserDAO.GetByID(l.ctx, userID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %v", err)
	}

	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 检查是否为文章作者
	if userID != authorID {
		return fmt.Errorf("无权限删除此文章")
	}

	return nil
}

// executeDelete 执行软删除操作
func (l *DeletePostLogic) executeDelete(id string) error {
	err := l.svcCtx.PostDAO.Delete(l.ctx, id)
	if err != nil {
		return fmt.Errorf("删除文章失败: %v", err)
	}

	return nil
}

// buildDeleteResponse 构建删除响应
func (l *DeletePostLogic) buildDeleteResponse() *types.PostDeleteResponse {
	return &types.PostDeleteResponse{
		Code:      200,
		Message:   "文章删除成功",
		Timestamp: time.Now().Format(time.RFC3339),
	}
}
