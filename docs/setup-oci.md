# راه‌اندازی OCI Always Free

Oracle برای Always Free منابعی را در home region برای طول عمر tenancy ارائه می‌کند. برای Ampere A1، سقف فعلی tenancy رایگان مجموع 2 OCPU و 12 GB RAM است. citeturn658806search2turn658806search0

Jamshidix از یک VM.Standard.A1.Flex با 2 OCPU و 12 GB RAM استفاده می‌کند.

مراحل:
1. tenancy و home region را آماده کنید.
2. Ubuntu 24.04 ARM64 انتخاب کنید.
3. یک public IPv4 داشته باشید.
4. TCP/443 را از اینترنت باز کنید.
5. SSH را فقط از CIDR مدیریتی خودتان مجاز کنید.
6. Terraform موجود در `infra/oci` را با مقادیر خود اجرا کنید، یا VM را دستی ایجاد کنید.
7. `server/install_ubuntu.sh` را اجرا کنید.
8. با `server/generate_server_config.sh` کلید REALITY و server.json را بسازید.
9. `sing-box check -c /etc/sing-box/server.json` را اجرا کنید.

Oracle هشدار می‌دهد که Always Free ممکن است با Out of Host Capacity مواجه شود. citeturn658806search2

نکتهٔ هزینه: این پروژه فقط منابع Always Free را مجاز می‌داند؛ Terraform نباید shape یا resource پولی ایجاد کند.
