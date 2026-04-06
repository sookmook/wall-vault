# Umhlahlandlela Womsebenzisi we-wall-vault
*(Kucishwa kamuva: 2026-04-06 — v0.1.23)*

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

**i-wall-vault = i-proxy ye-AI (isithunywa) + indawo yokugcina izikhiye ze-API, eyenzelwe OpenClaw**

Ukusebenzisa izinsizakalo ze-AI kudinga **ukhiye we-API**. Ukhiye we-API ufana ne **nephasi ledijithali** elifakazela ukuthi unelungelo lokusebenzisa leso nsizakalo. Lezi zikhiye zinomkhawulo wemibuzo ngosuku, futhi zingaba sengozini uma zingagcinwa kahle.

I-wall-vault igcina lezi zikhiye endaweni evikelekile (i-vault), bese isebenza njenge-**proxy** (isithunywa esimele) phakathi kuka-OpenClaw nezinsizakalo ze-AI. Nje nje: OpenClaw ixhumana no-wall-vault kuphela -- wonke umsebenzi onzima wenziwa yi-wall-vault.

Izinkinga i-wall-vault ezixazululayo:

- **Ukuphendulana kwazikhiye ngokuzenzakalela**: Uma ukhiye owodwa ufika emkhawulweni wawo noma ubanjwe okwesikhashana (i-cooldown), i-wall-vault iyaguqukelela ngokuthula kukhiye olulandelayo. OpenClaw iqhubeka ngaphandle kokuma.
- **Ukushintsha Kwezinsizakalo Ngokuzenzakalela (i-fallback)**: Uma Google ingaphenduli, iyaguqukelela ku-OpenRouter; uma nako kungasebenzi, iyaguqukela ku-Ollama, LM Studio, noma vLLM (i-AI ekhompyutheni yakho). Iseshini ayipheli. Uma insizakalo yokuqala ibuyela esimweni esijwayelekile, iyashintsha ibuyele kuyo ngokuzenzakalela kusukela esicwelweni esilandelayo (v0.1.18+, LM Studio/vLLM: v0.1.21+).
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
        ├─ Ollama / LM Studio / vLLM (Ikhompyutha yakho, indawo yokugcina yokugcina)
        └─ OpenAI / Anthropic API
```

---

## Ukufakela

### Linux / macOS

Vula itheminali ubese unamathisela le miyalo njengoba injalo.

```bash
# Linux (i-PC evamile, iseva -- amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` -- Ilanda ifayela ku-inthanethi.
- `chmod +x` -- Yenza ifayela elilandiwe "likwazi ukusebenza". Uma udlula lesi sinyathelo uzothola iphutha elithi "awukho umvume".

### Windows

Vula i-PowerShell (njengomphathi) bese usebenzisa le miyalo.

```powershell
# Ukulanda
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Engeza ku-PATH (isebenza ngemva kokuvula kabusha i-PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **Yini i-PATH?** Yuhlu lwamafolda ikhompyutha efuna imiyalo kuwo. Uma ungeza ku-PATH, ungabhala `wall-vault` kunoma yiliphi ifolda bese uyisebenzisa.

### Ukwakha Kusuka Kumthombo (kwabathuthukisi)

Lokhu kusebenza kuphela uma indawo yokuthuthukisa yolimi lwe-Go ifakelwe.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (inguqulo: v0.1.23.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Inguqulo yesitembu sesikhathi**: Uma wakha ngo-`make build`, inguqulo yakhiwa ngokuzenzakalela ngesimo esifana no-`v0.1.23.20260406.211004` oluqukethe usuku nesikhathi. Uma wakha ngokuqondile ngo-`go build ./...`, inguqulo izobonisa `"dev"` kuphela.

---

## Ukuqala Okokuqala

### Ukusebenzisa i-setup wizard

Ngemva kokufakela, qiniseka usebenzisa **i-wizard yokusetha** ngomyalo ongezansi. I-wizard izokuqondisa ngezinto ezidingekayo ngokubuza ngayinye ngayinye.

```bash
wall-vault setup
```

Izinyathelo i-wizard ezidlula kuzo yilezi:

```
1. Ukukhetha ulimi (izilimi ezingu-10 okubandakanya isiZulu)
2. Ukukhetha ithimu (light / dark / gold / cherry / ocean)
3. Imodi yokusebenza -- wedwa (standalone), noma nezixhumanisi eziningi (distributed)
4. Igama le-bot -- igama elizovela kudashubhodi
5. Ukusetha amaphothi -- okuzenzakalelayo: proxy 56244, vault 56243 (cindezela Enter uma ungadingi ukushintsha)
6. Ukukhetha izinsizakalo ze-AI -- Google / OpenRouter / Ollama / LM Studio / vLLM
7. Ukusetha isihluzi sethuluzi lokuphepha
8. Ukusetha ithokheni yomphathi -- iphasiwedi yokuvala izici zokuphatha ze-dashboard. Ingakhiqizwa ngokuzenzakalela
9. Ukusetha iphasiwedi yokubethela ukhiye we-API -- ukugcina okungeziwe kokuphepha (ukukhetha)
10. Indawo yokugcina ifayela lokusetha
```

> ⚠️ **Khumbula ithokheni yomphathi.** Uzoyidinga kamuva ukwengeza izikhiye kudashubhodi noma ukushintsha izilungiselelo. Uma uyilahla uzodinga ukuhlela ngokuqondile ifayela lokusetha.

Uma i-wizard iqedile, ifayela lokusetha `wall-vault.yaml` lizokwakhiwa ngokuzenzakalela.

### Ukusebenzisa

```bash
wall-vault start
```

Amaseva amabili azokuqala ngasikhathi sinye:

- **I-Proxy** (`http://localhost:56244`) -- isithunywa esixhuma OpenClaw nezinsizakalo ze-AI
- **I-Vault** (`http://localhost:56243`) -- ukuphatha ukhiye we-API nedashubhodi yewebhu

