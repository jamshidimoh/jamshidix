# Jamshidix

> **یک‌کلیک VPN برای عبور از فیلترهای شبکه، بدون هزینه‌ٔ ماهانه.**
>
> One-click VPN for bypassing network restrictions without monthly cost.

## ✨ ویژگی‌ها

- یک‌کلیک: فقط روی `Jamshidix.exe` دوبار کلیک کنید.
- خودکار: نصب sing-box، تولید کانفیگ، و فعال‌سازی TUN بدون فرمان دستی.
- رایگان: سرور روی OCI Always Free با هزینه صفر.
- امن: VLESS + REALITY.
- فارسی: رابط‌ها و پیام‌های تولیدی به فارسی هستند.

## 🚀 شروع سریع

### برای کاربر نهایی (ویندوز)

1. `Jamshidix.exe` را دانلود کنید.
2. روی فایل دوبار کلیک کنید.
3. در اولین اجرا، برنامه در صورت نیاز Administrator می‌گیرد.
4. اگر sing-box نصب نشده باشد، به‌صورت خودکار نصب می‌شود.
5. اگر `client.json` وجود نداشته باشد، از قالب ساخته می‌شود.
6. TUN شروع می‌شود و اتصال فعال می‌شود.

### برای توسعه‌دهنده

```bash
git clone https://github.com/jamshidimoh/jamshidix.git
cd jamshidix

go build -o Jamshidix.exe -ldflags="-H windowsgui" ./cmd/jamshidix
```

## 📁 ساختار پروژه

```text
jamshidix/
├── cmd/jamshidix/            # برنامه اصلی Go
│   ├── main.go               # ورودی اصلی
│   ├── ui.go                 # Wizard تنظیم اولیه
│   ├── installer.go          # نصب sing-box
│   ├── provisioner.go        # ساخت سرور OCI
│   ├── config_engine.go      # تولید client.json
│   ├── README.md             # مستندات اجرای کلیک‌دار
│   └── assets/
│       └── client.config.template.json
├── server/                   # اسکریپت‌های سرور
│   ├── install_ubuntu.sh
│   ├── generate_server_config.sh
│   ├── install_cloudflared.sh
│   └── README.md
├── infra/oci/                # Terraform برای سرور رایگان
│   ├── main.tf
│   ├── variables.tf
│   ├── outputs.tf
│   ├── versions.tf
│   ├── README.md
│   └── terraform.tfvars.example
├── config/                   # قالب‌های کانفیگ
│   ├── client.config.template.json
│   └── server.config.template.json
├── scripts/                  # اعتبارسنجی و ابزارهای CI
│   ├── validate_singbox.sh
│   ├── generate_ids.sh
│   ├── secret_audit.sh
│   └── cloudflare_connection_test.sh
├── docs/                     # مستندات پروژه
│   ├── SETUP_GUIDE.md
│   ├── ARCHITECTURE.md
│   └── FAQ.md
├── .gitignore
├── CHANGELOG.md
├── LICENSE
├── VERSION
├── README.md
└── go.mod
```

## 🧩 معماری نهایی

```text
کاربر روی Jamshidix.exe دوبار کلیک می‌کند
           ↓
   بررسی / نصب sing-box
           ↓
   پیدا کردن یا ساخت client.json
           ↓
   اگر سرور وجود ندارد:
      - OCI Always Free VM ساخته شود
      - sing-box روی سرور نصب شود
      - REALITY و 443 باز شود
           ↓
   شروع TUN و اتصال از طریق VLESS/REALITY
           ↓
   ترافیک سیستم و DNS از طریق سرور رایگان عبور می‌کند
```

## 🏗️ سرور رایگان و کانفیگ‌ها

این پروژه منطبق بر مدل رایگان است:

- OCI Always Free
- اعمال قوانین امنیتی مناسب برای 443/TCP
- قالب‌سازی خودکار `client.json`
- تولید `server.json` از یک قالب استاندارد
- اجرای `systemctl enable --now sing-box`

## 🔐 امنیت و کیفیت

- اعتبارسنجی SHA-256 برای sing-box
- اتصال با VLESS + REALITY
- استفاده از TUN برای پوشش ترافیک کل سیستم
- جلوگیری از نصب نسخه‌های غیرمعتبر

## 📝 لایسنس

MIT

## 🤝 مشارکت

هر کمک یا PR پذیرفته می‌شود.
