# راه‌اندازی OCI Always Free

1. یک حساب OCI Free Tier ایجاد کنید و فقط منابع Always Free بسازید.
2. برای VM از A1 Flex استفاده کنید و مجموع منابع را داخل سقف رسمی Always Free نگه دارید.
3. Ubuntu ARM64 انتخاب شود.
4. Public IPv4 اختصاص دهید.
5. فقط SSH و TCP/443 را باز کنید؛ SSH را تا حد امکان به IP مدیریتی محدود کنید.
6. `server/install_ubuntu.sh` را اجرا کنید.
7. روی خود VM، REALITY keypair و `server.json` را تولید کنید.
8. `sing-box check` و listening روی 443 را بررسی کنید.

ظرفیت رایگان region و امکان ثبت‌نام/دسترسی حساب باید هنگام اجرا جداگانه بررسی شود.
