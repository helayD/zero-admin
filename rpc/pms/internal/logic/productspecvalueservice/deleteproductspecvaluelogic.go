package productspecvalueservicelogic
import (
	"context"; "errors"; "time"
	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"; "github.com/feihua/zero-admin/rpc/pms/internal/svc"; "github.com/feihua/zero-admin/rpc/pms/pmsclient"; "github.com/zeromicro/go-zero/core/logc"; "github.com/zeromicro/go-zero/core/logx")

type DeleteProductSpecValueLogic struct { ctx context.Context; svcCtx *svc.ServiceContext; logx.Logger }
func NewDeleteProductSpecValueLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProductSpecValueLogic { return &DeleteProductSpecValueLogic{ctx:ctx, svcCtx:svcCtx, Logger:logx.WithContext(ctx)} }
func (l *DeleteProductSpecValueLogic) DeleteProductSpecValue(in *pmsclient.DeleteProductSpecValueReq) (*pmsclient.DeleteProductSpecValueResp, error) { currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.UpdateBy); if err != nil { return nil, err }; if _, err := logiccommon.EnsureSpecValueScope(l.ctx, l.svcCtx.DB, currentScope, in.Ids, "pms.product_spec_value.delete", in.UpdateBy, "", "delete spec value"); err != nil { return nil, err }; if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_spec_value").Where("id IN ?", in.Ids).Updates(map[string]interface{}{"is_deleted":1,"update_by":in.UpdateBy,"update_time":time.Now()}).Error; err != nil { logc.Errorf(l.ctx, "删除商品规格值失败,参数:%+v,异常:%s", in, err.Error()); return nil, errors.New("删除商品规格值失败") }; return &pmsclient.DeleteProductSpecValueResp{Pong:"ok"}, nil }
