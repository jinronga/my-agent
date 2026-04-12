package operations

import (
	"context"
	"fmt"
	"log"
	"time"
)

type Snapshot struct {
	DatasetVersion string
	ChunkCount     int
}

type SearchAdmin interface {
	CreateIndex(ctx context.Context, indexName string) error
	Backfill(ctx context.Context, indexName string, snapshot Snapshot) error
	Validate(ctx context.Context, indexName string, snapshot Snapshot) error
	ShadowRead(ctx context.Context, oldIndex, newIndex string) error
	UpdateAlias(ctx context.Context, alias, newIndex string) error
}

type RebuildJob struct {
	Admin SearchAdmin
	Alias string
	Now   func() time.Time
}

func (j RebuildJob) Run(ctx context.Context, currentIndex string, snapshot Snapshot) (string, error) {
	if j.Now == nil {
		j.Now = time.Now
	}

	nextIndex := fmt.Sprintf("idx:rag:prod:%s", j.Now().Format("20060102-150405"))
	log.Printf("creating next index %s", nextIndex)

	if err := j.Admin.CreateIndex(ctx, nextIndex); err != nil {
		return "", err
	}
	if err := j.Admin.Backfill(ctx, nextIndex, snapshot); err != nil {
		return "", err
	}
	if err := j.Admin.Validate(ctx, nextIndex, snapshot); err != nil {
		return "", err
	}
	if err := j.Admin.ShadowRead(ctx, currentIndex, nextIndex); err != nil {
		return "", err
	}
	if err := j.Admin.UpdateAlias(ctx, j.Alias, nextIndex); err != nil {
		return "", err
	}

	log.Printf("alias %s switched from %s to %s", j.Alias, currentIndex, nextIndex)
	return nextIndex, nil
}
