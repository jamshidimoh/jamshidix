# Jamshidix — Iran Free Tunnel

سرویس شخصی و متن‌باز برای ایجاد تونل امن از ایران، با معیار «صفر هزینهٔ سرویس».

## معیار هزینه

هیچ اشتراک، VPS پولی، API پولی، دامنهٔ پولی یا VPN تجاری نباید برای اجرای dataplane لازم باشد. زیرساخت فقط از سهمیهٔ رسمی رایگان provider استفاده می‌کند.

رایگان بودن provider به معنی تضمین ظرفیت، uptime یا امکان ثبت‌نام برای همهٔ کاربران نیست.

## معماری

`Windows → sing-box TUN → VLESS + REALITY → Free VM → Internet`

Chrome در حالت TUN به proxy جداگانه نیاز ندارد. mixed proxy محلی روی 127.0.0.1:2080 نیز برای fallback وجود دارد.

نسخهٔ stable مورد استفاده در این ریپو `sing-box 1.14.1` است. نسخهٔ 1.15.x فعلاً alpha است. citeturn482847search6

REALITY در sing-box از private key در server و public key در client استفاده می‌کند. citeturn248666search1

## زیرساخت رایگان

provider مرجع: Oracle Cloud Always Free A1.

طبق مستندات فعلی Oracle، سهمیهٔ Always Free برای A1 در tenancy رایگان معادل مجموع 2 OCPU و 12 GB RAM است. همچنین منابع Always Free در home region بدون هزینه ادامه پیدا می‌کنند، مشروط به ماندن در حدود رایگان. citeturn658806search2turn658806search0

این پروژه provider را hard-code نمی‌کند؛ adapterهای دیگر فقط در صورتی اضافه می‌شوند که VM دائمی، IP عمومی و egress رایگان واقعی داشته باشند.

## کنترل نشت

TUN با `auto_route=true` و `strict_route=true` اجرا می‌شود. مستندات sing-box می‌گویند strict_route در Windows از DNS leak ناشی از رفتار عادی multihomed DNS جلوگیری می‌کند. citeturn860646search2

DNS upstream از DoT به 1.1.1.1 استفاده می‌کند و connection آن از outbound پروکسی عبور می‌کند؛ detour برای DNS در sing-box جزو Dial Fields است. citeturn190631search1turn190631search0

## وضعیت توسعه

- server bootstrap
- تولید REALITY keypair
- client template
- Windows installer
- Windows kill-switch
- OCI Terraform
- schema validation و secret scanning در CI

## محدودیت واقعی

این پروژه خودِ اینترنت خارجی رایگان ایجاد نمی‌کند. برای استفادهٔ واقعی، حداقل یک نقطهٔ خروج خارجیِ رایگان و قابل‌دسترسی لازم است. ظرفیت free-tier، region و verification در اختیار provider است.

هرگز private key، UUID خصوصی، token یا کانفیگ واقعی را commit نکنید.
