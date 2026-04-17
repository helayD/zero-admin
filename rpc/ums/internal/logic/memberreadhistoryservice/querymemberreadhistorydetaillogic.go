package memberreadhistoryservicelogic

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

type QueryMemberReadHistoryDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryMemberReadHistoryDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMemberReadHistoryDetailLogic {
	return &QueryMemberReadHistoryDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询用户商品浏览历史记录详情
func (l *QueryMemberReadHistoryDetailLogic) QueryMemberReadHistoryDetail(in *umsclient.QueryMemberReadHistoryDetailReq) (*umsclient.QueryMemberReadHistoryDetailResp, error) {
	item, err := l.svcCtx.MemberBrowseRecordModel.FindOne(l.ctx, in.Id)
	switch {
	case errors.Is(err, model.ErrNotFound):
		logc.Errorf(l.ctx, "会员浏览记录不存在, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("会员浏览记录不存在")
	case errors.Is(err, model.ErrInvalidObjectId):
		return nil, errors.New("会员浏览记录不存在")
	case err != nil:
		logc.Errorf(l.ctx, "查询会员浏览记录异常, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("查询会员浏览记录异常")
	}

	return &umsclient.QueryMemberReadHistoryDetailResp{
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
