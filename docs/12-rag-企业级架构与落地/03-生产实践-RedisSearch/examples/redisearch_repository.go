package redisearch

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

type Chunk struct {
	Key            string
	TenantID       string
	DocID          string
	ChunkID        string
	SourceSystem   string
	DocType        string
	Title          string
	Body           string
	Path           string
	Tags           []string
	ACLRoles       []string
	ACLUsers       []string
	UpdatedAtUnix  int64
	DatasetVersion string
	IndexVersion   string
	Embedding      []float32
}

type Repository struct {
	client *redis.Client
	index  string
}

func NewRepository(client *redis.Client, index string) *Repository {
	return &Repository{client: client, index: index}
}

func (r *Repository) CreateIndex(ctx context.Context, prefix string, dim int) error {
	args := []interface{}{
		"FT.CREATE", r.index,
		"ON", "HASH",
		"PREFIX", "1", prefix,
		"SCHEMA",
		"tenant_id", "TAG",
		"doc_id", "TAG",
		"chunk_id", "TAG",
		"source_system", "TAG",
		"doc_type", "TAG",
		"title", "TEXT", "WEIGHT", "3.0",
		"body", "TEXT", "WEIGHT", "1.0",
		"path", "TEXT", "WEIGHT", "1.5",
		"tags", "TAG", "SEPARATOR", ",",
		"acl_roles", "TAG", "SEPARATOR", ",",
		"acl_users", "TAG", "SEPARATOR", ",",
		"updated_at", "NUMERIC", "SORTABLE",
		"dataset_version", "TAG",
		"index_version", "TAG",
		"embedding", "VECTOR", "HNSW", "10",
		"TYPE", "FLOAT32",
		"DIM", strconv.Itoa(dim),
		"DISTANCE_METRIC", "COSINE",
		"M", "16",
		"EF_CONSTRUCTION", "200",
		"EF_RUNTIME", "64",
	}
	return r.client.Do(ctx, args...).Err()
}

func (r *Repository) UpsertChunks(ctx context.Context, chunks []Chunk) error {
	pipe := r.client.Pipeline()
	for _, chunk := range chunks {
		fields := map[string]interface{}{
			"tenant_id":       chunk.TenantID,
			"doc_id":          chunk.DocID,
			"chunk_id":        chunk.ChunkID,
			"source_system":   chunk.SourceSystem,
			"doc_type":        chunk.DocType,
			"title":           chunk.Title,
			"body":            chunk.Body,
			"path":            chunk.Path,
			"tags":            strings.Join(chunk.Tags, ","),
			"acl_roles":       strings.Join(chunk.ACLRoles, ","),
			"acl_users":       strings.Join(chunk.ACLUsers, ","),
			"updated_at":      chunk.UpdatedAtUnix,
			"dataset_version": chunk.DatasetVersion,
			"index_version":   chunk.IndexVersion,
			"embedding":       float32SliceToBytes(chunk.Embedding),
		}
		pipe.HSet(ctx, chunk.Key, fields)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r *Repository) HybridSearch(ctx context.Context, tenantID string, roles []string, docTypes []string, queryVector []byte, k int) ([]map[string]interface{}, error) {
	roleFilter := strings.Join(roles, "|")
	docTypeFilter := strings.Join(docTypes, "|")
	query := fmt.Sprintf("(@tenant_id:{%s} @acl_roles:{%s} @doc_type:{%s})=>[KNN %d @embedding $query_vec AS vector_score]", tenantID, roleFilter, docTypeFilter, k)

	args := []interface{}{
		"FT.SEARCH", r.index, query,
		"PARAMS", "2", "query_vec", queryVector,
		"SORTBY", "vector_score",
		"RETURN", "7",
		"doc_id", "chunk_id", "title", "body", "path", "updated_at", "vector_score",
		"DIALECT", "2",
	}

	raw, err := r.client.Do(ctx, args...).Result()
	if err != nil {
		return nil, err
	}
	return decodeSearchResult(raw), nil
}

func (r *Repository) UpdateAlias(ctx context.Context, alias, nextIndex string) error {
	return r.client.Do(ctx, "FT.ALIASUPDATE", alias, nextIndex).Err()
}

func float32SliceToBytes(values []float32) []byte {
	result := make([]byte, 0, len(values)*4)
	for _, value := range values {
		bits := uint32FromFloat32(value)
		result = append(result,
			byte(bits),
			byte(bits>>8),
			byte(bits>>16),
			byte(bits>>24),
		)
	}
	return result
}

func uint32FromFloat32(v float32) uint32 {
	return math.Float32bits(v)
}

func decodeSearchResult(raw interface{}) []map[string]interface{} {
	values, ok := raw.([]interface{})
	if !ok || len(values) < 2 {
		return nil
	}

	results := make([]map[string]interface{}, 0, (len(values)-1)/2)
	for i := 1; i+1 < len(values); i += 2 {
		item := map[string]interface{}{
			"key": values[i],
		}

		fields, ok := values[i+1].([]interface{})
		if !ok {
			results = append(results, item)
			continue
		}
		for j := 0; j+1 < len(fields); j += 2 {
			name := fmt.Sprint(fields[j])
			item[name] = fields[j+1]
		}
		results = append(results, item)
	}
	return results
}
