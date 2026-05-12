package memberinfoservicelogic

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/feihua/zero-admin/rpc/ums/gen/model"
	"github.com/feihua/zero-admin/rpc/ums/gen/query"
	"github.com/feihua/zero-admin/rpc/ums/internal/svc"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	uuidlib "github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// LoginByCodeLogic 验证码登录注册合并接口（Story 3.1.1）。
//
// 流程:
//  1. 从 Redis 读 ums:sms:login:{mobile}，按 ":" 切出 code 与 errCount
//  2. 不匹配：errCount+1 回写（保留 TTL）；超 5 次直接 Del
//  3. 匹配：立即 Del（一次性）
//  4. 查 ums_member_info by mobile：
//     - 命中 + is_enabled=0 → 拒绝（验证码仍消费，避免被遍历探测）
//     - 命中 + is_enabled=1 → 走登录分支：写日志/抽卡/首登优惠券，签发 JWT
//     - 未命中 → 自动建号：member_id 来自 Redis INCR，nickname 自动生成，
//     password 写入随机 bcrypt 哈希（不再用于认证），插入后走登录分支
type LoginByCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginByCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginByCodeLogic {
	return &LoginByCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// LoginByCode 实现 MemberAuthService.LoginByCode
func (l *LoginByCodeLogic) LoginByCode(in *umsclient.LoginByCodeReq) (*umsclient.LoginByCodeResp, error) {
	mobile := strings.TrimSpace(in.Mobile)
	code := strings.TrimSpace(in.Code)
	if mobile == "" || code == "" {
		return nil, errors.New("手机号或验证码不能为空")
	}

	if err := l.verifyAndConsumeCode(mobile, code); err != nil {
		return nil, err
	}

	// 验证码已消费，开始查/建会员
	q := query.UmsMemberInfo
	member, err := q.WithContext(l.ctx).Where(q.Mobile.Eq(mobile)).First()
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return l.autoRegisterAndIssueToken(mobile, in.Source, in.Ip)
	case err != nil:
		logc.Errorf(l.ctx, "查询会员失败,手机号:%s,异常:%s", mobile, err.Error())
		return nil, errors.New("登录失败，请稍后重试")
	}

	if member.IsEnabled == 0 {
		logc.Errorf(l.ctx, "账号已被禁用,手机号:%s", mobile)
		return nil, errors.New("账号已被禁用，请联系客服")
	}

	token, err := l.issueToken(member.MemberID, member.Nickname, member.Mobile)
	if err != nil {
		return nil, err
	}

	runPostLoginActions(l.ctx, l.svcCtx.DB, l.svcCtx.RabbitMQ, postLoginParams{
		MemberID:         member.MemberID,
		Nickname:         member.Nickname,
		FirstLoginStatus: member.FirstLoginStatus,
		IP:               in.Ip,
		Source:           in.Source,
	})

	return &umsclient.LoginByCodeResp{
		Token:     token,
		IsNewUser: false,
		MemberId:  member.MemberID,
	}, nil
}

// verifyAndConsumeCode 校验验证码并消费 Redis key。
//   - 验证码错误：errCount+1 回写，超 5 次直接 Del
//   - 验证码正确：立即 Del 实现一次性
func (l *LoginByCodeLogic) verifyAndConsumeCode(mobile, code string) error {
	codeKey := smsLoginCodeKey(mobile)
	stored, err := l.svcCtx.Redis.GetCtx(l.ctx, codeKey)
	if err != nil {
		logc.Errorf(l.ctx, "读取验证码 Redis 失败,手机号:%s,异常:%s", mobile, err.Error())
		return errors.New("登录失败，请稍后重试")
	}
	if stored == "" {
		return errors.New("验证码错误或已过期")
	}

	storedCode, errCount := parseMemberSequence(stored)
	if storedCode == "" {
		// Redis 数据格式异常 — 直接清掉避免无限循环
		_, _ = l.svcCtx.Redis.DelCtx(l.ctx, codeKey)
		return errors.New("验证码错误或已过期")
	}

	if storedCode != code {
		newErrCount := errCount + 1
		if newErrCount >= smsMaxErrAttempts {
			_, _ = l.svcCtx.Redis.DelCtx(l.ctx, codeKey)
			return errors.New("验证码错误次数过多，请重新获取")
		}
		// 保留剩余 TTL：先取 TTL 再 SetEx
		ttl, ttlErr := l.svcCtx.Redis.TtlCtx(l.ctx, codeKey)
		if ttlErr != nil || ttl <= 0 {
			ttl = 300
		}
		if err := l.svcCtx.Redis.SetexCtx(l.ctx, codeKey, formatMemberSequence(storedCode, newErrCount), ttl); err != nil {
			logc.Errorf(l.ctx, "回写验证码错误计数失败,手机号:%s,异常:%s", mobile, err.Error())
		}
		return errors.New("验证码错误或已过期")
	}

	// 验证码命中 → 一次性消费
	if _, err := l.svcCtx.Redis.DelCtx(l.ctx, codeKey); err != nil {
		logc.Errorf(l.ctx, "消费验证码 Redis key 失败,手机号:%s,异常:%s", mobile, err.Error())
	}
	return nil
}

