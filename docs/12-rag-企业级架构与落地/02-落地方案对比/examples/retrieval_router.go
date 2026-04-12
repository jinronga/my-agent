package comparison

import (
	"context"
	"strings"
)

type QueryIntent string

const (
	IntentKeyword    QueryIntent = "keyword"
	IntentSemantic   QueryIntent = "semantic"
	IntentStructured QueryIntent = "structured"
	IntentCompliance QueryIntent = "compliance"
)

type RetrievalMode string

const (
	ModeLexical RetrievalMode = "lexical"
	ModeDense   RetrievalMode = "dense"
	ModeHybrid  RetrievalMode = "hybrid"
	ModeSQL     RetrievalMode = "sql"
)

type Query struct {
	Text      string
	Scene     string
	NeedFresh bool
}

type Plan struct {
	Intent      QueryIntent
	Mode        RetrievalMode
	NeedRerank  bool
	NeedLiveAPI bool
}

type Router struct{}

func (Router) BuildPlan(_ context.Context, q Query) Plan {
	lower := strings.ToLower(q.Text)

	switch {
	case q.Scene == "compliance":
		return Plan{
			Intent:      IntentCompliance,
			Mode:        ModeHybrid,
			NeedRerank:  true,
			NeedLiveAPI: false,
		}
	case looksStructured(lower):
		return Plan{
			Intent:      IntentStructured,
			Mode:        ModeSQL,
			NeedRerank:  false,
			NeedLiveAPI: true,
		}
	case looksKeyword(lower):
		return Plan{
			Intent:      IntentKeyword,
			Mode:        ModeLexical,
			NeedRerank:  true,
			NeedLiveAPI: q.NeedFresh,
		}
	default:
		return Plan{
			Intent:      IntentSemantic,
			Mode:        ModeHybrid,
			NeedRerank:  true,
			NeedLiveAPI: q.NeedFresh,
		}
	}
}

func looksStructured(q string) bool {
	return strings.Contains(q, "订单") ||
		strings.Contains(q, "审批") ||
		strings.Contains(q, "status") ||
		strings.Contains(q, "金额")
}

func looksKeyword(q string) bool {
	keywords := []string{"错误码", "api", "接口", "条款", "编号", "函数", "类"}
	for _, keyword := range keywords {
		if strings.Contains(q, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}
