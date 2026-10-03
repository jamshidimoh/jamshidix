# Jamshidix

> One-click free VPN for bypassing network restrictions.

## هدف پروژه

Jamshidix یک راه‌حل خودکار برای ساخت و فعال‌سازی تونل امن است که با استفاده از سرورهای رایگان OCI و sing-box طراحی شده است. کاربر فقط باید روی `Jamshidix.exe` کلیک کند تا برنامه:

- sing-box را نصب کند
- کانفیگ را از قالب بسازد
- سرور رایگان را فعال کند
- TUN را اجرا کند
- اتصال امن را فراهم کند

## معماری

```text
کاربر → Jamshidix.exe → ثبت/نصب sing-box → ساخت client.json → OCI VM → TUN → VLESS/REALITY
```

## مجموعهٔ فایل‌های اصلی

- `cmd/jamshidix/` : اپلیکیشن اصلی Go
- `server/` : اسکریپت‌های سرور Ubuntu
- `infra/oci/` : Terraform برای سرور رایگان OCI
- `config/` : قالب‌های کانفیگ
- `scripts/` : ابزارهای تست و اعتبارسنجی

## راه‌اندازی

```bash
git clone https://github.com/jamshidimoh/jamshidix.git
cd jamshidix

go build -o Jamshidix.exe -ldflags="-H windowsgui" ./cmd/jamshidix
```

## نکتهٔ مهم

طرح پروژه به‌صورتی تنظیم شده که فقط یک کلیک برای کاربر نهایی لازم باشد و در همین راستا، سرورها و کانفیگ‌ها با تمرکز بر گزینه‌های رایگان و خودکار سازماندهی شده‌اند.

