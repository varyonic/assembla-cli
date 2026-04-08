from __future__ import annotations

import sys

import click

from assembla_cli.client import AssemblaClient
from assembla_cli.commands.auth import auth
from assembla_cli.commands.comment import comment
from assembla_cli.commands.milestone import milestone
from assembla_cli.commands.space import space
from assembla_cli.commands.status import status
from assembla_cli.commands.ticket import ticket
from assembla_cli.commands.user import user
from assembla_cli.config import load_config


NO_AUTH_COMMANDS = {"auth"}


@click.group()
@click.option("--space", "space_override", envvar="ASSEMBLA_SPACE", help="Space ID or wiki name")
@click.option("--api-key", envvar="ASSEMBLA_API_KEY", help="Assembla API key")
@click.option("--api-secret", envvar="ASSEMBLA_API_SECRET", help="Assembla API secret")
@click.pass_context
def cli(ctx, space_override, api_key, api_secret):
    """Assembla CLI — manage tickets, comments, and spaces."""
    ctx.ensure_object(dict)

    if ctx.invoked_subcommand in NO_AUTH_COMMANDS:
        return

    config = load_config()
    key = api_key or config.get("api_key")
    secret = api_secret or config.get("api_secret")
    if not key or not secret:
        click.echo(
            "Error: Assembla credentials required.\n"
            "Set ASSEMBLA_API_KEY and ASSEMBLA_API_SECRET env vars,\n"
            "or create a .assembla.yml file with api_key and api_secret.\n"
            "Or run: assembla auth login",
            err=True,
        )
        sys.exit(1)

    ctx.obj["client"] = AssemblaClient(
        key, secret, config.get("api_url", "https://api.assembla.com"),
    )
    ctx.obj["space"] = space_override or config.get("space")


cli.add_command(auth)
cli.add_command(ticket)
cli.add_command(comment)
cli.add_command(space)
cli.add_command(status)
cli.add_command(milestone)
cli.add_command(user)


if __name__ == "__main__":
    cli()
