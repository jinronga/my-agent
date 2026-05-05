package policy

import (
	"errors"
	"testing"

	"rag-service/internal/types"
)

func TestNormalizeAndValidateQuery(t *testing.T) {
	req := types.QueryRequest{
		TenantID: " tenant_a ",
		UserID:   " u_001 ",
		Query:    " hello ",
	}

	if err := NormalizeAndValidateQuery(&req); err != nil {
		t.Fatalf("expected valid request: %v", err)
	}
	if req.TenantID != "tenant_a" || req.UserID != "u_001" || req.Query != "hello" {
		t.Fatalf("request was not normalized: %#v", req)
	}
	if req.MaxResults != DefaultMaxResults {
		t.Fatalf("expected default max results, got %d", req.MaxResults)
	}
}

func TestNormalizeAndValidateQueryRequiresTenant(t *testing.T) {
	req := types.QueryRequest{UserID: "u_001", Query: "hello"}

	err := NormalizeAndValidateQuery(&req)
	if !errors.Is(err, ErrMissingTenant) {
		t.Fatalf("expected ErrMissingTenant, got %v", err)
	}
}
