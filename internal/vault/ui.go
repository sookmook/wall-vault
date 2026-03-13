package vault

import (
	"fmt"
	"strings"
	"time"

	"github.com/sookmook/wall-vault/internal/theme"
)

func buildDashboard(s *Server, t *theme.Theme) string {
	keys := s.store.ListKeys()
	clients := s.store.ListClients()
	proxies := s.store.ListProxies()
	services := s.store.ListServices()

	css := buildCSS(t)
	agentCard := buildAgentsCard(clients, proxies, services)
	keyCard := buildKeysCard(keys, services)
	svcCard := buildServicesCard(services)
	js := buildJS(t.Name, s.cfg.Lang, s.startedAt.Unix(), services)

	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html lang="ko">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>wall-vault 키 금고</title>
<style>`)
	sb.WriteString(css)
	sb.WriteString(`</style>
</head>
<body>
<div class="topbar">
  <div class="topbar-brand">
    <span class="topbar-title">🔐 wall-vault</span>
  </div>
  <div class="topbar-controls">
    <div class="dropdown">
      <button class="dd-btn" id="dd-btn-lang" onclick="toggleDd(event,'dd-lang')">`)
	sb.WriteString(langLabel(s.cfg.Lang))
	sb.WriteString(` ▾</button>
      <div class="dd-menu" id="dd-lang">
        <div class="dd-item" data-val="ko" onclick="setLang('ko')">🌏 한국어</div>
        <div class="dd-item" data-val="en" onclick="setLang('en')">🌍 English</div>
        <div class="dd-item" data-val="zh" onclick="setLang('zh')">🌏 中文</div>
        <div class="dd-item" data-val="ja" onclick="setLang('ja')">🌏 日本語</div>
        <div class="dd-item" data-val="es" onclick="setLang('es')">🌍 Español</div>
        <div class="dd-item" data-val="hi" onclick="setLang('hi')">🌏 हिन्दी</div>
        <div class="dd-item" data-val="ar" onclick="setLang('ar')">🌍 العربية</div>
        <div class="dd-item" data-val="pt" onclick="setLang('pt')">🌍 Português</div>
        <div class="dd-item" data-val="fr" onclick="setLang('fr')">🌍 Français</div>
        <div class="dd-item" data-val="de" onclick="setLang('de')">🌍 Deutsch</div>
      </div>
    </div>
    <div class="dropdown">
      <button class="dd-btn" id="dd-btn-theme" onclick="toggleDd(event,'dd-theme')">`)
	sb.WriteString(themeLabel(t.Name))
	sb.WriteString(` ▾</button>
      <div class="dd-menu" id="dd-theme">
        <div class="dd-item" data-val="light"  onclick="setTheme('light')">☀️ light</div>
        <div class="dd-item" data-val="dark"   onclick="setTheme('dark')">🌑 dark</div>
        <div class="dd-item" data-val="gold"   onclick="setTheme('gold')">✨ gold</div>
        <div class="dd-item" data-val="cherry" onclick="setTheme('cherry')">🌸 cherry</div>
        <div class="dd-item" data-val="ocean"  onclick="setTheme('ocean')">🌊 ocean</div>
      </div>
    </div>
    <span class="badge" id="sse-badge">● 연결 중...</span>
  </div>
</div>
<div class="header">
  <img src="/logo" alt="wall-vault" class="header-logo" onerror="this.style.display='none'">
  <h1 id="page-title" data-i18n="title">AI 프록시 키 금고 대시보드</h1>
</div>
<div class="grid">`)
	sb.WriteString(agentCard)
	sb.WriteString(keyCard)
	sb.WriteString(svcCard)
	sb.WriteString(buildAddClientModal(services))
	sb.WriteString(buildAddKeyModal(services))
	sb.WriteString(buildEditClientModal(services))
	sb.WriteString(buildAddServiceModal())
	sb.WriteString(`</div>
<div class="footer">
  wall-vault v0.1.1 — <a href="https://github.com/sookmook/wall-vault">github.com/sookmook/wall-vault</a>
  &nbsp;|&nbsp; <a href="https://sookmook.org/">sookmook.org</a>
  &nbsp;|&nbsp; <a href="mailto:sookmook@gmail.com">sookmook@gmail.com</a>
  &nbsp;|&nbsp; ⏱ <span id="uptime"></span>
</div>
<div class="sse-indicator" id="sse-status">SSE: 연결 중...</div>
<script>`)
	sb.WriteString(js)
	sb.WriteString(`</script>
</body>
</html>`)
	return sb.String()
}

func langLabel(lang string) string {
	m := map[string]string{
		"ko": "🌏 한국어", "en": "🌍 English", "zh": "🌏 中文",
		"es": "🌍 Español", "hi": "🌏 हिन्दी", "ar": "🌍 العربية",
		"pt": "🌍 Português", "fr": "🌍 Français", "de": "🌍 Deutsch", "ja": "🌏 日本語",
	}
	if v, ok := m[lang]; ok {
		return v
	}
	return "🌏 한국어"
}

func themeLabel(name string) string {
	m := map[string]string{
		"light": "☀️ light", "dark": "🌑 dark", "gold": "✨ gold",
		"cherry": "🌸 cherry", "ocean": "🌊 ocean",
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
.topbar{position:sticky;top:0;z-index:500;display:flex;justify-content:space-between;align-items:center;padding:.45rem 1.4rem;background:var(--surface);border-bottom:1px solid var(--border);gap:.8rem;box-shadow:0 1px 3px rgba(0,0,0,.06)}
.topbar-brand{display:flex;align-items:center;gap:.5rem;flex-shrink:0}
.topbar-logo{height:24px;object-fit:contain}
.topbar-title{color:var(--accent);font-size:.88rem;font-weight:700;letter-spacing:1px;white-space:nowrap}
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
.header{display:flex;align-items:center;justify-content:center;gap:.55rem;padding:.45rem 1.5rem;border-bottom:1px solid var(--border);background:var(--surface)}
.header-logo{height:22px;object-fit:contain;flex-shrink:0;display:block;opacity:.8}
.header h1{color:var(--text-muted);font-size:.82rem;font-weight:500;letter-spacing:.3px;white-space:nowrap}
/* ── Badge ── */
.badge{display:inline-block;background:var(--surface);border:1px solid var(--green);color:var(--green);padding:.12rem .55rem;border-radius:20px;font-size:.72rem;font-weight:600;letter-spacing:.3px}
/* ── Grid & Cards ── */
.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(320px,1fr));gap:1rem;margin-bottom:1rem;padding:1.2rem 1.4rem}
.card{background:var(--surface);border:1px solid var(--border);border-radius:10px;padding:1.1rem 1.2rem;box-shadow:0 1px 4px rgba(0,0,0,.06),0 0 0 1px rgba(0,0,0,.02)}
.card h2{color:var(--accent);font-size:.88rem;font-weight:700;margin-bottom:.75rem;padding-bottom:.55rem;border-bottom:1px solid var(--border);display:flex;justify-content:space-between;align-items:center;letter-spacing:.2px}
.card h2 .count{color:var(--text-muted);font-size:.76rem;font-weight:400}
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
/* ── 상태 점 ── */
.dot{width:7px;height:7px;border-radius:50%;flex-shrink:0}
.dot-green{background:var(--green);box-shadow:0 0 5px var(--green)}
.dot-yellow{background:var(--yellow);box-shadow:0 0 4px var(--yellow)}
.dot-red{background:var(--red);box-shadow:0 0 4px var(--red)}
.dot-gray{background:var(--border)}
/* ── 에이전트 카드 ── */
.agent-item{padding:.65rem 0;border-bottom:1px solid var(--border)}
.agent-item:last-child{border-bottom:none;padding-bottom:0}
.agent-disabled{opacity:.4;filter:grayscale(.6)}
.agent-header{display:flex;align-items:flex-start;gap:.6rem}
.agent-name{font-size:.84rem;color:var(--text);font-weight:600;margin-bottom:.08rem;line-height:1.3}
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
/* ── 모델 폼 (가로 배치) ── */
.model-form{margin-top:.65rem;display:flex;gap:.35rem;align-items:center;flex-wrap:wrap}
.model-form select{background:var(--bg);color:var(--text);border:1px solid var(--border);padding:.28rem .5rem;border-radius:6px;font-size:.76rem;font-family:inherit;flex:0 0 auto;min-width:110px;max-width:160px;cursor:pointer}
.model-form input{background:var(--bg);color:var(--text);border:1px solid var(--border);padding:.28rem .5rem;border-radius:6px;font-size:.76rem;font-family:inherit;flex:1;min-width:90px}
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
.cherry-petal{position:fixed;top:0;pointer-events:none;z-index:9999;border-radius:50% 0 50% 0;background:radial-gradient(ellipse at 30% 20%,#ffe8f4 0%,#ff80b8 50%,#f01870 100%);box-shadow:0 0 6px #ff90c038}
/* ── Ocean 효과 ── */
@keyframes wave1{0%{transform:translateX(0)}100%{transform:translateX(-50%)}}
@keyframes wave2{0%{transform:translateX(-50%)}100%{transform:translateX(0)}}
@keyframes cloud-drift{0%{left:-260px;opacity:0}6%{opacity:.85}94%{opacity:.85}100%{left:112vw;opacity:0}}
@keyframes cloud-bob{0%,100%{transform:translateY(0) rotate(-1deg)}35%{transform:translateY(-24px) rotate(2deg)}68%{transform:translateY(-10px) rotate(-2deg)}}
@keyframes bubble-rise{0%{transform:translate(0,0);opacity:0}12%{opacity:.55}45%{transform:translate(16px,-40vh)}70%{transform:translate(-12px,-65vh)}92%{opacity:.35}100%{transform:translate(8px,-105vh);opacity:0}}
@keyframes ocean-sparkle{0%,100%{opacity:0;transform:scale(0)}50%{opacity:.8;transform:scale(1)}}
.ocean-fx{position:fixed;bottom:0;left:0;width:100%;height:160px;pointer-events:none;z-index:9999;overflow:hidden}
.ocean-wave{position:absolute;left:-100%;width:300%;border-radius:42% 42% 0 0}
.ocean-cloud{position:fixed;pointer-events:none;z-index:9998;white-space:nowrap;line-height:1}
.ocean-bubble{position:fixed;pointer-events:none;z-index:9997;border-radius:50%;background:radial-gradient(circle at 32% 28%,rgba(255,255,255,.55),rgba(90,200,240,.18));border:1px solid rgba(90,200,240,.38)}
.ocean-sparkle{position:fixed;pointer-events:none;z-index:9996;width:6px;height:6px;border-radius:50%;background:radial-gradient(circle,#90eeff,#0098d8);animation:ocean-sparkle ease-in-out infinite}
/* ── Gold 반짝임 ── */
@keyframes gold-twinkle{0%,100%{opacity:0;transform:scale(0) rotate(0deg)}45%{opacity:.25}50%{opacity:1;transform:scale(1) rotate(15deg)}55%{opacity:.25}}
@keyframes gold-drift{0%,100%{transform:translate(0,0)}25%{transform:translate(14px,-18px)}75%{transform:translate(-12px,14px)}}
.gold-spark{position:fixed;pointer-events:none;z-index:9999;color:#c89000;text-shadow:0 0 6px #ffd70090,0 0 18px #ffaa0060}`
}

