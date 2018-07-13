output "application address" {
  value = "${brightbox_cloudip.lb.fqdn}"
}

output "server1 address" {
  value = "${brightbox_cloudip.server1.fqdn}"
}

output "server2 address" {
  value = "${brightbox_cloudip.server2.fqdn}"
}

output "database address" {
  value = "${brightbox_cloudip.database.fqdn}"
}

output "database_username" {
  value = "${brightbox_database_server.database.admin_username}"
}

output "database_password" {
  value = "${brightbox_database_server.database.admin_password}"
}

output "orbit_backup_userid" {
  value = "${brightbox_container.backups.auth_user}"
}

output "orbit_backup_password" {
  value = "${brightbox_container.backups.auth_key}"
}

output "orbit_backup_url" {
  value = "${brightbox_container.backups.orbit_url}${brightbox_container.backups.account_id}/${brightbox_container.backups.name}"
}
