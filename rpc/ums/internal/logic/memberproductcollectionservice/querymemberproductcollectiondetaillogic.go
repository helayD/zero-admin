package memberproductcollectionservicelogic

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

type QueryMemberProductCollectionDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryMemberProductCollectionDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMemberProductCollectionDetailLogic {
	return &QueryMemberProductCollectionDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询用户收藏的商品详情
func (l *QueryMemberProductCollectionDetailLogic) QueryMemberProductCollectionDetail(in *umsclient.QueryMemberProductCollectionDetailReq) (*umsclient.QueryMemberProductCollectionDetailResp, error) {
	item, err := l.svcCtx.MemberProductCollectionModel.FindOne(l.ctx, in.Id)
	switch {
	case errors.Is(err, model.ErrNotFound):
		logc.Errorf(l.ctx, "会员商品收藏不存在, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("会员商品收藏不存在")
	case errors.Is(err, model.ErrInvalidObjectId):
		return nil, errors.New("会员商品收藏不存在")
	case err != nil:
		logc.Errorf(l.ctx, "查询会员商品收藏异常, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("查询会员商品收藏异常")
	}

	return &umsclient.QueryMemberProductCollectionDetailResp{
		Id:              item.ID.Hex(),
		MemberId:        item.MemberId,
		MemberNickName:  item.MemberNickName,
		MemberIcon:      item.MemberIcon,
		ProductId:       item.ProductId,
		ProductName:     item.ProductName,
		ProductPic:      item.ProductPic,
		ProductSubTitle: item.ProductSubTitle,
		ProductPrice:    item.ProductPrice,
		CreateTime:      time_util.TimeToStr(item.CreateAt),
	}, nil
}
