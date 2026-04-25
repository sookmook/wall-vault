#!/usr/bin/env bash
# wall-vault 피칭 덱 빌드 파이프라인.
#
# 1. signal-lights.png 합성 (PIL)
# 2. Mermaid 다이어그램 → PNG (mermaid.ink — 이 환경엔 Chrome/Chromium 부재)
# 3. python-pptx 로 .pptx 직접 생성 (Marp 도 puppeteer 의존이라 동일 부재)
#
# extract_palette.py 는 손으로 정한 palette.css 를 덮어쓰므로 빌드에서 제외.
# 필요 시 별도로 실행하면 참고용 자동 추출 값을 palette.css 에 다시 기록함.

set -euo pipefail

cd "$(dirname "$0")/.."   # docs/promo/

echo "▶ 신호등 모킹 합성 (PIL)"
python3 tools/make_signal_mock.py

echo "▶ Mermaid 다이어그램 렌더 (mermaid.ink)"
python3 tools/render_mermaid.py

echo "▶ PPTX 컴파일 (python-pptx)"
mkdir -p output
python3 tools/build_pptx.py

echo
echo "✓ 산출물:"
ls -lh output/
