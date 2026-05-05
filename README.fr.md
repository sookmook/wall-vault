<p align="center">
  <img src="docs/logo.png" alt="wall-vault" width="200">
</p>

# wall-vault

> **Coffre-fort à clés API + proxy IA dans un unique binaire Go.**
> Stocke vos clés en local avec AES-GCM, les fait tourner entre fournisseurs, bascule en cas d'échec, et propose un tableau de bord en temps réel.

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8.svg)](go.mod)

[English](README.md) · [한국어](README.ko.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [Español](README.es.md) · **Français** · [Deutsch](README.de.md) · [Português](README.pt.md) · [العربية](README.ar.md) · [हिन्दी](README.hi.md) · [Bahasa Indonesia](README.id.md) · [ภาษาไทย](README.th.md) · [Kiswahili](README.sw.md) · [Hausa](README.ha.md) · [नेपाली](README.ne.md) · [Монгол](README.mn.md) · [isiZulu](README.zu.md)

---

## De quoi s'agit-il

wall-vault s'intercale entre un agent IA (OpenClaw, Claude Code, Cursor, Continue, votre propre script) et les fournisseurs IA cloud ou locaux avec lesquels il dialogue. Deux composants en un seul binaire :

- **Vault** — stocke les clés API chiffrées au repos (AES-GCM avec un mot de passe maître), les fait tourner, suit l'usage et les temps de récupération par clé, diffuse les changements via SSE, et expose un tableau de bord web sur `:56243`.
- **Proxy** — expose des points d'entrée Gemini, Anthropic et compatibles OpenAI sur `:56244`, choisit une clé dans le coffre, dispatche vers l'amont configuré, et bascule vers le fournisseur suivant en cas d'échec.

Il prend en charge quatre formats de requête (Gemini `:generateContent`, Anthropic `/v1/messages`, OpenAI `/v1/chat/completions`, et Ollama natif `/api/chat`) et cinq catégories d'amont :

| Fournisseur | Notes |
|----------|-------|
| Google Gemini | API native ; rotation des clés par projet |
| Anthropic | Passthrough natif `/v1/messages` |
| OpenAI | `/v1/chat/completions` natif |
| OpenRouter | 340+ modèles, fallback automatique vers les variantes `:free` |
| Ollama / LM Studio / vLLM / llama.cpp / Jan / KoboldCpp / TabbyAPI / mlx-server / LiteLLM | Backends locaux compatibles OpenAI ; intégration via plugin yaml |

Ajouter un nouveau backend compatible OpenAI ne demande qu'un fichier yaml dans `~/.wall-vault/services/` — sans modification de code.

## Pourquoi vous pourriez en avoir besoin

- Vous jonglez avec trois ou quatre services IA et voulez une seule URL à laquelle l'agent s'adresse.
- Vous voulez qu'une clé free-tier en cooldown s'efface au profit de la suivante sans casser la session.
- Vous voulez que les mêmes clés alimentent plusieurs bots / IDE / scripts sur le même réseau local sans copier les identifiants.
- Vous préférez un tableau de bord aux variables d'environnement pour éditer vos clés API.
- Vous voulez une option local-first (Ollama / LM Studio) quand les quotas cloud sont épuisés.

## Démarrage rapide

### Installation (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

Ou téléchargez directement un binaire pré-construit :

```bash
# Linux amd64
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# macOS Apple Silicon
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o wall-vault && chmod +x wall-vault

# Linux arm64 (Raspberry Pi, serveurs ARM)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-arm64 \
  -o wall-vault && chmod +x wall-vault
```

### Installation (Windows)

```powershell
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe
```

### Premier lancement

```bash
wall-vault setup    # assistant interactif — choisit port, services, jeton admin, mot de passe maître
wall-vault start    # lance à la fois vault et proxy
```

Ouvrez `http://localhost:56243` (ou `https://...` une fois TLS activé — voir ci-dessous) dans un navigateur. Le tableau de bord vous demande le jeton admin imprimé par `setup`. Vous pouvez ensuite ajouter des clés API, enregistrer des clients et changer de modèle sans redémarrer.

---

## TLS (recommandé)

Par défaut, `wall-vault setup` écrit une configuration sans TLS, donc les deux écoutes répondent en HTTP simple. Les URL d'exemple de ce README utilisent `https://localhost:56244` parce que la plupart des agents (OpenClaw, Claude Code, Cursor) souhaitent un point d'entrée unique derrière TLS qui ne se cassera pas si vous déplacez ensuite le proxy sur un autre hôte. Pour correspondre à ces exemples, activez TLS une bonne fois pour toutes avec l'autorité de certification interne fournie :

```bash
# 1. Créez l'autorité de certification interne wall-vault (une seule fois, dans ~/.wall-vault/ca.{crt,key})
wall-vault cert init

# 2. Émettez un certificat hôte pour CETTE machine
#    Les SAN incluent le nom d'hôte, localhost, 127.0.0.1, et toute IP LAN détectée
wall-vault cert issue $(hostname)

# 3. Faites confiance à la CA dans le trousseau du système d'exploitation local
wall-vault cert install-trust

# 4. Basculez les écoutes en TLS
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
wall-vault start
```

Pour une autre machine sur votre LAN : copiez `~/.wall-vault/ca.crt` dessus et exécutez-y `wall-vault cert install-trust --ca <chemin>`. Une fois la CA approuvée partout, chaque machine du réseau peut atteindre le proxy via `https://<hôte>:56244` sans avertissement de certificat.

Si vous préférez rester en HTTP simple, laissez la configuration telle quelle et remplacez `https://` par `http://` dans les extraits clients ci-dessous. Les deux schémas fonctionnent ; la différence est de savoir quel port répond à un handshake TLS.

**Repli loopback.** Les clients du même hôte qui ne peuvent pas honorer la CA wall-vault (notamment le runtime Node fourni avec OpenClaw, qui réécrit `NODE_EXTRA_CA_CERTS` au démarrage) atteignent le proxy via un compagnon HTTP simple uniquement loopback sur `127.0.0.1:56245`. wall-vault l'active automatiquement quand TLS est actif.

---

## Connexion des clients

Pointez n'importe quel client IA sur `https://<hôte>:56244` (ou `http://...` si TLS est désactivé). Le proxy répond à quatre formats :

| Format | Chemin | Exemples de clients |
|--------|------|-----------------|
| Gemini | `/google/v1beta/models/<model>:generateContent` | OpenClaw, Gemini CLI, Antigravity |
| Anthropic | `/v1/messages` | Claude Code, SDKs Anthropic |
| OpenAI | `/v1/chat/completions` | Cursor, Continue, scripts personnalisés, la plupart des applis LLM |
| Ollama natif | `/api/chat` | Clients Ollama en passthrough |

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<votre-jeton-client-vault>
claude
```

Quand les crédits Anthropic en amont sont épuisés, le dispatch bascule vers les fournisseurs définis dans `fallback_services` pour ce client. Pour activer explicitement un fallback non-Claude :

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

(La valeur par défaut vide fait que le dispatch retourne une erreur, pour que tout mauvais routage apparaisse immédiatement.)

### Cursor / Continue

Dans Cursor **Settings → AI → OpenAI API** :

```
Base URL:  https://localhost:56244
API Key:   <votre-jeton-client-vault>
Model:     gemini-2.5-flash    # ou tout modèle connu de wall-vault
```

Continue (`config.json`) :

```json
{
  "models": [
    {
      "title": "wall-vault",
      "provider": "openai",
      "model": "gemini-2.5-flash",
      "apiBase": "https://localhost:56244/v1",
      "apiKey": "<votre-jeton-client-vault>"
    }
  ]
}
```

### OpenClaw

OpenClaw est un framework d'agent TUI que wall-vault a été initialement conçu pour servir. La modale **Add Agent** du tableau de bord définit le type d'agent à `openclaw` (ou `nanoclaw`) ; wall-vault écrit alors directement `~/.openclaw/openclaw.json`, y compris les URL des fournisseurs, le jeton du coffre et les entrées de modèle :

```bash
VAULT_CLIENT_ID=my-bot \
VAULT_URL=http://<vault-host>:56243 \
VAULT_TOKEN=<votre-jeton-client> \
wall-vault proxy
```

### curl / scripts

```bash
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer <votre-jeton-client-vault>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "ping"}]
  }'
