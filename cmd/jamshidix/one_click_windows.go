//go:build windows

package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	mbOK          = 0x00000000
	mbYesNo       = 0x00000004
	mbError       = 0x00000010
	mbQuestion    = 0x00000020
	mbWarning     = 0x00000030
	mbInfo        = 0x00000040
	idYes         = 6
	createNoWin   = 0x08000000
	detachedProc  = 0x00000008
	newProcGroup  = 0x00000200
	attachParent  = ^uintptr(0) // ATTACH_PARENT_PROCESS (-1)
	probeURL      = "http://connectivitycheck.gstatic.com/generate_204"
	startupWaitMs = 3000
)

var (
	user32        = syscall.NewLazyDLL("user32.dll")
	shell32       = syscall.NewLazyDLL("shell32.dll")
	kernel32      = syscall.NewLazyDLL("kernel32.dll")
	messageBoxW   = user32.NewProc("MessageBoxW")
	shellExecuteW = shell32.NewProc("ShellExecuteW")
	isUserAnAdmin = shell32.NewProc("IsUserAnAdmin")
	attachConsole = kernel32.NewProc("AttachConsole")
)

func msg(body string, flags uintptr) int {
	t, _ := syscall.UTF16PtrFromString("Jamshidix")
	b, _ := syscall.UTF16PtrFromString(body)
	r, _, _ := messageBoxW.Call(0, uintptr(unsafe.Pointer(b)), uintptr(unsafe.Pointer(t)), flags)
	return int(r)
}

// attachParentConsole lets CLI subcommands print when the EXE (built with
// -H=windowsgui) is started from cmd.exe or PowerShell.
func attachParentConsole() {
	if r, _, _ := attachConsole.Call(attachParent); r == 0 {
		return
	}
	if _, err := os.Stdout.Stat(); err != nil {
		if f, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0); err == nil {
			os.Stdout = f
		}
	}
	if _, err := os.Stderr.Stat(); err != nil {
		if f, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0); err == nil {
			os.Stderr = f
		}
	}
}

func isAdmin() bool {
	r, _, _ := isUserAnAdmin.Call()
	return r != 0
}

// relaunchElevated re-runs this EXE through the UAC consent prompt.
func relaunchElevatedArgs(args []string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	verb, _ := syscall.UTF16PtrFromString("runas")
	file, _ := syscall.UTF16PtrFromString(exe)
	dir, _ := syscall.UTF16PtrFromString(filepath.Dir(exe))
	argText, _ := syscall.UTF16PtrFromString(strings.Join(args, " "))
	r, _, callErr := shellExecuteW.Call(0, uintptr(unsafe.Pointer(verb)), uintptr(unsafe.Pointer(file)), uintptr(unsafe.Pointer(argText)), uintptr(unsafe.Pointer(dir)), 1)
	if r <= 32 {
		return fmt.Errorf("UAC elevation was cancelled or failed (code %d): %v", r, callErr)
	}
	return nil
}

func relaunchElevated() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	verb, _ := syscall.UTF16PtrFromString("runas")
	file, _ := syscall.UTF16PtrFromString(exe)
	dir, _ := syscall.UTF16PtrFromString(filepath.Dir(exe))
	r, _, callErr := shellExecuteW.Call(0, uintptr(unsafe.Pointer(verb)), uintptr(unsafe.Pointer(file)), 0, uintptr(unsafe.Pointer(dir)), 1)
	if r <= 32 {
		return fmt.Errorf("UAC elevation was cancelled or failed (code %d): %v", r, callErr)
	}
	return nil
}

