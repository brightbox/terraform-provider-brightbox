output "application_address" {
  value = brightbox_cloudip.lb.fqdn
}

output "server1_address" {
  value = brightbox_cloudip.server1.fqdn
}

output "server2_address" {
  value = brightbox_cloudip.server2.fqdn
}

output "database_address" {
  value = brightbox_cloudip.database.fqdn
}

output "database_username" {
  value = brightbox_database_server.database.admin_username
}

output "database_password" {
  value = brightbox_database_server.database.admin_password
}

output "orbit_backup_name" {
  value = brightbox_orbit_container.backups.name
}
