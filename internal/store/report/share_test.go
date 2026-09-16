package report

import (
	"testing"
	"time"

	"github.com/bryanster/blacklight/internal/store/storetest"
)

func TestShareRoundTripAndLookup(t *testing.T) {
	db := storetest.Migrated(t)
	r := NewShares(db)
	ctx := t.Context()
	engID := insertEngagement(t, db)
	reports, versions := NewReports(db), NewVersions(db)
	rep, err := reports.Create(ctx, NewReport{EngagementID: engID, Title: "Shared", CreatedBy: "admin"})
	if err != nil {
		t.Fatal(err)
	}
	version, err := versions.Insert(ctx, NewVersion{ReportID: rep.ID, BlindScope: "all", HTML: "<html></html>"})
	if err != nil {
		t.Fatal(err)
	}
	expires := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)
	password, label := "password-hash", "Guest link"
	maxGrants := 3
	share, err := r.Insert(ctx, NewShare{
		VersionID: version.ID, TokenHash: "token-1", PasswordHash: &password,
		ExpiresAt: &expires, CreatedBy: "admin", Label: &label, MaxGrants: &maxGrants,
	})
	if err != nil {
		t.Fatal(err)
	}
	if share.ID == "" || share.RevokedAt != nil || share.CreatedAt.IsZero() || share.VersionID != version.ID {
		t.Fatalf("insert returned unexpected share: %+v", share)
	}
	// The lookup helpers return nil, not an error, when nothing matches.
	for name, tc := range map[string]struct {
		lookup    func() (*ReportShare, error)
		tokenHash string
	}{
		"by token hash": {lookup: func() (*ReportShare, error) { return r.ByTokenHash(ctx, "token-1") }, tokenHash: "token-1"},
		"by id":         {lookup: func() (*ReportShare, error) { return r.ByID(ctx, share.ID) }, tokenHash: "token-1"},
		"missing token": {lookup: func() (*ReportShare, error) { return r.ByTokenHash(ctx, "unknown") }, tokenHash: ""},
		"missing id":    {lookup: func() (*ReportShare, error) { return r.ByID(ctx, "01900000-0000-7000-e000-0000000000aa") }, tokenHash: ""},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := tc.lookup()
			if err != nil {
				t.Fatal(err)
			}
			if tc.tokenHash == "" {
				if got != nil {
					t.Fatalf("unknown share returned: %+v", got)
				}
				return
			}
			if got == nil || got.TokenHash != tc.tokenHash || got.PasswordHash == nil || *got.PasswordHash != password ||
				got.ExpiresAt == nil || !got.ExpiresAt.Equal(expires) || got.Label == nil || *got.Label != label ||
				got.MaxGrants == nil || *got.MaxGrants != maxGrants || got.CreatedBy != "admin" {
				t.Fatalf("optional fields did not round-trip: %+v", got)
			}
		})
	}
}

