import os
from pathlib import Path

import yaml

CONFIG_DIR = Path.home() / ".config" / "assembla"
GLOBAL_CONFIG_FILE = CONFIG_DIR / "config.yml"


def _find_project_config() -> Path | None:
    """Walk up from cwd looking for .assembla.yml."""
    current = Path.cwd()
    for directory in [current, *current.parents]:
        candidate = directory / ".assembla.yml"
        if candidate.is_file():
            return candidate
    return None


def _read_yaml(path: Path) -> dict:
    if path.is_file():
        with open(path) as f:
            return yaml.safe_load(f) or {}
    return {}


def load_config() -> dict:
    """Load config with precedence: env vars > project .assembla.yml > ~/.config/assembla/config.yml."""
    config: dict = {}

    # 1. Global config (lowest priority)
    config.update(_read_yaml(GLOBAL_CONFIG_FILE))

    # 2. Project config
    project_file = _find_project_config()
    if project_file:
        config.update(_read_yaml(project_file))
        config["_project_config"] = str(project_file)

    # 3. Env vars (highest priority)
    env_map = {
        "ASSEMBLA_API_KEY": "api_key",
        "ASSEMBLA_API_SECRET": "api_secret",
        "ASSEMBLA_SPACE": "space",
        "ASSEMBLA_API_URL": "api_url",
    }
    for env_var, key in env_map.items():
        value = os.environ.get(env_var)
        if value:
            config[key] = value
            config.setdefault("_source", {})[key] = "env"

    config.setdefault("api_url", "https://api.assembla.com")

    return config


def save_global_config(data: dict) -> Path:
    """Save config to ~/.config/assembla/config.yml."""
    CONFIG_DIR.mkdir(parents=True, exist_ok=True)
    existing = _read_yaml(GLOBAL_CONFIG_FILE)
    existing.update(data)
    with open(GLOBAL_CONFIG_FILE, "w") as f:
        yaml.safe_dump(existing, f, default_flow_style=False)
    return GLOBAL_CONFIG_FILE


def save_project_config(data: dict, path: Path | None = None) -> Path:
    """Save config to .assembla.yml in the given directory (default: cwd)."""
    target = path or (Path.cwd() / ".assembla.yml")
    existing = _read_yaml(target)
    existing.update(data)
    with open(target, "w") as f:
        yaml.safe_dump(existing, f, default_flow_style=False)
    return target
