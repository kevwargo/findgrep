import sys
from itertools import chain
from subprocess import run as run_cmd

from findgrep.arguments import parse_cmdline
from findgrep.config import load_config
from findgrep.elisp_transient import print_elisp_transient


def run():
    config = load_config()
    args = parse_cmdline(config)

    if args.print_elisp_transient:
        return print_elisp_transient(args)

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


def build_opt_value_list(opt: dict) -> list[str]:
    resolved = opt["resolved"]
    target = opt["target"]

    if bool(opt.get("value")) != (resolved is None or resolved is False):
        return []

    target = target if isinstance(target, list) else [target]
    if not isinstance(resolved, bool):
        target.append(str(resolved))

    return target
