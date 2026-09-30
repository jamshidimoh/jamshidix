# Windows client

نسخهٔ هدف sing-box در ویندوز دو حالت دارد:

1. TUN برای پوشش ترافیک سیستم، از جمله Chrome.
2. mixed inbound روی `127.0.0.1:2080` برای fallback پروکسی.

کانفیگ نمونه در `config/client.config.template.json` است.

فایل واقعی client.json و هر credential باید خارج از Git نگهداری شود.
