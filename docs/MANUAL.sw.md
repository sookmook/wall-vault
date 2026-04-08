# Mwongozo wa Mtumiaji wa wall-vault
*(Last updated: 2026-04-08 — v0.1.25)*

---

## Yaliyomo

1. [wall-vault ni nini?](#wall-vault-ni-nini)
2. [Ufungaji](#ufungaji)
3. [Kuanza kwa Mara ya Kwanza (mchawi wa setup)](#kuanza-kwa-mara-ya-kwanza)
4. [Kusajili Ufunguo wa API](#kusajili-ufunguo-wa-api)
5. [Jinsi ya Kutumia Wakala (Proxy)](#jinsi-ya-kutumia-wakala-proxy)
6. [Dashibodi ya Ghala la Funguo](#dashibodi-ya-ghala-la-funguo)
7. [Hali ya Kusambazwa (Boti Nyingi)](#hali-ya-kusambazwa-boti-nyingi)
8. [Kuanzisha Kiotomatiki](#kuanzisha-kiotomatiki)
9. [Daktari (Doctor)](#daktari-doctor)
10. [RTK Kuokoa Tokeni](#rtk-kuokoa-tokeni)
11. [Marejeo ya Vigeuzi vya Mazingira](#marejeo-ya-vigeuzi-vya-mazingira)
12. [Utatuzi wa Matatizo](#utatuzi-wa-matatizo)

---

## wall-vault ni nini?

**wall-vault = Wakala (Proxy) wa AI + Ghala la Funguo za API kwa ajili ya OpenClaw**

Ili kutumia huduma za AI, unahitaji **ufunguo wa API**. Ufunguo wa API ni kama **kitambulisho cha kidijitali** kinachothibitisha kwamba "mtu huyu ana ruhusa ya kutumia huduma hii". Hata hivyo, kitambulisho hiki kina kikomo cha matumizi ya kila siku, na ikiwa hakidhibitiwi vizuri, kinaweza kufichuliwa.

wall-vault huhifadhi vitambulisho hivi katika ghala salama na kufanya kazi kama **wakala (proxy)** kati ya OpenClaw na huduma za AI. Kwa maneno rahisi, OpenClaw inahitaji tu kuunganishwa na wall-vault, na wall-vault itashughulikia mambo mengine yote.

Matatizo ambayo wall-vault inasuluhisha:

- **Mzunguko wa Kiotomatiki wa Funguo za API**: Ufunguo mmoja ukifikia kikomo au ukizuiliwa kwa muda (cooldown), inabadilika kimya kimya hadi ufunguo unaofuata. OpenClaw inaendelea kufanya kazi bila kukatika.
- **Ubadilishaji wa Kiotomatiki wa Huduma (Fallback)**: Google ikiwa haijibu, inabadilika hadi OpenRouter, na ikishindikana, inabadilika hadi Ollama/LM Studio/vLLM (AI ya ndani) iliyosanikishwa kwenye kompyuta yako. Kipindi hakikatiki. Huduma ya awali inapoanza tena, maombi yanayofuata yanarudi kiotomatiki (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Usawazishaji wa Papo Hapo (SSE)**: Ukibadilisha modeli kwenye dashibodi ya ghala, inajitokeza kwenye skrini ya OpenClaw ndani ya sekunde 1-3. SSE (Server-Sent Events) ni teknolojia ambapo seva inasukuma mabadiliko kwa wateja kwa wakati halisi.
- **Arifa za Papo Hapo**: Matukio kama funguo kuisha au huduma kushindwa yanaonyeshwa mara moja chini ya TUI (skrini ya terminali) ya OpenClaw.

> 💡 **Claude Code, Cursor, VS Code** pia zinaweza kuunganishwa, lakini madhumuni ya asili ya wall-vault ni kutumika na OpenClaw.

```
OpenClaw (TUI skrini ya terminali)
        │
        ▼
  wall-vault proxy (:56244)   ← usimamizi wa funguo, uelekezaji, fallback, matukio
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (modeli 340+)
        ├─ Ollama / LM Studio / vLLM (kompyuta yako, kimbilio la mwisho)
        └─ OpenAI / Anthropic API
```

---

## Ufungaji

### Linux / macOS

Fungua terminali na ubandike amri zilizo hapa chini.

```bash
# Linux (PC ya kawaida, seva — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Inapakua faili kutoka mtandaoni.
- `chmod +x` — Inafanya faili iliyopakuliwa "iweze kutekelezwa". Ukiacha hatua hii, utapata kosa la "ruhusa imekataliwa".

### Windows

Fungua PowerShell (kama msimamizi) na utekeleze amri zifuatazo.

```powershell
# Pakua
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Ongeza kwenye PATH (inatumika baada ya PowerShell kuanzishwa upya)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **PATH ni nini?** Ni orodha ya folda ambapo kompyuta inatafuta amri. Unahitaji kuongeza kwenye PATH ili uweze kuandika `wall-vault` na kuitekeleza kutoka folda yoyote.

### Kujenga kutoka kwa Msimbo wa Chanzo (kwa watengenezaji)

Hii inatumika tu ikiwa una mazingira ya maendeleo ya lugha ya Go yaliyosanikishwa.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (toleo: v0.1.25.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Toleo la muhuri wa wakati wa kujenga**: Ukijenga na `make build`, toleo litaundwa kiotomatiki katika muundo unaojumuisha tarehe na wakati kama `v0.1.25.20260408.022325`. Ukijenga moja kwa moja na `go build ./...`, toleo litaonyeshwa tu kama `"dev"`.

---

## Kuanza kwa Mara ya Kwanza

### Kuendesha Mchawi wa Setup

Baada ya ufungaji, hakikisha unatekeleza **mchawi wa usanidi** kwa amri ifuatayo. Mchawi utakuuliza maswali moja kwa moja na kukuongoza.

```bash
wall-vault setup
```

Hatua ambazo mchawi anafuata ni hizi:

```
1. Chagua lugha (lugha 10 ikiwa ni pamoja na Kiswahili)
2. Chagua mandhari (light / dark / gold / cherry / ocean)
3. Hali ya uendeshaji — chagua kama utatumia peke yako (standalone) au kwenye mashine nyingi (distributed)
4. Ingiza jina la boti — jina litakaloonyeshwa kwenye dashibodi
5. Mipangilio ya bandari — chaguomsingi: proxy 56244, ghala 56243 (bonyeza Enter kama huhitaji kubadilisha)
6. Chagua huduma za AI — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Mipangilio ya kichujio cha zana za usalama
8. Weka tokeni ya msimamizi — nenosiri la kufunga vipengele vya usimamizi wa dashibodi. Unaweza pia kuizalisha kiotomatiki
9. Weka nenosiri la usimbuaji wa ufunguo wa API — ikiwa unataka kuhifadhi funguo kwa usalama zaidi (hiari)
10. Njia ya kuhifadhi faili ya usanidi
```

> ⚠️ **Hakikisha unaikumbuka tokeni ya msimamizi.** Utaihitaji baadaye unapoongeza funguo au kubadilisha mipangilio kwenye dashibodi. Ukiipoteza, utahitaji kuhariri faili ya usanidi moja kwa moja.

Baada ya mchawi kukamilika, faili ya usanidi `wall-vault.yaml` itaundwa kiotomatiki.

### Kuendesha

```bash
wall-vault start
```

Seva mbili zitaanza kwa wakati mmoja:

- **Proxy** (`http://localhost:56244`) — wakala anayeunganisha OpenClaw na huduma za AI
- **Ghala la Funguo** (`http://localhost:56243`) — usimamizi wa ufunguo wa API na dashibodi ya wavuti

Fungua `http://localhost:56243` kwenye kivinjari chako ili kuona dashibodi mara moja.

---

## Kusajili Ufunguo wa API

Kuna njia nne za kusajili ufunguo wa API. **Kwa wanaoanza, Njia ya 1 (vigeuzi vya mazingira) inapendekezwa**.

### Njia ya 1: Vigeuzi vya Mazingira (Inayopendekezwa — Rahisi Zaidi)

Vigeuzi vya mazingira ni **thamani zilizowekwa mapema** ambazo programu inasoma inapoanza. Andika katika terminali kama ifuatavyo.

```bash
# Sajili ufunguo wa Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Sajili ufunguo wa OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Endesha baada ya kusajili
wall-vault start
```

Ikiwa una funguo nyingi, ziunganishe kwa koma (,). wall-vault itatumia funguo kwa zamu kiotomatiki (round robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Kidokezo**: Amri ya `export` inatumika tu kwenye kipindi cha sasa cha terminali. Ili kudumu hata baada ya kuanzisha upya kompyuta, ongeza mstari huo kwenye faili ya `~/.bashrc` au `~/.zshrc`.

### Njia ya 2: Dashibodi ya UI (Bonyeza kwa Kipanya)

1. Tembelea `http://localhost:56243` kwenye kivinjari
2. Kwenye kadi ya **🔑 Funguo za API** hapo juu, bonyeza kitufe cha `[+ Ongeza]`
3. Ingiza aina ya huduma, thamani ya ufunguo, lebo (jina la kumbukumbu), na kikomo cha kila siku, kisha uhifadhi

### Njia ya 3: REST API (kwa Otomatiki/Hati)

REST API ni njia ambapo programu hubadilishana data kupitia HTTP. Ni muhimu kwa kusajili kiotomatiki kupitia hati.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer tokeni-ya-msimamizi" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Ufunguo Mkuu",
    "daily_limit": 1000
  }'
```

### Njia ya 4: Bendera za Proxy (kwa Majaribio ya Muda Mfupi)

Inatumika wakati unataka kujaribu kwa muda bila usajili rasmi. Inapotea unapozima programu.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Jinsi ya Kutumia Wakala (Proxy)

### Matumizi na OpenClaw (Madhumuni Makuu)

Jinsi ya kusanidi OpenClaw ili iunganishwe na huduma za AI kupitia wall-vault.

Fungua faili ya `~/.openclaw/openclaw.json` na uongeze yaliyomo yafuatayo:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // tokeni ya wakala wa vault
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // muktadha wa 1M bila malipo
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Njia Rahisi Zaidi**: Bonyeza kitufe cha **🦞 Nakili Usanidi wa OpenClaw** kwenye kadi ya wakala wa dashibodi na snippet yenye tokeni na anwani zilizojazwa tayari itanakiliwa kwenye ubao wa kunakili. Bandika tu.

**`wall-vault/` kabla ya jina la modeli inaelekezwa wapi?**

wall-vault inaamua kiotomatiki huduma ipi ya AI itumie kulingana na jina la modeli:

| Muundo wa Modeli | Huduma Inayounganishwa |
|----------|--------------|
| `wall-vault/gemini-*` | Muunganisho wa moja kwa moja na Google Gemini |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | Muunganisho wa moja kwa moja na OpenAI |
| `wall-vault/claude-*` | Muunganisho na Anthropic kupitia OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (tokeni milioni 1 za muktadha bila malipo) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | Muunganisho na OpenRouter |
| `google/jina-la-modeli`, `openai/jina-la-modeli`, `anthropic/jina-la-modeli` n.k. | Muunganisho wa moja kwa moja na huduma husika |
| `custom/google/jina-la-modeli`, `custom/openai/jina-la-modeli` n.k. | Sehemu ya `custom/` inaondolewa na kuelekeza upya |
| `jina-la-modeli:cloud` | Sehemu ya `:cloud` inaondolewa na kuunganishwa na OpenRouter |

> 💡 **Muktadha (context) ni nini?** Ni kiasi cha mazungumzo ambacho AI inaweza kukumbuka kwa wakati mmoja. 1M (tokeni milioni moja) inamaanisha mazungumzo marefu sana au hati ndefu zinaweza kushughulikiwa kwa wakati mmoja.

### Muunganisho wa Moja kwa Moja kwa Muundo wa Gemini API (utangamano na zana zilizopo)

Ikiwa una zana zilizokuwa zikitumia Google Gemini API moja kwa moja, badilisha anwani tu hadi wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Au ikiwa zana yako inabainisha URL moja kwa moja:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Matumizi na OpenAI SDK (Python)

Unaweza pia kuunganisha wall-vault katika msimbo wa Python unaotumia AI. Badilisha tu `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # wall-vault inasimamia funguo za API kiotomatiki
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # muundo wa mtoa-huduma/modeli
    messages=[{"role": "user", "content": "Habari"}]
)
```

### Kubadilisha Modeli Wakati wa Uendeshaji

Ili kubadilisha modeli ya AI wakati wall-vault tayari inaendesha:

```bash
# Badilisha modeli kwa kuomba moja kwa moja kwa proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# Katika hali ya kusambazwa (boti nyingi), badilisha kwenye seva ya ghala → inaakisiwa mara moja kupitia SSE
curl -X PUT http://localhost:56243/admin/clients/kitambulisho-cha-boti \
  -H "Authorization: Bearer tokeni-ya-msimamizi" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Kuangalia Orodha ya Modeli Zinazopatikana

```bash
# Tazama orodha kamili
curl http://localhost:56244/api/models | python3 -m json.tool

# Tazama modeli za Google tu
curl "http://localhost:56244/api/models?service=google"

# Tafuta kwa jina (mfano: modeli zinazojumuisha "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Muhtasari wa Modeli Kuu kwa Huduma:**

| Huduma | Modeli Kuu |
|--------|----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ (Hunter Alpha 1M muktadha bila malipo, DeepSeek R1/V3, Qwen 2.5 n.k.) |
| Ollama | Inagundua seva ya ndani iliyosanikishwa kwenye kompyuta yako kiotomatiki |
| LM Studio | Seva ya ndani kwenye kompyuta yako (bandari 1234) |
| vLLM | Seva ya ndani kwenye kompyuta yako (bandari 8000) |

---

## Dashibodi ya Ghala la Funguo

Tembelea `http://localhost:56243` kwenye kivinjari ili kuona dashibodi.

**Mpangilio wa Skrini:**
- **Baa ya juu iliyoimarishwa (topbar)**: Nembo, kichaguzi cha lugha na mandhari, hali ya muunganisho wa SSE
- **Gridi ya kadi**: Kadi za wakala, huduma, na ufunguo wa API zilizopangwa kama vigae

### Kadi ya Ufunguo wa API

Kadi inayokuruhusu kusimamia funguo za API zilizosajiliwa kwa mtazamo mmoja.

- Inaonyesha orodha ya funguo zikigawanywa kwa huduma.
- `today_usage`: Idadi ya tokeni zilizofanikiwa kushughulikiwa leo (idadi ya herufi AI ilizosoma na kuandika)
- `today_attempts`: Jumla ya simu za leo (mafanikio + kushindwa)
- Kitufe cha `[+ Ongeza]` kusajili ufunguo mpya, na `✕` kufuta ufunguo.

> 💡 **Tokeni ni nini?** Ni kitengo ambacho AI inatumia kusindika maandishi. Takriban ni sawa na neno moja la Kiingereza, au herufi 1-2 za Kikorea. Malipo ya API kwa kawaida yanahesabiwa kulingana na idadi hii ya tokeni.

### Kadi ya Wakala

Kadi inayoonyesha hali ya boti (mawakala) waliounganishwa na proxy ya wall-vault.

**Hali ya muunganisho inaonyeshwa katika hatua 4:**

| Onyesho | Hali | Maana |
|------|------|------|
| 🟢 | Inaendesha | Proxy inafanya kazi kawaida |
| 🟡 | Imechelewa | Inajibu lakini polepole |
| 🔴 | Nje ya Mtandao | Proxy haijibu |
| ⚫ | Haijaungangishwa/Imezimwa | Proxy haijawahi kuunganishwa na ghala au imezimwa |

**Mwongozo wa Vitufe Chini ya Kadi ya Wakala:**

Unaposajili wakala na kubainisha **aina ya wakala**, vitufe vya urahisi vilivyokusudiwa aina hiyo vinaonekana kiotomatiki.

---

#### 🔘 Kitufe cha Kunakili Usanidi — Kinaunda mipangilio ya muunganisho kiotomatiki

Unapobonyeza kitufe, snippet ya usanidi yenye tokeni ya wakala huyo, anwani ya proxy, na taarifa za modeli zilizojazwa tayari inanakiliwa kwenye ubao wa kunakili. Bandika tu maudhui yaliyonakiliwa kwenye eneo lililoonyeshwa kwenye jedwali hapa chini ili kukamilisha usanidi wa muunganisho.

| Kitufe | Aina ya Wakala | Mahali pa Kubandika |
|------|-------------|-------------|
| 🦞 Nakili Usanidi wa OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Nakili Usanidi wa NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Nakili Usanidi wa Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Nakili Usanidi wa Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Nakili Usanidi wa VSCode | `vscode` | `~/.continue/config.json` |

**Mfano — Ikiwa ni aina ya Claude Code, yafuatayo yatanakiliwa:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "tokeni-ya-wakala-huyu"
}
```

**Mfano — Ikiwa ni aina ya VSCode (Continue):**

```yaml
# ~/.continue/config.yaml  ← bandika kwenye config.yaml, si config.json
name: My Config
version: 0.0.1
schema: v1

models:
  - name: wall-vault proxy
    provider: openai
    model: gemini-2.5-flash
    apiBase: http://192.168.1.20:56244/v1
    apiKey: tokeni-ya-wakala-huyu
    roles:
      - chat
      - edit
      - apply
```

> ⚠️ **Toleo jipya la Continue linatumia `config.yaml`.** Ikiwa `config.yaml` ipo, `config.json` itapuuzwa kabisa. Hakikisha unabandika kwenye `config.yaml`.

**Mfano — Ikiwa ni aina ya Cursor:**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : tokeni-ya-wakala-huyu

// Au vigeuzi vya mazingira:
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=tokeni-ya-wakala-huyu
```

> ⚠️ **Kunakili kwenye ubao wa kunakili kukishindikana**: Sera ya usalama ya kivinjari inaweza kuzuia kunakili. Ikiwa sanduku la maandishi litafunguka kama popup, tumia Ctrl+A kuchagua yote kisha Ctrl+C kunakili.

---

#### ⚡ Kitufe cha Kutumia Kiotomatiki — Bonyeza mara moja na usanidi umekamilika

Ikiwa aina ya wakala ni `cline`, `claude-code`, `openclaw`, au `nanoclaw`, kitufe cha **⚡ Tumia Usanidi** kinaonyeshwa kwenye kadi ya wakala. Unapobonyeza kitufe hiki, faili ya usanidi ya ndani ya wakala huyo inasasishwa kiotomatiki.

| Kitufe | Aina ya Wakala | Faili Inayolengwa |
|------|-------------|-------------|
| ⚡ Tumia Usanidi wa Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Tumia Usanidi wa Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Tumia Usanidi wa OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Tumia Usanidi wa NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Kitufe hiki kinatuma ombi kwa **localhost:56244** (proxy ya ndani). Proxy lazima iwe inaendesha kwenye mashine hiyo ili ifanye kazi.

---

#### 🔀 Kupanga Kadi kwa Kuburuta na Kudondosha (v0.1.17, iliyoboreshwa v0.1.25)

Unaweza **kuburuta** kadi za wakala kwenye dashibodi ili kuzipanga upya kwa mpangilio unaotaka.

1. Shika eneo la **taa za trafiki (●)** upande wa kushoto juu wa kadi na uburute kwa kipanya
2. Iache kwenye kadi unayotaka na mpangilio utabadilika

> 💡 Maudhui ya kadi (sehemu za kuingiza, vitufe n.k.) hayaburutwi. Unaweza kushika tu kutoka eneo la taa za trafiki.

#### 🟠 Kugundua Mchakato wa Wakala (v0.1.25)

Wakati proxy inafanya kazi vizuri lakini mchakato wa wakala wa ndani (NanoClaw, OpenClaw) umekufa, taa ya kadi inabadilika kuwa **rangi ya machungwa (inapepesa)** na ujumbe "Mchakato wa wakala umesimama" unaonyeshwa.

- 🟢 Kijani: Proxy + wakala vinafanya kazi vizuri
- 🟠 Machungwa (inapepesa): Proxy inafanya kazi vizuri, wakala amekufa
- 🔴 Nyekundu: Proxy nje ya mtandao
3. Mpangilio uliobadilishwa **unahifadhiwa kwenye seva mara moja** na unadumu hata baada ya kuburudisha

> 💡 Kwenye vifaa vya kugusa (simu/kibao bapa) bado hakutumiki. Tumia kivinjari cha kompyuta.

---

#### 🔄 Usawazishaji wa Modeli wa Pande Mbili (v0.1.16)

Ukibadilisha modeli ya wakala kwenye dashibodi ya ghala, usanidi wa ndani wa wakala huyo unasasishwa kiotomatiki.

**Kwa Cline:**
- Modeli ikibadilishwa kwenye ghala → Tukio la SSE → Proxy inasasisha sehemu ya modeli katika `globalState.json`
- Malengo ya kusasishwa: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` na ufunguo wa API haviathiriwi
- **Kupakia upya VS Code (`Ctrl+Alt+R` au `Ctrl+Shift+P` → `Developer: Reload Window`) kunahitajika**
  - Kwa sababu Cline haisomi faili ya usanidi tena inapoendesha

**Kwa Claude Code:**
- Modeli ikibadilishwa kwenye ghala → Tukio la SSE → Proxy inasasisha sehemu ya `model` katika `settings.json`
- Inatafuta kiotomatiki njia za WSL na Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Mwelekeo wa Nyuma (wakala → ghala):**
- Wakala (Cline, Claude Code n.k.) anapotuma ombi kwa proxy, proxy inajumuisha taarifa za huduma/modeli za mteja katika heartbeat
- Huduma/modeli inayotumika sasa inaonyeshwa kwa wakati halisi kwenye kadi ya wakala kwenye dashibodi ya ghala

> 💡 **Jambo Muhimu**: Proxy inatambua wakala kupitia tokeni ya Authorization ya ombi, na kuelekeza kiotomatiki kwa huduma/modeli iliyowekwa kwenye ghala. Hata Cline au Claude Code ikituma jina tofauti la modeli, proxy inabatilisha na usanidi wa ghala.

---

### Kutumia Cline kwenye VS Code — Mwongozo wa Kina

#### Hatua ya 1: Sanikisha Cline

Sanikisha **Cline** (ID: `saoudrizwan.claude-dev`) kutoka soko la viendelezi vya VS Code.

#### Hatua ya 2: Sajili Wakala kwenye Ghala

1. Fungua dashibodi ya ghala (`http://IP-ya-ghala:56243`)
2. Katika sehemu ya **Mawakala**, bonyeza **+ Ongeza**
3. Ingiza yafuatayo:

| Sehemu | Thamani | Maelezo |
|------|----|------|
| ID | `yangu_cline` | Kitambulisho cha kipekee (Kiingereza, bila nafasi) |
| Jina | `Cline Yangu` | Jina litakaloonyeshwa kwenye dashibodi |
| Aina ya Wakala | `cline` | ← Lazima uchague `cline` |
| Huduma | Chagua huduma ya kutumia (mfano: `google`) | |
| Modeli | Ingiza modeli ya kutumia (mfano: `gemini-2.5-flash`) | |

4. Bonyeza **Hifadhi** na tokeni itaundwa kiotomatiki

#### Hatua ya 3: Unganisha na Cline

**Njia A — Kutumia Kiotomatiki (Inayopendekezwa)**

1. Hakikisha **proxy** ya wall-vault inaendesha kwenye mashine hiyo (`localhost:56244`)
2. Bonyeza kitufe cha **⚡ Tumia Usanidi wa Cline** kwenye kadi ya wakala kwenye dashibodi
3. Arifa "Usanidi umetumika!" ikionekana, imefanikiwa
4. Pakia upya VS Code (`Ctrl+Alt+R`)

**Njia B — Usanidi wa Mikono**

Fungua mipangilio (⚙️) kwenye upau wa pembeni wa Cline na uweke:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://anwani-ya-proxy:56244/v1`
  - Kwenye mashine ile ile: `http://localhost:56244/v1`
  - Kwenye mashine nyingine kama seva ya Mini: `http://192.168.1.20:56244/v1`
- **API Key**: Tokeni iliyotolewa kutoka ghala (nakili kutoka kadi ya wakala)
- **Model ID**: Modeli iliyowekwa kwenye ghala (mfano: `gemini-2.5-flash`)

#### Hatua ya 4: Thibitisha

Tuma ujumbe wowote kwenye mazungumzo ya Cline. Ikiwa kila kitu ni sawa:
- **Alama ya kijani (● Inaendesha)** inaonyeshwa kwenye kadi ya wakala kwenye dashibodi ya ghala
- Huduma/modeli ya sasa inaonyeshwa kwenye kadi (mfano: `google / gemini-2.5-flash`)

#### Kubadilisha Modeli

Unapotaka kubadilisha modeli ya Cline, ibadilishe kwenye **dashibodi ya ghala**:

1. Badilisha kichupo cha huduma/modeli kwenye kadi ya wakala
2. Bonyeza **Tumia**
3. Pakia upya VS Code (`Ctrl+Alt+R`) — jina la modeli katika kijachini cha Cline litasasishwa
4. Modeli mpya itatumika kuanzia ombi linalofuata

> 💡 Kwa kweli, proxy inatambua ombi la Cline kupitia tokeni na kuelekeza kwa modeli ya usanidi wa ghala. Hata bila kupakia upya VS Code, **modeli inayotumika kweli inabadilika mara moja** — kupakia upya ni kusasisha onyesho la modeli kwenye UI ya Cline tu.

#### Kugundua Kukatika kwa Muunganisho

Unapofunga VS Code, kadi ya wakala kwenye dashibodi ya ghala inabadilika kuwa njano (imechelewa) baada ya takriban **sekunde 90**, na kuwa nyekundu (nje ya mtandao) baada ya **dakika 3**. (Kuanzia v0.1.18, ukaguzi wa hali kwa vipindi vya sekunde 15 unafanya ugunduzi wa nje ya mtandao kuwa wa haraka zaidi.)

#### Utatuzi wa Matatizo

| Dalili | Sababu | Suluhisho |
|------|------|------|
| Kosa la "muunganisho umeshindikana" kwenye Cline | Proxy haiendesha au anwani si sahihi | Thibitisha proxy kwa `curl http://localhost:56244/health` |
| Alama ya kijani haionekani kwenye ghala | Ufunguo wa API (tokeni) haujawekwa | Bonyeza kitufe cha **⚡ Tumia Usanidi wa Cline** tena |
| Modeli kwenye kijachini cha Cline haibadiliki | Cline ina akiba ya usanidi | Pakia upya VS Code (`Ctrl+Alt+R`) |
| Jina la modeli lisilo sahihi linaonyeshwa | Hitilafu ya zamani (ilisahihishwa katika v0.1.16) | Sasisha proxy hadi v0.1.16 au zaidi |

---

#### 🟣 Kitufe cha Kunakili Amri ya Kupeleka — Kinatumika wakati wa kusakinisha kwenye mashine mpya

Kinatumika wakati wa kusakinisha proxy ya wall-vault kwa mara ya kwanza kwenye kompyuta mpya na kuiunganisha na ghala. Ukibonyeza kitufe, hati nzima ya usakinishaji inanakiliwa. Ibandike kwenye terminali ya kompyuta mpya na uitekeleze, na yafuatayo yatashughulikiwa kwa wakati mmoja:

1. Sakinisha binary ya wall-vault (inarukwa ikiwa tayari imesanikishwa)
2. Sajili huduma ya mtumiaji ya systemd kiotomatiki
3. Anzisha huduma na uunganishe na ghala kiotomatiki

> 💡 Hati tayari ina tokeni ya wakala huyu na anwani ya seva ya ghala zilizojazwa, kwa hivyo unaweza kuitekeleza moja kwa moja baada ya kubandika bila marekebisho yoyote.

---

### Kadi ya Huduma

Kadi ya kuwasha/kuzima au kusanidi huduma za AI za kutumia.

- Swichi za kubadilisha kuwezesha/kuzima kwa kila huduma
- Ukiingiza anwani ya seva ya AI ya ndani (Ollama, LM Studio, vLLM n.k. inayoendesha kwenye kompyuta yako), modeli zinazopatikana zitagundulika kiotomatiki.
- **Onyesho la hali ya muunganisho wa huduma ya ndani**: Alama ya ● karibu na jina la huduma ni **kijani** ikiwa imeunganishwa, **kijivu** ikiwa haijaunganishwa
- **Taa za trafiki za kiotomatiki za huduma ya ndani** (v0.1.23+): Huduma za ndani (Ollama, LM Studio, vLLM) zinawezeshwa kiotomatiki zinapoweza kuunganishwa, na kuzimwa zinapokatika. Ukiwasha huduma, inabadilika kuwa ● kijani ndani ya sekunde 15 na kisanduku cha kuangalia kinawaka, na ukiizima, inazimwa kiotomatiki. Hii inafanya kazi sawa na jinsi huduma za wingu (Google, OpenRouter n.k.) zinavyobadilika kiotomatiki kulingana na uwepo wa ufunguo wa API.

> 💡 **Ikiwa huduma ya ndani inaendesha kwenye kompyuta nyingine**: Ingiza IP ya kompyuta hiyo kwenye sehemu ya kuingiza URL ya huduma. Mfano: `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio). Ikiwa huduma imefungwa kwa `127.0.0.1` tu badala ya `0.0.0.0`, haitaweza kufikiwa kupitia IP ya nje, kwa hivyo angalia anwani ya kufunga kwenye usanidi wa huduma.

### Kuingiza Tokeni ya Msimamizi

Unapojaribu kutumia vipengele muhimu kama kuongeza au kufuta funguo kwenye dashibodi, popup ya kuingiza tokeni ya msimamizi itaonekana. Ingiza tokeni uliyoiweka kwenye mchawi wa setup. Mara tu ukiiingiza, inadumu hadi ufunge kivinjari.

> ⚠️ **Ikiwa uthibitishaji utashindikana mara zaidi ya 10 ndani ya dakika 15, IP hiyo itazuiliwa kwa muda.** Ikiwa umesahau tokeni yako, angalia kipengee cha `admin_token` kwenye faili ya `wall-vault.yaml`.

---

## Hali ya Kusambazwa (Boti Nyingi)

Usanidi wa **kushiriki ghala moja ya funguo** wakati unaendesha OpenClaw kwenye kompyuta nyingi kwa wakati mmoja. Ni rahisi kwa sababu unahitaji tu kusimamia funguo katika sehemu moja.

### Mfano wa Usanidi

```
[Seva ya Ghala la Funguo]
  wall-vault vault    (ghala la funguo :56243, dashibodi)

[WSL Alpha]            [Raspberry Pi Gamma]    [Mac Mini Ndani]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ Usawazishaji wa SSE  ↕ Usawazishaji wa SSE    ↕ Usawazishaji wa SSE
```

Boti zote zinaangalia seva ya ghala katikati, kwa hivyo ukibadilisha modeli au kuongeza funguo kwenye ghala, inaakisiwa kwa boti zote mara moja.

### Hatua ya 1: Anzisha Seva ya Ghala la Funguo

Endesha kwenye kompyuta utakayoitumia kama seva ya ghala:

```bash
wall-vault vault
```

### Hatua ya 2: Sajili Kila Boti (Mteja)

Sajili taarifa za kila boti inayounganishwa na seva ya ghala mapema:

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer tokeni-ya-msimamizi" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "BotiA",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### Hatua ya 3: Anzisha Proxy kwenye Kila Kompyuta ya Boti

Endesha proxy kwenye kila kompyuta ambayo boti imesanikishwa, ukibainisha anwani ya seva ya ghala na tokeni:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Badilisha sehemu ya **`192.168.x.x`** na anwani halisi ya IP ya ndani ya kompyuta ya seva ya ghala. Unaweza kuithibitisha kupitia mipangilio ya router au amri ya `ip addr`.

---

## Kuanzisha Kiotomatiki

Ikiwa ni usumbufu kuwasha wall-vault kwa mkono kila wakati unapoanzisha upya kompyuta, isajili kama huduma ya mfumo. Mara tu ikiwa imesajiliwa, itaanza kiotomatiki wakati wa kupakia.

### Linux — systemd (Linux nyingi)

systemd ni mfumo unaowasha na kusimamia programu kiotomatiki kwenye Linux:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Angalia kumbukumbu:

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

Mfumo unaosimamia uanzishaji wa kiotomatiki wa programu kwenye macOS:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Pakua NSSM kutoka [nssm.cc](https://nssm.cc/download) na uiongeze kwenye PATH.
2. Katika PowerShell ya msimamizi:

```powershell
wall-vault doctor deploy windows
```

---

## Daktari (Doctor)

Amri ya `doctor` ni **zana inayojichunguza na kujirekebisha** kuangalia kama wall-vault imesanidiwa kwa usahihi.

```bash
wall-vault doctor check   # Chunguza hali ya sasa (inasoma tu, haibadilishi chochote)
wall-vault doctor fix     # Rekebisha matatizo kiotomatiki
wall-vault doctor all     # Uchunguzi + kurekebisha kiotomatiki kwa wakati mmoja
```

> 💡 Ikiwa kitu kinaonekana si sawa, endesha `wall-vault doctor all` kwanza. Inashughulikia matatizo mengi kiotomatiki.

---

## RTK Kuokoa Tokeni

*(v0.1.24+)*

**RTK (Zana ya Kuokoa Tokeni)** inashinikiza kiotomatiki matokeo ya amri za shell zinazotekelezwa na mawakala wa AI wa msimbo (kama Claude Code) ili kupunguza matumizi ya tokeni. Kwa mfano, matokeo ya mistari 15 ya `git status` yanapunguzwa hadi muhtasari wa mistari 2.

### Matumizi ya Msingi

```bash
# Funga amri kwa wall-vault rtk na matokeo yatachujwa kiotomatiki
wall-vault rtk git status          # Orodha ya faili zilizobadilishwa tu
wall-vault rtk git diff HEAD~1     # Mistari iliyobadilishwa + muktadha wa chini
wall-vault rtk git log -10         # Hash + ujumbe wa mstari mmoja
wall-vault rtk go test ./...       # Majaribio yaliyoshindikana tu
wall-vault rtk ls -la              # Amri zisizotumika zinakatwa kiotomatiki
```

### Amri Zinazotumika na Athari ya Kupunguza

| Amri | Njia ya Kuchuja | Kiwango cha Kupunguza |
|------|----------|--------|
| `git status` | Muhtasari wa faili zilizobadilishwa tu | ~87% |
| `git diff` | Mistari iliyobadilishwa + muktadha wa mistari 3 | ~60-94% |
| `git log` | Hash + ujumbe wa mstari wa kwanza | ~90% |
| `git push/pull/fetch` | Ondoa maendeleo, muhtasari tu | ~80% |
| `go test` | Onyesha kushindikana tu, hesabu kupita | ~88-99% |
| `go build/vet` | Onyesha makosa tu | ~90% |
| Amri nyingine yoyote | Mistari 50 ya kwanza + 50 ya mwisho, ukubwa wa juu 32KB | Inategemea |

### Mfumo wa Kuchuja wa Hatua 3

1. **Kichujio cha muundo kwa kila amri** — Kinaelewa muundo wa matokeo ya git, go n.k. na kuchukua sehemu zenye maana tu
2. **Usindikaji wa baadaye wa regex** — Ondoa misimbo ya rangi ya ANSI, punguza mistari tupu, jumlisha mistari inayojirudia
3. **Passthrough + kukata** — Amri zisizotumika huhifadhi mistari 50 ya kwanza/mwisho tu

### Muunganisho na Claude Code

Unaweza kusanidi amri zote za shell kupitia RTK kiotomatiki kupitia hook ya `PreToolUse` ya Claude Code.

```bash
# Sakinisha hook (inaongezwa kiotomatiki kwenye Claude Code settings.json)
wall-vault rtk hook install
```

Au ongeza kwa mkono kwenye `~/.claude/settings.json`:

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

> 💡 **Kuhifadhi Exit code**: RTK inarudisha msimbo wa kutoka wa amri ya asili kama ilivyo. Ikiwa amri itashindikana (exit code ≠ 0), AI pia itagundua kushindikana kwa usahihi.

> 💡 **Kulazimisha Kiingereza**: RTK inatekeleza amri na `LC_ALL=C` ili kutoa matokeo ya Kiingereza kila wakati bila kujali mipangilio ya lugha ya mfumo. Hii inahakikisha kichujio kinafanya kazi kwa usahihi.

---

## Marejeo ya Vigeuzi vya Mazingira

Vigeuzi vya mazingira ni njia ya kupitisha thamani za usanidi kwa programu. Ingiza kwa muundo wa `export jina-la-kigezo=thamani` kwenye terminali, au weka kwenye faili ya huduma ya kuanzisha kiotomatiki ili itumike kila wakati.

| Kigezo | Maelezo | Thamani ya Mfano |
|------|------|---------|
| `WV_LANG` | Lugha ya dashibodi | `ko`, `en`, `ja` |
| `WV_THEME` | Mandhari ya dashibodi | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Ufunguo wa Google API (nyingi kwa koma) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Ufunguo wa OpenRouter API | `sk-or-v1-...` |
| `WV_VAULT_URL` | Anwani ya seva ya ghala katika hali ya kusambazwa | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Tokeni ya uthibitishaji ya mteja (boti) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Tokeni ya msimamizi | `admin-token-here` |
| `WV_MASTER_PASS` | Nenosiri la usimbuaji wa ufunguo wa API | `my-password` |
| `WV_AVATAR` | Njia ya faili ya picha ya avatar (njia ya jamaa kutoka `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Anwani ya seva ya ndani ya Ollama | `http://192.168.x.x:11434` |

---

## Utatuzi wa Matatizo

### Wakati Proxy Haiwezi Kuanza

Mara nyingi, bandari tayari inatumika na programu nyingine.

```bash
ss -tlnp | grep 56244   # Angalia nani anatumia bandari 56244
wall-vault proxy --port 8080   # Anza na nambari tofauti ya bandari
```

### Wakati Kosa la Ufunguo wa API Linatokea (429, 402, 401, 403, 582)

| Msimbo wa Kosa | Maana | Jinsi ya Kushughulikia |
|----------|------|----------|
| **429** | Maombi mengi sana (matumizi yamezidi) | Subiri kidogo au ongeza ufunguo mwingine |
| **402** | Malipo yanahitajika au mkopo hautoshi | Jaza mkopo kwenye huduma husika |
| **401 / 403** | Ufunguo si sahihi au hakuna ruhusa | Thibitisha thamani ya ufunguo na usajili tena |
| **582** | Mzigo wa gateway (cooldown dakika 5) | Inaondolewa kiotomatiki baada ya dakika 5 |

```bash
# Angalia orodha na hali ya funguo zilizosajiliwa
curl -H "Authorization: Bearer tokeni-ya-msimamizi" http://localhost:56243/admin/keys

# Weka upya vihesabio vya matumizi ya funguo
curl -X POST -H "Authorization: Bearer tokeni-ya-msimamizi" http://localhost:56243/admin/keys/reset
```

### Wakati Wakala Anaonyeshwa kama "Haijaunganishwa"

"Haijaunganishwa" inamaanisha mchakato wa proxy hautumi ishara (heartbeat) kwa ghala. **Haimaanishi kwamba mipangilio haijahifadhiwa.** Proxy lazima iendesha ikijua anwani ya seva ya ghala na tokeni ili hali ya muunganisho ibadilke.

```bash
# Anza proxy ukibainisha anwani ya seva ya ghala, tokeni, na kitambulisho cha mteja
WV_VAULT_URL=http://anwani-ya-seva-ya-ghala:56243 \
WV_VAULT_TOKEN=tokeni-ya-mteja \
WV_VAULT_CLIENT_ID=kitambulisho-cha-mteja \
wall-vault proxy
```

Muunganisho ukifanikiwa, inabadilika kuwa 🟢 Inaendesha kwenye dashibodi ndani ya takriban sekunde 20.

### Wakati Ollama Haiwezi Kuunganishwa

Ollama ni programu inayoendesha AI moja kwa moja kwenye kompyuta yako. Kwanza hakikisha Ollama imewashwa.

```bash
curl http://localhost:11434/api/tags   # Orodha ya modeli ikionekana, ni kawaida
export OLLAMA_URL=http://192.168.x.x:11434   # Ikiwa inaendesha kwenye kompyuta nyingine
```

> ⚠️ Ikiwa Ollama haijibu, anza Ollama kwanza kwa amri ya `ollama serve`.

> ⚠️ **Modeli kubwa ni polepole**: Modeli kubwa kama `qwen3.5:35b`, `deepseek-r1` zinaweza kuchukua dakika kadhaa kuzalisha jibu. Hata ikionekana kama hakuna jibu, inaweza kuwa inashughulika kawaida, kwa hivyo subiri.

---

## Mabadiliko ya Hivi Karibuni (v0.1.16 ~ v0.1.25)

### v0.1.25 (2026-04-08)
- **Kugundua Mchakato wa Wakala**: Proxy inagundua hali ya uhai ya wakala wa ndani (NanoClaw/OpenClaw) na kuonyesha kwa taa ya machungwa kwenye dashibodi.
- **Uboreshaji wa Kishiko cha Kuburuta**: Imebadilishwa ili kadi zishikwe tu kutoka eneo la taa za trafiki (●) wakati wa kupanga. Sehemu za kuingiza au vitufe havikurutiwi kwa bahati mbaya.

### v0.1.24 (2026-04-06)
- **Amri ndogo ya RTK ya Kuokoa Tokeni**: `wall-vault rtk <command>` inachuja kiotomatiki matokeo ya amri za shell ili kupunguza matumizi ya tokeni ya mawakala wa AI kwa 60-90%. Inajumuisha vichujio maalum kwa amri kuu kama git, go, na amri zisizotumika pia zinakatwa kiotomatiki. Inaunganishwa kwa uwazi kupitia hook ya `PreToolUse` ya Claude Code.

### v0.1.23 (2026-04-06)
- **Urekebishaji wa Kubadilisha Modeli ya Ollama**: Tatizo ambapo kubadilisha modeli ya Ollama kwenye dashibodi ya ghala hakuakisiwa kwenye proxy limesahihishwa. Hapo awali ilitumia kigezo cha mazingira (`OLLAMA_MODEL`) tu, lakini sasa usanidi wa ghala unapewa kipaumbele.
- **Taa za Trafiki za Kiotomatiki za Huduma ya Ndani**: Ollama, LM Studio, vLLM zinawezeshwa kiotomatiki zinapoweza kuunganishwa, na kuzimwa kiotomatiki zinapokatika. Inafanya kazi kwa njia ile ile na ubadilishaji wa kiotomatiki unaotegemea ufunguo wa huduma za wingu.

### v0.1.22 (2026-04-05)
- **Urekebishaji wa Kukosa Sehemu Tupu ya content**: Modeli za fikira (gemini-3.1-pro, o1, claude thinking n.k.) zinapotumia kikomo chote cha max_tokens kwa hoja na kushindwa kuzalisha jibu halisi, proxy ilikuwa ikiacha sehemu za `content`/`text` kwenye JSON ya jibu kwa `omitempty`, na kusababisha wateja wa OpenAI/Anthropic SDK kupata kosa la `Cannot read properties of undefined (reading 'trim')`. Imebadilishwa kujumuisha sehemu kila wakati kulingana na ubainisho rasmi wa API.

### v0.1.21 (2026-04-05)
- **Msaada wa Modeli ya Gemma 4**: Modeli za familia ya Gemma kama `gemma-4-31b-it`, `gemma-4-26b-a4b-it` zinaweza kutumika kupitia Google Gemini API.
- **Msaada Rasmi wa Huduma za LM Studio / vLLM**: Hapo awali huduma hizi zilirukwa kwenye uelekezaji wa proxy na daima zilbadilishwa na Ollama. Sasa zinaelekezwa kawaida kupitia API inayooana na OpenAI.
- **Urekebishaji wa Onyesho la Huduma kwenye Dashibodi**: Hata fallback ikitokea, dashibodi daima inaonyesha huduma iliyowekwa na mtumiaji.
- **Onyesho la Hali ya Huduma ya Ndani**: Hali ya muunganisho wa huduma za ndani (Ollama, LM Studio, vLLM n.k.) inaonyeshwa kwa rangi ya alama ya ● wakati dashibodi inapakia.
- **Kigezo cha Mazingira cha Kichujio cha Zana**: Hali ya kupitisha zana inaweza kuwekwa na kigezo cha mazingira `WV_TOOL_FILTER=passthrough`.

### v0.1.20 (2026-03-28)
- **Uimarishaji Kamili wa Usalama**: Uboreshaji wa vipengee 12 vya usalama ikiwa ni pamoja na kuzuia XSS (sehemu 41), ulinganishaji wa tokeni wa muda usio na tofauti, vikwazo vya CORS, vikomo vya ukubwa wa ombi, kuzuia kupita njia, uthibitishaji wa SSE, uimarishaji wa kikomo cha kasi n.k.

### v0.1.19 (2026-03-27)
- **Kugundua Claude Code Mtandaoni**: Claude Code isiyopitia proxy pia inaonyeshwa kama mtandaoni kwenye dashibodi.

### v0.1.18 (2026-03-26)
- **Urekebishaji wa Kukwama kwa Huduma ya Fallback**: Baada ya kurudi kwa Ollama kutokana na kosa la muda, huduma ya asili ikianza tena, inarudi kiotomatiki.
- **Uboreshaji wa Kugundua Nje ya Mtandao**: Ukaguzi wa hali kwa vipindi vya sekunde 15 unafanya ugunduzi wa proxy kusimama kuwa wa haraka zaidi.

### v0.1.17 (2026-03-25)
- **Kupanga Kadi kwa Kuburuta na Kudondosha**: Kadi za wakala zinaweza kuburutwa ili kubadilisha mpangilio.
- **Kitufe cha Kutumia Usanidi cha Mstari**: Kitufe cha [⚡ Tumia Usanidi] kinaonyeshwa kwenye mawakala walio nje ya mtandao.
- **Aina ya wakala wa cokacdir imeongezwa**.

### v0.1.16 (2026-03-25)
- **Usawazishaji wa Modeli wa Pande Mbili**: Kubadilisha modeli ya Cline au Claude Code kwenye dashibodi ya ghala kunasababisha kuakisiwa kiotomatiki.

---

*Kwa taarifa za kina zaidi za API, tazama [API.md](API.md).*
