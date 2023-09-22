import builtins
import sys
from argparse import ArgumentParser
from copy import deepcopy
from itertools import chain
from pathlib import Path
from subprocess import run as run_cmd

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
            "target": ["!", "-name", "package-lock.json", "!", "-name", "yarn.lock"],
            "value": True,
        },
        "exclude-tests": {
            "alias": "T",
            "target": ["!", "-name", "test_*.py", "!", "-name", "*_test.go"],
            "value": True,
        },
        # File selection options
        "only-go": {"alias": "g", "target": ["-name", "*.go"]},
        "only-python": {"alias": "p", "target": ["-name", "*.py"]},
        "only-typescript": {"alias": "t", "target": ["-name", "*.ts"]},
        "only-java": {"alias": "J", "target": ["-name", "*.java"]},
        "only-graphql": {"alias": "q", "target": ["-name", "*.graphql"]},
        "only-el": {"alias": "E", "target": ["-name", "*.el"]},
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
    args = parse_cmdline(config)
    cmd = build_command(config, args.regexps)

    if args.print_cmd:
        print(" ".join(cmd))
    elif args.regexps:
        run_cmd(cmd)
    else:
        print("No regexp", file=sys.stderr)
        exit(1)


def build_command(config: dict, regexps: list[str]) -> list[str]:
    return [
        *("find", "."),
        *chain.from_iterable(build_opt_value_list(o) for o in config["find"].values() if "-path" in o["target"]),
        *("-type", "f"),
        *chain.from_iterable(build_opt_value_list(o) for o in config["find"].values() if "-name" in o["target"]),
        *("-exec", "grep", "--color=always"),
        *chain.from_iterable(build_opt_value_list(o) for o in config["grep"].values()),
        *chain.from_iterable(["-e", r] for r in regexps),
        *("{}", "+"),
    ]


class Namespace:
    def __init__(self):
        self._options = {}

    def register_option(self, name: str, opt: dict):
        self._options[name] = opt

    def __setattr__(self, name: str, value: bool | str | None):
        try:
            self._options[name]["resolved"] = value
        except (AttributeError, KeyError):
            super().__setattr__(name, value)


def parse_cmdline(config: dict) -> Namespace:
    parser = ArgumentParser()
    ns = Namespace()

    only_group = parser.add_mutually_exclusive_group()

    for section, options in config.items():
        for disabled in [n for n in options if options[n].get("disabled")]:
            del options[disabled]

        for name, opt in options.items():
            if type_ := opt.get("type"):
                kwargs = {"type": getattr(builtins, type_)}
            else:
                kwargs = {"action": "store_true"}
                if opt.get("value"):
                    name = f"no-{name}"

            if section == "find":
                kwargs["help"] = f"Adds '{' '.join(opt['target'])}' to find"
            elif section == "grep":
                kwargs["help"] = f"Adds '{opt['target']}' to grep"

            if name.startswith("only-"):
                arg = only_group.add_argument(f"-{opt['alias']}", f"--{name}", **kwargs)
            else:
                arg = parser.add_argument(f"-{opt['alias']}", f"--{name}", **kwargs)

            ns.register_option(arg.dest, opt)

    parser.add_argument("--print-cmd", action="store_true")
    parser.add_argument("regexps", nargs="*")

    return parser.parse_args(namespace=ns)


def build_opt_value_list(opt: dict) -> list[str]:
    resolved = opt["resolved"]
    target = opt["target"]

    if bool(opt.get("value")) != (resolved is None or resolved is False):
        return []

    target = target if isinstance(target, list) else [target]
    if not isinstance(resolved, bool):
        target.append(str(resolved))

    return target


def load_config() -> dict:
    config_files = find_config_files(Path.cwd())
    config = deepcopy(DEFAULT_CONFIG)
    for config_file in config_files:
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
