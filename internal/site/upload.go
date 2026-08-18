package site

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// UploadImages mirrors the local image tree to the R2 bucket with rclone.
func UploadImages() error {
	env, remote, err := r2Env()
	if err != nil {
		return err
	}

	src := filepath.Join(devAssetDir, imageDir)
	cmd := exec.Command("rclone", "sync", src, remote, "--progress")
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// r2Env reads the R2 credentials and returns the rclone remote definition as
// environment overrides plus the "remote:bucket" target.
func r2Env() ([]string, string, error) {
	account := os.Getenv("HP_R2_ACCOUNT_ID")
	key := os.Getenv("HP_R2_ACCESS_KEY_ID")
	secret := os.Getenv("HP_R2_SECRET_ACCESS_KEY")
	bucket := os.Getenv("HP_R2_BUCKET")

	for name, v := range map[string]string{
		"HP_R2_ACCOUNT_ID":        account,
		"HP_R2_ACCESS_KEY_ID":     key,
		"HP_R2_SECRET_ACCESS_KEY": secret,
		"HP_R2_BUCKET":            bucket,
	} {
		if v == "" {
			return nil, "", fmt.Errorf("%s is not set", name)
		}
	}

	env := []string{
		"RCLONE_CONFIG_R2_TYPE=s3",
		"RCLONE_CONFIG_R2_PROVIDER=Cloudflare",
		"RCLONE_CONFIG_R2_ACCESS_KEY_ID=" + key,
		"RCLONE_CONFIG_R2_SECRET_ACCESS_KEY=" + secret,
		"RCLONE_CONFIG_R2_ENDPOINT=https://" + account + ".r2.cloudflarestorage.com",
	}
	return env, "r2:" + bucket, nil
}
