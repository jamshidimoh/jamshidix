# Jamshidix.exe

## اجرای یک‌کلیکی

بعد از آماده بودن `client.json`، فقط روی `Jamshidix.exe` دوبارکلیک کنید.

EXE به‌صورت خودکار:
- در صورت نیاز درخواست Administrator می‌دهد.
- sing-box 1.14.1 را نصب/بررسی می‌کند.
- `client.json` کنار EXE یا در `%ProgramData%\\Jamshidix\\client.json` را پیدا می‌کند.
- کانفیگ را با parser واقعی sing-box اعتبارسنجی می‌کند.
- TUN را در پس‌زمینه اجرا می‌کند.

برای اجرای اول، می‌توانید `client.json` واقعی را کنار EXE قرار دهید. اگر فایل وجود نداشته باشد، EXE یک `client.config.template.json` در پوشهٔ داده می‌سازد و پوشه را باز می‌کند.

## فرمان‌های عیب‌یابی

```powershell
.\Jamshidix.exe status
.\Jamshidix.exe stop
.\Jamshidix.exe version
```

فرمان‌های قدیمی `install`، `config`، `check` و `run` نیز برای مدیریت دستی باقی مانده‌اند.

## Kill-switch و autostart

این دو قابلیت هنوز اسکریپت‌های PowerShell موجود در `client/windows/` هستند و در مسیر یک‌کلیکی پایه ادغام نشده‌اند.

## معماری

`Jamshidix.exe → sing-box.exe → TUN → VLESS/REALITY → gateway`
