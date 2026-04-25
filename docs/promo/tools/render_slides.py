"""Render the wall-vault pitch deck as 14 individual 1920×1080 PNGs.

The shipped sandbox lacks Chromium (so Marp / Playwright / mmdc in-process
all fail) AND has no CJK font (so python-pptx's text shapes degrade to
tofu-boxes when previewed by anyone whose viewer falls back to a Latin
font). This renderer sidesteps both problems by drawing every slide
pixel via PIL with a bundled Pretendard, then a separate step wraps the
PNGs into a PPTX as full-bleed background images.

Tradeoff: text is no longer editable inside PowerPoint. For a pitch deck
that's an acceptable price — the deliverable's aesthetic consistency
matters more than per-cell editability, and the source-of-truth narrative
lives in deck.md so authoring still happens in markdown.
"""

from __future__ import annotations

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
MUTED   = (0x5b, 0x68, 0x77)
MUTED2  = (0x8b, 0x95, 0xa0)
PRIMARY = (0xc8, 0x88, 0x1a)
ACCENT  = (0xe0, 0x9a, 0x32)
PAPER   = (0xff, 0xfa, 0xf2)
PAPER2  = (0xff, 0xf4, 0xe0)
LINE    = (0xea, 0xd8, 0xb6)
SUCCESS = (0x2f, 0x9d, 0x6b)
DANGER  = (0xc6, 0x4a, 0x3c)


# ── Geometry — 1920x1080, generous margins ────────────────────────────

W, H = 1920, 1080
MARGIN_X = 120
MARGIN_Y = 90
GUTTER = 48


def font(weight: str, size: int) -> ImageFont.FreeTypeFont:
    file = {
        "regular":   "Pretendard-Regular.otf",
        "medium":    "Pretendard-Medium.otf",
        "semibold":  "Pretendard-SemiBold.otf",
        "bold":      "Pretendard-Bold.otf",
        "extrabold": "Pretendard-ExtraBold.otf",
    }[weight]
    return ImageFont.truetype(str(FONTS / file), size)


