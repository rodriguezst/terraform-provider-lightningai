# Lightning AI Coder Template

This template creates a Lightning AI Studio as a [Coder](https://coder.com) workspace.

## Usage

1. Copy this directory into your Coder template repository.
2. Create a `bootstrap.sh` script with your workspace initialization logic.
3. Set the required variables in your Coder template:
   - `lightning_api_key` — your Lightning AI API key
   - `lightning_user_id` — your Lightning AI user ID
   - `lightning_project_id` — your Lightning AI project/teamspace ID

## How It Works

- When the Coder workspace starts (`start_count > 0`), the Lightning AI Studio is started.
- On the first start the `startup_script` (bootstrap.sh) is executed once inside the studio.
- When the workspace is stopped, the studio is stopped accordingly.

## Machine Types

| Machine     | Description                    |
|-------------|--------------------------------|
| cpu-2       | 2 vCPU, 7.5 GB RAM             |
| cpu-4       | 4 vCPU, 15 GB RAM              |
| cpu-8       | 8 vCPU, 30 GB RAM              |
| lit-l4-1    | 1x L4 GPU, 4 vCPU, 22 GB RAM  |
| lit-a100-1  | 1x A100 GPU, 12 vCPU, 85 GB RAM |