```

---

## Configuration

`wall-vault setup` écrit soit `./wall-vault.yaml` soit `~/.wall-vault/config.yaml`. Modifiez à la main les champs que l'assistant ne demande pas.

```yaml
mode: standalone     # standalone | distributed
lang: en             # en | ko | zh | ja | es | fr | de | pt | ar | hi | id | th | sw | ha | ne | mn | zu
theme: light         # light | dark | cherry | ocean | gold | autumn | winter

proxy:
  port: 56244
  host: ""           # par défaut : 127.0.0.1 standalone, 0.0.0.0 distributed
  client_id: my-bot
  vault_url: ""      # distributed : http://vault-host:56243
  vault_token: ""    # distributed : jeton client
  tool_filter: strip_all   # strip_all | whitelist | passthrough
  allowed_tools: []
  services: [google, openrouter, ollama]
  timeout: 300s
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  plain_port: 56245              # compagnon HTTP loopback uniquement quand TLS est actif
  ollama_keep_alive: "30m"       # "-1" jamais décharger, "0" décharger immédiatement
  ollama_num_ctx: 8192
  oai_stream_forward: false      # passthrough SSE backend réel optionnel
  anthropic_fallback_model: ""   # réécriture non-Claude optionnelle sur dispatch anthropic

vault:
  port: 56243
  host: ""
  admin_token: ""
  admin_ip_whitelist: []
  master_password: ""            # mot de passe de chiffrement de clé AES-GCM
  data_dir: ~/.wall-vault/data
  services_dir: ~/.wall-vault/services
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  bootstrap_port: 56247          # listener HTTP simple servant uniquement ca.crt

