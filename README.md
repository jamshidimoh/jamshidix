# Jamshidix — Iran Free Tunnel

سرویس شخصیِ رایگان و متن‌باز برای اتصال امن از ایران، با معیار اصلی «صفر هزینهٔ سرویس».

## سیاست هزینه

این پروژه نباید برای اجرای dataplane به اشتراک پولی، VPS پولی، API پولی، دامنهٔ پولی یا سرویس VPN تجاری وابسته باشد. فقط منابعی مجازند که در سهمیهٔ رایگانِ رسمی provider قرار دارند.

نکتهٔ مهم: رایگان بودن نرم‌افزار و سهمیهٔ رایگان provider با تضمین دائمیِ ظرفیت، uptime یا امکان ثبت‌نام یک کاربر در همهٔ کشورها یکسان نیست. بنابراین provider abstraction از ابتدا بخشی از معماری است.

## معماری نسخهٔ 0.1

\`Windows → sing-box TUN → VLESS + REALITY → Free VM → Internet\`

برای Chrome، در حالت TUN نیازی به تنظیم proxy جداگانه نیست. یک SOCKS5/HTTP محلی نیز برای fallback در نظر گرفته شده است.

هستهٔ تونل \`sing-box\` است. مسیر اصلی نسخهٔ 0.1 از VLESS + REALITY استفاده می‌کند؛ WireGuard به عنوان transport آزمایشی/پشتیبان بررسی خواهد شد.

## زیرساخت رایگان

Provider مرجع فعلی: Oracle Cloud Always Free A1.

پروژه به یک provider خاص hard-code نمی‌شود. هدف نهایی داشتن حداقل دو adapter رایگان است، تا در صورت از دسترس خارج شدن ظرفیت یک provider، لایهٔ client و tunnel تغییر نکند.

## وضعیت فعلی

- اسکلت repository آماده است.
- نصب و تولید کانفیگ sing-box در حال توسعه است.
- CI برای JSON و shell syntax فعال است.
- هیچ secret یا private key نباید وارد Git شود.
- IPv6 در مسیر اصلی تا زمان تکمیل routing/NAT صریحاً به عنوان مسیر قابل اعتماد تلقی نمی‌شود.

## مسیر توسعه

1. Provisioning رایگان VM.
2. نصب و pin کردن نسخهٔ sing-box.
3. تولید UUID و REALITY keypair روی سرور.
4. ساخت کانفیگ client بدون انتشار secret.
5. Windows TUN + DNS control + kill-switch.
6. health check و self-test.
7. provider adapters و failover.
8. تست عملی از شبکهٔ ایران.

این پروژه «اینترنت رایگان خارجی» تولید نمی‌کند؛ یک gateway رایگان/شخصی را مدیریت می‌کند و موفقیت نهایی به وجود یک نقطهٔ خروج رایگانِ قابل‌دسترس بستگی دارد.

هرگز \`server.json\` واقعی، private key، UUID خصوصی یا token را commit نکنید.
