"""Render a signal-light dashboard mock-up.

Renders a 3x4 grid of agent cards that visually mirrors what the live
vault dashboard shows: avatar circle, agent name, agent type chip, status
dot (green=LIVE, red=OFFLINE), service/model strip.

All names are deliberately generic so the artifact can sit in a public
pitch deck without leaking internal hostnames or persona names.
"""

from pathlib import Path
from PIL import Image, ImageDraw, ImageFont


PALETTE = {
    "ink": "#102236",
    "muted": "#5b6877",
    "primary": "#c8881a",
    "accent": "#e09a32",
    "paper": "#fffaf2",
    "line": "#ead8b6",
    "success": "#2f9d6b",
    "danger": "#c64a3c",
    "card_bg": "#ffffff",
    "card_alt": "#fff4e0",
}


CARDS = [
    ("운영 머신·비서",         "openclaw",    "live",    "ollama / 로컬"),
    ("AI 주치의",            "claude-code", "live",    "anthropic / claude"),
    ("코드 어시스턴트",         "cline",       "offline", "—"),
    ("윈도우 비서",          "claude-code", "live",    "anthropic / claude"),
    ("경량 노드 비서",         "openclaw",    "live",    "ollama / 로컬"),
    ("보조 노드 챗",          "claude-code", "live",    "anthropic / claude"),
    ("메인 서버 비서",         "openclaw",    "live",    "ollama / 로컬"),
    ("Mac 클라이언트",        "claude-code", "live",    "anthropic / claude"),
    ("나노 에이전트",         "nanoclaw",    "live",    "openrouter / claude"),
    ("친구 에이전트",         "claude-code", "live",    "anthropic / claude"),
    ("외부 분석 봇",          "econoworld",  "live",    "ollama / 로컬"),
    ("실험 에이전트",         "claude-code", "offline", "—"),
]


def load_font(size: int, bold: bool = False) -> ImageFont.FreeTypeFont:
    fonts_dir = Path(__file__).resolve().parents[1] / "fonts"
    primary = fonts_dir / ("Pretendard-Bold.otf" if bold else "Pretendard-Regular.otf")
    if primary.exists():
        try:
            return ImageFont.truetype(str(primary), size)
        except OSError:
            pass
    # CJK-capable system fallbacks (rare on this sandbox but useful elsewhere)
    for path in [
        "/usr/share/fonts/truetype/noto/NotoSansCJK-Bold.ttc" if bold else "/usr/share/fonts/truetype/noto/NotoSansCJK-Regular.ttc",
        "/System/Library/Fonts/Supplemental/AppleSDGothicNeo.ttc",
    ]:
        if Path(path).exists():
            try:
                return ImageFont.truetype(path, size)
            except OSError:
                continue
    return ImageFont.load_default()


def round_rect(draw: ImageDraw.ImageDraw, box, radius, fill=None, outline=None, width=1):
    draw.rounded_rectangle(box, radius=radius, fill=fill, outline=outline, width=width)


def render_card(img: Image.Image, x, y, w, h, name, agent_type, status, model):
    draw = ImageDraw.Draw(img, "RGBA")
    round_rect(draw, [x, y, x + w, y + h], 18, fill=PALETTE["card_bg"], outline=PALETTE["line"], width=2)

    # status dot — top right
    dot_r = 12
    dot_x = x + w - 32
    dot_y = y + 32
    dot_color = PALETTE["success"] if status == "live" else PALETTE["danger"]
    draw.ellipse([dot_x - dot_r, dot_y - dot_r, dot_x + dot_r, dot_y + dot_r], fill=dot_color)
    if status == "live":
        glow = Image.new("RGBA", (60, 60), (0, 0, 0, 0))
        glow_draw = ImageDraw.Draw(glow)
        glow_draw.ellipse([10, 10, 50, 50], fill=(47, 157, 107, 80))
        img.paste(glow, (dot_x - 30, dot_y - 30), glow)

    # avatar circle (lobster gold gradient)
    av_r = 30
    av_x = x + 40
    av_y = y + 40
    draw.ellipse([av_x, av_y, av_x + av_r * 2, av_y + av_r * 2], fill=PALETTE["primary"])
    initial_font = load_font(28, bold=True)
    initials = name[0]
    bbox = draw.textbbox((0, 0), initials, font=initial_font)
    tw = bbox[2] - bbox[0]
    th = bbox[3] - bbox[1]
    draw.text(
        (av_x + av_r - tw // 2, av_y + av_r - th // 2 - bbox[1]),
        initials,
        font=initial_font,
        fill=PALETTE["paper"],
    )

    # name
    name_font = load_font(22, bold=True)
    draw.text((x + 120, y + 38), name, font=name_font, fill=PALETTE["ink"])

    # agent type chip
    chip_font = load_font(15)
    chip_pad_x = 10
    chip_y = y + 70
    chip_bbox = draw.textbbox((0, 0), agent_type, font=chip_font)
    chip_w = chip_bbox[2] - chip_bbox[0] + chip_pad_x * 2
    round_rect(
        draw,
        [x + 120, chip_y, x + 120 + chip_w, chip_y + 24],
        12,
        fill=(200, 136, 26, 38),
        outline=PALETTE["primary"],
        width=1,
    )
    draw.text((x + 120 + chip_pad_x, chip_y + 4), agent_type, font=chip_font, fill=PALETTE["primary"])

    # divider
    draw.line([x + 24, y + h - 56, x + w - 24, y + h - 56], fill=PALETTE["line"], width=1)

    # service / model line
    svc_font = load_font(17)
    draw.text((x + 24, y + h - 42), model, font=svc_font, fill=PALETTE["muted"])


def main():
    repo = Path(__file__).resolve().parents[3]
    out = repo / "docs/promo/images/signal-lights.png"

    width, height = 1800, 1100
    cols, rows = 4, 3
    margin = 60
    gap = 28
    card_w = (width - margin * 2 - gap * (cols - 1)) // cols
    card_h = (height - margin * 2 - 100 - gap * (rows - 1)) // rows

    img = Image.new("RGB", (width, height), PALETTE["paper"])
    draw = ImageDraw.Draw(img, "RGBA")

    title_font = load_font(48, bold=True)
    sub_font = load_font(24)
    draw.text((margin, 30), "벽금고 통합 대시보드", font=title_font, fill=PALETTE["ink"])
    draw.text((margin, 86), "fleet 전 에이전트 신호등 — host 매칭 + agent_type 별 liveness", font=sub_font, fill=PALETTE["muted"])

    for i, (name, agent_type, status, model) in enumerate(CARDS):
        col = i % cols
        row = i // cols
        x = margin + col * (card_w + gap)
        y = margin + 100 + row * (card_h + gap)
        render_card(img, x, y, card_w, card_h, name, agent_type, status, model)

    img.save(out, "PNG", optimize=True)
    print(f"wrote {out}")


if __name__ == "__main__":
    main()
