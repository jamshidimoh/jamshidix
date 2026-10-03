# Windows client

مسیر پیشنهادی: `Jamshidix.exe` (یک‌کلیکی، فایل `cmd/jamshidix/README.md`).

اسکریپت‌های این پوشه برای مدیریت دستی/پیشرفته‌اند: install، render-client، run، autostart، kill-switch.

دو حالت sing-box:

1. TUN برای پوشش ترافیک سیستم، از جمله Chrome.
2. mixed inbound روی `127.0.0.1:2080` برای fallback پروکسی.

کانفیگ نمونه: `config/client.config.template.json`. فایل واقعی client.json و هر credential/لینک باید خارج از Git بماند.
