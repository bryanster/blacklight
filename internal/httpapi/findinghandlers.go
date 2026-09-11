package httpapi

import (
	"context"
	"fmt"

	"github.com/oapi-codegen/nullable"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/bryanster/blacklight/internal/authn"
	engagement "github.com/bryanster/blacklight/internal/engagement"
	"github.com/bryanster/blacklight/internal/httpapi/apierr"
	"github.com/bryanster/blacklight/internal/httpapi/gen"
	storengagement "github.com/bryanster/blacklight/internal/store/engagement"
)

// Finding handlers (M3-011).

// ListFindings returns every finding in an engagement, newest first.
func (h *handlers) ListFindings(ctx context.Context,
	request gen.ListFindingsRequestObject) (gen.ListFindingsResponseObject, error) {

	findings, err := h.engagements.ListFindings(ctx, request.EngagementId.String())
	if err != nil {
		return nil, err
	}

	wire := make([]gen.Finding, len(findings))
	for i, f := range findings {
		w, err := findingToWire(ctx, h, f)
		if err != nil {
			return nil, err
		}
		wire[i] = w
	}
	return gen.ListFindings200JSONResponse(wire), nil
}

// CreateFinding raises a new finding in an engagement.
func (h *handlers) CreateFinding(ctx context.Context,
	request gen.CreateFindingRequestObject) (gen.CreateFindingResponseObject, error) {

	actor, ok := authn.SubjectFrom(ctx)
	if !ok {
		return nil, apierr.Unauthenticated("no authenticated subject")
	}

	body := request.Body

	in := engagement.CreateFindingInput{
		EngagementID:   request.EngagementId.String(),
		Title:          body.Title,
		Description:    body.Description,
		Severity:       string(body.Severity),
		Recommendation: stringFromPtr(body.Recommendation),
		Owner:          stringFromPtr(body.Owner),
	}
	if body.Status != nil {
		in.Status = string(*body.Status)
	}
	if body.CreatedFromExecution != nil {
		in.CreatedFromExecution = body.CreatedFromExecution.String()
	}

	finding, err := h.engagements.CreateFinding(ctx, actor, in)
	if err != nil {
		return nil, fmt.Errorf("create finding: %w", err)
	}

	w, err := findingToWire(ctx, h, finding)
	if err != nil {
		return nil, err
	}
	return gen.CreateFinding201JSONResponse(w), nil
}

// GetFinding returns a single finding with its step ids.
func (h *handlers) GetFinding(ctx context.Context,
	request gen.GetFindingRequestObject) (gen.GetFindingResponseObject, error) {

	finding, err := h.engagements.GetFinding(ctx, request.FindingId.String())
	if err != nil {
		return nil, err
	}

	w, err := findingToWire(ctx, h, finding)
	if err != nil {
		return nil, err
	}
	return gen.GetFinding200JSONResponse(w), nil
}

// PatchFinding updates a finding.
func (h *handlers) PatchFinding(ctx context.Context,
	request gen.PatchFindingRequestObject) (gen.PatchFindingResponseObject, error) {

	actor, ok := authn.SubjectFrom(ctx)
	if !ok {
		return nil, apierr.Unauthenticated("no authenticated subject")
	}

	body := request.Body

	in := engagement.UpdateFindingInput{
		Title:          stringFromPtr(body.Title),
		Description:    stringFromPtr(body.Description),
		Recommendation: stringFromPtr(body.Recommendation),
		Owner:          stringFromPtr(body.Owner),
	}
	if body.Severity != nil {
		in.Severity = string(*body.Severity)
	}
	if body.Status != nil {
		in.Status = string(*body.Status)
	}

	finding, err := h.engagements.UpdateFinding(ctx, actor, request.FindingId.String(), in)
	if err != nil {
		return nil, fmt.Errorf("patch finding: %w", err)
	}

	w, err := findingToWire(ctx, h, finding)
	if err != nil {
		return nil, err
	}
	return gen.PatchFinding200JSONResponse(w), nil
}

