package userservicelogic

import (
	"context"
	"errors"
	"fmt"

	"github.com/feihua/zero-admin/rpc/sys/gen/model"
	"github.com/feihua/zero-admin/rpc/sys/gen/query"
	logiccommon "github.com/feihua/zero-admin/rpc/sys/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// AddUserLogic 新增用户
/*
Author: LiuFeiHua
Date: 2023/12/18 14:13
*/
type AddUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddUserLogic {
	return &AddUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddUser 新增用户
// 1.查询用户名称是否存在
// 2.查询手机号是否存在
// 3.查询邮箱是否存在
// 4.用户不存在时,则直接添加用户
// 5.清空用户与岗位关联
// 6.添加用户与岗位关联
// 7.清空用户与角色关联(防止脏数据)
func (l *AddUserLogic) AddUser(in *sysclient.AddUserReq) (*sysclient.AddUserResp, error) {
	currentScope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	if err = logiccommon.ValidateTenantWritable(l.ctx, l.svcCtx.DB, currentScope); err != nil {
		return nil, err
	}
	if _, err = logiccommon.ValidateDeptInScope(l.ctx, l.svcCtx.DB, currentScope, in.DeptId); err != nil {
		return nil, err
	}
	if err = logiccommon.ValidatePostIDsInScope(l.ctx, l.svcCtx.DB, currentScope, in.PostIds); err != nil {
		return nil, err
	}
	if err = logiccommon.ValidateRoleIDsInScope(l.ctx, l.svcCtx.DB, currentScope, in.RoleIds); err != nil {
		return nil, err
	}

	q := query.SysUser

	// 1.查询用户名称是否存在
	name := in.UserName
	count, err := q.WithContext(l.ctx).Where(q.UserName.Eq(name)).Count()

	if err != nil {
		logc.Errorf(l.ctx, "根据用户名称：%s,查询用户失败,异常:%s", name, err.Error())
		return nil, errors.New(fmt.Sprintf("新增用户失败"))
	}

	if count > 0 {
		return nil, errors.New(fmt.Sprintf("用户：%s,已存在", name))
	}

	// 2.查询手机号是否存在
	mobile := in.Mobile
	count, err = q.WithContext(l.ctx).Where(q.Mobile.Eq(mobile)).Count()

	if err != nil {
		logc.Errorf(l.ctx, "根据手机号：%s,查询用户失败,异常:%s", mobile, err.Error())
		return nil, errors.New(fmt.Sprintf("新增用户失败"))
	}

	if count > 0 {
		return nil, errors.New(fmt.Sprintf("手机号：%s,已存在", mobile))
	}

	// 3.查询邮箱是否存在
	email := in.Email
	count, err = q.WithContext(l.ctx).Where(q.Email.Eq(email)).Count()

	if err != nil {
		logc.Errorf(l.ctx, "根据邮箱：%s,查询用户失败,异常:%s", email, err.Error())
		return nil, errors.New(fmt.Sprintf("新增用户失败"))
	}

	if count > 0 {
		return nil, errors.New(fmt.Sprintf("邮箱：%s,已存在", email))
	}

	avatar := in.Avatar
	// 默认用户头像
	if len(avatar) == 0 {
		avatar = "https://gw.alipayobjects.com/zos/antfincdn/XAosXuNZyF/BiazfanxmamNRoxxVxka.png"
	}
	user := &model.SysUser{
		Mobile:   in.Mobile,   // 手机号码
		UserName: in.UserName, // 用户账号
		NickName: in.NickName, // 用户昵称
		UserType: in.UserType, // 用户类型（00系统用户）
		Avatar:   avatar,      // 头像路径
		Email:    in.Email,    // 用户邮箱
		Password: in.Password, // 密码
		Status:   in.Status,   // 状态(1:正常，0:禁用)
		DeptID:   in.DeptId,   // 部门ID
		Remark:   in.Remark,   // 备注
		CreateBy: in.CreateBy, // 创建者
	}

	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err = tx.Create(user).Error; err != nil {
			logc.Errorf(l.ctx, "新增用户异常,参数:%+v,异常:%s", user, err.Error())
			return err
		}

		if err = tx.Table("sys_user").
			Where("id = ?", user.ID).
			Updates(map[string]interface{}{
				"platform_id": currentScope.PlatformID,
				"tenant_id":   currentScope.TenantID,
				"merchant_id": currentScope.MerchantID,
			}).Error; err != nil {
			return err
		}

		if err = logiccommon.UpsertUserScopeBinding(l.ctx, tx, user.ID, in.DeptId, currentScope, in.ActivationStatus, in.RoleMode, in.CreateBy); err != nil {
			return err
		}

		if err = tx.Where("user_id = ?", user.ID).Delete(&model.SysUserPost{}).Error; err != nil {
			return err
		}
		if len(in.PostIds) > 0 {
			var userPosts []*model.SysUserPost
			for _, postId := range in.PostIds {
				userPosts = append(userPosts, &model.SysUserPost{UserID: user.ID, PostID: postId})
			}
			if err = tx.Create(&userPosts).Error; err != nil {
				return err
			}
		}

		if err = tx.Where("user_id = ?", user.ID).Delete(&model.SysUserRole{}).Error; err != nil {
			return err
		}
		if len(in.RoleIds) > 0 {
			var userRoles []*model.SysUserRole
			for _, roleId := range in.RoleIds {
				userRoles = append(userRoles, &model.SysUserRole{UserID: user.ID, RoleID: roleId})
			}
			if err = tx.Create(&userRoles).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		logc.Errorf(l.ctx, "新增用户异常,参数:%+v,异常:%s", user, err.Error())
		return nil, errors.New("新增用户异常")
	}
	return &sysclient.AddUserResp{}, nil
}
