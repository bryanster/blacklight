package evidence

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bryanster/blacklight/internal/config"
	storengagement "github.com/bryanster/blacklight/internal/store/engagement"
	"github.com/bryanster/blacklight/internal/store/storetest"
)

func TestStoreRoundTripAndDeduplication(t *testing.T) {
	ctx := context.Background()
	repo := storengagement.NewEvidenceBlobRepo(storetest.Migrated(t))
	s := NewStore(t.TempDir(), config.Evidence{MaxUploadBytes: 1024}, repo)
	const payload = "captured evidence\x00\xff"
	wantDigest := fmt.Sprintf("%x", sha256.Sum256([]byte(payload)))
	for upload := 1; upload <= 2; upload++ {
		digest, existed, size, err := s.Put(ctx, strings.NewReader(payload), "text/plain", "engagement")
		if err != nil {
			t.Fatal(err)
		}
		if digest != wantDigest || existed != (upload == 2) || size != int64(len(payload)) {
			t.Fatalf("upload %d: digest=%q existed=%v size=%d", upload, digest, existed, size)
		}
		blob, err := repo.GetBlob(ctx, digest)
		if err != nil {
			t.Fatal(err)
		}
		if blob.RefCount != upload || blob.Size != size || blob.MIME != "text/plain" {
			t.Fatalf("upload %d: persisted blob = %+v", upload, blob)
		}
		r, err := s.Open(digest)
		if err != nil {
			t.Fatal(err)
		}
		got, readErr := io.ReadAll(r)
		closeErr := r.Close()
		if readErr != nil || closeErr != nil || string(got) != payload {
			t.Fatalf("read = %q, errors = %v / %v", got, readErr, closeErr)
		}
		files := evidenceFiles(t, s.root)
		if len(files) != 1 || files[0] != filepath.Join(s.root, blob.StoragePath) {
			t.Fatalf("expected only persisted blob, got %v", files)
		}
	}
	if err := s.RemoveBlobFile(wantDigest); err != nil {
		t.Fatal(err)
	}
	if r, err := s.Open(wantDigest); err == nil {
		_ = r.Close()
		t.Fatal("removed blob is still readable")
	}
	if err := s.RemoveBlobFile(wantDigest); err != nil {
		t.Fatalf("repeated removal: %v", err)
	}
}

// Override only the database boundary under test; successful operations still
// use the real repository and its migrated database.
type faultBlobDB struct {
	blobDB
	used                        int64
	quotaErr, insertErr, refErr error
}

func (db faultBlobDB) EngagementBlobBytes(context.Context, string) (int64, error) {
	return db.used, db.quotaErr
}
func (db faultBlobDB) InsertBlob(ctx context.Context, digest, mime, path string, size int64) error {
	if db.insertErr != nil {
		return db.insertErr
	}
	return db.blobDB.InsertBlob(ctx, digest, mime, path, size)
}
func (db faultBlobDB) IncrementRef(ctx context.Context, digest string) error {
	if db.refErr != nil {
		return db.refErr
	}
	return db.blobDB.IncrementRef(ctx, digest)
}

func TestStoreUploadLimits(t *testing.T) {
	for _, tc := range []struct {
		name                     string
		uploadLimit, quota, used int64
		tooLarge, overQuota      bool
	}{
		{name: "exact file limit", uploadLimit: 4},
		{name: "one byte over file limit", uploadLimit: 3, tooLarge: true},
		{name: "exact engagement quota", uploadLimit: 4, quota: 10, used: 6},
		{name: "one byte over engagement quota", uploadLimit: 4, quota: 10, used: 7, overQuota: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := storengagement.NewEvidenceBlobRepo(storetest.Migrated(t))
			s := NewStore(t.TempDir(), config.Evidence{
				MaxUploadBytes: config.ByteSize(tc.uploadLimit), MaxEngagementBytes: config.ByteSize(tc.quota),
			}, faultBlobDB{blobDB: repo, used: tc.used})
			digest, _, size, err := s.Put(context.Background(), strings.NewReader("data"), "text/plain", "engagement")
			if !tc.tooLarge && !tc.overQuota {
				if err != nil || size != 4 || digest == "" {
					t.Fatalf("exact boundary rejected: digest=%q size=%d err=%v", digest, size, err)
				}
				return
			}
			var large *ErrTooLarge
			var quota *ErrEngagementQuota
			if tc.tooLarge && (!errors.As(err, &large) || large.Limit != tc.uploadLimit || large.Got != 4) {
				t.Fatalf("expected size limit error, got %v", err)
			}
			if tc.overQuota && (!errors.As(err, &quota) || quota.Limit != tc.quota || quota.Used != tc.used || quota.Got != 4) {
				t.Fatalf("expected quota error, got %v", err)
			}
			if files := evidenceFiles(t, s.root); len(files) != 0 {
				t.Fatalf("rejected upload left files: %v", files)
			}
		})
	}
}

