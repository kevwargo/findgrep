import sys
from subprocess import run as run_cmd
from typing import Iterable, Iterator

from findgrep.arguments import parse_cmdline
from findgrep.config import load_config
from findgrep.elisp_transient import print_elisp_transient


def run():
    config = load_config()
    args = parse_cmdline(config)

    if args.print_elisp_transient:
        return print_elisp_transient(config)

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
        *opt_args(o for o in config["find"].values() if "-path" in o["target"]),
        *("-type", "f"),
        *opt_args(o for o in config["find"].values() if "-name" in o["target"]),
        *("-exec", "grep", "--color=always"),
        *opt_args(o for o in config["grep"].values()),
        *regexp_args(regexps),
        *("{}", "+"),
    ]


def opt_args(options: Iterable[dict]) -> Iterator[str]:
    for opt in options:
        resolved = opt["resolved"]
        target = opt["target"]

        if bool(opt.get("value")) != (resolved is None or resolved is False):
            continue

        target = target if isinstance(target, list) else [target]
        if not isinstance(resolved, bool):
            target.append(str(resolved))

        yield from target


def regexp_args(regexps: list[str]) -> Iterator[str]:
    for r in regexps:
        yield "-e"
        yield r
