// Package activity stores the append-only record of what happened (M1-015).
//
// It is storage and nothing else. Verb vocabulary, redaction rules and the
// decision to write an entry live in internal/events. A repository here inserts
// a row and lists rows back; it never updates or deletes one.
package activity

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/bryanster/blacklight/internal/httpapi/apierr"
	"github.com/bryanster/blacklight/internal/store"
)

// DB is the part of the store this package needs. [store.Store] satisfies it.
type DB interface {
	Read() store.Reader
	Write(ctx context.Context, fn func(tx *sql.Tx) error) error
}

// Entry is one thing that happened, as the caller supplies it. The identifier
// and the timestamp (when zero) are assigned on insert.
type Entry struct {
	// EngagementID is empty for a platform event.
	EngagementID string

	// ActorID is empty when the actor is unknown (a failed login that named no
	// account).
	ActorID string

	Verb       string
	ObjectType string
	ObjectID   string

	// Delta is before/after for changed fields, already redacted. nil is stored
	// as SQL NULL.
	Delta json.RawMessage

	// At is when it happened. Zero means "now", truncated to the microsecond
	// DuckDB stores.
	At time.Time
}

// Row is one stored activity entry.
type Row struct {
	ID           string
	EngagementID string
	ActorID      string
	Verb         string
	ObjectType   string
	ObjectID     string
	Delta        json.RawMessage
	At           time.Time
}

// Entries reads and appends activity rows. Construct it with [New].
//
// There is deliberately no Update and no Delete. The table is append-only
// (0009_activity.sql); retention is a blctl command, not a method here.
type Entries struct {
	db DB
}

// New returns a repository over db.
func New(db DB) *Entries { return &Entries{db: db} }

// selectEntryColumns casts the JSON delta to VARCHAR: DuckDB scans a JSON
// column as a parsed value (map/slice), which database/sql cannot store in
// json.RawMessage and which deltaBytes would then re-marshal — losing object
// key order and integers above 2^53. CAST returns the raw stored JSON text.
const selectEntryColumns = `id, engagement_id, actor_id, verb, object_type, object_id,
	CAST(delta AS VARCHAR), "at"`

const selectEntry = `SELECT ` + selectEntryColumns + ` FROM app.activity `

