# OCI provisioning

این Terraform یک VM از نوع VM.Standard.A1.Flex با 2 OCPU و 12 GB RAM ایجاد می‌کند تا با سقف فعلی Always Free tenancy منطبق باشد.

اجرا:

terraform init
terraform fmt -check
terraform validate
terraform plan
terraform apply

قبل از اجرا، terraform.tfvars را از روی terraform.tfvars.example بسازید.

هرگز terraform.tfvars یا فایل‌های state را commit نکنید.

پس از ساخت VM، public IPv4 خروجی بگیرید و با SSH وارد شوید. سپس اسکریپت‌های server را اجرا کنید.
