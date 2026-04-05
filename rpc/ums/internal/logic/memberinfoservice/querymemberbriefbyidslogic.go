package memberinfoservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/pkg/operatefunnel"
	"github.com/feihua/zero-admin/rpc/ums/gen/query"
	"github.com/feihua/zero-admin/rpc/ums/internal/svc"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMemberBriefByIdsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryMemberBriefByIdsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMemberBriefByIdsLogic {
	return &QueryMemberBriefByIdsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryMemberBriefByIdsLogic) QueryMemberBriefByIds(in *umsclient.QueryMemberBriefByIdsReq) (*umsclient.QueryMemberBriefByIdsResp, error) {
	memberIDs := uniquePositiveInt64s(in.MemberIds)
	if len(memberIDs) == 0 {
		return &umsclient.QueryMemberBriefByIdsResp{List: []*umsclient.MemberBriefData{}}, nil
	}

	memberInfo := query.UmsMemberInfo
	rows, err := memberInfo.WithContext(l.ctx).
		Where(memberInfo.MemberID.In(memberIDs...), memberInfo.IsDeleted.Eq(0)).
		Find()
	if err != nil {
		logc.Errorf(l.ctx, "批量查询会员简要信息失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("批量查询会员简要信息失败")
	}

	list := make([]*umsclient.MemberBriefData, 0, len(rows))
	for _, item := range rows {
		if item == nil {
			continue
		}
		list = append(list, &umsclient.MemberBriefData{
			MemberId:       item.MemberID,
			NicknameMasked: operatefunnel.MaskRepeatPurchaseNickname(item.Nickname),
			MobileMasked:   operatefunnel.MaskRepeatPurchaseMobile(item.Mobile),
			LevelId:        item.LevelID,
		})
	}

	return &umsclient.QueryMemberBriefByIdsResp{List: list}, nil
}

func uniquePositiveInt64s(values []int64) []int64 {
	result := make([]int64, 0, len(values))
	seen := make(map[int64]struct{}, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
