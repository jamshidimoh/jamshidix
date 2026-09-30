# Cloudflare integration

## وضعیت

اتصال GitHub Actions به Cloudflare با Secretهای `CLOUDFLARE_EMAIL` و `CLOUDFLARE_API_KEY` با موفقیت آزمایش شده است. تست، endpoint کاربر و دسترسی Zone API را بررسی می‌کند.

## معماری فعلی Jamshidix

`Windows → sing-box TUN → VLESS/REALITY → OCI VM → Internet`

Cloudflare در این نسخه در نقش control plane و ابزار اختیاری Tunnel قرار دارد و مسیر اصلی VLESS/REALITY را تغییر نمی‌دهد.

## چرا Cloudflare Proxy معمولی جایگزین VPN نیست

Cloudflare DNS Proxy به‌طور معمول HTTP/HTTPS را روی مجموعه مشخصی از پورت‌ها پروکسی می‌کند. برای TCP/UDP دلخواه باید از Spectrum استفاده شود؛ جدول فعلی Spectrum نشان می‌دهد TCP/UDP عمومی برای Free در دسترس نیست و قابلیت‌های مربوط به آن به پلن‌های پولی/Enterprise وابسته است.

بنابراین قرار دادن رکورد DNS روی حالت orange-cloud به‌تنهایی VLESS/REALITY موجود را به یک VPN تحت Cloudflare تبدیل نمی‌کند.

## Cloudflare Tunnel

Cloudflare Tunnel روی همه پلن‌ها در دسترس است و اتصال origin به Cloudflare را به‌صورت outbound برقرار می‌کند. برای arbitrary TCP با Cloudflare Access، مستندات فعلی نیاز به `cloudflared` روی سرور و کلاینت و یک hostname تحت دامنه فعال در Cloudflare دارند.

برای Jamshidix این مسیر فقط به‌صورت optional در نظر گرفته شده است:

`Windows → cloudflared/Access → Cloudflare Tunnel → OCI → local TCP service`

مسیر فوق نباید بدون داشتن Zone/hostname و تست کامل جایگزین مسیر مستقیم VLESS/REALITY شود.

## فایل‌های آماده‌شده

- `server/install_cloudflared.sh` — نصب `cloudflared` روی Linux با SHA-256 pin برای نسخه `2026.9.3` و پشتیبانی از ARM64/AMD64.
- `client/windows/install-cloudflared.ps1` — نصب نسخه Windows AMD64 با SHA-256 verification.
- `.github/workflows/cloudflare-connection-test.yml` — تست دستی احراز هویت و دسترسی Cloudflare.

هیچ Tunnel Token، private key یا Global API Key در repository ذخیره نمی‌شود.

## پیش‌شرط فعال‌سازی Tunnel

برای استفاده عملی از arbitrary TCP، یک دامنه باید در Cloudflare فعال باشد و hostname موردنظر برای Tunnel ساخته شود. سپس Tunnel Token باید فقط در Secret/credential store مناسب قرار گیرد و روی سرور به‌صورت service اجرا شود. Cloudflare استفاده از API Token محدود را به‌جای Global API Key توصیه می‌کند.

## پورت اتصال Tunnel

`cloudflared` برای اتصال به Cloudflare به خروجی شبکه نیاز دارد. مستندات فعلی پورت TCP `7844` را برای fallback/HTTP2 و مسیرهای Tunnel ذکر می‌کنند؛ اگر این مسیر مسدود باشد Tunnel برقرار نمی‌شود.