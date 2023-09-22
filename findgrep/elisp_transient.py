from typing import Iterable, Iterator

HEADER = """
;; This expression is intended to be returned by `transient-setup-children'
;; and a transient group will be built from it.
"""


def print_elisp_transient(config: dict):
    print(HEADER)

    print(f"({render_sections(config)})")


def render_sections(config: dict) -> str:
    return "\n".join(
        [
            render_section("Exclude paths", filter_options(config["find"].values(), "!", "-path")),
            render_section("Exclude files", filter_options(config["find"].values(), "!", "-name"), True),
            render_section("Select files", filter_options(config["find"].values(), "-name"), True),
            render_section("Grep args", filter_options(config["grep"].values()), True),
        ]
    )


def filter_options(options: list[dict], *target_start) -> Iterator[dict]:
    for opt in options:
        if not target_start:
            yield opt
        if opt["target"][: len(target_start)] == list(target_start):
            yield opt


def render_section(name: str, options: Iterable[dict], indent=False) -> str:
    rendered_options = "\n".join(render_option(o) for o in options)
    prefix = " " if indent else ""
    return f'{prefix}["{name}"\n{rendered_options}]'


def render_option(opt: dict) -> str:
    key = opt["alias"]

    words = opt["name"].split("-")
    if words[0] == "no":
        words[0] = "Don't"
    words[0] = words[0].capitalize()
    name = " ".join(words)

    short_arg = f'-{opt["alias"]}'
    long_arg = f'--{opt["name"]}'

    if opt.get("type"):
        long_arg += "="
        transient_class = "option"
    else:
        transient_class = "switch"

    if g := opt.get("mutex-group"):
        mutex_group = f" :mutex-group {g}"
        transient_class = f"findgrep--{transient_class}-mutex"
    else:
        mutex_group = ""
        transient_class = f"transient-{transient_class}"

    return f'  ("{key}" "{name}" ("{short_arg}" "{long_arg}") :class {transient_class}{mutex_group})'
