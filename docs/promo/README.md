# wall-vault 투자·제휴 피칭 덱

`output/wall-vault-pitch.pptx` 가 최종 산출물입니다. 한국어 단독, 16:9, 14
슬라이드(타이틀 + 섹션 구분 2 장 + 본문 11 장 + CTA), 마케팅 감성 톤.

## 빠른 빌드

```bash
tools/build.sh
```

세 단계가 순서대로 실행됩니다.

1. `tools/make_signal_mock.py` — Pillow 로 신호등 대시보드 mock 합성
2. `tools/render_mermaid.py` — Mermaid 다이어그램 4 종을 mermaid.ink 공개
   렌더러로 PNG 화 (이 환경엔 Chrome/Chromium 이 없어 mmdc 가 못 뜸)
3. `tools/build_pptx.py` — python-pptx 로 `.pptx` 직접 생성

`output/` 은 `.gitignore` 됩니다. 트래킹되는 자산은 본문 소스(`deck.md`),
다이어그램 소스(`diagrams/*.mmd`), 이미지 소스(`images/*.png`), 빌드 도구
(`tools/*.py`, `tools/build.sh`), 테마 (`theme.css`, `palette.css`) 입니다.

## 사전 도구

- Python 3.10+ + `python-pptx`, `Pillow`
- `npm i -g @marp-team/marp-cli @mermaid-js/mermaid-cli` (선택 — `deck.md`
  의 Marp 직접 컴파일에는 필요. 이 환경에선 puppeteer 의 Chrome 부재로
  실패하므로 build.sh 는 사용하지 않음)

## 미리보기

`.pptx` 를 그대로 PowerPoint / Google Slides / Keynote / LibreOffice
Impress 어디든 올려서 확인. PDF 변환이 필요하면 PowerPoint 의 "다른
형식으로 저장 → PDF" 를 사용하시거나 Google Slides 에 업로드 후
"파일 → 다운로드 → PDF" 가 가장 손쉽습니다.

## 콘텐츠 진실성 (외부 피칭 안전성)

`docs/promo/` 전체에 내부 IP·호스트명·토큰·인격명이 들어가지 않도록
설계되어 있습니다. `tools/build.sh` 마지막 단계에 grep 검증을 더 붙이고
싶으면 다음 한 줄로 충분합니다.

```bash
grep -rE "192\\.168\\.|SMPC|dhvmszmffh|c34cec|모토코|작순이|미니|라즈" docs/promo/ \
  | grep -v "Binary file" | grep -v "\\.png" \
  && exit 1 || true
```

빌드 산출물(`output/*.pptx`) 은 git 추적 대상이 아니므로 외부 공유 시
한 번 더 시각 검수해 주세요 — 이미지에 텍스트로 정보가 들어가지는
않지만, 추후 슬라이드 본문이 갱신되면 자동 검증의 대상으로 들어가야
합니다.
