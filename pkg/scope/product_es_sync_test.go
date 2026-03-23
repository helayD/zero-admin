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

	body, err := json.Marshal(NewProductESSyncPayload(99, current, "trace-1"))
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
	if payload.TraceID != "trace-1" {
		t.Fatalf("unexpected trace id: %q", payload.TraceID)
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

func TestDecodeProductESDeletePayload(t *testing.T) {
	current, err := NormalizeGovernanceScope(SubjectTypeTenant, 1, 18, 0)
	if err != nil {
		t.Fatalf("NormalizeGovernanceScope returned error: %v", err)
	}

	body, err := json.Marshal(NewProductESDeletePayload([]int64{9, 9, 10}, current, "trace-delete"))
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	payload, resolved, err := DecodeProductESDeletePayload(body)
	if err != nil {
		t.Fatalf("DecodeProductESDeletePayload returned error: %v", err)
	}
	if payload.TraceID != "trace-delete" {
		t.Fatalf("unexpected trace id: %q", payload.TraceID)
	}
	if len(payload.IDs) != 2 || payload.IDs[0] != 9 || payload.IDs[1] != 10 {
		t.Fatalf("unexpected ids: %+v", payload.IDs)
	}
	if !resolved.SameScope(current) {
		t.Fatalf("unexpected scope: got %+v want %+v", resolved, current)
	}
}

func TestDecodeProductESDeletePayloadRejectsMissingScope(t *testing.T) {
	body := []byte(`{"ids":[1,2]}`)
	if _, _, err := DecodeProductESDeletePayload(body); err == nil {
		t.Fatalf("expected payload without scope to fail")
	}
}
