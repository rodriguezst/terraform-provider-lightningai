---
page_title: "Lightning AI Provider"
description: "The Lightning AI provider manages Lightning AI Studios. Use it to create, start, stop, and delete studios with optional startup scripts."
---

# Lightning AI Provider

> **⚠️ Disclaimer**: This Terraform provider is **unofficial** and **not affiliated with, endorsed by, or connected to Lightning AI** in any way. It is provided **without any warranty** of any kind, express or implied. The authors and contributors are **not responsible** for any damages, costs, or issues arising from its use, including but not limited to unexpected charges (e.g., machines not powering off properly, leaving compute running unintentionally, etc.). Use this provider at your own risk.

The Lightning AI provider allows you to manage [Lightning AI](https://lightning.ai) Studios
as Terraform resources. Use it to create, start, stop, and delete studios, execute startup
scripts, and integrate with platforms like [Coder](https://coder.com/).

## Features

- Create, start, stop, and delete Lightning AI Studios
- Execute startup scripts after studio creation or on every start
- Wait for full studio readiness (filesystem restore) before running startup scripts
- Drift detection: sync `running` state with actual studio status
- Immutable design: startup script changes trigger studio replacement

## Example Usage

```hcl
provider "lightning" {
  api_key    = var.lightning_api_key    # or set LIGHTNING_API_KEY
  user_id    = var.lightning_user_id    # or set LIGHTNING_USER_ID
  project_id = var.lightning_project_id # or set LIGHTNING_PROJECT_ID
}

resource "lightning_studio" "example" {
  name    = "my-studio"
  machine = "cpu-4"
  running = true

  startup_script      = file("${path.module}/bootstrap.sh")
  startup_script_mode = "once"
  startup_timeout     = "10m"
}
```

## Authentication

The provider requires three credentials, each of which can be set as a provider
attribute or via environment variable:

| Attribute    | Environment Variable   | Description                 |
|--------------|------------------------|-----------------------------|
| `api_key`    | `LIGHTNING_API_KEY`    | Lightning AI API key        |
| `user_id`    | `LIGHTNING_USER_ID`    | Lightning AI user ID        |
| `project_id` | `LIGHTNING_PROJECT_ID` | Lightning AI project/teamspace ID |

## Schema

### Optional

- `api_key` (String, Sensitive) Lightning AI API key. Can be set via `LIGHTNING_API_KEY` environment variable.
- `user_id` (String) Lightning AI user ID. Can be set via `LIGHTNING_USER_ID` environment variable.
- `project_id` (String) Lightning AI project/teamspace ID. Can be set via `LIGHTNING_PROJECT_ID` environment variable.
