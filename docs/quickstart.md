# Quick Start

هدف این راهنما راه‌اندازی یک نمونهٔ شخصی برای Windows است.

1. یک VM در OCI Always Free ایجاد کنید، ترجیحاً با Terraform پروژه.
2. public IPv4 و TCP/443 داشته باشید.
3. روی VM، `server/install_ubuntu.sh` را اجرا کنید.
4. با `server/generate_server_config.sh` UUID و REALITY keys را تولید کنید.
5. public key، UUID، short ID، public IP و handshake host را برای ساخت client بردارید.
6. در Windows، `client/windows/install.ps1` را به‌صورت Administrator اجرا کنید.
7. `render-client.ps1` را اجرا کنید تا client.json ساخته و validate شود.
8. قبل از استفادهٔ عادی، `enable-killswitch.ps1` را اجرا کنید.
9. `run.ps1` را به‌صورت Administrator اجرا کنید.
10. برای اجرای خودکار هنگام boot، `install-autostart.ps1` را اجرا کنید.

عیب‌یابی:

- سرور: `sudo systemctl status sing-box --no-pager`
- کانفیگ سرور: `sudo sing-box check -c /etc/sing-box/server.json`
- کانفیگ client: `%ProgramFiles%\Jamshidix\sing-box.exe check -c %ProgramData%\Jamshidix\client.json`
- بازگردانی kill-switch: `disable-killswitch.ps1`

تست موفق CI به معنی اثبات end-to-end از ایران نیست؛ پس از ساخت gateway باید connectivity واقعی و نشت DNS/IPv6 از شبکهٔ موردنظر آزمایش شود.