package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"rag-service/internal/observability"
	"rag-service/internal/policy"
	"rag-service/internal/types"
)

type QueryService interface {
	HandleQuery(context.Context, types.QueryRequest) (types.QueryResponse, error)
}

type Server struct {
	QueryService   QueryService
	RequestTimeout time.Duration
}

func (s Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("POST /v1/query", s.handleQuery)
	return mux
}

func (s Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if s.RequestTimeout > 0 {
		var cancel func()
		ctx, cancel = contextWithTimeout(ctx, s.RequestTimeout)
		defer cancel()
	}

	var req types.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err)
		return
	}

	req.TenantID = r.Header.Get("X-Tenant-ID")
	req.UserID = r.Header.Get("X-User-ID")
	req.Roles = splitRoles(r.Header.Get("X-Roles"))
	req.TraceID = observability.TraceIDFromRequest(r)

	resp, err := s.QueryService.HandleQuery(ctx, req)
	if err != nil {
		status := http.StatusInternalServerError
		code := "internal_error"
		if isValidationError(err) {
			status = http.StatusBadRequest
			code = "invalid_request"
		}
		writeError(w, status, code, err)
		return
	}

	w.Header().Set(observability.TraceHeader, resp.TraceID)
	writeJSON(w, http.StatusOK, resp)
}

func splitRoles(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	roles := make([]string, 0, len(parts))
	for _, part := range parts {
		if role := strings.TrimSpace(part); role != "" {
			roles = append(roles, role)
		}
	}
	return roles
}

func isValidationError(err error) bool {
	return errors.Is(err, policy.ErrMissingTenant) ||
		errors.Is(err, policy.ErrMissingUser) ||
		errors.Is(err, policy.ErrMissingQuery) ||
		errors.Is(err, policy.ErrQueryTooLong)
}

func writeError(w http.ResponseWriter, status int, code string, err error) {
	writeJSON(w, status, map[string]string{
		"error":   code,
		"message": err.Error(),
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
