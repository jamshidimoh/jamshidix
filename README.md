# Jamshidix

ابزار شخصی و رایگان برای ساخت یک تونل رمزنگاری‌شده در ویندوز با **VLESS + REALITY** (بر پایهٔ sing-box).

هدف: یک فایل `Jamshidix.exe` که با **یک دوبارکلیک** وصل شود.

```
Windows → Jamshidix.exe → sing-box (TUN) → VLESS/REALITY → سرور رایگان → اینترنت
```

## استفادهٔ سریع (کاربر ویندوز)

1. فایل `Jamshidix.exe` را از بخش **Releases** بگیرید و SHA-256 آن را با فایل `.sha256` کنار آن مقایسه کنید.
2. لینک `vless://...` سرور را کپی کنید (از خروجی اسکریپت سرور یا از کسی که سرور را اداره می‌کند).
3. روی `Jamshidix.exe` دوبارکلیک کنید و UAC را تأیید کنید.

برنامه خودش:
- sing-box (نسخهٔ رسمی، SHA-256 پین‌شده، **داخل خود EXE** — بدون دانلود) را نصب می‌کند؛
- لینک را از کلیپ‌بورد (یا `link.txt` کنار EXE) می‌خواند و کانفیگ را می‌سازد و با parser واقعی sing-box اعتبارسنجی می‌کند؛
- تونل را در پس‌زمینه بالا می‌آورد و **تست می‌کند که اینترنت واقعاً از تونل عبور می‌کند**؛
- کلیک دوم، برنامه را به حالت **قطع اتصال** می‌برد.

بار بعد نیازی به لینک نیست؛ فقط دوبارکلیک. برای تغییر سرور، لینک جدید را کپی کنید و دوباره کلیک کنید.

### دستورهای اضافه (CMD / PowerShell)

```
Jamshidix.exe status
Jamshidix.exe stop
Jamshidix.exe import <vless://...>      # یا --clipboard / --file link.txt
Jamshidix.exe autostart on|off          # اجرای خودکار در boot (Administrator)
Jamshidix.exe uninstall
```

لاگ تونل: `%ProgramData%\Jamshidix\jamshidix.log`

## راه‌اندازی سرور (یک بار)

نیاز: یک VM لینوکس (Ubuntu) با IPv4 عمومی و TCP/443. مرجع پروژه: OCI Always Free A1 (`infra/oci`، Terraform).

```bash
./server/install_ubuntu.sh               # نصب sing-box + باز کردن 443 در iptables
sudo ./server/generate_server_config.sh  # کلید/UUID می‌سازد و لینک vless:// چاپ می‌کند
sudo systemctl enable --now sing-box
```

لینک چاپ‌شده را کپی و مستقیم در ویندوز استفاده کنید. کلاینت به OCI وابسته نیست؛ با هر سرور VLESS+REALITY (tcp، `xtls-rprx-vision`) کار می‌کند.

## وضعیت «رایگان بودن»

| جزء | هزینه |
|---|---|
| Jamshidix، Go، GitHub Actions (ریپوی عمومی) | رایگان |
| sing-box (GPL-3.0) | رایگان |
| سرور OCI Always Free | رایگان داخل سهمیه |

محدودیت واقعی: ساخت حساب سرور خارجی (احراز هویت، ظرفیت، region) خارج از کنترل این پروژه است. جزئیات: `docs/free-services.md`.

## ساخت از سورس

```bash
go test ./...
# EXE تک‌فایلی (نیازمند آرشیو رسمی sing-box در cmd/jamshidix/assets؛ CI آن را خودکار می‌گیرد و hash را چک می‌کند)
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -tags embedsb -trimpath -ldflags="-s -w -H=windowsgui" -o dist/Jamshidix.exe ./cmd/jamshidix
```

انتشار: `VERSION` را بالا ببرید و تگ `vX.Y.Z` بزنید؛ workflow فایل EXE و SHA-256 را در Release می‌گذارد.

## محدودیت‌ها (صادقانه)

- EXE امضای دیجیتال ندارد؛ SmartScreen/آنتی‌ویروس ممکن است هشدار دهد. همیشه hash را بررسی کنید.
- Kill-switch هنوز اسکریپت PowerShell است (`client/windows/enable-killswitch.ps1`).
- CI فقط درستی کد/کانفیگ را اثبات می‌کند، نه اتصال واقعی از شبکهٔ شما. تست نهایی باید روی ویندوز و شبکهٔ هدف انجام شود.
- هیچ transport ای ضدفیلتر دائمی نیست؛ IP سرور ممکن است مسدود شود.

مستندات بیشتر: `docs/` (quickstart، architecture، threat-model).