func runOneClick() error {
	if !isAdmin() {
		if err := relaunchElevated(); err != nil {
			msg("برای ساخت تونل، Jamshidix باید با دسترسی Administrator اجرا شود.\n\n"+err.Error(), mbError)
		}
		return nil
	}

	if isSingBoxRunning() {
		if msg("تونل Jamshidix در حال اجراست.\n\nبرای قطع اتصال «Yes» را بزنید.", mbYesNo|mbQuestion) == idYes {
			if err := stopSingBox(); err != nil {
				msg("قطع تونل انجام نشد.\n\n"+err.Error(), mbError)
			} else {
				msg("تونل قطع شد.", mbInfo)
			}
		}
		return nil
	}

	if !existingSingBoxIsExpected() {
		if err := installSingBox(); err != nil {
			msg("نصب sing-box انجام نشد.\n\n"+err.Error(), mbError)
			return nil
		}
	}

	imported, err := ensureClientConfig()
	if err != nil {
		msg(err.Error(), mbWarning)
		return nil
	}
	if err := checkConfigQuiet(); err != nil {
		msg("کانفیگ معتبر نیست.\n\n"+err.Error(), mbError)
		return nil
	}
	if err := startDetached(); err != nil {
		msg("اجرای تونل شروع نشد.\n\n"+err.Error(), mbError)
		return nil
	}
	time.Sleep(startupWaitMs * time.Millisecond)
	if !isSingBoxRunning() {
		msg("sing-box بلافاصله متوقف شد.\n\n"+tailLog(700)+"\n\nلاگ کامل: "+logPath(), mbError)
		return nil
	}

	note := ""
	if imported {
		note = "\n\nلینک سرور از کلیپ‌بورد/فایل وارد شد."
	}
	if probeInternet() {
		msg("متصل شد و اینترنت از مسیر تونل پاسخ می‌دهد.\n\nبرای قطع اتصال، دوباره روی Jamshidix.exe کلیک کنید."+note, mbInfo)
	} else {
		msg("تونل اجرا شد، اما از طریق سرور اینترنتی پاسخ نگرفتیم.\n\nاحتمالاً سرور خاموش/مسدود است یا کانفیگ اشتباه است.\nبرای قطع تونل دوباره روی Jamshidix.exe کلیک کنید.\n\nلاگ: "+logPath()+note, mbWarning)
	}
	return nil
}

func checkConfigQuiet() error {
	cmd := exec.Command(singBoxPath(), "check", "-c", configPath())
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWin}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// startDetached launches sing-box hidden; output goes to jamshidix.log.
func startDetached() error {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	logf, err := os.Create(logPath())
	if err != nil {
		return err
	}
	defer logf.Close()
	cmd := exec.Command(singBoxPath(), "run", "-c", configPath())
	cmd.Dir = dataDir
	cmd.Stdout, cmd.Stderr = logf, logf
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWin | detachedProc | newProcGroup}
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := os.WriteFile(pidPath(), []byte(strconv.FormatUint(uint64(cmd.Process.Pid), 10)), 0o600); err != nil {
		_ = cmd.Process.Kill()
		return err
	}
	return cmd.Process.Release()
}

func tailLog(n int) string {
	b, err := os.ReadFile(logPath())
	if err != nil || len(b) == 0 {
		return "(بدون خروجی)"
	}
	if len(b) > n {
		b = b[len(b)-n:]
	}
	return strings.TrimSpace(string(b))
}

// probeInternet checks that traffic really flows through the tunnel.
func probeInternet() bool {
	client := &http.Client{Timeout: 7 * time.Second}
	for i := 0; i < 3; i++ {
		resp, err := client.Get(probeURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
				return true
			}
		}
		time.Sleep(2 * time.Second)
	}
	return false
}

func singBoxPID() (int, bool) {
	b, err := os.ReadFile(pidPath())
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil || pid <= 0 {
		return 0, false
	}
	return pid, true
}

func isSingBoxRunning() bool {
	pid, ok := singBoxPID()
	if !ok {
		return false
	}
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/FO", "CSV", "/NH")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWin}
	out, err := cmd.CombinedOutput()
	return err == nil && strings.Contains(strings.ToLower(string(out)), "sing-box.exe")
}

