terraform {
  required_providers {
    lightning = {
      source  = "lightningai/lightning"
      version = "~> 1.0"
    }
  }
}

provider "lightning" {
  api_key    = var.lightning_api_key
  user_id    = var.lightning_user_id
  project_id = var.lightning_project_id
}

variable "lightning_api_key" {
  type      = string
  sensitive = true
}

variable "lightning_user_id" {
  type = string
}

variable "lightning_project_id" {
  type = string
}

resource "lightning_studio" "example" {
  name    = "my-studio"
  machine = "cpu-4"
  running = true

  startup_script = <<-EOT
    #!/bin/bash
    set -e
    echo "Hello from startup script!"
    pip install -q torch torchvision
  EOT

  startup_script_mode = "once"
  startup_timeout     = "10m"
}

output "studio_id" {
  value = lightning_studio.example.id
}

output "studio_status" {
  value = lightning_studio.example.status
}

output "studio_public_ip" {
  value = lightning_studio.example.public_ip
}