doctor:
  interval: 5m
  auto_fix: true
  log_file: /tmp/wall-vault-doctor.log

hooks:
  on_model_change: ""            # commande shell (env : SERVICE, MODEL)
  on_key_exhausted: ""
  on_service_down: ""
  on_doctor_fix: ""
  openclaw_socket: ""
```

### Variables d'environnement

Chaque champ YAML possède une surcharge env qui prime sur le fichier. Les plus courantes :

| Variable | Description |
|----------|-------------|
| `WV_LANG`, `WV_THEME` | Langue et thème |
| `WV_PROXY_PORT`, `WV_PROXY_HOST` | Adresse d'écoute du proxy |
| `WV_VAULT_PORT`, `WV_VAULT_HOST` | Adresse d'écoute du vault |
| `WV_VAULT_URL`, `WV_VAULT_TOKEN` | Points d'entrée mode distribué |
| `WV_ADMIN_TOKEN`, `WV_MASTER_PASS` | Identifiants vault |
| `WV_KEY_GOOGLE`, `WV_KEY_OPENROUTER`, `WV_KEY_ANTHROPIC`, `WV_KEY_OPENAI` | Clés API (séparées par virgules pour plusieurs) |
| `WV_PROXY_TLS_ENABLED`, `WV_PROXY_TLS_CERT`, `WV_PROXY_TLS_KEY` | TLS proxy |
| `WV_VAULT_TLS_ENABLED`, `WV_VAULT_TLS_CERT`, `WV_VAULT_TLS_KEY` | TLS vault |
| `WV_PROXY_PLAIN_PORT` | Compagnon HTTP loopback (`0` pour désactiver) |
| `WV_VAULT_BOOTSTRAP_PORT` | Listener bootstrap CA (`0` pour désactiver) |
| `WV_OLLAMA_URL`, `WV_OLLAMA_KEEP_ALIVE`, `WV_OLLAMA_NUM_CTX` | Réglages Ollama |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | Surcharges backend local |
| `WV_TOKEN_SENTINEL_FALLBACK` | Substitution sentinelle « proxy-managed » loopback |
| `WV_OAI_STREAM_FORWARD` | Passthrough SSE backend réel compat OpenAI |
| `WV_ANTHROPIC_FALLBACK_MODEL` | Réécriture non-Claude optionnelle sur anthropic |

---

## Modes

### Standalone (par défaut)

Vault et proxy s'exécutent dans le même processus. Idéal pour un hôte unique qui héberge à la fois les clés et l'agent. Écoute uniquement sur loopback par défaut.

```bash
wall-vault start    # lance les deux
```

### Distributed

Le vault tourne sur un hôte (l'**hôte vault**) et stocke toutes les clés ; plusieurs proxies sur d'autres hôtes s'authentifient chacun avec un jeton par client. Utile quand plusieurs machines ont besoin des mêmes clés sans devoir les copier partout.

**Hôte vault :**

```bash
WV_VAULT_HOST=0.0.0.0 wall-vault vault
```

**Chaque hôte proxy :**

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<ce-jeton-client> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

La modale **Add Client** du tableau de bord génère un jeton, enregistre un type d'agent, et le proxy récupère sa configuration via SSE sans redémarrer.

---

## Plugin yaml (backend prêt à brancher)

Tout backend compatible OpenAI peut être ajouté comme yaml sous `~/.wall-vault/services/`. wall-vault le détecte au démarrage, l'enregistre comme service routable, et le dispatch + le set de détection compat OAI + le pont stream Gemini le voient tous sans modification de code.

```yaml
# ~/.wall-vault/services/llamacpp.yaml
id: llamacpp
name: llama.cpp
enabled: true
default_url: http://localhost:8080
endpoints:
  generate: /v1/chat/completions
  list_models: /v1/models
