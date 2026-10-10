package site

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/henryppercy/hp-source/internal/database"
)

// TestSnapshotDatabase verifies that VACUUM INTO produces a readable,
// self-contained copy of a live database.
func TestSnapshotDatabase(t *testing.T) {
	dir := t.TempDir()

	db, err := database.NewDB(filepath.Join(dir, "live.sqlite"))
	if err != nil {
		t.Fatalf("open source: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec("CREATE TABLE t (id INTEGER PRIMARY KEY, body TEXT)"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	if _, err := db.Exec("INSERT INTO t (body) VALUES ('hello')"); err != nil {
		t.Fatalf("insert: %v", err)
	}

	dest := filepath.Join(dir, "snapshot.sqlite")
	if err := snapshotDatabase(db, dest); err != nil {
		t.Fatalf("snapshot: %v", err)
	}

	snap, err := database.NewDB(dest)
	if err != nil {
		t.Fatalf("open snapshot: %v", err)
	}
	defer snap.Close()

	var body string
	if err := snap.QueryRow("SELECT body FROM t").Scan(&body); err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	if body != "hello" {
		t.Fatalf("snapshot body = %q, want %q", body, "hello")
	}
}

// TestSnapshotDatabaseQuotedPath verifies that destinations containing a
// single quote are escaped correctly in the VACUUM INTO statement.
func TestSnapshotDatabaseQuotedPath(t *testing.T) {
	dir := t.TempDir()

	db, err := database.NewDB(filepath.Join(dir, "live.sqlite"))
	if err != nil {
		t.Fatalf("open source: %v", err)
	}
	defer db.Close()

	destDir := filepath.Join(dir, "it's here")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	dest := filepath.Join(destDir, "snapshot.sqlite")

	if err := snapshotDatabase(db, dest); err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("stat snapshot: %v", err)
	}
}

// TestR2ConfigFromEnv verifies credential resolution: required variables must
// be set, the endpoint defaults to the global R2 endpoint, and endpoint/region
// overrides are honoured (needed for buckets in a jurisdiction such as the EU).
func TestR2ConfigFromEnv(t *testing.T) {
	names := r2EnvNames{
		account:  "TEST_R2_ACCOUNT",
		key:      "TEST_R2_KEY",
		secret:   "TEST_R2_SECRET",
		bucket:   "TEST_R2_BUCKET",
		endpoint: "TEST_R2_ENDPOINT",
		region:   "TEST_R2_REGION",
	}

	t.Run("missing required var", func(t *testing.T) {
		if _, err := r2ConfigFromEnv(names); err == nil {
			t.Fatal("expected error for unset env vars")
		}
	})

	t.Run("default endpoint and empty region", func(t *testing.T) {
		t.Setenv("TEST_R2_ACCOUNT", "acc")
		t.Setenv("TEST_R2_KEY", "key")
		t.Setenv("TEST_R2_SECRET", "secret")
		t.Setenv("TEST_R2_BUCKET", "bucket")

		cfg, err := r2ConfigFromEnv(names)
		if err != nil {
			t.Fatalf("config: %v", err)
		}
		if want := "https://acc.r2.cloudflarestorage.com"; cfg.endpoint != want {
			t.Fatalf("endpoint = %q, want %q", cfg.endpoint, want)
		}
		if cfg.region != "" {
			t.Fatalf("region = %q, want empty", cfg.region)
		}
	})

	t.Run("invalid endpoint is rejected", func(t *testing.T) {
		t.Setenv("TEST_R2_ACCOUNT", "acc")
		t.Setenv("TEST_R2_KEY", "key")
		t.Setenv("TEST_R2_SECRET", "secret")
		t.Setenv("TEST_R2_BUCKET", "bucket")
		t.Setenv("TEST_R2_ENDPOINT", "https://<acc>.eu.r2.cloudflarestorage.com")

		if _, err := r2ConfigFromEnv(names); err == nil {
			t.Fatal("expected error for endpoint with invalid characters")
		}
	})

	t.Run("endpoint and region overrides", func(t *testing.T) {
		t.Setenv("TEST_R2_ACCOUNT", "acc")
		t.Setenv("TEST_R2_KEY", "key")
		t.Setenv("TEST_R2_SECRET", "secret")
		t.Setenv("TEST_R2_BUCKET", "bucket")
		t.Setenv("TEST_R2_ENDPOINT", "https://acc.eu.r2.cloudflarestorage.com")
		t.Setenv("TEST_R2_REGION", "eu")

		cfg, err := r2ConfigFromEnv(names)
		if err != nil {
			t.Fatalf("config: %v", err)
		}
		if want := "https://acc.eu.r2.cloudflarestorage.com"; cfg.endpoint != want {
			t.Fatalf("endpoint = %q, want %q", cfg.endpoint, want)
		}
		if cfg.region != "eu" {
			t.Fatalf("region = %q, want %q", cfg.region, "eu")
		}
	})
}

// TestR2ConfigEnv verifies the rclone environment overrides, including the
// no-check-bucket flag required for object-scoped R2 tokens, and that the
// optional region is only emitted when set.
func TestR2ConfigEnv(t *testing.T) {
	cfg := r2Config{
		account:  "acc",
		key:      "key",
		secret:   "secret",
		bucket:   "bucket",
		endpoint: "https://acc.eu.r2.cloudflarestorage.com",
	}

	want := []string{
		"RCLONE_CONFIG_R2_TYPE=s3",
		"RCLONE_CONFIG_R2_PROVIDER=Cloudflare",
		"RCLONE_CONFIG_R2_ACCESS_KEY_ID=key",
		"RCLONE_CONFIG_R2_SECRET_ACCESS_KEY=secret",
		"RCLONE_CONFIG_R2_ENDPOINT=https://acc.eu.r2.cloudflarestorage.com",
		"RCLONE_CONFIG_R2_NO_CHECK_BUCKET=true",
	}
	if got := cfg.env(); !reflect.DeepEqual(got, want) {
		t.Fatalf("env = %v, want %v", got, want)
	}

	cfg.region = "eu"
	env := cfg.env()
	if got := env[len(env)-1]; got != "RCLONE_CONFIG_R2_REGION=eu" {
		t.Fatalf("last env entry = %q, want region override", got)
	}
}