func TestShareListingAndRevocation(t *testing.T) {
	db := storetest.Migrated(t)
	r := NewShares(db)
	ctx := t.Context()
	reports, versions := NewReports(db), NewVersions(db)
	rep, err := reports.Create(ctx, NewReport{EngagementID: insertEngagement(t, db), Title: "Listed", CreatedBy: "admin"})
	if err != nil {
		t.Fatal(err)
	}
	version, err := versions.Insert(ctx, NewVersion{ReportID: rep.ID, BlindScope: "all", HTML: "<html></html>"})
	if err != nil {
		t.Fatal(err)
	}
	first, err := r.Insert(ctx, NewShare{VersionID: version.ID, TokenHash: "share-1", CreatedBy: "admin"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := r.Insert(ctx, NewShare{VersionID: version.ID, TokenHash: "share-2", CreatedBy: "admin"})
	if err != nil {
		t.Fatal(err)
	}
	other, err := r.Insert(ctx, NewShare{VersionID: version.ID, TokenHash: "share-3", CreatedBy: "admin"})
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Revoke(ctx, second.ID); err != nil {
		t.Fatal(err)
	}
	if err := r.Revoke(ctx, second.ID); err == nil {
		t.Fatal("double revocation succeeded")
	}
	// Shares are scoped to their version and ordered newest first.
	listed, err := r.ListByVersion(ctx, version.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 3 || listed[0].ID != other.ID || listed[1].ID != second.ID || listed[2].ID != first.ID {
		t.Fatalf("listing order: %+v", listed)
	}
	if listed[1].RevokedAt == nil || listed[0].RevokedAt != nil || listed[2].RevokedAt != nil {
		t.Fatalf("revocation state: %+v", listed)
	}
	got, err := r.ByID(ctx, second.ID)
	if err != nil || got == nil || got.RevokedAt == nil {
		t.Fatalf("revocation not persisted: %+v, %v", got, err)
	}
	if _, err := r.ByTokenHash(ctx, "share-3"); err != nil {
		t.Fatal(err)
	}
	missing := "01900000-0000-7000-e000-0000000000aa"
	if err := r.Revoke(ctx, missing); err == nil {
		t.Fatal("revoking an unknown share succeeded")
	}
}

func TestGrantCapsClaimsAndRevocation(t *testing.T) {
	db := storetest.Migrated(t)
	shares, grants := NewShares(db), NewGrants(db)
	ctx := t.Context()
	reports, versions := NewReports(db), NewVersions(db)
	rep, err := reports.Create(ctx, NewReport{EngagementID: insertEngagement(t, db), Title: "Grants", CreatedBy: "admin"})
	if err != nil {
		t.Fatal(err)
	}
	version, err := versions.Insert(ctx, NewVersion{ReportID: rep.ID, BlindScope: "all", HTML: "<html></html>"})
	if err != nil {
		t.Fatal(err)
	}
	capped, err := shares.Insert(ctx, NewShare{VersionID: version.ID, TokenHash: "capped", CreatedBy: "admin"})
	if err != nil {
		t.Fatal(err)
	}
	uncapped, err := shares.Insert(ctx, NewShare{VersionID: version.ID, TokenHash: "uncapped", CreatedBy: "admin"})
	if err != nil {
		t.Fatal(err)
	}
	cap := 2
	alice, bob := "user-alice", "user-bob"
	userAlice, userBob := &alice, &bob
	// A user grant is claimed at creation; an anonymous grant is not.
	first, ok, err := grants.ClaimInsert(ctx, NewGrant{ShareID: capped.ID, UserID: userAlice}, &cap)
	if err != nil || !ok || first.ClaimedAt == nil || first.UserID == nil || *first.UserID != alice || first.InviteCodeHash != nil {
		t.Fatalf("first claim: grant=%+v, ok=%v, err=%v", first, ok, err)
	}
	second, ok, err := grants.ClaimInsert(ctx, NewGrant{ShareID: capped.ID}, &cap)
	if err != nil || !ok || second.ClaimedAt != nil || second.UserID != nil {
		t.Fatalf("anonymous claim: grant=%+v, ok=%v, err=%v", second, ok, err)
	}
	if _, ok, err := grants.ClaimInsert(ctx, NewGrant{ShareID: capped.ID}, &cap); err != nil || ok {
		t.Fatalf("grant cap not enforced: ok=%v, err=%v", ok, err)
	}
	// A revoked grant frees a slot for a new claim.
	if err := grants.Revoke(ctx, first.ID); err != nil {
		t.Fatal(err)
	}
	if err := grants.Revoke(ctx, first.ID); err == nil {
		t.Fatal("double grant revocation succeeded")
	}
	third, ok, err := grants.ClaimInsert(ctx, NewGrant{ShareID: capped.ID, UserID: userBob}, &cap)
	if err != nil || !ok || third.UserID == nil || *third.UserID != bob {
		t.Fatalf("revocation freed no slot: grant=%+v, ok=%v, err=%v", third, ok, err)
	}
	// Revoked grants do not count as membership or cap usage.
	active, err := grants.ByShareAndUser(ctx, capped.ID, bob)
	if err != nil || active == nil || active.ID != third.ID || active.RevokedAt != nil {
		t.Fatalf("active grant lookup: grant=%+v, err=%v", active, err)
	}
	if _, err := grants.ByShareAndUser(ctx, capped.ID, alice); err != nil {
		t.Fatal(err)
	}
	revoked, err := grants.ByShareAndUser(ctx, capped.ID, alice)
	if err != nil || revoked != nil {
		t.Fatalf("revoked grant is still active: %+v, %v", revoked, err)
	}
	// Unlimited shares accept more claims than the capped one.
	for range 3 {
		if _, ok, err := grants.ClaimInsert(ctx, NewGrant{ShareID: uncapped.ID}, nil); err != nil || !ok {
			t.Fatalf("unlimited share claim: ok=%v, err=%v", ok, err)
		}
	}
	listed, err := grants.ListByShare(ctx, uncapped.ID)
	if err != nil || len(listed) != 3 {
		t.Fatalf("uncapped grants: %+v, %v", listed, err)
	}
	capTwo := 0
	// A non-positive cap is documented as unlimited; it must not reject claims.
	if _, ok, err := grants.ClaimInsert(ctx, NewGrant{ShareID: uncapped.ID}, &capTwo); err != nil || !ok {
		t.Fatalf("non-positive cap rejected claims: ok=%v, err=%v", ok, err)
	}
	if err := grants.Revoke(ctx, "01900000-0000-7000-e000-0000000000aa"); err == nil {
		t.Fatal("revoking an unknown grant succeeded")
	}
}
