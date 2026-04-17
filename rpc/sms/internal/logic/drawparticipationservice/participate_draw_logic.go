package drawparticipationservicelogic

import (
	"context"
	"errors"
	"strings"
	"time"

	cardassetservicelogic "github.com/feihua/zero-admin/rpc/sms/internal/logic/cardassetservice"
	logiccommon "github.com/feihua/zero-admin/rpc/sms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type ParticipateDrawLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewParticipateDrawLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ParticipateDrawLogic {
	return &ParticipateDrawLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ParticipateDrawLogic) ParticipateDraw(in *smsclient.ParticipateDrawReq) (*smsclient.ParticipateDrawResp, error) {
	if in.ActivityId <= 0 {
		return nil, errors.New("活动ID不能为空")
	}
	if in.MemberId <= 0 {
		return nil, errors.New("会员ID不能为空")
	}
	if strings.TrimSpace(in.RequestId) == "" {
		return nil, errors.New("请求ID不能为空")
	}
	currentScope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}

	result := &smsclient.ParticipateDrawResp{}
	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		existing, err := loadRecordByRequest(l.ctx, tx, in.ActivityId, in.MemberId, in.RequestId)
		switch {
		case err == nil:
			if strings.TrimSpace(existing.ResultType) == drawResultTypeWon && strings.TrimSpace(existing.ResultStatus) == drawResultStatusWon {
				if _, err = cardassetservicelogic.EnsureCardInstanceByParticipationRecord(l.ctx, tx, existing.ID, "system", existing.TraceID); err != nil {
					return err
				}
			}
			record, detailErr := loadRecordDetailByID(l.ctx, tx, existing.ID)
			if detailErr != nil {
				return detailErr
			}
			result.Record = record
			result.EligibilityCode = firstNonEmpty(existing.FailureCode, drawEligibilityEligible)
			result.EligibilityMessage = firstNonEmpty(existing.FailureReason, drawResultStatusText(existing.ResultStatus))
			result.NextAction = eligibilityNextAction(existing.FailureCode)
			return nil
		case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
			return err
		}

		activity, err := loadActivitySnapshot(l.ctx, tx, currentScope, in.ActivityId, true)
		if err != nil {
			return err
		}
		member, err := loadMemberInfoSnapshot(l.ctx, tx, in.MemberId, true)
		if err != nil {
			return err
		}
		identity, err := loadMemberIdentitySnapshot(l.ctx, tx, in.MemberId)
		if err != nil {
			return err
		}
		totalCount, dailyCount, err := countConsumedRecords(l.ctx, tx, activity.ID, in.MemberId)
		if err != nil {
			return err
		}
		pools, err := loadPoolSnapshots(l.ctx, tx, activity.ID)
		if err != nil {
			return err
		}
		poolIDs := make([]int64, 0, len(pools))
		for _, pool := range pools {
			poolIDs = append(poolIDs, pool.ID)
		}
		poolTemplates, err := loadPoolTemplates(l.ctx, tx, activity.ID, poolIDs, true)
		if err != nil {
			return err
		}
		templateIDs := make([]int64, 0, len(poolTemplates))
		for _, item := range poolTemplates {
			templateIDs = append(templateIDs, item.TemplateID)
		}
		templateMap, err := loadCardTemplates(l.ctx, tx, templateIDs)
		if err != nil {
			return err
		}

		eligibility := buildEligibility(activity, member, identity, totalCount, dailyCount, hasAvailableInventory(poolTemplates), true)
		if eligibility.Code != drawEligibilityEligible {
			recordRow := buildRejectedRecord(activity, in.MemberId, in.RequestId, eligibility)
			if err = createParticipationRecord(l.ctx, tx, recordRow); err != nil {
				return err
			}
			record, detailErr := loadRecordDetailByID(l.ctx, tx, recordRow.ID)
			if detailErr != nil {
				return detailErr
			}
			result.Record = record
			result.EligibilityCode = eligibility.Code
			result.EligibilityMessage = eligibility.Message
			result.NextAction = eligibility.NextAction
			return nil
		}

		before := member.LotteryTimes
		after := before - activity.ConsumeAmount
		updateResult := tx.WithContext(l.ctx).
			Table(member.TableName()).
			Where("member_id = ? AND lottery_times >= ?", in.MemberId, activity.ConsumeAmount).
			Updates(map[string]interface{}{
				"lottery_times": after,
				"update_time":   time.Now(),
			})
		if updateResult.Error != nil {
			return updateResult.Error
		}
		if updateResult.RowsAffected == 0 {
			return errors.New("剩余抽奖次数不足")
		}

		winner := chooseWinner(poolTemplates)
		if winner != nil {
			inventoryResult := tx.WithContext(l.ctx).
				Table(drawPoolTemplateSnapshot{}.TableName()).
				Where("id = ? AND remaining_limit > 0", winner.ID).
				Update("remaining_limit", gorm.Expr("remaining_limit - 1"))
			if inventoryResult.Error != nil {
				return inventoryResult.Error
			}
			if inventoryResult.RowsAffected == 0 {
				if err = tx.WithContext(l.ctx).
					Table(member.TableName()).
					Where("member_id = ?", in.MemberId).
					Updates(map[string]interface{}{
						"lottery_times": before,
						"update_time":   time.Now(),
					}).Error; err != nil {
					return err
				}
				rejected := buildRejectedRecord(activity, in.MemberId, in.RequestId, drawEligibilitySummary{
					Status:               drawEligibilityInventoryExhausted,
					Code:                 drawEligibilityInventoryExhausted,
					Message:              "当前卡池库存刚刚变化，请稍后重试",
					NextAction:           drawNextActionRetryLater,
					RemainingLotteryTime: before,
					RealNameStatus:       eligibility.RealNameStatus,
					RealNameStatusText:   eligibility.RealNameStatusText,
					RealNameMasked:       eligibility.RealNameMasked,
					CredentialRef:        eligibility.CredentialRef,
					VerifiedAt:           eligibility.VerifiedAt,
				})
				if err = createParticipationRecord(l.ctx, tx, rejected); err != nil {
					return err
				}
				record, detailErr := loadRecordDetailByID(l.ctx, tx, rejected.ID)
				if detailErr != nil {
					return detailErr
				}
				result.Record = record
				result.EligibilityCode = drawEligibilityInventoryExhausted
				result.EligibilityMessage = "当前卡池库存刚刚变化，请稍后重试"
				result.NextAction = drawNextActionRetryLater
				return nil
			}
		}

		recordRow := buildWinningOrNotRecord(activity, in.MemberId, in.RequestId, eligibility, before, after, winner, templateMap)
		if err = createParticipationRecord(l.ctx, tx, recordRow); err != nil {
			return err
		}
		if strings.TrimSpace(recordRow.ResultType) == drawResultTypeWon && strings.TrimSpace(recordRow.ResultStatus) == drawResultStatusWon {
			if _, err = cardassetservicelogic.EnsureCardInstanceByParticipationRecord(l.ctx, tx, recordRow.ID, "system", recordRow.TraceID); err != nil {
				return err
			}
		}
		record, detailErr := loadRecordDetailByID(l.ctx, tx, recordRow.ID)
		if detailErr != nil {
			return detailErr
		}
		result.Record = record
		result.EligibilityCode = drawEligibilityEligible
		result.EligibilityMessage = eligibility.Message
		result.NextAction = drawNextActionNone
		return nil
	})
	if err != nil {
		logc.Errorf(l.ctx, "参与抽卡失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("参与抽卡失败")
	}

	return result, nil
}
