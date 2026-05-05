package retrieval

import (
	"context"
	"sort"
	"strings"
	"unicode"

	"rag-service/internal/policy"
	"rag-service/internal/types"
)

type MemoryRetriever struct {
	docs []types.Candidate
}

func NewMemoryRetriever(docs []types.Candidate) *MemoryRetriever {
	return &MemoryRetriever{docs: docs}
}

func NewDemoRetriever() *MemoryRetriever {
	return NewMemoryRetriever([]types.Candidate{
		{
			DocID:       "support-handbook",
			ChunkID:     "support-handbook:complaint",
			TenantID:    "tenant_a",
			Title:       "客服投诉处理手册",
			Content:     "客户投诉处理应先查询客户记录和历史工单，生成处理建议，不应自动关闭工单。",
			URL:         "https://example.internal/support/handbook",
			Score:       0.90,
			AllowedRole: "support_agent",
		},
		{
			DocID:       "security-policy",
			ChunkID:     "security-policy:external-send",
			TenantID:    "tenant_a",
			Title:       "外发安全策略",
			Content:     "对外发送客户信息前必须经过人工确认，并记录审批人和 trace id。",
			URL:         "https://example.internal/security/external-send",
			Score:       0.88,
			AllowedRole: "support_agent",
		},
		{
			DocID:       "engineering-runbook",
			ChunkID:     "engineering-runbook:incident",
			TenantID:    "tenant_a",
			Title:       "工程事故处理手册",
			Content:     "生产事故处理要先止血、保留 trace、回滚最近变更，再将失败样本加入回归集。",
			URL:         "https://example.internal/engineering/runbook",
			Score:       0.86,
			AllowedRole: "engineer",
		},
	})
}

func (r *MemoryRetriever) Retrieve(_ context.Context, req types.QueryRequest, limit int) ([]types.Candidate, error) {
	queryTerms := tokenize(req.Query)
	candidates := make([]types.Candidate, 0, len(r.docs))

	for _, doc := range r.docs {
		if doc.TenantID != req.TenantID {
			continue
		}
		if !policy.HasRole(req.Roles, doc.AllowedRole) {
			continue
		}

		score := scoreCandidate(doc, queryTerms)
		if score <= 0 {
			continue
		}
		doc.Score = score
		candidates = append(candidates, doc)
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	return candidates, nil
}

func tokenize(q string) []string {
	fields := strings.Fields(strings.ToLower(q))
	if len(fields) > 0 {
		return fields
	}
	terms := make([]string, 0, len(q))
	for _, r := range strings.ToLower(q) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			terms = append(terms, string(r))
		}
	}
	return terms
}

func scoreCandidate(doc types.Candidate, terms []string) float64 {
	text := strings.ToLower(doc.Title + " " + doc.Content)
	score := 0.0
	for _, term := range terms {
		if strings.Contains(text, term) {
			score += 1
		}
	}

	if score == 0 {
		return 0
	}
	return score + doc.Score
}
