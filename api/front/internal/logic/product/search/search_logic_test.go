package search

import (
	"context"
	"errors"
	"testing"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/rpc/search/search_client"
	"google.golang.org/grpc"
)

type fakeSearchClient struct {
	searchFn func(context.Context, *search_client.SearchReq, ...grpc.CallOption) (*search_client.SearchResp, error)
}

func (f *fakeSearchClient) Create(context.Context, *search_client.CreateReq, ...grpc.CallOption) (*search_client.CreateResp, error) {
	return &search_client.CreateResp{}, nil
}

func (f *fakeSearchClient) Delete(context.Context, *search_client.DeleteReq, ...grpc.CallOption) (*search_client.DeleteResp, error) {
	return &search_client.DeleteResp{}, nil
}

func (f *fakeSearchClient) SearchSimple(context.Context, *search_client.SearchSimpleReq, ...grpc.CallOption) (*search_client.SearchResp, error) {
	return &search_client.SearchResp{}, nil
}

func (f *fakeSearchClient) Search(ctx context.Context, in *search_client.SearchReq, opts ...grpc.CallOption) (*search_client.SearchResp, error) {
	return f.searchFn(ctx, in, opts...)
}

func (f *fakeSearchClient) Recommend(context.Context, *search_client.RecommendReq, ...grpc.CallOption) (*search_client.SearchResp, error) {
	return &search_client.SearchResp{}, nil
}

func (f *fakeSearchClient) SearchRelated(context.Context, *search_client.SearchRelatedReq, ...grpc.CallOption) (*search_client.SearchRelatedResp, error) {
	return &search_client.SearchRelatedResp{}, nil
}

func TestSearchReturnsFriendlyEmptyResponseWhenRpcFails(t *testing.T) {
	logic := NewSearchLogic(context.Background(), &svc.ServiceContext{
		SearchClient: &fakeSearchClient{
			searchFn: func(context.Context, *search_client.SearchReq, ...grpc.CallOption) (*search_client.SearchResp, error) {
				return nil, errors.New("es search error: index_not_found_exception")
			},
		},
	})

	resp, err := logic.Search(&types.SearchReq{Keyword: "海澜之家", PageNum: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if resp == nil || !resp.Empty || resp.Total != 0 || len(resp.Data) != 0 {
		t.Fatalf("Search response = %+v, want empty response", resp)
	}
	if resp.EmptyHint != "搜索服务暂时不可用，请稍后再试" {
		t.Fatalf("EmptyHint = %q", resp.EmptyHint)
	}
}
