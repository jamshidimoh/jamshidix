# ماتریس زیرساخت

| گزینه | هزینهٔ تکرارشونده | VM دائمی | IP عمومی | نقش |
|---|---:|---:|---:|---|
| OCI Always Free A1 | 0 داخل سهمیه | بله | بله | dataplane مرجع |
| Cloudflare Workers Free | 0 داخل سهمیه | خیر | خیر به شکل VM | control-plane احتمالی |
| GitHub Actions | 0 | خیر | نامناسب | CI |
| VPN تجاری رایگان | 0 | وابسته به provider | نامشخص | backend پروژه نیست |

Workers Free در حال حاضر 100,000 request/day و 10ms CPU برای هر invocation دارد و بنابراین برای control-plane سبک مناسب‌تر از dataplane دائمی است. citeturn482847search0turn482847search1

اصل انتخاب: VM persistence + public IP + egress + اجرای TCP/443 باید همزمان فراهم باشد.