Vula isiphequluli bese uya ku-`http://localhost:56243` ukuze ubone idashubhodi ngokuqondile.

---

## Ukubhalisa Ukhiye we-API

Kunezindlela ezine zokubhalisa ukhiye we-API. **Kwabaqalayo, indlela yoku-1 (izinguquko zemvelo) iyanconywa**.

### Indlela yoku-1: Izinguquko Zemvelo (Iyanconywa -- elula kakhulu)

Izinguquko zemvelo **yizinani ezisethwe kusengaphambili** ezifundwa uhlelo lwesoftware uma luqala. Bhala kutheminali kanje:

```bash
# Bhalisa ukhiye we-Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Bhalisa ukhiye we-OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Sebenzisa ngemva kokubhalisa
wall-vault start
```

Uma unezikhiye eziningi, zihlanganise ngokhomu (,). i-wall-vault izosebenzisa izikhiye ngokulandelana ngokuzenzakalela (round robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Ithiphu**: Umyalo we-`export` usebenza kuphela kuseshini yamanje yetheminali. Ukuze uhlale ngemva kokuqala kabusha ikhompyutha, engeza lo myalo kufayela elithi `~/.bashrc` noma `~/.zshrc`.

### Indlela yoku-2: Idashubhodi ye-UI (cindezela ngemawusi)

1. Vula isiphequluli bese uya ku-`http://localhost:56243`
2. Kukhadi ye-**🔑 Izikhiye ze-API** phezulu, cindezela inkinobho ethi `[+ Engeza]`
3. Faka uhlobo lwensizakalo, inani lokhiye, ilebuli (igama lesikhumbuzo), nomkhawulo wansuku zonke, bese ugcina

### Indlela yoku-3: I-REST API (yokuzenzakalela neskripthi)

I-REST API yindlela yezinhlelo zesoftware ukushintshana ngedatha nge-HTTP. Iwusizo ekubhalisweni okuzenzakalela ngeskripthi.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer ithokheni-yomphathi" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Ukhiye Oyinhloko",
    "daily_limit": 1000
  }'
```

### Indlela yoku-4: Amafulegi e-proxy (okuhlolelwa okwesikhashana)

Sebenzisa lokhu ukufaka ukhiye okwesikhashana ngaphandle kokubhalisa ngokusemthethweni. Ukhiye uzonyamalala uma uhlelo lwesoftware luvaliwe.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Ukusebenzisa i-Proxy

### Ukusebenzisa no-OpenClaw (inhloso enkulu)

Nansi indlela osetha ngayo ukuthi OpenClaw ixhume nezinsizakalo ze-AI nge-wall-vault.

Vula ifayela elithi `~/.openclaw/openclaw.json` bese wengeza okuqukethwe okulandelayo:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // ithokheni ye-agent ye-vault
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // i-context yamahhala ye-1M
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Indlela elula kakhulu**: Cindezela inkinobho ethi **🦞 Kopisha Okusetiwe kwe-OpenClaw** kukhadi ye-agent kudashubhodi. Isiqeshana esinethokeni nekheli eselicishelwe kakade sizokopishwa kubhodi lokunamathisela. Namathisela nje.

**`wall-vault/` ngaphambi kwegama lemodeli iqondiswa kuphi?**

Ngegama lemodeli, i-wall-vault iyazi ngokuzenzakalela ukuthi yiluphi uhlobo lwensizakalo ye-AI okufanele ithunyelwe kuyo isicelo:

| Isimo Semodeli | Insizakalo Exhunywe |
|----------|--------------|
| `wall-vault/gemini-*` | Google Gemini ngokuqondile |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI ngokuqondile |
| `wall-vault/claude-*` | Anthropic nge-OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (i-context yamahhala yamathokheni ayisigidi esi-1) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/igama-lemodeli`, `openai/igama-lemodeli`, `anthropic/igama-lemodeli` njll. | Insizakalo ehambisanayo ngokuqondile |
| `custom/google/igama-lemodeli`, `custom/openai/igama-lemodeli` njll. | Susa ingxenye ye-`custom/` bese uqondisa kabusha |
| `igama-lemodeli:cloud` | Susa ingxenye ye-`:cloud` bese uxhuma nge-OpenRouter |

> 💡 **Yini i-context?** Yinani lengxoxo i-AI ekwazi ukuyikhumbula ngasikhathi sinye. 1M (amathokheni ayisigidi esi-1) kusho ukuthi izingxoxo ezinde kakhulu noma imibhalo emide ingasetshenziswa ngasikhathi sinye.

### Ukuxhuma Ngokuqondile Ngesimo se-Gemini API (ukuhambisana namathuluzi akhona)

Uma unamathuluzi abesebenzisa Google Gemini API ngokuqondile, shintsha nje ikheli libe ngelika-wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Noma uma ithuluzi lakho lisebenzisa i-URL ngokuqondile:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Ukusebenzisa ne-OpenAI SDK (Python)

