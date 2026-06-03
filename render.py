#!/usr/bin/env python3
"""Bake mermaid-ascii diagrams into presentation.md from presentation.src.md.

Each diagram in the source is a normal mermaid-ascii invocation, e.g.:

    ```bash +exec_replace
    mermaid-ascii -x 22 -y 2 -f - <<'MMD'
    graph LR
    ...
    MMD
    ```

This runs that exact command and replaces the block with its rendered (colored
ANSI) output, emitted via `cat` — so presenting needs no external tools, just
`cat`. Per-diagram sizing lives in each block's flags (-x horizontal spacing,
-y vertical spacing, -p border padding). Re-run after editing the source:

    python3 render.py
"""
import re
import subprocess
import pathlib

src = pathlib.Path("presentation.src.md").read_text()
pat = re.compile(r"```bash \+exec_replace\n(mermaid-ascii [^\n]*?) <<'MMD'\n(.*?)\nMMD\n```", re.S)


def repl(m):
    cmd, mmd = m.group(1), m.group(2)
    out = subprocess.run(
        cmd, shell=True, input=mmd, capture_output=True, text=True, check=True
    ).stdout.rstrip("\n")
    return "```bash +exec_replace\ncat <<'ANSIEOF'\n" + out + "\nANSIEOF\n```"


res, n = pat.subn(repl, src)
pathlib.Path("presentation.md").write_text(res)
print(f"baked {n} diagrams into presentation.md")
