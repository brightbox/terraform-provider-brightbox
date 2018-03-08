variable "web_image" {
  description = "search string for server image"
  default     = "^ubuntu-xenial.*server$"
}

variable "web_type" {
  description = "server type for web servers"
  default     = "1gb.ssd"
}

variable "database_type" {
  description = "search string for database type"
  default     = "^SSD 4GB$"
}
