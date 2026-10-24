package engagement_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/bryanster/blacklight/internal/engagement"
	"github.com/bryanster/blacklight/internal/httpapi/apierr"
)

func TestCreateCommentValidatesTheBody(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	ctx := context.Background()

	if _, err := d.svc.CreateComment(ctx, d.actor(), d.eng.ID, d.exec.ID, ""); err == nil {
		t.Fatal("an empty comment was accepted")
	}
	if _, err := d.svc.CreateComment(ctx, d.actor(), d.eng.ID, d.exec.ID, strings.Repeat("x", 16385)); err == nil {
		t.Fatal("a comment over the 16 KiB limit was accepted")
	}

	// The boundary itself is allowed.
	if _, err := d.svc.CreateComment(ctx, d.actor(), d.eng.ID, d.exec.ID, strings.Repeat("x", 16384)); err != nil {
		t.Fatalf("a comment of exactly 16 KiB was refused: %v", err)
	}
}

func TestCreateCommentAttributesTheAuthorAndListsOldestFirst(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	ctx := context.Background()

	got, err := d.svc.CreateComment(ctx, d.actor(), d.eng.ID, d.exec.ID, "second")
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}
	if got.AuthorID != workbookActorID || got.Body != "second" || got.ExecutionID != d.exec.ID {
		t.Fatalf("comment = %+v, want the actor's body on the execution", got)
	}

	list, err := d.svc.ListComments(ctx, d.eng.ID, d.exec.ID)
	if err != nil {
		t.Fatalf("list comments: %v", err)
	}
	// The seeded "first" precedes the new one.
	if len(list) != 2 || list[0].Body != "first" || list[1].Body != "second" {
		t.Fatalf("comments = %+v, want first then second", list)
	}
}

func TestCreateCommentCrossEngagementIsNotFound(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	_, err := d.svc.CreateComment(context.Background(), d.actor(), d.other.ID, d.exec.ID, "hello")
	if !errors.Is(err, apierr.ErrNotFound) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func TestEditCommentKeepsThePreviousBodyAsARevision(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	ctx := context.Background()

	edited, err := d.svc.EditComment(ctx, d.actor(), engagement.EditCommentInput{
		CommentID:    d.comment.ID,
		EngagementID: d.eng.ID,
		ExecutionID:  d.exec.ID,
		Body:         "second",
	})
	if err != nil {
		t.Fatalf("edit comment: %v", err)
	}
	if edited.Body != "second" {
		t.Errorf("body = %q, want second", edited.Body)
	}
	if edited.EditedAt == nil {
		t.Error("edited_at was not stamped")
	}

	revs, err := d.svc.ListCommentRevisions(ctx, d.eng.ID, d.comment.ID)
	if err != nil {
		t.Fatalf("list revisions: %v", err)
	}
	if len(revs) != 1 || revs[0].Body != "first" || revs[0].EditedBy != workbookActorID {
		t.Fatalf("revisions = %+v, want the previous body once", revs)
	}
}

func TestEditCommentValidatesTheBodyAndTheBinding(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	ctx := context.Background()

	base := engagement.EditCommentInput{
		CommentID:    d.comment.ID,
		EngagementID: d.eng.ID,
		ExecutionID:  d.exec.ID,
		Body:         "changed",
	}

	empty := base
	empty.Body = ""
	if _, err := d.svc.EditComment(ctx, d.actor(), empty); err == nil {
		t.Fatal("an empty edit was accepted")
	}

	tooLong := base
	tooLong.Body = strings.Repeat("x", 16385)
	if _, err := d.svc.EditComment(ctx, d.actor(), tooLong); err == nil {
		t.Fatal("an edit over the 16 KiB limit was accepted")
	}

	wrongEngagement := base
	wrongEngagement.EngagementID = d.other.ID
	if _, err := d.svc.EditComment(ctx, d.actor(), wrongEngagement); !errors.Is(err, apierr.ErrNotFound) {
		t.Fatalf("cross-engagement edit: err = %v, want not found", err)
	}

	// None of the refusals wrote a revision or changed the body.
	got, err := d.svc.GetComment(ctx, d.comment.ID)
	if err != nil {
		t.Fatalf("get comment: %v", err)
	}
	if got.Body != "first" {
		t.Fatalf("refused edits changed the body to %q", got.Body)
	}
	revs, err := d.svc.ListCommentRevisions(ctx, d.eng.ID, d.comment.ID)
	if err != nil {
		t.Fatalf("list revisions: %v", err)
	}
	if len(revs) != 0 {
		t.Fatalf("refused edits wrote %d revisions, want 0", len(revs))
	}
}

func TestGetCommentInEngagementBindsThePathEngagement(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	ctx := context.Background()

	got, err := d.svc.GetCommentInEngagement(ctx, d.eng.ID, d.comment.ID)
	if err != nil {
		t.Fatalf("get comment in engagement: %v", err)
	}
	if got.ID != d.comment.ID {
		t.Fatalf("comment = %+v, want the seeded one", got)
	}

	// The same id under the other engagement is indistinguishable from missing.
	if _, err := d.svc.GetCommentInEngagement(ctx, d.other.ID, d.comment.ID); !errors.Is(err, apierr.ErrNotFound) {
		t.Fatalf("cross-engagement get: err = %v, want not found", err)
	}
}
