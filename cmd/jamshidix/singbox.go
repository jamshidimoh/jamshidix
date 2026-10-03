package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	appVersion     = "0.3.1"
	singBoxVersion = "1.14.1"
	singBoxSHA256  = "5197f16d492d93202dc623622149a6ed040f8eca263128f91d603f2b901baa89"
	maxArchiveSize = 200 << 20
)

func winDir(env, fallback string) string {
	if v := os.Getenv(env); v != "" {
		return v
	}
	return fallback
}

var (
	programDir = filepath.Join(winDir("ProgramFiles", `C:\Program Files`), "Jamshidix")
	dataDir    = filepath.Join(winDir("ProgramData", `C:\ProgramData`), "Jamshidix")
)

func singBoxPath() string { return filepath.Join(programDir, "sing-box.exe") }
func configPath() string  { return filepath.Join(dataDir, "client.json") }
func logPath() string     { return filepath.Join(dataDir, "jamshidix.log") }
func pidPath() string     { return filepath.Join(dataDir, "sing-box.pid") }

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// singBoxArchive returns the verified official archive: embedded when the
// build carries it, otherwise downloaded (with retries) from the upstream release.
func singBoxArchive() ([]byte, error) {
	data := embeddedSingBoxZip
	if len(data) == 0 {
		url := fmt.Sprintf("https://github.com/SagerNet/sing-box/releases/download/v%s/sing-box-%s-windows-amd64.zip", singBoxVersion, singBoxVersion)
		client := &http.Client{Timeout: 10 * time.Minute}
		var lastErr error
		for attempt := 1; attempt <= 3; attempt++ {
			data, lastErr = download(client, url)
			if lastErr == nil {
				break
			}
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
		if lastErr != nil {
			return nil, fmt.Errorf("download sing-box: %w", lastErr)
		}
	}
	if got := sha256Hex(data); !strings.EqualFold(got, singBoxSHA256) {
		return nil, fmt.Errorf("SHA-256 mismatch: expected %s, got %s", singBoxSHA256, got)
	}
	return data, nil
}

func download(c *http.Client, url string) ([]byte, error) {
	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %s", resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, maxArchiveSize))
}

// extractSingBox pulls only sing-box.exe out of the archive.
func extractSingBox(archive []byte) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, err
	}
	for _, zf := range zr.File {
		if !strings.EqualFold(filepath.Base(zf.Name), "sing-box.exe") {
			continue
		}
		r, err := zf.Open()
		if err != nil {
			return nil, err
		}
		defer r.Close()
		return io.ReadAll(io.LimitReader(r, maxArchiveSize))
	}
	return nil, errors.New("sing-box.exe not found in archive")
}

func installSingBox() error {
	archive, err := singBoxArchive()
	if err != nil {
		return err
	}
	exe, err := extractSingBox(archive)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(programDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	_ = stopSingBox() // a running binary cannot be replaced
	tmp := singBoxPath() + ".new"
	if err := os.WriteFile(tmp, exe, 0o755); err != nil {
		return err
	}
	if err := os.Rename(tmp, singBoxPath()); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	fmt.Println("Installed:", singBoxPath())
	return nil
}
