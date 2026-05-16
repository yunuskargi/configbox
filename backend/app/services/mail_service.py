import smtplib
import html
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
from email.header import Header
from app.database import SessionLocal
from app.models import Setting


def _get_smtp_settings() -> dict:
    db = SessionLocal()
    try:
        defaults = {
            "smtp_host": "", "smtp_port": "587", "smtp_username": "",
            "smtp_password": "", "smtp_use_tls": "true",
            "smtp_from_email": "", "smtp_from_name": "ConfBox",
        }
        result = {}
        for key, default in defaults.items():
            row = db.query(Setting).filter(Setting.key == key).first()
            result[key] = row.value if row else default
        result["smtp_port"] = int(result["smtp_port"])
        result["smtp_use_tls"] = result["smtp_use_tls"].lower() == "true"
        return result
    finally:
        db.close()


def _get_notify_settings() -> dict:
    db = SessionLocal()
    try:
        defaults = {
            "notify_on_success": "false", "notify_on_failure": "true",
            "notify_on_change": "false", "notify_daily_summary": "false",
            "notify_recipients": "",
        }
        result = {}
        for key, default in defaults.items():
            row = db.query(Setting).filter(Setting.key == key).first()
            result[key] = row.value if row else default
        for k in ("notify_on_success", "notify_on_failure", "notify_on_change", "notify_daily_summary"):
            result[k] = result[k].lower() == "true"
        return result
    finally:
        db.close()


VENDOR_LABELS = {
    "fortigate": "FortiGate",
    "juniper": "Juniper",
    "cisco": "Cisco",
    "paloalto": "Palo Alto",
}


def _vendor_label(vendor: str) -> str:
    return VENDOR_LABELS.get(vendor, vendor)


def _mail_wrapper(title: str, content: str, accent: str = "#2563eb") -> str:
    return f"""<!DOCTYPE html>
<html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"></head>
<body style="margin:0;padding:0;background:#f1f5f9;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,'Helvetica Neue',Arial,sans-serif">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background:#f1f5f9;padding:32px 16px">
<tr><td align="center">
<table role="presentation" width="560" cellpadding="0" cellspacing="0" style="max-width:560px;width:100%;background:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,0.06)">
  <tr><td style="background:linear-gradient(135deg,{accent},#1e40af);padding:28px 32px">
    <table role="presentation" width="100%" cellpadding="0" cellspacing="0">
      <tr>
        <td><span style="font-size:24px;font-weight:700;color:#ffffff;letter-spacing:-0.5px">{title}</span></td>
        <td align="right"><span style="font-size:11px;color:rgba(255,255,255,0.7);text-transform:uppercase;letter-spacing:1px">ConfBox</span></td>
      </tr>
    </table>
  </td></tr>
  <tr><td style="padding:28px 32px">{content}</td></tr>
  <tr><td style="padding:16px 32px;border-top:1px solid #f1f5f9">
    <table role="presentation" width="100%" cellpadding="0" cellspacing="0">
      <tr>
        <td><span style="color:#94a3b8;font-size:11px">ConfBox &mdash; Network Config Backup</span></td>
        <td align="right"><span style="color:#cbd5e1;font-size:10px">Otomatik bildirim</span></td>
      </tr>
    </table>
  </td></tr>
</table>
</td></tr></table>
</body></html>"""


def _info_row(label: str, value: str, bg: bool = False) -> str:
    bg_style = "background:#f8fafc;" if bg else ""
    return f'<tr><td style="{bg_style}padding:10px 16px;color:#64748b;font-size:13px;width:130px;border-bottom:1px solid #f1f5f9">{label}</td><td style="{bg_style}padding:10px 16px;font-size:13px;color:#1e293b;font-weight:500;border-bottom:1px solid #f1f5f9">{value}</td></tr>'


