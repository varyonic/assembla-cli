from __future__ import annotations

import click

from assembla_cli.formatting import print_detail, print_json, print_table


@click.group()
def milestone():
    """Manage milestones."""


@milestone.command("list")
@click.option("--all", "show_all", is_flag=True, help="Show all milestones (not just upcoming)")
@click.option("--completed", is_flag=True, help="Show completed milestones")
@click.option("--json", "as_json", is_flag=True, help="Output as JSON")
@click.pass_context
def milestone_list(ctx, show_all, completed, as_json):
    """List milestones in the space."""
    client = ctx.obj["client"]
    space = ctx.obj["space"]
    if show_all:
        path = f"/spaces/{space}/milestones/all"
    elif completed:
        path = f"/spaces/{space}/milestones/completed"
    else:
        path = f"/spaces/{space}/milestones/upcoming"
    data = client.get(path)
    if as_json:
        print_json(data)
    else:
        print_table(
            data or [],
            columns=["id", "title", "due_date", "is_completed", "planner_type"],
            headers=["ID", "Title", "Due Date", "Completed", "Type"],
        )


@milestone.command("show")
@click.argument("milestone_id")
@click.option("--json", "as_json", is_flag=True, help="Output as JSON")
@click.pass_context
def milestone_show(ctx, milestone_id, as_json):
    """Show milestone details."""
    client = ctx.obj["client"]
    space = ctx.obj["space"]
    data = client.get(f"/spaces/{space}/milestones/{milestone_id}")
    if as_json:
        print_json(data)
    else:
        print_detail(data, [
            "id", "title", "description", "due_date",
            "is_completed", "planner_type", "created_at", "updated_at",
        ])
