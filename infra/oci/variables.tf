variable "region" {
  description = "OCI home region for Always Free resources."
  type        = string
}

variable "tenancy_ocid" {
  type      = string
  sensitive = true
}

variable "user_ocid" {
  type      = string
  sensitive = true
}

variable "fingerprint" {
  type      = string
  sensitive = true
}

variable "private_key_path" {
  type      = string
  sensitive = true
}

variable "compartment_ocid" {
  description = "Compartment in which the VCN and VM will be created."
  type        = string
}

variable "availability_domain_name" {
  description = "Optional exact AD name. Empty uses the first AD."
  type        = string
  default     = ""
}

variable "ssh_public_key" {
  description = "OpenSSH public key installed on the VM."
  type        = string
}

variable "admin_cidr" {
  description = "CIDR allowed to SSH to the VM. Use your public IP/32 where possible."
  type        = string
}

variable "instance_name" {
  type    = string
  default = "jamshidix-gateway"
}

variable "vcn_cidr" {
  type    = string
  default = "10.77.0.0/16"
}

variable "subnet_cidr" {
  type    = string
  default = "10.77.1.0/24"
}

variable "boot_volume_size_gb" {
  type    = number
  default = 50

  validation {
    condition     = var.boot_volume_size_gb >= 50
    error_message = "OCI boot volume must be at least 50 GB."
  }
}
