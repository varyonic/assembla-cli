from __future__ import annotations

import click

from assembla_cli.formatting import print_detail, print_json, print_table


@click.group()
def user():
    """User information."""


@user.command("me")
@click.option("--json", "as_json", is_flag=True, help="Output as JSON")
@click.pass_context
def user_me(ctx, as_json):
    """Show the authenticated user."""
    client = ctx.obj["client"]
    data = client.get("/user")
    if as_json:
        print_json(data)
    else:
        print_detail(data, ["id", "login", "name", "email"])


@user.command("list")
@click.option("--json", "as_json", is_flag=True, help="Output as JSON")
@click.pass_context
def user_list(ctx, as_json):
    """List users in the configured space."""
    client = ctx.obj["client"]
    space = ctx.obj["space"]
    data = client.get(f"/spaces/{space}/users")
    if as_json:
        print_json(data)
    else:
        print_table(
            data or [],
            columns=["id", "login", "name", "email"],
            headers=["ID", "Login", "Name", "Email"],
        )
