terraform {
  required_providers {
    ansible-forms = {
      source = "hashicorp.com/se/ansible-forms"
    }
  }
  required_version = ">= 0.0.1"
}

provider "ansible-forms" {
  connection_profiles = [
    {
      name           = "cluster1"
      username       = var.username
      password       = var.password
      hostname       = "127.0.0.1:8443" # Publicly available by Ansible Forms
      validate_certs = var.validate_certs
    }
  ]
}

