# Mwongozo wa Mtumiaji wa wall-vault
*(Ilisasishwa mara ya mwisho: 2026-04-05 — v0.1.21)*

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
10. [Marejeo ya Vigeuzi vya Mazingira](#marejeo-ya-vigeuzi-vya-mazingira)
11. [Utatuzi wa Matatizo](#utatuzi-wa-matatizo)

---

## wall-vault ni nini?

**wall-vault = Wakala wa AI (proxy) + Ghala la Funguo za API kwa ajili ya OpenClaw**

Ili kutumia huduma za AI, unahitaji **ufunguo wa API** — yaani, "kibali cha kidijitali" kinachothibitisha kwamba una ruhusa ya kutumia huduma hiyo. Kibali hiki kina kikomo cha matumizi kwa siku, na kikiachwa bila ulinzi, kinaweza kuvamiwa.

wall-vault huhifadhi vibali hivi katika ghala salama na hufanya kazi kama **wakala (proxy)** kati ya OpenClaw na huduma za AI. Kwa urahisi: OpenClaw inaunganishwa na wall-vault peke yake, na wall-vault ndiye anayeshughulikia kila kitu kingine.

Matatizo ambayo wall-vault huyatatua:

- **Mzunguko wa Otomatiki wa Funguo**: Ufunguo ukifikia kikomo au kuzuiwa kwa muda (cooldown), wall-vault hubadilika kimya kimya kwenye ufunguo unaofuata. OpenClaw inaendelea kufanya kazi bila kukatizwa.
- **Ubadilishaji Otomatiki wa Huduma (Fallback)**: Kama Google haikujibu, inabadilika kwenye OpenRouter. Kama hiyo pia haikufanya kazi, inabadilika kwenye Ollama (AI inayofanya kazi ndani ya kompyuta yako). Kikao hakikatizwi. Huduma ya awali inaporejea, inabadilika kurudi kwake kiotomatiki kuanzia ombi linalofuata (v0.1.18+).
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
        └─ Ollama (kompyuta yako, hifadhi ya mwisho)
```

---

## Ufungaji

### Linux / macOS

Fungua terminal na ubandike amri zifuatazo moja kwa moja.

```bash
# Linux (PC ya kawaida, seva — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (Mac M1/M2/M3)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Kupakua faili kutoka kwenye mtandao.
- `chmod +x` — Kufanya faili iliyopakuliwa "iweze kutekelezwa". Ukiacha hatua hii, utapata hitilafu ya "ruhusa imekataliwa".

### Windows

Fungua PowerShell (kwa haki za msimamizi) na utekeleze amri zifuatazo.

```powershell
# Kupakua
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Kuongeza kwenye PATH (itatekelezwa baada ya kuanzisha upya PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **PATH ni nini?** Ni orodha ya folda ambazo kompyuta hutafuta amri. Ukiongeza kwenye PATH, unaweza kuandika `wall-vault` na kuitekeleza kutoka folda yoyote.

### Kujenga Kutoka kwa Msimbo (kwa Waendelezaji)

Hii inatumika tu kama una mazingira ya maendeleo ya lugha ya Go imewekwa.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (toleo: v0.1.6.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Toleo lenye muhuri wa wakati wa ujenzi**: Ukijenga kwa `make build`, toleo litaundwa kiotomatiki katika muundo wa `v0.1.6.20260314.231308` unaojumuisha tarehe na saa. Ukijenga moja kwa moja kwa `go build ./...`, toleo litaonyeshwa kama `"dev"` tu.

---

## Kuanza kwa Mara ya Kwanza

### Kuendesha Mchawi wa setup

Baada ya ufungaji, lazima kwanza uendeshe **mchawi wa usanidi** kwa amri hii. Mchawi atakuuliza maswali moja moja na kukuongoza.

```bash
wall-vault setup
```

Hatua ambazo mchawi hupitia:

```
1. Kuchagua lugha (lugha 10 ikiwemo Kiswahili)
2. Kuchagua mandhari (light / dark / gold / cherry / ocean)
3. Hali ya uendeshaji — kujua kama utatumia peke yako (standalone) au pamoja na mashine nyingi (distributed)
4. Jina la boti — jina litakaloonekana kwenye dashibodi
5. Usanidi wa bandari — chaguo-msingi: proxy 56244, ghala 56243 (bonyeza Enter kama hutaki kubadilisha)
6. Kuchagua huduma ya AI — Google / OpenRouter / Ollama
7. Usanidi wa kichujio cha usalama cha zana
8. Kuweka tokeni ya msimamizi — nenosiri la kufunga vipengele vya usimamizi wa dashibodi; inaweza kuundwa kiotomatiki
9. Kuweka nenosiri la usimbuaji wa funguo za API — kwa uhifadhi salama zaidi (ni hiari)
10. Njia ya kuhifadhi faili ya usanidi
```

> ⚠️ **Kumbuka tokeni ya msimamizi vizuri.** Utahitaji baadaye unapotaka kuongeza funguo au kubadilisha mipangilio kwenye dashibodi. Ukiisahau, utahitaji kuhariri faili ya usanidi moja kwa moja.

Mchawi ukikamilika, faili ya usanidi `wall-vault.yaml` itaundwa kiotomatiki.

### Kuanzisha

```bash
wall-vault start
```

Seva mbili zifuatazo zinaanza wakati huo huo:

- **Proxy** (`http://localhost:56244`) — wakala anayeunganisha OpenClaw na huduma za AI
- **Ghala la Funguo** (`http://localhost:56243`) — usimamizi wa funguo za API na dashibodi ya wavuti

Fungua kivinjari na nenda `http://localhost:56243` kuona dashibodi mara moja.

---

## Kusajili Ufunguo wa API

Kuna njia nne za kusajili ufunguo wa API. **Kwa wanaoanza, tunapendekeza Njia ya 1 (vigeuzi vya mazingira).**

### Njia ya 1: Vigeuzi vya Mazingira (Inapendekezwa — Rahisi Zaidi)

Vigeuzi vya mazingira (environment variables) ni **maadili yaliyowekwa mapema** ambayo programu husoma wakati wa kuanza. Ingiza kama ifuatavyo kwenye terminal:

```bash
# Kusajili ufunguo wa Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Kusajili ufunguo wa OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Kuanzisha baada ya kusajili
wall-vault start
```

Kama una funguo nyingi, ziunganishe kwa mkato (,). wall-vault itatumia funguo kwa zamu kiotomatiki (mzunguko):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Kidokezo**: Amri ya `export` inatumika kwa kikao cha terminal cha sasa tu. Ili ibaki hata baada ya kuanzisha upya kompyuta, ongeza mstari huo kwenye faili `~/.bashrc` au `~/.zshrc`.

### Njia ya 2: Kiolesura cha Dashibodi (Bonyeza na Panya)

1. Fungua kivinjari na nenda `http://localhost:56243`
2. Bonyeza kitufe cha `[+ Ongeza]` kwenye kadi ya **🔑 Funguo za API** juu
3. Ingiza aina ya huduma, thamani ya ufunguo, lebo (jina la kumbukumbu), na kikomo cha siku kisha uhifadhi

### Njia ya 3: REST API (kwa Otomatiki na Hati za Script)

REST API ni njia ambayo programu zinazungumzana kupitia HTTP. Hii ni muhimu wakati wa kusajili kiotomatiki kwa hati.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer tokeni-ya-msimamizi" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "ufunguo-mkuu",
    "daily_limit": 1000
  }'
```

### Njia ya 4: Bendera ya proxy (kwa Majaribio ya Muda Mfupi)

Inatumika kujaribu kwa muda bila usajili rasmi. Ufunguo hutoweka unapozima programu.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Jinsi ya Kutumia Wakala (Proxy)

### Kutumia na OpenClaw (Madhumuni Makuu)

Hivi ndivyo unavyoweka OpenClaw ili iunganike na huduma za AI kupitia wall-vault.

Fungua faili `~/.openclaw/openclaw.json` na uongeze yaliyomo yafuatayo:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "tokeni-yako-ya-wakala",   // tokeni ya wakala wa ghala
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // bure, muktadha wa 1M
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Njia Rahisi Zaidi**: Bonyeza kitufe cha **🦞 Nakili Mipangilio ya OpenClaw** kwenye kadi ya wakala kwenye dashibodi. Kipande cha usanidi chenye tokeni na anwani tayari imejazwa kitanakiliwa kwenye ubao wa kunakili. Ubandike tu.

**`wall-vault/` mwanzoni mwa jina la mfano inaelekeza wapi?**

wall-vault hutambua kiotomatiki ni huduma gani ya AI ya kutuma ombi kulingana na jina la mfano:

| Muundo wa Mfano | Huduma Inayounganishwa |
|-----------------|------------------------|
| `wall-vault/gemini-*` | Unganisho la moja kwa moja la Google Gemini |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | Unganisho la moja kwa moja la OpenAI |
| `wall-vault/claude-*` | Anthropic kupitia OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (muktadha wa token 1M bure) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | Unganisho la OpenRouter |
| `google/jina-la-mfano`, `openai/jina-la-mfano`, `anthropic/jina-la-mfano` n.k. | Unganisho la moja kwa moja la huduma husika |
| `custom/google/jina-la-mfano`, `custom/openai/jina-la-mfano` n.k. | Huondoa sehemu ya `custom/` kisha inarejesha |
| `jina-la-mfano:cloud` | Huondoa `:cloud` kisha inaunganisha kwenye OpenRouter |

> 💡 **Muktadha (context) ni nini?** Ni kiasi cha mazungumzo ambacho AI inaweza kukumbuka kwa wakati mmoja. 1M (tokeni milioni moja) inamaanisha mazungumzo marefu sana au hati ndefu zinaweza kushughulikiwa mara moja.

### Unganisho la Moja kwa Moja la Muundo wa Gemini API (Uoanifu na Zana Zilizopo)

Kama una zana inayotumia moja kwa moja Google Gemini API, badilisha anwani yake tu kwenye wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Au kama zana inakuruhusu kubainisha URL moja kwa moja:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Kutumia na OpenAI SDK (Python)

Unaweza pia kuunganisha wall-vault katika msimbo wa Python unaotumia AI. Badilisha `base_url` tu:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # wall-vault inashughulikia funguo za API
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # ingiza kwa muundo wa provider/mfano
    messages=[{"role": "user", "content": "Habari"}]
)
```

### Kubadilisha Mfano Wakati wa Uendeshaji

Ili kubadilisha mfano wa AI unaotumika wakati wall-vault tayari inafanya kazi:

```bash
# Ombi la moja kwa moja kwa proxy kubadilisha mfano
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# Katika hali ya kusambazwa (boti nyingi), badilisha kwenye seva ya ghala → itaonyeshwa mara moja kupitia SSE
curl -X PUT http://localhost:56243/admin/clients/id-ya-boti-yangu \
  -H "Authorization: Bearer tokeni-ya-msimamizi" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Kuangalia Orodha ya Mifano Inayopatikana

```bash
# Kuona orodha yote
curl http://localhost:56244/api/models | python3 -m json.tool

# Kuona mifano ya Google tu
curl "http://localhost:56244/api/models?service=google"

# Kutafuta kwa jina (mfano: mifano yenye "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Muhtasari wa Mifano Mikuu kwa Huduma:**

| Huduma | Mifano Mikuu |
|--------|-------------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ (Hunter Alpha muktadha wa 1M bure, DeepSeek R1/V3, Qwen 2.5 n.k.) |
| Ollama | Hugundua kiotomatiki seva ya ndani iliyowekwa kwenye kompyuta yako |

---

## Dashibodi ya Ghala la Funguo

Fungua kivinjari na nenda `http://localhost:56243` kuona dashibodi.

**Mpangilio wa Skrini:**
- **Upau wa Juu Uliowekwa (topbar)**: Nembo, kichaguzi cha lugha na mandhari, hali ya muunganisho wa SSE
- **Gridi ya Kadi**: Kadi za wakala, huduma, na funguo za API zimepangwa kama vigae

### Kadi ya Funguo za API

Kadi inayokuruhusu kusimamia funguo zote zilizosajiliwa kwa mtazamo mmoja.

- Inaonyesha orodha ya funguo imegawanywa kwa huduma.
- `today_usage`: Idadi ya tokeni zilizoshughulikiwa kwa mafanikio leo (idadi ya maneno AI iliyosoma na kuandika)
- `today_attempts`: Jumla ya maombi ya leo (pamoja na yaliyofanikiwa + yaliyoshindwa)
- Tumia kitufe cha `[+ Ongeza]` kusajili ufunguo mpya, na `✕` kufuta ufunguo.

> 💡 **Tokeni (token) ni nini?** Ni kitengo ambacho AI hutumia kushughulikia maandishi. Inakadiria kuwa takriban neno moja la Kiingereza, au herufi 1–2 za Kiswahili. Ada za API kawaida huhesabiwa kulingana na idadi ya tokeni.

### Kadi ya Wakala

Kadi inayoonyesha hali ya boti (wakala) zilizounganishwa kwenye proxy ya wall-vault.

**Hali ya muunganisho inaonyeshwa kwa hatua 4:**

| Alama | Hali | Maana |
|-------|------|-------|
| 🟢 | Inafanya kazi | Proxy inafanya kazi kawaida |
| 🟡 | Imechelewa | Inajibu lakini polepole |
| 🔴 | Nje ya mtandao | Proxy haijibui |
| ⚫ | Haijaunganishwa / imezimwa | Proxy haijawahi kuunganishwa na ghala au imezimwa |

**Mwongozo wa Vitufe vya Chini ya Kadi ya Wakala:**

Unapoweka wakala, ukibainisha **aina ya wakala**, vitufe vya urahisi vinavyofaa kwa aina hiyo vitaonekana kiotomatiki.

---

#### 🔘 Kitufe cha Nakili Mipangilio — Huunda Mipangilio ya Muunganisho Kiotomatiki

Ukibonyeza kitufe, kipande cha usanidi chenye tokeni ya wakala, anwani ya proxy, na taarifa za mfano tayari imejazwa kitanakiliwa kwenye ubao wa kunakili. Bandika tu kwenye eneo linalofaa kwenye jedwali hapa chini ili kumaliza mipangilio ya muunganisho.

| Kitufe | Aina ya Wakala | Eneo la Kubandika |
|--------|---------------|-------------------|
| 🦞 Nakili Mipangilio ya OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Nakili Mipangilio ya NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Nakili Mipangilio ya Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Nakili Mipangilio ya Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Nakili Mipangilio ya VSCode | `vscode` | `~/.continue/config.json` |

**Mfano — Kama aina ni Claude Code, hii ndiyo itakayonakiliwa:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.0.6:56244/v1",
  "apiKey": "tokeni-ya-wakala-huyu"
}
```

**Mfano — Kama aina ni VSCode (Continue):**

```json
// ~/.continue/config.json
{
  "models": [{
    "title": "wall-vault proxy",
    "provider": "openai",
    "model": "gemini-2.5-flash",
    "apiBase": "http://192.168.0.6:56244/v1",
    "apiKey": "tokeni-ya-wakala-huyu"
  }]
}
```

**Mfano — Kama aina ni Cursor:**

```
Base URL : http://192.168.0.6:56244/v1
API Key  : tokeni-ya-wakala-huyu

