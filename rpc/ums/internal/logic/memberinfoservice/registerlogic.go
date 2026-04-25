package memberinfoservicelogic

import (
	"context"
	"errors"
	"strconv"

	"github.com/feihua/zero-admin/rpc/ums/gen/model"
	"github.com/feihua/zero-admin/rpc/ums/gen/query"
	"github.com/zeromicro/go-zero/core/logc"
	"golang.org/x/crypto/bcrypt"

	"github.com/feihua/zero-admin/rpc/ums/internal/svc"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	memberIdSeqKey = "ums:member_id_seq"

	// syncMemberIDSeqScript keeps the Redis sequence at least as large as
	// the current max member_id in MySQL before allocating the next id.
	syncMemberIDSeqScript = `
local current = redis.call("GET", KEYS[1])
local floor = tonumber(ARGV[1])
if current == false or tonumber(current) < floor then
  redis.call("SET", KEYS[1], floor)
end
return redis.call("INCR", KEYS[1])
`
)

// RegisterLogic 注册会员信息
/*
Author: LiuFeiHua
Date: 2025/5/21 15:13
*/
type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Register 注册会员信息
func (l *RegisterLogic) Register(in *umsclient.RegisterReq) (*umsclient.RegisterResp, error) {
	q := query.UmsMemberInfo
	count, countErr := q.WithContext(l.ctx).Where(q.Mobile.Eq(in.Mobile)).Count()
	if countErr != nil {
		logc.Errorf(l.ctx, "查询手机号失败,手机号: %s,异常:%s", in.Mobile, countErr.Error())
		return nil, errors.New("注册失败")
	}
	if count > 0 {
		logc.Errorf(l.ctx, "手机号已注册,手机号: %s", in.Mobile)
		return nil, errors.New("手机号已注册")
	}

	// 密码 bcrypt 哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		logc.Errorf(l.ctx, "密码哈希失败,手机号: %s,异常:%s", in.Mobile, err.Error())
		return nil, errors.New("注册失败")
	}

	accessExpire := l.svcCtx.Config.JWT.AccessExpire
	secret := l.svcCtx.Config.JWT.AccessSecret
	id, err := insertMember(in, string(hashedPassword), l)
	if err != nil {
		logc.Errorf(l.ctx, "新增会员失败,手机号:%s,异常:%s", in.Mobile, err.Error())
		return nil, errors.New("新增会员失败")
	}
	token, err := createJwtToken(secret, in.Nickname, in.Mobile, accessExpire, id)

	if err != nil {
		logc.Errorf(l.ctx, "生成token失败,手机号:%s,异常:%s", in.Mobile, err.Error())
		return nil, err
	}

	return &umsclient.RegisterResp{
		Token: token,
	}, nil
}

func insertMember(in *umsclient.RegisterReq, hashedPassword string, l *RegisterLogic) (int64, error) {
	// 使用 Redis INCR 生成 member_id，避免并发冲突
	last, dbErr := query.UmsMemberInfo.WithContext(l.ctx).Order(query.UmsMemberInfo.MemberID.Desc()).First()
	var initVal int64 = 1000
	if dbErr == nil && last != nil {
		initVal = last.MemberID
	}

	newMemberId, err := nextMemberID(l, initVal)
	if err != nil {
		return 0, errors.New("生成会员ID失败")
	}

	member := &model.UmsMemberInfo{
		MemberID: int64(newMemberId), // 会员ID
		LevelID:  1,                  // 等级ID
		Nickname: in.Nickname,        // 昵称
		Mobile:   in.Mobile,          // 手机号码
		Source:   in.Source,          // 注册来源：0-PC，1-APP，2-小程序
		Password: hashedPassword,     // bcrypt 哈希密码
	}
	err = query.UmsMemberInfo.WithContext(l.ctx).Create(member)

	if err != nil {
		return 0, err
	}

	return member.MemberID, nil
}

func nextMemberID(l *RegisterLogic, floor int64) (int64, error) {
	value, err := l.svcCtx.Redis.EvalCtx(l.ctx, syncMemberIDSeqScript, []string{memberIdSeqKey}, floor)
	if err != nil {
		return 0, err
	}

	memberID, ok := redisValueToInt64(value)
	if !ok {
		return 0, errors.New("生成会员ID失败")
	}

	return memberID, nil
}

func redisValueToInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int64:
		return v, true
	case string:
		n, err := strconv.ParseInt(v, 10, 64)
		return n, err == nil
	case []byte:
		n, err := strconv.ParseInt(string(v), 10, 64)
		return n, err == nil
	default:
		return 0, false
	}
}
