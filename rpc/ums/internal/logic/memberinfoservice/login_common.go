package memberinfoservicelogic

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/feihua/zero-admin/pkg/mq"
	"github.com/feihua/zero-admin/rpc/ums/gen/model"
	"github.com/feihua/zero-admin/rpc/ums/gen/query"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"
)

// Story 3.1.1: 抽取 Login 与 LoginByCode 共享的「登录后置动作」
//   - 写登录日志
//   - 赠送每日登录抽卡次数
//   - 赠送每日登录积分
//   - 首登赠优惠券
//
// 旧的 loginlogic.go 与新的 loginbycodelogic.go 都通过本文件复用上述能力。

// postLoginParams 登录后置动作所需的最小信息集合。
type postLoginParams struct {
	MemberID         int64
	Nickname         string
	FirstLoginStatus int32 // 1=首次登录，0=已登录过
	IP               string
	Source           int32 // 0=PC，1=APP，2=小程序
}

// runPostLoginActions 执行登录成功后的副作用：登录日志 / 抽卡赠送 / 首登优惠券。
//
// 副作用任何一步失败都只打日志、不阻断主流程（保持与旧 LoginLogic 行为一致），
// 因此返回值只是为了方便单测断言。
func runPostLoginActions(ctx context.Context, db *gorm.DB, rabbit *mq.RabbitMQ, p postLoginParams) {
	now := time.Now()
	loginLog := &model.UmsMemberLoginLog{
		MemberID:   p.MemberID,
		CreateTime: now,
		MemberIP:   p.IP,
		City:       "todo",
		LoginType:  p.Source,
		Province:   "todo",
	}
	if err := query.UmsMemberLoginLog.WithContext(ctx).Create(loginLog); err != nil {
		logc.Errorf(ctx, "添加会员登录日志失败,参数：%+v,异常:%s", loginLog, err.Error())
		// 兼容旧行为：登录日志失败不阻断
	}

	if err := grantDailyLoginLotteryTimes(ctx, db, p.MemberID, now); err != nil {
		logc.Errorf(ctx, "每日登录赠送抽卡次数失败,memberId:%d,异常:%s", p.MemberID, err.Error())
	}

	if err := grantDailyLoginPoints(ctx, db, p.MemberID, now); err != nil {
		logc.Errorf(ctx, "每日登录赠送积分失败,memberId:%d,异常:%s", p.MemberID, err.Error())
	}

	if p.FirstLoginStatus == 1 {
		sendFirstLoginCouponMsg(ctx, rabbit, p.MemberID, p.Nickname, p.FirstLoginStatus)
	}
}

// sendFirstLoginCouponMsg 与旧 sendCouponMsg 行为等价，但去掉对 *LoginLogic 的耦合。
func sendFirstLoginCouponMsg(ctx context.Context, rabbit *mq.RabbitMQ, memberID int64, nickname string, firstLoginStatus int32) {
	if rabbit == nil {
		return
	}
	logc.Infof(ctx, "用户：%s,首次登录，将发放新手优惠券", nickname)

	param := map[string]any{
		"memberId":         memberID,
		"nickname":         nickname,
		"firstLoginStatus": firstLoginStatus,
	}
	body, err := sonic.Marshal(param)
	if err != nil {
		logc.Errorf(ctx, "序列化 JSON 失败: %v", err)
		return
	}
	if err := rabbit.SendMessage(
		"coupon.event.exchange",
		"direct",
		"first.login.queue",
		"first.login.key",
		body,
	); err != nil {
		logc.Errorf(ctx, "发送新手优惠券消息失败,参数：%+v,异常:%s", param, err.Error())
	}
}

// generateNewMemberID 通过 Redis INCR 序列分配下一个 member_id。
// 与 RegisterLogic 中 nextMemberID 等价，但解耦于 *RegisterLogic 接收者。
func generateNewMemberID(ctx context.Context, redisEval redisEvalCtx) (int64, error) {
	last, dbErr := query.UmsMemberInfo.WithContext(ctx).Order(query.UmsMemberInfo.MemberID.Desc()).First()
	var initVal int64 = 1000
	if dbErr == nil && last != nil {
		initVal = last.MemberID
	}

	value, err := redisEval(ctx, syncMemberIDSeqScript, []string{memberIdSeqKey}, initVal)
	if err != nil {
		return 0, err
	}
	memberID, ok := redisValueToInt64(value)
	if !ok {
		return 0, errors.New("生成会员ID失败")
	}
	return memberID, nil
}

// redisEvalCtx 抽象 redis.Redis.EvalCtx，便于测试时 mock。
type redisEvalCtx func(ctx context.Context, script string, keys []string, args ...interface{}) (any, error)

// formatMemberSequence Redis 验证码 + 错误计数序列化格式: "{code}:{errCount}"
func formatMemberSequence(code string, errCount int) string {
	return code + ":" + strconv.Itoa(errCount)
}
