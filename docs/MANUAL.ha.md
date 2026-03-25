# Jagoran Amfani da wall-vault
*(Ƙarshen sabuntawa: 2026-03-20 — v0.1.15)*

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
10. [Bayani kan Masu-Canji na Yanayi](#bayani-kan-masu-canji-na-yanayi)
11. [Warware Matsaloli](#warware-matsaloli)

---

## Menene wall-vault?

**wall-vault = Wakilin AI (Proxy) + Ɓaure na Maɓallin API — na OpenClaw**

Don amfani da ayyukan AI, kana buƙatar **maɓallin API** — wato "takardar shiga ta dijital" da take tabbatar cewa ana ba ka izinin amfani da ayyukan. Amma wannan takardar shiga tana da ƙayyadaddun amfani a kowace rana, kuma idan ba a lura da ita sosai ba, tana iya fita a hannun marasa izini.

wall-vault yana ajiye waɗannan maɓallan cikin **ɓaure mai aminci**, kuma yana aiki a matsayin **wakili (proxy)** tsakanin OpenClaw da ayyukan AI. A taƙaice, OpenClaw yana haɗawa da wall-vault kawai, sannan wall-vault yana kula da sauran abubuwan.

Matsalolin da wall-vault ke warware:

- **Jujjuya Maɓallin ta Atomatik**: Idan amfanin maɓalli ɗaya ya kai iyakansa ko ya tsaya na ɗan lokaci (cooldown), sai a yi amfani da na gaba a ɓoye. OpenClaw yana ci gaba da aiki ba tare da katse ba.
- **Musanya Ayyuka ta Atomatik (Fallback)**: Idan Google bai amsa ba, sai a yi amfani da OpenRouter; idan haka ma bai yi aiki ba, sai a koma Ollama (AI na gida akan kwamfutarka). Zaman ba ya katse.
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
        └─ Ollama (Kwamfutarka, na ƙarshe)
```

---

## Shigarwa

### Linux / macOS

Buɗe terminal ɗinka kuma manna umarnin da ke ƙasa kamar yadda yake.

```bash
# Linux (PC na yau da kullum, uwar garke — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (Mac M1/M2/M3)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Saukewa fayil ɗin daga intanet.
- `chmod +x` — Yana mai da fayil ɗin da aka saukewa ya zama "mai iya gudana". Idan ka manta wannan mataki, za ka samu kuskure mai cewa "babu izini".

### Windows

Buɗe PowerShell (a matsayin mai gudanarwa) kuma gudanar da umarnin da ke ƙasa.

```powershell
# Saukewa
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Ƙara zuwa PATH (za a yi aiki bayan sake farawa PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **Menene PATH?** Jerin manyan jakobi ne da kwamfuta ke bincika lokacin neman umarnin. Idan ka ƙara zuwa PATH, za ka iya rubuta `wall-vault` ko'ina a terminal kuma zai gudana.

### Gina daga Kodo (Don Masu Haɓakawa)

Wannan yana shafar kawai waɗanda suka shigar da yanayin haɓakar Go.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (sigar: v0.1.6.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Sigar da Lokaci na Gini**: Idan ka gina ta `make build`, za a samar da sigar ta atomatik a tsarin `v0.1.6.20260314.231308` — mai ƙunshe da kwanan wata da lokaci. Idan ka gina kai tsaye da `go build ./...`, sigar za ta nuna `"dev"` kawai.

---

## Farkon Farawa

### Gudanar da Magatakarda na setup

Bayan shigarwa, dole ne ka gudanar da **magatakarda na tsarawa** da wannan umarnin. Magatakarda yana tambayi abubuwa ɗaya bayan ɗaya yana kuma jagora ka.

```bash
wall-vault setup
```

Matakai da magatakarda ke bi sune:

```
1. Zaɓar harshe (harsuna 10 ciki har da Hausa)
2. Zaɓar jigon bayyanar (light / dark / gold / cherry / ocean)
3. Yanayin amfani — shi kaɗai (standalone) ko tare da kwamfutoci da yawa (distributed)
4. Shigar da sunan bot — sunan da za a nuna a dashboard
5. Saita na tashar — tsoho: proxy 56244, ɓaure 56243 (danna Enter kawai idan ba ka son canza)
6. Zaɓar ayyukan AI — na Google / OpenRouter / Ollama
7. Saita tsarin tsaro na kayan aiki
8. Saita token ɗin mai gudanarwa — kalmar sirri don kullewa manyan ayyukan dashboard. Za a iya samar da shi ta atomatik
9. Saita kalmar sirri ta ɓoyewa na maɓallin API — idan kana son ƙarin tsaro (zaɓi ne)
10. Wurin ajiye fayil ɗin tsarawa
```

> ⚠️ **Ka tuna token ɗin mai gudanarwa sosai.** Kana buƙatarsa daga baya don ƙara maɓallan ko canza tsarawa a dashboard. Idan ka manta, dole ne ka gyara fayil ɗin tsarawa kai tsaye.

Bayan kammala magatakarda, fayil ɗin tsarawa `wall-vault.yaml` za a samar da shi ta atomatik.

### Fara

```bash
wall-vault start
```

Uwar garken biyu za su fara lokaci ɗaya:

- **Proxy** (`http://localhost:56244`) — Wakili da ke haɗa OpenClaw da ayyukan AI
- **Ɓaure na Maɓalli** (`http://localhost:56243`) — Sarrafa maɓallin API da dashboard na yanar gizo

Buɗe `http://localhost:56243` a mai bincike naka don duba dashboard nan da nan.

---

## Yin Rajista na Maɓallin API

Akwai hanyoyi huɗu na yin rajista na maɓallin API. **Mun ba da shawarar Hanya ta 1 (masu-canji na yanayi) ga masu farawa.**

### Hanya ta 1: Masu-canji na Yanayi (An ba da shawarar — mafi sauƙi)

Masu-canji na yanayi sune **ƙimomi da aka saita a gaba** da shirin ke karantawa lokacin da ya fara. Ka rubuta a terminal kamar haka:

```bash
# Yin rajista na maɓallin Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Yin rajista na maɓallin OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Fara bayan yin rajista
wall-vault start
```

Idan kana da maɓallan da yawa, haɗa su da waƙafi (,). wall-vault zai yi amfani da su ɗaya bayan ɗaya ta atomatik (round robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Shawarar**: Umarnin `export` yana shafar zaman terminal na yanzu kawai. Don a ci gaba da shi bayan sake farawa kwamfuta, ƙara layin da ke sama zuwa fayil ɗin `~/.bashrc` ko `~/.zshrc`.

### Hanya ta 2: Dashboard UI (Ta Danna da Linzamin Kwamfuta)

1. Shiga `http://localhost:56243` a mai bincike
2. Danna maɓallin `[+ Ƙara]` a katin **🔑 Maɓallin API** na sama
3. Shigar da nau'in aiki, ƙimar maɓalli, lakabi (suna don tunawa), da iyakar yau, sannan adana

### Hanya ta 3: REST API (Don Atomatik da Rubutun Kwamfuta)

REST API hanya ce da shirye-shirye ke musayar bayanai ta hanyar HTTP. Yana da amfani don yin rajista ta atomatik ta hanyar rubutun kwamfuta.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer token_ɗin_mai_gudanarwa" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Maɓalli na Farko",
    "daily_limit": 1000
  }'
```

### Hanya ta 4: Tutar proxy (Don Gwaji na Ɗan Lokaci)

Ana amfani da wannan don saka maɓalli na ɗan lokaci ba tare da yin rajista na hukuma ba. Yana ɓacewa lokacin da shirin ya tsaya.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Yadda Ake Amfani da Proxy

### Amfani a OpenClaw (Babban Manufa)

Hanya ce ta saita OpenClaw don haɗawa da ayyukan AI ta hanyar wall-vault.

Buɗe fayil ɗin `~/.openclaw/openclaw.json` kuma ƙara abubuwan da ke ƙasa:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "token_ɗin_wakilin_naka",   // token na wakili na ɓaure
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // Kyauta 1M context
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Hanya mafi sauƙi**: Danna maɓallin **🦞 Kwafi Tsarawan OpenClaw** da ke kan katin wakili a dashboard, sannan snippet ɗin da ke ƙunshe da token da adireshi za a kwafa zuwa clipboard. Manna kawai.

**`wall-vault/` a gaban sunan ƙira — yana haɗawa zuwa ina?**

wall-vault yana yanke hukunci ta atomatik yadda ake aika buƙata bisa ga sunan ƙirar:

| Tsarin Sunan Ƙira | Aiki da Ake Haɗawa |
|----------|--------------|
| `wall-vault/gemini-*` | Haɗi kai tsaye zuwa Google Gemini |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | Haɗi kai tsaye zuwa OpenAI |
| `wall-vault/claude-*` | Haɗi zuwa Anthropic ta hanyar OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (Kyauta token miliyar 1 na context) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | Haɗi zuwa OpenRouter |
| `google/sunan_ƙira`, `openai/sunan_ƙira`, `anthropic/sunan_ƙira` da sauransu | Haɗi kai tsaye zuwa aiki mai dacewa |
| `custom/google/sunan_ƙira`, `custom/openai/sunan_ƙira` da sauransu | Cire `custom/` sannan sake turawa |
| `sunan_ƙira:cloud` | Cire `:cloud` sannan haɗa zuwa OpenRouter |

> 💡 **Menene Context?** Ita ce adadin tattaunawa da AI zai iya tunawa a lokaci ɗaya. 1M (miliyar token) yana nufin ana iya aika tattaunawa masu tsawo ko takardun da suka yi ƙamu a lokaci ɗaya.

### Haɗi Kai Tsaye ta Tsarin Gemini API (Don Kayan Tsohon Tsari)

Idan kana da kayan aiki da ke amfani kai tsaye da Google Gemini API, canza adireshi kawai zuwa wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Ko kayan aikin da ke ba da damar saka URL kai tsaye:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Amfani da OpenAI SDK (Python)

Za a iya haɗa wall-vault a cikin lambar Python da ke amfani da AI. Canza `base_url` kawai:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # wall-vault yana sarrafa maɓallin API
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # shigar da tsarin: mai_bada_aiki/ƙira
    messages=[{"role": "user", "content": "Sannu"}]
)
```

### Canza Ƙira Yayin Aiki

Don canza ƙirar AI yayin da wall-vault yana gudana:

```bash
# Canza ƙira ta hanyar buƙata kai tsaye zuwa proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# A yanayin rarraba (multi-bot), canza a uwar garken ɓaure → za a yi nuna nan da nan ta SSE
curl -X PUT http://localhost:56243/admin/clients/id_ɓot_naka \
  -H "Authorization: Bearer token_ɗin_mai_gudanarwa" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Duba Jerin Ƙirar da Ake da shi

```bash
# Duba jerin cikakke
curl http://localhost:56244/api/models | python3 -m json.tool

# Duba ƙirar Google kawai
curl "http://localhost:56244/api/models?service=google"

# Bincika da suna (misali: ƙirar da ke ƙunshe da "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Taƙaitaccen ƙirar muhimman ayyuka:**

| Aiki | Muhimman Ƙirar |
|--------|----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | Fiye da 346 (Hunter Alpha 1M context kyauta, DeepSeek R1/V3, Qwen 2.5, da sauransu) |
| Ollama | Gano uwar garken gida da aka shigar a kwamfutarka ta atomatik |

---

## Dashboard na Ɓaure

Shiga `http://localhost:56243` a mai bincike don duba dashboard.

**Yadda ake tsara allon:**
- **Batan Sama (topbar)**: Tambarin kamfani, zaɓin harshe da jigo, nuna halin haɗin SSE
- **Grid na Kati**: Katin wakili, ayyuka, da maɓallin API a cikin fayafai

### Katin Maɓallin API

Kati ne don sarrafa duk maɓallan API da aka yi rajista a wuri ɗaya.

- Yana nuna jerin maɓallan a rarrabe ta kowane aiki.
- `today_usage`: Adadin tokens da aka sarrafa da nasara a yau (haruffa da AI ta karanta da rubuta)
- `today_attempts`: Yawan kira a yau (nasara + gazawa)
- Danna `[+ Ƙara]` don yin rajista na sabon maɓalli, da `✕` don goge maɓalli.

> 💡 **Menene Token?** Naɗi ne da AI ke amfani da shi wajen sarrafa rubutu. Kusan daidai yake da kalma ɗaya ta Hausa ko harafi ɗaya na Ingilishi. Ana lissafin kuɗin API ta adadin token yawanci.

### Katin Wakili

Kati ne da ke nuna halin wakilai (bots) da ke haɗa zuwa proxy na wall-vault.

**Ana nuna halin haɗi a mataki huɗu:**

| Alamar | Hali | Ma'ana |
|------|------|------|
| 🟢 | Yana Gudana | Proxy yana aiki yadda ya kamata |
| 🟡 | Jinkiri | Amsa tana zuwa amma tana jinkiri |
| 🔴 | Ba Kan Layi | Proxy bai amsa ba |
| ⚫ | Ba a Haɗa / Ba a Kunna | Proxy bai taɓa haɗawa da ɓaure ba ko an kashe shi |

**Jagorar maɓallan da ke ƙasan katin wakili:**

Idan ka ƙayyade **nau'in wakili** yayin rajista, maɓallan da suka dace za su bayyana ta atomatik.

---

#### 🔘 Maɓallin Kwafi Tsarawa — Yana Samar da Tsarawa ta Haɗi ta Atomatik

Danna maɓallin yana kwafa snippet ɗin tsarawa wanda ya riga ya ƙunshi token, adireshi na proxy, da bayanan ƙira na wannan wakili zuwa clipboard. Manna kawai a wurin da ke cikin tebur a ƙasa don kammala tsarawa.

| Maɓalli | Nau'in Wakili | Inda Ake Manna |
|------|-------------|-------------|
| 🦞 Kwafi Tsarawan OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Kwafi Tsarawan NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Kwafi Tsarawan Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Kwafi Tsarawan Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Kwafi Tsarawan VSCode | `vscode` | `~/.continue/config.json` |

**Misali — idan nau'in Claude Code, wannan za a kwafa:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.0.6:56244/v1",
  "apiKey": "token_ɗin_wannan_wakili"
}
```

**Misali — idan nau'in VSCode (Continue):**

```json
// ~/.continue/config.json
{
  "models": [{
    "title": "wall-vault proxy",
    "provider": "openai",
    "model": "gemini-2.5-flash",
    "apiBase": "http://192.168.0.6:56244/v1",
    "apiKey": "token_ɗin_wannan_wakili"
  }]
}
```

**Misali — idan nau'in Cursor:**

```
Base URL : http://192.168.0.6:56244/v1
API Key  : token_ɗin_wannan_wakili

// Ko masu-canji na yanayi:
OPENAI_BASE_URL=http://192.168.0.6:56244/v1
OPENAI_API_KEY=token_ɗin_wannan_wakili
```

> ⚠️ **Idan kwafin clipboard bai yi aiki ba**: Wani lokaci manufofin tsaro na mai bincike na iya hana kwafin. Idan akwai akwatin rubutu da ya buɗe, danna Ctrl+A don zaɓar duka sannan Ctrl+C don kwafa.

---

#### ⚡ Maɓallin Saita ta Atomatik — Danna Sau Ɗaya, Tsarawa ta Gama

Idan nau'in wakili shine `cline`, `claude-code`, `openclaw`, `nanoclaw`, maɓallin **⚡ Saita Tsarawa** zai bayyana a kan katin wakili. Danna wannan maɓallin yana sabunta fayil ɗin tsarawa na gida na wannan wakili ta atomatik.

| Maɓalli | Nau'in Wakili | Fayil ɗin da Ake Sabuntawa |
|---------|--------------|---------------------------|
| ⚡ Saita Tsarawan Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Saita Tsarawan Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Saita Tsarawan OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Saita Tsarawan NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Wannan maɓalli yana aika buƙata zuwa **localhost:56244** (proxy na gida). Proxy dole ne ya kasance yana gudana a wannan kwamfutar kafin ya yi aiki.

---

#### 🔀 Ja da Sauka don Tsara Katunan (v0.1.17)

Kana iya **jan** katunan wakili a dashboard don sake tsara su yadda kake so.

1. Danna ka riƙe katun na wakili sannan ka ja shi
2. Saka shi a kan katun a wurin da kake so — tsarin zai canja nan take
3. Sabon tsarin **za a adana shi a uwar garke nan take** kuma zai kasance ko bayan sabunta shafin

> 💡 Na'urorin taɓawa (wayar hannu/kwamfutar hannu) ba su aiki a yanzu ba. Da fatan za a yi amfani da browser na tebur.

---

#### 🔄 Daidaita Ƙira ta Bangarorin Biyu (v0.1.16)

Idan ka canza ƙirar wakili a dashboard ɗin ɓaure, tsarawa na gida na wannan wakili za a sabunta ta atomatik.

**Na Cline:**
- Canza ƙira a ɓaure → lambar SSE → proxy yana sabunta filin ƙira a `globalState.json`
- Filolin da ake sabuntawa: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- Ba a taɓa `openAiBaseUrl` da maɓallin API ba
- **Ana buƙatar sake loda VS Code (`Ctrl+Alt+R` ko `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - Domin Cline ba ya sake karanta fayil ɗin tsarawa yayin da yake gudana

**Na Claude Code:**
- Canza ƙira a ɓaure → lambar SSE → proxy yana sabunta filin `model` a `settings.json`
- Yana bincika hanyoyin WSL da Windows ta atomatik (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Banagaren da ke akasin (wakili → ɓaure):**
- Idan wakili (Cline, Claude Code, da sauransu) ya aika buƙata ta hanyar proxy, proxy yana ƙara bayanan aiki·ƙira na abokin ciniki a heartbeat
- Katin wakili a dashboard ɗin ɓaure yana nuna aiki/ƙira da ake amfani da shi a yanzu a lokaci ɗaya

> 💡 **Muhimmi**: Proxy yana gane wakili ta hanyar token ɗin Authorization a buƙatar, kuma yana tura buƙata zuwa aiki/ƙira da aka saita a ɓaure ta atomatik. Ko Cline ko Claude Code ya aika wani sunan ƙira, proxy zai maye gurbinsa da tsarawan ɓaure.

---

### Amfani da Cline a VS Code — Cikakken Jagora

#### Mataki na 1: Shigar da Cline

Shigar da **Cline** (ID: `saoudrizwan.claude-dev`) daga VS Code Extension Marketplace.

#### Mataki na 2: Yi Rajista na Wakili a Ɓaure

1. Buɗe dashboard ɗin ɓaure (`http://IP_ɗin_ɓaure:56243`)
2. Danna **+ Ƙara** a sashen **wakili**
3. Shigar da bayanan kamar haka:

| Fili | Ƙima | Bayani |
|------|------|--------|
| ID | `cline_na` | Lambar ganewa na musamman (Turanci, ba tazara) |
| Suna | `Cline Na` | Sunan da za a nuna a dashboard |
| Nau'in Wakili | `cline` | ← Dole ne a zaɓi `cline` |
| Aiki | Zaɓi aiki da za a yi amfani da shi (misali: `google`) | |
| Ƙira | Shigar da ƙirar da za a yi amfani da ita (misali: `gemini-2.5-flash`) | |

4. Danna **Adana** sannan za a samar da token ta atomatik

#### Mataki na 3: Haɗa Cline

**Hanya A — Saita ta Atomatik (An Ba da Shawarar)**

1. Tabbata wall-vault **proxy** yana gudana a wannan kwamfutar (`localhost:56244`)
2. Danna maɓallin **⚡ Saita Tsarawan Cline** a katin wakili a dashboard
3. Idan saƙon "An kammala saita!" ya bayyana, an yi nasara
4. Sake loda VS Code (`Ctrl+Alt+R`)

**Hanya B — Saita da Hannu**

Buɗe tsarawa (⚙️) a gefen Cline:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://adireshi_proxy:56244/v1`
  - Kwamfuta ɗaya: `http://localhost:56244/v1`
  - Wata kwamfuta kamar Mac Mini: `http://192.168.0.6:56244/v1`
- **API Key**: Token daga ɓaure (kwafa daga katin wakili)
- **Model ID**: Ƙirar da aka saita a ɓaure (misali: `gemini-2.5-flash`)

#### Mataki na 4: Tabbatarwa

Aika kowace saƙo a tattaunawar Cline. Idan komai yana aiki daidai:
- Katin wakili a dashboard ɗin ɓaure zai nuna **koren ɗigo (● Yana Gudana)**
- Katin zai nuna aiki/ƙira na yanzu (misali: `google / gemini-2.5-flash`)

#### Canza Ƙira

Idan kana son canza ƙirar Cline, canza a **dashboard ɗin ɓaure**:

1. Canza jerin zaɓin aiki/ƙira a katin wakili
2. Danna **Amince**
3. Sake loda VS Code (`Ctrl+Alt+R`) — sunan ƙira a ƙasan Cline zai sabunta
4. Buƙata na gaba za ta yi amfani da sabuwar ƙira

> 💡 A zahiri, proxy yana gane buƙatar Cline ta hanyar token kuma yana tura ta zuwa ƙirar da aka saita a ɓaure. Ko ba a sake loda VS Code ba, **ƙirar da ake amfani da ita za ta canja nan da nan** — sake loda shine kawai don sabunta sunan ƙira a UI na Cline.

#### Gano Yankewar Haɗi

Idan ka rufe VS Code, katin wakili a dashboard ɗin ɓaure zai canza zuwa rawaya (jinkiri) bayan **daƙiƙa 90**, sannan ja (ba kan layi) bayan **minti 3**. (Tun daga v0.1.18, binciken yanayi kowane daƙiƙa 15 yana sa gano rashin haɗi ya yi sauri.)

#### Warware Matsaloli

| Alamar | Dalilin | Hanyar Warwarewa |
|--------|---------|------------------|
| Cline yana nuna "haɗi ya gaza" | Proxy ba ya gudana ko adireshi ba daidai ba | Gwada proxy da `curl http://localhost:56244/health` |
| Koren ɗigo ba ya bayyana a ɓaure | Ba a saita maɓallin API (token) ba | Danna maɓallin **⚡ Saita Tsarawan Cline** sake |
| Sunan ƙira a ƙasan Cline bai canja ba | Cline yana ajiye tsarawa | Sake loda VS Code (`Ctrl+Alt+R`) |
| An nuna sunan ƙira marar daidai | Tsohon matsala (an gyara a v0.1.16) | Sabunta proxy zuwa v0.1.16 ko sama |

---

#### 🟣 Maɓallin Kwafi Umarnin Shigarwa — Ana Amfani Da shi Idan Ana Shigar Da Sabuwar Kwamfuta

Ana amfani da wannan wajen shigar da wall-vault proxy a kwamfuta sabuwa kuma haɗa ta da ɓaure. Danna maɓallin yana kwafa rubutun shigarwa cikakke. Manna shi a terminal na sabuwar kwamfuta kuma gudanar, za a kula da abubuwa masu zuwa a lokaci ɗaya:

1. Shigar da wall-vault binary (za a tsallake idan an riga an shigar)
2. Yin rajista ta atomatik na ayyukan systemd na mai amfani
3. Fara aiki da haɗa zuwa ɓaure ta atomatik

> 💡 Token na wannan wakili da adireshin uwar garken ɓaure an riga an cika su a cikin rubutun, don haka za a iya gudanar da shi nan da nan bayan manna ba tare da gyara ƙarin ba.

---

### Katin Ayyuka

Kati ne don kunna ko kashe, da kuma saita ayyukan AI da za a yi amfani da su.

- Maɓallin canzawa don kunna/kashe kowane aiki
- Idan ka shigar da adireshi na uwar garken AI na gida (Ollama, LM Studio, vLLM da sauransu da ke gudana a kwamfutarka), za a nemo ƙirar da ake da shi ta atomatik.
- **Nuna Halin Haɗin Ayyuka na Gida**: Ɗigo ● kusa da sunan aiki yana **kore** idan an haɗa, **toka** idan ba a haɗa.
- **Daidaita Akwatin Zaɓi ta Atomatik**: Idan ayyukan gida (Ollama, da sauransu) suna gudana lokacin bude shafi, za a saka akwatin zaɓi ta atomatik.

> 💡 **Idan ayyukan gida na gudana a wata kwamfuta**: Shigar da IP ɗin wannan kwamfuta a filin URL na ayyuka. Misali: `http://192.168.0.6:11434` (Ollama), `http://192.168.0.6:1234` (LM Studio)

### Shigar da Token ɗin Mai Gudanarwa

Idan ka yi ƙoƙarin amfani da manyan ayyuka a dashboard kamar ƙara ko goge maɓallan, akwatin shigar da token ɗin mai gudanarwa zai bayyana. Shigar da token ɗin da ka saita a magatakarda na setup. Da zarar an shigar, za a ci gaba da shi har sai ka rufe mai bincike.

> ⚠️ **Idan gazawar tabbatarwa ta wuce sau 10 a cikin minti 15, za a toshe wannan IP na ɗan lokaci.** Idan ka manta token, duba `admin_token` a fayil ɗin `wall-vault.yaml`.

---

## Yanayin Rarraba (Multi-Bot)

Idan kana gudanar da OpenClaw a kwamfutoci da yawa lokaci ɗaya, wannan tsari yana ba ka damar **raba ɓaure maɓalli ɗaya**. Yana da sauƙi domin kana sarrafa maɓallan a wuri ɗaya kawai.

### Misali na Tsarin

```
[Uwar Garken Ɓaure]
  wall-vault vault    (ɓaure :56243, dashboard)

[WSL Alpha]            [Raspberry Pi Gamma]    [Mac Mini Local]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ Daidaita SSE        ↕ Daidaita SSE          ↕ Daidaita SSE
```

Duk bots suna duba uwar garken ɓaure na tsakiya, don haka idan ka canza ƙira ko ƙara maɓalli a ɓaure, za a nuna nan da nan a duk bots.

### Mataki na 1: Fara Uwar Garken Ɓaure

Gudanar da wannan a kwamfutar da za ta zama uwar garken ɓaure:

```bash
wall-vault vault
```

### Mataki na 2: Yin Rajista na Kowane Bot (Abokin Ciniki)

Yi rajista na bayanan kowane bot da zai haɗa zuwa uwar garken ɓaure a gaba:

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer token_ɗin_mai_gudanarwa" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "Bot A",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### Mataki na 3: Fara Proxy a Kowane Kwamfutar Bot

A kowane kwamfutar bot, gudanar da proxy tare da ƙayyade adireshi da token na uwar garken ɓaure:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Maye gurbin **`192.168.x.x`** da ainihin IP ɗin ciki na kwamfutar uwar garken ɓaure. Za a iya duba shi a saitunan router ko da umarnin `ip addr`.

---

## Saita Farawa ta Atomatik

Idan ya gajunka sake kunna wall-vault da hannu bayan kowane sake farawa kwamfuta, yi rajistanshi a matsayin ayyukan tsarin. Da zarar an yi rajista, zai fara ta atomatik a lokacin kunnawa.

### Linux — systemd (Mafi Yawan Linux)

systemd tsarin Linux ne don fara da sarrafa shirye-shirye ta atomatik:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Duba tarihin ayyuka:

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

Tsarin macOS ne da ke kula da farawa ta atomatik na shirye-shirye:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Saukewa NSSM daga [nssm.cc](https://nssm.cc/download) kuma ƙara shi zuwa PATH.
2. A PowerShell na mai gudanarwa:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor — Kayan Bincike

Umarnin `doctor` kayan aiki ne da ke **bincika kansa kuma gyara matsaloli** a wall-vault don tabbatar da an saita shi daidai.

```bash
wall-vault doctor check   # Bincika halin yanzu (karanta kawai, ba a canza komai ba)
wall-vault doctor fix     # Gyara matsaloli ta atomatik
wall-vault doctor all     # Bincike + Gyarawa ta Atomatik tare
```

> 💡 Idan wani abu ya yi kamar ba daidai ba, gwada `wall-vault doctor all` da farko. Yana gyara matsaloli da yawa ta atomatik.

---

## Bayani kan Masu-canji na Yanayi

Masu-canji na yanayi hanya ce ta ba da ƙimomi ga shiri. Za a iya shigar da su a terminal ta `export SUNAN_CANJI=ƙima`, ko a saka su a fayil ɗin ayyukan farawa ta atomatik don a ci gaba da amfani da su.

| Canji | Bayani | Misalin Ƙima |
|------|------|---------|
| `WV_LANG` | Harshen dashboard | `ko`, `en`, `ja` |
| `WV_THEME` | Jigon dashboard | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Maɓallin Google API (da yawa da waƙafi) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Maɓallin OpenRouter API | `sk-or-v1-...` |
| `WV_VAULT_URL` | Adireshi na uwar garken ɓaure a yanayin rarraba | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Token ɗin tabbatarwa na abokin ciniki (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Token ɗin mai gudanarwa | `admin-token-here` |
| `WV_MASTER_PASS` | Kalmar sirri ta ɓoyewa na maɓallin API | `my-password` |
| `WV_AVATAR` | Hanyar fayil ɗin hoton avatar (`~/.openclaw/` dangane da) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Adireshi na uwar garken Ollama na gida | `http://192.168.x.x:11434` |

---

## Warware Matsaloli

### Idan Proxy Bai Fara Ba

Yawanci ana yin amfani da tashar ta wata shirin da ke gudana.

```bash
ss -tlnp | grep 56244   # Duba wanda ke amfani da tashar 56244
wall-vault proxy --port 8080   # Fara da lambar tashar da ta bambanta
```

### Idan Akwai Kuskure na Maɓallin API (429, 402, 401, 403, 582)

| Lambar Kuskure | Ma'ana | Yadda Ake Warwarewa |
|----------|------|----------|
| **429** | Buƙatu da yawa (amfani ya wuce iyaka) | Jira ɗan lokaci ko ƙara wani maɓalli |
| **402** | Buƙatar biyan kuɗi ko bashi ya ƙare | Caji bashi a wannan aiki |
| **401 / 403** | Maɓalli ba daidai ba ko babu izini | Sake duba ƙimar maɓalli sannan sake yi rajista |
| **582** | Babbar matsin hanyar shiga (cooldown minti 5) | Za a saki ta atomatik bayan minti 5 |

```bash
# Duba jerin maɓallan da aka yi rajista da halinsu
curl -H "Authorization: Bearer token_ɗin_mai_gudanarwa" http://localhost:56243/admin/keys

# Sake saita mai ƙididdiga na amfanin maɓalli
curl -X POST -H "Authorization: Bearer token_ɗin_mai_gudanarwa" http://localhost:56243/admin/keys/reset
```

### Idan Wakili Ya Nuna "Ba a Haɗa"

"Ba a Haɗa" yana nufin tsarin proxy bai aika siginar zuciya (heartbeat) zuwa ɓaure ba. **Ba yana nufin ba a adana tsarawa ba.** Proxy dole ne ya san adireshi da token na uwar garken ɓaure don ya zama haɗi.

```bash
# Fara proxy tare da ƙayyade adireshi, token, da ID na abokin ciniki na ɓaure
WV_VAULT_URL=http://adireshin_uwar_garken_baure:56243 \
WV_VAULT_TOKEN=token_ɗin_abokin_ciniki \
WV_VAULT_CLIENT_ID=id_ɗin_abokin_ciniki \
wall-vault proxy
```

Idan haɗi ya yi nasara, za a canza zuwa 🟢 Yana Gudana a dashboard cikin kusan dakika 20.

### Idan Haɗin Ollama Bai Yi Aiki Ba

Ollama shiri ne da ke gudanar da AI kai tsaye a kwamfutarka. Da farko, tabbata Ollama yana kunna.

```bash
curl http://localhost:11434/api/tags   # Idan an nuna jerin ƙirar, yana aiki daidai
export OLLAMA_URL=http://192.168.x.x:11434   # Idan yana gudana a wata kwamfuta
```

> ⚠️ Idan Ollama bai amsa ba, fara Ollama da farko da umarnin `ollama serve`.

> ⚠️ **Ƙirar manyan AI suna da jinkiri**: Ƙirar da suka yi ƙamu kamar `qwen3.5:35b` ko `deepseek-r1` na iya ɗaukar mintuna da yawa don samar da amsa. Idan ya zama kamar babu amsa, yana iya zama yana aiki daidai — jira kawai.

---

*Don ƙarin bayani na API, duba [API.md](API.md).*
