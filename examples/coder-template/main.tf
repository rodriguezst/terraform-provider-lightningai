terraform {
  required_providers {
    coder = {
      source  = "coder/coder"
      version = "~> 0.12"
    }
    lightning = {
      source  = "lightningai/lightning"
      version = "~> 1.0"
    }
  }
}

data "coder_workspace" "me" {}

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

resource "lightning_studio" "dev" {
  name    = "coder-${data.coder_workspace.me.owner}-${data.coder_workspace.me.name}"
  machine = "cpu-4"
  running = data.coder_workspace.me.start_count > 0

  startup_script = file("${path.module}/bootstrap.sh")

  startup_script_mode = "once"
  startup_timeout     = "20m"
}

output "studio_id" {
  value = lightning_studio.dev.id
}
