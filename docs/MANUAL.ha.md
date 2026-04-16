# Jagorar Amfani na wall-vault
*(Last updated: 2026-04-16 — v0.2.2)*

---

## Abubuwan da ke Ciki

1. [Mene ne wall-vault?](#mene-ne-wall-vault)
2. [Shigarwa](#shigarwa)
3. [Farawa da Farko (masihirin setup)](#farawa-da-farko)
4. [Rajista API Key](#rajista-api-key)
5. [Yadda ake Amfani da Proxy](#yadda-ake-amfani-da-proxy)
6. [Dashboard na Key Vault](#dashboard-na-key-vault)
7. [Yanayin Rarraba (Multi Bot)](#yanayin-rarraba-multi-bot)
8. [Saita Farawa ta Atomatik](#saita-farawa-ta-atomatik)
9. [Doctor (Likita)](#doctor-likita)
10. [RTK Adana Token](#rtk-adana-token)
11. [Bayanin Muhalli Variables](#bayanin-muhalli-variables)
12. [Warware Matsaloli](#warware-matsaloli)

---

## Mene ne wall-vault?

**wall-vault = Wakili na AI (Proxy) + Akwatin API Key don OpenClaw**

Don amfani da sabis na AI, kana bukatar **API key**. API key ita ce kamar **katin shiga na dijital** wanda ke tabbatar da cewa "wannan mutum yana da hakkin amfani da wannan sabis". Amma wannan katin shiga yana iyaka na adadin amfani a rana, kuma idan ba a kula shi da kyau ba, akwai hadarin tonewa.

wall-vault yana adana wadannan katin shiga a cikin akwati mai tsaro, kuma tana aiki a matsayin **wakili (proxy)** tsakanin OpenClaw da sabis na AI. A takaice, OpenClaw tana bukatar kawai ta hada da wall-vault, sauran abubuwan da suka yi rikitarwa wall-vault ce za ta kula su.

Matsalolin da wall-vault ke warwarewa:

- **Juyawar API Key ta Atomatik**: Idan amfanin makulli guda ya kai iyaka ko kuma an hana shi na dan lokaci (cooldown), sai ya juya zuwa makulli na gaba a hankali. OpenClaw tana ci gaba da aiki ba tare da katse wa ba.
- **Sauya Sabis ta Atomatik (Fallback)**: Idan Google ba ta amsa ba, sai ta juya zuwa OpenRouter, idan wannan ma bai yi aiki ba, sai ta juya zuwa Ollama·LM Studio·vLLM (AI na cikin kwamfuta) da aka girka a kwamfutarka ta atomatik. Session ba ta katse ba. Idan sabis na asali ya dawo, bukatu na gaba za su koma ta atomatik (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Daidaitawa a Lokaci Guda (SSE)**: Idan ka canza model a dashboard na akwati, zai bayyana a fuskar OpenClaw a cikin dakika 1-3. SSE (Server-Sent Events) fasaha ce da server ke tura sauye-sauye zuwa client a lokaci guda.
- **Sanarwa a Lokaci Guda**: Abubuwa kamar karewa na key ko rushewar sabis suna bayyana nan da nan a kasan fuskar OpenClaw TUI (fuskar terminal).

> 💡 **Claude Code, Cursor, VS Code** ma za a iya hada su, amma dalilin asali na wall-vault shi ne amfani tare da OpenClaw.

```
OpenClaw (Fuskar Terminal na TUI)
        │
        ▼
  wall-vault Proxy (:56244)   ← Gudanar da key, turawa, fallback, abubuwa
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (model 340+)
        ├─ Ollama / LM Studio / vLLM (kwamfutarka, mafakar karshe)
        └─ OpenAI / Anthropic API
```

---

## Shigarwa

### Linux / macOS

Bude terminal kuma ka manna wadannan umarni kamar yadda suke.

```bash
# Linux (PC na yau da kullum, server — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Yana sauke fayil daga intanet.
- `chmod +x` — Yana sa fayil da aka sauke ya zama "mai iya gudana". Idan ka tsallake wannan mataki, za ka sami kuskuren "an hana izini".

### Windows

Bude PowerShell (a matsayin mai gudanarwa) kuma ka gudanar da wadannan umarni.

```powershell
# Sauke
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Kara zuwa PATH (yana aiki bayan sake bude PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **Mene ne PATH?** Jerin manyan fayiloli ne inda kwamfuta ke neman umarni. Kana bukatar ka kara shi zuwa PATH domin ka iya gudanar da `wall-vault` daga kowace babbar fayil.

### Gina daga Tushe (don Masu Bunkasa)

Wannan yana aiki ne kawai idan kana da yanayin bunkasa na harshen Go.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (sigar: v0.1.25.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Sigar Alamar Lokaci na Ginawa**: Idan ka gina da `make build`, sigar tana fitowa ta atomatik a tsarin da ya hada kwanan wata·lokaci kamar `v0.1.27.20260409`. Idan ka gina kai tsaye da `go build ./...`, sigar tana bayyana kawai a matsayin `"dev"`.

---

## Farawa da Farko

### Gudanar da masihirin setup

Bayan shigarwa, tabbatar ka gudanar da **masihirin saiti** da umarnin da ke kasa da farko. Masihirin zai jagorance ka ta hanyar tambayar ka abubuwan da ake bukata daya bayan daya.

```bash
wall-vault setup
```

Matakai da masihirin ke bi su kamar haka:

```
1. Zabi harshe (harsuna 10 ciki har da Hausa)
2. Zabi jigo (light / dark / gold / cherry / ocean)
3. Yanayin gudanarwa — zabi ko za ka yi amfani shi kadai (standalone) ko a kan injuna da yawa (distributed)
4. Shigar da sunan bot — sunan da zai bayyana a dashboard
5. Saita tashar — tushe: proxy 56244, vault 56243 (danna Enter kawai idan ba ka bukatar canjawa ba)
6. Zabi sabis na AI — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Saita tace kayan aiki na tsaro
8. Saita alamar mai gudanarwa — kalmar sirri don kulle fasalolin gudanarwar dashboard. Ana iya samar da shi ta atomatik
9. Saita kalmar sirri na rufe API key — idan kana so ka adana key cikin tsaro (zabi)
10. Hanyar adana fayil na saiti
```

> ⚠️ **Tabbatar ka tuna alamar mai gudanarwa.** Za ka bukata ta daga baya idan ka kara key ko canza saituna a dashboard. Idan ka rasa ta, za ka bukata ka gyara fayil na saiti kai tsaye.

Idan masihirin ya kammala, fayil na saiti `wall-vault.yaml` tana fitowa ta atomatik.

### Gudanarwa

```bash
wall-vault start
```

Server guda biyu suna farawa a lokaci guda:

- **Proxy** (`http://localhost:56244`) — Wakili da ke hada OpenClaw da sabis na AI
- **Key Vault** (`http://localhost:56243`) — Gudanar da API key da dashboard na yanar gizo

Bude `http://localhost:56243` a cikin browser ka don ganin dashboard nan da nan.

---

## Rajista API Key

Akwai hanyoyi hudu don rajista API key. **Ga masu farawa, Hanya ta 1 (muhalli variables) ana ba da shawarar ta**.

### Hanya ta 1: Muhalli Variables (Ana Ba da Shawarar — Mafi Sauki)

Muhalli variables su ne **dabi'u da aka saita tun farko** wadanda shirin ke karantawa idan ya fara. Shigar da wadannan a cikin terminal.

```bash
# Rajista key na Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Rajista key na OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Gudanar bayan rajista
wall-vault start
```

Idan kana da key da yawa, hada su da waƙafi (,). wall-vault za ta yi amfani da su a zagaye ta atomatik (round robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Shawarar**: Umarnin `export` yana aiki ne kawai ga zaman terminal na yanzu. Don ya ci gaba ko bayan sake kunna kwamfuta, kara layyin a cikin fayil `~/.bashrc` ko `~/.zshrc`.

### Hanya ta 2: UI na Dashboard (Danna da Linzami)

1. Bude `http://localhost:56243` a cikin browser
2. Danna maballin `[+ Kara]` a katin **🔑 API Key** a sama
3. Shigar da nau'in sabis, darajar key, lakabi (sunan tunawa), da iyakar rana sannan ka adana

### Hanya ta 3: REST API (don Atomatik·Rubutun)

REST API hanya ce da shirye-shirye ke musayar bayanai ta hanyar HTTP. Tana da amfani don rajista ta atomatik ta hanyar rubutun.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer alamar-mai-gudanarwa" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Key na Asali",
    "daily_limit": 1000
  }'
```

### Hanya ta 4: Tutar proxy (don Gwajin Dan Lokaci)

Yi amfani da wannan idan kana so ka shigar da key na dan lokaci don gwaji ba tare da rajista na gaskiya ba. Key tana bacewa idan ka rufe shirin.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Yadda ake Amfani da Proxy

### Amfani da OpenClaw (Babban Dalili)

Ga yadda ake saita OpenClaw don hada da sabis na AI ta hanyar wall-vault.

Bude fayil `~/.openclaw/openclaw.json` kuma ka kara abubuwan da ke kasa:

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
          { id: "wall-vault/hunter-alpha" },    // 1M context kyauta
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Hanya Mafi Sauki**: Danna maballin **🦞 Kwafa Saitin OpenClaw** a katin wakili a dashboard kuma snippet da alamar da adireshi da suka cika za a kwafa zuwa clipboard. Kawai ka manna shi.

**`wall-vault/` a gaban sunan model yana nufin ina?**

Ta hanyar kallon sunan model, wall-vault tana yanke shawara ta atomatik wace sabis na AI za ta tura bukatu:

| Tsarin Model | Sabis da ake Hadawa |
|----------|--------------|
| `wall-vault/gemini-*` | Google Gemini kai tsaye |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI kai tsaye |
| `wall-vault/claude-*` | Anthropic ta hanyar OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (1 miliyan token kyauta) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/sunan-model`, `openai/sunan-model`, `anthropic/sunan-model` d.s. | Hada kai tsaye zuwa sabis da ya dace |
| `custom/google/sunan-model`, `custom/openai/sunan-model` d.s. | Cire bangaren `custom/` kuma ka sake turawa |
| `sunan-model:cloud` | Cire bangaren `:cloud` kuma ka hada da OpenRouter |

> 💡 **Mene ne Context (mahallin magana)?** Shi ne adadin tattaunawa da AI za ta iya tunawa a lokaci guda. 1M (miliyan token) yana nufin za ta iya gudanar da tattaunawa ko takardun da suka yi tsawo a lokaci guda.

### Hada Kai Tsaye a Tsarin Gemini API (dacewa da kayan aiki da ke akwai)

Idan kana da kayan aiki da suka kasance suna amfani da Google Gemini API kai tsaye, kawai ka canza adireshi zuwa wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Ko idan kayan aikinka suna bayyana URL kai tsaye:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Amfani da OpenAI SDK (Python)

Za ka iya hada wall-vault a cikin lambar Python da ke amfani da AI. Kawai ka canza `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # wall-vault tana gudanar da API key ta atomatik
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # shigar a tsarin provider/model
    messages=[{"role": "user", "content": "Sannu"}]
)
```

### Canza Model Yayin Gudana

Don canza model na AI yayin da wall-vault tana gudana:

```bash
# Canza model ta hanyar neman proxy kai tsaye
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# A yanayin rarraba (multi bot), canza a server na vault → yana bayyana nan da nan ta SSE
curl -X PUT http://localhost:56243/admin/clients/id-na-bot-na \
  -H "Authorization: Bearer alamar-mai-gudanarwa" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Duba Jerin Model da Ake Samu

```bash
# Duba duk jerin
curl http://localhost:56244/api/models | python3 -m json.tool

# Duba model na Google kawai
curl "http://localhost:56244/api/models?service=google"

# Nema ta suna (misali: model da suka kunshi "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Takaitaccen Babban Model a Kowane Sabis:**

| Sabis | Babban Model |
|--------|----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ (Hunter Alpha 1M context kyauta, DeepSeek R1/V3, Qwen 2.5 d.s.) |
| Ollama | Yana gano server na cikin kwamfuta da aka girka ta atomatik |
| LM Studio | Server na cikin kwamfuta (tashar 1234) |
| vLLM | Server na cikin kwamfuta (tashar 8000) |

---

## Dashboard na Key Vault

Bude `http://localhost:56243` a cikin browser ka don ganin dashboard.

**Tsarin Fuska:**
- **Sandan sama da aka daidaita (topbar)**: Alamar, mai zabar harshe·jigo, alamar halin hadar SSE
- **Grid na Kati**: Katuna na wakili·sabis·API key an shirya su a tsarin tiles

### Katin API Key

Katin don gudanar da API key da aka rajista cikin saukin gani.

- Yana nuna jerin key da aka rarraba a kowane sabis.
- `today_usage`: Adadin token (haruffa da AI ya karanta da rubutawa) da aka gudanar cikin nasara a yau
- `today_attempts`: Jimlar kiraye-kiraye a yau (nasara + gazawa)
- Rajista sabuwar key da maballin `[+ Kara]`, kuma ka share key da `✕`.

> 💡 **Mene ne Token?** Nauyin da AI ke amfani da shi idan tana gudanar da rubutu. Kusan kalmar Turanci guda, ko haruffa 1-2 na wasu harsuna. Kudin API yawanci ana lissafa shi gwargwadon wannan adadin token.

### Katin Wakili

Katin da ke nuna halin bot (wakili) da suka hada da proxy na wall-vault.

**Halin hada yana bayyana a mataki 4:**

| Alama | Hali | Ma'ana |
|------|------|------|
| 🟢 | Yana Gudana | Proxy tana aiki yadda ya kamata |
| 🟡 | An Jinkirta | Amsa tana zuwa amma a hankali |
| 🔴 | Ba ta Kan Layi | Proxy ba ta amsa ba |
| ⚫ | Ba a Hada ba·An Kashe | Proxy ba ta taba hadawa da vault ba ko an kashe ta |

**Jagorar maballan a kasan katin wakili:**

Idan ka bayyana **nau'in wakili** yayin rajista wakili, maballan dacewa da nau'in nan suna bayyana ta atomatik.

---

#### 🔘 Maballin Kwafa Saiti — Yana samar da saitin hada ta atomatik

Idan ka danna maballin, snippet na saiti da alamar wakili, adireshi na proxy, da bayanan model da suka cika za a kwafa zuwa clipboard. Kawai ka manna abin da aka kwafa a wurin da aka nuna a tebur na kasa don kammala saitin hada.

| Maballin | Nau'in Wakili | Wurin Mannawa |
|------|-------------|-------------|
| 🦞 Kwafa Saitin OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Kwafa Saitin NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Kwafa Saitin Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Kwafa Saitin Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Kwafa Saitin VSCode | `vscode` | `~/.continue/config.json` |

**Misali — Idan nau'in Claude Code ne, abubuwan da ke kasa za a kwafa su:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "alamar-wannan-wakili"
}
```

**Misali — Idan nau'in VSCode (Continue) ne:**

```yaml
# ~/.continue/config.yaml  ← Manna a cikin config.yaml, ba config.json ba
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

> ⚠️ **Sabuwar sigar Continue tana amfani da `config.yaml`.** Idan `config.yaml` tana nan, `config.json` an yi watsi da ita gaba daya. Tabbatar ka manna a cikin `config.yaml`.

**Misali — Idan nau'in Cursor ne:**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : alamar-wannan-wakili

// Ko muhalli variables:
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=alamar-wannan-wakili
```

> ⚠️ **Kwafa zuwa clipboard bai yi aiki ba**: Ka'idojin tsaro na browser na iya hana kwafa. Idan akwatin rubutu ya bude a popup, zabi duka da Ctrl+A sannan kwafa da Ctrl+C.

---

#### ⚡ Maballin Aiwatar da Atomatik — Danna sau daya sai saiti ya kammala

Idan nau'in wakili `cline`, `claude-code`, `openclaw`, ko `nanoclaw` ne, maballin **⚡ Aiwatar da Saiti** yana bayyana a katin wakili. Idan ka danna wannan maballin, fayilolin saitin cikin gida na wakili da ya dace suna sabuntawa ta atomatik.

| Maballin | Nau'in Wakili | Fayil da ake Kai Hari |
|------|-------------|-------------|
| ⚡ Aiwatar da Saitin Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Aiwatar da Saitin Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Aiwatar da Saitin OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Aiwatar da Saitin NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Wannan maballin yana tura bukatu zuwa **localhost:56244** (proxy na cikin gida). Proxy dole ne ta kasance tana gudana a kan wannan injin domin ta yi aiki.

---

#### 🔀 Shirya Kati ta Ja da Sauke (v0.1.17, an inganta v0.1.25)

Za ka iya **ja** katuna na wakili a dashboard don sake shirya su bisa tsarin da kake so.

1. Ka kama yankin **fitilun traffik (●)** a saman hagu na kati da linzami kuma ka ja
2. Ka sauke shi a kan katin a matsayin da kake so kuma tsarin zai canja

> 💡 Jikin kati (wuraren shigarwa, maballai d.s.) ba sa jan su. Za ka iya kamawa ne kawai daga yankin fitilun traffik.

#### 🟠 Gano Tsarin Wakili (v0.1.25)

Idan proxy tana aiki yadda ya kamata amma tsarin wakili na cikin gida (NanoClaw, OpenClaw) ya mutu, fitilun traffik na kati yana canjawa zuwa **orange (yana kyaftawa)** kuma sakon "Tsarin wakili ya tsaya" ya bayyana.

- 🟢 Kore: Proxy + wakili yadda ya kamata
- 🟠 Orange (yana kyaftawa): Proxy yadda ya kamata, wakili ya mutu
- 🔴 Ja: Proxy ba ta kan layi
3. Tsarin da aka canza **an adana shi a server nan da nan** kuma yana ci gaba ko bayan sake sabuntawa

> 💡 A kan na'urorin tabawa (waya/tablet) har yanzu ba a tallafa ba. Yi amfani da browser na desktop.

---

#### 🔄 Daidaitawar Model ta Bangarori Biyu (v0.1.16)

Idan ka canza model na wakili a dashboard na vault, saitin cikin gida na wakili da ya dace yana sabuntawa ta atomatik.

**Don Cline:**
- Idan ka canza model a vault → taron SSE → proxy tana sabunta filin model a cikin `globalState.json`
- Filaye da ake sabuntawa: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` da API key ba a taba su ba
- **Ana bukatar sake sabuntar VS Code (`Ctrl+Alt+R` ko `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - Saboda Cline ba ta sake karantar fayil na saiti yayin gudana ba

**Don Claude Code:**
- Idan ka canza model a vault → taron SSE → proxy tana sabunta filin `model` a cikin `settings.json`
- Tana bincike ta atomatik a hanyoyin WSL da Windows duka (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Bangaren Baya (wakili → vault):**
- Idan wakili (Cline, Claude Code d.s.) ya tura bukatu zuwa proxy, proxy tana hada bayanan sabis·model na client a cikin heartbeat
- Sabis/model da ake amfani da shi yanzu yana bayyana a lokaci guda a katin wakili a dashboard na vault

> 💡 **Mafi Muhimmanci**: Proxy tana gano wakili ta hanyar alamar Authorization na bukatu, kuma tana turawa ta atomatik zuwa sabis/model da aka saita a vault. Ko da Cline ko Claude Code ta tura sunan model dabam, proxy tana soke shi da saitin vault.

---

### Amfani da Cline a VS Code — Cikakken Jagora

#### Mataki na 1: Girka Cline

Girka **Cline** (ID: `saoudrizwan.claude-dev`) daga Kasuwar Kari na VS Code.

#### Mataki na 2: Rajista Wakili a Vault

1. Bude dashboard na vault (`http://IP-na-vault:56243`)
2. Danna **+ Kara** a sashen **Wakili**
3. Shigar kamar haka:

| Fili | Daraja | Bayani |
|------|----|------|
| ID | `cline_na` | Alamar ganewa ta musamman (Turanci, ba tazara ba) |
| Suna | `Cline Na` | Sunan da zai bayyana a dashboard |
| Nau'in Wakili | `cline` | ← Dole ne ka zabi `cline` |
| Sabis | Zabi sabis (misali: `google`) | |
| Model | Shigar da model (misali: `gemini-2.5-flash`) | |

4. Danna **Adana** kuma alamar za ta fito ta atomatik

#### Mataki na 3: Hada da Cline

**Hanya A — Aiwatar da Atomatik (Ana Ba da Shawarar)**

1. Tabbatar **proxy** na wall-vault tana gudana a kan wannan injin (`localhost:56244`)
2. Danna maballin **⚡ Aiwatar da Saitin Cline** a katin wakili a dashboard
3. Idan ka ga sanarwar "An aiwatar da saiti cikin nasara!" ya yi nasara
4. Sake sabuntar VS Code (`Ctrl+Alt+R`)

**Hanya B — Saitin Hannu**

Bude saituna (⚙️) a gefen Cline:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://adireshi-na-proxy:56244/v1`
  - Injin guda: `http://localhost:56244/v1`
  - Wani injin kamar mini server: `http://192.168.1.20:56244/v1`
- **API Key**: Alamar da aka bayar daga vault (kwafa daga katin wakili)
- **Model ID**: Model da aka saita a vault (misali: `gemini-2.5-flash`)

#### Mataki na 4: Tabbatarwa

Tura kowace sako a tattaunawar Cline. Idan yadda ya kamata:
- Digon **kore (● Yana Gudana)** zai bayyana a katin wakili da ya dace a dashboard na vault
- Sabis/model na yanzu za su bayyana a katin (misali: `google / gemini-2.5-flash`)

#### Canza Model

Idan kana so ka canza model na Cline, canza a **dashboard na vault**:

1. Canza sabis/model a cikin menu na katin wakili
2. Danna **Aiwatar**
3. Sake sabuntar VS Code (`Ctrl+Alt+R`) — sunan model a kasan Cline zai sabuntu
4. Sabuwar model za ta kasance ana amfani da ita daga bukatu na gaba

> 💡 A gaskiya, proxy tana gano bukatun Cline ta alamar kuma tana turawa zuwa model na saitin vault. Ko ba ka sake sabuntar VS Code ba, **model da ake amfani da ita ta canja nan da nan** — sake sabuntawa don sabunta nuni na model a UI na Cline ne.

#### Gano Katsewar Hada

Idan ka rufe VS Code, katin wakili a dashboard na vault zai zama rawaya (an jinkirta) bayan kusan **dakika 1.5**, kuma ja (ba ta kan layi) bayan **minti 3**. (Daga v0.1.18, binciken hali a kowane dakika 15 ya hanzarta gano halin ba ta kan layi.)

#### Warware Matsaloli

| Alama | Dalili | Magani |
|------|------|------|
| Kuskuren "Hada ya gaza" a Cline | Proxy ba ta gudana ba ko adireshi ba daidai ba | Duba proxy da `curl http://localhost:56244/health` |
| Digon kore bai bayyana a vault ba | API key (alamar) ba a saita ta ba | Danna maballin **⚡ Aiwatar da Saitin Cline** sake |
| Model na kasan Cline bai canja ba | Cline tana adana saiti a cache | Sake sabuntar VS Code (`Ctrl+Alt+R`) |
| Sunan model mara daidai yana bayyana | Tsohuwar matsalar (an gyara a v0.1.16) | Sabunta proxy zuwa v0.1.16 ko sama |

---

#### 🟣 Maballin Kwafa Umarnin Turawa — Ana amfani da shi idan ana girka a sabon injin

Ana amfani da shi idan ana girka proxy na wall-vault a sabon kwamfuta da hadawa da vault a karon farko. Danna maballin kuma dukkan rubutun girka za a kwafa shi. Manna shi kuma gudanar da shi a terminal na sabon kwamfuta kuma wadannan za a gudanar da su a lokaci guda:

1. Girka binary na wall-vault (an tsallake idan an riga an girka)
2. Rajista ta atomatik na sabis na systemd na mai amfani
3. Fara sabis kuma hada ta atomatik da vault

> 💡 Rubutun ya kunshi alamar wannan wakili da adireshi na server na vault da suka cika, don haka za ka iya gudanar da shi nan da nan bayan mannawa ba tare da wani gyara ba.

---

### Katin Sabis

Katin don kunna, kashe, ko saita sabis na AI da za a yi amfani da su.

- Maballin kunna·kashe na kowane sabis
- Idan ka shigar da adireshi na server na AI na cikin gida (Ollama, LM Studio, vLLM d.s. da ke gudana a kwamfutarka), zai gano model da ake samu ta atomatik.
- **Nuni na halin hadar sabis na cikin gida**: Digon ● kusa da sunan sabis idan **kore** an hada, **toka** ba a hada ba
- **Fitilun traffik ta atomatik na sabis na cikin gida** (v0.1.23+): Sabis na cikin gida (Ollama, LM Studio, vLLM) suna kunna/kashe ta atomatik bisa yiwuwar hada. Idan ka kunna sabis, a cikin dakika 15 ● ya zama kore kuma akwatin zabi ya kunna, kuma idan ka kashe sabis, ya kashe ta atomatik. Hanya guda ce da sabis na cloud (Google, OpenRouter d.s.) da ke canjawa ta atomatik bisa yiwuwar API key.

> 💡 **Idan sabis na cikin gida tana gudana a wani kwamfuta**: Shigar da IP na wannan kwamfuta a wurin shigar URL na sabis. Misali: `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio). Idan sabis tana daure ne kawai ga `127.0.0.1` maimakon `0.0.0.0`, shiga ta IP na waje ba zai yi aiki ba, don haka ka duba adireshi na daure a saitin sabis.

### Shigar da Alamar Mai Gudanarwa

Idan ka yi kokarin amfani da fasaloli masu muhimmanci kamar kara·share key a dashboard, popup na shigar da alamar mai gudanarwa za ta bayyana. Shigar da alamar da ka saita a masihirin setup. Bayan shigarwa sau daya, tana ci gaba har ka rufe browser.

> ⚠️ **Idan gazawar tabbatarwa ta wuce sau 10 a cikin minti 15, za a toshe IP da ya dace na dan lokaci.** Idan ka manta alamar, duba abin `admin_token` a cikin fayil `wall-vault.yaml`.

---

## Yanayin Rarraba (Multi Bot)

Idan ana gudanar da OpenClaw a kwamfutoci da yawa a lokaci guda, wannan tsari ne inda **akwatin key guda ana raba shi**. Yana da sauki saboda kana bukatar gudanar da key a wuri guda kawai.

### Misalin Tsari

```
[Server na Key Vault]
  wall-vault vault    (Key Vault :56243, dashboard)

[WSL Alpha]           [Raspberry Pi Gamma]    [Mac Mini na Cikin Gida]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ Daidaitawar SSE     ↕ Daidaitawar SSE       ↕ Daidaitawar SSE
```

Dukkan bot suna kallon server na vault na tsakiya, don haka idan ka canza model ko kara key a vault, yana bayyana a dukkan bot nan da nan.

### Mataki na 1: Fara Server na Key Vault

Gudanar a kwamfutar da za a yi amfani da ita a matsayin server na vault:

```bash
wall-vault vault
```

### Mataki na 2: Rajista Kowane Bot (Client)

Rajista tun farko bayanan kowane bot da ke hadawa da server na vault:

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

### Mataki na 3: Fara Proxy a Kowane Kwamfutar Bot

Gudanar da proxy ta hanyar bayyana adireshi na server na vault da alamar a kowane kwamfutar da aka girka bot:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Canja bangaren **`192.168.x.x`** zuwa adireshi na IP na ciki na gaske na kwamfutar server na vault. Za ka iya duba shi a saitin router ko ta umarnin `ip addr`.

---

## Saita Farawa ta Atomatik

Idan wahalar kunna wall-vault da hannu a duk lokacin da ka sake kunna kwamfuta, rajista shi a matsayin sabis na tsarin. Bayan rajista sau daya, yana farawa ta atomatik a lokacin bootup.

### Linux — systemd (yawancin Linux)

systemd tsarin ne da ke fara·gudanar da shirye-shirye ta atomatik a Linux:

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

Tsarin ne da ke gudanar da fara shirye-shirye ta atomatik a macOS:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Sauke NSSM daga [nssm.cc](https://nssm.cc/download) kuma ka kara shi zuwa PATH.
2. A PowerShell na mai gudanarwa:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor (Likita)

Umarnin `doctor` kayan aiki ne da ke **bincike kansa da gyara kansa** idan an saita wall-vault yadda ya kamata.

```bash
wall-vault doctor check   # Bincike halin yanzu (karanta kawai, ba a canza komai ba)
wall-vault doctor fix     # Gyara matsaloli ta atomatik
wall-vault doctor all     # Bincike + gyara ta atomatik a lokaci guda
```

> 💡 Idan wani abu bai dace ba, gudanar da `wall-vault doctor all` da farko. Yana kama matsaloli da yawa ta atomatik.

---

## RTK Adana Token

*(v0.1.24+)*

**RTK (Kayan Adana Token)** yana matsawa ta atomatik fitowar umarnin shell da wakili na AI coding (kamar Claude Code) ke gudanarwa don rage amfanin token. Misali, fitowar layi 15 na `git status` tana ragewa zuwa takaitaccen layi 2.

### Asalin Amfani

```bash
# Nannade umarni da wall-vault rtk kuma fitowa za ta tace ta atomatik
wall-vault rtk git status          # Jerin fayiloli da suka canja kawai
wall-vault rtk git diff HEAD~1     # Layukan canja + mafi karancin mahallin magana
wall-vault rtk git log -10         # Hash + sakon layi daya a kowane
wall-vault rtk go test ./...       # Gwaje-gwajen da suka gaza kawai
wall-vault rtk ls -la              # Umarni marasa tallafi suna yankewa ta atomatik
```

### Umarni da ake Tallafawa da Tasirin Ragewa

| Umarni | Hanyar Tace | Kimar Ragewa |
|------|----------|--------|
| `git status` | Takaitaccen fayiloli da suka canja kawai | ~87% |
| `git diff` | Layuka da suka canja + mahallin magana na layi 3 | ~60-94% |
| `git log` | Hash + sakon layi na farko | ~90% |
| `git push/pull/fetch` | Cire ci gaba, takaitawa kawai | ~80% |
| `go test` | Nuna gazawa kawai, kidaya wanda suka wuce | ~88-99% |
| `go build/vet` | Nuna kuskure kawai | ~90% |
| Dukkan sauran umarni | Layi 50 na farko + layi 50 na karshe, kololuwa 32KB | Mai canjawa |

### Pipeline na Tace na Mataki 3

1. **Tace na tsari a kowane umarni** — Yana fahimtar tsarin fitowar git, go d.s. kuma yana ciro sassan da suka dace kawai
2. **Sarrafa bayan regex** — Cire lambobin launi na ANSI, rage layukan wofi, tattara layukan da suka maimaita
3. **Wucewa + yankewa** — Umarnin da ba a tallafa ba suna adana layi 50 na farko/karshe kawai

### Hadawa da Claude Code

Za ka iya saita hook na `PreToolUse` na Claude Code don dukkan umarnin shell su wuce ta RTK ta atomatik.

```bash
# Girka hook (an kara ta atomatik zuwa settings.json na Claude Code)
wall-vault rtk hook install
```

Ko kara da hannu zuwa `~/.claude/settings.json`:

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

> 💡 **Adana Exit code**: RTK tana mayar da exit code na asalin umarni kamar yadda yake. Idan umarni ya gaza (exit code ≠ 0), AI ma tana gano gazawa daidai.

> 💡 **Tilasta Turanci**: RTK tana gudanar da umarni da `LC_ALL=C` don samar da fitowa a Turanci koyaushe ba tare da la'akari da saitin harshen tsarin ba. Wannan yana tabbatar da cewa tace yana aiki daidai.

---

## Bayanin Muhalli Variables

Muhalli variables hanya ce don tura dabi'un saiti zuwa shiri. Shigar a tsarin `export sunan-kigezo=daraja` a terminal, ko saka a cikin fayil na sabis na farawa ta atomatik don ya kasance yana aiki koyaushe.

| Kigezo | Bayani | Misalin Daraja |
|------|------|---------|
| `WV_LANG` | Harshen dashboard | `ko`, `en`, `ja` |
| `WV_THEME` | Jigon dashboard | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Key na API na Google (da yawa da wakafi) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Key na API na OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Adireshi na server na vault a yanayin rarraba | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Alamar tabbatarwar client (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Alamar mai gudanarwa | `admin-token-here` |
| `WV_MASTER_PASS` | Kalmar sirri na rufe API key | `my-password` |
| `WV_AVATAR` | Hanyar fayil na hoton avatar (dandalin dangantaka daga `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Adireshi na server na cikin gida na Ollama | `http://192.168.x.x:11434` |

---

## Warware Matsaloli

### Idan Proxy Ba Ta Farawa

Sau da yawa saboda wani shiri ya riga ya yi amfani da tashar.

```bash
ss -tlnp | grep 56244   # Duba waye ke amfani da tashar 56244
wall-vault proxy --port 8080   # Fara da wata lambar tashar
```

### Idan Kuskuren API Key Ya Faru (429, 402, 401, 403, 582)

| Lambar Kuskure | Ma'ana | Magani |
|----------|------|----------|
| **429** | Bukatu da yawa (amfani ya wuce) | Jira dan lokaci ko kara wata key |
| **402** | Ana bukatar biya ko credit bai isa ba | Cika credit a sabis da ya dace |
| **401 / 403** | Key ba daidai ba ce ko ba ta da izini | Sake duba darajar key kuma sake rajista |
| **582** | Gateway ya cika (cooldown minti 5) | Ya warware kansa bayan minti 5 |

```bash
# Duba jerin da halin key da aka rajista
curl -H "Authorization: Bearer alamar-mai-gudanarwa" http://localhost:56243/admin/keys

# Sake saita kidaya amfanin key
curl -X POST -H "Authorization: Bearer alamar-mai-gudanarwa" http://localhost:56243/admin/keys/reset
```

### Idan Wakili Yana Bayyana a matsayin "Ba a Hada ba"

"Ba a Hada ba" yana nufin tsarin proxy ba ya tura sigina (heartbeat) zuwa vault. **Ba yana nufin ba a adana saiti ba.** Proxy tana bukatar gudana tana sanin adireshi na server na vault da alamar don ta canja zuwa halin hadawa.

```bash
# Fara proxy ka bayyana adireshi na server na vault, alamar, da ID na client
WV_VAULT_URL=http://adireshi-na-server-na-vault:56243 \
WV_VAULT_TOKEN=alamar-na-client \
WV_VAULT_CLIENT_ID=id-na-client \
wall-vault proxy
```

Idan hada ya yi nasara, zai canja zuwa 🟢 Yana Gudana a dashboard a cikin kusan dakika 20.

### Idan Hadar Ollama Ba Ta Aiki

Ollama shiri ne da ke gudanar da AI kai tsaye a kwamfutarka. Da farko ka duba ko an kunna Ollama.

```bash
curl http://localhost:11434/api/tags   # Idan jerin model ya bayyana, yadda ya kamata ne
export OLLAMA_URL=http://192.168.x.x:11434   # Idan tana gudana a wani kwamfuta
```

> ⚠️ Idan Ollama ba ta amsa ba, fara Ollama da farko da umarnin `ollama serve`.

> ⚠️ **Manyan model suna jinkiri**: Manyan model kamar `qwen3.5:35b`, `deepseek-r1` na iya daukar mintuna da yawa don samar da amsa. Ko ya zama kamar babu amsa, yana iya kasancewa ana gudanar da shi yadda ya kamata ne, don haka ka yi hakuri.

---

## Bayanin haɓaka v0.2

- `Service` ya karɓa `default_model` da `allowed_models`. Zaɓin model na musamman na kowane sabis yanzu ana saita shi kai tsaye akan katunan sabis.
- `Client.default_service` / `default_model` an sake sata sunansu kuma an sake tafsira su a matsayin `preferred_service` / `model_override`. Idan override bai cika ba, model na musamman na sabis aka yi amfani da shi.
- A farko na buguwar v0.2, jerin `vault.json` da ke akwai ya bugi ta atomatik, kuma halin da jiya aka sani ana ajiya shi a matsayin `vault.json.pre-v02.{timestamp}.bak`.
- An sake tsara dashboard zuwa yankunan uku: sidebar na hayin hagu, grid na kartuna na tsakiya, da slideover na gyarawa na hayin dama.
- Hanyoyin Admin API ba su canja ba, amma tsarin bukatu/amsoshin jini suka sauya — tsohuwan tsingiyoyin CLI za su bukatar sabuntawa daidai.

---

## Sabbin Abubuwan v0.2.1

- **Tura abubuwa da yawa na mutumedia kai tsaye (OpenAI → Gemini)**: `/v1/chat/completions` yanzu yana karɓar nau'ikan sassan abun ciki guda shida baya ga `text` — `input_audio`, `input_video`, `input_image`, `input_file`, da `image_url` (data URIs da kuma URLs na http(s) na waje ≤ 5 MB). Proxy tana mayar da kowanne zuwa `inlineData` na Gemini. Abokan ciniki masu jituwa da OpenAI kamar EconoWorld suna iya turawa blobs na sauti / hoto / bidiyo kai tsaye.
- **Nau'in agent na EconoWorld**: `POST /agent/apply` tare da `agentType: "econoworld"` yana rubuta saitunan wall-vault a cikin `analyzer/ai_config.json` na aikin. `workDir` yana karɓar jerin hanyoyi masu yuwuwa da aka raba da waƙafi kuma yana mayar da hanyoyin drive na Windows zuwa hanyoyin dauka na WSL.
- **Grid na keys na dashboard + CRUD**: keys 11 ana nuna su a matsayin kartuna masu ƙarami tare da slideover na + ƙara / ✕ share.
- **Ƙara sabis + sake tsara ta hanyar ja-da-zube**: grid na sabis ya samu maballin + ƙara da riƙo na ja (`⋮⋮`).
- **Header / footer / motsi na jigo / canza harshe** an mayar da su. Jigogi 7 (cherry/dark/light/ocean/gold/autumn/winter) suna kunna sakamakon barbashi a kan Layer a bayan kartuna amma sama da bango.
- **UX na kore slideover**: danna waje ko Esc yana rufe slideover.
- **Mai nuna halin SSE** a cikin footer (kore = an haɗa, orange = ana sake haɗawa, launin toka = an cire haɗi).

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

## Sauye-sauye na Kwanan Nan (v0.1.16 ~ v0.1.27)

### v0.1.27 (2026-04-09)
- **Gyaran sunan model na fallback na Ollama**: An gyara matsalar da sunan model mai prefix na provider (misali: `google/gemini-3.1-pro-preview`) ya kasance ana turawa zuwa Ollama kamar yadda yake lokacin fallback daga wani sabis. Yanzu ana maye gurbinsa ta atomatik da muhalli variable/model na asali.
- **Ragewa mai yawa na lokacin cooldown**: 429 rate limit 30min→5min, 402 biya 1hr→30min, 401/403 24hr→6hr. Hana halin da dukkan key ke cikin cooldown a lokaci guda da ya sa proxy ta tsaya gaba daya.
- **Sake gwadawa ta tilas lokacin cikakken cooldown**: Idan dukkan key suna cikin halin cooldown, key da za ta warware da farko ana sake gwada ta ta tilas don hana kin amincin bukatu.
- **Gyaran nunin jerin sabis**: Amsoshin `/status` suna nuna jerin sabis na gaske da aka daidaita daga vault (hana rashin anthropic d.s.).

### v0.1.25 (2026-04-08)
- **Gano tsarin wakili**: Proxy tana gano ko wakili na cikin gida (NanoClaw/OpenClaw) yana da rai kuma tana nuna shi da fitilun traffik na orange a dashboard.
- **Ingantaccen hannun ja**: An canza don a iya kamawa kawai daga yankin fitilun traffik (●) lokacin shirya kati. Ba zai yiwu a ja ta kuskure daga wuraren shigarwa ko maballai ba.

### v0.1.24 (2026-04-06)
- **Umarni na RTK don adana token**: `wall-vault rtk <command>` tana tace fitowar umarnin shell ta atomatik don rage amfanin token na wakilin AI da 60-90%. Tana da tacewa na musamman don babban umarni kamar git, go, kuma umarnin da ba a tallafa ba ma ana yankewa ta atomatik. Tana hadawa ba tare da wata illa ba ta hanyar hook na `PreToolUse` na Claude Code.

### v0.1.23 (2026-04-06)
- **Gyaran canza model na Ollama**: An gyara matsalar da canza model na Ollama a dashboard na vault ba ya bayyana a proxy ba. A da tana amfani da muhalli variable (`OLLAMA_MODEL`) kawai, yanzu saitin vault yana da fifiko.
- **Fitilun traffik ta atomatik na sabis na cikin gida**: Ollama·LM Studio·vLLM suna kunna ta atomatik idan za a iya hada su, kuma suna kashe ta atomatik idan sun katse. Hanya guda ce da canjawa ta atomatik na sabis na cloud bisa key.

### v0.1.22 (2026-04-05)
- **Gyaran filin content na wofi da ya bace**: Idan model na tunani (gemini-3.1-pro, o1, claude thinking d.s.) suka yi amfani da iyakar max_tokens don tunani kuma ba za su iya samar da amsa na gaske ba, proxy tana cire filaye `content`/`text` na JSON na amsa da `omitempty`, wanda ke sa SDK na client na OpenAI/Anthropic su ruguje da kuskuren `Cannot read properties of undefined (reading 'trim')`. An canza don koyaushe ya hada filaye kamar yadda ke cikin ka'idojin API na hukuma.

### v0.1.21 (2026-04-05)
- **Tallafin model na Gemma 4**: Model na dangin Gemma kamar `gemma-4-31b-it`, `gemma-4-26b-a4b-it` za a iya amfani da su ta hanyar Google Gemini API.
- **Tallafin hukuma na sabis na LM Studio / vLLM**: A da wadannan sabis sun kasance ana barin su a wajen turawa na proxy kuma koyaushe ana maye gurbin su da Ollama. Yanzu suna turawa yadda ya kamata ta hanyar API mai dacewa da OpenAI.
- **Gyaran nunin sabis a dashboard**: Ko da fallback ya faru, dashboard koyaushe yana nuna sabis da mai amfani ya saita.
- **Nunin halin sabis na cikin gida**: Halin hadar sabis na cikin gida (Ollama, LM Studio, vLLM d.s.) yana bayyana ta launi na digon ● idan an loda dashboard.
- **Muhalli variable na tace kayan aiki**: Yanayin wucewar kayan aiki (tools) za a iya saita shi da muhalli variable `WV_TOOL_FILTER=passthrough`.

### v0.1.20 (2026-03-28)
- **Kare tsaro mai zurfi**: Hana XSS (wurare 41), kwatanta alamar a lokaci daya, takurawa na CORS, iyakoki na girman bukatu, hana wucewar hanya, tabbatarwar SSE, karfafawa na iyakar gudu, da ingantattun abubuwa 12 na tsaro.

### v0.1.19 (2026-03-27)
- **Gano Claude Code a kan layi**: Claude Code da ba ta wucewa ta proxy ba ma tana bayyana a matsayin a kan layi a dashboard.

### v0.1.18 (2026-03-26)
- **Gyaran makale da sabis na fallback**: Bayan fallback na dan lokaci zuwa Ollama, idan sabis na asali ya dawo ya koma ta atomatik.
- **Ingantaccen gano ba ta kan layi**: Gano proxy da ta tsaya ya hanzarta da binciken hali a kowane dakika 15.

### v0.1.17 (2026-03-25)
- **Shirya kati ta ja da sauke**: Za a iya jan katuna na wakili don canza tsari.
- **Maballin aiwatar da saiti na cikin layi**: Maballin [⚡ Aiwatar da Saiti] yana bayyana ga wakili da ba su kan layi ba.
- **An kara nau'in wakili na cokacdir**.

### v0.1.16 (2026-03-25)
- **Daidaitawar model ta bangarori biyu**: Canza model na Cline·Claude Code a dashboard na vault yana bayyana ta atomatik.

---

*Don cikakken bayanan API, duba [API.md](API.md).*
