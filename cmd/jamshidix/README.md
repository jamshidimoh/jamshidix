# Jamshidix.exe

## اجرای یک‌کلیکی

1. لینک `vless://...` را کپی کنید.
2. روی `Jamshidix.exe` دوبارکلیک کنید.

ترتیب منبع کانفیگ:
1. لینک جدید در کلیپ‌بورد (اگر با لینک ذخیره‌شده فرق داشته باشد)
2. `%ProgramData%\Jamshidix\client.json` موجود
3. `link.txt` / `vless.txt` / `client.json` کنار EXE

کلیک دوم وقتی تونل فعال است، پیشنهاد قطع اتصال می‌دهد.

## نسخه‌ها

- `Jamshidix.exe` (release): sing-box داخل فایل است، بدون دانلود.
- `Jamshidix-lite.exe` (CI): بدون sing-box؛ در اولین اجرا از GitHub می‌گیرد و SHA-256 را چک می‌کند.

## دستورها

```
Jamshidix.exe status | stop | version | check | run | install
Jamshidix.exe import <vless://...> | --clipboard | --file <path>
Jamshidix.exe config --server-ip IP --uuid UUID --public-key KEY --short-id ID --handshake-host HOST [--port 443]
Jamshidix.exe autostart on|off
Jamshidix.exe uninstall
```

## توسعه

- منطق parse/render لینک پلتفرم‌مستقل است و در `link_test.go` تست می‌شود.
- `SINGBOX_BIN=/path/sing-box go test ./cmd/jamshidix` کانفیگ تولیدشده را با parser واقعی چک می‌کند.
- دو نسخه از template کلاینت وجود دارد (`config/` و `assets/`)؛ تست `TestTemplateCopiesInSync` یکسان بودنشان را تضمین می‌کند.

`Jamshidix.exe → sing-box.exe → TUN → VLESS/REALITY → gateway`
