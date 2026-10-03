// Jamshidix v0.3.0
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func main() {
	if runtime.GOOS != "windows" {
		fatal(errors.New("Jamshidix supports Windows only"))
	}
	if len(os.Args) < 2 || os.Args[1] == "--gui-elevated" {
		fatal(runGUI())
		return
	}
	if os.Args[1] == "--one-click-elevated" {
		fatal(runOneClick())
		return
	}
	attachParentConsole() // the GUI-subsystem build has no console of its own
	args := os.Args[2:]
	switch os.Args[1] {
	case "version":
		fmt.Printf("Jamshidix %s | sing-box %s\n", appVersion, singBoxVersion)
	case "install":
		fatal(installSingBox())
	case "import":
		fatal(importCmd(args))
	case "config":
		fatal(configCmd(args))
	case "check":
		fatal(checkConfig())
	case "run":
		fatal(runTunnel())
	case "stop":
		fatal(stopSingBox())
	case "status":
		fatal(status())
	case "autostart":
		fatal(autostartCmd(args))
	case "uninstall":
		fatal(uninstall())
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println("Jamshidix", appVersion)
	fmt.Println()
	fmt.Println("Double-click Jamshidix.exe to connect (or disconnect).")
	fmt.Println()
	fmt.Println("  Jamshidix.exe import <vless://...>   save a server link as the client config")
	fmt.Println("  Jamshidix.exe import --clipboard     same, from the clipboard")
	fmt.Println("  Jamshidix.exe import --file <path>   same, from a text file")
	fmt.Println("  Jamshidix.exe config --server-ip IP --uuid UUID --public-key KEY --short-id ID --handshake-host HOST [--port 443]")
	fmt.Println("  Jamshidix.exe install | check | run | stop | status | version")
	fmt.Println("  Jamshidix.exe autostart on|off       start the tunnel at boot (needs Administrator)")
	fmt.Println("  Jamshidix.exe uninstall              stop, remove autostart, sing-box and config (needs Administrator)")
	fmt.Println()
	fmt.Println("Kill-switch remains available as PowerShell scripts in client/windows.")
}

// saveProfile renders, writes and (when sing-box is present) validates client.json.
func saveProfile(p Profile) error {
	cfg, err := renderClientConfig(p)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	path := configPath()
	previous, prevErr := os.ReadFile(path)
	if err := os.WriteFile(path, cfg, 0o600); err != nil {
		return err
	}
	if _, err := os.Stat(singBoxPath()); err == nil {
		if err := runCommand(singBoxPath(), "check", "-c", path); err != nil {
			if prevErr == nil {
				_ = os.WriteFile(path, previous, 0o600)
			} else {
				_ = os.Remove(path)
			}
			return fmt.Errorf("sing-box validation failed: %w", err)
		}
	}
	return nil
}

func importCmd(args []string) error {
	var text string
	switch {
	case len(args) == 1 && args[0] == "--clipboard":
		t, err := readClipboard()
		if err != nil {
			return err
		}
		text = t
	case len(args) == 2 && args[0] == "--file":
		b, err := os.ReadFile(args[1])
		if err != nil {
			return err
		}
		text = string(b)
	case len(args) == 1:
		text = args[0]
	default:
		return errors.New("usage: import <vless://link> | --clipboard | --file <path>")
	}
	link := findLink(text)
	if link == "" {
		return errors.New("no vless:// link found")
	}
	p, err := parseVLESSLink(link)
	if err != nil {
		return err
	}
	if err := saveProfile(p); err != nil {
		return err
	}
	fmt.Println("Client configuration created:", configPath())
	return nil
}

func configCmd(args []string) error {
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var p Profile
	fs.StringVar(&p.Server, "server-ip", "", "")
	fs.StringVar(&p.UUID, "uuid", "", "")
	fs.StringVar(&p.PublicKey, "public-key", "", "")
	fs.StringVar(&p.ShortID, "short-id", "", "")
	fs.StringVar(&p.SNI, "handshake-host", "", "")
	fs.IntVar(&p.Port, "port", 443, "")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for name, v := range map[string]string{"server-ip": p.Server, "uuid": p.UUID, "public-key": p.PublicKey, "short-id": p.ShortID, "handshake-host": p.SNI} {
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("missing --%s", name)
		}
	}
	if err := saveProfile(p); err != nil {
		return err
	}
	fmt.Println("Client configuration created:", configPath())
	return nil
}

func checkConfig() error {
	if _, err := os.Stat(configPath()); err != nil {
		return fmt.Errorf("client config not found: %s", configPath())
	}
	return runCommand(singBoxPath(), "check", "-c", configPath())
}

func runTunnel() error {
	if err := checkConfig(); err != nil {
		return err
	}
	fmt.Println("Starting TUN in the foreground (Administrator required). Press Ctrl+C to stop.")
	cmd := exec.Command(singBoxPath(), "run", "-c", configPath())
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	return cmd.Run()
}

func status() error {
	fmt.Println("Jamshidix:", appVersion)
	fmt.Println("sing-box :", singBoxPath())
	fmt.Println("config   :", configPath())
	fmt.Println("log      :", logPath())
	if _, err := os.Stat(configPath()); err == nil {
		fmt.Println("config status: present")
	} else {
		fmt.Println("config status: missing")
	}
	if isSingBoxRunning() {
		fmt.Println("tunnel status: RUNNING")
	} else {
		fmt.Println("tunnel status: stopped")
	}
	return nil
}

func autostartCmd(args []string) error {
	if len(args) != 1 || (args[0] != "on" && args[0] != "off") {
		return errors.New("usage: autostart on|off")
	}
	return setAutostart(args[0] == "on")
}

func uninstall() error {
	_ = stopSingBox()
	_ = setAutostart(false)
	for _, dir := range []string{programDir, dataDir} {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}
	fmt.Println("Jamshidix components removed. Delete Jamshidix.exe manually if desired.")
	return nil
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func fatal(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}
