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
