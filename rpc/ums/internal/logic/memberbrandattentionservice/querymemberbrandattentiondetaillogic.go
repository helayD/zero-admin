package memberbrandattentionservicelogic

import (
	"context"
	"errors"
	"github.com/feihua/zero-admin/pkg/time_util"
	"github.com/feihua/zero-admin/rpc/ums/gen/model"

	"github.com/feihua/zero-admin/rpc/ums/internal/svc"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMemberBrandAttentionDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryMemberBrandAttentionDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMemberBrandAttentionDetailLogic {
	return &QueryMemberBrandAttentionDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询会员关注品牌详情
func (l *QueryMemberBrandAttentionDetailLogic) QueryMemberBrandAttentionDetail(in *umsclient.QueryMemberBrandAttentionDetailReq) (*umsclient.QueryMemberBrandAttentionDetailResp, error) {
	item, err := l.svcCtx.MemberBrandAttentionModel.FindOne(l.ctx, in.Id)
	switch {
	case errors.Is(err, model.ErrNotFound):
		logc.Errorf(l.ctx, "会员关注品牌不存在, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("会员关注品牌不存在")
	case errors.Is(err, model.ErrInvalidObjectId):
		return nil, errors.New("会员关注品牌不存在")
	case err != nil:
		logc.Errorf(l.ctx, "查询会员关注品牌异常, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("查询会员关注品牌异常")
	}

	return &umsclient.QueryMemberBrandAttentionDetailResp{
		Id:             item.ID.Hex(),
		MemberId:       item.MemberId,
		MemberNickName: item.MemberNickName,
		MemberIcon:     item.MemberIcon,
		BrandId:        item.BrandId,
		BrandName:      item.BrandName,
		BrandLogo:      item.BrandLogo,
		BrandCity:      item.BrandCity,
		CreateTime:     time_util.TimeToStr(item.CreateAt),
	}, nil
}
