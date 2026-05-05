package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

type QueryService interface {
	HandleQuery(ctx context.Context, query QueryRequest) (QueryResponse, error)
}

type QueryRequest struct {
	TenantID string
	UserID   string
	Query    string
	Scene    string
}

type QueryResponse struct {
	Answer    string
	Citations []string
}

type Server struct {
	QueryService QueryService
}

type staticQueryService struct{}

func (staticQueryService) HandleQuery(_ context.Context, query QueryRequest) (QueryResponse, error) {
	return QueryResponse{
		Answer: "example answer for query: " + query.Query,
		Citations: []string{
			"example-doc:chunk-1",
		},
	}, nil
}

func (s Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/v1/query", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
		defer cancel()

		resp, err := s.QueryService.HandleQuery(ctx, QueryRequest{
			TenantID: r.Header.Get("X-Tenant-ID"),
			UserID:   r.Header.Get("X-User-ID"),
			Query:    r.URL.Query().Get("q"),
			Scene:    r.URL.Query().Get("scene"),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, _ = w.Write([]byte(resp.Answer))
	})
	return mux
}

func main() {
	server := Server{
		QueryService: staticQueryService{},
	}

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           server.routes(),
		ReadHeaderTimeout: 3 * time.Second,
	}

	log.Printf("query-api listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
