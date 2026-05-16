# ConfBox

Open-source web application that automatically backs up network device configurations and stores them as plain files.

![Dashboard](https://img.shields.io/badge/stack-FastAPI%20%2B%20React-blue)
![License](https://img.shields.io/badge/license-AGPL--3.0-green)
![Docker](https://img.shields.io/badge/deploy-Docker%20Compose-blue)

## Supported Devices

| Vendor | Protocol | Detail |
|--------|----------|--------|
| **FortiGate** | REST API | Config backup via `/api/v2/monitor/system/config/backup` |
| **Juniper** | SSH (Netmiko) | `show configuration \| display set` |
| **Cisco** (IOS/NX-OS/ASA) | SSH (Netmiko) | `show running-config` |
| **Palo Alto** | PAN-OS XML API | Config export via XML API |

## Features

- Automated scheduled backups (cron-based)
- One-click manual backup
- Config diff / comparison
- CSV bulk device import
- Dashboard statistics and trend charts
- Location-based device management
- Email notifications (success/failure)
- Dark mode / light mode
- Multi-language support (English & Turkish)
- Role-based access control (Admin / Backup Admin)
- Two-factor authentication (TOTP)
- Audit log
- Encrypted credentials (AES-256-CBC)
- Rate limiting
- Plain file storage — even if the app crashes, you can access configs directly from the `backups/` directory

## Quick Start

### Requirements
- Docker & Docker Compose

### 1. Clone the repository

```bash
git clone https://github.com/yunuskargi/confbox.git
cd confbox
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

## API Documentation

Once the backend is running, you can access Swagger UI:

```
http://localhost:8000/docs
```

## License

This project is licensed under [AGPL-3.0](LICENSE).

## Contributing

Pull requests and issues are welcome. For major changes, please open an issue first.
