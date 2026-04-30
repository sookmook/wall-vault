# Incwadi Yomuntu Esebenzisayo ye-wall-vault
*(Last updated: 2026-04-16 — v0.2.2)*

---

## Okuqukethwe

1. [Yini i-wall-vault?](#yini-i-wall-vault)
2. [Ukufakwa](#ukufakwa)
3. [Ukuqala Okokuqala (umthakathi we-setup)](#ukuqala-okokuqala)
4. [Ukubhalisa i-API Key](#ukubhalisa-i-api-key)
5. [Indlela Yokusebenzisa i-Proxy](#indlela-yokusebenzisa-i-proxy)
6. [I-Dashboard ye-Key Vault](#i-dashboard-ye-key-vault)
7. [Imodi Yokusabalalisa (Multi Bot)](#imodi-yokusabalalisa-multi-bot)
8. [Ukusetha Ukuqala Ngokuzenzakalela](#ukusetha-ukuqala-ngokuzenzakalela)
9. [Doctor (Udokotela)](#doctor-udokotela)
10. [RTK Ukonga Ama-token](#rtk-ukonga-ama-token)
11. [Ireferensi Yokuguquguquka Kwemvelo](#ireferensi-yokuguquguquka-kwemvelo)
12. [Ukuxazulula Izinkinga](#ukuxazulula-izinkinga)

---

## Amanothi Wokuthuthukiswa kwe-v0.2

- `Service` ilwandle izakala `default_model` kanye `allowed_models`. Imodeli yokuzenzakalela nge-sevisi uya-sethiwa ngokuqondile ekhadini lesevisi.
- `Client.default_service` / `default_model` kubiyiselwe igama futhi kuhunyushwa njengo-`preferred_service` / `model_override`. Uma i-override ishiyile, imodeli yokuzenzakalela yesevisi iyasetshenziswa.
- Ku-startup kuqala kwe-v0.2, i-`vault.json` yesisikhathi sehogo iqondwe ngokuzenzakalela, futhi isimo esikuqaleni kesikhathi senqobo sigodiwe njengo-`vault.json.pre-v02.{timestamp}.bak`.
- I-dashboard ibukelwe kabusha ngezona (izinzuzo ezintathu): kibhoda esekunxele, umugadi wekhadi emidlweni, kanye no-edit slideover okunxele.
- Izindawo ze-API zeMlamali azishintshiwe, kodwa izisekelo zomsiga wokucekeleza/impendulo zibuyele nge-update — isikripti sase-CLI esidala sidinga ukubuyela nge-update.

---

## Izici Ezintsha ze-v0.2.1

- **Ukudluliselwa kwe-Multimodal (OpenAI → Gemini)**: `/v1/chat/completions` manje yamukela izinhlobo eziyisithupha zengxenye yokuqukethwe ngaphezu kuka-`text` — `input_audio`, `input_video`, `input_image`, `input_file`, kanye `image_url` (ama-data URI namaURL angaphandle we-http(s) ≤ 5 MB). I-proxy iguqula ngalunye kube i-`inlineData` ye-Gemini. Amakhasimende avumelana ne-OpenAI anjenge-EconoWorld angadlulisa ama-blob omsindo / isithombe / ividiyo ngqo.
- **Uhlobo lomenzeli we-EconoWorld**: `POST /agent/apply` ene-`agentType: "econoworld"` ibhala izilungiselelo ze-wall-vault ku-`analyzer/ai_config.json` yephrojekthi. I-`workDir` iyamukela uhla olwahlukaniswe ngokhefana lwezindlela zokhetho futhi iguqule izindlela ze-drive ze-Windows zibe izindlela zokumounta ze-WSL.
- **Igridi yezikhiye ze-Dashboard + CRUD**: izikhiye ezingu-11 zikhiqizwa njengamakhadi afingqiwe ane-slideover ye-+ add / ✕ delete.
- **Ukungezwa kwesevisi + ukuhleleka kabusha ngokudonsa-nokulahla**: igridi yesevisi ithola inkinobho ye-+ add kanye nesibambo sokudonsa (`⋮⋮`).
- **Isihloko / isihlathi / ukuhamba kwezihloko / isishintshi solimi** kubuyiselwe. Izihloko eziyisi-7 (cherry/dark/light/ocean/gold/autumn/winter) zidlala umthelela wazo wezinhlayiya kusendlalelo ngemuva kwamakhadi kodwa ngenhla kwesizinda.
- **I-UX yokuchitha i-Slideover**: ukuchofoza ngaphandle noma u-Esc kuvala i-slideover.
- **Isikhombi sesimo se-SSE + isibali-sikhathi sokusebenza**: embhedeni ophezulu (topbar), eduze kwesikhethi solimi/sendikimba, isibali-sikhathi se-`⏱ uptime` nesikhombi se-`● SSE` (okuluhlaza okwesibhakabhaka = kuxhunyiwe, i-orange = kuxhumeka kabusha, okugreyi = akuxhumekile) zibekwe ndawonye (zithuthelwe esihlathini zaya esihlokweni kusukela ku-v0.2.18 — isimo siyabonakala ngaphandle kokuskrola).

---

## v0.2.2 Stability & UX Improvements

- **Dispatch fast-skip**: cloud services whose keys are all on cooldown or exhausted are no longer force-retried. Dispatch moves to the next fallback immediately. Per-request tail latency dropped from ~15 s to ~1.5 s.
- **Fallback model swap**: each fallback step now applies the target service's own `default_model`. Previously a `gemini-2.5-flash` request would be handed to Anthropic/Ollama verbatim and rejected (400/404).
- **Anthropic credit-balance handling**: when Anthropic returns HTTP 400 with a "credit balance" body, the proxy promotes it to 402-equivalent and sets a 30 min cooldown so subsequent dispatches skip Anthropic automatically.
- **Service edit default_model dropdown polish**:
  - The server now renders the complete model list (Google 15, OpenRouter 345, etc.) into the `<select>` from the first open — no second round-trip required.
  - `↓ Move to Allowed` button demotes the current default into the allowed_models textarea and clears the default.
  - `✕ Clear` empties the default in place.
  - Collapsible `Custom input` details block lets you type a model ID directly when the dropdown is unreachable.
- **Agent edit/create model_override dropdown**: free text replaced by a `<select>` populated from the preferred service's `default_model` + `allowed_models`. Changing the preferred service auto-repopulates the override options.
- **ClientInput v0.2 fields**: POST `/admin/clients` now accepts v0.2 canonical `preferred_service` / `model_override` alongside legacy `default_service` / `default_model` (legacy is a fallback).

---

## Yini i-wall-vault?

**wall-vault = Ummeleli we-AI (Proxy) + Isikhwama se-API Key se-OpenClaw**

Ukuze usebenzise amasevisi e-AI, udinga **i-API key**. I-API key injenge **khadi lokungena ledijithali** elifakazela ukuthi "lo muntu unelungelo lokusebenzisa le sevisi". Kodwa-ke le khadi lokungena linemkhawulo wokusetshenziswa ngosuku, futhi uma lingaphathwa kahle, kukhona ingozi yokuvezwa.

I-wall-vault igcina lawa makhadi okungena esikwhameni esiphephile, futhi isebenza njenge **mmeleli (proxy)** phakathi kwe-OpenClaw namasevisi e-AI. Ngokufingqiwe, i-OpenClaw idinga kuphela ukuxhumana ne-wall-vault, bese i-wall-vault ilungisa zonke ezinye izinto eziyinkimbinkimbi.

Izinkinga i-wall-vault ezixazululayo:

- **Ukuzungeza kwe-API Key Ngokuzenzakalela**: Uma ukusetshenziswa kokhiye oyedwa kufinyelela umkhawulo noma kuvinjelwa isikhashana (cooldown), kushintsha ngokuthulile kuye kokhiye olandelayo. I-OpenClaw iyaqhubeka ukusebenza ngaphandle kokuphazanyiswa.
- **Ukushintsha Kwesevisi Ngokuzenzakalela (Fallback)**: Uma i-Google ingaphenduli, kushintshela ku-OpenRouter, futhi uma nalokho kungasebenzi, kushintshela ngokuzenzakalela ku-Ollama·LM Studio·vLLM (i-AI yendawo) efakwe kukhompyutha yakho. Iseshini ayiphuki. Uma isevisi yasekuqaleni ibuyela, izicelo ezilandelayo zibuyela ngokuzenzakalela (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Ukuvumelana Kwesikhathi Samanje (SSE)**: Uma ushintsha imodeli ku-dashboard yesikhwama, ivela kusikrini se-OpenClaw ngaphakathi kwamasekhondi angu-1-3. I-SSE (Server-Sent Events) ingubuchwepheshe lapho iseva idudulela izinguquko kumklayenti ngesikhathi samanje.
- **Izaziso Zesikhathi Samanje**: Izehlakalo njengokuphela kokhiye noma ukwehluleka kwesevisi zivela ngokushesha ngaphansi kwesikrini se-OpenClaw TUI (isikrini setheminali).

> 💡 **I-Claude Code, Cursor, VS Code** nazo zingaxhunyaniswa, kodwa inhloso yokuqala ye-wall-vault ukusetshenziswa ne-OpenClaw.

```
OpenClaw (Isikrini Setheminali ye-TUI)
        │
        ▼
  wall-vault Proxy (:56244)   ← Ukuphatha okhiye, ukuqondisa, fallback, izehlakalo
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (amamodeli angu-340+)
        ├─ Ollama / LM Studio / vLLM (ikhompyutha yakho, isiphephelo sokugcina)
        └─ OpenAI / Anthropic API
```

---

## Ukufakwa

### Linux / macOS

Vula itheminali uphinde unamathisele imiyalo elandelayo njengoba injalo.

```bash
# Linux (PC ejwayelekile, iseva — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Ilanda ifayela ku-inthanethi.
- `chmod +x` — Yenza ifayela elilandiwe "likwazi ukusebenza". Uma weqa lesi sinyathelo, uzothola iphutha lokuthi "imvume yenqatshiwe".

### Windows

Vula i-PowerShell (njengomphathi) bese usebenzisa imiyalo elandelayo.

```powershell
# Landa
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Engeza ku-PATH (isebenza ngemva kokuqala kabusha i-PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **Yini i-PATH?** Yuhlu lwamafolda lapho ikhompyutha ifuna khona imiyalo. Udinga ukuyengeza ku-PATH ukuze ukwazi ukusebenzisa `wall-vault` kunoma yiliphi ifolda.

### Ukwakha Kusukela Kumthombo (Ngabathuthukisi)

Lokhu kusebenza kuphela uma unemvelo yokuthuthukisa yolimi lwe-Go.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (inguqulo: v0.1.25.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Inguqulo Yesitembu Sesikhathi Sokwakha**: Uma wakha nge-`make build`, inguqulo ikhiqizwa ngokuzenzakalela ngefomethi ehlanganisa usuku·isikhathi njenge-`v0.1.27.20260409`. Uma wakha ngokuqondile nge-`go build ./...`, inguqulo ivela kuphela njenge-`"dev"`.

---

## Ukuqala Okokuqala

### Ukusebenzisa umthakathi we-setup

Ngemva kokufaka, qiniseka ukusebenzisa **umthakathi wokusetha** ngomyalo ongezansi kuqala. Umthakathi uzokuhola ngokukubuza izinto ezidingekayo ngayinye ngayinye.

```bash
wall-vault setup
```

Izinyathelo umthakathi azidlulayo yilezi:

```
1. Khetha ulimi (izilimi ezingu-10 okuhlanganisa isiZulu)
2. Khetha indikimba (light / dark / gold / cherry / ocean)
3. Imodi yokusebenza — khetha ukuthi uzosebenzisa wedwa (standalone) noma kumashini amaningi (distributed)
4. Faka igama le-bot — igama elizovela ku-dashboard
5. Ukusetha iphothi — okuzenzakalelayo: proxy 56244, vault 56243 (cindezela u-Enter nje uma ungadingi ukushintsha)
6. Khetha amasevisi e-AI — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Ukusetha isihlungi samathuluzi okuphepha
8. Setha ithokheni yomphathi — iphasiwedi yokuvala izici zokuphatha ze-dashboard. Ingakhiqizwa ngokuzenzakalela
9. Setha iphasiwedi yokubethela i-API key — uma ufuna ukugcina okhiye ngokuphepha okwengeziwe (ukukhetha)
10. Umzila wokugcina ifayela lokusetha
```

> ⚠️ **Qiniseka uyikhumbula ithokheni yomphathi.** Uzoyidinga kamuva uma wengeza okhiye noma ushintsha izilungiselelo ku-dashboard. Uma uyilahlekelwa, uzodinga ukuhlela ngokuqondile ifayela lokusetha.

Uma umthakathi eseqedile, ifayela lokusetha `wall-vault.yaml` likhiqizwa ngokuzenzakalela.

### Ukusebenzisa

```bash
wall-vault start
```

Amaseva amabili aqala ngesikhathi esifanayo:

- **I-Proxy** (`https://localhost:56244`) — Ummeleli oxhumanisa i-OpenClaw namasevisi e-AI
- **I-Key Vault** (`https://localhost:56243`) — Ukuphatha i-API key ne-dashboard yewebhu

Vula `https://localhost:56243` kubhrawuza yakho ukuze ubone i-dashboard ngokushesha.

---

## Ukubhalisa i-API Key

Kukhona izindlela ezine zokubhalisa i-API key. **Ngabaqalayo, iNdlela 1 (okuguquguquka kwemvelo) inconywa**.

### Indlela 1: Okuguquguquka Kwemvelo (Kunconywa — Kulula Kakhulu)

Okuguquguquka kwemvelo **amanani asethwe ngaphambili** uhlelo olufundayo uma luqala. Faka okulandelayo kutheminali.

```bash
# Bhalisa ukhiye we-Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Bhalisa ukhiye we-OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Sebenzisa ngemva kokubhalisa
wall-vault start
```

Uma unokhiye abaningi, bahlanganise ngokhefana (,). I-wall-vault izobaphendukisela ngokuzenzakalela (round robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Iseluleko**: Umyalo we-`export` usebenza kuphela kuseshini yamanje yetheminali. Ukuze uqhubeke ngisho nangemva kokuqala kabusha ikhompyutha, engeza lo mugqa kufayela `~/.bashrc` noma `~/.zshrc`.

### Indlela 2: I-UI ye-Dashboard (Cindezela Ngemawusi)

1. Vula `https://localhost:56243` kubhrawuza
2. Cindezela inkinobho ye-`[+ Engeza]` ekhadini le-**🔑 API Key** ngaphezulu
3. Faka uhlobo lwesevisi, inani lokhiye, ilebuli (igama lesikhumbuzo), nomkhawulo wansuku langa bese ugcina

### Indlela 3: I-REST API (Ngokuzenzakalela·Amaskripti)

I-REST API yindlela izinhlelo ezishintshanisa ngayo idatha nge-HTTP. Iwusizo ngokubhaliswa okuzenzakalelayo ngamaskripti.

```bash
curl -X POST https://localhost:56243/admin/keys \
  -H "Authorization: Bearer ithokheni-yomphathi" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Ukhiye Omkhulu",
    "daily_limit": 1000
  }'
```

### Indlela 4: Ifulegi ye-proxy (Ukuhlola Okwesikhashana)

Sebenzisa lokhu uma ufuna ukufaka ukhiye wesikhashana wokuhlola ngaphandle kokubhalisa okusemthethweni. Ukhiye unyamalala uma uvala uhlelo.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Indlela Yokusebenzisa i-Proxy

### Ukusebenzisa ne-OpenClaw (Inhloso Enkulu)

Nansi indlela yokusetha i-OpenClaw ukuze ixhumane namasevisi e-AI nge-wall-vault.

Vula ifayela `~/.openclaw/openclaw.json` bese wengeza okuqukethwe okulandelayo:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "https://localhost:56244/v1",
        apiKey: "your-agent-token",   // ithokheni yommeleli we-vault
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // 1M context yamahhala
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Indlela Elula Kakhulu**: Cindezela inkinobho ye-**🦞 Kopisha Ukusetha kwe-OpenClaw** ekhadini lommeleli ku-dashboard bese i-snippet enethokheni nekheli esivele igcwalisiwe ikopishelwa ku-clipboard. Namathisela nje.

**`wall-vault/` ngaphambi kwegama lemodeli iqondisaphi?**

Ngokubheka igama lemodeli, i-wall-vault inquma ngokuzenzakalela yiliphi isevisi le-AI elizothumelela isicelo:

| Ifomethi Yemodeli | Isevisi Exhunyiwe |
|----------|--------------|
| `wall-vault/gemini-*` | Google Gemini ngokuqondile |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI ngokuqondile |
| `wall-vault/claude-*` | Anthropic nge-OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (1 miliyoni yama-token yamahhala) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/igama-lemodeli`, `openai/igama-lemodeli`, `anthropic/igama-lemodeli` njll. | Ukuxhuma ngokuqondile nesevisi efanelekile |
| `custom/google/igama-lemodeli`, `custom/openai/igama-lemodeli` njll. | Susa ingxenye ye-`custom/` bese uqondisa kabusha |
| `igama-lemodeli:cloud` | Susa ingxenye ye-`:cloud` bese uxhuma ne-OpenRouter |

> 💡 **Yini i-Context (umongo)?** Yinani lengxoxo i-AI engalikhumbula ngesikhathi esisodwa. 1M (miliyoni yama-token) kusho ukuthi ingakwazi ukucubungula izingxoxo noma imibhalo emide kakhulu ngesikhathi esisodwa.

### Ukuxhuma Ngokuqondile Ngefomethi ye-Gemini API (ukuhambisana namathuluzi akhona)

Uma unamathuluzi abesebenzisa i-Google Gemini API ngokuqondile, shintsha ikheli kuphela ku-wall-vault:

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244/google
```

Noma uma ithuluzi lakho licacisa i-URL ngokuqondile:

```
https://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Ukusebenzisa ne-OpenAI SDK (Python)

Ungaxhuma futhi i-wall-vault kukhodi ye-Python esebenzisa i-AI. Shintsha kuphela `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="https://localhost:56244/v1",
    api_key="not-needed"  # i-wall-vault iphatha i-API key ngokuzenzakalela
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # faka ngefomethi ye-provider/model
    messages=[{"role": "user", "content": "Sawubona"}]
)
```

### Ukushintsha Imodeli Ngesikhathi Sokusebenza

Ukushintsha imodeli ye-AI ngesikhathi i-wall-vault isisebenza:

```bash
# Shintsha imodeli ngokucela i-proxy ngokuqondile
curl -X PUT https://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# Kuyi-mode yokusabalalisa (multi bot), shintsha kuseva ye-vault → ivela ngokushesha nge-SSE
curl -X PUT https://localhost:56243/admin/clients/id-yebot-yami \
  -H "Authorization: Bearer ithokheni-yomphathi" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Ukuhlola Uhlu Lwamamodeli Atholakalayo

```bash
# Buka uhlu oluphelele
curl https://localhost:56244/api/models | python3 -m json.tool

# Buka amamodeli e-Google kuphela
curl "https://localhost:56244/api/models?service=google"

# Sesha ngegama (isibonelo: amamodeli anokuthi "claude")
curl "https://localhost:56244/api/models?q=claude"
```

**Isifinyezo Samamodeli Amakhulu Ngesevisi:**

| Isevisi | Amamodeli Amakhulu |
|--------|----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ (Hunter Alpha 1M context yamahhala, DeepSeek R1/V3, Qwen 2.5 njll.) |
| Ollama | Ithola ngokuzenzakalela iseva yendawo efakwe kukhompyutha yakho |
| LM Studio | Iseva yendawo yekhompyutha (iphothi 1234) |
| vLLM | Iseva yendawo yekhompyutha (iphothi 8000) |
| llama.cpp | Iseva yendawo yekhompyutha (iphothi 8080) |

---

## I-Dashboard ye-Key Vault

Vula `https://localhost:56243` kubhrawuza yakho ukuze ubone i-dashboard.

**Isakhiwo Sesikrini:**
- **Ibha ephezulu ehleli (topbar)**: Ilogo, isikhethi solimi·sendikimba, isimo sokuxhuma kwe-SSE
- **I-Grid Yamakhadi**: Amakhadi ommeleli·esevisi·e-API key ahlelelwe njengezitayela

### Ikhadi le-API Key

Ikhadi lokuphatha okhiye be-API ababhalisiwe ngokushesha.

- Ibonisa uhlu lokhiye oluhlukaniswe ngesevisi.
- `today_usage`: Inani lama-token (izinhlamvu i-AI ezifundile nezibhalile) ezisetshenziswe ngempumelelo namhlanje
- `today_attempts`: Isamba samakholi namhlanje (impumelelo + ukwehluleka)
- Bhalisa ukhiye omusha ngenkinobho ye-`[+ Engeza]`, bese ususa ukhiye nge-`✕`.

> 💡 **Yini i-Token?** Iyiyunithi esetshenziswa yi-AI uma icubungula umbhalo. Cishe yigama linye lesiNgisi, noma izinhlamvu ezingu-1-2 zezinye izilimi. Izindleko ze-API ngokuvamile zibalwa ngokwaleli nani lama-token.

### Ikhadi Lommeleli

Ikhadi elibonisa isimo sama-bot (ababili) axhunyaniswe ne-proxy ye-wall-vault.

**Isimo sokuxhuma siboniswa ngamazinga angu-4:**

| Isibonakaliso | Isimo | Incazelo |
|------|------|------|
| 🟢 | Iyasebenza | I-proxy isebenza ngokujwayelekile |
| 🟡 | Ibambezelekile | Izimpendulo ziyeza kodwa ngokucotha |
| 🔴 | Ayikho ku-Inthanethi | I-proxy ayiphenduli |
| ⚫ | Ayixhunywanga·Ikhutshaziwe | I-proxy ayikaze ixhumane ne-vault noma ikhutshaziwe |

**Umhlahlandlela wamaqhosha ngaphansi kwekhadi lommeleli:**

Uma ucacisa **uhlobo lommeleli** uma ubhalisa ummeleli, amaqhosha okulula ahambisana nalolo hlobo avela ngokuzenzakalela.

---

#### 🔘 Iqhosha Lokukopisha Ukusetha — Lakha ukusetha kokuxhuma ngokuzenzakalela

Uma ucindezela iqhosha, i-snippet yokusetha enethokheni yommeleli, ikheli le-proxy, nolwazi lwemodeli esivele igcwalisiwe ikopishelwa ku-clipboard. Namathisela okukopiwe endaweni eboniswe ethebuleni elingezansi ukuze uqedele ukusetha kokuxhuma.

| Iqhosha | Uhlobo Lommeleli | Indawo Yokunamathisela |
|------|-------------|-------------|
| 🦞 Kopisha Ukusetha kwe-OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Kopisha Ukusetha kwe-NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Kopisha Ukusetha kwe-Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Kopisha Ukusetha kwe-Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Kopisha Ukusetha kwe-VSCode | `vscode` | `~/.continue/config.json` |

**Isibonelo — Uma uhlobo luyiClaude Code, okuqukethwe okufana nalokhu kuyakopishwa:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "ithokheni-yalo-mmeleli"
}
```

**Isibonelo — Uma uhlobo luyi-VSCode (Continue):**

```yaml
# ~/.continue/config.yaml  ← Namathisela ku-config.yaml, hhayi ku-config.json
name: My Config
version: 0.0.1
schema: v1

models:
  - name: wall-vault proxy
    provider: openai
    model: gemini-2.5-flash
    apiBase: http://192.168.1.20:56244/v1
    apiKey: ithokheni-yalo-mmeleli
    roles:
      - chat
      - edit
      - apply
```

> ⚠️ **Inguqulo yakamuva ye-Continue isebenzisa `config.yaml`.** Uma `config.yaml` ikhona, `config.json` izitshalelwa ngokuphelele. Qiniseka unamathisela ku-`config.yaml`.

**Isibonelo — Uma uhlobo luyiCursor:**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : ithokheni-yalo-mmeleli

// Noma okuguquguquka kwemvelo:
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=ithokheni-yalo-mmeleli
```

> ⚠️ **Ukukopisha ku-clipboard akusebenzi**: Izinqubomgomo zokuphepha zebhrawuza zingavimba ukukopisha. Uma ibhokisi lombhalo livela ku-popup, khetha konke nge-Ctrl+A bese ukopisha nge-Ctrl+C.

---

#### ⚡ Iqhosha Lokusebenzisa Ngokuzenzakalela — Cindezela kanye bese ukusetha kuqediwe

Uma uhlobo lommeleli kungu-`cline`, `claude-code`, `openclaw`, noma `nanoclaw`, iqhosha le-**⚡ Sebenzisa Ukusetha** livela ekhadini lommeleli. Uma ucindezela leli qhosha, amafayela okusetha endawo ommeleli ofanelekile abuyekezwa ngokuzenzakalela.

| Iqhosha | Uhlobo Lommeleli | Ifayela Eliqondiswe |
|------|-------------|-------------|
| ⚡ Sebenzisa Ukusetha kwe-Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Sebenzisa Ukusetha kwe-Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Sebenzisa Ukusetha kwe-OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Sebenzisa Ukusetha kwe-NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Leli qhosha lithumela isicelo ku-**localhost:56244** (i-proxy yendawo). I-proxy kufanele isebenze kuleyo mashini ukuze isebenze.

---

#### 🔀 Ukuhlelwa Kwamakhadi Ngokudonsa Nokudedela (v0.1.17, ithuthukisiwe v0.1.25)

Ungadonsa amakhadi ommeleli ku-dashboard ukuze uwahlelele kabusha ngendlela oyifunayo.

1. Bamba indawo ye-**thafikhilayithi (●)** phezulu kwesokunxele sekhadi ngemawusi bese udonsa
2. Idedele phezu kwekhadi endaweni oyifunayo bese uhlelo lushintsha

> 💡 Umzimba wekhadi (izindawo zokufaka, amaqhosha njll.) awudonseleki. Ungabamba kuphela endaweni yethafikhilayithi.

#### 🟠 Ukuthola Inqubo Yommeleli (v0.1.25)

Uma i-proxy isebenza ngokujwayelekile kodwa inqubo yommeleli wendawo (NanoClaw, OpenClaw) ifile, ithafikhilayithi yekhadi ishintsha iba **nsomi (imenyezela)** bese umyalezo othi "Inqubo yommeleli imile" uvela.

- 🟢 Luhlaza: I-proxy + ummeleli ngokujwayelekile
- 🟠 Nsomi (imenyezela): I-proxy ngokujwayelekile, ummeleli ufile
- 🔴 Bomvu: I-proxy ayikho ku-inthanethi
3. Uhlelo olushintshiwe **lugcinwa kuseva ngokushesha** futhi luhlala ngisho ngemva kokuvuselela ikhasi

> 💡 Kumadivayisi okuthinta (iselula/ithebhulethi) akukasekelwa okwamanje. Sebenzisa ibhrawuza yedeskhithopu.

---

#### 🔄 Ukuvumelana Kwemodeli Ngezinhlangothi Ezimbili (v0.1.16)

Uma ushintsha imodeli yommeleli ku-dashboard ye-vault, ukusetha kwendawo kommeleli ofanelekile kubuyekezwa ngokuzenzakalela.

**Nge-Cline:**
- Uma ushintsha imodeli ku-vault → isehlakalo se-SSE → i-proxy ibuyekeza inkambu yemodeli ku-`globalState.json`
- Izinkambu ezibuyekezwayo: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` ne-API key azithintwa
- **Kudingeka ukuvuselela kabusha i-VS Code (`Ctrl+Alt+R` noma `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - Ngoba i-Cline ayifundi kabusha ifayela lokusetha ngesikhathi isebenza

**Nge-Claude Code:**
- Uma ushintsha imodeli ku-vault → isehlakalo se-SSE → i-proxy ibuyekeza inkambu ye-`model` ku-`settings.json`
- Isesha ngokuzenzakalela izindlela zikamabili ze-WSL ne-Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Uhlangothi Olubuyayo (ummeleli → vault):**
- Uma ummeleli (Cline, Claude Code njll.) ethumela isicelo ku-proxy, i-proxy ifaka ulwazi lwesevisi·lwemodeli yomklayenti ofanelekile ku-heartbeat
- Isevisi/imodeli esetshenziswa manje ivela ngesikhathi samanje ekhadini lommeleli ku-dashboard ye-vault

> 💡 **Okubalulekile**: I-proxy ihlonza ummeleli ngethokheni ye-Authorization yesicelo, futhi iqondisa ngokuzenzakalela kusevisi/imodeli esethwe ku-vault. Ngisho noma i-Cline noma i-Claude Code ithumela igama lemodeli elihlukile, i-proxy iyabaphambukisa ngokusetha kwe-vault.

---

### Ukusebenzisa i-Cline ku-VS Code — Umhlahlandlela Ophelele

#### Isinyathelo 1: Faka i-Cline

Faka i-**Cline** (ID: `saoudrizwan.claude-dev`) kusuka e-VS Code Extension Marketplace.

#### Isinyathelo 2: Bhalisa Ummeleli ku-Vault

1. Vula i-dashboard ye-vault (`http://IP-ye-vault:56243`)
2. Cindezela **+ Engeza** esigabeni saba-**Mmeleli**
3. Faka njengokulandelayo:

| Inkambu | Inani | Incazelo |
|------|----|------|
| ID | `cline_yami` | Isihlonzi esiyingqayizivele (isiNgisi, ngaphandle kwezikhala) |
| Igama | `Cline Yami` | Igama elizovela ku-dashboard |
| Uhlobo Lommeleli | `cline` | ← Kufanele ukhethe `cline` |
| Isevisi | Khetha isevisi (isibonelo: `google`) | |
| Imodeli | Faka imodeli (isibonelo: `gemini-2.5-flash`) | |

4. Cindezela **Gcina** bese ithokheni ikhiqizwa ngokuzenzakalela

#### Isinyathelo 3: Xhuma ne-Cline

**Indlela A — Ukusebenzisa Ngokuzenzakalela (Kunconywa)**

1. Qiniseka ukuthi **i-proxy** ye-wall-vault iyasebenza kuleyo mashini (`localhost:56244`)
2. Cindezela iqhosha le-**⚡ Sebenzisa Ukusetha kwe-Cline** ekhadini lommeleli ku-dashboard
3. Uma ubona isaziso esithi "Ukusetha kusetshenziswe ngempumelelo!" kuphumelele
4. Vuselela kabusha i-VS Code (`Ctrl+Alt+R`)

**Indlela B — Ukusetha Ngesandla**

Vula izilungiselelo (⚙️) kubha yesokunxele ye-Cline:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://ikheli-le-proxy:56244/v1`
  - Imashini efanayo: `https://localhost:56244/v1`
  - Enye imashini njenge-mini server: `http://192.168.1.20:56244/v1`
- **API Key**: Ithokheni etholwe ku-vault (kopisha ekhadini lommeleli)
- **Model ID**: Imodeli esethwe ku-vault (isibonelo: `gemini-2.5-flash`)

#### Isinyathelo 4: Qinisekisa

Thumela noma yimuphi umyalezo engxoxweni ye-Cline. Uma kujwayelekile:
- Iqhaza **eliluhlaza (● Iyasebenza)** lizovela ekhadini lommeleli elifanelekile ku-dashboard ye-vault
- Isevisi/imodeli yamanje ivela ekhadini (isibonelo: `google / gemini-2.5-flash`)

#### Ukushintsha Imodeli

Uma ufuna ukushintsha imodeli ye-Cline, shintsha ku-**dashboard ye-vault**:

1. Shintsha isevisi/imodeli kumenyu yokudonsa yekhadi lommeleli
2. Cindezela **Sebenzisa**
3. Vuselela kabusha i-VS Code (`Ctrl+Alt+R`) — igama lemodeli ezansi kwe-Cline lizobuyekezwa
4. Imodeli entsha izosetshenziswa kusukela esiceli esilandelayo

> 💡 Eqinisweni, i-proxy ihlonza isicelo se-Cline ngethokheni bese isiqondisa kumodeli yokusetha ye-vault. Ngisho noma ungayivuseleli kabusha i-VS Code, **imodeli esetshenziswa ishintsha ngokushesha** — ukuvuselela kabusha kungokubuyekeza isiboniso semodeli ku-UI ye-Cline.

#### Ukuthola Ukuphuka Kokuxhuma

Uma uvala i-VS Code, ikhadi lommeleli ku-dashboard ye-vault lizoshintsha libe phuzi (libambezelekile) ngemva cishe **kwamasekhondi angu-90**, libe bomvu (alikho ku-inthanethi) ngemva **kwemizuzu engu-3**. (Kusukela ku-v0.1.18, ukuhlolwa kwesimo njalo ngamasekhondi angu-15 kwasheshisa ukutholwa kwesimo sokungabi ku-inthanethi.)

#### Ukuxazulula Izinkinga

| Isibonakaliso | Imbangela | Isixazululo |
|------|------|------|
| Iphutha lokuthi "Ukuxhuma kwehlulekile" ku-Cline | I-proxy ayisebenzi noma ikheli alifanele | Hlola i-proxy nge-`curl https://localhost:56244/health` |
| Iqhaza eliluhlaza aliveli ku-vault | I-API key (ithokheni) ayisethwanga | Cindezela iqhosha le-**⚡ Sebenzisa Ukusetha kwe-Cline** futhi |
| Imodeli ezansi kwe-Cline ayishintshi | I-Cline igcina ukusetha ku-cache | Vuselela kabusha i-VS Code (`Ctrl+Alt+R`) |
| Igama lemodeli elingalungile livela | Iphutha lakudala (lalungiswa ku-v0.1.16) | Buyekeza i-proxy ku-v0.1.16 noma ngaphezulu |

---

#### 🟣 Iqhosha Lokukopisha Umyalo Wokusabalalisa — Lisetshenziswa uma ufaka kumashini entsha

Lisetshenziswa uma ufaka i-proxy ye-wall-vault kukhompyutha entsha nokuyixhuma ne-vault okokuqala. Cindezela iqhosha bese iskripti sonke sokufaka siyakopishwa. Namathisela bese usisebenzisa kutheminali yekhompyutha entsha bese okulandelayo kusingathwa ngesikhathi esisodwa:

1. Faka i-binary ye-wall-vault (iyeqiwa uma isifakiwe kakade)
2. Ukubhalisa okuzenzakalela kwesevisi ye-systemd yomsebenzisi
3. Qala isevisi bese uxhuma ngokuzenzakalela ne-vault

> 💡 Iskripti sinayo ithokheni yalo mmeleli nekheli leseva ye-vault esivele kugcwalisiwe, ngakho ungasisebenzisa ngokushesha ngemva kokunamathisela ngaphandle kwanoma yiluphi ushintsho.

---

### Ikhadi Lesevisi

Ikhadi lokuvula, lokuvala, noma lokusetha amasevisi e-AI azosetshenziwa.

- Iswitshi yokuvula·yokuvala esevisi ngayinye
- Uma ufaka ikheli leseva ye-AI yendawo (i-Ollama, i-LM Studio, i-vLLM, i-llama.cpp njll. esebenza kukhompyutha yakho), izothola ngokuzenzakalela amamodeli atholakalayo.
- **Isiboniso sesimo sokuxhuma sesevisi yendawo**: Iqhaza le-● eceleni kwegama lesevisi uma **liluhlaza** kuxhunyiwe, **limpunga** akuxhunyiwe
- **Ithafikhilayithi yendawo yesevisi ngokuzenzakalela** (v0.1.23+): Amasevisi endawo (Ollama, LM Studio, vLLM, llama.cpp) ayavulwa/ayavala ngokuzenzakalela ngokuya ngokuthi ukuxhuma kungatholakala yini. Uma uvula isevisi, ngaphakathi kwamasekhondi angu-15 ● iba luhlaza nebhokisi lokuhlola livulwa, futhi uma uvala isevisi, kuvala ngokuzenzakalela. Indlela efanayo namasevisi efu (Google, OpenRouter njll.) ashintsha ngokuzenzakalela ngokuya ngokutholakala kwe-API key.
- **Iswitshi yemodi yokucabanga** (v0.2.17+): Ezansi ewindini lokuhlela isevisi yendawo, kuvela ibhokisi lokuhlola le-**modi yokucabanga**. Uma uyivula, i-proxy ifaka `"reasoning": true` emzimbeni we-chat-completions othunyelwa kuseva ephezulu, ukuze amamodeli asekela ukukhipha inqubo yokucabanga njenge-DeepSeek R1, Qwen QwQ abuyise nebhloki le-`<think>…</think>` kanye nayo. Iziseva ezingasazi lesi sikhala ziyasishaya indiva, ngakho ungasishiya kuvuliwe ngokuphephile ngisho nasekulethwembula okuxubile.

> 💡 **Uma isevisi yendawo isebenza kukhompyutha enye**: Faka i-IP yaleyo khompyutha endaweni yokufaka i-URL yesevisi. Isibonelo: `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio), `http://192.168.1.20:8080` (llama.cpp). Uma isevisi iboshwe ku-`127.0.0.1` kuphela esikhundleni sika-`0.0.0.0`, ukufinyelela nge-IP yangaphandle ngeke kusebenze, ngakho hlola ikheli lokuboshwa ezilungiselelweni zesevisi.

### Ukufaka Ithokheni Yomphathi

Uma uzama ukusebenzisa izici ezibalulekile njengokwengeza·ukususa okhiye ku-dashboard, i-popup yokufaka ithokheni yomphathi izovela. Faka ithokheni owayisetha kumthakathi we-setup. Ngemva kokufaka kanye, ihlala kuze uvale ibhrawuza.

> ⚠️ **Uma ukwehluleka kokuqinisekisa kudlula izikhathi ezingu-10 ngaphakathi kwemizuzu engu-15, i-IP efanelekile ivalwa okwesikhashana.** Uma ukhohlwe ithokheni, hlola into ye-`admin_token` kufayela le-`wall-vault.yaml`.

---

## Imodi Yokusabalalisa (Multi Bot)

Uma usebenzisa i-OpenClaw kumakhompyutha amaningi ngesikhathi esifanayo, lokhu kuyisakhiwo lapho **isikhwama sokhiye esisodwa sabiwa**. Kulula ngoba udinga kuphatha okhiye endaweni eyodwa kuphela.

### Isibonelo Sesakhiwo

```
[Iseva ye-Key Vault]
  wall-vault vault    (Key Vault :56243, dashboard)

[WSL Alpha]           [Raspberry Pi Gamma]    [Mac Mini Yendawo]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ Ukuvumelana kwe-SSE ↕ Ukuvumelana kwe-SSE  ↕ Ukuvumelana kwe-SSE
```

Wonke ama-bot abheka iseva ye-vault ephakathi, ngakho uma ushintsha imodeli noma wengeza ukhiye ku-vault, kuvela kuwo wonke ama-bot ngokushesha.

### Isinyathelo 1: Qala Iseva ye-Key Vault

Sebenzisa kukhompyutha ezosetshenziselwa iseva ye-vault:

```bash
wall-vault vault
```

### Isinyathelo 2: Bhalisa I-bot Ngayinye (Umklayenti)

Bhalisa ngaphambi kwesikhathi ulwazi lwe-bot ngayinye exhumana neseva ye-vault:

```bash
curl -X POST https://localhost:56243/admin/clients \
  -H "Authorization: Bearer ithokheni-yomphathi" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "BotA",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### Isinyathelo 3: Qala I-proxy Kukhompyutha Ye-bot Ngayinye

Sebenzisa i-proxy ngokucacisa ikheli leseva ye-vault nethokheni kukhompyutha ngayinye lapho i-bot ifakwe khona:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Shintsha ingxenye ye-**`192.168.x.x`** ngekheli le-IP langaphakathi langempela lekhompyutha yeseva ye-vault. Ungalihlola ezilungiselelweni zerawutha noma ngomyalo we-`ip addr`.

---

## Ukusetha Ukuqala Ngokuzenzakalela

Uma kukukhathaza ukuvula i-wall-vault ngesandla njalo uma uqala kabusha ikhompyutha, ibhalise njengesevisi yesistimu. Ngemva kokubhalisa kanye, iqala ngokuzenzakalela ngesikhathi sokuqala.

### Linux — systemd (i-Linux iningi)

I-systemd yisistimu eqala·ephatha izinhlelo ngokuzenzakalela ku-Linux:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Ukuhlola amalogi:

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

Yisistimu ephatha ukuqala ngokuzenzakalela kwezinhlelo ku-macOS:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Landa i-NSSM ku-[nssm.cc](https://nssm.cc/download) bese uyengeza ku-PATH.
2. Ku-PowerShell yomphathi:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor (Udokotela)

Umyalo we-`doctor` yithuluzi **elizihlolayo nelizilungisayo** uma i-wall-vault isethwe ngendlela efanele.

```bash
wall-vault doctor check   # Hlola isimo samanje (funda kuphela, ungashintshi lutho)
wall-vault doctor fix     # Lungisa izinkinga ngokuzenzakalela
wall-vault doctor all     # Hlola + lungisa ngokuzenzakalela ngesikhathi esisodwa
```

> 💡 Uma okuthile kubonakala kungalungile, sebenzisa `wall-vault doctor all` kuqala. Kubamba izinkinga eziningi ngokuzenzakalela.

---

## RTK Ukonga Ama-token

*(v0.1.24+)*

**RTK (Ithuluzi Lokonga Ama-token)** licindezela ngokuzenzakalela okukhishwa yimiyalo ye-shell esetshenziswa yi-AI coding agent (njengo-Claude Code) ukuze kuncishiswe ukusetshenziwa kwama-token. Isibonelo, okukhishwa kwemigqa engu-15 kwe-`git status` kucinyezwa kube isifinyezo semigqa engu-2.

### Ukusebenzisa Okuyisisekelo

```bash
# Goqela umyalo nge-wall-vault rtk bese okukhishwa kuhlungwa ngokuzenzakalela
wall-vault rtk git status          # Uhlu lwamafayela ashintshiwe kuphela
wall-vault rtk git diff HEAD~1     # Imigqa eshintshiwe + umongo omncane
wall-vault rtk git log -10         # Hash + umyalezo womugqa owodwa ngamunye
wall-vault rtk go test ./...       # Ukuhlolwa okwehlulekile kuphela
wall-vault rtk ls -la              # Imiyalo engasekelwa inqanyulwa ngokuzenzakalela
```

### Imiyalo Esekelwayo Nomthelela Wokunciphisa

| Umyalo | Indlela Yesihlungi | Amazinga Okunciphisa |
|------|----------|--------|
| `git status` | Isifinyezo samafayela ashintshiwe kuphela | ~87% |
| `git diff` | Imigqa eshintshiwe + umongo wemigqa engu-3 | ~60-94% |
| `git log` | Hash + umyalezo womugqa wokuqala | ~90% |
| `git push/pull/fetch` | Susa inqubekela phambili, isifinyezo kuphela | ~80% |
| `go test` | Bonisa okwehlulekile kuphela, bala okuphasiswe | ~88-99% |
| `go build/vet` | Bonisa amaphutha kuphela | ~90% |
| Yonke eminye imiyalo | Imigqa engu-50 yokuqala + imigqa engu-50 yokugcina, okungaphezu kuka-32KB | Iyaguquguquka |

### I-Pipeline Yesihlungi Yezinyathelo Ezingu-3

1. **Isihlungi sesakhiwo somyalo ngamunye** — Siqonda ifomethi yokukhishwa kwe-git, go njll. bese sikhipha izingxenye ezibalulekile kuphela
2. **Ukucubungula okulandelayo kwe-regex** — Susa amakhodi ombala we-ANSI, nciphisa imigqa engenalutho, hlanganisa imigqa ephindaphindayo
3. **Ukudlula + ukunqamula** — Imiyalo engasekelwayo igcina imigqa engu-50 yokuqala/yokugcina kuphela

### Ukuxhuma ne-Claude Code

Ungasetha i-hook ye-`PreToolUse` ye-Claude Code ukuze yonke imiyalo ye-shell idlule nge-RTK ngokuzenzakalela.

```bash
# Faka i-hook (yengezwa ngokuzenzakalela ku-settings.json ye-Claude Code)
wall-vault rtk hook install
```

Noma yengeze ngesandla ku-`~/.claude/settings.json`:

```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "Bash",
      "command": "wall-vault rtk rewrite"
    }]
  }
}
```

> 💡 **Ukugcinwa kwe-Exit code**: I-RTK ibuyisela i-exit code yomyalo woqobo njengoba injalo. Uma umyalo wehluleka (exit code ≠ 0), i-AI nayo ithola ukwehluleka ngokunembe.

> 💡 **Ukuphoqwa kwesiNgisi**: I-RTK isebenzisa imiyalo nge-`LC_ALL=C` ukuze ikhiqize okukhishwa kwesiNgisi njalo ngaphandle kokukhathazeka ngezilungiselelo zolimi lwesistimu. Lokhu kuqinisekisa ukuthi isihlungi sisebenza ngokunembe.

---

## Ireferensi Yokuguquguquka Kwemvelo

Okuguquguquka kwemvelo yindlela yokudlulisa amanani okusetha ohlelweni. Faka ngefomethi ye-`export igama-lokuguquguquka=inani` kutheminali, noma kufake kufayela lesevisi yokuqala ngokuzenzakalela ukuze kusebenze njalo.

| Okuguquguquka | Incazelo | Inani Lesibonelo |
|------|------|---------|
| `WV_LANG` | Ulimi lwe-dashboard | `ko`, `en`, `ja` |
| `WV_THEME` | Indikimba ye-dashboard | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | I-API key ye-Google (eziningi ngokhefana) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | I-API key ye-OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Ikheli leseva ye-vault kuyi-mode yokusabalalisa | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Ithokheni yokuqinisekisa yomklayenti (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Ithokheni yomphathi | `admin-token-here` |
| `WV_MASTER_PASS` | Iphasiwedi yokubethela i-API key | `my-password` |
| `WV_AVATAR` | Umzila wefayela lesithombe se-avatar (umzila ohlobene kusuka ku-`~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Ikheli leseva yendawo ye-Ollama | `http://192.168.x.x:11434` |

---

## Ukuxazulula Izinkinga

### Uma I-proxy Ingaqali

Ngokuvamile kungenxa yokuthi iphothi isisetshenziselwe esinye isinhlelo.

```bash
ss -tlnp | grep 56244   # Hlola ukuthi ubani osebenzisa iphothi 56244
wall-vault proxy --port 8080   # Qala ngenombolo yephothi ehlukile
```

### Uma Amaphutha E-API Key Enzeka (429, 402, 401, 403, 582)

| Ikhodi Yephutha | Incazelo | Isixazululo |
|----------|------|----------|
| **429** | Izicelo eziningi kakhulu (ukusetshenziswa kwedlulile) | Linda kancane noma wengeze okunye ukhiye |
| **402** | Inkokhelo idingekile noma ikhredithi ayenele | Gcwalisa ikhredithi kusevisi efanelekile |
| **401 / 403** | Ukhiye awulungile noma awunavumo | Hlola futhi inani lokhiye bese ubhalisa kabusha |
| **582** | I-Gateway inemithwalo eningi (cooldown imizuzu 5) | Ixazulula ngokuzenzakalela ngemva kwemizuzu 5 |

```bash
# Hlola uhlu nesimo sokhiye ababhalisiwe
curl -H "Authorization: Bearer ithokheni-yomphathi" https://localhost:56243/admin/keys

# Setha kabusha izibali zokusetshenziwa kokhiye
curl -X POST -H "Authorization: Bearer ithokheni-yomphathi" https://localhost:56243/admin/keys/reset
```

### Uma Ummeleli Evela Njenge-"Ayixhunywanga"

"Ayixhunywanga" kusho ukuthi inqubo ye-proxy ayithumeli isignali (heartbeat) ku-vault. **Akusho ukuthi ukusetha akugcinwanga.** I-proxy idinga ukusebenza uyazi ikheli leseva ye-vault nethokheni ukuze ishintshe ibe yisimo sokuxhumana.

```bash
# Qala i-proxy ucacise ikheli leseva ye-vault, ithokheni, ne-ID yomklayenti
WV_VAULT_URL=http://ikheli-leseva-ye-vault:56243 \
WV_VAULT_TOKEN=ithokheni-yomklayenti \
WV_VAULT_CLIENT_ID=id-yomklayenti \
wall-vault proxy
```

Uma ukuxhuma kuphumelela, kuzoshintsha kube 🟢 Iyasebenza ku-dashboard ngaphakathi cishe kwamasekhondi angu-20.

### Uma Ukuxhuma kwe-Ollama Kungasebenzi

I-Ollama uhlelo olusebenzisa i-AI ngokuqondile kukhompyutha yakho. Okokuqala hlola ukuthi i-Ollama ivuliwe yini.

```bash
curl http://localhost:11434/api/tags   # Uma uhlu lwamamodeli luvela, kujwayelekile
export OLLAMA_URL=http://192.168.x.x:11434   # Uma isebenza kukhompyutha enye
```

> ⚠️ Uma i-Ollama ingaphenduli, qala i-Ollama kuqala ngomyalo we-`ollama serve`.

> ⚠️ **Amamodeli amakhulu mancane**: Amamodeli amakhulu njengo-`qwen3.5:35b`, `deepseek-r1` angathatha imizuzu eminingana ukukhiqiza impendulo. Ngisho noma kubonakala sengathi ayikho impendulo, kungenzeka kucutshungulwa ngokujwayelekile, ngakho linda.

---

## Izinguquko Zakamuva (v0.1.16 ~ v0.1.27)

### v0.1.27 (2026-04-09)
- **Ukulungiswa kwegama lemodeli ye-fallback ye-Ollama**: Kulungiswe inkinga lapho igama lemodeli elinophawu lwangaphambi kwe-provider (isibonelo: `google/gemini-3.1-pro-preview`) lalithunyelwa ku-Ollama njengoba linjalo ngesikhathi se-fallback kusuka kwesinye isevisi. Manje kufakwa ngokuzenzakalela okuguquguquka kwemvelo/imodeli ezenzakalelayo.
- **Ukuncishiswa okukhulu kwesikhathi se-cooldown**: 429 rate limit 30min→5min, 402 inkokhelo 1hr→30min, 401/403 24hr→6hr. Ukuvimbela isimo lapho bonke okhiye bese-cooldown ngesikhathi esifanayo okwenza i-proxy ime ngokuphelele.
- **Ukuzama kabusha ngokuqinile ngesikhathi se-cooldown ephelele**: Uma bonke okhiye besezimeni se-cooldown, ukhiye ovuleka kuqala uzanywa kabusha ngokuqinile ukuze kuvinjwe ukunqatshwa kwezicelo.
- **Ukulungiswa kokuboniswa kohlu lwesevisi**: Izimpendulo ze-`/status` zibonisa uhlu lwangempela lwesevisi oluvumelaniswe kusuka ku-vault (ukuvimbela ukungabi khona kwe-anthropic njll.).

### v0.1.25 (2026-04-08)
- **Ukuthola inqubo yommeleli**: I-proxy ithola ukuthi ummeleli wendawo (NanoClaw/OpenClaw) uphila yini futhi ibonisa ngethafikhilayithi ensomi ku-dashboard.
- **Ukuthuthukiswa kwesibambi sokudonsa**: Kushintshiwe ukuze ubambe kuphela endaweni yethafikhilayithi (●) uma uhlela amakhadi. Akusakwazeki ukudonsa ngephutha ezindaweni zokufaka noma amaqhosha.

### v0.1.24 (2026-04-06)
- **Umyalo omncane we-RTK wokonga ama-token**: `wall-vault rtk <command>` ihlunga ngokuzenzakalela okukhishwa kwemiyalo ye-shell ukuze kuncishiswe ukusetshenziwa kwama-token ye-AI agent ngo-60-90%. Iqukethe izihlungi ezikhethekile zemiyalo emikhulu njengo-git, go, futhi imiyalo engasekelwayo nayo inqunywa ngokuzenzakalela. Ixhuma ngokusobala nge-hook ye-`PreToolUse` ye-Claude Code.

### v0.1.23 (2026-04-06)
- **Ukulungiswa kokushintsha imodeli ye-Ollama**: Kulungiswe inkinga lapho ukushintsha imodeli ye-Ollama ku-dashboard ye-vault kwakungaveli ku-proxy. Ngaphambilini kwakusetshenziswa kuphela okuguquguquka kwemvelo (`OLLAMA_MODEL`), manje ukusetha kwe-vault kunikezwa phambili.
- **Ithafikhilayithi yendawo yesevisi ngokuzenzakalela**: I-Ollama·LM Studio·vLLM iqalwa ngokuzenzakalela uma ukuxhuma kungatholakala, futhi icinywe ngokuzenzakalela uma iphukile. Indlela efanayo nokushintsha okuzenzakalelayo kwamasevisi efu ngokuya kokhiye.

### v0.1.22 (2026-04-05)
- **Ukulungiswa kwenkambu ye-content engenalutho engazange iboniswe**: Uma amamodeli okucabanga (gemini-3.1-pro, o1, claude thinking njll.) esebenzisa umkhawulo we-max_tokens ku-reasoning futhi engakwazi ukukhiqiza impendulo yangempela, i-proxy yayisusa izinkambu ze-`content`/`text` ze-JSON yempendulo nge-`omitempty`, okwenza ama-SDK omklayenti e-OpenAI/Anthropic awe ngephutha le-`Cannot read properties of undefined (reading 'trim')`. Kushintshiwe ukuze kuhlale kufakwa izinkambu njengoba kusho izimiso ezisemthethweni ze-API.

### v0.1.21 (2026-04-05)
- **Ukusekelwa kwemodeli ye-Gemma 4**: Amamodeli omndeni we-Gemma njengo-`gemma-4-31b-it`, `gemma-4-26b-a4b-it` angasetshenziswa nge-Google Gemini API.
- **Ukusekelwa okusemthethweni kwesevisi ye-LM Studio / vLLM**: Ngaphambilini la masevisi ayeshiywe ngaphandle ekuqondiseni kwe-proxy futhi njalo ayebuyiselwa ku-Ollama. Manje aqondiswa ngokujwayelekile nge-API ehambisana ne-OpenAI.
- **Ukulungiswa kokuboniswa kwesevisi ku-dashboard**: Ngisho noma i-fallback yenzekile, i-dashboard njalo ibonisa isevisi esethwe umsebenzisi.
- **Isiboniso sesimo sesevisi yendawo**: Isimo sokuxhuma samasevisi endawo (Ollama, LM Studio, vLLM njll.) siboniswa ngombala weqhaza le-● uma i-dashboard ilayishwa.
- **Okuguquguquka kwemvelo kwesihlungi samathuluzi**: Imodi yokudlulisa amathuluzi (tools) ingasethwa ngokuguquguquka kwemvelo `WV_TOOL_FILTER=passthrough`.

### v0.1.20 (2026-03-28)
- **Ukuqinisa okungqalile kokuphepha**: Ukuvimbela i-XSS (izindawo ezingu-41), ukuqhathanisa ama-token ngesikhathi esilinganayo, izihibe ze-CORS, imikhawulo yobukhulu besicelo, ukuvimbela ukudlula komzila, ukuqinisekisa kwe-SSE, ukuqinisa umkhawulo wejubane, nezinto ezingu-12 zokuphepha ezithuthukisiwe.

### v0.1.19 (2026-03-27)
- **Ukuthola i-Claude Code ku-inthanethi**: I-Claude Code engadluli nge-proxy nayo ivela njengoku-inthanethi ku-dashboard.

### v0.1.18 (2026-03-26)
- **Ukulungiswa kokunamathela kwesevisi ye-fallback**: Ngemva kwe-fallback yesikhashana ku-Ollama, uma isevisi yasekuqaleni ibuyela, ibuyela ngokuzenzakalela.
- **Ukuthuthukiswa kokuthola ukungabi ku-inthanethi**: Ukuthola i-proxy emile kusheshisiwe ngokuhlola isimo njalo ngamasekhondi angu-15.

### v0.1.17 (2026-03-25)
- **Ukuhlelwa kwamakhadi ngokudonsa nokudedela**: Amakhadi ommeleli angadonswa ukuze kushintshe uhlelo.
- **Iqhosha lokusebenzisa ukusetha emugqeni**: Iqhosha le-[⚡ Sebenzisa Ukusetha] livela kwabameleli abangekho ku-inthanethi.
- **Uhlobo lommeleli we-cokacdir lungezwe**.

### v0.1.16 (2026-03-25)
- **Ukuvumelana kwemodeli ngezinhlangothi ezimbili**: Ukushintsha imodeli ye-Cline·Claude Code ku-dashboard ye-vault kuvela ngokuzenzakalela.

---

*Ngolwazi olwengeziwe lwe-API, bheka [API.md](API.md).*