const insertEntry = `INSERT INTO app.activity
	(id, engagement_id, actor_id, verb, object_type, object_id, delta, "at")
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

// Insert writes one entry on the caller's transaction and returns it as stored.
//
// The caller MUST already be inside [store.DB.Write]. Sharing that transaction
// with the change the entry describes is the central design constraint of
// M1-015: if the change rolls back, the log row must not exist either.
func (r *Entries) Insert(ctx context.Context, tx *sql.Tx, e Entry) (Row, error) {
	if tx == nil {
		return Row{}, errors.New("activity: Insert requires the caller's write transaction")
	}
	if err := validate(e); err != nil {
		return Row{}, err
	}

	id, err := newID()
	if err != nil {
		return Row{}, err
	}
	at := e.At
	if at.IsZero() {
		at = now()
	} else {
		at = toStorage(at)
	}

	var delta any
	if len(e.Delta) > 0 {
		// DuckDB's JSON parameter binding accepts []byte or string, not
		// json.RawMessage-as-interface. Copy so a caller reusing the slice
		// cannot race the driver.
		buf := make([]byte, len(e.Delta))
		copy(buf, e.Delta)
		delta = buf
	}

	if _, err := tx.ExecContext(ctx, insertEntry,
		id,
		nullString(e.EngagementID),
		nullString(e.ActorID),
		e.Verb,
		e.ObjectType,
		e.ObjectID,
		delta,
		at,
	); err != nil {
		return Row{}, fmt.Errorf("activity: insert %s: %w", e.Verb, err)
	}

	row, err := scanEntry(tx.QueryRowContext(ctx, selectEntry+`WHERE id = ?`, id))
	if err != nil {
		return Row{}, fmt.Errorf("activity: read back %s: %w", id, err)
	}
	return row, nil
}

// Append opens a write transaction and inserts one entry. Use it for events
// that are not accompanying another mutation — a failed login, a lockout.
// Prefer [Entries.Insert] when a transaction already exists.
func (r *Entries) Append(ctx context.Context, e Entry) (Row, error) {
	var row Row
	err := r.db.Write(ctx, func(tx *sql.Tx) error {
		var err error
		row, err = r.Insert(ctx, tx, e)
		return err
	})
	if err != nil {
		return Row{}, err
	}
	return row, nil
}

// ListFilter selects and pages through activity rows.
type ListFilter struct {
	// ScopePlatform restricts the result to rows with no engagement (the
	// installation-wide feed). ScopeEngagement restricts it to one engagement.
	// Exactly one of the two must be set; listing everything is not a supported
	// operation because the two feeds have different audiences.
	ScopePlatform   bool
	ScopeEngagement string
	ActorID         string
	Verb            string
	ObjectType      string
	ObjectID        string
	Cursor          string
	Limit           int
}

// List returns a page of activity rows, newest first, with a stable id
// tiebreaker. nextCursor is empty when there is no further page.
func (r *Entries) List(ctx context.Context, f ListFilter) (rows []Row, nextCursor string, err error) {
	if f.ScopePlatform == (f.ScopeEngagement != "") {
		return nil, "", errors.New("activity: List requires exactly one of ScopePlatform or ScopeEngagement")
	}
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	var (
		b    strings.Builder
		args []any
	)
	b.WriteString(selectEntry)
	b.WriteString(`WHERE `)
	if f.ScopePlatform {
		b.WriteString(`engagement_id IS NULL`)
	} else {
		b.WriteString(`engagement_id = ?`)
		args = append(args, f.ScopeEngagement)
	}
	if f.ActorID != "" {
		b.WriteString(` AND actor_id = ?`)
		args = append(args, f.ActorID)
	}
	if f.Verb != "" {
		b.WriteString(` AND verb = ?`)
		args = append(args, f.Verb)
	}
	if f.ObjectType != "" {
		b.WriteString(` AND object_type = ?`)
		args = append(args, f.ObjectType)
	}
	if f.ObjectID != "" {
		b.WriteString(` AND object_id = ?`)
		args = append(args, f.ObjectID)
	}
	if f.Cursor != "" {
		at, id, cerr := decodeCursor(f.Cursor)
		if cerr != nil {
			return nil, "", apierr.Validation(apierr.Field("cursor", "is not a cursor this server issued"))
		}
		// (at, id) descending: take rows strictly older than the cursor.
		b.WriteString(` AND ("at" < ? OR ("at" = ? AND id < ?))`)
		args = append(args, at, at, id)
	}
	b.WriteString(` ORDER BY "at" DESC, id DESC LIMIT ?`)
	args = append(args, limit+1)

	rs, err := r.db.Read().QueryContext(ctx, b.String(), args...)
	if err != nil {
		return nil, "", fmt.Errorf("activity: list: %w", err)
	}
	defer rs.Close()

	for rs.Next() {
		row, err := scanEntry(rs)
		if err != nil {
			return nil, "", err
		}
		rows = append(rows, row)
	}
	if err := rs.Err(); err != nil {
		return nil, "", fmt.Errorf("activity: list: %w", err)
	}

	if len(rows) > limit {
		last := rows[limit-1]
		rows = rows[:limit]
		nextCursor = encodeCursor(last.At, last.ID)
	}
	if rows == nil {
		rows = []Row{}
	}
	return rows, nextCursor, nil
}

// ReplayAfterCursor returns rows for engagementID with id strictly after
// cursor, oldest first, up to limit. Cursor is an activity row id (UUIDv7).
// Used for SSE catch-up (M4-004).
func (r *Entries) ReplayAfterCursor(ctx context.Context, engagementID, cursor string, limit int) ([]Row, error) {
	query := selectEntry + `WHERE engagement_id = ? AND id > ? ORDER BY id ASC LIMIT ?`
	rs, err := r.db.Read().QueryContext(ctx, query, engagementID, cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("activity: replay: %w", err)
	}
	defer rs.Close()

	var rows []Row
	for rs.Next() {
		row, err := scanEntry(rs)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	if err := rs.Err(); err != nil {
		return nil, fmt.Errorf("activity: replay: %w", err)
	}
	if rows == nil {
		rows = []Row{}
	}
	return rows, nil
}

func validate(e Entry) error {
	switch {
	case strings.TrimSpace(e.Verb) == "":
		return errors.New("activity: verb is required")
	case strings.TrimSpace(e.ObjectType) == "":
		return errors.New("activity: object_type is required")
	case strings.TrimSpace(e.ObjectID) == "":
		return errors.New("activity: object_id is required")
	}
	return nil
}

func scanEntry(row interface{ Scan(...any) error }) (Row, error) {
	var (
		e                 Row
		engagement, actor sql.NullString
		delta             any
	)
	if err := row.Scan(
		&e.ID, &engagement, &actor, &e.Verb, &e.ObjectType, &e.ObjectID, &delta, &e.At,
	); err != nil {
		return Row{}, err
	}
	if engagement.Valid {
		e.EngagementID = engagement.String
	}
	if actor.Valid {
		e.ActorID = actor.String
	}
	raw, err := deltaBytes(delta)
	if err != nil {
		return Row{}, err
	}
	e.Delta = raw
	e.At = e.At.UTC()
	return e, nil
}

// deltaBytes normalises whatever the driver handed back for a JSON column into
// the raw bytes callers expect. DuckDB returns a map; other drivers may return
// []byte or string.
func deltaBytes(v any) (json.RawMessage, error) {
	switch d := v.(type) {
	case nil:
		return nil, nil
	case []byte:
		if len(d) == 0 {
			return nil, nil
		}
		return append(json.RawMessage(nil), d...), nil
	case string:
		if d == "" {
			return nil, nil
		}
		return json.RawMessage(d), nil
	default:
		raw, err := json.Marshal(d)
		if err != nil {
			return nil, fmt.Errorf("activity: encode delta: %w", err)
		}
		return raw, nil
	}
}

func newID() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("activity: mint id: %w", err)
	}
	return id.String(), nil
}

func now() time.Time { return toStorage(time.Now()) }

func toStorage(t time.Time) time.Time { return t.UTC().Truncate(time.Microsecond) }

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// cursorJoin separates the two halves of an opaque cursor. It is a character
// that cannot appear in an RFC 3339 nano timestamp or a UUID.
const cursorJoin = "|"

func encodeCursor(at time.Time, id string) string {
	raw := toStorage(at).Format(time.RFC3339Nano) + cursorJoin + id
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(cursor string) (time.Time, string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, "", err
	}
	atText, id, ok := strings.Cut(string(raw), cursorJoin)
	if !ok || id == "" {
		return time.Time{}, "", errors.New("malformed")
	}
	at, err := time.Parse(time.RFC3339Nano, atText)
	if err != nil {
		return time.Time{}, "", err
	}
	return toStorage(at), id, nil
}
