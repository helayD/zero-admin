package memberinfoservicelogic

import (
	"context"
	"errors"
	query "github.com/feihua/zero-admin/rpc/ums/gen/query"
	"github.com/feihua/zero-admin/rpc/ums/internal/svc"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"strings"
	"time"
)

// UpdateMemberInfoLogic 更新会员信息
/*
Author: LiuFeiHua
Date: 2025/05/21 14:18:26
*/
type UpdateMemberInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateMemberInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMemberInfoLogic {
	return &UpdateMemberInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateMemberInfo 更新会员信息
func (l *UpdateMemberInfoLogic) UpdateMemberInfo(in *umsclient.UpdateMemberInfoReq) (*umsclient.UpdateMemberInfoResp, error) {
	q := query.UmsMemberInfo.WithContext(l.ctx)

	// 1.根据会员信息id查询会员信息是否已存在
	_, err := q.Where(query.UmsMemberInfo.ID.Eq(in.Id)).First()

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		logc.Errorf(l.ctx, "会员不存在, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("会员不存在")
	case err != nil:
		logc.Errorf(l.ctx, "查询会员异常, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("查询会员异常")
	}

	now := time.Now()
	updates := map[string]any{
		"update_time": &now,
	}

	profileUpdateRequested := strings.TrimSpace(in.Nickname) != "" ||
		strings.TrimSpace(in.Mobile) != "" ||
		strings.TrimSpace(in.Avatar) != "" ||
		strings.TrimSpace(in.Signature) != "" ||
		strings.TrimSpace(in.Birthday) != "" ||
		in.Gender != 0

	if profileUpdateRequested {
		if strings.TrimSpace(in.Nickname) != "" {
			updates["nickname"] = in.Nickname
		}
		if strings.TrimSpace(in.Mobile) != "" {
			updates["mobile"] = in.Mobile
		}
		if strings.TrimSpace(in.Avatar) != "" {
			updates["avatar"] = in.Avatar
		}
		if strings.TrimSpace(in.Signature) != "" {
			updates["signature"] = in.Signature
		}
		if strings.TrimSpace(in.Birthday) != "" {
			birthday, parseErr := time.Parse("2006-01-02", in.Birthday)
			if parseErr != nil {
				return nil, errors.New("生日格式错误")
			}
			updates["birthday"] = &birthday
		}
		if in.Gender != 0 {
			updates["gender"] = in.Gender
		}
	}

	if strings.TrimSpace(in.Password) != "" {
		hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if hashErr != nil {
			logc.Errorf(l.ctx, "更新会员密码哈希失败,请求参数：%+v,异常信息:%s", in, hashErr.Error())
			return nil, errors.New("更新会员信息失败")
		}
		updates["password"] = string(hashedPassword)
	}

	if len(updates) == 1 {
		return &umsclient.UpdateMemberInfoResp{}, nil
	}

	// 2.会员信息存在时,则直接更新会员信息
	_, err = q.Where(query.UmsMemberInfo.ID.Eq(in.Id)).Updates(updates)

	if err != nil {
		logc.Errorf(l.ctx, "更新会员信息失败,参数:%+v,异常:%s", updates, err.Error())
		return nil, errors.New("更新会员信息失败")
	}

	return &umsclient.UpdateMemberInfoResp{}, nil
}