// DeleteFinding removes a finding and its step join rows.
func (h *handlers) DeleteFinding(ctx context.Context,
	request gen.DeleteFindingRequestObject) (gen.DeleteFindingResponseObject, error) {

	actor, ok := authn.SubjectFrom(ctx)
	if !ok {
		return nil, apierr.Unauthenticated("no authenticated subject")
	}

	if err := h.engagements.DeleteFinding(ctx, actor, request.FindingId.String()); err != nil {
		return nil, fmt.Errorf("delete finding: %w", err)
	}
	return gen.DeleteFinding204Response{}, nil
}

// SetFindingSteps replaces the step set for a finding.
func (h *handlers) SetFindingSteps(ctx context.Context,
	request gen.SetFindingStepsRequestObject) (gen.SetFindingStepsResponseObject, error) {

	actor, ok := authn.SubjectFrom(ctx)
	if !ok {
		return nil, apierr.Unauthenticated("no authenticated subject")
	}

	stepIDs := make([]string, len(request.Body.StepIds))
	for i, id := range request.Body.StepIds {
		stepIDs[i] = id.String()
	}

	if err := h.engagements.SetFindingSteps(ctx, actor, request.FindingId.String(), stepIDs); err != nil {
		return nil, fmt.Errorf("set finding steps: %w", err)
	}
	return gen.SetFindingSteps204Response{}, nil
}

// findingToWire converts a store finding to its OpenAPI representation,
// including linked step ids.
func findingToWire(ctx context.Context, h *handlers, f storengagement.Finding) (gen.Finding, error) {
	id := mustParseUUID(f.ID)
	engID := mustParseUUID(f.EngagementID)

	w := gen.Finding{
		Id:             id,
		EngagementId:   engID,
		Title:          f.Title,
		Description:    f.Description,
		Severity:       gen.FindingSeverity(f.Severity),
		Recommendation: f.Recommendation,
		Owner:          f.Owner,
		Status:         gen.FindingStatus(f.Status),
		CreatedAt:      f.CreatedAt,
		UpdatedAt:      f.UpdatedAt,
	}

	if f.CreatedFromExecution != "" {
		cf := mustParseUUID(f.CreatedFromExecution)
		w.CreatedFromExecution = nullable.NewNullableWithValue(cf)
	}

	// Load linked steps.
	steps, err := h.engagements.FindingSteps(ctx, f.ID)
	if err != nil {
		return gen.Finding{}, err
	}
	stepIDs, err := h.visibleFindingStepIDs(ctx, f.EngagementID, steps)
	if err != nil {
		return gen.Finding{}, err
	}
	w.StepIds = stepIDs

	return w, nil
}

// visibleFindingStepIDs returns the linked step ids a finding carries on the
// wire for the calling seat. In a blind engagement the blue seat must not
// learn that an unrevealed step exists (docs/authz.md), so each link is
// resolved live — a step revealed after the finding was linked is returned,
// an unrevealed one is dropped. This is the same filter the archive export
// applies to finding→step links (archivehandler.go); the findings themselves
// stay visible to blue, so a finding whose links are all hidden arrives with
// an empty stepIds. A lookup failure fails the request rather than leaking
// the id, matching the handler-side blind checks (evidenceConcealed).
func (h *handlers) visibleFindingStepIDs(ctx context.Context,
	engagementID string, steps []storengagement.Step,
) ([]openapi_types.UUID, error) {
	scope, err := h.stepBlindScope(ctx, engagementID)
	if err != nil {
		return nil, err
	}
	if !scope.Withholds() {
		stepIDs := make([]openapi_types.UUID, len(steps))
		for i, s := range steps {
			stepIDs[i] = mustParseUUID(s.ID)
		}
		return stepIDs, nil
	}

	stepIDs := make([]openapi_types.UUID, 0, len(steps))
	for _, s := range steps {
		revealed, err := h.IsStepRevealed(ctx, s.ID)
		if err != nil {
			return nil, fmt.Errorf("step visibility: %w", err)
		}
		if revealed {
			stepIDs = append(stepIDs, mustParseUUID(s.ID))
		}
	}
	return stepIDs, nil
}

// stringFromPtr returns the string value from a pointer, or "" if nil.
func stringFromPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
