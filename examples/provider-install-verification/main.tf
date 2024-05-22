terraform {
  required_providers {
    ansibleforms = {
      source = "hashicorp.com/se/ansible-forms"
    }
  }
}

provider "ansible-forms" {}