Ungaxhuma futhi i-wall-vault kukhodi ye-Python esebenzisa i-AI. Shintsha i-`base_url` kuphela:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # i-wall-vault iphatha ukhiye we-API
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # faka ngesimo se-provider/model
    messages=[{"role": "user", "content": "Sawubona"}]
)
```

### Ukushintsha Imodeli Ngesikhathi Yokusebenza

Ukushintsha imodeli ye-AI ngesikhathi i-wall-vault isisebenza:

```bash
# Shintsha imodeli ngokucela ngokuqondile ku-proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# Kuimodi eyahlukaniswayo (ama-bot amaningi), shintsha kuseva ye-vault → kuzoboniswa ngokuphazima nge-SSE
curl -X PUT http://localhost:56243/admin/clients/i-bot-id-yami \
  -H "Authorization: Bearer ithokheni-yomphathi" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Ukuhlola Uhlu Lwamamodeli Atholakalayo

```bash
# Buka uhlu lonke
curl http://localhost:56244/api/models | python3 -m json.tool

# Amamodeli e-Google kuphela
curl "http://localhost:56244/api/models?service=google"

# Sesha ngegama (isibonelo: amamodeli anokuthi "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Isifinyezo samamodeli ayinhloko ngensizakalo:**

| Insizakalo | Amamodeli Ayinhloko |
|--------|----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | Angaphezu kuka-346 (Hunter Alpha 1M i-context yamahhala, DeepSeek R1/V3, Qwen 2.5 njll.) |
| Ollama | Ithola ngokuzenzakalela amamodeli eseva yasekhaya ekhompyutheni yakho |
| LM Studio | Iseva yasekhaya ekhompyutheni yakho (iphothi 1234) |
| vLLM | Iseva yasekhaya ekhompyutheni yakho (iphothi 8000) |

---

## Idashubhodi ye-Vault

Vula isiphequluli bese uya ku-`http://localhost:56243` ukuze ubone idashubhodi.

**Ukwakheka kwesikrini:**
- **Ibha ephezulu enamathisiwe (topbar)**: Ilogo, isikhethi solimi nethimu, isimo sokuxhuma se-SSE
- **Igridi yamakhadi**: Amakhadi e-agent, ensizakalo, nokhiye we-API ahlelwe njengamathayili

### Ikhadi Lokhiye we-API

Ikhadi elikuvumela ukuphatha zonke izikhiye ze-API ezibhalisiwe ngombono owodwa.

- Libonisa uhlu lwezikhiye oluhlukaniswe ngensizakalo.
- `today_usage`: Amathokheni (inani lamagama i-AI eyafundile neyawabhalayo) acutshungulwe ngempumelelo namhlanje
- `today_attempts`: Isamba sezicelo zanamhlanje (impumelelo + ukuhluleka)
- Inkinobho ethi `[+ Engeza]` yokubhalisa ukhiye omusha, no-`✕` wokususa ukhiye.

> 💡 **Yini ithokheni?** Isilinganiso esisetshenziselwa i-AI ukucubungula umbhalo. Cishe yigama elilodwa lesiNgisi, noma izinhlamvu ezingu-1-2 zezinye izilimi. Izindleko ze-API ngokuvamile zibalwa ngokwenani lamathokheni.

### Ikhadi Le-Agent

Ikhadi elibonisa isimo sama-bot (ama-agent) axhume ne-proxy ye-wall-vault.

**Isimo sokuxhuma siboniswa ngamazinga ama-4:**

| Isibonakaliso | Isimo | Incazelo |
|------|------|------|
| 🟢 | Iyasebenza | I-proxy isebenza kahle |
| 🟡 | Ukubambezeleka | Iyaphendula kodwa kancane |
| 🔴 | Ayikho ku-inthanethi | I-proxy ayiphenduli |
| ⚫ | Ayixhunyiwe/Ivaliwe | I-proxy ayikaze ixhume ne-vault noma ivaliwe |

**Ukuchazwa kwezinkinobho ngaphansi kwekhadi le-agent:**

Uma ubhalisa i-agent bese ucacisa **uhlobo lwe-agent**, izinkinobho zokwenza lula ezifanele lolo hlobo zizovela ngokuzenzakalela.

---

#### 🔘 Inkinobho Yokukopisha Okusetiwe -- Yakha okusetiwe kokuxhuma ngokuzenzakalela

Uma ucindezela inkinobho, isiqeshana sokusetha esinethokeni, ikheli le-proxy, nolwazi lwemodeli ye-agent sikopishwa kubhodi lokunamathisela. Namathisela nje endaweni eboniswe ethebuleni elingezansi bese okusetiwe kokuxhuma kuqedwa.

| Inkinobho | Uhlobo Lwe-Agent | Indawo Yokunamathisela |
|------|-------------|-------------|
| 🦞 Kopisha Okusetiwe kwe-OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Kopisha Okusetiwe kwe-NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Kopisha Okusetiwe kwe-Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Kopisha Okusetiwe kwe-Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Kopisha Okusetiwe kwe-VSCode | `vscode` | `~/.continue/config.json` |

**Isibonelo -- Uma kuyuhlobo lwe-Claude Code, yilokhu okuzokopishwa:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.0.6:56244/v1",
  "apiKey": "ithokheni-yale-agent"
}
```

**Isibonelo -- Uma kuyuhlobo lwe-VSCode (Continue):**

```yaml
# ~/.continue/config.yaml  ← namathisela ku-config.yaml, hhayi config.json
name: My Config
version: 0.0.1
schema: v1

