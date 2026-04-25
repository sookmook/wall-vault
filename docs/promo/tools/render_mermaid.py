"""Render Mermaid diagrams to PNG via mermaid.ink (public renderer).

Used as a fallback when local mmdc cannot launch Chromium because system
libs (libasound, libnss, ...) are unavailable. The diagrams shipped in
docs/promo/diagrams/ are deliberately generic — no internal hostnames,
tokens, or persona names — so transit through a public renderer is safe.

mermaid.ink expects pako-compressed + base64url payload at /img/pako:.
We use simple base64url (no pako) for portability, served at /img/.
"""

import base64
import sys
import urllib.parse
import urllib.request
from pathlib import Path

REPO = Path(__file__).resolve().parents[3]
SRC = REPO / "docs/promo/diagrams"
OUT = REPO / "docs/promo/images"


def b64url(s: bytes) -> str:
    return base64.urlsafe_b64encode(s).decode("ascii").rstrip("=")


def render(mmd_path: Path) -> Path:
    body = mmd_path.read_text(encoding="utf-8")
    encoded = b64url(body.encode("utf-8"))
    url = f"https://mermaid.ink/img/{encoded}?type=png&bgColor=fffaf2&width=2200"
    out_path = OUT / (mmd_path.stem + ".png")
    print(f"  GET {url[:80]}...")
    req = urllib.request.Request(url, headers={"User-Agent": "wall-vault-promo/0.1"})
    with urllib.request.urlopen(req, timeout=30) as r:
        if r.status != 200:
            raise RuntimeError(f"HTTP {r.status} for {mmd_path.name}")
        data = r.read()
    out_path.write_bytes(data)
    print(f"  → {out_path} ({len(data)} bytes)")
    return out_path


def main():
    sources = sorted(SRC.glob("*.mmd"))
    if not sources:
        print(f"no .mmd files in {SRC}", file=sys.stderr)
        return 1
    for s in sources:
        print(f"▶ {s.name}")
        try:
            render(s)
        except Exception as e:
            print(f"  ✗ {e}", file=sys.stderr)
            return 2
    print(f"\n✓ {len(sources)} diagrams rendered to {OUT}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
