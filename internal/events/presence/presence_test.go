package presence

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func heartbeat(t *testing.T, r *Registry, e Entry, wantJoined bool) {
	t.Helper()
	joined, err := r.Heartbeat(e)
	if err != nil || joined != wantJoined {
		t.Fatalf("Heartbeat(%s): joined=%v err=%v, want joined=%v", e.UserID, joined, err, wantJoined)
	}
}

// The registry has no clock seam. Age entries under its mutex to exercise
// ordering and expiry without sleeps or changing the production API.
func seenAt(r *Registry, id uuid.UUID, at time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e := r.entries[id]
	e.LastSeenAt = at
	r.entries[id] = e
}

func TestHeartbeatUpdatesAndCollapsesTabs(t *testing.T) {
	r := New(Options{HeartbeatTTL: time.Hour})
	first := Entry{PresenceID: uuid.New(), UserID: "alice", EngagementID: "one", DisplayName: "Alice", Focus: Focus{StepID: "old"}}
	second := first
	second.PresenceID = uuid.New()
	second.Focus = Focus{StepID: "new", ExecutionID: "execution"}
	heartbeat(t, r, first, true)
	seenAt(r, first.PresenceID, time.Now().Add(-time.Minute))
	heartbeat(t, r, second, true)
	heartbeat(t, r, Entry{PresenceID: uuid.New(), UserID: "alice", EngagementID: "two"}, true)
	got := r.Snapshot("one")
	if len(got) != 1 || got[0].UserID != "alice" || got[0].DisplayName != "Alice" || got[0].TabCount != 2 || got[0].Focus == nil || *got[0].Focus != second.Focus {
		t.Fatalf("collapsed snapshot: %+v", got)
	}
	// Snapshots must not expose mutable registry state.
	got[0].Focus.StepID = "changed by caller"
	if focus := r.Snapshot("one")[0].Focus; focus == nil || *focus != second.Focus {
		t.Fatalf("snapshot mutation changed stored focus: %+v", focus)
	}
	if !r.Leave(first.PresenceID) || r.Leave(first.PresenceID) {
		t.Fatal("Leave must report whether a tab was present")
	}
	second.Focus = Focus{}
	second.DisplayName = "Alice updated"
	heartbeat(t, r, second, false)
	got = r.Snapshot("one")
	if len(got) != 1 || got[0].TabCount != 1 || got[0].Focus != nil || got[0].DisplayName != second.DisplayName {
		t.Fatalf("updated snapshot: %+v", got)
	}
	if r.Count() != 2 || r.CountEngagement("one") != 1 || len(r.Snapshot("missing")) != 0 {
		t.Fatal("tab counts or engagement isolation changed")
	}
}

func TestLeaveUserIsScopedToEngagement(t *testing.T) {
	r := New(Options{})
	for _, e := range []Entry{
		{PresenceID: uuid.New(), UserID: "alice", EngagementID: "one"},
		{PresenceID: uuid.New(), UserID: "alice", EngagementID: "one"},
		{PresenceID: uuid.New(), UserID: "bob", EngagementID: "one"},
		{PresenceID: uuid.New(), UserID: "alice", EngagementID: "two"},
	} {
		heartbeat(t, r, e, true)
	}
	if removed := r.LeaveUser("alice", "one"); removed != 2 {
		t.Fatalf("removed %d tabs, want 2", removed)
	}
	if removed := r.LeaveUser("alice", "one"); removed != 0 {
		t.Fatalf("second removal = %d", removed)
	}
	assertUsers(t, r, "one", "bob")
	assertUsers(t, r, "two", "alice")
	if r.Count() != 2 {
		t.Fatalf("remaining tabs = %d", r.Count())
	}
}

func TestPresenceExpiresOnSnapshotAndSweep(t *testing.T) {
	r := New(Options{HeartbeatTTL: time.Hour})
	for _, e := range []Entry{
		{PresenceID: uuid.New(), UserID: "expired", EngagementID: "one"},
		{PresenceID: uuid.New(), UserID: "expired", EngagementID: "two"},
		{PresenceID: uuid.New(), UserID: "live", EngagementID: "one"},
	} {
		heartbeat(t, r, e, true)
		if e.UserID == "expired" {
			seenAt(r, e.PresenceID, time.Now().Add(-2*time.Hour))
		}
	}
	assertUsers(t, r, "one", "live")
	if r.Count() != 2 {
		t.Fatalf("snapshot did not remove expired tab: count=%d", r.Count())
	}
	if removed := r.Sweep(); removed != 1 {
		t.Fatalf("Sweep removed %d, want 1", removed)
	}
	if removed := r.Sweep(); removed != 0 {
		t.Fatalf("second Sweep removed %d", removed)
	}
	assertUsers(t, r, "two")
	assertUsers(t, r, "one", "live")
}

