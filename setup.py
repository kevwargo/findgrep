from setuptools import find_packages, setup

setup(
    name="findgrep",
    packages=find_packages(),
    entry_points={
        "console_scripts": [
            "findgrep = findgrep:run",
        ]
    },
)