models:
  - name: wall-vault proxy
    provider: openai
    model: gemini-2.5-flash
    apiBase: http://192.168.0.6:56244/v1
    apiKey: ithokheni-yale-agent
    roles:
      - chat
      - edit
      - apply
```

> ⚠️ **Inguqulo entsha ye-Continue isebenzisa `config.yaml`.** Uma `config.yaml` ikhona, `config.json` izoshaywa indiva ngokuphelele. Qiniseka unamathisela ku-`config.yaml`.

**Isibonelo -- Uma kuyuhlobo lwe-Cursor:**

```
Base URL : http://192.168.0.6:56244/v1
API Key  : ithokheni-yale-agent

// Noma izinguquko zemvelo:
OPENAI_BASE_URL=http://192.168.0.6:56244/v1
OPENAI_API_KEY=ithokheni-yale-agent
```

> ⚠️ **Uma ukukopisha kubhodi lokunamathisela kungasebenzi**: Izinqubomgomo zokuphepha zesiphequluli zingavimba ukukopisha. Uma ibhokisi lokubhala livuleka njengediyalogi, cindezela Ctrl+A ukukhetha konke bese Ctrl+C ukukopisha.

---

#### ⚡ Inkinobho Yokusebenzisa Ngokuzenzakalela -- Cindezela kanye bese okusetiwe kuqediwe

Uma uhlobo lwe-agent lungu-`cline`, `claude-code`, `openclaw`, noma `nanoclaw`, inkinobho ethi **⚡ Sebenzisa Okusetiwe** izoboniswa kukhadi le-agent. Uma ucindezela le nkinobho, ifayela lokusethwa lasekhaya le-agent lizobuyekezwa ngokuzenzakalela.

| Inkinobho | Uhlobo Lwe-Agent | Ifayela Elisetshenziswa |
|------|-------------|-------------|
| ⚡ Sebenzisa Okusetiwe kwe-Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Sebenzisa Okusetiwe kwe-Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Sebenzisa Okusetiwe kwe-OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Sebenzisa Okusetiwe kwe-NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Le nkinobho ithumela isicelo ku-**localhost:56244** (i-proxy yasekhaya). I-proxy kufanele isebenze kule mashini ukuze isebenze.

---

#### 🔀 Ukuhlelwa Kwamakhadi Ngokudonsela Nokudedela (v0.1.17)

Ungakwazi **ukudonsela** amakhadi e-agent kudashubhodi bese uwahlelela ngendlela oyithandayo.

1. Bamba ikhadi le-agent ngemawusi bese ulidonsela
2. Lidedele phezu kwekhadi olithandayo bese uhlelo lushintsha
3. Uhlelo olusha **lugcinwa ngokuphazima kuseva** futhi luhlala ngemva kokulayisha kabusha

> 💡 Izixhumanisi zokuthinta (amaselula/amathebhulethi) azikasekelwa okwamanje. Sebenzisa isiphequluli sedeskithophu.

---

#### 🔄 Ukuvumelaniswa Kwamamodeli Nhlangothi Zombili (v0.1.16)

Uma ushintsha imodeli ye-agent kudashubhodi ye-vault, okusetiwe kwasekhaya kwe-agent kubuyekezwa ngokuzenzakalela.

**Nge-Cline:**
- Ukushintsha imodeli ku-vault → umcimbi we-SSE → i-proxy ibuyekeza ingxenye yemodeli ku-`globalState.json`
- Izingxenye ezibuyekezwayo: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` nokhiye we-API akuthintwa
- **Ukulayisha kabusha i-VS Code (`Ctrl+Alt+R` noma `Ctrl+Shift+P` → `Developer: Reload Window`) kuyadingeka**
  - I-Cline ayiphindi ifunde ifayela lokusetha ngesikhathi isebenza

**Nge-Claude Code:**
- Ukushintsha imodeli ku-vault → umcimbi we-SSE → i-proxy ibuyekeza ingxenye ye-`model` ku-`settings.json`
- Ifuna ngokuzenzakalela izindlela ze-WSL ne-Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Umkhombandlela ophambene (i-agent → i-vault):**
- Uma i-agent (Cline, Claude Code njll.) ithumela isicelo ku-proxy, i-proxy ifaka ulwazi lwensizakalo/lwemodeli yeklayenti ku-heartbeat
- Insizakalo/imodeli esetshenziswa okwamanje iboniswa ngesikhathi sangempela kukhadi le-agent kudashubhodi ye-vault

> 💡 **Okungumsuka**: I-proxy ihlonza i-agent ngethokheni ye-Authorization yesicelo bese iqondisa ngokuzenzakalela ensizakalweni/emodeli esethwe ku-vault. Noma i-Cline noma Claude Code ithumela igama lemodeli elihlukile, i-proxy iyashintshela kokusetiwe kwe-vault.

---

### Ukusebenzisa i-Cline ku-VS Code -- Umhlahlandlela Obanzi

#### Isinyathelo soku-1: Fakela i-Cline

Fakela **i-Cline** (ID: `saoudrizwan.claude-dev`) kusuka emakethe yezengezelo ye-VS Code.

#### Isinyathelo soku-2: Bhalisa i-agent ku-vault

1. Vula idashubhodi ye-vault (`http://i-IP-ye-vault:56243`)
2. Cindezela **+ Engeza** engxenyeni ye-**Agent**
3. Faka ngale ndlela:

