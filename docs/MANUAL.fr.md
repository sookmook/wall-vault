# Manuel Utilisateur wall-vault
*(Dernière mise à jour: 2026-03-20 — v0.1.15)*

---

## Table des matières

1. [Qu'est-ce que wall-vault ?](#quest-ce-que-wall-vault-)
2. [Installation](#installation)
3. [Premier démarrage (assistant setup)](#premier-démarrage-assistant-setup)
4. [Enregistrer une clé API](#enregistrer-une-clé-api)
5. [Utiliser le proxy](#utiliser-le-proxy)
6. [Tableau de bord du coffre-fort](#tableau-de-bord-du-coffre-fort)
7. [Mode distribué (multi-bots)](#mode-distribué-multi-bots)
8. [Démarrage automatique](#démarrage-automatique)
9. [Doctor — l'outil de diagnostic](#doctor--loutil-de-diagnostic)
10. [Référence des variables d'environnement](#référence-des-variables-denvironnement)
11. [Résolution des problèmes](#résolution-des-problèmes)

---

## Qu'est-ce que wall-vault ?

**wall-vault = proxy IA + coffre-fort de clés API, conçu pour OpenClaw**

Pour utiliser un service d'intelligence artificielle, vous avez besoin d'une **clé API** — c'est-à-dire un **badge numérique** qui prouve que vous avez le droit d'utiliser ce service. Ce badge a une limite d'utilisation quotidienne et peut être compromis s'il est mal géré.

wall-vault conserve tous ces badges dans un coffre-fort sécurisé, et joue le rôle de **proxy (intermédiaire)** entre OpenClaw et les services IA. En clair : OpenClaw ne connaît que wall-vault, et wall-vault s'occupe de tout le reste.

Ce que wall-vault résout pour vous :

- **Rotation automatique des clés** : quand une clé atteint sa limite ou est temporairement bloquée (cooldown), wall-vault bascule discrètement vers la clé suivante. OpenClaw continue de fonctionner sans interruption.
- **Basculement automatique de service (fallback)** : si Google ne répond pas, wall-vault passe à OpenRouter ; si OpenRouter est aussi indisponible, il bascule vers Ollama (IA locale installée sur votre machine). La session n'est jamais interrompue.
- **Synchronisation en temps réel (SSE)** : si vous changez de modèle dans le tableau de bord, le changement se reflète dans OpenClaw en 1 à 3 secondes. SSE (Server-Sent Events) est une technologie où le serveur pousse les mises à jour vers le client en temps réel.
- **Notifications en temps réel** : les événements comme l'épuisement d'une clé ou une panne de service s'affichent immédiatement en bas de l'interface TUI (terminal) d'OpenClaw.

> 💡 **Claude Code, Cursor et VS Code** peuvent également être connectés à wall-vault, mais l'usage principal reste avec OpenClaw.

```
OpenClaw (interface TUI dans le terminal)
        │
        ▼
  wall-vault proxy (:56244)   ← gestion des clés, routage, fallback, événements
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (plus de 340 modèles)
        └─ Ollama (votre machine, dernier recours)
```

---

## Installation

### Linux / macOS

Ouvrez un terminal et collez la commande correspondant à votre système.

```bash
# Linux (PC standard, serveur — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (Mac M1/M2/M3)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — télécharge le fichier depuis Internet.
- `chmod +x` — rend le fichier téléchargé exécutable. Sans cette étape, vous obtiendrez une erreur « Permission refusée ».

### Windows

Ouvrez PowerShell en tant qu'administrateur et exécutez les commandes suivantes.

```powershell
# Téléchargement
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Ajout au PATH (effectif après redémarrage de PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **Qu'est-ce que le PATH ?** C'est la liste des dossiers où l'ordinateur cherche les commandes. En ajoutant wall-vault au PATH, vous pourrez taper `wall-vault` depuis n'importe quel dossier.

### Compilation depuis les sources (pour les développeurs)

Uniquement si vous avez l'environnement de développement Go installé.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (version: v0.1.6.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Version avec horodatage** : en compilant avec `make build`, la version inclut automatiquement la date et l'heure, par exemple `v0.1.6.20260314.231308`. En compilant avec `go build ./...`, la version affichée sera simplement `"dev"`.

---

## Premier démarrage (assistant setup)

### Lancer l'assistant de configuration

Après l'installation, lancez obligatoirement l'**assistant de configuration** avec la commande suivante. Il vous guidera étape par étape.

```bash
wall-vault setup
```

L'assistant vous pose les questions suivantes :

```
1. Choix de la langue (10 langues disponibles, dont le français)
2. Choix du thème (light / dark / gold / cherry / ocean)
3. Mode de fonctionnement — seul (standalone) ou en réseau multi-machines (distributed)
4. Nom du bot — le nom affiché dans le tableau de bord
5. Ports — valeurs par défaut : proxy 56244, coffre-fort 56243 (appuyez sur Entrée pour garder les valeurs)
6. Services IA — choisissez parmi Google, OpenRouter, Ollama
7. Filtres de sécurité des outils
8. Token administrateur — le mot de passe pour protéger les fonctions d'administration. Génération automatique possible.
9. Mot de passe de chiffrement des clés API — pour un stockage encore plus sécurisé (optionnel)
10. Chemin du fichier de configuration
```

> ⚠️ **Notez bien votre token administrateur.** Vous en aurez besoin pour ajouter des clés ou modifier des paramètres dans le tableau de bord. Si vous le perdez, vous devrez modifier le fichier de configuration manuellement.

Une fois l'assistant terminé, le fichier de configuration `wall-vault.yaml` est créé automatiquement.

### Démarrage

```bash
wall-vault start
```

Deux serveurs démarrent simultanément :

- **Proxy** (`http://localhost:56244`) — l'intermédiaire entre OpenClaw et les services IA
- **Coffre-fort** (`http://localhost:56243`) — gestion des clés API et tableau de bord web

Ouvrez `http://localhost:56243` dans votre navigateur pour accéder au tableau de bord.

---

## Enregistrer une clé API

Il y a quatre façons d'enregistrer une clé API. **Pour débuter, la méthode 1 (variables d'environnement) est recommandée.**

### Méthode 1 : Variables d'environnement (recommandée — la plus simple)

Une variable d'environnement est une **valeur préconfigurée** que le programme lit au démarrage. Entrez les commandes suivantes dans votre terminal.

```bash
# Enregistrer une clé Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Enregistrer une clé OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Lancer wall-vault après enregistrement
wall-vault start
```

Si vous avez plusieurs clés, séparez-les par des virgules. wall-vault les utilisera en rotation automatique (round-robin) :

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Astuce** : la commande `export` ne s'applique qu'à la session de terminal en cours. Pour conserver ces valeurs après un redémarrage, ajoutez ces lignes à votre fichier `~/.bashrc` ou `~/.zshrc`.

### Méthode 2 : Interface du tableau de bord (avec la souris)

1. Ouvrez `http://localhost:56243` dans votre navigateur
2. Dans la carte **🔑 Clés API** en haut, cliquez sur le bouton `[+ Ajouter]`
3. Renseignez le type de service, la valeur de la clé, un libellé (nom mémo) et la limite journalière, puis sauvegardez

### Méthode 3 : API REST (pour les scripts et l'automatisation)

L'API REST est un moyen pour les programmes d'échanger des données via HTTP. Utile pour l'enregistrement automatisé via script.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer VOTRE_TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Clé principale",
    "daily_limit": 1000
  }'
```

### Méthode 4 : Drapeau proxy (pour un test rapide)

Pour tester temporairement sans enregistrement officiel. La clé disparaît à l'arrêt du programme.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Utiliser le proxy

### Utilisation avec OpenClaw (usage principal)

Voici comment configurer OpenClaw pour qu'il se connecte aux services IA via wall-vault.

Ouvrez le fichier `~/.openclaw/openclaw.json` et ajoutez-y le contenu suivant :

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // token de l'agent vault
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

> 💡 **Plus facile** : cliquez sur le bouton **🦞 Copier la config OpenClaw** dans la carte agent du tableau de bord — un extrait de configuration déjà rempli avec votre token et votre adresse est copié dans le presse-papiers. Il ne vous reste qu'à coller.

**Vers quel service `wall-vault/` dans le nom du modèle pointe-t-il ?**

wall-vault détermine automatiquement le service IA à utiliser en lisant le nom du modèle :

| Format du modèle | Service utilisé |
|----------|--------------|
| `wall-vault/gemini-*` | Connexion directe à Google Gemini |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | Connexion directe à OpenAI |
| `wall-vault/claude-*` | Connexion à Anthropic via OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (contexte 1M tokens gratuit) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | Connexion via OpenRouter |
| `google/nom-du-modèle`, `openai/nom-du-modèle`, `anthropic/nom-du-modèle`, etc. | Connexion directe au service correspondant |
| `custom/google/nom-du-modèle`, `custom/openai/nom-du-modèle`, etc. | Supprime le préfixe `custom/` puis re-route |
| `nom-du-modèle:cloud` | Supprime le suffixe `:cloud` puis route vers OpenRouter |

> 💡 **Qu'est-ce que le contexte ?** C'est la quantité de conversation que l'IA peut mémoriser en une seule fois. 1M (un million de tokens) permet de traiter de très longues conversations ou de très longs documents d'un seul coup.

### Connexion directe au format API Gemini (compatibilité avec les outils existants)

Si vous avez déjà un outil qui utilise l'API Google Gemini directement, il suffit de rediriger l'adresse vers wall-vault :

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Ou, si votre outil permet de spécifier une URL directement :

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Utilisation avec le SDK OpenAI (Python)

Vous pouvez connecter wall-vault à du code Python qui utilise l'IA. Il suffit de changer `base_url` :

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # wall-vault gère les clés API lui-même
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # format provider/modèle
    messages=[{"role": "user", "content": "Bonjour"}]
)
```

### Changer de modèle à la volée

Pour changer le modèle IA utilisé pendant que wall-vault est en cours d'exécution :

```bash
# Changer le modèle directement via le proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# En mode distribué (multi-bots), changer depuis le coffre-fort → répercussion immédiate via SSE
curl -X PUT http://localhost:56243/admin/clients/ID-DE-VOTRE-BOT \
  -H "Authorization: Bearer VOTRE_TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Consulter la liste des modèles disponibles

```bash
# Voir tous les modèles
curl http://localhost:56244/api/models | python3 -m json.tool

# Voir uniquement les modèles Google
curl "http://localhost:56244/api/models?service=google"

# Rechercher par nom (exemple : modèles contenant "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Résumé des principaux modèles par service :**

| Service | Principaux modèles |
|--------|----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | Plus de 346 modèles (Hunter Alpha 1M contexte gratuit, DeepSeek R1/V3, Qwen 2.5, etc.) |
| Ollama | Détection automatique du serveur local installé sur votre machine |

---

## Tableau de bord du coffre-fort

Ouvrez `http://localhost:56243` dans votre navigateur pour accéder au tableau de bord.

**Structure de l'interface :**
- **Barre supérieure fixe (topbar)** : logo, sélecteurs de langue et de thème, indicateur d'état de la connexion SSE
- **Grille de cartes** : cartes agents, services et clés API disposées en tuiles

### Carte des clés API

Cette carte vous permet de gérer en un coup d'œil toutes les clés API enregistrées.

- Affiche la liste des clés organisée par service.
- `today_usage` : nombre de tokens (unités de texte traitées par l'IA) utilisés avec succès aujourd'hui
- `today_attempts` : nombre total d'appels aujourd'hui (succès + échecs)
- Utilisez le bouton `[+ Ajouter]` pour enregistrer une nouvelle clé, et `✕` pour en supprimer une.

> 💡 **Qu'est-ce qu'un token ?** C'est l'unité que l'IA utilise pour traiter du texte. Grosso modo, un token correspond à un mot en anglais, ou à 1 à 2 caractères en coréen. La facturation des API est généralement calculée d'après le nombre de tokens.

### Carte des agents

Cette carte affiche l'état des bots (agents) connectés au proxy wall-vault.

**L'état de connexion s'affiche en 4 niveaux :**

| Indicateur | État | Signification |
|------|------|------|
| 🟢 | En cours d'exécution | Le proxy fonctionne normalement |
| 🟡 | En attente | Répond, mais lentement |
| 🔴 | Hors ligne | Le proxy ne répond pas |
| ⚫ | Non connecté / désactivé | Le proxy ne s'est jamais connecté au coffre-fort, ou est désactivé |

**Guide des boutons en bas de la carte agent :**

Lorsque vous enregistrez un agent, vous spécifiez son **type**. Les boutons de raccourci appropriés apparaissent alors automatiquement.

---

#### 🔘 Bouton « Copier la config » — génère automatiquement la configuration de connexion

En cliquant sur ce bouton, un extrait de configuration déjà rempli avec le token de l'agent, l'adresse du proxy et les informations de modèle est copié dans le presse-papiers. Il suffit de le coller à l'emplacement indiqué dans le tableau ci-dessous.

| Bouton | Type d'agent | Où coller |
|------|-------------|-------------|
| 🦞 Copier la config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Copier la config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Copier la config Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Copier la config Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Copier la config VSCode | `vscode` | `~/.continue/config.json` |

**Exemple — pour un agent de type Claude Code, voici ce qui est copié :**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "le-token-de-cet-agent"
}
```

**Exemple — pour un agent de type VSCode (Continue) :**

```json
// ~/.continue/config.json
{
  "models": [{
    "title": "wall-vault proxy",
    "provider": "openai",
    "model": "gemini-2.5-flash",
    "apiBase": "http://192.168.1.20:56244/v1",
    "apiKey": "le-token-de-cet-agent"
  }]
}
```

**Exemple — pour un agent de type Cursor :**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : le-token-de-cet-agent

// Ou via variable d'environnement :
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=le-token-de-cet-agent
```

> ⚠️ **Si la copie dans le presse-papiers ne fonctionne pas** : les politiques de sécurité du navigateur peuvent bloquer cette action. Si une zone de texte s'ouvre dans une fenêtre pop-up, utilisez Ctrl+A pour tout sélectionner, puis Ctrl+C pour copier.

---

#### 🟣 Bouton « Copier la commande de déploiement » — pour installer sur une nouvelle machine

Ce bouton est utile quand vous installez le proxy wall-vault sur un nouvel ordinateur et souhaitez le connecter au coffre-fort. Un clic copie l'intégralité du script d'installation. Collez-le dans le terminal du nouvel ordinateur et exécutez-le ; il effectue en une seule fois :

1. Installation du binaire wall-vault (ignorée s'il est déjà installé)
2. Enregistrement automatique en tant que service utilisateur systemd
3. Démarrage du service et connexion automatique au coffre-fort

> 💡 Le script contient déjà le token de cet agent et l'adresse du serveur coffre-fort — aucune modification manuelle n'est nécessaire après le collage.

---

### Carte des services

Cette carte permet d'activer ou de désactiver chaque service IA, et d'en configurer les paramètres.

- Interrupteur d'activation/désactivation par service
- Entrez l'adresse d'un serveur IA local (Ollama, LM Studio, vLLM, etc. tournant sur votre machine) pour que wall-vault détecte automatiquement les modèles disponibles.
- **Indicateur d'état de connexion locale** : le point ● à côté du nom du service est **vert** si connecté, **gris** sinon.
- **Synchronisation automatique des cases à cocher** : si un service local (Ollama, etc.) est en cours d'exécution lors de l'ouverture de la page, la case correspondante se coche automatiquement.

> 💡 **Si votre service local tourne sur une autre machine** : saisissez l'IP de cette machine dans le champ URL du service. Exemple : `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio)

### Saisie du token administrateur

Lorsque vous essayez d'effectuer une action importante dans le tableau de bord (ajouter ou supprimer une clé, etc.), une fenêtre pop-up vous demande le token administrateur. Saisissez le token configuré lors de l'assistant setup. Il reste mémorisé jusqu'à la fermeture du navigateur.

> ⚠️ **Si vous échouez plus de 10 fois en moins de 15 minutes, votre adresse IP sera temporairement bloquée.** Si vous avez oublié votre token, consultez l'entrée `admin_token` dans le fichier `wall-vault.yaml`.

---

## Mode distribué (multi-bots)

Quand vous utilisez OpenClaw sur plusieurs machines simultanément, vous pouvez **partager un seul coffre-fort** entre toutes. La gestion des clés se fait en un seul endroit, ce qui simplifie tout.

### Exemple de configuration

```
[Serveur coffre-fort]
  wall-vault vault    (coffre-fort :56243, tableau de bord)

[WSL Alpha]             [Raspberry Pi Gamma]    [Mac Mini local]
  wall-vault proxy        wall-vault proxy          wall-vault proxy
  openclaw TUI            openclaw TUI              openclaw TUI
  ↕ sync SSE              ↕ sync SSE                ↕ sync SSE
```

Tous les bots pointent vers le serveur coffre-fort central. Un changement de modèle ou l'ajout d'une clé dans le coffre-fort se répercute instantanément sur tous les bots.

### Étape 1 : Démarrer le serveur coffre-fort

Sur la machine qui servira de serveur coffre-fort :

```bash
wall-vault vault
```

### Étape 2 : Enregistrer chaque bot (client)

Enregistrez à l'avance les informations de chaque bot qui se connectera au serveur coffre-fort :

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer VOTRE_TOKEN_ADMIN" \
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

Sur chaque machine hébergeant un bot, démarrez le proxy en indiquant l'adresse et le token du serveur coffre-fort :

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Remplacez **`192.168.x.x`** par l'adresse IP interne réelle de la machine serveur coffre-fort. Vous pouvez la trouver dans les paramètres de votre routeur ou avec la commande `ip addr`.

---

## Démarrage automatique

Si vous ne souhaitez pas démarrer wall-vault manuellement à chaque redémarrage, enregistrez-le comme service système. Une fois enregistré, il démarrera automatiquement au démarrage.

### Linux — systemd (la plupart des distributions Linux)

systemd est le système qui gère le démarrage automatique des programmes sous Linux :

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

launchd est le système qui gère l'exécution automatique des programmes sous macOS :

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Téléchargez NSSM depuis [nssm.cc](https://nssm.cc/download) et ajoutez-le au PATH.
2. Dans PowerShell en tant qu'administrateur :

```powershell
wall-vault doctor deploy windows
```

---

## Doctor — l'outil de diagnostic

La commande `doctor` est un **outil de diagnostic et de réparation automatique** qui vérifie que wall-vault est correctement configuré.

```bash
wall-vault doctor check   # Diagnostic de l'état actuel (lecture seule, aucune modification)
wall-vault doctor fix     # Correction automatique des problèmes
wall-vault doctor all     # Diagnostic + correction automatique en une seule commande
```

> 💡 Si quelque chose semble anormal, commencez par `wall-vault doctor all`. Il résout automatiquement de nombreux problèmes courants.

---

## Référence des variables d'environnement

Les variables d'environnement sont un moyen de passer des valeurs de configuration à un programme. Entrez-les sous la forme `export NOM_VARIABLE=valeur` dans le terminal, ou placez-les dans votre fichier de service pour qu'elles soient toujours appliquées.

| Variable | Description | Exemple de valeur |
|------|------|---------|
| `WV_LANG` | Langue du tableau de bord | `fr`, `en`, `ko`, `ja` |
| `WV_THEME` | Thème du tableau de bord | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Clé(s) API Google (séparées par des virgules) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Clé API OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Adresse du serveur coffre-fort en mode distribué | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Token d'authentification du client (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Token administrateur | `admin-token-here` |
| `WV_MASTER_PASS` | Mot de passe de chiffrement des clés API | `my-password` |
| `WV_AVATAR` | Chemin vers le fichier image avatar (relatif à `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Adresse du serveur Ollama local | `http://192.168.x.x:11434` |

---

## Résolution des problèmes

### Le proxy ne démarre pas

Le port est souvent déjà utilisé par un autre programme.

```bash
ss -tlnp | grep 56244   # Vérifier quel programme utilise le port 56244
wall-vault proxy --port 8080   # Démarrer sur un autre numéro de port
```

### Erreurs de clé API (429, 402, 401, 403, 582)

| Code d'erreur | Signification | Que faire |
|----------|------|----------|
| **429** | Trop de requêtes (quota dépassé) | Attendez un peu ou ajoutez une autre clé |
| **402** | Paiement requis ou crédit insuffisant | Rechargez le crédit sur le service concerné |
| **401 / 403** | Clé incorrecte ou accès non autorisé | Vérifiez la valeur de la clé et re-enregistrez-la |
| **582** | Surcharge de la passerelle (cooldown 5 min) | Se lève automatiquement après 5 minutes |

```bash
# Vérifier la liste et l'état des clés enregistrées
curl -H "Authorization: Bearer VOTRE_TOKEN_ADMIN" http://localhost:56243/admin/keys

# Réinitialiser les compteurs d'utilisation des clés
curl -X POST -H "Authorization: Bearer VOTRE_TOKEN_ADMIN" http://localhost:56243/admin/keys/reset
```

### Un agent s'affiche comme « Non connecté »

« Non connecté » signifie que le processus proxy n'envoie pas de signal (heartbeat) au coffre-fort. **Cela ne veut pas dire que la configuration est perdue.** Le proxy doit connaître l'adresse du serveur coffre-fort et le token pour passer à l'état connecté.

```bash
# Démarrer le proxy en spécifiant l'adresse du coffre-fort, le token et l'ID client
WV_VAULT_URL=http://ADRESSE_DU_COFFRE:56243 \
WV_VAULT_TOKEN=TOKEN_CLIENT \
WV_VAULT_CLIENT_ID=ID_CLIENT \
wall-vault proxy
```

Une fois la connexion établie, le tableau de bord passe à 🟢 En cours d'exécution en moins de 20 secondes.

### Impossible de se connecter à Ollama

Ollama est un programme qui exécute l'IA directement sur votre machine. Vérifiez d'abord qu'Ollama est bien lancé.

```bash
curl http://localhost:11434/api/tags   # Si une liste de modèles s'affiche, tout est normal
export OLLAMA_URL=http://192.168.x.x:11434   # Si Ollama tourne sur une autre machine
```

> ⚠️ Si Ollama ne répond pas, lancez-le d'abord avec la commande `ollama serve`.

> ⚠️ **Les grands modèles sont lents à répondre** : des modèles imposants comme `qwen3.5:35b` ou `deepseek-r1` peuvent mettre plusieurs minutes à générer une réponse. Si rien ne semble se passer, c'est probablement normal — patientez.

---

*Pour des informations plus détaillées sur l'API, consultez [API.md](API.md).*
