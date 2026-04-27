package system_config

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// SaveSystemConfigLogic 保存系统配置
type SaveSystemConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSaveSystemConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveSystemConfigLogic {
	return &SaveSystemConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// SaveSystemConfig 保存系统配置
func (l *SaveSystemConfigLogic) SaveSystemConfig(req *types.SaveSystemConfigReq) (*types.SaveSystemConfigResp, error) {
	if err := validateSystemConfig(req); err != nil {
		return nil, err
	}

	operator, err := common.GetUserName(l.ctx)
	if err != nil {
		operator = "system"
	}

	if err = SaveSystemConfig(l.ctx, l.svcCtx, req, operator); err != nil {
		logc.Errorf(l.ctx, "保存系统配置失败,参数:%+v,异常:%s", req, err.Error())
		return nil, err
	}

	return &types.SaveSystemConfigResp{
		Code:    "000000",
		Message: "保存系统配置成功",
	}, nil
}

func validateSystemConfig(req *types.SaveSystemConfigReq) error {
	if req == nil {
		return errors.New("系统配置不能为空")
	}
	if req.OSS.MaxSizeMB <= 0 {
		return errors.New("OSS上传大小上限必须大于0")
	}
	return nil
}
