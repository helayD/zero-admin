package config

import (
	"os"
	"testing"

	"github.com/zeromicro/go-zero/core/conf"
)

func TestAdminAPIYamlLoadsWithReservedSystemConfig(t *testing.T) {
	content, err := os.ReadFile("../../etc/admin-api.yaml")
	if err != nil {
		t.Fatalf("read admin-api.yaml failed: %v", err)
	}

	var c Config
	if err = conf.LoadFromYamlBytes(content, &c); err != nil {
		t.Fatalf("load admin-api.yaml failed: %v", err)
	}
}
