# Jagoran Amfani da wall-vault
*(Ƙarshen sabuntawa: 2026-04-06 — v0.1.24)*

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

**wall-vault = Wakilin AI (Proxy) + Ɓaure na Maɓallin API — na OpenClaw**

Don amfani da ayyukan AI, kana buƙatar **maɓallin API**. Maɓallin API shi ne **takardar shiga ta dijital** wanda ke tabbatar cewa ana ba ka izinin amfani da ayyukan. Amma wannan takardar shiga tana da ƙayyadaddun amfani a kowace rana, kuma idan ba a lura da ita sosai ba, tana iya fita a hannun marasa izini.

wall-vault yana ajiye waɗannan maɓallan cikin **ɓaure mai aminci**, kuma yana aiki a matsayin **wakili (proxy)** tsakanin OpenClaw da ayyukan AI. A taƙaice, OpenClaw yana haɗawa da wall-vault kawai, sannan wall-vault yana kula da sauran abubuwan.

Matsalolin da wall-vault ke warware:

- **Jujjuya Maɓallin ta Atomatik**: Idan amfanin maɓalli ɗaya ya kai iyakansa ko ya tsaya na ɗan lokaci (cooldown), sai a yi amfani da na gaba a ɓoye. OpenClaw yana ci gaba da aiki ba tare da katse ba.
- **Musanya Ayyuka ta Atomatik (Fallback)**: Idan Google bai amsa ba, sai a yi amfani da OpenRouter; idan haka ma bai yi aiki ba, sai a koma Ollama, LM Studio, ko vLLM (AI na gida akan kwamfutarka). Zaman ba ya katse. Idan ayyukan asali sun dawo, za a koma kan su ta atomatik daga buƙatar gaba (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Daidaita Aiki a Lokaci Ɗaya (SSE)**: Idan ka canza ƙirar a dashboard ɗin ɓaure, za a nuna canje-canjen akan allon OpenClaw cikin dakika 1–3. SSE (Server-Sent Events) wata fasaha ce da ke ba wa sabar damar tura sabuntawa zuwa ga abokin ciniki a lokaci ɗaya.
- **Sanarwa a Lokaci Ɗaya**: Idan maɓalli ya ƙare ko akwai matsalar aiki, za a nuna wannan a ƙasan allon TUI na OpenClaw nan da nan.

> 💡 **Claude Code, Cursor, da VS Code** suma za a iya haɗa su, amma babban manufar wall-vault ita ce amfani tare da OpenClaw.

```
OpenClaw (Allon TUI na Terminal)
        │
        ▼
  wall-vault proxy (:56244)   ← Sarrafa maɓalli, routing, fallback, abubuwan da ke faruwa
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (fiye da ƙirar 340)
        ├─ Ollama / LM Studio / vLLM (Kwamfutarka, na ƙarshe)
        └─ OpenAI / Anthropic API
```

---

## Shigarwa

### Linux / macOS

Buɗe terminal ɗinka sannan ka manna waɗannan umarni kamar yadda suke.

```bash
# Linux (PC na sabar — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Yana sauko da fayil daga intanet.
- `chmod +x` — Yana sa fayil ɗin da aka sauko da shi ya "iya aiki". Idan ka tsallake wannan matakin za ka sami kuskuren "babu izini".

### Windows

Buɗe PowerShell (a matsayin mai gudanarwa) ka aiwatar da waɗannan umarni.

```powershell
# Sauko
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Ƙara PATH (yana aiki bayan sake buɗe PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **Menene PATH?** Jerin manyan fayiloli ne da kwamfuta ke bincika umarni a ciki. Idan ka ƙara cikin PATH, za ka iya rubuta `wall-vault` daga kowace babban fayil ka aiwatar da shi.

### Gina daga Tushe (na Masu Haɓaka)

Wannan yana aiki ne kawai idan an shigar da yanayin haɓaka harshen Go.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (sigar: v0.1.24.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Sigar alama lokaci**: Idan ka gina da `make build`, sigar za ta ƙirƙiru ta atomatik a tsari kamar `v0.1.24.20260406.211004` tare da kwanan wata da lokaci. Idan ka gina kai tsaye da `go build ./...`, sigar za ta nuna `"dev"` kawai.

---

## Farkon Farawa

### Aiwatar da magatakarda na setup

Bayan shigarwa, dole ne ka aiwatar da **magatakarda na saiti** ta amfani da umarnin da ke ƙasa. Magatakarda za ta jagorance ka cikin abubuwan da ake buƙata ɗaya bayan ɗaya ta hanyar tambayoyi.

```bash
wall-vault setup
```

Matakai da magatakarda ke bi su ne kamar haka:

```
1. Zaɓin harshe (harsuna 10 ciki har da Hausa)
2. Zaɓin jigo (light / dark / gold / cherry / ocean)
3. Yanayin aiki — kai kaɗai (standalone), ko tare da na'urori da yawa (distributed)
4. Sunan bot — sunan da zai bayyana a dashboard
5. Saita tasha — tsohuwar darajar: proxy 56244, ɓaure 56243 (danna Enter idan ba ka buƙatar canzawa ba)
6. Zaɓin ayyukan AI — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Saita tace kayan aikin tsaro
8. Saita alamar mai gudanarwa — kalmar sirri don kulle ayyukan gudanarwa na dashboard. Ana iya ƙirƙira ta atomatik
9. Saita kalmar sirrin ɓoye maɓallin API — don ajiye maɓallan cikin aminci sosai (zaɓi)
10. Wurin ajiye fayil ɗin saiti
```

> ⚠️ **Ku tuna alamar mai gudanarwa.** Za ka buƙace ta daga baya don ƙara maɓallan a dashboard ko canza saiti. Idan ka manta da ita za ka buƙaci gyara fayil ɗin saiti kai tsaye.

Idan magatakarda ta ƙare, fayil ɗin saiti `wall-vault.yaml` za ta ƙirƙiru ta atomatik.

### Aiwatarwa

```bash
wall-vault start
```

Sabobi biyu za fara aiki a lokaci guda:

- **Proxy** (`http://localhost:56244`) — wakilin da ke haɗa OpenClaw da ayyukan AI
- **Ɓaure na Maɓallan** (`http://localhost:56243`) — sarrafa maɓallin API da dashboard na yanar gizo

Buɗe browser ka je `http://localhost:56243` don ganin dashboard nan da nan.

---

## Yin Rajista na Maɓallin API

Akwai hanyoyi huɗu don yin rajista da maɓallin API. **Ga masu farawa, hanya ta 1 (masu-canji na yanayi) ce aka ba da shawara**.

### Hanya ta 1: Masu-Canji na Yanayi (Ana Ba Da Shawara — Mafi Sauƙi)

Masu-canji na yanayi su ne **ƙimomi da aka saita tun farko** waɗanda shiri ke karantawa lokacin da ya fara. Rubuta a terminal kamar haka:

```bash
# Rajista maɓallin Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Rajista maɓallin OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Aiwatarwa bayan rajista
wall-vault start
```

Idan kana da maɓallan da yawa, haɗa su da waƙafi (,). wall-vault zai yi amfani da maɓallan a jere ta atomatik (round robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Shawara**: Umarnin `export` yana aiki ne kawai a zaman terminal na yanzu. Don ya kasance bayan sake kunna kwamfuta, ƙara wannan layin cikin fayil ɗin `~/.bashrc` ko `~/.zshrc`.

### Hanya ta 2: Dashboard na UI (dannawa da linzami)

1. Buɗe browser ka je `http://localhost:56243`
2. A kan katin **🔑 Maɓallin API** a sama, danna maɓallin `[+ Ƙara]`
3. Shigar da irin ayyuka, ƙimar maɓalli, lakabi (sunan tunawa), da iyakar yau da kullum, sannan ka ajiye

### Hanya ta 3: REST API (don aiki ta atomatik da rubutun)

REST API hanya ce ta shirye-shirye don musanyar bayanai ta HTTP. Yana da amfani don rajista ta atomatik ta hanyar rubutun.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer alamar-mai-gudanarwa" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Babban Maɓalli",
    "daily_limit": 1000
  }'
```

### Hanya ta 4: Alamomin proxy (don ɗan gwaji na ɗan lokaci)

Yi amfani da wannan don sanya maɓalli na ɗan lokaci ba tare da rajista na hukuma ba. Maɓallin zai ɓace idan shirin ya tsaya.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Yadda Ake Amfani da Proxy

### Amfani da OpenClaw (Babban Manufa)

Yadda ake saita OpenClaw don haɗawa da ayyukan AI ta hanyar wall-vault.

Buɗe fayil ɗin `~/.openclaw/openclaw.json` ka ƙara abubuwan da ke ƙasa:

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
          { id: "wall-vault/hunter-alpha" },    // mahallin kyauta na 1M
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Hanya mafi sauƙi**: Danna maɓallin **🦞 Kwafi Saitin OpenClaw** a katin wakili a dashboard. Yankin da ke ɗauke da alama da adireshi wanda aka riga aka cika za a kwafa zuwa clipboard. Ka manna shi kawai.

**`wall-vault/` a gaban sunan ƙirar yana kai ina?**

Da sunan ƙirar, wall-vault ta san ta atomatik wace ayyukan AI za ta aika buƙata zuwa gare ta:

| Tsarin Ƙira | Ayyukan da Ake Haɗawa |
|----------|--------------|
| `wall-vault/gemini-*` | Google Gemini kai tsaye |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI kai tsaye |
| `wall-vault/claude-*` | Anthropic ta hanyar OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (mahallin kyauta na token miliyan 1) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/sunan-ƙira`, `openai/sunan-ƙira`, `anthropic/sunan-ƙira` d.s. | Ayyukan kai tsaye |
| `custom/google/sunan-ƙira`, `custom/openai/sunan-ƙira` d.s. | Cire ɓangaren `custom/` sannan a sake turawa |
| `sunan-ƙira:cloud` | Cire ɓangaren `:cloud` sannan a haɗa ta hanyar OpenRouter |

> 💡 **Menene mahallin (context)?** Shine adadin tattaunawar da AI ke iya tunawa a lokaci guda. 1M (token miliyan 1) yana nufin ana iya sarrafa tattaunawa mai tsawo sosai ko takardu masu tsawo a lokaci guda.

### Haɗawa Kai Tsaye da Tsarin Gemini API (don dacewa da kayan aiki na yanzu)

Idan kana da kayan aiki da suka kasance suna amfani da Google Gemini API kai tsaye, kawai ka canza adireshin zuwa na wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Ko kuma idan kayan aikinka na amfani da URL kai tsaye:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Amfani da OpenAI SDK (Python)

Za ka iya haɗa wall-vault cikin lambar Python da ke amfani da AI. Kawai ka canza `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # wall-vault yana sarrafa maɓallin API
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # shigar da tsarin provider/model
    messages=[{"role": "user", "content": "Sannu"}]
)
```

### Canza Ƙira Yayin Aiki

Don canza ƙirar AI yayin da wall-vault ke aiki:

```bash
# Canza ƙirar ta hanyar neman kai tsaye ga proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# A yanayin rarraba (boti da yawa), canza a sabar ɓaure → za a nuna nan da nan ta SSE
curl -X PUT http://localhost:56243/admin/clients/id-na-bot \
  -H "Authorization: Bearer alamar-mai-gudanarwa" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Duba Jerin Ƙirori da Ake Samu

```bash
# Duba dukkan jerin
curl http://localhost:56244/api/models | python3 -m json.tool

# Ƙirorin Google kawai
curl "http://localhost:56244/api/models?service=google"

# Bincika da suna (misali: ƙirori da ke ɗauke da "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Taƙaitaccen manyan ƙirori ta ayyuka:**

| Ayyuka | Manyan Ƙirori |
|--------|----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | Fiye da 346 (Hunter Alpha 1M mahallin kyauta, DeepSeek R1/V3, Qwen 2.5 d.s.) |
| Ollama | Yana gano ƙirorin sabar gida a kwamfutarka ta atomatik |
| LM Studio | Sabar gida a kwamfutarka (tasha 1234) |
| vLLM | Sabar gida a kwamfutarka (tasha 8000) |

---

## Dashboard na Ɓaure

Buɗe browser ka je `http://localhost:56243` don ganin dashboard.

**Tsarin allon:**
- **Layin sama mai ɗaurewa (topbar)**: Alama, zaɓin harshe da jigo, matsayin haɗin SSE
- **Grid ɗin katuna**: Katuna na wakili, ayyuka, da maɓallin API sun tsara a tsarin tiles

### Katin Maɓallin API

Kati wanda ke ba ka damar sarrafa dukkan maɓallin API da aka rajista a duba guda ɗaya.

- Yana nuna jerin maɓallan da aka raba ta ayyuka.
- `today_usage`: Token (adadin haruffa da AI ta karanta da rubuta) da aka sarrafa cikin nasara a yau
- `today_attempts`: Jimlar buƙatu na yau (nasara + gazawa)
- Maɓallin `[+ Ƙara]` don rajista sabon maɓalli, da `✕` don share maɓalli.

> 💡 **Menene token?** Shine ma'aunin da AI ke amfani da shi don sarrafa rubutu. Kusan kalma ɗaya ta Turanci, ko haruffa 1–2 na wasu harsuna. Kuɗin API yawanci ana ƙididdige su ta hanyar adadin token.

### Katin Wakili

Kati da ke nuna matsayin bot (wakili) da ke haɗe da proxy na wall-vault.

**Matsayin haɗin ana nuna shi a mataki 4:**

| Alama | Matsayi | Ma'ana |
|------|------|------|
| 🟢 | Yana Aiki | Proxy yana aiki yadda ya kamata |
| 🟡 | Jinkiri | Yana amsawa amma a hankali |
| 🔴 | Ba ya Aiki | Proxy ba ya amsawa |
| ⚫ | Ba a Haɗa/An Kashe | Proxy bai taɓa haɗawa da ɓaure ba ko an kashe shi |

**Bayani kan maɓallan ƙasan katin wakili:**

Lokacin da aka rajista wakili kuma aka ƙayyade **irin wakilin**, maɓallan sauƙi masu dacewa da irin wakilin za su bayyana ta atomatik.

---

#### 🔘 Maɓallin Kwafi Saiti — Yana ƙirƙira saitin haɗin ta atomatik

Idan ka danna maɓallin, yanki na saiti wanda ke ɗauke da alamar wakili, adireshin proxy, da bayanan ƙirar da aka riga aka cika za a kwafa zuwa clipboard. Kawai ka manna a wurin da aka nuna a teburin da ke ƙasa kuma saitin haɗin zai ƙare.

| Maɓalli | Irin Wakili | Wurin Mannawa |
|------|-------------|-------------|
| 🦞 Kwafi Saitin OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Kwafi Saitin NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Kwafi Saitin Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Kwafi Saitin Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Kwafi Saitin VSCode | `vscode` | `~/.continue/config.json` |

**Misali — Idan irin Claude Code ne, ga abin da za a kwafa:**

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

> ⚠️ **Sabuwar sigar Continue tana amfani da `config.yaml`.** Idan `config.yaml` yana nan, za a yi watsi da `config.json` gaba ɗaya. Tabbatar ka manna a cikin `config.yaml`.

**Misali — Idan irin Cursor ne:**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : alamar-wannan-wakili

// Ko masu-canji na yanayi:
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=alamar-wannan-wakili
```

> ⚠️ **Idan kwafin clipboard bai yi aiki ba**: Manufofin tsaro na browser na iya hana kwafin. Idan akwatin rubutu ya buɗe a matsayin pop-up, danna Ctrl+A don zaɓar duka sannan Ctrl+C don kwafa.

---

#### ⚡ Maɓallin Aiwatar da Atomatik — Danna sau ɗaya saitin ya ƙare

Idan irin wakilin shine `cline`, `claude-code`, `openclaw`, ko `nanoclaw`, maɓallin **⚡ Aiwatar da Saiti** zai bayyana a katin wakili. Idan ka danna wannan maɓallin, fayil ɗin saitin gida na wakilin zai sabunta ta atomatik.

| Maɓalli | Irin Wakili | Fayil ɗin da Ake Aiwatarwa |
|------|-------------|-------------|
| ⚡ Aiwatar da Saitin Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Aiwatar da Saitin Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Aiwatar da Saitin OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Aiwatar da Saitin NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Wannan maɓallin yana aika buƙata zuwa **localhost:56244** (proxy na gida). Dole ne proxy ya kasance tana aiki a kan wannan na'urar don ta yi aiki.

---

#### 🔀 Tsara Katuna ta Ja-da-Saki (v0.1.17)

Za ka iya **jawo** katuna na wakili a dashboard ka sake tsara su yadda kake so.

1. Ka kama katin wakili da linzami ka jawo shi
2. Ka saka shi a kan katin da kake so kuma tsarin zai canza
3. Sabon tsarin **an ajiye shi nan da nan a sabar** kuma zai kasance ko bayan sake sabuntawa

> 💡 Na'urori masu taɓawa (wayoyin hannu/tablets) ba su da tallafin yanzu. Yi amfani da browser na kwamfutar tebur.

---

#### 🔄 Daidaita Ƙira A Bangarori Biyu (v0.1.16)

Idan ka canza ƙirar wakili a dashboard ɗin ɓaure, saitin gida na wakilin zai sabunta ta atomatik.

**Ga Cline:**
- Canza ƙira a ɓaure → lambar SSE → proxy yana sabunta yankin ƙira a cikin `globalState.json`
- Yankuna da ake sabuntawa: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` da maɓallin API ba a taɓa su ba
- **Ana buƙatar sake ɗaukar VS Code (`Ctrl+Alt+R` ko `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - Cline ba ya sake karanta fayil ɗin saiti yayin aiki

**Ga Claude Code:**
- Canza ƙira a ɓaure → lambar SSE → proxy yana sabunta yankin `model` a cikin `settings.json`
- Yana bincika hanyoyin WSL da Windows ta atomatik (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Akasin hanya (wakili → ɓaure):**
- Lokacin da wakili (Cline, Claude Code d.s.) ya aika buƙata zuwa proxy, proxy yana ƙara bayanan ayyuka/ƙira na abokin ciniki a cikin heartbeat
- Ayyukan/ƙira da ake amfani da su a yanzu suna bayyana a lokaci ɗaya a katin wakili a dashboard ɗin ɓaure

> 💡 **Muhimmin abu**: Proxy yana gane wakili ta hanyar alamar Authorization na buƙatar kuma yana turawa ta atomatik zuwa ayyukan/ƙirar da aka saita a ɓaure. Ko da Cline ko Claude Code ya aika wani sunan ƙira daban, proxy yana maye gurbinsa da saitin ɓaure.

---

### Amfani da Cline a VS Code — Cikakken Jagora

#### Mataki na 1: Shigar da Cline

Shigar da **Cline** (ID: `saoudrizwan.claude-dev`) daga kasuwar ƙarin VS Code.

#### Mataki na 2: Rajista wakili a ɓaure

1. Buɗe dashboard ɗin ɓaure (`http://IP-na-ɓaure:56243`)
2. Danna **+ Ƙara** a sashen **Wakili**
3. Shigar da kamar haka:

| Yanki | Ƙima | Bayani |
|------|----|------|
| ID | `cline_na` | Alamar musamman (haruffa na Turanci, ba tazara) |
| Suna | `Cline Na` | Sunan da zai bayyana a dashboard |
| Irin Wakili | `cline` | ← dole ne ka zaɓi `cline` |
| Ayyuka | Zaɓi ayyukan da za a yi amfani (misali: `google`) | |
| Ƙira | Shigar da ƙirar da za a yi amfani (misali: `gemini-2.5-flash`) | |

4. Danna **Ajiye** kuma alama za ta ƙirƙiru ta atomatik

#### Mataki na 3: Haɗa Cline

**Hanya A — Aiwatar da atomatik (ana ba da shawara)**

1. Tabbatar cewa **proxy** na wall-vault yana aiki a kan wannan na'urar (`localhost:56244`)
2. Danna maɓallin **⚡ Aiwatar da Saitin Cline** a katin wakili a dashboard
3. Idan sanarwar "An aiwatar da saiti!" ta bayyana, an yi nasara
4. Sake ɗaukar VS Code (`Ctrl+Alt+R`)

**Hanya B — Saitin hannu**

Buɗe saiti (⚙️) a gefen Cline:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://adireshin-proxy:56244/v1`
  - Na'ura ɗaya: `http://localhost:56244/v1`
  - Wata na'ura kamar sabar Mini: `http://192.168.1.20:56244/v1`
- **API Key**: Alama da aka bayar daga ɓaure (kwafa daga katin wakili)
- **Model ID**: Ƙirar da aka saita a ɓaure (misali: `gemini-2.5-flash`)

#### Mataki na 4: Tabbatarwa

Aika kowace saƙo a cikin akwatin tattaunawar Cline. Idan yana aiki yadda ya kamata:
- Aya kore (● Yana Aiki) za ta bayyana a katin wakili a dashboard ɗin ɓaure
- Ayyuka/ƙira na yanzu za a nuna a katin (misali: `google / gemini-2.5-flash`)

#### Canza Ƙira

Idan kana son canza ƙirar Cline, canza a **dashboard ɗin ɓaure**:

1. Canza menu na ayyuka/ƙira a katin wakili
2. Danna **Aiwatar**
3. Sake ɗaukar VS Code (`Ctrl+Alt+R`) — sunan ƙirar a ƙasan Cline zai sabunta
4. Daga buƙatar gaba, sabuwar ƙirar za ta fara amfani

> 💡 A zahiri, proxy yana gane buƙatar Cline ta hanyar alama kuma yana turawa zuwa ƙirar saitin ɓaure. Ko ba ka sake ɗaukar VS Code ba **ƙirar da ake amfani da ita ta canza nan da nan** — sake ɗaukar don sabunta nuni na ƙira a UI na Cline ne kawai.

#### Gano Katsewar Haɗin

Idan ka rufe VS Code, katin wakili a dashboard ɗin ɓaure zai zama rawaya (jinkiri) bayan kusan **daƙiƙa 90**, kuma ja (ba ya aiki) bayan **minti 3**. (Daga v0.1.18, binciken matsayi na daƙiƙa 15 ya sa gano kasancewar ba a kan layi ba ya zama da sauri.)

#### Warware Matsaloli

| Alamar | Dalilin | Maganin |
|------|------|------|
| Kuskuren "haɗin ya gaza" a Cline | Proxy ba ya aiki ko adireshi ba daidai ba ne | Tabbatar da proxy da `curl http://localhost:56244/health` |
| Aya kore ba ta bayyana a ɓaure | Maɓallin API (alama) ba a saita shi ba | Danna maɓallin **⚡ Aiwatar da Saitin Cline** sake |
| Ƙirar a ƙasan Cline ba ta canja ba | Cline ta ajiye saitin a cikin cache | Sake ɗaukar VS Code (`Ctrl+Alt+R`) |
| Sunan ƙira marar kyau ya bayyana | Tsohuwar matsala (an gyara a v0.1.16) | Sabunta proxy zuwa v0.1.16 ko sama |

---

#### 🟣 Maɓallin Kwafi Umarnin Rarraba — Ana amfani da shi lokacin shigarwa a sabuwar na'ura

Ana amfani da shi lokacin da ake shigar da proxy na wall-vault a karon farko a sabuwar kwamfuta kuma ana haɗa shi da ɓaure. Danna maɓallin kuma cikakken rubutun shigarwa za a kwafa shi. Manna a terminal na sabuwar kwamfutar ka aiwatar — abubuwan da ke ƙasa za a yi su gaba ɗaya:

1. Shigar da fayil ɗin wall-vault (za a tsallake idan an riga an shigar)
2. Rajista ta atomatik ta hanyar ayyukan mai amfani na systemd
3. Fara aiki da haɗawa da ɓaure ta atomatik

> 💡 Alamar wannan wakilin da adireshin sabar ɓaure an riga an cika su cikin rubutun, don haka za ka iya aiwatar da shi nan da nan bayan mannawa ba tare da wani gyara ba.

---

### Katin Ayyuka

Kati don kunna da kashe ko saita ayyukan AI da za ka yi amfani da su.

- Maɓallan canza yanayin kunna da kashe na kowace ayyuka
- Idan ka shigar da adireshin sabar AI na gida (Ollama, LM Studio, vLLM d.s. da ke aiki a kwamfutarka), za ta gano ƙirorin da ake samu ta atomatik.
- **Nuni matsayin haɗin ayyukan gida**: Aya ● kusa da sunan ayyuka idan **kore** an haɗa, idan **toka** ba a haɗa ba
- **Fitilun hanya ta atomatik na ayyukan gida** (v0.1.23+): Ayyukan gida (Ollama, LM Studio, vLLM) suna kunna da kashe ta atomatik dangane da ko za a iya haɗa su. Idan ka kunna ayyuka, cikin daƙiƙa 15 aya ● za ta zama kore kuma akwatin tabbatarwa zai kunna; idan ka kashe ayyuka, za ta kashe ta atomatik. Wannan yana aiki ta hanya ɗaya da ayyukan girgije (Google, OpenRouter d.s.) da ke kunna da kashe ta atomatik dangane da kasancewar maɓallin API.

> 💡 **Idan ayyukan gida yana aiki a wata kwamfuta**: Shigar da IP na wannan kwamfutar a cikin akwatin URL na ayyuka. Misali: `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio). Idan ayyukan an ɗaura shi ga `127.0.0.1` maimakon `0.0.0.0`, ba za a iya isa gare shi ta IP na waje ba, don haka ka duba adireshin ɗaurin a saitin ayyukan.

### Shigar da Alamar Mai Gudanarwa

Idan ka yi ƙoƙarin amfani da ayyuka masu muhimmanci kamar ƙara ko share maɓallan a dashboard, pop-up na shigar da alamar mai gudanarwa zai bayyana. Shigar da alamar da ka saita a magatakarda na setup. Da zarar ka shigar, zai kasance har sai ka rufe browser.

> ⚠️ **Idan tabbatarwa ta gaza fiye da sau 10 cikin minti 15, za a toshe wannan IP na ɗan lokaci.** Idan ka manta alamar, duba abin `admin_token` a fayil ɗin `wall-vault.yaml`.

---

## Yanayin Rarraba (Multi-Bot)

Lokacin da ake gudanar da OpenClaw a kwamfutoci da yawa a lokaci guda, wannan tsarin ne na **raba ɓaure ɗaya na maɓallan**. Yana da sauƙi domin kana buƙatar sarrafa maɓallan a wuri ɗaya kawai.

### Misalin Tsari

```
[Sabar Ɓaure na Maɓallan]
  wall-vault vault    (ɓaure na maɓallan :56243, dashboard)

[WSL Alpha]            [Raspberry Pi Gamma]    [Mac Mini Local]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ daidaita SSE        ↕ daidaita SSE          ↕ daidaita SSE
```

Dukkan bot suna kallon sabar ɓaure a tsakiya, don haka canza ƙira ko ƙara maɓalli a ɓaure yana bayyana nan da nan a dukkan bot.

### Mataki na 1: Fara Sabar Ɓaure na Maɓallan

Aiwatar a kwamfutar da za ta kasance sabar ɓaure:

```bash
wall-vault vault
```

### Mataki na 2: Rajista Kowace Bot (Abokin Ciniki)

Rajista bayanan kowace bot da za ta haɗa da sabar ɓaure tun da wuri:

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

Aiwatar proxy a kowace kwamfutar da ke da bot ta hanyar ƙayyade adireshin sabar ɓaure da alama:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Maye gurbin **`192.168.x.x`** da ainihin adireshin IP na ciki na kwamfutar sabar ɓaure. Za ka iya bincika ta hanyar saitin router ko umarnin `ip addr`.

---

## Saita Farawa ta Atomatik

Idan wahalar kunna wall-vault da hannu a kowane lokacin da aka sake kunna kwamfuta, rajista shi a matsayin ayyukan tsarin. Da zarar an rajista, zai fara ta atomatik lokacin boot.

### Linux — systemd (yawancin Linux)

systemd shine tsarin Linux don fara da sarrafa shirye-shirye ta atomatik:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Duba tarihin abubuwan:

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

Tsarin da ke kula da aiwatar da shirye-shirye ta atomatik a macOS:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Sauko da NSSM daga [nssm.cc](https://nssm.cc/download) ka ƙara shi cikin PATH.
2. A PowerShell na mai gudanarwa:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor — Kayan Bincike

Umarnin `doctor` kayan aiki ne na wall-vault da ke **bincika kansa da gyara kansa**.

```bash
wall-vault doctor check   # Bincika yanayin yanzu (karantawa kawai, babu abin da ake canzawa)
wall-vault doctor fix     # Gyara matsaloli ta atomatik
wall-vault doctor all     # Bincike + gyara ta atomatik gaba ɗaya
```

> 💡 Idan wani abu ya yi kamar bai dace ba, aiwatar da `wall-vault doctor all` da fari. Yana magance matsaloli da yawa ta atomatik.


---

## RTK Tanadin Token

*(v0.1.24+)*

**RTK (Kayan Tanadin Token)** yana matsar fitowar umarnin sheli da wakilin AI na kodawa (Claude Code d.s.) ke aiwatarwa ta atomatik, yana rage yawan amfanin token. Misali, fitowar layi 15 na `git status` za a matsa zuwa taƙaitaccen layi 2.

### Asalin Amfani

```bash
# Nannade umarnin da wall-vault rtk kuma fitowa za ta tacewa ta atomatik
wall-vault rtk git status          # Yana nuna jerin fayiloli da suka canza kawai
wall-vault rtk git diff HEAD~1     # Layukan da suka canza + ƙaramin mahallin kawai
wall-vault rtk git log -10         # Hash + saƙon layi ɗaya kowane shigarwa
wall-vault rtk go test ./...       # Yana nuna gwaje-gwaje da suka gaza kawai
wall-vault rtk ls -la              # Umarnin da ba a tallafa ba ana yanke su ta atomatik
```

### Umarnin da Ake Tallafawa da Tasirin Tanadi

| Umarnin | Hanyar Tacewa | Adadin Tanadi |
|------|----------|--------|
| `git status` | Taƙaitaccen fayiloli da suka canza kawai | ~87% |
| `git diff` | Layukan da suka canza + mahallin layi 3 | ~60-94% |
| `git log` | Hash + saƙon layi na farko | ~90% |
| `git push/pull/fetch` | Cire ci gaba, taƙaitacce kawai | ~80% |
| `go test` | Nuna gazawa kawai, ƙidaya nasarori | ~88-99% |
| `go build/vet` | Nuna kuskurori kawai | ~90% |
| Sauran umarnin duka | Layi 50 na farko + 50 na ƙarshe, matsakaicin 32KB | Yana canzawa |

### Bututun Tacewa na Mataki 3

1. **Tace tsari ta umarnin** — Yana fahimtar tsarin fitowar git, go d.s. kuma yana ciro ɓangarorin da ke da ma'ana kawai
2. **Sarrafa bayan regex** — Cire lambobin launi na ANSI, rage layukan banza, taƙaita layukan da suka maimaitu
3. **Wucewa + yankewa** — Umarnin da ba a tallafa ba suna riƙe layi 50 na farko da 50 na ƙarshe kawai

### Haɗawa da Claude Code

Za ka iya saita ta hanyar ƙugiya `PreToolUse` na Claude Code don dukkan umarnin sheli su wuce ta RTK ta atomatik.

```bash
# Shigar da ƙugiya (ana ƙara ta ta atomatik cikin settings.json na Claude Code)
wall-vault rtk hook install
```

Ko ƙara da hannu cikin `~/.claude/settings.json`:

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

> 💡 **Adana lambar fita**: RTK yana mayar da lambar fita ta umarnin asali kamar yadda take. Idan umarnin ya gaza (exit code ≠ 0), AI ma za ta gane gazawar daidai.

> 💡 **Tilasta Turanci**: RTK yana aiwatar da umarnin da `LC_ALL=C` don samar da fitowar Turanci koyaushe ba tare da la'akari da saitin harshen tsarin ba. Wannan yana tabbatar da cewa tacewa yana aiki daidai.

---

## Bayani kan Masu-Canji na Yanayi

Masu-canji na yanayi hanya ce ta isar da ƙimomi na saiti zuwa ga shiri. Shigar da su a tsarin `export sunan-mai-canji=ƙima` a terminal, ko sanya su a fayil ɗin ayyukan farawa ta atomatik don su yi aiki koyaushe.

| Mai-Canji | Bayani | Misalin Ƙima |
|------|------|---------|
| `WV_LANG` | Harshen dashboard | `ko`, `en`, `ja` |
| `WV_THEME` | Jigon dashboard | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Maɓallin API na Google (da yawa da waƙafi) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Maɓallin API na OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Adireshin sabar ɓaure a yanayin rarraba | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Alamar tabbatarwa na abokin ciniki (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Alamar mai gudanarwa | `admin-token-here` |
| `WV_MASTER_PASS` | Kalmar sirrin ɓoye maɓallin API | `my-password` |
| `WV_AVATAR` | Hanyar fayil ɗin hoton avatar (dangi da `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Adireshin sabar gida na Ollama | `http://192.168.x.x:11434` |

---

## Warware Matsaloli

### Proxy Ba Ya Farawa

Mafi yawan lokaci tashar tana amfani da wani shiri.

```bash
ss -tlnp | grep 56244   # Duba wanene ke amfani da tasha 56244
wall-vault proxy --port 8080   # Fara da wani lambar tasha
```

### Kuskuren Maɓallin API (429, 402, 401, 403, 582)

| Lambar Kuskure | Ma'ana | Yadda Ake Magana |
|----------|------|----------|
| **429** | Buƙatu da yawa (an wuce iyaka) | Jira ɗan lokaci ko ƙara wani maɓalli |
| **402** | Ana buƙatar biyan kuɗi ko bashi ya ƙare | Ƙara kuɗi a ayyukan da abin ya shafa |
| **401 / 403** | Maɓallin ba daidai ba ne ko babu izini | Sake tabbatar da ƙimar maɓallin ka sake rajista |
| **582** | Cunkoson ƙofa (cooldown minti 5) | Zai buɗe ta atomatik bayan minti 5 |

```bash
# Duba jerin maɓallan da aka rajista da matsayinsu
curl -H "Authorization: Bearer alamar-mai-gudanarwa" http://localhost:56243/admin/keys

# Sake saita ma'aunin amfani na maɓalli
curl -X POST -H "Authorization: Bearer alamar-mai-gudanarwa" http://localhost:56243/admin/keys/reset
```

### Wakili Yana Nuna "Ba a Haɗa"

"Ba a Haɗa" yana nufin aikin proxy ba ya aika sigina (heartbeat) zuwa ɓaure. **Ba yana nufin ba a ajiye saiti ba.** Proxy dole ne ta san adireshin sabar ɓaure da alama kuma ta kasance tana aiki don ta canza zuwa yanayin haɗawa.

```bash
# Fara proxy ta hanyar ƙayyade adireshin sabar ɓaure, alama, da ID na abokin ciniki
WV_VAULT_URL=http://adireshin-sabar-ɓaure:56243 \
WV_VAULT_TOKEN=alamar-abokin-ciniki \
WV_VAULT_CLIENT_ID=id-na-abokin-ciniki \
wall-vault proxy
```

Idan haɗin ya yi nasara, zai canza zuwa 🟢 Yana Aiki a dashboard cikin kusan daƙiƙa 20.

### Ollama Ba Ya Haɗawa

Ollama shiri ne na gudanar da AI kai tsaye a kwamfutarka. Da fari tabbatar Ollama yana aiki.

```bash
curl http://localhost:11434/api/tags   # Idan jerin ƙirori ya bayyana, yana aiki yadda ya kamata
export OLLAMA_URL=http://192.168.x.x:11434   # Idan yana aiki a wata kwamfuta
```

> ⚠️ Idan Ollama ba ya amsawa, fara Ollama da farko ta hanyar umarnin `ollama serve`.

> ⚠️ **Manyan ƙirori suna da jinkirin amsawa**: Manyan ƙirori kamar `qwen3.5:35b`, `deepseek-r1` na iya ɗaukar mintuna da yawa don samar da amsa. Ko da ya yi kamar babu amsa, yana iya sarrafa shi yadda ya kamata, don haka ka jira.

---

## Canje-Canjen Kwanan Nan (v0.1.16 ~ v0.1.24)

### v0.1.24 (2026-04-06)
- **Ƙaramin umarnin RTK na tanadin token**: `wall-vault rtk <command>` yana tace fitowar umarnin sheli ta atomatik don rage yawan amfanin token na wakilin AI da 60-90%. Yana ɗauke da tacewa na musamman ga manyan umarnin kamar git, go, kuma umarnin da ba a tallafa ba ma ana yanke su ta atomatik. Yana haɗuwa ba tare da matsala ba ta hanyar ƙugiya `PreToolUse` na Claude Code.

### v0.1.23 (2026-04-06)
- **Gyaran canza ƙirar Ollama**: An gyara matsalar da canza ƙirar Ollama a dashboard ɗin ɓaure bai bayyana a proxy na gaske ba. A baya, mai-canji na yanayi (`OLLAMA_MODEL`) ne kawai ake amfani da shi, amma yanzu saitin ɓaure ne ake ba fifiko.
- **Fitilun hanya ta atomatik na ayyukan gida**: Ollama, LM Studio, da vLLM suna kunna ta atomatik idan za a iya haɗa su kuma suna kashe ta atomatik idan sun katse. Wannan yana aiki ta hanya ɗaya da musanyar ayyukan girgije ta atomatik da ke dogara ga maɓalli.

### v0.1.22 (2026-04-05)
- **Gyaran yankin content marar komai da ya ɓace**: Lokacin da ƙirorin tunani (gemini-3.1-pro, o1, claude thinking d.s.) suka yi amfani da iyakar max_tokens duka don tunani kuma suka kasa samar da amsa na gaske, proxy ya cire yankuna `content`/`text` na JSON na amsa ta `omitempty`, wanda ya haifar da kuskuren `Cannot read properties of undefined (reading 'trim')` a abokan ciniki na SDK na OpenAI/Anthropic. An canza shi don a haɗa yankuna koyaushe bisa ƙa'idojin API na hukuma.

### v0.1.21 (2026-04-05)
- **Tallafin ƙirorin Gemma 4**: Ƙirorin dangin Gemma kamar `gemma-4-31b-it`, `gemma-4-26b-a4b-it` za a iya amfani da su ta hanyar Google Gemini API.
- **Tallafin ayyukan LM Studio / vLLM na hukuma**: A baya waɗannan ayyukan sun ɓace a cikin hanyar proxy kuma koyaushe an maye gurbinsu da Ollama. Yanzu ana tura su yadda ya kamata ta hanyar API mai jituwa da OpenAI.
- **Gyaran nuni na ayyuka a dashboard**: Ko da fallback ya faru, dashboard koyaushe yana nuna ayyukan da mai amfani ya saita.
- **Nuni matsayin ayyukan gida**: Matsayin haɗin ayyukan gida (Ollama, LM Studio, vLLM d.s.) ana nuna shi da launin aya ● lokacin da dashboard ke ɗaukawa.
- **Mai-canji na yanayi na tace kayan aiki**: Yanayin isar kayan aiki (tools) za a iya saita shi da mai-canji na yanayi `WV_TOOL_FILTER=passthrough`.

### v0.1.20 (2026-03-28)
- **Ƙarfafa tsaro mai zurfi**: Kariya daga XSS (wurare 41), kwatanta alama na lokaci na kullum, ƙuntatawa CORS, iyakokin girman buƙata, kariya daga bi hanya, tabbatarwa SSE, ƙarfafa iyakar gudu d.s. abubuwa 12 na tsaro da aka inganta.

### v0.1.19 (2026-03-27)
- **Gano Claude Code a kan layi**: Claude Code da ba ta bi ta proxy ba suma ana nuna ta a kan layi a dashboard.

### v0.1.18 (2026-03-26)
- **Gyaran ayyukan fallback da ya makale**: Bayan fallback zuwa Ollama saboda kuskure na ɗan lokaci, idan ayyukan asali ta dawo, tana canzawa ta koma ta atomatik.
- **Inganta gano rashin kasancewa a kan layi**: Binciken matsayi na daƙiƙa 15 ya sa gano katsewar proxy ya zama da sauri.

### v0.1.17 (2026-03-25)
- **Tsara katuna ta ja-da-saki**: Za a iya jawo katuna na wakili don sake tsara tsarin su.
- **Maɓallan aiwatar da saiti a cikin layi**: Maɓallin [⚡ Aiwatar da Saiti] yana bayyana ga wakilai marasa kasancewa a kan layi.
- **Ƙarin irin wakilin cokacdir**.

### v0.1.16 (2026-03-25)
- **Daidaita ƙira a bangarori biyu**: Canza ƙirar Cline ko Claude Code a dashboard ɗin ɓaure yana bayyana ta atomatik.

---

*Don ƙarin bayani na API, duba [API.md](API.md).*
