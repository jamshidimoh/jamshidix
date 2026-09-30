# Jamshidix — Iran Free Tunnel

سرویس شخصی و متن‌باز برای ایجاد تونل امن از ایران، با معیار «صفر هزینهٔ سرویس».

## معیار هزینه

هیچ اشتراک، VPS پولی، API پولی، دامنهٔ پولی یا VPN تجاری نباید برای اجرای dataplane لازم باشد. زیرساخت فقط از سهمیهٔ رسمی رایگان provider استفاده می‌کند.

این شرط به معنی تضمین همیشگیِ ظرفیت یا امکان ثبت‌نام برای همهٔ کاربران نیست.

## معماری

`Windows → sing-box TUN → VLESS + REALITY → Free VM → Internet`

Chrome در حالت TUN به proxy جداگانه نیاز ندارد. mixed proxy محلی روی `127.0.0.1:2080` برای برنامه‌های سازگار وجود دارد.

نسخهٔ stable فعلی در این ریپو `sing-box 1.14.1` است.

## زیرساخت رایگان

provider مرجع: Oracle Cloud Always Free A1.

سقف فعلی A1 برای tenancy رایگان مجموعاً 2 OCPU و 12 GB RAM است. Always Free به home region محدود است و ظرفیت، احراز هویت و availability می‌تواند در زمان اجرا تغییر کند.

Terraform پروژه shape را روی `VM.Standard.A1.Flex` با 2 OCPU و 12 GB RAM قفل می‌کند.

## کنترل نشت

TUN با `auto_route=true` و `strict_route=true` اجرا می‌شود. DNS نیز از DoT استفاده می‌کند و مسیر DNS از outbound پروکسی عبور می‌کند.

IPv6 forwarding روی gateway عمداً خاموش است تا قبل از پیاده‌سازی routing/NAT صریح، مسیر IPv6 ناخواسته ایجاد نشود.

## امنیت

private key سرور فقط روی gateway نگهداری می‌شود. کانفیگ واقعی client/server، UUID خصوصی، token و private key نباید commit شوند.

برای نصب binary رسمی sing-box، SHA-256 release قبل از نصب بررسی می‌شود.

Windows kill-switch می‌تواند outbound پیش‌فرض سیستم را مسدود کند؛ ابتدا تونل را آماده کنید و برای بازگردانی از `disable-killswitch.ps1` استفاده کنید.

## وضعیت نسخه

آخرین snapshot تمام gateهای CI را پاس کرده است: JSON، ShellCheck، PowerShell، Secret Audit، sing-box parser، Terraform format و Terraform validate.

تست end-to-end از یک شبکهٔ واقعی ایران و یک VM واقعی هنوز بخشی از provisioning کاربر است؛ CI نمی‌تواند latency، DPI یا دسترسی free-tier را از ایران اثبات کند.

## فایل اجرایی Windows

فایل `Jamshidix.exe` به‌صورت GUI و one-click ساخته می‌شود: با دوبارکلیک، در صورت نیاز UAC را می‌گیرد، sing-box رسمی را بررسی/نصب می‌کند، `client.json` را پیدا و اعتبارسنجی می‌کند و TUN را در پس‌زمینه اجرا می‌کند. برای استفاده، `client.json` واقعی را یک‌بار کنار EXE قرار دهید یا در `%ProgramData%\\Jamshidix\\client.json` بسازید. برای kill-switch و autostart، اسکریپت‌های PowerShell موجود باقی مانده‌اند.

Build مستقل با Go انجام می‌شود و SHA-256 آرشیو sing-box قبل از نصب بررسی می‌شود.

## اسناد

- `docs/quickstart.md`
- `cmd/jamshidix/README.md` — اجرای one-click ویندوز
- `docs/setup-oci.md`
- `docs/architecture.md`
- `docs/provider-matrix.md`
- `docs/threat-model.md`
- `infra/oci/`
- `client/windows/`

## منابع رسمی

- Oracle Cloud Free Tier: https://www.oracle.com/cloud/free/
- Oracle Always Free resources: https://docs.oracle.com/en-us/iaas/Content/FreeTier/freetier_topic-Always_Free_Resources.htm
- OCI Terraform provider: https://registry.terraform.io/providers/oracle/oci/latest
- sing-box releases: https://github.com/SagerNet/sing-box/releases
- sing-box documentation: https://sing-box.sagernet.org/

## محدودیت واقعی

این پروژه اینترنت خارجی «رایگان» تولید نمی‌کند. برای استفادهٔ عملی، باید یک نقطهٔ خروج خارجیِ رایگان و قابل‌دسترسی داشته باشید. هزینهٔ project در design صفر است، ولی provider می‌تواند ظرفیت، verification یا سیاست دسترسی خود را تغییر دهد.

هرگز private key، credential یا کانفیگ واقعی را در Git قرار ندهید.