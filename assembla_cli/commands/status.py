from __future__ import annotations

import click

from assembla_cli.formatting import print_json, print_table


@click.group()
def status():
    """Manage ticket statuses."""


@status.command("list")
@click.option("--json", "as_json", is_flag=True, help="Output as JSON")
@click.pass_context
def status_list(ctx, as_json):
    """List all ticket statuses in the space."""
    client = ctx.obj["client"]
    space = ctx.obj["space"]
    data = client.get(f"/spaces/{space}/tickets/statuses")
    if as_json:
        print_json(data)
    else:
        print_table(
            data or [],
            columns=["id", "name", "state", "list_order"],
            headers=["ID", "Name", "State", "Order"],
        )
