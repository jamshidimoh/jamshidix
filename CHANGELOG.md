# Changelog

## 0.2.0 — 2026-10-03

- **Single-click for real:** paste/copy a `vless://` link, double-click the EXE. The link is imported from the clipboard or `link.txt`; no manual `client.json`.
- **Self-contained EXE:** release builds embed the official sing-box (SHA-256 pinned), so no download on first run. The lite build still downloads with retries and timeouts.
- Second click offers to disconnect; after start the EXE verifies that traffic really flows through the tunnel and logs to `%ProgramData%\Jamshidix\jamshidix.log`.
- Fixed: `status`/`stop`/etc. printed nothing from the GUI-subsystem build (now attaches to the parent console).
- Fixed: elevation no longer relies on an environment variable (risk of a UAC relaunch loop); uses `IsUserAnAdmin` + `ShellExecute runas`.
- Config rendering is now structured JSON editing with strict validation (UUID, key, short id, host, flow, fingerprint) instead of text substitution.
- New CLI: `import`, `autostart on|off`, `uninstall`, `--port`.
- Server: `generate_server_config.sh` prints a ready `vless://` link and refuses to overwrite keys without `FORCE=1`; `install_ubuntu.sh` now opens TCP/443 in iptables (OCI images block it by default) and no longer enables IP forwarding/NAT, which a user-space proxy does not need; dropped `CAP_NET_ADMIN`.
- Removed the unused deprecated `block` outbound from the server template.
- Tests: link parsing/rendering, template-sync, real `sing-box check`. CI also vets the Windows build and publishes a Release on `v*` tags with a SHA-256 file.
- Cloudflare test uses a scoped API Token instead of the Global API Key; secret audit also catches committed `vless://` links.

## 0.1.1 — 2026-09-30

- One-click Windows launcher: automatic elevation, sing-box check/install, config discovery, validation and background TUN start.
- GUI subsystem build so double-click does not open a console window.


## 0.1.0 — 2026-09-30

- Free-tier-first OCI A1 provisioning.
- sing-box 1.14.1 with VLESS + REALITY.
- Windows TUN, DNS control, kill-switch and autostart tooling.
- SHA-256 verification for official sing-box binaries.
- CI validation for JSON, shell, PowerShell, secrets, sing-box and Terraform.
- No paid service is required by the repository design.

This release is CI-validated. End-to-end availability from a real Iranian network depends on the external free-tier VM, provider capacity and network conditions.
