package policy

import (
	"errors"
	"strings"

	"rag-service/internal/types"
)

const (
	DefaultMaxResults = 3
	MaxResultsLimit   = 10
	MaxQueryLength    = 2000
)

var (
	ErrMissingTenant = errors.New("tenant id is required")
	ErrMissingUser   = errors.New("user id is required")
	ErrMissingQuery  = errors.New("query is required")
	ErrQueryTooLong  = errors.New("query is too long")
)

func NormalizeAndValidateQuery(req *types.QueryRequest) error {
	req.TenantID = strings.TrimSpace(req.TenantID)
	req.UserID = strings.TrimSpace(req.UserID)
	req.Query = strings.TrimSpace(req.Query)
	req.Scene = strings.TrimSpace(req.Scene)

	switch {
	case req.TenantID == "":
		return ErrMissingTenant
	case req.UserID == "":
		return ErrMissingUser
	case req.Query == "":
		return ErrMissingQuery
	case len([]rune(req.Query)) > MaxQueryLength:
		return ErrQueryTooLong
	}

	if req.MaxResults <= 0 {
		req.MaxResults = DefaultMaxResults
	}
	if req.MaxResults > MaxResultsLimit {
		req.MaxResults = MaxResultsLimit
	}
	return nil
}

func HasRole(userRoles []string, required string) bool {
	if required == "" {
		return true
	}
	for _, role := range userRoles {
		if strings.TrimSpace(role) == required {
			return true
		}
	}
	return false
}
