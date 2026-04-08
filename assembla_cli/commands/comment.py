from __future__ import annotations

import click

from assembla_cli.formatting import print_json, print_table


@click.group()
def comment():
    """Manage ticket comments."""


@comment.command("list")
@click.argument("ticket_number", type=int)
@click.option("--json", "as_json", is_flag=True, help="Output as JSON")
@click.pass_context
def comment_list(ctx, ticket_number, as_json):
    """List comments on a ticket."""
    client = ctx.obj["client"]
    space = ctx.obj["space"]
    data = client.get(f"/spaces/{space}/tickets/{ticket_number}/ticket_comments")
    if as_json:
        print_json(data)
    else:
        print_table(
            data or [],
            columns=["id", "author_name", "comment", "created_on"],
            headers=["ID", "Author", "Comment", "Created"],
        )


@comment.command("add")
@click.argument("ticket_number", type=int)
@click.argument("body")
@click.option("--json", "as_json", is_flag=True, help="Output as JSON")
@click.pass_context
def comment_add(ctx, ticket_number, body, as_json):
    """Add a comment to a ticket.

    Usage: assembla comment add 123 "This is my comment"
    """
    client = ctx.obj["client"]
    space = ctx.obj["space"]
    data = client.post(
        f"/spaces/{space}/tickets/{ticket_number}/ticket_comments",
        data={"ticket_comment": {"comment": body}},
    )
    if as_json:
        print_json(data)
    else:
        click.echo(f"Comment added to ticket #{ticket_number}")
