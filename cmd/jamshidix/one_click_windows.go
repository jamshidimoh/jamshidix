//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	messageBoxInformation = 0x00000040
	messageBoxWarning     = 0x00000030
	messageBoxError       = 0x00000010
	createNoWindow        = 0x08000000
)

var (
	user32              = syscall.NewLazyDLL("user32.dll")
	messageBoxW         = user32.NewProc("MessageBoxW")
	errOneClickRelaunch = errors.New("elevated instance launched")
)

func runOneClick() error {
	if err := ensureElevated(); err != nil {
		if errors.Is(err, errOneClickRelaunch) {
			return nil
		}
		showMessage("Jamshidix", "اجرای Jamshidix با دسترسی Administrator شروع نشد.\n\n"+err.Error(), messageBoxError)
		return nil
	}

	if isSingBoxRunning() {
		showMessage("Jamshidix", "تونل Jamshidix در حال اجراست.", messageBoxInformation)
		return nil
	}

	if !existingSingBoxIsExpected() {
		if err := installSingBox(); err != nil {
			showMessage("Jamshidix", "نصب sing-box انجام نشد.\n\n"+err.Error(), messageBoxError)
			return nil
		}
	}

	if err := ensureClientConfig(); err != nil {
		showMessage("Jamshidix", err.Error(), messageBoxWarning)
		return nil
	}
	if err := checkConfig(); err != nil {
		showMessage("Jamshidix", "کانفیگ معتبر نیست.\n\n"+err.Error(), messageBoxError)
		return nil
	}

	path := filepath.Join(dataDir, "client.json")
	if err := startDetached(singBoxPath(), "run", "-c", path); err != nil {
		showMessage("Jamshidix", "اجرای تونل شروع نشد.\n\n"+err.Error(), messageBoxError)
		return nil
	}
	if waitForProcess(3) {
		showMessage("Jamshidix", "تونل با موفقیت شروع شد.\n\nTUN: JamshidixTunnel\nProxy: 127.0.0.1:2080", messageBoxInformation)
		return nil
	}
	showMessage("Jamshidix", "sing-box شروع نشد یا بلافاصله متوقف شد.\n\nبرای بررسی، Jamshidix.exe status را از CMD اجرا کنید.", messageBoxError)
	return nil
}

func showMessage(title, body string, flags uintptr) {
	t, _ := syscall.UTF16PtrFromString(title)
	b, _ := syscall.UTF16PtrFromString(body)
	_, _, _ = messageBoxW.Call(0, uintptr(unsafe.Pointer(b)), uintptr(unsafe.Pointer(t)), flags)
}

func startDetached(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

func isSingBoxRunning() bool {
	out, err := exec.Command("tasklist", "/FI", "IMAGENAME eq sing-box.exe", "/FO", "CSV", "/NH").CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(out)), `"sing-box.exe"`)
}

func stopSingBox() error {
	if !isSingBoxRunning() {
		return nil
	}
	cmd := exec.Command("taskkill", "/IM", "sing-box.exe", "/F", "/T")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return errors.New(msg)
		}
		return err
	}
	return nil
}

func existingSingBoxIsExpected() bool {
	if _, err := os.Stat(singBoxPath()); err != nil {
		return false
	}
	out, err := exec.Command(singBoxPath(), "version").CombinedOutput()
	return err == nil && strings.Contains(string(out), singBoxVersion)
}

func ensureClientConfig() error {
	path := filepath.Join(dataDir, "client.json")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "client.json")
		if _, statErr := os.Stat(candidate); statErr == nil {
			if err := copyFile(candidate, path); err != nil {
				return err
			}
			return nil
		}
	}

	templatePath := filepath.Join(dataDir, "client.config.template.json")
	if err := os.WriteFile(templatePath, clientTemplate, 0600); err != nil {
		return err
	}
	openConfigFolder()
	return fmt.Errorf("اولین اجرا نیاز به client.json دارد.\n\nفایل نمونه در این مسیر ساخته شد:\n%s\n\nکانفیگ واقعی را با همین نام کنار Jamshidix.exe قرار دهید؛ سپس فقط با دوبارکلیک Jamshidix.exe اجرا کنید.", templatePath)
}

func ensureElevated() error {
	if os.Getenv("JAMSHIDIX_ELEVATED") == "1" {
		return nil
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, err = filepath.Abs(exe)
	if err != nil {
		return err
	}
	quoted := strings.ReplaceAll(exe, "'", "''")
	ps := fmt.Sprintf("$p='%s'; Start-Process -FilePath $p -ArgumentList '--one-click-elevated' -Verb RunAs -WindowStyle Hidden", quoted)
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", ps)
	cmd.Env = append(os.Environ(), "JAMSHIDIX_ELEVATED=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("administrator elevation failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return errOneClickRelaunch
}

func openConfigFolder() {
	_ = exec.Command("explorer.exe", dataDir).Start()
}

func waitForProcess(seconds int) bool {
	deadline := time.Now().Add(time.Duration(seconds) * time.Second)
	for time.Now().Before(deadline) {
		if isSingBoxRunning() {
			return true
		}
		time.Sleep(250 * time.Millisecond)
	}
	return false
}
