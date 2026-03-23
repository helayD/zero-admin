package pmsclient

import (
	"reflect"
	"testing"
)

func requireField(t *testing.T, typ reflect.Type, name string, want reflect.Type) {
	t.Helper()
	field, ok := typ.FieldByName(name)
	if !ok {
		t.Fatalf("%s missing field %s", typ.Name(), name)
	}
	if want != nil && field.Type != want {
		t.Fatalf("%s field %s type mismatch: got %v want %v", typ.Name(), name, field.Type, want)
	}
}

func TestPmsGovernanceScopeContracts(t *testing.T) {
	scopePtr := reflect.TypeOf((*GovernanceScope)(nil))
	cases := []struct {
		name   string
		typ    reflect.Type
		fields map[string]reflect.Type
	}{
		{
			name: "attribute requests and responses",
			typ:  reflect.TypeOf(AddProductAttributeReq{}),
			fields: map[string]reflect.Type{"Scope": scopePtr},
		},
		{
			name: "attribute delete request",
			typ:  reflect.TypeOf(DeleteProductAttributeReq{}),
			fields: map[string]reflect.Type{"UpdateBy": reflect.TypeOf(int64(0)), "Scope": scopePtr},
		},
		{
			name: "attribute detail response",
			typ:  reflect.TypeOf(QueryProductAttributeDetailResp{}),
			fields: map[string]reflect.Type{"ScopeType": reflect.TypeOf(""), "PlatformId": reflect.TypeOf(int64(0)), "TenantId": reflect.TypeOf(int64(0)), "MerchantId": reflect.TypeOf(int64(0))},
		},
		{
			name: "attribute list request",
			typ:  reflect.TypeOf(QueryProductAttributeListReq{}),
			fields: map[string]reflect.Type{"Scope": scopePtr},
		},
		{
			name: "attribute list row",
			typ:  reflect.TypeOf(ProductAttributeListData{}),
			fields: map[string]reflect.Type{"ScopeType": reflect.TypeOf(""), "PlatformId": reflect.TypeOf(int64(0)), "TenantId": reflect.TypeOf(int64(0)), "MerchantId": reflect.TypeOf(int64(0))},
		},
		{
			name: "attribute group update/status/detail/list",
			typ:  reflect.TypeOf(UpdateProductAttributeGroupReq{}),
			fields: map[string]reflect.Type{"Scope": scopePtr},
		},
		{
			name: "attribute group status req",
			typ:  reflect.TypeOf(UpdateProductAttributeGroupStatusReq{}),
			fields: map[string]reflect.Type{"Scope": scopePtr},
		},
		{
			name: "attribute group detail resp",
			typ:  reflect.TypeOf(QueryProductAttributeGroupDetailResp{}),
			fields: map[string]reflect.Type{"ScopeType": reflect.TypeOf(""), "PlatformId": reflect.TypeOf(int64(0)), "TenantId": reflect.TypeOf(int64(0)), "MerchantId": reflect.TypeOf(int64(0))},
		},
		{
			name: "attribute group list row",
			typ:  reflect.TypeOf(ProductAttributeGroupListData{}),
			fields: map[string]reflect.Type{"ScopeType": reflect.TypeOf(""), "PlatformId": reflect.TypeOf(int64(0)), "TenantId": reflect.TypeOf(int64(0)), "MerchantId": reflect.TypeOf(int64(0))},
		},
		{
			name: "spec update/status/detail/list",
			typ:  reflect.TypeOf(UpdateProductSpecReq{}),
			fields: map[string]reflect.Type{"Scope": scopePtr},
		},
		{
			name: "spec status req",
			typ:  reflect.TypeOf(UpdateProductSpecStatusReq{}),
			fields: map[string]reflect.Type{"Scope": scopePtr},
		},
		{
			name: "spec detail resp",
			typ:  reflect.TypeOf(QueryProductSpecDetailResp{}),
			fields: map[string]reflect.Type{"ScopeType": reflect.TypeOf(""), "PlatformId": reflect.TypeOf(int64(0)), "TenantId": reflect.TypeOf(int64(0)), "MerchantId": reflect.TypeOf(int64(0))},
		},
		{
			name: "spec list row",
			typ:  reflect.TypeOf(ProductSpecListData{}),
			fields: map[string]reflect.Type{"ScopeType": reflect.TypeOf(""), "PlatformId": reflect.TypeOf(int64(0)), "TenantId": reflect.TypeOf(int64(0)), "MerchantId": reflect.TypeOf(int64(0))},
		},
		{
			name: "spec value update/status/detail/list",
			typ:  reflect.TypeOf(UpdateProductSpecValueReq{}),
			fields: map[string]reflect.Type{"Scope": scopePtr},
		},
		{
			name: "spec value status req",
			typ:  reflect.TypeOf(UpdateProductSpecValueStatusReq{}),
			fields: map[string]reflect.Type{"Scope": scopePtr},
		},
		{
			name: "spec value detail resp",
			typ:  reflect.TypeOf(QueryProductSpecValueDetailResp{}),
			fields: map[string]reflect.Type{"ScopeType": reflect.TypeOf(""), "PlatformId": reflect.TypeOf(int64(0)), "TenantId": reflect.TypeOf(int64(0)), "MerchantId": reflect.TypeOf(int64(0))},
		},
		{
			name: "spec value list row",
			typ:  reflect.TypeOf(ProductSpecValueListData{}),
			fields: map[string]reflect.Type{"ScopeType": reflect.TypeOf(""), "PlatformId": reflect.TypeOf(int64(0)), "TenantId": reflect.TypeOf(int64(0)), "MerchantId": reflect.TypeOf(int64(0))},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for field, want := range tc.fields {
				requireField(t, tc.typ, field, want)
			}
		})
	}
}
