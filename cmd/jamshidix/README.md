# Jamshidix.exe

## نصب

PowerShell یا CMD را به‌صورت Administrator باز کنید:

```powershell
.\Jamshidix.exe install
```

## ساخت کانفیگ

```powershell
.\Jamshidix.exe config --server-ip "<PUBLIC_IP>" --uuid "<UUID>" --public-key "<REALITY_PUBLIC_KEY>" --short-id "<SHORT_ID>" --handshake-host "<HANDSHAKE_HOST>"
```

## بررسی و اجرا

```powershell
.\Jamshidix.exe check
.\Jamshidix.exe status
.\Jamshidix.exe run
```

`run` تونل TUN را در همان console اجرا می‌کند و معمولاً به Administrator نیاز دارد.

## Kill-switch و autostart

برای این دو قابلیت، اسکریپت‌های `client/windows/enable-killswitch.ps1`، `disable-killswitch.ps1` و `install-autostart.ps1` همچنان مسیر رسمی پروژه هستند.

## معماری

`Jamshidix.exe → sing-box.exe → TUN → VLESS/REALITY → gateway`