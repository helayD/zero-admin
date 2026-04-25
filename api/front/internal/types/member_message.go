package types

// QueryMemberMessageReq - 查询消息列表
type QueryMemberMessageReq struct {
	MessageType int64 `form:"messageType,optional"` // 消息类型（可选）
	Status      int64 `form:"status,default=-1"`    // 状态 0:未读 1:已读（可选）
	PageNum     int64 `form:"pageNum,default=1"`
	PageSize    int64 `form:"pageSize,default=20"`
}

// MemberMessageResp - 单条消息
type MemberMessageResp struct {
	ID             int64                    `json:"id"`
	MessageType    int64                    `json:"messageType"`
	Title          string                   `json:"title"`
	Content        string                   `json:"content"`
	ImageURL       string                   `json:"imageUrl"`
	LinkType       string                   `json:"linkType"`
	LinkID         string                   `json:"linkId"`
	RelatedOrderID int64                    `json:"relatedOrderId"`
	Intent         *MemberMessageIntentResp `json:"intent,omitempty"`
	Status         int64                    `json:"status"`
	ReadTime       string                   `json:"readTime"`
	CreateTime     string                   `json:"createTime"`
}

// MemberMessageIntentResp - 消息唤回意图
type MemberMessageIntentResp struct {
	IntentType       string `json:"intentType"`
	TargetType       string `json:"targetType"`
	TargetID         *int64 `json:"targetId,omitempty"`
	TargetTab        *int64 `json:"targetTab,omitempty"`
	FallbackType     string `json:"fallbackType"`
	FallbackTargetID *int64 `json:"fallbackTargetId,omitempty"`
	FallbackTab      *int64 `json:"fallbackTab,omitempty"`
	RequiresAuth     bool   `json:"requiresAuth"`
	MinAppVersion    string `json:"minAppVersion"`
	IntentID         string `json:"intentId"`
	IssuedAt         string `json:"issuedAt"`
	FailureReason    string `json:"failureReason"`
	RecoveryHint     string `json:"recoveryHint"`
	Blocked          bool   `json:"blocked"`
	Source           string `json:"source"`
}

// QueryMemberMessageListResp - 消息列表响应
type QueryMemberMessageListResp struct {
	Code     int64               `json:"code"`
	Message  string              `json:"message"`
	Data     []MemberMessageResp `json:"data"`
	Total    int64               `json:"total"`
	PageNum  int64               `json:"pageNum"`
	PageSize int64               `json:"pageSize"`
}

// MemberMessageDetailReq - 消息详情请求
type MemberMessageDetailReq struct {
	ID int64 `path:"id,optional" form:"id,optional" json:"id,optional"`
}

// MemberMessageDetailResp - 消息详情响应
type MemberMessageDetailResp struct {
	Code    int64              `json:"code"`
	Message string             `json:"message"`
	Data    *MemberMessageResp `json:"data"`
}

// MarkMessageReadReq - 标记单条已读请求
type MarkMessageReadReq struct {
	ID int64 `path:"id,optional" form:"id,optional" json:"id,optional"`
}

// BaseResp - 基础响应
type BaseResp struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

// MarkAllMessagesReadResp - 标记全部已读响应
type MarkAllMessagesReadResp struct {
	Code         int64  `json:"code"`
	Message      string `json:"message"`
	UpdatedCount int64  `json:"updatedCount"`
}

// QueryUnreadCountResp - 未读消息数量响应
type QueryUnreadCountResp struct {
	Code        int64  `json:"code"`
	Message     string `json:"message"`
	UnreadCount int64  `json:"unreadCount"`
}

// BaseReq - 基础请求
type BaseReq struct {
}
