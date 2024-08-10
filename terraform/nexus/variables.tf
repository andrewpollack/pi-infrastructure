variable "nexus_url" {
    description = "URL to nexus repository."
    type        = string
}

variable "nexus_username" {
    description = "Username to nexus repository."
    type        = string
}

variable "nexus_password" {
    description = "Password to nexus repository."
    type        = string
}

variable "nexus_docker_repo_port" {
    description = "Port for nexus' hosted docker repository."
    type        = number
}
 