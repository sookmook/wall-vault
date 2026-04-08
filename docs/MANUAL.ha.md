# Jagoran Amfani da wall-vault
*(Last updated: 2026-04-08 — v0.1.25)*

---

## Jerin Abubuwa

1. [Menene wall-vault?](#menene-wall-vault)
2. [Shigarwa](#shigarwa)
3. [Farkon Farawa (Magatakarda na setup)](#farkon-farawa)
4. [Yin Rajista na Maɓallin API](#yin-rajista-na-maballin-api)
5. [Yadda Ake Amfani da Proxy](#yadda-ake-amfani-da-proxy)
6. [Dashboard na Ɓaure](#dashboard-na-baure)
7. [Yanayin Rarraba (Multi-Bot)](#yanayin-rarraba-multi-bot)
8. [Saita Farawa ta Atomatik](#saita-farawa-ta-atomatik)
9. [Doctor — Kayan Bincike](#doctor-kayan-bincike)
10. [RTK Tanadin Token](#rtk-tanadin-token)
11. [Bayani kan Masu-Canji na Yanayi](#bayani-kan-masu-canji-na-yanayi)
12. [Warware Matsaloli](#warware-matsaloli)

---

## Menene wall-vault?

**wall-vault = Wakili na AI (Proxy) + Ɓauren Maɓallin API don OpenClaw**

Don amfani da sabis na AI, kana buƙatar **maɓallin API**. Maɓallin API kamar **katin shiga na dijital** ne wanda ke tabbatar da "wannan mutumin yana da ikon amfani da wannan sabis ɗin". Sai dai, wannan katin shiga yana iyakar amfani a kowace rana, kuma idan ba a kula shi da hankali ba, yana iya bayyana.

wall-vault na adana waɗannan katin shiga a cikin ɓaure mai tsaro kuma yana aiki a matsayin **wakili (proxy)** tsakanin OpenClaw da sabis na AI. A taƙaice, OpenClaw kawai yana buƙatar haɗawa da wall-vault, sauran abubuwan da suka fi rikitarwa wall-vault zai magance su.

Matsaloli da wall-vault ke warwarewa:

- **Sauya Maɓallin API ta Atomatik**: Idan maɓalli ɗaya ya kai iyaka ko aka dakatar da shi na ɗan lokaci (cooldown), za a canza shi a hankali zuwa maɓalli na gaba. OpenClaw na ci gaba da aiki ba tare da katsewar ba.
- **Sauya Sabis ta Atomatik (Fallback)**: Idan Google ba ya amsa ba, za a canza zuwa OpenRouter, idan hakan ma bai yi ba, za a canza zuwa Ollama/LM Studio/vLLM (AI na gida) da aka shigar a kwamfutarka. Zaman ba ya katse ba. Idan sabis na asali ya dawo, buƙatun da ke tafe za su koma ta atomatik (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Daidaita Ainihin Lokaci (SSE)**: Idan ka canza ƙirar a dashboard na ɓaure, zai bayyana a fuskar OpenClaw cikin daƙiƙa 1-3. SSE (Server-Sent Events) fasaha ce inda sabar ke tura canje-canje zuwa abokan ciniki a ainihin lokaci.
- **Sanarwa ta Ainihin Lokaci**: Abubuwan da suka faru kamar maɓalli sun ƙare ko sabis ya gaza ana nuna su nan da nan a ƙasan TUI (fuskar terminal) na OpenClaw.

> 💡 **Claude Code, Cursor, VS Code** ma ana iya haɗa su, amma manufar asali ta wall-vault ita ce a yi amfani da ita tare da OpenClaw.

```
OpenClaw (TUI fuskar terminal)
        │
        ▼
  wall-vault proxy (:56244)   ← sarrafa maɓalli, jagoranci, fallback, abubuwa
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (ƙira 340+)
        ├─ Ollama / LM Studio / vLLM (kwamfutarka, mafita na ƙarshe)
        └─ OpenAI / Anthropic API
```

---

## Shigarwa

### Linux / macOS

Buɗe terminal kuma ka manna umarni a ƙasa.

```bash
# Linux (PC na yau da kullum, sabar — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Yana sauke fayil daga yanar gizo.
- `chmod +x` — Yana sa fayil ɗin da aka sauke ya zama "ana iya gudanarwa". Idan ka tsallake wannan matakin, za ka samu kuskuren "ba a ba da izini ba".

### Windows

Buɗe PowerShell (a matsayin mai gudanarwa) kuma ka gudanar da umarni masu zuwa.

```powershell
# Sauke
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Ƙara zuwa PATH (yana aiki bayan sake farawa PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **Menene PATH?** Jerin manyan fayil ne inda kwamfuta ke neman umarni. Kana buƙatar ƙara zuwa PATH don ka iya rubuta `wall-vault` kuma ka gudanar da shi daga kowace babban fayil.

### Gina daga Tushen Lambar (don masu haɓakawa)

Wannan yana aiki ne kawai idan kana da yanayin haɓakawa na harshen Go da aka shigar.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (sigar: v0.1.25.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Sigar tambarin lokaci na gina**: Idan ka gina da `make build`, sigar za ta samar ta atomatik a tsarin da ya haɗa kwanan wata da lokaci kamar `v0.1.25.20260408.022325`. Idan ka gina kai tsaye da `go build ./...`, sigar za ta bayyana a matsayin `"dev"` kawai.

---

## Farkon Farawa

### Gudanar da Magatakarda na Setup

Bayan shigarwa, tabbatar ka gudanar da **magatakarda na saiti** da umarni mai zuwa. Magatakarda za ta tambaye ka tambayoyi ɗaya bayan ɗaya kuma ta jagorance ka.

```bash
wall-vault setup
```

Matakan da magatakarda ke bi su ne:

```
1. Zaɓi harshe (harsuna 10 haɗe da Hausa)
2. Zaɓi jigo (light / dark / gold / cherry / ocean)
3. Yanayin aiki — zaɓi ko za ka yi amfani da shi ka kaɗai (standalone) ko a na'urori da yawa (distributed)
4. Shigar da sunan bot — sunan da za a nuna a dashboard
5. Saita tashar — tsohuwar ƙima: proxy 56244, ɓaure 56243 (danna Enter idan ba ka buƙatar canzawa ba)
6. Zaɓi sabis na AI — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Saita tacen kayan aikin tsaro
8. Saita alamar mai gudanarwa — kalmar sirri don kulle fasalolin gudanarwa na dashboard. Ana iya samar da ita ta atomatik
9. Saita kalmar sirrin ɓoye maɓallin API — idan kana son adana maɓalli cikin aminci (zaɓi)
10. Hanyar adana fayil na saiti
```

> ⚠️ **Tabbatar ka tuna alamar mai gudanarwa.** Za ka buƙace ta daga baya lokacin da kake ƙara maɓalli ko canza saiti a dashboard. Idan ka rasa ta, za ka buƙaci gyara fayil na saiti kai tsaye.

Bayan magatakarda ta ƙare, fayil na saiti `wall-vault.yaml` za ta samar ta atomatik.

### Gudanarwa

```bash
wall-vault start
```

Sabar biyu za su fara a lokaci ɗaya:

- **Proxy** (`http://localhost:56244`) — wakili da ke haɗa OpenClaw da sabis na AI
- **Ɓauren Maɓalli** (`http://localhost:56243`) — sarrafa maɓallin API da dashboard na gidan yanar gizo

Buɗe `http://localhost:56243` a cikin mai binciken ka don ganin dashboard nan da nan.

---

## Yin Rajista na Maɓallin API

Akwai hanyoyi huɗu don yin rajista na maɓallin API. **Ga masu farawa, ana ba da shawarar Hanya ta 1 (masu-canji na yanayi)**.

### Hanya ta 1: Masu-Canji na Yanayi (Ana Ba da Shawarar — Mafi Sauƙi)

Masu-canji na yanayi sune **ƙimomi da aka riga aka saita** waɗanda shirin ke karantawa lokacin da ya fara. Rubuta a terminal kamar haka.

```bash
# Yin rajista da maɓallin Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Yin rajista da maɓallin OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Gudanar bayan yin rajista
wall-vault start
```

Idan kana da maɓalli da yawa, haɗa su da waƙafi (,). wall-vault za ta yi amfani da maɓalli a jeri ta atomatik (round robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Tukwici**: Umarni na `export` yana aiki ne kawai a zaman terminal na yanzu. Don ya ci gaba ko bayan sake farawa kwamfuta, ƙara layin a fayil na `~/.bashrc` ko `~/.zshrc`.

### Hanya ta 2: Dashboard na UI (Danna da linzami)

1. Ziyarci `http://localhost:56243` a cikin mai bincike
2. A katin **🔑 Maɓallin API** a sama, danna maɓallin `[+ Ƙara]`
3. Shigar da irin sabis, ƙimar maɓalli, alamar (sunan tunani), da iyakar yau da kullum, sa'an nan ka ajiye

### Hanya ta 3: REST API (don Atomatik/Rubutun)

REST API hanya ce inda shirye-shirye ke musayar bayanai ta HTTP. Yana da amfani don yin rajista ta atomatik ta hanyar rubutu.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer alamar-mai-gudanarwa" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Maɓallin Babba",
    "daily_limit": 1000
  }'
```

### Hanya ta 4: Tutar Proxy (don Gwajin na Ɗan Lokaci)

Ana amfani da shi lokacin da kake so ka gwada na ɗan lokaci ba tare da yin rajista na hukuma ba. Yana ɓacewa lokacin da ka kashe shirin.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Yadda Ake Amfani da Proxy

### Amfani da OpenClaw (Babban Manufa)

Yadda ake saita OpenClaw don haɗawa da sabis na AI ta wall-vault.

Buɗe fayil na `~/.openclaw/openclaw.json` kuma ka ƙara abubuwan da ke ƙasa:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // alamar wakili na vault
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // context na 1M kyauta
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Hanya Mafi Sauƙi**: Danna maɓallin **🦞 Kwafi Saitin OpenClaw** a katin wakili na dashboard kuma snippet da alamar da adireshi da aka riga aka cika za a kwafa zuwa clipboard. Ka manna kawai.

**`wall-vault/` kafin sunan ƙirar yana nufin ina?**

wall-vault yana yanke shawara ta atomatik wace sabis na AI za a yi amfani da ita dangane da sunan ƙirar:

| Tsarin Ƙirar | Sabis da Ake Haɗawa |
|----------|--------------|
| `wall-vault/gemini-*` | Haɗin kai tsaye da Google Gemini |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | Haɗin kai tsaye da OpenAI |
| `wall-vault/claude-*` | Haɗawa da Anthropic ta OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (tokeni miliyan 1 na context kyauta) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | Haɗawa da OpenRouter |
| `google/sunan-ƙirar`, `openai/sunan-ƙirar`, `anthropic/sunan-ƙirar` da sauransu | Haɗin kai tsaye da sabis ɗin da ya dace |
| `custom/google/sunan-ƙirar`, `custom/openai/sunan-ƙirar` da sauransu | Sashen `custom/` ana cirewa kuma ana sake jagora |
| `sunan-ƙirar:cloud` | Sashen `:cloud` ana cirewa kuma ana haɗawa da OpenRouter |

> 💡 **Menene context?** Adadin tattaunawa da AI ke iya tunawa a lokaci ɗaya. 1M (tokeni miliyan ɗaya) yana nufin za a iya sarrafa tattaunawa mai tsawo ko takardu masu tsawo a lokaci ɗaya.

### Haɗin Kai Tsaye da Tsarin Gemini API (dacewa da kayan aikin da ake da su)

Idan kana da kayan aiki da suka kasance suna amfani da Google Gemini API kai tsaye, kawai canza adireshin zuwa wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Ko idan kayan aikinka suna bayyana URL kai tsaye:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Amfani da OpenAI SDK (Python)

Kuma za ka iya haɗa wall-vault a cikin lambar Python da ke amfani da AI. Kawai canza `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # wall-vault yana sarrafa maɓallin API ta atomatik
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # tsarin mai-bayarwa/ƙirar
    messages=[{"role": "user", "content": "Sannu"}]
)
```

### Canza Ƙirar Yayin Gudanarwa

Don canza ƙirar AI yayin da wall-vault tana aiki:

```bash
# Canza ƙirar ta hanyar nema kai tsaye ga proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# A yanayin rarraba (boti da yawa), canza a sabar ɓaure → ana nuna nan da nan ta SSE
curl -X PUT http://localhost:56243/admin/clients/id-na-bot \
  -H "Authorization: Bearer alamar-mai-gudanarwa" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Duba Jerin Ƙira da Ake Samu

```bash
# Duba jerin gaba ɗaya
curl http://localhost:56244/api/models | python3 -m json.tool

# Duba ƙirar Google kawai
curl "http://localhost:56244/api/models?service=google"

# Bincika da suna (misali: ƙira da suka haɗa da "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Taƙaitaccen Babban Ƙira ta Sabis:**

| Sabis | Babban Ƙira |
|--------|----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ (Hunter Alpha 1M context kyauta, DeepSeek R1/V3, Qwen 2.5 da sauransu) |
| Ollama | Yana gano sabar na gida da aka shigar a kwamfutarka ta atomatik |
| LM Studio | Sabar na gida a kwamfutarka (tashar 1234) |
| vLLM | Sabar na gida a kwamfutarka (tashar 8000) |

---

## Dashboard na Ɓaure

Ziyarci `http://localhost:56243` a cikin mai bincike don ganin dashboard.

**Tsarin Fuskar:**
- **Sandar sama mai ɗaurewa (topbar)**: Alamar kamfani, mai zaɓin harshe da jigo, yanayin haɗin SSE
- **Grid na kati**: Katin wakili, sabis, da maɓallin API da aka jera kamar tiles

### Katin Maɓallin API

Kati da ke ba ka damar sarrafa maɓallin API da aka yi rajista a duba ɗaya.

- Yana nuna jerin maɓalli da aka raba ta sabis.
- `today_usage`: Adadin tokeni da aka sarrafa cikin nasara yau (adadin haruffa da AI ta karanta da rubuta)
- `today_attempts`: Jimlar kiran yau (nasara + gazawa)
- Maɓallin `[+ Ƙara]` don yin rajista da sabon maɓalli, da `✕` don share maɓalli.

> 💡 **Menene tokeni?** Ƙayyadaddun sashe ne da AI ke amfani da shi wajen sarrafa rubutu. Kusan daidai yake da kalma ɗaya na Turanci, ko haruffa 1-2 na Koriya. Kuɗin API yawanci ana kididdige su dangane da adadin tokeni.

### Katin Wakili

Kati da ke nuna yanayin bots (masu wakilci) da aka haɗa da proxy na wall-vault.

**Ana nuna yanayin haɗin a matakai 4:**

| Nuni | Yanayi | Ma'ana |
|------|------|------|
| 🟢 | Yana Aiki | Proxy yana aiki yadda ya kamata |
| 🟡 | An Jinkirta | Yana amsawa amma a hankali |
| 🔴 | Ba a Kan Layi Ba | Proxy ba ya amsawa |
| ⚫ | Ba a Haɗa/An Kashe | Proxy bai taɓa haɗawa da ɓaure ba ko an kashe shi |

**Jagorar Maɓallan da ke Ƙasan Katin Wakili:**

Lokacin da ka yi rajista da wakili kuma ka bayyana **irin wakili**, maɓallan sauƙi da aka yi niyya don wannan irin suna bayyana ta atomatik.

---

#### 🔘 Maɓallin Kwafin Saiti — Yana samar da saitin haɗin ta atomatik

Lokacin da ka danna maɓallin, snippet na saiti da ke ɗauke da alamar wannan wakili, adireshin proxy, da bayanan ƙirar da aka riga aka cika ana kwafa zuwa clipboard. Kawai ka manna abubuwan da aka kwafa a wurin da aka nuna a tebur na ƙasa don kammala saitin haɗin.

| Maɓalli | Irin Wakili | Wurin Mannawa |
|------|-------------|-------------|
| 🦞 Kwafi Saitin OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Kwafi Saitin NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Kwafi Saitin Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Kwafi Saitin Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Kwafi Saitin VSCode | `vscode` | `~/.continue/config.json` |

**Misali — Idan irin Claude Code ne, abubuwan da ke ƙasa za a kwafa su:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "alamar-wannan-wakili"
}
```

**Misali — Idan irin VSCode (Continue) ne:**

```yaml
# ~/.continue/config.yaml  ← manna a config.yaml, ba config.json ba
name: My Config
version: 0.0.1
schema: v1

models:
  - name: wall-vault proxy
    provider: openai
    model: gemini-2.5-flash
    apiBase: http://192.168.1.20:56244/v1
    apiKey: alamar-wannan-wakili
    roles:
      - chat
      - edit
      - apply
```

> ⚠️ **Sabon sigar Continue yana amfani da `config.yaml`.** Idan `config.yaml` ya wanzu, `config.json` za a yi watsi da shi gaba ɗaya. Tabbatar ka manna a cikin `config.yaml`.

**Misali — Idan irin Cursor ne:**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : alamar-wannan-wakili

// Ko masu-canji na yanayi:
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=alamar-wannan-wakili
```

> ⚠️ **Idan kwafin clipboard bai yi aiki ba**: Manufar tsaro na mai bincike na iya hana kwafi. Idan akwatin rubutu ya buɗe a matsayin popup, yi amfani da Ctrl+A don zaɓar duka sa'an nan Ctrl+C don kwafi.

---

#### ⚡ Maɓallin Aiwatar ta Atomatik — Danna sau ɗaya kuma saitin ya ƙare

Idan irin wakili shine `cline`, `claude-code`, `openclaw`, ko `nanoclaw`, maɓallin **⚡ Aiwatar da Saiti** yana bayyana a katin wakili. Lokacin da ka danna wannan maɓallin, fayil na saitin na gida na wannan wakili ana sabunta shi ta atomatik.

| Maɓalli | Irin Wakili | Fayil da Ake Niyya |
|------|-------------|-------------|
| ⚡ Aiwatar da Saitin Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Aiwatar da Saitin Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Aiwatar da Saitin OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Aiwatar da Saitin NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Wannan maɓallin yana aika buƙata zuwa **localhost:56244** (proxy na gida). Dole ne proxy ya kasance tana aiki a wannan na'urar don ya yi aiki.

---

#### 🔀 Tsara Kati ta Ja da Sauke (v0.1.17, an inganta v0.1.25)

Za ka iya **jan** katin wakili a dashboard don sake tsara su a tsarin da kake so.

1. Kama yankin **hasken zirga-zirga (●)** a saman hagu na kati da linzami kuma ka ja
2. Sauke shi a kan katin da kake so kuma tsarin zai canja

> 💡 Abubuwan cikin kati (wuraren shigarwa, maɓallai da sauransu) ba a iya jan su ba. Za ka iya kamawa ne kawai daga yankin hasken zirga-zirga.

#### 🟠 Ganowa Tsarin Wakili (v0.1.25)

Lokacin da proxy ke aiki yadda ya kamata amma tsarin wakili na gida (NanoClaw, OpenClaw) ya mutu, hasken katin ya canja zuwa **launin lemu (yana kyaftawa)** kuma saƙon "Tsarin wakili ya tsaya" yana bayyana.

- 🟢 Kore: Proxy + wakili suna aiki yadda ya kamata
- 🟠 Launin lemu (yana kyaftawa): Proxy yana aiki yadda ya kamata, wakili ya mutu
- 🔴 Ja: Proxy ba a kan layi ba
3. Tsarin da aka canza **ana adana shi a sabar nan da nan** kuma yana ci gaba ko bayan sake lodi

> 💡 A na'urorin tabawa (wayar hannu/kwamfutar hannu) har yanzu ba a goyon baya ba. Yi amfani da mai binciken kwamfuta.

---

#### 🔄 Daidaita Ƙirar ta Bangarori Biyu (v0.1.16)

Idan ka canza ƙirar wakili a dashboard na ɓaure, saitin na gida na wannan wakili ana sabunta shi ta atomatik.

**Don Cline:**
- Idan aka canza ƙirar a ɓaure → Abin da ya faru na SSE → Proxy yana sabunta sashen ƙirar a `globalState.json`
- Maƙasudi na sabuntawa: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` da maɓallin API ba a taɓa su ba
- **Sake lodin VS Code (`Ctrl+Alt+R` ko `Ctrl+Shift+P` → `Developer: Reload Window`) yana da buƙata**
  - Saboda Cline ba ta sake karanta fayil na saiti yayin aiki ba

**Don Claude Code:**
- Idan aka canza ƙirar a ɓaure → Abin da ya faru na SSE → Proxy yana sabunta sashen `model` a `settings.json`
- Yana bincika hanyoyin WSL da Windows ta atomatik (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Bangaren Baya (wakili → ɓaure):**
- Lokacin da wakili (Cline, Claude Code da sauransu) ya aika buƙata zuwa proxy, proxy yana haɗa bayanan sabis/ƙirar abokin ciniki a heartbeat
- Sabis/ƙirar da ake amfani da ita a yanzu ana nuna ta a ainihin lokaci a katin wakili a dashboard na ɓaure

> 💡 **Mahimmin Abu**: Proxy yana gane wakili ta alamar Authorization na buƙatar, kuma yana jagoranci ta atomatik zuwa sabis/ƙirar da aka saita a ɓaure. Ko Cline ko Claude Code ta aika sunan ƙirar daban, proxy yana maye gurbinsa da saitin ɓaure.

---

### Amfani da Cline a VS Code — Cikakken Jagora

#### Mataki na 1: Shigar da Cline

Shigar da **Cline** (ID: `saoudrizwan.claude-dev`) daga kasuwar ƙari na VS Code.

#### Mataki na 2: Yi Rajista da Wakili a Ɓaure

1. Buɗe dashboard na ɓaure (`http://IP-na-ɓaure:56243`)
2. A sashen **Masu Wakilci**, danna **+ Ƙara**
3. Shigar da abubuwan da ke ƙasa:

| Filin | Ƙima | Bayani |
|------|----|------|
| ID | `cline_na` | Musamman mai ganowa (Turanci, ba tazara) |
| Suna | `Cline Nawa` | Sunan da za a nuna a dashboard |
| Irin Wakili | `cline` | ← Dole ne ka zaɓi `cline` |
| Sabis | Zaɓi sabis don amfani (misali: `google`) | |
| Ƙirar | Shigar da ƙirar don amfani (misali: `gemini-2.5-flash`) | |

4. Danna **Ajiye** kuma za a samar da alama ta atomatik

#### Mataki na 3: Haɗa da Cline

**Hanya A — Aiwatar ta Atomatik (Ana Ba da Shawarar)**

1. Tabbatar **proxy** na wall-vault yana aiki a wannan na'urar (`localhost:56244`)
2. Danna maɓallin **⚡ Aiwatar da Saitin Cline** a katin wakili a dashboard
3. Idan sanarwar "An aiwatar da saiti!" ta bayyana, ya yi nasara
4. Sake lodin VS Code (`Ctrl+Alt+R`)

**Hanya B — Saita da Hannu**

Buɗe saituna (⚙️) a gefen Cline kuma ka saita:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://adireshin-proxy:56244/v1`
  - A wannan na'urar: `http://localhost:56244/v1`
  - A wata na'urar kamar sabar Mini: `http://192.168.1.20:56244/v1`
- **API Key**: Alamar da aka bayar daga ɓaure (kwafa daga katin wakili)
- **Model ID**: Ƙirar da aka saita a ɓaure (misali: `gemini-2.5-flash`)

#### Mataki na 4: Tabbatarwa

Aika kowace saƙo a tattaunawar Cline. Idan komai ya yi daidai:
- **Alamar kore (● Yana Aiki)** tana bayyana a katin wakili a dashboard na ɓaure
- Sabis/ƙirar na yanzu ana nuna ta a katin (misali: `google / gemini-2.5-flash`)

#### Canza Ƙirar

Lokacin da kake so ka canza ƙirar Cline, canza ta a **dashboard na ɓaure**:

1. Canza zaɓin sabis/ƙirar a katin wakili
2. Danna **Aiwatar**
3. Sake lodin VS Code (`Ctrl+Alt+R`) — sunan ƙirar a ƙasan Cline zai sabunta
4. Za a yi amfani da sabuwar ƙirar tun daga buƙatar da ke tafe

> 💡 A zahiri, proxy yana gane buƙatar Cline ta alama kuma yana jagoran zuwa ƙirar saitin ɓaure. Ko ba tare da sake lodin VS Code ba, **ƙirar da ake amfani da ita a zahiri tana canzawa nan da nan** — sake lodin don sabunta nunin ƙirar a UI na Cline ne kawai.

#### Ganowa Katsewar Haɗin

Lokacin da ka rufe VS Code, katin wakili a dashboard na ɓaure yana canzawa zuwa rawaya (an jinkirta) bayan kusan **dakika 90**, kuma zuwa ja (ba a kan layi ba) bayan **minti 3**. (Tun daga v0.1.18, binciken yanayi a kowane dakika 15 yana sa ganowa ba a kan layi ba ta fi sauri.)

#### Warware Matsaloli

| Alamar | Dalili | Magani |
|------|------|------|
| Kuskuren "haɗin ya gaza" a Cline | Proxy ba ya aiki ko adireshin ba daidai ba ne | Tabbatar da proxy da `curl http://localhost:56244/health` |
| Alamar kore ba ta bayyana a ɓaure ba | Maɓallin API (alama) ba a saita ba | Sake danna maɓallin **⚡ Aiwatar da Saitin Cline** |
| Ƙirar a ƙasan Cline ba ta canjawa ba | Cline yana adana saiti | Sake lodin VS Code (`Ctrl+Alt+R`) |
| Ana nuna sunan ƙirar da ba daidai ba | Tsohon kuskure (an gyara a v0.1.16) | Sabunta proxy zuwa v0.1.16 ko sama |

---

#### 🟣 Maɓallin Kwafin Umarni na Tura — Ana amfani da shi lokacin shigarwa a sabuwar na'ura

Ana amfani da shi lokacin da ake shigar da proxy na wall-vault a karon farko a sabuwar kwamfuta kuma ana haɗa ta da ɓaure. Lokacin da ka danna maɓallin, dukkan rubutun shigarwa ana kwafa shi. Manna shi a terminal na sabuwar kwamfuta kuma ka gudanar da shi, kuma abubuwan da ke ƙasa za a magance su a lokaci ɗaya:

1. Shigar da binary na wall-vault (ana tsallake idan an riga an shigar)
2. Yi rajista da sabis na masu amfani na systemd ta atomatik
3. Fara sabis kuma haɗa da ɓaure ta atomatik

> 💡 Rubutun ya riga ya ƙunshi alamar wannan wakili da adireshin sabar ɓaure, don haka za ka iya gudanar da shi kai tsaye bayan mannawa ba tare da wani gyara ba.

---

### Katin Sabis

Kati don kunna/kashe ko saita sabis na AI don amfani.

- Maɓallan sauya kunna/kashe ga kowace sabis
- Idan ka shigar da adireshin sabar AI na gida (Ollama, LM Studio, vLLM da sauransu da ke gudana a kwamfutarka), ƙirar da ake samu za a gano su ta atomatik.
- **Nunin yanayin haɗin sabis na gida**: Alamar ● kusa da sunan sabis shine **kore** idan an haɗa, **toka** idan ba a haɗa ba
- **Hasken zirga-zirga na atomatik na sabis na gida** (v0.1.23+): Sabis na gida (Ollama, LM Studio, vLLM) ana kunna su ta atomatik lokacin da za a iya haɗawa, kuma ana kashe su ta atomatik lokacin da suka katse. Lokacin da ka kunna sabis, yana canzawa zuwa ● kore cikin daƙiƙa 15 kuma akwatin dubawa yana kunna, kuma idan ka kashe shi, ana kashe shi ta atomatik. Wannan yana aiki daidai da yadda sabis na girgije (Google, OpenRouter da sauransu) ke sauya atomatik dangane da kasancewar maɓallin API.

> 💡 **Idan sabis na gida yana gudana a wata kwamfuta**: Shigar da IP na wannan kwamfutar a wurin shigar URL na sabis. Misali: `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio). Idan sabis ɗin ya ɗaure zuwa `127.0.0.1` kawai maimakon `0.0.0.0`, ba za a iya samun dama ta IP na waje ba, don haka bincika adireshin ɗaurin a saitin sabis.

### Shigar da Alamar Mai Gudanarwa

Lokacin da kake ƙoƙarin amfani da fasaloli masu muhimmanci kamar ƙara ko share maɓalli a dashboard, popup na shigar da alamar mai gudanarwa zai bayyana. Shigar da alamar da ka saita a magatakarda na setup. Da zarar ka shigar da ita, tana ci gaba har sai ka rufe mai bincike.

> ⚠️ **Idan tabbatarwa ta gaza fiye da sau 10 cikin minti 15, za a dakatar da wannan IP na ɗan lokaci.** Idan ka manta da alamarka, duba kashi na `admin_token` a fayil na `wall-vault.yaml`.

---

## Yanayin Rarraba (Multi-Bot)

Tsari na **raba ɓaure ɗaya na maɓalli** lokacin gudanar da OpenClaw a kwamfutoci da yawa a lokaci ɗaya. Yana da sauƙi saboda kana buƙatar sarrafa maɓalli a wuri ɗaya kawai.

### Misalin Tsari

```
[Sabar Ɓauren Maɓalli]
  wall-vault vault    (ɓauren maɓalli :56243, dashboard)

[WSL Alpha]            [Raspberry Pi Gamma]    [Mac Mini na Gida]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ Daidaita SSE        ↕ Daidaita SSE          ↕ Daidaita SSE
```

Dukkan bots suna duban sabar ɓaure a tsakiya, don haka idan ka canza ƙirar ko ƙara maɓalli a ɓaure, ana nuna shi ga dukkan bots nan da nan.

### Mataki na 1: Fara Sabar Ɓauren Maɓalli

Gudanar a kwamfutar da za ka yi amfani da ita a matsayin sabar ɓaure:

```bash
wall-vault vault
```

### Mataki na 2: Yi Rajista da Kowace Bot (Abokin Ciniki)

Yi rajista da bayanan kowace bot da ke haɗawa da sabar ɓaure tun farko:

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer alamar-mai-gudanarwa" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "BotA",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### Mataki na 3: Fara Proxy a Kowace Kwamfutar Bot

Gudanar da proxy a kowace kwamfutar da bot ke ciki, ka bayyana adireshin sabar ɓaure da alama:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Canza sashen **`192.168.x.x`** da ainihin adireshin IP na ciki na kwamfutar sabar ɓaure. Za ka iya tabbatar da shi ta saitin router ko umarni na `ip addr`.

---

## Saita Farawa ta Atomatik

Idan yana da wuya a kunna wall-vault da hannu kowace lokaci da ka sake farawa kwamfuta, yi rajista da shi a matsayin sabis na tsarin. Da zarar an yi rajista, zai fara ta atomatik lokacin da ake lodawa.

### Linux — systemd (yawancin Linux)

systemd tsarin ne da ke kunna da sarrafa shirye-shirye ta atomatik a Linux:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Duba bayanan aiki:

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

Tsarin da ke sarrafa fara shirye-shirye ta atomatik a macOS:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Sauke NSSM daga [nssm.cc](https://nssm.cc/download) kuma ka ƙara shi zuwa PATH.
2. A PowerShell na mai gudanarwa:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor — Kayan Bincike

Umarni na `doctor` kayan aiki ne da ke **bincika kansa da gyara kansa** don tabbatar da cewa wall-vault an saita shi yadda ya kamata.

```bash
wall-vault doctor check   # Bincika yanayin yanzu (karantawa kawai, ba ya canza komai)
wall-vault doctor fix     # Gyara matsaloli ta atomatik
wall-vault doctor all     # Bincike + gyara ta atomatik a lokaci ɗaya
```

> 💡 Idan wani abu ya yi kamar ba daidai ba, gudanar da `wall-vault doctor all` da farko. Yana magance matsaloli da yawa ta atomatik.

---

## RTK Tanadin Token

*(v0.1.24+)*

**RTK (Kayan Tanadin Token)** yana matse sakamakon umarni na shell da masu wakilci na AI na lambar (kamar Claude Code) ke gudanarwa ta atomatik don rage amfani da tokeni. Misali, sakamakon layi 15 na `git status` yana raguwa zuwa taƙaitaccen layi 2.

### Amfani na Asali

```bash
# Nannaɗe umarni da wall-vault rtk kuma za a tace sakamako ta atomatik
wall-vault rtk git status          # Jerin fayiloli da aka canza kawai
wall-vault rtk git diff HEAD~1     # Layukan da aka canza + ƙaramin mahallin
wall-vault rtk git log -10         # Hash + saƙon layi ɗaya
wall-vault rtk go test ./...       # Gwaje-gwaje da suka gaza kawai
wall-vault rtk ls -la              # Umarni marasa goyon baya ana yanke su ta atomatik
```

### Umarni da Ake Goyon Baya da Tasirin Ragewa

| Umarni | Hanyar Tacewa | Adadin Ragewa |
|------|----------|--------|
| `git status` | Taƙaitaccen fayiloli da aka canza kawai | ~87% |
| `git diff` | Layukan da aka canza + mahallin layi 3 | ~60-94% |
| `git log` | Hash + saƙon layi na farko | ~90% |
| `git push/pull/fetch` | Cire ci gaba, taƙaitacce kawai | ~80% |
| `go test` | Nuna gazawa kawai, ƙidaya nasara | ~88-99% |
| `go build/vet` | Nuna kurakurai kawai | ~90% |
| Duk sauran umarni | Layi 50 na farko + 50 na ƙarshe, matsakaicin 32KB | Ya danganta |

### Tsarin Tacewa na Mataki 3

1. **Tacen tsari ga kowace umarni** — Yana fahimtar tsarin sakamakon git, go da sauransu kuma yana ciro sassan da suka dace kawai
2. **Sarrafa bayan regex** — Cire lambobin launi na ANSI, rage layukan kofe, taƙaita layukan da suka maimaita
3. **Passthrough + yanke** — Umarni marasa goyon baya suna riƙe da layi 50 na farko/ƙarshe kawai

### Haɗawa da Claude Code

Za ka iya saita dukkan umarni na shell su bi ta RTK ta atomatik ta hook na `PreToolUse` na Claude Code.

```bash
# Shigar da hook (ana ƙarawa ta atomatik zuwa Claude Code settings.json)
wall-vault rtk hook install
```

Ko ƙara da hannu zuwa `~/.claude/settings.json`:

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

> 💡 **Kiyaye Exit code**: RTK yana mayar da lambar fitarwa na ainihin umarni kamar yadda yake. Idan umarni ya gaza (exit code ≠ 0), AI ma yana gano gazawar daidai.

> 💡 **Tilasta Turanci**: RTK yana gudanar da umarni da `LC_ALL=C` don samar da sakamakon Turanci koyaushe ba tare da la'akari da saitin harshen tsarin ba. Wannan yana tabbatar da cewa tacen yana aiki daidai.

---

## Bayani kan Masu-Canji na Yanayi

Masu-canji na yanayi hanya ce ta isar da ƙimomi na saiti zuwa shirin. Shigar a tsarin `export sunan-mai-canji=ƙima` a terminal, ko saka a fayil na sabis na farawa ta atomatik don ya kasance yana aiki koyaushe.

| Mai-Canji | Bayani | Misalin Ƙima |
|------|------|---------|
| `WV_LANG` | Harshen dashboard | `ko`, `en`, `ja` |
| `WV_THEME` | Jigon dashboard | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Maɓallin Google API (da yawa da waƙafi) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Maɓallin OpenRouter API | `sk-or-v1-...` |
| `WV_VAULT_URL` | Adireshin sabar ɓaure a yanayin rarraba | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Alamar tabbatarwa na abokin ciniki (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Alamar mai gudanarwa | `admin-token-here` |
| `WV_MASTER_PASS` | Kalmar sirrin ɓoye maɓallin API | `my-password` |
| `WV_AVATAR` | Hanyar fayil na hoton avatar (dangantaka daga `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Adireshin sabar Ollama na gida | `http://192.168.x.x:11434` |

---

## Warware Matsaloli

### Lokacin da Proxy Ba Ya Farawa

Sau da yawa, tashar tana amfani da wani shirin.

```bash
ss -tlnp | grep 56244   # Duba wanene ke amfani da tashar 56244
wall-vault proxy --port 8080   # Fara da wata lambar tashar
```

### Lokacin da Kuskuren Maɓallin API Ya Faru (429, 402, 401, 403, 582)

| Lambar Kuskure | Ma'ana | Yadda Ake Magancewa |
|----------|------|----------|
| **429** | Buƙatu da yawa (matumizi sun wuce ƙima) | Jira kaɗan ko ƙara wani maɓalli |
| **402** | Ana buƙatar biyan kuɗi ko kuɗin ba su isa ba | Ƙara kuɗi a sabis ɗin da ya dace |
| **401 / 403** | Maɓalli ba daidai ba ne ko babu izini | Sake tabbatar da ƙimar maɓalli kuma ka sake yin rajista |
| **582** | Nauyin gateway (cooldown minti 5) | Ana sakewa ta atomatik bayan minti 5 |

```bash
# Duba jerin da yanayin maɓalli da aka yi rajista
curl -H "Authorization: Bearer alamar-mai-gudanarwa" http://localhost:56243/admin/keys

# Sake saita ƙididdigan amfani da maɓalli
curl -X POST -H "Authorization: Bearer alamar-mai-gudanarwa" http://localhost:56243/admin/keys/reset
```

### Lokacin da Wakili Ya Nuna "Ba a Haɗa Ba"

"Ba a haɗa ba" yana nufin tsarin proxy ba ya aika sigina (heartbeat) zuwa ɓaure. **Ba ya nufin cewa ba a adana saituna ba.** Dole ne proxy ya kasance yana aiki yana sanin adireshin sabar ɓaure da alama don yanayin haɗin ya canja.

```bash
# Fara proxy ka bayyana adireshin sabar ɓaure, alama, da ID na abokin ciniki
WV_VAULT_URL=http://adireshin-sabar-ɓaure:56243 \
WV_VAULT_TOKEN=alamar-abokin-ciniki \
WV_VAULT_CLIENT_ID=ID-na-abokin-ciniki \
wall-vault proxy
```

Idan haɗin ya yi nasara, zai canja zuwa 🟢 Yana Aiki a dashboard cikin kusan dakika 20.

### Lokacin da Ollama Ba Ya Haɗawa

Ollama shirin ne da ke gudanar da AI kai tsaye a kwamfutarka. Da farko tabbatar Ollama tana aiki.

```bash
curl http://localhost:11434/api/tags   # Idan jerin ƙira sun bayyana, al'ada ne
export OLLAMA_URL=http://192.168.x.x:11434   # Idan tana gudana a wata kwamfuta
```

> ⚠️ Idan Ollama ba ta amsa ba, fara Ollama da farko da umarni na `ollama serve`.

> ⚠️ **Manyan ƙira suna da jinkiri**: Manyan ƙira kamar `qwen3.5:35b`, `deepseek-r1` na iya ɗaukar mintuna da yawa don samar da amsa. Ko ya yi kamar babu amsa, yana iya kasancewa ana sarrafa al'ada ne, don haka ka jira.

---

## Sabbin Canje-canje (v0.1.16 ~ v0.1.25)

### v0.1.25 (2026-04-08)
- **Ganowa Tsarin Wakili**: Proxy yana gano yanayin rayuwa na wakili na gida (NanoClaw/OpenClaw) kuma yana nuna shi da hasken launin lemu a dashboard.
- **Inganta Hannun Ja**: An canza don kati su kama ne kawai daga yankin hasken zirga-zirga (●) lokacin tsarawa. Wuraren shigarwa ko maɓallai ba a jan su da kuskure ba.

### v0.1.24 (2026-04-06)
- **Ƙaramin Umarni na RTK na Tanadin Token**: `wall-vault rtk <command>` yana tace sakamakon umarni na shell ta atomatik don rage amfani da tokeni na masu wakilci na AI da 60-90%. Ya haɗa da tacen musamman don manyan umarni kamar git, go, kuma umarni marasa goyon baya ma ana yanke su ta atomatik. Yana haɗawa cikin ganuwa ta hook na `PreToolUse` na Claude Code.

### v0.1.23 (2026-04-06)
- **Gyaran Canza Ƙirar Ollama**: Matsalar inda canza ƙirar Ollama a dashboard na ɓaure ba ta bayyana a proxy an gyara. A da can yana amfani da mai-canji na yanayi (`OLLAMA_MODEL`) kawai, amma yanzu ana ba da fifiko ga saitin ɓaure.
- **Hasken Zirga-zirga na Atomatik na Sabis na Gida**: Ollama, LM Studio, vLLM ana kunna su ta atomatik lokacin da za a iya haɗawa, kuma ana kashe su ta atomatik lokacin da suka katse. Yana aiki daidai da sauya atomatik da ke dogara ga maɓalli na sabis na girgije.

### v0.1.22 (2026-04-05)
- **Gyaran Rashin Filin content mai Fanko**: Lokacin da ƙirar tunani (gemini-3.1-pro, o1, claude thinking da sauransu) suka yi amfani da iyakar max_tokens gaba ɗaya don tunani kuma suka gaza samar da ainihin amsa, proxy ya kasance yana barin filayen `content`/`text` a JSON na amsa da `omitempty`, wanda ya sa abokan cinikin OpenAI/Anthropic SDK su samu kuskure na `Cannot read properties of undefined (reading 'trim')`. An canza don haɗa filaye koyaushe bisa ga ƙayyadaddun hukumar API.

### v0.1.21 (2026-04-05)
- **Goyon Bayan Ƙirar Gemma 4**: Ƙirar dangin Gemma kamar `gemma-4-31b-it`, `gemma-4-26b-a4b-it` za a iya amfani da su ta Google Gemini API.
- **Goyon Bayan Sabis na LM Studio / vLLM na Hukuma**: A da can an tsallake waɗannan sabis a jagorar proxy kuma koyaushe ana maye gurbinsu da Ollama. Yanzu ana jagoran su yadda ya kamata ta API mai dacewa da OpenAI.
- **Gyaran Nunin Sabis a Dashboard**: Ko fallback ya faru, dashboard koyaushe yana nuna sabis ɗin da mai amfani ya saita.
- **Nunin Yanayin Sabis na Gida**: Yanayin haɗin sabis na gida (Ollama, LM Studio, vLLM da sauransu) ana nuna shi ta launin alamar ● lokacin da dashboard ke lodawa.
- **Mai-Canji na Yanayi na Tacen Kayan Aiki**: Yanayin isar da kayan aiki za a iya saita shi da mai-canji na yanayi `WV_TOOL_FILTER=passthrough`.

### v0.1.20 (2026-03-28)
- **Ƙarfafa Tsaro na Gaba Ɗaya**: Inganta abubuwa 12 na tsaro haɗe da hana XSS (wurare 41), kwatanta alama na lokaci maras bambanci, ƙuntatawa CORS, iyakokin girman buƙata, hana bi ta hanya, tabbatarwa na SSE, ƙarfafa iyakar gudu da sauransu.

### v0.1.19 (2026-03-27)
- **Ganowa Claude Code a Kan Layi**: Claude Code da ba ya bi ta proxy ma ana nuna shi a matsayin a kan layi a dashboard.

### v0.1.18 (2026-03-26)
- **Gyaran Makale na Sabis na Fallback**: Bayan komawa Ollama saboda kuskure na ɗan lokaci, idan sabis na asali ya dawo, ana komawa ta atomatik.
- **Inganta Ganowa Ba a Kan Layi Ba**: Binciken yanayi a kowane dakika 15 yana sa ganowa proxy ya tsaya ta fi sauri.

### v0.1.17 (2026-03-25)
- **Tsara Kati ta Ja da Sauke**: Za a iya jan katin wakili don canza tsari.
- **Maɓallin Aiwatar da Saiti na Layi**: Maɓallin [⚡ Aiwatar da Saiti] yana bayyana a masu wakilci da ba su a kan layi ba.
- **An ƙara irin wakili na cokacdir**.

### v0.1.16 (2026-03-25)
- **Daidaita Ƙirar ta Bangarori Biyu**: Canza ƙirar Cline ko Claude Code a dashboard na ɓaure yana haifar da bayyanar ta atomatik.

---

*Don ƙarin cikakken bayani na API, duba [API.md](API.md).*
