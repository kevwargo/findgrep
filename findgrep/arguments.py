import builtins
from argparse import ArgumentParser, HelpFormatter
from collections import defaultdict
from functools import partial


class Arguments:
    def __init__(self):
        self._options = {}

    def register_option(self, name: str, opt: dict):
        self._options[name] = opt

    def __setattr__(self, name: str, value: bool | str | None):
        try:
            self._options[name]["resolved"] = value
        except (AttributeError, KeyError):
            super().__setattr__(name, value)


def parse_cmdline(config: dict) -> Arguments:
    parser = ArgumentParser(formatter_class=partial(HelpFormatter, max_help_position=40))
    args = Arguments()
    mutex_groups = defaultdict(parser.add_mutually_exclusive_group)

    for section, options in config.items():
        for name, opt in options.items():
            if type_ := opt.get("type"):
                kwargs = {"type": getattr(builtins, type_)}
            else:
                kwargs = {"action": "store_true"}
                if opt.get("value"):
                    name = f"no-{name}"

            opt["name"] = name

            if section == "find":
                kwargs["help"] = f"Adds '{' '.join(opt['target'])}' to find"
            elif section == "grep":
                kwargs["help"] = f"Adds '{opt['target']}' to grep"

            arg_container = mutex_groups[g] if (g := opt.get("mutex-group")) else parser
            arg = arg_container.add_argument(f"-{opt['alias']}", f"--{name}", **kwargs)
            args.register_option(arg.dest, opt)

    parser.add_argument("--print-cmd", action="store_true")
    parser.add_argument("--print-elisp-transient", action="store_true")
    parser.add_argument("regexps", nargs="*")

    return parser.parse_args(namespace=args)
