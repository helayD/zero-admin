package productspecvalueservicelogic
import (
	"context"; "errors"; "strconv"; "time"
	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"; "github.com/feihua/zero-admin/rpc/pms/internal/svc"; "github.com/feihua/zero-admin/rpc/pms/pmsclient"; "github.com/zeromicro/go-zero/core/logc"; "github.com/zeromicro/go-zero/core/logx")

type UpdateProductSpecValueStatusLogic struct { ctx context.Context; svcCtx *svc.ServiceContext; logx.Logger }
func NewUpdateProductSpecValueStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductSpecValueStatusLogic { return &UpdateProductSpecValueStatusLogic{ctx:ctx, svcCtx:svcCtx, Logger:logx.WithContext(ctx)} }
func (l *UpdateProductSpecValueStatusLogic) UpdateProductSpecValueStatus(in *pmsclient.UpdateProductSpecValueStatusReq) (*pmsclient.UpdateProductSpecValueStatusResp, error) { currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.UpdateBy); if err != nil { return nil, err }; if _, err := logiccommon.EnsureSpecValueScope(l.ctx, l.svcCtx.DB, currentScope, in.Ids, "pms.product_spec_value.status", in.UpdateBy, "", "status="+strconv.Itoa(int(in.Status))); err != nil { return nil, err }; if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_spec_value").Where("id IN ?", in.Ids).Updates(map[string]interface{}{"status":in.Status,"update_by":in.UpdateBy,"update_time":time.Now()}).Error; err != nil { logc.Errorf(l.ctx, "更新商品规格值状态失败,参数:%+v,异常:%s", in, err.Error()); return nil, errors.New("更新商品规格值状态失败") }; return &pmsclient.UpdateProductSpecValueStatusResp{Pong:"ok"}, nil }
