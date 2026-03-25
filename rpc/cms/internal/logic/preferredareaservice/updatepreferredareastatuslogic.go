package preferredareaservicelogic

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/feihua/zero-admin/rpc/cms/cmsclient"
	"github.com/feihua/zero-admin/rpc/cms/gen/query"
	logiccommon "github.com/feihua/zero-admin/rpc/cms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/cms/internal/svc"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// UpdatePreferredAreaStatusLogic 更新优选专区
/*
Author: LiuFeiHua
Date: 2025/01/23 15:24:00
*/
type UpdatePreferredAreaStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdatePreferredAreaStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePreferredAreaStatusLogic {
	return &UpdatePreferredAreaStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdatePreferredAreaStatus 更新优选专区状态
func (l *UpdatePreferredAreaStatusLogic) UpdatePreferredAreaStatus(in *cmsclient.UpdatePreferredAreaStatusReq) (*cmsclient.UpdatePreferredAreaStatusResp, error) {
	q := query.CmsPreferredArea
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.UpdateBy)
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsurePreferredAreaScope(l.ctx, l.svcCtx.DB, currentScope, in.Ids, "cms.preferred_area.status", in.UpdateBy, "showStatus="+strconv.Itoa(int(in.ShowStatus))); err != nil {
		return nil, err
	}

	// 发布校验：当目标状态为1（显示/发布）时，对每个优选专区执行发布前置校验
	if in.ShowStatus == 1 {
		for _, id := range in.Ids {
			detail, qErr := q.WithContext(l.ctx).Where(q.ID.Eq(id)).First()
			if qErr != nil {
				logc.Errorf(l.ctx, "查询优选专区失败,id:%d,异常:%s", id, qErr.Error())
				return nil, fmt.Errorf("优选专区(ID:%d)不存在", id)
			}

			var missing []string
			if strings.TrimSpace(detail.Name) == "" {
				missing = append(missing, "专区名称")
			}
			if strings.TrimSpace(detail.Pic) == "" {
				missing = append(missing, "展示图片")
			}
			if len(missing) > 0 {
				return nil, fmt.Errorf("优选专区「%s」发布校验未通过：缺少%s", detail.Name, strings.Join(missing, "、"))
			}
		}
	}

	_, err = q.WithContext(l.ctx).Where(q.ID.In(in.Ids...)).Update(q.ShowStatus, in.ShowStatus)

	if err != nil {
		logc.Errorf(l.ctx, "更新优选专区状态失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("更新优选专区状态失败")
	}

	return &cmsclient.UpdatePreferredAreaStatusResp{}, nil
}
