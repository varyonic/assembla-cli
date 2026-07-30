# Assembla CLI

A command-line interface for the Assembla API. Manage tickets, comments, spaces, milestones, and users from your terminal.

## Installation

### Homebrew (macOS & Linux)

```bash
brew install eugene-software/tap/assembla-cli
```

### Go

```bash
go install github.com/eugene-software/assembla-cli@latest
```

### Binary

Download from [GitHub Releases](https://github.com/eugene-software/assembla-cli/releases).

## Setup

Get your API credentials at [assembla.com/user/edit/manage_clients](https://www.assembla.com/user/edit/manage_clients).

### Option 1: Interactive login

```bash
assembla auth login
```

### Option 2: Environment variables

```bash
export ASSEMBLA_API_KEY="your-api-key"
export ASSEMBLA_API_SECRET="your-api-secret"
export ASSEMBLA_SPACE="your-space-wiki-name"
```

### Option 3: Config file

Create `.assembla.yml` in your project root (or `~/.config/assembla/config.yml` globally):

```yaml
api_key: "your-api-key"
api_secret: "your-api-secret"
space: "your-space-wiki-name"
```

## Quick Start

```bash
# Authenticate (interactive)
assembla auth login

# List tickets
assembla ticket list

# Show a specific ticket
assembla ticket show 12345

# Create a ticket
assembla ticket create -t "Bug: login fails" -d "Steps to reproduce..."

# Move a ticket to a new status
assembla ticket move 12345 "In Progress"

# Add a comment
assembla comment add 12345 "Working on this now"

# List spaces
assembla space list

# List statuses
assembla status list

# List milestones
assembla milestone list

# Show current user
assembla user me
```

## Configuration

Configuration is loaded with the following precedence (highest first):

1. Environment variables (`ASSEMBLA_API_KEY`, `ASSEMBLA_API_SECRET`, `ASSEMBLA_SPACE`, `ASSEMBLA_API_URL`)
2. Project config (`.assembla.yml` in current or parent directory) — `api_key`, `api_secret`, `space`
3. Global config (`~/.config/assembla/config.yml`) — the above plus `api_url`

Only these keys are read; anything else in a config file is ignored with a warning.

`api_url` is deliberately **not** settable from a project `.assembla.yml`. That file
is found by searching parent directories, so it can arrive with a cloned
repository, and it must not be able to decide where your credentials are sent. For
a non-default endpoint use the global config or `ASSEMBLA_API_URL`.

### Global Flags

- `--space` - Override the default space
- `--api-key` - Override the API key
- `--api-secret` - Override the API secret

### JSON Output

All data commands support `--json` for machine-readable output:

```bash
assembla ticket list --json
assembla ticket show 12345 --json
```

## Commands

| Command | Description |
|---------|-------------|
| `auth login` | Authenticate with Assembla |
| `auth logout` | Remove stored credentials |
| `auth status` | Show authentication status |
| `ticket list` | List tickets |
| `ticket show` | Show ticket details |
| `ticket create` | Create a new ticket |
| `ticket update` | Update an existing ticket |
| `ticket move` | Move ticket to a new status |
| `comment list` | List ticket comments |
| `comment add` | Add a comment to a ticket |
| `space list` | List available spaces |
| `space show` | Show space details |
| `status list` | List ticket statuses |
| `milestone list` | List milestones |
| `milestone show` | Show milestone details |
| `user me` | Show current user |
| `user list` | List users in space |

## License

MIT
