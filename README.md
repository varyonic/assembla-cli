# assembla-cli

Command-line tool for managing Assembla tickets, comments, milestones, and spaces.

## Installation

```bash
pip install git+https://github.com/eugene-software/assembla-cli.git
```

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

## Usage

```bash
# Auth
assembla auth login
assembla auth status

# Tickets
assembla ticket list
assembla ticket list --status "In Progress" --assignee "john"
assembla ticket show 12345
assembla ticket create --title "Bug report" --description "Details here"
assembla ticket update 12345 --status "Fixed"
assembla ticket move 12345 "In Progress"

# Comments
assembla comment list 12345
assembla comment add 12345 "This is fixed in v2.1"

# Spaces
assembla space list
assembla space show

# Statuses
assembla status list

# Milestones
assembla milestone list
assembla milestone list --all
assembla milestone show <milestone-id>

# Users
assembla user me
assembla user list
```

### Global options

```bash
assembla --space other-space ticket list    # Override space
assembla ticket list --json                 # JSON output
```

## License

MIT
