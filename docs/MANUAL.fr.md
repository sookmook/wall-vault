# Manuel Utilisateur wall-vault
*(Last updated: 2026-04-16 — v0.2.2)*

---

## Table des matières

1. [Qu'est-ce que wall-vault ?](#quest-ce-que-wall-vault)
2. [Installation](#installation)
3. [Premiers pas (Assistant de configuration)](#premiers-pas)
4. [Enregistrement des clés API](#enregistrement-des-clés-api)
5. [Utilisation du proxy](#utilisation-du-proxy)
6. [Tableau de bord du coffre-fort](#tableau-de-bord-du-coffre-fort)
7. [Mode distribué (Multi-Bot)](#mode-distribué-multi-bot)
8. [Configuration du démarrage automatique](#configuration-du-démarrage-automatique)
9. [Doctor (Diagnostic)](#doctor-diagnostic)
10. [RTK Économie de tokens](#rtk-économie-de-tokens)
11. [Référence des variables d'environnement](#référence-des-variables-denvironnement)
12. [Dépannage](#dépannage)

---

## Qu'est-ce que wall-vault ?

**wall-vault = Proxy IA + Coffre-fort de clés API pour OpenClaw**

Pour utiliser les services d'IA, vous avez besoin de **clés API**. Une clé API est comme un **badge d'accès numérique** qui prouve que « cette personne est autorisée à utiliser ce service ». Cependant, ces badges ont des limites d'utilisation quotidiennes, et il y a toujours un risque d'exposition en cas de mauvaise gestion.

wall-vault stocke ces badges dans un coffre-fort sécurisé et agit comme un **proxy (intermédiaire)** entre OpenClaw et les services d'IA. En termes simples, OpenClaw n'a qu'à se connecter à wall-vault, et wall-vault s'occupe du reste.

Problèmes résolus par wall-vault :

- **Rotation automatique des clés API** : Lorsqu'une clé atteint sa limite d'utilisation ou est temporairement bloquée (cooldown), le système bascule silencieusement vers la clé suivante. OpenClaw continue de fonctionner sans interruption.
- **Fallback automatique de service** : Si Google ne répond pas, le système bascule vers OpenRouter. Si cela échoue aussi, il bascule automatiquement vers l'IA installée localement (Ollama, LM Studio, vLLM). Votre session n'est jamais interrompue. Lorsque le service d'origine est rétabli, il revient automatiquement à la requête suivante (v0.1.18+, LM Studio/vLLM : v0.1.21+).
- **Synchronisation en temps réel (SSE)** : Lorsque vous changez le modèle dans le tableau de bord du coffre-fort, le changement est reflété dans l'écran OpenClaw en 1 à 3 secondes. SSE (Server-Sent Events) est une technologie où le serveur pousse les changements aux clients en temps réel.
- **Notifications en temps réel** : Les événements comme l'épuisement des clés ou les pannes de service sont immédiatement affichés en bas du TUI OpenClaw (écran terminal).

> :bulb: **Claude Code, Cursor et VS Code** peuvent également être connectés, mais le but principal de wall-vault est d'être utilisé avec OpenClaw.

```
OpenClaw (écran terminal TUI)
        |
        v
  wall-vault proxy (:56244)   <- Gestion des clés, routage, fallback, événements
        |
        +-- Google Gemini API
        +-- OpenRouter API (340+ modèles)
        +-- Ollama / LM Studio / vLLM (machine locale, dernier recours)
        +-- OpenAI / Anthropic API
```

---

## Notes de mise à niveau v0.2

- `Service` a acquis `default_model` et `allowed_models`. Le modèle par défaut pour chaque service est maintenant défini directement sur la fiche de service.
- `Client.default_service` / `default_model` ont été renommés et réinterprétés comme `preferred_service` / `model_override`. Si l'override est vide, le modèle par défaut du service est utilisé.
- Au premier démarrage en v0.2, le fichier `vault.json` existant est auto-migré, et l'état pré-migration est préservé sous `vault.json.pre-v02.{timestamp}.bak`.
- Le tableau de bord a été restructuré en trois zones : une barre latérale gauche, une grille de cartes centrales, et un volet d'édition latéral droit.
- Les chemins de l'API Admin sont inchangés, mais les schémas de corps de requête/réponse ont été mis à jour — les anciens scripts CLI devront être mis à jour en conséquence.

---

## Nouveautés de la v0.2.1

- **Passerelle multimodale (OpenAI → Gemini)** : `/v1/chat/completions` accepte désormais six types de parties de contenu en plus de `text` — `input_audio`, `input_video`, `input_image`, `input_file` et `image_url` (URI de données et URL http(s) externes ≤ 5 Mo). Le proxy convertit chacune d'elles en `inlineData` de Gemini. Les clients compatibles OpenAI tels qu'EconoWorld peuvent diffuser directement des blobs audio / image / vidéo.
- **Type d'agent EconoWorld** : `POST /agent/apply` avec `agentType: "econoworld"` écrit les paramètres wall-vault dans le fichier `analyzer/ai_config.json` du projet. `workDir` accepte une liste de chemins candidats séparés par des virgules et convertit les chemins de lecteur Windows en chemins de montage WSL.
- **Grille de clés + CRUD du tableau de bord** : 11 clés s'affichent sous forme de cartes compactes avec un volet latéral + ajouter / ✕ supprimer.
- **Ajout de service + réorganisation par glisser-déposer** : la grille des services gagne un bouton + ajouter et une poignée de glissement (`⋮⋮`).
- **En-tête / pied de page / animations de thème / sélecteur de langue** restaurés. Les 7 thèmes (cherry/dark/light/ocean/gold/autumn/winter) jouent leur effet de particules sur une couche derrière les cartes mais au-dessus de l'arrière-plan.
- **UX de fermeture du volet latéral** : un clic à l'extérieur ou la touche Esc ferme le volet latéral.
- **Indicateur d'état SSE** dans le pied de page (vert = connecté, orange = reconnexion, gris = déconnecté).

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

## Installation

### Linux / macOS

Ouvrez un terminal et collez les commandes suivantes :

```bash
# Linux (PC standard, serveur — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Télécharge le fichier depuis Internet.
- `chmod +x` — Rend le fichier téléchargé « exécutable ». Si vous sautez cette étape, vous obtiendrez une erreur « permission refusée ».

### Windows

Ouvrez PowerShell (en tant qu'administrateur) et exécutez :

```powershell
# Téléchargement
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Ajouter au PATH (prend effet après redémarrage de PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> :bulb: **Qu'est-ce que le PATH ?** C'est une liste de dossiers où votre ordinateur cherche les commandes. Vous devez ajouter wall-vault au PATH pour pouvoir exécuter `wall-vault` depuis n'importe quel dossier.

### Compilation depuis les sources (pour développeurs)

Applicable uniquement si l'environnement de développement Go est installé.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (version : v0.1.25.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> :bulb: **Version avec horodatage de build** : Lors de la compilation avec `make build`, la version est automatiquement générée dans un format comme `v0.1.27.20260409` incluant la date et l'heure. Si vous compilez directement avec `go build ./...`, la version affichera uniquement `"dev"`.

---

## Premiers pas

### Exécution de l'assistant de configuration

Après l'installation, assurez-vous d'exécuter l'**assistant de configuration** avec la commande suivante. L'assistant vous guidera à travers les éléments nécessaires un par un.

```bash
wall-vault setup
```

L'assistant procède aux étapes suivantes :

```
1. Sélection de la langue (10 langues dont le français)
2. Sélection du thème (light / dark / gold / cherry / ocean)
3. Mode d'exploitation — standalone (machine unique) ou distribué (plusieurs machines)
4. Nom du bot — le nom affiché sur le tableau de bord
5. Configuration des ports — par défaut : proxy 56244, coffre-fort 56243 (Entrée pour garder les valeurs par défaut)
6. Sélection des services IA — choix parmi Google / OpenRouter / Ollama / LM Studio / vLLM
7. Paramètres du filtre de sécurité des outils
8. Token administrateur — un mot de passe qui verrouille les fonctions d'administration du tableau de bord. Génération automatique possible
9. Mot de passe de chiffrement des clés API — pour un stockage plus sécurisé des clés (optionnel)
10. Chemin de sauvegarde du fichier de configuration
```

> :warning: **Mémorisez votre token administrateur.** Vous en aurez besoin plus tard pour ajouter des clés ou modifier les paramètres dans le tableau de bord. Si vous le perdez, vous devrez modifier le fichier de configuration manuellement.

Une fois l'assistant terminé, un fichier de configuration `wall-vault.yaml` est automatiquement créé.

### Démarrage

```bash
wall-vault start
```

Les deux serveurs suivants démarrent simultanément :

- **Proxy** (`http://localhost:56244`) — L'intermédiaire connectant OpenClaw et les services IA
- **Coffre-fort** (`http://localhost:56243`) — Gestion des clés API et tableau de bord web

Ouvrez `http://localhost:56243` dans votre navigateur pour accéder au tableau de bord.

---

## Enregistrement des clés API

Il existe quatre méthodes pour enregistrer des clés API. **La méthode 1 (variables d'environnement) est recommandée pour les débutants.**

### Méthode 1 : Variables d'environnement (recommandée — la plus simple)

Les variables d'environnement sont des **valeurs prédéfinies** que les programmes lisent au démarrage. Tapez simplement ceci dans votre terminal :

```bash
# Enregistrer une clé Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Enregistrer une clé OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Démarrer après l'enregistrement
wall-vault start
```

Si vous avez plusieurs clés, séparez-les par des virgules. wall-vault les utilisera en rotation automatiquement (round robin) :

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> :bulb: **Astuce** : La commande `export` ne s'applique qu'à la session terminal en cours. Pour persister après un redémarrage, ajoutez les lignes à votre fichier `~/.bashrc` ou `~/.zshrc`.

### Méthode 2 : Interface du tableau de bord (pointer et cliquer)

1. Ouvrez `http://localhost:56243` dans votre navigateur
2. Cliquez sur le bouton `[+ Ajouter]` dans la carte **:key: Clés API** en haut
3. Entrez le type de service, la valeur de la clé, le libellé (nom mémo) et la limite quotidienne, puis enregistrez

### Méthode 3 : API REST (pour l'automatisation/scripts)

L'API REST est une méthode permettant aux programmes d'échanger des données via HTTP. Utile pour l'enregistrement automatisé par script.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer token-admin" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Clé principale",
    "daily_limit": 1000
  }'
```

### Méthode 4 : Flags du proxy (pour les tests rapides)

Utilisez ceci pour des tests temporaires sans enregistrement formel. Les clés disparaissent à la fermeture du programme.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Utilisation du proxy

### Utilisation avec OpenClaw (objectif principal)

Voici comment configurer OpenClaw pour se connecter aux services IA via wall-vault.

Ouvrez `~/.openclaw/openclaw.json` et ajoutez le contenu suivant :

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // token d'agent du coffre-fort
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // contexte 1M gratuit
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> :bulb: **Méthode plus facile** : Appuyez sur le bouton **:lobster: Copier la config OpenClaw** sur la carte d'agent du tableau de bord. Un snippet avec le token et l'adresse pré-remplis sera copié dans votre presse-papiers. Il suffit de le coller.

**Où le préfixe `wall-vault/` dans le nom du modèle redirige-t-il ?**

wall-vault détermine automatiquement quel service IA doit recevoir la requête en fonction du nom du modèle :

| Format du modèle | Service routé |
|-----------------|--------------|
| `wall-vault/gemini-*` | Directement vers Google Gemini |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | Directement vers OpenAI |
| `wall-vault/claude-*` | Anthropic via OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (contexte 1M tokens gratuit) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/nom-modèle`, `openai/nom-modèle`, `anthropic/nom-modèle`, etc. | Directement vers le service correspondant |
| `custom/google/nom-modèle`, `custom/openai/nom-modèle`, etc. | Supprime le préfixe `custom/` et redirige |
| `nom-modèle:cloud` | Supprime le suffixe `:cloud` et redirige vers OpenRouter |

> :bulb: **Qu'est-ce que le contexte ?** C'est la quantité de conversation qu'une IA peut retenir en une fois. 1M (un million de tokens) signifie qu'elle peut traiter de très longues conversations ou documents en une seule passe.

### Connexion directe via le format API Gemini (compatibilité avec les outils existants)

Si vous avez des outils qui utilisent directement l'API Google Gemini, changez simplement l'URL vers wall-vault :

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Ou si l'outil spécifie les URL directement :

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Utilisation avec le SDK OpenAI (Python)

Vous pouvez connecter wall-vault au code Python qui utilise l'IA. Changez simplement le `base_url` :

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # wall-vault gère les clés API pour vous
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # Utilisez le format provider/model
    messages=[{"role": "user", "content": "Bonjour"}]
)
```

### Changer de modèle en cours d'exécution

Pour changer le modèle IA pendant que wall-vault est déjà en cours d'exécution :

```bash
# Changer le modèle via une requête directe au proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# En mode distribué (multi-bot), changez sur le serveur coffre-fort -> synchronisé instantanément via SSE
curl -X PUT http://localhost:56243/admin/clients/mon-bot-id \
  -H "Authorization: Bearer token-admin" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Lister les modèles disponibles

```bash
# Voir la liste complète
curl http://localhost:56244/api/models | python3 -m json.tool

# Voir uniquement les modèles Google
curl "http://localhost:56244/api/models?service=google"

# Rechercher par nom (ex : modèles contenant "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Modèles principaux par service :**

| Service | Modèles principaux |
|---------|-------------------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ (Hunter Alpha 1M contexte gratuit, DeepSeek R1/V3, Qwen 2.5, etc.) |
| Ollama | Détection automatique depuis le serveur installé localement |
| LM Studio | Serveur local (port 1234) |
| vLLM | Serveur local (port 8000) |
| llama.cpp | Serveur local (port 8080) |

---

## Tableau de bord du coffre-fort

Accédez au tableau de bord en ouvrant `http://localhost:56243` dans votre navigateur.

**Disposition de l'écran :**
- **Barre supérieure (fixe)** : Logo, sélecteur de langue/thème, statut de connexion SSE
- **Grille de cartes** : Cartes d'agents, de services et de clés API disposées en tuiles

### Carte des clés API

Une carte pour gérer toutes les clés API enregistrées en un coup d'oeil.

- Affiche la liste des clés groupées par service.
- `today_usage` : Tokens traités avec succès aujourd'hui (caractères lus et écrits par l'IA)
- `today_attempts` : Total des appels aujourd'hui (succès + échecs combinés)
- Utilisez le bouton `[+ Ajouter]` pour enregistrer de nouvelles clés, et `x` pour les supprimer.

> :bulb: **Que sont les tokens ?** Les tokens sont les unités utilisées par l'IA pour traiter le texte. Environ un mot anglais, ou 1-2 caractères français. La tarification des API est généralement calculée sur le nombre de tokens.

### Carte d'agent

Une carte affichant le statut des bots (agents) connectés au proxy wall-vault.

**Le statut de connexion est affiché en 4 niveaux :**

| Indicateur | Statut | Signification |
|-----------|--------|---------------|
| :green_circle: | En cours d'exécution | Le proxy fonctionne normalement |
| :yellow_circle: | Retardé | Répond mais lentement |
| :red_circle: | Hors ligne | Le proxy ne répond pas |
| :black_circle: | Non connecté / Désactivé | Le proxy n'a jamais été connecté au coffre-fort ou est désactivé |

**Guide des boutons en bas de la carte d'agent :**

Lorsque vous enregistrez un agent et spécifiez le **type d'agent**, des boutons pratiques apparaissent automatiquement pour ce type.

---

#### :radio_button: Bouton Copier la configuration — Génère automatiquement les paramètres de connexion

En cliquant sur le bouton, un snippet de configuration avec le token de l'agent, l'adresse du proxy et les informations du modèle pré-remplis est copié dans votre presse-papiers. Collez simplement le contenu copié à l'emplacement indiqué dans le tableau ci-dessous.

| Bouton | Type d'agent | Emplacement de collage |
|--------|-------------|----------------------|
| :lobster: Copier config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| :crab: Copier config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| :orange_circle: Copier config Claude Code | `claude-code` | `~/.claude/settings.json` |
| :keyboard: Copier config Cursor | `cursor` | Cursor -> Settings -> AI |
| :computer: Copier config VSCode | `vscode` | `~/.continue/config.json` |

**Exemple — Pour le type Claude Code, voici ce qui est copié :**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "token-de-cet-agent"
}
```

**Exemple — Pour le type VSCode (Continue) :**

```yaml
# ~/.continue/config.yaml  <- coller dans config.yaml, PAS config.json
name: My Config
version: 0.0.1
schema: v1

models:
  - name: wall-vault proxy
    provider: openai
    model: gemini-2.5-flash
    apiBase: http://192.168.1.20:56244/v1
    apiKey: token-de-cet-agent
    roles:
      - chat
      - edit
      - apply
```

> :warning: **Les versions récentes de Continue utilisent `config.yaml`.** Si `config.yaml` existe, `config.json` est complètement ignoré. Assurez-vous de coller dans `config.yaml`.

**Exemple — Pour le type Cursor :**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : token-de-cet-agent

// Ou variables d'environnement :
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=token-de-cet-agent
```

> :warning: **Si la copie dans le presse-papiers ne fonctionne pas** : Les politiques de sécurité du navigateur peuvent bloquer la copie. Si une boîte de texte popup apparaît, sélectionnez tout avec Ctrl+A et copiez avec Ctrl+C.

---

#### :zap: Bouton d'application automatique — Un clic et c'est configuré

Pour les agents de type `cline`, `claude-code`, `openclaw` ou `nanoclaw`, un bouton **:zap: Appliquer la config** apparaît sur la carte d'agent. Cliquer sur ce bouton met automatiquement à jour le fichier de configuration local de l'agent.

| Bouton | Type d'agent | Fichier cible |
|--------|-------------|--------------|
| :zap: Appliquer config Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| :zap: Appliquer config Claude Code | `claude-code` | `~/.claude/settings.json` |
| :zap: Appliquer config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| :zap: Appliquer config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> :warning: Ce bouton envoie une requête à **localhost:56244** (proxy local). Le proxy doit être en cours d'exécution sur cette machine.

---

#### :twisted_rightwards_arrows: Tri des cartes par glisser-déposer (v0.1.17, amélioré v0.1.25)

Vous pouvez **glisser** les cartes d'agents sur le tableau de bord pour les réorganiser dans n'importe quel ordre.

1. Saisissez la zone **feu tricolore (●)** en haut à gauche d'une carte avec votre souris et glissez
2. Déposez-la sur la carte à la position souhaitée pour échanger leur ordre

> :bulb: Le corps de la carte (champs de saisie, boutons, etc.) ne peut pas être glissé. Vous ne pouvez saisir que depuis la zone du feu tricolore.

#### :orange_circle: Détection du processus agent (v0.1.25)

Lorsque le proxy fonctionne normalement mais qu'un processus d'agent local (NanoClaw, OpenClaw) est mort, le feu tricolore de la carte passe en **orange (clignotant)** et affiche un message « Processus agent arrêté ».

- :green_circle: Vert : Proxy + agent normal
- :orange_circle: Orange (clignotant) : Proxy normal, agent mort
- :red_circle: Rouge : Proxy hors ligne
3. L'ordre modifié est **sauvegardé immédiatement sur le serveur** et persiste après actualisation de la page

> :bulb: Les appareils tactiles (mobile/tablette) ne sont pas encore supportés. Utilisez un navigateur de bureau.

---

#### :arrows_counterclockwise: Synchronisation bidirectionnelle des modèles (v0.1.16)

Lorsque vous changez le modèle d'un agent dans le tableau de bord du coffre-fort, la configuration locale de l'agent est automatiquement mise à jour.

**Pour Cline :**
- Changement de modèle dans le coffre-fort -> événement SSE -> le proxy met à jour les champs de modèle dans `globalState.json`
- Champs mis à jour : `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` et la clé API ne sont pas modifiés
- **Rechargement de VS Code requis (`Ctrl+Alt+R` ou `Ctrl+Shift+P` -> `Developer: Reload Window`)**
  - Parce que Cline ne relit pas les fichiers de configuration en cours d'exécution

**Pour Claude Code :**
- Changement de modèle dans le coffre-fort -> événement SSE -> le proxy met à jour le champ `model` dans `settings.json`
- Recherche automatique dans les chemins WSL et Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Direction inverse (agent -> coffre-fort) :**
- Lorsque les agents (Cline, Claude Code, etc.) envoient des requêtes au proxy, celui-ci inclut les informations service/modèle du client dans le heartbeat
- La carte d'agent du tableau de bord affiche le service/modèle actuellement utilisé en temps réel

> :bulb: **Point clé** : Le proxy identifie les agents par le token Authorization dans les requêtes et redirige automatiquement vers le service/modèle configuré dans le coffre-fort. Même si Cline ou Claude Code envoie un nom de modèle différent, le proxy le remplace par le paramètre du coffre-fort.

---

### Utiliser Cline dans VS Code — Guide détaillé

#### Étape 1 : Installer Cline

Installez **Cline** (ID : `saoudrizwan.claude-dev`) depuis le Marketplace d'extensions VS Code.

#### Étape 2 : Enregistrer l'agent dans le coffre-fort

1. Ouvrez le tableau de bord du coffre-fort (`http://IP-coffre-fort:56243`)
2. Cliquez sur **+ Ajouter** dans la section **Agents**
3. Remplissez les champs suivants :

| Champ | Valeur | Description |
|-------|--------|-------------|
| ID | `mon_cline` | Identifiant unique (alphanumérique, sans espaces) |
| Nom | `Mon Cline` | Nom affiché sur le tableau de bord |
| Type d'agent | `cline` | <- Doit sélectionner `cline` |
| Service | Choisir le service (ex : `google`) | |
| Modèle | Entrer le modèle (ex : `gemini-2.5-flash`) | |

4. Cliquez sur **Enregistrer** pour générer automatiquement un token

#### Étape 3 : Connecter Cline

**Méthode A — Application automatique (recommandée)**

1. Vérifiez que le **proxy** wall-vault est en cours d'exécution sur cette machine (`localhost:56244`)
2. Cliquez sur le bouton **:zap: Appliquer config Cline** sur la carte d'agent du tableau de bord
3. Succès lorsque vous voyez la notification « Configuration appliquée ! »
4. Rechargez VS Code (`Ctrl+Alt+R`)

**Méthode B — Configuration manuelle**

Ouvrez les paramètres (:gear:) dans la barre latérale de Cline :
- **API Provider** : `OpenAI Compatible`
- **Base URL** : `http://adresse-proxy:56244/v1`
  - Même machine : `http://localhost:56244/v1`
  - Machine différente (ex : Mac Mini) : `http://192.168.1.20:56244/v1`
- **API Key** : Token émis par le coffre-fort (copier depuis la carte d'agent)
- **Model ID** : Modèle configuré dans le coffre-fort (ex : `gemini-2.5-flash`)

#### Étape 4 : Vérification

Envoyez n'importe quel message dans la fenêtre de chat Cline. Si tout fonctionne correctement :
- La carte d'agent correspondante sur le tableau de bord affiche un **point vert (En cours d'exécution)**
- La carte affiche le service/modèle actuel (ex : `google / gemini-2.5-flash`)

#### Changer de modèle

Pour changer le modèle de Cline, faites-le depuis le **tableau de bord du coffre-fort** :

1. Changez le menu déroulant service/modèle sur la carte d'agent
2. Cliquez sur **Appliquer**
3. Rechargez VS Code (`Ctrl+Alt+R`) — Le nom du modèle dans le pied de page de Cline se met à jour
4. Le nouveau modèle est utilisé à partir de la requête suivante

> :bulb: En pratique, le proxy identifie les requêtes de Cline par le token et les redirige vers le modèle configuré dans le coffre-fort. Même sans rechargement de VS Code, **le modèle réellement utilisé change immédiatement** — le rechargement sert uniquement à mettre à jour l'affichage du modèle dans l'interface Cline.

#### Détection de déconnexion

Lorsque vous fermez VS Code, la carte d'agent sur le tableau de bord passe en jaune (retardé) après environ **90 secondes**, puis en rouge (hors ligne) après **3 minutes**. (Depuis v0.1.18, des vérifications de statut toutes les 15 secondes ont accéléré la détection hors ligne.)

#### Dépannage

| Symptôme | Cause | Solution |
|----------|-------|----------|
| Erreur « Connexion échouée » dans Cline | Proxy non démarré ou mauvaise adresse | Vérifier le proxy avec `curl http://localhost:56244/health` |
| Le point vert n'apparaît pas dans le coffre-fort | Clé API (token) non configurée | Cliquer à nouveau sur **:zap: Appliquer config Cline** |
| Le modèle dans le pied de page Cline ne change pas | Cline met en cache les paramètres | Recharger VS Code (`Ctrl+Alt+R`) |
| Mauvais nom de modèle affiché | Ancien bug (corrigé dans v0.1.16) | Mettre à jour le proxy vers v0.1.16 ou ultérieur |

---

#### :purple_circle: Bouton Copier la commande de déploiement — Pour installer sur de nouvelles machines

Utilisez ceci lors de la première installation du proxy wall-vault sur un nouvel ordinateur et de sa connexion au coffre-fort. En cliquant sur le bouton, le script d'installation complet est copié. Collez-le dans le terminal du nouvel ordinateur et exécutez-le — tout est géré en une seule fois :

1. Installation du binaire wall-vault (ignoré si déjà installé)
2. Enregistrement automatique du service utilisateur systemd
3. Démarrage du service et connexion automatique au coffre-fort

> :bulb: Le script contient déjà le token de cet agent et l'adresse du serveur coffre-fort, vous pouvez donc l'exécuter immédiatement après le collage sans aucune modification.

---

### Carte de service

Une carte pour activer/désactiver et configurer les services IA.

- Interrupteur d'activation/désactivation par service
- Entrez l'adresse des serveurs IA locaux (Ollama, LM Studio, vLLM, llama.cpp, etc. sur votre machine) et les modèles disponibles sont automatiquement découverts.
- **Statut de connexion du service local** : Le point à côté du nom du service est **vert** si connecté, **gris** sinon
- **Feu tricolore automatique du service local** (v0.1.23+) : Les services locaux (Ollama, LM Studio, vLLM, llama.cpp) sont automatiquement activés/désactivés en fonction de la connectivité. Lorsqu'un service devient accessible, il passe au vert et la case s'active dans les 15 secondes ; lorsqu'il devient inaccessible, il se désactive automatiquement. Fonctionne de la même manière que les services cloud (Google, OpenRouter, etc.) qui basculent automatiquement en fonction de la disponibilité des clés API.
- **Bascule du mode raisonnement** (v0.2.17+) : Une case à cocher **mode raisonnement** apparaît en bas du formulaire d'édition des services locaux. Lorsqu'elle est activée, le proxy ajoute `"reasoning": true` au corps des chat-completions envoyé en amont, permettant aux modèles qui prennent en charge la sortie du processus de réflexion — comme DeepSeek R1 ou Qwen QwQ — de renvoyer également des blocs `<think>…</think>`. Les serveurs qui ne connaissent pas ce champ l'ignorent, vous pouvez donc la laisser activée en toute sécurité même pour des charges de travail mixtes.

> :bulb: **Si un service local s'exécute sur un autre ordinateur** : Entrez l'IP de cet ordinateur dans le champ URL du service. Exemple : `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio), `http://192.168.1.20:8080` (llama.cpp). Si le service est lié à `127.0.0.1` plutôt qu'à `0.0.0.0`, l'accès par IP externe ne fonctionnera pas — vérifiez l'adresse de liaison dans les paramètres du service.

### Saisie du token administrateur

Lorsque vous essayez d'utiliser des fonctions importantes comme l'ajout ou la suppression de clés dans le tableau de bord, une popup de saisie du token administrateur apparaît. Entrez le token défini lors de l'assistant de configuration. Une fois saisi, il persiste jusqu'à la fermeture du navigateur.

> :warning: **Si les échecs d'authentification dépassent 10 en 15 minutes, cette IP est temporairement bloquée.** Si vous avez oublié votre token, vérifiez le champ `admin_token` dans votre fichier `wall-vault.yaml`.

---

## Mode distribué (Multi-Bot)

Lorsque vous exécutez OpenClaw simultanément sur plusieurs ordinateurs, cette configuration **partage un seul coffre-fort de clés**. C'est pratique car vous n'avez besoin de gérer les clés qu'à un seul endroit.

### Exemple de configuration

```
[Serveur coffre-fort]
  wall-vault vault    (coffre-fort :56243, tableau de bord)

[WSL Alpha]           [Raspberry Pi Gamma]   [Mac Mini Local]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  <-> sync SSE          <-> sync SSE            <-> sync SSE
```

Tous les bots pointent vers le serveur coffre-fort central, donc les changements de modèle ou les ajouts de clés dans le coffre-fort sont instantanément reflétés sur tous les bots.

### Étape 1 : Démarrer le serveur coffre-fort

Exécutez ceci sur l'ordinateur qui servira de serveur coffre-fort :

```bash
wall-vault vault
```

### Étape 2 : Enregistrer chaque bot (client)

Pré-enregistrez les informations de chaque bot qui se connectera au serveur coffre-fort :

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer token-admin" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "Bot A",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### Étape 3 : Démarrer le proxy sur chaque machine bot

Sur chaque machine bot, démarrez le proxy avec l'adresse du serveur coffre-fort et le token :

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> :bulb: Remplacez **`192.168.x.x`** par l'adresse IP interne réelle de la machine serveur coffre-fort. Vous pouvez la trouver dans les paramètres du routeur ou via la commande `ip addr`.

---

## Configuration du démarrage automatique

S'il est fastidieux de démarrer manuellement wall-vault à chaque redémarrage, enregistrez-le comme service système. Une fois enregistré, il démarre automatiquement au démarrage.

### Linux — systemd (la plupart des distributions Linux)

systemd est le système qui démarre et gère automatiquement les programmes sous Linux :

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Consulter les journaux :

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

Le système responsable du démarrage automatique des programmes sous macOS :

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Téléchargez NSSM depuis [nssm.cc](https://nssm.cc/download) et ajoutez-le au PATH.
2. Dans un PowerShell administrateur :

```powershell
wall-vault doctor deploy windows
```

---

## Doctor (Diagnostic)

La commande `doctor` est un outil qui **auto-diagnostique et répare** les problèmes de configuration de wall-vault.

```bash
wall-vault doctor check   # Diagnostiquer l'état actuel (lecture seule, ne change rien)
wall-vault doctor fix     # Réparer automatiquement les problèmes
wall-vault doctor all     # Diagnostic + réparation automatique en une étape
```

> :bulb: Si quelque chose semble anormal, exécutez d'abord `wall-vault doctor all`. Il détecte et corrige automatiquement de nombreux problèmes.

---

## RTK Économie de tokens

*(v0.1.24+)*

**RTK (Token Reduction Kit)** compresse automatiquement la sortie des commandes shell exécutées par les agents de codage IA (comme Claude Code), réduisant l'utilisation de tokens. Par exemple, 15 lignes de sortie `git status` peuvent être réduites à un résumé de 2 lignes.

### Utilisation de base

```bash
# Enveloppez les commandes avec wall-vault rtk pour filtrer automatiquement la sortie
wall-vault rtk git status          # Affiche uniquement la liste des fichiers modifiés
wall-vault rtk git diff HEAD~1     # Lignes modifiées + contexte minimal uniquement
wall-vault rtk git log -10         # Hash + message d'une ligne chacun
wall-vault rtk go test ./...       # Affiche uniquement les tests échoués
wall-vault rtk ls -la              # Les commandes non supportées sont automatiquement tronquées
```

### Commandes supportées et économies

| Commande | Méthode de filtrage | Économie |
|----------|-------------------|----------|
| `git status` | Résumé des fichiers modifiés uniquement | ~87% |
| `git diff` | Lignes modifiées + 3 lignes de contexte | ~60-94% |
| `git log` | Hash + première ligne du message | ~90% |
| `git push/pull/fetch` | Progression supprimée, résumé uniquement | ~80% |
| `go test` | Échecs uniquement, réussites comptées | ~88-99% |
| `go build/vet` | Erreurs uniquement | ~90% |
| Toutes les autres commandes | 50 premières + 50 dernières lignes, max 32 Ko | Variable |

### Pipeline de filtrage en 3 étapes

1. **Filtre structurel spécifique à la commande** — Comprend le format de sortie de git, go, etc. et extrait les parties significatives
2. **Post-traitement par expressions régulières** — Supprime les codes couleur ANSI, réduit les lignes vides, agrège les lignes en double
3. **Passage direct + troncation** — Les commandes non supportées ne conservent que les 50 premières/dernières lignes

### Intégration Claude Code

Vous pouvez configurer le hook `PreToolUse` de Claude Code pour faire passer automatiquement toutes les commandes shell par RTK.

```bash
# Installer le hook (automatiquement ajouté à settings.json de Claude Code)
wall-vault rtk hook install
```

Ou ajouter manuellement à `~/.claude/settings.json` :

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

> :bulb: **Préservation du code de sortie** : RTK retourne le code de sortie de la commande originale sans modification. Si une commande échoue (code de sortie != 0), l'IA détecte précisément l'échec.

> :bulb: **Anglais forcé** : RTK exécute les commandes avec `LC_ALL=C`, garantissant une sortie en anglais indépendamment des paramètres de langue du système. Ceci est nécessaire pour que les filtres fonctionnent correctement.

---

## Référence des variables d'environnement

Les variables d'environnement sont un moyen de transmettre des valeurs de configuration aux programmes. Tapez `export VARIABLE=valeur` dans le terminal, ou placez-les dans les fichiers de service de démarrage automatique pour une application permanente.

| Variable | Description | Exemple |
|----------|-------------|---------|
| `WV_LANG` | Langue du tableau de bord | `ko`, `en`, `ja` |
| `WV_THEME` | Thème du tableau de bord | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Clé API Google (séparées par des virgules pour plusieurs) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Clé API OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Adresse du serveur coffre-fort en mode distribué | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Token d'authentification client (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Token administrateur | `admin-token-here` |
| `WV_MASTER_PASS` | Mot de passe de chiffrement des clés API | `my-password` |
| `WV_AVATAR` | Chemin du fichier image avatar (relatif à `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Adresse du serveur local Ollama | `http://192.168.x.x:11434` |

---

## Dépannage

### Le proxy ne démarre pas

Le port est probablement déjà utilisé par un autre programme.

```bash
ss -tlnp | grep 56244   # Vérifier ce qui utilise le port 56244
wall-vault proxy --port 8080   # Démarrer sur un port différent
```

### Erreurs de clé API (429, 402, 401, 403, 582)

| Code d'erreur | Signification | Résolution |
|--------------|---------------|------------|
| **429** | Trop de requêtes (quota dépassé) | Patienter ou ajouter plus de clés |
| **402** | Paiement requis ou crédits épuisés | Recharger les crédits sur le service |
| **401 / 403** | Clé invalide ou pas de permission | Vérifier la valeur de la clé et ré-enregistrer |
| **582** | Surcharge de la passerelle (cooldown de 5 minutes) | Se résout automatiquement après 5 minutes |

```bash
# Vérifier la liste des clés enregistrées et leur statut
curl -H "Authorization: Bearer token-admin" http://localhost:56243/admin/keys

# Réinitialiser les compteurs d'utilisation des clés
curl -X POST -H "Authorization: Bearer token-admin" http://localhost:56243/admin/keys/reset
```

### L'agent affiche « Non connecté »

« Non connecté » signifie que le processus proxy n'envoie pas de heartbeats au coffre-fort. **Cela ne signifie pas que les paramètres n'ont pas été sauvegardés.** Le proxy doit être en cours d'exécution avec l'adresse du serveur coffre-fort et le token pour entrer dans un état connecté.

```bash
# Démarrer le proxy avec l'adresse du serveur coffre-fort, le token et l'ID client
WV_VAULT_URL=http://serveur-coffre-fort:56243 \
WV_VAULT_TOKEN=token-client \
WV_VAULT_CLIENT_ID=id-client \
wall-vault proxy
```

Une fois connecté, le tableau de bord affiche :green_circle: En cours d'exécution dans les 20 secondes environ.

### Ollama ne se connecte pas

Ollama est un programme qui exécute l'IA directement sur votre ordinateur. Vérifiez d'abord si Ollama est en cours d'exécution.

```bash
curl http://localhost:11434/api/tags   # Si une liste de modèles apparaît, ça fonctionne
export OLLAMA_URL=http://192.168.x.x:11434   # Si exécuté sur un autre ordinateur
```

> :warning: Si Ollama ne répond pas, démarrez-le d'abord avec la commande `ollama serve`.

> :warning: **Les grands modèles sont lents** : Les grands modèles comme `qwen3.5:35b` ou `deepseek-r1` peuvent prendre plusieurs minutes pour générer une réponse. Même si rien ne semble se passer, le traitement est peut-être en cours — veuillez patienter.

---

## Changements récents (v0.1.16 ~ v0.1.27)

### v0.1.27 (2026-04-09)
- **Correction du nom de modèle en fallback Ollama** : Correction d'un problème où les noms de modèles préfixés par le fournisseur (ex : `google/gemini-3.1-pro-preview`) étaient transmis directement à Ollama lors du fallback depuis d'autres services. Maintenant automatiquement remplacé par la variable d'environnement/le modèle par défaut.
- **Durée de cooldown significativement réduite** : 429 limite de débit 30min->5min, 402 paiement 1h->30min, 401/403 24h->6h. Prévient la paralysie totale du proxy lorsque toutes les clés entrent en cooldown simultanément.
- **Nouvelle tentative forcée en cas de cooldown total** : Lorsque toutes les clés sont en cooldown, la clé la plus proche de l'expiration est retentée de force pour éviter le rejet des requêtes.
- **Correction de l'affichage de la liste des services** : La réponse `/status` affiche maintenant la liste réelle des services synchronisés depuis le vault (empêche l'omission d'anthropic etc.).

### v0.1.25 (2026-04-08)
- **Détection du processus agent** : Le proxy détecte si les agents locaux (NanoClaw/OpenClaw) sont vivants et affiche un feu tricolore orange sur le tableau de bord.
- **Amélioration de la poignée de glissement** : Le tri des cartes ne permet désormais de saisir que depuis la zone du feu tricolore. Empêche le glissement accidentel depuis les champs de saisie ou les boutons.

### v0.1.24 (2026-04-06)
- **Sous-commande RTK d'économie de tokens** : `wall-vault rtk <command>` filtre automatiquement la sortie des commandes shell, réduisant l'utilisation de tokens des agents IA de 60-90%. Inclut des filtres intégrés pour les commandes principales comme git et go, et tronque automatiquement les commandes non supportées. S'intègre de manière transparente avec Claude Code via le hook `PreToolUse`.

### v0.1.23 (2026-04-06)
- **Correction du changement de modèle Ollama** : Correction d'un problème où le changement du modèle Ollama dans le tableau de bord du coffre-fort n'était pas reflété sur le proxy réel. Auparavant, seule la variable d'environnement (`OLLAMA_MODEL`) était utilisée ; maintenant les paramètres du coffre-fort ont la priorité.
- **Feu tricolore automatique du service local** : Ollama, LM Studio et vLLM sont automatiquement activés quand accessibles et désactivés quand déconnectés. Fonctionne de la même manière que la bascule automatique basée sur les clés pour les services cloud.

### v0.1.22 (2026-04-05)
- **Correction de l'omission du champ content vide** : Lorsque les modèles thinking (gemini-3.1-pro, o1, claude thinking, etc.) utilisaient tous les max_tokens pour le reasoning et ne pouvaient pas produire de réponse réelle, le proxy omettait les champs `content`/`text` du JSON de réponse via `omitempty`, causant le crash des clients SDK OpenAI/Anthropic avec `Cannot read properties of undefined (reading 'trim')`. Corrigé pour toujours inclure les champs selon les spécifications API officielles.

### v0.1.21 (2026-04-05)
- **Support du modèle Gemma 4** : Les modèles de la série Gemma comme `gemma-4-31b-it` et `gemma-4-26b-a4b-it` peuvent maintenant être utilisés via l'API Google Gemini.
- **Support complet des services LM Studio / vLLM** : Auparavant, ces services manquaient dans le routage du proxy et retombaient toujours sur Ollama. Maintenant correctement routés via l'API compatible OpenAI.
- **Correction de l'affichage du service dans le tableau de bord** : Même pendant le fallback, le tableau de bord affiche toujours le service configuré par l'utilisateur.
- **Affichage du statut du service local** : Au chargement du tableau de bord, le statut de connexion des services locaux (Ollama, LM Studio, vLLM, etc.) est affiché via la couleur du point.
- **Variable d'environnement du filtre d'outils** : Mode de passage des outils configurable avec la variable d'environnement `WV_TOOL_FILTER=passthrough`.

### v0.1.20 (2026-03-28)
- **Durcissement complet de la sécurité** : Prévention XSS (41 emplacements), comparaison de tokens en temps constant, restrictions CORS, limites de taille des requêtes, prévention de traversée de chemin, authentification SSE, durcissement du limiteur de débit, et 12 autres améliorations de sécurité.

### v0.1.19 (2026-03-27)
- **Détection en ligne de Claude Code** : Les instances Claude Code ne passant pas par le proxy sont également affichées comme en ligne sur le tableau de bord.

### v0.1.18 (2026-03-26)
- **Correction de la persistance du service de fallback** : Après des erreurs temporaires causant un fallback vers Ollama, retour automatique au service d'origine lorsqu'il est rétabli.
- **Amélioration de la détection hors ligne** : Des vérifications de statut toutes les 15 secondes permettent une détection plus rapide des pannes de proxy.

### v0.1.17 (2026-03-25)
- **Tri des cartes par glisser-déposer** : Les cartes d'agents peuvent être glissées pour réorganiser.
- **Bouton d'application de config en ligne** : Les agents hors ligne affichent un bouton [:zap: Appliquer la config].
- **Type d'agent cokacdir ajouté**.

### v0.1.16 (2026-03-25)
- **Synchronisation bidirectionnelle des modèles** : Le changement du modèle pour Cline ou Claude Code dans le tableau de bord du coffre-fort est automatiquement reflété.

---

*Pour des informations API plus détaillées, consultez [API.md](API.md).*
