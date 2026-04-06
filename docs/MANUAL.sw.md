# Mwongozo wa Mtumiaji wa wall-vault
*(Ilisasishwa mara ya mwisho: 2026-04-06 — v0.1.24)*

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

**wall-vault = Wakala wa AI (proxy) + Ghala la Funguo za API kwa ajili ya OpenClaw**

Ili kutumia huduma za AI, unahitaji **ufunguo wa API**. Ufunguo wa API ni kama **kibali cha kidijitali** kinachothibitisha kwamba una ruhusa ya kutumia huduma hiyo. Kibali hiki kina kikomo cha matumizi kwa siku, na kikiachwa bila ulinzi, kinaweza kuvamiwa.

wall-vault huhifadhi vibali hivi katika ghala salama na hufanya kazi kama **wakala (proxy)** kati ya OpenClaw na huduma za AI. Kwa urahisi: OpenClaw inaunganishwa na wall-vault peke yake, na wall-vault ndiye anayeshughulikia kila kitu kingine.

Matatizo ambayo wall-vault huyatatua:

- **Mzunguko wa Otomatiki wa Funguo**: Ufunguo ukifikia kikomo au kuzuiwa kwa muda (cooldown), wall-vault hubadilika kimya kimya kwenye ufunguo unaofuata. OpenClaw inaendelea kufanya kazi bila kukatizwa.
- **Ubadilishaji Otomatiki wa Huduma (Fallback)**: Kama Google haikujibu, inabadilika kwenye OpenRouter. Kama hiyo pia haikufanya kazi, inabadilika kwenye Ollama, LM Studio, au vLLM (AI inayofanya kazi ndani ya kompyuta yako). Kikao hakikatizwi. Huduma ya awali inaporejea, inabadilisha kurudi kwake kiotomatiki kuanzia ombi linalofuata (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Usawazishaji wa Wakati Halisi (SSE)**: Ukibadilisha mfano wa AI kwenye dashibodi ya ghala, OpenClaw itaonyesha mabadiliko hayo ndani ya sekunde 1–3. SSE (Server-Sent Events) ni teknolojia ambapo seva hutuma mabadiliko moja kwa moja kwa mteja wakati wowote yanapotokea.
- **Arifa za Wakati Halisi**: Matukio kama ufunguo kuisha au tatizo la huduma yanaonyeshwa mara moja kwenye chini ya skrini ya TUI ya OpenClaw.

> 💡 **Claude Code, Cursor, VS Code** pia zinaweza kuunganishwa na wall-vault, lakini madhumuni ya msingi ya wall-vault ni kutumika pamoja na OpenClaw.

```
OpenClaw (skrini ya TUI ya terminal)
        │
        ▼
  wall-vault proxy (:56244)   ← usimamizi wa funguo, uelekezaji, fallback, matukio
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (mifano 340+)
        ├─ Ollama / LM Studio / vLLM (kompyuta yako, hifadhi ya mwisho)
        └─ OpenAI / Anthropic API
```

---

## Ufungaji

### Linux / macOS

Fungua terminal na ubandike amri hizi kama zilivyo.

```bash
# Linux (PC ya kawaida, seva — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Inapakua faili kutoka kwenye mtandao.
- `chmod +x` — Inafanya faili iliyopakuliwa iweze "kutekelezwa". Ukipitisha hatua hii utapata kosa la "ruhusa imekataliwa".

### Windows

Fungua PowerShell (kama msimamizi) na uendeshe amri zifuatazo.

```powershell
# Kupakua
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Kuongeza PATH (inatumika baada ya kufungua tena PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **PATH ni nini?** Ni orodha ya folda ambazo kompyuta hutafuta amri ndani yake. Ukiongeza kwenye PATH, unaweza kuandika `wall-vault` kutoka folda yoyote na kuiendesha.

### Kujenga kutoka Chanzo (kwa wasanidi programu)

Hii inatumika tu ikiwa mazingira ya programu ya Go yamesakinishwa.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (toleo: v0.1.24.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Toleo la muhuri wa wakati**: Unapojenga kwa `make build`, toleo linaundwa otomatiki kwa umbizo kama `v0.1.24.20260406.211004` lenye tarehe na saa. Ukijenga moja kwa moja kwa `go build ./...`, toleo litaonyesha `"dev"` tu.

---

## Kuanza kwa Mara ya Kwanza

### Kuendesha mchawi wa setup

Baada ya ufungaji, hakikisha umetekeleza **mchawi wa usanidi** kwa amri ifuatayo. Mchawi atakuongoza kupitia vipengele vyote muhimu kwa kuuliza moja kwa moja.

```bash
wall-vault setup
```

Hatua ambazo mchawi anazipitia ni hizi:

```
1. Uchaguzi wa lugha (lugha 10 ikiwa ni pamoja na Kiswahili)
2. Uchaguzi wa mandhari (light / dark / gold / cherry / ocean)
3. Hali ya uendeshaji — peke yako (standalone), au pamoja na vifaa vingi (distributed)
4. Jina la boti — jina litakaloonekana kwenye dashibodi
5. Mpangilio wa bandari — chaguo-msingi: proxy 56244, ghala 56243 (bonyeza Enter ikiwa huhitaji kubadilisha)
6. Uchaguzi wa huduma za AI — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Mpangilio wa kichujio cha zana za usalama
8. Mpangilio wa tokeni ya msimamizi — nenosiri la kufunga vipengele vya usimamizi wa dashibodi. Inaweza kuzalishwa otomatiki
9. Mpangilio wa nenosiri la usimbuaji wa ufunguo wa API — kwa uhifadhi salama zaidi (hiari)
10. Eneo la kuhifadhi faili ya usanidi
```

> ⚠️ **Kumbuka tokeni ya msimamizi.** Utaihitaji baadaye kuongeza funguo kwenye dashibodi au kubadilisha mipangilio. Ukiipoteza utahitaji kuhariri moja kwa moja faili ya usanidi.

Mchawi ukikamilika, faili ya usanidi `wall-vault.yaml` itaundwa otomatiki.

### Kuendesha

```bash
wall-vault start
```

Seva mbili zitaanza kwa wakati mmoja:

- **Proxy** (`http://localhost:56244`) — wakala anayeunganisha OpenClaw na huduma za AI
- **Ghala la Funguo** (`http://localhost:56243`) — usimamizi wa ufunguo wa API na dashibodi ya wavuti

Fungua kivinjari na uende `http://localhost:56243` kuona dashibodi moja kwa moja.

---

## Kusajili Ufunguo wa API

Kuna njia nne za kusajili ufunguo wa API. **Kwa wanaoanza, njia ya 1 (vigeuzi vya mazingira) inashauriwa**.

### Njia ya 1: Vigeuzi vya Mazingira (Inashauriwa — rahisi zaidi)

Vigeuzi vya mazingira ni **thamani zilizowekwa mapema** ambazo programu husoma wakati inapoanza. Andika kwenye terminal kama hivi:

```bash
# Kusajili ufunguo wa Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Kusajili ufunguo wa OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Kuendesha baada ya kusajili
wall-vault start
```

Ikiwa una funguo nyingi, unganisha kwa koma (,). wall-vault itatumia funguo kwa zamu otomatiki (round robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Kidokezo**: Amri ya `export` inatumika tu kwenye kikao cha sasa cha terminal. Ili ibaki hata baada ya kuanzisha upya kompyuta, ongeza mstari huo kwenye faili ya `~/.bashrc` au `~/.zshrc`.

### Njia ya 2: Dashibodi ya UI (bonyeza kwa kipanya)

1. Fungua kivinjari na uende `http://localhost:56243`
2. Kwenye kadi ya **🔑 API Funguo** juu, bonyeza kitufe cha `[+ Ongeza]`
3. Ingiza aina ya huduma, thamani ya ufunguo, lebo (jina la kumbukumbu), na kikomo cha kila siku, kisha uhifadhi

### Njia ya 3: REST API (kwa ajili ya otomatiki na hati)

REST API ni njia ya programu kubadilishana data kupitia HTTP. Inafaa kwa usajili wa otomatiki kupitia hati.

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

### Njia ya 4: Alama za proxy (kwa majaribio ya muda mfupi)

Tumia hii kuweka ufunguo kwa muda bila usajili rasmi. Ufunguo utatoweka programu inapozimwa.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Jinsi ya Kutumia Wakala (Proxy)

### Kutumia na OpenClaw (madhumuni ya msingi)

Hivi ndivyo unavyoweka OpenClaw iunganishwe na huduma za AI kupitia wall-vault.

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
          { id: "wall-vault/hunter-alpha" },    // muktadha wa bure wa 1M
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Njia rahisi zaidi**: Bonyeza kitufe cha **🦞 Nakili Usanidi wa OpenClaw** kwenye kadi ya wakala katika dashibodi. Kipande chenye tokeni na anwani tayari kimejazwa kitanakiliwa kwenye ubao wa kunakilia. Bandika tu.

**`wall-vault/` kabla ya jina la mfano inaelekezwa wapi?**

Kulingana na jina la mfano, wall-vault inajua otomatiki ni huduma ipi ya AI ya kupeleka ombi:

| Umbizo la Mfano | Huduma Inayounganishwa |
|----------|--------------|
| `wall-vault/gemini-*` | Google Gemini moja kwa moja |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI moja kwa moja |
| `wall-vault/claude-*` | Anthropic kupitia OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (muktadha wa bure wa tokeni milioni 1) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/jina-la-mfano`, `openai/jina-la-mfano`, `anthropic/jina-la-mfano` n.k. | Huduma husika moja kwa moja |
| `custom/google/jina-la-mfano`, `custom/openai/jina-la-mfano` n.k. | Ondoa sehemu ya `custom/` na uelekeze upya |
| `jina-la-mfano:cloud` | Ondoa sehemu ya `:cloud` na uunganishe kupitia OpenRouter |

> 💡 **Muktadha (context) ni nini?** Ni kiasi cha mazungumzo ambacho AI kinaweza kukumbuka kwa wakati mmoja. 1M (tokeni milioni 1) inamaanisha mazungumzo marefu sana au nyaraka ndefu zinaweza kushughulikiwa mara moja.

### Kuunganisha Moja kwa Moja kwa Umbizo la Gemini API (utangamano wa zana zilizopo)

Ikiwa una zana zilizokuwa zikitumia Google Gemini API moja kwa moja, badilisha tu anwani kuwa ya wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Au ikiwa zana yako inatumia URL moja kwa moja:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Kutumia na OpenAI SDK (Python)

Unaweza pia kuunganisha wall-vault kwenye nambari ya Python inayotumia AI. Badilisha `base_url` tu:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # wall-vault inasimamia ufunguo wa API
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # ingiza kwa umbizo la provider/model
    messages=[{"role": "user", "content": "Habari"}]
)
```

### Kubadilisha Mfano Wakati wa Uendeshaji

Ili kubadilisha mfano wa AI wakati wall-vault tayari inaendesha:

```bash
# Kubadilisha mfano kwa kuomba moja kwa moja kwa proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# Katika hali ya kusambazwa (boti nyingi), badilisha kwenye seva ya ghala → itaonyeshwa mara moja kupitia SSE
curl -X PUT http://localhost:56243/admin/clients/boti-yangu-id \
  -H "Authorization: Bearer tokeni-ya-msimamizi" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Kuangalia Orodha ya Mifano Inayopatikana

```bash
# Angalia orodha yote
curl http://localhost:56244/api/models | python3 -m json.tool

# Mifano ya Google peke yake
curl "http://localhost:56244/api/models?service=google"

# Tafuta kwa jina (mfano: mifano yenye "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Muhtasari wa mifano kuu kwa huduma:**

| Huduma | Mifano Kuu |
|--------|----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | Zaidi ya 346 (Hunter Alpha 1M muktadha bure, DeepSeek R1/V3, Qwen 2.5 n.k.) |
| Ollama | Inagundua otomatiki mifano ya seva ya ndani kwenye kompyuta yako |
| LM Studio | Seva ya ndani kwenye kompyuta yako (bandari 1234) |
| vLLM | Seva ya ndani kwenye kompyuta yako (bandari 8000) |

---

## Dashibodi ya Ghala la Funguo

Fungua kivinjari na uende `http://localhost:56243` kuona dashibodi.

**Mpangilio wa skrini:**
- **Baa ya juu iliyobandikwa (topbar)**: Nembo, kichaguzi cha lugha na mandhari, hali ya muunganisho wa SSE
- **Gridi ya kadi**: Kadi za wakala, huduma, na ufunguo wa API zimepangwa kwa vigae

### Kadi ya Ufunguo wa API

Kadi inayokuruhusu kusimamia funguo zote za API zilizosajiliwa kwa mtazamo mmoja.

- Inaonyesha orodha ya funguo zilizogawanywa kwa huduma.
- `today_usage`: Tokeni (idadi ya herufi ambazo AI imesoma na kuandika) zilizochakatwa kwa mafanikio leo
- `today_attempts`: Jumla ya maombi ya leo (mafanikio + kushindwa)
- Kitufe cha `[+ Ongeza]` cha kusajili ufunguo mpya, na `✕` kufuta ufunguo.

> 💡 **Tokeni (token) ni nini?** Ni kipimo kinachotumiwa na AI kuchakata maandishi. Kwa takriban ni neno moja la Kiingereza, au herufi 1–2 za lugha nyingine. Gharama za API kawaida huhesabiwa kulingana na idadi ya tokeni.

### Kadi ya Wakala

Kadi inayoonyesha hali ya boti (wakala) zilizounganishwa na proxy ya wall-vault.

**Hali ya muunganisho inaonyeshwa katika viwango 4:**

| Alama | Hali | Maana |
|------|------|------|
| 🟢 | Inaendesha | Proxy inafanya kazi vizuri |
| 🟡 | Kuchelewa | Inajibu lakini polepole |
| 🔴 | Nje ya mtandao | Proxy haijibu |
| ⚫ | Haijaunganishwa/Imezimwa | Proxy haijawahi kuunganishwa na ghala au imezimwa |

**Maelezo ya vitufe chini ya kadi ya wakala:**

Unaposajili wakala na kubainisha **aina ya wakala**, vitufe vya urahisi vinavyofaa aina hiyo vitaonekana otomatiki.

---

#### 🔘 Kitufe cha Kunakili Usanidi — Kinaunda usanidi wa muunganisho otomatiki

Ukibonyeza kitufe, kipande cha usanidi chenye tokeni, anwani ya proxy, na taarifa za mfano za wakala huyo kitanakiliwa kwenye ubao wa kunakilia. Bandika tu kwenye mahali panapoonyeshwa kwenye jedwali hapa chini na usanidi wa muunganisho utakamilika.

| Kitufe | Aina ya Wakala | Mahali pa Kubandika |
|------|-------------|-------------|
| 🦞 Nakili Usanidi wa OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Nakili Usanidi wa NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Nakili Usanidi wa Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Nakili Usanidi wa Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Nakili Usanidi wa VSCode | `vscode` | `~/.continue/config.json` |

**Mfano — Ikiwa ni aina ya Claude Code, hii ndiyo itakayonakiliwa:**

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

> ⚠️ **Ikiwa kunakili kwenye ubao wa kunakilia hakufanyi kazi**: Sera za usalama za kivinjari zinaweza kuzuia kunakili. Ikiwa sanduku la maandishi litafunguka kama dirisha-dogo, bonyeza Ctrl+A kuchagua yote kisha Ctrl+C kunakili.

---

#### ⚡ Kitufe cha Kutumia Otomatiki — Bonyeza mara moja na usanidi umekamilika

Ikiwa aina ya wakala ni `cline`, `claude-code`, `openclaw`, au `nanoclaw`, kitufe cha **⚡ Tumia Usanidi** kitaonyeshwa kwenye kadi ya wakala. Ukibonyeza kitufe hiki, faili ya usanidi wa ndani ya wakala itasasishwa otomatiki.

| Kitufe | Aina ya Wakala | Faili Inayotumika |
|------|-------------|-------------|
| ⚡ Tumia Usanidi wa Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Tumia Usanidi wa Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Tumia Usanidi wa OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Tumia Usanidi wa NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Kitufe hiki kinatuma ombi kwa **localhost:56244** (proxy ya ndani). Proxy lazima iwe inaendesha kwenye mashine hiyo ili ifanye kazi.

---

#### 🔀 Kupanga Kadi kwa Kuburuta na Kudondosha (v0.1.17)

Unaweza **kuburuta** kadi za wakala kwenye dashibodi na kuzipanga kwa mpangilio unaotaka.

1. Shika kadi ya wakala kwa kipanya na uiburute
2. Iache juu ya kadi unayotaka na mpangilio utabadilika
3. Mpangilio mpya **unahifadhiwa mara moja kwenye seva** na unabaki hata baada ya kupakia upya

> 💡 Vifaa vya kuguswa (simu za mkononi/kompyuta kibao) bado havijauniwa. Tumia kivinjari cha kompyuta ya mezani.

---

#### 🔄 Usawazishaji wa Mifano Pande Mbili (v0.1.16)

Ukibadilisha mfano wa wakala kwenye dashibodi ya ghala, usanidi wa ndani wa wakala husasishwa otomatiki.

**Kwa Cline:**
- Kubadilisha mfano kwenye ghala → tukio la SSE → proxy inasasisha sehemu ya mfano katika `globalState.json`
- Sehemu zinazosasishwa: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` na ufunguo wa API hazigusiwi
- **Kupakia upya VS Code (`Ctrl+Alt+R` au `Ctrl+Shift+P` → `Developer: Reload Window`) kunahitajika**
  - Cline haisomi tena faili ya usanidi wakati inaendesha

**Kwa Claude Code:**
- Kubadilisha mfano kwenye ghala → tukio la SSE → proxy inasasisha sehemu ya `model` katika `settings.json`
- Inatafuta otomatiki njia za WSL na Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Mwelekeo wa nyuma (wakala → ghala):**
- Wakala (Cline, Claude Code n.k.) anapotuma ombi kwa proxy, proxy inajumuisha taarifa za huduma/mfano wa mteja katika heartbeat
- Huduma/mfano inayotumiwa sasa inaonyeshwa kwa wakati halisi kwenye kadi ya wakala katika dashibodi ya ghala

> 💡 **Jambo kuu**: Proxy inatambua wakala kwa tokeni ya Authorization ya ombi na kuelekeza otomatiki kwa huduma/mfano iliyowekwa kwenye ghala. Hata kama Cline au Claude Code inatuma jina tofauti la mfano, proxy inabadilisha kuwa mpangilio wa ghala.

---

### Kutumia Cline kwenye VS Code — Mwongozo wa Kina

#### Hatua ya 1: Sakinisha Cline

Sakinisha **Cline** (ID: `saoudrizwan.claude-dev`) kutoka soko la nyongeza za VS Code.

#### Hatua ya 2: Sajili wakala kwenye ghala

1. Fungua dashibodi ya ghala (`http://IP-ya-ghala:56243`)
2. Bonyeza **+ Ongeza** katika sehemu ya **Wakala**
3. Ingiza kama ifuatavyo:

| Sehemu | Thamani | Maelezo |
|------|----|------|
| ID | `cline_yangu` | Kitambulisho cha kipekee (herufi za Kiingereza, bila nafasi) |
| Jina | `Cline Yangu` | Jina litakaloonyeshwa kwenye dashibodi |
| Aina ya Wakala | `cline` | ← lazima uchague `cline` |
| Huduma | Chagua huduma ya kutumia (mfano: `google`) | |
| Mfano | Ingiza mfano wa kutumia (mfano: `gemini-2.5-flash`) | |

4. Bonyeza **Hifadhi** na tokeni itazalishwa otomatiki

#### Hatua ya 3: Unganisha Cline

**Njia A — Kutumia otomatiki (inashauriwa)**

1. Hakikisha **proxy** ya wall-vault inaendesha kwenye mashine hiyo (`localhost:56244`)
2. Bonyeza kitufe cha **⚡ Tumia Usanidi wa Cline** kwenye kadi ya wakala katika dashibodi
3. Ikiwa arifa "Usanidi umetumika!" inaonekana, umefanikiwa
4. Pakia upya VS Code (`Ctrl+Alt+R`)

**Njia B — Usanidi wa mwenyewe**

Fungua mipangilio (⚙️) kwenye upau wa pembeni wa Cline:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://anwani-ya-proxy:56244/v1`
  - Mashine ile ile: `http://localhost:56244/v1`
  - Mashine nyingine kama seva ya Mini: `http://192.168.1.20:56244/v1`
- **API Key**: Tokeni iliyotolewa kutoka ghala (nakili kutoka kadi ya wakala)
- **Model ID**: Mfano uliowekwa kwenye ghala (mfano: `gemini-2.5-flash`)

#### Hatua ya 4: Thibitisha

Tuma ujumbe wowote kwenye sanduku la mazungumzo la Cline. Ikiwa ni sawa:
- Nukta ya kijani (● Inaendesha) itaonekana kwenye kadi ya wakala katika dashibodi ya ghala
- Huduma/mfano wa sasa itaonyeshwa kwenye kadi (mfano: `google / gemini-2.5-flash`)

#### Kubadilisha Mfano

Unapotaka kubadilisha mfano wa Cline, badilisha kwenye **dashibodi ya ghala**:

1. Badilisha menyu kunjuzi ya huduma/mfano kwenye kadi ya wakala
2. Bonyeza **Tumia**
3. Pakia upya VS Code (`Ctrl+Alt+R`) — jina la mfano kwenye chini ya Cline litasasishwa
4. Kuanzia ombi linalofuata, mfano mpya utatumiwa

> 💡 Kwa kweli, proxy inatambua ombi la Cline kwa tokeni na kuelekeza kwa mfano wa mpangilio wa ghala. Hata usipopakia upya VS Code **mfano unaotumika hubadilika mara moja** — kupakia upya ni kwa ajili ya kusasisha onyesho la mfano kwenye UI ya Cline tu.

#### Kugundua Kukatika kwa Muunganisho

Ukifunga VS Code, kadi ya wakala kwenye dashibodi ya ghala itageuka njano (kuchelewa) baada ya takriban **sekunde 90**, na nyekundu (nje ya mtandao) baada ya **dakika 3**. (Kuanzia v0.1.18, ukaguzi wa hali wa sekunde 15 umefanya kugundua kuwa nje ya mtandao kuwa haraka zaidi.)

#### Utatuzi wa Matatizo

| Dalili | Sababu | Suluhisho |
|------|------|------|
| Kosa la "kushindwa kuunganisha" kwenye Cline | Proxy haiendesha au anwani si sahihi | Thibitisha proxy kwa `curl http://localhost:56244/health` |
| Nukta ya kijani haionekani kwenye ghala | Ufunguo wa API (tokeni) haujawekwa | Bonyeza tena kitufe cha **⚡ Tumia Usanidi wa Cline** |
| Mfano kwenye chini ya Cline haubadiliki | Cline imehifadhi usanidi kwenye kashe | Pakia upya VS Code (`Ctrl+Alt+R`) |
| Jina la mfano lisilo sahihi linaonekana | Hitilafu ya zamani (ilisahihishwa katika v0.1.16) | Sasisha proxy hadi v0.1.16 au zaidi |

---

#### 🟣 Kitufe cha Kunakili Amri ya Kupeleka — Kinatumika wakati wa kusakinisha kwenye mashine mpya

Kinatumika unaposakinisha proxy ya wall-vault kwa mara ya kwanza kwenye kompyuta mpya na kuiunganisha na ghala. Bonyeza kitufe na hati nzima ya usakinishaji itanakiliwa. Bandika kwenye terminal ya kompyuta mpya na uiendeshe — yafuatayo yatashughulikiwa mara moja:

1. Usakinishaji wa faili ya wall-vault (itarukwa ikiwa tayari imesakinishwa)
2. Usajili wa otomatiki wa huduma ya mtumiaji ya systemd
3. Kuanza huduma na kuunganisha otomatiki na ghala

> 💡 Tokeni ya wakala huyu na anwani ya seva ya ghala tayari zimejazwa ndani ya hati, kwa hivyo unaweza kuiendesha mara moja baada ya kubandika bila marekebisho yoyote.

---

### Kadi ya Huduma

Kadi ya kuwasha na kuzima au kusanidi huduma za AI utakazotumia.

- Swichi za kila huduma za kuwasha na kuzima
- Ukiingiza anwani ya seva ya AI ya ndani (Ollama, LM Studio, vLLM n.k. inayoendesha kwenye kompyuta yako), itagundua otomatiki mifano inayopatikana.
- **Onyesho la hali ya muunganisho wa huduma ya ndani**: Nukta ● karibu na jina la huduma ikiwa **kijani** imeunganishwa, ikiwa **kijivu** haijaunganishwa
- **Taa za trafiki otomatiki za huduma ya ndani** (v0.1.23+): Huduma za ndani (Ollama, LM Studio, vLLM) zinawashwa na kuzimwa otomatiki kulingana na iwapo zinaweza kuunganishwa. Ukiwasha huduma, ndani ya sekunde 15 nukta ● itageuka kijani na kisanduku cha kukagua kitawashwa; ukizima huduma, itazimwa otomatiki. Hii inafanya kazi kwa njia ile ile na huduma za wingu (Google, OpenRouter n.k.) zinazowashwa na kuzimwa otomatiki kulingana na uwepo wa ufunguo wa API.

> 💡 **Ikiwa huduma ya ndani inaendesha kwenye kompyuta nyingine**: Ingiza IP ya kompyuta hiyo kwenye kisanduku cha URL ya huduma. Mfano: `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio). Ikiwa huduma imefungwa kwa `127.0.0.1` badala ya `0.0.0.0`, itashindwa kufikia kupitia IP ya nje, kwa hivyo kagua anwani ya kufunga katika mpangilio wa huduma.

### Kuingiza Tokeni ya Msimamizi

Unapojaribu kutumia vipengele muhimu kama kuongeza au kufuta funguo kwenye dashibodi, dirisha-dogo la kuingiza tokeni ya msimamizi litaonekana. Ingiza tokeni uliyoiweka wakati wa mchawi wa setup. Mara tu ukiingiza, itabaki hadi utakapofunga kivinjari.

> ⚠️ **Ikiwa uthibitisho utashindwa zaidi ya mara 10 ndani ya dakika 15, IP hiyo itazuiwa kwa muda.** Ikiwa umesahau tokeni, angalia kipengele cha `admin_token` kwenye faili ya `wall-vault.yaml`.

---

## Hali ya Kusambazwa (Boti Nyingi)

Unapoendeshea OpenClaw kwenye kompyuta nyingi kwa wakati mmoja, huu ni mpangilio wa **kushiriki ghala moja la funguo**. Ni rahisi kwa sababu unahitaji kusimamia funguo mahali pamoja tu.

### Mfano wa Mpangilio

```
[Seva ya Ghala la Funguo]
  wall-vault vault    (ghala la funguo :56243, dashibodi)

[WSL Alpha]            [Raspberry Pi Gamma]    [Mac Mini Local]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ usawazishaji SSE    ↕ usawazishaji SSE      ↕ usawazishaji SSE
```

Boti zote zinaelekeza kwenye seva ya ghala katikati, kwa hivyo kubadilisha mfano au kuongeza ufunguo kwenye ghala kunaonyeshwa mara moja kwenye boti zote.

### Hatua ya 1: Anza Seva ya Ghala la Funguo

Endesha kwenye kompyuta itakayotumika kama seva ya ghala:

```bash
wall-vault vault
```

### Hatua ya 2: Sajili Kila Boti (Mteja)

Sajili mapema taarifa za kila boti itakayounganishwa na seva ya ghala:

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

### Hatua ya 3: Anza Proxy kwenye Kila Kompyuta ya Boti

Endesha proxy kwenye kila kompyuta yenye boti kwa kubainisha anwani ya seva ya ghala na tokeni:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Badilisha **`192.168.x.x`** na anwani halisi ya IP ya ndani ya kompyuta ya seva ya ghala. Unaweza kuikagua kupitia mpangilio wa router au amri ya `ip addr`.

---

## Kuanzisha Kiotomatiki

Ikiwa ni kero kuwasha wall-vault kwa mkono kila kompyuta inapoanzishwa upya, sajili kama huduma ya mfumo. Mara tu ukisajili, itaanza otomatiki wakati wa boot.

### Linux — systemd (Linux nyingi)

systemd ni mfumo wa Linux wa kuanza na kusimamia programu otomatiki:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Kuangalia kumbukumbu:

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

Mfumo unaosimamia utekelezaji otomatiki wa programu kwenye macOS:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Pakua NSSM kutoka [nssm.cc](https://nssm.cc/download) na uiongeze kwenye PATH.
2. Kwenye PowerShell ya msimamizi:

```powershell
wall-vault doctor deploy windows
```

---

## Daktari (Doctor)

Amri ya `doctor` ni zana ya wall-vault ya **kujichunguza na kujisahihisha**.

```bash
wall-vault doctor check   # Chunguza hali ya sasa (soma tu, hakuna kinachobadilishwa)
wall-vault doctor fix     # Rekebisha matatizo otomatiki
wall-vault doctor all     # Uchunguzi + urekebishaji wa otomatiki mara moja
```

> 💡 Ikiwa kitu kinaonekana si sawa, endesha `wall-vault doctor all` kwanza. Kinashughulikia matatizo mengi otomatiki.


---

## RTK Kuokoa Tokeni

*(v0.1.24+)*

**RTK (Zana ya Kuokoa Tokeni)** inabana otomatiki matokeo ya amri za sheli ambazo wakala wa AI wa kodishaji (Claude Code n.k.) huendesha, na kupunguza matumizi ya tokeni. Kwa mfano, matokeo ya mistari 15 ya `git status` yanabana kuwa muhtasari wa mistari 2.

### Matumizi ya Msingi

```bash
# Funika amri kwa wall-vault rtk na matokeo yatachujwa otomatiki
wall-vault rtk git status          # Inaonyesha orodha ya faili zilizobadilishwa tu
wall-vault rtk git diff HEAD~1     # Mistari iliyobadilishwa + muktadha wa chini tu
wall-vault rtk git log -10         # Hash + ujumbe wa mstari mmoja kwa kila ingizo
wall-vault rtk go test ./...       # Inaonyesha majaribio yaliyoshindwa tu
wall-vault rtk ls -la              # Amri zisizouniwa zinakatwa otomatiki
```

### Amri Zinazouniwa na Athari za Kuokoa

| Amri | Njia ya Kuchuja | Kiwango cha Kuokoa |
|------|----------|--------|
| `git status` | Muhtasari wa faili zilizobadilishwa tu | ~87% |
| `git diff` | Mistari iliyobadilishwa + muktadha wa mistari 3 | ~60-94% |
| `git log` | Hash + ujumbe wa mstari wa kwanza | ~90% |
| `git push/pull/fetch` | Ondoa maendeleo, muhtasari tu | ~80% |
| `go test` | Onyesha kushindwa tu, hesabu kupita | ~88-99% |
| `go build/vet` | Onyesha makosa tu | ~90% |
| Amri nyingine zote | Mistari 50 ya mwanzo + 50 ya mwisho, upeo 32KB | Inabadilika |

### Bomba la Kuchuja la Hatua 3

1. **Kichujio cha muundo kwa amri** — Kinaelewa muundo wa matokeo ya git, go n.k. na kuchota sehemu zenye maana tu
2. **Uchakataji wa baadaye wa regex** — Kuondoa misimbo ya rangi ya ANSI, kupunguza mistari tupu, kujumlisha mistari inayorudiwa
3. **Kupitisha + kukata** — Amri zisizouniwa huhifadhi mistari 50 ya mwanzo na 50 ya mwisho tu

### Kuunganisha na Claude Code

Unaweza kusanidi kwa kiungo cha `PreToolUse` cha Claude Code ili amri zote za sheli zipite kupitia RTK otomatiki.

```bash
# Sakinisha kiungo (kinaongezwa otomatiki kwenye settings.json ya Claude Code)
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

> 💡 **Uhifadhi wa exit code**: RTK inarudisha msimbo wa kutoka wa amri ya asili kama ilivyo. Amri ikishindwa (exit code ≠ 0), AI pia itagundua kushindwa kwa usahihi.

> 💡 **Kulazimisha Kiingereza**: RTK inaendesha amri kwa `LC_ALL=C` ili kuzalisha matokeo ya Kiingereza kila wakati bila kujali mpangilio wa lugha ya mfumo. Hii inahakikisha kichujio kinafanya kazi kwa usahihi.

---

## Marejeo ya Vigeuzi vya Mazingira

Vigeuzi vya mazingira ni njia ya kupitisha thamani za usanidi kwenye programu. Ingiza kwa umbizo la `export jina-la-kigezo=thamani` kwenye terminal, au viweke kwenye faili ya huduma ya kuanza otomatiki ili vitumike kila wakati.

| Kigezo | Maelezo | Thamani ya Mfano |
|------|------|---------|
| `WV_LANG` | Lugha ya dashibodi | `ko`, `en`, `ja` |
| `WV_THEME` | Mandhari ya dashibodi | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Ufunguo wa API wa Google (nyingi kwa koma) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Ufunguo wa API wa OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Anwani ya seva ya ghala katika hali ya kusambazwa | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Tokeni ya uthibitisho ya mteja (boti) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Tokeni ya msimamizi | `admin-token-here` |
| `WV_MASTER_PASS` | Nenosiri la usimbuaji wa ufunguo wa API | `my-password` |
| `WV_AVATAR` | Njia ya faili ya picha ya avatar (njia inayohusiana na `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Anwani ya seva ya ndani ya Ollama | `http://192.168.x.x:11434` |

---

## Utatuzi wa Matatizo

### Proxy Haianzishi

Mara nyingi bandari tayari inatumika na programu nyingine.

```bash
ss -tlnp | grep 56244   # Angalia ni nani anayetumia bandari 56244
wall-vault proxy --port 8080   # Anza kwa nambari ya bandari nyingine
```

### Makosa ya Ufunguo wa API (429, 402, 401, 403, 582)

| Nambari ya Kosa | Maana | Jinsi ya Kushughulikia |
|----------|------|----------|
| **429** | Maombi mengi sana (kikomo cha matumizi kimezidishwa) | Subiri kidogo au ongeza ufunguo mwingine |
| **402** | Malipo yanahitajika au salio limeisha | Jumuisha salio kwenye huduma husika |
| **401 / 403** | Ufunguo si sahihi au huna ruhusa | Thibitisha tena thamani ya ufunguo na usajili upya |
| **582** | Msongamano wa lango (cooldown dakika 5) | Itafunguliwa otomatiki baada ya dakika 5 |

```bash
# Angalia orodha ya funguo zilizosajiliwa na hali zake
curl -H "Authorization: Bearer tokeni-ya-msimamizi" http://localhost:56243/admin/keys

# Weka upya vihesabio vya matumizi ya ufunguo
curl -X POST -H "Authorization: Bearer tokeni-ya-msimamizi" http://localhost:56243/admin/keys/reset
```

### Wakala Anaonyeshwa kama "Haijaunganishwa"

"Haijaunganishwa" inamaanisha mchakato wa proxy hautumi ishara (heartbeat) kwa ghala. **Haimaanishi kuwa usanidi haujahifadhiwa.** Proxy lazima iwe inajua anwani ya seva ya ghala na tokeni na kuendesha ili kubadilika kuwa hali ya kuunganishwa.

```bash
# Anza proxy kwa kubainisha anwani ya seva ya ghala, tokeni, na ID ya mteja
WV_VAULT_URL=http://anwani-ya-seva-ya-ghala:56243 \
WV_VAULT_TOKEN=tokeni-ya-mteja \
WV_VAULT_CLIENT_ID=id-ya-mteja \
wall-vault proxy
```

Muunganisho ukifanikiwa, itabadilika kuwa 🟢 Inaendesha kwenye dashibodi ndani ya takriban sekunde 20.

### Ollama Haiunganishwi

Ollama ni programu ya kuendesha AI moja kwa moja kwenye kompyuta yako. Kwanza hakikisha Ollama inafanya kazi.

```bash
curl http://localhost:11434/api/tags   # Ikiwa orodha ya mifano inaonekana, ni sawa
export OLLAMA_URL=http://192.168.x.x:11434   # Ikiwa inaendesha kwenye kompyuta nyingine
```

> ⚠️ Ikiwa Ollama haijibu, anza kwanza Ollama kwa amri ya `ollama serve`.

> ⚠️ **Mifano mikubwa ina muda mrefu wa kujibu**: Mifano mikubwa kama `qwen3.5:35b`, `deepseek-r1` inaweza kuchukua dakika kadhaa kuzalisha jibu. Hata ikionekana kama hakuna jibu, inaweza kuwa inachakata kawaida, kwa hivyo subiri.

---

## Mabadiliko ya Hivi Karibuni (v0.1.16 ~ v0.1.24)

### v0.1.24 (2026-04-06)
- **Amri ndogo ya RTK ya kuokoa tokeni**: `wall-vault rtk <command>` inachuja otomatiki matokeo ya amri za sheli ili kupunguza matumizi ya tokeni ya wakala wa AI kwa 60-90%. Inajumuisha vichujio maalum kwa amri kuu kama git, go, na amri zisizouniwa pia zinakatwa otomatiki. Inaunganishwa kwa uwazi na kiungo cha `PreToolUse` cha Claude Code.

### v0.1.23 (2026-04-06)
- **Marekebisho ya kubadilisha mfano wa Ollama**: Ilirekebishwa tatizo ambapo kubadilisha mfano wa Ollama kwenye dashibodi ya ghala hakukuonyeshwa kwenye proxy halisi. Hapo awali, ni kigezo cha mazingira (`OLLAMA_MODEL`) peke yake kilichotumiwa, lakini sasa mpangilio wa ghala unapewa kipaumbele.
- **Taa za trafiki otomatiki za huduma ya ndani**: Ollama, LM Studio, na vLLM zinawashwa otomatiki zinapoweza kuunganishwa na kuzimwa otomatiki zinapokatika. Hii inafanya kazi kwa njia ile ile na ubadilishaji otomatiki wa huduma za wingu unaotegemea ufunguo.

### v0.1.22 (2026-04-05)
- **Marekebisho ya sehemu tupu ya content iliyokosekana**: Mifano ya kufikiri (gemini-3.1-pro, o1, claude thinking n.k.) ilipotumia kikomo cha max_tokens yote kwa reasoning na kushindwa kutoa jibu halisi, proxy iliondoa sehemu za `content`/`text` za JSON ya jibu kwa `omitempty`, na kusababisha makosa ya `Cannot read properties of undefined (reading 'trim')` kwenye wateja wa SDK wa OpenAI/Anthropic. Imebadilishwa ili sehemu zijumuishwe kila wakati kulingana na vipimo rasmi vya API.

### v0.1.21 (2026-04-05)
- **Msaada wa mifano ya Gemma 4**: Mifano ya familia ya Gemma kama `gemma-4-31b-it`, `gemma-4-26b-a4b-it` inaweza kutumiwa kupitia Google Gemini API.
- **Msaada rasmi wa huduma za LM Studio / vLLM**: Hapo awali huduma hizi zilirukwa katika uelekezaji wa proxy na kila mara zilibadilishwa na Ollama. Sasa zinaelekezwa vizuri kupitia API inayoendana na OpenAI.
- **Marekebisho ya onyesho la huduma kwenye dashibodi**: Hata wakati wa fallback, dashibodi inaonyesha kila wakati huduma iliyowekwa na mtumiaji.
- **Onyesho la hali ya huduma ya ndani**: Hali ya muunganisho wa huduma za ndani (Ollama, LM Studio, vLLM n.k.) inaonyeshwa kwa rangi ya nukta ● wakati dashibodi inapakia.
- **Kigezo cha mazingira cha kichujio cha zana**: Hali ya kupitisha zana (tools) inaweza kuwekwa kwa kigezo cha mazingira `WV_TOOL_FILTER=passthrough`.

### v0.1.20 (2026-03-28)
- **Uimarishaji wa kina wa usalama**: Kinga dhidi ya XSS (maeneo 41), ulinganisho wa tokeni wa muda wa kudumu, vizuizi vya CORS, vikomo vya ukubwa wa ombi, kinga dhidi ya kupita njia, uthibitisho wa SSE, uimarishaji wa kizuizi cha kasi n.k. vipengele 12 vya usalama vilivyoboreshwa.

### v0.1.19 (2026-03-27)
- **Kugundua Claude Code mtandaoni**: Claude Code ambayo haipiti kupitia proxy pia inaonyeshwa mtandaoni kwenye dashibodi.

### v0.1.18 (2026-03-26)
- **Marekebisho ya huduma ya fallback iliyokwama**: Baada ya fallback kwa Ollama kwa sababu ya kosa la muda mfupi, ikiwa huduma ya awali inrejea, inabadilika kurudi otomatiki.
- **Uboreshaji wa kugundua nje ya mtandao**: Ukaguzi wa hali wa sekunde 15 umefanya kugundua kukatika kwa proxy kuwa haraka zaidi.

### v0.1.17 (2026-03-25)
- **Kupanga kadi kwa kuburuta na kudondosha**: Kadi za wakala zinaweza kuburutwa kupanga upya mpangilio.
- **Vitufe vya kutumia usanidi ndani ya mstari**: Kitufe cha [⚡ Tumia Usanidi] kinaonyeshwa kwa wakala wasio mtandaoni.
- **Aina mpya ya wakala cokacdir**.

### v0.1.16 (2026-03-25)
- **Usawazishaji wa mifano pande mbili**: Kubadilisha mfano wa Cline au Claude Code kwenye dashibodi ya ghala kunaonyeshwa otomatiki.

---

*Kwa taarifa zaidi za API, tazama [API.md](API.md).*
