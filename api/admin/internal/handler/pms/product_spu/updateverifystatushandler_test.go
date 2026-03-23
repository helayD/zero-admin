package product_spu

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func TestUpdateProductSpuStatusReqParsesDetailFromJSONBody(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/pms/product/updateVerifyStatus",
		bytes.NewBufferString(`{"ids":[2001],"status":1,"detail":"审核通过"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	var parsed types.UpdateProductSpuStatusReq
	if err := httpx.Parse(req, &parsed); err != nil {
		t.Fatalf("parse request failed: %v", err)
	}
	if parsed.Detail != "审核通过" {
		t.Fatalf("expected detail to be parsed from json body, got %q", parsed.Detail)
	}
	if len(parsed.Ids) != 1 || parsed.Ids[0] != 2001 || parsed.Status != 1 {
		t.Fatalf("unexpected parsed payload: %+v", parsed)
	}
}