auth:
  type: none
request_format: openai
inline_no_think_for_qwen3: false   # à activer si votre backend retire le marqueur
```

La topologie hub (un wall-vault devant un autre) est prise en charge via `tls_internal_ca: true`, `auth.type: bearer`, et `preserve_model_id: true`.

---

## Compilation depuis les sources

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
./wall-vault setup
./wall-vault start
```

Compilation croisée pour l'ensemble pris en charge :

```bash
make build-all   # linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
```

Les versions suivent `v{major}.{minor}.{patch}.{YYYYMMDD}.{HHmmss}` ; `BASE_VERSION` dans le Makefile fixe le préfixe.

### Disposition du projet

```
wall-vault/
├── main.go                     # dispatch CLI (start/proxy/vault/setup/cert/doctor)
├── cmd/
│   ├── setup/                  # assistant de configuration interactif
│   └── cert/                   # CA interne + émetteur de certificat TLS par hôte
├── internal/
│   ├── config/                 # chargeur YAML + env, chargeur de plugins
│   ├── proxy/                  # dispatch des requêtes, rotation des clés, convertisseurs de format
│   ├── vault/                  # store AES-GCM, tableau de bord, broker SSE
│   ├── doctor/                 # sonde de santé + auto-fix
│   ├── hooks/                  # déclencheurs d'événements par commande shell
│   └── i18n/                   # chaînes UI dans 17 langues
├── configs/services/           # plugins yaml fournis (lmstudio, vllm, ollama, …)
└── docs/                       # MANUAL, référence API, 16 variantes locales
```

---

## Documentation

- [Manuel utilisateur](docs/MANUAL.en.md) — installation, tableau de bord, agents, dépannage
- [Référence API](docs/API.en.md) — chaque point d'entrée avec les formats requête/réponse
- [CHANGELOG](CHANGELOG.md)

---

## Pile technique

- Go 1.25, binaire statique unique
- [templ](https://templ.guide) pour le tableau de bord rendu côté serveur, [HTMX](https://htmx.org) pour les mises à jour partielles
- AES-GCM (clé dérivée par PBKDF2) pour le chiffrement des clés au repos
- Server-Sent Events pour la synchronisation de configuration en direct entre vault et proxies
- CA interne auto-signée + certificats par hôte (pas de DNS public ni Let's Encrypt requis)

## Licence

GPL-3.0. Voir [LICENSE](LICENSE).

## Contribuer

Les pull requests sont les bienvenues. Voir [CONTRIBUTING.md](CONTRIBUTING.md). Pour des changements plus importants, merci d'ouvrir d'abord un issue pour discuter du design.
