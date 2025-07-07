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

type DeleteTagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteTagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTagLogic {
	return &DeleteTagLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTagLogic) DeleteTag(req *types.TagDeleteRequest) (resp *types.TagDeleteResponse, err error) {
	// 输入验证
	if err := l.validateDeleteRequest(req); err != nil {
		l.Errorf("标签删除请求验证失败: %v", err)
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

	// 检查标签是否关联了文章
	if existingTag.PostCount > 0 {
		l.Errorf("尝试删除关联了 %d 篇文章的标签: %s (ID: %s)", existingTag.PostCount, existingTag.Name, existingTag.ID.Hex())
		return nil, fmt.Errorf("无法删除标签 '%s'，该标签关联了 %d 篇文章。请先移除相关文章的标签关联后再删除", existingTag.Name, existingTag.PostCount)
	}

	// 执行软删除
	if err := l.svcCtx.TagDAO.Delete(l.ctx, req.ID); err != nil {
		l.Errorf("删除标签失败: %v", err)
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("标签不存在或已被删除")
		}
		return nil, fmt.Errorf("删除标签失败: %v", err)
	}

	l.Infof("标签删除成功: %s (ID: %s)", existingTag.Name, existingTag.ID.Hex())

	// 构建响应
	resp = &types.TagDeleteResponse{
		Code:      200,
		Message:   "标签删除成功",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return resp, nil
}

// validateDeleteRequest 验证删除请求
func (l *DeleteTagLogic) validateDeleteRequest(req *types.TagDeleteRequest) error {
	if req == nil {
		return fmt.Errorf("请求不能为空")
	}

	if strings.TrimSpace(req.ID) == "" {
		return fmt.Errorf("标签ID不能为空")
	}

	return nil
}
