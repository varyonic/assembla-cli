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
2. Project config (`.assembla.yml` in current or parent directory)
3. Global config (`~/.config/assembla/config.yml`)

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