def _badge(text: str, color: str, bg: str) -> str:
    return f'<span style="display:inline-block;padding:4px 12px;background:{bg};color:{color};border-radius:20px;font-size:12px;font-weight:600">{text}</span>'


def send_email(to: str, subject: str, body_html: str):
    smtp = _get_smtp_settings()
    if not smtp["smtp_host"] or not to:
        return

    msg = MIMEMultipart("alternative")
    msg["From"] = f"{smtp['smtp_from_name']} <{smtp['smtp_from_email']}>"
    msg["To"] = to
    msg["Subject"] = Header(subject, "utf-8")
    msg.attach(MIMEText(body_html, "html", "utf-8"))

    server = smtplib.SMTP(smtp["smtp_host"], smtp["smtp_port"])
    if smtp["smtp_use_tls"]:
        server.starttls()
    if smtp["smtp_username"]:
        server.login(smtp["smtp_username"], smtp["smtp_password"])
    server.sendmail(smtp["smtp_from_email"], to.split(","), msg.as_string())
    server.quit()


def send_test_email(to: str):
    content = f"""
    <div style="text-align:center;padding:20px 0">
      <div style="width:64px;height:64px;margin:0 auto 16px;background:#dcfce7;border-radius:50%;display:flex;align-items:center;justify-content:center">
        <span style="font-size:28px">&#9989;</span>
      </div>
      <p style="margin:0;font-size:18px;font-weight:600;color:#16a34a">SMTP Bağlantısı Başarılı</p>
      <p style="margin:8px 0 0;color:#64748b;font-size:14px">Mail bildirimleri düzgün çalışıyor.</p>
    </div>
    """
    send_email(to, "ConfBox - Test Email", _mail_wrapper("Test Bildirimi", content, "#16a34a"))


def notify_backup(device_name: str, vendor: str, status: str, error: str = None,
                   file_path: str = None, file_size: int = 0, location: str = None,
                   vdom: str = None, triggered_by: str = "manual"):
    notify = _get_notify_settings()
    recipients = notify["notify_recipients"]
    if not recipients:
        return

    if status == "success" and not notify["notify_on_success"]:
        return
    if status == "failed" and not notify["notify_on_failure"]:
        return

    from datetime import datetime
    from app.config import APP_TIMEZONE
    now = datetime.now(APP_TIMEZONE).strftime("%Y-%m-%d %H:%M")

    is_ok = status == "success"
    status_text = "Başarılı" if is_ok else "Başarısız"
    status_color = "#16a34a" if is_ok else "#dc2626"
    status_bg = "#dcfce7" if is_ok else "#fee2e2"
    status_icon = "&#9989;" if is_ok else "&#10060;"
    accent = "#2563eb" if is_ok else "#dc2626"
    trigger_label = "Manuel" if triggered_by == "manual" else "Zamanlanmış"

    size_str = ""
    if file_size:
        if file_size > 1024 * 1024:
            size_str = f"{file_size / (1024*1024):.1f} MB"
        elif file_size > 1024:
            size_str = f"{file_size / 1024:.1f} KB"
        else:
            size_str = f"{file_size} B"

    status_banner = f"""
    <div style="background:{status_bg};border-radius:12px;padding:16px 20px;margin-bottom:20px;text-align:center">
      <span style="font-size:24px">{status_icon}</span>
      <div style="margin-top:8px;font-size:18px;font-weight:700;color:{status_color}">{status_text}</div>
      <div style="margin-top:4px;color:#64748b;font-size:13px">{html.escape(device_name)}</div>
    </div>
    """

    rows = _info_row("Cihaz", f"<strong>{html.escape(device_name)}</strong>")
    rows += _info_row("Vendor", _badge(_vendor_label(vendor), "#1e40af", "#eff6ff"), True)
    if vdom:
        rows += _info_row("VDOM", html.escape(vdom))
    if location:
        rows += _info_row("Lokasyon", html.escape(location), True)
    rows += _info_row("Tetikleyen", trigger_label)
    rows += _info_row("Zaman", now, True)
    if size_str:
        rows += _info_row("Dosya Boyutu", size_str)
    if file_path:
        rows += _info_row("Dosya Yolu", f'<span style="font-size:11px;word-break:break-all;color:#64748b">{html.escape(file_path)}</span>', True)

    error_html = ""
    if error:
        error_html = f"""
        <div style="margin-top:20px;padding:16px;background:#fef2f2;border:1px solid #fecaca;border-radius:12px">
          <div style="font-size:12px;font-weight:600;color:#dc2626;text-transform:uppercase;letter-spacing:0.5px;margin-bottom:8px">Hata Detayı</div>
          <div style="color:#991b1b;font-size:13px;line-height:1.5">{html.escape(error)}</div>
        </div>
        """

    content = f"""
    {status_banner}
    <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="border:1px solid #e2e8f0;border-radius:12px;overflow:hidden">
      {rows}
    </table>
    {error_html}
    """

    status_emoji = "✅" if is_ok else "❌"
    send_email(recipients, f"ConfBox {status_emoji} {device_name} - Backup {status_text}", _mail_wrapper("Backup Bildirimi", content, accent))


