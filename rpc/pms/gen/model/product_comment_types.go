package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ProductComment struct {
	ID               bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ProductId        int64         `bson:"productId,omitempty" json:"productId,omitempty"`                     // 商品id
	MemberNickName   string        `bson:"memberNickName,omitempty" json:"memberNickName,omitempty"`           // 评价者昵称
	ProductName      string        `bson:"productName,omitempty" json:"productName,omitempty"`                 // 商品名称
	Star             int32         `bson:"star,omitempty" json:"star,omitempty"`                               // 评价星数：0->5
	MemberIp         string        `bson:"memberIp,omitempty" json:"memberIp,omitempty"`                       // 评价的ip
	ShowStatus       int32         `bson:"showStatus,omitempty" json:"showStatus,omitempty"`                   // 是否显示，0->屏蔽，1->通过
	AuditStatus      int32         `bson:"auditStatus,omitempty" json:"auditStatus,omitempty"`                 // 审核状态：0-待审核，1-审核通过，2-审核拒绝，3-已屏蔽
	AuditRemark      string        `bson:"auditRemark,omitempty" json:"auditRemark,omitempty"`                 // 审核备注/拒绝原因
	AuditorId        int64         `bson:"auditorId,omitempty" json:"auditorId,omitempty"`                    // 审核人ID
	AuditorName      string        `bson:"auditorName,omitempty" json:"auditorName,omitempty"`                // 审核人名称
	AuditedAt        time.Time     `bson:"auditedAt,omitempty" json:"auditedAt,omitempty"`                    // 审核时间
	Hidden           int32         `bson:"hidden,omitempty" json:"hidden,omitempty"`                           // 是否屏蔽：0-显示，1-屏蔽
	AppealStatus     int32         `bson:"appealStatus,omitempty" json:"appealStatus,omitempty"`               // 申诉状态：0-未申诉，1-申诉中，2-申诉通过，3-申诉驳回
	AppealReason     string        `bson:"appealReason,omitempty" json:"appealReason,omitempty"`               // 申诉原因
	AppealReply      string        `bson:"appealReply,omitempty" json:"appealReply,omitempty"`                 // 申诉回复
	AppealedAt       time.Time     `bson:"appealedAt,omitempty" json:"appealedAt,omitempty"`                  // 申诉时间
	AppealHandledAt  time.Time     `bson:"appealHandledAt,omitempty" json:"appealHandledAt,omitempty"`        // 申诉处理时间
	ProductAttribute string        `bson:"productAttribute,omitempty" json:"productAttribute,omitempty"`       // 购买时的商品属性
	CollectCount     int32         `bson:"collectCount,omitempty" json:"collectCount,omitempty"`               // 点赞数
	ReadCount        int32         `bson:"readCount,omitempty" json:"readCount,omitempty"`                     // 阅读数
	Content          string        `bson:"content,omitempty" json:"content,omitempty"`                         // 内容
	Pics             string        `bson:"pics,omitempty" json:"pics,omitempty"`                               // 图片地址，逗号分隔
	MemberIcon       string        `bson:"memberIcon,omitempty" json:"memberIcon,omitempty"`                   // 评价者头像
	ReplayCount      int32         `bson:"replayCount,omitempty" json:"replayCount,omitempty"`                 // 回复数量
	MemberId         int64         `bson:"memberId,omitempty" json:"memberId,omitempty"`                       // 会员ID
	PlatformId       int64         `bson:"platformId,omitempty" json:"platformId,omitempty"`                   // 平台ID
	TenantId         int64         `bson:"tenantId,omitempty" json:"tenantId,omitempty"`                       // 租户ID
	MerchantId       int64         `bson:"merchantId,omitempty" json:"merchantId,omitempty"`                   // 商户ID
	OrderId          int64         `bson:"orderId,omitempty" json:"orderId,omitempty"`                         // 订单ID
	UpdateBy         string        `bson:"updateBy,omitempty" json:"updateBy,omitempty"`                       // 处理人
	UpdateAt         time.Time     `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt         time.Time     `bson:"createAt,omitempty" json:"createAt,omitempty"`
}
