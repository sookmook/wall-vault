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

	css := buildCSS(t)
	bots := buildBotsCard(proxies)
	keyCard := buildKeysCard(keys)
	clientCard := buildClientsCard(clients)
	js := buildJS()

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
<div class="header">
  <h1>🔐 wall-vault</h1>
  <p>AI 프록시 키 금고 대시보드</p>
  <span class="badge" id="sse-badge">● 연결 중...</span>
</div>
<div class="grid">`)
	sb.WriteString(bots)
	sb.WriteString(keyCard)
	sb.WriteString(clientCard)
	sb.WriteString(buildAddClientModal())
	sb.WriteString(buildAddKeyModal())
	sb.WriteString(`</div>
<div class="footer">
  wall-vault v0.1.0 — <a href="https://github.com/sookmook/wall-vault">github.com/sookmook/wall-vault</a>
  &nbsp;|&nbsp; 테마: `)
	sb.WriteString(t.Name)
	sb.WriteString(` &nbsp;|&nbsp; <span id="clock"></span>
</div>
<div class="sse-indicator" id="sse-status">SSE: 연결 중...</div>
<script>`)
	sb.WriteString(js)
	sb.WriteString(`</script>
</body>
</html>`)
	return sb.String()
}

func buildCSS(t *theme.Theme) string {
	return `:root {` + t.CSSVars() + `}
