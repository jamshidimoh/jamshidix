package main

import (
	"archive/zip"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	appVersion     = "0.1.1"
	singBoxVersion = "1.14.1"
	singBoxSHA256  = "5197f16d492d93202dc623622149a6ed040f8eca263128f91d603f2b901baa89"
)

//go:embed assets/client.config.template.json
var clientTemplate []byte

var programDir = filepath.Join(os.Getenv("ProgramFiles"), "Jamshidix")
var dataDir = filepath.Join(os.Getenv("ProgramData"), "Jamshidix")

func main() {
	if runtime.GOOS != "windows" {
		fatal(errors.New("Jamshidix supports Windows only"))
	}
	if len(os.Args) < 2 {
		fatal(runOneClick())
		return
	}
	switch os.Args[1] {
	case "version":
		fmt.Printf("Jamshidix %s | sing-box %s\n", appVersion, singBoxVersion)
	case "install":
		fatal(installSingBox())
	case "config":
		fatal(renderConfig(os.Args[2:]))
	case "check":
		fatal(checkConfig())
	case "run":
		fatal(runTunnel())
	case "stop":
		fatal(stopSingBox())
	case "status":
		fatal(status())
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println("Jamshidix Windows executable")
	fmt.Println()
	fmt.Println("  Jamshidix.exe install")
	fmt.Println("  Jamshidix.exe config --server-ip <IP> --uuid <UUID> --public-key <KEY> --short-id <ID> --handshake-host <HOST>")
	fmt.Println("  Jamshidix.exe check")
	fmt.Println("  Jamshidix.exe run")
	fmt.Println("  Jamshidix.exe status")
	fmt.Println("  Jamshidix.exe version")
	fmt.Println()
	fmt.Println("Kill-switch/autostart remain available as PowerShell scripts in client/windows.")
}

func installSingBox() error {
	if err := os.MkdirAll(programDir, 0755); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp("", "jamshidix-install-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	url := fmt.Sprintf("https://github.com/SagerNet/sing-box/releases/download/v%s/sing-box-%s-windows-amd64.zip", singBoxVersion, singBoxVersion)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download sing-box: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download sing-box: HTTP %s", resp.Status)
	}
	archivePath := filepath.Join(tmp, "sing-box.zip")
	f, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	if _, err = io.Copy(f, resp.Body); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	got, err := sha256File(archivePath)
	if err != nil {
		return err
	}
	if !strings.EqualFold(got, singBoxSHA256) {
		return fmt.Errorf("SHA-256 mismatch: expected %s, got %s", singBoxSHA256, got)
	}
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer zr.Close()
	var source string
	for _, zf := range zr.File {
		if strings.EqualFold(filepath.Base(zf.Name), "sing-box.exe") {
			source = filepath.Join(tmp, "sing-box.exe")
			r, err := zf.Open()
			if err != nil {
				return err
			}
			out, err := os.Create(source)
			if err == nil {
				_, err = io.Copy(out, r)
			}
			_ = r.Close()
			_ = out.Close()
			if err != nil {
				return err
			}
			break
		}
	}
	if source == "" {
		return errors.New("sing-box.exe not found in archive")
	}
	target := filepath.Join(programDir, "sing-box.exe")
	if err := copyFile(source, target); err != nil {
		return err
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}
	fmt.Println("Installed:", target)
	fmt.Println("Configuration directory:", dataDir)
	return nil
}

func renderConfig(args []string) error {
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	serverIP := fs.String("server-ip", "", "")
	uuid := fs.String("uuid", "", "")
	publicKey := fs.String("public-key", "", "")
	shortID := fs.String("short-id", "", "")
	handshakeHost := fs.String("handshake-host", "", "")
	if err := fs.Parse(args); err != nil {
		return err
	}
	fields := map[string]string{"server-ip": *serverIP, "uuid": *uuid, "public-key": *publicKey, "short-id": *shortID, "handshake-host": *handshakeHost}
	for k, v := range fields {
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("missing --%s", k)
		}
	}
	cfg := string(clientTemplate)
	replacements := map[string]string{
		"REPLACE_WITH_SERVER_IP":          *serverIP,
		"REPLACE_WITH_UUID":               *uuid,
		"REPLACE_WITH_REALITY_PUBLIC_KEY": *publicKey,
		"REPLACE_WITH_SHORT_ID":           *shortID,
		"REPLACE_WITH_HANDSHAKE_HOST":     *handshakeHost,
	}
	for old, value := range replacements {
		cfg = strings.ReplaceAll(cfg, old, value)
	}
	var parsed any
	if err := json.Unmarshal([]byte(cfg), &parsed); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dataDir, "client.json")
	if err := os.WriteFile(path, []byte(cfg), 0600); err != nil {
		return err
	}
	if _, err := os.Stat(singBoxPath()); err == nil {
		if err := runCommand(singBoxPath(), "check", "-c", path); err != nil {
			_ = os.Remove(path)
			return fmt.Errorf("sing-box validation failed: %w", err)
		}
	}
	fmt.Println("Client configuration created:", path)
	return nil
}

func checkConfig() error {
	path := filepath.Join(dataDir, "client.json")
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("client config not found: %s", path)
	}
	return runCommand(singBoxPath(), "check", "-c", path)
}

func runTunnel() error {
	path := filepath.Join(dataDir, "client.json")
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("client config not found: %s", path)
	}
	if err := runCommand(singBoxPath(), "check", "-c", path); err != nil {
		return err
	}
	fmt.Println("Starting TUN. Administrator privileges may be required.")
	cmd := exec.Command(singBoxPath(), "run", "-c", path)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	return cmd.Run()
}

func status() error {
	cfg := filepath.Join(dataDir, "client.json")
	fmt.Println("Jamshidix:", appVersion)
	fmt.Println("sing-box:", singBoxPath())
	fmt.Println("config:", cfg)
	if _, err := os.Stat(cfg); err == nil {
		fmt.Println("config status: present")
	} else {
		fmt.Println("config status: missing")
	}
	out, err := exec.Command("tasklist", "/FI", "IMAGENAME eq sing-box.exe").CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Printf("%s", out)
	return nil
}

func singBoxPath() string {
	return filepath.Join(programDir, "sing-box.exe")
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func fatal(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}
