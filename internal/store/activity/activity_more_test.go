package activity_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/bryanster/blacklight/internal/httpapi/apierr"
	"github.com/bryanster/blacklight/internal/store/activity"
	"github.com/bryanster/blacklight/internal/store/storetest"
)

func TestActivityValidateRejectsIncompleteEntries(t *testing.T) {
	repo := activity.New(storetest.Migrated(t))
	for _, tc := range []struct {
		name                       string
		verb, objectType, objectID string
	}{
		{"missing verb", " ", "session", "sess-1"},
		{"missing object type", "session.login", " ", "sess-1"},
		{"missing object id", "session.login", "session", " "},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := repo.Append(t.Context(), activity.Entry{Verb: tc.verb, ObjectType: tc.objectType, ObjectID: tc.objectID})
			if err == nil || errors.Is(err, &apierr.Error{}) {
				t.Fatalf("expected plain validation error, got %v", err)
			}
		})
	}
}

func TestActivityAppendNormalisesAndStoresDelta(t *testing.T) {
	repo := activity.New(storetest.Migrated(t))
	ctx := t.Context()
	at := time.Date(2026, 3, 4, 5, 6, 7, 890123456, time.FixedZone("test", 2*3600))
	row, err := repo.Append(ctx, activity.Entry{
		EngagementID: "eng-1", ActorID: "actor-1", Verb: "comment.updated",
		ObjectType: "comment", ObjectID: "c-1",
		Delta: json.RawMessage(`{"body":{"before":"x","after":"y"}}`), At: at,
	})
	if err != nil {
		t.Fatal(err)
	}
	if row.ID == "" {
		t.Fatal("expected assigned id")
	}
	if !row.At.UTC().Equal(at.Truncate(time.Microsecond).UTC()) {
		t.Fatalf("At = %v, want microsecond UTC %v", row.At, at)
	}
	if string(row.Delta) != `{"body":{"before":"x","after":"y"}}` {
		t.Fatalf("Delta = %s", row.Delta)
	}
	// A caller mutating its slice afterwards must not affect what was stored.
	delta := json.RawMessage(`{"state":"original"}`)
	lockoutRow, err := repo.Append(ctx, activity.Entry{Verb: "lockout", ObjectType: "session", ObjectID: "s-1", Delta: delta})
	if err != nil {
		t.Fatal(err)
	}
	copy(delta, []byte(`{"state":"mutated"}`))
	rows, _, err := repo.List(ctx, activity.ListFilter{ScopePlatform: true, Verb: "lockout"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || string(rows[0].Delta) != `{"state":"original"}` || rows[0].ID != lockoutRow.ID {
		t.Fatalf("stored delta followed caller mutation: %+v, want %s", rows, lockoutRow.ID)
	}
}

func TestActivityListScopesAndCursorErrors(t *testing.T) {
	repo := activity.New(storetest.Migrated(t))
	ctx := t.Context()
	if _, _, err := repo.List(ctx, activity.ListFilter{}); err == nil {
		t.Fatal("unscoped listing accepted")
	}
	if _, _, err := repo.List(ctx, activity.ListFilter{ScopePlatform: true, ScopeEngagement: "eng-1"}); err == nil {
		t.Fatal("double-scoped listing accepted")
	}
	engagement, err := repo.Append(ctx, activity.Entry{EngagementID: "eng-1", Verb: "finding.raised", ObjectType: "finding", ObjectID: "f-1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Append(ctx, activity.Entry{Verb: "session.login", ObjectType: "session", ObjectID: "s-1"}); err != nil {
		t.Fatal(err)
	}
	rows, next, err := repo.List(ctx, activity.ListFilter{ScopeEngagement: "eng-1", ObjectType: "finding", ObjectID: "f-1"})
	if err != nil || next != "" {
		t.Fatalf("engagement scope: %v, %v", next, err)
	}
	if len(rows) != 1 || rows[0].ID != engagement.ID || rows[0].EngagementID != "eng-1" {
		t.Fatalf("scoped rows = %+v", rows)
	}
	platform, _, err := repo.List(ctx, activity.ListFilter{ScopePlatform: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(platform) != 1 || platform[0].EngagementID != "" || platform[0].ID == engagement.ID {
		t.Fatalf("platform scope leaked engagement rows: %+v", platform)
	}
	_, _, err = repo.List(ctx, activity.ListFilter{ScopePlatform: true, Cursor: "not-a-cursor"})
	var typed *apierr.Error
	if !errors.As(err, &typed) || typed.Status() != 400 {
		t.Fatalf("invalid cursor: %v", err)
	}
}

func TestActivityReplayAfterCursor(t *testing.T) {
	db := storetest.Migrated(t)
	repo := activity.New(db)
	ctx := t.Context()
	engagement := "eng-1"
	var ids []string
	// Deterministic UUIDv7 ids require controlling the insert timestamp; later
	// ids must sort after earlier ones for replay to have a stable order.
	at := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
	err := db.Write(ctx, func(tx *sql.Tx) error {
		for i := range 3 {
			row, err := repo.Insert(ctx, tx, activity.Entry{
				EngagementID: engagement, Verb: "comment.created", ObjectType: "comment", ObjectID: "c-1", At: at.Add(time.Duration(i) * time.Second),
			})
			if err != nil {
				return err
			}
			ids = append(ids, row.ID)
		}
		// A platform row and an older row that must be excluded from replay.
		if _, err := repo.Insert(ctx, tx, activity.Entry{Verb: "session.login", ObjectType: "session", ObjectID: "s-1", At: at.Add(3 * time.Second)}); err != nil {
			return err
		}
		if _, err := repo.Insert(ctx, tx, activity.Entry{EngagementID: engagement, Verb: "comment.updated", ObjectType: "comment", ObjectID: "c-1", At: at.Add(-time.Second)}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	rows, err := repo.ReplayAfterCursor(ctx, engagement, ids[0], 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 || rows[0].ID != ids[1] || rows[1].ID != ids[2] || rows[2].Verb != "comment.updated" {
		t.Fatalf("replay = %+v", rows)
	}
	// A limit truncates the page, oldest first; the remaining row is fetchable.
	rows, err = repo.ReplayAfterCursor(ctx, engagement, ids[0], 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].ID != ids[1] {
		t.Fatalf("limited replay = %+v", rows)
	}
	rows, err = repo.ReplayAfterCursor(ctx, engagement, ids[1], 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].ID != ids[2] || rows[1].Verb != "comment.updated" {
		t.Fatalf("second replay page = %+v", rows)
	}
	rows, err = repo.ReplayAfterCursor(ctx, engagement, ids[2], 10)
	if err != nil {
		t.Fatal(err)
	}
	// The out-of-order comment.updated row was inserted last, so its UUIDv7 id
	// sorts after the seeded ids and it belongs to this page too.
	if len(rows) != 1 || rows[0].Verb != "comment.updated" {
		t.Fatalf("terminal replay returned %+v", rows)
	}
	// An unknown engagement id returns an empty page, not an error.
	if _, err := repo.ReplayAfterCursor(ctx, "01900000-0000-7000-e000-0000000000aa", ids[0], 10); err != nil {
		t.Fatal(err)
	}
}

func TestActivityZeroAtUsesStoredClock(t *testing.T) {
	repo := activity.New(storetest.Migrated(t))
	ctx := t.Context()
	before := time.Now().UTC().Add(-time.Second)
	row, err := repo.Append(ctx, activity.Entry{Verb: "finding.raised", ObjectType: "finding", ObjectID: "f-1"})
	if err != nil {
		t.Fatal(err)
	}
	if row.At.Before(before) || row.At.After(time.Now().UTC().Add(time.Second)) || row.At.Location() != time.UTC {
		t.Fatalf("zero At not assigned by store clock: %v", row.At)
	}
	rows, _, err := repo.List(ctx, activity.ListFilter{ScopePlatform: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].ID != row.ID {
		t.Fatalf("clock-assigned row = %+v", rows)
	}
	_ = reflect.DeepEqual(row.ID, rows[0].ID)
}
