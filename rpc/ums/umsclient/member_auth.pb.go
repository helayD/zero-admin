// Hand-written protobuf message types for MemberAuthService.
// 不通过 protoc 生成，遵循 sysclient.ChannelIntegrationTemplateData 的手写模式：
//   - 仅实现 protobuf legacy reflection 必需的 3 个方法（Reset/String/ProtoMessage）
//   - 通过 struct tag 完成 wire-format 编解码
// 这样既避免 protoc 重新生成 784KB 的 ums.pb.go，也保证 grpc-go 兼容。

package umsclient

import "fmt"

// SendSmsCodeReq 发送短信验证码请求
type SendSmsCodeReq struct {
	Mobile string `protobuf:"bytes,1,opt,name=mobile,proto3" json:"mobile,omitempty"`
	Scene  int32  `protobuf:"varint,2,opt,name=scene,proto3" json:"scene,omitempty"` // 1=member_login，预留其他场景
}

func (x *SendSmsCodeReq) Reset()         { *x = SendSmsCodeReq{} }
func (x *SendSmsCodeReq) String() string { return fmt.Sprintf("%+v", *x) }
func (*SendSmsCodeReq) ProtoMessage()    {}

// SendSmsCodeResp 发送短信验证码响应
//
// 安全约束: 响应体绝不包含验证码原文，即使 mock 模式。
type SendSmsCodeResp struct {
	ExpireSeconds int32 `protobuf:"varint,1,opt,name=expire_seconds,json=expireSeconds,proto3" json:"expire_seconds,omitempty"`
}

func (x *SendSmsCodeResp) Reset()         { *x = SendSmsCodeResp{} }
func (x *SendSmsCodeResp) String() string { return fmt.Sprintf("%+v", *x) }
func (*SendSmsCodeResp) ProtoMessage()    {}

// LoginByCodeReq 验证码登录注册合并接口请求
type LoginByCodeReq struct {
	Mobile string `protobuf:"bytes,1,opt,name=mobile,proto3" json:"mobile,omitempty"`
	Code   string `protobuf:"bytes,2,opt,name=code,proto3" json:"code,omitempty"`
	Source int32  `protobuf:"varint,3,opt,name=source,proto3" json:"source,omitempty"` // 0=PC，1=APP，2=小程序
	Ip     string `protobuf:"bytes,4,opt,name=ip,proto3" json:"ip,omitempty"`
}

func (x *LoginByCodeReq) Reset()         { *x = LoginByCodeReq{} }
func (x *LoginByCodeReq) String() string { return fmt.Sprintf("%+v", *x) }
func (*LoginByCodeReq) ProtoMessage()    {}

// LoginByCodeResp 验证码登录注册合并接口响应
//
// IsNewUser=true 表示该次调用触发了自动建号；客户端可据此展示新人欢迎语。
type LoginByCodeResp struct {
	Token     string `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
	IsNewUser bool   `protobuf:"varint,2,opt,name=is_new_user,json=isNewUser,proto3" json:"is_new_user,omitempty"`
	MemberId  int64  `protobuf:"varint,3,opt,name=member_id,json=memberId,proto3" json:"member_id,omitempty"`
}

func (x *LoginByCodeResp) Reset()         { *x = LoginByCodeResp{} }
func (x *LoginByCodeResp) String() string { return fmt.Sprintf("%+v", *x) }
func (*LoginByCodeResp) ProtoMessage()    {}
