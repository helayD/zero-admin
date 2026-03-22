package subjectproductrelationservicelogic

import (
	"context"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/cms/gen/model"
	logiccommon "github.com/feihua/zero-admin/rpc/cms/internal/logic/common"

	"github.com/feihua/zero-admin/rpc/cms/gen/query"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/feihua/zero-admin/rpc/cms/cmsclient"
	"github.com/feihua/zero-admin/rpc/cms/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// QuerySubjectProductRelationListLogic 查询专题商品关系列表
/*
Author: LiuFeiHua
Date: 2024/6/11 16:42
*/
type QuerySubjectProductRelationListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQuerySubjectProductRelationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QuerySubjectProductRelationListLogic {
	return &QuerySubjectProductRelationListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QuerySubjectProductRelationList 查询专题商品关系列表
func (l *QuerySubjectProductRelationListLogic) QuerySubjectProductRelationList(in *cmsclient.QuerySubjectProductRelationListReq) (*cmsclient.QuerySubjectProductRelationListResp, error) {
	current, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		logc.Errorf(l.ctx, "查询关联专题列表scope非法,参数:%+v,异常:%s", in, err.Error())
		return nil, err
	}

	var ids []int64
	err = pkgscope.ApplyGovernanceScope(
		l.svcCtx.DB.WithContext(l.ctx).Model(&model.CmsSubjectProductRelation{}),
		current,
		"",
	).Select("subject_id").Where(query.CmsSubjectProductRelation.ProductID.Eq(in.ProductId)).Scan(&ids).Error

	if err != nil {
		logc.Errorf(l.ctx, "查询关联专题列表信息失败,参数:%+v,异常:%s", in, err.Error())
		return nil, err
	}

	return &cmsclient.QuerySubjectProductRelationListResp{
		SubjectIds: ids,
	}, nil
}
