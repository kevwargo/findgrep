import builtins
from argparse import ArgumentParser
from copy import deepcopy
from pathlib import Path

import yaml

CONFIG_FILE = ".findgrep.yml"

DEFAULT_CONFIG = {
    "find": {
        # Path options
        "exclude-cdk": {"alias": "c", "target": ["!", "-path", "*/cdk.out/*"], "value": True},
        "exclude-node-modules": {"alias": "N", "target": ["!", "-path", "*/node_modules/*"], "value": True},
        "exclude-cover": {"alias": "V", "target": ["!", "-path", "*/cover/*"], "value": True},
        "exclude-git": {"alias": "G", "target": ["!", "-path", "*/.git/*"], "value": True},
        "exclude-venv": {"alias": "v", "target": ["!", "-path", "*/.venv/*"], "value": True},
        # Exclude name options
        "exclude-autosave": {"alias": "~", "target": ["!", "-name", "*~"], "value": True},
        "exclude-temp-sockets": {"alias": ".#", "target": ["!", "-name", ".#*"], "value": True},
        "exclude-emacs-temp": {"alias": "#", "target": ["!", "-name", "#*#"], "value": True},
        "exclude-d-ts": {"alias": "D", "target": ["!", "-name", "*.d.ts"], "value": True},
        "exclude-js": {"alias": "j", "target": ["!", "-name", "*.js"], "value": True},
        "exclude-locks": {
            "alias": "L",
            "target": ["!", "-name", "package-lock.json", "!", "-name", "yarn.lock"],
            "value": True,
        },
        "exclude-tests": {
            "alias": "T",
            "target": ["!", "-name", "test_*.py", "!", "-name", "*_test.go"],
            "value": True,
        },
        # Include name options
        "include-go": {"alias": "g", "target": ["-name", "*.go"]},
        "include-python": {"alias": "p", "target": ["-name", "*.py"]},
        "include-typescript": {"alias": "t", "target": ["-name", "*.ts"]},
        "include-java": {"alias": "J", "target": ["-name", "*.java"]},
        "include-graphql": {"alias": "q", "target": ["-name", "*.graphql"]},
        "include-el": {"alias": "E", "target": ["-name", "*.el"]},
    },
    "grep": {
        "ignore-binary": {"alias": "I", "target": "-I", "value": True},
        "line-numbers": {"alias": "n", "target": "-n", "value": True},
        "show-filename": {"alias": "H", "target": "-H", "value": True},
        "filenames-only": {"alias": "l", "target": "-l"},
        "extended-regexp": {"alias": "e", "target": "-E"},
        "perl-regexp": {"alias": "P", "target": "-P"},
        "whole-word": {"alias": "w", "target": "-w"},
        "ignore-case": {"alias": "i", "target": "-i"},
        "before": {"alias": "B", "target": "-B", "type": "int"},
        "after": {"alias": "A", "target": "-A", "type": "int"},
    },
}


def run():
    config = load_config()
    parser = ArgumentParser()

    for options in config.values():
        for name, opt in options.items():
            if opt.get("disabled"):
                continue

            if type_ := opt.get("type"):
                kwargs = {"type": getattr(builtins, type_)}
            else:
                kwargs = {"action": "store_true"}

            parser.add_argument(f"-{opt['alias']}", f"--{name}", **kwargs)

    print(parser.parse_args())


def load_config() -> dict:
    config_files = find_config_files(Path.cwd())
    config = deepcopy(DEFAULT_CONFIG)
    for config_file in config_files:
        print(f"Parsing {config_file}...")
        local_config = yaml.load(config_file.read_text(), Loader=yaml.SafeLoader)
        if isinstance(local_config, dict):
            merge_in(config, local_config)

    return config


def find_config_files(start: Path) -> list[Path]:
    files = []

    while start != start.parent:
        config_file = start / CONFIG_FILE
        if config_file.is_file():
            files.append(config_file)
        start = start.parent

    return reversed(files)


def merge_in(target: dict, source: dict):
    for k, v in target.items():
        if k not in source:
            continue
        if isinstance(v, dict):
            merge_in(v, source[k])
        else:
            target[k] = source[k]

    for k, v in source.items():
        if k not in target:
            target[k] = v