func TestHeartbeatRefreshesExpiry(t *testing.T) {
	r := New(Options{HeartbeatTTL: time.Hour})
	e := Entry{PresenceID: uuid.New(), UserID: "alice", EngagementID: "one"}
	heartbeat(t, r, e, true)
	seenAt(r, e.PresenceID, time.Now().Add(-2*time.Hour))
	heartbeat(t, r, e, false)
	if removed := r.Sweep(); removed != 0 {
		t.Fatalf("fresh heartbeat expired: %d", removed)
	}
	assertUsers(t, r, "one", "alice")
}

func TestPresenceCapsEvictOldestWithoutEvictingOnUpdate(t *testing.T) {
	for _, tc := range []struct {
		name          string
		opts          Options
		oldEngagement string
	}{
		{name: "per engagement", opts: Options{MaxPerEngagement: 2}, oldEngagement: "one"},
		{name: "global", opts: Options{MaxGlobal: 3}, oldEngagement: "two"},
		{name: "combined", opts: Options{MaxPerEngagement: 2, MaxGlobal: 3}, oldEngagement: "one"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.opts.HeartbeatTTL = time.Hour
			r := New(tc.opts)
			old := Entry{PresenceID: uuid.New(), UserID: "old", EngagementID: tc.oldEngagement}
			keep := Entry{PresenceID: uuid.New(), UserID: "keep", EngagementID: "one"}
			other := Entry{PresenceID: uuid.New(), UserID: "other", EngagementID: "three"}
			heartbeat(t, r, old, true)
			seenAt(r, old.PresenceID, time.Now().Add(-time.Minute))
			heartbeat(t, r, keep, true)
			heartbeat(t, r, other, true)
			heartbeat(t, r, keep, false)
			if r.Count() != 3 {
				t.Fatalf("update evicted a tab: %d", r.Count())
			}
			heartbeat(t, r, Entry{PresenceID: uuid.New(), UserID: "new", EngagementID: "one"}, true)
			assertUsers(t, r, "one", "keep", "new")
			assertUsers(t, r, "three", "other")
			if r.Leave(old.PresenceID) || r.Count() != 3 {
				t.Fatal("oldest tab was not the only eviction")
			}
		})
	}
}

func TestPresenceConcurrentLifecycle(t *testing.T) {
	r := New(Options{HeartbeatTTL: time.Hour})
	const users = 16
	var wg sync.WaitGroup
	for i := range users {
		wg.Add(1)
		go func() {
			defer wg.Done()
			e := Entry{PresenceID: uuid.New(), UserID: fmt.Sprint(i), EngagementID: "one"}
			joined, err := r.Heartbeat(e)
			if err != nil || !joined {
				t.Errorf("join: %v / %v", joined, err)
				return
			}
			for range 4 {
				joined, err = r.Heartbeat(e)
				if err != nil || joined {
					t.Errorf("update: %v / %v", joined, err)
				}
				r.Snapshot("one")
				r.Sweep()
			}
		}()
	}
	wg.Wait()
	if r.Count() != users || len(r.Snapshot("one")) != users {
		t.Fatalf("lost concurrent users: %d", r.Count())
	}
	for i := range users {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if n := r.LeaveUser(fmt.Sprint(i), "one"); n != 1 {
				t.Errorf("removed %d tabs, want 1", n)
			}
		}()
	}
	wg.Wait()
	if r.Count() != 0 {
		t.Fatalf("tabs remain after leaving: %d", r.Count())
	}
}

func TestStartSweepStops(t *testing.T) {
	r := New(Options{})
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() { defer close(done); r.StartSweep(stop) }()
	close(stop)
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("StartSweep did not stop")
	}
}

func assertUsers(t *testing.T, r *Registry, engagement string, want ...string) {
	t.Helper()
	got := make([]string, 0)
	for _, e := range r.Snapshot(engagement) {
		got = append(got, e.UserID)
	}
	sort.Strings(got)
	expected := append([]string{}, want...)
	sort.Strings(expected)
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("engagement %s: users=%v, want %v", engagement, got, expected)
	}
}
