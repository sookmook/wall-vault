"""Render the wall-vault investor pitch deck — 14 slides at 1920×1080.

Every diagram is drawn with PIL primitives instead of Mermaid: the
sandbox lacks both Chromium (mmdc) and any CJK font, and Mermaid output
looks more like an engineering whiteboard than a marketing pitch
anyway. Custom diagrams with branded gradients, big typography, status
badges, and consistent visual language carry the investor narrative
much further.

Pipeline: this script writes 14 PNGs into output/slides/. tools/wrap_pptx.py
then bundles them into a single .pptx as full-bleed backgrounds.
"""

from __future__ import annotations

import math
from dataclasses import dataclass
from pathlib import Path

from PIL import Image, ImageDraw, ImageFilter, ImageFont


REPO = Path(__file__).resolve().parents[3]
PROMO = REPO / "docs/promo"
FONTS = PROMO / "fonts"
IMG = PROMO / "images"
OUT = PROMO / "output/slides"


# ── Brand palette ──────────────────────────────────────────────────────

INK     = (0x10, 0x22, 0x36)
INK_SOFT = (0x1c, 0x33, 0x4d)
INK_DARK = (0x06, 0x14, 0x22)
MUTED   = (0x5b, 0x68, 0x77)
MUTED2  = (0x8b, 0x95, 0xa0)
PRIMARY = (0xc8, 0x88, 0x1a)
PRIMARY_LT = (0xe0, 0xa0, 0x42)
ACCENT  = (0xe0, 0x9a, 0x32)
PAPER   = (0xff, 0xfa, 0xf2)
PAPER2  = (0xff, 0xf4, 0xe0)
PAPER3  = (0xfa, 0xea, 0xc8)
LINE    = (0xea, 0xd8, 0xb6)
LINE_DK = (0xc8, 0xa8, 0x70)
SUCCESS = (0x2f, 0x9d, 0x6b)
SUCCESS_BG = (0xdf, 0xf5, 0xea)
DANGER  = (0xc6, 0x4a, 0x3c)
DANGER_BG  = (0xfb, 0xe2, 0xdd)
INFO    = (0x3a, 0x6f, 0xa6)
INFO_BG = (0xe2, 0xee, 0xf8)


# ── Geometry ──────────────────────────────────────────────────────────

W, H = 1920, 1080
MARGIN_X = 110
MARGIN_Y = 80


def font(weight: str, size: int) -> ImageFont.FreeTypeFont:
    file = {
        "regular":   "Pretendard-Regular.otf",
        "medium":    "Pretendard-Medium.otf",
        "semibold":  "Pretendard-SemiBold.otf",
        "bold":      "Pretendard-Bold.otf",
        "extrabold": "Pretendard-ExtraBold.otf",
    }[weight]
    return ImageFont.truetype(str(FONTS / file), size)


# ── Drawing primitives ────────────────────────────────────────────────


def measure(text: str, fnt: ImageFont.FreeTypeFont, draw: ImageDraw.ImageDraw | None = None) -> tuple[int, int]:
    if draw is None:
        img = Image.new("RGB", (10, 10))
        draw = ImageDraw.Draw(img)
    bbox = draw.textbbox((0, 0), text, font=fnt)
    return bbox[2] - bbox[0], bbox[3] - bbox[1]


def gradient_bg(color_a, color_b, vertical: bool = True) -> Image.Image:
    img = Image.new("RGB", (W, H), color_a)
    draw = ImageDraw.Draw(img)
    if vertical:
        for y in range(H):
            t = y / (H - 1)
            r = int(color_a[0] + (color_b[0] - color_a[0]) * t)
            g = int(color_a[1] + (color_b[1] - color_a[1]) * t)
            b = int(color_a[2] + (color_b[2] - color_a[2]) * t)
            draw.line([(0, y), (W, y)], fill=(r, g, b))
    else:
        for x in range(W):
            t = x / (W - 1)
            r = int(color_a[0] + (color_b[0] - color_a[0]) * t)
            g = int(color_a[1] + (color_b[1] - color_a[1]) * t)
            b = int(color_a[2] + (color_b[2] - color_a[2]) * t)
            draw.line([(x, 0), (x, H)], fill=(r, g, b))
    return img