| Inkambu | Inani | Incazelo |
|------|----|------|
| ID | `i_cline_yami` | Isihlonzinkulumo esiyingqayizivele (izinhlamvu zesiNgisi, ngaphandle kwesikhala) |
| Igama | `I-Cline Yami` | Igama elizoboniswa kudashubhodi |
| Uhlobo Lwe-Agent | `cline` | ← kumele ukhethe `cline` |
| Insizakalo | Khetha insizakalo ozoyisebenzisa (isibonelo: `google`) | |
| Imodeli | Faka imodeli ozoyisebenzisa (isibonelo: `gemini-2.5-flash`) | |

4. Cindezela **Gcina** bese ithokheni ikhiqizwa ngokuzenzakalela

#### Isinyathelo soku-3: Xhuma i-Cline

**Indlela A -- Ukusebenzisa ngokuzenzakalela (iyanconywa)**

1. Qiniseka ukuthi **i-proxy** ye-wall-vault iyasebenza kule mashini (`localhost:56244`)
2. Cindezela inkinobho ethi **⚡ Sebenzisa Okusetiwe kwe-Cline** kukhadi le-agent kudashubhodi
3. Uma isaziso esithi "Okusetiwe kusetshenzisiwe!" sibonakala, kuphumelele
4. Layisha kabusha i-VS Code (`Ctrl+Alt+R`)

**Indlela B -- Ukusetha ngesandla**

Vula izilungiselelo (⚙️) kubha yecala ye-Cline:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://ikheli-le-proxy:56244/v1`
  - Imashini efanayo: `http://localhost:56244/v1`
  - Enye imashini njenge-Mini server: `http://192.168.0.6:56244/v1`
- **API Key**: Ithokheni ekhishwe ku-vault (kopisha kukhadi le-agent)
- **Model ID**: Imodeli esethwe ku-vault (isibonelo: `gemini-2.5-flash`)

#### Isinyathelo soku-4: Qinisekisa

Thumela noma yimuphi umlayezo kubhokisi lengxoxo le-Cline. Uma kulungile:
- Ichashazi eliluhlaza (● Iyasebenza) lizobonakala kukhadi le-agent kudashubhodi ye-vault
- Insizakalo/imodeli yamanje izoboniswa kukhadi (isibonelo: `google / gemini-2.5-flash`)

#### Ukushintsha Imodeli

Uma ufuna ukushintsha imodeli ye-Cline, shintsha ku-**dashubhodi ye-vault**:

1. Shintsha imenyu yokudonsela yensizakalo/yemodeli kukhadi le-agent
2. Cindezela **Sebenzisa**
3. Layisha kabusha i-VS Code (`Ctrl+Alt+R`) -- igama lemodeli ngaphansi kwe-Cline lizobuyekezwa
4. Kusukela esicwelweni esilandelayo, imodeli entsha izosetshenziwa

> 💡 Eqinisweni, i-proxy ihlonza isicelo se-Cline ngethokheni bese iqondisa emodeli yokusetha ye-vault. Noma ungalayishi kabusha i-VS Code **imodeli esetshenziswayo ishintsha ngokuphazima** -- ukulayisha kabusha kungokokubuyekeza ukuboniswa kwemodeli ku-UI ye-Cline kuphela.

#### Ukuthola Ukuphuka Kokuxhuma

Uma uvala i-VS Code, ikhadi le-agent kudashubhodi ye-vault lizophenduka phuzi (ukubambezeleka) ngemva kwamasekendi angama-**90**, futhi libomvu (ayikho ku-inthanethi) ngemva kwemizuzu engu-**3**. (Kusukela ku-v0.1.18, ukuhlolwa kwesimo kwamasekhendi angu-15 kwenze ukuthola ukuthi ayikho ku-inthanethi kube ngokushesha.)

#### Ukuxazulula Izinkinga

| Isibonakaliso | Imbangela | Isixazululo |
|------|------|------|
| Iphutha lokuthi "ukuxhuma kuhlulekile" ku-Cline | I-proxy ayisebenzi noma ikheli alilungile | Qinisekisa i-proxy ngo-`curl http://localhost:56244/health` |
| Ichashazi eliluhlaza alibonakali ku-vault | Ukhiye we-API (ithokheni) akasethiwe | Cindezela kabusha inkinobho ethi **⚡ Sebenzisa Okusetiwe kwe-Cline** |
| Imodeli ngaphansi kwe-Cline ayishintshi | I-Cline igcine okusetiwe ku-cache | Layisha kabusha i-VS Code (`Ctrl+Alt+R`) |
| Igama lemodeli elingalungile libonakala | Iphutha elidala (lilungisiwe ku-v0.1.16) | Buyekeza i-proxy ibe ngu-v0.1.16 nangaphezulu |

---

#### 🟣 Inkinobho Yokukopisha Umyalo Wokusabalalisa -- Isetshenziswa uma ufakela kumashini entsha

Isetshenziswa uma ufakela i-proxy ye-wall-vault okokuqala kumashini entsha futhi uyixhuma ne-vault. Cindezela inkinobho bese iskripthi sonke sokufakela sikopishwa. Namathisela kutheminali yekhompyutha entsha bese usisebenzisa -- okulandelayo kuzosetshenzwa ngasikhathi sinye:

1. Ukufakelwa kwefayela ye-wall-vault (kuzodlulwa uma sekufakelwe)
2. Ukubhaliswa ngokuzenzakalela kwensizakalo yomsebenzisi ye-systemd
3. Ukuqalisa insizakalo nokuxhuma ngokuzenzakalela ne-vault

