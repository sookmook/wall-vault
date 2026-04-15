# Mwongozo wa Mtumiaji wa wall-vault
*(Last updated: 2026-04-16 — v0.2.1)*

---

## Yaliyomo

1. [wall-vault ni nini?](#wall-vault-ni-nini)
2. [Ufungaji](#ufungaji)
3. [Kuanza kwa Mara ya Kwanza (mchawi wa setup)](#kuanza-kwa-mara-ya-kwanza)
4. [Usajili wa API Key](#usajili-wa-api-key)
5. [Jinsi ya Kutumia Proxy](#jinsi-ya-kutumia-proxy)
6. [Dashibodi ya Key Vault](#dashibodi-ya-key-vault)
7. [Hali ya Kusambazwa (Multi Bot)](#hali-ya-kusambazwa-multi-bot)
8. [Mpangilio wa Kuanza Kiotomatiki](#mpangilio-wa-kuanza-kiotomatiki)
9. [Doctor (Daktari)](#doctor-daktari)
10. [RTK Kuokoa Tokeni](#rtk-kuokoa-tokeni)
11. [Rejea ya Vigezo vya Mazingira](#rejea-ya-vigezo-vya-mazingira)
12. [Utatuzi wa Matatizo](#utatuzi-wa-matatizo)

---

## wall-vault ni nini?

**wall-vault = Wakala wa AI (Proxy) + Kabati la API Key kwa ajili ya OpenClaw**

Ili kutumia huduma za AI, unahitaji **API key**. API key ni kama **kitambulisho cha dijitali** kinachothibitisha kuwa "mtu huyu ana haki ya kutumia huduma hii". Hata hivyo, kitambulisho hiki kina kikomo cha matumizi kwa siku, na kikisimamiwa vibaya, kinaweza kufichuliwa.

wall-vault huhifadhi vitambulisho hivi kwenye kabati salama, na hufanya kazi kama **wakala (proxy)** kati ya OpenClaw na huduma za AI. Kwa ufupi, OpenClaw inahitaji tu kuunganishwa na wall-vault, na wall-vault inashughulikia mambo yote mengine magumu.

Matatizo ambayo wall-vault inasuluhisha:

- **Mzunguko wa Kiotomatiki wa API Key**: Wakati matumizi ya key moja yanafikia kikomo au inapigwa marufuku kwa muda (cooldown), inabadilika kimya kimya hadi key inayofuata. OpenClaw inaendelea kufanya kazi bila kukatizwa.
- **Ubadilishaji wa Kiotomatiki wa Huduma (Fallback)**: Ikiwa Google haijibu, inabadilika hadi OpenRouter, na ikiwa hiyo pia haifanyi kazi, inabadilika kiotomatiki hadi Ollama·LM Studio·vLLM (AI ya ndani) iliyofungwa kwenye kompyuta yako. Kipindi hakikatiki. Huduma ya awali ikirejea, maombi yanayofuata yatabadilika kiotomatiki (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Usawazishaji wa Wakati Halisi (SSE)**: Ukibadilisha modeli kwenye dashibodi ya kabati, inajitokeza kwenye skrini ya OpenClaw ndani ya sekunde 1-3. SSE (Server-Sent Events) ni teknolojia ambapo seva inasukuma mabadiliko kwa mteja kwa wakati halisi.
- **Arifa za Wakati Halisi**: Matukio kama uchoshaji wa key au kushindikana kwa huduma yanaonyeshwa mara moja chini ya skrini ya OpenClaw TUI (skrini ya terminal).

> 💡 **Claude Code, Cursor, VS Code** pia zinaweza kuunganishwa, lakini kusudi la awali la wall-vault ni kutumika pamoja na OpenClaw.

```
OpenClaw (Skrini ya Terminal ya TUI)
        │
        ▼
  wall-vault Proxy (:56244)   ← Usimamizi wa key, uelekezaji, fallback, matukio
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (modeli 340+)
        ├─ Ollama / LM Studio / vLLM (kompyuta yako, kimbilio la mwisho)
        └─ OpenAI / Anthropic API
```

---

## Ufungaji

### Linux / macOS

Fungua terminal na ubandike amri zifuatazo kama zilivyo.

```bash
# Linux (PC ya kawaida, seva — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Inapakua faili kutoka mtandaoni.
- `chmod +x` — Inafanya faili iliyopakuliwa "iweze kutekelezwa". Ukikosa hatua hii, utapata kosa la "ruhusa imekataliwa".

### Windows

Fungua PowerShell (kama msimamizi) na utekeleze amri zifuatazo.

```powershell
# Pakua
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Ongeza kwenye PATH (inatumika baada ya kuanzisha upya PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **PATH ni nini?** Ni orodha ya folda ambapo kompyuta inatafuta amri. Unahitaji kuiongeza kwenye PATH ili uweze kutekeleza `wall-vault` kutoka folda yoyote.

### Kujenga kutoka Chanzo (kwa Waendelezaji)

Hii inatumika tu ikiwa una mazingira ya maendeleo ya lugha ya Go.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (toleo: v0.1.25.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Toleo la Muhuri wa Wakati wa Kujenga**: Ukijenga kwa `make build`, toleo linazalishwa kiotomatiki katika muundo unaojumuisha tarehe·muda kama `v0.1.27.20260409`. Ukijenga moja kwa moja kwa `go build ./...`, toleo linaonyeshwa tu kama `"dev"`.

---

## Kuanza kwa Mara ya Kwanza

### Kuendesha mchawi wa setup

Baada ya ufungaji, hakikisha umetekeleza **mchawi wa usanidi** kwa amri ifuatayo kwanza. Mchawi atakuongoza kwa kukuuliza vitu muhimu moja baada ya nyingine.

```bash
wall-vault setup
```

Hatua ambazo mchawi anazipitia ni kama zifuatazo:

```
1. Chagua lugha (lugha 10 ikiwa ni pamoja na Kiswahili)
2. Chagua mandhari (light / dark / gold / cherry / ocean)
3. Hali ya uendeshaji — chagua kama utatumia peke yako (standalone) au kwenye mashine nyingi (distributed)
4. Ingiza jina la bot — jina litakaloonyeshwa kwenye dashibodi
5. Mpangilio wa bandari — chaguomsingi: proxy 56244, vault 56243 (bonyeza Enter tu ikiwa huhitaji kubadilisha)
6. Chagua huduma za AI — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Mpangilio wa kichuja cha zana za usalama
8. Weka tokeni ya msimamizi — nenosiri la kufunga vipengele vya usimamizi wa dashibodi. Inaweza kuzalishwa kiotomatiki
9. Weka nenosiri la usimbaji fiche wa API key — wakati unataka kuhifadhi key kwa usalama zaidi (hiari)
10. Njia ya kuhifadhi faili ya usanidi
```

> ⚠️ **Hakikisha umekumbuka tokeni ya msimamizi.** Utaihitaji baadaye unapoongeza key au kubadilisha mipangilio kwenye dashibodi. Ukiipoteza, utahitaji kuhariri faili ya usanidi moja kwa moja.

Mchawi ukimalizika, faili ya usanidi `wall-vault.yaml` inazalishwa kiotomatiki.

### Kuendesha

```bash
wall-vault start
```

Seva mbili zinaanza wakati huo huo:

- **Proxy** (`http://localhost:56244`) — Wakala anayeunganisha OpenClaw na huduma za AI
- **Key Vault** (`http://localhost:56243`) — Usimamizi wa API key na dashibodi ya wavuti

Fungua `http://localhost:56243` kwenye kivinjari chako ili kuona dashibodi mara moja.

---

## Usajili wa API Key

Kuna njia nne za kusajili API key. **Kwa wanaoanza, Njia 1 (vigezo vya mazingira) inapendekezwa**.

### Njia 1: Vigezo vya Mazingira (Inashauriwa — Rahisi Zaidi)

Vigezo vya mazingira ni **maadili yaliyowekwa mapema** ambayo programu inayasoma inapoanza. Ingiza yafuatayo kwenye terminal.

```bash
# Sajili key ya Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Sajili key ya OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Endesha baada ya usajili
wall-vault start
```

Ikiwa una key nyingi, ziunganishe kwa koma (,). wall-vault itazitumia kwa mzunguko kiotomatiki (round robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Kidokezo**: Amri ya `export` inatumika tu kwa kipindi cha sasa cha terminal. Ili ibakie hata baada ya kuanzisha upya kompyuta, ongeza mstari huo kwenye faili ya `~/.bashrc` au `~/.zshrc`.

### Njia 2: UI ya Dashibodi (Bonyeza kwa Panya)

1. Fungua `http://localhost:56243` kwenye kivinjari
2. Bonyeza kitufe cha `[+ Ongeza]` kwenye kadi ya **🔑 API Key** juu
3. Ingiza aina ya huduma, thamani ya key, lebo (jina la kumbukumbu), na kikomo cha kila siku kisha uhifadhi

### Njia 3: REST API (kwa Otomatiki·Hati)

REST API ni njia ya programu kubadilishana data kupitia HTTP. Ni muhimu kwa usajili wa kiotomatiki kupitia hati.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer tokeni-ya-msimamizi" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Key Kuu",
    "daily_limit": 1000
  }'
```

### Njia 4: Bendera ya proxy (kwa Majaribio ya Haraka)

Tumia hii wakati unataka kuingiza key kwa muda kwa majaribio bila usajili rasmi. Key inatoweka ukifunga programu.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Jinsi ya Kutumia Proxy

### Kutumia na OpenClaw (Kusudi Kuu)

Hivi ndivyo unavyosanidi OpenClaw ili kuunganishwa na huduma za AI kupitia wall-vault.

Fungua faili `~/.openclaw/openclaw.json` na uongeze yaliyomo yafuatayo:

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
          { id: "wall-vault/hunter-alpha" },    // 1M context bure
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Njia Rahisi Zaidi**: Bonyeza kitufe cha **🦞 Nakili Usanidi wa OpenClaw** kwenye kadi ya wakala kwenye dashibodi na snippet yenye tokeni na anwani tayari imejazwa itanakiliwa kwenye clipboard. Bandika tu.

**`wall-vault/` mbele ya jina la modeli inaelekeza wapi?**

Kwa kuangalia jina la modeli, wall-vault inaamua kiotomatiki ni huduma gani ya AI itatuma ombi:

| Muundo wa Modeli | Huduma Inayounganishwa |
|----------|--------------|
| `wall-vault/gemini-*` | Google Gemini moja kwa moja |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI moja kwa moja |
| `wall-vault/claude-*` | Anthropic kupitia OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (tokeni milioni 1 bure) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/jina-la-modeli`, `openai/jina-la-modeli`, `anthropic/jina-la-modeli` n.k. | Muunganisho wa moja kwa moja na huduma husika |
| `custom/google/jina-la-modeli`, `custom/openai/jina-la-modeli` n.k. | Ondoa sehemu ya `custom/` na uelekeze upya |
| `jina-la-modeli:cloud` | Ondoa sehemu ya `:cloud` na uunganishe na OpenRouter |

> 💡 **Context (muktadha) ni nini?** Ni kiasi cha mazungumzo ambacho AI inaweza kukumbuka kwa wakati mmoja. 1M (tokeni milioni) inamaanisha inaweza kushughulikia mazungumzo au hati ndefu sana kwa wakati mmoja.

### Kuunganisha Moja kwa Moja kwa Muundo wa Gemini API (utangamano na zana zilizopo)

Ikiwa una zana iliyokuwa ikitumia Google Gemini API moja kwa moja, badilisha tu anwani hadi wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Au ikiwa zana yako inabainisha URL moja kwa moja:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Kutumia na OpenAI SDK (Python)

Unaweza pia kuunganisha wall-vault kwenye msimbo wa Python unaotumia AI. Badilisha tu `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # wall-vault inasimamia API key kiotomatiki
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # ingiza katika muundo wa provider/model
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

# Katika hali ya kusambazwa (multi bot), badilisha kwenye seva ya vault → inajitokeza mara moja kupitia SSE
curl -X PUT http://localhost:56243/admin/clients/id-ya-bot-yangu \
  -H "Authorization: Bearer tokeni-ya-msimamizi" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Kuangalia Orodha ya Modeli Zinazopatikana

```bash
# Angalia orodha nzima
curl http://localhost:56244/api/models | python3 -m json.tool

# Angalia modeli za Google tu
curl "http://localhost:56244/api/models?service=google"

# Tafuta kwa jina (mfano: modeli zenye "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Muhtasari wa Modeli Kuu kwa Huduma:**

| Huduma | Modeli Kuu |
|--------|----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ (Hunter Alpha 1M context bure, DeepSeek R1/V3, Qwen 2.5 n.k.) |
| Ollama | Hugundua kiotomatiki seva ya ndani iliyofungwa kwenye kompyuta yako |
| LM Studio | Seva ya ndani ya kompyuta (bandari 1234) |
| vLLM | Seva ya ndani ya kompyuta (bandari 8000) |

---

## Dashibodi ya Key Vault

Fungua `http://localhost:56243` kwenye kivinjari chako ili kuona dashibodi.

**Mpangilio wa Skrini:**
- **Upau wa juu ulioshikiliwa (topbar)**: Nembo, kichaguzi cha lugha·mandhari, hali ya muunganisho wa SSE
- **Gridi ya Kadi**: Kadi za wakala·huduma·API key zimepangwa kama vigae

### Kadi ya API Key

Kadi ya kusimamia API key zilizosajiliwa kwa mtazamo mmoja.

- Inaonyesha orodha ya key zilizogawanywa kwa huduma.
- `today_usage`: Idadi ya tokeni (herufi ambazo AI imesoma na kuandika) zilizoshughulikiwa kwa mafanikio leo
- `today_attempts`: Jumla ya milio leo (mafanikio + kushindikana)
- Sajili key mpya kwa kitufe cha `[+ Ongeza]`, na ufute key kwa `✕`.

> 💡 **Tokeni ni nini?** Ni kipimo kinachotumiwa na AI inaporubani maandishi. Ni takriban neno moja la Kiingereza, au herufi 1-2 za lugha nyingine. Ada ya API kwa kawaida inahesabiwa kulingana na idadi hii ya tokeni.

### Kadi ya Wakala

Kadi inayoonyesha hali ya bot (wakala) zilizounganishwa na proxy ya wall-vault.

**Hali ya muunganisho inaonyeshwa kwa ngazi 4:**

| Alama | Hali | Maana |
|------|------|------|
| 🟢 | Inaendesha | Proxy inafanya kazi kawaida |
| 🟡 | Imecheleweshwa | Majibu yanakuja lakini polepole |
| 🔴 | Nje ya Mtandao | Proxy haijibu |
| ⚫ | Haijaunganishwa·Imezimwa | Proxy haijawahi kuunganishwa na vault au imezimwa |

**Mwongozo wa vitufe chini ya kadi ya wakala:**

Unaposajili wakala, ukibainisha **aina ya wakala**, vitufe vya urahisi vinavyolingana na aina hiyo vinaonekana kiotomatiki.

---

#### 🔘 Kitufe cha Nakili Usanidi — Kinaunda usanidi wa muunganisho kiotomatiki

Ukibonyeza kitufe, snippet ya usanidi yenye tokeni ya wakala, anwani ya proxy, na taarifa za modeli tayari imejazwa inanakiliwa kwenye clipboard. Bandika tu yaliyonakiliwa kwenye eneo lililoonyeshwa kwenye jedwali hapa chini ili kukamilisha usanidi wa muunganisho.

| Kitufe | Aina ya Wakala | Mahali pa Kubandika |
|------|-------------|-------------|
| 🦞 Nakili Usanidi wa OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Nakili Usanidi wa NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Nakili Usanidi wa Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Nakili Usanidi wa Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Nakili Usanidi wa VSCode | `vscode` | `~/.continue/config.json` |

**Mfano — Ikiwa aina ni Claude Code, yaliyomo kama haya yananakiliwa:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "tokeni-ya-wakala-huu"
}
```

**Mfano — Ikiwa aina ni VSCode (Continue):**

```yaml
# ~/.continue/config.yaml  ← Bandika kwenye config.yaml, si config.json
name: My Config
version: 0.0.1
schema: v1

models:
  - name: wall-vault proxy
    provider: openai
    model: gemini-2.5-flash
    apiBase: http://192.168.1.20:56244/v1
    apiKey: tokeni-ya-wakala-huu
    roles:
      - chat
      - edit
      - apply
```

> ⚠️ **Toleo la hivi karibuni la Continue linatumia `config.yaml`.** Ikiwa `config.yaml` ipo, `config.json` inapuuzwa kabisa. Hakikisha umebandika kwenye `config.yaml`.

**Mfano — Ikiwa aina ni Cursor:**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : tokeni-ya-wakala-huu

// Au vigezo vya mazingira:
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=tokeni-ya-wakala-huu
```

> ⚠️ **Kunakili kwenye clipboard hakufanyi kazi**: Sera za usalama za kivinjari zinaweza kuzuia kunakili. Ikiwa sanduku la maandishi linafunguka kwenye popup, chagua yote kwa Ctrl+A kisha nakili kwa Ctrl+C.

---

#### ⚡ Kitufe cha Kutumia Kiotomatiki — Bonyeza mara moja na usanidi umekamilika

Ikiwa aina ya wakala ni `cline`, `claude-code`, `openclaw`, au `nanoclaw`, kitufe cha **⚡ Tumia Usanidi** kinaonyeshwa kwenye kadi ya wakala. Ukibonyeza kitufe hiki, faili za usanidi za ndani za wakala husika zinasasishwa kiotomatiki.

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

1. Shika eneo la **taa ya trafiki (●)** upande wa juu kushoto wa kadi kwa panya na uburute
2. Idondoshe juu ya kadi katika nafasi unayotaka na mpangilio utabadilika

> 💡 Mwili wa kadi (sehemu za kuingiza, vitufe n.k.) hauburutiki. Unaweza kushika tu kutoka eneo la taa ya trafiki.

#### 🟠 Kugundua Mchakato wa Wakala (v0.1.25)

Wakati proxy inafanya kazi kawaida lakini mchakato wa wakala wa ndani (NanoClaw, OpenClaw) umekufa, taa ya trafiki ya kadi inabadilika kuwa **machungwa (inayofifia)** na ujumbe wa "Mchakato wa wakala umesimama" unaonyeshwa.

- 🟢 Kijani: Proxy + wakala kwa kawaida
- 🟠 Machungwa (inafifia): Proxy kawaida, wakala amekufa
- 🔴 Nyekundu: Proxy nje ya mtandao
3. Mpangilio uliobadilika **unahifadhiwa kwenye seva mara moja** na unabaki hata ukisasisha ukurasa

> 💡 Kwenye vifaa vya kugusa (simu/kibao) bado haijaunga mkono. Tumia kivinjari cha kompyuta ya mezani.

---

#### 🔄 Usawazishaji wa Pande Mbili wa Modeli (v0.1.16)

Ukibadilisha modeli ya wakala kwenye dashibodi ya vault, usanidi wa ndani wa wakala husika unasasishwa kiotomatiki.

**Kwa Cline:**
- Ukibadilisha modeli kwenye vault → tukio la SSE → proxy inasasisha sehemu ya modeli kwenye `globalState.json`
- Sehemu zinazosasishwa: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` na API key hazigusiwi
- **Unahitaji kusasisha VS Code upya (`Ctrl+Alt+R` au `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - Kwa sababu Cline haisomi faili ya usanidi tena inaporusha

**Kwa Claude Code:**
- Ukibadilisha modeli kwenye vault → tukio la SSE → proxy inasasisha sehemu ya `model` kwenye `settings.json`
- Inatafuta kiotomatiki njia zote mbili za WSL na Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Upande wa Nyuma (wakala → vault):**
- Wakala (Cline, Claude Code n.k.) anapotuma ombi kwa proxy, proxy inajumuisha taarifa za huduma·modeli za mteja husika kwenye heartbeat
- Huduma/modeli inayotumika sasa inaonyeshwa kwa wakati halisi kwenye kadi ya wakala kwenye dashibodi ya vault

> 💡 **Jambo Muhimu**: Proxy inatambua wakala kwa tokeni ya Authorization ya ombi, na inaelekeza kiotomatiki hadi huduma/modeli iliyowekwa kwenye vault. Hata kama Cline au Claude Code inatuma jina tofauti la modeli, proxy inabatilisha kwa usanidi wa vault.

---

### Kutumia Cline katika VS Code — Mwongozo wa Kina

#### Hatua ya 1: Funga Cline

Funga **Cline** (ID: `saoudrizwan.claude-dev`) kutoka Soko la Nyongeza la VS Code.

#### Hatua ya 2: Sajili Wakala kwenye Vault

1. Fungua dashibodi ya vault (`http://IP-ya-vault:56243`)
2. Bonyeza **+ Ongeza** kwenye sehemu ya **Wakala**
3. Ingiza kama ifuatavyo:

| Sehemu | Thamani | Maelezo |
|------|----|------|
| ID | `cline_yangu` | Kitambulisho cha pekee (Kiingereza, bila nafasi) |
| Jina | `Cline Yangu` | Jina litakaloonyeshwa kwenye dashibodi |
| Aina ya Wakala | `cline` | ← Lazima uchague `cline` |
| Huduma | Chagua huduma (mfano: `google`) | |
| Modeli | Ingiza modeli (mfano: `gemini-2.5-flash`) | |

4. Bonyeza **Hifadhi** na tokeni itazalishwa kiotomatiki

#### Hatua ya 3: Unganisha na Cline

**Njia A — Kutumia Kiotomatiki (Inashauriwa)**

1. Hakikisha **proxy** ya wall-vault inaendesha kwenye mashine hiyo (`localhost:56244`)
2. Bonyeza kitufe cha **⚡ Tumia Usanidi wa Cline** kwenye kadi ya wakala kwenye dashibodi
3. Ukiona arifa "Usanidi umetumika kwa mafanikio!" imefanikiwa
4. Sasisha VS Code upya (`Ctrl+Alt+R`)

**Njia B — Usanidi wa Mikono**

Fungua mipangilio (⚙️) kwenye upau wa kando wa Cline:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://anwani-ya-proxy:56244/v1`
  - Mashine ile ile: `http://localhost:56244/v1`
  - Mashine nyingine kama seva ya mini: `http://192.168.1.20:56244/v1`
- **API Key**: Tokeni iliyotolewa kutoka vault (nakili kutoka kadi ya wakala)
- **Model ID**: Modeli iliyowekwa kwenye vault (mfano: `gemini-2.5-flash`)

#### Hatua ya 4: Thibitisha

Tuma ujumbe wowote kwenye mazungumzo ya Cline. Ikiwa ni kawaida:
- Nukta ya **kijani (● Inaendesha)** itaonyeshwa kwenye kadi ya wakala husika kwenye dashibodi ya vault
- Huduma/modeli ya sasa itaonyeshwa kwenye kadi (mfano: `google / gemini-2.5-flash`)

#### Kubadilisha Modeli

Unapotaka kubadilisha modeli ya Cline, badilisha kwenye **dashibodi ya vault**:

1. Badilisha huduma/modeli kwenye menyu ya kushuka ya kadi ya wakala
2. Bonyeza **Tumia**
3. Sasisha VS Code upya (`Ctrl+Alt+R`) — jina la modeli kwenye kijachini cha Cline litasasishwa
4. Modeli mpya itatumika kuanzia ombi linalofuata

> 💡 Kwa kweli, proxy inatambua ombi la Cline kwa tokeni na kuielekeza kwa modeli ya usanidi wa vault. Hata ikiwa husasishi VS Code, **modeli inayotumika hubadilika mara moja** — kusasisha ni kwa ajili ya kusasisha onyesho la modeli kwenye UI ya Cline.

#### Kugundua Kukatika kwa Muunganisho

Ukifunga VS Code, kadi ya wakala kwenye dashibodi ya vault itageuka njano (imecheleweshwa) baada ya takriban **sekunde 90**, na nyekundu (nje ya mtandao) baada ya **dakika 3**. (Kuanzia v0.1.18, uchunguzi wa hali kwa vipindi vya sekunde 15 uliharakisha ugunduzi wa hali ya nje ya mtandao.)

#### Utatuzi wa Matatizo

| Dalili | Sababu | Suluhisho |
|------|------|------|
| Kosa la "Muunganisho umeshindikana" kwenye Cline | Proxy haiendeshwi au anwani si sahihi | Angalia proxy kwa `curl http://localhost:56244/health` |
| Nukta ya kijani haionekani kwenye vault | API key (tokeni) haijawekwa | Bonyeza kitufe cha **⚡ Tumia Usanidi wa Cline** tena |
| Modeli ya kijachini cha Cline haibadiliki | Cline inahifadhi usanidi kwenye cache | Sasisha VS Code upya (`Ctrl+Alt+R`) |
| Jina lisilo sahihi la modeli linaonyeshwa | Hitilafu ya zamani (iliyorekebishwa v0.1.16) | Sasisha proxy hadi v0.1.16 au zaidi |

---

#### 🟣 Kitufe cha Nakili Amri ya Kusambaza — Kinatumika unapofunga kwenye mashine mpya

Kinatumika unapofunga proxy ya wall-vault kwenye kompyuta mpya na kuiunganisha na vault kwa mara ya kwanza. Bonyeza kitufe na hati nzima ya ufungaji inanakiliwa. Bandika na uitekeleze kwenye terminal ya kompyuta mpya na yafuatayo yanashughulikiwa kwa wakati mmoja:

1. Funga binary ya wall-vault (inarukwa ikiwa tayari imefungwa)
2. Usajili wa kiotomatiki wa huduma ya systemd ya mtumiaji
3. Anza huduma na uunganishe kiotomatiki na vault

> 💡 Hati ina tokeni ya wakala huu na anwani ya seva ya vault tayari imejazwa, kwa hivyo unaweza kuitekeleza mara moja baada ya kubandika bila marekebisho yoyote.

---

### Kadi ya Huduma

Kadi ya kuwasha, kuzima, au kusanidi huduma za AI za kutumia.

- Swichi ya kuzima·kuwasha kwa kila huduma
- Ukiingiza anwani ya seva ya AI ya ndani (Ollama, LM Studio, vLLM n.k. inayoendesha kwenye kompyuta yako), itagundua kiotomatiki modeli zinazopatikana.
- **Onyesho la hali ya muunganisho wa huduma ya ndani**: Nukta ya ● kando ya jina la huduma ikiwa **kijani** imeunganishwa, **kijivu** haijaunganishwa
- **Taa ya trafiki ya kiotomatiki ya huduma ya ndani** (v0.1.23+): Huduma za ndani (Ollama, LM Studio, vLLM) zinawashwa/kuzimwa kiotomatiki kulingana na upatikanaji wa muunganisho. Ukiwasha huduma, ndani ya sekunde 15 ● inaonekana kijani na kisanduku cha kuangalia kinawashwa, na ukizima huduma, inazimwa kiotomatiki. Njia ile ile na huduma za wingu (Google, OpenRouter n.k.) zinazobadilika kiotomatiki kulingana na upatikanaji wa API key.

> 💡 **Ikiwa huduma ya ndani inaendesha kwenye kompyuta nyingine**: Ingiza IP ya kompyuta hiyo kwenye sehemu ya kuingiza URL ya huduma. Mfano: `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio). Ikiwa huduma imefungwa kwa `127.0.0.1` tu badala ya `0.0.0.0`, kufikia kwa IP ya nje hakutafanya kazi, kwa hivyo angalia anwani ya kufunga kwenye usanidi wa huduma.

### Kuingiza Tokeni ya Msimamizi

Unapojaribu kutumia vipengele muhimu kama kuongeza·kufuta key kwenye dashibodi, popup ya kuingiza tokeni ya msimamizi itaonekana. Ingiza tokeni uliyoweka kwenye mchawi wa setup. Baada ya kuingiza mara moja, inabaki hadi ufunge kivinjari.

> ⚠️ **Ikiwa kushindikana kwa uthibitishaji kunazidi mara 10 ndani ya dakika 15, IP husika itazuiwa kwa muda.** Ikiwa umesahau tokeni, angalia kipengee cha `admin_token` kwenye faili ya `wall-vault.yaml`.

---

## Hali ya Kusambazwa (Multi Bot)

Wakati wa kuendesha OpenClaw kwenye kompyuta nyingi kwa wakati mmoja, hii ni usanidi ambapo **vault moja ya key inashirikiwa**. Ni rahisi kwa sababu unahitaji tu kusimamia key mahali pamoja.

### Mfano wa Usanidi

```
[Seva ya Key Vault]
  wall-vault vault    (Key Vault :56243, dashibodi)

[WSL Alpha]           [Raspberry Pi Gamma]    [Mac Mini ya Ndani]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ Usawazishaji wa SSE ↕ Usawazishaji wa SSE  ↕ Usawazishaji wa SSE
```

Bot zote zinaangalia seva ya kati ya vault, kwa hivyo ukibadilisha modeli au kuongeza key kwenye vault, inajitokeza kwenye bot zote mara moja.

### Hatua ya 1: Anza Seva ya Key Vault

Endesha kwenye kompyuta ambayo itatumika kama seva ya vault:

```bash
wall-vault vault
```

### Hatua ya 2: Sajili Kila Bot (Mteja)

Sajili mapema taarifa za kila bot inayounganishwa na seva ya vault:

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer tokeni-ya-msimamizi" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "BotA",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### Hatua ya 3: Anza Proxy kwenye Kila Kompyuta ya Bot

Endesha proxy kwa kubainisha anwani ya seva ya vault na tokeni kwenye kila kompyuta ambapo bot imefungwa:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Badilisha sehemu ya **`192.168.x.x`** na anwani halisi ya IP ya ndani ya kompyuta ya seva ya vault. Unaweza kuiangalia kwenye mipangilio ya router au kwa amri ya `ip addr`.

---

## Mpangilio wa Kuanza Kiotomatiki

Ikiwa ni kero kuwasha wall-vault kwa mikono kila ukianzisha upya kompyuta, isajili kama huduma ya mfumo. Baada ya kusajili mara moja, itaanza kiotomatiki wakati wa bootup.

### Linux — systemd (Linux nyingi)

systemd ni mfumo unaoendesha·kusimamia programu kiotomatiki kwenye Linux:

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

Ni mfumo unaosimamia utekelezaji wa kiotomatiki wa programu kwenye macOS:

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

## Doctor (Daktari)

Amri ya `doctor` ni zana ambayo **inajigundua na kujirekebishA** ikiwa wall-vault imesanidiwa vizuri.

```bash
wall-vault doctor check   # Gundua hali ya sasa (soma tu, usibadilishe kitu)
wall-vault doctor fix     # Rekebisha matatizo kiotomatiki
wall-vault doctor all     # Gundua + rekebisha kiotomatiki kwa wakati mmoja
```

> 💡 Ikiwa kitu kinaonekana si sawa, endesha `wall-vault doctor all` kwanza. Inashughulikia matatizo mengi kiotomatiki.

---

## RTK Kuokoa Tokeni

*(v0.1.24+)*

**RTK (Zana ya Kuokoa Tokeni)** inashinikiza kiotomatiki matokeo ya amri za shell zinazotekelezwa na wakala wa AI coding (kama Claude Code) ili kupunguza matumizi ya tokeni. Kwa mfano, matokeo ya mistari 15 ya `git status` yanashindiliwa kuwa muhtasari wa mistari 2.

### Matumizi ya Msingi

```bash
# Funika amri kwa wall-vault rtk na matokeo yachujwa kiotomatiki
wall-vault rtk git status          # Orodha ya faili zilizobadilishwa tu
wall-vault rtk git diff HEAD~1     # Mistari iliyobadilishwa + muktadha wa chini
wall-vault rtk git log -10         # Hash + ujumbe wa mstari mmoja kwa kila moja
wall-vault rtk go test ./...       # Majaribio yaliyoshindikana tu
wall-vault rtk ls -la              # Amri zisizotumika zinakatwa kiotomatiki
```

### Amri Zinazotumika na Athari za Kupunguza

| Amri | Njia ya Kichuja | Kiwango cha Kupunguza |
|------|----------|--------|
| `git status` | Muhtasari wa faili zilizobadilishwa tu | ~87% |
| `git diff` | Mistari iliyobadilishwa + muktadha wa mistari 3 | ~60-94% |
| `git log` | Hash + ujumbe wa mstari wa kwanza | ~90% |
| `git push/pull/fetch` | Ondoa maendeleo, muhtasari tu | ~80% |
| `go test` | Onyesha kushindikana tu, hesabu zilizopita | ~88-99% |
| `go build/vet` | Onyesha makosa tu | ~90% |
| Amri nyingine zote | Mistari 50 ya kwanza + mistari 50 ya mwisho, upeo 32KB | Inabadilika |

### Pipeline ya Kichuja cha Hatua 3

1. **Kichuja cha muundo kwa amri** — Kinaelewa muundo wa matokeo ya git, go n.k. na kuchota sehemu zenye maana tu
2. **Uchakataji wa baadaye wa regex** — Ondoa misimbo ya rangi ya ANSI, punguza mistari tupu, jumlisha mistari inayojirudia
3. **Kupitisha + kukata** — Amri zisizotumika huhifadhi mistari 50 ya kwanza/mwisho tu

### Kuunganisha na Claude Code

Unaweza kusanidi hook ya `PreToolUse` ya Claude Code ili amri zote za shell zipitie RTK kiotomatiki.

```bash
# Funga hook (huongezwa kiotomatiki kwenye settings.json ya Claude Code)
wall-vault rtk hook install
```

Au ongeza kwa mikono kwenye `~/.claude/settings.json`:

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

> 💡 **Uhifadhi wa Exit code**: RTK inarudisha exit code ya amri ya awali kama ilivyo. Ikiwa amri inashindikana (exit code ≠ 0), AI pia inagundua kushindikana kwa usahihi.

> 💡 **Kulazimisha Kiingereza**: RTK inatekeleza amri kwa `LC_ALL=C` ili kuzalisha matokeo ya Kiingereza kila wakati bila kujali mipangilio ya lugha ya mfumo. Hii inahakikisha kichuja kinafanya kazi kwa usahihi.

---

## Rejea ya Vigezo vya Mazingira

Vigezo vya mazingira ni njia ya kupitisha thamani za usanidi kwa programu. Ingiza katika muundo wa `export jina-la-kigezo=thamani` kwenye terminal, au weka kwenye faili ya huduma ya kuanza kiotomatiki ili itumike daima.

| Kigezo | Maelezo | Thamani ya Mfano |
|------|------|---------|
| `WV_LANG` | Lugha ya dashibodi | `ko`, `en`, `ja` |
| `WV_THEME` | Mandhari ya dashibodi | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | API key ya Google (nyingi kwa koma) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | API key ya OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Anwani ya seva ya vault katika hali ya kusambazwa | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Tokeni ya uthibitishaji ya mteja (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Tokeni ya msimamizi | `admin-token-here` |
| `WV_MASTER_PASS` | Nenosiri la usimbaji fiche wa API key | `my-password` |
| `WV_AVATAR` | Njia ya faili ya picha ya avatar (njia ya jamaa kutoka `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Anwani ya seva ya ndani ya Ollama | `http://192.168.x.x:11434` |

---

## Utatuzi wa Matatizo

### Wakati Proxy Haianzii

Mara nyingi ni kwa sababu bandari tayari inatumika na programu nyingine.

```bash
ss -tlnp | grep 56244   # Angalia nani anatumia bandari 56244
wall-vault proxy --port 8080   # Anza kwa nambari tofauti ya bandari
```

### Wakati Makosa ya API Key Yanatokea (429, 402, 401, 403, 582)

| Msimbo wa Kosa | Maana | Suluhisho |
|----------|------|----------|
| **429** | Maombi mengi sana (matumizi yamezidi) | Subiri kidogo au ongeza key nyingine |
| **402** | Malipo yanahitajika au krediti haitoshi | Jaza krediti kwenye huduma husika |
| **401 / 403** | Key si sahihi au haina ruhusa | Angalia tena thamani ya key na uisajili upya |
| **582** | Gateway imezidiwa (cooldown dakika 5) | Inaisha kiotomatiki baada ya dakika 5 |

```bash
# Angalia orodha na hali ya key zilizosajiliwa
curl -H "Authorization: Bearer tokeni-ya-msimamizi" http://localhost:56243/admin/keys

# Weka upya vihesabio vya matumizi ya key
curl -X POST -H "Authorization: Bearer tokeni-ya-msimamizi" http://localhost:56243/admin/keys/reset
```

### Wakati Wakala Anaonyeshwa kama "Haijaunganishwa"

"Haijaunganishwa" inamaanisha mchakato wa proxy hautumi ishara (heartbeat) kwa vault. **Haimaanishi kuwa usanidi haujahifadhiwa.** Proxy inahitaji kuendesha ukijua anwani ya seva ya vault na tokeni ili kubadilika kuwa hali ya kuunganishwa.

```bash
# Anza proxy ukibainisha anwani ya seva ya vault, tokeni, na ID ya mteja
WV_VAULT_URL=http://anwani-ya-seva-ya-vault:56243 \
WV_VAULT_TOKEN=tokeni-ya-mteja \
WV_VAULT_CLIENT_ID=id-ya-mteja \
wall-vault proxy
```

Muunganisho ukifanikiwa, itabadilika kuwa 🟢 Inaendesha kwenye dashibodi ndani ya takriban sekunde 20.

### Wakati Muunganisho wa Ollama Haifanyi Kazi

Ollama ni programu inayoendesha AI moja kwa moja kwenye kompyuta yako. Kwanza angalia ikiwa Ollama imewashwa.

```bash
curl http://localhost:11434/api/tags   # Ikiwa orodha ya modeli inaonekana, ni kawaida
export OLLAMA_URL=http://192.168.x.x:11434   # Ikiwa inaendesha kwenye kompyuta nyingine
```

> ⚠️ Ikiwa Ollama haijibu, anza kwanza Ollama kwa amri ya `ollama serve`.

> ⚠️ **Modeli kubwa ni polepole**: Modeli kubwa kama `qwen3.5:35b`, `deepseek-r1` zinaweza kuchukua dakika kadhaa kuzalisha jibu. Hata ikiwa inaonekana hakuna jibu, inaweza kuwa bado inashughulika kawaida, kwa hivyo subiri.

---

## Maelezo ya Uboreshaji wa v0.2

- `Service` ilipokea `default_model` na `allowed_models`. Modeli chaguomsingi inayolingana na huduma sasa imewekwa moja kwa moja kwenye kadi ya huduma.
- `Client.default_service` / `default_model` zimebadilishwa jina na kubadilishwa maana kuwa `preferred_service` / `model_override`. Ikiwa kizuizi kipo tupu, modeli chaguomsingi ya huduma itatumika.
- Wakati wa kuanzisha v0.2 kwa mara ya kwanza, faili ya `vault.json` iliyopo inabadilishwa kiotomatiki, na hali kabla ya ubadilishaji inahifadhwa kama `vault.json.pre-v02.{timestamp}.bak`.
- Dashibodi imebadilishwa kuwa maeneo matatu: upau wa upande wa kushoto, taifa la kadi katikati, na sehemu ya kuhariri upande wa kulia.
- Njia za Admin API hazibadilishwi, lakini vigezo vya ombi/jibu vimebadilishwa — hati za CLI za zamani zitahitaji kusasishwa ipasavyo.

---

## Vipengele Vipya vya v0.2.1

- **Upitishaji wa Multimodal (OpenAI → Gemini)**: `/v1/chat/completions` sasa inakubali aina sita za sehemu za maudhui zaidi ya `text` — `input_audio`, `input_video`, `input_image`, `input_file`, na `image_url` (data URIs pamoja na URL za nje za http(s) ≤ 5 MB). Proxy hubadilisha kila moja kuwa `inlineData` ya Gemini. Wateja wanaoendana na OpenAI kama EconoWorld wanaweza kutiririsha blobs za sauti / picha / video moja kwa moja.
- **Aina ya wakala ya EconoWorld**: `POST /agent/apply` ikiwa na `agentType: "econoworld"` inaandika mipangilio ya wall-vault ndani ya `analyzer/ai_config.json` ya mradi. `workDir` inakubali orodha ya njia wagombea zilizotenganishwa kwa koma na kubadilisha njia za kiendeshi cha Windows kuwa njia za mlimo wa WSL.
- **Gridi ya funguo + CRUD ya dashibodi**: funguo 11 zinaonyeshwa kama kadi zilizokusanywa zikiwa na slideover ya + ongeza / ✕ futa.
- **Kuongeza huduma + kupanga upya kwa kuburuta-na-kuacha**: gridi ya huduma inapata kitufe cha + ongeza pamoja na kishiko cha kuburuta (`⋮⋮`).
- **Kichwa / chini / uhuishaji wa mandhari / kibadilishi cha lugha** vimerejeshwa. Mandhari 7 (cherry/dark/light/ocean/gold/autumn/winter) huchezesha athari yao ya chembe kwenye safu nyuma ya kadi lakini juu ya mandharinyuma.
- **Uzoefu wa kufunga slideover**: kubofya nje au Esc kunafunga slideover.
- **Kiashiria cha hali cha SSE** chini ya ukurasa (kijani = imeunganishwa, chungwa = inaunganisha upya, kijivu = imekatika).

---

## Mabadiliko ya Hivi Karibuni (v0.1.16 ~ v0.1.27)

### v0.1.27 (2026-04-09)
- **Urekebishaji wa jina la modeli ya fallback ya Ollama**: Ilirekebishwa tatizo ambapo jina la modeli yenye kiambishi awali cha provider (mfano: `google/gemini-3.1-pro-preview`) lilipelekwa kwa Ollama kama lilivyo wakati wa fallback kutoka huduma nyingine. Sasa linabadilishwa kiotomatiki na kigezo cha mazingira/modeli chaguomsingi.
- **Kupunguza sana muda wa cooldown**: 429 rate limit 30min→5min, 402 malipo 1hr→30min, 401/403 24hr→6hr. Kuzuia hali ambapo key zote ziko kwenye cooldown kwa wakati mmoja na kusababisha proxy kusimama kabisa.
- **Kujaribu tena kwa lazima wakati wa cooldown kamili**: Wakati key zote ziko kwenye hali ya cooldown, key inayofunguliwa mapema zaidi inajaribiwa tena kwa lazima ili kuzuia kukataliwa kwa ombi.
- **Urekebishaji wa onyesho la orodha ya huduma**: Jibu la `/status` linaonyesha orodha halisi ya huduma iliyosawazishwa kutoka vault (kuzuia kukosekana kwa anthropic n.k.).

### v0.1.25 (2026-04-08)
- **Kugundua mchakato wa wakala**: Proxy inagundua iwapo wakala wa ndani (NanoClaw/OpenClaw) yuko hai na kuonyesha kwa taa ya trafiki ya machungwa kwenye dashibodi.
- **Uboreshaji wa kishiko cha kuburuta**: Kubadilishwa ili uweze kushika tu kutoka eneo la taa ya trafiki (●) wakati wa kupanga kadi. Haiwezekani tena kuburuta kwa bahati mbaya kutoka sehemu za kuingiza au vitufe.

### v0.1.24 (2026-04-06)
- **Amri ndogo ya RTK ya kuokoa tokeni**: `wall-vault rtk <command>` inachuja kiotomatiki matokeo ya amri za shell ili kupunguza matumizi ya tokeni ya wakala wa AI kwa 60-90%. Inajumuisha vichuja maalum kwa amri kuu kama git, go, na amri zisizotumika pia zinakatwa kiotomatiki. Inaunganishwa kwa uwazi kupitia hook ya `PreToolUse` ya Claude Code.

### v0.1.23 (2026-04-06)
- **Urekebishaji wa kubadilisha modeli ya Ollama**: Ilirekebishwa tatizo ambapo kubadilisha modeli ya Ollama kwenye dashibodi ya vault hakukujitokeza kwenye proxy. Hapo awali ilitumia tu kigezo cha mazingira (`OLLAMA_MODEL`), sasa usanidi wa vault unapewa kipaumbele.
- **Taa ya trafiki ya kiotomatiki ya huduma ya ndani**: Ollama·LM Studio·vLLM zinaanzishwa kiotomatiki zikiweza kuunganishwa, na kuzimwa kiotomatiki zikikatika. Njia ile ile na ubadilishaji wa kiotomatiki wa huduma za wingu kulingana na key.

### v0.1.22 (2026-04-05)
- **Urekebishaji wa sehemu ya content tupu inayokosekana**: Wakati modeli za thinking (gemini-3.1-pro, o1, claude thinking n.k.) zinatumia kikomo cha max_tokens kwa reasoning na haziwezi kuzalisha jibu halisi, proxy ilikuwa ikiondoa sehemu za `content`/`text` za jibu la JSON kwa `omitempty`, na kusababisha SDK za mteja za OpenAI/Anthropic kuanguka kwa kosa la `Cannot read properties of undefined (reading 'trim')`. Kubadilishwa ili kujumuisha sehemu kila wakati kama ilivyo kwenye vipimo rasmi vya API.

### v0.1.21 (2026-04-05)
- **Msaada wa modeli ya Gemma 4**: Modeli za familia ya Gemma kama `gemma-4-31b-it`, `gemma-4-26b-a4b-it` zinaweza kutumika kupitia Google Gemini API.
- **Msaada rasmi wa huduma ya LM Studio / vLLM**: Hapo awali huduma hizi zilikuwa zikiachwa nje ya uelekezaji wa proxy na kila wakati kubadilishwa na Ollama. Sasa zinaelekezwa kawaida kupitia API inayolingana na OpenAI.
- **Urekebishaji wa onyesho la huduma kwenye dashibodi**: Hata fallback ikitokea, dashibodi daima inaonyesha huduma iliyowekwa na mtumiaji.
- **Onyesho la hali ya huduma ya ndani**: Hali ya muunganisho wa huduma za ndani (Ollama, LM Studio, vLLM n.k.) inaonyeshwa kwa rangi ya nukta ya ● dashibodi inapopakiwa.
- **Kigezo cha mazingira cha kichuja cha zana**: Hali ya kupitisha zana (tools) inaweza kuwekwa kwa kigezo cha mazingira `WV_TOOL_FILTER=passthrough`.

### v0.1.20 (2026-03-28)
- **Uimarishaji wa kina wa usalama**: Kuzuia XSS (sehemu 41), kulinganisha tokeni kwa muda sawa, vizuizi vya CORS, vikomo vya ukubwa wa ombi, kuzuia kupita njia, uthibitishaji wa SSE, uimarishaji wa kikomo cha kasi, na uboreshaji wa vipengee 12 vya usalama.

### v0.1.19 (2026-03-27)
- **Kugundua Claude Code mtandaoni**: Claude Code ambayo haipiti proxy pia inaonyeshwa kama mtandaoni kwenye dashibodi.

### v0.1.18 (2026-03-26)
- **Urekebishaji wa kushikamana na huduma ya fallback**: Baada ya fallback ya muda kwa Ollama, huduma ya awali ikirejea inarudi kiotomatiki.
- **Uboreshaji wa ugunduzi wa nje ya mtandao**: Ugunduzi wa proxy iliyosimama uliharakishwa kwa uchunguzi wa hali kwa vipindi vya sekunde 15.

### v0.1.17 (2026-03-25)
- **Kupanga kadi kwa kuburuta na kudondosha**: Kadi za wakala zinaweza kuvutwa ili kubadilisha mpangilio.
- **Kitufe cha kutumia usanidi ndani ya mstari**: Kitufe cha [⚡ Tumia Usanidi] kinaonyeshwa kwa wakala walio nje ya mtandao.
- **Aina ya wakala wa cokacdir imeongezwa**.

### v0.1.16 (2026-03-25)
- **Usawazishaji wa pande mbili wa modeli**: Kubadilisha modeli ya Cline·Claude Code kwenye dashibodi ya vault kunajitokeza kiotomatiki.

---

*Kwa taarifa zaidi za API, tazama [API.md](API.md).*
