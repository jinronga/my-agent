package app

import (
	"context"
	"testing"

	"rag-service/internal/retrieval"
	"rag-service/internal/types"
)

func TestHandleQueryReturnsGroundedAnswer(t *testing.T) {
	service := NewService(retrieval.NewDemoRetriever())

	resp, err := service.HandleQuery(context.Background(), types.QueryRequest{
		TenantID: "tenant_a",
		UserID:   "u_001",
		Roles:    []string{"support_agent"},
		Query:    "客户投诉",
		TraceID:  "trace_test",
	})
	if err != nil {
		t.Fatalf("expected no error: %v", err)
	}
	if resp.Refused {
		t.Fatalf("expected answer, got refusal: %#v", resp)
	}
	if len(resp.Citations) == 0 {
		t.Fatal("expected citations")
	}
	if resp.TraceID != "trace_test" {
		t.Fatalf("expected trace id to be propagated, got %q", resp.TraceID)
	}
}

func TestHandleQueryRefusesWithoutAuthorizedEvidence(t *testing.T) {
	service := NewService(retrieval.NewDemoRetriever())

	resp, err := service.HandleQuery(context.Background(), types.QueryRequest{
		TenantID: "tenant_a",
		UserID:   "u_001",
		Roles:    []string{"viewer"},
		Query:    "客户投诉",
	})
	if err != nil {
		t.Fatalf("expected no error: %v", err)
	}
	if !resp.Refused {
		t.Fatalf("expected refusal, got %#v", resp)
	}
}