def radial_glow(center, color_rgba, radius: int, base: Image.Image) -> Image.Image:
    layer = Image.new("RGBA", (W, H), (0, 0, 0, 0))
    d = ImageDraw.Draw(layer)
    cx, cy = center
    d.ellipse([cx - radius, cy - radius, cx + radius, cy + radius], fill=color_rgba)
    layer = layer.filter(ImageFilter.GaussianBlur(radius // 4))
    return Image.alpha_composite(base.convert("RGBA"), layer).convert("RGB")


def soft_shadow_card(canvas: Image.Image, box, *, offset=(0, 12), blur=18, alpha=70) -> Image.Image:
    """Drop shadow for a rounded card — drawn on a separate transparent
    layer so it doesn't blur the rest of the canvas."""
    layer = Image.new("RGBA", (W, H), (0, 0, 0, 0))
    d = ImageDraw.Draw(layer)
    sx, sy = offset
    d.rounded_rectangle(
        [box[0] + sx, box[1] + sy, box[2] + sx, box[3] + sy],
        radius=24, fill=(0, 0, 0, alpha),
    )
    layer = layer.filter(ImageFilter.GaussianBlur(blur))
    return Image.alpha_composite(canvas.convert("RGBA"), layer)


def dot_pattern(canvas: Image.Image, color_rgba, spacing: int = 30, size: int = 2) -> Image.Image:
    """Subtle dotted background pattern — investor decks often use these."""
    layer = Image.new("RGBA", (W, H), (0, 0, 0, 0))
    d = ImageDraw.Draw(layer)
    for x in range(0, W, spacing):
        for y in range(0, H, spacing):
            d.ellipse([x, y, x + size, y + size], fill=color_rgba)
    return Image.alpha_composite(canvas.convert("RGBA"), layer)


def text(draw, pos, txt, fnt, fill, *, spacing=8):
    draw.text(pos, txt, font=fnt, fill=fill, spacing=spacing)


def text_wrapped(draw, pos, txt, fnt, fill, max_w, line_h):
    words = txt.split()
    lines = []
    cur = ""
    for w in words:
        test = (cur + " " + w).strip()
        if measure(test, fnt, draw)[0] <= max_w:
            cur = test
        else:
            if cur:
                lines.append(cur)
            cur = w
    if cur:
        lines.append(cur)
    x, y = pos
    for line in lines:
        draw.text((x, y), line, font=fnt, fill=fill)
        y += line_h
    return y


def fit_image(path: Path, max_w: int, max_h: int) -> Image.Image:
    img = Image.open(path).convert("RGBA")
    img.thumbnail((max_w, max_h), Image.LANCZOS)
    return img


def paste_image(canvas, img, center=None, top_left=None):
    if center is not None:
        x = center[0] - img.width // 2
        y = center[1] - img.height // 2
    else:
        x, y = top_left
    if img.mode == "RGBA":
        canvas.paste(img, (x, y), img)
    else:
        canvas.paste(img, (x, y))
    return (x, y, x + img.width, y + img.height)


def page_chrome(canvas, page, total):
    draw = ImageDraw.Draw(canvas)
    y = H - 50
    draw.line([(MARGIN_X, y), (W - MARGIN_X, y)], fill=LINE, width=1)
    f = font("medium", 18)
    text(draw, (MARGIN_X, y + 14), "wall-vault · 투자·제휴 피칭 · 2026", f, MUTED)
    page_str = f"{page:02d} / {total:02d}"
    pw, _ = measure(page_str, f, draw)
    text(draw, (W - MARGIN_X - pw, y + 14), page_str, f, MUTED)


def watermark_logo(canvas, logo, *, opacity=24, size=140, pos="tr"):
    wm = logo.copy()
    wm.thumbnail((size, size), Image.LANCZOS)
    alpha = wm.split()[-1]
    alpha = alpha.point(lambda v: min(v, opacity))
    wm.putalpha(alpha)
    if pos == "tr":
        x, y = W - size - 60, 60
    else:
        x, y = 60, 60
    canvas.paste(wm, (x, y), wm)


def eyebrow_title(draw, eyebrow, title, *, y=None):
    if y is None:
        y = MARGIN_Y + 30
    eb_font = font("bold", 26)
    text(draw, (MARGIN_X, y), eyebrow, eb_font, PRIMARY)
    _, eh = measure(eyebrow, eb_font, draw)
    title_font = font("extrabold", 64)
    text(draw, (MARGIN_X, y + eh + 12), title, title_font, INK)
    _, th = measure(title, title_font, draw)
    rule_y = y + eh + 12 + th + 26
    draw.line([(MARGIN_X, rule_y), (MARGIN_X + 240, rule_y)], fill=PRIMARY, width=4)
    return rule_y + 30


def big_title(draw, title, *, color=INK, y=None, accent=False):
    if y is None:
        y = MARGIN_Y + 40
    title_font = font("extrabold", 64)
    text(draw, (MARGIN_X, y), title, title_font, color)
    _, th = measure(title, title_font, draw)
    rule_y = y + th + 26
    draw.line([(MARGIN_X, rule_y), (MARGIN_X + 240, rule_y)], fill=PRIMARY, width=4)
    return rule_y + 30


def chip(draw, x, y, label, *, fnt=None, solid=True, fill=None, color=None, h=52, pad_x=24):
    if fnt is None:
        fnt = font("bold", 22)
    if solid:
        bg = fill or PRIMARY
        fg = color or PAPER
    else:
        bg = fill or (255, 244, 224)
        fg = color or PRIMARY
    cw, ch = measure(label, fnt)
    box_w = cw + pad_x * 2
    draw.rounded_rectangle([x, y, x + box_w, y + h], radius=h // 2, fill=bg, outline=PRIMARY if not solid else None, width=2 if not solid else 0)
    text(draw, (x + pad_x, y + (h - ch) // 2 - 4), label, fnt, fg)
    return box_w


def card(draw, box, *, fill=PAPER, outline=LINE, width=2, radius=24):
    draw.rounded_rectangle(box, radius=radius, fill=fill, outline=outline, width=width)


def kpi_card(draw, box, num, lbl, *, num_color=PRIMARY, fill=PAPER):
    card(draw, box, fill=fill)
    nf = font("extrabold", 88)
    nw, nh = measure(num, nf)
    cx = (box[0] + box[2]) // 2
    cy = (box[1] + box[3]) // 2
    text(draw, (cx - nw // 2, cy - nh // 2 - 16), num, nf, num_color)
    lf = font("semibold", 22)
    lw, lh = measure(lbl, lf)
    text(draw, (cx - lw // 2, cy + nh // 2), lbl, lf, MUTED)


def arrow(draw, p1, p2, *, color=MUTED, width=3, head=12):
    draw.line([p1, p2], fill=color, width=width)
    # arrowhead
    dx = p2[0] - p1[0]
    dy = p2[1] - p1[1]
    ang = math.atan2(dy, dx)
    a1 = ang + math.radians(150)
    a2 = ang - math.radians(150)
    h1 = (p2[0] + head * math.cos(a1), p2[1] + head * math.sin(a1))
    h2 = (p2[0] + head * math.cos(a2), p2[1] + head * math.sin(a2))
    draw.polygon([p2, h1, h2], fill=color)


def section_panel(canvas, draw, x, y, w, h, *, title, subtitle="", fill=PAPER):
    """A bordered panel grouping related nodes — used in architecture etc."""
    canvas2 = soft_shadow_card(canvas, [x, y, x + w, y + h], offset=(0, 6), blur=10, alpha=30)
    draw2 = ImageDraw.Draw(canvas2)
    draw2.rounded_rectangle([x, y, x + w, y + h], radius=20, fill=fill, outline=LINE, width=2)
    title_font = font("bold", 22)
    sub_font = font("medium", 18)
    text(draw2, (x + 28, y + 18), title, title_font, MUTED)
    if subtitle:
        text(draw2, (x + 28, y + 44), subtitle, sub_font, MUTED2)
    return canvas2


def node_box(draw, box, *, label, kicker="", body="", fill=PAPER, ink=INK, accent=PRIMARY, radius=14):
    draw.rounded_rectangle(box, radius=radius, fill=fill, outline=LINE, width=2)
    cx = (box[0] + box[2]) // 2
    cy = (box[1] + box[3]) // 2
    if kicker:
        kf = font("bold", 16)
        kw, kh = measure(kicker, kf)
        text(draw, (cx - kw // 2, box[1] + 16), kicker, kf, accent)
    lf = font("bold", 22)
    lw, lh = measure(label, lf)
    if kicker:
        ly = box[1] + 16 + 22
    else:
        ly = cy - lh // 2 - (10 if body else 0)
    text(draw, (cx - lw // 2, ly), label, lf, ink)
    if body:
        bf = font("regular", 16)
        for i, ln in enumerate(body.split("\n")):
            bw, bh = measure(ln, bf)
            text(draw, (cx - bw // 2, ly + lh + 8 + i * 22), ln, bf, MUTED)


# ── Slide builders ────────────────────────────────────────────────────


def s01_title(logo):
    bg = gradient_bg(PAPER, PAPER2)
    bg = radial_glow((W // 2, int(H * 0.40)), (200, 136, 26, 80), 850, bg)
    bg = radial_glow((int(W * 0.85), int(H * 0.85)), (224, 154, 50, 50), 500, bg)
    bg = dot_pattern(bg, (200, 136, 26, 14), spacing=42, size=2)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)

    big_logo = logo.copy()
    big_logo.thumbnail((680, 680), Image.LANCZOS)
    paste_image(canvas, big_logo, center=(W // 2, int(H * 0.38)))

    sub_font = font("extrabold", 64)
    line1 = "AI 가 절대 끊기지 않는"
    line2 = "멀티벤더 게이트웨이"
    w1, h1 = measure(line1, sub_font, draw)
    w2, h2 = measure(line2, sub_font, draw)
    head_y = int(H * 0.72)
    text(draw, ((W - w1) // 2, head_y), line1, sub_font, INK)
    text(draw, ((W - w2) // 2, head_y + h1 + 6), line2, sub_font, PRIMARY)

    rule_y = head_y + h1 + h2 + 28
    draw.line([((W - 280) // 2, rule_y), ((W + 280) // 2, rule_y)], fill=PRIMARY, width=3)

    tag_font = font("medium", 26)
    tag = "키 금고  ·  지능형 라우팅  ·  fleet observability  —  한 바이너리에"
    tw, _ = measure(tag, tag_font, draw)
    text(draw, ((W - tw) // 2, rule_y + 26), tag, tag_font, MUTED)

    # ornamental corners
    for cx, cy in [(80, 80), (W - 80, 80), (80, H - 80), (W - 80, H - 80)]:
        draw.rectangle([cx - 30, cy - 1, cx + 30, cy + 1], fill=PRIMARY)
        draw.rectangle([cx - 1, cy - 30, cx + 1, cy + 30], fill=PRIMARY)

    return canvas.convert("RGB")


def s02_section(num, label, line1, line2):
    bg = gradient_bg(INK, INK_DARK)
    bg = radial_glow((int(W * 0.20), int(H * 0.50)), (200, 136, 26, 60), 750, bg)
    bg = radial_glow((int(W * 0.80), int(H * 0.50)), (224, 154, 50, 45), 600, bg)
    bg = dot_pattern(bg, (200, 136, 26, 16), spacing=44, size=2)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)

    label_font = font("bold", 28)
    label_text = f"PART · {num}"
    lw, _ = measure(label_text, label_font, draw)
    text(draw, ((W - lw) // 2, int(H * 0.22)), label_text, label_font, PRIMARY)

    sub_font = font("semibold", 36)
    sw, sh = measure(label, sub_font, draw)
    text(draw, ((W - sw) // 2, int(H * 0.28)), label, sub_font, PAPER2)

    title_font = font("extrabold", 116)
    tw1, th1 = measure(line1, title_font, draw)
    tw2, th2 = measure(line2, title_font, draw)
    title_y = int(H * 0.42)
    text(draw, ((W - tw1) // 2, title_y), line1, title_font, PAPER)
    text(draw, ((W - tw2) // 2, title_y + th1 + 8), line2, title_font, ACCENT)

    rule_y = title_y + th1 + th2 + 60
    draw.line([((W - 200) // 2, rule_y), ((W + 200) // 2, rule_y)], fill=PRIMARY, width=3)

    return canvas.convert("RGB")


def s03_problem(logo):
    bg = gradient_bg(PAPER, PAPER2)
    bg = dot_pattern(bg, (200, 136, 26, 12), spacing=38, size=2)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)
    watermark_logo(canvas, logo)

    head_font = font("extrabold", 60)
    head1 = "AI 가 끊기는 순간, "
    head2 = "비용은 두 배가 됩니다"
    h1w, h1h = measure(head1, head_font, draw)
    head_y = MARGIN_Y + 20
    text(draw, (MARGIN_X, head_y), head1, head_font, INK)
    text(draw, (MARGIN_X + h1w, head_y), head2, head_font, PRIMARY)
    rule_y = head_y + h1h + 24
    draw.line([(MARGIN_X, rule_y), (MARGIN_X + 240, rule_y)], fill=PRIMARY, width=4)

    sub_font = font("medium", 24)
    text(draw, (MARGIN_X, rule_y + 20), "현장에서 매일 발생하는 다섯 가지 단절 시나리오", sub_font, MUTED)

    # 5 problem cards in a 2x3 layout (last cell becomes a quote)
    card_w = 540
    card_h = 220
    gap = 24
    items = [
        ("01", "키 만료 · 쿨다운", "단일 클라우드 키가 만료되면\nfleet 전체가 즉시 정지합니다."),
        ("02", "벤더별 모델 ID 차이", "Anthropic / OpenAI / Google /\nOpenRouter 모두 다른 네임스페이스."),
        ("03", "관측 가시성 0", "여러 머신·여러 에이전트가\n어디서 멈췄는지 안 보입니다."),
        ("04", "조용한 모델 치환", "Fallback 이 다른 모델로\n바꿔치기 — 응답 품질 붕괴."),
        ("05", "운영자 묶임", "매일 \"왜 또 멈췄지?\"\n장애 대응에 사람이 묶입니다."),
    ]
    grid_y = rule_y + 70
    for i, (num, ttl, desc) in enumerate(items):
        col = i % 3
        row = i // 3
        x = MARGIN_X + col * (card_w + gap)
        y = grid_y + row * (card_h + gap)
        canvas = soft_shadow_card(canvas, [x, y, x + card_w, y + card_h], offset=(0, 10), blur=14, alpha=44)
        draw = ImageDraw.Draw(canvas)
        draw.rounded_rectangle([x, y, x + card_w, y + card_h], radius=22, fill=PAPER, outline=LINE, width=2)
        # left accent stripe
        draw.rounded_rectangle([x, y, x + 8, y + card_h], radius=4, fill=PRIMARY)
        # number badge
        nf = font("extrabold", 34)
        text(draw, (x + 32, y + 24), num, nf, PRIMARY_LT)
        # title
        tf = font("bold", 26)
        text(draw, (x + 32, y + 66), ttl, tf, INK)
        # body
        bf = font("regular", 20)
        text(draw, (x + 32, y + 110), desc, bf, MUTED)

    # quote card in the 6th cell
    x = MARGIN_X + 2 * (card_w + gap)
    y = grid_y + 1 * (card_h + gap)
    qbox = [x, y, x + card_w, y + card_h]
    canvas = soft_shadow_card(canvas, qbox, offset=(0, 10), blur=14, alpha=44)
    draw = ImageDraw.Draw(canvas)
    draw.rounded_rectangle(qbox, radius=22, fill=INK, outline=PRIMARY, width=2)
    qf = font("extrabold", 78)
    text(draw, (x + 28, y + 8), "“", qf, PRIMARY)
    body_f = font("semibold", 22)
    body = "키 하나가\n전체를 정지시킵니다.\n그게 어디서 멈췄는지조차\n보이지 않습니다."
    text(draw, (x + 32, y + 80), body, body_f, PAPER, spacing=10)

    return canvas.convert("RGB")


def s04_definition(logo):
    bg = gradient_bg(PAPER, PAPER2)
    bg = radial_glow((W // 2, int(H * 0.45)), (200, 136, 26, 50), 700, bg)
    bg = dot_pattern(bg, (200, 136, 26, 12), spacing=38, size=2)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)
    watermark_logo(canvas, logo)

    label_font = font("semibold", 26)
    label = "wall-vault · 한 줄 정의"
    lw, _ = measure(label, label_font, draw)
    text(draw, ((W - lw) // 2, MARGIN_Y + 20), label, label_font, MUTED)

    # headline — horizontally centered, two anchored phrases joined by gold +
    h_font = font("extrabold", 76)
    plus_font = font("extrabold", 60)
    part_a = "AES-GCM 암호화 키 금고"
    plus = "  +  "
    part_b = "지능형 멀티벤더 프록시"
    aw, ah = measure(part_a, h_font, draw)
    pw, ph = measure(plus, plus_font, draw)
    bw, bh = measure(part_b, h_font, draw)
    total = aw + pw + bw
    y = int(H * 0.18)
    x = (W - total) // 2
    text(draw, (x, y), part_a, h_font, INK)
    text(draw, (x + aw, y + 14), plus, plus_font, PRIMARY)
    text(draw, (x + aw + pw, y), part_b, h_font, INK)
    draw.line([(x, y + ah + 14), (x + aw, y + ah + 14)], fill=PRIMARY, width=4)
    draw.line([(x + aw + pw, y + ah + 14), (x + aw + pw + bw, y + ah + 14)], fill=PRIMARY, width=4)

    # 4 feature cards row
    feat_y = int(H * 0.40)
    feat_h = 280
    feat_gap = 24
    feat_w = (W - MARGIN_X * 2 - feat_gap * 3) // 4
    features = [
        ("암호화 저장",  "AES-GCM",     "마스터 키로 키 자체를 암호화 저장.\n탈취돼도 평문 노출 0."),
        ("자동 로테이션","Round-Robin", "쿨다운 자동 감지 + 다음 키로\n무중단 전환."),
        ("정직한 라우팅","Strict default","모델 무단 치환 없음.\nfallback 헤더로 가시성 100%."),
        ("실시간 신호등","SSE 동기화",  "fleet 전 에이전트 상태를\n한 화면에서 즉시 확인."),
    ]
    for i, (ttl, kicker, body) in enumerate(features):
        x = MARGIN_X + i * (feat_w + feat_gap)
        y = feat_y
        canvas = soft_shadow_card(canvas, [x, y, x + feat_w, y + feat_h], offset=(0, 10), blur=14, alpha=44)
        draw = ImageDraw.Draw(canvas)
        draw.rounded_rectangle([x, y, x + feat_w, y + feat_h], radius=22, fill=PAPER, outline=LINE, width=2)
        # number circle
        cir_r = 28
        draw.ellipse([x + 28, y + 28, x + 28 + cir_r * 2, y + 28 + cir_r * 2], fill=PRIMARY)
        nf = font("extrabold", 28)
        nstr = f"{i+1:02d}"
        nw, nh = measure(nstr, nf)
        text(draw, (x + 28 + cir_r - nw // 2, y + 28 + cir_r - nh // 2 - 4), nstr, nf, PAPER)
        # title
        tf = font("extrabold", 28)
        text(draw, (x + 28, y + 110), ttl, tf, INK)
        # kicker
        kf = font("bold", 18)
        text(draw, (x + 28, y + 148), kicker, kf, PRIMARY_LT)
        # body
        bf = font("regular", 19)
        text(draw, (x + 28, y + 184), body, bf, MUTED, spacing=4)

    # bottom platform strip
    strip_y = feat_y + feat_h + 60
    strip_h = 100
    strip_box = [MARGIN_X, strip_y, W - MARGIN_X, strip_y + strip_h]
    draw.rounded_rectangle(strip_box, radius=22, fill=INK)
    f1 = font("medium", 22)
    f2 = font("bold", 26)
    text(draw, (MARGIN_X + 40, strip_y + 22), "한 개의 Go 바이너리", f2, PRIMARY)
    text(draw, (MARGIN_X + 40, strip_y + 56), "무설치 의존성 · 설치 1 분", f1, PAPER)
    chips = ["Linux", "macOS", "Windows", "WSL", "ARM SBC"]
    cx = W - MARGIN_X - 40
    chip_y = strip_y + (strip_h - 40) // 2
    for c in reversed(chips):
        cf = font("bold", 18)
        cw, ch = measure(c, cf)
        cw_total = cw + 32
        cx -= cw_total
        draw.rounded_rectangle([cx, chip_y, cx + cw_total, chip_y + 40], radius=20, outline=PRIMARY, width=2)
        text(draw, (cx + 16, chip_y + (40 - ch) // 2 - 4), c, cf, PRIMARY_LT)
        cx -= 12

    return canvas.convert("RGB")


def s05_architecture(logo):
    """Custom architecture diagram — replaces the Mermaid render."""
    bg = gradient_bg(PAPER, PAPER2)
    bg = dot_pattern(bg, (200, 136, 26, 10), spacing=42, size=2)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)
    watermark_logo(canvas, logo)

    big_title(draw, "아키텍처 — 한 장으로 보는 동작")

    # 3 tiers laid out vertically: clients (top), gateway (middle), upstream (bottom)
    # Actually for visual horizontal flow: clients (left) → gateway (center) → upstream (right)
    # Use a wide horizontal layout with tier panels.

    tier_top = 240
    tier_h = 580
    panel_color = (255, 245, 220)

    # ── left tier: clients ──
    cx, cy, cw = MARGIN_X, tier_top, 380
    canvas = section_panel(canvas, draw, cx, cy, cw, tier_h, title="클라이언트", subtitle="모든 형식 그대로 호출", fill=panel_color)
    draw = ImageDraw.Draw(canvas)
    cli_items = [
        ("OpenClaw",    "Gemini API"),
        ("Claude Code", "Anthropic API"),
        ("Cline · Cursor","OpenAI API"),
        ("커스텀 앱",    "OpenRouter API"),
    ]
    item_y = cy + 100
    item_h = 96
    for i, (label, sub) in enumerate(cli_items):
        bx = cx + 28
        by = item_y + i * (item_h + 16)
        bw_local = cw - 56
        draw.rounded_rectangle([bx, by, bx + bw_local, by + item_h], radius=14, fill=PAPER, outline=LINE, width=2)
        # left accent
        draw.rounded_rectangle([bx, by, bx + 6, by + item_h], radius=3, fill=PRIMARY)
        text(draw, (bx + 24, by + 18), label, font("bold", 24), INK)
        text(draw, (bx + 24, by + 52), sub, font("regular", 18), MUTED)

    # ── center tier: wall-vault gateway ──
    gx = cx + cw + 80
    gw = 400
    canvas = section_panel(canvas, draw, gx, cy, gw, tier_h, title="wall-vault 게이트웨이", subtitle="단일 진입점 · 단일 진실", fill=(16, 34, 54))
    draw = ImageDraw.Draw(canvas)

    # Two stacked dark cards: proxy + vault
    inner_x = gx + 24
    inner_w = gw - 48
    # proxy card
    px, py, ph = inner_x, cy + 100, 220
    draw.rounded_rectangle([px, py, px + inner_w, py + ph], radius=16, fill=INK_SOFT, outline=PRIMARY, width=2)
    text(draw, (px + 24, py + 18), "프록시", font("bold", 18), PRIMARY)
    text(draw, (px + 24, py + 44), ":56244", font("medium", 16), MUTED2)
    text(draw, (px + 24, py + 80), "지능형 라우팅", font("extrabold", 26), PAPER)
    text(draw, (px + 24, py + 120), "Strict-by-default", font("semibold", 18), PRIMARY_LT)
    text(draw, (px + 24, py + 150), "fallback 헤더 가시성", font("regular", 16), MUTED2)
    text(draw, (px + 24, py + 178), "OAI · Anthropic · Gemini 동시 노출", font("regular", 16), MUTED2)

    # vault card
    vy = py + ph + 20
    vh = 220
    draw.rounded_rectangle([px, vy, px + inner_w, vy + vh], radius=16, fill=INK_SOFT, outline=PRIMARY, width=2)
    text(draw, (px + 24, vy + 18), "벽금고", font("bold", 18), PRIMARY)
    text(draw, (px + 24, vy + 44), ":56243", font("medium", 16), MUTED2)
    text(draw, (px + 24, vy + 80), "AES-GCM 키 금고", font("extrabold", 26), PAPER)
    text(draw, (px + 24, vy + 120), "SSE 실시간 브로드캐스트", font("semibold", 18), PRIMARY_LT)
    text(draw, (px + 24, vy + 150), "키 로테이션 · 자동 쿨다운", font("regular", 16), MUTED2)
    text(draw, (px + 24, vy + 178), "fleet 신호등 통합", font("regular", 16), MUTED2)

    # SSE connector between proxy and vault
    sse_x1 = px + inner_w // 2
    sse_y1 = py + ph
    sse_y2 = vy
    draw.line([(sse_x1, sse_y1), (sse_x1, sse_y2)], fill=PRIMARY, width=3)
    sse_lf = font("bold", 14)
    text(draw, (sse_x1 + 8, (sse_y1 + sse_y2) // 2 - 8), "SSE", sse_lf, PRIMARY_LT)

    # ── right tier: upstream ──
    ux = gx + gw + 80
    uw = W - MARGIN_X - ux
    canvas = section_panel(canvas, draw, ux, cy, uw, tier_h, title="업스트림 (멀티 벤더)", subtitle="170+ 모델 자동 라우팅", fill=panel_color)
    draw = ImageDraw.Draw(canvas)
    up_items = [
        ("Anthropic",     "claude-opus / sonnet"),
        ("Google Gemini", "gemini-pro / flash"),
        ("OpenAI",        "gpt-4 / gpt-5 family"),
        ("OpenRouter",    "340+ 모델 게이트"),
        ("Ollama 로컬",    "qwen3 · gemma · llama"),
    ]
    item_h = 88
    item_y = cy + 100
    for i, (label, sub) in enumerate(up_items):
        bx = ux + 28
        by = item_y + i * (item_h + 8)
        bw_local = uw - 56
        draw.rounded_rectangle([bx, by, bx + bw_local, by + item_h], radius=14, fill=PAPER, outline=LINE, width=2)
        draw.rounded_rectangle([bx + bw_local - 6, by, bx + bw_local, by + item_h], radius=3, fill=PRIMARY)
        text(draw, (bx + 24, by + 14), label, font("bold", 22), INK)
        text(draw, (bx + 24, by + 46), sub, font("regular", 16), MUTED)

    # ── arrows between tiers ──
    arrow_color = PRIMARY
    arrow(draw, (cx + cw + 12, cy + tier_h // 2), (gx - 12, cy + tier_h // 2), color=arrow_color, width=4, head=18)
    arrow(draw, (gx + gw + 12, cy + tier_h // 2), (ux - 12, cy + tier_h // 2), color=arrow_color, width=4, head=18)

    # ── bottom KPI strip ──
    strip_y = tier_top + tier_h + 32
    strip_h = 80
    strip_box = [MARGIN_X, strip_y, W - MARGIN_X, strip_y + strip_h]
    draw.rounded_rectangle(strip_box, radius=20, fill=INK)
    metrics = [("4", "API 포맷"), ("170+", "모델"), ("17", "로케일"), ("4 머신+", "fleet 운영"), ("AES-GCM", "키 금고")]
    seg_w = (W - MARGIN_X * 2) // len(metrics)
    for i, (n, l) in enumerate(metrics):
        seg_x = MARGIN_X + i * seg_w
        if i > 0:
            draw.line([(seg_x, strip_y + 16), (seg_x, strip_y + strip_h - 16)], fill=MUTED, width=1)
        nf = font("extrabold", 32)
        nw, nh = measure(n, nf)
        text(draw, (seg_x + (seg_w - nw) // 2, strip_y + 12), n, nf, PRIMARY)
        lf = font("medium", 16)
        lw, lh = measure(l, lf)
        text(draw, (seg_x + (seg_w - lw) // 2, strip_y + 50), l, lf, PAPER2)

    return canvas.convert("RGB")


def s07_diff_multivendor(logo):
    bg = gradient_bg(PAPER, PAPER2)
    bg = dot_pattern(bg, (200, 136, 26, 10), spacing=42, size=2)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)
    watermark_logo(canvas, logo)

    eyebrow_title(draw, "차별점 ①", "멀티벤더를 진짜로 통합")

    # left: bullets
    bullets = [
        ("4 가지 API 포맷 동시 노출",  "Gemini · OpenAI · Anthropic · OpenRouter — 한 게이트웨이가 모두 받습니다"),
        ("170+ 모델 자동 라우팅",      "로컬 Ollama 부터 Claude / Gemini / GPT 까지 — provider/model 형식만으로 도착"),
        ("코드 변경 0 의 클라이언트 통합", "Cline · Cursor · Claude Code · OpenClaw — 자기 형식 그대로 호출"),
        ("벤더 SDK 4 개 → 게이트웨이 1 개", "팀 학습 곡선 ↓  운영 비용 ↓  벤더 락-인 ↓"),
    ]
    bul_font = font("semibold", 30)
    desc_font = font("regular", 21)
    y = 320
    for primary_txt, desc_txt in bullets:
        # numbered dot
        draw.ellipse([MARGIN_X, y + 12, MARGIN_X + 16, y + 28], fill=PRIMARY)
        text(draw, (MARGIN_X + 38, y), primary_txt, bul_font, INK)
        _, bh = measure(primary_txt, bul_font, draw)
        text(draw, (MARGIN_X + 38, y + bh + 6), desc_txt, desc_font, MUTED)
        y += bh + 70

    # right: KPI grid 2x2 + supporting band
    grid_x = 1180
    grid_y = 290
    cell_w = 260
    cell_h = 230
    gap = 24
    kpis = [("4", "API 포맷", PRIMARY),
            ("170+", "모델", ACCENT),
            ("10+", "에이전트 타입", INK),
            ("17", "로케일", SUCCESS)]
    for i, (num, lbl, col) in enumerate(kpis):
        c = i % 2
        r = i // 2
        x = grid_x + c * (cell_w + gap)
        yy = grid_y + r * (cell_h + gap)
        canvas = soft_shadow_card(canvas, [x, yy, x + cell_w, yy + cell_h], offset=(0, 10), blur=14, alpha=44)
        draw = ImageDraw.Draw(canvas)
        draw.rounded_rectangle([x, yy, x + cell_w, yy + cell_h], radius=22, fill=PAPER, outline=LINE, width=2)
        # accent stripe top
        draw.rounded_rectangle([x, yy, x + cell_w, yy + 8], radius=4, fill=col)
        nf = font("extrabold", 96)
        nw, nh = measure(num, nf)
        text(draw, (x + (cell_w - nw) // 2, yy + 30), num, nf, col)
        lf = font("semibold", 22)
        lw, _ = measure(lbl, lf)
        text(draw, (x + (cell_w - lw) // 2, yy + cell_h - 56), lbl, lf, MUTED)

    # supporting band — vendor logos as text chips
    band_y = grid_y + 2 * (cell_h + gap) + 14
    band_box = [grid_x, band_y, grid_x + 2 * cell_w + gap, band_y + 90]
    draw.rounded_rectangle(band_box, radius=20, fill=INK)
    text(draw, (grid_x + 24, band_y + 18), "지원 벤더", font("bold", 18), PRIMARY)
    cx_band = grid_x + 24
    cy_band = band_y + 50
    vendors = ["Anthropic", "Google", "OpenAI", "OpenRouter", "Ollama"]
    for v in vendors:
        cf = font("bold", 16)
        vw, vh = measure(v, cf)
        bw_v = vw + 24
        draw.rounded_rectangle([cx_band, cy_band, cx_band + bw_v, cy_band + 28], radius=14, outline=PRIMARY, width=1)
        text(draw, (cx_band + 12, cy_band + 4), v, cf, PRIMARY_LT)
        cx_band += bw_v + 8

    return canvas.convert("RGB")


def s08_diff_routing(logo):
    """Horizontal dispatch flow — pipeline rendered left-to-right with
    a bottom triplet of outcomes, plus a right-side narrative + config
    panel. Spreads the visual load instead of stacking everything."""
    bg = gradient_bg(PAPER, PAPER2)
    bg = dot_pattern(bg, (200, 136, 26, 10), spacing=42, size=2)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)
    watermark_logo(canvas, logo)

    eyebrow_title(draw, "차별점 ②", "정직한 라우팅 (Strict-by-default)")

    # ── horizontal pipeline: 요청 → 토큰 조회 → primary 시도 ──
    pipe_y = 290
    pipe_h = 110
    n_count = 3
    pipe_avail = 1100
    pipe_gap = 32
    pipe_w = (pipe_avail - pipe_gap * (n_count - 1)) // n_count
    pipe_x0 = MARGIN_X
    pipeline = [
        ("요청 도착",      "Bearer 토큰 + body.model",       INK,   PRIMARY),
        ("벽금고 조회",     "preferred_service · model_override", PAPER, INK),
        ("Primary 호출",   "preferred_service 그대로 시도",     PAPER, INK),
    ]
    pipe_centers = []
    for i, (title, sub, fill, ink) in enumerate(pipeline):
        x = pipe_x0 + i * (pipe_w + pipe_gap)
        canvas = soft_shadow_card(canvas, [x, pipe_y, x + pipe_w, pipe_y + pipe_h], offset=(0, 8), blur=12, alpha=40)
        draw = ImageDraw.Draw(canvas)
        outline = PRIMARY if fill == INK else LINE
        draw.rounded_rectangle([x, pipe_y, x + pipe_w, pipe_y + pipe_h], radius=18, fill=fill, outline=outline, width=2)
        if fill == INK:
            tcolor, scolor = PRIMARY, PAPER2
        else:
            tcolor, scolor = ink, MUTED
        text(draw, (x + 24, pipe_y + 22), title, font("bold", 26), tcolor)
        text(draw, (x + 24, pipe_y + 60), sub, font("regular", 19), scolor)
        pipe_centers.append((x + pipe_w, pipe_y + pipe_h // 2))
        if i < n_count - 1:
            arrow(draw,
                  (x + pipe_w + 4, pipe_y + pipe_h // 2),
                  (x + pipe_w + pipe_gap - 4, pipe_y + pipe_h // 2),
                  color=PRIMARY, width=3, head=12)

    # split arrow from last pipeline node downwards into 3 outcomes
    split_origin = (pipe_x0 + (n_count - 1) * (pipe_w + pipe_gap) + pipe_w // 2, pipe_y + pipe_h)
    split_y = split_origin[1] + 60
    draw.line([split_origin, (split_origin[0], split_y)], fill=PRIMARY, width=3)

    # ── 3 outcomes side by side ──
    out_y = split_y + 30
    out_h = 240
    out_count = 3
    out_avail = pipe_avail
    out_gap = 32
    out_w = (out_avail - out_gap * (out_count - 1)) // out_count
    outs = [
        {
            "title":   "200 OK",
            "kicker":  "primary 성공",
            "body":    ["X-WV-Used-Service", "X-WV-Used-Model", "(Fallback-Reason 없음)"],
            "fill":    SUCCESS_BG,
            "outline": SUCCESS,
            "title_color": SUCCESS,
        },
        {
            "title":   "502",
            "kicker":  "strict — 기본 동작",
            "body":    ["primary 실패 = 즉시 502", "모델 무단 치환 없음", "(원인 그대로 반환)"],
            "fill":    DANGER_BG,
            "outline": DANGER,
            "title_color": DANGER,
        },
        {
            "title":   "Fallback",
            "kicker":  "opt-in (FallbackServices)",
            "body":    ["체인 순서대로 시도", "X-WV-Fallback-Reason 노출", "조용한 치환 ✗"],
            "fill":    PAPER2,
            "outline": PRIMARY,
            "title_color": PRIMARY,
        },
    ]
    for i, o in enumerate(outs):
        x = pipe_x0 + i * (out_w + out_gap)
        canvas = soft_shadow_card(canvas, [x, out_y, x + out_w, out_y + out_h], offset=(0, 10), blur=14, alpha=50)
        draw = ImageDraw.Draw(canvas)
        draw.rounded_rectangle([x, out_y, x + out_w, out_y + out_h], radius=20, fill=o["fill"], outline=o["outline"], width=2)
        # connector arrow from split_origin to top of this card
        arrow(draw, (split_origin[0], split_y), (x + out_w // 2, out_y - 4), color=o["outline"], width=3, head=12)
        text(draw, (x + 28, out_y + 22), o["title"], font("extrabold", 40), o["title_color"])
        text(draw, (x + 28, out_y + 80), o["kicker"], font("bold", 18), MUTED)
        for j, line in enumerate(o["body"]):
            text(draw, (x + 28, out_y + 122 + j * 30), "·  " + line, font("medium", 18), INK)

    # ── right side: narrative + config snippet ──
    rx = pipe_x0 + pipe_avail + 50
    rw = W - MARGIN_X - rx
    bullets = [
        ("Primary 실패 ≠ 모델 치환",
         "기본은 502 — 운영자만 명시적으로 fallback 허용"),
        ("응답 헤더로 라우팅 가시성",
         "X-WV-Used-Service · Used-Model · Fallback-Reason"),
        ("호출자 한 줄 헤더 검증",
         "운영 사고를 응답 시점에 즉시 포착"),
    ]
    by = pipe_y - 10
    for primary_txt, desc_txt in bullets:
        draw.ellipse([rx, by + 14, rx + 16, by + 30], fill=PRIMARY)
        text(draw, (rx + 28, by), primary_txt, font("bold", 22), INK)
        _, bh = measure(primary_txt, font("bold", 22), draw)
        text_wrapped(draw, (rx + 28, by + bh + 4), desc_txt, font("regular", 17), MUTED, rw - 30, 22)
        by += 100

    # config panel
    code_y = by + 6
    code_h = (out_y + out_h) - code_y
    draw.rounded_rectangle([rx, code_y, rx + rw, code_y + code_h], radius=18, fill=INK_DARK, outline=PRIMARY, width=2)
    text(draw, (rx + 22, code_y + 18), "vault client config", font("bold", 18), PRIMARY_LT)
    cf_mono = font("medium", 16)
    body = (
        "{\n"
        '  "preferred_service":\n'
        '    "anthropic",\n'
        '  "model_override":\n'
        '    "claude-opus-4-7",\n'
        '  "fallback_services": []\n'
        "  // ← strict (default)\n"
        "}\n\n"
        "// fallback 원할 때만:\n"
        '"fallback_services": [\n'
        '  "openrouter",\n'
        '  "ollama"\n'
        "]"
    )
    text(draw, (rx + 22, code_y + 50), body, cf_mono, PAPER, spacing=2)

    return canvas.convert("RGB")


def s09_diff_observability(mock, logo):
    bg = gradient_bg(PAPER, PAPER2)
    bg = dot_pattern(bg, (200, 136, 26, 10), spacing=42, size=2)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)
    watermark_logo(canvas, logo)

    eyebrow_title(draw, "차별점 ③", "Fleet observability — 한 곳에서 다 봅니다")

    img = mock.copy()
    avail_w = W - MARGIN_X * 2
    img.thumbnail((avail_w, 600), Image.LANCZOS)
    paste_image(canvas, img, center=(W // 2, 600))

    # bottom row of insights
    box_y = 920
    box_h = 90
    cols = 3
    gap = 24
    box_w = (W - MARGIN_X * 2 - gap * (cols - 1)) // cols
    insights = [
        ("실시간",  "SSE 1–3초 내 동기화"),
        ("정확함",  "host 매칭 + 타입별 liveness"),
        ("복원력",  "프록시·벤더 죽어도 신호등 유지"),
    ]
    for i, (k, v) in enumerate(insights):
        x = MARGIN_X + i * (box_w + gap)
        draw.rounded_rectangle([x, box_y, x + box_w, box_y + box_h], radius=18, fill=INK)
        text(draw, (x + 24, box_y + 16), k, font("extrabold", 24), PRIMARY)
        text(draw, (x + 24, box_y + 52), v, font("medium", 18), PAPER2)

    return canvas.convert("RGB")


def s10_proof(logo):
    """Custom fleet topology — replaces Mermaid render."""
    bg = gradient_bg(PAPER, PAPER2)
    bg = dot_pattern(bg, (200, 136, 26, 10), spacing=42, size=2)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)
    watermark_logo(canvas, logo)

    big_title(draw, "실측 운영 — 끊김 없는 fleet")

    # central vault on bottom-center, 4 host nodes on top forming a fan
    cy = 480
    central_w = 480
    central_h = 140
    cx = (W - central_w) // 2
    central_box = [cx, cy + 240, cx + central_w, cy + 240 + central_h]
    canvas = soft_shadow_card(canvas, central_box, offset=(0, 12), blur=18, alpha=70)
    draw = ImageDraw.Draw(canvas)
    draw.rounded_rectangle(central_box, radius=22, fill=INK, outline=PRIMARY, width=3)
    text(draw, (cx + 36, cy + 250), "벽금고 :56243", font("extrabold", 28), PRIMARY)
    text(draw, (cx + 36, cy + 290), "중앙 키·설정·신호등 · SSE 브로드캐스트", font("medium", 20), PAPER2)
    # central indicator
    draw.ellipse([cx + central_w - 60, cy + 264, cx + central_w - 36, cy + 288], fill=SUCCESS)

    # 4 hosts on top
    host_w = 320
    host_h = 200
    host_y = cy
    total_w = 4 * host_w + 3 * 32
    start_x = (W - total_w) // 2
    hosts = [
        ("워크 머신",     "WSL · Linux",    "여러 에이전트 호스팅", SUCCESS),
        ("메인 서버",     "macOS · GPU",    "벽금고 + 로컬 추론",   SUCCESS),
        ("저전력 노드",   "ARM SBC",        "경량 에이전트",         SUCCESS),
        ("보조 노드",     "Linux · 전용",    "특화 봇",               SUCCESS),
    ]
    host_centers = []
    for i, (name, os_lbl, role, status) in enumerate(hosts):
        hx = start_x + i * (host_w + 32)
        hbox = [hx, host_y, hx + host_w, host_y + host_h]
        canvas = soft_shadow_card(canvas, hbox, offset=(0, 8), blur=12, alpha=40)
        draw = ImageDraw.Draw(canvas)
        draw.rounded_rectangle(hbox, radius=22, fill=PAPER, outline=LINE, width=2)
        # top accent
        draw.rounded_rectangle([hx, host_y, hx + host_w, host_y + 8], radius=4, fill=PRIMARY)
        # status dot
        draw.ellipse([hx + host_w - 40, host_y + 24, hx + host_w - 20, host_y + 44], fill=status)
        text(draw, (hx + 24, host_y + 28), name, font("extrabold", 26), INK)
        text(draw, (hx + 24, host_y + 66), os_lbl, font("bold", 18), PRIMARY_LT)
        text(draw, (hx + 24, host_y + 96), role, font("regular", 17), MUTED)
        # mini chips
        chips = ["openclaw", "claude-code"] if i in (0, 1) else ["openclaw"] if i == 2 else ["nanoclaw"]
        cx_chip = hx + 24
        cy_chip = host_y + 140
        for c in chips:
            cf = font("bold", 14)
            cw, _ = measure(c, cf)
            cw_total = cw + 18
            draw.rounded_rectangle([cx_chip, cy_chip, cx_chip + cw_total, cy_chip + 26], radius=12, fill=(255, 244, 224), outline=PRIMARY, width=1)
            text(draw, (cx_chip + 9, cy_chip + 4), c, cf, PRIMARY)
            cx_chip += cw_total + 6
        host_centers.append((hx + host_w // 2, host_y + host_h))

    # SSE arrows from each host to central
    central_top = (cx + central_w // 2, central_box[1])
    for h_center in host_centers:
        arrow(draw, h_center, central_top, color=PRIMARY, width=2, head=10)

    # right-side floating SSE label
    sse_label_x = central_box[0] + central_w + 24
    sse_label_y = central_box[1] + 10
    text(draw, (sse_label_x, sse_label_y), "← SSE", font("extrabold", 20), PRIMARY)
    text(draw, (sse_label_x, sse_label_y + 30), "1–3 초 동기화", font("medium", 16), MUTED)

    # Bottom KPI strip
    strip_y = central_box[3] + 50
    strip_h = 100
    strip_box = [MARGIN_X, strip_y, W - MARGIN_X, strip_y + strip_h]
    draw.rounded_rectangle(strip_box, radius=20, fill=INK)
    metrics = [("4+", "fleet 머신"), ("12", "활성 에이전트"), ("100%", "신호등 정직성"),
               ("1-3s", "SSE 지연"), ("∞", "수평 확장")]
    seg_w = (W - MARGIN_X * 2) // len(metrics)
    for i, (n, l) in enumerate(metrics):
        sx = MARGIN_X + i * seg_w
        if i > 0:
            draw.line([(sx, strip_y + 18), (sx, strip_y + strip_h - 18)], fill=MUTED, width=1)
        nf = font("extrabold", 36)
        nw, _ = measure(n, nf)
        text(draw, (sx + (seg_w - nw) // 2, strip_y + 14), n, nf, PRIMARY)
        lf = font("medium", 16)
        lw, _ = measure(l, lf)
        text(draw, (sx + (seg_w - lw) // 2, strip_y + 60), l, lf, PAPER2)

    return canvas.convert("RGB")


def s11_global(logo):
    """17 locales — Korean transliterations for non-Korean scripts so
    Pretendard renders all of them cleanly."""
    bg = gradient_bg(PAPER, PAPER2)
    bg = radial_glow((W // 2, int(H * 0.55)), (200, 136, 26, 28), 700, bg)
    bg = dot_pattern(bg, (200, 136, 26, 10), spacing=42, size=2)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)
    watermark_logo(canvas, logo)

    big_title(draw, "17 개 언어로 시작합니다")

    sub_font = font("medium", 26)
    text(draw, (MARGIN_X, MARGIN_Y + 130),
         "대시보드 · CLI · 시스템 메시지 — 클릭 한 번으로 전환", sub_font, MUTED)

    # 17 language cards in a 6x3 grid (last row centered with 5 entries)
    # ISO codes instead of emoji flags — Pretendard has no flag glyphs and
    # ISO codes are more universally legible in a pitch context anyway.
    langs = [
        ("한국어",     "Korean",      "KO"),
        ("영어",       "English",     "EN"),
        ("일본어",     "Japanese",    "JA"),
        ("중국어",     "Chinese",     "ZH"),
        ("독일어",     "Deutsch",     "DE"),
        ("프랑스어",   "Français",    "FR"),
        ("스페인어",   "Español",     "ES"),
        ("포르투갈어", "Português",   "PT"),
        ("인도네시아어","Bahasa",     "ID"),
        ("태국어",     "Thai",        "TH"),
        ("힌디어",     "Hindi",       "HI"),
        ("네팔어",     "Nepali",      "NE"),
        ("몽골어",     "Mongolian",   "MN"),
        ("아랍어",     "Arabic",      "AR"),
        ("하우사어",   "Hausa",       "HA"),
        ("스와힐리어", "Swahili",     "SW"),
        ("줄루어",     "Zulu",        "ZU"),
    ]
    cols = 6
    rows = 3
    cell_w = 280
    cell_h = 130
    gap_x = 18
    gap_y = 18
    grid_w = cols * cell_w + gap_x * (cols - 1)
    sx = (W - grid_w) // 2
    sy = 350
    for i, (kr, native, iso) in enumerate(langs):
        col = i % cols
        row = i // cols
        # last row: 5 items centered
        if row == 2:
            count_in_row = len(langs) - 12
            row_w = count_in_row * cell_w + (count_in_row - 1) * gap_x
            sx_row = (W - row_w) // 2
            x = sx_row + col * (cell_w + gap_x)
        else:
            x = sx + col * (cell_w + gap_x)
        y = sy + row * (cell_h + gap_y)
        fill = PAPER if (col + row) % 2 == 0 else PAPER2
        draw.rounded_rectangle([x, y, x + cell_w, y + cell_h], radius=18, fill=fill, outline=LINE, width=2)
        draw.rounded_rectangle([x, y, x + 6, y + cell_h], radius=3, fill=PRIMARY)
        # ISO code badge in top-right
        badge_w = 56
        badge_h = 32
        bx = x + cell_w - badge_w - 18
        by = y + 18
        draw.rounded_rectangle([bx, by, bx + badge_w, by + badge_h], radius=8, fill=PRIMARY)
        cf = font("extrabold", 16)
        cw, ch = measure(iso, cf)
        text(draw, (bx + (badge_w - cw) // 2, by + (badge_h - ch) // 2 - 3), iso, cf, PAPER)
        # Korean name (primary)
        nf = font("extrabold", 28)
        text(draw, (x + 26, y + 22), kr, nf, INK)
        # native script underneath
        ef = font("medium", 18)
        text(draw, (x + 26, y + 64), native, ef, MUTED)

    cap_font = font("medium", 22)
    cap = "한국 본사 + 동남아 · 중동 · 아프리카 · 라틴 시장 동시 대응"
    cw, _ = measure(cap, cap_font, draw)
    text(draw, ((W - cw) // 2, H - 130), cap, cap_font, MUTED)

    return canvas.convert("RGB")


def s12_compare(logo):
    bg = gradient_bg(PAPER, PAPER2)
    bg = dot_pattern(bg, (200, 136, 26, 10), spacing=42, size=2)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)
    watermark_logo(canvas, logo)

    big_title(draw, "경쟁 비교 — 암호화 · 가시성 · 실시간성")

    rows = [
        ("기능",                            "wall-vault",  "LiteLLM",  "OpenRouter SDK", "벤더 SDK 직접"),
        ("AES-GCM 암호화 키 금고",          "✓",           "✗",         "✗",               "✗"),
        ("4 API 포맷 동시 노출",             "✓",           "✓",         "✗",               "✗"),
        ("Strict-by-default + fallback 헤더","✓",           "✗",         "✗",               "✗"),
        ("SSE 실시간 fleet 동기화",          "✓",           "✗",         "✗",               "✗"),
        ("에이전트 페르소나 · 음성 통합",   "✓",           "✗",         "✗",               "✗"),
        ("17 개국 다국어 UI",                "✓",           "✗",         "✗",               "✗"),
        ("한 바이너리 · 무의존",             "✓",           "Python",    "JS",              "언어별"),
    ]
    table_x = MARGIN_X
    table_y = 280
    table_w = W - MARGIN_X * 2
    col_widths = [int(table_w * 0.36)] + [int(table_w * 0.16)] * 4
    row_h = 80

    # header
    draw.rounded_rectangle([table_x, table_y, table_x + sum(col_widths), table_y + row_h], radius=16, fill=INK)
    cx = table_x
    for i, (val, w) in enumerate(zip(rows[0], col_widths)):
        cw, ch = measure(val, font("bold", 24))
        if i == 0:
            text(draw, (cx + 28, table_y + (row_h - ch) // 2 - 2), val, font("bold", 24), PRIMARY)
        elif i == 1:
            text(draw, (cx + (w - cw) // 2, table_y + (row_h - ch) // 2 - 2), val, font("extrabold", 24), PRIMARY)
        else:
            text(draw, (cx + (w - cw) // 2, table_y + (row_h - ch) // 2 - 2), val, font("medium", 22), PAPER2)
        cx += w

    # body
    for r, row in enumerate(rows[1:], start=1):
        ry = table_y + r * row_h
        if r % 2 == 1:
            draw.rectangle([table_x, ry, table_x + sum(col_widths), ry + row_h], fill=PAPER2)
        # highlight wall-vault column
        wv_x = table_x + col_widths[0]
        draw.rectangle([wv_x, ry, wv_x + col_widths[1], ry + row_h], fill=(232, 184, 80, 40) if r % 2 == 0 else (232, 184, 80, 56))
        cx = table_x
        for c, (val, w) in enumerate(zip(row, col_widths)):
            if c == 0:
                cf = font("semibold", 24)
                cw, ch = measure(val, cf)
                text(draw, (cx + 28, ry + (row_h - ch) // 2 - 2), val, cf, INK)
            elif val == "✓":
                if c == 1:
                    fnt = font("extrabold", 36)
                    color = SUCCESS
                else:
                    fnt = font("bold", 30)
                    color = MUTED
                cw, ch = measure(val, fnt)
                text(draw, (cx + (w - cw) // 2, ry + (row_h - ch) // 2 - 4), val, fnt, color)
            elif val == "✗":
                fnt = font("bold", 30)
                cw, ch = measure(val, fnt)
                text(draw, (cx + (w - cw) // 2, ry + (row_h - ch) // 2 - 4), val, fnt, DANGER)
            else:
                cf = font("medium", 22)
                cw, ch = measure(val, cf)
                text(draw, (cx + (w - cw) // 2, ry + (row_h - ch) // 2 - 2), val, cf, MUTED)
            cx += w

    # outer border
    draw.rounded_rectangle(
        [table_x, table_y, table_x + sum(col_widths), table_y + row_h * len(rows)],
        radius=16, outline=LINE, width=2,
    )
    # vertical dividers
    cx = table_x
    for w in col_widths[:-1]:
        cx += w
        draw.line([(cx, table_y + row_h), (cx, table_y + row_h * len(rows))], fill=LINE, width=1)

    # footer takeaway band
    fy = table_y + row_h * len(rows) + 36
    fbox = [MARGIN_X, fy, W - MARGIN_X, fy + 80]
    draw.rounded_rectangle(fbox, radius=18, fill=INK)
    text(draw, (MARGIN_X + 28, fy + 24),
         "wall-vault 만 일곱 가지 차별점을 모두 갖춥니다.",
         font("extrabold", 26), PRIMARY)
    text(draw, (MARGIN_X + 28, fy + 56), "암호화 + 가시성 + 실시간성 + 글로벌 + 단일 바이너리.",
         font("medium", 18), PAPER2)

    return canvas.convert("RGB")


def s13_roadmap(logo):
    """Custom 4-phase roadmap — replaces Mermaid Gantt entirely."""
    bg = gradient_bg(PAPER, PAPER2)
    bg = dot_pattern(bg, (200, 136, 26, 10), spacing=42, size=2)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)
    watermark_logo(canvas, logo)

    big_title(draw, "로드맵 — 4 단계 전략")
    sub_font = font("medium", 26)
    text(draw, (MARGIN_X, MARGIN_Y + 130),
         "안정화 → 운영 자동화 → 확장 → 생태계.  분기 단위 실행.",
         sub_font, MUTED)

    # 4 phase columns
    phases = [
        {
            "step":   "01",
            "term":   "Q2 · 2026",
            "title":  "안정화",
            "tag":    "현재",
            "tag_color": SUCCESS,
            "tag_bg":    SUCCESS_BG,
            "items": [
                ("멀티벤더 라우팅 · 키 금고 핵심", "완료"),
                ("호스트 기반 신호등",            "완료"),
                ("Strict-by-default 라우팅",       "완료"),
                ("4 머신 fleet 안정 가동",          "완료"),
            ],
            "color": SUCCESS,
        },
        {
            "step":   "02",
            "term":   "Q3 · 2026",
            "title":  "운영 자동화",
            "tag":    "진행 중",
            "tag_color": PRIMARY,
            "tag_bg":    PAPER3,
            "items": [
                ("음성 인터페이스 통합",         "진행 중"),
                ("주기적 자율 기동 (cron)",      "진행 중"),
                ("감사 로그·이상 탐지",           "예정"),
                ("자동 키 회전 정책 엔진",         "예정"),
            ],
            "color": PRIMARY,
        },
        {
            "step":   "03",
            "term":   "Q4 · 2026",
            "title":  "확장",
            "tag":    "예정",
            "tag_color": INFO,
            "tag_bg":    INFO_BG,
            "items": [
                ("팀 · 조직 멀티테넌시",          "예정"),
                ("BYOK 셀프서비스 포털",          "예정"),
                ("엣지 캐싱 · 코스트 분석",        "예정"),
                ("엔터프라이즈 SSO / SCIM",        "예정"),
            ],
            "color": INFO,
        },
        {
            "step":   "04",
            "term":   "2027+",
            "title":  "생태계",
            "tag":    "계획",
            "tag_color": MUTED,
            "tag_bg":    (236, 234, 230),
            "items": [
                ("플러그인 SDK 공개",            "계획"),
                ("공식 클라이언트 라이브러리",     "계획"),
                ("커뮤니티 마켓플레이스",          "계획"),
                ("벤더 인증 프로그램",             "계획"),
            ],
            "color": MUTED,
        },
    ]

    # 4 column layout
    n = len(phases)
    avail_w = W - MARGIN_X * 2
    gap = 24
    col_w = (avail_w - gap * (n - 1)) // n
    col_top = 320
    col_bottom = H - 100
    col_h = col_bottom - col_top

    # Connector spine running horizontally across all columns
    spine_y = col_top + 90
    draw.line([(MARGIN_X + 50, spine_y), (W - MARGIN_X - 50, spine_y)], fill=LINE, width=3)

    for i, ph in enumerate(phases):
        x = MARGIN_X + i * (col_w + gap)
        y = col_top
        # column card
        canvas = soft_shadow_card(canvas, [x, y, x + col_w, y + col_h], offset=(0, 12), blur=18, alpha=44)
        draw = ImageDraw.Draw(canvas)
        draw.rounded_rectangle([x, y, x + col_w, y + col_h], radius=22, fill=PAPER, outline=LINE, width=2)
        # top accent stripe
        draw.rounded_rectangle([x, y, x + col_w, y + 8], radius=4, fill=ph["color"])

        # phase number circle on the spine
        circle_r = 36
        circle_cx = x + col_w // 2
        circle_cy = spine_y
        draw.ellipse([circle_cx - circle_r, circle_cy - circle_r, circle_cx + circle_r, circle_cy + circle_r],
                     fill=ph["color"], outline=PAPER, width=4)
        nf = font("extrabold", 28)
        nw, nh = measure(ph["step"], nf)
        text(draw, (circle_cx - nw // 2, circle_cy - nh // 2 - 4), ph["step"], nf, PAPER)

        # term label
        tf = font("bold", 18)
        text(draw, (x + 28, y + 32), ph["term"], tf, MUTED)
        # title
        ttl_f = font("extrabold", 36)
        text(draw, (x + 28, y + 60), ph["title"], ttl_f, INK)
        # status tag
        stag = ph["tag"]
        sf = font("bold", 16)
        sw, sh = measure(stag, sf)
        chip_x = x + 28
        chip_y = y + 160
        chip_w = sw + 24
        chip_h = 32
        draw.rounded_rectangle([chip_x, chip_y, chip_x + chip_w, chip_y + chip_h],
                               radius=16, fill=ph["tag_bg"], outline=ph["tag_color"], width=1)
        text(draw, (chip_x + 12, chip_y + (chip_h - sh) // 2 - 4), stag, sf, ph["tag_color"])

        # items
        item_y = y + 220
        for item_text, status in ph["items"]:
            # status icon
            if status == "완료":
                draw.ellipse([x + 28, item_y + 8, x + 28 + 18, item_y + 26], fill=SUCCESS)
                tick_f = font("extrabold", 14)
                text(draw, (x + 32, item_y + 6), "✓", tick_f, PAPER)
            elif status == "진행 중":
                draw.ellipse([x + 28, item_y + 8, x + 28 + 18, item_y + 26], fill=PRIMARY)
                tick_f = font("extrabold", 14)
                text(draw, (x + 33, item_y + 4), "·", tick_f, PAPER)
            elif status == "예정":
                draw.ellipse([x + 28, item_y + 8, x + 28 + 18, item_y + 26], outline=INFO, width=2)
            else:
                draw.ellipse([x + 28, item_y + 8, x + 28 + 18, item_y + 26], outline=MUTED2, width=2)
            text_wrapped(draw, (x + 56, item_y + 4), item_text, font("medium", 19), INK, col_w - 90, 26)
            item_y += 60

    return canvas.convert("RGB")


def s14_cta(logo):
    bg = gradient_bg(PAPER, PAPER2)
    bg = radial_glow((W // 2, int(H * 0.40)), (200, 136, 26, 80), 800, bg)
    bg = radial_glow((int(W * 0.85), int(H * 0.85)), (224, 154, 50, 50), 500, bg)
    bg = dot_pattern(bg, (200, 136, 26, 14), spacing=42, size=2)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)

    mini = logo.copy()
    mini.thumbnail((180, 180), Image.LANCZOS)
    paste_image(canvas, mini, center=(W // 2, MARGIN_Y + 130))

    title_font = font("extrabold", 110)
    title = "함께 만들어 갈까요?"
    tw, th = measure(title, title_font, draw)
    text(draw, ((W - tw) // 2, int(H * 0.32)), title, title_font, INK)

    subtitle_font = font("medium", 30)
    sub = "연구실 운영자에서 출발했습니다.  이제 시장 차례입니다."
    sw, _ = measure(sub, subtitle_font, draw)
    text(draw, ((W - sw) // 2, int(H * 0.32) + th + 18), sub, subtitle_font, MUTED)

    # 4 CTA cards
    cta_y = int(H * 0.58)
    cta_h = 200
    cards = [
        ("01", "데모 요청",    "라이브 데모로\n동작 확인"),
        ("02", "파일럿 도입",  "사내 환경 1 주\n무료 검증"),
        ("03", "기술 제휴",    "API · SDK · 공동\n로드맵 협업"),
        ("04", "투자 검토",    "라운드 정보\nIR 자료 공유"),
    ]
    n = len(cards)
    gap = 28
    cta_w = (W - MARGIN_X * 2 - gap * (n - 1)) // n
    for i, (num, ttl, desc) in enumerate(cards):
        x = MARGIN_X + i * (cta_w + gap)
        canvas = soft_shadow_card(canvas, [x, cta_y, x + cta_w, cta_y + cta_h], offset=(0, 12), blur=18, alpha=60)
        draw = ImageDraw.Draw(canvas)
        draw.rounded_rectangle([x, cta_y, x + cta_w, cta_y + cta_h], radius=22, fill=INK, outline=PRIMARY, width=2)
        text(draw, (x + 28, cta_y + 22), num, font("extrabold", 30), PRIMARY)
        text(draw, (x + 28, cta_y + 64), ttl, font("extrabold", 32), PAPER)
        text(draw, (x + 28, cta_y + 116), desc, font("medium", 18), PAPER2, spacing=4)

    # signature rule + tagline
    rule_y = cta_y + cta_h + 60
    draw.line([((W - 280) // 2, rule_y), ((W + 280) // 2, rule_y)], fill=PRIMARY, width=3)
    sig_font = font("medium", 22)
    sig = "wall-vault — AI 가 절대 끊기지 않는 멀티벤더 게이트웨이"
    sw, _ = measure(sig, sig_font, draw)
    text(draw, ((W - sw) // 2, rule_y + 24), sig, sig_font, MUTED)

    return canvas.convert("RGB")


# ── Orchestration ─────────────────────────────────────────────────────


def main():
    OUT.mkdir(parents=True, exist_ok=True)

    logo = Image.open(IMG / "logo.png").convert("RGBA")
    signal_lights = Image.open(IMG / "signal-lights.png").convert("RGBA")

    @dataclass
    class Spec:
        idx: int
        builder: callable
        chrome: bool = True

    builders = [
        Spec(1,  lambda: s01_title(logo),                                            chrome=False),
        Spec(2,  lambda: s02_section(1, "문제 의식", "AI 가 멈추는 순간,", "일도 멈춥니다"), chrome=False),
        Spec(3,  lambda: s03_problem(logo)),
        Spec(4,  lambda: s04_definition(logo)),
        Spec(5,  lambda: s05_architecture(logo)),
        Spec(6,  lambda: s02_section(2, "차별점", "시중 솔루션과", "다른 세 가지"),         chrome=False),
        Spec(7,  lambda: s07_diff_multivendor(logo)),
        Spec(8,  lambda: s08_diff_routing(logo)),
        Spec(9,  lambda: s09_diff_observability(signal_lights, logo)),
        Spec(10, lambda: s10_proof(logo)),
        Spec(11, lambda: s11_global(logo)),
        Spec(12, lambda: s12_compare(logo)),
        Spec(13, lambda: s13_roadmap(logo)),
        Spec(14, lambda: s14_cta(logo),                                              chrome=False),
    ]
    total = len(builders)

    for spec in builders:
        img = spec.builder()
        if spec.chrome:
            page_chrome(img, spec.idx, total)
        out_path = OUT / f"slide-{spec.idx:02d}.png"
        img.save(out_path, "PNG", optimize=True)
        print(f"  ✓ slide-{spec.idx:02d}  ({out_path.stat().st_size // 1024} KB)")


if __name__ == "__main__":
    main()
