# Izikhombisi Zomsebenzisi ze-wall-vault
*(Kucishwa kamuva: 2026-04-05 — v0.1.22)*

---

## Okuqukethwe

1. [Yini i-wall-vault?](#yini-i-wall-vault)
2. [Ukufakela](#ukufakela)
3. [Ukuqala Okokuqala (i-setup wizard)](#ukuqala-okokuqala)
4. [Ukubhalisa Ukhiye we-API](#ukubhalisa-ukhiye-we-api)
5. [Ukusebenzisa i-Proxy](#ukusebenzisa-i-proxy)
6. [Idashubhodi ye-Vault](#idashubhodi-ye-vault)
7. [Imodi Eyahlukaniswayo (Ama-Bot Amaningi)](#imodi-eyahlukaniswayo-ama-bot-amaningi)
8. [Ukusethwa Kokuqala Ngokuzenzakalela](#ukusethwa-kokuqala-ngokuzenzakalela)
9. [I-Doctor: Isihloli](#i-doctor-isihloli)
10. [Izinguquko Zemvelo](#izinguquko-zemvelo)
11. [Ukuxazulula Izinkinga](#ukuxazulula-izinkinga)

---

## Yini i-wall-vault?

**i-wall-vault = i-proxy ye-AI (isithunywa) + indawo yokugcina izikhiye ye-API, eyenzelwe OpenClaw**

Ukusebenzisa izinsizakalo ze-AI kudinga **ukhiye we-API** — okuwuphawu ledijithali (ipasi ledijithali) olufakazela ukuthi unelungelo lokusebenzisa leso nsizakalo. Lezi zikhiye zinomkhawulo wemibuzo ngosuku, futhi zingaba sengozini uma zingagcinwa kahle.

I-wall-vault igcina lezi zikhiye endaweni evikelekile (i-vault), bese isebenza njenge-**proxy** (isithunywa esimele) phakathi kuka-OpenClaw nezinsizakalo ze-AI. Nje nje: OpenClaw ixhumana no-wall-vault kuphela — wonke umsebenzi onzima wenziwa yi-wall-vault.

Izinkinga i-wall-vault ezixazululayo:

- **Ukuphendulana kwazikhiye ngokuzenzakalela**: Uma ukhiye owodwa ufika emkhawulweni wawo noma ubanjwe okwesikhashana (i-cooldown), i-wall-vault iyaguqukelela ngokuthula kukhiye olulandelayo. OpenClaw iqhubeka ngaphandle kokuma.
- **Ukushintsha Kwezinsizakalo Ngokuzenzakalela (i-fallback)**: Uma Google ingaphenduli, iyaguqukelela ku-OpenRouter; uma nako kungasebenzi, iyaguqukela ku-Ollama (i-AI ekhompyutheni yakho). Iseshini ayipheli. Uma insizakalo yokuqala ibuyela esimweni esijwayelekile, iyashintsha ibuyele kuyo ngokuzenzakalela kusukela esicwelweni esilandelayo (v0.1.18+).
- **Ukuvumelaniswa Ngesikhathi Sangempela (SSE)**: Uma ushintsha imodeli ku-vault dashboard, izinguquko zibonakala ku-OpenClaw phakathi kwamasekhendi angu-1 no-3. I-SSE (Server-Sent Events) yiteknoloji lapho iserver ithumela iziguquko zangempela-isikhathi ngokuqondile kuklayenti.
- **Izaziso Zangempela-Isikhathi**: Izikhombisi zokuphela kwazikhiye noma ukuphazamiseka kwensizakalo zibonakala ngokuphazima ngephansi kwe-TUI (isikrini setheminali) ye-OpenClaw.

> 💡 **Claude Code, Cursor, ne-VS Code** nako kunganxuswa, kodwa inhloso enkulu ye-wall-vault ukusebenzisana no-OpenClaw.

```
OpenClaw (Isikrini Setheminali TUI)
        │
        ▼
  i-wall-vault proxy (:56244)   ← Ukuphatha izikhiye, ukuqondisa, i-fallback, imicimbi
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (Amamodeli angaphezu kuka-340)
        └─ Ollama (Ikhompyutha yakho, indawo yokugcina yokugcina)
```

---

## Ukufakela

### Linux / macOS

Vula itheminali bese ulayisha lezi miqulu yemiyalo:

```bash
# Linux (ikhompyutha evamile, iserver — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (Mac M1/M2/M3)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Landa ifayela ku-inthanethi.
- `chmod +x` — Yenza ifayela eliladiwe libe "elingasethenziswa". Uma weqa lesi sinyathelo, uzothola iphutha elithi "Akukho Imvume".

### Windows

Vula i-PowerShell (njenge-admin) bese ufaka le miyalo:

```powershell
# Ukudownloada
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Engeza ku-PATH (isebenza ngemuva kokuvula kabusha i-PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **I-PATH iyini?** Iqoqo lamafolda ikhompyutha ekwahlola kuwo uma icinga umyalo. Uma wengeza ku-PATH, ungalayisha `wall-vault` noma yikuphi ngaphandle kokuchaza indlela ephelele.

### Ukwakha Kusuka Emthonjeni (Kubakhiqizi)

Lokhu kuphela kubakhiqizi abafakele indawo yokusebenza ye-Go.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (inguqulo: v0.1.6.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Inguqulo nesikhathi sokwakha**: Uma wakha ngo-`make build`, inguqulo ikhiqizwa ngokuzenzakalela ngobukhuni obufana no-`v0.1.6.20260314.231308`. Uma wenza `go build ./...` ngokuqondile, inguqulo ibonakala njenge-`"dev"` kuphela.

---

## Ukuqala Okokuqala

### Ukuqalisa i-setup Wizard

Ngemuva kokufakela, qala **i-wizard yokusetha** nge-:

```bash
wall-vault setup
```

I-wizard ikusiza ngezinyathelo lezi:

```
1. Ukukhetha ulimi (izilimi ezingu-10 kuhlanganisa isiNgisi)
2. Ukukhetha isihloko (light / dark / gold / cherry / ocean)
3. Imodi yokusebenza — standalone (wena wedwa) noma distributed (amakhompyutha amaningi)
4. Igama le-bot — igama elizobonakala ku-dashboard
5. Ukusetha izikhibhalo — default: proxy 56244, vault 56243 (chofoza i-Enter uma ungakuthintanga)
6. Ukukhetha insizakalo ye-AI — Google / OpenRouter / Ollama
7. Ukusetha izinhluzo zezinsizakalo
8. Ukusetha izikhiye ze-admin — iphasiwedi evikela izici zokuphatha ku-dashboard; ingakhiqizwa ngokuzenzakalela
9. Iphasiwedi yokukhipha izikhiye ze-API — uma ufuna ukugcina izikhiye ngokuphepha okwengeziwe (kuzokhetha)
10. Indawo yokugcina ifayela lezilungiselelo
```

> ⚠️ **Khumbula izikhiye ze-admin zakho.** Uzidinga ngemuva ukwengeza izikhiye noma ukushintsha izilungiselelo ku-dashboard. Uma uzilahlekelwa, uzodinga ukuhlela ngokuqondile ifayela lezilungiselelo.

Uma i-wizard iqedile, ifayela lezilungiselelo `wall-vault.yaml` likhiqizwa ngokuzenzakalela.

### Ukuqalisa

```bash
wall-vault start
```

Iserver ezimbili ziqalisa ngasikhathi sinye:

- **Proxy** (`http://localhost:56244`) — isithunywa esixhumanisa OpenClaw nezinsizakalo ze-AI
- **Vault** (`http://localhost:56243`) — Ukuphatha izikhiye we-API ne-dashboard yewebhu

Vula `http://localhost:56243` kwidwebhulayiza ukuze ubone i-dashboard ngokuphazima.

---

## Ukubhalisa Ukhiye we-API

Kuneendlela ezine zokubhalisa ukhiye we-API (ipasi ledijithali). **Abaqalayo kunconywa Indlela 1 (izinguquko zemvelo).**

### Indlela 1: Izinguquko Zemvelo (Inconywa — Elilula Kakhulu)

Izinguquko zemvelo ziyizinqolobane ezisethiwe ngaphambili izinhlelo ezizihlola ekuqaleni. Faka le miyalo etheminalini:

```bash
# Bhalisa ukhiye we-Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Bhalisa ukhiye we-OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Qalisa ngemuva kokubhalisa
wall-vault start
```

Uma uneizikhiye eziningi, hlukanisa nge-comma (,). I-wall-vault izozishintshashintsha ngokuzenzakalela (round-robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Iseluleko**: Umyalo we-`export` usebenza kuphela esikhathini samanje setheminali. Ukuze zihlale ngemuva kokuvula kabusha ikhompyutha, engeza lezo zinhla ku-`~/.bashrc` noma `~/.zshrc`.

### Indlela 2: Isixhobo se-Dashboard UI (Ukuchofoza Ngemouse)

1. Vula `http://localhost:56243` kwidwebhulayiza
2. Chofoza inkinobho ethi `[+ Engeza]` enkhadini ethi **🔑 API Keys**
3. Faka uhlobo lwensizakalo, inani lekhiye, uphawu (igama lememori), nomkhawulo wosuku bese ugcina

### Indlela 3: I-REST API (Ukusetshenziswa Kwezikripti)

I-REST API iyindlela inhlelo ezithumela ngayo idatha ngokusetshenziswa kwe-HTTP. Iluleme ukubhalisa ngokuzenzakalela ezikriptini.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer your-admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "main-key",
    "daily_limit": 1000
  }'
```

### Indlela 4: I-proxy Flag (Uhlelo Lokuhlola Okwesikhashana)

Setshenziswa ukuhlola ngezokuvikela ngaphandle kokubhalisa okusemthethweni. Ikhiye yalahleka uma uhlelo luphela.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Ukusebenzisa i-Proxy

### Ukusebenzisa no-OpenClaw (Inhloso Enkulu)

Iqiniso lokuseta i-OpenClaw ukuze ixhumane nezinsizakalo ze-AI nge-wall-vault:

Vula `~/.openclaw/openclaw.json` bese ufaka okulandelayo:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // vault agent token
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // Free 1M context
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Indlela elula kakhulu**: Chofoza inkinobho ethi **🦞 OpenClaw Config Copy** enkhadini ye-agent ku-dashboard — isigcebi esinezinsipho zithunyelwe esigcwele sikhopyiwa ku-clipboard. Namathisela kuphela.

**I-`wall-vault/` phambi kwegama lemodeli ixhumanisa kuphi?**

I-wall-vault ihlola igama lemodeli ukuze inqume ukuthumela isicelo kowuphi umhlinzeki we-AI:

| Ifomathi Yemodeli | Insizakalo Exhunyiwe |
|-------------------|---------------------|
| `wall-vault/gemini-*` | Ixhumanisa ngokuqondile ku-Google Gemini |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | Ixhumanisa ngokuqondile ku-OpenAI |
| `wall-vault/claude-*` | Ixhumanisa ku-Anthropic nge-OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (1M token context mahhala) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | Ixhumanisa ku-OpenRouter |
| `google/model-name`, `openai/model-name`, `anthropic/model-name`, njll. | Ixhumanisa ngokuqondile kuleso nsizakalo |
| `custom/google/model-name`, `custom/openai/model-name`, njll. | Isusa `custom/` bese iqondisa kabusha |
| `model-name:cloud` | Isusa `:cloud` bese ixhumanisa ku-OpenRouter |

> 💡 **I-context iyini?** Ingxenye yenkulumo i-AI engayikhumbula ngasikhathi sinye. Ama-token angu-1M (isigidi) anikeza i-AI amandla okuphatha izingxoxo ezinde kakhulu futhi imibhalo emikhulu ngokuphazima.

### Ukuxhuma Ngokuqondile Ngefomathi ye-Gemini API (Ukuxhuma Namathuluzi Akade)

Uma ungathuluzi asebenzisa ngokuqondile i-Google Gemini API, shintshanisa i-URL kuphela ubheke ku-wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Noma, uma ithuluzi likuvumela ukuchaza i-URL ngokuqondile:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Ukusebenzisa ne-OpenAI SDK (Python)

Ungaxhuma i-wall-vault ekhowdini ye-Python ye-AI. Shintsha kuphela `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # wall-vault manages API keys for you
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # provider/model format
    messages=[{"role": "user", "content": "Hello"}]
)
```

### Ukushintsha Imodeli Ngesikhathi Esisebenzayo

Ukushintsha imodeli ye-AI esetshenziswa ngenkathi i-wall-vault isesebenza:

```bash
# Shintsha imodeli ngokuhlola i-proxy ngokuqondile
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# Ngimodi eyahlukaniswayo (ama-bot amaningi), shintsha ku-vault server → iziguquko zibonakala ngokuphazima nge-SSE
curl -X PUT http://localhost:56243/admin/clients/your-bot-id \
  -H "Authorization: Bearer your-admin-token" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Ukubona Uhlu Lwamamodeli Atholakalayo

```bash
# Bona uhlu oluphelele
curl http://localhost:56244/api/models | python3 -m json.tool

# Bona kuphela amamodeli e-Google
curl "http://localhost:56244/api/models?service=google"

# Sesha ngegama (isib. amamodeli anemazwi ethi "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Izincazelo Zamamodeli Abalulekile Ngomsebenzisi:**

| Insizakalo | Amamodeli Abalulekile |
|------------|----------------------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | Angaphezu kuka-346 (Hunter Alpha 1M context mahhala, DeepSeek R1/V3, Qwen 2.5, njll.) |
| Ollama | Ukuhlola ngokuzenzakalela iserver yendawo efakwe ekhompyutheni yakho |

---

## Idashubhodi ye-Vault

Vula `http://localhost:56243` kwidwebhulayiza ukuze ubheke i-dashboard.

**Ukwakheka Kwesikrini:**
- **Ibha Ephezulu (topbar)**: Ilogo, ukukhetha ulimi nesihloko, ukubonisa isimo sokuqhagamshelana kwe-SSE
- **Igridi Yamakhadi**: Amakhadi e-agent, insizakalo, nezinkhiye ze-API ahlelwe njengamathayela

### Ikhadi Lazikhiye ze-API

Ikhadi okuphathwa ngalo izikhiye ze-API ezibhalisiwe ngasikhathi sinye:

- Ibonisa uhlu lwazikhiye ngokwehlukaniswa ngomsebenzisi.
- `today_usage`: Inani lamathokheni aphumelele angalelo langa (izinhlamvu i-AI ezizifunde nezibhale)
- `today_attempts`: Inani lokubizelwa okuphelele kwanamhlanje (okuphumelele + okwehlulekile)
- Bhalisa izikhiye ezintsha ngenkinobho ethi `[+ Engeza]`, futhi susa nge-`✕`.

> 💡 **Ithokheni iyini?** Iyunithi i-AI esebenzisa ukucubungula umbhalo. Ngokwesihlahla, ilingana namazwi e-Ngisi angu-1, noma izinhlamvu ze-Hangul ezingu-1–2. Imali ye-API ibalelwa ngokuvamile ngale mithokheni.

### Ikhadi Le-agent

Ikhadi elibonisa isimo samabhoti (ama-agent) axhunywe ku-wall-vault proxy:

**Isimo Sokuqhagamshelana Sibonakala Ngezigaba Ezine:**

| Ukubonisa | Isimo | Incazelo |
|-----------|-------|----------|
| 🟢 | Iyasebenza | I-proxy isebenza kahle |
| 🟡 | Ilinzela | Izimpendulo ziyeza kodwa zinephutha lelanga |
| 🔴 | Ayikho Ku-inthanethi | I-proxy ayiphenduli |
| ⚫ | Ayixhunyiwe / Ivaliwe | I-proxy ayikaxhumananga ne-vault noma ivaliwe |

**Izikhombisi Zamakinobho Angaphansi Kwekhadi Le-agent:**

Uma ubhalisa i-agent, chaza **uhlobo lwe-agent** ukuze amakinobho afanele avele ngokuzenzakalela.

---

#### 🔘 Iminobho Yokukhopiya Izilungiselelo — Ikhiqiza Ngokuzenzakalela Izilungiselelo Zokuqhagamshelana

Chofoza inkinobho ukuze uthole isigcebi sezilungiselelo esinezinsipho, ikheli lika-proxy, namulokumodeli ethunyelwe esikhumuloni. Namathisela nje endaweni efanele eshownwe etafuleni elingezansi.

| Inkinobho | Uhlobo Lwe-agent | Namathisela Lapho |
|-----------|-----------------|-------------------|
| 🦞 OpenClaw Config Copy | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 NanoClaw Config Copy | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Claude Code Config Copy | `claude-code` | `~/.claude/settings.json` |
| ⌨ Cursor Config Copy | `cursor` | Cursor → Settings → AI |
| 💻 VSCode Config Copy | `vscode` | `~/.continue/config.json` |

**Isibonelo — Uhlobo lwe-Claude Code lukhopiye lokhu:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "this-agent-token"
}
```

**Isibonelo — Uhlobo lwe-VSCode (Continue):**

```json
// ~/.continue/config.json
{
  "models": [{
    "title": "wall-vault proxy",
    "provider": "openai",
    "model": "gemini-2.5-flash",
    "apiBase": "http://192.168.1.20:56244/v1",
    "apiKey": "this-agent-token"
  }]
}
```

**Isibonelo — Uhlobo lwe-Cursor:**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : this-agent-token

// Or as environment variables:
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=this-agent-token
```

> ⚠️ **Uma ukhophi we-clipboard ungasebenzi**: Inqubomgomo yokuphepha yedwebhulayiza ingavimba ukukhopisha. Uma ibhokisi lombhalo livuleka ngokuvuthuza, chofoza Ctrl+A ukuze ukhethe konke bese Ctrl+C ukukhopisha.

---

#### ⚡ Inkinobho Yokusebenzisa Ngokuzenzakalela — Chofoza Kanye, Izilungiselelo Ziqediwe

Uma uhlobo lwe-agent lungu-`cline`, `claude-code`, `openclaw`, `nanoclaw`, inkinobho ethi **⚡ Sebenzisa Izilungiselelo** izovela ekhadini le-agent. Uma uchofoza le nkinobho, ifayela lezilungiselelo zendawo ye-agent lishintshelwa ngokuzenzakalela.

| Inkinobho | Uhlobo Lwe-agent | Ifayela Elishintshelwayo |
|-----------|-----------------|------------------------|
| ⚡ Sebenzisa Izilungiselelo ze-Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Sebenzisa Izilungiselelo ze-Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Sebenzisa Izilungiselelo ze-OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Sebenzisa Izilungiselelo ze-NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Le nkinobho ithumela isicelo ku-**localhost:56244** (i-proxy yendawo). I-proxy kumele isebenze kuleyo khompyutha ukuze isebenze.

---

#### 🔀 Hudula Bese Uyeka Ukuhlela Amakhadi (v0.1.17)

Ungahudula **amakhadi** e-agent ku-dashboard ukuze uwahlele ngokulandelana okufunayo.

1. Chofoza ubambe ikhadi le-agent bese uyahudula
2. Liyeke phezu kwekhadi endaweni oyifunayo — ukulandelana kuzoshintsha ngokushesha
3. Ukulandelana okusha **kugcinwa ku-server ngokushesha** futhi kuzohlala noma ususa ikhasi kabusha

> 💡 Amadivayisi okuthinta (iselula/ithebhulethi) awakasekelwa okwamanje. Sicela usebenzise isiphequluli sekhompyutha.

---

#### 🔄 Ukuvumelaniswa Kwemodeli Ngezinhlangothi Ezimbili (v0.1.16)

Uma ushintsha imodeli ye-agent ku-vault dashboard, izilungiselelo zendawo ze-agent zishintshelwa ngokuzenzakalela.

**Nge-Cline:**
- Shintsha imodeli ku-vault → umcimbi we-SSE → i-proxy ishintshe ingxenye yemodeli ku-`globalState.json`
- Izingxenye ezishintshelwayo: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` nokhiye we-API akuthintwa
- **Kudingeka ukuvuselela i-VS Code (`Ctrl+Alt+R` noma `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - Ngoba i-Cline ayisifundi kabusha ifayela lezilungiselelo ngenkathi isebenza

**Nge-Claude Code:**
- Shintsha imodeli ku-vault → umcimbi we-SSE → i-proxy ishintshe ingxenye ethi `model` ku-`settings.json`
- Ifuna ngokuzenzakalela izindlela ze-WSL ne-Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Uhlangothi oluphambene (i-agent → vault):**
- Uma i-agent (Cline, Claude Code, njll.) ithumela isicelo nge-proxy, i-proxy ifaka ulwazi lwensizakalo·imodeli yeklayenti ku-heartbeat
- Ikhadi le-agent ku-vault dashboard libonisa insizakalo/imodeli esetshenziswa njengamanje ngesikhathi sangempela

> 💡 **Okusemqoka**: I-proxy ihlonza i-agent ngezikhiye ze-Authorization esicelweni, bese iqondisa ngokuzenzakalela ensizakalweni/emodeli esethwe ku-vault. Noma i-Cline noma i-Claude Code ithumele igama lemodeli elinomahluko, i-proxy iyokubhala ngaphezulu ngezilungiselelo ze-vault.

---

### Ukusebenzisa i-Cline ku-VS Code — Umhlahlandlela Obanzi

#### Isinyathelo 1: Faka i-Cline

Faka **Cline** (ID: `saoudrizwan.claude-dev`) ku-VS Code Extension Marketplace.

#### Isinyathelo 2: Bhalisa i-agent ku-Vault

1. Vula i-vault dashboard (`http://IP-ye-vault:56243`)
2. Chofoza **+ Engeza** esigabeni **se-agent**
3. Gcwalisa loku okulandelayo:

| Ingxenye | Inani | Incazelo |
|----------|-------|----------|
| ID | `my_cline` | Isazisi esiyingqayizivele (ngesiNgisi, ngaphandle kwezikhala) |
| Igama | `My Cline` | Igama elizovela ku-dashboard |
| Uhlobo Lwe-agent | `cline` | ← Kumele ukhethe `cline` |
| Insizakalo | Khetha insizakalo ozoyisebenzisa (isib: `google`) | |
| Imodeli | Faka imodeli ozoyisebenzisa (isib: `gemini-2.5-flash`) | |

4. Chofoza **Gcina** bese izikhiye zikhiqizwa ngokuzenzakalela

#### Isinyathelo 3: Xhuma i-Cline

**Indlela A — Ukusebenzisa Ngokuzenzakalela (Inconywa)**

1. Qiniseka ukuthi i-wall-vault **proxy** isebenza kuleyo khompyutha (`localhost:56244`)
2. Chofoza inkinobho ethi **⚡ Sebenzisa Izilungiselelo ze-Cline** ekhadini le-agent ku-dashboard
3. Uma ubona umyalezo othi "Izilungiselelo zisebenziswe!" usuphumelele
4. Vuselela i-VS Code (`Ctrl+Alt+R`)

**Indlela B — Ukusetha Ngesandla**

Vula izilungiselelo (⚙️) ku-sidebar ye-Cline:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://ikheli-le-proxy:56244/v1`
  - Ikhompyutha efanayo: `http://localhost:56244/v1`
  - Elinye ikhompyutha njenge-Mac Mini: `http://192.168.1.20:56244/v1`
- **API Key**: izikhiye ezitholakale ku-vault (khopisha ekhadini le-agent)
- **Model ID**: imodeli esethwe ku-vault (isib: `gemini-2.5-flash`)

#### Isinyathelo 4: Qinisekisa

Thumela noma yimuphi umyalezo engxoxweni ye-Cline. Uma isebenza kahle:
- Ikhadi le-agent ku-vault dashboard lizobonisa **indilinga eluhlaza (● Iyasebenza)**
- Ikhadi lizobonisa insizakalo/imodeli yamanje (isib: `google / gemini-2.5-flash`)

#### Ukushintsha Imodeli

Uma ufuna ukushintsha imodeli ye-Cline, shintsha ku-**vault dashboard**:

1. Shintsha orodha yensizakalo/yemodeli ekhadini le-agent
2. Chofoza **Sebenzisa**
3. Vuselela i-VS Code (`Ctrl+Alt+R`) — igama lemodeli ku-footer ye-Cline lizoshintshelwa
4. Isicelo esilandelayo sizosebenzisa imodeli entsha

> 💡 Empeleni i-proxy ihlonza isicelo se-Cline ngezikhiye bese iqondisa kwimodeli esethwe ku-vault. Noma ungavuseleli i-VS Code, **imodeli esetshenziswayo iguquka ngokuphazima** — ukuvuselela kungokokushintshelwa kwegama lemodeli ku-UI ye-Cline kuphela.

#### Ukubona Ukukhishwa Kokuqhagamshelana

Uma uvala i-VS Code, ikhadi le-agent ku-vault dashboard liguqukela kuphuzi (ilindile) ngemuva **kwamasekhondi angu-90**, bese liba bomvu (aye ikho ku-inthanethi) ngemuva **kwemizuzu engu-3**. (Kusukela ku-v0.1.18, ukuhlola isimo njalo ngamasekhondi angu-15 kwenza ukutholwa kokungaxhunyiwe kube ngokushesha.)

#### Ukuxazulula Izinkinga

| Isimpawu | Imbangela | Isixazululo |
|----------|-----------|-------------|
| I-Cline ibonisa "ukuxhuma kuhlulekile" | I-proxy ayisebenzi noma ikheli alikho kahle | Hlola i-proxy ngo-`curl http://localhost:56244/health` |
| Indilinga eluhlaza ayibonakali ku-vault | Ukhiye we-API (izikhiye) awukasetshwa | Chofoza inkinobho ethi **⚡ Sebenzisa Izilungiselelo ze-Cline** futhi |
| Igama lemodeli ku-footer ye-Cline aliguquki | I-Cline igcina izilungiselelo ku-cache | Vuselela i-VS Code (`Ctrl+Alt+R`) |
| Igama lemodeli elingalungile libonakala | Iphutha elidala (lilungiswe ku-v0.1.16) | Shintshelela i-proxy ku-v0.1.16 noma ngaphezulu |

---

#### 🟣 Inkinobho Yokukhopiya Umyalo Wokusabalalisa — Ukufakela Ekhompyutheni Entsha

Isetshenziswa ekufakeleni i-wall-vault proxy ekhompyutheni entsha nokuyixhuma ne-vault. Chofoza inkinobho ukukhopisha isikripti sifakelo esiphelele. Namathisela itheminalini yekhompyutha entsha bese uqalisa — okulandelayo kuphathwa ngesikhathi sinye:

1. Ukufakela i-wall-vault binary (kuphula uma isifakiwe)
2. Ukubhalisa ngokuzenzakalela isevisi somsebenzisi we-systemd
3. Ukuqalisa isevisi nokuxhuma ngokuzenzakalela ne-vault

> 💡 Isikripti sihlanganisa izikhiye ze-agent nezokuxhuma ne-vault server izilungiselelo, ngakho ngaphandle kokushintsha okuthe xaxa, namathisela bese uqalisa.

---

### Ikhadi Lezinsizakalo

Ikhadi lokuvula noma ukuvala nokuseta izinsizakalo ze-AI:

- Ukushintsha ukusebenza / ukuvala ngokuguquguquka ngensizakalo ngayinye
- Uma ufaka ikheli yeserver ye-AI yendawo (Ollama, LM Studio, vLLM, njll. esebenza ekhompyutheni yakho), i-wall-vault izowathola ngokuzenzakalela amamodeli atholakalayo.
- **Ukubonisa Isimo Sokuqhagamshelana Kwensizakalo Yendawo**: Indingilizi ● eceleni kwegama lensizakalo — **oluhlaza** = ixhunywe, **impunga** = ayixhunyiwe
- **Isikhombisi Sesimo Sensizakalo Yendawo**: Uma uvula ikhasi, izinsizakalo zendawo ezisebenzayo (njenge-Ollama), indingilizi ● iguqulela oluhlaza — kodwa isimo sokukhetha asishintshwa.

> 💡 **Uma insizakalo yendawo isebenza kwelinye ikhompyutha**: Faka i-IP yekhompyutha leyo enkinobheni ye-URL yensizakalo. Isib: `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio)

### Ukufaka Izikhiye Ze-Admin

Uma uzama ukusebenzisa izici ezibalulekile ku-dashboard (njengokwengeza noma ukususa izikhiye), i-popup yokufaka izikhiye ze-admin izovela. Faka izikhiye owazisethayo ku-setup wizard. Uma usezifakile, ziqhubeka kuze kube wuvala idwebhulayiza.

> ⚠️ **Uma ukuqinisekiswa kuhluleka izikhathi ezingu-10 ngaphakathi kwemizuzu engu-15, i-IP yakho ivaliwe okwesikhashana.** Uma ukhohliwe izikhiye zakho, hlola ingxenye ethi `admin_token` kufayela `wall-vault.yaml`.

---

## Imodi Eyahlukaniswayo (Ama-Bot Amaningi)

Uma usebenzisa i-OpenClaw ezikhompyutheni eziningi ngasikhathi sinye, ungashiya **i-vault eyodwa enabelwana ngayo**. Ukuphathwa kwezikhiye kwenziwa endaweni eyodwa — kulula kakhulu.

### Isibonelo Sohlelo

```
[I-Vault Server]
  wall-vault vault    (vault :56243, dashboard)

[WSL Alpha]            [Raspberry Pi Gamma]    [Mac Mini Local]
  wall-vault proxy      wall-vault proxy          wall-vault proxy
  openclaw TUI          openclaw TUI              openclaw TUI
  ↕ SSE sync            ↕ SSE sync                ↕ SSE sync
```

Wonke amabhoti abheka iserver ye-vault phakathi — lapho ushintsha imodeli noma ufaka izikhiye ku-vault, izinguquko zibonakala ngokuphazima kuwo wonke amabhoti.

### Isinyathelo 1: Qalisa I-Vault Server

Qalisa ekhompyutheni ethiwa nge-vault server:

```bash
wall-vault vault
```

### Isinyathelo 2: Bhalisa Amabhoti (Amaklayenti) Ngamunye

Bhalisa ingatho yabo bot ngamunyandwa axhumana ne-vault server:

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer your-admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "Bot A",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### Isinyathelo 3: Qalisa I-proxy Ekhompyutheni Yamabhoti Ngayinye

Ekhompyutheni ngayinye yabot, qalisa i-proxy ngokuchaza ikheli le-vault server nezikhiye:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Shintsha **`192.168.x.x`** ukuze ibe yi-IP yangempela yendawo yekhompyutha ye-vault server. Ungayithola ngokusebenzisa izilungiselelo ze-router noma umyalo othi `ip addr`.

---

## Ukusethwa Kokuqala Ngokuzenzakalela

Uma ukuqalisa i-wall-vault ngesandla ngokuvula ikhompyutha kukhathaza, ibhalisa njengasevisi yohlelo. Uma ibhalisiwe, iqala ngokuzenzakalela ekuqaleni kwekhompyutha.

### Linux — systemd (Iningi Lwelinux)

I-systemd iyisizindalwazi seLinux sokuqalisa nokuphatha izinhlelo ngokuzenzakalela:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Ukubheka amalogi:

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

Isistimu ye-macOS ehlelela izinhlelo ukuqalisa ngokuzenzakalela:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Landa i-NSSM ku-[nssm.cc](https://nssm.cc/download) bese uyengeza ku-PATH.
2. Ku-PowerShell ne-admin rights:

```powershell
wall-vault doctor deploy windows
```

---

## I-Doctor: Isihloli

Umyalo we-`doctor` uyithuluzi **elizihlola futhi elizixazulula ngokwalo** izinkinga ze-wall-vault.

```bash
wall-vault doctor check   # Hlola isimo samanje (funda kuphela, akushintshi lutho)
wall-vault doctor fix     # Lungisa izinkinga ngokuzenzakalela
wall-vault doctor all     # Ukuhlola + Ukulungisa ngesikhathi sinye
```

> 💡 Uma kukhona okungahambisani, qalisa `wall-vault doctor all` kuqala. Ixazulula izinkinga eziningi ngokuzenzakalela.

---

## Izinguquko Zemvelo

Izinguquko zemvelo ziyindlela yokudlulisela izinqolobane kuzo izikhungo. Faka nge-`export VARIABLE=value` etheminalini, noma zinike isevisi yokuziqalisa ngokuzenzakalela ukuze zihlale zisebenza.

| Inguquko | Incazelo | Isibonelo Senani |
|----------|----------|-----------------|
| `WV_LANG` | Ulimi lwe-dashboard | `ko`, `en`, `ja` |
| `WV_THEME` | Isihloko se-dashboard | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Ukhiye we-Google API (izingi eziningi nge-comma) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Ukhiye we-OpenRouter API | `sk-or-v1-...` |
| `WV_VAULT_URL` | Ikheli le-vault server ngimodi eyahlukaniswayo | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Izikhiye ze-authentication zamaklayenti (amabhoti) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Izikhiye ze-admin | `admin-token-here` |
| `WV_MASTER_PASS` | Iphasiwedi yokukhipha izikhiye ze-API | `my-password` |
| `WV_AVATAR` | Indlela yefayela lesithombe se-avatar (i-relative path ngaphansi kwe-`~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Ikheli le-Ollama local server | `http://192.168.x.x:11434` |

---

## Ukuxazulula Izinkinga

### Uma I-proxy Ingaqalisi

Kaningi ikhodi yomgwaqo isesetshenziswa inhlelo enye.

```bash
ss -tlnp | grep 56244   # Bheka ukuthi ubani osebenzisa ikhodi 56244
wall-vault proxy --port 8080   # Qalisa ngekhodi elinye
```

### Uma Kukhona Amaphutha Ezikhiye ze-API (429, 402, 401, 403, 582)

| Ikhodi Yephutha | Incazelo | Ukuphendula |
|-----------------|----------|-------------|
| **429** | Izicelo eziningi kakhulu (ukusetshenziswa kudlulile) | Linda okwesikhashana noma wengeze ezinye izikhiye |
| **402** | Ukubhadala kuyadingeka noma amakredikhiyeli awanele | Gcwalisa amakredikhiyeli ensizakalo efanele |
| **401 / 403** | Ukhiye ongalungile noma akukho imvume | Hlola kabusha inani lekhiye bese ubhalisa kabusha |
| **582** | Ukudinwa kwe-gateway (cooldown yemizuzu engu-5) | Ikhishwa ngokuzenzakalela ngemizuzu engu-5 |

```bash
# Bona uhlu lwazikhiye ezibhalisiwe nesimo sazo
curl -H "Authorization: Bearer your-admin-token" http://localhost:56243/admin/keys

# Setha kabusha izibali zokusetshenziswa kwezikhiye
curl -X POST -H "Authorization: Bearer your-admin-token" http://localhost:56243/admin/keys/reset
```

### Uma I-agent Ibonakala Njenge-"Ayixhunyiwe"

"Ayixhunyiwe" kusho ukuthi inqubo ye-proxy ayithumeli izimpawu (i-heartbeat) ne-vault. **Akusho ukuthi izilungiselelo azigcinwanga.** I-proxy idinga ukwazi ikheli le-vault server nezikhiye ukuze ibonakale ixhunyiwe.

```bash
# Qalisa i-proxy ngokuchaza ikheli le-vault server, izikhiye, ne-client ID
WV_VAULT_URL=http://vault-server-address:56243 \
WV_VAULT_TOKEN=client-token \
WV_VAULT_CLIENT_ID=client-id \
wall-vault proxy
```

Uma ukukhonza kuphumelela, i-dashboard ishintsha kube 🟢 Iyasebenza ngaphakathi kwamasekhendi angu-20.

### Uma Ollama Ingaxhumi

I-Ollama inhlelo esebenzisa i-AI ngokuqondile ekhompyutheni yakho. Qiniseka ukuthi i-Ollama isebenza kuqala.

```bash
curl http://localhost:11434/api/tags   # Uma uhlu lwamamodeli luvela, kukhona ukusebenza
export OLLAMA_URL=http://192.168.x.x:11434   # Uma isebenza kwelinye ikhompyutha
```

> ⚠️ Uma i-Ollama ingaphenduli, qalisa i-Ollama kuqala ngesicelo `ollama serve`.

> ⚠️ **Amamodeli amakhulu aphendula kancane**: Amamodeli amakhulu anjenge-`qwen3.5:35b` noma `deepseek-r1` angathatha imizuzu eminingana ukuze aphendule. Uma kubonakala sengathi akukho mphendulo, ingabe lokho kushintshwa okujwayelekile — linda nje.

---

## Izinguquko Zakamuva (v0.1.16 ~ v0.1.22)

### v0.1.22 (2026-04-05)
- **Ukulungisa: inkambu ye-content engenalutho yalahlwa**: Lapho amamodeli okucabanga (gemini-3.1-pro, o1, claude thinking, njll.) ephelelwa yi-max_tokens ekucabangeni ngaphambi kokukhiqiza okuphumayo, i-proxy ibilahla inkambu ye-`content`/`text` engenalutho nge-`omitempty`. Amakhasimende e-OpenAI/Anthropic SDK (Claude Code, Cline, njll.) aphahlazeka nge-"Cannot read properties of undefined (reading 'trim')". Manje njalo ikhipha inkambu ngokwemisebenzi esemthethweni ye-API.

### v0.1.21 (2026-04-05)
- **Ukwesekwa kwamamodeli e-Gemma 4**: Amamodeli e-Gemma (gemma-4-31b-it, gemma-4-26b-a4b-it) manje athunyelwa nge-Google Gemini API.
- **Ukwesekwa kwe-LM Studio / vLLM**: Lezi zinsizakalo zendawo manje zithunyelwa ngendlela efanele esikhundleni sokubuyela ku-Ollama.
- **Ukulungisa iDashboard**: Ihlala ibonisa insizakalo esethiwe, hhayi insizakalo yokubuyisela.
- **Ukhetho lwensizakalo yendawo lugciniwe**: IDashboard ayisavali izinsizakalo zendawo ngokuzenzakalela uma ikhasi lifakwa.
- **Okuguquguqukayo kwesihlungi sethuluzi**: Ukweseka `WV_TOOL_FILTER=passthrough`.

### v0.1.20 (2026-03-28)
- **Ukuqiniswa kokuphepha okwanele**: Ukuvimbela i-XSS (amaphuzu angu-41), ukuqhathanisa izikhiye ngesikhathi esilingana, ukuvinjelwa kwe-CORS, imikhawulo yobukhulu besicelo nokunye.

### v0.1.19 (2026-03-27)
- **Ukutholwa kwe-Claude Code ku-inthanethi**: I-Claude Code iboniswa ku-inthanethi ku-dashboard noma iqhela iproksi.

### v0.1.18 (2026-03-26)
- **Ukulungisa ukubuyiselwa**: Ibuyela ngokuzenzakalela ensizakalweni ekhethiwe uma itholakalayo.
- **Ukuthola ukungaxhunywanga okuthuthukisiwe**: Ukuhlola isimo njalo ngemizuzwana eyi-15.

### v0.1.17 (2026-03-25)
- **Ukuhlelwa kabusha kwamakhadi ngokudonsela nokudedela**.
- **Izinkinobho zokusebenzisa ngaphakathi komugqa zama-ajenti angaxhunyiwe**.
- **Uhlobo lwe-ajenti ye-cokacdir lwengeziwe**.

### v0.1.16 (2026-03-25)
- **Ukuvumelaniswa kwamamodeli ngezinhlangothi zombili** kwe-Cline ne-Claude Code.

---

*Ukuze uthole ulwazi olwengeziwe we-API, bheka ku-[API.md](API.md).*