// Au kwa vigeuzi vya mazingira:
OPENAI_BASE_URL=http://192.168.0.6:56244/v1
OPENAI_API_KEY=tokeni-ya-wakala-huyu
```

> ⚠️ **Kama kunakili kwenye ubao wa kunakili hakufanyi kazi**: Sera ya usalama ya kivinjari inaweza kuzuia unakili. Kama sanduku la maandishi litafunguka, bonyeza Ctrl+A kuchagua yote kisha Ctrl+C kunakili.

---

#### ⚡ Kitufe cha Kutumia Mipangilio Kiotomatiki — Bonyeza Mara Moja, Usanidi Uliokamilika

Kama aina ya wakala ni `cline`, `claude-code`, `openclaw`, `nanoclaw`, kitufe cha **⚡ Tumia Mipangilio** kitaonekana kwenye kadi ya wakala. Ukibonyeza kitufe hiki, faili ya mipangilio ya ndani ya wakala husasishwa kiotomatiki.

| Kitufe | Aina ya Wakala | Faili Inayosasishwa |
|--------|---------------|---------------------|
| ⚡ Tumia Mipangilio ya Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Tumia Mipangilio ya Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Tumia Mipangilio ya OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Tumia Mipangilio ya NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Kitufe hiki kinatuma ombi kwa **localhost:56244** (proxy ya ndani). Proxy lazima iwe inafanya kazi kwenye kompyuta hiyo ili kitufe kifanye kazi.

---

#### 🔀 Buruta na Udondoshe ili Kupanga Kadi (v0.1.17)

Unaweza **kuburuta** kadi za wakala kwenye dashibodi ili kuzipanga upya kwa mpangilio unaotaka.

1. Bonyeza na ushikilie kadi ya wakala kisha uiburuze
2. Idondoshe kwenye kadi katika nafasi unayotaka — mpangilio utabadilika mara moja
3. Mpangilio mpya **utahifadhiwa kwenye seva mara moja** na utabaki hata ukisasisha ukurasa

> 💡 Vifaa vya kugusa (simu/kompyuta kibao) bado havitumiki kwa sasa. Tafadhali tumia kivinjari cha kompyuta ya mezani.

---

#### 🔄 Usawazishaji wa Mfano wa Pande Mbili (v0.1.16)

Ukibadilisha mfano wa wakala kwenye dashibodi ya ghala, mipangilio ya ndani ya wakala huyo itasasishwa kiotomatiki.

**Kwa Cline:**
- Kubadilisha mfano kwenye ghala → tukio la SSE → proxy inasasisha sehemu ya mfano katika `globalState.json`
- Sehemu zinazosasishwa: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` na ufunguo wa API hazibadilishwi
- **Inahitajika kupakia upya VS Code (`Ctrl+Alt+R` au `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - Kwa sababu Cline haisomi upya faili ya mipangilio wakati inafanya kazi

**Kwa Claude Code:**
- Kubadilisha mfano kwenye ghala → tukio la SSE → proxy inasasisha sehemu ya `model` katika `settings.json`
- Inatafuta njia kiotomatiki kwa WSL na Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Upande wa pili (wakala → ghala):**
- Wakala (Cline, Claude Code n.k.) anapotuma ombi kupitia proxy, proxy inajumuisha taarifa za huduma·mfano wa mteja kwenye heartbeat
- Kadi ya wakala kwenye dashibodi ya ghala inaonyesha huduma/mfano inayotumiwa kwa wakati halisi

> 💡 **Muhimu**: Proxy inatambua wakala kwa tokeni ya Authorization katika ombi, na kuelekeza kiotomatiki kwenye huduma/mfano iliyowekwa kwenye ghala. Hata kama Cline au Claude Code inatuma jina tofauti la mfano, proxy itabadilisha na mipangilio ya ghala.

---

### Kutumia Cline kwenye VS Code — Mwongozo wa Kina

#### Hatua ya 1: Kufunga Cline

Funga **Cline** (ID: `saoudrizwan.claude-dev`) kutoka VS Code Extension Marketplace.

#### Hatua ya 2: Kusajili Wakala kwenye Ghala

1. Fungua dashibodi ya ghala (`http://IP-ya-ghala:56243`)
2. Bonyeza **+ Ongeza** katika sehemu ya **wakala**
3. Ingiza taarifa zifuatazo:

| Sehemu | Thamani | Maelezo |
|--------|---------|---------|
| ID | `cline_yangu` | Kitambulisho cha kipekee (herufi za Kiingereza, bila nafasi) |
| Jina | `Cline Yangu` | Jina litakaloonekana kwenye dashibodi |
| Aina ya Wakala | `cline` | ← Lazima uchague `cline` |
| Huduma | Chagua huduma ya kutumia (mfano: `google`) | |
| Mfano | Ingiza mfano wa kutumia (mfano: `gemini-2.5-flash`) | |

4. Bonyeza **Hifadhi** na tokeni itaundwa kiotomatiki

#### Hatua ya 3: Kuunganisha Cline

**Njia A — Kutumia Kiotomatiki (Inapendekezwa)**

1. Hakikisha wall-vault **proxy** inafanya kazi kwenye kompyuta hiyo (`localhost:56244`)
2. Bonyeza kitufe cha **⚡ Tumia Mipangilio ya Cline** kwenye kadi ya wakala kwenye dashibodi
3. Ujumbe "Mipangilio imetumika!" ukionekana, umefanikiwa
4. Pakia upya VS Code (`Ctrl+Alt+R`)

**Njia B — Kuweka kwa Mkono**

Fungua mipangilio (⚙️) kwenye upau wa pembeni wa Cline:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://anwani-ya-proxy:56244/v1`
  - Kompyuta hiyo hiyo: `http://localhost:56244/v1`
  - Kompyuta nyingine kama Mac Mini: `http://192.168.0.6:56244/v1`
- **API Key**: tokeni iliyotolewa kutoka ghala (nakili kutoka kadi ya wakala)
- **Model ID**: mfano uliowekwa kwenye ghala (mfano: `gemini-2.5-flash`)

#### Hatua ya 4: Kuthibitisha

Tuma ujumbe wowote kwenye mazungumzo ya Cline. Ikiwa inafanya kazi vizuri:
- Kadi ya wakala kwenye dashibodi ya ghala itaonyesha **nukta ya kijani (● Inafanya kazi)**
- Kadi itaonyesha huduma/mfano wa sasa (mfano: `google / gemini-2.5-flash`)

#### Kubadilisha Mfano

Unapotaka kubadilisha mfano wa Cline, badilisha kwenye **dashibodi ya ghala**:

1. Badilisha orodha ya huduma/mfano kwenye kadi ya wakala
2. Bonyeza **Tumia**
3. Pakia upya VS Code (`Ctrl+Alt+R`) — jina la mfano kwenye sehemu ya chini ya Cline litasasishwa
4. Ombi linalofuata litatumia mfano mpya

> 💡 Kwa kweli, proxy inatambua ombi la Cline kwa tokeni na kuelekeza kwenye mfano uliowekwa kwenye ghala. Hata bila kupakia upya VS Code, **mfano unaotumika hubadilishwa mara moja** — kupakia upya ni tu kusasisha jina la mfano kwenye UI ya Cline.

#### Kugundua Kukatika kwa Muunganisho

Ukifunga VS Code, kadi ya wakala kwenye dashibodi ya ghala itabadilika kuwa ya njano (imechelewa) baada ya **sekunde 90**, na nyekundu (nje ya mtandao) baada ya **dakika 3**. (Tangu v0.1.18, ukaguzi wa hali kila sekunde 15 unawezesha kugundua hali ya nje ya mtandao haraka zaidi.)

#### Utatuzi wa Matatizo

| Dalili | Sababu | Suluhisho |
|--------|--------|-----------|
| Cline inaonyesha "muunganisho umeshindwa" | Proxy haifanyi kazi au anwani si sahihi | Angalia proxy kwa `curl http://localhost:56244/health` |
| Nukta ya kijani haionekani kwenye ghala | Ufunguo wa API (tokeni) haujawekwa | Bonyeza kitufe cha **⚡ Tumia Mipangilio ya Cline** tena |
| Jina la mfano kwenye sehemu ya chini ya Cline hailabadiliki | Cline inakasha mipangilio | Pakia upya VS Code (`Ctrl+Alt+R`) |
| Jina la mfano lisilo sahihi linaonyeshwa | Hitilafu ya zamani (ilirekebishwa katika v0.1.16) | Sasisha proxy hadi v0.1.16 au zaidi |

---

#### 🟣 Kitufe cha Nakili Amri za Usambazaji — Inatumika Wakati wa Kufunga kwenye Mashine Mpya

Inatumika unapofunga proxy ya wall-vault kwa mara ya kwanza kwenye kompyuta mpya na kuiunganisha na ghala. Ukibonyeza kitufe, hati ya ufungaji nzima itanakiliwa. Ibandike kwenye terminal ya kompyuta mpya na uitekeleze. Hatua zifuatazo zitashughulikiwa kwa wakati mmoja:

1. Ufungaji wa faili ya binary ya wall-vault (itarukwa kama tayari imewekwa)
2. Usajili wa otomatiki wa huduma ya mtumiaji ya systemd
3. Kuanzisha huduma na kuunganisha kiotomatiki na ghala

> 💡 Hati ina tokeni ya wakala huyu na anwani ya seva ya ghala tayari imejazwa, kwa hivyo unaweza kuitekeleza moja kwa moja bila mabadiliko yoyote baada ya kubandika.

---

### Kadi ya Huduma

Kadi ya kuwasha/kuzima na kusanidi huduma za AI unazotaka kutumia.

- Swichi ya kuwasha/kuzima kwa kila huduma
- Ukiingiza anwani ya seva ya AI ya ndani (Ollama, LM Studio, vLLM n.k. inayofanya kazi kwenye kompyuta yako), mifano inayopatikana itapatikana kiotomatiki.
- **Hali ya Muunganisho wa Huduma za Ndani**: Nukta ● karibu na jina la huduma ikiwa **ya kijani** inamaanisha imeunganishwa, **ya kijivu** inamaanisha haijaungana.
- **Kiashirio cha hali ya huduma za ndani**: Wakati ukifungua ukurasa, kama huduma za ndani (kama Ollama) zinafanya kazi, nukta ● inageuka kijani — lakini hali ya kisanduku cha kuangalia haibadilishwi.

> 💡 **Kama huduma ya ndani inafanya kazi kwenye kompyuta nyingine**: Ingiza IP ya kompyuta hiyo kwenye sehemu ya URL ya huduma. Mfano: `http://192.168.0.6:11434` (Ollama), `http://192.168.0.6:1234` (LM Studio)

### Kuingiza Tokeni ya Msimamizi

Unapojaribu kutumia vipengele muhimu kama kuongeza au kufuta funguo kwenye dashibodi, sanduku la kuingiza tokeni ya msimamizi litaonekana. Ingiza tokeni uliyoweka wakati wa mchawi wa setup. Ukisha kuingiza, itabaki mpaka ufunge kivinjari.

> ⚠️ **Kama majaribio ya uthibitisho yanazidi 10 ndani ya dakika 15, IP hiyo itazuiwa kwa muda.** Ukisahau tokeni yako, angalia kipengele cha `admin_token` kwenye faili `wall-vault.yaml`.

---

## Hali ya Kusambazwa (Boti Nyingi)

Unapoendesha OpenClaw kwenye kompyuta nyingi kwa wakati mmoja, unaweza kusanidi **ghala moja la funguo kushirikiwa**. Hii inafanya usimamizi wa funguo uwe rahisi kwa sababu unafanywa mahali pamoja tu.

### Mfano wa Usanidi

```
[Seva ya Ghala la Funguo]
  wall-vault vault    (ghala la funguo :56243, dashibodi)

[WSL Alpha]              [Raspberry Pi Gamma]    [Mac Mini ya Ndani]
  wall-vault proxy         wall-vault proxy         wall-vault proxy
  openclaw TUI             openclaw TUI             openclaw TUI
  ↕ SSE usawazishaji       ↕ SSE usawazishaji       ↕ SSE usawazishaji
```

Boti zote zinaangalia seva ya ghala iliyo katikati, kwa hivyo ukibadilisha mfano au kuongeza ufunguo kwenye ghala, mabadiliko yataonyeshwa mara moja kwenye boti zote.

### Hatua ya 1: Kuanzisha Seva ya Ghala la Funguo

Tekeleza kwenye kompyuta utakayoitumia kama seva ya ghala:

```bash
wall-vault vault
```

### Hatua ya 2: Kusajili Kila Boti (Mteja)

Sajili mapema taarifa za kila boti itakayounganika kwenye seva ya ghala:

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

### Hatua ya 3: Kuanzisha Proxy kwenye Kila Kompyuta ya Boti

Kwenye kila kompyuta yenye boti, tekeleza proxy ukibainisha anwani ya seva ya ghala na tokeni:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 **`192.168.x.x`** badilisha na IP halisi ya ndani ya kompyuta ya seva ya ghala. Unaweza kuipata kwenye mipangilio ya router au kwa amri ya `ip addr`.

---

## Kuanzisha Kiotomatiki

Kama ni kuchosha kuanzisha wall-vault kwa mkono kila wakati unaowasha upya kompyuta, isajili kama huduma ya mfumo. Ukisajili mara moja, itaanza kiotomatiki wakati wa kuanza kwa mfumo.

### Linux — systemd (Linuxes Nyingi)

systemd ni mfumo wa Linux wa kuanzisha na kusimamia programu kiotomatiki:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Kuona kumbukumbu:

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
2. Kwenye PowerShell yenye haki za msimamizi:

```powershell
wall-vault doctor deploy windows
```

---

## Daktari (Doctor)

Amri ya `doctor` ni **zana inayojitathmini na kujirekebishea** ili kuhakikisha wall-vault imesanidiwa vizuri.

```bash
wall-vault doctor check   # Kutambua hali ya sasa (kusoma tu, haibadilishi chochote)
wall-vault doctor fix     # Kurekebisha matatizo kiotomatiki
wall-vault doctor all     # Kutambua + kurekebisha kiotomatiki mara moja
```

> 💡 Kama kitu kinaonekana si sawa, jaribu kwanza `wall-vault doctor all`. Matatizo mengi yanasuluhiwa kiotomatiki.

---

## Marejeo ya Vigeuzi vya Mazingira

Vigeuzi vya mazingira ni njia ya kupitisha maadili ya usanidi kwenye programu. Ingiza kwa muundo wa `export JINA_LA_KIGEUZIO=thamani` kwenye terminal, au viweke kwenye faili ya huduma ya uanzishaji wa kiotomatiki ili viwe hai kila wakati.

| Kigeuzio | Maelezo | Thamani ya Mfano |
|----------|---------|------------------|
| `WV_LANG` | Lugha ya dashibodi | `ko`, `en`, `ja` |
| `WV_THEME` | Mandhari ya dashibodi | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Funguo za Google API (nyingi zinaachana na mkato) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Ufunguo wa OpenRouter API | `sk-or-v1-...` |
| `WV_VAULT_URL` | Anwani ya seva ya ghala katika hali ya kusambazwa | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Tokeni ya uthibitisho ya mteja (boti) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Tokeni ya msimamizi | `admin-token-here` |
| `WV_MASTER_PASS` | Nenosiri la usimbuaji wa funguo za API | `my-password` |
| `WV_AVATAR` | Njia ya faili ya picha ya avatar (njia ya jamaa kulingana na `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Anwani ya seva ya ndani ya Ollama | `http://192.168.x.x:11434` |

---

## Utatuzi wa Matatizo

### Proxy Haianzii

Mara nyingi bandari tayari inatumika na programu nyingine.

```bash
ss -tlnp | grep 56244   # Angalia ni nani anayetumia bandari 56244
wall-vault proxy --port 8080   # Anza kwa nambari nyingine ya bandari
```

### Hitilafu za Ufunguo wa API (429, 402, 401, 403, 582)

| Nambari ya Hitilafu | Maana | Jinsi ya Kushughulikia |
|--------------------|-------|------------------------|
| **429** | Maombi mengi sana (matumizi yamezidi) | Subiri kidogo au ongeza ufunguo mwingine |
| **402** | Malipo yanahitajika au salio halioshi | Jaza upya salio kwenye huduma husika |
| **401 / 403** | Ufunguo si sahihi au huna ruhusa | Hakikisha thamani ya ufunguo na usajili upya |
| **582** | Mzigo mkubwa wa lango (cooldown ya dakika 5) | Itafunguliwa kiotomatiki baada ya dakika 5 |

```bash
# Angalia orodha ya funguo zilizosajiliwa na hali zake
curl -H "Authorization: Bearer tokeni-ya-msimamizi" http://localhost:56243/admin/keys

# Anza upya kihesabu cha matumizi ya funguo
curl -X POST -H "Authorization: Bearer tokeni-ya-msimamizi" http://localhost:56243/admin/keys/reset
```

### Wakala Anaonyeshwa kama "Haijaunganishwa"

"Haijaunganishwa" inamaanisha mchakato wa proxy hauitumii ishara (heartbeat) kwenye ghala. **Haimaanishi mipangilio haijahifadhiwa.** Proxy lazima ianzishwe ikijua anwani ya seva ya ghala na tokeni ili hali ibadilike hadi imeunganishwa.

```bash
# Anzisha proxy ukibainisha anwani ya seva ya ghala, tokeni, na ID ya mteja
WV_VAULT_URL=http://anwani-ya-seva-ya-ghala:56243 \
WV_VAULT_TOKEN=tokeni-ya-mteja \
WV_VAULT_CLIENT_ID=id-ya-mteja \
wall-vault proxy
```

Muunganisho ukifanikiwa, dashibodi itaonyesha 🟢 Inafanya kazi ndani ya sekunde 20 takriban.

### Ollama Haiunganiki

Ollama ni programu inayotekeleza AI moja kwa moja ndani ya kompyuta yako. Kwanza hakikisha Ollama imewashwa.

```bash
curl http://localhost:11434/api/tags   # Kama orodha ya mifano itaonekana, ni kawaida
export OLLAMA_URL=http://192.168.x.x:11434   # Kama inafanya kazi kwenye kompyuta nyingine
```

> ⚠️ Kama Ollama haijibu, ianzishe kwanza kwa amri ya `ollama serve`.

> ⚠️ **Mifano mikubwa inajibu polepole**: Mifano mikubwa kama `qwen3.5:35b` au `deepseek-r1` inaweza kuchukua dakika kadhaa kuunda jibu. Hata kama inaonekana kama haijibu, inaweza kuwa inafanya kazi kawaida — subiri.

---

## Mabadiliko ya Hivi Karibuni (v0.1.16 ~ v0.1.21)

### v0.1.21 (2026-04-05)
- **Msaada wa mifano ya Gemma 4**: Mifano ya Gemma (gemma-4-31b-it, gemma-4-26b-a4b-it) sasa inaelekezwa kupitia Google Gemini API.
- **Msaada wa LM Studio / vLLM**: Huduma hizi za ndani sasa zinatumwa kwa usahihi badala ya kurudi nyuma kwa Ollama.
- **Marekebisho ya dashibodi**: Daima inaonyesha huduma iliyosanidiwa, si huduma ya akiba.
- **Kisanduku cha kuangalia cha huduma za ndani kinahifadhiwa**: Dashibodi haizimi tena huduma za ndani kiotomatiki ukurasa unapopakia.
- **Kigezo cha mazingira cha kichujio cha zana**: Msaada wa `WV_TOOL_FILTER=passthrough`.

### v0.1.20 (2026-03-28)
- **Uimarishaji wa usalama wa kina**: Kuzuia XSS (pointi 41), ulinganisho wa tokeni kwa muda uliowekwa, vizuizi vya CORS, vikomo vya ukubwa wa maombi na zaidi.

### v0.1.19 (2026-03-27)
- **Kugundua Claude Code mtandaoni**: Claude Code inaonyeshwa mtandaoni kwenye dashibodi hata inapopita proksi.

### v0.1.18 (2026-03-26)
- **Marekebisho ya urejeshaji wa akiba**: Inarudi kiotomatiki kwa huduma inayopendekezwa inapopatikana.
- **Kugundua nje ya mtandao kuliboreshwa**: Uchunguzi wa hali kila sekunde 15.

### v0.1.17 (2026-03-25)
- **Kupanga upya kadi kwa kuvuta na kudondosha**.
- **Vitufe vya kutumia ndani ya mstari kwa mawakala ambao hawajaunganishwa**.
- **Aina ya wakala wa cokacdir imeongezwa**.

### v0.1.16 (2026-03-25)
- **Usawazishaji wa mifano wa njia mbili** kwa Cline na Claude Code.

---

*Kwa taarifa zaidi za API, angalia [API.md](API.md).*
