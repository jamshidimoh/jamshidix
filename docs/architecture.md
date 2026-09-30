# معماری فنی

## مسیر اصلی

`Windows → local sing-box TUN → VLESS/REALITY → free VM → Internet`

## لایه‌ها

### Edge

Linux VM با IP عمومی؛ در سناریوی OCI ترجیحاً ARM64.

### Tunnel Gateway

sing-box با inbound از نوع VLESS و TLS REALITY.

### Egress

سرور، ترافیک client مجاز را با مسیر IPv4 به اینترنت عمومی می‌رساند.

### Client

Windows از TUN استفاده می‌کند. Chrome در حالت TUN بدون proxy اختصاصی کار می‌کند؛ mixed proxy برای fallback وجود دارد.

## الزامات امنیتی

- private key فقط روی سرور.
- credential واقعی خارج از Git.
- پورت‌های غیرضروری بسته.
- IPv6 تا زمان تکمیل routing/NAT قابل اتکا نیست.
- DNS policy و kill-switch باید قبل از اعلام نسخهٔ stable تست شوند.

## اصل طراحی

provider و transport باید pluggable باشند. منطق client نباید به OCI یا یک hostname خاص hard-code شود.
