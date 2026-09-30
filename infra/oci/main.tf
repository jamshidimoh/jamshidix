data "oci_identity_availability_domains" "ads" {
  compartment_id = var.tenancy_ocid
}

data "oci_core_images" "ubuntu" {
  compartment_id           = var.compartment_ocid
  operating_system         = "Ubuntu"
  operating_system_version = "24.04"
  shape                    = "VM.Standard.A1.Flex"
  state                    = "AVAILABLE"
  sort_by                  = "TIMECREATED"
  sort_order               = "DESC"
}

locals {
  availability_domain = var.availability_domain_name != "" ? var.availability_domain_name : data.oci_identity_availability_domains.ads.availability_domains[0].name
  image_id            = data.oci_core_images.ubuntu.images[0].id
}

resource "oci_core_vcn" "main" {
  cidr_blocks    = [var.vcn_cidr]
  compartment_id = var.compartment_ocid
  display_name   = "jamshidix-vcn"
  dns_label      = "jamshidix"
}

resource "oci_core_internet_gateway" "main" {
  compartment_id = var.compartment_ocid
  display_name   = "jamshidix-igw"
  vcn_id         = oci_core_vcn.main.id
}

resource "oci_core_route_table" "public" {
  compartment_id = var.compartment_ocid
  vcn_id         = oci_core_vcn.main.id
  display_name   = "jamshidix-public-rt"

  route_rules {
    network_entity_id = oci_core_internet_gateway.main.id
    destination       = "0.0.0.0/0"
    destination_type  = "CIDR_BLOCK"
  }
}

resource "oci_core_security_list" "gateway" {
  compartment_id = var.compartment_ocid
  vcn_id         = oci_core_vcn.main.id
  display_name   = "jamshidix-security"

  ingress_security_rules {
    protocol    = "6"
    source      = "0.0.0.0/0"
    description = "Jamshidix VLESS/REALITY"

    tcp_options {
      destination_port_range {
        min = 443
        max = 443
      }
    }
  }

  ingress_security_rules {
    protocol    = "6"
    source      = var.admin_cidr
    description = "SSH administration"

    tcp_options {
      destination_port_range {
        min = 22
        max = 22
      }
    }
  }

  egress_security_rules {
    protocol    = "all"
    destination = "0.0.0.0/0"
  }
}

resource "oci_core_subnet" "public" {
  cidr_block                 = var.subnet_cidr
  compartment_id             = var.compartment_ocid
  display_name               = "jamshidix-public-subnet"
  dns_label                  = "public"
  prohibit_public_ip_on_vnic = false
  route_table_id             = oci_core_route_table.public.id
  security_list_ids          = [oci_core_security_list.gateway.id]
  vcn_id                     = oci_core_vcn.main.id
}

resource "oci_core_instance" "gateway" {
  availability_domain = local.availability_domain
  compartment_id      = var.compartment_ocid
  display_name        = var.instance_name
  shape               = "VM.Standard.A1.Flex"

  shape_config {
    ocpus         = 2
    memory_in_gbs = 12
  }

  create_vnic_details {
    assign_public_ip = true
    hostname_label   = "jamshidix"
    subnet_id        = oci_core_subnet.public.id
  }

  metadata = {
    ssh_authorized_keys = var.ssh_public_key
  }

  source_details {
    boot_volume_size_in_gbs = var.boot_volume_size_gb
    source_id               = local.image_id
    source_type             = "image"
  }

  freeform_tags = {
    project = "jamshidix"
    cost    = "always-free-only"
  }
}

data "oci_core_vnic" "gateway" {
  vnic_id = oci_core_instance.gateway.primary_vnic_id
}
