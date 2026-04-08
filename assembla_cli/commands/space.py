from __future__ import annotations

import click

from assembla_cli.formatting import print_detail, print_json, print_table


@click.group()
def space():
    """Manage spaces."""


@space.command("list")
@click.option("--json", "as_json", is_flag=True, help="Output as JSON")
@click.pass_context
def space_list(ctx, as_json):
    """List all spaces you belong to."""
    client = ctx.obj["client"]
    data = client.get("/spaces")
    if as_json:
        print_json(data)
    else:
        print_table(
            data or [],
            columns=["id", "wiki_name", "name", "status"],
            headers=["ID", "Wiki Name", "Name", "Status"],
        )


@space.command("show")
@click.argument("space_id", required=False)
@click.option("--json", "as_json", is_flag=True, help="Output as JSON")
@click.pass_context
def space_show(ctx, space_id, as_json):
    """Show space details. Uses configured space if not provided."""
    client = ctx.obj["client"]
    space_id = space_id or ctx.obj["space"]
    data = client.get(f"/spaces/{space_id}")
    if as_json:
        print_json(data)
    else:
        print_detail(data, [
            "id", "wiki_name", "name", "status",
            "created_at", "updated_at", "description",
        ])
