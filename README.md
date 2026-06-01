<div align="center">

# ConfigBox

### Modern, open-source network configuration backup manager

Automated config backups for FortiGate, Cisco, Juniper & Palo Alto — with a clean web UI, diff viewer, scheduling, and Docker deployment in minutes.

A free alternative to **RANCID**, **Oxidized**, and **SolarWinds NCM**.

[![License](https://img.shields.io/badge/license-AGPL--3.0-green.svg)](LICENSE)
[![Stack](https://img.shields.io/badge/stack-Go%20%2B%20React-00ADD8.svg)](#tech-stack)
[![Docker](https://img.shields.io/badge/deploy-Docker%20Compose-2496ED.svg)](#quick-start)
[![GitHub Stars](https://img.shields.io/github/stars/yunuskargi/configbox?style=flat&color=yellow)](https://github.com/yunuskargi/configbox/stargazers)
[![GitHub Issues](https://img.shields.io/github/issues/yunuskargi/configbox.svg)](https://github.com/yunuskargi/configbox/issues)
[![Last Commit](https://img.shields.io/github/last-commit/yunuskargi/configbox.svg)](https://github.com/yunuskargi/configbox/commits/main)

[Quick Start](#quick-start) · [Features](#features) · [Supported Devices](#supported-devices) · [Updating](#updating--upgrading)

![ConfigBox Demo](docs/demo.gif)

</div>

## Why ConfigBox?

Tools like **RANCID** and **Oxidized** have been around for years, but they show their age — CLI-only, hard to install, no built-in user management, no notifications. **SolarWinds NCM** solves these but costs thousands per year.

ConfigBox is the modern alternative: a single Docker command to deploy, a clean web UI for everyone on the team, built-in 2FA, email alerts, and S3/Google Drive sync — all free and open-source.

### How does it compare?

| Feature | **ConfigBox** | RANCID | Oxidized | SolarWinds NCM |
|---|:---:|:---:|:---:|:---:|
| Modern web UI | ✅ | ❌ | ⚠️ basic | ✅ |
| Docker single-command install | ✅ | ❌ | ⚠️ | ❌ |
| Config diff viewer (built-in) | ✅ | via CVS | via Git | ✅ |
| Scheduled + manual backups | ✅ | ✅ | ✅ | ✅ |
| Email notifications | ✅ | ⚠️ basic | ⚠️ basic | ✅ |
| Role-based access control | ✅ | ❌ | ❌ | ✅ |
| Two-factor authentication | ✅ | ❌ | ❌ | ✅ |
| Audit log | ✅ | ❌ | ❌ | ✅ |
| Remote backup (S3 / Google Drive) | ✅ | ❌ | ❌ | ❌ |
| Multi-language UI | ✅ EN/TR | ❌ | ❌ | ⚠️ |
| Dark mode | ✅ | ❌ | ❌ | ❌ |
| **License** | AGPL-3.0 | BSD | Apache-2.0 | Proprietary |
| **Price** | Free | Free | Free | $$$$ |

## Supported Devices

| Vendor | Protocol | Detail |
|--------|----------|--------|
| **FortiGate** | REST API | Config backup via `/api/v2/monitor/system/config/backup` |
| **Juniper** | SSH | `show configuration | display set` |
| **Cisco** (IOS/NX-OS/ASA) | SSH | `show running-config` |
| **Palo Alto** | PAN-OS XML API | Config export via XML API |

## Features

### 🔄 Backup & Storage
- **Automated scheduled backups** — cron-based, per-device schedules
- **One-click manual backup** from the dashboard
- **Built-in config diff** — compare any two backups side-by-side
- **Remote backup** to S3-compatible storage (AWS, MinIO, R2, B2) or Google Drive
- **Automatic archival** — gzip compression of old backups to save disk
- **Plain file storage** — even if the app stops, configs are readable in `backups/`
- **CSV bulk import** — onboard hundreds of devices in seconds

### 🔐 Security
- **Two-factor authentication (TOTP)** for all users
- **AES-256-CBC encrypted credentials** (API tokens, SSH passwords, SMTP)
- **Role-based access control** (Admin / Backup Admin)
- **Comprehensive audit log** — every action tracked with user, IP, timestamp
- **Rate limiting** on auth endpoints and downloads
- **Single-use download tokens** — backup files cannot be re-fetched with a leaked URL
- **SSRF / gzip-bomb / path-traversal protection**

### 📊 Monitoring & Notifications
- **Dashboard** with statistics, trend charts, recent activity
- **Email notifications** — success / failure / config change / daily summary
- **Location-based device grouping** with filtering
- **Vendor + location filters** on the device list

### 🌐 Platform & UX
- **Multi-vendor support** — FortiGate, Cisco (IOS/NX-OS/ASA), Juniper, Palo Alto
- **Dark mode** / light mode
- **Multi-language UI** — English & Turkish
- **Modern web UI** built with React + Tailwind
- **Lightweight** — ~30 MB Docker image, single Go binary
- **Self-hosted** — your configs never leave your infrastructure

## Quick Start

### Requirements
- Docker & Docker Compose

### 1. Clone the repository

```bash
git clone https://github.com/yunuskargi/configbox.git
cd configbox
```

### 2. Configure environment variables

```bash
cp .env.example .env
# Change the JWT_SECRET value in .env!
```

### 3. Run

```bash
docker compose up -d
```

The application will be available at `http://localhost:6161`.

### 4. Login

- **Username:** `admin`
- **Password:** `admin`

> It is recommended to change your password after first login.

## CLI Commands

```bash
# Reset a user's password
docker compose exec backend /configbox reset-password <username> <new-password>
```

## Backup File Structure

```
backups/
├── fortigate/
│   └── device-name/
│       ├── 2024-01-15_020000.conf
│       └── 2024-01-16_020000.conf
├── juniper/
├── cisco/
└── paloalto/
```

## Remote Backup (S3 / Google Drive)

ConfigBox can automatically upload a copy of each backup to S3-compatible storage (AWS, MinIO, Cloudflare R2, Backblaze B2) or Google Drive. Configure via **Settings → Remote Backup** in the web UI — setup guides are included.

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Backend | Go (Chi router, sqlx, golang.org/x/crypto/ssh) |
| Frontend | React + Vite |
| Database | SQLite (WAL mode) |
| Auth | JWT + bcrypt + TOTP |
| Encryption | AES-256-CBC |
| Scheduler | robfig/cron |

## Updating / Upgrading

Your data is safe during updates:
- **Database** → stored in Docker named volume `db-data`, persists across container rebuilds
- **Config backups** → stored in `./backups` bind mount on your host, untouched during updates
- **Schema** → uses `CREATE TABLE IF NOT EXISTS`, no manual migration needed

### Update Steps

```bash
cd configbox

# Pull latest source
git pull

# Rebuild and restart (containers are recreated automatically, data is preserved)
docker compose up -d --build
```

### Important Notes

> **Do NOT change `JWT_SECRET` in `.env` after initial setup.** All device credentials (API tokens, SSH passwords) are encrypted with this key. Changing it will make existing credentials unreadable — you would need to re-enter all device passwords.

> **Do NOT delete the `db-data` Docker volume.** It contains your SQLite database with all devices, users, backup history, and settings. If you need to check: `docker volume ls | grep db-data`

> **Backup your `.env` file** before updating. If you accidentally lose it, you lose your `JWT_SECRET` and encrypted credentials cannot be recovered.

## Security

- Default login is `admin/admin` — you will be asked to change it on first login
- All credentials (device passwords, API keys, SMTP) are encrypted in the database
- If you expose ConfigBox to the internet, put a reverse proxy with SSL in front (nginx, Traefik, Caddy)
- See `.env.example` for optional settings like `ENCRYPTION_KEY`, `TRUSTED_PROXY`, and `FORCE_HTTPS`

## License

This project is licensed under [AGPL-3.0](LICENSE).

## Contributing

Pull requests and issues are welcome. For major changes, please open an issue first.