> 💡 Ithokheni yale agent nekheli leseva ye-vault sekugcwalisiwe ngaphakathi kweskripthi, ngakho ungasisebenzisa ngokuqondile ngemva kokunamathisela ngaphandle kwezinguquko.

---

### Ikhadi Lensizakalo

Ikhadi lokuvula nokuvala noma ukusetha izinsizakalo ze-AI ozozisebenzisa.

- Izishintshi zokuvula nokuvala zensizakalo ngayinye
- Uma ufaka ikheli leseva ye-AI yasekhaya (Ollama, LM Studio, vLLM njll. esebenza ekhompyutheni yakho), izothola ngokuzenzakalela amamodeli atholakalayo.
- **Ukuboniswa kwesimo sokuxhuma sensizakalo yasekhaya**: Ichashazi ● eliseduze negama lensizakalo uma **liluhlaza** ixhunyiwe, uma **limpunga** ayixhunyiwe
- **Izibani zendlela ezenzakalelayo zensizakalo yasekhaya** (v0.1.23+): Izinsizakalo zasekhaya (Ollama, LM Studio, vLLM) zivulwa zivalwe ngokuzenzakalela ngokwemilinganiso yokuthi zingaxhunwa yini. Uma uvula insizakalo, ngaphakathi kwamasekhendi angu-15 ichashazi ● lizophenduka luhlaza nebhokisi lokumaka lizovulwa; uma uvala insizakalo, izovalwa ngokuzenzakalela. Lokhu kusebenza ngendlela efanayo nezinsizakalo zefu (Google, OpenRouter njll.) ezivulwa zivalwe ngokuzenzakalela ngokuya ngokuthi ukhiye we-API ukhona yini.

> 💡 **Uma insizakalo yasekhaya isebenza kwenye ikhompyutha**: Faka i-IP yaleyo khompyutha kubhokisi le-URL yensizakalo. Isibonelo: `http://192.168.0.6:11434` (Ollama), `http://192.168.0.6:1234` (LM Studio). Uma insizakalo iboshwe ku-`127.0.0.1` hhayi ku-`0.0.0.0`, ngeke ikwazi ukufinyelelwa nge-IP yangaphandle, ngakho hlola ikheli lokuboshwa kumasethwa ensizakalo.

### Ukufaka Ithokheni Yomphathi

Uma uzama ukusebenzisa izici ezibalulekile njengokwengeza noma ukususa izikhiye kudashubhodi, idiyalogi yokufaka ithokheni yomphathi izobonakala. Faka ithokheni oyisethile ngesikhathi se-setup wizard. Uma usuyifakile, izohlala ize uvale isiphequluli.

> ⚠️ **Uma ukuqinisekiswa kuhluleka ngaphezu kwamazikhathi angu-10 phakathi kwemizuzu engu-15, leyo IP izovalelwa okwesikhashana.** Uma ukhohlwe ithokheni, hlola into ethi `admin_token` kufayela elithi `wall-vault.yaml`.

---

## Imodi Eyahlukaniswayo (Ama-Bot Amaningi)

Uma usebenzisa OpenClaw kumakhompyutha amaningi ngasikhathi sinye, lokhu kuyisakhiwo **sokwabelana nge-vault eyodwa yokhiye**. Kulula ngoba udinga ukuphatha izikhiye endaweni eyodwa kuphela.

### Isibonelo Sesakhiwo

```
[Iseva ye-Vault Yokhiye]
  wall-vault vault    (i-vault yokhiye :56243, idashubhodi)

[WSL Alpha]            [Raspberry Pi Gamma]    [Mac Mini Local]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ ukuvumelanisa SSE   ↕ ukuvumelanisa SSE     ↕ ukuvumelanisa SSE
```

Onke ama-bot abheke iseva ye-vault esimaphakathi, ngakho ukushintsha imodeli noma ukwengeza ukhiye ku-vault kubonakala ngokuphazima kuwo wonke ama-bot.

### Isinyathelo soku-1: Qala Iseva ye-Vault Yokhiye

Sebenzisa kukhompyutha ezosetshenziselwa njengiseva ye-vault:

```bash
wall-vault vault
```

### Isinyathelo soku-2: Bhalisa I-Bot Ngayinye (Iklayenti)

Bhalisa ngaphambili ulwazi lwe-bot ngayinye ezoxhuma kuseva ye-vault:

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer ithokheni-yomphathi" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "I-BotA",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### Isinyathelo soku-3: Qala I-Proxy Kukhompyutha Ye-Bot Ngayinye

Sebenzisa i-proxy kukhompyutha ngayinye ene-bot ngokucacisa ikheli leseva ye-vault nethokheni:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Buyisela **`192.168.x.x`** ngekheli langempela le-IP yangaphakathi yekhompyutha yeseva ye-vault. Ungalihlola ngokusetha kweroutha noma ngomyalo othi `ip addr`.

---

## Ukusethwa Kokuqala Ngokuzenzakalela

Uma kuyinkathazo ukuvula i-wall-vault ngesandla njalo uma ikhompyutha iqala kabusha, yibhalise njengensizakalo yesistimu. Uma usuyibhalisile, izoqala ngokuzenzakalela ngesikhathi se-boot.

### Linux -- systemd (i-Linux eningi)

I-systemd yisistimu ye-Linux yokuqala nokuphatha izinhlelo zesoftware ngokuzenzakalela:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Ukuhlola amalogi:

