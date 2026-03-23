package scope

import "testing"

func TestProductESSyncPayloadCarriesEventMeta(t *testing.T) {
	current := DefaultScope(1, 10, 301)
	meta := ProductEventMeta{
		Action:     "pms.product_spu.publish_status",
		ActorID:    2001,
		ActorName:  "reviewer",
		OccurredAt: "2026-03-23T21:52:00+08:00",
		Version:    1742737920000,
	}

	payload := NewProductESSyncPayload(123, current, "trace-1", meta)
	if payload.Action != meta.Action || payload.ActorID != meta.ActorID || payload.ActorName != meta.ActorName {
		t.Fatalf("expected event meta to be copied, got %+v", payload)
	}
	if payload.Version != meta.Version || payload.OccurredAt != meta.OccurredAt {
		t.Fatalf("expected version/timestamp to be copied, got %+v", payload)
	}
}

func TestProductESDeletePayloadDeduplicatesIDsAndCarriesEventMeta(t *testing.T) {
	current := DefaultScope(1, 10, 301)
	meta := ProductEventMeta{
		Action:     "pms.product_spu.verify_status",
		ActorID:    2002,
		ActorName:  "operator",
		OccurredAt: "2026-03-23T21:53:00+08:00",
		Version:    1742737980000,
	}

	payload := NewProductESDeletePayload([]int64{9, 9, 0, 10}, current, "trace-2", meta)
	if len(payload.IDs) != 2 || payload.IDs[0] != 9 || payload.IDs[1] != 10 {
		t.Fatalf("expected unique positive ids, got %+v", payload.IDs)
	}
	if payload.Action != meta.Action || payload.ActorID != meta.ActorID || payload.Version != meta.Version {
		t.Fatalf("expected event meta to be copied, got %+v", payload)
	}
}
