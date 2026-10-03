# Jamshidix

کلاینت رایگان و تک‌فایلی ویندوز برای اتصال به سرورهای عمومی رایگان و اجرای تونل TUN بر پایهٔ sing-box.

## هدف محصول

کاربر فقط `Jamshidix.exe` را اجرا می‌کند. برنامه:

1. فهرست سرورهای عمومی رایگان را از چند mirror دریافت می‌کند.
2. در صورت نبود دسترسی به اینترنت، آخرین فهرست cache‌شده را استفاده می‌کند.
3. سرورها را بر اساس سلامت، سرعت، اولویت و region نمایش می‌دهد.
4. دسترسی هر سرور را از شبکهٔ محلی بررسی می‌کند.
5. با انتخاب کاربر، کانفیگ sing-box را خودکار می‌سازد، اعتبارسنجی می‌کند و TUN را بالا می‌آورد.
6. خروجی عمومی اینترنت را بعد از اتصال بررسی می‌کند.
7. با «قطع اتصال» تونل را متوقف می‌کند.

رابط برنامه بدون نیاز به نصب جداگانهٔ sing-box، PowerShell script، `client.json` یا واردکردن دستی لینک VLESS طراحی شده است.

## رابط کاربری

در صفحهٔ اصلی:

- «بروزرسانی سرورها» برای دریافت جدیدترین directory.
- مرتب‌سازی بر اساس «اولویت»، «سرعت» و «منطقه».
- فیلتر region.
- نمایش latency محلی یا وضعیت دسترسی محلی.
- نمایش priority.
- دکمهٔ «اتصال» و «قطع اتصال».
- با دوبارکلیک روی یک server نیز اتصال شروع می‌شود.

برنامه هر ۲۰ دقیقه یک refresh خودکار را نیز انجام می‌دهد.

## معماری

`Windows → Jamshidix.exe → sing-box TUN → selected public node → Internet`

فهرست سرورها از این زنجیره عبور می‌کند:

`public sources → collector → validation → deduplication → remote health check → directory/nodes.json → mirrors → Jamshidix`

کلاینت چند mirror مستقل دارد و در صورت شکست همهٔ mirrorها، از cache محلی استفاده می‌کند.

## منابع عمومی

Collector فقط nodeهایی را وارد directory می‌کند که با policy فعلی کلاینت سازگار باشند: VLESS + REALITY + TCP + `xtls-rprx-vision`.

منابع فعلی شامل پروژه‌های عمومی به‌روزشونده مانند ebrasha، Baarcuda، 0xRadikal، morpheusadam و feedهای Vlessnode/freevlessnode هستند. این منابع «زیرساخت مورد اعتماد» محسوب نمی‌شوند؛ فقط منبع discovery هستند.

## رایگان بودن

اجزای پروژه بر مبنای ابزارهای رایگان و free-tier طراحی شده‌اند:

- Jamshidix، Go و sing-box: رایگان.
- GitHub Actions برای repository عمومی: رایگان.
- collector و directory publishing: GitHub Actions.
- OCI Always Free به‌عنوان fallback اختیاری سمت سرور.

هیچ subscription پولی یا VPN تجاری برای عملکرد اصلی پروژه لازم نیست.

## نکتهٔ مهم دربارهٔ شبکهٔ فیلترشده

برنامه برای این سناریو طراحی شده که مسیر مستقیم به برخی منابع عمومی روی شبکهٔ کلاینت در دسترس نباشد. به همین دلیل discovery روی چند mirror مستقل انجام می‌شود و directory در cache محلی نگه داشته می‌شود.

با این حال هیچ نرم‌افزاری نمی‌تواند بدون هیچ مسیر bootstrap، داده‌ای را از اینترنتی که تمام مسیرهای ممکن آن مسدود شده‌اند دریافت کند. بنابراین cache محلی و داشتن چند mirror بخش ضروری معماری هستند.

## امنیت

Private key هیچ server عمومی یا سرور شخصی نباید داخل repository یا EXE قرار گیرد.

فهرست nodeهای عمومی untrusted است. سلامت فنی یک node به‌معنی قابل‌اعتماد بودن اپراتور آن نیست. برای داده‌های حساس از node عمومی رایگان استفاده نکنید.

`sing-box` در Release داخل EXE قرار می‌گیرد و archive رسمی آن با SHA-256 pinned بررسی می‌شود.

## ساخت

`go test ./...`

برای ساخت EXE خودکفا:

`GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -tags embedsb -trimpath -ldflags="-s -w -H=windowsgui" -o dist/Jamshidix.exe ./cmd/jamshidix`

Workflow انتشار، sing-box رسمی pinned را داخل EXE embed می‌کند و فایل SHA-256 تولید می‌کند.

## سمت سرور اختصاصی اختیاری

در `infra/oci` یک معماری OCI Always Free برای gateway شخصی وجود دارد:

`OCI VM → sing-box VLESS/REALITY → Internet`

این مسیر جایگزین directory عمومی نیست؛ fallback اختصاصی است.

## محدودیت واقعی

- availability سرورهای رایگان عمومی دائمی نیست.
- IP یا دامنهٔ یک node ممکن است بعداً از کار بیفتد یا مسدود شود.
- TCP reachability به‌تنهایی تضمین نمی‌کند handshake نهایی موفق شود؛ اتصال واقعی روی کلاینت دوباره بررسی می‌شود.
- اجرای TUN در Windows به دسترسی Administrator نیاز دارد.
- EXE فعلاً امضای Authenticode ندارد.
