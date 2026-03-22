package scope

import (
	"encoding/json"
	"testing"
)

func TestDecodeProductESSyncPayload(t *testing.T) {
	current, err := NormalizeGovernanceScope(SubjectTypeMerchant, 1, 18, 301)
	if err != nil {
		t.Fatalf("NormalizeGovernanceScope returned error: %v", err)
	}

	body, err := json.Marshal(NewProductESSyncPayload(99, current))
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	payload, resolved, err := DecodeProductESSyncPayload(body)
	if err != nil {
		t.Fatalf("DecodeProductESSyncPayload returned error: %v", err)
	}
	if payload.ID != 99 {
		t.Fatalf("unexpected payload id: %d", payload.ID)
	}
	if !resolved.SameScope(current) {
		t.Fatalf("unexpected scope: got %+v want %+v", resolved, current)
	}
}

func TestDecodeProductESSyncPayloadRejectsMissingScope(t *testing.T) {
	body := []byte(`{"id":1}`)
	if _, _, err := DecodeProductESSyncPayload(body); err == nil {
		t.Fatalf("expected payload without scope to fail")
	}
}
