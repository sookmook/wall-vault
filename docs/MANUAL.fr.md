# Manuel Utilisateur wall-vault
*(Derniere mise a jour: 2026-04-06 — v0.1.23)*

---

## Table des matieres

1. [Qu'est-ce que wall-vault ?](#quest-ce-que-wall-vault-)
2. [Installation](#installation)
3. [Premier demarrage (assistant setup)](#premier-demarrage-assistant-setup)
4. [Enregistrer une cle API](#enregistrer-une-cle-api)
5. [Utiliser le proxy](#utiliser-le-proxy)
6. [Tableau de bord du coffre-fort](#tableau-de-bord-du-coffre-fort)
7. [Mode distribue (multi-bots)](#mode-distribue-multi-bots)
8. [Demarrage automatique](#demarrage-automatique)
9. [Doctor — l'outil de diagnostic](#doctor--loutil-de-diagnostic)
10. [Reference des variables d'environnement](#reference-des-variables-denvironnement)
11. [Resolution des problemes](#resolution-des-problemes)

---

## Qu'est-ce que wall-vault ?

**wall-vault = proxy IA + coffre-fort de cles API, concu pour OpenClaw**

Pour utiliser un service d'intelligence artificielle, vous avez besoin d'une **cle API** — c'est-a-dire un **badge numerique** qui prouve que vous avez le droit d'utiliser ce service. Ce badge a une limite d'utilisation quotidienne et peut etre compromis s'il est mal gere.

wall-vault conserve tous ces badges dans un coffre-fort securise, et joue le role de **proxy (intermediaire)** entre OpenClaw et les services IA. En clair : OpenClaw ne connait que wall-vault, et wall-vault s'occupe de tout le reste.

Ce que wall-vault resout pour vous :

- **Rotation automatique des cles** : quand une cle atteint sa limite ou est temporairement bloquee (cooldown), wall-vault bascule discretement vers la cle suivante. OpenClaw continue de fonctionner sans interruption.
- **Basculement automatique de service (fallback)** : si Google ne repond pas, wall-vault passe a OpenRouter ; si OpenRouter est aussi indisponible, il bascule vers Ollama, LM Studio ou vLLM (IA locale installee sur votre machine). La session n'est jamais interrompue. Lorsque le service d'origine est retabli, le retour s'effectue automatiquement des la requete suivante (v0.1.18+, LM Studio/vLLM : v0.1.21+).
- **Synchronisation en temps reel (SSE)** : si vous changez de modele dans le tableau de bord, le changement se reflete dans OpenClaw en 1 a 3 secondes. SSE (Server-Sent Events) est une technologie ou le serveur pousse les mises a jour vers le client en temps reel.
- **Notifications en temps reel** : les evenements comme l'epuisement d'une cle ou une panne de service s'affichent immediatement en bas de l'interface TUI (terminal) d'OpenClaw.

> 💡 **Claude Code, Cursor et VS Code** peuvent egalement etre connectes a wall-vault, mais l'usage principal reste avec OpenClaw.

```
OpenClaw (interface TUI dans le terminal)
        │
        ▼
  wall-vault proxy (:56244)   ← gestion des cles, routage, fallback, evenements
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (plus de 340 modeles)
        ├─ Ollama / LM Studio / vLLM (votre machine, dernier recours)
        └─ OpenAI / Anthropic API
```

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

- `curl -L ...` — Telecharge un fichier depuis Internet.
- `chmod +x` — Rend le fichier telecharge executable. Si vous sautez cette etape, vous obtiendrez une erreur « permission refusee ».

### Windows

Ouvrez PowerShell (en tant qu'administrateur) et executez les commandes suivantes :

```powershell
# Telechargement
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Ajouter au PATH (prend effet apres le redemarrage de PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **Qu'est-ce que le PATH ?** C'est la liste des dossiers ou votre ordinateur cherche les commandes. En ajoutant wall-vault au PATH, vous pouvez lancer `wall-vault` depuis n'importe quel repertoire.

### Compiler depuis les sources (pour les developpeurs)

Cela ne s'applique que si vous avez un environnement de developpement Go installe.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (version : v0.1.23.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Version avec horodatage** : en compilant avec `make build`, la version est generee automatiquement au format `v0.1.23.20260406.211004` incluant la date et l'heure. En compilant directement avec `go build ./...`, la version affichera simplement `"dev"`.

---

## Premier demarrage (assistant setup)

### Lancer l'assistant de configuration

Apres l'installation, lancez imperativement l'**assistant de configuration**. Il vous guidera etape par etape en vous posant les questions necessaires.

```bash
wall-vault setup
```

Voici les etapes suivies par l'assistant :

```
1. Selection de la langue (10 langues dont le francais)
2. Selection du theme (light / dark / gold / cherry / ocean)
3. Mode de fonctionnement — seul (standalone) ou multi-machines (distributed)
4. Nom du bot — le nom affiche dans le tableau de bord
5. Configuration des ports — par defaut : proxy 56244, coffre-fort 56243 (appuyez sur Entree si OK)
6. Selection des services IA — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Configuration du filtre de securite des outils
8. Token administrateur — un mot de passe pour securiser les fonctions d'administration ; peut etre genere automatiquement
9. Mot de passe de chiffrement des cles API — pour un stockage encore plus securise (optionnel)
10. Emplacement du fichier de configuration
```

> ⚠️ **N'oubliez pas votre token administrateur.** Vous en aurez besoin pour ajouter des cles ou modifier les parametres dans le tableau de bord. Si vous le perdez, vous devrez editer manuellement le fichier de configuration.

Une fois l'assistant termine, un fichier de configuration `wall-vault.yaml` est automatiquement cree.

### Demarrage

```bash
wall-vault start
```

Deux serveurs demarrent simultanement :

- **Proxy** (`http://localhost:56244`) — l'intermediaire entre OpenClaw et les services IA
- **Coffre-fort** (`http://localhost:56243`) — gestion des cles API et tableau de bord web

Ouvrez `http://localhost:56243` dans votre navigateur pour voir le tableau de bord immediatement.

---

## Enregistrer une cle API

Il existe quatre methodes pour enregistrer des cles API. **Pour les debutants, la methode 1 (variables d'environnement) est recommandee.**

### Methode 1 : Variables d'environnement (recommandee — la plus simple)

Les variables d'environnement sont des **valeurs predefinies** qu'un programme lit au demarrage. Tapez simplement ceci dans votre terminal :

```bash
# Enregistrer une cle Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Enregistrer une cle OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Demarrer apres l'enregistrement
wall-vault start
```

Si vous avez plusieurs cles, separez-les par des virgules. wall-vault les utilisera en rotation automatique (round-robin) :

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Astuce** : la commande `export` ne s'applique qu'a la session de terminal en cours. Pour la rendre persistante apres un redemarrage, ajoutez la ligne dans votre fichier `~/.bashrc` ou `~/.zshrc`.

### Methode 2 : Interface du tableau de bord (clic de souris)

1. Ouvrez `http://localhost:56243` dans votre navigateur
2. Cliquez sur le bouton `[+ Ajouter]` dans la carte **🔑 Cles API** en haut
3. Saisissez le type de service, la valeur de la cle, un label (nom de memo), et la limite quotidienne, puis enregistrez

### Methode 3 : API REST (pour l'automatisation/les scripts)

L'API REST permet aux programmes d'echanger des donnees via HTTP. Utile pour l'enregistrement automatise par script.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer votre-token-admin" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Cle principale",
    "daily_limit": 1000
  }'
```

### Methode 4 : Drapeaux proxy (pour un test rapide)

Permet d'injecter temporairement une cle pour tester sans enregistrement formel. La cle disparait a l'arret du programme.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Utiliser le proxy

### Utilisation avec OpenClaw (usage principal)

Voici comment configurer OpenClaw pour se connecter aux services IA via wall-vault.

Ouvrez le fichier `~/.openclaw/openclaw.json` et ajoutez le contenu suivant :

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "votre-token-agent",   // token agent du coffre-fort
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // 1M de contexte gratuit
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Methode encore plus simple** : cliquez sur le bouton **🦞 Copier la config OpenClaw** sur la carte agent du tableau de bord — un snippet avec le token et l'adresse deja remplis est copie dans le presse-papiers. Il suffit de le coller.

**Ou le prefixe `wall-vault/` dans les noms de modeles mene-t-il ?**

wall-vault determine automatiquement a quel service IA envoyer la requete en fonction du nom du modele :

| Format du modele | Routage vers |
|-----------------|-------------|
| `wall-vault/gemini-*` | Google Gemini directement |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI directement |
| `wall-vault/claude-*` | Anthropic via OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (contexte gratuit de 1M tokens) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/nom-modele`, `openai/nom-modele`, `anthropic/nom-modele`, etc. | Directement au service concerne |
| `custom/google/nom-modele`, `custom/openai/nom-modele`, etc. | Supprime le prefixe `custom/` et reroute |
| `nom-modele:cloud` | Supprime le suffixe `:cloud` et route vers OpenRouter |

> 💡 **Qu'est-ce que le contexte ?** C'est la quantite de conversation qu'une IA peut retenir en une seule fois. 1M (un million de tokens) signifie qu'elle peut traiter de tres longues conversations ou documents en une seule session.

### Connexion directe au format Gemini API (compatibilite avec les outils existants)

Si vous avez des outils qui utilisent deja directement l'API Google Gemini, changez simplement l'adresse pour wall-vault :

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Ou si l'outil prend une URL directe :

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Utilisation avec le SDK OpenAI (Python)

Vous pouvez aussi connecter wall-vault a du code Python utilisant l'IA. Il suffit de changer le `base_url` :

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # wall-vault gere les cles API pour vous
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # au format provider/model
    messages=[{"role": "user", "content": "Bonjour"}]
)
```

### Changer de modele en cours d'execution

Pour changer de modele IA pendant que wall-vault est en cours d'execution :

```bash
# Changer le modele en envoyant une requete au proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# En mode distribue (multi-bots), changer sur le serveur coffre-fort → synchronisation SSE instantanee
curl -X PUT http://localhost:56243/admin/clients/mon-bot-id \
  -H "Authorization: Bearer votre-token-admin" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Consulter la liste des modeles disponibles

```bash
# Voir la liste complete
curl http://localhost:56244/api/models | python3 -m json.tool

# Voir uniquement les modeles Google
curl "http://localhost:56244/api/models?service=google"

# Rechercher par nom (ex : modeles contenant « claude »)
curl "http://localhost:56244/api/models?q=claude"
```

**Principaux modeles par service :**

| Service | Modeles principaux |
|---------|-------------------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ (Hunter Alpha contexte 1M gratuit, DeepSeek R1/V3, Qwen 2.5, etc.) |
| Ollama | Detection automatique des modeles installes localement |
| LM Studio | Serveur local (port 1234) |
| vLLM | Serveur local (port 8000) |

---

## Tableau de bord du coffre-fort

Ouvrez `http://localhost:56243` dans votre navigateur pour acceder au tableau de bord.

**Disposition :**
- **Barre superieure (fixe)** : logo, selecteurs de langue/theme, indicateur de connexion SSE
- **Grille de cartes** : cartes agents, services et cles API disposees en mosaique

### Cartes Cles API

Ces cartes permettent de gerer vos cles API enregistrees en un coup d'oeil.

- Les cles sont regroupees par service.
- `today_usage` : nombre de tokens (unites de texte que l'IA lit/ecrit) traites avec succes aujourd'hui
- `today_attempts` : nombre total d'appels aujourd'hui (succes + echecs)
- Utilisez le bouton `[+ Ajouter]` pour enregistrer de nouvelles cles et `✕` pour les supprimer.

> 💡 **Qu'est-ce qu'un token ?** C'est l'unite utilisee par l'IA pour traiter le texte. Environ un mot anglais, ou 1 a 2 caracteres francais. Les tarifs API sont generalement calcules en fonction du nombre de tokens.

### Cartes Agents

Ces cartes montrent l'etat des bots (agents) connectes au proxy wall-vault.

**L'etat de connexion est affiche sur 4 niveaux :**

| Indicateur | Etat | Signification |
|-----------|------|---------------|
| 🟢 | En cours d'execution | Le proxy fonctionne normalement |
| 🟡 | Retarde | Repond mais lentement |
| 🔴 | Hors ligne | Le proxy ne repond pas |
| ⚫ | Non connecte / Desactive | Le proxy ne s'est jamais connecte au coffre-fort, ou est desactive |

**Boutons en bas des cartes agents :**

Quand vous enregistrez un agent avec un **type d'agent** specifique, des boutons pratiques adaptes a ce type apparaissent automatiquement.

---

#### 🔘 Bouton Copier la configuration — genere automatiquement les parametres de connexion

En cliquant sur ce bouton, un snippet de configuration avec le token, l'adresse du proxy et les informations du modele deja remplis est copie dans le presse-papiers. Collez simplement le contenu a l'emplacement indique dans le tableau ci-dessous pour finaliser la configuration.

| Bouton | Type d'agent | Ou coller |
|--------|-------------|-----------|
| 🦞 Copier config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Copier config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Copier config Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Copier config Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Copier config VSCode | `vscode` | `~/.continue/config.json` |

**Exemple — Pour le type Claude Code, voici ce qui est copie :**

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
# ~/.continue/config.yaml  ← coller dans config.yaml, PAS config.json
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

> ⚠️ **La derniere version de Continue utilise `config.yaml`.** Si `config.yaml` existe, `config.json` est completement ignore. Assurez-vous de coller dans `config.yaml`.

**Exemple — Pour le type Cursor :**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : token-de-cet-agent

// Ou en variables d'environnement :
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=token-de-cet-agent
```

> ⚠️ **Si la copie dans le presse-papiers ne fonctionne pas** : les politiques de securite du navigateur peuvent bloquer la copie. Si une fenetre popup avec une zone de texte apparait, selectionnez tout avec Ctrl+A puis copiez avec Ctrl+C.

---

#### ⚡ Bouton Application automatique — un clic et c'est configure

Pour les agents de type `cline`, `claude-code`, `openclaw` ou `nanoclaw`, la carte agent affiche un bouton **⚡ Appliquer la config**. Cliquer sur ce bouton met automatiquement a jour le fichier de configuration local de l'agent.

| Bouton | Type d'agent | Fichier cible |
|--------|-------------|---------------|
| ⚡ Appliquer config Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Appliquer config Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Appliquer config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Appliquer config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Ce bouton envoie une requete a **localhost:56244** (proxy local). Le proxy doit etre en cours d'execution sur cette machine pour que cela fonctionne.

---

#### 🔀 Tri des cartes par glisser-deposer (v0.1.17)

Vous pouvez **glisser** les cartes agents du tableau de bord pour les reorganiser dans l'ordre souhaite.

1. Attrapez une carte agent avec la souris et faites-la glisser
2. Deposez-la sur une autre carte pour echanger leurs positions
3. Le nouvel ordre est **immediatement sauvegarde sur le serveur** et persiste apres un rafraichissement

> 💡 Les appareils tactiles (mobile/tablette) ne sont pas encore pris en charge. Utilisez un navigateur de bureau.

---

#### 🔄 Synchronisation bidirectionnelle des modeles (v0.1.16)

Quand vous changez le modele d'un agent dans le tableau de bord du coffre-fort, la configuration locale de l'agent est automatiquement mise a jour.

**Pour Cline :**
- Changement de modele dans le coffre-fort → evenement SSE → le proxy met a jour le champ modele dans `globalState.json`
- Champs mis a jour : `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` et la cle API ne sont pas touches
- **Un rechargement de VS Code est necessaire (`Ctrl+Alt+R` ou `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - Car Cline ne relit pas le fichier de configuration pendant l'execution

**Pour Claude Code :**
- Changement de modele dans le coffre-fort → evenement SSE → le proxy met a jour le champ `model` dans `settings.json`
- Recherche automatique des chemins WSL et Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Direction inverse (agent → coffre-fort) :**
- Quand un agent (Cline, Claude Code, etc.) envoie une requete au proxy, celui-ci inclut les informations service/modele du client dans le heartbeat
- La carte agent du tableau de bord affiche le service/modele actuellement utilise en temps reel

> 💡 **Point cle** : le proxy identifie les agents par leur token d'Authorization dans les requetes, et les redirige automatiquement vers le service/modele configure dans le coffre-fort. Meme si Cline ou Claude Code envoie un nom de modele different, le proxy le remplace par la configuration du coffre-fort.

---

### Utiliser Cline avec VS Code — Guide detaille

#### Etape 1 : Installer Cline

Installez **Cline** (ID : `saoudrizwan.claude-dev`) depuis le marketplace Extensions de VS Code.

#### Etape 2 : Enregistrer l'agent dans le coffre-fort

1. Ouvrez le tableau de bord du coffre-fort (`http://IP-coffre-fort:56243`)
2. Cliquez sur **+ Ajouter** dans la section **Agents**
3. Remplissez comme suit :

| Champ | Valeur | Description |
|-------|--------|-------------|
| ID | `mon_cline` | Identifiant unique (alphanumerique, sans espaces) |
| Nom | `Mon Cline` | Nom affiche dans le tableau de bord |
| Type d'agent | `cline` | ← Choisir imperativement `cline` |
| Service | Selectionnez le service a utiliser (ex : `google`) | |
| Modele | Saisissez le modele a utiliser (ex : `gemini-2.5-flash`) | |

4. Cliquez sur **Enregistrer** — un token est genere automatiquement

#### Etape 3 : Connecter Cline

**Methode A — Application automatique (recommandee)**

1. Verifiez que le **proxy** wall-vault est en cours d'execution sur cette machine (`localhost:56244`)
2. Cliquez sur le bouton **⚡ Appliquer config Cline** sur la carte agent du tableau de bord
3. Si la notification « Configuration appliquee avec succes ! » apparait, c'est bon
4. Rechargez VS Code (`Ctrl+Alt+R`)

**Methode B — Configuration manuelle**

Ouvrez les Parametres (⚙️) dans la barre laterale Cline :
- **API Provider** : `OpenAI Compatible`
- **Base URL** : `http://adresse-proxy:56244/v1`
  - Meme machine : `http://localhost:56244/v1`
  - Autre machine (ex : Mac Mini) : `http://192.168.1.20:56244/v1`
- **API Key** : le token delivre par le coffre-fort (copiez-le depuis la carte agent)
- **Model ID** : le modele configure dans le coffre-fort (ex : `gemini-2.5-flash`)

#### Etape 4 : Verification

Envoyez n'importe quel message dans la fenetre de chat Cline. Si tout fonctionne :
- La carte agent dans le tableau de bord affiche un **point vert (● En cours d'execution)**
- La carte montre le service/modele en cours (ex : `google / gemini-2.5-flash`)

#### Changer de modele

Pour changer le modele de Cline, faites-le depuis le **tableau de bord du coffre-fort** :

1. Changez le menu deroulant service/modele sur la carte agent
2. Cliquez sur **Appliquer**
3. Rechargez VS Code (`Ctrl+Alt+R`) — le nom du modele dans le pied de page de Cline sera mis a jour
4. Le nouveau modele sera utilise a partir de la prochaine requete

> 💡 En pratique, le proxy identifie les requetes de Cline par le token et les dirige vers le modele configure dans le coffre-fort. Meme sans rechargement de VS Code, **le modele reel utilise change immediatement** — le rechargement sert uniquement a mettre a jour l'affichage du modele dans l'interface de Cline.

#### Detection de deconnexion

Quand VS Code est ferme, la carte agent du tableau de bord passe au jaune (retarde) apres environ **90 secondes**, et au rouge (hors ligne) apres **3 minutes**. (A partir de v0.1.18, la detection hors ligne est plus rapide grace a des verifications d'etat toutes les 15 secondes.)

#### Resolution des problemes

| Symptome | Cause | Solution |
|----------|-------|----------|
| Erreur « connexion echouee » dans Cline | Proxy non demarre ou mauvaise adresse | Verifier le proxy avec `curl http://localhost:56244/health` |
| Le point vert n'apparait pas dans le coffre-fort | Cle API (token) non configuree | Cliquez a nouveau sur le bouton **⚡ Appliquer config Cline** |
| Le modele dans le pied de page de Cline ne change pas | Cline met en cache les parametres | Rechargez VS Code (`Ctrl+Alt+R`) |
| Mauvais nom de modele affiche | Ancien bug (corrige en v0.1.16) | Mettez a jour le proxy vers v0.1.16 ou plus recent |

---

#### 🟣 Bouton Copier la commande de deploiement — pour l'installation sur une nouvelle machine

Utilisez ce bouton lors de la premiere installation du proxy wall-vault sur un nouvel ordinateur et sa connexion au coffre-fort. Cliquer sur le bouton copie l'integralite du script d'installation. Collez-le dans le terminal du nouvel ordinateur et executez-le pour effectuer tout ceci en une seule fois :

1. Installation du binaire wall-vault (ignore si deja installe)
2. Enregistrement automatique du service utilisateur systemd
3. Demarrage du service et connexion automatique au coffre-fort

> 💡 Le script contient deja le token de cet agent et l'adresse du serveur coffre-fort, vous pouvez donc l'executer directement apres le collage sans modification.

---

### Cartes Services

Ces cartes permettent d'activer/desactiver et de configurer les services IA.

- Bouton bascule pour activer/desactiver chaque service
- Saisissez l'adresse d'un serveur IA local (Ollama, LM Studio, vLLM, etc. tournant sur votre machine) pour decouvrir automatiquement les modeles disponibles
- **Indicateur de connexion du service local** : un point ● a cote du nom du service est **vert** si connecte, **gris** sinon
- **Signal automatique des services locaux** (v0.1.23+) : les services locaux (Ollama, LM Studio, vLLM) sont automatiquement actives/desactives selon leur disponibilite. Quand un service devient accessible, le point ● passe au vert et la case est cochee en moins de 15 secondes ; quand le service tombe, il est automatiquement desactive. Cela fonctionne de la meme maniere que la bascule automatique des services cloud (Google, OpenRouter, etc.) en fonction de la disponibilite des cles API.

> 💡 **Si le service local tourne sur un autre ordinateur** : saisissez l'IP de cet ordinateur dans le champ URL du service. Exemple : `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio). Si le service n'est lie qu'a `127.0.0.1` au lieu de `0.0.0.0`, l'acces via une IP externe ne fonctionnera pas — verifiez l'adresse de liaison dans les parametres du service.

### Saisie du token administrateur

Lorsque vous essayez d'utiliser des fonctions importantes comme l'ajout ou la suppression de cles dans le tableau de bord, une fenetre de saisie du token administrateur apparait. Entrez le token que vous avez defini lors de l'assistant de configuration. Une fois saisi, il reste valide jusqu'a la fermeture du navigateur.

> ⚠️ **Si l'authentification echoue plus de 10 fois en 15 minutes, cette IP sera temporairement bloquee.** Si vous avez oublie votre token, consultez le champ `admin_token` dans le fichier `wall-vault.yaml`.

---

## Mode distribue (multi-bots)

Lorsque vous faites tourner OpenClaw sur plusieurs ordinateurs simultanement, vous pouvez **partager un seul coffre-fort de cles**. C'est pratique car vous n'avez a gerer les cles qu'a un seul endroit.

### Exemple de configuration

```
[Serveur Coffre-fort]
  wall-vault vault    (coffre-fort :56243, tableau de bord)

[WSL Alpha]           [Raspberry Pi Gamma]    [Mac Mini Local]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ synchro SSE         ↕ synchro SSE           ↕ synchro SSE
```

Tous les bots pointent vers le serveur coffre-fort central, de sorte que toute modification de modele ou ajout de cle dans le coffre-fort est immediatement repercutee sur tous les bots.

### Etape 1 : Demarrer le serveur coffre-fort

Sur l'ordinateur qui servira de serveur coffre-fort :

```bash
wall-vault vault
```

### Etape 2 : Enregistrer chaque bot (client)

Enregistrez les informations de chaque bot qui se connectera au serveur coffre-fort :

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer votre-token-admin" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "Bot A",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### Etape 3 : Demarrer le proxy sur chaque machine bot

Sur chaque ordinateur ou un bot est installe, lancez le proxy avec l'adresse du serveur coffre-fort et le token :

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Remplacez **`192.168.x.x`** par l'adresse IP interne reelle de l'ordinateur serveur coffre-fort. Vous pouvez la trouver dans les parametres de votre routeur ou avec la commande `ip addr`.

---

## Demarrage automatique

S'il est fastidieux de demarrer manuellement wall-vault a chaque redemarrage de votre ordinateur, enregistrez-le comme service systeme. Une fois enregistre, il demarre automatiquement au demarrage.

### Linux — systemd (la plupart des distributions Linux)

systemd est le systeme qui demarre et gere automatiquement les programmes sous Linux :

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Consulter les logs :

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

Le systeme responsable du demarrage automatique des programmes sous macOS :

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Telechargez NSSM depuis [nssm.cc](https://nssm.cc/download) et ajoutez-le au PATH.
2. Dans un PowerShell en mode administrateur :

```powershell
wall-vault doctor deploy windows
```

---

## Doctor — l'outil de diagnostic

La commande `doctor` est un outil qui **diagnostique et repare** automatiquement la configuration de wall-vault.

```bash
wall-vault doctor check   # Diagnostiquer l'etat actuel (lecture seule, ne modifie rien)
wall-vault doctor fix     # Reparer automatiquement les problemes
wall-vault doctor all     # Diagnostic + reparation automatique en une seule etape
```

> 💡 Si quelque chose semble anormal, essayez d'abord `wall-vault doctor all`. Il detecte et corrige automatiquement de nombreux problemes.

---

## Reference des variables d'environnement

Les variables d'environnement permettent de transmettre des valeurs de configuration a un programme. Saisissez-les dans le terminal au format `export VARIABLE=valeur`, ou ajoutez-les dans votre fichier de service de demarrage automatique pour une application permanente.

| Variable | Description | Valeur d'exemple |
|----------|-------------|-----------------|
| `WV_LANG` | Langue du tableau de bord | `ko`, `en`, `ja` |
| `WV_THEME` | Theme du tableau de bord | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Cle API Google (separees par des virgules pour plusieurs) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Cle API OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Adresse du serveur coffre-fort en mode distribue | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Token d'authentification client (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Token administrateur | `admin-token-here` |
| `WV_MASTER_PASS` | Mot de passe de chiffrement des cles API | `my-password` |
| `WV_AVATAR` | Chemin du fichier image avatar (relatif a `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Adresse du serveur local Ollama | `http://192.168.x.x:11434` |

---

## Resolution des problemes

### Le proxy ne demarre pas

Le port est souvent deja utilise par un autre programme.

```bash
ss -tlnp | grep 56244   # Verifier qui utilise le port 56244
wall-vault proxy --port 8080   # Demarrer sur un port different
```

### Erreurs de cle API (429, 402, 401, 403, 582)

| Code d'erreur | Signification | Que faire |
|--------------|---------------|-----------|
| **429** | Trop de requetes (quota depasse) | Attendre un peu ou ajouter d'autres cles |
| **402** | Paiement requis ou credits epuises | Recharger les credits sur le service concerne |
| **401 / 403** | Cle invalide ou pas d'autorisation | Reverifier la valeur de la cle et la re-enregistrer |
| **582** | Surcharge de la passerelle (cooldown de 5 minutes) | Se resout automatiquement apres 5 minutes |

```bash
# Verifier la liste des cles enregistrees et leur etat
curl -H "Authorization: Bearer votre-token-admin" http://localhost:56243/admin/keys

# Reinitialiser les compteurs d'utilisation des cles
curl -X POST -H "Authorization: Bearer votre-token-admin" http://localhost:56243/admin/keys/reset
```

### L'agent s'affiche comme « Non connecte »

« Non connecte » signifie que le processus proxy n'envoie pas de signal (heartbeat) au coffre-fort. **Cela ne signifie pas que la configuration n'a pas ete sauvegardee.** Le proxy doit etre lance avec l'adresse du serveur coffre-fort et le token pour etablir une connexion.

```bash
# Demarrer le proxy avec l'adresse du coffre-fort, le token et l'ID client
WV_VAULT_URL=http://serveur-coffre:56243 \
WV_VAULT_TOKEN=token-client \
WV_VAULT_CLIENT_ID=id-client \
wall-vault proxy
```

Une fois connecte, le tableau de bord affichera 🟢 En cours d'execution dans les 20 secondes environ.

### Problemes de connexion Ollama

Ollama est un programme qui execute l'IA directement sur votre ordinateur. Verifiez d'abord qu'Ollama est en cours d'execution.

```bash
curl http://localhost:11434/api/tags   # Si une liste de modeles apparait, tout fonctionne
export OLLAMA_URL=http://192.168.x.x:11434   # Si Ollama tourne sur un autre ordinateur
```

> ⚠️ Si Ollama ne repond pas, demarrez-le d'abord avec `ollama serve`.

> ⚠️ **Les gros modeles sont lents a repondre** : les modeles volumineux comme `qwen3.5:35b` ou `deepseek-r1` peuvent mettre plusieurs minutes a generer une reponse. Meme si rien ne semble se passer, le traitement est peut-etre en cours — soyez patient.

---

## Changements recents (v0.1.16 ~ v0.1.23)

### v0.1.23 (2026-04-06)
- **Correction du changement de modele Ollama** : correction d'un probleme ou le changement du modele Ollama dans le tableau de bord du coffre-fort n'etait pas repercute dans le proxy. Auparavant, seule la variable d'environnement (`OLLAMA_MODEL`) etait utilisee, mais desormais les parametres du coffre-fort sont prioritaires.
- **Signal automatique des services locaux** : Ollama, LM Studio et vLLM sont automatiquement actives lorsqu'ils sont accessibles et desactives lorsqu'ils ne le sont pas. Fonctionne de la meme maniere que la bascule automatique basee sur les cles pour les services cloud.

### v0.1.22 (2026-04-05)
- **Correction du champ content vide** : quand les modeles de reflexion (gemini-3.1-pro, o1, claude thinking, etc.) utilisent tous les max_tokens pour le raisonnement et ne peuvent pas produire de reponse effective, le proxy omettait les champs `content`/`text` du JSON de reponse via `omitempty`, ce qui faisait planter les clients SDK OpenAI/Anthropic avec `Cannot read properties of undefined (reading 'trim')`. Corrige pour toujours inclure les champs conformement a la specification officielle de l'API.

### v0.1.21 (2026-04-05)
- **Support des modeles Gemma 4** : les modeles de la famille Gemma comme `gemma-4-31b-it` et `gemma-4-26b-a4b-it` peuvent desormais etre utilises via l'API Google Gemini.
- **Support des services LM Studio / vLLM** : auparavant, ces services etaient absents du routage proxy et retombaient toujours sur Ollama. Desormais correctement routes via l'API compatible OpenAI.
- **Correction de l'affichage des services dans le tableau de bord** : meme en cas de fallback, le tableau de bord affiche toujours le service configure par l'utilisateur.
- **Affichage de l'etat des services locaux** : lors du chargement du tableau de bord, l'etat de connexion des services locaux (Ollama, LM Studio, vLLM, etc.) est indique par la couleur du point ●.
- **Variable d'environnement pour le filtre d'outils** : utilisez `WV_TOOL_FILTER=passthrough` pour definir le mode de transmission des outils.

### v0.1.20 (2026-03-28)
- **Renforcement complet de la securite** : prevention XSS (41 emplacements), comparaison de tokens en temps constant, restrictions CORS, limites de taille de requete, prevention de traversee de chemin, authentification SSE, renforcement du limiteur de debit et 12 autres ameliorations de securite.

### v0.1.19 (2026-03-27)
- **Detection en ligne de Claude Code** : les instances Claude Code ne passant pas par le proxy sont desormais affichees comme en ligne dans le tableau de bord.

### v0.1.18 (2026-03-26)
- **Correction du collage du service de fallback** : apres une erreur temporaire causant un fallback vers Ollama, le retour au service d'origine se fait automatiquement lorsqu'il est retabli.
- **Amelioration de la detection hors ligne** : des verifications d'etat toutes les 15 secondes permettent une detection plus rapide des pannes de proxy.

### v0.1.17 (2026-03-25)
- **Tri des cartes par glisser-deposer** : les cartes agents peuvent etre reorganisees par glisser-deposer.
- **Bouton d'application de configuration en ligne** : le bouton [⚡ Appliquer la config] est affiche sur les cartes agents hors ligne.
- **Type d'agent cokacdir ajoute**.

### v0.1.16 (2026-03-25)
- **Synchronisation bidirectionnelle des modeles** : le changement d'un modele Cline ou Claude Code dans le tableau de bord du coffre-fort est automatiquement repercute.

---

*Pour des informations API plus detaillees, consultez [API.md](API.md).*
