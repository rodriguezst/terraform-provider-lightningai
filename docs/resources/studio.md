---
page_title: "lightning_studio Resource - Lightning AI"
description: "Manages the lifecycle of a Lightning AI Studio."
---

# lightning_studio (Resource)

Manages the lifecycle of a Lightning AI Studio. This resource creates a studio,
optionally starts it on a specified machine type, and can execute startup scripts
after the studio is fully ready.

## Example Usage

### Basic Studio

```hcl
resource "lightning_studio" "example" {
  name    = "my-studio"
  machine = "cpu-4"
  running = true
}
```

### Studio with Startup Script

```hcl
resource "lightning_studio" "with_script" {
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
```

### Coder Integration

```hcl
resource "lightning_studio" "workspace" {
  name    = "coder-${data.coder_workspace.me.owner}-${data.coder_workspace.me.name}"
  machine = "cpu-4"
  running = data.coder_workspace.me.start_count > 0

  startup_script      = file("${path.module}/bootstrap.sh")
  startup_script_mode = "always"
  startup_timeout     = "20m"
}
```

## Startup Script Behavior

- **`once` mode** (default): The startup script runs only when the studio is first created.
- **`always` mode**: The startup script runs on every start, including restarts after
  `running` transitions from `false` to `true`.

The provider waits for the studio to be fully ready (filesystem restore complete) before
executing startup scripts. This ensures that persisted user data in
`/teamspace/studios/this_studio` is available when the script runs.

## Machine Types

| Machine      | Description                       |
|--------------|-----------------------------------|
| `cpu-2`      | 2 vCPU, 7.5 GB RAM               |
| `cpu-4`      | 4 vCPU, 15 GB RAM                |
| `cpu-8`      | 8 vCPU, 30 GB RAM                |
| `lit-t4-1`   | 1x T4 GPU, 4 vCPU, 15 GB RAM    |
| `lit-l4-1`   | 1x L4 GPU, 4 vCPU, 22 GB RAM    |
| `lit-a100-1` | 1x A100 GPU, 12 vCPU, 85 GB RAM  |

## Schema

### Required

- `name` (String) Studio name. Changes force replacement.

### Optional

- `machine` (String) Machine type used when starting the studio (e.g., `cpu-4`, `lit-l4-1`).
- `running` (Boolean) Desired runtime state of the studio. Defaults to `true`.
- `interruptible` (Boolean) Use spot/preemptible compute. Defaults to `false`.
- `startup_script` (String) Multiline script executed after studio start. Changes trigger resource replacement.
- `startup_script_mode` (String) When to run the startup script: `once` (only at creation) or `always` (every start). Defaults to `once`.
- `startup_timeout` (String) Maximum time to wait for startup script execution (e.g., `10m`, `30m`). Defaults to `10m`.
- `public_ip` (String) Public IP address of the studio, if available.

### Read-Only

- `id` (String) Studio unique identifier.
- `status` (String) Current state of the studio.