```bash
journalctl --user -u wall-vault -f
```

### macOS -- launchd

Isistimu ephatha ukusebenziswa kwezinhlelo zesoftware ngokuzenzakalela ku-macOS:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows -- NSSM

1. Landa i-NSSM ku-[nssm.cc](https://nssm.cc/download) bese uyengeza ku-PATH.
2. Ku-PowerShell yomphathi:

```powershell
wall-vault doctor deploy windows
```

---

## I-Doctor: Isihloli

Umyalo we-`doctor` yithuluzi le-wall-vault **lokuzicwaninga nokulungisa ngokwalo**.

```bash
wall-vault doctor check   # Cwaninga isimo samanje (funda kuphela, akukho okushintshwayo)
wall-vault doctor fix     # Lungisa izinkinga ngokuzenzakalela
wall-vault doctor all     # Ukucwaninga + ukulungisa ngokuzenzakalela ngasikhathi sinye
```

> 💡 Uma kukhona okubonakala kungalungile, sebenzisa `wall-vault doctor all` kuqala. Kuxazulula izinkinga eziningi ngokuzenzakalela.

---

## Izinguquko Zemvelo

Izinguquko zemvelo yindlela yokudlulisa amanani okusetha kuzinhlelo zesoftware. Faka ngokuthi `export igama-lokuguquka=inani` kutheminali, noma kufake kufayela lensizakalo yokuqala ngokuzenzakalela ukuze kusebenze njalo.

| Inguquko | Incazelo | Inani Lesibonelo |
|------|------|---------|
| `WV_LANG` | Ulimi lwedashubhodi | `ko`, `en`, `ja` |
| `WV_THEME` | Ithimu yedashubhodi | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Ukhiye we-API we-Google (eziningi ngokhomu) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Ukhiye we-API we-OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Ikheli leseva ye-vault kuimodi eyahlukaniswayo | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Ithokheni yokuqinisekiswa yeklayenti (i-bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Ithokheni yomphathi | `admin-token-here` |
| `WV_MASTER_PASS` | Iphasiwedi yokubethela ukhiye we-API | `my-password` |
| `WV_AVATAR` | Indlela yefayela lesithombe se-avatar (indlela ehlobene no-`~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Ikheli leseva yasekhaya ye-Ollama | `http://192.168.x.x:11434` |

---

## Ukuxazulula Izinkinga

### I-Proxy Ayiqali

Ngokuvamile iphothi seyisetshenziselwa esinye isoftware.

```bash
ss -tlnp | grep 56244   # Hlola ukuthi ubani osebenzisa iphothi 56244
wall-vault proxy --port 8080   # Qala ngenombolo yephothi ehlukile
```

### Amaphutha Okhiye we-API (429, 402, 401, 403, 582)

| Ikhodi Yephutha | Incazelo | Indlela Yokwenza |
|----------|------|----------|
| **429** | Izicelo eziningi kakhulu (umkhawulo wokusetshenziswa udluliwe) | Linda kancane noma engeza esinye isikhiye |
| **402** | Kudingeka inkokhelo noma ibhalansi iphelile | Gcwalisa ibhalansi ensizakalweni ehambisanayo |
| **401 / 403** | Ukhiye awulungile noma awukho umvume | Qinisekisa kabusha inani lokhiye bese ubhalisa kabusha |
| **582** | Ukuminyana kwegateway (cooldown imizuzu emi-5) | Izovulwa ngokuzenzakalela ngemva kwemizuzu emi-5 |

```bash
# Hlola uhlu lwezikhiye ezibhalisiwe nesimo sazo
curl -H "Authorization: Bearer ithokheni-yomphathi" http://localhost:56243/admin/keys

# Setha kabusha izibalo zokusetshenziswa kokhiye
curl -X POST -H "Authorization: Bearer ithokheni-yomphathi" http://localhost:56243/admin/keys/reset
```

### I-Agent Iboniswa Njengokuthi "Ayixhunyiwe"

"Ayixhunyiwe" kusho ukuthi inqubo ye-proxy ayithumeli isignali (heartbeat) ku-vault. **Akusho ukuthi okusetiwe akugcinwanga.** I-proxy kumele yazi ikheli leseva ye-vault nethokheni futhi isebenze ukuze iguqulele esimweni sokuxhuma.

```bash
# Qala i-proxy ngokucacisa ikheli leseva ye-vault, ithokheni, ne-ID yeklayenti
WV_VAULT_URL=http://ikheli-leseva-ye-vault:56243 \
WV_VAULT_TOKEN=ithokheni-yeklayenti \
WV_VAULT_CLIENT_ID=i-id-yeklayenti \
wall-vault proxy
```

Uma ukuxhuma kuphumelela, izoshintsha ibe ngu-🟢 Iyasebenza kudashubhodi phakathi kwamasekhendi angama-20.

### I-Ollama Ayixhumi

I-Ollama uhlelo lwesoftware lokusebenzisa i-AI ngokuqondile ekhompyutheni yakho. Qala ngokuqinisekisa ukuthi i-Ollama iyasebenza.

```bash
curl http://localhost:11434/api/tags   # Uma uhlu lwamamodeli lubonakala, kulungile
export OLLAMA_URL=http://192.168.x.x:11434   # Uma isebenza kwenye ikhompyutha
```

> ⚠️ Uma i-Ollama ingaphenduli, qala i-Ollama kuqala ngomyalo othi `ollama serve`.

> ⚠️ **Amamodeli amakhulu abambezeleka ekuphenduleni**: Amamodeli amakhulu afana no-`qwen3.5:35b`, `deepseek-r1` angathatha imizuzu eminingana ukukhiqiza impendulo. Noma kubonakala sengathi akukho mpendulo, kungase kucutshungulwa ngokujwayelekile, ngakho linda.

---

## Izinguquko Zamuva (v0.1.16 ~ v0.1.23)

### v0.1.23 (2026-04-06)
- **Ukulungiswa kokushintsha imodeli ye-Ollama**: Kulungisiwe inkinga lapho ukushintsha imodeli ye-Ollama kudashubhodi ye-vault kungabonakali ku-proxy yangempela. Ngaphambilini, kusetshenziswe kuphela inguquko yemvelo (`OLLAMA_MODEL`), kodwa manje okusetiwe kwe-vault kuya phambili.
- **Izibani zendlela ezenzakalelayo zensizakalo yasekhaya**: Ollama, LM Studio, no-vLLM zivulwa ngokuzenzakalela uma zingaxhunwa futhi zivalwa ngokuzenzakalela uma ziphukile. Lokhu kusebenza ngendlela efanayo nokushintsha ngokuzenzakalela kwezinsizakalo zefu okuncike kukhiye.

### v0.1.22 (2026-04-05)
- **Ukulungiswa kwengxenye ye-content engenalutho engekhoyo**: Uma amamodeli okucabanga (gemini-3.1-pro, o1, claude thinking njll.) esebenzisa umkhawulo we-max_tokens wonke ku-reasoning futhi ehluleka ukukhiqiza impendulo yangempela, i-proxy isuse izingxenye ze-`content`/`text` ze-JSON yempendulo nge-`omitempty`, okubangela amaphutha athi `Cannot read properties of undefined (reading 'trim')` kumakhlayenti e-SDK e-OpenAI/Anthropic. Kushintshiwe ukuze izingxenye zihlale zifakwe ngokuya ngezincazelo ezisemthethweni ze-API.

### v0.1.21 (2026-04-05)
- **Ukusekelwa kwamamodeli e-Gemma 4**: Amamodeli omndeni we-Gemma afana no-`gemma-4-31b-it`, `gemma-4-26b-a4b-it` angasetshenziswa nge-Google Gemini API.
- **Ukusekelwa ngokusemthethweni kwezinsizakalo ze-LM Studio / vLLM**: Ngaphambilini lezi zinsizakalo zedlulwe ekuqondiseni kwe-proxy futhi njalo zibuyiselwa ku-Ollama. Manje ziqondiswa kahle nge-API ehambisana ne-OpenAI.
- **Ukulungiswa kokuboniswa kwensizakalo kudashubhodi**: Noma nge-fallback, idashubhodi ihlale ibonisa insizakalo esethwe umsebenzisi.
- **Ukuboniswa kwesimo sensizakalo yasekhaya**: Isimo sokuxhuma sezinsizakalo zasekhaya (Ollama, LM Studio, vLLM njll.) siboniswa ngombala wechashazi ● uma idashubhodi ilayisha.
- **Inguquko yemvelo yesihluzi sethuluzi**: Imodi yokudlulisa amathuluzi (tools) ingasethwa ngomguquko wemvelo othi `WV_TOOL_FILTER=passthrough`.

### v0.1.20 (2026-03-28)
- **Ukuqiniswa okujulile kokuphepha**: Ukuvikela i-XSS (izindawo ezingu-41), ukuqhathaniswa kwethokheni kwesikhathi esingashintshi, izivimbelo ze-CORS, imikhawulo yobukhulu besicelo, ukuvikela ukuhambahamba kwendlela, ukuqinisekiswa kwe-SSE, ukuqiniswa komkhawulo wesivinini njll. izinto ezingu-12 zokuphepha ezithuthukisiwe.

### v0.1.19 (2026-03-27)
- **Ukutholwa kwe-Claude Code ku-inthanethi**: I-Claude Code engadluli nge-proxy nayo iboniswa ku-inthanethi kudashubhodi.

### v0.1.18 (2026-03-26)
- **Ukulungiswa kwensizakalo ye-fallback ebambekile**: Ngemva kwe-fallback ku-Ollama ngenxa yephutha lesikhashana, uma insizakalo yokuqala ibuyela, ishintsha ibuyele kuyo ngokuzenzakalela.
- **Ukuthuthukiswa kokuthola ukungabi ku-inthanethi**: Ukuhlolwa kwesimo kwamasekhendi angu-15 kwenze ukuthola ukuphuka kwe-proxy kube ngokushesha.

### v0.1.17 (2026-03-25)
- **Ukuhlelwa kwamakhadi ngokudonsela nokudedela**: Amakhadi e-agent angadonselwa ukuze ahlelwe kabusha.
- **Izinkinobho zokusebenzisa okusetiwe ngaphakathi emyaleni**: Inkinobho ethi [⚡ Sebenzisa Okusetiwe] iboniswa kuma-agent angekho ku-inthanethi.
- **Uhlobo olusha lwe-agent cokacdir**.

### v0.1.16 (2026-03-25)
- **Ukuvumelaniswa kwamamodeli nhlangothi zombili**: Ukushintsha imodeli ye-Cline noma Claude Code kudashubhodi ye-vault kubonakala ngokuzenzakalela.

---

*Ukuthola ulwazi olwengeziwe lwe-API, bheka [API.md](API.md).*
