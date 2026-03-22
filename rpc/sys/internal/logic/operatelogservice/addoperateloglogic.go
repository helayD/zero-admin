package operatelogservicelogic

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/feihua/zero-admin/rpc/sys/gen/model"
	"github.com/feihua/zero-admin/rpc/sys/gen/query"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
	"path"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// AddOperateLogLogic 添加操作日志
/*
Author: LiuFeiHua
Date: 2023/12/18 17:08
*/
type AddOperateLogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddOperateLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddOperateLogLogic {
	return &AddOperateLogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

const (
	operateLogTitleMax         = 50
	operateLogMethodMax        = 200
	operateLogRequestMethodMax = 10
	operateLogOperateNameMax   = 50
	operateLogDeptNameMax      = 50
	operateLogURLMax           = 255
	operateLogLocationMax      = 255
	operateLogParamMax         = 2000
	operateLogJSONResultMax    = 2000
	operateLogPlatformMax      = 50
	operateLogBrowserMax       = 50
	operateLogVersionMax       = 50
	operateLogOSMax            = 50
	operateLogArchMax          = 50
	operateLogEngineMax        = 50
	operateLogEngineDetailsMax = 50
	operateLogExtraMax         = 1000
	operateLogErrorMsgMax      = 2000
)

// AddOperateLog 添加操作日志
func (l *AddOperateLogLogic) AddOperateLog(in *sysclient.AddOperateLogReq) (*sysclient.AddOperateLogResp, error) {

	uri := strings.Split(in.OperateUrl, "?")[0]

	key := l.svcCtx.RedisKey + "background_url"
	name, _ := l.svcCtx.Redis.HgetCtx(l.ctx, key, uri)

	if name == "" {
		type menuNameRow struct {
			MenuName string `gorm:"column:menu_name"`
		}
		var row menuNameRow
		_ = l.svcCtx.DB.WithContext(l.ctx).
			Table("sys_menu").
			Select("menu_name").
			Where("background_url = ? OR FIND_IN_SET(?, REPLACE(background_url, ' ', '')) > 0", uri, uri).
			Order("id ASC").
			Limit(1).
			Take(&row).Error
		name = row.MenuName
		if name == "" {
			_, _ = l.svcCtx.Redis.HdelCtx(l.ctx, l.svcCtx.RedisKey+"background_url", uri)
			name = fallbackOperateLogTitle(uri, in.JsonResult)
		} else {
			_ = l.svcCtx.Redis.HsetCtx(l.ctx, key, uri, name)
		}
	}

	sysLog := &model.SysOperateLog{
		Title:           trimToRunes(name, operateLogTitleMax),                     // 模块标题
		BusinessType:    in.BusinessType,                                           // 业务类型（0其它 1新增 2修改 3删除）
		Method:          trimToRunes(in.Method, operateLogMethodMax),               // 方法名称
		RequestMethod:   trimToRunes(in.RequestMethod, operateLogRequestMethodMax), // 请求方式
		OperatorType:    in.OperatorType,                                           // 操作类别（0其它 1后台用户 2手机端用户）
		OperateName:     trimToRunes(in.OperateName, operateLogOperateNameMax),     // 操作人员
		DeptName:        trimToRunes(in.DeptName, operateLogDeptNameMax),           // 部门名称
		OperateURL:      trimToRunes(in.OperateUrl, operateLogURLMax),              // 请求URL
		OperateIP:       in.OperateIp,                                              // 主机地址
		OperateLocation: trimToRunes(in.OperateLocation, operateLogLocationMax),    // 操作地点
		OperateParam:    trimToRunes(in.OperateParam, operateLogParamMax),          // 请求参数
		JSONResult:      trimToRunes(in.JsonResult, operateLogJSONResultMax),       // 返回参数
		Platform:        trimToRunes(in.Platform, operateLogPlatformMax),           // 平台信息
		Browser:         trimToRunes(in.Browser, operateLogBrowserMax),             // 浏览器类型
		Version:         trimToRunes(in.Version, operateLogVersionMax),             // 浏览器版本
		Os:              trimToRunes(in.Os, operateLogOSMax),                       // 操作系统
		Arch:            trimToRunes(in.Arch, operateLogArchMax),                   // 体系结构信息
		Engine:          trimToRunes(in.Engine, operateLogEngineMax),               // 渲染引擎信息
		EngineDetails:   trimToRunes(in.EngineDetails, operateLogEngineDetailsMax), // 渲染引擎详细信息
		Extra:           trimToRunes(in.Extra, operateLogExtraMax),                 // 其他信息（可选）
		Status:          in.Status,                                                 // 操作状态(0:异常,正常)
		ErrorMsg:        trimToRunes(in.ErrorMsg, operateLogErrorMsgMax),           // 错误消息
		OperateTime:     time.Now(),                                                // 操作时间
		CostTime:        in.CostTime,                                               // 消耗时间
	}

	if strings.Contains(in.JsonResult, "000000") {
		sysLog.Status = 1
	}

	err := query.SysOperateLog.WithContext(l.ctx).Create(sysLog)
	if err != nil {
		logc.Errorf(l.ctx, "添加操作日志失败,参数:%+v,异常:%s", sysLog, err.Error())
		return nil, errors.New("添加操作日志失败")
	}

	return &sysclient.AddOperateLogResp{}, nil
}

func fallbackOperateLogTitle(uri, jsonResult string) string {
	type resultMessage struct {
		Message string `json:"message"`
	}

	var result resultMessage
	if jsonResult != "" && json.Unmarshal([]byte(jsonResult), &result) == nil {
		message := strings.TrimSpace(result.Message)
		if message != "" {
			message = strings.TrimSuffix(message, "成功")
			message = strings.TrimSuffix(message, "失败")
			message = strings.TrimSpace(message)
			if message != "" {
				return message
			}
		}
	}

	name := strings.TrimSpace(path.Base(uri))
	if name == "" || name == "." || name == "/" {
		return "未登记接口"
	}

	return name
}

func trimToRunes(s string, limit int) string {
	if limit <= 0 || s == "" {
		return ""
	}

	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}

	return string(runes[:limit])
}
