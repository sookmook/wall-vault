# Umhlahlandlela Womsebenzisi we-wall-vault
*(Last updated: 2026-04-08 — v0.1.25)*

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
10. [I-RTK Ukonga Amathokheni](#i-rtk-ukonga-amathokheni)
11. [Izinguquko Zemvelo](#izinguquko-zemvelo)
12. [Ukuxazulula Izinkinga](#ukuxazulula-izinkinga)

---

## Yini i-wall-vault?

**wall-vault = Ummeleli we-AI (Proxy) + Isikhwama Sokhiye be-API ye-OpenClaw**

Ukuze usebenzise amasevisi e-AI, udinga **ukhiye we-API**. Ukhiye we-API unjenge **khadi lokungena ledijithali** elifakazela ukuthi "lo muntu unelungelo lokusebenzisa le sevisi". Kodwa-ke, le khadi lokungena linemkhawulo wokusebenzisa kwansuku zonke, futhi uma ungaliphathanga kahle, lingavezwa.

i-wall-vault igcina la makhadi okungena endaweni ephephile futhi isebenza njenge **mmeleli (proxy)** phakathi kwe-OpenClaw namasevisi e-AI. Ngamazwi alula, i-OpenClaw idinga ukuxhumana ne-wall-vault kuphela, konke okunye okuyinkimbinkimbi i-wall-vault izokukusingatha.

Izinkinga i-wall-vault ezixazululayo:

- **Ukushintshashintsha Okhiye be-API Ngokuzenzakalela**: Uma ukhiye owodwa ufinyelela umkhawulo noma uvinjwa okwesikhashana (cooldown), ishintsha ngokuthula kuya kukhiye olandelayo. I-OpenClaw iqhubeka isebenza ngaphandle kokuphazamiseka.
- **Ukushintsha Amasevisi Ngokuzenzakalela (Fallback)**: Uma i-Google ingaphenduli, ishintshela ku-OpenRouter, nalokho kungasebenzi, ishintshela ku-Ollama/LM Studio/vLLM (i-AI yendawo) efakwe kukhompuyutha yakho. Iseshini ayiphuki. Uma isevisi yokuqala ibuya, izicelo ezilandelayo zibuyela ngokuzenzakalela (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Ukuvumelanisa Ngesikhathi Sangempela (SSE)**: Uma ushintsha imodeli kudashubhodi ye-vault, ibonakala kusikrini se-OpenClaw phakathi kwamasekhondi angu-1-3. I-SSE (Server-Sent Events) ubuchwepheshe lapho iseva isunduza izinguquko kumakhasimende ngesikhathi sangempela.
- **Izaziso Zangempela**: Izehlakalo ezinjengokuphela kokhiye noma ukwehluleka kwesevisi ziboniswa ngokushesha ngezansi kwe-TUI (isikrini setheminali) ye-OpenClaw.

> 💡 **I-Claude Code, i-Cursor, i-VS Code** nazo zingaxhunyaniswa, kodwa inhloso yokuqala ye-wall-vault ukusetshenziswa ne-OpenClaw.

```
OpenClaw (TUI isikrini setheminali)
        │
        ▼
  wall-vault proxy (:56244)   ← ukuphatha okhiye, ukuqondisa, i-fallback, izehlakalo
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (amamodeli angu-340+)
        ├─ Ollama / LM Studio / vLLM (ikhompuyutha yakho, isiphephelo sokugcina)
        └─ OpenAI / Anthropic API
```

---

## Ukufakela

### Linux / macOS

Vula itheminali ubese unamathisela imiyalo engezansi.

```bash
# Linux (i-PC evamile, iseva — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Ilanda ifayili ku-inthanethi.
- `chmod +x` — Yenza ifayili elandiwe "ikwazi ukusebenza". Uma udlula le sinyathelo, uzothola iphutha elithi "awunalo imvume".

### Windows

Vula i-PowerShell (njengomphathi) bese usebenzisa imiyalo elandelayo.

```powershell
# Landa
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Engeza ku-PATH (isebenza ngemuva kokuqala kabusha i-PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **Yini i-PATH?** Uhlu lwamafolda lapho ikhompuyutha ifuna khona imiyalo. Udinga ukungeza ku-PATH ukuze ukwazi ukuthayipha `wall-vault` bese uyisebenzisa kunoma iliphi ifolda.

### Ukwakha Kusuka Kumthombo Wekhodi (okwabathuthukisi)

Lokhu kusebenza kuphela uma unemvelo yokuthuthukisa yolimi lwe-Go efakiwe.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (inguqulo: v0.1.25.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Inguqulo yesitembu sesikhathi sokwakha**: Uma wakha nge-`make build`, inguqulo izakhiwa ngokuzenzakalela ngefomethi ehlanganisa usuku nesikhathi njengo-`v0.1.25.20260408.022325`. Uma wakha ngokuqondile nge-`go build ./...`, inguqulo izoboniswa njengo-`"dev"` kuphela.

---

## Ukuqala Okokuqala

### Ukusebenzisa I-Setup Wizard

Ngemuva kokufakela, qiniseka ukusebenzisa **i-wizard yokusetha** ngomyalo olandelayo. I-wizard izokubuza imibuzo ngayinye ngayinye bese ikuqondisa.

```bash
wall-vault setup
```

Izinyathelo i-wizard ezilandelayo yilezi:

```
1. Khetha ulimi (izilimi ezingu-10 kuhlanganisa isiZulu)
2. Khetha indikimba (light / dark / gold / cherry / ocean)
3. Imodi yokusebenza — khetha ukuthi uzosebenzisa wedwa (standalone) noma kumishini eminingi (distributed)
4. Faka igama le-bot — igama elizoboniswa kudashubhodi
5. Izilungiselelo zechweba — okuzenzakalelayo: proxy 56244, vault 56243 (cindezela Enter uma ungadingi ukushintsha)
6. Khetha amasevisi e-AI — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Izilungiselelo zesihlungi samathuluzi okuphepha
8. Setha ithokheni yomphathi — iphasiwedi yokukhiya izici zokuphatha zedashubhodi. Ungakwazi futhi ukuyenza ngokuzenzakalela
9. Setha iphasiwedi yokubhala ngekhodi ukhiye we-API — uma ufuna ukugcina okhiye ngokuphephile okwengeziwe (okungenasidingo)
10. Indlela yokugcina ifayili yezilungiselelo
```

> ⚠️ **Qiniseka ukukhumbula ithokheni yomphathi.** Uzoyidinga kamuva uma ungeza okhiye noma ushintsha izilungiselelo kudashubhodi. Uma uyilahla, uzodinga ukuhlela ifayili yezilungiselelo ngokuqondile.

Ngemuva kokuqeda i-wizard, ifayili yezilungiselelo ethi `wall-vault.yaml` izakhiwa ngokuzenzakalela.

### Ukusebenzisa

```bash
wall-vault start
```

Amaseva amabili aqala ngesikhathi esisodwa:

- **I-Proxy** (`http://localhost:56244`) — ummeleli oxhuma i-OpenClaw namasevisi e-AI
- **Isikhwama Sokhiye** (`http://localhost:56243`) — ukuphatha ukhiye we-API nedashubhodi yewebhu

Vula `http://localhost:56243` kubhrawuza yakho ukuze ubone idashubhodi ngokushesha.

---

## Ukubhalisa Ukhiye we-API

Kunezindlela ezine zokubhalisa ukhiye we-API. **Kwabaqalayo, Indlela 1 (izinguquko zemvelo) iyatuswa**.

### Indlela 1: Izinguquko Zemvelo (Ituswa — Elula Kakhulu)

Izinguquko zemvelo **amanani asethwe ngaphambili** uhlelo oluwafundayo uma luqala. Thayipha kutheminali kanje.

```bash
# Bhalisa ukhiye we-Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Bhalisa ukhiye we-OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Sebenzisa ngemuva kokubhalisa
wall-vault start
```

Uma unokhiye abaningi, baxhume ngokhefana (,). i-wall-vault izosebenzisa okhiye ngokulandelana ngokuzenzakalela (round robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Icebiso**: Umyalo ka-`export` usebenza kuphela kwiseshini yetheminali yamanje. Ukuze uqhubeke ngisho ngemuva kokuqala kabusha ikhompuyutha, engeza lo mqolo kufayili ye-`~/.bashrc` noma `~/.zshrc`.

### Indlela 2: Idashubhodi ye-UI (Cindezela Ngemawusi)

1. Vakashela `http://localhost:56243` kubhrawuza
2. Ekhadini ye-**🔑 Okhiye be-API** phezulu, cindezela inkinobho ka-`[+ Engeza]`
3. Faka uhlobo lwesevisi, inani lokhiye, ilebula (igama lokukhumbuza), nomkhawulo wansuku zonke, bese ugcina

### Indlela 3: I-REST API (Yokuzenzakalela/Yemibhalo)

I-REST API indlela lapho izinhlelo zishintshana khona idatha nge-HTTP. Iwusizo ekubhaliseni ngokuzenzakalela ngemibhalo.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer ithokheni-yomphathi" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Ukhiye Omkhulu",
    "daily_limit": 1000
  }'
```

### Indlela 4: Amafulegi e-Proxy (Yokuhlola Okwesikhashana)

Isetshenziselwa uma ufuna ukuhlola okwesikhashana ngaphandle kokubhalisa ngokusemthethweni. Inyamalala uma ucima uhlelo.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Ukusebenzisa i-Proxy

### Ukusebenziswa Ne-OpenClaw (Inhloso Enkulu)

Indlela yokulungisa i-OpenClaw ukuze ixhume namasevisi e-AI nge-wall-vault.

Vula ifayili ye-`~/.openclaw/openclaw.json` bese wengeza okuqukethwe okulandelayo:

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
          { id: "wall-vault/hunter-alpha" },    // i-context ye-1M yamahhala
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Indlela Elula Kakhulu**: Cindezela inkinobho ethi **🦞 Kopisha Izilungiselelo ze-OpenClaw** ekhadini ye-agent kudashubhodi bese isiniphethi esineithokheni nekheli esigcwalisiwe kakade sikopishelwa kubhodi lokunamathisela. Namathisela nje.

**`wall-vault/` ngaphambi kwegama lemodeli iqondiswa kuphi?**

i-wall-vault inquma ngokuzenzakalela ukuthi iliphi isevisi ye-AI ezolisebenzisa ngokusekelwe egameni lemodeli:

| Ifomethi Yemodeli | Isevisi Exhunywayo |
|----------|--------------|
| `wall-vault/gemini-*` | Ukuxhuma ngokuqondile ne-Google Gemini |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | Ukuxhuma ngokuqondile ne-OpenAI |
| `wall-vault/claude-*` | Ukuxhuma ne-Anthropic nge-OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (amathokheni e-context angusigidi elingu-1 amahhala) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | Ukuxhuma ne-OpenRouter |
| `google/igama-lemodeli`, `openai/igama-lemodeli`, `anthropic/igama-lemodeli` njll. | Ukuxhuma ngokuqondile nesevisi efanele |
| `custom/google/igama-lemodeli`, `custom/openai/igama-lemodeli` njll. | Ingxenye ka-`custom/` isuswa bese kuqondiswa kabusha |
| `igama-lemodeli:cloud` | Ingxenye ye-`:cloud` isuswa bese kuxhunyaniswa ne-OpenRouter |

> 💡 **Yini i-context?** Isilinganiso sengxoxo i-AI ekwazi ukuyikhumbula ngesikhathi esisodwa. 1M (amathokheni ayisigidi elingu-1) kusho ukuthi ingxoxo ende kakhulu noma amadokhumenti amade angasingathwa ngesikhathi esisodwa.

### Ukuxhuma Ngokuqondile Ngefomethi ye-Gemini API (ukuhambelana namathuluzi akhona)

Uma unamathuluzi abekade esebenzisa i-Google Gemini API ngokuqondile, shintsha ikheli kuphela ku-wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Noma uma ithuluzi lakho licacisa i-URL ngokuqondile:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Ukusebenzisa ne-OpenAI SDK (Python)

Ungaxhuma futhi i-wall-vault kukhodi ye-Python esebenzisa i-AI. Shintsha `base_url` kuphela:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # i-wall-vault iphatha okhiye be-API ngokuzenzakalela
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # ifomethi yomhlinzeki/imodeli
    messages=[{"role": "user", "content": "Sawubona"}]
)
```

### Ukushintsha Imodeli Ngesikhathi Kusebenzwa

Ukushintsha imodeli ye-AI ngesikhathi i-wall-vault isasebenza:

```bash
# Shintsha imodeli ngokucela ngokuqondile ku-proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# Kumodi eyahlukaniswayo (ama-bot amaningi), shintsha kuseva ye-vault → iboniswa ngokushesha nge-SSE
curl -X PUT http://localhost:56243/admin/clients/i-id-ye-bot \
  -H "Authorization: Bearer ithokheni-yomphathi" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Ukuhlola Uhlu Lwamamodeli Atholakalayo

```bash
# Buka uhlu oluphelele
curl http://localhost:56244/api/models | python3 -m json.tool

# Buka amamodeli e-Google kuphela
curl "http://localhost:56244/api/models?service=google"

# Sesha ngegama (isib.: amamodeli afaka "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Isifinyezo Samamodeli Asemqoka Ngesevisi:**

| Isevisi | Amamodeli Asemqoka |
|--------|----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ (Hunter Alpha 1M i-context yamahhala, DeepSeek R1/V3, Qwen 2.5 njll.) |
| Ollama | Ithola ngokuzenzakalela iseva yendawo efakwe kukhompuyutha yakho |
| LM Studio | Iseva yendawo kukhompuyutha yakho (ichweba 1234) |
| vLLM | Iseva yendawo kukhompuyutha yakho (ichweba 8000) |

---

## Idashubhodi ye-Vault

Vakashela `http://localhost:56243` kubhrawuza ukuze ubone idashubhodi.

**Isakhiwo Sesikrini:**
- **Ibha ephezulu ehlala njalo (topbar)**: Ilogo, isikhethi solimi nendikimba, isimo sokuxhuma se-SSE
- **Igridi yamakhadi**: Amakhadi e-agent, isevisi, nokhiye we-API abekwe njengamathayili

### Ikhadi Lokhiye we-API

Ikhadi elikuvumela ukuphatha okhiye be-API ababhalisiwe ngokubuka okukodwa.

- Libonisa uhlu lokhiye oluhlukaniswe ngesevisi.
- `today_usage`: Inani lamathokheni aphumelele namuhla (inani lamalethi i-AI eliwafundile nalibhale)
- `today_attempts`: Isamba sezingcingo namuhla (ukuphumelela + ukwehluleka)
- Inkinobho ka-`[+ Engeza]` ukubhalisa ukhiye omusha, no-`✕` ukususa ukhiye.

> 💡 **Yini ithokheni?** Iyunithi i-AI eyisebenzisayo ukusingatha umbhalo. Cishe ilingana negama lesiNgisi elilodwa, noma izinhlamvu ezingu-1-2 zesiKorea. Izindleko ze-API ngokuvamile zibalwa ngokusekelwe kuleli nani lamathokheni.

### Ikhadi Le-Agent

Ikhadi elibonisa isimo sama-bot (ama-agent) axhunywene ne-proxy ye-wall-vault.

**Isimo sokuxhuma siboniswa ngezigaba ezingu-4:**

| Ukuboniswa | Isimo | Incazelo |
|------|------|------|
| 🟢 | Iyasebenza | I-proxy isebenza kahle |
| 🟡 | Ibambezelekile | Iyaphendula kodwa kancane |
| 🔴 | Ayikho ku-inthanethi | I-proxy ayiphenduli |
| ⚫ | Ayixhunyiwe/Ivaliwe | I-proxy ayikaze ixhume ne-vault noma ivaliwe |

**Isiqondiso Sezinkinobho Ngezansi Kwekhadi Le-Agent:**

Uma ubhalisa i-agent bese ucacisa **uhlobo lwe-agent**, izinkinobho zokushesha ezihloselwe lolo hlobo zibonakala ngokuzenzakalela.

---

#### 🔘 Inkinobho Yokukopisha Izilungiselelo — Yenza izilungiselelo zokuxhuma ngokuzenzakalela

Uma ucindezela inkinobho, isiniphethi sezilungiselelo esineithokheni yalelo agent, ikheli le-proxy, nolwazi lwemodeli esigcwalisiwe kakade sikopishelwa kubhodi lokunamathisela. Namathisela nje okuqukethwe okukopishiwe endaweni eboniswe ethebuleni elingezansi ukuze uqedele ukusetha ukuxhuma.

| Inkinobho | Uhlobo Lwe-Agent | Indawo Yokunamathisela |
|------|-------------|-------------|
| 🦞 Kopisha Izilungiselelo ze-OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Kopisha Izilungiselelo ze-NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Kopisha Izilungiselelo ze-Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Kopisha Izilungiselelo ze-Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Kopisha Izilungiselelo ze-VSCode | `vscode` | `~/.continue/config.json` |

**Isibonelo — Uma kunguhlobo lwe-Claude Code, okulandelayo kuzokopishwa:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "ithokheni-yaleli-agent"
}
```

**Isibonelo — Uma kunguhlobo lwe-VSCode (Continue):**

```yaml
# ~/.continue/config.yaml  ← namathisela ku-config.yaml, hhayi config.json
name: My Config
version: 0.0.1
schema: v1

models:
  - name: wall-vault proxy
    provider: openai
    model: gemini-2.5-flash
    apiBase: http://192.168.1.20:56244/v1
    apiKey: ithokheni-yaleli-agent
    roles:
      - chat
      - edit
      - apply
```

> ⚠️ **Inguqulo entsha ye-Continue isebenzisa `config.yaml`.** Uma `config.yaml` ikhona, `config.json` izoziwa ngokuphelele. Qiniseka ukuthi unamathisela ku-`config.yaml`.

**Isibonelo — Uma kunguhlobo lwe-Cursor:**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : ithokheni-yaleli-agent

// Noma izinguquko zemvelo:
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=ithokheni-yaleli-agent
```

> ⚠️ **Uma ukukopisha kubhodi lokunamathisela kungasebenzi**: Inqubomgomo yokuphepha yebhrawuza ingavimba ukukopisha. Uma ibhokisi lombhalo livuleka njengesivivinyeli, sebenzisa Ctrl+A ukukhetha konke bese Ctrl+C ukukopisha.

---

#### ⚡ Inkinobho Yokusebenzisa Ngokuzenzakalela — Cindezela kanye bese ukusetha kuqediwe

Uma uhlobo lwe-agent lungu-`cline`, `claude-code`, `openclaw`, noma `nanoclaw`, inkinobho ethi **⚡ Sebenzisa Izilungiselelo** ibonakala ekhadini le-agent. Uma ucindezela le nkinobho, ifayili yezilungiselelo zendawo yalelo agent ibuyekezwa ngokuzenzakalela.

| Inkinobho | Uhlobo Lwe-Agent | Ifayili Eliqondiswe |
|------|-------------|-------------|
| ⚡ Sebenzisa Izilungiselelo ze-Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Sebenzisa Izilungiselelo ze-Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Sebenzisa Izilungiselelo ze-OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Sebenzisa Izilungiselelo ze-NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Le nkinobho ithumela isicelo ku-**localhost:56244** (i-proxy yendawo). I-proxy kufanele ibe isebenza kule mishini ukuze isebenze.

---

#### 🔀 Ukuhlelela Amakhadi Ngokudonsela Nokuwehlisa (v0.1.17, kuthuthukisiwe v0.1.25)

Ungakwazi **ukudonsela** amakhadi e-agent kudashubhodi ukuwahlelela kabusha ngokulandelana okufunayo.

1. Bamba indawo ye-**signesha zezimoto (●)** phezulu kwesokunxele sekhadi ngemawusi bese udonsela
2. Wehlisela ekhadini ofuna ukuya kulo bese ukulandelana kushintsha

> 💡 Okuqukethwe kwekhadi (izindawo zokufaka, izinkinobho njll.) akudonseleki. Ungabamba kuphela endaweni yezignesha zezimoto.

#### 🟠 Ukuthola Inqubo Ye-Agent (v0.1.25)

Uma i-proxy isebenza kahle kodwa inqubo ye-agent yendawo (NanoClaw, OpenClaw) ife, isignesha sekhadi ishintsha ibe **nsomi (iyacwazimula)** bese umyalezo othi "Inqubo ye-agent imile" uboniswa.

- 🟢 Luhlaza: I-Proxy + i-agent kusebenza kahle
- 🟠 Nsomi (iyacwazimula): I-proxy isebenza kahle, i-agent ifile
- 🔴 Bomvu: I-proxy ayikho ku-inthanethi
3. Ukulandelana okushintshiwe **kugcinwa kuseva ngokushesha** futhi kuhlala ngisho ngemuva kokuvuselela

> 💡 Kumadivayisi okuthinta (iselula/ithebhulethi) akukasekelwa. Sebenzisa ibhrawuza yedeskthopu.

---

#### 🔄 Ukuvumelanisa Imodeli Nhlangothi Zombili (v0.1.16)

Uma ushintsha imodeli ye-agent kudashubhodi ye-vault, izilungiselelo zendawo zalelo agent zibuyekezwa ngokuzenzakalela.

**Nge-Cline:**
- Uma imodeli ishintshwa ku-vault → Isehlakalo se-SSE → I-proxy ibuyekeza ingxenye yemodeli ku-`globalState.json`
- Izimpokophelo zokubuyekeza: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` nokhiye we-API akuthintwa
- **Ukuvuselela i-VS Code (`Ctrl+Alt+R` noma `Ctrl+Shift+P` → `Developer: Reload Window`) kuyadingeka**
  - Ngoba i-Cline ayifundi kabusha ifayili yezilungiselelo ngesikhathi isebenza

**Nge-Claude Code:**
- Uma imodeli ishintshwa ku-vault → Isehlakalo se-SSE → I-proxy ibuyekeza ingxenye ye-`model` ku-`settings.json`
- Isesha ngokuzenzakalela izindlela ze-WSL ne-Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Inhlangothi Ehlehliselayo (i-agent → i-vault):**
- Uma i-agent (Cline, Claude Code njll.) ithumela isicelo ku-proxy, i-proxy ifaka ulwazi lwesevisi/imodeli yekhasimende ku-heartbeat
- Isevisi/imodeli esetshenziswa njengamanje iboniswa ngesikhathi sangempela ekhadini le-agent kudashubhodi ye-vault

> 💡 **Okubalulekile**: I-proxy ibona i-agent ngethokheni ye-Authorization yesicelo, bese iqondisa ngokuzenzakalela kusevisi/imodeli esethwe ku-vault. Ngisho noma i-Cline noma i-Claude Code ithumela igama lemodeli elihlukile, i-proxy ibeka phezu kwayo ngezilungiselelo ze-vault.

---

### Ukusebenzisa i-Cline ku-VS Code — Umhlahlandlela Ogcwele

#### Isinyathelo 1: Faka i-Cline

Faka **i-Cline** (ID: `saoudrizwan.claude-dev`) kusuka emakethe yezandiso ye-VS Code.

#### Isinyathelo 2: Bhalisa I-Agent ku-Vault

1. Vula idashubhodi ye-vault (`http://IP-ye-vault:56243`)
2. Esigabeni sika-**Ama-Agent**, cindezela **+ Engeza**
3. Faka okulandelayo:

| Inkambu | Inani | Incazelo |
|------|----|------|
| ID | `i_cline_yami` | Isibonisi esiyingqayizivele (isiNgisi, ngaphandle kwezikhala) |
| Igama | `I-Cline Yami` | Igama elizoboniswa kudashubhodi |
| Uhlobo Lwe-Agent | `cline` | ← Kufanele ukhethe `cline` |
| Isevisi | Khetha isevisi ozoyisebenzisa (isib.: `google`) | |
| Imodeli | Faka imodeli ozoyisebenzisa (isib.: `gemini-2.5-flash`) | |

4. Cindezela **Gcina** bese ithokheni yenziwa ngokuzenzakalela

#### Isinyathelo 3: Xhuma Ku-Cline

**Indlela A — Ukusebenzisa Ngokuzenzakalela (Ituswa)**

1. Qiniseka ukuthi i-**proxy** ye-wall-vault iyasebenza kule mishini (`localhost:56244`)
2. Cindezela inkinobho ethi **⚡ Sebenzisa Izilungiselelo ze-Cline** ekhadini le-agent kudashubhodi
3. Uma isaziso esithi "Izilungiselelo zisetshenzisiwe!" sibonakala, kuphumelele
4. Vuselela i-VS Code (`Ctrl+Alt+R`)

**Indlela B — Ukusetha Ngesandla**

Vula izilungiselelo (⚙️) kubha yaseceleni ye-Cline bese usetha:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://ikheli-le-proxy:56244/v1`
  - Kule mishini efanayo: `http://localhost:56244/v1`
  - Kumishini enye njenge-Mini server: `http://192.168.1.20:56244/v1`
- **API Key**: Ithokheni ekhishwe ku-vault (kopisha ekhadini le-agent)
- **Model ID**: Imodeli esethwe ku-vault (isib.: `gemini-2.5-flash`)

#### Isinyathelo 4: Qinisekisa

Thumela noma imuphi umyalezo engxoxweni ye-Cline. Uma konke kulungile:
- **Uphawu oluluhlaza (● Iyasebenza)** lubonakala ekhadini le-agent kudashubhodi ye-vault
- Isevisi/imodeli yamanje iboniswa ekhadini (isib.: `google / gemini-2.5-flash`)

#### Ukushintsha Imodeli

Uma ufuna ukushintsha imodeli ye-Cline, yishintshe ku-**dashubhodi ye-vault**:

1. Shintsha ukukhetha kwesevisi/imodeli ekhadini le-agent
2. Cindezela **Sebenzisa**
3. Vuselela i-VS Code (`Ctrl+Alt+R`) — igama lemodeli kuphansi kwe-Cline lizobuyekezwa
4. Imodeli entsha izosetshenziswa kusukela esiclelweni esilandelayo

> 💡 Eqinisweni, i-proxy ibona isicelo se-Cline ngethokheni bese iqondisa kumodeli yezilungiselelo ze-vault. Ngisho ngaphandle kokuvuselela i-VS Code, **imodeli okusetshenziswa yona ngempela ishintsha ngokushesha** — ukuvuselela kungokokubuyekeza ukuboniswa kwemodeli ku-UI ye-Cline kuphela.

#### Ukuthola Ukuphuka Kokuxhuma

Uma uvala i-VS Code, ikhadi le-agent kudashubhodi ye-vault lishintsha libe phuzi (ibambezelekile) ngemuva kwamasekhondi angama-**90**, bese liba bomvu (ayikho ku-inthanethi) ngemuva kwe-**mizuzu engu-3**. (Kusukela ku-v0.1.18, ukuhlola isimo ngamasekhondi angu-15 kwenza ukuthola ukungabikhona ku-inthanethi kusheshe kakhulu.)

#### Ukuxazulula Izinkinga

| Isimpthomu | Imbangela | Isixazululo |
|------|------|------|
| Iphutha lokuthi "ukuxhuma kwehlulekile" ku-Cline | I-proxy ayisebenzi noma ikheli asilona | Qinisekisa i-proxy nge-`curl http://localhost:56244/health` |
| Uphawu oluluhlaza alubonakali ku-vault | Ukhiye we-API (ithokheni) awusethiwe | Cindezela inkinobho ethi **⚡ Sebenzisa Izilungiselelo ze-Cline** futhi |
| Imodeli ngezansi kwe-Cline ayishintshi | I-Cline igcina izilungiselelo kwi-cache | Vuselela i-VS Code (`Ctrl+Alt+R`) |
| Igama lemodeli elingalungile liboniswa | Iphutha elidala (lilungiswe ku-v0.1.16) | Buyekeza i-proxy kuya ku-v0.1.16 nangaphezulu |

---

#### 🟣 Inkinobho Yokukopisha Umyalo Wokuthumela — Isetshenziselwa uma ufakela kumshini omusha

Isetshenziselwa uma ufaka i-proxy ye-wall-vault okokuqala kumshini omusha bese uyixhuma ne-vault. Uma ucindezela inkinobho, isikripthi sonke sokufakela sikopishwa. Sinamathisele kutheminali yekhompuyutha entsha bese uyisisebenzisa, bese okulandelayo kushingathwa ngesikhathi esisodwa:

1. Faka i-binary ye-wall-vault (iyekwa uma isifakiwe kakade)
2. Bhalisa isevisi yomsebenzi ye-systemd ngokuzenzakalela
3. Qala isevisi bese uxhuma ne-vault ngokuzenzakalela

> 💡 Isikripthi sesivele sineithokheni yaleli agent nekheli leseva ye-vault eligcwalisiwe, ngakho ungasisebenzisa ngokuqondile ngemuva kokunamathisela ngaphandle kwezinguquko.

---

### Ikhadi Lesevisi

Ikhadi lokuvula/lokuvala noma lokulungisa amasevisi e-AI okuwasebenzisa.

- Izishintshi zokuvula/ukuvala isevisi ngayinye
- Uma ufaka ikheli leseva ye-AI yendawo (Ollama, LM Studio, vLLM njll. esebenza kukhompuyutha yakho), amamodeli atholakalayo azotholakala ngokuzenzakalela.
- **Ukubonisa isimo sokuxhuma kwesevisi yendawo**: Uphawu lwe-● eceleni kwegama lesevisi lu-**luhlaza** uma ixhunyiwe, **mpunga** uma ingaxhunyiwe
- **Izignesha zezimoto zokuzenzakalela zesevisi yendawo** (v0.1.23+): Amasevisi endawo (Ollama, LM Studio, vLLM) avulwa ngokuzenzakalela uma engaxhunyaniswa, bese avalwa ngokuzenzakalela uma enqamuka. Uma uvula isevisi, ishintsha ibe ● luhlaza phakathi kwamasekhondi angu-15 nebhokisi lokumaka livulwa, futhi uma uyicisha, ivalwa ngokuzenzakalela. Lokhu kusebenza ngendlela efanayo njengalokho amasevisi efu (Google, OpenRouter njll.) ashintsha ngokuzenzakalela ngokusekelwe ekubakhona kokhiye we-API.

> 💡 **Uma isevisi yendawo isebenza kumshini omunye**: Faka i-IP yaleyo khompuyutha endaweni yokufaka i-URL yesevisi. Isib.: `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio). Uma isevisi iboshwe ku-`127.0.0.1` kuphela hhayi ku-`0.0.0.0`, ngeke ikwazi ukufinyelelwa nge-IP yangaphandle, ngakho hlola ikheli lokuboshwa ezilungiselelweni zesevisi.

### Ukufaka Ithokheni Yomphathi

Uma uzama ukusebenzisa izici ezibalulekile njengokungeza noma ukususa okhiye kudashubhodi, isivivinyeli sokufaka ithokheni yomphathi sizobonakala. Faka ithokheni oyisethe ku-setup wizard. Uma usuyifakile, ihlala kuze kuvale ibhrawuza.

> ⚠️ **Uma ukuqinisekiswa kwehluleka ngaphezu kwesikhathi esinqu-10 phakathi kwemizuzu engu-15, leyo IP izovinjwa okwesikhashana.** Uma ukhohlwe ithokheni yakho, hlola into ethi `admin_token` kufayili ye-`wall-vault.yaml`.

---

## Imodi Eyahlukaniswayo (Ama-Bot Amaningi)

Ukusetha **kokwabelana isikhwama sokhiye esisodwa** uma usebenzisa i-OpenClaw kumakhompuyutha amaningi ngesikhathi esisodwa. Kulula ngoba udinga ukuphatha okhiye endaweni eyodwa kuphela.

### Isibonelo Sokusetha

```
[Iseva Yesikhwama Sokhiye]
  wall-vault vault    (isikhwama sokhiye :56243, idashubhodi)

[WSL Alpha]            [Raspberry Pi Gamma]    [Mac Mini Wendawo]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ Ukuvumelanisa SSE   ↕ Ukuvumelanisa SSE     ↕ Ukuvumelanisa SSE
```

Onke ama-bot abheka iseva ye-vault maphakathi, ngakho uma ushintsha imodeli noma wengeza okhiye ku-vault, kuboniswa kuwo onke ama-bot ngokushesha.

### Isinyathelo 1: Qala Iseva Yesikhwama Sokhiye

Sebenzisa kumshini ozowusebenzisa njengeseva ye-vault:

```bash
wall-vault vault
```

### Isinyathelo 2: Bhalisa I-Bot Ngayinye (Ikhasimende)

Bhalisa ulwazi lwe-bot ngayinye exhuma kuseva ye-vault kusengaphambili:

```bash
curl -X POST http://localhost:56243/admin/clients \
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

### Isinyathelo 3: Qala I-Proxy Kumshini We-Bot Ngamunye

Sebenzisa i-proxy kumshini ngamunye lapho i-bot ifakwe khona, ucacise ikheli leseva ye-vault nethokheni:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Shintsha ingxenye ye-**`192.168.x.x`** nge-IP yangaphakathi yangempela yekhompuyutha yeseva ye-vault. Ungayiqinisekisa ngezilungiselelo ze-router noma umyalo we-`ip addr`.

---

## Ukusethwa Kokuqala Ngokuzenzakalela

Uma kukhathaza ukuvula i-wall-vault ngesandla njalo uma uqala kabusha ikhompuyutha, yibhalise njengesevisi yesistimu. Uma isibhaliswe, izoqala ngokuzenzakalela uma kulayishwa.

### Linux — systemd (i-Linux eningi)

i-systemd uhlelo oluvula nolulawula izinhlelo ngokuzenzakalela ku-Linux:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Hlola amalogo:

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

Uhlelo olulawula ukuqala ngokuzenzakalela kwezinhlelo ku-macOS:

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

## I-Doctor: Isihloli

Umyalo we-`doctor` yi-**thuluzi elizihlola nelizilungisa** ukuqinisekisa ukuthi i-wall-vault isethwe ngendlela efanele.

```bash
wall-vault doctor check   # Hlola isimo samanje (ifunda kuphela, ayishintshi lutho)
wall-vault doctor fix     # Lungisa izinkinga ngokuzenzakalela
wall-vault doctor all     # Ukuhlola + ukulungisa ngokuzenzakalela ngesikhathi esisodwa
```

> 💡 Uma kukhona okubonakala kungalungile, sebenzisa `wall-vault doctor all` kuqala. Isingatha izinkinga eziningi ngokuzenzakalela.

---

## I-RTK Ukonga Amathokheni

*(v0.1.24+)*

**I-RTK (Ithuluzi Lokonga Amathokheni)** icindezela ngokuzenzakalela imiphumela yemiyalo ye-shell esenziwa ama-agent e-AI ekhodi (njenge-Claude Code) ukunciphisa ukusetshenziswa kwamathokheni. Isibonelo, imiphumela yemigqa engu-15 ye-`git status` incishiswa ibe isifinyezo semigqa engu-2.

### Ukusetshenziswa Okuyisisekelo

```bash
# Songa umyalo nge-wall-vault rtk bese imiphumela ihluziwa ngokuzenzakalela
wall-vault rtk git status          # Uhlu lwamafayili ashintshiwe kuphela
wall-vault rtk git diff HEAD~1     # Imigqa eshintshiwe + i-context encane
wall-vault rtk git log -10         # I-hash + umyalezo womugqa owodwa
wall-vault rtk go test ./...       # Ukuhlolwa okwehlulekile kuphela
wall-vault rtk ls -la              # Imiyalo engasekelwe isikwa ngokuzenzakalela
```

### Imiyalo Esekelwayo Nomthelela Wokunciphisa

| Umyalo | Indlela Yokuhluza | Izinga Lokunciphisa |
|------|----------|--------|
| `git status` | Isifinyezo samafayili ashintshiwe kuphela | ~87% |
| `git diff` | Imigqa eshintshiwe + i-context yemigqa engu-3 | ~60-94% |
| `git log` | I-hash + umyalezo womugqa wokuqala | ~90% |
| `git push/pull/fetch` | Susa inqubekelaphambili, isifinyezo kuphela | ~80% |
| `go test` | Bonisa ukwehluleka kuphela, bala ukuphumelela | ~88-99% |
| `go build/vet` | Bonisa amaphutha kuphela | ~90% |
| Yonke eminye imiyalo | Imigqa engu-50 yokuqala + 50 yokugcina, ubungako obukhulu 32KB | Kuyehluka |

### Uhlelo Lokuhluza Lwezinyathelo Ezingu-3

1. **Isihlungi sesakhiwo somyalo ngamunye** — Siqonda ifomethi yemiphumela ye-git, go njll. bese sikhipha izingxenye ezinomqondo kuphela
2. **Ukusingatha okulandela i-regex** — Susa amakhodi ombala we-ANSI, nciphisa imigqa engenalutho, hlanganisa imigqa ephindaphindiwe
3. **Passthrough + ukusika** — Imiyalo engasekelwe igcina imigqa engu-50 yokuqala/yokugcina kuphela

### Ukuxhuma Ne-Claude Code

Ungalungisa yonke imiyalo ye-shell ukuze idlule nge-RTK ngokuzenzakalela nge-hook ye-`PreToolUse` ye-Claude Code.

```bash
# Faka i-hook (yengezwa ngokuzenzakalela ku-Claude Code settings.json)
wall-vault rtk hook install
```

Noma yengeza ngesandla ku-`~/.claude/settings.json`:

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

> 💡 **Ukugcinwa kwe-Exit code**: I-RTK ibuyisa ikhodi yokuphuma yomyalo wokuqala njengoba injalo. Uma umyalo wehluleka (exit code ≠ 0), i-AI nayo ithola ukwehluleka ngokufanele.

> 💡 **Ukuphoqa IsiNgisi**: I-RTK isebenzisa imiyalo nge-`LC_ALL=C` ukuze ikhiqize imiphumela yesiNgisi njalo ngaphandle kokucabangela izilungiselelo zolimi lwesistimu. Lokhu kuqinisekisa ukuthi isihlungi sisebenza ngokufanele.

---

## Izinguquko Zemvelo

Izinguquko zemvelo yindlela yokudlulisela amanani ezilungiselelo kuhlelo. Faka ngefomethi ye-`export igama-lenguquko=inani` kutheminali, noma faka kufayili yesevisi yokuqala ngokuzenzakalela ukuze isebenze njalo.

| Inguquko | Incazelo | Inani Lesibonelo |
|------|------|---------|
| `WV_LANG` | Ulimi lwedashubhodi | `ko`, `en`, `ja` |
| `WV_THEME` | Indikimba yedashubhodi | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Ukhiye we-Google API (abaningi ngokhefana) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Ukhiye we-OpenRouter API | `sk-or-v1-...` |
| `WV_VAULT_URL` | Ikheli leseva ye-vault kumodi eyahlukaniswayo | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Ithokheni yokuqinisekiswa yekhasimende (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Ithokheni yomphathi | `admin-token-here` |
| `WV_MASTER_PASS` | Iphasiwedi yokubhala ngekhodi ukhiye we-API | `my-password` |
| `WV_AVATAR` | Indlela yefayili yesithombe se-avatar (indlela ehlobene kusuka ku-`~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Ikheli leseva yendawo ye-Ollama | `http://192.168.x.x:11434` |

---

## Ukuxazulula Izinkinga

### Uma I-Proxy Ingaqali

Ngokuvamile, ichweba selivele lisetshenziswa uhlelo olunye.

```bash
ss -tlnp | grep 56244   # Hlola ukuthi ubani osebenzisa ichweba 56244
wall-vault proxy --port 8080   # Qala ngenombolo yechweba ehlukile
```

### Uma Iphutha Lokhiye we-API Livela (429, 402, 401, 403, 582)

| Ikhodi Yephutha | Incazelo | Indlela Yokusingatha |
|----------|------|----------|
| **429** | Izicelo eziningi kakhulu (ukusetshenziswa kufinyelele umkhawulo) | Linda kancane noma engeza omunye ukhiye |
| **402** | Ukukhokha kuyadingeka noma i-credit ayenele | Gcwalisa i-credit kusevisi efanele |
| **401 / 403** | Ukhiye awulungile noma awunayo imvume | Qinisekisa inani lokhiye bese ubhalisa kabusha |
| **582** | Umthwalo we-gateway (cooldown imizuzu engu-5) | Isuswa ngokuzenzakalela ngemuva kwemizuzu engu-5 |

```bash
# Hlola uhlu nesimo sokhiye ababhalisiwe
curl -H "Authorization: Bearer ithokheni-yomphathi" http://localhost:56243/admin/keys

# Setha kabusha izibali zokusetshenziswa kokhiye
curl -X POST -H "Authorization: Bearer ithokheni-yomphathi" http://localhost:56243/admin/keys/reset
```

### Uma I-Agent Iboniswa Njengokuthi "Ayixhunyiwe"

"Ayixhunyiwe" kusho ukuthi inqubo ye-proxy ayithumeli isignali (heartbeat) ku-vault. **Akusho ukuthi izilungiselelo azigciniwe.** I-proxy kufanele ibe isebenza yazi ikheli leseva ye-vault nethokheni ukuze isimo sokuxhuma sishintshe.

```bash
# Qala i-proxy ucacise ikheli leseva ye-vault, ithokheni, ne-ID yekhasimende
WV_VAULT_URL=http://ikheli-leseva-ye-vault:56243 \
WV_VAULT_TOKEN=ithokheni-yekhasimende \
WV_VAULT_CLIENT_ID=i-ID-yekhasimende \
wall-vault proxy
```

Uma ukuxhuma kuphumelela, kushintsha kube 🟢 Iyasebenza kudashubhodi phakathi kwamasekhondi angama-20.

### Uma I-Ollama Ingakwazi Ukuxhuma

I-Ollama uhlelo olusebenzisa i-AI ngokuqondile kukhompuyutha yakho. Kuqala qinisekisa ukuthi i-Ollama ivuliwe.

```bash
curl http://localhost:11434/api/tags   # Uma uhlu lwamamodeli lubonakala, kujwayelekile
export OLLAMA_URL=http://192.168.x.x:11434   # Uma isebenza kumshini omunye
```

> ⚠️ Uma i-Ollama ingaphenduli, qala i-Ollama kuqala ngomyalo we-`ollama serve`.

> ⚠️ **Amamodeli amakhulu alengezela**: Amamodeli amakhulu njenge-`qwen3.5:35b`, `deepseek-r1` angathatha imizuzu eminingi ukukhiqiza impendulo. Ngisho ibonakala sengathi akukho mpendulo, ingase isethwe ngokujwayelekile, ngakho linda.

---

## Izinguquko Zakamuva (v0.1.16 ~ v0.1.25)

### v0.1.25 (2026-04-08)
- **Ukuthola Inqubo Ye-Agent**: I-proxy ithola isimo sokuphila se-agent yendawo (NanoClaw/OpenClaw) bese ibonisa ngesignesha ensomi kudashubhodi.
- **Ukuthuthukiswa Kwesibambi Sokudonsela**: Kushintshiwe ukuze amakhadi abanjwe kuphela endaweni yezignesha zezimoto (●) ngesikhathi sokuhlelela. Izindawo zokufaka noma izinkinobho azidonseleki ngephutha.

### v0.1.24 (2026-04-06)
- **Umyalo Omncane we-RTK Wokonga Amathokheni**: `wall-vault rtk <command>` ihluza ngokuzenzakalela imiphumela yemiyalo ye-shell ukunciphisa ukusetshenziswa kwamathokheni ama-agent e-AI ngo-60-90%. Ifaka izihlungi ezikhethekile zemiyalo emqoka njenge-git, go, nemiyalo engasekelwe nayo isikwa ngokuzenzakalela. Ixhuma ngobuso obungabonakali nge-hook ye-`PreToolUse` ye-Claude Code.

### v0.1.23 (2026-04-06)
- **Ukulungiswa Kokushintsha Imodeli Ye-Ollama**: Inkinga lapho ukushintsha imodeli ye-Ollama kudashubhodi ye-vault kungabonakali ku-proxy ilungisiwe. Ngaphambili yayisebenzisa inguquko yemvelo (`OLLAMA_MODEL`) kuphela, kodwa manje izilungiselelo ze-vault zinikwa ukubaluleka kuqala.
- **Izignesha Zezimoto Zokuzenzakalela Zesevisi Yendawo**: Ollama, LM Studio, vLLM zivulwa ngokuzenzakalela uma zingaxhunyaniswa, bese zivalwa ngokuzenzakalela uma zinqamuka. Isebenza ngendlela efanayo nokushintsha ngokuzenzakalela okusekelwe kukhiye kwamasevisi efu.

### v0.1.22 (2026-04-05)
- **Ukulungiswa Kokungabi Kwenkambu ye-content Engenalutho**: Uma amamodeli okucabanga (gemini-3.1-pro, o1, claude thinking njll.) esebenzisa umkhawulo wonke we-max_tokens ekucabangeni bese ehluleka ukukhiqiza impendulo yangempela, i-proxy yayishiya izinkambu ze-`content`/`text` ku-JSON yempendulo nge-`omitempty`, okubangela amakhasimende e-OpenAI/Anthropic SDK athole iphutha elithi `Cannot read properties of undefined (reading 'trim')`. Kushintshiwe ukuze kufakwe izinkambu njalo ngokulandela izimiso ze-API ezisemthethweni.

### v0.1.21 (2026-04-05)
- **Ukusekelwa Kwemodeli ye-Gemma 4**: Amamodeli omndeni we-Gemma anjenge-`gemma-4-31b-it`, `gemma-4-26b-a4b-it` angasetshenziswa nge-Google Gemini API.
- **Ukusekelwa Ngokusemthethweni Kwamasevisi e-LM Studio / vLLM**: Ngaphambili la masevisi ayekweqiwa ekuqondiseni kwe-proxy futhi njalo ayeshintshwa nge-Ollama. Manje aqondiswa ngendlela efanele nge-API ehambelana ne-OpenAI.
- **Ukulungiswa Kokuboniswa Kwesevisi Kudashubhodi**: Ngisho noma i-fallback yenzeka, idashubhodi ibonisa njalo isevisi esethwe umsebenzisi.
- **Ukuboniswa Kwesimo Sesevisi Yendawo**: Isimo sokuxhuma samasevisi endawo (Ollama, LM Studio, vLLM njll.) siboniswa ngombala wophawu lwe-● uma idashubhodi ilayisha.
- **Inguquko Yemvelo Yesihlungi Samathuluzi**: Imodi yokudlulisa amathuluzi ingasethwa ngokuguquka kwemvelo `WV_TOOL_FILTER=passthrough`.

### v0.1.20 (2026-03-28)
- **Ukuqiniswa Kokuphepha Okubandakanyayo**: Ukuthuthukiswa kwezinto ezingu-12 zokuphepha kufaka phakathi ukuvimbela i-XSS (izindawo ezingu-41), ukuqhathanisa amathokheni esikhathi esingaguquki, izihibe ze-CORS, imikhawulo yobukhulu besicelo, ukuvimbela ukuhamba ngezikhala, ukuqinisekiswa kwe-SSE, ukuqiniswa kwesikhawuli sejubane njll.

### v0.1.19 (2026-03-27)
- **Ukutholwa Kwe-Claude Code Ku-Inthanethi**: I-Claude Code engadluli ku-proxy nayo iboniswa njengoku-inthanethi kudashubhodi.

### v0.1.18 (2026-03-26)
- **Ukulungiswa Kokunamathela Kwesevisi ye-Fallback**: Ngemuva kokubuyela ku-Ollama ngenxa yephutha lesikhashana, uma isevisi yokuqala ibuya, ibuyela ngokuzenzakalela.
- **Ukuthuthukiswa Kokuthola Ukungabikhona Ku-Inthanethi**: Ukuhlola isimo ngamasekhondi angu-15 kwenza ukuthola ukumisa kwe-proxy kusheshe kakhulu.

### v0.1.17 (2026-03-25)
- **Ukuhlelela Amakhadi Ngokudonsela Nokuwehlisa**: Amakhadi e-agent angadonselwa ukushintsha ukulandelana.
- **Inkinobho Yokusebenzisa Izilungiselelo Ngaphakathi**: Inkinobho ka-[⚡ Sebenzisa Izilungiselelo] ibonakala kuma-agent angekho ku-inthanethi.
- **Uhlobo lwe-agent lwe-cokacdir lungezwe**.

### v0.1.16 (2026-03-25)
- **Ukuvumelanisa Imodeli Nhlangothi Zombili**: Ukushintsha imodeli ye-Cline noma ye-Claude Code kudashubhodi ye-vault kubangela ukuboniswa ngokuzenzakalela.

---

*Ukuthola ulwazi olunzulu lwe-API, bheka [API.md](API.md).*
