# ماتریس زیرساخت

| گزینه | هزینهٔ تکرارشونده | VM دائمی | IP عمومی | نقش |
|---|---:|---:|---:|---|
| OCI Always Free A1 | 0 داخل سهمیه | بله | بله | dataplane مرجع |
| Cloudflare Workers Free | 0 داخل سهمیه | خیر | خیر به شکل VM | control-plane احتمالی |
| GitHub Actions | 0 | خیر | نامناسب | CI |
| VPN تجاری رایگان | 0 | وابسته به provider | نامشخص | backend پروژه نیست |

Cloudflare Workers برای control-plane سبک قابل استفاده است، اما جای یک gateway VM دائمی را در این معماری نمی‌گیرد.

اصل انتخاب: VM persistence + public IP + egress + اجرای TCP/443 باید همزمان فراهم باشد.

این پروژه به provider تجاری یا اشتراک VPN وابسته نیست.