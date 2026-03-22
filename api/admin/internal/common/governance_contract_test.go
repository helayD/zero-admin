package common

import (
	"reflect"
	"testing"

	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/cms/cmsclient"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/feihua/zero-admin/rpc/search/search"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestAdminQueryContractsExposeGovernanceScopeFields(t *testing.T) {
	assertStructHasFields(t, reflect.TypeOf(types.QueryProductSpuListReq{}), "ScopeType", "PlatformId", "TenantId", "MerchantId")
	assertStructHasFields(t, reflect.TypeOf(types.QueryProductSpuDetailReq{}), "ScopeType", "PlatformId", "TenantId", "MerchantId")
	assertStructHasFields(t, reflect.TypeOf(types.QueryProductSkuListReq{}), "ScopeType", "PlatformId", "TenantId", "MerchantId")
	assertStructHasFields(t, reflect.TypeOf(types.QueryProductSkuDetailReq{}), "ScopeType", "PlatformId", "TenantId", "MerchantId")
	assertStructHasFields(t, reflect.TypeOf(types.QueryOrderMainListReq{}), "ScopeType", "PlatformId", "TenantId", "MerchantId")
	assertStructHasFields(t, reflect.TypeOf(types.QueryOrderMainDetailReq{}), "ScopeType", "PlatformId", "TenantId", "MerchantId")
	assertStructHasFields(t, reflect.TypeOf(types.QueryCouponListReq{}), "ScopeType", "PlatformId", "TenantId", "MerchantId")
	assertStructHasFields(t, reflect.TypeOf(types.QueryCouponDetailReq{}), "ScopeType", "PlatformId", "TenantId", "MerchantId")
	assertStructHasFields(t, reflect.TypeOf(types.QuerySubjectListReq{}), "ScopeType", "PlatformId", "TenantId", "MerchantId")
	assertStructHasFields(t, reflect.TypeOf(types.QuerySubjectDetailReq{}), "ScopeType", "PlatformId", "TenantId", "MerchantId")
}

func TestRpcQueryContractsExposeGovernanceScopeField(t *testing.T) {
	assertProtoHasField(t, (&pmsclient.QueryProductSpuListReq{}).ProtoReflect().Descriptor(), "scope")
	assertProtoHasField(t, (&pmsclient.QueryProductSpuDetailReq{}).ProtoReflect().Descriptor(), "scope")
	assertProtoHasField(t, (&pmsclient.QueryProductSpuByIdsReq{}).ProtoReflect().Descriptor(), "scope")
	assertProtoHasField(t, (&pmsclient.QueryProductSkuListReq{}).ProtoReflect().Descriptor(), "scope")
	assertProtoHasField(t, (&pmsclient.QueryProductSkuDetailReq{}).ProtoReflect().Descriptor(), "scope")
	assertProtoHasField(t, (&omsclient.QueryOrderListReq{}).ProtoReflect().Descriptor(), "scope")
	assertProtoHasField(t, (&omsclient.QueryOrderDetailReq{}).ProtoReflect().Descriptor(), "scope")
	assertProtoHasField(t, (&smsclient.QueryCouponListReq{}).ProtoReflect().Descriptor(), "scope")
	assertProtoHasField(t, (&smsclient.QueryCouponDetailReq{}).ProtoReflect().Descriptor(), "scope")
	assertProtoHasField(t, (&smsclient.QueryCouponByScopeIdReq{}).ProtoReflect().Descriptor(), "scope")
	assertProtoHasField(t, (&smsclient.QueryCouponByCodeReq{}).ProtoReflect().Descriptor(), "scope")
	assertProtoHasField(t, (&cmsclient.QuerySubjectListReq{}).ProtoReflect().Descriptor(), "scope")
	assertProtoHasField(t, (&cmsclient.QuerySubjectDetailReq{}).ProtoReflect().Descriptor(), "scope")
	assertProtoHasField(t, (&cmsclient.SubjectListByIdsReq{}).ProtoReflect().Descriptor(), "scope")
}

func TestSearchContractsExposeGovernanceScopeField(t *testing.T) {
	assertProtoHasField(t, (&search.ProductData{}).ProtoReflect().Descriptor(), "scope")
	assertProtoHasField(t, (&search.SearchSimpleReq{}).ProtoReflect().Descriptor(), "scope")
	assertProtoHasField(t, (&search.SearchReq{}).ProtoReflect().Descriptor(), "scope")
	assertProtoHasField(t, (&search.RecommendReq{}).ProtoReflect().Descriptor(), "scope")
	assertProtoHasField(t, (&search.SearchRelatedReq{}).ProtoReflect().Descriptor(), "scope")
}

func assertStructHasFields(t *testing.T, target reflect.Type, fields ...string) {
	t.Helper()

	for _, field := range fields {
		if _, ok := target.FieldByName(field); !ok {
			t.Fatalf("%s missing field %s", target.Name(), field)
		}
	}
}

func assertProtoHasField(t *testing.T, descriptor protoreflect.MessageDescriptor, field protoreflect.Name) {
	t.Helper()

	if descriptor.Fields().ByName(field) == nil {
		t.Fatalf("%s missing proto field %s", descriptor.FullName(), field)
	}
}
