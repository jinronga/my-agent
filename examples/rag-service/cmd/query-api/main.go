package main

import (
	"log"
	"net/http"
	"time"

	"rag-service/internal/app"
	"rag-service/internal/httpapi"
	"rag-service/internal/retrieval"
)

func main() {
	queryService := app.NewService(retrieval.NewDemoRetriever())
	server := httpapi.Server{
		QueryService:   queryService,
		RequestTimeout: 8 * time.Second,
	}

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           server.Routes(),
		ReadHeaderTimeout: 3 * time.Second,
	}

	log.Printf("query-api listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
