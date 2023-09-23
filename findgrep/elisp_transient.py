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
    words[0] = "Don't" if words[0] == "no" else words[0].capitalize()
    description = " ".join(words)

    short_arg = f'-{opt["alias"]}'
    long_arg = f'--{opt["name"]}'

    if opt.get("type"):
        long_arg += "="
        argument_class = "findgrep-option"
    else:
        argument_class = "findgrep-switch"

    mutex_group = opt.get("mutex-group") or "nil"

    definition = f'"{key}" "{description}" ("{short_arg}" "{long_arg}")'
    keywords = f":class {argument_class} :mutex-group {mutex_group}"

    return f"  ({definition}\n   {keywords})"
