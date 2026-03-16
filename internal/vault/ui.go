package vault

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sookmook/wall-vault/internal/i18n"
	"github.com/sookmook/wall-vault/internal/theme"
)

func buildDashboard(s *Server, t *theme.Theme) string {
	keys := s.store.ListKeys()
	clients := s.store.ListClients()
	proxies := s.store.ListProxies()
	services := s.store.ListServices()

	css := buildCSS(t)
	agentCard := buildAgentsCard(clients, proxies, services)
	// Merge active key IDs from all connected proxies
	activeKeys := make(map[string]string)
	for _, p := range proxies {
		for svc, keyID := range p.ActiveKeys {
			activeKeys[svc] = keyID
		}
	}
	keyCard := buildKeysCard(keys, services, activeKeys)
	svcCard := buildServicesCard(services)
	js := buildJS(t.Name, s.cfg.Lang, s.startedAt.Unix(), services, keys, s.cfg.Vault.AdminToken)

	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html lang="ko">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>벽금고(wall-vault) 대시보드</title>
<style>`)
	sb.WriteString(css)
	sb.WriteString(`</style>
</head>
<body>
<div class="topbar">
  <div class="topbar-brand">
    <img src="/logo" alt="wall-vault" class="topbar-logo">
    <span class="topbar-title">wall-vault</span>
  </div>
  <div class="topbar-controls">
    <div class="dropdown">
      <button class="dd-btn" id="dd-btn-lang" onclick="toggleDd(event,'dd-lang')">`)
	sb.WriteString(langLabel(s.cfg.Lang))
	sb.WriteString(` ▾</button>
      <div class="dd-menu" id="dd-lang">`)
	for _, code := range i18n.Supported {
		label := i18n.LangLabel(code)
		sb.WriteString(fmt.Sprintf(`        <div class="dd-item" data-val=%q onclick="setLang('%s')">%s</div>`+"\n", code, code, label))
	}
	sb.WriteString(`      </div>
    </div>
    <div class="dropdown">
      <button class="dd-btn" id="dd-btn-theme" onclick="toggleDd(event,'dd-theme')">`)
	sb.WriteString(themeLabel(t.Name))
	sb.WriteString(` ▾</button>
      <div class="dd-menu" id="dd-theme">`)
	for _, name := range theme.List() {
		sb.WriteString(fmt.Sprintf(`        <div class="dd-item" data-val=%q onclick="setTheme('%s')">%s</div>`+"\n", name, name, themeLabel(name)))
	}
	sb.WriteString(`      </div>
    </div>
    <span class="badge" id="sse-badge">● 연결 중...</span>
  </div>
</div>
<div class="header">
  <h1 id="page-title" data-i18n="title">벽금고(wall-vault) 대시보드</h1>
</div>
<div class="grid">`)
	sb.WriteString(agentCard)
	sb.WriteString(svcCard)
	sb.WriteString(keyCard)
	sb.WriteString(buildAddClientModal(services))
	sb.WriteString(buildAddKeyModal(services))
	sb.WriteString(buildEditClientModal(services))
	sb.WriteString(buildAddServiceModal())
	sb.WriteString(fmt.Sprintf(`</div>
<div class="footer">
  wall-vault %s — <a href="https://github.com/sookmook/wall-vault">github.com/sookmook/wall-vault</a>
  &nbsp;|&nbsp; <a href="https://sookmook.org/">sookmook.org</a>
  &nbsp;|&nbsp; <a href="mailto:sookmook@gmail.com">sookmook@gmail.com</a>
  &nbsp;|&nbsp; ⏱ <span id="uptime"></span>
</div>
<div class="sse-indicator" id="sse-status">SSE: 연결 중...</div>
<script>`, Version))
	sb.WriteString(js)
	sb.WriteString(`</script>
</body>
</html>`)
	return sb.String()
}

func langLabel(code string) string {
	label := i18n.LangLabel(code)
	if label != "" {
		return label
	}
	return i18n.LangLabel("ko")
}

// buildI18NJS: dynamically generate JS I18N and LANG_LABELS from locales/*.json
func buildI18NJS() string {
	var sb strings.Builder
	sb.WriteString("// ── I18N ──\nconst I18N={\n")
	allLangs := i18n.AllLangs()
	for idx, code := range i18n.Supported {
		m := allLangs[code]
		data, _ := json.Marshal(m)
		if idx < len(i18n.Supported)-1 {
			sb.WriteString(code + ":" + string(data) + ",\n")
		} else {
			sb.WriteString(code + ":" + string(data) + "\n")
		}
	}
	sb.WriteString("};\n")
	sb.WriteString("const LANG_LABELS={")
	parts := make([]string, 0, len(i18n.Supported))
	for _, code := range i18n.Supported {
		parts = append(parts, fmt.Sprintf("%q:%q", code, i18n.LangLabel(code)))
	}
	sb.WriteString(strings.Join(parts, ","))
	sb.WriteString("};\n")
	return sb.String()
}

func themeLabel(name string) string {
	m := map[string]string{
		"light": "☀️ light", "dark": "🌑 dark", "gold": "✨ gold",
		"cherry": "🌸 cherry", "ocean": "🌊 ocean",
		"autumn": "🍂 autumn", "winter": "❄️ winter",
	}
	if v, ok := m[name]; ok {
		return v
	}
	return "🎨 " + name
}

func buildCSS(t *theme.Theme) string {
	return `:root {` + t.CSSVars() + `}
*{box-sizing:border-box;margin:0;padding:0}
body{background:var(--bg);color:var(--text);font-family:-apple-system,BlinkMacSystemFont,'Segoe UI','Noto Sans',sans-serif;padding:0;min-height:100vh;font-size:14px}
a{color:var(--accent);text-decoration:none}
/* ── Topbar ── */
.topbar{position:sticky;top:0;z-index:500;display:flex;justify-content:space-between;align-items:center;padding:.5rem 1.5rem;background:var(--surface);border-bottom:1px solid var(--border);gap:.8rem;box-shadow:0 1px 5px rgba(0,0,0,.08)}
.topbar-brand{display:flex;align-items:center;gap:.5rem;flex-shrink:0}
.topbar-logo{height:38px;width:auto;flex-shrink:0;display:block}
.topbar-title{color:var(--accent);font-size:.92rem;font-weight:800;letter-spacing:1.2px;white-space:nowrap}
.topbar-controls{display:flex;align-items:center;gap:.5rem}
/* ── Dropdown ── */
.dropdown{position:relative}
.dd-btn{background:transparent;color:var(--text-muted);border:1px solid var(--border);padding:.22rem .7rem;border-radius:6px;cursor:pointer;font-size:.78rem;font-family:inherit;white-space:nowrap;transition:border-color .15s,color .15s}
.dd-btn:hover{color:var(--text);border-color:var(--accent)}
.dd-menu{display:none;position:absolute;right:0;top:calc(100% + 5px);background:var(--surface);border:1px solid var(--border);border-radius:8px;min-width:148px;z-index:9999;box-shadow:0 8px 24px rgba(0,0,0,.14);overflow:hidden}
.dd-menu.open{display:block}
.dd-item{padding:.35rem .9rem;font-size:.8rem;color:var(--text-muted);cursor:pointer;transition:background .1s}
.dd-item:hover{background:var(--bg);color:var(--text)}
.dd-item.active{color:var(--accent);font-weight:600}
/* ── Header (슬림) ── */
.header{display:flex;align-items:center;justify-content:center;padding:1rem 1.5rem;border-bottom:1px solid var(--border);background:var(--surface)}
.header h1{color:var(--text);font-size:1.5rem;font-weight:700;letter-spacing:.3px;white-space:nowrap}
/* ── Badge ── */
.badge{display:inline-block;background:var(--surface);border:1px solid var(--green);color:var(--green);padding:.12rem .55rem;border-radius:20px;font-size:.72rem;font-weight:600;letter-spacing:.3px}
/* ── Grid & Cards ── */
.grid{display:grid;grid-template-columns:repeat(2,1fr);gap:1.1rem;margin-bottom:1.2rem;padding:1.1rem 1.5rem}
.card{position:relative;z-index:2;background:var(--surface);border:1px solid var(--border);border-top:3px solid var(--accent);border-radius:10px;padding:1rem 1.1rem;box-shadow:0 1px 4px rgba(0,0,0,.05)}
.card-hdr{display:flex;align-items:center;justify-content:space-between;margin-bottom:.75rem;padding-bottom:.5rem;border-bottom:1px solid var(--border)}
.card-hdr h2{color:var(--accent);font-size:.82rem;font-weight:700;letter-spacing:.4px;text-transform:uppercase;display:flex;align-items:center;gap:.4rem}
.card-hdr h2 .count{color:var(--text-muted);font-size:.76rem;font-weight:400;text-transform:none;letter-spacing:0}
.card h2{color:var(--accent);font-size:.88rem;font-weight:700;margin-bottom:.75rem;padding-bottom:.55rem;border-bottom:1px solid var(--border);display:flex;justify-content:space-between;align-items:center;letter-spacing:.2px}
.card h2 .count{color:var(--text-muted);font-size:.76rem;font-weight:400}
/* ── 에이전트 섹션 (전체 폭) ── */
.agents-section{grid-column:1/-1;position:relative;z-index:2;padding-bottom:1.2rem;margin-bottom:.6rem;border-bottom:2px solid var(--border)}
.section-banner{grid-column:1/-1;display:flex;align-items:center;justify-content:space-between;background:var(--surface);border:1px solid var(--border);border-top:3px solid var(--accent);border-radius:10px;padding:.65rem 1.1rem;margin-bottom:.8rem;box-shadow:0 1px 4px rgba(0,0,0,.05)}
.section-banner h2{font-size:.82rem;font-weight:700;color:var(--accent);letter-spacing:.5px;text-transform:uppercase;display:flex;align-items:center;gap:.4rem}
.section-banner h2 .count{color:var(--text-muted);font-size:.76rem;font-weight:400;text-transform:none;letter-spacing:0}
.section-hdr{display:flex;align-items:center;gap:.5rem;margin-bottom:.75rem}
.section-hdr h2{color:var(--accent);font-size:.88rem;font-weight:700;display:flex;align-items:center;gap:.4rem;letter-spacing:.2px;flex:1}
.section-hdr h2 .count{color:var(--text-muted);font-size:.76rem;font-weight:400}
.agents-grid{display:grid;grid-template-columns:repeat(2,1fr);gap:.8rem}
/* ── 에이전트 개별 카드 ── */
.agent-card{background:var(--surface);border:1px solid var(--border);border-left:4px solid var(--accent);border-radius:10px;padding:1rem 1.1rem;display:flex;flex-direction:column;gap:.4rem;transition:box-shadow .18s,transform .18s;box-shadow:0 1px 4px rgba(0,0,0,.06)}
.agent-card:hover{box-shadow:0 4px 16px rgba(0,0,0,.12);transform:translateY(-1px)}
.agent-card.agent-disabled{border-left-color:var(--text-muted);opacity:.5}
.agent-card.ac-live{border-left-color:var(--green)}
.agent-card.ac-delay{border-left-color:var(--yellow)}
.agent-card.ac-offline{border-left-color:var(--red)}
.agent-card.ac-noconn{border-left-color:var(--text-muted)}
.ac-top{display:flex;align-items:flex-start;gap:.5rem}
.ac-type-icon{font-size:5.28rem;line-height:1;flex-shrink:0;margin-top:0;filter:drop-shadow(0 2px 6px rgba(0,0,0,.20))}
.ac-avatar{width:5.28rem;height:5.28rem;border-radius:50%;object-fit:cover;flex-shrink:0;border:2px solid var(--border)}
.ac-info{flex:1;min-width:0}
.ac-btns{display:flex;gap:.25rem;flex-shrink:0;margin-top:.05rem}
/* ── 기본 유틸 ── */
.row{display:flex;justify-content:space-between;align-items:center;margin:.3rem 0;font-size:.82rem;gap:.5rem}
.label{color:var(--text-muted);flex-shrink:0}
.val{color:var(--text);text-align:right;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
/* ── API 키 ── */
.key-item{margin:.45rem 0}
.key-header{display:flex;justify-content:space-between;align-items:center;font-size:.78rem;margin-bottom:.25rem}
.key-label{color:var(--text);font-weight:500}
.key-meta{color:var(--text-muted);font-size:.7rem}
.bar-track{background:var(--border);border-radius:4px;height:5px;overflow:hidden}
.bar-fill{height:5px;border-radius:4px;transition:width .4s ease}
.bar-green{background:var(--green)}
.bar-yellow{background:var(--yellow)}
.bar-red{background:var(--red)}
.bar-gray{background:var(--text-muted)}
.key-active .key-label{color:var(--accent)}
.key-active .bar-track{box-shadow:0 0 0 1px var(--accent)40}
/* ── 상태 점 ── */
.dot{width:7px;height:7px;border-radius:50%;flex-shrink:0}
.dot-green{background:var(--green);box-shadow:0 0 5px var(--green)}
.dot-yellow{background:var(--yellow);box-shadow:0 0 4px var(--yellow)}
.dot-red{background:var(--red);box-shadow:0 0 4px var(--red)}
.dot-gray{background:var(--border)}
/* ── 에이전트 카드 (레거시 호환 — agent-card에서 사용) ── */
.agent-name{font-size:.88rem;color:var(--text);font-weight:700;margin-bottom:.08rem;line-height:1.3;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.agent-live{font-size:.72rem;color:var(--green);margin-bottom:.3rem;font-weight:500}
.agent-type-badge{display:inline-block;font-size:.62rem;padding:.02rem .28rem;border-radius:4px;background:var(--accent);color:#fff;opacity:.75;margin-left:.35rem;vertical-align:middle;font-weight:600;letter-spacing:.3px}
.agent-desc{font-size:.72rem;color:var(--text-muted);margin:.1rem 0 .15rem;font-style:italic}
.agent-meta{font-size:.68rem;color:var(--text-muted);margin:.06rem 0}
.agent-uptime{font-size:.67rem;color:var(--text-muted);margin:.08rem 0;font-variant-numeric:tabular-nums}
/* ── 서비스 카드 ── */
.svc-item{padding:.42rem 0;border-bottom:1px solid var(--border)}
.svc-item:last-child{border-bottom:none}
/* ── SSE 인디케이터 ── */
.sse-indicator{position:fixed;bottom:.8rem;right:.8rem;font-size:.7rem;color:var(--text-muted);background:var(--surface);border:1px solid var(--border);padding:.22rem .6rem;border-radius:6px;z-index:400;box-shadow:0 1px 4px rgba(0,0,0,.08)}
/* ── 모델 폼 (카드 내 수직 배치, 기본 숨김) ── */
.model-form{display:none;flex-direction:column;gap:.3rem;border-top:1px solid var(--border);padding-top:.5rem;margin-top:.2rem}
.model-form.open{display:flex}
/* ── 에이전트 카드 액션 버튼 행 ── */
.ac-actions{display:flex;gap:.4rem;margin-top:.45rem}
.btn-action-wide{flex:1;background:transparent;border:1px solid var(--border);color:var(--text-muted);padding:.3rem .5rem;border-radius:6px;cursor:pointer;font-size:.73rem;font-family:inherit;transition:all .18s;white-space:nowrap}
.btn-action-wide:hover{color:var(--accent);border-color:var(--accent)}
.model-form-row{display:flex;gap:.3rem;align-items:center}
.model-form select{background:var(--bg);color:var(--text);border:1px solid var(--border);padding:.26rem .4rem;border-radius:6px;font-size:.75rem;font-family:inherit;flex:1;min-width:0;cursor:pointer}
.model-form input{background:var(--bg);color:var(--text);border:1px solid var(--border);padding:.26rem .4rem;border-radius:6px;font-size:.75rem;font-family:inherit;flex:1;min-width:0}
.model-form select:focus,.model-form input:focus{outline:none;border-color:var(--accent)}
/* ── 버튼 ── */
.btn{background:var(--accent);color:#fff;border:none;padding:.28rem .7rem;border-radius:6px;cursor:pointer;font-size:.76rem;font-family:inherit;font-weight:500;transition:background .15s,opacity .15s;white-space:nowrap}
.btn:hover{background:var(--accent-hover)}
.btn-sm{background:transparent;color:var(--accent);border:1px solid var(--accent);padding:.12rem .45rem;border-radius:5px;cursor:pointer;font-size:.7rem;font-family:inherit;margin-left:.4rem;font-weight:500;transition:all .15s}
.btn-sm:hover{background:var(--accent);color:#fff}
.btn-del{background:transparent;color:var(--text-muted);border:none;cursor:pointer;font-size:.72rem;padding:.05rem .3rem;line-height:1;border-radius:3px}
.btn-del:hover{color:var(--red)}
.btn-action{background:transparent;color:var(--text-muted);border:1px solid var(--border);cursor:pointer;font-size:.82rem;padding:.24rem .5rem;line-height:1.2;border-radius:5px;font-family:inherit;transition:all .15s}
.btn-action:hover{color:var(--accent);border-color:var(--accent);background:var(--bg)}
.btn-action-del:hover{color:var(--red);border-color:var(--red)}
/* ── 에이전트 종류 뱃지 (타입별 색상) ── */
.atbadge{display:inline-flex;align-items:center;font-size:.6rem;padding:.05rem .32rem;border-radius:4px;color:#fff;font-weight:600;opacity:.85;margin-left:.3rem;vertical-align:middle}
.atb-openclaw{background:#c0392b}
.atb-claude{background:#e07020}
.atb-cursor{background:#2471b0}
.atb-vscode{background:#2471b0}
.atb-gemini{background:#1a73e8}
.atb-custom{background:var(--text-muted)}
/* ── 에이전트 상태 행 ── */
.agent-status{font-size:.74rem;margin:.2rem 0 .3rem;display:flex;align-items:flex-start;gap:.5rem;flex-wrap:wrap}
.status-live{color:var(--green);font-weight:600}
.status-delay{color:var(--yellow)}
.status-offline{color:var(--red)}
.status-muted,.status-dc{color:var(--text-muted)}
.status-hint{color:var(--text-muted);font-size:.67rem;font-style:italic}
.status-version{color:var(--text-muted);font-size:.67rem}
/* ── 에이전트 설정 복사 버튼 (카드 하단, 전체 폭) ── */
.ac-cfg{margin-top:.1rem}
/* ── 설정 복사 버튼 ── */
.btn-cfg{display:flex;align-items:center;justify-content:center;gap:.28rem;width:100%;background:transparent;border:1px solid var(--border);color:var(--text-muted);padding:.28rem .55rem;border-radius:6px;cursor:pointer;font-size:.72rem;font-family:inherit;transition:all .15s;white-space:nowrap}
.btn-cfg:hover{background:var(--bg);color:var(--accent);border-color:var(--accent)}
.btn-cfg-openclaw:hover{color:#c0392b;border-color:#c0392b}
.btn-cfg-claude:hover{color:#e07020;border-color:#e07020}
/* ── 모델 저장 버튼 ── */
.btn-save{background:var(--accent);color:#fff;border:none;padding:.26rem .65rem;border-radius:6px;cursor:pointer;font-size:.75rem;font-family:inherit;font-weight:500;transition:all .15s;white-space:nowrap}
.btn-save:hover{background:var(--accent-hover);transform:translateY(-1px)}
/* ── 푸터 ── */
.footer{text-align:center;color:var(--text-muted);font-size:.7rem;padding:.7rem 1.5rem;border-top:1px solid var(--border)}
/* ── 모달 ── */
.modal-overlay{display:none;position:fixed;inset:0;background:rgba(0,0,0,.45);z-index:100;align-items:center;justify-content:center;backdrop-filter:blur(2px)}
.modal-overlay.open{display:flex}
.modal{background:var(--surface);border:1px solid var(--border);border-radius:12px;padding:1.5rem;min-width:340px;max-width:92vw;max-height:90vh;overflow-y:auto;box-shadow:0 16px 48px rgba(0,0,0,.2)}
.modal h3{color:var(--accent);margin-bottom:1rem;font-size:.95rem;font-weight:700}
.modal label{display:block;color:var(--text-muted);font-size:.75rem;font-weight:500;margin:.65rem 0 .2rem}
.modal input,.modal select{background:var(--bg);color:var(--text);border:1px solid var(--border);padding:.35rem .6rem;border-radius:6px;font-size:.82rem;width:100%;font-family:inherit;transition:border-color .15s}
.modal input:focus,.modal select:focus{outline:none;border-color:var(--accent)}
.modal-btns{display:flex;gap:.6rem;margin-top:1.1rem;justify-content:flex-end}
.msg{font-size:.76rem;margin-top:.5rem;min-height:1rem;color:var(--red)}
/* ── Cherry 꽃잎 — 애니메이션은 JS에서 고유 keyframe 생성 ── */
.cherry-petal{position:fixed;top:0;pointer-events:none;z-index:1;border-radius:50% 0 50% 0;background:radial-gradient(ellipse at 30% 20%,#ffe8f4 0%,#ff80b8 50%,#f01870 100%);box-shadow:0 0 6px #ff90c038}
/* ── Ocean 효과 ── */
@keyframes wave1{0%{transform:translateX(0)}100%{transform:translateX(-50%)}}
@keyframes wave2{0%{transform:translateX(-50%)}100%{transform:translateX(0)}}
@keyframes cloud-drift{0%{left:-260px;opacity:0}6%{opacity:.85}94%{opacity:.85}100%{left:112vw;opacity:0}}
@keyframes cloud-bob{0%,100%{transform:translateY(0) rotate(-1deg)}35%{transform:translateY(-24px) rotate(2deg)}68%{transform:translateY(-10px) rotate(-2deg)}}
@keyframes bubble-rise{0%{transform:translate(0,0);opacity:0}12%{opacity:.55}45%{transform:translate(16px,-40vh)}70%{transform:translate(-12px,-65vh)}92%{opacity:.35}100%{transform:translate(8px,-105vh);opacity:0}}
@keyframes ocean-sparkle{0%,100%{opacity:0;transform:scale(0)}50%{opacity:.8;transform:scale(1)}}
.ocean-fx{position:fixed;bottom:0;left:0;width:100%;height:160px;pointer-events:none;z-index:1;overflow:hidden}
.ocean-wave{position:absolute;left:-100%;width:300%;border-radius:42% 42% 0 0}
.ocean-cloud{position:fixed;pointer-events:none;z-index:1;white-space:nowrap;line-height:1}
.ocean-bubble{position:fixed;pointer-events:none;z-index:1;border-radius:50%;background:radial-gradient(circle at 32% 28%,rgba(255,255,255,.55),rgba(90,200,240,.18));border:1px solid rgba(90,200,240,.38)}
.ocean-sparkle{position:fixed;pointer-events:none;z-index:1;width:6px;height:6px;border-radius:50%;background:radial-gradient(circle,#90eeff,#0098d8);animation:ocean-sparkle ease-in-out infinite}
/* ── Gold 반짝임 ── */
@keyframes gold-twinkle{0%,100%{opacity:0;transform:scale(0) rotate(0deg)}45%{opacity:.25}50%{opacity:1;transform:scale(1) rotate(15deg)}55%{opacity:.25}}
@keyframes gold-drift{0%,100%{transform:translate(0,0)}25%{transform:translate(14px,-18px)}75%{transform:translate(-12px,14px)}}
.gold-spark{position:fixed;pointer-events:none;z-index:1;color:#c89000;text-shadow:0 0 6px #ffd70090,0 0 18px #ffaa0060}
/* ── Autumn 단풍잎 ── */
.autumn-leaf{position:fixed;top:0;pointer-events:none;z-index:1;font-size:18px;line-height:1;filter:sepia(1) hue-rotate(-20deg) saturate(1.6) brightness(0.55)}
/* ── Winter 뱅글벵글 ── */
@keyframes winter-spin{0%{transform:rotate(0deg)}100%{transform:rotate(360deg)}}
@keyframes winter-drift{0%{left:-8vw;opacity:0}5%{opacity:.9}95%{opacity:.9}100%{left:108vw;opacity:0}}
@keyframes winter-bob{0%,100%{transform:translateY(0)}50%{transform:translateY(-28px)}}
@keyframes snowfall{0%{transform:translateY(-10vh) translateX(0) rotate(0deg);opacity:0}5%{opacity:.8}95%{opacity:.8}100%{transform:translateY(110vh) translateX(40px) rotate(720deg);opacity:0}}
.winter-char{position:fixed;pointer-events:none;z-index:1;line-height:1;white-space:nowrap}
.snowflake{position:fixed;pointer-events:none;z-index:1;color:rgba(180,220,255,.8);font-size:14px;line-height:1}`
}

func buildJS(currentTheme, currentLang string, startedAt int64, services []*ServiceConfig, keys []*APIKey, adminToken string) string {
	// key count per service
	keyCounts := map[string]int{}
	for _, k := range keys {
		keyCounts[k.Service]++
	}
	// serialize service list as JS object (ID, Name, IsLocal, LocalURL, KeyCount)
	var svcJSParts []string
	for _, sv := range services {
		isLocal := "false"
		if sv.IsLocal() {
			isLocal = "true"
		}
		kc := keyCounts[sv.ID]
		svcJSParts = append(svcJSParts, fmt.Sprintf(`%q:{name:%q,local:%s,localUrl:%q,keyCount:%d}`, sv.ID, sv.Name, isLocal, sv.LocalURL, kc))
	}
	svcJSMap := "{" + strings.Join(svcJSParts, ",") + "}"

	return fmt.Sprintf(`const _SERVICES=%s;`+"\n", svcJSMap) +
		"const _SERVER_TOKEN=`" + adminToken + "`;\n" +
		buildI18NJS() + fmt.Sprintf(`
let curLang='ko';
function T(k){return(I18N[curLang]||I18N.ko)[k]||k;}
function applyLang(lang){
  curLang=lang;
  const t=I18N[lang]||I18N.ko;
  // data-i18n 텍스트
  document.querySelectorAll('[data-i18n]').forEach(el=>{const k=el.dataset.i18n;if(t[k]!==undefined)el.textContent=t[k];});
  // placeholders
  document.querySelectorAll('[data-i18n-ph]').forEach(el=>{const k=el.dataset.i18nPh;if(t[k]!==undefined)el.placeholder=t[k];});
  // title 속성
  document.querySelectorAll('[data-i18n-title]').forEach(el=>{const k=el.dataset.i18nTitle;if(t[k]!==undefined)el.title=t[k];});
  // 카운트 접미사 (3개/3件/3个/3)
  document.querySelectorAll('[data-i18n-cnt]').forEach(el=>{el.textContent=el.dataset.count+t.cnt;});
  // 봇 ago 텍스트 재포맷
  document.querySelectorAll('.bot-ago').forEach(el=>{
    const sec=parseInt(el.dataset.agoSec);
    el.textContent=sec<60?sec+t.ago_s:Math.floor(sec/60)+t.ago_m;
  });
  // SSE 상태 유지
  const badge=document.getElementById('sse-badge');
  if(badge){const isRun=badge.dataset.running==='1';badge.textContent=isRun?'● '+t.sse_run.slice(2):'● '+t.sse_conn.slice(2);}
  const sseEl=document.getElementById('sse-status');
  if(sseEl&&sseEl.dataset.state){
    const s=sseEl.dataset.state;
    sseEl.textContent=s==='ok'?t.sse_st_ok:s==='retry'?t.sse_st_retry:t.sse_st_conn;
  }
  // lang 버튼 업데이트
  document.getElementById('dd-btn-lang').textContent=(LANG_LABELS[lang]||lang)+' ▾';
  document.querySelectorAll('#dd-lang .dd-item').forEach(el=>el.classList.toggle('active',el.dataset.val===lang));
}

// 가동 시간 카운터
const SERVER_START = %d;
function updateUptime(){
  const sec = Math.floor(Date.now()/1000 - SERVER_START);
  const d = Math.floor(sec/86400);
  const h = Math.floor((sec%%86400)/3600);
  const m = Math.floor((sec%%3600)/60);
  const s = sec%%60;
  let txt = '';
  if(d>0) txt+=d+T('upd')+' ';
  if(d>0||h>0) txt+=h+T('uph')+' ';
  if(d>0||h>0||m>0) txt+=m+T('upm')+' ';
  txt+=s+T('ups');
  document.getElementById('uptime').textContent = txt;
}
setInterval(updateUptime, 1000); updateUptime();

// 에이전트별 가동 시간 카운터
function fmtAgentUptime(sec) {
  const t = I18N[curLang]||I18N.ko;
  const d=Math.floor(sec/86400), h=Math.floor((sec%%86400)/3600), m=Math.floor((sec%%3600)/60), s=sec%%60;
  let txt='';
  if(d>0) txt+=d+t.upd+' ';
  if(d>0||h>0) txt+=h+t.uph+' ';
  if(d>0||h>0||m>0) txt+=m+t.upm+' ';
  txt+=s+t.ups;
  return txt;
}
function updateAgentUptimes() {
  const now = Math.floor(Date.now()/1000);
  document.querySelectorAll('.agent-card[data-started-sec]').forEach(el => {
    const started = parseInt(el.dataset.startedSec);
    if (!started) return;
    let up = el.querySelector('.agent-uptime');
    if (!up) {
      up = document.createElement('div');
      up.className = 'agent-uptime';
      const status = el.querySelector('.agent-status');
      if (status) status.after(up); else {
        const name = el.querySelector('.agent-name');
        if (name) name.after(up);
      }
    }
    up.textContent = '⏱ ' + fmtAgentUptime(now - started);
  });
}
setInterval(updateAgentUptimes, 1000); updateAgentUptimes();

// 드롭다운 토글
function toggleDd(evt, id){
  evt.stopPropagation();
  const menu=document.getElementById(id);
  const isOpen=menu.classList.contains('open');
  document.querySelectorAll('.dd-menu').forEach(m=>m.classList.remove('open'));
  if(!isOpen) menu.classList.add('open');
}
document.addEventListener('click',()=>document.querySelectorAll('.dd-menu').forEach(m=>m.classList.remove('open')));

// 언어 변경
// LANG_LABELS is injected by buildI18NJS()
function setLang(lang){
  const tok=getAdminToken(); if(!tok) return;
  fetch('/admin/lang',{method:'PUT',headers:{'Content-Type':'application/json','Authorization':'Bearer '+tok},body:JSON.stringify({lang:lang})})
  .then(r=>r.json()).then(d=>{
    if(d.error){if(d.error==='unauthorized'){clearAdminToken();alert(T('err_token'));}else alert(T('err')+d.error);return;}
    applyLang(lang);
    document.getElementById('dd-lang').classList.remove('open');
  }).catch(e=>alert(T('err')+e));
}

// 테마 데이터`, startedAt) + `
const THEMES = {
  light:  {'--bg':'#f0f2f5','--surface':'#ffffff','--border':'#dde1e9','--text':'#1d2433','--text-muted':'#6b7385','--green':'#1a8a3c','--yellow':'#c47d00','--red':'#d93025','--blue':'#1a6fd4','--accent':'#5c5fc4','--accent-hover':'#4a4db0'},
  dark:   {'--bg':'#0d0d10','--surface':'#18181f','--border':'#2e2e3a','--text':'#e2e4ea','--text-muted':'#606470','--green':'#3dbf6a','--yellow':'#f5a623','--red':'#f05050','--blue':'#4da6ff','--accent':'#9b8afb','--accent-hover':'#b3a5ff'},
  gold:   {'--bg':'#faf8ee','--surface':'#fffff8','--border':'#e0cc70','--text':'#2a1e00','--text-muted':'#806a14','--green':'#2a8020','--yellow':'#b07000','--red':'#c03020','--blue':'#1850a0','--accent':'#a06800','--accent-hover':'#c07800'},
  cherry: {'--bg':'#fff4f7','--surface':'#fffbfd','--border':'#f0c0d0','--text':'#320820','--text-muted':'#9a5068','--green':'#0e9040','--yellow':'#c86800','--red':'#d81040','--blue':'#0e58b0','--accent':'#d8105e','--accent-hover':'#f0206e'},
  ocean:  {'--bg':'#e6f4ff','--surface':'#f8fdff','--border':'#70c4e8','--text':'#052038','--text-muted':'#26789a','--green':'#007858','--yellow':'#b06800','--red':'#c82828','--blue':'#0070b8','--accent':'#0086c8','--accent-hover':'#10a0e0'},
  autumn: {'--bg':'#ede0c8','--surface':'#f5e8d2','--border':'#b88848','--text':'#201008','--text-muted':'#6a3e18','--green':'#4a7810','--yellow':'#986000','--red':'#a82010','--blue':'#284070','--accent':'#b03008','--accent-hover':'#cc4818'},
  winter: {'--bg':'#f4f8ff','--surface':'#ffffff','--border':'#b8d4f0','--text':'#1a2840','--text-muted':'#5878a0','--green':'#087850','--yellow':'#d89800','--red':'#c82020','--blue':'#1868c0','--accent':'#1870d8','--accent-hover':'#3090f8'}
};

// ── 효과 요소 관리 ──
let fxElems = [];
function clearFx(){fxElems.forEach(e=>e.remove());fxElems=[];}

// ── Cherry: 꽃잎마다 고유 지그재그 keyframe ──
function createCherryFx(){
  clearFx();
  const N=30;
  const style=document.createElement('style');
  let css='';
  for(let i=0;i<N;i++){
    const steps=8;
    const amp=55+Math.random()*85;
    let kf='@keyframes pzz'+i+'{';
    for(let s=0;s<=steps;s++){
      const pct=(s/steps*100).toFixed(0);
      const y=(s/steps*110).toFixed(1);
      const x=((s%2===0)?1:-1)*amp*(0.55+Math.random()*0.45);
      const rz=(s*(780/steps)).toFixed(0);
      const ry=(s%2===0?35:-35);
      const op=s===0||s===steps?0:s===1?0.75:0.88;
      kf+=pct+'%{transform:translate('+x.toFixed(0)+'px,'+y+'vh) rotateZ('+rz+'deg) rotateY('+ry+'deg);opacity:'+op+'}';
    }
    kf+='}';
    css+=kf;
  }
  style.textContent=css;
  document.head.appendChild(style);
  fxElems.push(style);
  for(let i=0;i<N;i++){
    const p=document.createElement('div');
    p.className='cherry-petal';
    const sz=9+Math.random()*14;
    const dur=38+Math.random()*22;
    const startY=Math.random()*100;
    const br=Math.random()>0.5?'50% 0 50% 0':'0 50% 0 50%';
    p.style.cssText=
      'left:'+Math.random()*108+'vw;'+
      'width:'+sz+'px;height:'+(sz*1.4).toFixed(0)+'px;'+
      'border-radius:'+br+';'+
      'animation:pzz'+i+' '+dur+'s linear -'+(dur*startY/100).toFixed(1)+'s infinite;';
    document.body.appendChild(p);
    fxElems.push(p);
  }
}

// ── Ocean: 구름 두둥실 + 방울 천천히 ──
function createOceanFx(){
  clearFx();
  const wc=document.createElement('div');wc.className='ocean-fx';
  [['rgba(0,152,216,.28)',78,32,'wave1'],['rgba(0,120,190,.20)',56,48,'wave2'],['rgba(32,184,248,.14)',40,22,'wave1']].forEach(([col,h,spd,anim],i)=>{
    const w=document.createElement('div');w.className='ocean-wave';
    w.style.cssText='bottom:'+(i*20)+'px;background:'+col+';height:'+h+'px;animation:'+anim+' '+spd+'s linear infinite;';
    wc.appendChild(w);
  });
  document.body.appendChild(wc);fxElems.push(wc);
  ['☁','⛅','🌤','☁','🌥','☁'].forEach((e,i)=>{
    const c=document.createElement('div');c.className='ocean-cloud';
    const dd=80+i*22;
    const bd=12+Math.random()*10;
    c.textContent=e;
    c.style.cssText=
      'top:'+(1+i*7)+'vh;'+
      'font-size:'+(2.2+Math.random()*2.8)+'rem;'+
      'animation:cloud-drift '+dd+'s linear -'+(Math.random()*dd).toFixed(0)+'s infinite,'+
               'cloud-bob '+bd+'s ease-in-out -'+(Math.random()*bd).toFixed(1)+'s infinite;'+
      'opacity:.9;';
    document.body.appendChild(c);fxElems.push(c);
  });
  for(let i=0;i<10;i++){
    const b=document.createElement('div');b.className='ocean-bubble';
    const sz=5+Math.random()*12;
    const dur=32+Math.random()*22;
    b.style.cssText=
      'left:'+Math.random()*100+'vw;'+
      'bottom:'+Math.random()*20+'vh;'+
      'width:'+sz+'px;height:'+sz+'px;'+
      'animation:bubble-rise '+dur+'s ease-in-out -'+(Math.random()*dur).toFixed(1)+'s infinite;';
    document.body.appendChild(b);fxElems.push(b);
  }
  for(let i=0;i<10;i++){
    const s=document.createElement('div');s.className='ocean-sparkle';
    s.style.cssText=
      'left:'+Math.random()*100+'vw;bottom:'+(5+Math.random()*50)+'vh;'+
      'animation-duration:'+(4+Math.random()*5)+'s;'+
      'animation-delay:-'+Math.random()*8+'s;';
    document.body.appendChild(s);fxElems.push(s);
  }
}

// ── Gold: 별빛 반짝반짝 ──
function createGoldFx(){
  clearFx();
  const chars=['✦','✧','⋆','✶','★','✦','✧','⋆','✦'];
  for(let i=0;i<32;i++){
    const g=document.createElement('div');g.className='gold-spark';
    const sz=10+Math.random()*22;
    const td=6+Math.random()*10;
    const dd=20+Math.random()*20;
    g.textContent=chars[Math.floor(Math.random()*chars.length)];
    g.style.cssText=
      'left:'+Math.random()*100+'vw;top:'+Math.random()*100+'vh;'+
      'font-size:'+sz+'px;'+
      'animation:gold-twinkle '+td+'s ease-in-out -'+(Math.random()*td).toFixed(1)+'s infinite,'+
               'gold-drift '+dd+'s ease-in-out -'+(Math.random()*dd).toFixed(1)+'s infinite;';
    document.body.appendChild(g);fxElems.push(g);
  }
}

// ── Autumn: 단풍잎 솔솔 ──
function createAutumnFx(){
  clearFx();
  const leaves=['🍂','🍂','🍁','🍂','🍂'];
  const N=28;
  const style=document.createElement('style');
  let css='';
  for(let i=0;i<N;i++){
    const amp=40+Math.random()*70;
    const steps=7;
    let kf='@keyframes aleaf'+i+'{';
    for(let s=0;s<=steps;s++){
      const pct=(s/steps*100).toFixed(0);
      const y=(s/steps*110).toFixed(1);
      const x=((s%2===0)?1:-1)*amp*(0.5+Math.random()*0.5);
      const rz=(s*(540/steps)).toFixed(0);
      const op=s===0||s===steps?0:0.85;
      kf+=pct+'%{transform:translate('+x.toFixed(0)+'px,'+y+'vh) rotate('+rz+'deg);opacity:'+op+'}';
    }
    kf+='}';
    css+=kf;
  }
  style.textContent=css;
  document.head.appendChild(style);fxElems.push(style);
  for(let i=0;i<N;i++){
    const l=document.createElement('div');
    l.className='autumn-leaf';
    const dur=32+Math.random()*20;
    const startY=Math.random()*100;
    l.textContent=leaves[Math.floor(Math.random()*leaves.length)];
    l.style.cssText=
      'left:'+Math.random()*108+'vw;'+
      'font-size:'+(14+Math.random()*12)+'px;'+
      'animation:aleaf'+i+' '+dur+'s linear -'+(dur*startY/100).toFixed(1)+'s infinite;';
    document.body.appendChild(l);fxElems.push(l);
  }
}

// ── Winter: 눈사람+트리 뱅글벵글 + 눈송이 ──
function createWinterFx(){
  clearFx();
  const chars=['☃️','🎄','❄️','⛄','🎄','☃️','❄️'];
  chars.forEach((ch,i)=>{
    const c=document.createElement('div');
    c.className='winter-char';
    const sz=1.8+Math.random()*2.2;
    const top=5+i*12;
    const dd=60+i*18;
    const bd=4+Math.random()*5;
    const sd=-(Math.random()*dd).toFixed(1);
    c.textContent=ch;
    c.style.cssText=
      'top:'+top+'vh;'+
      'font-size:'+sz+'rem;'+
      'animation:winter-drift '+dd+'s linear '+sd+'s infinite,'+
               'winter-bob '+bd+'s ease-in-out -'+(Math.random()*bd).toFixed(1)+'s infinite;';
    document.body.appendChild(c);fxElems.push(c);
  });
  // 회전하는 눈결정
  for(let i=0;i<6;i++){
    const s=document.createElement('div');
    s.className='winter-char';
    s.textContent='❄';
    const sz=0.9+Math.random()*1.2;
    const top=10+i*14;
    const spd=3+Math.random()*4;
    s.style.cssText=
      'top:'+top+'vh;left:'+(5+i*16)+'vw;'+
      'font-size:'+sz+'rem;color:rgba(140,190,255,.7);'+
      'animation:winter-spin '+spd+'s linear infinite;';
    document.body.appendChild(s);fxElems.push(s);
  }
  // 눈송이 내리기
  for(let i=0;i<20;i++){
    const f=document.createElement('div');
    f.className='snowflake';
    f.textContent='❅';
    const dur=8+Math.random()*12;
    f.style.cssText=
      'left:'+Math.random()*100+'vw;'+
      'font-size:'+(10+Math.random()*10)+'px;'+
      'animation:snowfall '+dur+'s linear -'+(Math.random()*dur).toFixed(1)+'s infinite;';
    document.body.appendChild(f);fxElems.push(f);
  }
}

function applyThemeCss(name){
  const vars=THEMES[name]; if(!vars) return;
  const root=document.documentElement;
  for(const [k,v] of Object.entries(vars)) root.style.setProperty(k,v);
  document.querySelectorAll('.theme-btn').forEach(b=>b.classList.toggle('active',b.id==='theme-'+name));
  if(name==='cherry') createCherryFx();
  else if(name==='ocean') createOceanFx();
  else if(name==='gold') createGoldFx();
  else if(name==='autumn') createAutumnFx();
  else if(name==='winter') createWinterFx();
  else clearFx();
}
const THEME_LABELS={'light':'☀️ light','dark':'🌑 dark','gold':'✨ gold','cherry':'🌸 cherry','ocean':'🌊 ocean','autumn':'🍂 autumn','winter':'❄️ winter'};
function setTheme(name){
  const tok=getAdminToken(); if(!tok) return;
  fetch('/admin/theme',{method:'PUT',headers:{'Content-Type':'application/json','Authorization':'Bearer '+tok},body:JSON.stringify({theme:name})})
  .then(r=>r.json()).then(d=>{
    if(d.error){if(d.error==='unauthorized'){clearAdminToken();alert(T('err_token'));}else alert(T('err')+d.error);return;}
    applyThemeCss(name);
    document.getElementById('dd-btn-theme').textContent=(THEME_LABELS[name]||name)+' ▾';
    document.querySelectorAll('#dd-theme .dd-item').forEach(el=>el.classList.toggle('active',el.dataset.val===name));
    document.getElementById('dd-theme').classList.remove('open');
  }).catch(e=>alert(T('err')+e));
}
// 초기화
applyThemeCss('` + currentTheme + `');
applyLang('` + currentLang + `');
document.querySelectorAll('#dd-theme .dd-item').forEach(el=>el.classList.toggle('active',el.dataset.val==='` + currentTheme + `'));

// Key state cache: id → {baseText, cdMs}  — used by 1-second countdown ticker
let _keyCache = {};

// Format cooldown countdown string from timestamp (ms)
function _fmtCountdown(untilMs) {
  const rem = Math.floor((untilMs - Date.now()) / 1000);
  if (rem <= 0) return null;
  const m = Math.floor(rem / 60), s = rem % 60;
  return m > 0 ? m + T('upm') + ' ' + s + T('ups') : s + T('ups');
}

// 1-second ticker: update cooldown countdowns with no network request
function _tickCooldowns() {
  const now = Date.now();
  for (const id in _keyCache) {
    const c = _keyCache[id];
    if (!c.cdMs) continue;
    const el = document.querySelector('[data-key-id="' + id + '"]');
    if (!el) continue;
    const meta = el.querySelector('.key-meta');
    const bar  = el.querySelector('.bar-fill');
    if (c.cdMs > now) {
      const cd = _fmtCountdown(c.cdMs);
      if (meta) meta.textContent = c.baseText + (cd ? ' (' + cd + ')' : '');
      if (bar)  bar.className = 'bar-fill bar-yellow';
    } else {
      // cooldown expired — restore to green
      if (meta) meta.textContent = c.baseText;
      if (bar)  bar.className = 'bar-fill bar-green';
      _keyCache[id].cdMs = 0;
    }
  }
}
setInterval(_tickCooldowns, 1000);

// Refresh key usage directly from SSE usage_update payload — no fetch needed
function refreshKeyUsage(data) {
  if (!data || !Array.isArray(data.keys)) return;
  const now = Date.now();
  // Pre-pass: for unlimited keys compute the SUM of activity per service.
  // Bar width = key_activity / service_total * 100, so each key shows its
  // SHARE of today's requests. This avoids the "all-100% or all-0%" problem
  // that occurs with max-relative scaling when keys have equal usage.
  // Activity = max(today_attempts, today_usage) so rate-limited-only keys count.
  const svcSum = {};
  data.keys.forEach(k => {
    if (k.daily_limit === 0) {
      const ref = Math.max(k.today_attempts || 0, k.today_usage || 0);
      svcSum[k.service] = (svcSum[k.service] || 0) + ref;
    }
  });
  data.keys.forEach(k => {
    const el = document.querySelector('[data-key-id="' + k.id + '"]');
    if (!el) return;
    const bar  = el.querySelector('.bar-fill');
    const meta = el.querySelector('.key-meta');
    const cdMs = k.cooldown_until ? new Date(k.cooldown_until).getTime() : 0;
    const onCooldown = cdMs > now;
    const usage    = k.today_usage    || 0;
    const attempts = k.today_attempts || 0;
    const exhausted = k.daily_limit > 0 && usage >= k.daily_limit;
    // update progress bar
    if (bar) {
      let pct = 0;
      if (k.daily_limit > 0) {
        // absolute scaling against daily limit
        pct = Math.min(100, Math.round(usage * 100 / k.daily_limit));
        if (pct === 0 && usage > 0) pct = 4;
      } else {
        // share-of-total: bar shows this key's fraction of today's service requests
        const ref = Math.max(attempts, usage);
        if (svcSum[k.service] > 0 && ref > 0) {
          pct = Math.min(100, Math.round(ref * 100 / svcSum[k.service]));
          if (pct === 0) pct = 4;
        }
      }
      bar.style.width = pct + '%';
      bar.className = 'bar-fill ' + (onCooldown ? 'bar-yellow' : (exhausted || pct >= 97 ? 'bar-red' : 'bar-green'));
    }
    // update meta text
    // Format: "N/limit" | "N req" | "N/limit (M att)" | "N req (M att)" | "M att"
    let baseText;
    if (k.daily_limit > 0) {
      baseText = usage + '/' + k.daily_limit;
      if (attempts > usage) baseText += ' (' + attempts + ' ' + T('key_att') + ')';
    } else if (usage > 0) {
      baseText = usage + ' ' + T('key_reqs');
      if (attempts > usage) baseText += ' (' + attempts + ' ' + T('key_att') + ')';
    } else if (attempts > 0) {
      // all requests rate-limited; show attempt count so it's not blank
      baseText = attempts + ' ' + T('key_att');
    } else {
      baseText = '0 ' + T('key_reqs');
    }
    if (meta) {
      const cd = onCooldown ? _fmtCountdown(cdMs) : null;
      meta.textContent = baseText + (cd ? ' (' + cd + ')' : '');
    }
    // store in cache for 1-second countdown ticker
    _keyCache[k.id] = {baseText, cdMs: onCooldown ? cdMs : 0};
  });
}

// SSE 연결
let es;
function connectSSE() {
  es = new EventSource('/api/events');
  es.onopen = () => {
    const t=I18N[curLang]||I18N.ko;
    const st=document.getElementById('sse-status');
    st.textContent=t.sse_st_ok; st.dataset.state='ok';
    const badge=document.getElementById('sse-badge');
    badge.textContent=t.sse_run; badge.dataset.running='1';
    badge.style.borderColor='var(--green)'; badge.style.color='var(--green)';
  };
  es.onmessage = (e) => {
    try {
      const d = JSON.parse(e.data);
      if (d.type === 'config_change') {
        // 에이전트 카드 서비스/모델 필드 즉시 반영 후 드롭다운 갱신
        const cd = d.data || {};
        if (cd.client_id) applyAgentConfigChange(cd.client_id, cd.service, cd.model);
        else refreshModelDropdowns();
      } else if (d.type === 'service_changed') {
        // 서비스 추가/삭제 → 체크 후 에이전트 메뉴 전체 갱신
        autoCheckServices().then(() => refreshServiceSelects()).then(() => refreshModelDropdowns());
      } else if (d.type === 'key_added' || d.type === 'key_deleted') {
        // 키 변경 → 서비스 활성 상태 재판정 후 에이전트 메뉴 갱신
        autoCheckServices().then(() => refreshServiceSelects()).then(() => refreshModelDropdowns());
      } else if (d.type === 'usage_reset') {
        // 가벼운 reload (일일 사용량 초기화)
        setTimeout(() => location.reload(), 500);
      } else if (d.type === 'usage_update') {
        // 키 사용량 실시간 갱신 (하트비트 동기화)
        refreshKeyUsage(d.data || {});
      }
    } catch {}
  };
  es.onerror = () => {
    const t=I18N[curLang]||I18N.ko;
    const st=document.getElementById('sse-status');
    st.textContent=t.sse_st_retry; st.dataset.state='retry';
    es.close();
    setTimeout(connectSSE, 3000);
  };
}
connectSSE();

// 서비스 select 갱신: /admin/services에서 현재 목록 받아 모든 svc-* select 업데이트
function refreshServiceSelects() {
  const tok = localStorage.getItem('wv_admin_token')||'';
  return fetch('/admin/services', {headers:{'Authorization':'Bearer '+tok}})
  .then(r=>r.json()).then(svcs=>{
    if(!Array.isArray(svcs)) return;
    // only proxy-enabled services are shown in agent model dropdowns
    const enabled = svcs.filter(s=>s.enabled && s.proxy_enabled);
    function _svcOpts(prev) {
      return enabled.map(s=>'<option value="'+s.id+'"'+(s.id===prev?' selected':'')+'>'+( s.name||s.id)+'</option>').join('');
    }
    // 에이전트 카드의 svc-* select 갱신
    document.querySelectorAll('.agent-svc-sel').forEach(sel=>{
      const prev = sel.value;
      sel.innerHTML = _svcOpts(prev);
      if(!sel.value && enabled.length) sel.value = enabled[0].id;
    });
    // 모달의 서비스 select 갱신
    ['ac-service','ec-service','ak-service'].forEach(id=>{
      const sel=document.getElementById(id);
      if(!sel) return;
      const prev=sel.value;
      sel.innerHTML = '<option value="">\u2014 \uc120\ud0dd \u2014</option>' + _svcOpts(prev);
    });
  }).catch(()=>{});
}

// 모델 드롭다운 갱신: 각 에이전트 카드의 현재 서비스로 모델 목록 재조회
function refreshModelDropdowns() {
  document.querySelectorAll('.agent-svc-sel').forEach(el=>{
    const id = el.id.replace('svc-','');
    if(id && el.value) onAgentServiceChange('mdl-'+id,'mdl-sel-'+id, el.value);
  });
}

// 에이전트 편집 시 원본 데이터 보관 (변경된 필드만 PUT 전송하기 위함)
let _ecOrig = {};

// Admin Token 헬퍼 (서버 내장 토큰 우선, 없으면 localStorage, 그것도 없으면 prompt)
function getAdminToken() {
  if (typeof _SERVER_TOKEN !== 'undefined' && _SERVER_TOKEN) return _SERVER_TOKEN;
  let token = localStorage.getItem('wv_admin_token');
  if (!token) {
    token = prompt(T('admin_prompt'));
    if (!token) return null;
    localStorage.setItem('wv_admin_token', token);
  }
  return token;
}
function clearAdminToken() { localStorage.removeItem('wv_admin_token'); }

// 키 추가 모달
function openAddKey() {
  document.getElementById('modal-addkey').classList.add('open');
  document.getElementById('ak-msg').textContent = '';
  document.getElementById('ak-key').value = '';
  document.getElementById('ak-label').value = '';
  document.getElementById('ak-limit').value = '0';
}
function closeAddKey() {
  document.getElementById('modal-addkey').classList.remove('open');
}
function submitAddKey() {
  const token = getAdminToken();
  if (!token) return;
  const svc = document.getElementById('ak-service').value;
  const key = document.getElementById('ak-key').value.trim();
  const label = document.getElementById('ak-label').value.trim();
  const limit = parseInt(document.getElementById('ak-limit').value) || 0;
  if (!key) { document.getElementById('ak-msg').textContent = T('err_key'); return; }
  document.getElementById('ak-msg').textContent = T('adding');
  fetch('/admin/keys', {
    method: 'POST',
    headers: {'Content-Type':'application/json','Authorization':'Bearer '+token},
    body: JSON.stringify({service:svc, key:key, label:label, daily_limit:limit})
  }).then(r => r.json()).then(d => {
    if (d.error) {
      if (d.error === 'unauthorized') { clearAdminToken(); alert(T('err_token')); }
      else document.getElementById('ak-msg').textContent = T('err')+d.error;
    } else {
      closeAddKey();
      setTimeout(() => location.reload(), 500);
    }
  }).catch(e => { document.getElementById('ak-msg').textContent = T('err')+e; });
}

// 키 삭제
function deleteKey(id) {
  if (!confirm(T('del_key'))) return;
  const token = getAdminToken();
  if (!token) return;
  fetch('/admin/keys/'+id, {
    method: 'DELETE',
    headers: {'Authorization':'Bearer '+token}
  }).then(r => r.json()).then(d => {
    if (d.error) {
      if (d.error === 'unauthorized') { clearAdminToken(); alert(T('err_token')); }
      else alert(T('err')+d.error);
    } else { location.reload(); }
  });
}

// ── 에이전트 카드 실시간 반영 (config_change SSE) ──
// 에이전트 카드의 서비스 select, 모델 input을 새 값으로 업데이트하고 모델 드롭다운 재조회
function applyAgentConfigChange(clientId, service, model) {
  const svcSel = document.getElementById('svc-'+clientId);
  const mdlInp = document.getElementById('mdl-'+clientId);
  const mdlSel = document.getElementById('mdl-sel-'+clientId);
  // 서비스 select 업데이트
  if (svcSel && service) {
    svcSel.value = service;
    if (svcSel.value !== service) {
      // select에 해당 옵션이 없으면 목록 갱신 후 재적용
      refreshServiceSelects().then(() => { if (svcSel) svcSel.value = service; });
    }
  }
  // 모델 input 업데이트
  if (mdlInp && model) {
    mdlInp.value = model;
    if (mdlSel) mdlSel.value = '';
  }
  // 모델 드롭다운 재조회
  const curSvc = (svcSel && svcSel.value) || service;
  if (curSvc) onAgentServiceChange('mdl-'+clientId, 'mdl-sel-'+clientId, curSvc);
  // 해당 에이전트 카드의 status-live 텍스트 업데이트 (실행 중인 경우)
  if (svcSel && service && model) {
    const card = svcSel.closest('.agent-card');
    if (card) {
      const live = card.querySelector('.status-live');
      if (live) {
        const parts = live.textContent.split('—');
        if (parts.length >= 2) {
          live.textContent = parts[0] + '— ' + service + ' / ' + model;
        }
      }
    }
  }
}

// ── 서비스 자동 체크 ──
// 클라우드: 키 없는 서비스 자동 비활성화 / 키 있으면 자동 활성화
// 로컬: /admin/models 프로브 결과에 따라 활성/비활성
async function _setSvcEnabled(id, enabled, token) {
  const urlEl = document.getElementById('svc-url-'+id);
  try {
    await fetch('/admin/services/'+id, {
      method: 'PUT',
      headers: {'Content-Type':'application/json','Authorization':'Bearer '+token},
      body: JSON.stringify({enabled, local_url: urlEl ? urlEl.value : ''})
    });
    const cb = document.getElementById('svc-en-'+id);
    if (cb) cb.checked = enabled;
  } catch {}
}
async function _checkLocalSvc(id, svcEnabled, token) {
  const cb = document.getElementById('svc-en-'+id);
  try {
    const resp = await fetch('/admin/models?service='+id, {headers:{'Authorization':'Bearer '+token}});
    const up = resp.ok && (((await resp.json()).models)||[]).length > 0;
    if (up !== svcEnabled) await _setSvcEnabled(id, up, token);
    else if (cb) cb.checked = up;
  } catch { if (svcEnabled) await _setSvcEnabled(id, false, token); }
}
// 호출할 때마다 최신 키/서비스 현황을 서버에서 조회해 처리
async function autoCheckServices() {
  const token = localStorage.getItem('wv_admin_token');
  if (!token) return;
  try {
    const [kr, sr] = await Promise.all([
      fetch('/admin/keys',     {headers:{'Authorization':'Bearer '+token}}),
      fetch('/admin/services', {headers:{'Authorization':'Bearer '+token}})
    ]);
    if (!kr.ok || !sr.ok) return;
    const [keys, svcs] = await Promise.all([kr.json(), sr.json()]);
    const keyCounts = {};
    (keys||[]).forEach(k => { keyCounts[k.service] = (keyCounts[k.service]||0)+1; });
    const checks = [];
    for (const svc of (svcs||[])) {
      const meta = _SERVICES[svc.id];
      const isLocal = meta ? meta.local : false;
      if (isLocal) {
        checks.push(_checkLocalSvc(svc.id, svc.enabled, token));
      } else {
        const want = (keyCounts[svc.id]||0) > 0;
        if (svc.enabled !== want) checks.push(_setSvcEnabled(svc.id, want, token));
      }
    }
    await Promise.all(checks);
  } catch {}
}

// ── 에이전트 카드 모델 목록 초기화 (페이지 로드 시) ──
document.addEventListener('DOMContentLoaded', function() {
  autoCheckServices().then(() => refreshServiceSelects()).then(() => refreshModelDropdowns());
});

// ── 모달 공통 유틸 ──
function closeModal(prefix) {
  document.getElementById('modal-'+prefix).classList.remove('open');
}
function _readClientForm(prefix) {
  const ipwlRaw = document.getElementById(prefix+'-ipwl').value.trim();
  const avatarEl = document.getElementById(prefix+'-avatar');
  return {
    id:      document.getElementById(prefix+'-id').value.trim(),
    name:    document.getElementById(prefix+'-name').value.trim(),
    token:   document.getElementById(prefix+'-token').value.trim(),
    default_service: document.getElementById(prefix+'-service').value,
    default_model:   document.getElementById(prefix+'-mdl').value.trim(),
    agent_type:  document.getElementById(prefix+'-agent-type').value,
    work_dir:    document.getElementById(prefix+'-workdir').value.trim(),
    description: document.getElementById(prefix+'-desc').value.trim(),
    ip_whitelist: ipwlRaw ? ipwlRaw.split(',').map(s=>s.trim()).filter(s=>s) : [],
    avatar:  avatarEl ? (avatarEl.value||'') : '',
    enabled: document.getElementById(prefix+'-enabled').checked,
  };
}
function _clearClientForm(prefix) {
  ['id','name','token','mdl','workdir','desc','ipwl'].forEach(f => {
    const el=document.getElementById(prefix+'-'+f);if(el)el.value='';
  });
  const at=document.getElementById(prefix+'-agent-type');if(at)at.value='';
  const en=document.getElementById(prefix+'-enabled');if(en)en.checked=true;
  const msg=document.getElementById(prefix+'-msg');if(msg)msg.textContent='';
  const ms=document.getElementById(prefix+'-mdl-sel');
  if(ms) ms.innerHTML='<option value="">'+T('sel_model_or_enter')+'</option>';
  const av=document.getElementById(prefix+'-avatar');if(av)av.value='';
  const avp=document.getElementById(prefix+'-avatar-preview');if(avp){avp.src='';avp.style.display='none';}
}
// 아바타 파일 → base64 data URI 변환 후 hidden input에 저장
function loadAvatarPreview(fileInput, hiddenId, previewId) {
  const file = fileInput.files[0];
  if (!file) return;
  const reader = new FileReader();
  reader.onload = e => {
    const data = e.target.result;
    document.getElementById(hiddenId).value = data;
    const prev = document.getElementById(previewId);
    if (prev) { prev.src = data; prev.style.display = 'block'; }
  };
  reader.readAsDataURL(file);
}

// ── OpenClaw 설정 복사 ──
function copyOpenClawConfig(clientId) {
  const token = getAdminToken(); if(!token) return;
  fetch('/admin/clients/'+clientId, {headers:{'Authorization':'Bearer '+token}})
  .then(r=>r.json()).then(c=>{
    if(c.error){alert(T('err')+c.error);return;}
    const svc = c.default_service||'google';
    const mdl = c.default_model||'gemini-2.5-flash';
    const proxyPort = location.port==='56243' ? '56244' : '56244';
    const baseUrl = location.protocol+'//'+location.hostname+':'+proxyPort+'/v1';
    const tok = c.token||'YOUR_AGENT_TOKEN';
    const cfg = '// Add to ~/.openclaw/openclaw.json\n'
      + '{\n'
      + '  models: {\n'
      + '    providers: {\n'
      + '      "wall-vault": {\n'
      + '        baseUrl: "' + baseUrl + '",\n'
      + '        apiKey: "' + tok + '",\n'
      + '        api: "openai-completions",\n'
      + '        models: [ { id: "wall-vault/' + mdl + '" } ]\n'
      + '      }\n'
      + '    }\n'
      + '  },\n'
      + '  agents: { defaults: { model: { primary: "wall-vault/' + mdl + '" } } }\n'
      + '}';
    navigator.clipboard.writeText(cfg).then(()=>{
      alert(T('cfg_openclaw')+'\n'+T('cfg_ok')+'\n'+T('cfg_openclaw_hint'));
    }).catch(()=>{
      prompt(T('cfg_manual'), cfg);
    });
  }).catch(e=>alert(T('err')+e));
}

// ── 에이전트 타입별 설정 복사 (claude-code / cursor / vscode) ──
function copyAgentConfig(clientId, agentType) {
  const token = getAdminToken(); if(!token) return;
  fetch('/admin/clients/'+clientId, {headers:{'Authorization':'Bearer '+token}})
  .then(r=>r.json()).then(c=>{
    if(c.error){alert(T('err')+c.error);return;}
    const baseUrl = location.protocol+'//'+location.hostname+':56244/v1';
    const tok = c.token||'YOUR_AGENT_TOKEN';
    const mdl = c.default_model||'gemini-2.5-flash';
    let cfg='', title='', hint='';
    if(agentType==='claude-code'){
      title=T('cfg_claude')+' '+T('cfg_ok');
      hint=T('cfg_claude_hint');
      cfg='// ~/.claude/settings.json\n'
        +'{\n'
        +'  "apiProvider": "openai",\n'
        +'  "baseUrl": "'+baseUrl+'",\n'
        +'  "apiKey": "'+tok+'"\n'
        +'}';
    } else if(agentType==='cursor'){
      title=T('cfg_cursor')+' '+T('cfg_ok');
      hint=T('cfg_cursor_hint');
      cfg='// Cursor: Settings > AI > OpenAI API\n'
        +'Base URL : '+baseUrl+'\n'
        +'API Key  : '+tok+'\n\n'
        +'// or environment variable:\n'
        +'OPENAI_BASE_URL='+baseUrl+'\n'
        +'OPENAI_API_KEY='+tok;
    } else if(agentType==='vscode'){
      title=T('cfg_vscode')+' '+T('cfg_ok');
      hint=T('cfg_vscode_hint');
      cfg='// ~/.continue/config.json  (Continue extension)\n'
        +'{\n'
        +'  "models": [{\n'
        +'    "title": "wall-vault proxy",\n'
        +'    "provider": "openai",\n'
        +'    "model": "'+mdl+'",\n'
        +'    "apiBase": "'+baseUrl+'",\n'
        +'    "apiKey": "'+tok+'"\n'
        +'  }]\n'
        +'}';
    } else if(agentType==='gemini-cli'){
      const geminiBase = location.protocol+'//'+location.hostname+':56244';
      title=T('cfg_gemini_cli')+' '+T('cfg_ok');
      hint=T('cfg_gemini_cli_hint');
      cfg='# set via environment variable (paste in terminal):\n'
        +'export GEMINI_API_BASE_URL='+geminiBase+'\n'
        +'export GEMINI_API_KEY='+tok+'\n\n'
        +'# or add to ~/.gemini/settings.json:\n'
        +'{\n'
        +'  "apiBaseUrl": "'+geminiBase+'",\n'
        +'  "apiKey": "'+tok+'"\n'
        +'}';
    } else if(agentType==='antigravity'){
      const geminiBase = location.protocol+'//'+location.hostname+':56244';
      title=T('cfg_antigravity')+' '+T('cfg_ok');
      hint=T('cfg_antigravity_hint');
      cfg='# set via environment variable:\n'
        +'export GEMINI_API_BASE_URL='+geminiBase+'\n'
        +'export GEMINI_API_KEY='+tok+'\n\n'
        +'# or add to ~/.gemini/settings.json:\n'
        +'{\n'
        +'  "apiBaseUrl": "'+geminiBase+'",\n'
        +'  "apiKey": "'+tok+'"\n'
        +'}';
    }
    if(!cfg) return;
    navigator.clipboard.writeText(cfg).then(()=>{
      alert(title+'\n'+hint);
    }).catch(()=>{
      prompt(T('cfg_manual'), cfg);
    });
  }).catch(e=>alert(T('err')+e));
}

// ── 에이전트 추가 모달 ──
function openAddClient() {
  _clearClientForm('ac');
  onAgentServiceChange('ac-mdl', 'ac-mdl-sel', 'google');
  document.getElementById('modal-ac').classList.add('open');
}
function submitModal(prefix) {
  const token = getAdminToken(); if (!token) return;
  const data = _readClientForm(prefix);
  const isEdit = (prefix === 'ec');
  if (!data.id && !isEdit) { document.getElementById(prefix+'-msg').textContent = T('err_id'); return; }
  if (!data.name.trim()) { document.getElementById(prefix+'-msg').textContent = T('err_name'); return; }
  document.getElementById(prefix+'-msg').textContent = isEdit ? T('saving') : T('adding');
  const url = isEdit ? '/admin/clients/'+data.id : '/admin/clients';
  const method = isEdit ? 'PUT' : 'POST';
  let body;
  if (isEdit) {
    // 변경된 필드만 전송 — 보내지 않은 필드는 Go에서 nil로 해석 = 변경 없음
    const o = _ecOrig;
    const origId = document.getElementById('ec-orig-id').value;
    const url2 = '/admin/clients/' + origId;
    body = { enabled: data.enabled }; // enabled는 항상 전송 (체크박스 의도 명확)
    if (data.token) body.token = data.token; // 토큰: 입력 시만 전송 (공백=기존 유지)
    if (data.id && data.id !== origId) body.new_id = data.id; // ID 변경
    if (data.name !== (o.name||'')) body.name = data.name;
    if (data.default_service && data.default_service !== (o.default_service||'')) body.default_service = data.default_service;
    if (data.default_model !== (o.default_model||'')) body.default_model = data.default_model;
    if (data.agent_type !== (o.agent_type||'')) body.agent_type = data.agent_type;
    if (data.work_dir !== (o.work_dir||'')) body.work_dir = data.work_dir;
    if (data.description !== (o.description||'')) body.description = data.description;
    const newIps = data.ip_whitelist.join(','), oldIps = (o.ip_whitelist||[]).join(',');
    if (newIps !== oldIps) body.ip_whitelist = data.ip_whitelist;
    if (data.avatar !== (o.avatar||'')) body.avatar = data.avatar;
    fetch(url2, {method:'PUT', headers:{'Content-Type':'application/json','Authorization':'Bearer '+token}, body:JSON.stringify(body)})
    .then(r=>r.json()).then(d=>{
      if (d.error) {
        if (d.error==='unauthorized'){clearAdminToken();alert(T('err_token'));}
        else document.getElementById(prefix+'-msg').textContent=T('err')+d.error;
      } else {
        closeModal(prefix);
        setTimeout(()=>location.reload(),500);
      }
    }).catch(e=>{document.getElementById(prefix+'-msg').textContent=T('err')+e;});
    return;
  } else {
    body = data;
  }
  fetch(url, {method, headers:{'Content-Type':'application/json','Authorization':'Bearer '+token}, body:JSON.stringify(body)})
  .then(r=>r.json()).then(d=>{
    if (d.error) {
      if (d.error==='unauthorized'){clearAdminToken();alert(T('err_token'));}
      else document.getElementById(prefix+'-msg').textContent=T('err')+d.error;
    } else {
      if (!isEdit && d.token) alert(T('tok_info')+d.token);
      closeModal(prefix);
      setTimeout(()=>location.reload(),500);
    }
  }).catch(e=>{document.getElementById(prefix+'-msg').textContent=T('err')+e;});
}

// ── 에이전트 편집 모달 ──
function openEditClient(id) {
  const token = getAdminToken(); if (!token) return;
  fetch('/admin/clients/'+id, {headers:{'Authorization':'Bearer '+token}})
  .then(r=>r.json()).then(c=>{
    if(c.error){if(c.error==='unauthorized'){clearAdminToken();alert(T('err_token'));}else alert(T('err')+c.error);return;}
    _ecOrig = c; // 원본 보관 — submitModal에서 변경된 필드만 전송하기 위함
    document.getElementById('ec-id').value = c.id||'';
    document.getElementById('ec-orig-id').value = c.id||'';
    document.getElementById('ec-name').value = c.name||'';
    // 토큰: 항상 빈칸, 플레이스홀더로 상태 표시
    const tokenEl = document.getElementById('ec-token');
    tokenEl.value = '';
    tokenEl.placeholder = c.token ? T('ph_token_edit') : T('ph_auto');
    document.getElementById('ec-service').value = c.default_service||'google';
    document.getElementById('ec-mdl').value = c.default_model||'';
    document.getElementById('ec-agent-type').value = c.agent_type||'';
    document.getElementById('ec-workdir').value = c.work_dir||'';
    document.getElementById('ec-desc').value = c.description||'';
    document.getElementById('ec-ipwl').value = (c.ip_whitelist||[]).join(', ');
    document.getElementById('ec-enabled').checked = c.enabled!==false;
    document.getElementById('ec-msg').textContent = '';
    // 아바타 미리보기
    const avHid = document.getElementById('ec-avatar');
    const avPrev = document.getElementById('ec-avatar-preview');
    if (avHid) avHid.value = c.avatar||'';
    if (avPrev) { if(c.avatar){avPrev.src=c.avatar;avPrev.style.display='block';}else{avPrev.src='';avPrev.style.display='none';} }
    // 모델 목록 로드
    onAgentServiceChange('ec-mdl', 'ec-mdl-sel', c.default_service||'google');
    document.getElementById('modal-ec').classList.add('open');
  }).catch(e=>alert(T('err')+e));
}

// ── 에이전트 종류 변경 시 작업 디렉토리 처리 ──
// • 필드가 비어있으면 → 새 타입의 기본값으로 자동 채움
// • 필드에 값이 있어도 → 플레이스홀더는 항상 새 타입 힌트로 업데이트
function onAgentTypeChange(type, wdId) {
  const el = document.getElementById(wdId);
  if (!el) return;
  const defaults = {
    'openclaw':    '~/.openclaw',
    'claude-code': '~/.claude',
    'cursor':      '~/projects',
    'vscode':      '~/projects',
  };
  const d = defaults[type] || '';
  el.placeholder = d || '/path/to/workdir';
  // 비어있을 때만 자동 채움 — 이미 입력된 값(커스텀 경로)은 건드리지 않음
  if (!el.value && d) el.value = d;
}

// ── 모델 드롭다운 (서비스 카드 등 datalist용 — 기존 호환) ──
function onServiceChange(inputId, service, listId) {
  fetch('/admin/models?service='+service, {headers:{'Authorization':'Bearer '+(localStorage.getItem('wv_admin_token')||'')}})
  .then(r=>r.json()).then(data=>{
    const dl = document.getElementById(listId);
    if(!dl) return;
    dl.innerHTML='';
    (data.models||[]).forEach(m=>{
      const opt=document.createElement('option');opt.value=m.id;opt.label=m.name||m.id;dl.appendChild(opt);
    });
  }).catch(()=>{});
}
// ── 에이전트 모달 전용: 서비스 변경 시 모델 select 채움 ──
function onAgentServiceChange(inputId, selId, service) {
  const sel = document.getElementById(selId);
  if(!sel) return;
  sel.innerHTML = '<option value="">'+T('detecting')+'</option>';
  fetch('/admin/models?service='+service, {headers:{'Authorization':'Bearer '+(localStorage.getItem('wv_admin_token')||_SERVER_TOKEN||'')}})
  .then(r=>r.json()).then(data=>{
    sel.innerHTML = '<option value="">'+T('sel_model_or_enter')+'</option>';
    (data.models||[]).forEach(m=>{
      const opt=document.createElement('option');
      opt.value=m.id;
      opt.textContent=m.name||m.id;
      sel.appendChild(opt);
    });
    const inp=document.getElementById(inputId);
    if(inp&&inp.value){
      sel.value=inp.value;
      if(!sel.value){
        // 현재 값이 목록에 없음 → 직접 입력 중 표시
        const cur=document.createElement('option');
        cur.value=inp.value;
        cur.textContent='✏ '+inp.value;
        cur.selected=true;
        sel.insertBefore(cur,sel.firstChild.nextSibling);
      }
    }
  }).catch(()=>{
    const inp=document.getElementById(inputId);
    sel.innerHTML='<option value="">'+T('sel_model_or_enter')+'</option>';
    if(inp&&inp.value){
      const cur=document.createElement('option');
      cur.value=inp.value;cur.textContent='✏ '+inp.value;cur.selected=true;
      sel.appendChild(cur);
    }
  });
}
// ── 모델 select 선택 시 input에 반영 ──
function onModelSelect(selId, inputId) {
  const sel=document.getElementById(selId);
  const inp=document.getElementById(inputId);
  if(sel&&inp&&sel.value) inp.value=sel.value;
}

// ── 서비스 관리 ──
function toggleService(id, enabled) {
  const token=getAdminToken();if(!token)return;
  const urlEl=document.getElementById('svc-url-'+id);
  const proxyEl=document.getElementById('svc-proxy-'+id);
  fetch('/admin/services/'+id,{method:'PUT',headers:{'Content-Type':'application/json','Authorization':'Bearer '+token},
    body:JSON.stringify({enabled:enabled,local_url:urlEl?urlEl.value:'',proxy_enabled:proxyEl?proxyEl.checked:false})})
  .then(r=>r.json()).then(d=>{if(d.error){alert(T('err')+d.error);}})
  .catch(e=>alert(T('err')+e));
}
function saveServiceURL(id) {
  const token=getAdminToken();if(!token)return;
  const urlEl=document.getElementById('svc-url-'+id);if(!urlEl)return;
  const enEl=document.getElementById('svc-en-'+id);
  const proxyEl=document.getElementById('svc-proxy-'+id);
  fetch('/admin/services/'+id,{method:'PUT',headers:{'Content-Type':'application/json','Authorization':'Bearer '+token},
    body:JSON.stringify({local_url:urlEl.value,enabled:enEl?enEl.checked:true,proxy_enabled:proxyEl?proxyEl.checked:false})})
  .then(r=>r.json()).then(d=>{if(d.error)alert(T('err')+d.error);})
  .catch(e=>alert(T('err')+e));
}
function toggleProxyService(id, enabled) {
  const token=getAdminToken();if(!token)return;
  const urlEl=document.getElementById('svc-url-'+id);
  const enEl=document.getElementById('svc-en-'+id);
  fetch('/admin/services/'+id,{method:'PUT',headers:{'Content-Type':'application/json','Authorization':'Bearer '+token},
    body:JSON.stringify({enabled:enEl?enEl.checked:true,local_url:urlEl?urlEl.value:'',proxy_enabled:enabled})})
  .then(r=>r.json()).then(d=>{if(d.error){alert(T('err')+d.error);}})
  .catch(e=>alert(T('err')+e));
}
function detectLocalModels(id) {
  const urlEl=document.getElementById('svc-url-'+id);if(!urlEl)return;
  urlEl.placeholder=T('detecting');
  const token=getAdminToken();
  fetch('/admin/models?service='+id,{headers:{'Authorization':'Bearer '+(token||'')}})
  .then(r=>r.json()).then(data=>{
    const n=(data.models||[]).length;
    urlEl.placeholder=T('detected')+' ('+n+')';
  }).catch(()=>{urlEl.placeholder='';});
}
function deleteService(id) {
  if(!confirm(T('del_service')))return;
  const token=getAdminToken();if(!token)return;
  fetch('/admin/services/'+id,{method:'DELETE',headers:{'Authorization':'Bearer '+token}})
  .then(r=>r.json()).then(d=>{
    if(d.error){if(d.error==='unauthorized'){clearAdminToken();alert(T('err_token'));}else alert(T('err')+d.error);}
    else location.reload();
  });
}
function openAddService() {
  ['asvc-id','asvc-name','asvc-url'].forEach(id=>{document.getElementById(id).value='';});
  document.getElementById('asvc-enabled').checked=true;
  document.getElementById('asvc-msg').textContent='';
  document.getElementById('modal-addsvc').classList.add('open');
}
function submitAddService() {
  const token=getAdminToken();if(!token)return;
  const id=document.getElementById('asvc-id').value.trim();
  const name=document.getElementById('asvc-name').value.trim();
  const url=document.getElementById('asvc-url').value.trim();
  const enabled=document.getElementById('asvc-enabled').checked;
  if(!id){document.getElementById('asvc-msg').textContent=T('err_id');return;}
  document.getElementById('asvc-msg').textContent=T('adding');
  fetch('/admin/services',{method:'POST',headers:{'Content-Type':'application/json','Authorization':'Bearer '+token},
    body:JSON.stringify({id,name:name||id,local_url:url,enabled,custom:true})})
  .then(r=>r.json()).then(d=>{
    if(d.error){if(d.error==='unauthorized'){clearAdminToken();alert(T('err_token'));}else document.getElementById('asvc-msg').textContent=T('err')+d.error;}
    else{closeModal('addsvc');setTimeout(()=>location.reload(),500);}
  }).catch(e=>{document.getElementById('asvc-msg').textContent=T('err')+e;});
}

// 에이전트 삭제
function deleteClient(id) {
  if (!confirm(T('del_agent'))) return;
  const token = getAdminToken(); if (!token) return;
  fetch('/admin/clients/'+id,{method:'DELETE',headers:{'Authorization':'Bearer '+token}})
  .then(r=>r.json()).then(d=>{
    if(d.error){if(d.error==='unauthorized'){clearAdminToken();alert(T('err_token'));}else alert(T('err')+d.error);}
    else location.reload();
  });
}

// 모델 폼 토글 (에이전트 카드)
function toggleModelForm(clientId) {
  const form = document.querySelector('.model-form[data-client="'+clientId+'"]');
  if (!form) return;
  form.classList.toggle('open');
}

// 모델 변경 (에이전트 카드 인라인) — reload 없이 인라인 피드백
function changeModel(clientId) {
  const svc = document.getElementById('svc-'+clientId).value;
  const model = document.getElementById('mdl-'+clientId).value.trim();
  if (!model) return alert(T('err_model'));
  const token = getAdminToken(); if (!token) return;
  // 버튼 찾기 (onclick="changeModel('...')" 버튼)
  const btn = document.querySelector('[onclick="changeModel(\''+clientId+'\')"]');
  if(btn) { btn.disabled=true; btn.textContent='…'; }
  fetch('/admin/clients/'+clientId,{method:'PUT',headers:{'Content-Type':'application/json','Authorization':'Bearer '+token},
    body:JSON.stringify({default_service:svc,default_model:model})})
  .then(r=>r.json()).then(d=>{
    if(d.error){
      if(d.error==='unauthorized'){clearAdminToken();alert(T('err_token'));}
      else alert(T('err')+d.error);
      if(btn){btn.disabled=false;btn.setAttribute('data-i18n','apply');btn.textContent=T('apply');}
    } else {
      // 저장 성공 — 버튼에 ✓ 표시 후 원래 텍스트로 복귀
      if(btn){
        btn.disabled=false;
        btn.textContent='\u2713';
        btn.style.color='var(--green)';
        setTimeout(()=>{btn.textContent=T('apply');btn.style.color='';},2000);
      }
      // agent-live 영역에도 저장 확인 표시 (미연결 상태여도 저장됐음을 알림)
      const agentItem = btn ? btn.closest('.agent-card') : null;
      if(agentItem){
        const liveDiv = agentItem.querySelector('.agent-live');
        if(liveDiv){
          const orig = liveDiv.innerHTML;
          const saved = document.createElement('span');
          saved.style.cssText='color:var(--green);font-size:.7rem;margin-left:.4rem';
          saved.textContent='\u2713 \uc800\uc7a5\ub428';
          liveDiv.appendChild(saved);
          setTimeout(()=>{ try{liveDiv.removeChild(saved);}catch(e){liveDiv.innerHTML=orig;} },3000);
        }
      }
    }
  }).catch(e=>{
    alert(T('err')+e);
    if(btn){btn.disabled=false;btn.setAttribute('data-i18n','apply');btn.textContent=T('apply');}
  });
}`
}

// buildServiceOptions: generate HTML options for service select element
// only includes services that are both enabled and proxy_enabled
func buildServiceOptions(services []*ServiceConfig, selected string) string {
	var sb strings.Builder
	for _, sv := range services {
		if !sv.Enabled || !sv.ProxyEnabled {
			continue
		}
		sel := ""
		if sv.ID == selected {
			sel = " selected"
		}
		sb.WriteString(fmt.Sprintf(`<option value="%s"%s>%s</option>`, sv.ID, sel, sv.Name))
	}
	return sb.String()
}

// trimServicePrefix strips "service/" prefix from model ID to avoid displaying "openrouter / openrouter/anthropic/..."
func trimServicePrefix(service, model string) string {
	prefix := service + "/"
	if strings.HasPrefix(model, prefix) {
		return model[len(prefix):]
	}
	return model
}

// resolveAvatarDataURI resolves an avatar to a data URI.
// avatarVal can be:
//   - a base64 data URI (data:image/...) → returned as-is
//   - a relative path under ~/.openclaw/ (e.g. "workspace/avatar.png") → read and encoded
//   - empty string → falls back to the default workspace avatar (workspace/avatar.png)
func resolveAvatarDataURI(avatarVal string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	ocBase := filepath.Join(home, ".openclaw")

	// already a data URI
	if strings.HasPrefix(avatarVal, "data:") {
		return avatarVal
	}

	// resolve path: use provided relative path, or default workspace avatar
	relPath := avatarVal
	if relPath == "" {
		relPath = filepath.Join("workspace", "avatar.png")
	}

	data, err := os.ReadFile(filepath.Join(ocBase, relPath))
	if err != nil {
		return ""
	}

	// detect mime type from file extension
	mime := "image/png"
	switch strings.ToLower(filepath.Ext(relPath)) {
	case ".jpg", ".jpeg", ".hpg":
		mime = "image/jpeg"
	case ".webp":
		mime = "image/webp"
	case ".gif":
		mime = "image/gif"
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)
}

func buildAgentsCard(clients []*Client, proxies []*ProxyStatus, services []*ServiceConfig) string {
	pmap := make(map[string]*ProxyStatus, len(proxies))
	for _, p := range proxies {
		pmap[p.ClientID] = p
	}

	// default workspace avatar used when a client has no avatar configured
	defaultAvatar := resolveAvatarDataURI("")

	typeIcons := map[string]string{
		"openclaw":    "🦞",
		"claude-code": "🟠",
		"cursor":      "⌨",
		"vscode":      "💻",
		"gemini-cli":  "💎",
		"antigravity": "🚀",
		"custom":      "⚙",
	}
	typeCls := map[string]string{
		"openclaw":    "atb-openclaw",
		"claude-code": "atb-claude",
		"cursor":      "atb-cursor",
		"vscode":      "atb-vscode",
		"gemini-cli":  "atb-gemini",
		"antigravity": "atb-gemini",
		"custom":      "atb-custom",
	}

	var sb strings.Builder
	// ── 섹션 배너 (전체 폭) ──
	sb.WriteString(fmt.Sprintf(
		`<div class="agents-section"><div class="section-banner"><h2><span data-i18n="agents">🤖 에이전트</span> <span class="count" data-count="%d" data-i18n-cnt="">%d개</span></h2><button class="btn-sm" onclick="openAddClient()" data-i18n="add">+ 추가</button></div>`,
		len(clients), len(clients),
	))
	if len(clients) == 0 {
		sb.WriteString(`<div style="color:var(--text-muted);font-size:.82rem" data-i18n="no_agents">등록된 에이전트 없음</div></div>`)
		return sb.String()
	}

	sb.WriteString(`<div class="agents-grid">`)

	for _, c := range clients {
		p := pmap[c.ID]

		// ── 연결 상태 칩 ──
		dotClass := "dot-gray"
		cardStatusCls := "ac-noconn"
		var statusChip string
		if !c.Enabled {
			statusChip = `<div class="agent-status"><span class="status-muted" data-i18n="st_disabled">— 비활성화됨</span></div>`
		} else if p == nil {
			switch c.AgentType {
			case "claude-code":
				statusChip = `<div class="agent-status">` +
					`<span class="status-dc">◎ Claude Code</span>` +
					`<span class="status-hint" data-i18n="st_claude_hint">ANTHROPIC_BASE_URL=http://localhost:56244 설정 후 재시작</span>` +
					`</div>`
			case "cursor", "vscode":
				statusChip = `<div class="agent-status">` +
					`<span class="status-dc">◎ ` + c.AgentType + `</span>` +
					`<span class="status-hint" data-i18n="st_editor_hint">Base URL을 http://localhost:56244 로 설정하세요</span>` +
					`</div>`
			case "gemini-cli", "antigravity":
				statusChip = `<div class="agent-status">` +
					`<span class="status-dc">◎ ` + c.AgentType + `</span>` +
					`<span class="status-hint" data-i18n="st_gemini_hint">GEMINI_API_BASE_URL=http://localhost:56244 설정 후 재시작</span>` +
					`</div>`
			default:
				statusChip = `<div class="agent-status">` +
					`<span class="status-dc" data-i18n="st_no_proxy">● 프록시 미연결</span>` +
					`<span class="status-hint" data-i18n="st_proxy_hint">프록시를 VAULT_TOKEN으로 실행하면 연결됩니다</span>` +
					`</div>`
			}
		} else {
			age := time.Since(p.UpdatedAt)
			ageSec := int(age.Seconds())
			switch {
			case age < 3*time.Minute:
				dotClass = "dot-green"
				cardStatusCls = "ac-live"
				statusChip = fmt.Sprintf(
					`<div class="agent-status"><span class="status-live"><span data-i18n="st_running">● 실행 중</span> — %s / %s</span> <span class="status-hint"><span class="bot-ago" data-ago-sec="%d">%.0f초 전</span></span> <span class="status-version">%s</span></div>`,
					p.Service, trimServicePrefix(p.Service, p.Model), ageSec, age.Seconds(), p.Version,
				)
			case age < 10*time.Minute:
				dotClass = "dot-yellow"
				cardStatusCls = "ac-delay"
				statusChip = fmt.Sprintf(
					`<div class="agent-status"><span class="status-delay"><span data-i18n="st_delayed">◑ 지연</span> <span class="bot-ago" data-ago-sec="%d">%.0f분 전</span> — %s / %s</span></div>`,
					ageSec, age.Minutes(), p.Service, trimServicePrefix(p.Service, p.Model),
				)
			default:
				dotClass = "dot-red"
				cardStatusCls = "ac-offline"
				statusChip = fmt.Sprintf(
					`<div class="agent-status"><span class="status-offline"><span data-i18n="st_offline">✕ 오프라인</span> (<span class="bot-ago" data-ago-sec="%d">%.0f분 전</span>)</span></div>`,
					ageSec, age.Minutes(),
				)
			}
		}

		displayName := c.Name
		if displayName == "" {
			displayName = c.ID
		}
		disabledClass := ""
		if !c.Enabled {
			disabledClass = " agent-disabled"
		}

		// ── 에이전트 종류 아이콘 & 뱃지 ──
		typeIcon := "🤖"
		if ic, ok := typeIcons[c.AgentType]; ok {
			typeIcon = ic
		}
		typeBadge := ""
		if c.AgentType != "" {
			cls := "atb-custom"
			if bc, ok := typeCls[c.AgentType]; ok {
				cls = bc
			}
			typeBadge = fmt.Sprintf(`<span class="atbadge %s">%s</span>`, cls, c.AgentType)
		}

		// ── 설명 & 메타 ──
		metaLines := ""
		if c.WorkDir != "" {
			metaLines += fmt.Sprintf(`<div class="agent-meta">📁 %s</div>`, c.WorkDir)
		}
		if c.Description != "" {
			metaLines += fmt.Sprintf(`<div class="agent-desc">%s</div>`, c.Description)
		}
		if len(c.IPWhitelist) > 0 {
			metaLines += fmt.Sprintf(`<div class="agent-meta">🔒 %s</div>`, strings.Join(c.IPWhitelist, ", "))
		}

		// ── 타입별 설정 복사 버튼 ──
		var cfgButton string
		switch c.AgentType {
		case "openclaw":
			cfgButton = fmt.Sprintf(
				`<button class="btn-cfg btn-cfg-openclaw" onclick="copyOpenClawConfig('%s')" data-i18n-title="cfg_openclaw_title" title="~/.openclaw/openclaw.json 설정 복사" data-i18n="cfg_openclaw">🦞 OpenClaw 설정 복사</button>`,
				c.ID)
		case "claude-code":
			cfgButton = fmt.Sprintf(
				`<button class="btn-cfg btn-cfg-claude" onclick="copyAgentConfig('%s','claude-code')" data-i18n-title="cfg_claude_title" title="~/.claude/settings.json 설정 복사" data-i18n="cfg_claude">🟠 Claude Code 설정 복사</button>`,
				c.ID)
		case "cursor":
			cfgButton = fmt.Sprintf(
				`<button class="btn-cfg" onclick="copyAgentConfig('%s','cursor')" data-i18n-title="cfg_cursor_title" title="Cursor AI 프록시 설정 복사" data-i18n="cfg_cursor">⌨ Cursor 설정 복사</button>`,
				c.ID)
		case "vscode":
			cfgButton = fmt.Sprintf(
				`<button class="btn-cfg" onclick="copyAgentConfig('%s','vscode')" data-i18n-title="cfg_vscode_title" title="VS Code / Continue 프록시 설정 복사" data-i18n="cfg_vscode">💻 VSCode 설정 복사</button>`,
				c.ID)
		case "gemini-cli":
			cfgButton = fmt.Sprintf(
				`<button class="btn-cfg" onclick="copyAgentConfig('%s','gemini-cli')" data-i18n-title="cfg_gemini_cli_title" title="Gemini CLI 프록시 설정 복사" data-i18n="cfg_gemini_cli">💎 Gemini CLI 설정 복사</button>`,
				c.ID)
		case "antigravity":
			cfgButton = fmt.Sprintf(
				`<button class="btn-cfg" onclick="copyAgentConfig('%s','antigravity')" data-i18n-title="cfg_antigravity_title" title="Antigravity IDE 프록시 설정 복사" data-i18n="cfg_antigravity">🚀 Antigravity 설정 복사</button>`,
				c.ID)
				default:
			cfgButton = fmt.Sprintf(
				`<button class="btn-cfg" onclick="copyOpenClawConfig('%s')" data-i18n-title="cfg_copy_title" title="프록시 설정 복사" data-i18n="cfg_copy">📋 설정 복사</button>`,
				c.ID)
		}

		// ── uptime 속성 ──
		uptimeAttr := ""
		if p != nil && !p.StartedAt.IsZero() {
			uptimeAttr = fmt.Sprintf(` data-started-sec="%d"`, p.StartedAt.Unix())
		}

		// ── 개별 에이전트 카드 조립 ──
		var item strings.Builder
		item.WriteString(fmt.Sprintf(`<div class="agent-card %s%s"%s>`, cardStatusCls, disabledClass, uptimeAttr))
		// 카드 상단: 상태점 + 이름/뱃지 + 편집/삭제 버튼
		item.WriteString(`<div class="ac-top">`)
		item.WriteString(fmt.Sprintf(`<div class="dot %s" style="margin-top:.3rem"></div>`, dotClass))
		// avatar priority: client avatar path/URI > default workspace avatar > emoji icon
		// c.Avatar may be a data URI, a relative path under ~/.openclaw/, or empty
		avatarSrc := ""
		if c.Avatar != "" {
			avatarSrc = resolveAvatarDataURI(c.Avatar)
		}
		if avatarSrc == "" && (c.AgentType == "openclaw" || c.AgentType == "") {
			avatarSrc = defaultAvatar
		}
		if avatarSrc != "" {
			item.WriteString(fmt.Sprintf(`<img src="%s" class="ac-avatar" alt="%s">`, avatarSrc, displayName))
		} else {
			item.WriteString(fmt.Sprintf(`<div class="ac-type-icon">%s</div>`, typeIcon))
		}
		item.WriteString(`<div class="ac-info">`)
		item.WriteString(fmt.Sprintf(`<div class="agent-name">%s <span style="color:var(--text-muted);font-size:.7rem">(%s)</span>%s</div>`,
			displayName, c.ID, typeBadge))
		item.WriteString(statusChip)
		if metaLines != "" {
			item.WriteString(metaLines)
		}
		item.WriteString(`</div>`) // ac-info
		item.WriteString(`<div class="ac-btns">`)
		item.WriteString(fmt.Sprintf(`<button class="btn-action" onclick="openEditClient('%s')" data-i18n-title="edit" title="편집">✎</button>`, c.ID))
		item.WriteString(fmt.Sprintf(`<button class="btn-action btn-action-del" onclick="deleteClient('%s')" data-i18n-title="btn_del" title="삭제">✕</button>`, c.ID))
		item.WriteString(`</div>`) // ac-btns
		item.WriteString(`</div>`) // ac-top
		// 하단 액션 버튼 행 (모델 변경 토글 + 설정 복사)
		item.WriteString(`<div class="ac-actions">`)
		item.WriteString(fmt.Sprintf(`<button class="btn-action-wide" onclick="toggleModelForm('%s')" data-i18n="toggle_model">⚙ 모델 변경</button>`, c.ID))
		// 설정 복사 버튼 (btn-action-wide 스타일 적용)
		cfgWide := strings.ReplaceAll(cfgButton, `class="btn-cfg`, `class="btn-action-wide btn-cfg`)
		item.WriteString(cfgWide)
		item.WriteString(`</div>`) // ac-actions
		// 모델 폼 (기본 숨김, 토글로 표시)
		item.WriteString(fmt.Sprintf(`<div class="model-form" data-client="%s">`, c.ID))
		item.WriteString(fmt.Sprintf(`<div class="model-form-row"><select id="svc-%s" class="agent-svc-sel" onchange="onAgentServiceChange('mdl-%s','mdl-sel-%s',this.value)">%s</select>`,
			c.ID, c.ID, c.ID, buildServiceOptions(services, c.DefaultService)))
		item.WriteString(fmt.Sprintf(`<select id="mdl-sel-%s" onchange="onModelSelect('mdl-sel-%s','mdl-%s')"><option value="" data-i18n="sel_model">— 모델 선택 —</option></select></div>`,
			c.ID, c.ID, c.ID))
		item.WriteString(fmt.Sprintf(`<div class="model-form-row"><input id="mdl-%s" type="text" data-i18n-ph="ph_mdl" placeholder="모델명" value="%s" oninput="document.getElementById('mdl-sel-%s').value=''">`,
			c.ID, c.DefaultModel, c.ID))
		item.WriteString(fmt.Sprintf(`<button class="btn btn-save" onclick="changeModel('%s')" data-i18n="apply">💾 저장</button></div>`, c.ID))
		item.WriteString(`</div>`) // model-form
		item.WriteString(`</div>`) // agent-card
		sb.WriteString(item.String())
	}
	sb.WriteString(`</div>`) // agents-grid
	sb.WriteString(`</div>`) // agents-section
	return sb.String()
}

func buildKeysCard(keys []*APIKey, services []*ServiceConfig, activeKeys map[string]string) string {
	_ = services // 향후 서비스 그룹 표시 확장 가능
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		`<div class="card"><div class="card-hdr"><h2><span data-i18n="keys">🔑 API 키</span> <span class="count" data-count="%d" data-i18n-cnt="">%d개</span></h2><button class="btn-sm" onclick="openAddKey()" data-i18n="add">+ 추가</button></div>`,
		len(keys), len(keys),
	))

	if len(keys) == 0 {
		sb.WriteString(`<div style="color:var(--text-muted);font-size:.82rem" data-i18n="no_keys">등록된 키 없음</div></div>`)
		return sb.String()
	}

	byService := make(map[string][]*APIKey)
	svcOrder := []string{}
	for _, k := range keys {
		if _, exists := byService[k.Service]; !exists {
			svcOrder = append(svcOrder, k.Service)
		}
		byService[k.Service] = append(byService[k.Service], k)
	}

	for _, svc := range svcOrder {
		svcKeys := byService[svc]
		// For unlimited keys: compute sum of activity (max of usage, attempts) so bar
		// shows each key's share of today's total service requests (not max-relative).
		sumAct := 0
		for _, k := range svcKeys {
			if k.DailyLimit == 0 {
				ref := k.TodayUsage
				if k.TodayAttempts > ref {
					ref = k.TodayAttempts
				}
				sumAct += ref
			}
		}
		sb.WriteString(fmt.Sprintf(`<div style="margin-bottom:.8rem"><div style="font-size:.78rem;color:var(--accent);margin-bottom:.4rem">▸ %s</div>`, svc))
		for _, k := range svcKeys {
			var barPct int
			if k.DailyLimit > 0 {
				barPct = k.TodayUsage * 100 / k.DailyLimit
				if barPct == 0 && k.TodayUsage > 0 {
					barPct = 4
				}
			} else if sumAct > 0 {
				// share-of-total: this key's fraction of service activity today
				ref := k.TodayUsage
				if k.TodayAttempts > ref {
					ref = k.TodayAttempts
				}
				if ref > 0 {
					barPct = ref * 100 / sumAct
					if barPct == 0 {
						barPct = 4
					}
				}
			}
			if barPct > 100 {
				barPct = 100
			}

			barClass := "bar-green"
			statusIcon := ""
			if k.IsOnCooldown() {
				barClass = "bar-yellow"
				statusIcon = "⏸ "
			} else if k.IsExhausted() || k.UsagePct() >= 97 {
				barClass = "bar-red"
				statusIcon = "✗ "
			}
			isActive := activeKeys[k.Service] == k.ID
			if isActive {
				statusIcon = "▶ "
			}

			label := k.Label
			if label == "" {
				label = k.ID[:8]
			}
			var meta string
			if k.DailyLimit > 0 {
				meta = fmt.Sprintf("%d/%d", k.TodayUsage, k.DailyLimit)
				if k.TodayAttempts > k.TodayUsage {
					meta += fmt.Sprintf(` (%d <span data-i18n="key_att">시도</span>)`, k.TodayAttempts)
				}
			} else if k.TodayUsage > 0 {
				meta = fmt.Sprintf(`%d <span data-i18n="key_reqs">요청</span>`, k.TodayUsage)
				if k.TodayAttempts > k.TodayUsage {
					meta += fmt.Sprintf(` (%d <span data-i18n="key_att">시도</span>)`, k.TodayAttempts)
				}
			} else if k.TodayAttempts > 0 {
				meta = fmt.Sprintf(`%d <span data-i18n="key_att">시도</span>`, k.TodayAttempts)
			} else {
				meta = fmt.Sprintf(`0 <span data-i18n="key_reqs">요청</span>`)
			}
			if k.IsOnCooldown() {
				remain := time.Until(k.CooldownUntil)
				meta += fmt.Sprintf(` (%.0f<span data-i18n="key_in_min">분 후</span>)`, remain.Minutes())
			}

			activeClass := ""
			if isActive {
				activeClass = " key-active"
			}
			sb.WriteString(fmt.Sprintf(
				`<div class="key-item%s" data-key-id="%s"><div class="key-header"><span class="key-label">%s%s</span><span style="display:flex;align-items:center;gap:.4rem"><span class="key-meta">%s</span><button class="btn-del" onclick="deleteKey('%s')" data-i18n-title="btn_del" title="삭제">✕</button></span></div><div class="bar-track"><div class="bar-fill %s" style="width:%d%%"></div></div></div>`,
				activeClass, k.ID, statusIcon, label, meta, k.ID, barClass, barPct,
			))
		}
		sb.WriteString(`</div>`)
	}
	sb.WriteString(`</div>`)
	return sb.String()
}


func sel(b bool) string {
	if b {
		return " selected"
	}
	return ""
}

// buildServicesCard: service management card
func buildServicesCard(services []*ServiceConfig) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		`<div class="card"><div class="card-hdr"><h2><span data-i18n="services">⚙️ 서비스</span> <span class="count" data-count="%d" data-i18n-cnt="">%d개</span></h2><button class="btn-sm" onclick="openAddService()" data-i18n="add">+ 추가</button></div>`,
		len(services), len(services),
	))

	for _, sv := range services {
		enabledChecked := ""
		if sv.Enabled {
			enabledChecked = " checked"
		}
		proxyChecked := ""
		if sv.ProxyEnabled {
			proxyChecked = " checked"
		}
		proxyCheckbox := fmt.Sprintf(
			`<label style="display:flex;align-items:center;gap:.3rem;font-size:.72rem;color:var(--text-muted);cursor:pointer;margin-left:.2rem">
  <input type="checkbox" id="svc-proxy-%s"%s style="accent-color:var(--accent);cursor:pointer" onchange="toggleProxyService('%s',this.checked)">
  <span data-i18n="proxy_use">프록시 사용</span>
</label>`,
			sv.ID, proxyChecked, sv.ID,
		)
		localURLField := ""
		if sv.IsLocal() {
			var defaultPort string
			switch sv.ID {
			case "ollama":
				defaultPort = "11434"
			case "lmstudio":
				defaultPort = "1234"
			case "vllm":
				defaultPort = "8000"
			}
			placeholder := "http://localhost:" + defaultPort
			if defaultPort == "" {
				placeholder = "http://localhost:PORT"
			}
			localURLField = fmt.Sprintf(
				`<div style="display:flex;gap:.4rem;margin-top:.3rem">
  <input id="svc-url-%s" type="text" style="flex:1;background:var(--bg);color:var(--text);border:1px solid var(--border);padding:.25rem .5rem;border-radius:4px;font-size:.75rem;font-family:inherit" placeholder="%s" value="%s">
  <button class="btn-sm" onclick="saveServiceURL('%s')" data-i18n="save">저장</button>
  <button class="btn-sm" onclick="detectLocalModels('%s')" data-i18n-title="auto_detect" title="자동 감지">🔍</button>
</div>`,
				sv.ID, placeholder, sv.LocalURL, sv.ID, sv.ID,
			)
		}
		deleteBtn := ""
		if sv.Custom {
			deleteBtn = fmt.Sprintf(`<button class="btn-del" onclick="deleteService('%s')" data-i18n-title="btn_del" title="삭제">✕</button>`, sv.ID)
		}
		sb.WriteString(fmt.Sprintf(
			`<div class="svc-item" id="svc-item-%s">
<div style="display:flex;align-items:center;gap:.5rem">
  <label style="display:flex;align-items:center;gap:.35rem;cursor:pointer;flex:1">
    <input type="checkbox" id="svc-en-%s"%s style="accent-color:var(--accent);cursor:pointer" onchange="toggleService('%s',this.checked)">
    <span style="font-size:.82rem;color:var(--text)">%s</span>
  </label>
  %s
  %s
</div>
%s
</div>`,
			sv.ID,
			sv.ID, enabledChecked, sv.ID,
			sv.Name,
			proxyCheckbox,
			deleteBtn,
			localURLField,
		))
	}
	sb.WriteString(`</div>`)
	return sb.String()
}

func buildClientModalBody(prefix, titleKey string, services []*ServiceConfig) string {
	svcOpts := buildServiceOptions(services, "google")
	// Fields: ID → name → agent type → workdir → service → model (select+input) → desc → IP → token → enabled
	// %s: 22 total (1 titleKey + 1 svcOpts + 20 prefix)
	return fmt.Sprintf(`
<div class="modal">
  <h3 data-i18n="%s">🤖 에이전트</h3>
  <label data-i18n="lbl_id">ID</label>
  <input id="%s-id" type="text" placeholder="my-bot">
  <label data-i18n="lbl_name">이름</label>
  <input id="%s-name" type="text" placeholder="My Bot">
  <label data-i18n="lbl_agent_type">에이전트 종류</label>
  <select id="%s-agent-type" onchange="onAgentTypeChange(this.value,'%s-workdir')">
    <option value="" data-i18n="opt_select">— 선택 —</option>
    <option value="openclaw">openclaw</option>
    <option value="claude-code">claude-code</option>
    <option value="gemini-cli">gemini-cli</option>
    <option value="antigravity">antigravity</option>
    <option value="cursor">cursor</option>
    <option value="vscode">vscode</option>
    <option value="custom">custom</option>
  </select>
  <label data-i18n="lbl_work_dir">작업 디렉토리</label>
  <input id="%s-workdir" type="text" placeholder="/home/user/project">
  <label data-i18n="lbl_defsvc">기본 서비스</label>
  <select id="%s-service" onchange="onAgentServiceChange('%s-mdl','%s-mdl-sel',this.value)">%s</select>
  <label data-i18n="lbl_defmdl">기본 모델</label>
  <select id="%s-mdl-sel" onchange="onModelSelect('%s-mdl-sel','%s-mdl')" style="margin-bottom:.3rem">
    <option value="" data-i18n="sel_model_or_enter">— 모델 선택 또는 직접 입력 —</option>
  </select>
  <input id="%s-mdl" type="text" data-i18n-ph="ph_mdl" placeholder="gemini-2.5-flash" oninput="document.getElementById('%s-mdl-sel').value=''">
  <label data-i18n="lbl_description">설명</label>
  <input id="%s-desc" type="text" data-i18n-ph="ph_desc">
  <label data-i18n="lbl_ip_whitelist">허용 IP</label>
  <input id="%s-ipwl" type="text" placeholder="192.168.1.1, 10.0.0.0/24">
  <label data-i18n="lbl_avatar">아바타 이미지 (openclaw)</label>
  <div style="display:flex;align-items:center;gap:.6rem;margin-bottom:.2rem">
    <img id="%s-avatar-preview" src="" style="width:48px;height:48px;border-radius:50%%;object-fit:cover;border:1px solid var(--border);display:none">
    <input id="%s-avatar" type="hidden">
    <input type="file" accept="image/*" style="font-size:.75rem;color:var(--text-muted)" onchange="loadAvatarPreview(this,'%s-avatar','%s-avatar-preview')">
  </div>
  <label data-i18n="lbl_tok">토큰</label>
  <input id="%s-token" type="text" data-i18n-ph="ph_auto" placeholder="자동 생성" autocomplete="off">
  <label style="display:flex;align-items:center;gap:.5rem;cursor:pointer;margin-top:.4rem">
    <input id="%s-enabled" type="checkbox" checked style="width:auto">
    <span data-i18n="lbl_enabled">활성화</span>
  </label>
  <div class="msg" id="%s-msg"></div>
  <div class="modal-btns">
    <button class="btn" style="background:var(--surface);color:var(--text)" onclick="closeModal('%s')" data-i18n="cancel">취소</button>
    <button class="btn" onclick="submitModal('%s')" data-i18n="add_btn">추가</button>
  </div>
</div>`,
		titleKey,
		prefix, prefix,
		prefix, prefix,
		prefix,
		prefix, prefix, prefix, svcOpts,
		prefix, prefix, prefix,
		prefix, prefix,
		prefix, prefix, prefix, prefix, // avatar: preview, hidden, file-onchange x2
		prefix, prefix, prefix, prefix,
		prefix, prefix,
		prefix,
	)
}

func buildAddClientModal(services []*ServiceConfig) string {
	body := buildClientModalBody("ac", "m_add_client", services)
	return `<div class="modal-overlay" id="modal-ac" onclick="if(event.target===this)closeModal('ac')">` + body + `</div>`
}

func buildEditClientModal(services []*ServiceConfig) string {
	body := buildClientModalBody("ec", "m_edit_client", services)
	// ID 필드는 편집 시 수정 가능 (변경 시 new_id로 전송)
	body = strings.Replace(body,
		`<input id="ec-id" type="text" placeholder="my-bot">`,
		`<input id="ec-id" type="text" placeholder="my-bot"><input id="ec-orig-id" type="hidden">`,
		1,
	)
	// 추가 버튼 텍스트를 "저장"으로 변경
	body = strings.Replace(body,
		`data-i18n="add_btn">추가</button>`,
		`data-i18n="save">저장</button>`,
		1,
	)
	return `<div class="modal-overlay" id="modal-ec" onclick="if(event.target===this)closeModal('ec')">` + body + `</div>`
}

// buildAddServiceModal: custom service add modal
func buildAddServiceModal() string {
	return `<div class="modal-overlay" id="modal-addsvc" onclick="if(event.target===this)closeModal('addsvc')">
<div class="modal">
  <h3 data-i18n="m_add_service">⚙️ 서비스 추가</h3>
  <label data-i18n="lbl_svc_id">서비스 ID</label>
  <input id="asvc-id" type="text" placeholder="my-api">
  <label data-i18n="lbl_svc_name">서비스 이름</label>
  <input id="asvc-name" type="text" placeholder="My API">
  <label data-i18n="lbl_local_url">서버 URL</label>
  <input id="asvc-url" type="text" placeholder="http://localhost:8080">
  <label style="display:flex;align-items:center;gap:.5rem;cursor:pointer;margin-top:.4rem">
    <input id="asvc-enabled" type="checkbox" checked style="width:auto">
    <span data-i18n="lbl_enabled">활성화</span>
  </label>
  <div class="msg" id="asvc-msg"></div>
  <div class="modal-btns">
    <button class="btn" style="background:var(--surface);color:var(--text)" onclick="closeModal('addsvc')" data-i18n="cancel">취소</button>
    <button class="btn" onclick="submitAddService()" data-i18n="add_btn">추가</button>
  </div>
</div>
</div>`
}

func buildAddKeyModal(services []*ServiceConfig) string {
	svcOpts := buildServiceOptions(services, "google")
	return fmt.Sprintf(`
<div class="modal-overlay" id="modal-addkey" onclick="if(event.target===this)closeAddKey()">
<div class="modal">
  <h3 data-i18n="m_add_key">🔑 API 키 추가</h3>
  <label data-i18n="lbl_service">서비스</label>
  <select id="ak-service">%s</select>
  <label data-i18n="lbl_apikey">API 키</label>
  <input id="ak-key" type="password" data-i18n-ph="ph_key" placeholder="AIzaSy..." autocomplete="off">
  <label data-i18n="lbl_lbl">레이블 (선택)</label>
  <input id="ak-label" type="text" data-i18n-ph="ph_lbl" placeholder="my-key-1">
  <label data-i18n="lbl_limit">일일 한도 (0 = 무제한)</label>
  <input id="ak-limit" type="number" value="0" min="0">
  <div class="msg" id="ak-msg"></div>
  <div class="modal-btns">
    <button class="btn" style="background:var(--surface);color:var(--text)" onclick="closeAddKey()" data-i18n="cancel">취소</button>
    <button class="btn" onclick="submitAddKey()" data-i18n="add_btn">추가</button>
  </div>
</div>
</div>`, svcOpts)
}
