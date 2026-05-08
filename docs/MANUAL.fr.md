# Manuel utilisateur de wall-vault

[English](MANUAL.md) · [한국어](MANUAL.ko.md) · [中文](MANUAL.zh.md) · [日本語](MANUAL.ja.md) · [Español](MANUAL.es.md) · **Français** · [Deutsch](MANUAL.de.md) · [Português](MANUAL.pt.md) · [العربية](MANUAL.ar.md) · [हिन्दी](MANUAL.hi.md) · [Bahasa Indonesia](MANUAL.id.md) · [ภาษาไทย](MANUAL.th.md) · [Kiswahili](MANUAL.sw.md) · [Hausa](MANUAL.ha.md) · [नेपाली](MANUAL.ne.md) · [Монгол](MANUAL.mn.md) · [isiZulu](MANUAL.zu.md)

Ce manuel couvre l'installation, la configuration et l'exploitation de wall-vault. Pour une vue d'ensemble rapide, voir le [README](../README.md). Pour les détails de l'API HTTP, voir la [référence API](API.md).

## Sommaire

1. [Ce que fait wall-vault](#ce-que-fait-wall-vault)
2. [Installation](#installation)
3. [Première exécution avec l'assistant de configuration](#première-exécution-avec-lassistant-de-configuration)
4. [Activation de TLS](#activation-de-tls)
5. [Enregistrement des clés API](#enregistrement-des-clés-api)
6. [Connexion des agents](#connexion-des-agents)
7. [Le tableau de bord](#le-tableau-de-bord)
8. [Mode distribué](#mode-distribué)
9. [Démarrage automatique](#démarrage-automatique)
10. [Plugins yaml](#plugins-yaml)
11. [Doctor](#doctor)
12. [Hooks](#hooks)
13. [Variables d'environnement](#variables-denvironnement)
14. [Dépannage](#dépannage)

---

## Ce que fait wall-vault

wall-vault est un binaire Go unique qui regroupe deux services coopérants :

- **Le coffre (vault)** stocke les clés API chiffrées au repos (AES-GCM avec un mot de passe maître), suit l'utilisation et les périodes de refroidissement par clé, diffuse les modifications via Server-Sent Events (SSE) et propose un tableau de bord web sur `:56243` pour les opérateurs humains.
- **Le proxy** expose des endpoints Gemini, Anthropic, compatibles OpenAI et Ollama-natifs sur `:56244`. Tout client IA pointant vers le proxy utilise les clés du coffre — les clients ne les voient jamais. Lorsqu'un fournisseur en amont échoue, la distribution bascule vers le fournisseur suivant dans l'ordre.

C'est utile lorsque :

- Vous avez des clés pour plusieurs fournisseurs et voulez une seule URL à laquelle l'agent s'adresse.
- Vous voulez qu'une clé free-tier en cooldown s'efface sans interrompre la session.
- Vous voulez que les mêmes clés alimentent plusieurs bots, IDE ou scripts sur le même LAN sans copier les identifiants.
- Vous voulez un tableau de bord, et non des variables d'environnement, pour modifier les clés et changer de modèle.
- Vous voulez un repli local (Ollama, LM Studio, vLLM) lorsque les limites cloud sont épuisées.

```
   AI client (OpenClaw, Claude Code, Cursor, …)
            │
            ▼
   wall-vault proxy  :56244
            │  (selects key, dispatches, falls back on failure)
            ├──► Google Gemini
            ├──► Anthropic
            ├──► OpenAI
            ├──► OpenRouter (340+ models, auto :free fallback)
            └──► Local OAI-compat backends (Ollama / LM Studio / vLLM / …)

   vault (AES-GCM key store + dashboard)  :56243
            ▲
            │  SSE broadcast on change
   Multiple proxies on different hosts can share one vault.
```

---

## Installation

### Linux / macOS one-liner

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

Le script détecte automatiquement le système d'exploitation et l'architecture, télécharge le bon binaire dans `~/.local/bin/wall-vault` et le rend exécutable. Si `~/.local/bin` n'est pas dans votre `PATH`, ajoutez-le :

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### Téléchargement manuel

Des binaires précompilés sont publiés à chaque release sur `https://github.com/sookmook/wall-vault/releases`.

```bash
# Linux amd64
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# Linux arm64 (Raspberry Pi, ARM servers)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-arm64 \
  -o wall-vault && chmod +x wall-vault

# macOS Apple Silicon
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o wall-vault && chmod +x wall-vault

# macOS Intel
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-amd64 \
  -o wall-vault && chmod +x wall-vault
```

### Windows

```powershell
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe
```

### Compilation depuis les sources

Nécessite Go 1.25 ou plus récent.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
```

`make build-all` compile en croisé pour les cinq plateformes prises en charge. Les binaires sont générés dans `bin/`.

---

## Première exécution avec l'assistant de configuration

```bash
wall-vault setup
```

L'assistant vous demande, dans l'ordre :

1. **Langue** — choisit l'une des 17 locales d'interface. Détectée automatiquement à partir de `$LANG` ; l'assistant propose une liste malgré tout.
2. **Thème** — `light` (par défaut), `dark`, `cherry`, `ocean`, `gold`, `autumn`, `winter`. Purement cosmétique.
3. **Mode** — `standalone` (hôte unique, par défaut) ou `distributed` (coffre sur un hôte, proxies sur d'autres).
4. **Nom du bot** — un slug `client_id` libre. Le coffre l'utilise pour cibler la configuration par client (surcharges de modèle, chaînes de fallback).
5. **Port du proxy** — défaut `56244`.
6. **Port du coffre** — défaut `56243` (standalone uniquement).
7. **Sélection des services** — un y/N pour chacun de : Google Gemini, OpenRouter, Anthropic, OpenAI, Ollama, LM Studio, vLLM. Plusieurs choix sont acceptés ; chacun écrit son indication de variable d'environnement à la fin.
8. **Filtre d'outils** — `strip_all` (par défaut ; bloque toutes les définitions d'outils entrants pour la sécurité) ou `passthrough` (laisse passer tous les outils).
9. **Jeton admin** — laissez vide pour génération automatique. Le tableau de bord exige ce jeton pour la connexion.
10. **Mot de passe maître** — laissez vide pour aucun chiffrement (NON recommandé) ; définissez une valeur pour chiffrer le magasin de clés au repos avec AES-GCM.
11. **Chemin de sauvegarde** — par défaut `wall-vault.yaml` dans le répertoire courant. Le chargeur consulte également `~/.wall-vault/config.yaml`.

Après l'enregistrement, l'assistant exécute `doctor.FixTrust` pour que tout agent installé localement (OpenClaw, Claude Code, Cline) reçoive automatiquement l'autorité de certification interne de wall-vault dans son magasin de confiance. Si aucun agent de ce type n'est installé, l'étape affiche `SKIP` et n'écrit rien.

Puis démarrez le binaire :

```bash
wall-vault start
```

`start` exécute à la fois le coffre et le proxy dans un seul processus (mode standalone). Pour le mode distribué, utilisez `wall-vault vault` sur l'hôte du coffre et `wall-vault proxy` sur chaque hôte proxy.

Ouvrez `http://localhost:56243` dans un navigateur. Connectez-vous avec le jeton admin que l'assistant a affiché.

---

## Activation de TLS

Les valeurs par défaut de l'assistant laissent les deux écouteurs en HTTP simple. La plupart des agents (OpenClaw, Claude Code, Cursor) fonctionnent mieux face à un seul endpoint HTTPS, donc TLS est recommandé pour tout déploiement qui dépasse la machine locale.

wall-vault est livré avec sa propre autorité de certification interne, vous n'avez donc pas besoin d'un nom DNS public ni de Let's Encrypt.

```bash
# 1. Create the internal CA — written to ~/.wall-vault/ca.{crt,key}.
#    The CA is good for 10 years by default; override with --ca-years.
wall-vault cert init

# 2. Issue a host certificate. Subject Alternative Names automatically include:
#       hostname, "localhost", "127.0.0.1", and any non-loopback LAN IP detected.
#    Override the issuer dir with --dir, validity with --host-years.
wall-vault cert issue $(hostname)

# 3. Trust the CA in this machine's OS keychain.
#    Linux: writes to /etc/ssl/certs/ via update-ca-certificates (needs sudo).
#    macOS: adds to the System keychain via security add-trusted-cert (needs sudo).
#    Windows: imports into CurrentUser\Root via certutil (no admin needed).
wall-vault cert install-trust

# 4. Enable TLS on both listeners.
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"

wall-vault start
```

Pour étendre la confiance à d'autres machines du LAN, copiez `~/.wall-vault/ca.crt` et exécutez `wall-vault cert install-trust --ca <path>` sur chacune. Le coffre expose également `ca.crt` via un petit écouteur HTTP simple sur `:56247` (le **port d'amorçage**) pour le cas du dilemme où un nouveau client a besoin de l'autorité de certification pour communiquer en HTTPS.

### Compagnon HTTP en boucle locale

Certains agents — notamment le runtime Node intégré d'OpenClaw — réécrivent `NODE_EXTRA_CA_CERTS` au démarrage du processus, supprimant tout indice d'autorité de certification fourni par l'opérateur. Ils ne peuvent pas honorer l'autorité de certification de wall-vault depuis l'intérieur du démon, même après `cert install-trust`. wall-vault contourne cela en liant un **écouteur HTTP simple supplémentaire restreint à la boucle locale** sur `127.0.0.1:56245` chaque fois que TLS est activé. Les clients sur le même hôte atteignent le proxy via ce port sans TLS du tout ; les clients du LAN continuent à utiliser l'écouteur TLS.

Désactivez avec `WV_PROXY_PLAIN_PORT=0` si vous n'en avez pas besoin.

### `wall-vault cert list`

Affiche chaque certificat sous `~/.wall-vault/` avec son sujet, sa fenêtre de validité et ses SAN.

```
$ wall-vault cert list
ca.crt          subject=wall-vault internal CA   not-after=2036-05-05
hostname.crt    subject=hostname                 not-after=2031-05-05   SAN=hostname,localhost,127.0.0.1,192.168.…
```

---

## Enregistrement des clés API

Deux façons : le tableau de bord ou les variables d'environnement.

### Tableau de bord (recommandé)

1. Connectez-vous sur `https://localhost:56243` avec le jeton admin.
2. Cliquez sur **+ API key** dans la carte des clés.
3. Choisissez un service (Google, OpenRouter, Anthropic, OpenAI, …).
4. Collez la clé. Enregistrez.

Plusieurs clés par service sont acceptées ; le proxy effectue un round-robin entre elles et ignore celles qui ont atteint un cooldown par clé.

### Variables d'environnement (amorçage unique)

```bash
export WV_KEY_GOOGLE="AIzaSyA1...,AIzaSyB2...,AIzaSyC3..."   # comma-separated
export WV_KEY_OPENROUTER="sk-or-v1-…"
export WV_KEY_ANTHROPIC="sk-ant-…"
export WV_KEY_OPENAI="sk-…"
wall-vault start
```

Les clés fournies de cette façon sont écrites dans le magasin chiffré au premier démarrage. Les démarrages suivants les lisent depuis le disque ; vous pouvez désactiver les variables d'environnement après la première exécution.

### Cooldowns et rotation

Chaque appel réussi incrémente le `usage_count` de la clé et rafraîchit `last_used`. Sur HTTP 429 / 402 / 403, le proxy met la clé en **cooldown** (par défaut : 60 minutes pour 429, 24 heures pour 402, 12 heures pour 403). La distribution suivante choisit une autre clé pour ce service. Lorsque toutes les clés d'un service sont en cooldown, le proxy saute rapidement ce service entièrement et essaie le fournisseur suivant dans la chaîne de fallback.

Les cooldowns sont visibles par clé dans le tableau de bord avec un compte à rebours.

---

## Connexion des agents

### OpenClaw

OpenClaw est le client cible d'origine. Utilisez la modale **+ Add agent** du tableau de bord :

- Définissez **Agent type** sur `openclaw` ou `nanoclaw`.
- Définissez **Work directory** — pour OpenClaw, cela se remplit automatiquement avec `~/.openclaw`.
- Choisissez un **preferred service** et éventuellement un **model override**.
- Cliquez sur **Apply**. wall-vault écrit directement `~/.openclaw/openclaw.json` (URLs des fournisseurs, jeton du coffre, entrées de modèles).

Lorsque vous changez de modèle depuis le tableau de bord, OpenClaw prend en compte le changement via SSE en 1 à 3 secondes — sans redémarrage.

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

Lorsque les crédits Anthropic en amont sont épuisés, la distribution bascule vers les services listés dans `fallback_services` de ce client. Par défaut, un id de modèle non-Claude envoyé à la distribution anthropic renvoie une erreur, afin que les erreurs de routage apparaissent immédiatement. Acceptez la réécriture automatique :

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

### Cursor

Dans Cursor **Settings → AI → OpenAI API** :

```
Base URL:  https://localhost:56244
API Key:   <your-vault-client-token>
Model:     gemini-2.5-flash    # or any model wall-vault knows
```

### Continue (VS Code, JetBrains)

`config.json` :

```json
{
  "models": [
    {
      "title": "wall-vault",
      "provider": "openai",
      "model": "gemini-2.5-flash",
      "apiBase": "https://localhost:56244/v1",
      "apiKey": "<your-vault-client-token>"
    }
  ]
}
```

### HTTP personnalisé

```bash
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer <your-vault-client-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "hello"}]
  }'
```

Le même endpoint accepte le streaming (`"stream": true`) lorsque `proxy.oai_stream_forward: true` est défini.

---

## Le tableau de bord

`https://localhost:56243`. Cinq cartes sur la grille d'accueil :

- **Keys** — chaque clé API, regroupée par service. Ajouter, modifier, supprimer ; voir l'utilisation et le cooldown.
- **Services** — Google / OpenRouter / Anthropic / OpenAI / Ollama / LM Studio / vLLM / llama.cpp, plus tout plugin yaml dans `~/.wall-vault/services/`. Définissez `default_model`, `allowed_models` par service, l'URL de base et le bouton de raisonnement.
- **Clients (agents)** — chaque client enregistré (bot OpenClaw, session Claude Code, instance Cursor, …). Attribuez le service préféré, la surcharge de modèle, la chaîne de fallback.
- **Proxies** — chaque proxy qui s'est authentifié auprès de ce coffre. Statut en direct (en ligne/hors ligne), dernière vue, modèle actuel.
- **Settings** — jeton admin, rotation du mot de passe maître, thème, langue.

Chaque carte a un volet d'édition (côté droit). Un clic à l'extérieur ou `Esc` le ferme. Les modifications sont propagées à tous les proxies connectés via SSE en quelques secondes.

Le **pied de page** affiche un indicateur SSE (vert = connecté, orange = reconnexion, gris = déconnecté) et la version du build en direct.

---

## Mode distribué

Lorsque vous avez plusieurs machines qui ont toutes besoin des mêmes clés, exécutez le coffre sur un hôte et les proxies sur chacun des autres.

### Hôte du coffre

```bash
WV_VAULT_HOST=0.0.0.0 \
WV_ADMIN_TOKEN=<admin> \
WV_MASTER_PASS=<master> \
wall-vault vault
```

Le tableau de bord est désormais accessible sur `https://<vault-host>:56243`. Ajoutez un agent pour chaque proxy distant dans la carte **Clients** ; chacun génère un `vault_token` unique.

### Hôtes des proxies

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<that-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

Le proxy s'authentifie auprès du coffre, ouvre un flux SSE et applique toute configuration qu'il reçoit (service préféré, surcharge de modèle, chaîne de fallback). Les modifications ultérieures du coffre arrivent en quelques secondes sans redémarrage.

Pour les installations couvrant le LAN, activez TLS sur l'hôte du coffre (`WV_VAULT_TLS_ENABLED=1` + les variables d'environnement de cert/clé) et exécutez chaque hôte proxy via la même étape `wall-vault cert install-trust` afin que les appels HTTPS du proxy vers le coffre soient approuvés.

---

## Démarrage automatique

### systemd (Linux)

```ini
# ~/.config/systemd/user/wall-vault-proxy.service
[Unit]
Description=wall-vault proxy
After=network-online.target

[Service]
Type=simple
ExecStart=%h/.local/bin/wall-vault proxy
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
```

```bash
systemctl --user enable --now wall-vault-proxy
loginctl enable-linger $USER       # so the unit keeps running after logout
```

Pour le coffre sur le même hôte, écrivez un `wall-vault-vault.service` parallèle. Pour le mode standalone, une seule unité appelant `wall-vault start` suffit.

### launchd (macOS)

```xml
<!-- ~/Library/LaunchAgents/com.wall-vault.proxy.plist -->
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key><string>com.wall-vault.proxy</string>
  <key>ProgramArguments</key>
  <array><string>/usr/local/bin/wall-vault</string><string>proxy</string></array>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><true/>
  <key>StandardOutPath</key><string>/tmp/wall-vault.proxy.log</string>
  <key>StandardErrorPath</key><string>/tmp/wall-vault.proxy.err</string>
</dict>
</plist>
```

```bash
launchctl load ~/Library/LaunchAgents/com.wall-vault.proxy.plist
```

### Windows

Utilisez `nssm` pour envelopper `wall-vault.exe start` en tant que service Windows, ou une entrée `schtasks` qui s'exécute à l'ouverture de session de l'utilisateur.

---

## Plugins yaml

Tout backend compatible OpenAI peut être ajouté sans modification de code en déposant un yaml dans `~/.wall-vault/services/`. wall-vault le charge au démarrage et enregistre le service pour la distribution, l'ensemble de détection OAI-compat et le pont Gemini-stream.

```yaml
# ~/.wall-vault/services/llamacpp.yaml
id: llamacpp                 # unique service id
name: llama.cpp              # human label
enabled: true                # disabled plugins are skipped at load

default_url: http://localhost:8080   # operator override; env wins (WV_LLAMACPP_URL)
endpoints:
  generate: /v1/chat/completions
  list_models: /v1/models

auth:
  type: none                 # none | bearer | query_param | header
  param: ""                  # for query_param: the param name (e.g. "key")

request_format: openai       # openai | gemini | ollama | raw

model_fetch:
  enabled: true              # let the dashboard auto-detect models
  dynamic: true              # re-fetch on every dashboard open
  auto_detect_url: true      # try /v1/models even when not declared

concurrency:
  max: 1                     # max concurrent requests to this backend
  queue_size: 10
  wait_notify: true          # show "queued" hint to TUI agents

error_codes:
  503:
    cooldown: 5m
    message: "llama.cpp not responding"

# Opt in to qwen3-family inline /no_think directive when reasoning is off.
# Set true if your backend's chat template strips the marker (LM Studio's
# jinja, Ollama's /v1 layer). Other backends typically echo the literal
# text back, so this stays opt-in per yaml.
inline_no_think_for_qwen3: false

# Hub topology — point at another wall-vault. Required when this plugin
# fronts a remote wall-vault (so the receiving wall-vault sees the
# publisher prefix and routes correctly) and so the bearer token in
# proxy.vault_token is sent as Authorization.
preserve_model_id: false
tls_internal_ca: false       # add ~/.wall-vault/ca.crt to client trust pool
```

L'ensemble fourni dans `configs/services/` (lmstudio, vllm, llamacpp, tgwui, localai, jan, koboldcpp, tabbyapi, mlx-server, litellm-proxy, ollama, google, openrouter) est livré désactivé par défaut. Copiez celui que vous voulez dans `~/.wall-vault/services/`, mettez `enabled: true`, redémarrez.

---

## Doctor

`wall-vault doctor` exécute une sonde de santé unique sur l'ensemble de l'installation :

```
✓ vault listener  (https://localhost:56243)
✓ proxy listener  (https://localhost:56244)
✓ master password set
⚠ Google: 2 keys, all on cooldown
✓ Anthropic: 1 key healthy
✗ Ollama: not reachable at http://localhost:11434
```

Chaque ligne est l'un de :

- `✓` — sain
- `⚠` — dégradé mais fonctionnel (une clé en cooldown, quota faible, etc.)
- `✗` — cassé
- `SKIP` — non configuré / non applicable sur cet hôte

Un second mode démon exécute la même sonde tous les `doctor.interval` (par défaut 5 minutes) et écrit les résultats dans `doctor.log_file` (par défaut `/tmp/wall-vault-doctor.log`). Lorsque `doctor.auto_fix` est activé, il tente également de réparer les dérives courantes (config OpenClaw obsolète, confiance TLS manquante, services redémarrables).

Déclenchez une exécution unique depuis le tableau de bord via la carte **Doctor** ou `wall-vault doctor`.

---

## Hooks

Exécutez une commande shell sur des événements clés :

```yaml
hooks:
  on_model_change:   "logger 'wall-vault: $SERVICE/$MODEL'"
  on_key_exhausted:  "notify-send 'wall-vault' '$SERVICE keys all on cooldown'"
  on_service_down:   "/usr/local/bin/page-oncall.sh $SERVICE '$ERROR'"
  on_doctor_fix:     "echo \"$AGENT: $LEVEL $MSG\" >> ~/wall-vault.audit.log"
  openclaw_socket:   ""    # if set, OpenClaw TUI receives events over this Unix socket
```

Chaque hook reçoit des variables d'environnement spécifiques à l'événement (`SERVICE`, `MODEL`, `ERROR`, `AGENT`, `LEVEL`, `MSG`). Les hooks s'exécutent de manière asynchrone avec un timeout de 5 secondes — le proxy ne se bloque jamais sur un hook lent.

---

## Variables d'environnement

| Variable | Champ YAML |
|----------|------------|
| `WV_LANG` | `lang` |
| `WV_THEME` | `theme` |
| `WV_PROXY_PORT` | `proxy.port` |
| `WV_PROXY_HOST` | `proxy.host` |
| `WV_VAULT_PORT` | `vault.port` |
| `WV_VAULT_HOST` | `vault.host` |
| `WV_VAULT_URL` | `proxy.vault_url` (distributed) |
| `WV_VAULT_TOKEN` | `proxy.vault_token` |
| `WV_ADMIN_TOKEN` | `vault.admin_token` |
| `WV_MASTER_PASS` | `vault.master_password` |
| `WV_AVATAR` | `proxy.avatar` |
| `WV_TOOL_FILTER` | `proxy.tool_filter` |
| `WV_CC_CLIENT_ID` | `proxy.claude_code_client_id` |
| `WV_PROXY_TLS_ENABLED` | `proxy.tls.enabled` |
| `WV_PROXY_TLS_CERT` | `proxy.tls.cert_file` |
| `WV_PROXY_TLS_KEY` | `proxy.tls.key_file` |
| `WV_PROXY_TLS_REQUIRED` | `proxy.tls.required` (refuse de démarrer si TLS désactivé — bloque le repli en clair) |
| `WV_PROXY_ALLOW_CIDRS` | `proxy.allow_cidrs` (liste séparée par des virgules, ex. `192.168.0.0/16,10.0.0.0/8` ; loopback toujours autorisé) |
| `WV_VAULT_TLS_ENABLED` | `vault.tls.enabled` |
| `WV_VAULT_TLS_CERT` | `vault.tls.cert_file` |
| `WV_VAULT_TLS_KEY` | `vault.tls.key_file` |
| `WV_VAULT_BOOTSTRAP_PORT` | `vault.bootstrap_port` |
| `WV_PROXY_PLAIN_PORT` | `proxy.plain_port` |
| `WV_KEY_GOOGLE` | Import unique : clés Google séparées par des virgules |
| `WV_KEY_OPENROUTER` | Import unique : clés OpenRouter |
| `WV_KEY_ANTHROPIC` | Import unique : clés Anthropic |
| `WV_KEY_OPENAI` | Import unique : clés OpenAI |
| `WV_OLLAMA_URL` | Surcharge d'URL Ollama par hôte (instance unique) |
| `WV_OLLAMA_URLS` | URLs Ollama séparées par virgules (dispatch multi-instance) |
| `WV_OLLAMA_KEEP_ALIVE` | `proxy.ollama_keep_alive` |
| `WV_OLLAMA_NUM_CTX` | `proxy.ollama_num_ctx` |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | Surcharge d'URL par backend (instance unique) |
| `WV_TOKEN_SENTINEL_FALLBACK` | `proxy.token_sentinel_fallback` |
| `WV_OAI_STREAM_FORWARD` | `proxy.oai_stream_forward` |
| `WV_ANTHROPIC_FALLBACK_MODEL` | `proxy.anthropic_fallback_model` |
| `WV_INJECT_MODEL_IDENTITY` | `proxy.inject_model_identity` (garde-fou d'identité par message système, désactivé par défaut) |
| `WV_PROMPT_TOKEN_CAP` | Plafond d'auto-troncature par hôte pour prompts OAI-compat locaux (entier positif = activer, 0 = off) |
| `WV_DISPATCH_TRACE` | Définir `1` pour journaliser le service/modèle résolu et la raison de chaque dispatch (off par défaut) |
| `WV_ECONOWORLD_MAX_TOKENS` | `proxy.econoworld_max_tokens` |
| `WV_ECONOWORLD_STREAM` | `proxy.econoworld_stream` |
| `WV_ECONOWORLD_REQUEST_TIMEOUT` | `proxy.econoworld_request_timeout` |

Chaque variable d'environnement, lorsqu'elle est définie, prévaut sur le fichier YAML.

---

## Dépannage

### `connection refused` sur `:56244`

Soit le proxy ne tourne pas, soit il est lié à un hôte différent. Vérifiez :

```bash
ss -lnp | grep 56244
systemctl --user status wall-vault-proxy   # Linux
launchctl list | grep wall-vault           # macOS
```

S'il tourne sur un port différent, votre configuration a `proxy.port` surchargé — vérifiez `~/.wall-vault/config.yaml`.

### `x509: certificate signed by unknown authority`

Le client n'approuve pas l'autorité de certification interne de wall-vault. Exécutez `wall-vault cert install-trust` sur la machine cliente. Pour les agents dont le runtime ignore le magasin de confiance de l'OS (par exemple Node avec un `NODE_EXTRA_CA_CERTS` codé en dur), utilisez le compagnon HTTP en boucle locale sur `127.0.0.1:56245` (uniquement sur le même hôte) ou définissez `WV_PROXY_TLS_ENABLED=0` pour revenir à HTTP simple.

### `token not registered with vault`

L'`Authorization: Bearer <token>` du client ne correspond à aucun client enregistré. Vérifiez le jeton sous **Clients** dans le tableau de bord. Si vous avez copié un jeton littéral comme `proxy-managed`, `dummy` ou `""` depuis une configuration obsolète, remplacez-le par le vrai jeton client.

### `Anthropic dispatch needs a Claude model id`

Comportement par défaut depuis v0.2.63 : un id de modèle non-Claude envoyé à la distribution anthropic renvoie une erreur. Soit corrigez le routage (n'envoyez pas `gemini-2.5-flash` à anthropic), soit acceptez la réécriture automatique via `proxy.anthropic_fallback_model`.

### `unknown service: <id>`

La distribution a vu un id de service qu'aucun plugin yaml n'a revendiqué. Vérifiez :

```bash
ls ~/.wall-vault/services/        # any plugin yaml present?
cat ~/.wall-vault/services/<id>.yaml | grep enabled
```

Si le yaml existe mais est `enabled: false`, basculez-le. S'il manque entièrement, copiez-le depuis `configs/services/` dans l'arborescence des sources.

### Réponse vide sur un modèle de raisonnement

`qwen3.6`, `deepseek-r1` et la famille GPT-`o1` émettent parfois uniquement `reasoning_content` et laissent `content` vide. Depuis v0.2.63, wall-vault bascule automatiquement vers le texte de raisonnement — si vous voyez toujours des réponses vides, le backend ne renvoie aucun des deux champs. Vérifiez les journaux en amont.

Pour LM Studio avec qwen3 spécifiquement, définissez `inline_no_think_for_qwen3: true` dans le plugin yaml afin que le raisonnement soit désactivé en ligne. Les fichiers lmstudio.yaml et ollama.yaml intégrés le font déjà.

### Le tableau de bord affiche « toutes les clés en cooldown » mais je viens d'en ajouter une

La nouvelle clé est saine mais le chemin de distribution peut encore être en cooldown pour une clé plus ancienne. Essayez une nouvelle requête — le proxy effectue un round-robin par appel, et une clé saine sera choisie ensuite.

### Le coffre ne se déverrouille pas avec le mot de passe maître

Mauvais mot de passe. Il n'y a pas de récupération — wall-vault ne livre délibérément pas de porte dérobée. Si vous avez réellement perdu le mot de passe maître, le seul chemin est de supprimer `~/.wall-vault/data/vault.json`, redémarrer avec un nouveau mot de passe et réajouter les clés.

### Limites OpenRouter free-tier atteintes

Définissez `proxy.services` pour inclure `openrouter` et ajoutez au moins une clé OpenRouter. Le proxy bascule automatiquement d'un modèle payant vers sa variante `:free` lorsque le chemin payant renvoie 402 / 429.

### `journalctl --user -u wall-vault-proxy` est vide

Les journaux systemd `--user` vont au journal de l'utilisateur qui le lance. Si vous avez démarré l'unité en tant que `root` ou via `sudo`, le journal se trouve dans l'instance système — essayez `journalctl -u wall-vault-proxy` sans `--user`.

---

## Plus d'informations

- Référence de l'API HTTP — voir [API.md](API.md)
- Code source — `https://github.com/sookmook/wall-vault`
- Rapports de bogues / demandes de fonctionnalités — GitHub Issues
- Historique des releases — [CHANGELOG.md](../CHANGELOG.md)
