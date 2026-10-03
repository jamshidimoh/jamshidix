# Jamshidix.exe — اجرای یک‌کلیکی

## نسخهٔ جدید (v1.0)

### عملکرد کلی

- در اولین اجرا، برنامه خودکار sing-box را نصب می‌کند.
- اگر سرور رایگان OCI موجود نباشد، پیشنهاد ایجاد سرور می‌دهد.
- اگر `client.json` وجود ندارد، از قالب ساخته می‌شود.
- TUN فعال می‌شود و ترافیک سیستم از طریق سرور رایگان عبور می‌کند.

### فرمان‌های دستی

```powershell
.\Jamshidix.exe install
.\Jamshidix.exe config --server-ip <IP> --uuid <UUID> --public-key <KEY> --short-id <ID> --handshake-host <HOST>
.\Jamshidix.exe check
.\Jamshidix.exe run
.\Jamshidix.exe status
.\Jamshidix.exe stop
.\Jamshidix.exe version
```

### ساخت محصول نهایی

```bash
go build -o Jamshidix.exe -ldflags="-H windowsgui" ./cmd/jamshidix
```

## هدف نهایی

- فقط یک کلیک برای کاربر نهایی
- استفاده از سرورهای رایگان و OCI Always Free
- کانفیگ‌های نمونه و تولید خودکار
- پشتیبانی از VLESS + REALITY و TUN
- سازگاری با اجراهای واگذارشده به کاربران عادی

