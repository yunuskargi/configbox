# Changelog

All notable changes to ConfigBox are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2026-06-06

### Added
- **Brocade vendor support** — VDX (NOS), ICX (FastIron), MLX (NetIron)
- **Extreme Networks SLX vendor support** — SLX-OS
- **Location-based device filtering** alongside the vendor filter on the Devices page
- **Batched backup notifications** — when multiple devices back up within a 3-minute window, results are combined into a single summary email instead of N separate ones
- **System openssh fallback** — automatically falls back to the system openssh client (via sshpass) when Go's SSH library fails to interoperate with legacy SSH servers (e.g. OpenSSH 6.x preauth crash)
- **Keyboard-interactive SSH auth** — Juniper QFX/EX and other devices that only accept keyboard-interactive auth (not plain password) now work out of the box
- **CRUD feedback toasts** on device add/edit/delete
- **Demo GIF** in the README for a quick visual overview

### Fixed
- **Config change detection** — `getPreviousConfig()` was called after writing the new file, so it always compared the file to itself and reported "no change". Now reads previous config before writing.
- **Trailing whitespace in device credentials** — auto-trimmed on save to prevent silent SSH auth failures from copy-paste
- **SSH output truncation on slow devices** — added settle detection: the session waits for the device's stdout stream to go idle for 3 seconds before sending exit, preventing the previous fixed 2-second cutoff from truncating long configs
- **Email status pill layout** — S3 / Google Drive status badges now render as inline-block pills with `white-space: nowrap`, no more awkward wrapping in narrow cells
- **Dockerfile Go version** updated to `golang:1.25-alpine` to match `go.mod` requirement

## [1.0.0] - 2026-05-15

Initial public release.

### Features
- Automated config backups for FortiGate, Cisco (IOS/NX-OS/ASA), Juniper, and Palo Alto
- Web UI with dashboard, device management, backup history, and built-in config diff viewer
- Scheduled (cron-based) and manual one-click backups
- Email notifications: success / failure / config change / daily summary
- Remote backup to S3-compatible storage and Google Drive
- Automatic gzip archival of old backups
- Two-factor authentication (TOTP)
- Role-based access control (Admin / Backup Admin)
- Comprehensive audit log
- AES-256-CBC encrypted credentials
- Rate limiting on auth and download endpoints
- Single-use download tokens
- Multi-language UI (English & Turkish)
- Dark mode / light mode
- Docker Compose deployment, ~30 MB image
