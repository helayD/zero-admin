package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ProductComment struct {
	ID               bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ProductId        int64         `bson:"productId,omitempty" json:"productId,omitempty"`                     // 商品id
	MemberNickName   string        `bson:"memberNickName,omitempty" json:"memberNickName,omitempty"`           // 评价者昵称
	ProductName      string        `bson:"productName,omitempty" json:"productName,omitempty"`               // 商品名称
	Star             int32         `bson:"star,omitempty" json:"star,omitempty"`                             // 评价星数：0->5
	MemberIp         string        `bson:"memberIp,omitempty" json:"memberIp,omitempty"`                     // 评价的ip
	ShowStatus       int32         `bson:"showStatus,omitempty" json:"showStatus,omitempty"`                   // 是否显示，0->屏蔽，1->通过
	ProductAttribute string        `bson:"productAttribute,omitempty" json:"productAttribute,omitempty"`       // 购买时的商品属性
	CollectCount     int32         `bson:"collectCount,omitempty" json:"collectCount,omitempty"`               // 点赞数
	ReadCount        int32         `bson:"readCount,omitempty" json:"readCount,omitempty"`                   // 阅读数
	Content          string        `bson:"content,omitempty" json:"content,omitempty"`                       // 内容
	Pics             string        `bson:"pics,omitempty" json:"pics,omitempty"`                             // 图片地址，逗号分隔
	MemberIcon       string        `bson:"memberIcon,omitempty" json:"memberIcon,omitempty"`                 // 评价者头像
	ReplayCount      int32         `bson:"replayCount,omitempty" json:"replayCount,omitempty"`               // 回复数量
	MemberId         int64         `bson:"memberId,omitempty" json:"memberId,omitempty"`                     // 会员ID
	PlatformId       int64         `bson:"platformId,omitempty" json:"platformId,omitempty"`                   // 平台ID
	TenantId         int64         `bson:"tenantId,omitempty" json:"tenantId,omitempty"`                     // 租户ID
	MerchantId       int64         `bson:"merchantId,omitempty" json:"merchantId,omitempty"`                 // 商户ID
	OrderId          int64         `bson:"orderId,omitempty" json:"orderId,omitempty"`                       // 订单ID
	UpdateBy         string        `bson:"updateBy,omitempty" json:"updateBy,omitempty"`                     // 处理人
	UpdateAt         time.Time     `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt         time.Time     `bson:"createAt,omitempty" json:"createAt,omitempty"`
}
