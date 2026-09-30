# راه‌اندازی OCI Always Free

## منابع

Jamshidix برای نسخهٔ مرجع از `VM.Standard.A1.Flex` با 2 OCPU و 12 GB RAM استفاده می‌کند تا داخل سقف فعلی Always Free tenancy باقی بماند.

## ساخت VM

در `infra/oci/`:

```bash
cp terraform.tfvars.example terraform.tfvars
terraform init
terraform fmt -check
terraform validate
terraform plan
terraform apply
```

`terraform.tfvars` و state را commit نکنید.

`region` باید home region tenancy شما باشد. ظرفیت Always Free ممکن است در یک Availability Domain موقتاً موجود نباشد.

## شبکه

Terraform این موارد را ایجاد می‌کند:
- public subnet
- Internet Gateway
- route table
- TCP/443 برای Jamshidix
- SSH فقط برای `admin_cidr`

برای SSH تا حد امکان `admin_cidr` را به IP عمومی فعلی خودتان با /32 محدود کنید.

## راه‌اندازی sing-box

پس از دریافت public IP:

```bash
ssh ubuntu@<PUBLIC_IP>
chmod +x server/install_ubuntu.sh server/generate_server_config.sh
./server/install_ubuntu.sh
sudo ./server/generate_server_config.sh
sudo systemctl enable --now sing-box
sudo systemctl status sing-box --no-pager
```

کلید REALITY روی خود سرور تولید می‌شود. private key را به client منتقل نکنید.

## نکتهٔ هزینه

فقط منابع Always Free را استفاده کنید و حساب را به منابع پولی upgrade نکنید. verification، ظرفیت region و شرایط provider خارج از کنترل این پروژه است.

منابع رسمی:
https://www.oracle.com/cloud/free/
https://docs.oracle.com/en-us/iaas/Content/FreeTier/freetier_topic-Always_Free_Resources.htm