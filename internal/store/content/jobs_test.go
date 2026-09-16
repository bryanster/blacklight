package content_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/bryanster/blacklight/internal/httpapi/apierr"
	"github.com/bryanster/blacklight/internal/store/content"
	"github.com/bryanster/blacklight/internal/store/storetest"
)

func TestJobSelectionAndListFilters(t *testing.T) {
	db := storetest.Migrated(t)
	r := content.NewJobs(db)
	ctx := t.Context()
	if _, ok, err := r.FindActive(ctx); err != nil || ok {
		t.Fatalf("empty active slot: %v, %v", ok, err)
	}
	if _, ok, err := r.NextQueued(ctx); err != nil || ok {
		t.Fatalf("empty queue: %v, %v", ok, err)
	}
	var jobs []content.Job
	for i, source := range []string{content.SourceIDAttack, content.SourceIDAtomic, content.SourceIDAttack, content.SourceIDAtomic} {
		job, err := r.Create(ctx, content.NewJob{SourceID: source, Kind: content.JobKindSync, CreatedBy: "admin"})
		if err != nil {
			t.Fatal(err)
		}
		// Explicit timestamps avoid relying on wall-clock granularity for FIFO.
		if err := db.Write(ctx, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, `UPDATE content.content_sync_job SET created_at = ? WHERE id = ?`, time.Date(2026, 1, 1, 0, i, 0, 0, time.UTC), job.ID)
			return err
		}); err != nil {
			t.Fatal(err)
		}
		jobs = append(jobs, job)
	}
	setStatus := func(index int, status content.JobStatus) {
		t.Helper()
		if _, err := r.Update(ctx, jobs[index].ID, content.JobUpdate{Status: status}); err != nil {
			t.Fatal(err)
		}
	}
	setStatus(0, content.JobStatusSucceeded)
	setStatus(1, content.JobStatusRunning)
	active, ok, err := r.FindActive(ctx)
	if err != nil || !ok || active.ID != jobs[1].ID {
		t.Fatalf("active selection: %+v, %v, %v", active, ok, err)
	}
	queued, ok, err := r.NextQueued(ctx)
	if err != nil || !ok || queued.ID != jobs[2].ID {
		t.Fatalf("FIFO selection: %+v, %v, %v", queued, ok, err)
	}
	for _, tc := range []struct {
		name   string
		filter content.ListFilter
		want   []string
	}{
		{"newest first and limit", content.ListFilter{Limit: 2}, []string{jobs[3].ID, jobs[2].ID}},
		{"source", content.ListFilter{SourceID: content.SourceIDAttack}, []string{jobs[2].ID, jobs[0].ID}},
		{"status", content.ListFilter{Status: content.JobStatusQueued}, []string{jobs[3].ID, jobs[2].ID}},
		{"combined", content.ListFilter{SourceID: content.SourceIDAtomic, Status: content.JobStatusQueued}, []string{jobs[3].ID}},
		{"no matches", content.ListFilter{Status: content.JobStatusFailed}, []string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			list, err := r.List(ctx, tc.filter)
			if err != nil {
				t.Fatal(err)
			}
			ids := make([]string, 0, len(list))
			for _, job := range list {
				ids = append(ids, job.ID)
			}
			if !reflect.DeepEqual(ids, tc.want) {
				t.Fatalf("listed IDs = %v, want %v", ids, tc.want)
			}
		})
	}
	setStatus(1, content.JobStatusCancelling)
	active, ok, err = r.FindActive(ctx)
	if err != nil || !ok || active.ID != jobs[1].ID {
		t.Fatalf("cancelling job released slot: %+v, %v, %v", active, ok, err)
	}
	for i := 1; i < len(jobs); i++ {
		setStatus(i, content.JobStatusSucceeded)
	}
	if _, ok, err := r.FindActive(ctx); err != nil || ok {
		t.Fatalf("terminal jobs retained active slot: %v, %v", ok, err)
	}
	if _, ok, err := r.NextQueued(ctx); err != nil || ok {
		t.Fatalf("terminal jobs remained queued: %v, %v", ok, err)
	}
}

func TestJobProgressPreservesCheckpointAndTimestamps(t *testing.T) {
	r := content.NewJobs(storetest.Migrated(t))
	ctx := t.Context()
	job, err := r.Create(ctx, content.NewJob{SourceID: content.SourceIDAttack, Kind: content.JobKindSync})
	if err != nil {
		t.Fatal(err)
	}
	started := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	checkpoint := json.RawMessage(`{"page":3,"cursor":"next"}`)
	if _, err := r.Update(ctx, job.ID, content.JobUpdate{Status: content.JobStatusRunning, StartedAt: started, Checkpoint: checkpoint}); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Update(ctx, job.ID, content.JobUpdate{Status: content.JobStatusRunning, Phase: "apply", ProgressCurrent: 4, ProgressTotal: 10}); err != nil {
		t.Fatal(err)
	}
	got, err := r.ByID(ctx, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	var saved, want map[string]any
	if err := json.Unmarshal(got.Checkpoint, &saved); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(checkpoint, &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(saved, want) || !got.StartedAt.Equal(started) || !got.FinishedAt.IsZero() || got.ProgressCurrent != 4 || got.ProgressTotal != 10 || got.Phase != "apply" {
		t.Fatalf("progress update lost saved state: %+v", got)
	}
	finished := started.Add(time.Minute)
	if _, err := r.Update(ctx, job.ID, content.JobUpdate{Status: content.JobStatusSucceeded, FinishedAt: finished, Checkpoint: json.RawMessage{}}); err != nil {
		t.Fatal(err)
	}
	got, err = r.ByID(ctx, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	var cleared map[string]any
	if err := json.Unmarshal(got.Checkpoint, &cleared); err != nil {
		t.Fatal(err)
	}
	if cleared == nil || len(cleared) != 0 || !got.StartedAt.Equal(started) || !got.FinishedAt.Equal(finished) {
		t.Fatalf("explicit reset or completion lost state: %+v", got)
	}
}

func TestJobInvalidCheckpointRollsBackUpdate(t *testing.T) {
	r := content.NewJobs(storetest.Migrated(t))
	ctx := t.Context()
	job, err := r.Create(ctx, content.NewJob{SourceID: content.SourceIDAtomic, Kind: content.JobKindSync})
	if err != nil {
		t.Fatal(err)
	}
	before, err := r.ByID(ctx, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Update(ctx, job.ID, content.JobUpdate{Status: content.JobStatusRunning, Phase: "apply", Checkpoint: json.RawMessage(`{"invalid":`)}); err == nil {
		t.Fatal("invalid checkpoint accepted")
	}
	after, err := r.ByID(ctx, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("failed update changed job: %+v, want %+v", after, before)
	}
	const missing = "01900000-0000-7000-e000-0000000000aa"
	_, err = r.Update(ctx, missing, content.JobUpdate{Status: content.JobStatusRunning})
	var typed *apierr.Error
	if !errors.As(err, &typed) || typed.Status() != 404 {
		t.Fatalf("missing update should return not found: %v", err)
	}
}
