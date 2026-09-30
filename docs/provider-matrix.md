# ماتریس انتخاب زیرساخت

| گزینه | هزینهٔ تکرارشونده | dataplane دائمی | IP عمومی | جایگاه |
|---|---:|---:|---:|---|
| OCI Always Free A1 | 0 در محدودهٔ سهمیه | بله | بله | provider مرجع |
| Cloudflare Workers Free | 0 | مناسب application/control-plane | خیر به شکل VPS | فقط control-plane احتمالی |
| GitHub Actions | 0 | خیر؛ runner موقت | نامناسب | رد |
| Web hosting رایگان | معمولاً 0 | اغلب sleep/محدود | معمولاً خیر | رد |
| VPN تجاری رایگان | 0 | وابسته به provider | ممکن است | backend پروژه نیست |

اصل انتخاب: CPU/RAM + IP عمومی + persistence + egress + اجرای TCP/443 باید همزمان فراهم باشد.

در production هیچ مسیر پرداختی نباید لازم باشد. provider باید از هستهٔ tunnel مستقل بماند.
