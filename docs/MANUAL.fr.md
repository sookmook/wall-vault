# Manuel Utilisateur wall-vault
*(Last updated: 2026-04-06 — v0.1.24)*

---

## Table des matieres

1. [Qu'est-ce que wall-vault ?](#quest-ce-que-wall-vault)
2. [Installation](#installation)
3. [Premiers pas (assistant de configuration)](#premiers-pas)
4. [Enregistrer des cles API](#enregistrer-des-cles-api)
5. [Utilisation du proxy](#utilisation-du-proxy)
6. [Tableau de bord du coffre-fort](#tableau-de-bord-du-coffre-fort)
7. [Mode distribue (multi-bot)](#mode-distribue-multi-bot)
8. [Configuration du demarrage automatique](#configuration-du-demarrage-automatique)
9. [Doctor — Outil d'auto-diagnostic](#doctor--outil-dauto-diagnostic)
10. [RTK Economie de tokens](#rtk-economie-de-tokens)
11. [Reference des variables d'environnement](#reference-des-variables-denvironnement)
12. [Depannage](#depannage)

---

## Qu'est-ce que wall-vault ?

**wall-vault = un proxy AI + coffre-fort de cles API pour OpenClaw**

Pour utiliser les services d'IA, vous avez besoin de **cles API** — considerez-les comme un **badge numerique** qui prouve que vous etes autorise a utiliser un service particulier. Ces badges ont des limites d'utilisation quotidiennes et peuvent etre compromis s'ils sont mal geres.

wall-vault conserve vos badges en securite dans un coffre-fort chiffre et agit comme un **proxy (intermediaire)** entre OpenClaw et les services d'IA. En bref, OpenClaw n'a besoin de communiquer qu'avec wall-vault — wall-vault gere toute la complexite en coulisses.

Voici ce que wall-vault prend en charge pour vous :

- **Rotation automatique des cles** : Quand une cle atteint sa limite ou est temporairement bloquee (refroidissement), wall-vault passe silencieusement a la cle suivante. OpenClaw continue de fonctionner sans interruption.
- **Basculement automatique de service (fallback)** : Si Google ne repond pas, il bascule vers OpenRouter. Si cela echoue aussi, il passe automatiquement a Ollama, LM Studio ou vLLM (IA locale) sur votre machine. Votre session n'est jamais interrompue. Quand le service original se retablit, il revient automatiquement a partir de la requete suivante (v0.1.18+, LM Studio/vLLM : v0.1.21+).
- **Synchronisation en temps reel (SSE)** : Changez le modele dans le tableau de bord du coffre-fort, et OpenClaw le refllete en 1 a 3 secondes. SSE (Server-Sent Events) est une technologie ou le serveur envoie des mises a jour aux clients en temps reel.
- **Notifications en temps reel** : Les evenements comme l'epuisement des cles ou les pannes de service apparaissent immediatement en bas de l'interface TUI (terminal) d'OpenClaw.

> 💡 **Claude Code, Cursor et VS Code** peuvent egalement etre connectes, mais l'objectif principal de wall-vault est de fonctionner avec OpenClaw.

```
OpenClaw (terminal TUI)
        │
        ▼
  wall-vault proxy (:56244)   ← gestion des cles, routage, fallback, evenements
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (340+ modeles)
        ├─ Ollama / LM Studio / vLLM (machine locale, dernier recours)
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
# Telecharger
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Ajouter au PATH (prend effet apres le redemarrage de PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **Qu'est-ce que le PATH ?** C'est une liste de dossiers ou votre ordinateur cherche les commandes. Ajouter wall-vault au PATH vous permet d'executer `wall-vault` depuis n'importe quel repertoire.

### Compilation depuis les sources (pour les developpeurs)

Ceci ne s'applique que si vous avez un environnement de developpement Go installe.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (version : v0.1.24.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Version avec horodatage de compilation** : Lorsque vous compilez avec `make build`, la version est automatiquement generee dans un format comme `v0.1.24.20260406.225957` incluant la date et l'heure. Si vous compilez directement avec `go build ./...`, la version affichera simplement `"dev"`.

---

## Premiers pas

### Executer l'assistant de configuration

Apres l'installation, assurez-vous d'executer d'abord l'**assistant de configuration**. L'assistant vous guidera etape par etape en demandant les informations necessaires.

```bash
wall-vault setup
```

Voici les etapes de l'assistant :

```
1. Selection de la langue (10 langues dont le francais)
2. Selection du theme (light / dark / gold / cherry / ocean)
3. Mode de fonctionnement — autonome (utilisateur unique) ou distribue (plusieurs machines)
4. Nom du bot — le nom affiche sur le tableau de bord
5. Configuration des ports — par defaut : proxy 56244, coffre-fort 56243 (appuyez sur Entree pour garder les valeurs par defaut)
6. Selection des services IA — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Configuration du filtre de securite des outils
8. Token administrateur — un mot de passe pour verrouiller les fonctions d'administration du tableau de bord ; peut etre genere automatiquement
9. Mot de passe de chiffrement des cles API — pour un stockage plus securise des cles (optionnel)
10. Emplacement de sauvegarde du fichier de configuration
```

> ⚠️ **Assurez-vous de retenir votre token administrateur.** Vous en aurez besoin plus tard pour ajouter des cles ou modifier les parametres du tableau de bord. Si vous l'oubliez, vous devrez modifier manuellement le fichier de configuration.

Une fois l'assistant termine, un fichier de configuration `wall-vault.yaml` est automatiquement cree.

### Demarrage

```bash
wall-vault start
```

Deux serveurs demarrent simultanement :

- **Proxy** (`http://localhost:56244`) — l'intermediaire connectant OpenClaw aux services d'IA
- **Coffre-fort de cles** (`http://localhost:56243`) — gestion des cles API et tableau de bord web

Ouvrez `http://localhost:56243` dans votre navigateur pour voir le tableau de bord immediatement.

---

## Enregistrer des cles API

Il existe quatre methodes pour enregistrer des cles API. **Pour les debutants, la methode 1 (variables d'environnement) est recommandee.**

### Methode 1 : Variables d'environnement (recommandee — la plus simple)

Les variables d'environnement sont des **valeurs predefinies** qu'un programme lit au demarrage. Tapez simplement ce qui suit dans votre terminal :

```bash
# Enregistrer une cle Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Enregistrer une cle OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Demarrer apres l'enregistrement
wall-vault start
```

Si vous avez plusieurs cles, separez-les par des virgules. wall-vault les utilisera en rotation automatiquement (round-robin) :

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Conseil** : La commande `export` ne s'applique qu'a la session terminal en cours. Pour la persister apres les redemarrages, ajoutez la ligne a votre fichier `~/.bashrc` ou `~/.zshrc`.

### Methode 2 : Interface du tableau de bord (pointer et cliquer)

1. Ouvrez `http://localhost:56243` dans votre navigateur
2. Cliquez sur le bouton `[+ Ajouter]` dans la carte superieure **🔑 Cles API**
3. Entrez le type de service, la valeur de la cle, un libelle (nom memo) et la limite quotidienne, puis sauvegardez

### Methode 3 : API REST (pour l'automatisation/les scripts)

L'API REST est un moyen pour les programmes d'echanger des donnees via HTTP. Utile pour l'enregistrement automatise par scripts.

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

### Methode 4 : Options du proxy (pour des tests rapides)

Utilisez ceci pour injecter temporairement une cle pour les tests sans enregistrement formel. La cle disparait quand le programme est arrete.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Utilisation du proxy

### Utilisation avec OpenClaw (objectif principal)

Voici comment configurer OpenClaw pour se connecter aux services d'IA via wall-vault.

Ouvrez `~/.openclaw/openclaw.json` et ajoutez ce qui suit :

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

> 💡 **Methode plus facile** : Cliquez sur le bouton **🦞 Copier la configuration OpenClaw** sur la carte d'agent dans le tableau de bord — il copie un extrait avec le token et l'adresse deja remplis. Il suffit de coller.

**Ou le prefixe `wall-vault/` dans les noms de modeles redirige-t-il ?**

wall-vault determine automatiquement vers quel service d'IA envoyer la requete en fonction du nom du modele :

| Format du modele | Redirige vers |
|-----------------|--------------|
| `wall-vault/gemini-*` | Google Gemini direct |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI direct |
| `wall-vault/claude-*` | Anthropic via OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (contexte 1M tokens gratuit) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/nom-modele`, `openai/nom-modele`, `anthropic/nom-modele`, etc. | Direct vers ce service |
| `custom/google/nom-modele`, `custom/openai/nom-modele`, etc. | Supprime le prefixe `custom/` et redirige |
| `nom-modele:cloud` | Supprime le suffixe `:cloud` et redirige vers OpenRouter |

> 💡 **Qu'est-ce que le contexte ?** C'est la quantite de conversation qu'une IA peut retenir a la fois. 1M (un million de tokens) signifie qu'elle peut traiter de tres longues conversations ou documents en une seule session.

### Format direct de l'API Gemini (pour la compatibilite avec les outils existants)

Si vous avez des outils qui utilisent deja directement l'API Google Gemini, changez simplement l'adresse vers wall-vault :

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Ou si l'outil prend une URL directe :

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Utilisation avec le SDK OpenAI (Python)

Vous pouvez egalement connecter wall-vault au code Python qui utilise l'IA. Changez simplement la `base_url` :

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # wall-vault gere les cles API pour vous
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # utilisez le format provider/model
    messages=[{"role": "user", "content": "Bonjour"}]
)
```

### Changer de modele en cours d'execution

Pour changer le modele d'IA pendant que wall-vault est deja en cours d'execution :

```bash
# Changer le modele en envoyant une requete au proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# En mode distribue (multi-bot), changer sur le serveur coffre-fort → synchronise instantanement via SSE
curl -X PUT http://localhost:56243/admin/clients/mon-bot-id \
  -H "Authorization: Bearer votre-token-admin" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Verifier les modeles disponibles

```bash
# Voir la liste complete
curl http://localhost:56244/api/models | python3 -m json.tool

# Voir uniquement les modeles Google
curl "http://localhost:56244/api/models?service=google"

# Rechercher par nom (ex : modeles contenant "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Modeles cles par service :**

| Service | Modeles cles |
|---------|-------------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ (Hunter Alpha contexte 1M gratuit, DeepSeek R1/V3, Qwen 2.5, etc.) |
| Ollama | Detection automatique des modeles installes localement |
| LM Studio | Serveur local (port 1234) |
| vLLM | Serveur local (port 8000) |

---

## Tableau de bord du coffre-fort

Ouvrez `http://localhost:56243` dans votre navigateur pour voir le tableau de bord.

**Disposition :**
- **Barre superieure (fixe)** : Logo, selecteurs de langue/theme, indicateur de connexion SSE
- **Grille de cartes** : Cartes d'agents, de services et de cles API disposees en tuiles

### Cartes de cles API

Ces cartes offrent une vue d'ensemble de vos cles API enregistrees.

- Les cles sont organisees par service.
- `today_usage` : Nombre de tokens (unites de texte lues/ecrites par l'IA) traites avec succes aujourd'hui
- `today_attempts` : Nombre total d'appels aujourd'hui (reussis + echoues)
- Utilisez le bouton `[+ Ajouter]` pour enregistrer de nouvelles cles et `✕` pour les supprimer.

> 💡 **Qu'est-ce qu'un token ?** C'est l'unite utilisee par l'IA pour traiter le texte. Environ un mot anglais ou 1 a 2 caracteres francais. La tarification des API est generalement basee sur le nombre de tokens.

### Cartes d'agents

Ces cartes montrent le statut des bots (agents) connectes au proxy wall-vault.

**Le statut de connexion a 4 niveaux :**

| Indicateur | Statut | Signification |
|-----------|--------|--------------|
| 🟢 | En cours d'execution | Le proxy fonctionne normalement |
| 🟡 | Retarde | Repond mais lentement |
| 🔴 | Hors ligne | Le proxy ne repond pas |
| ⚫ | Non connecte / Desactive | Le proxy ne s'est jamais connecte au coffre-fort ou est desactive |

**Boutons en bas des cartes d'agents :**

Lorsque vous enregistrez un agent avec un **type d'agent** specifique, des boutons de commodite correspondants apparaissent automatiquement.

---

#### 🔘 Bouton Copier la configuration — genere automatiquement les parametres de connexion

Cliquer sur ce bouton copie un extrait de configuration dans le presse-papiers avec le token, l'adresse du proxy et les informations du modele de l'agent deja remplis. Collez-le simplement a l'emplacement indique dans le tableau ci-dessous pour completer la configuration de connexion.

| Bouton | Type d'agent | Ou coller |
|--------|-------------|-----------|
| 🦞 Copier la config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Copier la config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Copier la config Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Copier la config Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Copier la config VSCode | `vscode` | `~/.continue/config.json` |

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

> ⚠️ **Si la copie dans le presse-papiers ne fonctionne pas** : Les politiques de securite du navigateur peuvent bloquer la copie. Si un popup avec une zone de texte apparait, utilisez Ctrl+A pour tout selectionner, puis Ctrl+C pour copier.

---

#### ⚡ Bouton Application automatique — un clic et c'est fait

Pour les agents de type `cline`, `claude-code`, `openclaw` ou `nanoclaw`, la carte d'agent affiche un bouton **⚡ Appliquer la configuration**. Cliquer sur ce bouton met automatiquement a jour le fichier de configuration local de l'agent.

| Bouton | Type d'agent | Fichier cible |
|--------|-------------|--------------|
| ⚡ Appliquer la config Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Appliquer la config Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Appliquer la config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Appliquer la config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Ce bouton envoie une requete a **localhost:56244** (proxy local). Le proxy doit etre en cours d'execution sur cette machine pour que cela fonctionne.

---

#### 🔀 Tri des cartes par glisser-deposer (v0.1.17)

Vous pouvez **glisser** les cartes d'agents sur le tableau de bord pour les reorganiser dans l'ordre souhaite.

1. Saisissez une carte d'agent avec votre souris et faites-la glisser
2. Deposez-la sur une autre carte pour echanger les positions
3. Le nouvel ordre est **immediatement sauvegarde sur le serveur** et persiste apres rafraichissement

> 💡 Les appareils tactiles (mobile/tablette) ne sont pas encore supportes. Utilisez un navigateur de bureau.

---

#### 🔄 Synchronisation bidirectionnelle des modeles (v0.1.16)

Lorsque vous changez le modele d'un agent dans le tableau de bord du coffre-fort, la configuration locale de l'agent est automatiquement mise a jour.

**Pour Cline :**
- Changement de modele dans le coffre-fort → evenement SSE → le proxy met a jour le champ modele dans `globalState.json`
- Champs mis a jour : `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` et la cle API ne sont pas touches
- **Rechargement de VS Code requis (`Ctrl+Alt+R` ou `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - Parce que Cline ne relit pas le fichier de configuration pendant l'execution

**Pour Claude Code :**
- Changement de modele dans le coffre-fort → evenement SSE → le proxy met a jour le champ `model` dans `settings.json`
- Recherche automatiquement les chemins WSL et Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Direction inverse (agent → coffre-fort) :**
- Quand un agent (Cline, Claude Code, etc.) envoie une requete au proxy, le proxy inclut les informations service/modele de ce client dans le heartbeat
- La carte d'agent dans le tableau de bord du coffre-fort montre le service/modele actuellement utilise en temps reel

> 💡 **Point cle** : Le proxy identifie les agents par leur token d'Authorization dans les requetes et redirige automatiquement vers le service/modele configure dans le coffre-fort. Meme si Cline ou Claude Code envoie un nom de modele different, le proxy le remplace par la configuration du coffre-fort.

---

### Utiliser Cline avec VS Code — Guide detaille

#### Etape 1 : Installer Cline

Installez **Cline** (ID : `saoudrizwan.claude-dev`) depuis le Marketplace des extensions VS Code.

#### Etape 2 : Enregistrer l'agent dans le coffre-fort

1. Ouvrez le tableau de bord du coffre-fort (`http://IP-du-coffre:56243`)
2. Cliquez sur **+ Ajouter** dans la section **Agents**
3. Remplissez les champs suivants :

| Champ | Valeur | Description |
|-------|--------|-------------|
| ID | `my_cline` | Identifiant unique (alphanumerique, sans espaces) |
| Nom | `My Cline` | Nom affiche sur le tableau de bord |
| Type d'agent | `cline` | ← Doit selectionner `cline` |
| Service | Selectionner le service a utiliser (ex : `google`) | |
| Modele | Entrer le modele a utiliser (ex : `gemini-2.5-flash`) | |

4. Cliquez sur **Sauvegarder** — un token est automatiquement genere

#### Etape 3 : Connecter a Cline

**Methode A — Application automatique (recommandee)**

1. Assurez-vous que le **proxy** wall-vault fonctionne sur cette machine (`localhost:56244`)
2. Cliquez sur le bouton **⚡ Appliquer la config Cline** sur la carte d'agent dans le tableau de bord
3. Si la notification « Configuration appliquee avec succes ! » apparait, c'est reussi
4. Rechargez VS Code (`Ctrl+Alt+R`)

**Methode B — Configuration manuelle**

Ouvrez les Parametres (⚙️) dans la barre laterale de Cline :
- **API Provider** : `OpenAI Compatible`
- **Base URL** : `http://adresse-du-proxy:56244/v1`
  - Meme machine : `http://localhost:56244/v1`
  - Machine differente (ex : Mac Mini) : `http://192.168.1.20:56244/v1`
- **API Key** : Le token delivre par le coffre-fort (copier depuis la carte d'agent)
- **Model ID** : Le modele configure dans le coffre-fort (ex : `gemini-2.5-flash`)

#### Etape 4 : Verification

Envoyez n'importe quel message dans la fenetre de chat Cline. Si ca fonctionne :
- La carte d'agent dans le tableau de bord du coffre-fort montre un **point vert (● En cours d'execution)**
- La carte affiche le service/modele actuel (ex : `google / gemini-2.5-flash`)

#### Changer le modele

Quand vous voulez changer le modele de Cline, faites-le depuis le **tableau de bord du coffre-fort** :

1. Changez le menu deroulant service/modele sur la carte d'agent
2. Cliquez sur **Appliquer**
3. Rechargez VS Code (`Ctrl+Alt+R`) — le nom du modele dans le pied de page de Cline sera mis a jour
4. Le nouveau modele est utilise a partir de la requete suivante

> 💡 En pratique, le proxy identifie les requetes de Cline par le token et les redirige vers le modele configure dans le coffre-fort. Meme sans recharger VS Code, **le modele reellement utilise change immediatement** — le rechargement sert uniquement a mettre a jour l'affichage du modele dans l'interface de Cline.

#### Detection de la deconnexion

Quand VS Code est ferme, la carte d'agent dans le tableau de bord du coffre-fort devient jaune (retardee) apres environ **90 secondes** et rouge (hors ligne) apres **3 minutes**. (A partir de v0.1.18, la detection hors ligne est plus rapide grace aux verifications de statut toutes les 15 secondes.)

#### Depannage

| Symptome | Cause | Solution |
|----------|-------|----------|
| Erreur « Connexion echouee » dans Cline | Proxy non demarre ou mauvaise adresse | Verifier le proxy avec `curl http://localhost:56244/health` |
| Le point vert n'apparait pas dans le coffre-fort | Cle API (token) non configuree | Cliquer a nouveau sur le bouton **⚡ Appliquer la config Cline** |
| Le modele du pied de page Cline ne change pas | Cline met en cache les parametres | Recharger VS Code (`Ctrl+Alt+R`) |
| Mauvais nom de modele affiche | Ancien bug (corrige en v0.1.16) | Mettre a jour le proxy vers v0.1.16 ou plus |

---

#### 🟣 Bouton Copier la commande de deploiement — pour installer sur une nouvelle machine

Utilisez ceci lors de la premiere installation du proxy wall-vault sur un nouvel ordinateur et sa connexion au coffre-fort. Cliquer sur le bouton copie l'ensemble du script d'installation. Collez-le dans le terminal du nouvel ordinateur et executez-le pour :

1. Installer le binaire wall-vault (ignore si deja installe)
2. Enregistrer automatiquement un service utilisateur systemd
3. Demarrer le service et se connecter automatiquement au coffre-fort

> 💡 Le script contient deja le token et l'adresse du serveur coffre-fort de cet agent, vous pouvez donc l'executer immediatement apres le collage sans aucune modification.

---

### Cartes de services

Ces cartes vous permettent d'activer/desactiver et de configurer les services d'IA.

- Interrupteur pour activer/desactiver chaque service
- Entrez l'adresse d'un serveur d'IA local (Ollama, LM Studio, vLLM, etc. sur votre ordinateur) pour decouvrir automatiquement les modeles disponibles
- **Statut de connexion du service local** : Un point ● a cote du nom du service est **vert** si connecte, **gris** sinon
- **Signalisation automatique des services locaux** (v0.1.23+) : Les services locaux (Ollama, LM Studio, vLLM) sont automatiquement actives/desactives en fonction de la disponibilite de la connexion. Quand un service devient accessible, il passe en ● vert et la case a cocher s'active en 15 secondes ; quand le service tombe, il se desactive automatiquement. Cela fonctionne de la meme maniere que le basculement automatique des services cloud (Google, OpenRouter, etc.) base sur la disponibilite des cles API.

> 💡 **Si le service local fonctionne sur un autre ordinateur** : Entrez l'IP de cet ordinateur dans le champ URL du service. Exemple : `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio). Si le service est lie uniquement a `127.0.0.1` au lieu de `0.0.0.0`, l'acces par IP externe ne fonctionnera pas — verifiez le parametre d'adresse de liaison du service.

### Saisie du token administrateur

Lorsque vous essayez d'utiliser des fonctions importantes comme l'ajout ou la suppression de cles dans le tableau de bord, un popup de saisie du token administrateur apparait. Entrez le token que vous avez defini pendant l'assistant de configuration. Une fois saisi, il reste valide jusqu'a la fermeture du navigateur.

> ⚠️ **Si l'authentification echoue plus de 10 fois en 15 minutes, cette IP sera temporairement bloquee.** Si vous avez oublie votre token, verifiez le champ `admin_token` dans le fichier `wall-vault.yaml`.

---

## Mode distribue (multi-bot)

Lorsque vous executez OpenClaw simultanement sur plusieurs ordinateurs, vous pouvez **partager un seul coffre-fort de cles**. C'est pratique car vous n'avez a gerer les cles qu'a un seul endroit.

### Exemple de configuration

```
[Serveur coffre-fort de cles]
  wall-vault vault    (coffre-fort :56243, tableau de bord)

[WSL Alpha]           [Raspberry Pi Gamma]    [Mac Mini Local]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ sync SSE            ↕ sync SSE              ↕ sync SSE
```

Tous les bots pointent vers le serveur coffre-fort central. Quand vous changez un modele ou ajoutez une cle dans le coffre-fort, c'est immediatement refllete sur tous les bots.

### Etape 1 : Demarrer le serveur coffre-fort

Executez ceci sur l'ordinateur qui servira de serveur coffre-fort :

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

### Etape 3 : Demarrer le proxy sur chaque ordinateur bot

Sur chaque ordinateur ou un bot est installe, executez le proxy avec l'adresse du serveur coffre-fort et le token :

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Remplacez **`192.168.x.x`** par l'adresse IP interne reelle de l'ordinateur serveur coffre-fort. Vous pouvez la trouver via les parametres de votre routeur ou la commande `ip addr`.

---

## Configuration du demarrage automatique

S'il est fastidieux de demarrer manuellement wall-vault a chaque redemarrage de votre ordinateur, enregistrez-le comme service systeme. Une fois enregistre, il demarre automatiquement au demarrage.

### Linux — systemd (la plupart des distributions Linux)

systemd est le systeme qui demarre et gere automatiquement les programmes sous Linux :

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Verifier les logs :

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

1. Telechargez NSSM depuis [nssm.cc](https://nssm.cc/download) et ajoutez-le a votre PATH.
2. Dans un PowerShell administrateur :

```powershell
wall-vault doctor deploy windows
```

---

## Doctor — Outil d'auto-diagnostic

La commande `doctor` est un outil qui **diagnostique et repare** automatiquement la configuration de wall-vault.

```bash
wall-vault doctor check   # Diagnostiquer l'etat actuel (lecture seule, ne change rien)
wall-vault doctor fix     # Reparer automatiquement les problemes
wall-vault doctor all     # Diagnostiquer + reparer automatiquement en une etape
```

> 💡 Si quelque chose semble anormal, essayez d'abord d'executer `wall-vault doctor all`. Il detecte et corrige de nombreux problemes automatiquement.

---

## RTK Economie de tokens

*(v0.1.24+)*

**RTK (Outil d'economie de tokens)** compresse automatiquement la sortie des commandes shell executees par les agents de codage IA (comme Claude Code), reduisant ainsi l'utilisation de tokens. Par exemple, la sortie de 15 lignes de `git status` est condensee en un resume de 2 lignes.

### Utilisation basique

```bash
# Enveloppez les commandes avec wall-vault rtk pour filtrer automatiquement la sortie
wall-vault rtk git status          # affiche uniquement la liste des fichiers modifies
wall-vault rtk git diff HEAD~1     # lignes modifiees + contexte minimal uniquement
wall-vault rtk git log -10         # hash + message sur une ligne chacun
wall-vault rtk go test ./...       # affiche uniquement les tests echoues
wall-vault rtk ls -la              # les commandes non supportees sont automatiquement tronquees
```

### Commandes supportees et economies

| Commande | Methode de filtrage | Economie |
|----------|-------------------|----------|
| `git status` | Resume des fichiers modifies uniquement | ~87% |
| `git diff` | Lignes modifiees + 3 lignes de contexte | ~60-94% |
| `git log` | Hash + premiere ligne du message | ~90% |
| `git push/pull/fetch` | Suppression de la progression, resume uniquement | ~80% |
| `go test` | Afficher uniquement les echecs, compter les succes | ~88-99% |
| `go build/vet` | Afficher uniquement les erreurs | ~90% |
| Toutes les autres commandes | 50 premieres + 50 dernieres lignes, max 32Ko | Variable |

### Pipeline de filtrage en 3 etapes

1. **Filtre structurel par commande** — Comprend les formats de sortie de git, go, etc. et extrait uniquement les parties significatives
2. **Post-traitement par expressions regulieres** — Supprime les codes couleur ANSI, reduit les lignes vides, agregge les lignes en double
3. **Passage direct + troncature** — Les commandes non supportees ne conservent que les 50 premieres/dernieres lignes

### Integration avec Claude Code

Vous pouvez configurer un hook `PreToolUse` de Claude Code pour acheminer automatiquement toutes les commandes shell via RTK.

```bash
# Installer le hook (ajoute automatiquement au settings.json de Claude Code)
wall-vault rtk hook install
```

Ou ajouter manuellement a `~/.claude/settings.json` :

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

> 💡 **Conservation du code de sortie** : RTK retourne le code de sortie de la commande originale tel quel. Si une commande echoue (code de sortie ≠ 0), l'IA detecte correctement l'echec.

> 💡 **Sortie en anglais forcee** : RTK execute les commandes avec `LC_ALL=C`, produisant une sortie en anglais independamment des parametres de langue du systeme. Cela garantit le bon fonctionnement des filtres.

---

## Reference des variables d'environnement

Les variables d'environnement sont un moyen de transmettre des valeurs de configuration a un programme. Entrez-les dans le terminal avec `export VARIABLE=valeur`, ou ajoutez-les a votre fichier de service de demarrage automatique pour une application permanente.

| Variable | Description | Valeur d'exemple |
|----------|-------------|-----------------|
| `WV_LANG` | Langue du tableau de bord | `ko`, `en`, `ja` |
| `WV_THEME` | Theme du tableau de bord | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Cle API Google (separees par des virgules) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Cle API OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Adresse du serveur coffre-fort en mode distribue | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Token d'authentification client (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Token administrateur | `admin-token-here` |
| `WV_MASTER_PASS` | Mot de passe de chiffrement des cles API | `my-password` |
| `WV_AVATAR` | Chemin du fichier image d'avatar (relatif a `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Adresse du serveur local Ollama | `http://192.168.x.x:11434` |

---

## Depannage

### Le proxy ne demarre pas

Le port est souvent deja utilise par un autre programme.

```bash
ss -tlnp | grep 56244   # Verifier ce qui utilise le port 56244
wall-vault proxy --port 8080   # Demarrer sur un autre port
```

### Erreurs de cles API (429, 402, 401, 403, 582)

| Code d'erreur | Signification | Que faire |
|-------------|-------------|-----------|
| **429** | Trop de requetes (quota depasse) | Attendre un moment ou ajouter d'autres cles |
| **402** | Paiement requis ou credits epuises | Recharger les credits sur ce service |
| **401 / 403** | Cle invalide ou pas de permission | Reverifier la valeur de la cle et re-enregistrer |
| **582** | Surcharge de la passerelle (refroidissement de 5 minutes) | Se resout automatiquement apres 5 minutes |

```bash
# Verifier la liste des cles enregistrees et leur statut
curl -H "Authorization: Bearer votre-token-admin" http://localhost:56243/admin/keys

# Reinitialiser les compteurs d'utilisation des cles
curl -X POST -H "Authorization: Bearer votre-token-admin" http://localhost:56243/admin/keys/reset
```

### Agent affiche comme « Non connecte »

« Non connecte » signifie que le processus proxy n'envoie pas de heartbeats au coffre-fort. **Cela ne signifie pas que la configuration n'a pas ete sauvegardee.** Le proxy doit fonctionner avec l'adresse du serveur coffre-fort et le token pour etablir une connexion.

```bash
# Demarrer le proxy avec l'adresse du serveur coffre-fort, le token et l'ID client
WV_VAULT_URL=http://serveur-coffre:56243 \
WV_VAULT_TOKEN=token-client \
WV_VAULT_CLIENT_ID=id-client \
wall-vault proxy
```

Une fois connecte, le tableau de bord affichera 🟢 En cours d'execution dans environ 20 secondes.

### Problemes de connexion Ollama

Ollama est un programme qui execute l'IA directement sur votre ordinateur. D'abord, assurez-vous qu'Ollama fonctionne.

```bash
curl http://localhost:11434/api/tags   # Si une liste de modeles apparait, ca fonctionne
export OLLAMA_URL=http://192.168.x.x:11434   # Si fonctionne sur un autre ordinateur
```

> ⚠️ Si Ollama ne repond pas, demarrez-le d'abord avec `ollama serve`.

> ⚠️ **Les grands modeles sont lents a repondre** : Les grands modeles comme `qwen3.5:35b` ou `deepseek-r1` peuvent prendre plusieurs minutes pour generer une reponse. Meme si rien ne semble se passer, le traitement est peut-etre en cours — soyez patient.

---

## Changements recents (v0.1.16 ~ v0.1.24)

### v0.1.24 (2026-04-06)
- **Sous-commande RTK d'economie de tokens** : `wall-vault rtk <command>` filtre automatiquement la sortie des commandes shell pour reduire l'utilisation de tokens des agents IA de 60 a 90%. Inclut des filtres integres pour les commandes cles comme git et go, et tronque automatiquement les commandes non supportees. S'integre de maniere transparente avec Claude Code via le hook `PreToolUse`.

### v0.1.23 (2026-04-06)
- **Correction du changement de modele Ollama** : Correction d'un probleme ou le changement du modele Ollama dans le tableau de bord du coffre-fort n'etait pas refllete dans le proxy reel. Auparavant, seule la variable d'environnement (`OLLAMA_MODEL`) etait utilisee, maintenant les parametres du coffre-fort ont la priorite.
- **Signalisation automatique des services locaux** : Ollama, LM Studio et vLLM sont automatiquement actives lorsqu'accessibles et desactives lorsqu'inaccessibles. Fonctionne de la meme maniere que le basculement automatique base sur les cles pour les services cloud.

### v0.1.22 (2026-04-05)
- **Correction du champ content vide** : Quand les modeles de reflexion (gemini-3.1-pro, o1, claude thinking, etc.) utilisent tous les max_tokens pour le raisonnement et ne peuvent pas produire de reponse reelle, le proxy omettait les champs `content`/`text` du JSON de reponse via `omitempty`, ce qui faisait planter les clients SDK OpenAI/Anthropic avec `Cannot read properties of undefined (reading 'trim')`. Corrige pour toujours inclure les champs selon la specification API officielle.

### v0.1.21 (2026-04-05)
- **Support du modele Gemma 4** : Les modeles de la famille Gemma comme `gemma-4-31b-it` et `gemma-4-26b-a4b-it` peuvent maintenant etre utilises via l'API Google Gemini.
- **Support des services LM Studio / vLLM** : Auparavant, ces services manquaient dans le routage du proxy et basculaient toujours vers Ollama. Maintenant correctement routes via l'API compatible OpenAI.
- **Correction de l'affichage des services du tableau de bord** : Meme en cas de fallback, le tableau de bord affiche toujours le service configure par l'utilisateur.
- **Affichage du statut des services locaux** : Affiche le statut de connexion des services locaux (Ollama, LM Studio, vLLM, etc.) avec des couleurs de points ● au chargement du tableau de bord.
- **Variable d'environnement du filtre d'outils** : Utilisez la variable d'environnement `WV_TOOL_FILTER=passthrough` pour definir le mode de passage des outils.

### v0.1.20 (2026-03-28)
- **Renforcement complet de la securite** : Prevention XSS (41 emplacements), comparaison de tokens en temps constant, restrictions CORS, limites de taille de requete, prevention de traversee de chemin, authentification SSE, renforcement du limiteur de debit, et 12 autres ameliorations de securite.

### v0.1.19 (2026-03-27)
- **Detection en ligne de Claude Code** : Les instances de Claude Code ne passant pas par le proxy sont maintenant affichees comme en ligne dans le tableau de bord.

### v0.1.18 (2026-03-26)
- **Correction du blocage du service de fallback** : Apres une erreur temporaire causant un fallback vers Ollama, il revient automatiquement au service original lorsqu'il se retablit.
- **Amelioration de la detection hors ligne** : Les verifications de statut a intervalles de 15 secondes rendent la detection des pannes de proxy plus rapide.

### v0.1.17 (2026-03-25)
- **Tri des cartes par glisser-deposer** : Les cartes d'agents peuvent etre glissees et deposees pour changer leur ordre.
- **Bouton d'application de configuration en ligne** : Le bouton [⚡ Appliquer la configuration] est affiche sur les cartes d'agents hors ligne.
- **Type d'agent cokacdir ajoute**.

### v0.1.16 (2026-03-25)
- **Synchronisation bidirectionnelle des modeles** : Le changement d'un modele Cline ou Claude Code dans le tableau de bord du coffre-fort est automatiquement refllete.

---

*Pour des informations API plus detaillees, voir [API.md](API.md).*
