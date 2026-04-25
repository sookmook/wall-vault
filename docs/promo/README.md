# wall-vault 투자·제휴 피칭 덱

`output/wall-vault-pitch.pptx` 가 최종 산출물입니다. 한국어 단독, 16:9, 14
슬라이드(타이틀 + 섹션 구분 2 장 + 본문 11 장 + CTA), 마케팅 감성 톤.

## 빠른 빌드

```bash
tools/build.sh
```

네 단계가 순서대로 실행됩니다.

1. `tools/make_signal_mock.py` — Pillow + 번들 Pretendard 로 신호등
   대시보드 mock 합성
2. `tools/render_mermaid.py` — Mermaid 다이어그램 4 종을 mermaid.ink
   공개 렌더러로 PNG 화 (sandbox 에 Chromium 부재라 mmdc 사용 불가)
3. `tools/render_slides.py` — PIL 로 14 장 슬라이드 PNG 직접 렌더
   (시각 통제 100%, 폰트 substitution 없음)
4. `tools/wrap_pptx.py` — python-pptx 로 PNG 들을 풀블리드 백그라운드로
   래핑한 `.pptx` 생성

> **트레이드오프**: PPTX 의 텍스트는 편집 불가능합니다 (각 슬라이드가
> 단일 PNG). 본문 source-of-truth 는 `deck.md` — 텍스트 수정은 거기서
> 하고 빌드를 다시 돌리시면 됩니다.

`output/` 은 `.gitignore` 됩니다. 트래킹되는 자산은 본문 소스(`deck.md`),
다이어그램 소스(`diagrams/*.mmd`), 정적 이미지 소스(`images/*.png` 중
로고만 — 나머지는 빌드 산출물이지만 빠른 시작을 위해 함께 추적), 빌드
도구(`tools/*.py`, `tools/build.sh`), 테마(`theme.css`, `palette.css`),
번들 폰트(`fonts/Pretendard-*.otf`) 입니다.

## 사전 의존성

- Python 3.10+ — `python-pptx`, `Pillow`
- 번들 폰트 (Pretendard 5 weights, MIT) — `fonts/` 에 이미 포함
- 인터넷 — `render_mermaid.py` 가 mermaid.ink 호출에 사용

## 미리보기

`.pptx` 를 그대로 PowerPoint / Google Slides / Keynote / LibreOffice
Impress 어디든 올려서 확인. PDF 가 필요하면 PowerPoint 의 "다른 형식으로
저장 → PDF" 를 쓰시거나 Google Slides 에 업로드 후 "파일 → 다운로드
→ PDF" 가 가장 손쉽습니다.

## 콘텐츠 진실성 (외부 피칭 안전성)

`docs/promo/` 전체에 내부 IP · 호스트명 · 토큰 · 인격명이 들어가지
않도록 설계돼 있습니다. 외부 공개 직전 검증:

```bash
grep -rE "192\\.168\\.|SMPC|dhvmszmffh|c34cec|모토코|작순이|미니|라즈" docs/promo/ \
  | grep -v "Binary file" | grep -v "\\.png" \
  && exit 1 || true
```
