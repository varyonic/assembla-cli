from __future__ import annotations

import click

from assembla_cli.formatting import print_detail, print_json, print_table


@click.group()
def ticket():
    """Manage tickets."""


@ticket.command("list")
@click.option("-s", "--status", help="Filter by status name")
@click.option("-a", "--assignee", help="Filter by assignee login")
@click.option("-m", "--milestone", help="Filter by milestone ID")
@click.option("-p", "--page", default=1, type=int, help="Page number")
@click.option("-n", "--per-page", default=25, type=int, help="Results per page")
@click.option("--json", "as_json", is_flag=True, help="Output as JSON")
@click.pass_context
def ticket_list(ctx, status, assignee, milestone, page, per_page, as_json):
    """List tickets in the space."""
    client = ctx.obj["client"]
    space = ctx.obj["space"]
    params = {"page": page, "per_page": per_page}
    if status:
        params["ticket_status"] = status
    if assignee:
        params["assigned_to_id"] = assignee
    if milestone:
        params["milestone_id"] = milestone
    data = client.get(f"/spaces/{space}/tickets", **params)
    if as_json:
        print_json(data)
    else:
        print_table(
            data or [],
            columns=["number", "summary", "status_name", "assigned_to_id", "priority"],
            headers=["#", "Summary", "Status", "Assignee", "Priority"],
        )


@ticket.command("show")
@click.argument("number", type=int)
@click.option("--json", "as_json", is_flag=True, help="Output as JSON")
@click.pass_context
def ticket_show(ctx, number, as_json):
    """Show a ticket by number."""
    client = ctx.obj["client"]
    space = ctx.obj["space"]
    data = client.get(f"/spaces/{space}/tickets/{number}")
    if as_json:
        print_json(data)
    else:
        print_detail(data, [
            "number", "summary", "description", "status_name",
            "priority", "assigned_to_id", "milestone_id",
            "created_on", "updated_at",
        ])


@ticket.command("create")
@click.option("-t", "--title", required=True, help="Ticket summary/title")
@click.option("-d", "--description", default="", help="Ticket description")
@click.option("-s", "--status", help="Status name")
@click.option("-p", "--priority", type=int, help="Priority (1=highest)")
@click.option("-a", "--assignee", help="Assignee user ID")
@click.option("-m", "--milestone", help="Milestone ID")
@click.option("--json", "as_json", is_flag=True, help="Output as JSON")
@click.pass_context
def ticket_create(ctx, title, description, status, priority, assignee, milestone, as_json):
    """Create a new ticket."""
    client = ctx.obj["client"]
    space = ctx.obj["space"]
    payload: dict = {"ticket": {"summary": title, "description": description}}
    if status:
        payload["ticket"]["status"] = status
    if priority is not None:
        payload["ticket"]["priority"] = priority
    if assignee:
        payload["ticket"]["assigned_to_id"] = assignee
    if milestone:
        payload["ticket"]["milestone_id"] = milestone
    data = client.post(f"/spaces/{space}/tickets", data=payload)
    if as_json:
        print_json(data)
    else:
        click.echo(f"Created ticket #{data['number']}: {data['summary']}")


@ticket.command("update")
@click.argument("number", type=int)
@click.option("-t", "--title", help="New summary/title")
@click.option("-d", "--description", help="New description")
@click.option("-s", "--status", help="New status name")
@click.option("-p", "--priority", type=int, help="New priority")
@click.option("-a", "--assignee", help="New assignee user ID")
@click.option("-m", "--milestone", help="New milestone ID")
@click.option("--json", "as_json", is_flag=True, help="Output as JSON")
@click.pass_context
def ticket_update(ctx, number, title, description, status, priority, assignee, milestone, as_json):
    """Update a ticket by number."""
    client = ctx.obj["client"]
    space = ctx.obj["space"]
    payload: dict = {"ticket": {}}
    if title is not None:
        payload["ticket"]["summary"] = title
    if description is not None:
        payload["ticket"]["description"] = description
    if status is not None:
        payload["ticket"]["status"] = status
    if priority is not None:
        payload["ticket"]["priority"] = priority
    if assignee is not None:
        payload["ticket"]["assigned_to_id"] = assignee
    if milestone is not None:
        payload["ticket"]["milestone_id"] = milestone
    if not payload["ticket"]:
        click.echo("Nothing to update. Provide at least one option.", err=True)
        raise SystemExit(1)
    data = client.put(f"/spaces/{space}/tickets/{number}", data=payload)
    if as_json:
        print_json(data)
    else:
        click.echo(f"Updated ticket #{number}")


@ticket.command("move")
@click.argument("number", type=int)
@click.argument("status")
@click.pass_context
def ticket_move(ctx, number, status):
    """Move a ticket to a new status.

    Usage: assembla ticket move 123 "Ready for Review"
    """
    client = ctx.obj["client"]
    space = ctx.obj["space"]
    client.put(f"/spaces/{space}/tickets/{number}", data={"ticket": {"status": status}})
    click.echo(f"Ticket #{number} moved to \"{status}\"")