type failingEvidenceReader struct{ err error }

func (r failingEvidenceReader) Read([]byte) (int, error) { return 0, r.err }

func TestStoreFailedUploadCleanup(t *testing.T) {
	failure := errors.New("injected storage failure")
	for _, name := range []string{"nil reader", "empty reader", "partial read failure", "cancelled context", "quota lookup failure", "insert failure"} {
		t.Run(name, func(t *testing.T) {
			repo := storengagement.NewEvidenceBlobRepo(storetest.Migrated(t))
			db := faultBlobDB{blobDB: repo}
			ctx := context.Background()
			var src io.Reader = strings.NewReader("data")
			var wantErr error
			switch name {
			case "nil reader":
				src = nil
			case "empty reader":
				src = strings.NewReader("")
			case "partial read failure":
				src = io.MultiReader(src, failingEvidenceReader{failure})
				wantErr = failure
			case "cancelled context":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
				wantErr = context.Canceled
			case "quota lookup failure":
				db.quotaErr, wantErr = failure, failure
			case "insert failure":
				db.insertErr, wantErr = failure, failure
			}
			s := NewStore(t.TempDir(), config.Evidence{MaxUploadBytes: 1024, MaxEngagementBytes: 1024}, db)
			_, _, _, err := s.Put(ctx, src, "text/plain", "engagement")
			if err == nil || (wantErr != nil && !errors.Is(err, wantErr)) {
				t.Fatalf("expected failure %v, got %v", wantErr, err)
			}
			if files := evidenceFiles(t, s.root); len(files) != 0 {
				t.Fatalf("failed upload left files: %v", files)
			}
			if _, err := repo.GetBlob(context.Background(), fmt.Sprintf("%x", sha256.Sum256([]byte("data")))); err == nil {
				t.Fatal("failed upload persisted a blob row")
			}
		})
	}
}

func TestStoreFailedDuplicatePreservesOriginal(t *testing.T) {
	ctx := context.Background()
	repo := storengagement.NewEvidenceBlobRepo(storetest.Migrated(t))
	failure := errors.New("reference update failed")
	s := NewStore(t.TempDir(), config.Evidence{MaxUploadBytes: 1024}, faultBlobDB{blobDB: repo, refErr: failure})
	digest, _, _, err := s.Put(ctx, strings.NewReader("original"), "text/plain", "engagement")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := s.Put(ctx, strings.NewReader("original"), "text/plain", "engagement"); !errors.Is(err, failure) {
		t.Fatalf("expected reference failure, got %v", err)
	}
	blob, err := repo.GetBlob(ctx, digest)
	if err != nil {
		t.Fatal(err)
	}
	if blob.RefCount != 1 {
		t.Fatalf("reference count = %d, want 1", blob.RefCount)
	}
	got, err := os.ReadFile(filepath.Join(s.root, blob.StoragePath))
	if err != nil || string(got) != "original" {
		t.Fatalf("original = %q, err=%v", got, err)
	}
	if files := evidenceFiles(t, s.root); len(files) != 1 {
		t.Fatalf("files after duplicate failure: %v", files)
	}
}

func evidenceFiles(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}
