package report

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/bryanster/blacklight/internal/httpapi/apierr"
	"github.com/bryanster/blacklight/internal/store/storetest"
)

func TestTemplateReplacementIsOrderedAndAtomic(t *testing.T) {
	db := storetest.Migrated(t)
	r := NewTemplates(db)
	ctx := t.Context()
	tmpl, err := r.Create(ctx, NewTemplate{EngagementID: insertEngagement(t, db), Name: "Template", CreatedBy: "admin"})
	if err != nil {
		t.Fatal(err)
	}
	original := []NewTemplateBlock{
		{BlockID: "rich_text", Params: json.RawMessage(`{"html":"first"}`)},
		{BlockID: "rich_text", Params: json.RawMessage(`{"html":"second"}`)},
	}
	if _, err := r.ReplaceBlocks(ctx, tmpl.ID, original); err != nil {
		t.Fatal(err)
	}
	before, err := r.BlocksByTemplate(ctx, tmpl.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(before) != 2 {
		t.Fatalf("blocks = %+v", before)
	}
	for i, block := range before {
		if block.Ordinal != i || block.BlockID != original[i].BlockID || !reflect.DeepEqual(block.Params, original[i].Params) {
			t.Fatalf("block %d = %+v", i, block)
		}
	}
	// The first insert succeeds, then invalid JSON rejects the second. The
	// deletion and successful insert must both roll back, preserving the draft.
	if _, err := r.ReplaceBlocks(ctx, tmpl.ID, []NewTemplateBlock{
		{BlockID: "rich_text", Params: json.RawMessage(`{"html":"replacement"}`)},
		{BlockID: "rich_text", Params: json.RawMessage(`{"broken":`)},
	}); err == nil {
		t.Fatal("invalid replacement succeeded")
	}
	after, err := r.BlocksByTemplate(ctx, tmpl.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("failed replacement changed blocks: got %+v, want %+v", after, before)
	}
	if _, err := r.ReplaceBlocks(ctx, tmpl.ID, original[1:]); err != nil {
		t.Fatal(err)
	}
	after, err = r.BlocksByTemplate(ctx, tmpl.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != 1 || after[0].Ordinal != 0 || string(after[0].Params) != string(original[1].Params) {
		t.Fatalf("replacement retained old blocks or ordinals: %+v", after)
	}
	if _, err := r.ReplaceBlocks(ctx, tmpl.ID, nil); err != nil {
		t.Fatal(err)
	}
	after, err = r.BlocksByTemplate(ctx, tmpl.ID)
	if err != nil || len(after) != 0 {
		t.Fatalf("clear blocks: %+v, %v", after, err)
	}
}

func TestTemplateMutationCallbackFailureRollsBack(t *testing.T) {
	db := storetest.Migrated(t)
	r := NewTemplates(db)
	ctx := t.Context()
	engID := insertEngagement(t, db)
	failure := errors.New("audit write failed")
	fail := func(context.Context, *sql.Tx) error { return failure }
	if _, err := r.Create(ctx, NewTemplate{EngagementID: engID, Name: "Uncommitted", CreatedBy: "admin"}, fail); !errors.Is(err, failure) {
		t.Fatalf("create error = %v", err)
	}
	list, err := r.ListByEngagement(ctx, engID)
	if err != nil || len(list) != 0 {
		t.Fatalf("failed create persisted: %+v, %v", list, err)
	}
	tmpl, err := r.Create(ctx, NewTemplate{EngagementID: engID, Name: "Original", CreatedBy: "admin"})
	if err != nil {
		t.Fatal(err)
	}
	before, err := r.ByID(ctx, tmpl.ID)
	if err != nil {
		t.Fatal(err)
	}
	name := "Uncommitted rename"
	if _, err := r.Update(ctx, tmpl.ID, TemplateUpdate{Name: &name}, fail); !errors.Is(err, failure) {
		t.Fatalf("update error = %v", err)
	}
	after, err := r.ByID(ctx, tmpl.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("failed update changed template: %+v, want %+v", after, before)
	}
	name = "Renamed"
	if _, err := r.Update(ctx, tmpl.ID, TemplateUpdate{Name: &name}); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Update(ctx, tmpl.ID, TemplateUpdate{}); err != nil {
		t.Fatal(err)
	}
	after, err = r.ByID(ctx, tmpl.ID)
	if err != nil || after.Name != name || after.CreatedBy != before.CreatedBy || !after.CreatedAt.Equal(before.CreatedAt) {
		t.Fatalf("rename/no-op update lost metadata: %+v, %v", after, err)
	}
	list, err = r.ListByEngagement(ctx, engID)
	if err != nil || len(list) != 1 || list[0].ID != tmpl.ID || list[0].Name != name {
		t.Fatalf("list after rename: %+v, %v", list, err)
	}
	list, err = r.ListByEngagement(ctx, "01900000-0000-7000-e000-0000000000aa")
	if err != nil || len(list) != 0 {
		t.Fatalf("unrelated engagement returned templates: %+v, %v", list, err)
	}
}

func TestTemplateMissingMutationsReturnNotFound(t *testing.T) {
	r := NewTemplates(storetest.Migrated(t))
	const missing = "01900000-0000-7000-e000-0000000000aa"
	for _, tc := range []struct {
		name string
		run  func() error
	}{
		{"read", func() error { _, err := r.ByID(t.Context(), missing); return err }},
		{"update", func() error { _, err := r.Update(t.Context(), missing, TemplateUpdate{}); return err }},
		{"replace blocks", func() error { _, err := r.ReplaceBlocks(t.Context(), missing, nil); return err }},
		{"delete", func() error { return r.Delete(t.Context(), missing) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var typed *apierr.Error
			if err := tc.run(); !errors.As(err, &typed) || typed.Status() != 404 {
				t.Fatalf("expected not found, got %v", err)
			}
		})
	}
}
