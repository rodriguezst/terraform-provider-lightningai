# terraform-provider-lightningai

Terraform provider for managing [Lightning AI](https://lightning.ai) Studios.

## Features

- Create, start, stop, and delete Lightning AI Studios
- Execute startup scripts after studio creation or on every start
- Waits for full studio readiness (filesystem restore) before running startup scripts
- Drift detection: syncs `running` state with actual studio status
- Immutable design: startup script changes trigger studio replacement
- Coder workspace integration support

## Requirements

- [Go](https://golang.org/) 1.21+
- [Terraform](https://www.terraform.io/) 1.0+

## Building

```sh
go build -o terraform-provider-lightning .
```

## Development Setup

For local development, configure Terraform to use your locally-built binary by creating a `~/.terraformrc` file:

```hcl
provider_installation {
  dev_overrides {
    "lightningai/lightning" = "/path/to/terraform-provider-lightningai"
  }
  direct {}
}
```

Then build the provider:

```sh
go build -o terraform-provider-lightning .
```

No `terraform init` is needed when using `dev_overrides`.

## Usage

### Provider Configuration

```hcl
provider "lightning" {
  api_key    = "your-api-key"       # or set LIGHTNING_API_KEY
  user_id    = "your-user-id"       # or set LIGHTNING_USER_ID
  project_id = "your-project-id"   # or set LIGHTNING_PROJECT_ID
}
```

All three attributes support environment variable fallback:

| Attribute    | Environment Variable      |
|-------------|--------------------------|
| `api_key`    | `LIGHTNING_API_KEY`       |
| `user_id`    | `LIGHTNING_USER_ID`       |
| `project_id` | `LIGHTNING_PROJECT_ID`    |

### Resource: `lightning_studio`

```hcl
resource "lightning_studio" "example" {
  name    = "my-studio"
  machine = "cpu-4"

  running = true

  startup_script = file("${path.module}/bootstrap.sh")

  startup_script_mode = "once"
  startup_timeout     = "10m"
}
```

#### Attributes

| Attribute            | Type   | Required | Description                                                    |
|----------------------|--------|----------|----------------------------------------------------------------|
| `name`               | string | Yes      | Studio name. Changes force replacement.                        |
| `machine`            | string | No       | Machine type (used when starting studio)                       |
| `running`            | bool   | No       | Desired runtime state (default: `true`)                        |
| `interruptible`      | bool   | No       | Use spot/preemptible compute (default: `false`)                |
| `startup_script`     | string | No       | Script executed after studio starts. Changes force replacement |
| `startup_script_mode`| string | No       | `once` or `always` (default: `once`)                           |
| `startup_timeout`    | string | No       | Max time for script execution (default: `10m`)                 |
| `id`                 | string | Computed | Studio unique identifier                                       |
| `status`             | string | Computed | Current state (e.g. `CLOUD_SPACE_INSTANCE_STATE_RUNNING`)      |
| `public_ip`          | string | Computed | Public IP address, if available                                |

#### Startup Script Behavior

- **`once` mode** (default): The startup script runs only when the studio is first created.
- **`always` mode**: The startup script runs on every start (including restarts after `running` transitions from `false` to `true`).

The provider waits for the studio to be fully ready (filesystem restore complete) before executing startup scripts. This ensures that persisted user data in `/teamspace/studios/this_studio` is available when the script runs.

### Machine Types

| Machine      | Description                      |
|--------------|----------------------------------|
| `cpu-2`      | 2 vCPU, 7.5 GB RAM              |
| `cpu-4`      | 4 vCPU, 15 GB RAM               |
| `cpu-8`      | 8 vCPU, 30 GB RAM               |
| `lit-t4-1`   | 1x T4 GPU, 4 vCPU, 15 GB RAM   |
| `lit-l4-1`   | 1x L4 GPU, 4 vCPU, 22 GB RAM   |
| `lit-a100-1` | 1x A100 GPU, 12 vCPU, 85 GB RAM |

## Authentication

Set via environment variables:

```sh
export LIGHTNING_API_KEY="your-api-key"
export LIGHTNING_USER_ID="your-user-id"
export LIGHTNING_PROJECT_ID="your-project-id"
```

Or configure directly in the provider block.

## Coder Integration

This provider is designed to work with [Coder](https://coder.com/) workspaces. The key pattern is using Coder's `start_count` to control the `running` attribute:

```hcl
resource "lightning_studio" "workspace" {
  name    = "coder-${data.coder_workspace.me.name}"
  machine = data.coder_parameter.machine.value
  running = data.coder_workspace.me.start_count > 0

  startup_script      = data.coder_parameter.startup_script.value
  startup_script_mode = "always"
}
```

When Coder starts a workspace, `start_count > 0` evaluates to `true` and the studio starts. When Coder stops a workspace, it evaluates to `false` and the studio stops. On workspace deletion, `terraform destroy` removes the studio entirely.

See [examples/coder-template](examples/coder-template/) for a complete example.
