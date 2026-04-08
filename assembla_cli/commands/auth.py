from __future__ import annotations

import click
import requests

from assembla_cli.config import (
    GLOBAL_CONFIG_FILE,
    load_config,
    save_global_config,
    save_project_config,
)


@click.group()
def auth():
    """Authenticate with Assembla."""


@auth.command("login")
@click.option(
    "--scope", type=click.Choice(["global", "project"]), default="global",
    help="Where to store credentials (default: global ~/.config/assembla/config.yml)",
)
def auth_login(scope):
    """Interactive login — stores API key, secret, and default space.

    Get your credentials at: https://www.assembla.com/user/edit/manage_clients
    """
    click.echo("Get your API key/secret at: https://www.assembla.com/user/edit/manage_clients\n")

    api_key = click.prompt("API Key")
    api_secret = click.prompt("API Secret")

    # Verify credentials
    click.echo("\nVerifying credentials... ", nl=False)
    resp = requests.get(
        "https://api.assembla.com/v1/user",
        headers={"X-Api-Key": api_key, "X-Api-Secret": api_secret},
    )
    if not resp.ok:
        click.echo("FAILED")
        click.echo(f"Error {resp.status_code}: {resp.text}", err=True)
        raise SystemExit(1)

    user = resp.json()
    click.echo(f"OK (logged in as {user.get('name', user.get('login'))})")

    # List spaces and let user pick a default
    resp = requests.get(
        "https://api.assembla.com/v1/spaces",
        headers={"X-Api-Key": api_key, "X-Api-Secret": api_secret},
    )
    space_id = None
    if resp.ok:
        spaces = resp.json()
        if spaces:
            click.echo(f"\nAvailable spaces ({len(spaces)}):")
            for i, s in enumerate(spaces, 1):
                click.echo(f"  {i}. {s['name']} ({s['wiki_name']})")
            choice = click.prompt(
                "\nDefault space (number or wiki_name, Enter to skip)",
                default="", show_default=False,
            )
            if choice:
                if choice.isdigit() and 1 <= int(choice) <= len(spaces):
                    space_id = spaces[int(choice) - 1]["wiki_name"]
                else:
                    space_id = choice

    data = {"api_key": api_key, "api_secret": api_secret}
    if space_id:
        data["space"] = space_id

    if scope == "global":
        path = save_global_config(data)
    else:
        path = save_project_config(data)

    click.echo(f"\nCredentials saved to {path}")
    if space_id:
        click.echo(f"Default space: {space_id}")


@auth.command("logout")
@click.option(
    "--scope", type=click.Choice(["global", "project"]), default="global",
    help="Which config to remove",
)
def auth_logout(scope):
    """Remove stored credentials."""
    if scope == "global":
        if GLOBAL_CONFIG_FILE.is_file():
            GLOBAL_CONFIG_FILE.unlink()
            click.echo(f"Removed {GLOBAL_CONFIG_FILE}")
        else:
            click.echo("No global config found.")
    else:
        from pathlib import Path
        project_file = Path.cwd() / ".assembla.yml"
        if project_file.is_file():
            if click.confirm(f"Remove {project_file}?"):
                project_file.unlink()
                click.echo(f"Removed {project_file}")
        else:
            click.echo("No project config found in current directory.")


@auth.command("status")
def auth_status():
    """Show current authentication status and config sources."""
    config = load_config()

    api_key = config.get("api_key")
    if not api_key:
        click.echo("Not authenticated. Run: assembla auth login")
        return

    masked_key = api_key[:4] + "..." + api_key[-4:] if len(api_key) > 8 else "****"
    click.echo(f"API Key:  {masked_key}")
    click.echo(f"Space:    {config.get('space', '(not set)')}")
    click.echo(f"API URL:  {config.get('api_url')}")

    # Show config sources
    if GLOBAL_CONFIG_FILE.is_file():
        click.echo(f"\nGlobal config:  {GLOBAL_CONFIG_FILE}")
    project_config = config.get("_project_config")
    if project_config:
        click.echo(f"Project config: {project_config}")

    # Verify credentials work
    click.echo("\nVerifying... ", nl=False)
    api_secret = config.get("api_secret")
    if not api_secret:
        click.echo("MISSING API SECRET")
        return
    resp = requests.get(
        f"{config['api_url']}/v1/user",
        headers={"X-Api-Key": api_key, "X-Api-Secret": api_secret},
    )
    if resp.ok:
        user = resp.json()
        click.echo(f"OK (logged in as {user.get('name', user.get('login'))})")
    else:
        click.echo(f"FAILED ({resp.status_code})")
