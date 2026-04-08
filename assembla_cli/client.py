from __future__ import annotations

import sys
from typing import Any

import click
import requests


class AssemblaClient:
    """HTTP client for the Assembla REST API v1."""

    def __init__(self, api_key: str, api_secret: str, api_url: str = "https://api.assembla.com"):
        self.api_url = api_url.rstrip("/")
        self.session = requests.Session()
        self.session.headers.update({
            "X-Api-Key": api_key,
            "X-Api-Secret": api_secret,
            "Content-Type": "application/json",
        })

    # -- generic request helpers ------------------------------------------

    def _request(self, method: str, path: str, **kwargs: Any) -> requests.Response:
        url = f"{self.api_url}/v1{path}"
        resp = self.session.request(method, url, **kwargs)
        if not resp.ok:
            _handle_error(resp)
        return resp

    def get(self, path: str, **params: Any) -> Any:
        resp = self._request("GET", path, params=params)
        return resp.json() if resp.content else None

    def post(self, path: str, data: dict | None = None) -> Any:
        resp = self._request("POST", path, json=data)
        return resp.json() if resp.content else None

    def put(self, path: str, data: dict | None = None) -> Any:
        resp = self._request("PUT", path, json=data)
        return resp.json() if resp.content else None

    def delete(self, path: str) -> None:
        self._request("DELETE", path)

    # -- pagination -------------------------------------------------------

    def get_all(self, path: str, per_page: int = 25, **params: Any) -> list[Any]:
        """Fetch all pages of a paginated endpoint."""
        results: list[Any] = []
        page = 1
        while True:
            page_params = {**params, "page": page, "per_page": per_page}
            data = self.get(path, **page_params)
            if not data:
                break
            results.extend(data)
            if len(data) < per_page:
                break
            page += 1
        return results


def _handle_error(resp: requests.Response) -> None:
    try:
        body = resp.json()
        msg = body.get("error", body.get("errors", resp.text))
    except Exception:
        msg = resp.text
    click.echo(f"Error {resp.status_code}: {msg}", err=True)
    sys.exit(1)
