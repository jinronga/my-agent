package app

import (
	"context"
	"fmt"
	"strings"

	"rag-service/internal/policy"
	"rag-service/internal/types"
)

type Retriever interface {
	Retrieve(ctx context.Context, req types.QueryRequest, limit int) ([]types.Candidate, error)
}

type Service struct {
	retriever Retriever
}

func NewService(retriever Retriever) *Service {
	return &Service{retriever: retriever}
}

func (s *Service) HandleQuery(ctx context.Context, req types.QueryRequest) (types.QueryResponse, error) {
	if err := policy.NormalizeAndValidateQuery(&req); err != nil {
		return types.QueryResponse{}, err
	}

	candidates, err := s.retriever.Retrieve(ctx, req, req.MaxResults)
	if err != nil {
		return types.QueryResponse{}, err
	}
	if len(candidates) == 0 {
		return types.QueryResponse{
			Answer:  "没有找到当前用户有权限访问的可靠证据，已拒答。",
			TraceID: req.TraceID,
			Refused: true,
			Reason:  "no_authorized_evidence",
		}, nil
	}

	return types.QueryResponse{
		Answer:    groundedAnswer(req, candidates),
		Citations: citationsFromCandidates(candidates),
		TraceID:   req.TraceID,
		Refused:   false,
	}, nil
}

func groundedAnswer(req types.QueryRequest, candidates []types.Candidate) string {
	var b strings.Builder
	b.WriteString("基于当前可访问证据，建议如下：\n")
	for i, candidate := range candidates {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, candidate.Content))
	}
	b.WriteString("涉及对外发送、删除、审批或生产变更时，应先进入人工确认。")
	if req.Scene != "" {
		b.WriteString(fmt.Sprintf("\n场景：%s", req.Scene))
	}
	return b.String()
}

func citationsFromCandidates(candidates []types.Candidate) []types.Citation {
	citations := make([]types.Citation, 0, len(candidates))
	for _, candidate := range candidates {
		citations = append(citations, types.Citation{
			DocID:   candidate.DocID,
			ChunkID: candidate.ChunkID,
			Title:   candidate.Title,
			URL:     candidate.URL,
			Score:   candidate.Score,
		})
	}
	return citations
}
