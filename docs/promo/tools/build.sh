#!/usr/bin/env bash
# wall-vault 피칭 덱 빌드 파이프라인.
#
# 1. signal-lights.png 합성 (PIL — bundled Pretendard)
# 2. Mermaid 다이어그램 → PNG (mermaid.ink — sandbox lacks Chromium)
# 3. PIL 로 14 장 슬라이드 PNG 직접 렌더 (full visual control, no font subs)
# 4. python-pptx 로 PNG 들을 풀블리드 래핑 → .pptx
#
# Pretendard OTF 5종이 docs/promo/fonts/ 에 번들돼 있어야 합니다 (Git
# 추적). 로컬에서 처음 빌드할 때 다운로드 스크립트를 한 번 돌리세요.

set -euo pipefail

cd "$(dirname "$0")/.."   # docs/promo/

echo "▶ 신호등 모킹 합성 (PIL · Pretendard)"
python3 tools/make_signal_mock.py

echo "▶ Mermaid 다이어그램 렌더 (mermaid.ink)"
python3 tools/render_mermaid.py

echo "▶ 슬라이드 14 장 PNG 렌더 (PIL · Pretendard)"
mkdir -p output/slides
python3 tools/render_slides.py

echo "▶ PPTX 래핑 (python-pptx)"
python3 tools/wrap_pptx.py

echo
echo "✓ 산출물:"
ls -lh output/
