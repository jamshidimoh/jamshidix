# Quick Start

هدف: یک نمونهٔ شخصی برای Windows.

## سرور (یک بار)

1. یک VM در OCI Always Free بسازید (`infra/oci`).
2. روی VM: `./server/install_ubuntu.sh` (کنار نصب sing-box، TCP/443 را در iptables هم باز و پایدار می‌کند).
3. `sudo ./server/generate_server_config.sh` را اجرا کنید؛ UUID و کلیدها ساخته می‌شود و یک **لینک `vless://`** چاپ می‌شود (در `/etc/sing-box/client-link.txt` هم ذخیره می‌شود).
4. `sudo systemctl enable --now sing-box`

## ویندوز

1. لینک را کپی کنید.
2. `Jamshidix.exe` را دوبارکلیک کنید (UAC را تأیید کنید).

بار بعد فقط دوبارکلیک. کلیک دوم = قطع اتصال.

مسیر دستی (اسکریپت‌های PowerShell در `client/windows`) برای کاربران پیشرفته باقی مانده است. `enable-killswitch.ps1` را فقط وقتی فعال کنید که تونل پایدار است؛ برای بازگشت `disable-killswitch.ps1`.

## عیب‌یابی

- سرور: `sudo systemctl status sing-box --no-pager` و `sudo sing-box check -c /etc/sing-box/server.json`
- کلاینت: `Jamshidix.exe status` و لاگ `%ProgramData%\Jamshidix\jamshidix.log`
- اگر «تونل اجرا شد ولی اینترنت پاسخ نمی‌دهد» دیدید: سرور خاموش/مسدود است، 443 در provider یا iptables بسته است، یا لینک اشتباه است.

موفقیت CI به معنی اثبات اتصال از ایران نیست؛ پس از ساخت gateway، اتصال واقعی و نشت DNS/IPv6 را در شبکهٔ خودتان تست کنید.
