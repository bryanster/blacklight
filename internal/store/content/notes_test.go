package content_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/bryanster/blacklight/internal/httpapi/apierr"
	"github.com/bryanster/blacklight/internal/store/content"
	"github.com/bryanster/blacklight/internal/store/storetest"
)

func TestNoteUpdateAndDeleteLifecycle(t *testing.T) {
	r := content.NewNotes(storetest.Migrated(t))
	ctx := t.Context()
	original, err := r.Create(ctx, content.Note{
		SourceID: content.SourceIDCustom, Version: content.VersionCurrent,
		ExternalID: "kb-one", Title: "Original", BodyMarkdown: "Original body",
		Tags: json.RawMessage(`["old"]`), TechniqueExternalID: "T1059",
	})
	if err != nil {
		t.Fatal(err)
	}
	// Update rewrites mutable fields, not the note's source/version identity.
	_, err = r.Update(ctx, content.Note{
		ID: original.ID, SourceID: content.SourceIDAtomic, Version: "other", ExternalID: "changed",
		Title: "Updated", BodyMarkdown: "## New body", Tags: json.RawMessage(`["blue","reviewed"]`),
		TechniqueExternalID: "T1059.001",
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := r.ByExternalID(ctx, original.SourceID, original.Version, original.ExternalID)
	if err != nil {
		t.Fatal(err)
	}
	var tags []string
	if err := json.Unmarshal(got.Tags, &tags); err != nil {
		t.Fatal(err)
	}
	if got.ID != original.ID || got.SourceID != original.SourceID || got.Version != original.Version || got.ExternalID != original.ExternalID || !got.CreatedAt.Equal(original.CreatedAt) {
		t.Fatalf("update changed identity: %+v", got)
	}
	if got.Title != "Updated" || got.BodyMarkdown != "## New body" || got.TechniqueExternalID != "T1059.001" || !reflect.DeepEqual(tags, []string{"blue", "reviewed"}) {
		t.Fatalf("mutable fields did not persist: %+v", got)
	}
	if _, err := r.Update(ctx, content.Note{ID: original.ID, Tags: json.RawMessage(`[]`)}); err != nil {
		t.Fatal(err)
	}
	got, err = r.ByID(ctx, original.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(got.Tags, &tags); err != nil {
		t.Fatal(err)
	}
	if got.Title != "" || got.BodyMarkdown != "" || got.TechniqueExternalID != "" || len(tags) != 0 {
		t.Fatalf("empty mutable values were not persisted: %+v", got)
	}
	if err := r.Delete(ctx, original.ID); err != nil {
		t.Fatal(err)
	}
	_, err = r.ByID(ctx, original.ID)
	assertNoteNotFound(t, err)
	_, err = r.ByExternalID(ctx, original.SourceID, original.Version, original.ExternalID)
	assertNoteNotFound(t, err)
	assertNoteNotFound(t, r.Delete(ctx, original.ID))
	_, err = r.Update(ctx, content.Note{ID: original.ID, Title: "Must not resurrect"})
	assertNoteNotFound(t, err)
}

func TestNoteListFiltersAndVisibility(t *testing.T) {
	db := storetest.Migrated(t)
	r := content.NewNotes(db)
	ctx := t.Context()
	// Deliberately insert out of external-ID order. The three search hits are
	// in different columns, and the subtechnique must not match its parent.
	fixtures := []content.Note{
		{SourceID: content.SourceIDCustom, Version: "v1", ExternalID: "c-note", Title: "Third", BodyMarkdown: "Contains NeEdLe", TechniqueExternalID: "T1059.001"},
		{SourceID: content.SourceIDCustom, Version: "v1", ExternalID: "a-needle", Title: "First", TechniqueExternalID: "T1059"},
		{SourceID: content.SourceIDCustom, Version: "v2", ExternalID: "b-note", Title: "NEEDLE in title", TechniqueExternalID: "T1059"},
		{SourceID: content.SourceIDAtomic, Version: "v1", ExternalID: "d-disabled", Title: "Disabled source"},
		{SourceID: content.SourceIDCustom, Version: content.StagingVersion, ExternalID: "e-staging", Title: "In flight"},
	}
	for _, note := range fixtures {
		if _, err := r.Create(ctx, note); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := content.NewSources(db).SetEnabled(ctx, content.SourceIDAtomic, false); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		filter content.NoteListFilter
		want   []string
	}{
		{"ordered without staging", content.NoteListFilter{}, []string{"a-needle", "b-note", "c-note", "d-disabled"}},
		{"enabled sources", content.NoteListFilter{EnabledOnly: true}, []string{"a-needle", "b-note", "c-note"}},
		{"disabled source explicit", content.NoteListFilter{SourceID: content.SourceIDAtomic}, []string{"d-disabled"}},
		{"disabled source and enabled only", content.NoteListFilter{SourceID: content.SourceIDAtomic, EnabledOnly: true}, []string{}},
		{"source and version", content.NoteListFilter{SourceID: content.SourceIDCustom, Version: "v1"}, []string{"a-needle", "c-note"}},
		{"case insensitive search columns", content.NoteListFilter{Q: "nEeDlE"}, []string{"a-needle", "b-note", "c-note"}},
		{"exact technique", content.NoteListFilter{Technique: "T1059"}, []string{"a-needle", "b-note"}},
		{"combined filters", content.NoteListFilter{SourceID: content.SourceIDCustom, Version: "v1", Q: "needle", Technique: "T1059.001", EnabledOnly: true}, []string{"c-note"}},
		{"limit after ordering", content.NoteListFilter{Limit: 2}, []string{"a-needle", "b-note"}},
		{"no search match", content.NoteListFilter{Q: "absent term"}, []string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			list, err := r.List(ctx, tc.filter)
			if err != nil {
				t.Fatal(err)
			}
			got := make([]string, 0, len(list))
			for _, note := range list {
				got = append(got, note.ExternalID)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("listed = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNoteExternalIdentityIsScoped(t *testing.T) {
	r := content.NewNotes(storetest.Migrated(t))
	ctx := t.Context()
	var saved []content.Note
	for _, note := range []content.Note{
		{SourceID: content.SourceIDCustom, Version: "v1", ExternalID: "shared", Title: "Custom one"},
		{SourceID: content.SourceIDCustom, Version: "v2", ExternalID: "shared", Title: "Custom two"},
		{SourceID: content.SourceIDAtomic, Version: "v1", ExternalID: "shared", Title: "Atomic one"},
	} {
		got, err := r.Create(ctx, note)
		if err != nil {
			t.Fatal(err)
		}
		saved = append(saved, got)
	}
	for _, note := range saved {
		got, err := r.ByExternalID(ctx, note.SourceID, note.Version, note.ExternalID)
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != note.ID || got.Title != note.Title {
			t.Fatalf("lookup crossed identity boundary: %+v, want %+v", got, note)
		}
	}
	duplicate := saved[0]
	duplicate.ID = ""
	duplicate.Title = "Duplicate"
	if _, err := r.Create(ctx, duplicate); !isConflict(err) {
		t.Fatalf("duplicate natural key: %v", err)
	}
	got, err := r.ByID(ctx, saved[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, saved[0]) {
		t.Fatalf("duplicate changed original: %+v", got)
	}
	_, err = r.ByExternalID(ctx, content.SourceIDAtomic, "v2", "shared")
	assertNoteNotFound(t, err)
}

func TestNoteCallbackFailureRollsBackMutations(t *testing.T) {
	for _, operation := range []string{"create", "update", "delete"} {
		t.Run(operation, func(t *testing.T) {
			r := content.NewNotes(storetest.Migrated(t))
			ctx := t.Context()
			original, err := r.Create(ctx, content.Note{SourceID: content.SourceIDCustom, Version: "v1", ExternalID: "original", Title: "Keep me", Tags: json.RawMessage(`["tag"]`)})
			if err != nil {
				t.Fatal(err)
			}
			failure := errors.New("activity write failed")
			fail := func(context.Context, *sql.Tx) error { return failure }
			switch operation {
			case "create":
				_, err = r.Create(ctx, content.Note{SourceID: original.SourceID, Version: original.Version, ExternalID: "uncommitted", Title: "Must roll back"}, fail)
			case "update":
				_, err = r.Update(ctx, content.Note{ID: original.ID, Title: "Must roll back"}, fail)
			case "delete":
				err = r.Delete(ctx, original.ID, fail)
			}
			if !errors.Is(err, failure) {
				t.Fatalf("mutation error = %v, want callback failure", err)
			}
			got, err := r.ByID(ctx, original.ID)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, original) {
				t.Fatalf("failed %s changed original: %+v", operation, got)
			}
			list, err := r.List(ctx, content.NoteListFilter{SourceID: original.SourceID})
			if err != nil {
				t.Fatal(err)
			}
			if len(list) != 1 || list[0].ID != original.ID {
				t.Fatalf("failed %s changed stored notes: %+v", operation, list)
			}
		})
	}
}

func assertNoteNotFound(t *testing.T, err error) {
	t.Helper()
	var typed *apierr.Error
	if !errors.As(err, &typed) || typed.Status() != 404 {
		t.Fatalf("expected not found, got %v", err)
	}
}