// autoRegisterAndIssueToken 未注册分支：自动建号 + 签发 JWT。
func (l *LoginByCodeLogic) autoRegisterAndIssueToken(mobile string, source int32, ip string) (*umsclient.LoginByCodeResp, error) {
	memberID, err := generateNewMemberID(l.ctx, l.svcCtx.Redis.EvalCtx)
	if err != nil {
		logc.Errorf(l.ctx, "自动建号生成 member_id 失败,手机号:%s,异常:%s", mobile, err.Error())
		return nil, errors.New("登录失败，请稍后重试")
	}

	nickname, err := l.allocateNickname(mobile)
	if err != nil {
		return nil, err
	}

	// 密码字段写入随机 bcrypt 哈希（schema 兼容，不再用于认证比对）
	hashed, err := bcrypt.GenerateFromPassword([]byte(uuidlib.NewString()), bcrypt.DefaultCost)
	if err != nil {
		logc.Errorf(l.ctx, "生成随机密码哈希失败,手机号:%s,异常:%s", mobile, err.Error())
		return nil, errors.New("登录失败，请稍后重试")
	}

	newMember := &model.UmsMemberInfo{
		MemberID:         memberID,
		LevelID:          1,
		Nickname:         nickname,
		Mobile:           mobile,
		Source:           source,
		Password:         string(hashed),
		IsEnabled:        1,
		FirstLoginStatus: 1,
	}
	if err := query.UmsMemberInfo.WithContext(l.ctx).Create(newMember); err != nil {
		logc.Errorf(l.ctx, "自动建号写入会员失败,手机号:%s,异常:%s", mobile, err.Error())
		return nil, errors.New("登录失败，请稍后重试")
	}

	token, err := l.issueToken(memberID, nickname, mobile)
	if err != nil {
		return nil, err
	}

	runPostLoginActions(l.ctx, l.svcCtx.DB, l.svcCtx.RabbitMQ, postLoginParams{
		MemberID:         memberID,
		Nickname:         nickname,
		FirstLoginStatus: 1,
		IP:               ip,
		Source:           source,
	})

	return &umsclient.LoginByCodeResp{
		Token:     token,
		IsNewUser: true,
		MemberId:  memberID,
	}, nil
}

// allocateNickname 生成"用户_XXXX"昵称：先取手机号后 4 位，冲突时追加 4 位随机数字。
func (l *LoginByCodeLogic) allocateNickname(mobile string) (string, error) {
	if len(mobile) < 4 {
		return "", errors.New("手机号格式不正确")
	}
	suffix := mobile[len(mobile)-4:]
	candidate := "用户_" + suffix

	q := query.UmsMemberInfo
	for attempt := 0; attempt < 5; attempt++ {
		count, err := q.WithContext(l.ctx).Where(q.Nickname.Eq(candidate)).Count()
		if err != nil {
			logc.Errorf(l.ctx, "校验昵称冲突失败,昵称:%s,异常:%s", candidate, err.Error())
			return "", errors.New("登录失败，请稍后重试")
		}
		if count == 0 {
			return candidate, nil
		}
		// 冲突 → 追加 4 位随机数字重试
		rnd, err := rand.Int(rand.Reader, big.NewInt(10000))
		if err != nil {
			return "", fmt.Errorf("生成随机昵称失败: %w", err)
		}
		candidate = "用户_" + suffix + strconv.FormatInt(rnd.Int64(), 10)
	}
	// 极端情况下 5 次仍冲突，附加 uuid 前 6 位作为兜底
	candidate = "用户_" + suffix + strings.ReplaceAll(uuidlib.NewString(), "-", "")[:6]
	return candidate, nil
}

// issueToken 签发 JWT token，复用既有 createJwtToken。
// claims 结构（memberId/memberName/mobile）保持与旧 LoginLogic 一致以避免破坏前端。
func (l *LoginByCodeLogic) issueToken(memberID int64, nickname, mobile string) (string, error) {
	accessExpire := l.svcCtx.Config.JWT.AccessExpire
	secret := l.svcCtx.Config.JWT.AccessSecret
	token, err := createJwtToken(secret, nickname, mobile, accessExpire, memberID)
	if err != nil {
		logc.Errorf(l.ctx, "生成 token 失败,memberId:%d,异常:%s", memberID, err.Error())
		return "", errors.New("生成token失败")
	}
	return token, nil
}

// parseMemberSequence 解析 Redis 中 "{code}:{errCount}" 格式。
// 失败时返回空 code，调用方据此走「数据异常」路径。
func parseMemberSequence(raw string) (string, int) {
	idx := strings.LastIndex(raw, ":")
	if idx < 0 {
		// 兼容旧版本只存 code 不存计数的情况
		return raw, 0
	}
	codePart := raw[:idx]
	countPart := raw[idx+1:]
	count, err := strconv.Atoi(countPart)
	if err != nil {
		return "", 0
	}
	return codePart, count
}
