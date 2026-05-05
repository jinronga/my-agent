package types

type QueryRequest struct {
	TenantID   string   `json:"-"`
	UserID     string   `json:"-"`
	Roles      []string `json:"-"`
	TraceID    string   `json:"-"`
	Query      string   `json:"query"`
	Scene      string   `json:"scene"`
	MaxResults int      `json:"max_results"`
}

type Citation struct {
	DocID   string  `json:"doc_id"`
	ChunkID string  `json:"chunk_id"`
	Title   string  `json:"title"`
	URL     string  `json:"url,omitempty"`
	Score   float64 `json:"score"`
}

type QueryResponse struct {
	Answer    string     `json:"answer"`
	Citations []Citation `json:"citations"`
	TraceID   string     `json:"trace_id"`
	Refused   bool       `json:"refused"`
	Reason    string     `json:"reason,omitempty"`
}

type Candidate struct {
	DocID       string
	ChunkID     string
	TenantID    string
	Title       string
	Content     string
	URL         string
	Score       float64
	AllowedRole string
}
