output "gateway_public_ip" {
  description = "Public IPv4 address for the Jamshidix gateway."
  value       = data.oci_core_vnic.gateway.public_ip_address
}

output "gateway_private_ip" {
  value = data.oci_core_vnic.gateway.private_ip_address
}

output "selected_image" {
  value = local.image_id
}

output "ssh_command" {
  value = "ssh ubuntu@\${data.oci_core_vnic.gateway.public_ip_address}"
}
