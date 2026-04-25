"""Extract a 5-colour palette from logo.png and emit CSS variables.

Cluster pixels by RGB, take the most populous bins, and emit primary +
accent + ink + paper + muted variables that theme.css consumes via
:root { --wv-* }. Skipping near-white and near-black so the palette
captures the logo's brand colours, not page chrome.
"""
import sys
from collections import Counter
from pathlib import Path
from PIL import Image


def quantise(rgb, step=24):
    return tuple((c // step) * step for c in rgb)


def luminance(rgb):
    r, g, b = [c / 255 for c in rgb]
    return 0.2126 * r + 0.7152 * g + 0.0722 * b


def extract(image_path: Path, n: int = 5) -> list[tuple[int, int, int]]:
    img = Image.open(image_path).convert("RGBA")
    img = img.resize((200, 200))
    counts = Counter()
    for r, g, b, a in img.getdata():
        if a < 128:
            continue
        lum = luminance((r, g, b))
        if lum < 0.06 or lum > 0.94:
            continue
        counts[quantise((r, g, b))] += 1
    swatches = [c for c, _ in counts.most_common(n * 4)]
    picked: list[tuple[int, int, int]] = []
    for c in swatches:
        if all(sum(abs(a - b) for a, b in zip(c, p)) > 60 for p in picked):
            picked.append(c)
        if len(picked) >= n:
            break
    while len(picked) < n:
        picked.append((128, 128, 128))
    return picked


def hex_(rgb):
    return "#%02x%02x%02x" % rgb


def main():
    repo = Path(__file__).resolve().parents[3]
    logo = repo / "docs/promo/images/logo.png"
    palette = extract(logo)
    palette.sort(key=luminance)
    ink, muted_dark, primary, accent, paper = palette
    out = repo / "docs/promo/palette.css"
    css = f""":root {{
    --wv-ink:     {hex_(ink)};
    --wv-muted:   {hex_(muted_dark)};
    --wv-primary: {hex_(primary)};
    --wv-accent:  {hex_(accent)};
    --wv-paper:   {hex_(paper)};
}}
"""
    out.write_text(css, encoding="utf-8")
    print(f"wrote {out}")
    for name, rgb in zip(["ink", "muted", "primary", "accent", "paper"], palette):
        print(f"  --wv-{name:8} = {hex_(rgb)}")


if __name__ == "__main__":
    sys.exit(main())