def notify_config_change(device_name: str, vendor: str, diff_text: str,
                          location: str = None, vdom: str = None):
    notify = _get_notify_settings()
    if not notify["notify_on_change"]:
        return
    recipients = notify["notify_recipients"]
    if not recipients:
        return

    info_parts = [f"<strong>{html.escape(device_name)}</strong>", _vendor_label(vendor)]
    if vdom:
        info_parts.append(f"VDOM: {html.escape(vdom)}")
    if location:
        info_parts.append(html.escape(location))

    diff_html = html.escape(diff_text) if diff_text else "Detay yok"

    content = f"""
    <div style="background:#fef3c7;border-radius:12px;padding:16px 20px;margin-bottom:20px;text-align:center">
      <span style="font-size:24px">&#9888;&#65039;</span>
      <div style="margin-top:8px;font-size:16px;font-weight:700;color:#b45309">Config Değişikliği Tespit Edildi</div>
      <div style="margin-top:4px;color:#92400e;font-size:13px">{' &bull; '.join(info_parts)}</div>
    </div>
    <div style="background:#0f172a;border-radius:12px;padding:20px;overflow-x:auto">
      <pre style="margin:0;font-size:12px;color:#e2e8f0;white-space:pre-wrap;line-height:1.6;font-family:'SF Mono',Monaco,Consolas,'Courier New',monospace">{diff_html}</pre>
    </div>
    """

    send_email(recipients, f"ConfBox ⚠️ {device_name} - Config Değişikliği", _mail_wrapper("Config Değişikliği", content, "#d97706"))


