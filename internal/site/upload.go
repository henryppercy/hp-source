package site

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// UploadImages mirrors the local image tree to the R2 bucket with rclone.
func UploadImages() error {
	cfg, err := r2ConfigFromEnv(r2EnvNames{
		account:  "HP_R2_ACCOUNT_ID",
		key:      "HP_R2_ACCESS_KEY_ID",
		secret:   "HP_R2_SECRET_ACCESS_KEY",
		bucket:   "HP_R2_BUCKET",
		endpoint: "HP_R2_ENDPOINT",
		region:   "HP_R2_REGION",
	})
	if err != nil {
		return err
	}

	src := filepath.Join(devAssetDir, imageDir)
	return cfg.rclone("sync", src, cfg.remote(), "--progress")
}

// backupLatestKey is the stable object key holding the newest snapshot,
// excluded from bucket lifecycle expiry so one backup always survives.
const backupLatestKey = "latest/data.sqlite"

// BackupDatabase snapshots the live database to a timestamped file and uploads
// it to snapshots/ in the R2 backup bucket with rclone, then copies it to
// backupLatestKey (a server-side copy within the same bucket).
func BackupDatabase(db *sql.DB) error {
	cfg, err := r2ConfigFromEnv(r2EnvNames{
		account:  "HP_R2_ACCOUNT_ID",
		key:      "HP_R2_DB_ACCESS_KEY_ID",
		secret:   "HP_R2_DB_SECRET_ACCESS_KEY",
		bucket:   "HP_R2_DB_BUCKET",
		endpoint: "HP_R2_DB_ENDPOINT",
		region:   "HP_R2_DB_REGION",
	})
	if err != nil {
		return err
	}

	name := fmt.Sprintf("data-%s.sqlite", time.Now().UTC().Format("20060102-150405"))
	snapshot := filepath.Join(os.TempDir(), name)
	defer os.Remove(snapshot)

	if err := snapshotDatabase(db, snapshot); err != nil {
		return err
	}

	key := "snapshots/" + name
	if err := cfg.rclone("copyto", snapshot, cfg.remote()+"/"+key, "--progress"); err != nil {
		return err
	}
	return cfg.rclone("copyto", cfg.remote()+"/"+key, cfg.remote()+"/"+backupLatestKey)
}

// snapshotDatabase writes a consistent, compacted copy of the database to dest
// using VACUUM INTO. The source runs in WAL mode, so readers and writers are
// not blocked while the snapshot is taken. dest must not already exist.
func snapshotDatabase(db *sql.DB, dest string) error {
	// The destination is passed as a SQL string literal with single quotes
	// doubled; VACUUM INTO does not accept bound parameters.
	dest = strings.ReplaceAll(dest, "'", "''")
	if _, err := db.ExecContext(context.Background(), "VACUUM INTO '"+dest+"'"); err != nil {
		return fmt.Errorf("failed to snapshot database: %w", err)
	}
	return nil
}

// r2EnvNames names the environment variables holding one set of R2 credentials.
// endpoint and region are optional overrides.
type r2EnvNames struct {
	account, key, secret, bucket string
	endpoint, region             string
}

// r2Config holds the credentials and target bucket for one rclone remote.
type r2Config struct {
	account, key, secret, bucket string
	endpoint, region             string
}

// r2ConfigFromEnv reads R2 credentials from the named environment variables,
// erroring on the first required one that is unset. The endpoint defaults to
// the global R2 endpoint. Buckets created in a jurisdiction (such as the EU)
// need the jurisdiction endpoint instead, e.g.
// https://<account-id>.eu.r2.cloudflarestorage.com. Region defaults to
// rclone's Cloudflare default ("auto"), the only region R2 accepts.
func r2ConfigFromEnv(names r2EnvNames) (r2Config, error) {
	cfg := r2Config{
		account:  os.Getenv(names.account),
		key:      os.Getenv(names.key),
		secret:   os.Getenv(names.secret),
		bucket:   os.Getenv(names.bucket),
		endpoint: os.Getenv(names.endpoint),
		region:   os.Getenv(names.region),
	}
	for name, v := range map[string]string{
		names.account: cfg.account,
		names.key:     cfg.key,
		names.secret:  cfg.secret,
		names.bucket:  cfg.bucket,
	} {
		if v == "" {
			return r2Config{}, fmt.Errorf("%s is not set", name)
		}
	}
	if cfg.endpoint == "" {
		cfg.endpoint = "https://" + cfg.account + ".r2.cloudflarestorage.com"
	} else if err := validEndpoint(cfg.endpoint); err != nil {
		return r2Config{}, fmt.Errorf("%s: %w", names.endpoint, err)
	}
	return cfg, nil
}

// validEndpoint checks that endpoint is an https URL whose host is a plain
// DNS name.
func validEndpoint(endpoint string) error {
	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("not a valid URL: %q", endpoint)
	}
	if u.Scheme != "https" || u.Host == "" {
		return fmt.Errorf("must be an https URL, got %q", endpoint)
	}
	for _, r := range u.Host {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-':
		default:
			return fmt.Errorf("host %q contains invalid character %q (copy-pasted placeholder?)", u.Host, r)
		}
	}
	return nil
}

// remote returns the rclone target for the configured bucket.
func (c r2Config) remote() string {
	return "r2:" + c.bucket
}

func (c r2Config) env() []string {
	env := []string{
		"RCLONE_CONFIG_R2_TYPE=s3",
		"RCLONE_CONFIG_R2_PROVIDER=Cloudflare",
		"RCLONE_CONFIG_R2_ACCESS_KEY_ID=" + c.key,
		"RCLONE_CONFIG_R2_SECRET_ACCESS_KEY=" + c.secret,
		"RCLONE_CONFIG_R2_ENDPOINT=" + c.endpoint,
		"RCLONE_CONFIG_R2_NO_CHECK_BUCKET=true",
	}
	if c.region != "" {
		env = append(env, "RCLONE_CONFIG_R2_REGION="+c.region)
	}
	return env
}

// rclone runs an rclone command against the configured remote, streaming
// stdout and stderr to the caller. The remote is defined entirely through
// environment overrides, so no rclone.conf is needed.
func (c r2Config) rclone(args ...string) error {
	cmd := exec.Command("rclone", args...)
	cmd.Env = append(os.Environ(), c.env()...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
