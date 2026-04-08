"""Shared output formatting helpers."""
from __future__ import annotations

import json
from typing import Any, Sequence

import click
from tabulate import tabulate


def print_table(rows: Sequence[dict], columns: list[str], headers: list[str] | None = None) -> None:
    """Print a list of dicts as a table."""
    if not rows:
        click.echo("No results.")
        return
    table = [[row.get(col, "") for col in columns] for row in rows]
    click.echo(tabulate(table, headers=headers or columns, tablefmt="simple"))


def print_json(data: Any) -> None:
    click.echo(json.dumps(data, indent=2, default=str))


def print_detail(data: dict, fields: list[str]) -> None:
    """Print a single record as key: value lines."""
    max_key = max(len(f) for f in fields)
    for field in fields:
        value = data.get(field, "")
        click.echo(f"  {field:<{max_key}}  {value}")