func stopSingBox() error {
	pid, ok := singBoxPID()
	if !ok {
		return nil
	}
	cmd := exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/F", "/T")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWin}
	out, err := cmd.CombinedOutput()
	if err != nil {
		if m := strings.TrimSpace(string(out)); m != "" {
			return errors.New(m)
		}
		return err
	}
	_ = os.Remove(pidPath())
	return nil
}

func existingSingBoxIsExpected() bool {
	if _, err := os.Stat(singBoxPath()); err != nil {
		return false
	}
	cmd := exec.Command(singBoxPath(), "version")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWin}
	out, err := cmd.CombinedOutput()
	return err == nil && strings.Contains(string(out), singBoxVersion)
}

func readClipboard() (string, error) {
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command",
		"[Console]::OutputEncoding=[Text.Encoding]::UTF8; Get-Clipboard -Raw")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWin}
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("cannot read clipboard: %w", err)
	}
	return string(out), nil
}

func readLinkFile(path string) (Profile, string, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Profile{}, "", false
	}
	link := findLink(string(b))
	if link == "" {
		return Profile{}, "", false
	}
	p, err := parseVLESSLink(link)
	return p, link, err == nil
}

func storedLinkPath() string { return filepath.Join(dataDir, "link.txt") }

// ensureClientConfig picks the config source, in this order:
//  1. a new vless:// link on the clipboard (different from the stored one),
//  2. the existing %ProgramData%\Jamshidix\client.json,
//  3. link.txt / vless.txt / client.json next to the EXE.
//
// The bool reports whether a link was imported on this run.
func ensureClientConfig() (bool, error) {
	if text, err := readClipboard(); err == nil {
		if link := findLink(text); link != "" {
			stored, _ := os.ReadFile(storedLinkPath())
			if strings.TrimSpace(string(stored)) != link {
				if p, err := parseVLESSLink(link); err == nil {
					if err := saveProfile(p); err != nil {
						return false, err
					}
					_ = os.WriteFile(storedLinkPath(), []byte(link), 0o600)
					return true, nil
				}
			}
		}
	}
	if _, err := os.Stat(configPath()); err == nil {
		return false, nil
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		for _, name := range []string{"link.txt", "vless.txt"} {
			if p, link, ok := readLinkFile(filepath.Join(dir, name)); ok {
				if err := saveProfile(p); err != nil {
					return false, err
				}
				_ = os.WriteFile(storedLinkPath(), []byte(link), 0o600)
				return true, nil
			}
		}
		if cand := filepath.Join(dir, "client.json"); fileExists(cand) {
			if err := os.MkdirAll(dataDir, 0o755); err != nil {
				return false, err
			}
			b, err := os.ReadFile(cand)
			if err != nil {
				return false, err
			}
			return true, os.WriteFile(configPath(), b, 0o600)
		}
	}
	return false, errors.New("هنوز سروری تنظیم نشده است.\n\n۱) لینک vless:// سرور را کپی کنید\n۲) دوباره روی Jamshidix.exe کلیک کنید\n\n(یا لینک را در فایل link.txt کنار EXE بگذارید.)")
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func setAutostart(on bool) error {
	var cmd *exec.Cmd
	if on {
		if !fileExists(configPath()) {
			return fmt.Errorf("client config not found: %s (run Jamshidix.exe once first)", configPath())
		}
		tr := fmt.Sprintf(`"%s" run -c "%s"`, singBoxPath(), configPath())
		cmd = exec.Command("schtasks", "/Create", "/TN", "Jamshidix", "/TR", tr, "/SC", "ONSTART", "/RU", "SYSTEM", "/RL", "HIGHEST", "/F")
	} else {
		cmd = exec.Command("schtasks", "/Delete", "/TN", "Jamshidix", "/F")
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWin}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("schtasks failed (run as Administrator): %s", strings.TrimSpace(string(out)))
	}
	fmt.Println("Autostart:", map[bool]string{true: "enabled", false: "disabled"}[on])
	return nil
}
