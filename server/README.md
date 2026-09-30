# Server

سرور هدف یک VM رایگان خارج از ایران است. مرجع نسخهٔ اول، OCI Always Free A1 است.

ترتیب کلی:
```bash
chmod +x install_ubuntu.sh generate_server_config.sh
./install_ubuntu.sh
sudo ./generate_server_config.sh
sudo systemctl enable --now sing-box
```

در provider نیز علاوه بر فایروال Linux باید TCP/443 باز باشد.