*{box-sizing:border-box;margin:0;padding:0}
body{background:var(--bg);color:var(--text);font-family:'Courier New',monospace;padding:1.5rem;min-height:100vh}
a{color:var(--accent);text-decoration:none}
.header{text-align:center;margin-bottom:1.5rem;padding-bottom:1rem;border-bottom:1px solid var(--border)}
.header h1{color:var(--accent);font-size:1.4rem;letter-spacing:2px}
.header p{color:var(--text-muted);font-size:.8rem;margin-top:.3rem}
.badge{display:inline-block;background:var(--surface);border:1px solid var(--green);color:var(--green);padding:.15rem .6rem;border-radius:4px;font-size:.75rem;margin-top:.4rem}
.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(320px,1fr));gap:1rem;margin-bottom:1rem}
.card{background:var(--surface);border:1px solid var(--border);border-radius:6px;padding:1.2rem}
.card h2{color:var(--accent);font-size:.95rem;margin-bottom:.8rem;display:flex;justify-content:space-between;align-items:center}
.card h2 .count{color:var(--text-muted);font-size:.8rem}
.row{display:flex;justify-content:space-between;align-items:center;margin:.3rem 0;font-size:.82rem;gap:.5rem}
.label{color:var(--text-muted);flex-shrink:0}
.val{color:var(--text);text-align:right;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.key-item{margin:.5rem 0}
.key-header{display:flex;justify-content:space-between;font-size:.78rem;margin-bottom:.2rem}
.key-label{color:var(--text)}
.key-meta{color:var(--text-muted);font-size:.72rem}
.bar-track{background:#ffffff10;border-radius:3px;height:8px;overflow:hidden}
.bar-fill{height:8px;border-radius:3px;transition:width .3s}
.bar-green{background:var(--green)}
.bar-yellow{background:var(--yellow)}
.bar-red{background:var(--red)}
.bar-gray{background:#555}
.bot-card{display:flex;align-items:center;gap:.6rem;padding:.4rem 0;border-bottom:1px solid var(--border)}
.bot-card:last-child{border-bottom:none}
.dot{width:8px;height:8px;border-radius:50%;flex-shrink:0}
.dot-green{background:var(--green)}
.dot-red{background:var(--red)}
.dot-gray{background:#555}
.bot-info{flex:1;min-width:0}
.bot-name{font-size:.82rem;color:var(--text)}
.bot-detail{font-size:.72rem;color:var(--text-muted);margin-top:.1rem;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.sse-indicator{position:fixed;bottom:1rem;right:1rem;font-size:.72rem;color:var(--text-muted);background:var(--surface);border:1px solid var(--border);padding:.3rem .7rem;border-radius:4px}
.model-form{margin-top:.8rem}
.model-form select,.model-form input{background:var(--bg);color:var(--text);border:1px solid var(--border);padding:.3rem .6rem;border-radius:4px;font-size:.8rem;width:100%;margin-bottom:.4rem;font-family:inherit}
.btn{background:var(--accent);color:#fff;border:none;padding:.3rem .8rem;border-radius:4px;cursor:pointer;font-size:.8rem;font-family:inherit}
.btn:hover{background:var(--accent-hover)}
.btn-sm{background:transparent;color:var(--accent);border:1px solid var(--accent);padding:.15rem .5rem;border-radius:4px;cursor:pointer;font-size:.72rem;font-family:inherit;margin-left:.5rem}
.btn-sm:hover{background:var(--accent);color:#fff}
.btn-del{background:transparent;color:var(--text-muted);border:none;cursor:pointer;font-size:.72rem;padding:.05rem .3rem;line-height:1;border-radius:3px}
.btn-del:hover{color:var(--red)}
.footer{text-align:center;color:var(--text-muted);font-size:.72rem;margin-top:1.5rem;padding-top:.8rem;border-top:1px solid var(--border)}
.modal-overlay{display:none;position:fixed;inset:0;background:#00000088;z-index:100;align-items:center;justify-content:center}
.modal-overlay.open{display:flex}
.modal{background:var(--surface);border:1px solid var(--border);border-radius:8px;padding:1.5rem;min-width:320px;max-width:90vw}
.modal h3{color:var(--accent);margin-bottom:1rem;font-size:1rem}
.modal label{display:block;color:var(--text-muted);font-size:.78rem;margin:.6rem 0 .2rem}
.modal input,.modal select{background:var(--bg);color:var(--text);border:1px solid var(--border);padding:.35rem .6rem;border-radius:4px;font-size:.82rem;width:100%;font-family:inherit}
.modal-btns{display:flex;gap:.6rem;margin-top:1rem;justify-content:flex-end}
.msg{font-size:.78rem;margin-top:.5rem;min-height:1rem}`
}

func buildJS() string {
	return `
// 시계
function updateClock(){
  document.getElementById('clock').textContent = new Date().toLocaleTimeString('ko-KR');
}
setInterval(updateClock, 1000); updateClock();

// SSE 연결
let es;
function connectSSE() {
  es = new EventSource('/api/events');
  es.onopen = () => {
    document.getElementById('sse-status').textContent = 'SSE: 연결됨';
    document.getElementById('sse-badge').textContent = '● 실행 중';
    document.getElementById('sse-badge').style.borderColor = 'var(--green)';
    document.getElementById('sse-badge').style.color = 'var(--green)';
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
    document.getElementById('sse-status').textContent = 'SSE: 재연결 중...';
    es.close();
    setTimeout(connectSSE, 3000);
  };
}
connectSSE();

// Admin Token 헬퍼
function getAdminToken() {
  let token = localStorage.getItem('wv_admin_token');
  if (!token) {
    token = prompt('Admin Token (저장됨):');
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
  if (!key) { document.getElementById('ak-msg').textContent = '키를 입력하세요'; return; }
  document.getElementById('ak-msg').textContent = '추가 중...';
  fetch('/admin/keys', {
    method: 'POST',
    headers: {'Content-Type':'application/json','Authorization':'Bearer '+token},
    body: JSON.stringify({service:svc, key:key, label:label, daily_limit:limit})
  }).then(r => r.json()).then(d => {
    if (d.error) {
      if (d.error === 'unauthorized') { clearAdminToken(); alert('토큰 오류'); }
      else document.getElementById('ak-msg').textContent = '오류: '+d.error;
    } else {
      closeAddKey();
      setTimeout(() => location.reload(), 500);
    }
  }).catch(e => { document.getElementById('ak-msg').textContent = '오류: '+e; });
}

// 키 삭제
function deleteKey(id) {
  if (!confirm('이 API 키를 삭제하시겠습니까?')) return;
  const token = getAdminToken();
  if (!token) return;
  fetch('/admin/keys/'+id, {
    method: 'DELETE',
    headers: {'Authorization':'Bearer '+token}
  }).then(r => r.json()).then(d => {
    if (d.error) {
      if (d.error === 'unauthorized') { clearAdminToken(); alert('토큰 오류'); }
      else alert('오류: '+d.error);
    } else { location.reload(); }
  });
}

// 클라이언트 추가 모달
function openAddClient() {
  document.getElementById('modal-addclient').classList.add('open');
  document.getElementById('ac-msg').textContent = '';
  ['ac-id','ac-name','ac-token','ac-model'].forEach(id => document.getElementById(id).value = '');
}
function closeAddClient() {
  document.getElementById('modal-addclient').classList.remove('open');
}
function submitAddClient() {
  const token = getAdminToken();
  if (!token) return;
  const id = document.getElementById('ac-id').value.trim();
  const name = document.getElementById('ac-name').value.trim();
  const clientToken = document.getElementById('ac-token').value.trim();
  const svc = document.getElementById('ac-service').value;
  const model = document.getElementById('ac-model').value.trim();
  if (!id) { document.getElementById('ac-msg').textContent = 'ID를 입력하세요'; return; }
  document.getElementById('ac-msg').textContent = '추가 중...';
  fetch('/admin/clients', {
    method: 'POST',
    headers: {'Content-Type':'application/json','Authorization':'Bearer '+token},
    body: JSON.stringify({id:id, name:name, token:clientToken, default_service:svc, default_model:model})
  }).then(r => r.json()).then(d => {
    if (d.error) {
      if (d.error === 'unauthorized') { clearAdminToken(); alert('토큰 오류'); }
      else document.getElementById('ac-msg').textContent = '오류: '+d.error;
    } else {
      if (d.token) alert('클라이언트 토큰 (저장하세요):\n\n'+d.token);
      closeAddClient();
      setTimeout(() => location.reload(), 500);
    }
  }).catch(e => { document.getElementById('ac-msg').textContent = '오류: '+e; });
}

// 모델 변경
function changeModel(clientId) {
  const svc = document.getElementById('svc-'+clientId).value;
  const model = document.getElementById('mdl-'+clientId).value.trim();
  if (!model) return alert('모델명을 입력하세요');
  const token = getAdminToken();
  if (!token) return;
  fetch('/admin/clients/'+clientId, {
    method: 'PUT',
    headers: {'Content-Type':'application/json','Authorization':'Bearer '+token},
    body: JSON.stringify({default_service:svc, default_model:model})
  }).then(r => r.json()).then(d => {
    if (d.error) {
      if (d.error === 'unauthorized') { clearAdminToken(); alert('토큰 오류'); }
      else alert('오류: '+d.error);
    } else { location.reload(); }
  });
}`
}

func buildBotsCard(proxies []*ProxyStatus) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<div class="card"><h2>🤖 봇 상태 <span class="count">%d개</span></h2>`, len(proxies)))
	if len(proxies) == 0 {
		sb.WriteString(`<div style="color:var(--text-muted);font-size:.82rem">연결된 봇 없음</div>`)
	}
	for _, p := range proxies {
		age := time.Since(p.UpdatedAt)
		dotClass := "dot-green"
		if age >= 3*time.Minute {
			dotClass = "dot-gray"
		}
		var ago string
		if age < time.Minute {
			ago = fmt.Sprintf("%.0f초 전", age.Seconds())
		} else {
			ago = fmt.Sprintf("%.0f분 전", age.Minutes())
		}
		sb.WriteString(fmt.Sprintf(
			`<div class="bot-card"><div class="dot %s"></div><div class="bot-info"><div class="bot-name">%s <span style="color:var(--text-muted);font-size:.72rem">%s</span></div><div class="bot-detail">%s / %s — %s</div></div></div>`,
			dotClass, p.ClientID, p.Version, p.Service, p.Model, ago,
		))
	}
	sb.WriteString(`</div>`)
	return sb.String()
}

func buildKeysCard(keys []*APIKey) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		`<div class="card"><h2>🔑 API 키 <span class="count">%d개</span><button class="btn-sm" onclick="openAddKey()">+ 추가</button></h2>`,
		len(keys),
	))

	if len(keys) == 0 {
		sb.WriteString(`<div style="color:var(--text-muted);font-size:.82rem">등록된 키 없음</div></div>`)
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
					barPct = 3
				}
			} else if maxU > 0 {
				barPct = k.TodayUsage * 80 / maxU
			} else {
				barPct = 3
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
				`<div class="key-item"><div class="key-header"><span class="key-label">%s%s</span><span style="display:flex;align-items:center;gap:.4rem"><span class="key-meta">%s</span><button class="btn-del" onclick="deleteKey('%s')" title="삭제">✕</button></span></div><div class="bar-track"><div class="bar-fill %s" style="width:%dpx;max-width:100%%"></div></div></div>`,
				statusIcon, label, meta, k.ID, barClass, barPct,
			))
		}
		sb.WriteString(`</div>`)
	}
	sb.WriteString(`</div>`)
	return sb.String()
}