# ── Drawing helpers ────────────────────────────────────────────────────


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
    """Soft radial glow centred at `center`, blended additively onto base."""
    layer = Image.new("RGBA", (W, H), (0, 0, 0, 0))
    d = ImageDraw.Draw(layer)
    cx, cy = center
    d.ellipse(
        [cx - radius, cy - radius, cx + radius, cy + radius],
        fill=color_rgba,
    )
    layer = layer.filter(ImageFilter.GaussianBlur(radius // 4))
    return Image.alpha_composite(base.convert("RGBA"), layer).convert("RGB")


def soft_shadow(box, color_rgba, blur: int, base: Image.Image) -> Image.Image:
    layer = Image.new("RGBA", (W, H), (0, 0, 0, 0))
    d = ImageDraw.Draw(layer)
    d.rounded_rectangle(box, radius=24, fill=color_rgba)
    layer = layer.filter(ImageFilter.GaussianBlur(blur))
    return Image.alpha_composite(base.convert("RGBA"), layer).convert("RGB")


def draw_card(
    draw: ImageDraw.ImageDraw,
    box,
    *,
    fill=PAPER,
    outline=LINE,
    width=2,
    radius=24,
):
    draw.rounded_rectangle(box, radius=radius, fill=fill, outline=outline, width=width)


def draw_text(
    draw: ImageDraw.ImageDraw,
    pos,
    text: str,
    fnt: ImageFont.FreeTypeFont,
    fill,
    *,
    spacing: int = 8,
):
    draw.text(pos, text, font=fnt, fill=fill, spacing=spacing)


def fit_image(path: Path, max_w: int, max_h: int) -> Image.Image:
    img = Image.open(path).convert("RGBA")
    img.thumbnail((max_w, max_h), Image.LANCZOS)
    return img


def paste_image(canvas: Image.Image, img: Image.Image, center=None, top_left=None):
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


def page_chrome(canvas: Image.Image, page: int, total: int):
    """Bottom footer line with page indicator (skipped on lead/section slides)."""
    draw = ImageDraw.Draw(canvas)
    y = H - 50
    draw.line([(MARGIN_X, y), (W - MARGIN_X, y)], fill=LINE, width=1)
    f = font("medium", 18)
    draw_text(draw, (MARGIN_X, y + 14), "wall-vault · 투자·제휴 피칭 · 2026", f, MUTED)
    page_str = f"{page:02d} / {total:02d}"
    pw, ph = measure(page_str, f, draw)
    draw_text(draw, (W - MARGIN_X - pw, y + 14), page_str, f, MUTED)


def watermark_logo(canvas: Image.Image, logo: Image.Image, opacity: int = 28):
    """Faint logo watermark in top-right corner."""
    wm = logo.copy()
    wm.thumbnail((180, 180), Image.LANCZOS)
    alpha = wm.split()[-1]
    alpha = alpha.point(lambda v: min(v, opacity))
    wm.putalpha(alpha)
    canvas.paste(wm, (W - 200, 50), wm)


# ── Slide builders ────────────────────────────────────────────────────


def s01_title(logo: Image.Image) -> Image.Image:
    # Logo already contains the "wall-vault" wordmark, so the title slide
    # leans on the logo as the wordmark and uses oversized space + a single
    # tagline + decorative rule. No redundant text wordmark.
    bg = gradient_bg(PAPER, PAPER2)
    bg = radial_glow((W // 2, int(H * 0.40)), (200, 136, 26, 70), 800, bg)
    bg = radial_glow((int(W * 0.88), int(H * 0.85)), (224, 154, 50, 40), 500, bg)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)

    big_logo = logo.copy()
    big_logo.thumbnail((720, 720), Image.LANCZOS)
    paste_image(canvas, big_logo, center=(W // 2, int(H * 0.40)))

    sub_font = font("extrabold", 60)
    sub = "AI 가 절대 끊기지 않는"
    sub2 = "멀티벤더 게이트웨이"
    sw, sh = measure(sub, sub_font, draw)
    sw2, sh2 = measure(sub2, sub_font, draw)
    head_y = int(H * 0.74)
    draw_text(draw, ((W - sw) // 2, head_y), sub, sub_font, INK)
    draw_text(draw, ((W - sw2) // 2, head_y + sh + 6), sub2, sub_font, PRIMARY)

    # decorative rule under the headline
    rule_y = head_y + sh + sh2 + 28
    draw.line([((W - 260) // 2, rule_y), ((W + 260) // 2, rule_y)], fill=PRIMARY, width=3)

    tag_font = font("medium", 26)
    tag = "키 금고 · 지능형 라우팅 · fleet observability — 한 바이너리에"
    tagw, _ = measure(tag, tag_font, draw)
    draw_text(draw, ((W - tagw) // 2, rule_y + 26), tag, tag_font, MUTED)

    return canvas.convert("RGB")


def s02_section_problem() -> Image.Image:
    bg = gradient_bg(INK, INK_SOFT)
    bg = radial_glow((W // 2, int(H * 0.5)), (200, 136, 26, 50), 800, bg)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)

    label_font = font("bold", 28)
    label = "PART · 1"
    lw, _ = measure(label, label_font, draw)
    draw_text(draw, ((W - lw) // 2, int(H * 0.32)), label, label_font, PRIMARY)

    title_font = font("extrabold", 110)
    title = "AI 가 멈추는 순간,"
    tw, th = measure(title, title_font, draw)
    draw_text(draw, ((W - tw) // 2, int(H * 0.40)), title, title_font, PAPER)

    title2 = "일도 멈춥니다"
    tw2, _ = measure(title2, title_font, draw)
    draw_text(draw, ((W - tw2) // 2, int(H * 0.40) + th + 8), title2, title_font, ACCENT)

    return canvas.convert("RGB")


def s03_problem() -> Image.Image:
    bg = gradient_bg(PAPER, PAPER2)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)

    head_font = font("extrabold", 64)
    accent_font = font("extrabold", 64)
    sub_font = font("semibold", 28)

    # title with mixed color
    head1 = "AI 가 끊기는 순간, "
    head2 = "비용은 두 배가 됩니다"
    h1w, h1h = measure(head1, head_font, draw)
    head_y = MARGIN_Y + 30
    draw_text(draw, (MARGIN_X, head_y), head1, head_font, INK)
    draw_text(draw, (MARGIN_X + h1w, head_y), head2, accent_font, PRIMARY)

    # accent rule
    rule_y = head_y + h1h + 28
    draw.line([(MARGIN_X, rule_y), (MARGIN_X + 240, rule_y)], fill=PRIMARY, width=4)

    # bullets
    bullets = [
        ("클라우드 API 키 만료 · 쿨다운 · credit-out", "한 키가 죽으면 fleet 전체 정지"),
        ("벤더별 모델 ID 네임스페이스 차이", "이식할 때마다 코드 변경 비용 누적"),
        ("멀티 머신 · 멀티 에이전트 운영의 가시성", "어디서 뭐가 멈췄는지 안 보임"),
        ("fallback 의 무단 모델 치환", "응답 품질이 조용히 무너짐"),
        ("매일 “왜 또 멈췄지?” 라는 운영 부담", "장애 대응에 사람이 묶임"),
    ]
    bul_font = font("semibold", 26)
    desc_font = font("regular", 22)
    y = rule_y + 60
    bullet_x = MARGIN_X
    for primary_txt, desc_txt in bullets:
        # bullet dot
        draw.ellipse([bullet_x, y + 10, bullet_x + 14, y + 24], fill=PRIMARY)
        draw_text(draw, (bullet_x + 30, y), primary_txt, bul_font, INK)
        _, bh = measure(primary_txt, bul_font, draw)
        draw_text(draw, (bullet_x + 30, y + bh + 6), desc_txt, desc_font, MUTED)
        y += bh + 60

    # quote card on the right (drop-shadow on a transparent layer, NOT a
    # blurred copy of the canvas — the latter would blur every primitive
    # we already drew on the left).
    qx, qy, qw, qh = 1200, rule_y + 40, 580, 460
    shadow = Image.new("RGBA", (W, H), (0, 0, 0, 0))
    shadow_draw = ImageDraw.Draw(shadow)
    shadow_draw.rounded_rectangle(
        [qx + 8, qy + 14, qx + qw + 8, qy + qh + 14],
        radius=24, fill=(0, 0, 0, 56),
    )
    shadow = shadow.filter(ImageFilter.GaussianBlur(12))
    canvas = Image.alpha_composite(canvas, shadow)
    draw = ImageDraw.Draw(canvas)
    draw.rounded_rectangle([qx, qy, qx + qw, qy + qh], radius=24, fill=PAPER2, outline=LINE, width=2)
    # left bar
    draw.rounded_rectangle([qx, qy, qx + 8, qy + qh], radius=4, fill=PRIMARY)
    # quote mark
    qm_font = font("extrabold", 96)
    draw_text(draw, (qx + 32, qy + 10), "“", qm_font, PRIMARY)
    # body
    quote_lines = [
        "키 하나 죽으면",
        "fleet 전체가 멈춥니다.",
        "",
        "그게 어디서 어떻게 멈췄는지조차",
        "보이지 않습니다.",
    ]
    qf = font("semibold", 30)
    qy_cursor = qy + 130
    for line in quote_lines:
        if line:
            draw_text(draw, (qx + 50, qy_cursor), line, qf, INK)
            _, lh = measure(line, qf, draw)
            qy_cursor += lh + 14
        else:
            qy_cursor += 26

    return canvas.convert("RGB")


def s04_definition() -> Image.Image:
    bg = gradient_bg(PAPER, PAPER2)
    bg = radial_glow((W // 2, int(H * 0.5)), (200, 136, 26, 36), 700, bg)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)

    label_font = font("semibold", 26)
    label = "wall-vault · 한 줄 정의"
    lw, lh = measure(label, label_font, draw)
    draw_text(draw, ((W - lw) // 2, MARGIN_Y + 40), label, label_font, MUTED)

    # the headline
    h_font = font("extrabold", 76)
    plus_font = font("extrabold", 60)
    part_a = "AES-GCM 암호화 키 금고"
    plus = "  +  "
    part_b = "지능형 멀티벤더 프록시"
    aw, ah = measure(part_a, h_font, draw)
    pw, ph = measure(plus, plus_font, draw)
    bw, bh = measure(part_b, h_font, draw)
    total = aw + pw + bw
    y = int(H * 0.30)
    x = (W - total) // 2
    draw_text(draw, (x, y), part_a, h_font, INK)
    draw_text(draw, (x + aw, y + 14), plus, plus_font, PRIMARY)
    draw_text(draw, (x + aw + pw, y), part_b, h_font, INK)

    # underline accent under each part
    draw.line([(x, y + ah + 12), (x + aw, y + ah + 12)], fill=PRIMARY, width=4)
    draw.line(
        [(x + aw + pw, y + ah + 12), (x + aw + pw + bw, y + ah + 12)],
        fill=PRIMARY,
        width=4,
    )

    # chips
    chips = ["암호화 저장", "자동 키 로테이션", "정직한 라우팅", "실시간 신호등"]
    chip_font = font("bold", 24)
    chip_pad_x = 28
    chip_h = 60
    chip_y = int(H * 0.58)
    widths = []
    for c in chips:
        cw, _ = measure(c, chip_font, draw)
        widths.append(cw + chip_pad_x * 2)
    gap = 24
    total_w = sum(widths) + gap * (len(chips) - 1)
    x = (W - total_w) // 2
    for c, w in zip(chips, widths):
        draw.rounded_rectangle([x, chip_y, x + w, chip_y + chip_h], radius=30, fill=PRIMARY)
        cw, ch = measure(c, chip_font, draw)
        draw_text(
            draw,
            (x + (w - cw) // 2, chip_y + (chip_h - ch) // 2 - 4),
            c,
            chip_font,
            PAPER,
        )
        x += w + gap

    # bottom caption
    cap_font = font("regular", 26)
    cap = "한 개의 Go 바이너리, 무설치 의존성, 17 개 언어, Linux / macOS / Windows / WSL"
    cw, ch = measure(cap, cap_font, draw)
    draw_text(draw, ((W - cw) // 2, int(H * 0.78)), cap, cap_font, MUTED)

    return canvas.convert("RGB")


def s05_architecture(diagram: Image.Image) -> Image.Image:
    return image_focus_slide(
        title="아키텍처 — 한 장으로 보는 동작",
        diagram=diagram,
        caption="벽금고(:56243) 가 키·설정의 단일 소스, 프록시(:56244) 가 지능형 라우터.  SSE 로 실시간 동기화.",
    )


def image_focus_slide(*, title: str, diagram: Image.Image, caption: str) -> Image.Image:
    bg = gradient_bg(PAPER, PAPER2)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)

    title_font = font("extrabold", 56)
    draw_text(draw, (MARGIN_X, MARGIN_Y + 40), title, title_font, INK)
    _, th = measure(title, title_font, draw)
    rule_y = MARGIN_Y + 40 + th + 22
    draw.line([(MARGIN_X, rule_y), (MARGIN_X + 240, rule_y)], fill=PRIMARY, width=4)

    # diagram area
    avail_w = W - MARGIN_X * 2
    avail_h = int(H * 0.62)
    img = diagram.copy()
    img.thumbnail((avail_w, avail_h), Image.LANCZOS)
    paste_image(canvas, img, center=(W // 2, rule_y + 60 + avail_h // 2))

    # caption
    cap_font = font("medium", 24)
    cw, ch = measure(caption, cap_font, draw)
    draw_text(draw, ((W - cw) // 2, H - 130), caption, cap_font, MUTED)

    return canvas.convert("RGB")


def s06_section_diff() -> Image.Image:
    bg = gradient_bg(INK, INK_SOFT)
    bg = radial_glow((int(W * 0.15), int(H * 0.5)), (200, 136, 26, 50), 700, bg)
    bg = radial_glow((int(W * 0.85), int(H * 0.5)), (224, 154, 50, 40), 600, bg)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)

    label_font = font("bold", 28)
    label = "PART · 2"
    lw, _ = measure(label, label_font, draw)
    draw_text(draw, ((W - lw) // 2, int(H * 0.32)), label, label_font, PRIMARY)

    title_font = font("extrabold", 110)
    title = "시중 솔루션과"
    title2 = "다른 세 가지"
    tw1, th1 = measure(title, title_font, draw)
    tw2, th2 = measure(title2, title_font, draw)
    draw_text(draw, ((W - tw1) // 2, int(H * 0.40)), title, title_font, PAPER)
    draw_text(draw, ((W - tw2) // 2, int(H * 0.40) + th1 + 8), title2, title_font, ACCENT)

    return canvas.convert("RGB")


def s07_diff_multivendor() -> Image.Image:
    bg = gradient_bg(PAPER, PAPER2)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)

    _eyebrow_title(draw, "차별점 ①", "멀티벤더를 진짜로 통합")

    # bullets on the left
    bullets = [
        ("4 가지 API 포맷 동시 노출", "Gemini · OpenAI · Anthropic · OpenRouter"),
        ("170+ 모델 자동 라우팅", "로컬 Ollama 부터 Claude / Gemini / GPT 까지"),
        ("클라이언트는 자기 형식 그대로 호출", "Cline · Cursor · Claude Code · OpenClaw — 코드 변경 0"),
        ("벤더 SDK 4 개 중복 학습 → 1 개 게이트웨이로", "팀 학습 곡선·운영 비용 절감"),
    ]
    bul_font = font("semibold", 30)
    desc_font = font("regular", 22)
    y = 320
    for primary_txt, desc_txt in bullets:
        draw.ellipse([MARGIN_X, y + 10, MARGIN_X + 14, y + 24], fill=PRIMARY)
        draw_text(draw, (MARGIN_X + 32, y), primary_txt, bul_font, INK)
        _, bh = measure(primary_txt, bul_font, draw)
        draw_text(draw, (MARGIN_X + 32, y + bh + 6), desc_txt, desc_font, MUTED)
        y += bh + 60

    # KPI grid on the right
    kpis = [("4",   "API 포맷"), ("170+", "모델"),
            ("10+", "에이전트 타입"), ("17",  "로케일")]
    grid_x = 1180
    grid_y = 320
    cell_w = 260
    cell_h = 220
    gap = 28
    for i, (num, lab) in enumerate(kpis):
        col = i % 2
        row = i // 2
        x = grid_x + col * (cell_w + gap)
        yy = grid_y + row * (cell_h + gap)
        draw.rounded_rectangle([x, yy, x + cell_w, yy + cell_h], radius=24, fill=PAPER, outline=LINE, width=2)
        nf = font("extrabold", 96)
        nw, nh = measure(num, nf, draw)
        draw_text(draw, (x + (cell_w - nw) // 2, yy + 30), num, nf, PRIMARY)
        lf = font("semibold", 22)
        lw, lh = measure(lab, lf, draw)
        draw_text(draw, (x + (cell_w - lw) // 2, yy + cell_h - lh - 28), lab, lf, MUTED)

    return canvas.convert("RGB")


def s08_diff_routing(diagram: Image.Image) -> Image.Image:
    bg = gradient_bg(PAPER, PAPER2)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)

    _eyebrow_title(draw, "차별점 ②", "정직한 라우팅 (Strict-by-default)")

    # diagram on left
    img = diagram.copy()
    img.thumbnail((900, 720), Image.LANCZOS)
    paste_image(canvas, img, center=(550, 660))

    # bullets on right
    points = [
        ("Primary 실패 ≠ 다른 모델 치환",
         "fallback 은 명시적 opt-in 만"),
        ("실제 사용 서비스·모델·사유를 응답 헤더로",
         "X-WV-Used-Service / Used-Model / Fallback-Reason"),
        ("호출자가 한 줄 헤더로 검증",
         "운영 사고를 응답 시점에 즉시 포착"),
        ("“조용한 실패” 가 사라집니다",
         "기존 멀티벤더 게이트웨이가 가장 자주 어기는 약속"),
    ]
    bul_font = font("semibold", 26)
    desc_font = font("regular", 21)
    y = 320
    rx = 1100
    for primary_txt, desc_txt in points:
        draw.ellipse([rx, y + 10, rx + 14, y + 24], fill=PRIMARY)
        draw_text(draw, (rx + 30, y), primary_txt, bul_font, INK)
        _, bh = measure(primary_txt, bul_font, draw)
        draw_text(draw, (rx + 30, y + bh + 6), desc_txt, desc_font, MUTED)
        y += bh + 60

    return canvas.convert("RGB")


def s09_diff_observability(mock: Image.Image) -> Image.Image:
    bg = gradient_bg(PAPER, PAPER2)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)

    _eyebrow_title(draw, "차별점 ③", "Fleet observability — 한 곳에서 다 봅니다")

    # full-width mock
    img = mock.copy()
    avail_w = W - MARGIN_X * 2
    img.thumbnail((avail_w, 700), Image.LANCZOS)
    paste_image(canvas, img, center=(W // 2, 670))

    return canvas.convert("RGB")


def s10_proof(topology: Image.Image) -> Image.Image:
    bg = gradient_bg(PAPER, PAPER2)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)

    title_font = font("extrabold", 56)
    draw_text(draw, (MARGIN_X, MARGIN_Y + 40), "실측 운영 — 끊김 없는 fleet", title_font, INK)
    _, th = measure("실측 운영 — 끊김 없는 fleet", title_font, draw)
    rule_y = MARGIN_Y + 40 + th + 22
    draw.line([(MARGIN_X, rule_y), (MARGIN_X + 240, rule_y)], fill=PRIMARY, width=4)

    # topology on left
    img = topology.copy()
    img.thumbnail((900, 700), Image.LANCZOS)
    paste_image(canvas, img, center=(540, 640))

    # KPIs + bullets on right
    bullets = [
        ("WSL · macOS GPU · ARM SBC · 전용 노드",
         "OS · 아키텍처 무관 · 같은 바이너리"),
        ("호스트 매칭 + agent_type 별 liveness",
         "거짓 초록불 · 거짓 빨간불 모두 제거"),
        ("AES-GCM 키 로테이션 + 자동 쿨다운",
         "credit-out 일에도 즉시 로컬 추론 전환"),
        ("단일 머신 → 수십 노드까지 같은 도구",
         "확장 시 학습·재배포 비용 0"),
    ]
    bul_font = font("semibold", 26)
    desc_font = font("regular", 21)
    y = 280
    rx = 1080
    for primary_txt, desc_txt in bullets:
        draw.ellipse([rx, y + 10, rx + 14, y + 24], fill=PRIMARY)
        draw_text(draw, (rx + 30, y), primary_txt, bul_font, INK)
        _, bh = measure(primary_txt, bul_font, draw)
        draw_text(draw, (rx + 30, y + bh + 6), desc_txt, desc_font, MUTED)
        y += bh + 60

    return canvas.convert("RGB")


def s11_global() -> Image.Image:
    bg = gradient_bg(PAPER, PAPER2)
    bg = radial_glow((W // 2, int(H * 0.55)), (200, 136, 26, 30), 700, bg)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)

    title_font = font("extrabold", 64)
    draw_text(draw, (MARGIN_X, MARGIN_Y + 40), "17 개 언어로 시작합니다", title_font, INK)
    _, th = measure("17 개 언어로 시작합니다", title_font, draw)
    rule_y = MARGIN_Y + 40 + th + 22
    draw.line([(MARGIN_X, rule_y), (MARGIN_X + 240, rule_y)], fill=PRIMARY, width=4)

    sub_font = font("medium", 28)
    sub = "대시보드 · CLI · 시스템 메시지 — 클릭 한 번으로 전환"
    draw_text(draw, (MARGIN_X, rule_y + 24), sub, sub_font, MUTED)

    # language chips grid (5 columns)
    langs = [
        "한국어", "English", "日本語", "中文", "Deutsch",
        "Français", "Español", "Português", "Bahasa", "ไทย",
        "हिन्दी", "नेपाली", "Монгол", "العربية", "Hausa",
        "Kiswahili", "isiZulu",
    ]
    cols = 6
    chip_w = 240
    chip_h = 80
    gap = 24
    grid_w = cols * chip_w + gap * (cols - 1)
    start_x = (W - grid_w) // 2
    start_y = 460
    chip_font = font("semibold", 28)
    for i, lang in enumerate(langs):
        col = i % cols
        row = i // cols
        x = start_x + col * (chip_w + gap)
        y = start_y + row * (chip_h + gap)
        # alternate fill for visual rhythm
        fill = PAPER if (i % 2 == 0) else PAPER2
        draw.rounded_rectangle([x, y, x + chip_w, y + chip_h], radius=18, fill=fill, outline=LINE, width=2)
        lw, lh = measure(lang, chip_font, draw)
        draw_text(draw, (x + (chip_w - lw) // 2, y + (chip_h - lh) // 2 - 4), lang, chip_font, INK)

    cap_font = font("medium", 22)
    cap = "한국 본사 + 동남아 · 중동 · 아프리카 · 라틴 시장 동시 대응"
    cw, ch = measure(cap, cap_font, draw)
    draw_text(draw, ((W - cw) // 2, H - 140), cap, cap_font, MUTED)

    return canvas.convert("RGB")


def s12_compare() -> Image.Image:
    bg = gradient_bg(PAPER, PAPER2)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)

    title_font = font("extrabold", 56)
    draw_text(draw, (MARGIN_X, MARGIN_Y + 40), "경쟁 비교 — 암호화 · 가시성 · 실시간성", title_font, INK)
    _, th = measure("경쟁 비교 — 암호화 · 가시성 · 실시간성", title_font, draw)
    rule_y = MARGIN_Y + 40 + th + 22
    draw.line([(MARGIN_X, rule_y), (MARGIN_X + 240, rule_y)], fill=PRIMARY, width=4)

    rows = [
        ("기능", "wall-vault", "LiteLLM", "OpenRouter SDK", "벤더 SDK 직접"),
        ("AES-GCM 암호화 키 금고",        "✓", "✗", "✗", "✗"),
        ("4 API 포맷 동시 노출",           "✓", "✓", "✗", "✗"),
        ("Strict-by-default + fallback 헤더","✓", "✗", "✗", "✗"),
        ("SSE 실시간 fleet 동기화",        "✓", "✗", "✗", "✗"),
        ("에이전트 페르소나 · 음성 통합",  "✓", "✗", "✗", "✗"),
        ("17 개국 다국어 UI",              "✓", "✗", "✗", "✗"),
        ("한 바이너리 · 무의존",           "✓", "Python", "JS", "언어별"),
    ]

    table_x = MARGIN_X
    table_y = rule_y + 50
    table_w = W - MARGIN_X * 2
    col_widths = [int(table_w * 0.36)] + [int(table_w * 0.16)] * 4
    row_h = 70

    head_font = font("bold", 22)
    cell_font = font("medium", 20)
    yes_font = font("extrabold", 30)

    # header
    draw.rounded_rectangle([table_x, table_y, table_x + sum(col_widths), table_y + row_h], radius=16, fill=INK)
    cx = table_x
    for i, (val, w) in enumerate(zip(rows[0], col_widths)):
        align_center = i > 0
        cw, ch = measure(val, head_font, draw)
        if align_center:
            draw_text(draw, (cx + (w - cw) // 2, table_y + (row_h - ch) // 2 - 2), val, head_font, PAPER)
        else:
            draw_text(draw, (cx + 24, table_y + (row_h - ch) // 2 - 2), val, head_font, PAPER)
        cx += w

    # body
    for r, row in enumerate(rows[1:], start=1):
        ry = table_y + r * row_h
        if r % 2 == 1:
            draw.rectangle([table_x, ry, table_x + sum(col_widths), ry + row_h], fill=PAPER2)
        cx = table_x
        for c, (val, w) in enumerate(zip(row, col_widths)):
            if c == 0:
                cf = font("semibold", 22)
                cw, ch = measure(val, cf, draw)
                draw_text(draw, (cx + 24, ry + (row_h - ch) // 2 - 2), val, cf, INK)
            elif val == "✓":
                color = SUCCESS if c == 1 else MUTED
                fnt = yes_font if c == 1 else font("bold", 26)
                cw, ch = measure(val, fnt, draw)
                draw_text(draw, (cx + (w - cw) // 2, ry + (row_h - ch) // 2 - 2), val, fnt, color)
            elif val == "✗":
                cw, ch = measure(val, font("bold", 26), draw)
                draw_text(draw, (cx + (w - cw) // 2, ry + (row_h - ch) // 2 - 2), val, font("bold", 26), DANGER)
            else:
                cw, ch = measure(val, cell_font, draw)
                draw_text(draw, (cx + (w - cw) // 2, ry + (row_h - ch) // 2 - 2), val, cell_font, MUTED)
            cx += w

    # outer border
    draw.rounded_rectangle(
        [table_x, table_y, table_x + sum(col_widths), table_y + row_h * len(rows)],
        radius=16, outline=LINE, width=2,
    )
    # vertical separators between body rows
    cx = table_x
    for w in col_widths[:-1]:
        cx += w
        draw.line([(cx, table_y + row_h), (cx, table_y + row_h * len(rows))], fill=LINE, width=1)

    return canvas.convert("RGB")


def s13_roadmap(roadmap: Image.Image) -> Image.Image:
    return image_focus_slide(
        title="로드맵 — 음성 → 자율 기동 → 멀티테넌시",
        diagram=roadmap,
        caption="안정화 → 운영 자동화 → 확장 → 생태계.  분기 단위 마일스톤.",
    )


def s14_cta(logo: Image.Image) -> Image.Image:
    bg = gradient_bg(PAPER, PAPER2)
    bg = radial_glow((W // 2, int(H * 0.45)), (200, 136, 26, 60), 800, bg)
    canvas = bg.convert("RGBA")
    draw = ImageDraw.Draw(canvas)

    # mini logo top center
    mini = logo.copy()
    mini.thumbnail((150, 150), Image.LANCZOS)
    paste_image(canvas, mini, center=(W // 2, MARGIN_Y + 100))

    title_font = font("extrabold", 110)
    title = "함께 만들어 갈까요?"
    tw, th = measure(title, title_font, draw)
    draw_text(draw, ((W - tw) // 2, int(H * 0.30)), title, title_font, INK)

    # 4 CTA chips
    labels = ["데모 요청", "파일럿 도입", "기술 제휴", "투자 검토"]
    chip_font = font("bold", 28)
    pad = 36
    h_chip = 80
    widths = []
    for c in labels:
        cw, _ = measure(c, chip_font, draw)
        widths.append(cw + pad * 2)
    gap = 32
    total_w = sum(widths) + gap * (len(labels) - 1)
    x = (W - total_w) // 2
    chip_y = int(H * 0.55)
    for c, w in zip(labels, widths):
        draw.rounded_rectangle([x, chip_y, x + w, chip_y + h_chip], radius=40, fill=PRIMARY)
        cw, ch = measure(c, chip_font, draw)
        draw_text(draw, (x + (w - cw) // 2, chip_y + (h_chip - ch) // 2 - 4), c, chip_font, PAPER)
        x += w + gap

    tag_font = font("medium", 30)
    tag = "연구실 운영자에서 출발했습니다.  이제 시장 차례입니다."
    tw, th = measure(tag, tag_font, draw)
    draw_text(draw, ((W - tw) // 2, int(H * 0.74)), tag, tag_font, MUTED)

    # rule
    rule_y = int(H * 0.74) - 30
    draw.line([((W - 280) // 2, rule_y), ((W + 280) // 2, rule_y)], fill=PRIMARY, width=3)

    return canvas.convert("RGB")


def _eyebrow_title(draw: ImageDraw.ImageDraw, eyebrow: str, title: str):
    eb_font = font("bold", 26)
    draw_text(draw, (MARGIN_X, MARGIN_Y + 30), eyebrow, eb_font, PRIMARY)
    _, eh = measure(eyebrow, eb_font, draw)
    title_font = font("extrabold", 60)
    draw_text(draw, (MARGIN_X, MARGIN_Y + 30 + eh + 12), title, title_font, INK)
    _, th = measure(title, title_font, draw)
    rule_y = MARGIN_Y + 30 + eh + 12 + th + 22
    draw.line([(MARGIN_X, rule_y), (MARGIN_X + 240, rule_y)], fill=PRIMARY, width=4)


# ── Orchestration ─────────────────────────────────────────────────────


def main():
    OUT.mkdir(parents=True, exist_ok=True)

    logo = Image.open(IMG / "logo.png").convert("RGBA")
    architecture = Image.open(IMG / "architecture.png").convert("RGBA")
    dispatch = Image.open(IMG / "dispatch.png").convert("RGBA")
    fleet_topology = Image.open(IMG / "fleet-topology.png").convert("RGBA")
    roadmap = Image.open(IMG / "roadmap.png").convert("RGBA")
    signal_lights = Image.open(IMG / "signal-lights.png").convert("RGBA")

    @dataclass
    class SlideSpec:
        idx: int
        builder: callable
        chrome: bool = True

    builders = [
        SlideSpec(1,  lambda: s01_title(logo),                chrome=False),
        SlideSpec(2,  s02_section_problem,                    chrome=False),
        SlideSpec(3,  s03_problem),
        SlideSpec(4,  s04_definition),
        SlideSpec(5,  lambda: s05_architecture(architecture)),
        SlideSpec(6,  s06_section_diff,                       chrome=False),
        SlideSpec(7,  s07_diff_multivendor),
        SlideSpec(8,  lambda: s08_diff_routing(dispatch)),
        SlideSpec(9,  lambda: s09_diff_observability(signal_lights)),
        SlideSpec(10, lambda: s10_proof(fleet_topology)),
        SlideSpec(11, s11_global),
        SlideSpec(12, s12_compare),
        SlideSpec(13, lambda: s13_roadmap(roadmap)),
        SlideSpec(14, lambda: s14_cta(logo),                  chrome=False),
    ]
    total = len(builders)

    for spec in builders:
        img = spec.builder()
        if spec.chrome:
            page_chrome(img, spec.idx, total)
        out_path = OUT / f"slide-{spec.idx:02d}.png"
        img.save(out_path, "PNG", optimize=True)
        print(f"  ✓ {out_path.name}  ({out_path.stat().st_size // 1024} KB)")


if __name__ == "__main__":
    main()
