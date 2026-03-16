# terraform-provider-lightningai

Terraform provider for managing [Lightning AI](https://lightning.ai) Studios.

## Features

- Create, start, stop, and delete Lightning AI Studios
- Execute startup scripts after studio creation
- Immutable design: startup script changes trigger studio replacement
- Coder workspace integration support

## Requirements

- [Go](https://golang.org/) 1.21+
- [Terraform](https://www.terraform.io/) 1.0+

## Building

```sh
make build
```

## Installing Locally

```sh
make install
```

This copies the provider binary to `~/.terraform.d/plugins/lightningai/lightning/1.0.0/<platform>/`.

## Usage

### Provider Configuration

```hcl
provider "lightning" {
  api_key    = "your-api-key"       # or set LIGHTNING_API_KEY
  user_id    = "your-user-id"       # or set LIGHTNING_USER_ID
  project_id = "your-project-id"
}
```

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
| `name`               | string | Yes      | Studio name                                                    |
| `machine`            | string | No       | Machine type (used when starting studio)                       |
| `running`            | bool   | No       | Desired runtime state (default: `true`)                        |
| `interruptible`      | bool   | No       | Use spot/preemptible compute                                   |
| `startup_script`     | string | No       | Script executed after studio starts. Changes force replacement |
| `startup_script_mode`| string | No       | `once` or `always` (default: `once`)                           |
| `startup_timeout`    | string | No       | Max time for script execution (default: `10m`)                 |
| `id`                 | string | Computed | Studio unique identifier                                       |
| `status`             | string | Computed | Current state                                                  |
| `public_ip`          | string | Computed | Public IP address, if available                                |

### Machine Types

| Machine      | Description                     |
|--------------|---------------------------------|
| `cpu-2`      | 2 vCPU, 7.5 GB RAM              |
| `cpu-4`      | 4 vCPU, 15 GB RAM               |
| `cpu-8`      | 8 vCPU, 30 GB RAM               |
| `lit-l4-1`   | 1x L4 GPU, 4 vCPU, 22 GB RAM   |
| `lit-a100-1` | 1x A100 GPU, 12 vCPU, 85 GB RAM |

## Authentication

Set via environment variables:

```sh
export LIGHTNING_API_KEY="your-api-key"
export LIGHTNING_USER_ID="your-user-id"
```

Or configure directly in the provider block.

## Coder Integration

See [examples/coder-template](examples/coder-template/) for an example of using this provider with Coder workspaces.
