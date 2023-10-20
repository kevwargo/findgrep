from copy import deepcopy
from pathlib import Path

import yaml

CONFIG_FILE = ".findgrep.yml"

DEFAULT_CONFIG = {
    "find": {
        # Path options
        "exclude-cdk": {"alias": "c", "target": ["!", "-path", "*/cdk.out/*"], "value": True},
        "exclude-node-modules": {"alias": "N", "target": ["!", "-path", "*/node_modules/*"], "value": True},
        "exclude-git": {"alias": "G", "target": ["!", "-path", "*/.git/*"], "value": True},
        "exclude-venv": {"alias": "v", "target": ["!", "-path", "*/.venv/*"], "value": True},
        "exclude-cover": {"alias": "V", "target": ["!", "-path", "*/cover/*"], "value": True},
        "exclude-serverless": {"alias": "S", "target": ["!", "-path", "*/.serverless/*"], "value": True},
        # Exclude name options
        "exclude-autosave": {"alias": "~", "target": ["!", "-name", "*~"], "value": True},
        "exclude-temp-sockets": {"alias": ".#", "target": ["!", "-name", ".#*"], "value": True},
        "exclude-emacs-temp": {"alias": "#", "target": ["!", "-name", "#*#"], "value": True},
        "exclude-d-ts": {"alias": "D", "target": ["!", "-name", "*.d.ts"], "value": True},
        "exclude-js": {"alias": "j", "target": ["!", "-name", "*.js"], "value": True},
        "exclude-locks": {
            "alias": "L",
            "target": ["!", "-name", "package-lock.json", "!", "-name", "yarn.lock", "!", "-name", "Pipfile.lock"],
            "value": True,
        },
        "exclude-tests": {
            "alias": "T",
            "target": ["!", "-name", "test_*.py", "!", "-name", "*_test.go"],
            "value": True,
        },
        # File selection options
        "only-go": {"alias": "g", "target": ["-name", "*.go"], "mutex-group": "select"},
        "only-python": {"alias": "p", "target": ["-name", "*.py"], "mutex-group": "select"},
        "only-typescript": {"alias": "t", "target": ["-name", "*.ts"], "mutex-group": "select"},
        "only-java": {"alias": "J", "target": ["-name", "*.java"], "mutex-group": "select"},
        "only-graphql": {"alias": "Q", "target": ["-name", "*.graphql"], "mutex-group": "select"},
        "only-el": {"alias": "E", "target": ["-name", "*.el"], "mutex-group": "select"},
    },
    "grep": {
        "ignore-binary": {"alias": "I", "target": "-I", "value": True},
        "line-numbers": {"alias": "n", "target": "-n", "value": True},
        "show-filename": {"alias": "H", "target": "-H", "value": True},
        "filenames-only": {"alias": "l", "target": "-l"},
        "fixed-string": {"alias": "F", "target": "-F", "mutex-group": "pattern"},
        "extended-regexp": {"alias": "e", "target": "-E", "mutex-group": "pattern"},
        "perl-regexp": {"alias": "P", "target": "-P", "mutex-group": "pattern"},
        "whole-word": {"alias": "w", "target": "-w"},
        "ignore-case": {"alias": "i", "target": "-i"},
        "before": {"alias": "B", "target": "-B", "type": "int"},
        "after": {"alias": "A", "target": "-A", "type": "int"},
    },
}


def load_config() -> dict:
    config_files = _find_files(Path.cwd())
    config = deepcopy(DEFAULT_CONFIG)
    for config_file in config_files:
        local_config = yaml.load(config_file.read_text(), Loader=yaml.SafeLoader)
        if isinstance(local_config, dict):
            _merge_in(config, local_config)

    for section, options in config.items():
        config[section] = {name: opt for name, opt in options.items() if not opt.get("disabled")}

    return config


def _find_files(start: Path) -> list[Path]:
    files = []

    while start != start.parent:
        config_file = start / CONFIG_FILE
        if config_file.is_file():
            files.append(config_file)
        start = start.parent

    return reversed(files)


def _merge_in(target: dict, source: dict):
    for k, v in target.items():
        if k not in source:
            continue
        if isinstance(v, dict):
            _merge_in(v, source[k])
        else:
            target[k] = source[k]

    for k, v in source.items():
        if k not in target:
            target[k] = v
