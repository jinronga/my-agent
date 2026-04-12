package architecture

import (
	"context"
	"errors"
	"time"
)

type SourceRecord struct {
	TenantID     string
	SourceSystem string
	ObjectType   string
	ObjectID     string
	Version      string
	UpdatedAt    time.Time
	Deleted      bool
	ACL          ACL
	Body         []byte
	Metadata     map[string]string
}

type ACL struct {
	Users []string
	Roles []string
	Tags  []string
}

type Chunk struct {
	ChunkID      string
	DocumentID   string
	Content      string
	Embedding    []float32
	Metadata     map[string]string
	ACL          ACL
	UpdatedAt    time.Time
	IndexVersion string
}

type QueryContext struct {
	TenantID   string
	UserID     string
	Roles      []string
	Query      string
	Scene      string
	MaxResults int
}

type Candidate struct {
	ChunkID   string
	Content   string
	Score     float64
	Citations []string
	Metadata  map[string]string
}

type GenerationRequest struct {
	Query      QueryContext
	Candidates []Candidate
}

type GenerationResult struct {
	Answer    string
	Citations []string
	NeedAgent bool
}

type AgentTask struct {
	Query     QueryContext
	Answer    string
	Citations []string
}

type TaskResult struct {
	Status    string
	Output    string
	ToolCalls []string
}

type SourceConnector interface {
	PullChanges(ctx context.Context, cursor string) ([]SourceRecord, string, error)
}

type DocumentProcessor interface {
	Process(ctx context.Context, record SourceRecord) ([]Chunk, error)
}

type IndexWriter interface {
	UpsertChunks(ctx context.Context, chunks []Chunk) error
	DeleteDocument(ctx context.Context, tenantID, documentID string) error
}

type Retriever interface {
	Retrieve(ctx context.Context, query QueryContext) ([]Candidate, error)
}

type Generator interface {
	Generate(ctx context.Context, req GenerationRequest) (GenerationResult, error)
}

type Orchestrator interface {
	Run(ctx context.Context, task AgentTask) (TaskResult, error)
}

type IngestionService struct {
	Connector SourceConnector
	Processor DocumentProcessor
	Writer    IndexWriter
}

func (s IngestionService) SyncOnce(ctx context.Context, cursor string) (string, error) {
	records, nextCursor, err := s.Connector.PullChanges(ctx, cursor)
	if err != nil {
		return cursor, err
	}

	for _, record := range records {
		documentID := record.SourceSystem + ":" + record.ObjectType + ":" + record.ObjectID
		if record.Deleted {
			if err := s.Writer.DeleteDocument(ctx, record.TenantID, documentID); err != nil {
				return cursor, err
			}
			continue
		}

		chunks, err := s.Processor.Process(ctx, record)
		if err != nil {
			return cursor, err
		}
		if len(chunks) == 0 {
			return cursor, errors.New("processor returned no chunks")
		}
		if err := s.Writer.UpsertChunks(ctx, chunks); err != nil {
			return cursor, err
		}
	}

	return nextCursor, nil
}

type QueryService struct {
	Retriever    Retriever
	Generator    Generator
	Orchestrator Orchestrator
}

func (s QueryService) HandleQuery(ctx context.Context, query QueryContext) (string, error) {
	candidates, err := s.Retriever.Retrieve(ctx, query)
	if err != nil {
		return "", err
	}

	result, err := s.Generator.Generate(ctx, GenerationRequest{
		Query:      query,
		Candidates: candidates,
	})
	if err != nil {
		return "", err
	}

	if !result.NeedAgent {
		return result.Answer, nil
	}

	taskResult, err := s.Orchestrator.Run(ctx, AgentTask{
		Query:     query,
		Answer:    result.Answer,
		Citations: result.Citations,
	})
	if err != nil {
		return "", err
	}

	return taskResult.Output, nil
}