def send_daily_summary():
    notify = _get_notify_settings()
    if not notify["notify_daily_summary"]:
        return
    recipients = notify["notify_recipients"]
    if not recipients:
        return

    from datetime import datetime, timedelta
    from sqlalchemy import func
    from app.config import APP_TIMEZONE
    from app.models import Backup, Device

    db = SessionLocal()
    try:
        since = datetime.now(APP_TIMEZONE) - timedelta(hours=24)

        total = db.query(func.count(Backup.id)).filter(Backup.created_at >= since).scalar() or 0
        success = db.query(func.count(Backup.id)).filter(Backup.created_at >= since, Backup.status == "success").scalar() or 0
        failed = db.query(func.count(Backup.id)).filter(Backup.created_at >= since, Backup.status == "failed").scalar() or 0
        total_devices = db.query(func.count(Device.id)).filter(Device.is_active == True).scalar() or 0
        rate = round((success / total * 100), 1) if total > 0 else 0

        failed_list = (
            db.query(Backup)
            .join(Device)
            .filter(Backup.created_at >= since, Backup.status == "failed")
            .order_by(Backup.created_at.desc())
            .all()
        )

        now = datetime.now(APP_TIMEZONE).strftime("%Y-%m-%d")

        if total == 0:
            summary_icon = "&#128260;"
            summary_text = "Son 24 saatte backup işlemi yapılmadı"
            summary_color = "#64748b"
            summary_bg = "#f1f5f9"
        elif failed == 0:
            summary_icon = "&#9989;"
            summary_text = f"Tüm backup'lar başarılı!"
            summary_color = "#16a34a"
            summary_bg = "#dcfce7"
        else:
            summary_icon = "&#9888;&#65039;"
            summary_text = f"{failed} başarısız backup var"
            summary_color = "#dc2626"
            summary_bg = "#fee2e2"

        stat_card = lambda label, value, color: f"""
        <td style="padding:12px;text-align:center;width:25%">
          <div style="font-size:24px;font-weight:700;color:{color}">{value}</div>
          <div style="font-size:11px;color:#94a3b8;margin-top:4px;text-transform:uppercase;letter-spacing:0.5px">{label}</div>
        </td>"""

        content = f"""
        <div style="background:{summary_bg};border-radius:12px;padding:16px 20px;margin-bottom:24px;text-align:center">
          <span style="font-size:24px">{summary_icon}</span>
          <div style="margin-top:8px;font-size:16px;font-weight:700;color:{summary_color}">{summary_text}</div>
          <div style="margin-top:4px;color:#64748b;font-size:12px">{now}</div>
        </div>

        <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="border:1px solid #e2e8f0;border-radius:12px;overflow:hidden;margin-bottom:20px">
          <tr style="background:#f8fafc">
            {stat_card("Aktif Cihaz", total_devices, "#1e293b")}
            {stat_card("Toplam", total, "#1e293b")}
            {stat_card("Başarılı", success, "#16a34a")}
            {stat_card("Başarısız", failed, "#dc2626" if failed > 0 else "#94a3b8")}
          </tr>
          <tr>
            <td colspan="4" style="padding:12px 16px;text-align:center;border-top:1px solid #e2e8f0">
              <div style="background:#e2e8f0;border-radius:20px;height:8px;overflow:hidden">
                <div style="background:#16a34a;height:100%;width:{rate}%;border-radius:20px"></div>
              </div>
              <div style="margin-top:6px;font-size:12px;color:#64748b">Başarı oranı: <strong style="color:#1e293b">{rate}%</strong></div>
            </td>
          </tr>
        </table>
        """

        if failed_list:
            rows = ""
            for i, b in enumerate(failed_list[:10]):
                err = html.escape(b.error_message or "Bilinmiyor")[:80]
                bg = "background:#fef2f2;" if i % 2 == 0 else ""
                rows += f'<tr><td style="{bg}padding:8px 14px;font-size:13px;font-weight:500;color:#1e293b;border-bottom:1px solid #fecaca">{html.escape(b.device.name)}</td><td style="{bg}padding:8px 14px;font-size:12px;color:#991b1b;border-bottom:1px solid #fecaca">{err}</td></tr>'
            content += f"""
            <div style="margin-top:4px">
              <div style="font-size:13px;font-weight:600;color:#dc2626;margin-bottom:8px">Başarısız Backup'lar</div>
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="border:1px solid #fecaca;border-radius:12px;overflow:hidden">
                <tr style="background:#fef2f2"><th style="padding:8px 14px;text-align:left;font-size:11px;color:#991b1b;text-transform:uppercase;letter-spacing:0.5px">Cihaz</th><th style="padding:8px 14px;text-align:left;font-size:11px;color:#991b1b;text-transform:uppercase;letter-spacing:0.5px">Hata</th></tr>
                {rows}
              </table>
            </div>
            """

        send_email(recipients, f"ConfBox 📊 Günlük Özet - {now}", _mail_wrapper("Günlük Backup Özeti", content, "#2563eb"))
    finally:
        db.close()
