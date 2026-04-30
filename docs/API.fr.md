# Manuel API wall-vault

Ce document décrit en détail tous les points de terminaison HTTP API de wall-vault.

---

## Table des matières

- [Authentification](#authentification)
- [API Proxy (:56244)](#api-proxy-56244)
  - [Vérification de santé](#get-health)
  - [Statut](#get-status)
  - [Liste des modèles](#get-apimodels)
  - [Changer de modèle](#put-apiconfigmodel)
  - [Mode réflexion](#put-apiconfigthink-mode)
  - [Recharger la configuration](#post-reload)
  - [API Gemini](#post-googlev1betamodelsmgeneratecontent)
  - [Streaming Gemini](#post-googlev1betamodelsmstreamgeneratecontent)
  - [API compatible OpenAI](#post-v1chatcompletions)
- [API Coffre-fort de clés (:56243)](#api-coffre-fort-de-clés-56243)
  - [API publique](#api-publique-aucune-authentification-requise)
  - [Flux d'événements SSE](#get-apievents)
  - [API réservée au proxy](#api-réservée-au-proxy-jeton-client)
  - [API Admin — Clés](#api-admin--clés-api)
  - [API Admin — Clients](#api-admin--clients)
  - [API Admin — Services](#api-admin--services)
  - [API Admin — Liste des modèles](#api-admin--liste-des-modèles)
  - [API Admin — Statut des proxys](#api-admin--statut-des-proxys)
- [Types d'événements SSE](#types-dévénements-sse)
- [Routage fournisseur/modèle](#routage-fournisseurmodèle)
- [Schéma de données](#schéma-de-données)
- [Réponses d'erreur](#réponses-derreur)
- [Exemples cURL](#exemples-curl)

---

## Authentification

| Portée | Méthode | En-tête |
|--------|---------|---------|
| API Admin | Jeton Bearer | `Authorization: Bearer <admin_token>` |
| Proxy → Coffre-fort | Jeton Bearer | `Authorization: Bearer <client_token>` |
| API Proxy | Aucune (local) | — |

Si `admin_token` n'est pas défini (chaîne vide), toutes les API admin sont accessibles sans authentification.

### Politique de sécurité

- **Limitation de débit** : Si l'authentification de l'API admin échoue plus de 10 fois en 15 minutes, l'IP est temporairement bloquée (`429 Too Many Requests`)
- **Liste blanche d'IP** : Seules les IP/CIDR enregistrées dans le champ `ip_whitelist` de l'agent (`Client`) sont autorisées à accéder à `/api/keys`. Si le tableau est vide, toutes les IP sont autorisées.
- **Protection theme/lang** : `/admin/theme` et `/admin/lang` nécessitent également l'authentification par jeton admin

---

## API Proxy (:56244)

Le serveur sur lequel le proxy s'exécute. Port par défaut `56244`.

---

### `GET /health`

Vérification de santé. Retourne toujours 200 OK.

**Exemple de réponse :**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "client": "bot-a"
}
```

---

### `GET /status`

Statut détaillé du proxy.

**Exemple de réponse :**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "client": "bot-a",
  "service": "google",
  "model": "gemini-2.5-flash",
  "sse": true,
  "filter": "strip_all",
  "services": ["google", "openrouter", "ollama"],
  "mode": "distributed"
}
```

| Champ | Type | Description |
|-------|------|-------------|
| `service` | string | Service par défaut actuel |
| `model` | string | Modèle par défaut actuel |
| `sse` | bool | Si la connexion SSE au coffre-fort est établie |
| `filter` | string | Mode de filtrage des outils |
| `services` | []string | Liste des services actifs |
| `mode` | string | `standalone` \| `distributed` |

---

### `GET /api/models`

Lister les modèles disponibles. Utilise un cache TTL (10 minutes par défaut).

**Paramètres de requête :**

| Paramètre | Description | Exemple |
|-----------|-------------|---------|
| `service` | Filtre par service | `?service=google` |
| `q` | Recherche par ID/nom de modèle | `?q=gemini` |

**Exemple de réponse :**
```json
{
  "models": [
    {
      "id": "gemini-2.5-pro",
      "name": "Gemini 2.5 Pro",
      "service": "google",
      "context_length": 1048576,
      "free": false
    },
    {
      "id": "openrouter/hunter-alpha",
      "name": "Hunter Alpha (1M ctx, free)",
      "service": "openrouter",
      "context_length": 1048576,
      "free": true
    }
  ],
  "count": 2
}
```

| Champ | Type | Description |
|-------|------|-------------|
| `id` | string | ID du modèle |
| `name` | string | Nom d'affichage du modèle |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` etc. |
| `context_length` | int | Taille de la fenêtre de contexte |
| `free` | bool | S'il s'agit d'un modèle gratuit (OpenRouter) |

---

### `PUT /api/config/model`

Changer le service et le modèle actuels.

**Corps de la requête :**
```json
{
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

**Réponse :**
```json
{
  "status": "ok",
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

> **Remarque :** En mode distribué, il est recommandé d'utiliser `PUT /admin/clients/{id}` du coffre-fort au lieu de cette API. Les modifications du coffre-fort sont automatiquement propagées via SSE en 1 à 3 secondes.

---

### `PUT /api/config/think-mode`

Basculer le mode réflexion (no-op, réservé pour extension future).

**Réponse :**
```json
{"status": "ok"}
```

---

### `POST /reload`

Resynchroniser immédiatement les paramètres client et les clés depuis le coffre-fort.

**Réponse :**
```json
{"status": "reloading"}
```

La resynchronisation s'exécute de manière asynchrone et se termine dans les 1 à 2 secondes suivant la réception de la réponse.

---

### `POST /google/v1beta/models/{model}:generateContent`

Proxy API Gemini (non-streaming).

**Paramètre de chemin :**
- `{model}` : ID du modèle. Si le préfixe `gemini-` est présent, le service Google est automatiquement sélectionné.

**Corps de la requête :** [Format de requête Gemini generateContent](https://ai.google.dev/api/generate-content)

```json
{
  "contents": [
    {
      "role": "user",
      "parts": [{"text": "안녕하세요"}]
    }
  ],
  "generationConfig": {
    "temperature": 0.7,
    "maxOutputTokens": 1024
  }
}
```

**Corps de la réponse :** Format de réponse Gemini generateContent

**Filtre d'outils :** Lorsque `tool_filter: strip_all` est défini, le tableau `tools` de la requête est automatiquement supprimé.

**Chaîne de repli :** Si le service désigné échoue → repli dans l'ordre des services configurés → Ollama (dernier recours).

---

### `POST /google/v1beta/models/{model}:streamGenerateContent`

Proxy streaming API Gemini. Le format de requête est identique au non-streaming. La réponse est un flux SSE :

```
data: {"candidates":[{"content":{"parts":[{"text":"안"}],...},...}]}

data: [DONE]
```

---

### `POST /v1/chat/completions`

API compatible OpenAI. Convertit en interne au format Gemini pour le traitement.

**Corps de la requête :**
```json
{
  "model": "gemini-2.5-flash",
  "messages": [
    {"role": "system", "content": "당신은 도움이 되는 어시스턴트입니다."},
    {"role": "user", "content": "안녕하세요"}
  ],
  "temperature": 0.7,
  "max_tokens": 1024,
  "stream": false
}
```

**Support du préfixe fournisseur dans le champ `model` (OpenClaw 3.11+) :**

| Exemple de modèle | Routage |
|-------------------|---------|
| `gemini-2.5-flash` | Service configuré actuel |
| `google/gemini-2.5-pro` | Google direct |
| `openai/gpt-4o` | OpenAI direct |
| `anthropic/claude-opus-4-6` | Via OpenRouter |
| `openrouter/meta-llama/llama-3.3-70b` | OpenRouter direct |
| `wall-vault/gemini-2.5-flash` | Détection auto → Google |
| `wall-vault/claude-opus-4-6` | Détection auto → OpenRouter Anthropic |
| `wall-vault/gpt-4o` | Détection auto → OpenAI |
| `wall-vault/hunter-alpha` | OpenRouter (gratuit, 1M contexte) |
| `moonshot/kimi-k2.5` | Via OpenRouter |
| `opencode-go/model` | Via OpenRouter |
| `kimi-k2.5:cloud` | Suffixe `:cloud` → OpenRouter |

Pour plus de détails, voir [Routage fournisseur/modèle](#routage-fournisseurmodèle).

**Corps de la réponse :**
```json
{
  "choices": [
    {
      "message": {
        "role": "assistant",
        "content": "안녕하세요! 무엇을 도와드릴까요?"
      },
      "finish_reason": "stop",
      "index": 0
    }
  ]
}
```

> **Suppression automatique des jetons de contrôle de modèle :** Si la réponse contient des délimiteurs GLM-5 / DeepSeek / ChatML (`<|im_start|>`, `[gMASK]`, `[sop]`, etc.), ils sont automatiquement supprimés.

---

## API Coffre-fort de clés (:56243)

Le serveur sur lequel le coffre-fort de clés s'exécute. Port par défaut `56243`.

---

### API publique (Aucune authentification requise)

#### `GET /`

Interface web du tableau de bord. Accès via navigateur.

---

#### `GET /api/status`

Statut du coffre-fort.

**Exemple de réponse :**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "keys": 3,
  "clients": 2,
  "sse": 2
}
```

---

#### `GET /api/clients`

Liste des clients enregistrés (informations publiques uniquement, jetons exclus).

---

### `GET /api/events`

Flux d'événements SSE (Server-Sent Events) en temps réel.

**En-têtes :**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**Reçu immédiatement à la connexion :**
```
data: {"type":"connected","clients":2}
```

**Exemples d'événements :**
```
data: {"type":"config_change","data":{"client_id":"bot-a","service":"google","model":"gemini-2.5-pro"}}

data: {"type":"key_added","data":{"service":"openrouter"}}

data: {"type":"key_deleted","data":{"service":"google"}}

data: {"type":"service_changed","data":{"action":"updated","id":"ollama"}}

data: {"type":"usage_reset","data":{"time":"2026-03-13T00:00:30Z"}}
```

Pour les types d'événements détaillés, voir [Types d'événements SSE](#types-dévénements-sse).

---

### API réservée au proxy (Jeton client)

Nécessite l'en-tête `Authorization: Bearer <client_token>`. Les jetons admin sont également acceptés.

#### `GET /api/keys`

Liste des clés API déchiffrées fournies au proxy.

**Paramètres de requête :**

| Paramètre | Description |
|-----------|-------------|
| `service` | Filtre par service (ex. : `?service=google`) |

**Exemple de réponse :**
```json
[
  {
    "id": "key-abc123",
    "service": "google",
    "plain_key": "AIzaSy...",
    "daily_limit": 1000,
    "today_usage": 42,
    "today_attempts": 45
  }
]
```

> **Sécurité :** Retourne les clés en texte clair. Seules les clés des services autorisés par le paramètre `allowed_services` du client sont retournées.

---

#### `GET /api/services`

Liste des services pour le proxy. Retourne un tableau d'ID de services où `proxy_enabled=true`.

**Exemple de réponse :**
```json
["google", "ollama"]
```

Si le tableau est vide, le proxy utilise tous les services sans restriction.

---

#### `POST /api/heartbeat`

Envoyer le statut du proxy (exécuté automatiquement toutes les 20 secondes).

**Corps de la requête :**
```json
{
  "client_id": "bot-a",
  "version": "v0.1.6.20260314.231308",
  "service": "google",
  "model": "gemini-2.5-flash",
  "sse_connected": true,
  "host": "bot-a-host",
  "avatar": "data:image/png;base64,...",
  "key_usage":     {"key-abc123": 42, "key-def456": 0},
  "key_attempts":  {"key-abc123": 45, "key-def456": 3},
  "key_cooldowns": {"key-abc123": "2026-03-15T14:30:00Z"}
}
```

| Champ | Type | Description |
|-------|------|-------------|
| `client_id` | string | ID du client |
| `version` | string | Version du proxy (inclut l'horodatage de build, ex. `v0.1.6.20260314.231308`) |
| `service` | string | Service actuel |
| `model` | string | Modèle actuel |
| `sse_connected` | bool | Si SSE est connecté |
| `host` | string | Nom d'hôte |
| `avatar` | string | base64 data URI of the proxy's local avatar image (sent by proxy, auto-persisted to client record by vault). Set via `WV_AVATAR` env var (relative path under `~/.openclaw/`). |
| `key_usage` | map[string]int | Key ID → successful tokens used today. Vault calls `SetKeyUsage()` (absolute value sync). 429/402/582 errors do **not** increment this. |
| `key_attempts` | map[string]int | Key ID → total API calls today (success + rate-limited). Vault calls `SetKeyAttempts()`. 429/402/582 errors only increment this, not `key_usage`. |
| `key_cooldowns` | map[string]string | Key ID → cooldown end time (RFC3339). Vault calls `SetKeyCooldownIfLater()`. |

**Réponse :**
```json
{"status": "ok"}
```

---

### API Admin — Clés API

Nécessite l'en-tête `Authorization: Bearer <admin_token>`.

#### `GET /admin/keys`

Lister toutes les clés API enregistrées (clés en texte clair exclues).

**Exemple de réponse :**
```json
[
  {
    "id": "key-abc123",
    "service": "google",
    "label": "메인 키",
    "today_usage": 42,
    "today_attempts": 45,
    "daily_limit": 1000,
    "cooldown_until": "0001-01-01T00:00:00Z",
    "last_error": 0,
    "created_at": "2026-03-13T12:00:00Z",
    "available": true,
    "usage_pct": 4
  }
]
```

| Champ | Type | Description |
|-------|------|-------------|
| `today_usage` | int | Jetons de requêtes réussies aujourd'hui (n'inclut pas les erreurs 429/402/582) |
| `today_attempts` | int | Total des appels API aujourd'hui (succès + limités en débit) |
| `available` | bool | Si disponible sans temps de refroidissement ni limite |
| `usage_pct` | int | Pourcentage d'utilisation de la limite quotidienne (`daily_limit=0` → 0) |
| `cooldown_until` | RFC3339 | Fin du temps de refroidissement (valeur zéro signifie aucun) |
| `last_error` | int | Dernier code d'erreur HTTP |

---

#### `POST /admin/keys`

Enregistrer une nouvelle clé API. Un événement SSE `key_added` est diffusé immédiatement lors de l'enregistrement.

**Corps de la requête :**
```json
{
  "service": "google",
  "key": "AIzaSy...",
  "label": "메인 키",
  "daily_limit": 1000
}
```

| Champ | Requis | Description |
|-------|--------|-------------|
| `service` | ✅ | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| personnalisé |
| `key` | ✅ | Clé API en texte clair |
| `label` | — | Étiquette d'identification |
| `daily_limit` | — | Limite d'utilisation quotidienne (0 = illimité) |

---

#### `DELETE /admin/keys/{id}`

Supprimer une clé API. Un événement SSE `key_deleted` est diffusé après la suppression.

**Réponse :**
```json
{"status": "deleted"}
```

---

#### `POST /admin/keys/reset`

Réinitialiser l'utilisation quotidienne de toutes les clés. L'événement SSE `usage_reset` est diffusé.

**Réponse :**
```json
{
  "status": "reset",
  "time": "2026-03-13T15:00:00Z"
}
```

---

### API Admin — Clients

#### `GET /admin/clients`

Lister tous les clients (jetons inclus).

---

#### `POST /admin/clients`

Enregistrer un nouveau client.

**Corps de la requête :**
```json
{
  "id": "my-bot",
  "name": "내 봇",
  "token": "my-secret-token",
  "default_service": "google",
  "default_model": "gemini-2.5-flash",
  "allowed_services": ["google", "openrouter"],
  "agent_type": "openclaw",
  "work_dir": "~/.openclaw",
  "description": "OpenClaw 에이전트",
  "ip_whitelist": ["10.0.0.1", "10.0.0.0/24"],
  "enabled": true
}
```

| Champ | Requis | Description |
|-------|--------|-------------|
| `id` | ✅ | ID client unique |
| `name` | — | Nom d'affichage |
| `token` | — | Jeton d'authentification (généré automatiquement si omis) |
| `default_service` | — | Service par défaut |
| `default_model` | — | Modèle par défaut |
| `allowed_services` | — | Liste des services autorisés (tableau vide = tous autorisés) |
| `agent_type` | — | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | — | Répertoire de travail de l'agent |
| `description` | — | Description de l'agent |
| `ip_whitelist` | — | Liste d'IP autorisées (tableau vide = toutes autorisées, CIDR supporté) |
| `enabled` | — | Si activé (par défaut `true`) |

---

#### `GET /admin/clients/{id}`

Obtenir un client spécifique (jeton inclus).

---

#### `PUT /admin/clients/{id}`

Mettre à jour les paramètres d'un client. **Diffusion SSE `config_change` → reflété sur le proxy en 1 à 3 secondes.**

**Corps de la requête (inclure uniquement les champs à modifier) :**
```json
{
  "default_service": "openrouter",
  "default_model": "anthropic/claude-opus-4-6",
  "enabled": true
}
```

**Réponse :**
```json
{"status": "updated"}
```

---

#### `DELETE /admin/clients/{id}`

Supprimer un client.

---

### API Admin — Services

#### `GET /admin/services`

Lister les services enregistrés.

**Exemple de réponse :**
```json
[
  {"id": "google",      "name": "Google Gemini",   "enabled": true,  "custom": false},
  {"id": "openai",      "name": "OpenAI",          "enabled": true,  "custom": false},
  {"id": "anthropic",   "name": "Anthropic",       "enabled": false, "custom": false},
  {"id": "openrouter",  "name": "OpenRouter",      "enabled": true,  "custom": false},
  {"id": "ollama",      "name": "Ollama (Local)",  "enabled": true,  "custom": false,
   "local_url": "http://localhost:11434"},
  {"id": "lmstudio",    "name": "LM Studio",       "enabled": false, "custom": false},
  {"id": "vllm",        "name": "vLLM",            "enabled": false, "custom": false},
  {"id": "github-copilot","name":"GitHub Copilot", "enabled": false, "custom": false}
]
```

8 services intégrés : `google`, `openai`, `anthropic`, `openrouter`, `github-copilot`, `ollama`, `lmstudio`, `vllm`

---

#### `POST /admin/services`

Ajouter un service personnalisé. L'événement SSE `service_changed` est diffusé après l'ajout → **les menus déroulants du tableau de bord sont mis à jour immédiatement**.

**Corps de la requête :**
```json
{
  "id": "my-llm",
  "name": "사내 LLM 서버",
  "local_url": "http://10.0.0.50:8080",
  "enabled": true
}
```

---

#### `PUT /admin/services/{id}`

Mettre à jour les paramètres d'un service. L'événement SSE `service_changed` est diffusé après les modifications.

**Corps de la requête :**
```json
{
  "local_url": "http://192.168.x.x:11434",
  "enabled": true
}
```

---

#### `DELETE /admin/services/{id}`

Supprimer un service personnalisé. L'événement SSE `service_changed` est diffusé après la suppression.

Tentative de suppression d'un service intégré (`custom: false`) :
```json
{"error": "기본 서비스는 삭제할 수 없습니다: google"}
```

---

### API Admin — Liste des modèles

#### `GET /admin/models`

Lister les modèles par service. Utilise un cache TTL (10 minutes).

**Paramètres de requête :**

| Paramètre | Description | Exemple |
|-----------|-------------|---------|
| `service` | Filtre par service | `?service=google` |
| `q` | Recherche de modèle | `?q=gemini` |

**Récupération des modèles par service :**

| Service | Méthode | Nombre |
|---------|---------|--------|
| `google` | Liste statique | 8 (embedding inclus) |
| `openai` | Liste statique | 9 |
| `anthropic` | Liste statique | 6 |
| `github-copilot` | Liste statique | 6 |
| `openrouter` | Requête API dynamique (repli sur 14 modèles sélectionnés en cas d'échec) | 340+ |
| `ollama` | Requête dynamique au serveur local (7 recommandés si pas de réponse) | Variable |
| `lmstudio` | Requête dynamique au serveur local | Variable |
| `vllm` | Requête dynamique au serveur local | Variable |
| Personnalisé | Compatible OpenAI `/v1/models` | Variable |

**Modèles de repli OpenRouter (lorsque l'API ne répond pas) :**

| Modèle | Remarques |
|--------|-----------|
| `openrouter/hunter-alpha` | Gratuit, 1M contexte |
| `openrouter/healer-alpha` | Gratuit, omni-modal |
| `moonshot/kimi-k2.5` | 256K contexte |
| `z-ai/glm-5`, `z-ai/glm-4.7-flash` | — |
| `deepseek/deepseek-r1`, `deepseek/deepseek-chat` | — |
| `qwen/qwen-2.5-72b-instruct` | 131K contexte |
| `minimax/minimax-m2.5` | — |
| `meta-llama/llama-3.3-70b-instruct` | 131K contexte |

---

### API Admin — Statut des proxys

#### `GET /admin/proxies`

Dernier statut de heartbeat de tous les proxys connectés.

---

## Types d'événements SSE

Événements reçus du flux `/api/events` du coffre-fort :

| `type` | Déclencheur | Contenu `data` | Réaction du tableau de bord |
|--------|-------------|----------------|----------------------------|
| `connected` | Immédiatement à la connexion SSE | `{"clients": N}` | — |
| `config_change` | Paramètres client modifiés | `{"client_id","service","model"}` | Menu déroulant de modèle de la carte agent actualisé |
| `key_added` | Nouvelle clé API enregistrée | `{"service": "google"}` | Menu déroulant de modèle actualisé |
| `key_deleted` | Clé API supprimée | `{"service": "google"}` | Menu déroulant de modèle actualisé |
| `service_changed` | Service ajouté/modifié/supprimé | `{"action":"added"\|"updated"\|"deleted","id":"...","proxy_services":["google","ollama"]}` | Select de service + menu déroulant de modèle actualisé immédiatement ; liste des services de dispatch du proxy mise à jour en temps réel |
| `usage_update` | Lors du heartbeat du proxy (toutes les 20s) | `{"keys":[{"id","service","today_usage","today_attempts","daily_limit","cooldown_until"},...]}` | Barres et chiffres d'utilisation des clés mis à jour instantanément, compte à rebours du temps de refroidissement lancé. Données SSE utilisées directement sans fetch. Barres avec mise à l'échelle proportionnelle au total (pour les clés illimitées). |
| `usage_reset` | Réinitialisation de l'utilisation quotidienne | `{"time": "RFC3339"}` | Rafraîchissement de la page |

**Traitement des événements côté proxy :**

```
config_change reçu
  → Si client_id correspond à son propre ID
    → service, model mis à jour immédiatement
    → hooksMgr.Fire(EventModelChanged)
```

---

## Routage fournisseur/modèle

Lorsqu'un format `provider/model` est spécifié dans le champ `model` de `/v1/chat/completions`, le routage automatique est appliqué (compatible OpenClaw 3.11).

### Règles de routage par préfixe

| Préfixe | Cible de routage | Exemple |
|---------|-----------------|---------|
| `google/` | Google direct | `google/gemini-2.5-pro` |
| `openai/` | OpenAI direct | `openai/gpt-4o` |
| `anthropic/` | Via OpenRouter | `anthropic/claude-opus-4-6` |
| `ollama/` | Ollama direct | `ollama/qwen3.5:35b` |
| `custom/` | Re-parsing récursif (suppression de `custom/` puis re-routage) | `custom/google/gemini-2.5-flash` → Google |
| `openrouter/` | OpenRouter (chemin brut conservé) | `openrouter/meta-llama/llama-3.3-70b` |
| `opencode/`, `opencode-go/`, `opencode-zen/` | OpenRouter | `opencode-go/model` |
| `moonshot/`, `kimi-coding/` | OpenRouter (chemin complet conservé) | `moonshot/kimi-k2.5` |
| `groq/`, `mistral/`, `minimax/`, `cohere/`, `perplexity/` | OpenRouter | `groq/llama-3.1-70b` |
| `together/`, `huggingface/`, `nvidia/`, `venice/` | OpenRouter | — |
| `meta-llama/`, `qwen/`, `deepseek/`, `01-ai/` | OpenRouter (chemin complet) | `deepseek/deepseek-r1` |

### Détection automatique du préfixe `wall-vault/`

Le préfixe propre à wall-vault détermine automatiquement le service à partir de l'ID du modèle.

| Motif d'ID de modèle | Routage |
|----------------------|---------|
| `wall-vault/gemini-*` | Google |
| `wall-vault/gpt-*`, `wall-vault/o1`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI |
| `wall-vault/claude-*` | OpenRouter (chemin Anthropic) |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (gratuit, 1M ctx) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*`, `wall-vault/qwen*` | OpenRouter |
| Autres | OpenRouter |

### Gestion du suffixe `:cloud`

Le suffixe `:cloud` au format de tag Ollama est automatiquement supprimé et routé vers OpenRouter.

```
kimi-k2.5:cloud  →  OpenRouter, ID de modèle : kimi-k2.5
glm-5:cloud      →  OpenRouter, ID de modèle : glm-5
```

### Exemple d'intégration OpenClaw openclaw.json

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "https://localhost:56244/v1",
        apiKey: "YOUR_AGENT_TOKEN",
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/hunter-alpha" },
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  },
  agents: {
    defaults: {
      model: {
        primary: "wall-vault/gemini-2.5-flash",
        fallbacks: ["wall-vault/hunter-alpha"]
      }
    }
  }
}
```

Cliquez sur le **bouton 🐾** d'une carte agent pour copier automatiquement l'extrait de configuration pour cet agent dans le presse-papiers.

---

## Schéma de données

### APIKey

| Champ | Type | Description |
|-------|------|-------------|
| `id` | string | ID unique au format UUID |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| personnalisé |
| `encrypted_key` | string | Clé chiffrée AES-GCM (Base64) |
| `label` | string | Étiquette d'identification |
| `today_usage` | int | Jetons de requêtes réussies aujourd'hui (n'inclut pas les erreurs 429/402/582) |
| `today_attempts` | int | Total des appels API aujourd'hui (succès + limités en débit ; réinitialisé à minuit) |
| `daily_limit` | int | Limite quotidienne (0 = illimité) |
| `cooldown_until` | time.Time | Fin du temps de refroidissement |
| `last_error` | int | Dernier code d'erreur HTTP |
| `created_at` | time.Time | Date d'enregistrement |

**Politique de temps de refroidissement :**

| Erreur HTTP | Temps de refroidissement |
|-------------|-------------------------|
| 429 (Too Many Requests) | 30 minutes |
| 402 (Payment Required) | 24 heures |
| 400 / 401 / 403 | 24 heures |
| 582 (Gateway Overload) | 5 minutes |
| Erreur réseau | 10 minutes |

> **429/402/582** : Le temps de refroidissement est défini + `today_attempts` est incrémenté. `today_usage` reste inchangé (seuls les jetons réussis sont comptés).
> **Ollama (service local)** : `callOllama` utilise un client HTTP dédié avec `Timeout: 0` (illimité). L'inférence de grands modèles peut prendre de dizaines de secondes à plusieurs minutes, le délai d'expiration par défaut de 60 secondes n'est donc pas appliqué.

### Client

| Champ | Type | Description |
|-------|------|-------------|
| `id` | string | ID client unique |
| `name` | string | Nom d'affichage |
| `token` | string | Jeton d'authentification |
| `default_service` | string | Service par défaut |
| `default_model` | string | Modèle par défaut (peut être au format `provider/model`) |
| `allowed_services` | []string | Services autorisés (tableau vide = tous) |
| `agent_type` | string | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | string | Répertoire de travail de l'agent |
| `description` | string | Description |
| `ip_whitelist` | []string | Liste d'IP autorisées (CIDR supporté) |
| `avatar` | string | Agent avatar — relative path under `~/.openclaw/` (e.g. `workspace/avatar.png`, `workspace/avatars/profile.hpg`) OR base64 data URI (`data:image/png;base64,...`). Supported extensions: `.png`, `.jpg`/`.jpeg`/`.hpg`, `.webp`, `.gif`. |
| `enabled` | bool | Si `false`, retourne `403` lors de l'accès à `/api/keys` |
| `created_at` | time.Time | Date d'enregistrement |

### ServiceConfig

| Champ | Type | Description |
|-------|------|-------------|
| `id` | string | ID unique du service |
| `name` | string | Nom d'affichage |
| `local_url` | string | URL du serveur local (Ollama/LMStudio/vLLM/personnalisé) |
| `enabled` | bool | Si activé |
| `custom` | bool | S'il s'agit d'un service ajouté par l'utilisateur |
| `proxy_enabled` | bool | Whether this service is included in proxy dispatch and agent model dropdowns. Controlled by the "프록시 사용" checkbox in the Services card. Only `proxy_enabled: true` services appear in agent service/model selectors. |

### ProxyStatus (Heartbeat)

| Champ | Type | Description |
|-------|------|-------------|
| `client_id` | string | ID du client |
| `version` | string | Version du proxy (ex. `v0.1.6.20260314.231308`) |
| `service` | string | Service actuel |
| `model` | string | Modèle actuel |
| `sse_connected` | bool | Si SSE est connecté |
| `host` | string | Nom d'hôte |
| `avatar` | string | base64 data URI of local avatar (auto-synced from proxy via heartbeat) |
| `updated_at` | time.Time | Dernière mise à jour |
| `vault.today_usage` | int | Utilisation de jetons aujourd'hui |
| `vault.daily_limit` | int | Limite quotidienne |
| `vault.key_status` | string | `active` \| `cooldown` \| `exhausted` |

---

## Réponses d'erreur

```json
{"error": "오류 메시지"}
```

| Code | Signification |
|------|---------------|
| 200 | Succès |
| 400 | Requête invalide |
| 401 | Échec d'authentification |
| 403 | Accès refusé (client désactivé, IP bloquée) |
| 404 | Ressource non trouvée |
| 405 | Méthode non autorisée |
| 429 | Limite de débit dépassée |
| 500 | Erreur interne du serveur |
| 502 | Erreur API en amont (tous les replis ont échoué) |

---

## Exemples cURL

```bash
# ─── Proxy ────────────────────────────────────────────────────────────────────

# Vérification de santé
curl https://localhost:56244/health

# Statut
curl https://localhost:56244/status

# Liste des modèles (tous)
curl https://localhost:56244/api/models

# Modèles Google uniquement
curl "https://localhost:56244/api/models?service=google"

# Rechercher des modèles gratuits
curl "https://localhost:56244/api/models?q=alpha"

# Changer de modèle (local)
curl -X PUT https://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service":"google","model":"gemini-2.5-pro"}'

# Recharger la configuration
curl -X POST https://localhost:56244/reload

# Appel API Gemini direct
curl -X POST "https://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent" \
  -H "Content-Type: application/json" \
  -d '{"contents":[{"role":"user","parts":[{"text":"안녕하세요"}]}]}'

# Compatible OpenAI (modèle par défaut)
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# Format OpenClaw provider/model
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# Modèle gratuit 1M contexte
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/hunter-alpha","messages":[{"role":"user","content":"안녕"}]}'

# ─── Coffre-fort de clés (Public) ─────────────────────────────────────────────

curl https://localhost:56243/api/status
curl https://localhost:56243/api/clients
curl -s https://localhost:56243/api/events --max-time 3

# ─── Coffre-fort de clés (Admin) ──────────────────────────────────────────────

ADMIN="Authorization: Bearer admin-token"

# Liste des clés
curl -H "$ADMIN" https://localhost:56243/admin/keys

# Ajouter une clé Google
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"google","key":"AIzaSy...","label":"메인 키","daily_limit":1000}'

# Ajouter une clé OpenAI
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openai","key":"sk-...","label":"GPT 키"}'

# Ajouter une clé OpenRouter
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openrouter","key":"sk-or-v1-...","label":"OR 키"}'

# Supprimer une clé (diffusion SSE key_deleted)
curl -X DELETE https://localhost:56243/admin/keys/key-abc123 -H "$ADMIN"

# Réinitialiser l'utilisation quotidienne
curl -X POST https://localhost:56243/admin/keys/reset -H "$ADMIN"

# Liste des clients
curl -H "$ADMIN" https://localhost:56243/admin/clients

# Ajouter un client (OpenClaw)
curl -X POST https://localhost:56243/admin/clients \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"bot-a","name":"봇 A","agent_type":"openclaw","work_dir":"~/.openclaw","default_service":"google","default_model":"gemini-2.5-flash"}'

# Changer le modèle d'un client (mise à jour SSE instantanée)
curl -X PUT https://localhost:56243/admin/clients/bot-a \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"default_service":"openrouter","default_model":"wall-vault/hunter-alpha"}'

# Désactiver un client
curl -X PUT https://localhost:56243/admin/clients/my-bot \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":false}'

# Supprimer un client
curl -X DELETE https://localhost:56243/admin/clients/my-bot -H "$ADMIN"

# Liste des services
curl -H "$ADMIN" https://localhost:56243/admin/services

# Définir l'URL locale Ollama (diffusion SSE service_changed)
curl -X PUT https://localhost:56243/admin/services/ollama \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"local_url":"http://192.168.x.x:11434","enabled":true}'

# Activer le service OpenAI
curl -X PUT https://localhost:56243/admin/services/openai \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":true}'

# Ajouter un service personnalisé (diffusion SSE service_changed)
curl -X POST https://localhost:56243/admin/services \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"my-llm","name":"사내 LLM","local_url":"http://10.0.0.50:8080","enabled":true}'

# Supprimer un service personnalisé
curl -X DELETE https://localhost:56243/admin/services/my-llm -H "$ADMIN"

# Liste des modèles
curl -H "$ADMIN" https://localhost:56243/admin/models
curl -H "$ADMIN" "https://localhost:56243/admin/models?service=openrouter"
curl -H "$ADMIN" "https://localhost:56243/admin/models?q=hunter"

# Statut des proxys (heartbeat)
curl -H "$ADMIN" https://localhost:56243/admin/proxies

# ─── Mode distribué — Proxy → Coffre-fort ────────────────────────────────────

# Obtenir les clés déchiffrées
curl https://localhost:56243/api/keys \
  -H "Authorization: Bearer your-bot-a-token"

# Envoyer un heartbeat
curl -X POST https://localhost:56243/api/heartbeat \
  -H "Authorization: Bearer your-bot-a-token" \
  -H "Content-Type: application/json" \
  -d '{"client_id":"bot-a","version":"v0.1.6.20260314.231308","service":"google","model":"gemini-2.5-flash","sse_connected":true}'
```

---

## Middleware

Appliqué automatiquement à toutes les requêtes :

| Middleware | Fonction |
|-----------|----------|
| **Logger** | Journalisation au format `[method] path status latencyms` |
| **CORS** | `Access-Control-Allow-Origin: *` |
| **Recovery** | Récupération après panique, retourne une réponse 500 |

---

*Dernière mise à jour : 16/03/2026 — v0.1.7 : dashboard title renamed to "벽금고(wall-vault) 대시보드", logo moved to sticky topbar, today_attempts tracking, HTTP 582 cooldown (5min), share-of-total bar scaling, custom/ routing fix, Ollama Timeout:0, key_att i18n, avatar heartbeat sync, build timestamp versioning, proxy-only service filter*
