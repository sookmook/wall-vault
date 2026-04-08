# Manuel Utilisateur wall-vault
*(Dernière mise à jour : 2026-04-08 — v0.1.25)*

---

## Table des matières

1. [Qu'est-ce que wall-vault ?](#quest-ce-que-wall-vault)
2. [Installation](#installation)
3. [Premiers pas (assistant de configuration)](#premiers-pas)
4. [Enregistrement des clés API](#enregistrement-des-clés-api)
5. [Utilisation du proxy](#utilisation-du-proxy)
6. [Tableau de bord du coffre-fort](#tableau-de-bord-du-coffre-fort)
7. [Mode distribué (multi-bot)](#mode-distribué-multi-bot)
8. [Configuration du démarrage automatique](#configuration-du-démarrage-automatique)
9. [Doctor (outil de diagnostic)](#doctor-outil-de-diagnostic)
10. [RTK Économie de tokens](#rtk-économie-de-tokens)
11. [Référence des variables d'environnement](#référence-des-variables-denvironnement)
12. [Dépannage](#dépannage)

---

## Qu'est-ce que wall-vault ?

**wall-vault = Proxy AI + coffre-fort de clés API pour OpenClaw**

Pour utiliser les services d'IA, vous avez besoin de **clés API**. Une clé API est comme un **badge numérique** qui prouve que « cette personne est autorisée à utiliser ce service ». Cependant, ces badges ont des limites d'utilisation quotidiennes, et s'ils sont mal gérés, ils risquent d'être exposés.

wall-vault stocke ces badges dans un coffre-fort sécurisé et joue le rôle de **proxy (intermédiaire)** entre OpenClaw et les services d'IA. En termes simples, OpenClaw n'a qu'à se connecter à wall-vault, et wall-vault s'occupe automatiquement de tout le reste.

Problèmes résolus par wall-vault :

- **Rotation automatique des clés API** : Lorsqu'une clé atteint sa limite d'utilisation ou est temporairement bloquée (refroidissement), le système bascule silencieusement vers la clé suivante. OpenClaw continue de fonctionner sans interruption.
- **Basculement automatique des services (fallback)** : Si Google ne répond pas, il bascule vers OpenRouter ; si celui-ci échoue aussi, il bascule vers Ollama, LM Studio ou vLLM (IA locale) installés sur votre ordinateur. Les sessions ne sont jamais interrompues. Lorsque le service d'origine se rétablit, il rebascule automatiquement à la prochaine requête (v0.1.18+, LM Studio/vLLM : v0.1.21+).
- **Synchronisation en temps réel (SSE)** : Lorsque vous changez un modèle dans le tableau de bord du coffre-fort, cela se reflète sur l'écran OpenClaw en 1 à 3 secondes. SSE (Server-Sent Events) est une technologie où le serveur pousse les changements vers les clients en temps réel.
- **Notifications en temps réel** : Les événements comme l'épuisement des clés ou les pannes de service sont immédiatement affichés en bas de l'interface TUI d'OpenClaw (écran terminal).

> 💡 **Claude Code, Cursor et VS Code** peuvent également être connectés, mais l'objectif principal de wall-vault est d'être utilisé avec OpenClaw.

```
OpenClaw (interface TUI terminal)
        │
        ▼
  wall-vault proxy (:56244)   ← gestion des clés, routage, fallback, événements
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (340+ modèles)
        ├─ Ollama / LM Studio / vLLM (votre ordinateur, dernier recours)
        └─ OpenAI / Anthropic API
```

---

## Installation

### Linux / macOS

Ouvrez un terminal et collez les commandes suivantes.

```bash
# Linux (PC standard, serveur — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Télécharge un fichier depuis Internet.
- `chmod +x` — Rend le fichier téléchargé « exécutable ». Si vous sautez cette étape, vous obtiendrez une erreur « permission refusée ».

### Windows

Ouvrez PowerShell (en tant qu'administrateur) et exécutez les commandes suivantes.

```powershell
# Téléchargement
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Ajouter au PATH (prend effet après redémarrage de PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **Qu'est-ce que le PATH ?** C'est la liste des dossiers où votre ordinateur cherche les commandes. Vous devez ajouter wall-vault au PATH pour pouvoir exécuter `wall-vault` depuis n'importe quel dossier.

### Compilation depuis les sources (pour développeurs)

Uniquement applicable si un environnement de développement Go est installé.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (version : v0.1.25.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Version avec horodatage de compilation** : Lors de la compilation avec `make build`, la version est automatiquement générée dans un format comme `v0.1.25.20260408.022325` incluant la date et l'heure. Si vous compilez directement avec `go build ./...`, la version affichera uniquement `"dev"`.

---

## Premiers pas

### Exécuter l'assistant de configuration

Après l'installation, vous devez d'abord exécuter l'**assistant de configuration** avec la commande suivante. L'assistant vous guide pas à pas à travers les paramètres requis.

```bash
wall-vault setup
```

L'assistant parcourt les étapes suivantes :

```
1. Sélection de la langue (10 langues dont le coréen)
2. Sélection du thème (light / dark / gold / cherry / ocean)
3. Mode de fonctionnement — utilisation individuelle (standalone) ou partagée (distributed)
4. Nom du bot — le nom affiché sur le tableau de bord
5. Configuration des ports — par défaut : proxy 56244, coffre-fort 56243 (Entrée pour garder les valeurs par défaut)
6. Sélection des services d'IA — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Configuration du filtre de sécurité des outils
8. Jeton administrateur — un mot de passe pour verrouiller les fonctions d'administration du tableau de bord. Génération automatique possible
9. Mot de passe de chiffrement des clés API — pour un stockage des clés plus sécurisé (optionnel)
10. Emplacement de sauvegarde du fichier de configuration
```

> ⚠️ **Assurez-vous de mémoriser votre jeton administrateur.** Vous en aurez besoin plus tard pour ajouter des clés ou modifier des paramètres dans le tableau de bord. Si vous le perdez, vous devrez modifier directement le fichier de configuration.

Une fois l'assistant terminé, un fichier de configuration `wall-vault.yaml` est automatiquement généré.

### Exécution

```bash
wall-vault start
```

Deux serveurs démarrent simultanément :

- **Proxy** (`http://localhost:56244`) — l'intermédiaire entre OpenClaw et les services d'IA
- **Coffre-fort de clés** (`http://localhost:56243`) — gestion des clés API et tableau de bord web

Ouvrez `http://localhost:56243` dans votre navigateur pour accéder au tableau de bord.

---

## Enregistrement des clés API

Il existe quatre méthodes pour enregistrer des clés API. **La méthode 1 (variables d'environnement) est recommandée pour les débutants.**

### Méthode 1 : Variables d'environnement (recommandée — la plus simple)

Les variables d'environnement sont des **valeurs prédéfinies** que les programmes lisent au démarrage. Entrez-les dans votre terminal comme ceci :

```bash
# Enregistrer une clé Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Enregistrer une clé OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Démarrer après l'enregistrement
wall-vault start
```

Si vous avez plusieurs clés, connectez-les avec des virgules. wall-vault les utilisera automatiquement en rotation (round-robin) :

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Conseil** : La commande `export` ne s'applique qu'à la session terminal en cours. Pour persister après les redémarrages, ajoutez les lignes dans votre fichier `~/.bashrc` ou `~/.zshrc`.

### Méthode 2 : Interface du tableau de bord (pointer et cliquer)

1. Ouvrez `http://localhost:56243` dans votre navigateur
2. Cliquez sur `[+ Ajouter]` dans la carte **🔑 Clés API** en haut
3. Entrez le type de service, la valeur de la clé, l'étiquette (nom descriptif) et la limite quotidienne, puis enregistrez

### Méthode 3 : API REST (pour l'automatisation/les scripts)

L'API REST est une méthode permettant aux programmes d'échanger des données via HTTP. Utile pour l'enregistrement automatisé par script.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer VOTRE_JETON_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Clé principale",
    "daily_limit": 1000
  }'
```

### Méthode 4 : Drapeaux proxy (pour tests rapides)

Pour des tests temporaires sans enregistrement formel. Les clés sont perdues à la fermeture du programme.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Utilisation du proxy

### Utilisation avec OpenClaw (objectif principal)

Voici comment configurer OpenClaw pour se connecter aux services d'IA via wall-vault.

Ouvrez `~/.openclaw/openclaw.json` et ajoutez le contenu suivant :

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // jeton d'agent vault
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // 1M contexte gratuit
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Méthode plus simple** : Cliquez sur le bouton **🦞 Copier la configuration OpenClaw** sur la carte d'agent du tableau de bord. Un extrait avec le jeton et l'adresse déjà remplis sera copié dans votre presse-papiers. Il suffit de le coller.

**Vers où le préfixe `wall-vault/` dans les noms de modèles redirige-t-il ?**

wall-vault détermine automatiquement vers quel service d'IA envoyer les requêtes en fonction du nom du modèle :

| Format du modèle | Service cible |
|-----------------|--------------|
| `wall-vault/gemini-*` | Directement vers Google Gemini |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | Directement vers OpenAI |
| `wall-vault/claude-*` | Anthropic via OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (1M tokens contexte gratuit) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/modèle`, `openai/modèle`, `anthropic/modèle`, etc. | Directement vers le service correspondant |
| `custom/google/modèle`, `custom/openai/modèle`, etc. | Supprime le préfixe `custom/` et redirige |
| `modèle:cloud` | Supprime le suffixe `:cloud` et redirige vers OpenRouter |

> 💡 **Qu'est-ce que le contexte ?** C'est la quantité de conversation qu'une IA peut mémoriser en une fois. 1M (un million de tokens) signifie qu'elle peut traiter de très longues conversations ou documents en une seule passe.

### Connexion directe au format Gemini API (compatibilité avec les outils existants)

Si vous avez des outils qui utilisaient directement l'API Google Gemini, changez simplement l'adresse pour wall-vault :

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Ou si l'outil spécifie les URL directement :

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Utilisation avec le SDK OpenAI (Python)

Vous pouvez également connecter wall-vault depuis du code Python utilisant l'IA. Changez simplement la `base_url` :

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # wall-vault gère les clés API pour vous
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # utiliser le format provider/model
    messages=[{"role": "user", "content": "Bonjour"}]
)
```

### Changer de modèle en cours d'exécution

Pour changer le modèle d'IA pendant que wall-vault est déjà en cours d'exécution :

```bash
# Changer de modèle par requête directe au proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# En mode distribué (multi-bot), changer depuis le serveur coffre-fort → synchronisé instantanément via SSE
curl -X PUT http://localhost:56243/admin/clients/mon-bot-id \
  -H "Authorization: Bearer VOTRE_JETON_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Vérifier les modèles disponibles

```bash
# Voir la liste complète
curl http://localhost:56244/api/models | python3 -m json.tool

# Voir uniquement les modèles Google
curl "http://localhost:56244/api/models?service=google"

# Rechercher par nom (ex. : modèles contenant "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Principaux modèles par service :**

| Service | Principaux modèles |
|---------|-------------------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ (Hunter Alpha 1M contexte gratuit, DeepSeek R1/V3, Qwen 2.5, etc.) |
| Ollama | Détection automatique des modèles sur votre serveur local |
| LM Studio | Serveur local (port 1234) |
| vLLM | Serveur local (port 8000) |

---

## Tableau de bord du coffre-fort

Ouvrez `http://localhost:56243` dans votre navigateur pour accéder au tableau de bord.

**Disposition :**
- **Barre supérieure (fixe)** : Logo, sélecteur de langue/thème, statut de connexion SSE
- **Grille de cartes** : Cartes d'agents, de services et de clés API disposées en mosaïque

### Carte des clés API

Une carte pour gérer d'un coup d'œil toutes les clés API enregistrées.

- Affiche les listes de clés organisées par service.
- `today_usage` : Nombre de tokens (caractères lus/écrits par l'IA) traités avec succès aujourd'hui
- `today_attempts` : Nombre total d'appels aujourd'hui (succès et échecs inclus)
- Bouton `[+ Ajouter]` pour enregistrer de nouvelles clés, `✕` pour les supprimer.

> 💡 **Qu'est-ce qu'un token ?** C'est une unité utilisée par l'IA pour traiter le texte. Un token correspond approximativement à un mot anglais, ou 1-2 caractères français. La tarification des API est généralement calculée en fonction du nombre de tokens.

### Carte d'agent

Une carte affichant le statut des bots (agents) connectés au proxy wall-vault.

**Le statut de connexion est affiché en 4 niveaux :**

| Indicateur | Statut | Signification |
|-----------|--------|--------------|
| 🟢 | En cours d'exécution | Le proxy fonctionne normalement |
| 🟡 | Retardé | Répond mais lentement |
| 🔴 | Hors ligne | Le proxy ne répond pas |
| ⚫ | Non connecté / Désactivé | Le proxy ne s'est jamais connecté au coffre-fort ou est désactivé |

**Guide des boutons en bas de la carte d'agent :**

Lorsque vous enregistrez un agent et spécifiez le **type d'agent**, des boutons pratiques correspondant à ce type apparaissent automatiquement.

---

#### 🔘 Bouton Copier la configuration — génère automatiquement les paramètres de connexion

En cliquant sur le bouton, un extrait de configuration avec le jeton, l'adresse proxy et les informations du modèle de l'agent déjà remplis est copié dans votre presse-papiers. Collez simplement le contenu copié à l'emplacement indiqué dans le tableau ci-dessous pour terminer la configuration de connexion.

| Bouton | Type d'agent | Emplacement de collage |
|--------|-------------|----------------------|
| 🦞 Copier la config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Copier la config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Copier la config Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Copier la config Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Copier la config VSCode | `vscode` | `~/.continue/config.json` |

**Exemple — pour le type Claude Code, ce contenu est copié :**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.0.6:56244/v1",
  "apiKey": "jeton-de-cet-agent"
}
```

**Exemple — pour le type VSCode (Continue) :**

```yaml
# ~/.continue/config.yaml  ← coller dans config.yaml, pas config.json
name: My Config
version: 0.0.1
schema: v1

models:
  - name: wall-vault proxy
    provider: openai
    model: gemini-2.5-flash
    apiBase: http://192.168.0.6:56244/v1
    apiKey: jeton-de-cet-agent
    roles:
      - chat
      - edit
      - apply
```

> ⚠️ **La dernière version de Continue utilise `config.yaml`.** Si `config.yaml` existe, `config.json` est complètement ignoré. Assurez-vous de coller dans `config.yaml`.

**Exemple — pour le type Cursor :**

```
Base URL : http://192.168.0.6:56244/v1
API Key  : jeton-de-cet-agent

// Ou variables d'environnement :
OPENAI_BASE_URL=http://192.168.0.6:56244/v1
OPENAI_API_KEY=jeton-de-cet-agent
```

> ⚠️ **Si la copie dans le presse-papiers ne fonctionne pas** : Les politiques de sécurité du navigateur peuvent bloquer la copie. Si une boîte de texte popup apparaît, utilisez Ctrl+A pour tout sélectionner, puis Ctrl+C pour copier.

---

#### ⚡ Bouton d'application automatique — un clic pour terminer la configuration

Pour les types d'agents `cline`, `claude-code`, `openclaw` ou `nanoclaw`, un bouton **⚡ Appliquer la configuration** apparaît sur la carte d'agent. Cliquer sur ce bouton met automatiquement à jour le fichier de configuration local de l'agent.

| Bouton | Type d'agent | Fichier cible |
|--------|-------------|--------------|
| ⚡ Appliquer la config Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Appliquer la config Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Appliquer la config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Appliquer la config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Ce bouton envoie des requêtes à **localhost:56244** (proxy local). Le proxy doit être en cours d'exécution sur cette machine pour que cela fonctionne.

---

#### 🔀 Tri des cartes par glisser-déposer (v0.1.17, amélioré v0.1.25)

Vous pouvez **glisser** les cartes d'agents du tableau de bord pour les réorganiser dans l'ordre souhaité.

1. Saisissez la zone **feu de signalisation (●)** en haut à gauche de la carte avec votre souris et faites glisser
2. Déposez-la sur une autre carte pour échanger les positions

> 💡 Le corps de la carte (champs de saisie, boutons, etc.) n'est pas déplaçable. Vous ne pouvez saisir que depuis la zone du feu de signalisation.

#### 🟠 Détection du processus d'agent (v0.1.25)

Lorsque le proxy fonctionne normalement mais que le processus d'agent local (NanoClaw, OpenClaw) est mort, le feu de signalisation de la carte devient **orange (clignotant)** et affiche un message « Processus d'agent arrêté ».

- 🟢 Vert : Proxy + agent tous deux normaux
- 🟠 Orange (clignotant) : Proxy normal, agent mort
- 🔴 Rouge : Proxy hors ligne
3. L'ordre modifié est **immédiatement enregistré sur le serveur** et persiste après actualisation

> 💡 Les appareils tactiles (mobile/tablette) ne sont pas encore pris en charge. Veuillez utiliser un navigateur de bureau.

---

#### 🔄 Synchronisation bidirectionnelle des modèles (v0.1.16)

Lorsque vous changez le modèle d'un agent dans le tableau de bord du coffre-fort, la configuration locale de l'agent est automatiquement mise à jour.

**Pour Cline :**
- Changer le modèle dans le coffre-fort → événement SSE → le proxy met à jour les champs de modèle dans `globalState.json`
- Champs mis à jour : `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` et la clé API ne sont pas touchés
- **Un rechargement de VS Code (`Ctrl+Alt+R` ou `Ctrl+Shift+P` → `Developer: Reload Window`) est nécessaire**
  - Car Cline ne relit pas les fichiers de configuration pendant l'exécution

**Pour Claude Code :**
- Changer le modèle dans le coffre-fort → événement SSE → le proxy met à jour le champ `model` dans `settings.json`
- Recherche automatiquement les chemins WSL et Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Direction inverse (agent → coffre-fort) :**
- Lorsqu'un agent (Cline, Claude Code, etc.) envoie une requête au proxy, le proxy inclut les informations de service/modèle de ce client dans le heartbeat
- La carte d'agent sur le tableau de bord du coffre-fort affiche le service/modèle actuellement utilisé en temps réel

> 💡 **Point clé** : Le proxy identifie les agents par le jeton d'autorisation dans les requêtes et route automatiquement vers le service/modèle configuré dans le coffre-fort. Même si Cline ou Claude Code envoie un nom de modèle différent, le proxy le remplace par la configuration du coffre-fort.

---

### Utiliser Cline dans VS Code — Guide détaillé

#### Étape 1 : Installer Cline

Installez **Cline** (ID : `saoudrizwan.claude-dev`) depuis le marketplace d'extensions VS Code.

#### Étape 2 : Enregistrer l'agent dans le coffre-fort

1. Ouvrez le tableau de bord du coffre-fort (`http://IP_COFFRE:56243`)
2. Cliquez sur **+ Ajouter** dans la section **Agents**
3. Entrez les informations suivantes :

| Champ | Valeur | Description |
|-------|--------|-------------|
| ID | `mon_cline` | Identifiant unique (alphanumérique, sans espaces) |
| Nom | `Mon Cline` | Nom affiché sur le tableau de bord |
| Type d'agent | `cline` | ← doit sélectionner `cline` |
| Service | Sélectionner le service souhaité (ex. : `google`) | |
| Modèle | Entrer le modèle souhaité (ex. : `gemini-2.5-flash`) | |

4. Cliquez sur **Enregistrer** — un jeton est automatiquement généré

#### Étape 3 : Connecter à Cline

**Méthode A — Application automatique (recommandée)**

1. Vérifiez que le **proxy** wall-vault est en cours d'exécution sur cette machine (`localhost:56244`)
2. Cliquez sur le bouton **⚡ Appliquer la config Cline** sur la carte d'agent du tableau de bord
3. Succès lorsque la notification « Configuration appliquée ! » apparaît
4. Rechargez VS Code (`Ctrl+Alt+R`)

**Méthode B — Configuration manuelle**

Ouvrez les paramètres (⚙️) dans la barre latérale de Cline :
- **API Provider** : `OpenAI Compatible`
- **Base URL** : `http://ADRESSE_PROXY:56244/v1`
  - Même machine : `http://localhost:56244/v1`
  - Autre machine (ex. : Mac Mini) : `http://192.168.0.6:56244/v1`
- **API Key** : Jeton émis par le coffre-fort (copier depuis la carte d'agent)
- **Model ID** : Modèle défini dans le coffre-fort (ex. : `gemini-2.5-flash`)

#### Étape 4 : Vérification

Envoyez n'importe quel message dans le chat de Cline. Si tout fonctionne :
- La carte d'agent sur le tableau de bord affiche un **point vert (● En cours d'exécution)**
- La carte affiche le service/modèle actuel (ex. : `google / gemini-2.5-flash`)

#### Changer de modèle

Lorsque vous souhaitez changer le modèle de Cline, changez-le depuis le **tableau de bord du coffre-fort** :

1. Changez le menu déroulant service/modèle sur la carte d'agent
2. Cliquez sur **Appliquer**
3. Rechargez VS Code (`Ctrl+Alt+R`) — le nom du modèle dans le pied de page de Cline est mis à jour
4. Le nouveau modèle est utilisé à partir de la prochaine requête

> 💡 En pratique, le proxy identifie les requêtes de Cline par le jeton et route vers le modèle configuré dans le coffre-fort. Même sans recharger VS Code, **le modèle réellement utilisé change immédiatement** — le rechargement sert uniquement à mettre à jour l'affichage du modèle dans l'interface de Cline.

#### Détection de déconnexion

Lorsque vous fermez VS Code, la carte d'agent sur le tableau de bord devient jaune (retardé) après environ **90 secondes**, puis rouge (hors ligne) après **3 minutes**. (À partir de v0.1.18, les vérifications d'état toutes les 15 secondes accélèrent la détection hors ligne.)

#### Dépannage

| Symptôme | Cause | Solution |
|----------|-------|----------|
| Erreur « Connexion échouée » dans Cline | Proxy non démarré ou mauvaise adresse | Vérifier le proxy avec `curl http://localhost:56244/health` |
| Le point vert n'apparaît pas dans le coffre-fort | Clé API (jeton) non configurée | Cliquer à nouveau sur **⚡ Appliquer la config Cline** |
| Le modèle en pied de page de Cline ne change pas | Cline met en cache la configuration | Recharger VS Code (`Ctrl+Alt+R`) |
| Mauvais nom de modèle affiché | Ancien bug (corrigé en v0.1.16) | Mettre à jour le proxy vers v0.1.16+ |

---

#### 🟣 Bouton Copier la commande de déploiement — pour l'installation sur de nouvelles machines

Utilisé lors de la première installation du proxy wall-vault sur un nouvel ordinateur et de la connexion au coffre-fort. Cliquer sur le bouton copie le script d'installation complet. Collez-le dans le terminal du nouvel ordinateur et exécutez-le — les opérations suivantes sont effectuées en une seule fois :

1. Installation du binaire wall-vault (ignoré si déjà installé)
2. Enregistrement automatique du service utilisateur systemd
3. Démarrage du service et connexion automatique au coffre-fort

> 💡 Le script contient déjà le jeton et l'adresse du serveur coffre-fort de cet agent, vous pouvez donc l'exécuter immédiatement après le collage sans aucune modification.

---

### Carte de service

Une carte pour activer/désactiver et configurer les services d'IA.

- Interrupteurs à bascule pour activer/désactiver chaque service
- Entrez l'adresse d'un serveur d'IA local (Ollama, LM Studio, vLLM, etc. sur votre ordinateur) pour découvrir automatiquement les modèles disponibles.
- **Affichage du statut de connexion des services locaux** : Le point ● à côté du nom du service est **vert** quand connecté, **gris** quand non connecté
- **Feu de signalisation automatique des services locaux** (v0.1.23+) : Les services locaux (Ollama, LM Studio, vLLM) sont automatiquement activés/désactivés en fonction de la disponibilité de connexion. Lorsqu'un service se connecte, le point ● devient vert et la case à cocher est cochée dans les 15 secondes ; en cas de déconnexion, il est automatiquement désactivé. Fonctionne de la même manière que la bascule automatique des services cloud (Google, OpenRouter, etc.) basée sur la disponibilité des clés API.

> 💡 **Si votre service local s'exécute sur un autre ordinateur** : Entrez l'IP de cet ordinateur dans le champ URL du service. Exemple : `http://192.168.0.6:11434` (Ollama), `http://192.168.0.6:1234` (LM Studio). Si le service n'est lié qu'à `127.0.0.1` au lieu de `0.0.0.0`, l'accès via IP externe ne fonctionnera pas — vérifiez l'adresse de liaison dans les paramètres du service.

### Saisie du jeton administrateur

Lorsque vous essayez d'utiliser des fonctions importantes comme l'ajout ou la suppression de clés dans le tableau de bord, une fenêtre de saisie du jeton administrateur apparaît. Entrez le jeton que vous avez défini lors de l'assistant de configuration. Une fois entré, il reste valide jusqu'à la fermeture du navigateur.

> ⚠️ **Si les échecs d'authentification dépassent 10 en 15 minutes, l'IP est temporairement bloquée.** Si vous avez oublié le jeton, vérifiez le champ `admin_token` dans le fichier `wall-vault.yaml`.

---

## Mode distribué (multi-bot)

Une configuration pour **partager un seul coffre-fort de clés** lorsque OpenClaw est utilisé sur plusieurs ordinateurs simultanément. C'est pratique car la gestion des clés se fait en un seul endroit.

### Exemple de configuration

```
[Serveur coffre-fort de clés]
  wall-vault vault    (coffre-fort :56243, tableau de bord)

[WSL Alpha]          [Raspberry Pi Gamma]    [Mac Mini Local]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ sync SSE            ↕ sync SSE              ↕ sync SSE
```

Tous les bots pointent vers le serveur coffre-fort central, donc lorsque vous changez un modèle ou ajoutez une clé dans le coffre-fort, cela se reflète instantanément sur tous les bots.

### Étape 1 : Démarrer le serveur coffre-fort de clés

Exécutez ceci sur l'ordinateur qui servira de serveur coffre-fort :

```bash
wall-vault vault
```

### Étape 2 : Enregistrer chaque bot (client)

Pré-enregistrez les informations de chaque bot qui se connectera au serveur coffre-fort :

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer VOTRE_JETON_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "Bot A",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### Étape 3 : Démarrer le proxy sur chaque ordinateur bot

Sur chaque ordinateur avec un bot, démarrez le proxy avec l'adresse du serveur coffre-fort et le jeton :

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Remplacez **`192.168.x.x`** par l'adresse IP interne réelle de l'ordinateur serveur coffre-fort. Vous pouvez la vérifier via les paramètres de votre routeur ou la commande `ip addr`.

---

## Configuration du démarrage automatique

Si c'est fastidieux de démarrer manuellement wall-vault à chaque redémarrage de l'ordinateur, enregistrez-le comme service système. Une fois enregistré, il démarre automatiquement au démarrage.

### Linux — systemd (la plupart des distributions Linux)

systemd est le système qui démarre et gère automatiquement les programmes sous Linux :

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Voir les journaux :

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

## Doctor (outil de diagnostic)

La commande `doctor` est un outil qui **auto-diagnostique et répare** la configuration de wall-vault.

```bash
wall-vault doctor check   # Diagnostiquer l'état actuel (lecture seule, ne change rien)
wall-vault doctor fix     # Corriger automatiquement les problèmes
wall-vault doctor all     # Diagnostic + correction automatique en une étape
```

> 💡 Si quelque chose semble anormal, essayez d'abord `wall-vault doctor all`. Cela détecte automatiquement de nombreux problèmes.

---

## RTK Économie de tokens

*(v0.1.24+)*

**RTK (Outil d'économie de tokens)** compresse automatiquement la sortie des commandes shell exécutées par les agents de codage IA (comme Claude Code), réduisant la consommation de tokens. Par exemple, 15 lignes de sortie `git status` sont compressées en un résumé de 2 lignes.

### Utilisation de base

```bash
# Envelopper les commandes avec wall-vault rtk pour filtrer automatiquement la sortie
wall-vault rtk git status          # affiche uniquement la liste des fichiers modifiés
wall-vault rtk git diff HEAD~1     # lignes modifiées + contexte minimal uniquement
wall-vault rtk git log -10         # hash + message sur une ligne
wall-vault rtk go test ./...       # affiche uniquement les tests échoués
wall-vault rtk ls -la              # les commandes non prises en charge sont automatiquement tronquées
```

### Commandes prises en charge et économies

| Commande | Méthode de filtrage | Économie |
|----------|-------------------|----------|
| `git status` | Résumé des fichiers modifiés uniquement | ~87% |
| `git diff` | Lignes modifiées + 3 lignes de contexte | ~60-94% |
| `git log` | Hash + première ligne du message | ~90% |
| `git push/pull/fetch` | Suppression de la progression, résumé uniquement | ~80% |
| `go test` | Afficher les échecs uniquement, compter les succès | ~88-99% |
| `go build/vet` | Afficher les erreurs uniquement | ~90% |
| Toutes les autres commandes | 50 premières + 50 dernières lignes, max 32Ko | Variable |

### Pipeline de filtrage en 3 étapes

1. **Filtre structurel spécifique à la commande** — Comprend les formats de sortie de git, go, etc. et n'extrait que les parties significatives
2. **Post-traitement par regex** — Supprime les codes couleur ANSI, réduit les lignes vides, agrège les lignes dupliquées
3. **Passage direct + troncature** — Les commandes non prises en charge ne conservent que les 50 premières/dernières lignes

### Intégration avec Claude Code

Vous pouvez configurer toutes les commandes shell pour passer automatiquement par RTK en utilisant le hook `PreToolUse` de Claude Code.

```bash
# Installer le hook (ajouté automatiquement à Claude Code settings.json)
wall-vault rtk hook install
```

Ou ajouter manuellement dans `~/.claude/settings.json` :

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

> 💡 **Préservation du code de sortie** : RTK retourne le code de sortie de la commande originale tel quel. Si une commande échoue (code de sortie ≠ 0), l'IA détecte correctement l'échec.

> 💡 **Sortie forcée en anglais** : RTK exécute les commandes avec `LC_ALL=C` pour toujours produire une sortie en anglais, quel que soit le paramètre de langue du système. Cela garantit le bon fonctionnement des filtres.

---

## Référence des variables d'environnement

Les variables d'environnement sont un moyen de passer des valeurs de configuration aux programmes. Entrez-les dans votre terminal sous la forme `export VARIABLE=valeur`, ou ajoutez-les dans votre fichier de service de démarrage automatique pour un effet permanent.

| Variable | Description | Valeur d'exemple |
|----------|-------------|-----------------|
| `WV_LANG` | Langue du tableau de bord | `ko`, `en`, `ja` |
| `WV_THEME` | Thème du tableau de bord | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Clé API Google (virgule pour plusieurs) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Clé API OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Adresse du serveur coffre-fort en mode distribué | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Jeton d'authentification du client (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Jeton administrateur | `admin-token-here` |
| `WV_MASTER_PASS` | Mot de passe de chiffrement des clés API | `my-password` |
| `WV_AVATAR` | Chemin du fichier image d'avatar (relatif à `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Adresse du serveur local Ollama | `http://192.168.x.x:11434` |

---

## Dépannage

### Quand le proxy ne démarre pas

Le port est souvent déjà utilisé par un autre programme.

```bash
ss -tlnp | grep 56244   # Vérifier ce qui utilise le port 56244
wall-vault proxy --port 8080   # Démarrer avec un numéro de port différent
```

### Erreurs de clé API (429, 402, 401, 403, 582)

| Code d'erreur | Signification | Solution |
|--------------|--------------|----------|
| **429** | Trop de requêtes (limite d'utilisation dépassée) | Attendre un moment ou ajouter plus de clés |
| **402** | Paiement requis ou crédits épuisés | Recharger les crédits sur le service concerné |
| **401 / 403** | Clé invalide ou pas de permission | Revérifier la valeur de la clé et réenregistrer |
| **582** | Surcharge du gateway (refroidissement de 5 minutes) | Se résout automatiquement après 5 minutes |

```bash
# Vérifier la liste des clés enregistrées et leur statut
curl -H "Authorization: Bearer VOTRE_JETON_ADMIN" http://localhost:56243/admin/keys

# Réinitialiser les compteurs d'utilisation des clés
curl -X POST -H "Authorization: Bearer VOTRE_JETON_ADMIN" http://localhost:56243/admin/keys/reset
```

### Quand l'agent affiche « Non connecté »

« Non connecté » signifie que le processus proxy n'envoie pas de signaux heartbeat au coffre-fort. **Cela ne signifie pas que les paramètres ne sont pas enregistrés.** Le proxy doit être en cours d'exécution avec l'adresse du serveur coffre-fort et le jeton pour apparaître comme connecté.

```bash
# Démarrer le proxy avec l'adresse du serveur coffre-fort, le jeton et l'ID client
WV_VAULT_URL=http://SERVEUR_COFFRE:56243 \
WV_VAULT_TOKEN=jeton-client \
WV_VAULT_CLIENT_ID=id-client \
wall-vault proxy
```

Une fois connecté avec succès, le tableau de bord affiche 🟢 En cours d'exécution dans les 20 secondes environ.

### Quand Ollama ne se connecte pas

Ollama est un programme qui exécute l'IA directement sur votre ordinateur. Vérifiez d'abord si Ollama est en cours d'exécution.

```bash
curl http://localhost:11434/api/tags   # Si la liste des modèles apparaît, ça fonctionne
export OLLAMA_URL=http://192.168.x.x:11434   # Si exécuté sur un autre ordinateur
```

> ⚠️ Si Ollama ne répond pas, démarrez-le d'abord avec la commande `ollama serve`.

> ⚠️ **Les gros modèles sont lents** : Les gros modèles comme `qwen3.5:35b` ou `deepseek-r1` peuvent prendre plusieurs minutes pour générer une réponse. Même s'il semble n'y avoir aucune réponse, le traitement peut être en cours normalement — veuillez patienter.

---

## Changements récents (v0.1.16 ~ v0.1.25)

### v0.1.25 (2026-04-08)
- **Détection du processus d'agent** : Le proxy détecte si les agents locaux (NanoClaw/OpenClaw) sont actifs et affiche un feu de signalisation orange sur le tableau de bord.
- **Amélioration de la poignée de glissement** : Le tri des cartes ne fonctionne que depuis la zone du feu de signalisation (●). Empêche le glissement accidentel depuis les champs de saisie ou les boutons.

### v0.1.24 (2026-04-06)
- **Sous-commande RTK d'économie de tokens** : `wall-vault rtk <command>` filtre automatiquement la sortie des commandes shell, réduisant la consommation de tokens des agents IA de 60-90%. Filtres intégrés pour les commandes principales comme git et go, avec troncature automatique pour les commandes non prises en charge. Intégration transparente avec les hooks `PreToolUse` de Claude Code.

### v0.1.23 (2026-04-06)
- **Correction du changement de modèle Ollama** : Correction du problème où le changement de modèle Ollama dans le tableau de bord du coffre-fort n'était pas effectivement reflété dans le proxy. Auparavant, seule la variable d'environnement (`OLLAMA_MODEL`) était utilisée, maintenant les paramètres du coffre-fort ont priorité.
- **Feu de signalisation automatique des services locaux** : Ollama, LM Studio et vLLM s'activent automatiquement lorsqu'ils sont connectables et se désactivent automatiquement en cas de déconnexion. Même mécanisme que la bascule automatique des services cloud basée sur les clés.

### v0.1.22 (2026-04-05)
- **Correction du champ content vide** : Correction du problème où les modèles de réflexion (gemini-3.1-pro, o1, claude thinking, etc.) qui utilisaient tous les max_tokens pour le raisonnement sans produire de réponses réelles faisaient que le proxy omettait les champs `content`/`text` via `omitempty`, faisant planter les clients SDK OpenAI/Anthropic avec des erreurs `Cannot read properties of undefined (reading 'trim')`. Modifié pour toujours inclure les champs conformément à la spécification API officielle.

### v0.1.21 (2026-04-05)
- **Support des modèles Gemma 4** : Les modèles Gemma comme `gemma-4-31b-it` et `gemma-4-26b-a4b-it` peuvent maintenant être utilisés via l'API Google Gemini.
- **Support officiel LM Studio / vLLM** : Auparavant, ces services étaient absents du routage proxy et revenaient toujours à Ollama. Maintenant correctement routés via l'API compatible OpenAI.
- **Correction de l'affichage des services dans le tableau de bord** : Le tableau de bord affiche toujours le service configuré par l'utilisateur même en cas de fallback.
- **Affichage du statut des services locaux** : Affiche le statut de connexion des services locaux (Ollama, LM Studio, vLLM, etc.) via la couleur du point ● au chargement du tableau de bord.
- **Variable d'environnement du filtre d'outils** : Le mode de passage des outils peut être défini avec la variable d'environnement `WV_TOOL_FILTER=passthrough`.

### v0.1.20 (2026-03-28)
- **Renforcement complet de la sécurité** : Prévention XSS (41 points), comparaison de jetons en temps constant, restriction CORS, limites de taille des requêtes, prévention de traversée de chemin, authentification SSE, renforcement du limiteur de débit et 12 améliorations de sécurité au total.

### v0.1.19 (2026-03-27)
- **Détection en ligne de Claude Code** : Claude Code fonctionnant sans passer par le proxy est maintenant affiché comme en ligne sur le tableau de bord.

### v0.1.18 (2026-03-26)
- **Correction du blocage du service de fallback** : Après un fallback temporaire vers Ollama, retour automatique au service d'origine lors de sa récupération.
- **Amélioration de la détection hors ligne** : Les vérifications d'état toutes les 15 secondes accélèrent la détection d'arrêt du proxy.

### v0.1.17 (2026-03-25)
- **Tri des cartes par glisser-déposer** : Les cartes d'agents peuvent être réorganisées par glissement.
- **Boutons d'application de configuration en ligne** : Des boutons [⚡ Appliquer la config] apparaissent sur les cartes d'agents hors ligne.
- **Type d'agent cokacdir ajouté**.

### v0.1.16 (2026-03-25)
- **Synchronisation bidirectionnelle des modèles** : Le changement des modèles Cline ou Claude Code depuis le tableau de bord du coffre-fort est automatiquement reflété.

---

*Pour des informations API plus détaillées, consultez [API.md](API.md).*