func buildJS(currentTheme, currentLang string, startedAt int64, services []*ServiceConfig) string {
	// 서비스 목록을 JS 객체로 직렬화 (ID, Name, IsLocal)
	var svcJSParts []string
	for _, sv := range services {
		isLocal := "false"
		if sv.IsLocal() {
			isLocal = "true"
		}
		svcJSParts = append(svcJSParts, fmt.Sprintf(`%q:{name:%q,local:%s}`, sv.ID, sv.Name, isLocal))
	}
	svcJSMap := "{" + strings.Join(svcJSParts, ",") + "}"

	return fmt.Sprintf(`const _SERVICES=%s;`+"\n", svcJSMap) + fmt.Sprintf(`
// ── I18N ──
const I18N={
ko:{title:'AI 프록시 키 금고 대시보드',agents:'🤖 에이전트',keys:'🔑 API 키',services:'⚙️ 서비스',cnt:'개',add:'+ 추가',add_btn:'추가',apply:'적용',save:'저장',cancel:'취소',edit:'편집',no_agents:'등록된 에이전트 없음',no_keys:'등록된 키 없음',lbl_service:'서비스',lbl_model:'모델',sse_conn:'● 연결 중...',sse_run:'● 연결됨',sse_st_ok:'SSE: 연결됨',sse_st_conn:'SSE: 연결 중...',sse_st_retry:'SSE: 재연결 중...',upd:'일',uph:'h',upm:'m',ups:'s',ago_s:'초 전',ago_m:'분 전',del_key:'이 API 키를 삭제하시겠습니까?',del_agent:'이 에이전트를 삭제하시겠습니까?',del_service:'이 서비스를 삭제하시겠습니까?',err_model:'모델명을 입력하세요',err_key:'키를 입력하세요',err_id:'ID를 입력하세요',err_token:'토큰 오류',err:'오류: ',adding:'추가 중...',saving:'저장 중...',tok_info:'에이전트 토큰 (저장하세요):\n\n',admin_prompt:'Admin Token:',m_add_client:'🤖 에이전트 추가',m_edit_client:'✏️ 에이전트 편집',m_add_key:'🔑 API 키 추가',m_add_service:'⚙️ 서비스 추가',lbl_id:'ID (영문·숫자·하이픈)',lbl_name:'이름',lbl_tok:'토큰 (빈칸이면 자동 생성)',lbl_defsvc:'기본 서비스',lbl_defmdl:'기본 모델',lbl_apikey:'API 키',lbl_lbl:'레이블 (선택)',lbl_limit:'일일 한도 (0 = 무제한)',ph_auto:'자동 생성',ph_key:'AIzaSy... 또는 sk-or-...',ph_lbl:'my-key-1',ph_mdl:'gemini-2.5-flash',ph_desc:'비서 AI 에이전트',lbl_agent_type:'에이전트 종류',lbl_work_dir:'작업 디렉토리',lbl_description:'설명',lbl_ip_whitelist:'허용 IP (쉼표 구분, 빈칸=모두)',lbl_enabled:'활성화',lbl_local_url:'서버 URL',lbl_svc_id:'서비스 ID',lbl_svc_name:'서비스 이름',opt_select:'— 선택 —',detecting:'감지 중...',detected:'감지 완료'},
en:{title:'AI Proxy Key Vault Dashboard',agents:'🤖 Agents',keys:'🔑 API Keys',services:'⚙️ Services',cnt:'',add:'+ Add',add_btn:'Add',apply:'Apply',save:'Save',cancel:'Cancel',edit:'Edit',no_agents:'No agents registered',no_keys:'No keys registered',lbl_service:'Service',lbl_model:'Model',sse_conn:'● Connecting...',sse_run:'● Connected',sse_st_ok:'SSE: Connected',sse_st_conn:'SSE: Connecting...',sse_st_retry:'SSE: Reconnecting...',upd:'d',uph:'h',upm:'m',ups:'s',ago_s:'s ago',ago_m:'m ago',del_key:'Delete this API key?',del_agent:'Delete this agent?',del_service:'Delete this service?',err_model:'Enter model name',err_key:'Enter API key',err_id:'Enter ID',err_token:'Token error',err:'Error: ',adding:'Adding...',saving:'Saving...',tok_info:'Agent token (save this):\n\n',admin_prompt:'Admin Token:',m_add_client:'🤖 Add Agent',m_edit_client:'✏️ Edit Agent',m_add_key:'🔑 Add API Key',m_add_service:'⚙️ Add Service',lbl_id:'ID (letters, numbers, hyphens)',lbl_name:'Name',lbl_tok:'Token (auto-generated if empty)',lbl_defsvc:'Default Service',lbl_defmdl:'Default Model',lbl_apikey:'API Key',lbl_lbl:'Label (optional)',lbl_limit:'Daily limit (0 = unlimited)',ph_auto:'auto-generated',ph_key:'AIzaSy... or sk-or-...',ph_lbl:'my-key-1',ph_mdl:'gemini-2.5-flash',ph_desc:'AI assistant agent',lbl_agent_type:'Agent Type',lbl_work_dir:'Working Directory',lbl_description:'Description',lbl_ip_whitelist:'Allowed IPs (comma-separated, empty=all)',lbl_enabled:'Enabled',lbl_local_url:'Server URL',lbl_svc_id:'Service ID',lbl_svc_name:'Service Name',opt_select:'— Select —',detecting:'Detecting...',detected:'Detected'},
zh:{title:'AI代理密钥保险库仪表板',agents:'🤖 代理',keys:'🔑 API密钥',services:'⚙️ 服务',cnt:'个',add:'+ 添加',add_btn:'添加',apply:'应用',save:'保存',cancel:'取消',edit:'编辑',no_agents:'无已注册代理',no_keys:'无已注册密钥',lbl_service:'服务',lbl_model:'模型',sse_conn:'● 连接中...',sse_run:'● 已连接',sse_st_ok:'SSE: 已连接',sse_st_conn:'SSE: 连接中...',sse_st_retry:'SSE: 重连中...',upd:'天',uph:'时',upm:'分',ups:'秒',ago_s:'秒前',ago_m:'分前',del_key:'删除此API密钥？',del_agent:'删除此代理？',del_service:'删除此服务？',err_model:'请输入模型名',err_key:'请输入密钥',err_id:'请输入ID',err_token:'Token错误',err:'错误: ',adding:'添加中...',saving:'保存中...',tok_info:'代理Token（请保存）:\n\n',admin_prompt:'管理员Token:',m_add_client:'🤖 添加代理',m_edit_client:'✏️ 编辑代理',m_add_key:'🔑 添加API密钥',m_add_service:'⚙️ 添加服务',lbl_id:'ID（字母·数字·连字符）',lbl_name:'名称',lbl_tok:'Token（空则自动生成）',lbl_defsvc:'默认服务',lbl_defmdl:'默认模型',lbl_apikey:'API密钥',lbl_lbl:'标签（可选）',lbl_limit:'每日限制（0=无限）',ph_auto:'自动生成',ph_key:'AIzaSy... 或 sk-or-...',ph_lbl:'my-key-1',ph_mdl:'gemini-2.5-flash',ph_desc:'AI助手代理',lbl_agent_type:'代理类型',lbl_work_dir:'工作目录',lbl_description:'描述',lbl_ip_whitelist:'允许IP（逗号分隔，空=全部）',lbl_enabled:'启用',lbl_local_url:'服务器URL',lbl_svc_id:'服务ID',lbl_svc_name:'服务名称',opt_select:'— 选择 —',detecting:'检测中...',detected:'检测完成'},
ja:{title:'AIプロキシ キー金庫ダッシュボード',agents:'🤖 エージェント',keys:'🔑 APIキー',services:'⚙️ サービス',cnt:'件',add:'+ 追加',add_btn:'追加',apply:'適用',save:'保存',cancel:'キャンセル',edit:'編集',no_agents:'登録済みエージェントなし',no_keys:'登録済みキーなし',lbl_service:'サービス',lbl_model:'モデル',sse_conn:'● 接続中...',sse_run:'● 接続済み',sse_st_ok:'SSE: 接続済み',sse_st_conn:'SSE: 接続中...',sse_st_retry:'SSE: 再接続中...',upd:'日',uph:'時',upm:'分',ups:'秒',ago_s:'秒前',ago_m:'分前',del_key:'このAPIキーを削除しますか？',del_agent:'このエージェントを削除しますか？',del_service:'このサービスを削除しますか？',err_model:'モデル名を入力してください',err_key:'キーを入力してください',err_id:'IDを入力してください',err_token:'トークンエラー',err:'エラー: ',adding:'追加中...',saving:'保存中...',tok_info:'エージェントトークン（保存してください）:\n\n',admin_prompt:'管理者トークン:',m_add_client:'🤖 エージェント追加',m_edit_client:'✏️ エージェント編集',m_add_key:'🔑 APIキー追加',m_add_service:'⚙️ サービス追加',lbl_id:'ID（英数字・ハイフン）',lbl_name:'名前',lbl_tok:'トークン（空なら自動生成）',lbl_defsvc:'デフォルトサービス',lbl_defmdl:'デフォルトモデル',lbl_apikey:'APIキー',lbl_lbl:'ラベル（任意）',lbl_limit:'1日の上限（0=無制限）',ph_auto:'自動生成',ph_key:'AIzaSy... または sk-or-...',ph_lbl:'my-key-1',ph_mdl:'gemini-2.5-flash',ph_desc:'AIアシスタントエージェント',lbl_agent_type:'エージェント種別',lbl_work_dir:'作業ディレクトリ',lbl_description:'説明',lbl_ip_whitelist:'許可IP（カンマ区切り、空=全許可）',lbl_enabled:'有効',lbl_local_url:'サーバーURL',lbl_svc_id:'サービスID',lbl_svc_name:'サービス名',opt_select:'— 選択 —',detecting:'検出中...',detected:'検出完了'},
es:{title:'Panel del Almacén de Claves del Proxy IA',agents:'🤖 Agentes',keys:'🔑 Claves API',services:'⚙️ Servicios',cnt:'',add:'+ Añadir',add_btn:'Añadir',apply:'Aplicar',save:'Guardar',cancel:'Cancelar',edit:'Editar',no_agents:'Sin agentes registrados',no_keys:'Sin claves registradas',lbl_service:'Servicio',lbl_model:'Modelo',sse_conn:'● Conectando...',sse_run:'● Conectado',sse_st_ok:'SSE: Conectado',sse_st_conn:'SSE: Conectando...',sse_st_retry:'SSE: Reconectando...',upd:'d',uph:'h',upm:'m',ups:'s',ago_s:'s atrás',ago_m:'m atrás',del_key:'¿Eliminar esta clave API?',del_agent:'¿Eliminar este agente?',del_service:'¿Eliminar este servicio?',err_model:'Introduce el nombre del modelo',err_key:'Introduce la clave',err_id:'Introduce el ID',err_token:'Error de token',err:'Error: ',adding:'Añadiendo...',saving:'Guardando...',tok_info:'Token del agente (guárdalo):\n\n',admin_prompt:'Token de administrador:',m_add_client:'🤖 Añadir agente',m_edit_client:'✏️ Editar agente',m_add_key:'🔑 Añadir clave API',m_add_service:'⚙️ Añadir servicio',lbl_id:'ID (letras, números, guiones)',lbl_name:'Nombre',lbl_tok:'Token (auto si vacío)',lbl_defsvc:'Servicio predeterminado',lbl_defmdl:'Modelo predeterminado',lbl_apikey:'Clave API',lbl_lbl:'Etiqueta (opcional)',lbl_limit:'Límite diario (0=sin límite)',ph_auto:'auto',ph_key:'AIzaSy... o sk-or-...',ph_lbl:'my-key-1',ph_mdl:'gemini-2.5-flash',ph_desc:'Agente asistente IA',lbl_agent_type:'Tipo de agente',lbl_work_dir:'Directorio de trabajo',lbl_description:'Descripción',lbl_ip_whitelist:'IPs permitidas (separadas por coma, vacío=todas)',lbl_enabled:'Habilitado',lbl_local_url:'URL del servidor',lbl_svc_id:'ID de servicio',lbl_svc_name:'Nombre de servicio',opt_select:'— Seleccionar —',detecting:'Detectando...',detected:'Detectado'},
hi:{title:'AI प्रॉक्सी की वॉल्ट डैशबोर्ड',agents:'🤖 एजेंट',keys:'🔑 API कुंजियाँ',services:'⚙️ सेवाएं',cnt:'',add:'+ जोड़ें',add_btn:'जोड़ें',apply:'लागू',save:'सहेजें',cancel:'रद्द करें',edit:'संपादित',no_agents:'कोई एजेंट नहीं',no_keys:'कोई कुंजी नहीं',lbl_service:'सेवा',lbl_model:'मॉडल',sse_conn:'● जोड़ रहे हैं...',sse_run:'● जुड़ा',sse_st_ok:'SSE: जुड़ा',sse_st_conn:'SSE: जोड़ रहे...',sse_st_retry:'SSE: पुनः जोड़ रहे...',upd:'दिन',uph:'घं',upm:'मि',ups:'से',ago_s:'से पहले',ago_m:'मि पहले',del_key:'इस API कुंजी को हटाएं?',del_agent:'इस एजेंट को हटाएं?',del_service:'इस सेवा को हटाएं?',err_model:'मॉडल नाम दर्ज करें',err_key:'कुंजी दर्ज करें',err_id:'ID दर्ज करें',err_token:'टोकन त्रुटि',err:'त्रुटि: ',adding:'जोड़ रहे...',saving:'सहेज रहे...',tok_info:'एजेंट टोकन (सहेजें):\n\n',admin_prompt:'एडमिन टोकन:',m_add_client:'🤖 एजेंट जोड़ें',m_edit_client:'✏️ एजेंट संपादित करें',m_add_key:'🔑 API कुंजी जोड़ें',m_add_service:'⚙️ सेवा जोड़ें',lbl_id:'ID (अक्षर·अंक·हाइफन)',lbl_name:'नाम',lbl_tok:'टोकन (खाली = स्वतः)',lbl_defsvc:'डिफ़ॉल्ट सेवा',lbl_defmdl:'डिफ़ॉल्ट मॉडल',lbl_apikey:'API कुंजी',lbl_lbl:'लेबल (वैकल्पिक)',lbl_limit:'दैनिक सीमा (0=असीमित)',ph_auto:'स्वतः',ph_key:'AIzaSy... या sk-or-...',ph_lbl:'my-key-1',ph_mdl:'gemini-2.5-flash',ph_desc:'AI सहायक एजेंट',lbl_agent_type:'एजेंट प्रकार',lbl_work_dir:'कार्य निर्देशिका',lbl_description:'विवरण',lbl_ip_whitelist:'अनुमत IP (कॉमा से अलग, खाली=सभी)',lbl_enabled:'सक्षम',lbl_local_url:'सर्वर URL',lbl_svc_id:'सेवा ID',lbl_svc_name:'सेवा नाम',opt_select:'— चुनें —',detecting:'पता लगा रहे...',detected:'पता चला'},
ar:{title:'لوحة تحكم خزينة مفاتيح وكيل الذكاء الاصطناعي',agents:'🤖 العملاء',keys:'🔑 مفاتيح API',services:'⚙️ الخدمات',cnt:'',add:'+ إضافة',add_btn:'إضافة',apply:'تطبيق',save:'حفظ',cancel:'إلغاء',edit:'تعديل',no_agents:'لا عملاء مسجلين',no_keys:'لا مفاتيح مسجلة',lbl_service:'الخدمة',lbl_model:'النموذج',sse_conn:'● جارٍ الاتصال...',sse_run:'● متصل',sse_st_ok:'SSE: متصل',sse_st_conn:'SSE: جارٍ الاتصال...',sse_st_retry:'SSE: إعادة الاتصال...',upd:'ي',uph:'س',upm:'د',ups:'ث',ago_s:'ث مضت',ago_m:'د مضت',del_key:'حذف مفتاح API هذا؟',del_agent:'حذف هذا العميل؟',del_service:'حذف هذه الخدمة؟',err_model:'أدخل اسم النموذج',err_key:'أدخل المفتاح',err_id:'أدخل المعرف',err_token:'خطأ في الرمز',err:'خطأ: ',adding:'جارٍ الإضافة...',saving:'جارٍ الحفظ...',tok_info:'رمز العميل (احفظه):\n\n',admin_prompt:'رمز المسؤول:',m_add_client:'🤖 إضافة عميل',m_edit_client:'✏️ تعديل عميل',m_add_key:'🔑 إضافة مفتاح API',m_add_service:'⚙️ إضافة خدمة',lbl_id:'المعرف (حروف·أرقام·شرطة)',lbl_name:'الاسم',lbl_tok:'الرمز (تلقائي إن فارغ)',lbl_defsvc:'الخدمة الافتراضية',lbl_defmdl:'النموذج الافتراضي',lbl_apikey:'مفتاح API',lbl_lbl:'التسمية (اختياري)',lbl_limit:'الحد اليومي (0=غير محدود)',ph_auto:'تلقائي',ph_key:'AIzaSy... أو sk-or-...',ph_lbl:'my-key-1',ph_mdl:'gemini-2.5-flash',ph_desc:'وكيل مساعد الذكاء الاصطناعي',lbl_agent_type:'نوع الوكيل',lbl_work_dir:'دليل العمل',lbl_description:'الوصف',lbl_ip_whitelist:'IPs المسموح بها (فاصلة، فارغ=الكل)',lbl_enabled:'مفعل',lbl_local_url:'عنوان الخادم',lbl_svc_id:'معرف الخدمة',lbl_svc_name:'اسم الخدمة',opt_select:'— اختر —',detecting:'جارٍ الكشف...',detected:'تم الكشف'},
pt:{title:'Painel do Cofre de Chaves do Proxy de IA',agents:'🤖 Agentes',keys:'🔑 Chaves API',services:'⚙️ Serviços',cnt:'',add:'+ Adicionar',add_btn:'Adicionar',apply:'Aplicar',save:'Salvar',cancel:'Cancelar',edit:'Editar',no_agents:'Sem agentes registrados',no_keys:'Sem chaves registradas',lbl_service:'Serviço',lbl_model:'Modelo',sse_conn:'● Conectando...',sse_run:'● Conectado',sse_st_ok:'SSE: Conectado',sse_st_conn:'SSE: Conectando...',sse_st_retry:'SSE: Reconectando...',upd:'d',uph:'h',upm:'m',ups:'s',ago_s:'s atrás',ago_m:'m atrás',del_key:'Excluir esta chave API?',del_agent:'Excluir este agente?',del_service:'Excluir este serviço?',err_model:'Insira o nome do modelo',err_key:'Insira a chave',err_id:'Insira o ID',err_token:'Erro de token',err:'Erro: ',adding:'Adicionando...',saving:'Salvando...',tok_info:'Token do agente (salve):\n\n',admin_prompt:'Token de administrador:',m_add_client:'🤖 Adicionar agente',m_edit_client:'✏️ Editar agente',m_add_key:'🔑 Adicionar chave API',m_add_service:'⚙️ Adicionar serviço',lbl_id:'ID (letras, números, hifens)',lbl_name:'Nome',lbl_tok:'Token (auto se vazio)',lbl_defsvc:'Serviço padrão',lbl_defmdl:'Modelo padrão',lbl_apikey:'Chave API',lbl_lbl:'Rótulo (opcional)',lbl_limit:'Limite diário (0=ilimitado)',ph_auto:'automático',ph_key:'AIzaSy... ou sk-or-...',ph_lbl:'my-key-1',ph_mdl:'gemini-2.5-flash',ph_desc:'Agente assistente de IA',lbl_agent_type:'Tipo de agente',lbl_work_dir:'Diretório de trabalho',lbl_description:'Descrição',lbl_ip_whitelist:'IPs permitidos (vírgula, vazio=todos)',lbl_enabled:'Habilitado',lbl_local_url:'URL do servidor',lbl_svc_id:'ID do serviço',lbl_svc_name:'Nome do serviço',opt_select:'— Selecionar —',detecting:'Detectando...',detected:'Detectado'},
fr:{title:'Tableau de bord du coffre de clés proxy IA',agents:'🤖 Agents',keys:'🔑 Clés API',services:'⚙️ Services',cnt:'',add:'+ Ajouter',add_btn:'Ajouter',apply:'Appliquer',save:'Enregistrer',cancel:'Annuler',edit:'Modifier',no_agents:'Aucun agent enregistré',no_keys:'Aucune clé enregistrée',lbl_service:'Service',lbl_model:'Modèle',sse_conn:'● Connexion...',sse_run:'● Connecté',sse_st_ok:'SSE: Connecté',sse_st_conn:'SSE: Connexion...',sse_st_retry:'SSE: Reconnexion...',upd:'j',uph:'h',upm:'m',ups:'s',ago_s:'s',ago_m:'min',del_key:'Supprimer cette clé API ?',del_agent:"Supprimer cet agent ?",del_service:'Supprimer ce service ?',err_model:'Entrez le nom du modèle',err_key:'Entrez la clé',err_id:"Entrez l'ID",err_token:'Erreur de token',err:'Erreur : ',adding:'Ajout en cours...',saving:'Enregistrement...',tok_info:'Token agent (sauvegardez):\n\n',admin_prompt:'Token administrateur:',m_add_client:'🤖 Ajouter un agent',m_edit_client:'✏️ Modifier un agent',m_add_key:'🔑 Ajouter une clé API',m_add_service:'⚙️ Ajouter un service',lbl_id:'ID (lettres, chiffres, tirets)',lbl_name:'Nom',lbl_tok:'Token (auto si vide)',lbl_defsvc:'Service par défaut',lbl_defmdl:'Modèle par défaut',lbl_apikey:'Clé API',lbl_lbl:'Libellé (optionnel)',lbl_limit:'Limite quotidienne (0=illimité)',ph_auto:'auto',ph_key:'AIzaSy... ou sk-or-...',ph_lbl:'my-key-1',ph_mdl:'gemini-2.5-flash',ph_desc:"Agent assistant IA",lbl_agent_type:"Type d'agent",lbl_work_dir:'Répertoire de travail',lbl_description:'Description',lbl_ip_whitelist:'IPs autorisées (virgule, vide=toutes)',lbl_enabled:'Activé',lbl_local_url:'URL du serveur',lbl_svc_id:'ID du service',lbl_svc_name:'Nom du service',opt_select:'— Sélectionner —',detecting:'Détection...',detected:'Détecté'},
de:{title:'KI-Proxy-Schlüsseltresor-Dashboard',agents:'🤖 Agenten',keys:'🔑 API-Schlüssel',services:'⚙️ Dienste',cnt:'',add:'+ Hinzufügen',add_btn:'Hinzufügen',apply:'Anwenden',save:'Speichern',cancel:'Abbrechen',edit:'Bearbeiten',no_agents:'Keine Agenten registriert',no_keys:'Keine Schlüssel registriert',lbl_service:'Dienst',lbl_model:'Modell',sse_conn:'● Verbinde...',sse_run:'● Verbunden',sse_st_ok:'SSE: Verbunden',sse_st_conn:'SSE: Verbinde...',sse_st_retry:'SSE: Wiederverbinde...',upd:'T',uph:'h',upm:'m',ups:'s',ago_s:'s her',ago_m:'min her',del_key:'Diesen API-Schlüssel löschen?',del_agent:'Diesen Agenten löschen?',del_service:'Diesen Dienst löschen?',err_model:'Modellname eingeben',err_key:'Schlüssel eingeben',err_id:'ID eingeben',err_token:'Token-Fehler',err:'Fehler: ',adding:'Hinzufügen...',saving:'Speichern...',tok_info:'Agenten-Token (speichern):\n\n',admin_prompt:'Admin-Token:',m_add_client:'🤖 Agenten hinzufügen',m_edit_client:'✏️ Agenten bearbeiten',m_add_key:'🔑 API-Schlüssel hinzufügen',m_add_service:'⚙️ Dienst hinzufügen',lbl_id:'ID (Buchstaben, Zahlen, Bindestriche)',lbl_name:'Name',lbl_tok:'Token (auto wenn leer)',lbl_defsvc:'Standarddienst',lbl_defmdl:'Standardmodell',lbl_apikey:'API-Schlüssel',lbl_lbl:'Bezeichnung (optional)',lbl_limit:'Tageslimit (0=unbegrenzt)',ph_auto:'automatisch',ph_key:'AIzaSy... oder sk-or-...',ph_lbl:'my-key-1',ph_mdl:'gemini-2.5-flash',ph_desc:'KI-Assistent Agent',lbl_agent_type:'Agententyp',lbl_work_dir:'Arbeitsverzeichnis',lbl_description:'Beschreibung',lbl_ip_whitelist:'Erlaubte IPs (Komma, leer=alle)',lbl_enabled:'Aktiviert',lbl_local_url:'Server-URL',lbl_svc_id:'Dienst-ID',lbl_svc_name:'Dienstname',opt_select:'— Auswählen —',detecting:'Erkenne...',detected:'Erkannt'}
};
let curLang='ko';
function T(k){return(I18N[curLang]||I18N.ko)[k]||k;}
function applyLang(lang){
  curLang=lang;
  const t=I18N[lang]||I18N.ko;
  // data-i18n 텍스트
  document.querySelectorAll('[data-i18n]').forEach(el=>{const k=el.dataset.i18n;if(t[k]!==undefined)el.textContent=t[k];});
  // placeholders
  document.querySelectorAll('[data-i18n-ph]').forEach(el=>{const k=el.dataset.i18nPh;if(t[k]!==undefined)el.placeholder=t[k];});
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
  document.querySelectorAll('.agent-item[data-started-sec]').forEach(el => {
    const started = parseInt(el.dataset.startedSec);
    if (!started) return;
    let up = el.querySelector('.agent-uptime');
    if (!up) {
      up = document.createElement('div');
      up.className = 'agent-uptime';
      const live = el.querySelector('.agent-live');
      if (live) live.after(up); else {
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
const LANG_LABELS={'ko':'🌏 한국어','en':'🌍 English','zh':'🌏 中文','ja':'🌏 日本語','es':'🌍 Español','hi':'🌏 हिन्दी','ar':'🌍 العربية','pt':'🌍 Português','fr':'🌍 Français','de':'🌍 Deutsch'};
function setLang(lang){
  const tok=localStorage.getItem('wv_admin_token')||'';
  fetch('/admin/lang',{method:'PUT',headers:{'Content-Type':'application/json','Authorization':'Bearer '+tok},body:JSON.stringify({lang:lang})})
  .then(r=>r.json()).then(d=>{
    if(d.error){alert(T('err')+d.error);return;}
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
  ocean:  {'--bg':'#e6f4ff','--surface':'#f8fdff','--border':'#70c4e8','--text':'#052038','--text-muted':'#26789a','--green':'#007858','--yellow':'#b06800','--red':'#c82828','--blue':'#0070b8','--accent':'#0086c8','--accent-hover':'#10a0e0'}
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
    const dur=22+Math.random()*16;
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

function applyThemeCss(name){
  const vars=THEMES[name]; if(!vars) return;
  const root=document.documentElement;
  for(const [k,v] of Object.entries(vars)) root.style.setProperty(k,v);
  document.querySelectorAll('.theme-btn').forEach(b=>b.classList.toggle('active',b.id==='theme-'+name));
  if(name==='cherry') createCherryFx();
  else if(name==='ocean') createOceanFx();
  else if(name==='gold') createGoldFx();
  else clearFx();
}
const THEME_LABELS={'light':'☀️ light','dark':'🌑 dark','gold':'✨ gold','cherry':'🌸 cherry','ocean':'🌊 ocean'};
function setTheme(name){
  const tok=localStorage.getItem('wv_admin_token')||'';
  fetch('/admin/theme',{method:'PUT',headers:{'Content-Type':'application/json','Authorization':'Bearer '+tok},body:JSON.stringify({theme:name})})
  .then(r=>r.json()).then(d=>{
    if(d.error){alert(T('err')+d.error);return;}
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
      if (d.type === 'config_change' || d.type === 'key_added') {
        setTimeout(() => location.reload(), 800);
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

// Admin Token 헬퍼
function getAdminToken() {
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

// ── 에이전트 카드 모델 목록 초기화 (페이지 로드 시) ──
document.addEventListener('DOMContentLoaded', function() {
  document.querySelectorAll('.agent-svc-sel').forEach(function(el) {
    const id = el.id.replace('svc-', '');
    if (id && el.value) onAgentServiceChange('mdl-'+id, 'mdl-sel-'+id, el.value);
  });
});

// ── 모달 공통 유틸 ──
function closeModal(prefix) {
  document.getElementById('modal-'+prefix).classList.remove('open');
}
function _readClientForm(prefix) {
  const ipwlRaw = document.getElementById(prefix+'-ipwl').value.trim();
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
  if(ms) ms.innerHTML='<option value="">— 모델 선택 또는 직접 입력 —</option>';
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
      alert('OpenClaw 설정이 클립보드에 복사되었습니다.\n~/.openclaw/openclaw.json에 붙여넣으세요.');
    }).catch(()=>{
      prompt('아래 내용을 복사하세요:', cfg);
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
  document.getElementById(prefix+'-msg').textContent = isEdit ? T('saving') : T('adding');
  const url = isEdit ? '/admin/clients/'+data.id : '/admin/clients';
  const method = isEdit ? 'PUT' : 'POST';
  const body = isEdit
    ? {name:data.name, token:data.token||undefined, default_service:data.default_service, default_model:data.default_model, agent_type:data.agent_type, work_dir:data.work_dir, description:data.description, ip_whitelist:data.ip_whitelist, enabled:data.enabled}
    : data;
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
    document.getElementById('ec-id').value = c.id||'';
    document.getElementById('ec-name').value = c.name||'';
    document.getElementById('ec-token').value = '';
    document.getElementById('ec-service').value = c.default_service||'google';
    document.getElementById('ec-mdl').value = c.default_model||'';
    document.getElementById('ec-agent-type').value = c.agent_type||'';
    document.getElementById('ec-workdir').value = c.work_dir||'';
    document.getElementById('ec-desc').value = c.description||'';
    document.getElementById('ec-ipwl').value = (c.ip_whitelist||[]).join(', ');
    document.getElementById('ec-enabled').checked = c.enabled!==false;
    document.getElementById('ec-msg').textContent = '';
    // 모델 목록 로드
    onAgentServiceChange('ec-mdl', 'ec-mdl-sel', c.default_service||'google');
    document.getElementById('modal-ec').classList.add('open');
  }).catch(e=>alert(T('err')+e));
}

// ── 에이전트 종류 변경 시 작업 디렉토리 힌트 ──
function onAgentTypeChange(type, wdId) {
  const el = document.getElementById(wdId);
  if (!el || el.value) return; // 이미 값이 있으면 덮어쓰지 않음
  const hints = {
    'openclaw':    '~/.openclaw',
    'claude-code': '~/.claude',
    'cursor':      '~/projects',
    'vscode':      '~/projects',
  };
  el.placeholder = hints[type] || '/path/to/workdir';
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
  sel.innerHTML = '<option value="">— 로딩 중... —</option>';
  fetch('/admin/models?service='+service, {headers:{'Authorization':'Bearer '+(localStorage.getItem('wv_admin_token')||'')}})
  .then(r=>r.json()).then(data=>{
    sel.innerHTML = '<option value="">— 모델 선택 또는 직접 입력 —</option>';
    (data.models||[]).forEach(m=>{
      const opt=document.createElement('option');
      opt.value=m.id;
      opt.textContent=(m.name&&m.name!==m.id)?(m.name+' · '+m.id):m.id;
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
    sel.innerHTML='<option value="">— 선택 또는 직접 입력 —</option>';
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
  fetch('/admin/services/'+id,{method:'PUT',headers:{'Content-Type':'application/json','Authorization':'Bearer '+token},
    body:JSON.stringify({enabled:enabled,local_url:urlEl?urlEl.value:''})})
  .then(r=>r.json()).then(d=>{if(d.error){alert(T('err')+d.error);}})
  .catch(e=>alert(T('err')+e));
}
function saveServiceURL(id) {
  const token=getAdminToken();if(!token)return;
  const urlEl=document.getElementById('svc-url-'+id);if(!urlEl)return;
  const enEl=document.getElementById('svc-en-'+id);
  fetch('/admin/services/'+id,{method:'PUT',headers:{'Content-Type':'application/json','Authorization':'Bearer '+token},
    body:JSON.stringify({local_url:urlEl.value,enabled:enEl?enEl.checked:true})})
  .then(r=>r.json()).then(d=>{if(d.error)alert(T('err')+d.error);})
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

// 모델 변경 (에이전트 카드 인라인)
function changeModel(clientId) {
  const svc = document.getElementById('svc-'+clientId).value;
  const model = document.getElementById('mdl-'+clientId).value.trim();
  if (!model) return alert(T('err_model'));
  const token = getAdminToken(); if (!token) return;
  fetch('/admin/clients/'+clientId,{method:'PUT',headers:{'Content-Type':'application/json','Authorization':'Bearer '+token},
    body:JSON.stringify({default_service:svc,default_model:model})})
  .then(r=>r.json()).then(d=>{
    if(d.error){if(d.error==='unauthorized'){clearAdminToken();alert(T('err_token'));}else alert(T('err')+d.error);}
    else location.reload();
  });
}`
}

// buildServiceOptions: 서비스 select 옵션 HTML 생성
func buildServiceOptions(services []*ServiceConfig, selected string) string {
	var sb strings.Builder
	for _, sv := range services {
		if !sv.Enabled {
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

// buildAgentsCard: 에이전트 카드 (등록 클라이언트 + 실시간 heartbeat 통합)
func buildAgentsCard(clients []*Client, proxies []*ProxyStatus, services []*ServiceConfig) string {
	// clientID → ProxyStatus 맵
	pmap := make(map[string]*ProxyStatus, len(proxies))
	for _, p := range proxies {
		pmap[p.ClientID] = p
	}


	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		`<div class="card"><h2><span data-i18n="agents">🤖 에이전트</span> <span class="count" data-count="%d" data-i18n-cnt="">%d개</span><button class="btn-sm" onclick="openAddClient()" data-i18n="add">+ 추가</button></h2>`,
		len(clients), len(clients),
	))
	if len(clients) == 0 {
		sb.WriteString(`<div style="color:var(--text-muted);font-size:.82rem" data-i18n="no_agents">등록된 에이전트 없음</div></div>`)
		return sb.String()
	}

	for _, c := range clients {
		p := pmap[c.ID]

		// 온라인 상태 판별 — 4단계: green(<3분)/yellow(3-10분)/red(>10분)/gray(비활성·미연결)
		dotClass := "dot-gray"
		liveDetail := ""
		if !c.Enabled {
			liveDetail = `<div class="agent-live" style="color:var(--text-muted)">비활성화</div>`
		} else if p == nil {
			liveDetail = `<div class="agent-live" style="color:var(--text-muted)">미연결</div>`
		} else {
			age := time.Since(p.UpdatedAt)
			ageSec := int(age.Seconds())
			var ago string
			if age < time.Minute {
				ago = fmt.Sprintf("%.0f초 전", age.Seconds())
			} else {
				ago = fmt.Sprintf("%.0f분 전", age.Minutes())
			}
			switch {
			case age < 3*time.Minute:
				dotClass = "dot-green"
				liveDetail = fmt.Sprintf(
					`<div class="agent-live">%s / %s <span style="color:var(--text-muted)">— <span class="bot-ago" data-ago-sec="%d">%s</span></span> <span style="color:var(--text-muted);font-size:.7rem">%s</span></div>`,
					p.Service, p.Model, ageSec, ago, p.Version,
				)
			case age < 10*time.Minute:
				dotClass = "dot-yellow"
				liveDetail = fmt.Sprintf(
					`<div class="agent-live" style="color:var(--yellow)">지연 %s — %s / %s</div>`,
					ago, p.Service, p.Model,
				)
			default:
				dotClass = "dot-red"
				liveDetail = fmt.Sprintf(
					`<div class="agent-live" style="color:var(--red)">오프라인 (%s)</div>`,
					ago,
				)
			}
		}

		displayName := c.Name
		if displayName == "" {
			displayName = c.ID
		}

		// 비활성화 클래스
		disabledClass := ""
		if !c.Enabled {
			disabledClass = " agent-disabled"
		}

		// agent_type 뱃지
		typeBadge := ""
		if c.AgentType != "" {
			typeBadge = fmt.Sprintf(`<span class="agent-type-badge">%s</span>`, c.AgentType)
		}

		// 설명
		descLine := ""
		if c.Description != "" {
			descLine = fmt.Sprintf(`<div class="agent-desc">%s</div>`, c.Description)
		}

		// 메타 정보 (work_dir, ip_whitelist)
		metaLines := ""
		if c.WorkDir != "" {
			metaLines += fmt.Sprintf(`<div class="agent-meta">📁 %s</div>`, c.WorkDir)
		}
		if len(c.IPWhitelist) > 0 {
			metaLines += fmt.Sprintf(`<div class="agent-meta">🔒 %s</div>`, strings.Join(c.IPWhitelist, ", "))
		}

		startedAtSec := int64(0)
		if p != nil && !p.StartedAt.IsZero() {
			startedAtSec = p.StartedAt.Unix()
		}
		uptimeAttr := ""
		if startedAtSec > 0 {
			uptimeAttr = fmt.Sprintf(` data-started-sec="%d"`, startedAtSec)
		}
		sb.WriteString(fmt.Sprintf(`<div class="agent-item%s"%s>
<div class="agent-header">
  <div class="dot %s" style="margin-top:.15rem;flex-shrink:0"></div>
  <div style="flex:1;min-width:0">
    <div class="agent-name">%s <span style="color:var(--text-muted);font-size:.72rem">%s</span>%s</div>
    %s%s%s
    <div class="model-form">
      <select id="svc-%s" class="agent-svc-sel" onchange="onAgentServiceChange('mdl-%s','mdl-sel-%s',this.value)">%s</select>
      <select id="mdl-sel-%s" onchange="onModelSelect('mdl-sel-%s','mdl-%s')" style="margin-bottom:.25rem">
        <option value="">— 모델 선택 —</option>
      </select>
      <input id="mdl-%s" type="text" data-i18n-ph="ph_mdl" placeholder="모델명" value="%s" oninput="document.getElementById('mdl-sel-%s').value=''">
      <button class="btn" onclick="changeModel('%s')" data-i18n="apply">적용</button>
    </div>
  </div>
  <div style="display:flex;flex-direction:row;gap:.35rem;flex-shrink:0;align-items:flex-start">
    <button class="btn-action" onclick="copyOpenClawConfig('%s')" title="OpenClaw 설정 복사">🐾</button>
    <button class="btn-action" onclick="openEditClient('%s')" title="편집">✎</button>
    <button class="btn-action btn-action-del" onclick="deleteClient('%s')" title="삭제">✕</button>
  </div>
</div>
</div>`,
			disabledClass, uptimeAttr,
			dotClass,
			displayName, c.ID, typeBadge,
			descLine, metaLines, liveDetail,
			c.ID, c.ID, c.ID, buildServiceOptions(services, c.DefaultService),
			c.ID, c.ID, c.ID,
			c.ID, c.DefaultModel, c.ID,
			c.ID,
			c.ID,
			c.ID, c.ID,
		))
	}
	sb.WriteString(`</div>`)
	return sb.String()
}

func buildKeysCard(keys []*APIKey, services []*ServiceConfig) string {
	_ = services // 향후 서비스 그룹 표시 확장 가능
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		`<div class="card"><h2><span data-i18n="keys">🔑 API 키</span> <span class="count" data-count="%d" data-i18n-cnt="">%d개</span><button class="btn-sm" onclick="openAddKey()" data-i18n="add">+ 추가</button></h2>`,
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
		maxU := 0
		for _, k := range svcKeys {
			if k.TodayUsage > maxU {
				maxU = k.TodayUsage
			}
		}
		sb.WriteString(fmt.Sprintf(`<div style="margin-bottom:.8rem"><div style="font-size:.78rem;color:var(--accent);margin-bottom:.4rem">▸ %s</div>`, svc))
		for _, k := range svcKeys {
			var barPct int
			if k.DailyLimit > 0 {
				barPct = k.TodayUsage * 100 / k.DailyLimit
				if barPct == 0 && k.TodayUsage > 0 {
					barPct = 4 // 사용량 있으면 최소 4%
				}
			} else if maxU > 0 {
				barPct = k.TodayUsage * 100 / maxU
				if barPct == 0 && k.TodayUsage > 0 {
					barPct = 4
				}
			}
			// 사용량 0이면 barPct = 0 (빈 바)
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

			label := k.Label
			if label == "" {
				label = k.ID[:8]
			}
			var meta string
			if k.DailyLimit > 0 {
				meta = fmt.Sprintf("%d/%d", k.TodayUsage, k.DailyLimit)
			} else {
				meta = fmt.Sprintf("%d 요청", k.TodayUsage)
			}
			if k.IsOnCooldown() {
				remain := time.Until(k.CooldownUntil)
				meta += fmt.Sprintf(" (%.0f분 후)", remain.Minutes())
			}

			sb.WriteString(fmt.Sprintf(
				`<div class="key-item"><div class="key-header"><span class="key-label">%s%s</span><span style="display:flex;align-items:center;gap:.4rem"><span class="key-meta">%s</span><button class="btn-del" onclick="deleteKey('%s')" title="삭제">✕</button></span></div><div class="bar-track"><div class="bar-fill %s" style="width:%d%%"></div></div></div>`,
				statusIcon, label, meta, k.ID, barClass, barPct,
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

// buildServicesCard: 서비스 관리 카드
func buildServicesCard(services []*ServiceConfig) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		`<div class="card"><h2><span data-i18n="services">⚙️ 서비스</span> <span class="count" data-count="%d" data-i18n-cnt="">%d개</span><button class="btn-sm" onclick="openAddService()" data-i18n="add">+ 추가</button></h2>`,
		len(services), len(services),
	))

	for _, sv := range services {
		enabledChecked := ""
		if sv.Enabled {
			enabledChecked = " checked"
		}
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
  <button class="btn-sm" onclick="detectLocalModels('%s')" title="자동 감지">🔍</button>
</div>`,
				sv.ID, placeholder, sv.LocalURL, sv.ID, sv.ID,
			)
		}
		deleteBtn := ""
		if sv.Custom {
			deleteBtn = fmt.Sprintf(`<button class="btn-del" onclick="deleteService('%s')" title="삭제">✕</button>`, sv.ID)
		}
		sb.WriteString(fmt.Sprintf(
			`<div class="svc-item" id="svc-item-%s">
<div style="display:flex;align-items:center;gap:.5rem">
  <label style="display:flex;align-items:center;gap:.35rem;cursor:pointer;flex:1">
    <input type="checkbox" id="svc-en-%s"%s style="accent-color:var(--accent);cursor:pointer" onchange="toggleService('%s',this.checked)">
    <span style="font-size:.82rem;color:var(--text)">%s</span>
    <span style="font-size:.68rem;color:var(--text-muted)">%s</span>
  </label>
  %s
</div>
%s
</div>`,
			sv.ID,
			sv.ID, enabledChecked, sv.ID,
			sv.Name, sv.ID,
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
    <option value="">— 모델 선택 또는 직접 입력 —</option>
  </select>
  <input id="%s-mdl" type="text" data-i18n-ph="ph_mdl" placeholder="gemini-2.5-flash" oninput="document.getElementById('%s-mdl-sel').value=''">
  <label data-i18n="lbl_description">설명</label>
  <input id="%s-desc" type="text" data-i18n-ph="ph_desc">
  <label data-i18n="lbl_ip_whitelist">허용 IP</label>
  <input id="%s-ipwl" type="text" placeholder="192.168.0.1, 10.0.0.0/24">
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
	// ID 필드는 편집 시 읽기 전용
	body = strings.Replace(body,
		`<input id="ec-id" type="text" placeholder="my-bot">`,
		`<input id="ec-id" type="text" readonly style="opacity:.6">`,
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

// buildAddServiceModal: 커스텀 서비스 추가 모달
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