func buildClientsCard(clients []*Client) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		`<div class="card"><h2>⚙️ 클라이언트 <span class="count">%d개</span><button class="btn-sm" onclick="openAddClient()">+ 추가</button></h2>`,
		len(clients),
	))
	if len(clients) == 0 {
		sb.WriteString(`<div style="color:var(--text-muted);font-size:.82rem">등록된 클라이언트 없음</div></div>`)
		return sb.String()
	}
	for _, c := range clients {
		allowed := strings.Join(c.AllowedServices, ", ")
		if allowed == "" {
			allowed = "모두"
		}
		sb.WriteString(fmt.Sprintf(`<div style="margin-bottom:1rem;padding-bottom:1rem;border-bottom:1px solid var(--border)">
<div class="row"><span class="label">ID</span><span class="val">%s</span></div>
<div class="row"><span class="label">서비스</span><span class="val">%s</span></div>
<div class="row"><span class="label">모델</span><span class="val">%s</span></div>
<div class="model-form">
<select id="svc-%s"><option value="google"%s>google</option><option value="openrouter"%s>openrouter</option><option value="ollama"%s>ollama</option></select>
<input id="mdl-%s" type="text" placeholder="모델명" value="%s">
<button class="btn" onclick="changeModel('%s')">적용</button>
</div></div>`,
			c.ID, c.DefaultService, c.DefaultModel,
			c.ID, sel(c.DefaultService == "google"), sel(c.DefaultService == "openrouter"), sel(c.DefaultService == "ollama"),
			c.ID, c.DefaultModel, c.ID,
		))
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

func buildAddClientModal() string {
	return `
<div class="modal-overlay" id="modal-addclient" onclick="if(event.target===this)closeAddClient()">
<div class="modal">
  <h3>⚙️ 클라이언트 추가</h3>
  <label>ID (영문·숫자·하이픈)</label>
  <input id="ac-id" type="text" placeholder="my-bot">
  <label>이름</label>
  <input id="ac-name" type="text" placeholder="My Bot">
  <label>토큰 (빈칸이면 자동 생성)</label>
  <input id="ac-token" type="text" placeholder="자동 생성" autocomplete="off">
  <label>기본 서비스</label>
  <select id="ac-service">
    <option value="google">google</option>
    <option value="openrouter">openrouter</option>
    <option value="ollama">ollama</option>
  </select>
  <label>기본 모델</label>
  <input id="ac-model" type="text" placeholder="gemini-2.5-flash">
  <div class="msg" id="ac-msg"></div>
  <div class="modal-btns">
    <button class="btn" style="background:var(--surface);color:var(--text)" onclick="closeAddClient()">취소</button>
    <button class="btn" onclick="submitAddClient()">추가</button>
  </div>
</div>
</div>`
}

func buildAddKeyModal() string {
	return `
<div class="modal-overlay" id="modal-addkey" onclick="if(event.target===this)closeAddKey()">
<div class="modal">
  <h3>🔑 API 키 추가</h3>
  <label>서비스</label>
  <select id="ak-service">
    <option value="google">google</option>
    <option value="openrouter">openrouter</option>
    <option value="ollama">ollama</option>
  </select>
  <label>API 키</label>
  <input id="ak-key" type="password" placeholder="AIzaSy... 또는 sk-or-..." autocomplete="off">
  <label>레이블 (선택)</label>
  <input id="ak-label" type="text" placeholder="my-key-1">
  <label>일일 한도 (0 = 무제한)</label>
  <input id="ak-limit" type="number" value="0" min="0">
  <div class="msg" id="ak-msg"></div>
  <div class="modal-btns">
    <button class="btn" style="background:var(--surface);color:var(--text)" onclick="closeAddKey()">취소</button>
    <button class="btn" onclick="submitAddKey()">추가</button>
  </div>
</div>
</div>`
}
