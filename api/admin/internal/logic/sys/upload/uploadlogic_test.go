package upload

import (
	"mime/multipart"
	"net/textproto"
	"strings"
	"testing"
)

func TestBuildObjectURLUsesConfiguredPublicURL(t *testing.T) {
	got := buildObjectURL("https://speed.maibanjk.com/", "oss-cn-shenzhen.aliyuncs.com", "mbjq", "uploads/20260427/a.png")
	want := "https://speed.maibanjk.com/uploads/20260427/a.png"
	if got != want {
		t.Fatalf("buildObjectURL() = %q, want %q", got, want)
	}
}

func TestBuildObjectURLFallsBackToBucketEndpoint(t *testing.T) {
	got := buildObjectURL("", "https://oss-cn-shenzhen.aliyuncs.com", "mbjq", "uploads/20260427/a.png")
	want := "https://mbjq.oss-cn-shenzhen.aliyuncs.com/uploads/20260427/a.png"
	if got != want {
		t.Fatalf("buildObjectURL() = %q, want %q", got, want)
	}
}

func TestBuildObjectKeyKeepsOnlySafeGeneratedNameAndExt(t *testing.T) {
	got := buildObjectKey(`C:\tmp\商品.JPG`)
	if !strings.HasPrefix(got, "uploads/") {
		t.Fatalf("object key should use uploads prefix, got %q", got)
	}
	if !strings.HasSuffix(got, ".jpg") {
		t.Fatalf("object key should keep lowercase extension, got %q", got)
	}
	if strings.Contains(got, "\\") || strings.Contains(got, "商品") {
		t.Fatalf("object key should not keep original path/name, got %q", got)
	}
}

func TestResolveContentType(t *testing.T) {
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="file"; filename="demo.png"`)
	part := &multipart.Part{Header: header}

	if got := resolveContentType(part); got != "image/png" {
		t.Fatalf("resolveContentType() = %q, want image/png", got)
	}
}
