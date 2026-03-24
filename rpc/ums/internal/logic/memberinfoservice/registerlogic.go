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
	const memberIdSeqKey = "ums:member_id_seq"
	rds := l.svcCtx.Redis

	// 初始化序列：用 Setnx 原子操作，避免 TOCTOU 竞态
	last, dbErr := query.UmsMemberInfo.WithContext(l.ctx).Order(query.UmsMemberInfo.MemberID.Desc()).First()
	var initVal int64 = 1000
	if dbErr == nil && last != nil {
		initVal = last.MemberID
	}
	// Setnx 只在 key 不存在时设置，并发安全
	_, _ = rds.Setnx(memberIdSeqKey, strconv.FormatInt(initVal, 10))

	newMemberId, err := rds.Incr(memberIdSeqKey)
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
