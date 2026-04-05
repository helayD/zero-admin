package channeltemplate

import "testing"

func TestNormalizeTargetCodeSupportsCanonicalChannelAlias(t *testing.T) {
	targetCode, err := NormalizeTargetCode(TemplateTypeChannel, "mini-program")
	if err != nil {
		t.Fatalf("NormalizeTargetCode returned error: %v", err)
	}
	if targetCode != TargetMiniProgram {
		t.Fatalf("unexpected target code: %s", targetCode)
	}
}

func TestNormalizeTargetCodeRejectsUnsupportedChannelTarget(t *testing.T) {
	_, err := NormalizeTargetCode(TemplateTypeChannel, "app")
	if err == nil {
		t.Fatal("expected unsupported channel target to fail")
	}
}

func TestValidateStatusTransition(t *testing.T) {
	if err := ValidateStatusTransition(StatusDraft, StatusEnabled); err != nil {
		t.Fatalf("expected draft -> enabled to pass, got %v", err)
	}
	if err := ValidateStatusTransition(StatusEnabled, StatusDisabled); err != nil {
		t.Fatalf("expected enabled -> disabled to pass, got %v", err)
	}
	if err := ValidateStatusTransition(StatusArchived, StatusEnabled); err == nil {
		t.Fatal("expected archived -> enabled to fail")
	}
}

func TestValidateSecretRefMapRejectsPlaintextSecret(t *testing.T) {
	err := ValidateSecretRefMap(map[string]string{
		"appSecretRef": "plain-text-secret",
	})
	if err == nil {
		t.Fatal("expected plaintext secret ref to fail")
	}
}

func TestValidateSecretRefMapAcceptsCredentialReference(t *testing.T) {
	err := ValidateSecretRefMap(map[string]string{
		"appSecretRef": "credential://sys/channel/mini-program/app-secret",
	})
	if err != nil {
		t.Fatalf("ValidateSecretRefMap returned error: %v", err)
	}
}

func TestValidateIntentContractsRequiresSemanticRouteKey(t *testing.T) {
	err := ValidateIntentContracts(TargetH5, map[string]IntentContract{
		"home": {
			Intent:   "home",
			RouteKey: "https://mall.example.com/#/home?tab=1",
		},
	})
	if err == nil {
		t.Fatal("expected raw URL route key to fail")
	}
}

func TestValidateIntentContractsRequiresContractsForMessageTargets(t *testing.T) {
	err := ValidateIntentContracts(TargetMemberMessage, nil)
	if err == nil {
		t.Fatal("expected member_message contracts to be required")
	}
}

func TestValidateIntentContractsAcceptsSemanticIntentRoutes(t *testing.T) {
	err := ValidateIntentContracts(TargetMiniProgram, map[string]IntentContract{
		"product_detail": {
			Intent:   "product_detail",
			RouteKey: "product_detail",
			Params: map[string]string{
				"id": "skuId",
			},
		},
	})
	if err != nil {
		t.Fatalf("ValidateIntentContracts returned error: %v", err)
	}
}
