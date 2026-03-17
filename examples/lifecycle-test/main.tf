# Lifecycle test for terraform-provider-lightningai
#
# This simulates Coder workspace behaviour WITHOUT the Coder provider.
# In Coder templates, the key mechanism is:
#
#   running = data.coder_workspace.me.start_count > 0
#
# When Coder starts a workspace  -> start_count = 1 -> running = true
# When Coder stops a workspace   -> start_count = 0 -> running = false
# When Coder deletes a workspace -> terraform destroy
#
# We replicate this with a simple variable:
#
#   terraform apply -var="running=true"   # simulate workspace start
#   terraform apply -var="running=false"  # simulate workspace stop
#   terraform destroy                     # simulate workspace delete

terraform {
  required_providers {
    lightning = {
      source = "rodriguezst/lightningai"
    }
  }
}

provider "lightning" {
  # Reads from env: LIGHTNING_API_KEY, LIGHTNING_USER_ID, LIGHTNING_PROJECT_ID
}

# ---------- variables ----------

variable "running" {
  description = "Simulates Coder's start_count > 0 (true = started, false = stopped)"
  type        = bool
  default     = true
}

variable "studio_name" {
  description = "Name of the studio to create"
  type        = string
  default     = "lifecycle-test"
}

variable "machine" {
  description = "Machine type"
  type        = string
  default     = "cpu-4"
}

# ---------- resource ----------

resource "lightning_studio" "workspace" {
  name    = var.studio_name
  machine = var.machine
  running = var.running

  startup_script = <<-EOT
    #!/bin/bash
    set -e
    echo "startup script running at $(date)" > /tmp/startup-marker.txt
    echo "hostname: $(hostname)" >> /tmp/startup-marker.txt
    echo "user: $(whoami)" >> /tmp/startup-marker.txt
    echo "startup script completed successfully"
  EOT

  startup_script_mode = "always"
  startup_timeout     = "5m"
}

# ---------- outputs ----------

output "studio_id" {
  description = "The studio's unique identifier"
  value       = lightning_studio.workspace.id
}

output "studio_status" {
  description = "Current status of the studio"
  value       = lightning_studio.workspace.status
}

output "studio_running" {
  description = "Whether the studio is running"
  value       = lightning_studio.workspace.running
}

output "studio_public_ip" {
  description = "Public IP if available"
  value       = lightning_studio.workspace.public_ip
}
