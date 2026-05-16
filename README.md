# ConfBox

Network cihazlarinin konfigurasyon yedeklerini otomatik alan, dosya olarak saklayan acik kaynakli web uygulamasi.

![Dashboard](https://img.shields.io/badge/stack-FastAPI%20%2B%20React-blue)
![License](https://img.shields.io/badge/license-AGPL--3.0-green)
![Docker](https://img.shields.io/badge/deploy-Docker%20Compose-blue)

## Desteklenen Cihazlar

| Vendor | Protokol | Detay |
|--------|----------|-------|
| **FortiGate** | REST API | Config backup via `/api/v2/monitor/system/config/backup` |
| **Juniper** | SSH (Netmiko) | `show configuration \| display set` |
| **Cisco** (IOS/NX-OS/ASA) | SSH (Netmiko) | `show running-config` |
| **Palo Alto** | PAN-OS XML API | Config export via XML API |

## Ozellikler

- Otomatik zamanlanmis yedekleme (cron tabanli)
- Manuel tek tusla backup
- Config diff / karsilastirma
- CSV ile toplu cihaz aktarimi
- Dashboard istatistikleri ve trend grafikleri
- Lokasyon bazli cihaz yonetimi
- E-posta bildirimleri (basarili/basarisiz backup)
- Dark mode / gece modu
- Rol tabanli erisim (Admin / Backup Admin)
- 2FA (TOTP) destegi
- Audit log
- Sifrelenmis kimlik bilgileri (AES-256-CBC)
- Rate limiting
- Duz dosya saklama -- uygulama cokse bile `backups/` dizinine gidip dosyalara erisilebilir

## Hizli Kurulum

### Gereksinimler
- Docker & Docker Compose

### 1. Repoyu klonlayin

```bash
git clone https://github.com/yunuskargi/confbox.git
cd confbox
```

### 2. Ortam degiskenlerini ayarlayin

```bash
cp .env.example .env
# .env dosyasindaki JWT_SECRET degerini degistirin!
```

### 3. Calistirin

```bash
docker compose up -d
```

Uygulama `http://localhost` adresinde calisacaktir.

### 4. Giris yapin

- **Kullanici:** `admin`
- **Sifre:** `admin`

> Ilk giristen sonra sifrenizi degistirmeniz onerilir.

## Yedek Dosya Yapisi

```
backups/
├── fortigate/
│   └── cihaz-adi/
│       ├── 2024-01-15_020000.conf
│       └── 2024-01-16_020000.conf
├── juniper/
├── cisco/
└── paloalto/
```

## API Dokumantasyonu

Backend calistiktan sonra Swagger UI'a erisebilirsiniz:

```
http://localhost:8000/docs
```

## Lisans

Bu proje [AGPL-3.0](LICENSE) lisansi altinda yayinlanmistir.

## Katki

Pull request'ler ve issue'lar memnuniyetle karsilanir. Buyuk degisiklikler icin once bir issue aciniz.
