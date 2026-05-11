// Story 10.11 / Task 4.1 / Q1 spike
// 临时验证 github.com/FISCO-BCOS/go-sdk/v3 可被本仓库 import 且 cgo 链接通过。
// Phase 2 Task 5 完成后删除本文件。

//go:build cgo_spike

package fisco

import (
	"testing"

	bcosclient "github.com/FISCO-BCOS/go-sdk/v3/client"
)

func TestCgoSpike_GoSdkV3_Imports(t *testing.T) {
	cfg := &bcosclient.Config{
		IsSMCrypto: false,
		GroupID:    "group0",
		Host:       "127.0.0.1",
		Port:       20200,
		DisableSsl: true,
	}
	if cfg == nil {
		t.Fatal("nil cfg")
	}
}
