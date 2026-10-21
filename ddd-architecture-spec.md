# Spécification d'Architecture : Refonte DDD, Injection de Dépendances & Contraintes 40L / 300L

## 1. Objectif et Périmètre

Ce document spécifie la refonte architecturale complète (**Full-Stack : Backend Go + Frontend Svelte 5 / TypeScript + CSS**) de **Fana Karaoke** selon les principes :
1. **Domain-Driven Design (DDD)** avec **Architecture Hexagonale (Ports & Adapters)** côté Go et côté TypeScript/Svelte, et un **cœur de domaine fonctionnel** (fonctions pures sur structures immuables / copiées).
2. **Injection de Dépendances (DI)** manuelle et explicite par constructeurs côté Go (`cmd/server/main.go` comme unique Composition Root) et via un conteneur typé injecté par **Svelte Context** côté Frontend (`web/src/main.ts` / `App.svelte`).
3. **Contraintes strictes de taille** :
   - **Maximum 300 lignes par fichier** (à la fois en lignes brutes totales **et** en SLOC) sur l'ensemble des fichiers sources (`.go`, `.ts`, `.svelte`, `.css`).
   - **Maximum 40 lignes par fonction ou méthode** (à la fois en lignes brutes de la déclaration `func`/`function` de l'ouverture à l'accolade fermante **et** en SLOC).
4. **Harnais de vérification automatisé** (`internal/architecture/architecture_test.go`) vérifiant par analyse AST/texte les limites 300L/40L et l'isolation stricte des couches DDD (Go et TypeScript/Svelte).
5. **Iso-fonctionnalité & Rétrocompatibilité** : Le contrat HTTP REST, le protocole WebSocket, les cookies signés et le comportement utilisateur restent 100 % identiques. Les 4 suites Playwright (`flow.mjs`, `autoplay.mjs`, `audit.mjs`, `voting.mjs`) doivent passer de bout en bout.

---

## 2. Diagnostic de l'Architecture Actuelle

### 2.1 Violations actuelles de la limite des 300 lignes par fichier

| Fichier actuel | Lignes | Problèmes architecturaux identifiés |
| :--- | :---: | :--- |
| `web/src/views/GuestRoom.svelte` | **1 449** | Monolithe mêlant appels HTTP, gestion WebSocket, état réactif, logique de vote/recherche/hôte, 3 modales et ~700 lignes de CSS. |
| `web/src/views/TvRoom.svelte` | **1 407** | Monolithe mêlant télémétrie YouTube, raccourcis clavier, état WS, tiroir de file (drag & drop), overlays et ~700 lignes de CSS. |
| `internal/httpapi/server.go` | **627** | Couplage direct à `*app.Manager` concret, routage, middlewares, sécurité CSRF/Origin, 18 handlers HTTP, décodage JSON et serveur SPA dans un seul fichier. |
| `web/src/views/Welcome.svelte` | **610** | Vue d'accueil combinant formulaire de création/jonction, résolution de code compact, hero animé et ~400 lignes de CSS. |
| `internal/app/mutations.go` | **544** | Mêle orchestration applicative, verrouillage `sync.Mutex`, mutations directes de maps et règles métier (tri de file, majorité de vote, avancement lecture). |
| `internal/search/invidious.go` | **535** | Combine client HTTP Invidious, rotation d'instances en panne, probing YouTube oEmbed, heuristique karaoké et cache mémoire TTL/Stale sous mutex. |
| `web/src/lib/i18n.ts` | **500** | Dictionnaires `en` et `fr`, store Svelte et formatage des notices/erreurs regroupés dans un seul fichier. |
| `web/src/app.css` | **480** | Variables CSS, reset, typographie, boutons, cartes, formulaires, badges et toasts dans une seule feuille globale. |
| `internal/app/manager.go` | **430** | God-object contenant l'état en mémoire (`rooms`, `devices`, `hostTokens`), le `sync.Mutex`, la génération aléatoire crypto, la projection `Snapshot` et les règles d'éligibilité. |
| `web/src/features/tv/Player.svelte` | **354** | Mêle chargement du script YouTube IFrame API, machine à états d'autoplay muet/sonore, timers de récupération et rendu de l'overlay. |

### 2.2 Violations majeures de la limite des 40 lignes par fonction
- **`cmd/server/main.go`** : `main()` fait **91 lignes** (configuration, instanciation, démarrage HTTP et arrêt gracieux mélangés).
- **`internal/app/mutations.go`** : `AddSong` (**99 lignes**), `StartSkipVote` (**57 lignes**), `CastSkipVote` (**44 lignes**), `Playback` (**56 lignes**).
- **`internal/app/manager.go`** : `snapshotLocked` (**98 lignes**), `CreateRoom` (**44 lignes**).
- **`internal/app/lifecycle.go`** : `Tick` (**60 lignes**).
- **`internal/search/invidious.go`** : `New` (**41 lignes**), `fetchSearch` (**47 lignes**), `parseSearchItems` (**52 lignes**).
- **`internal/httpapi/server.go`** : `Handler` (**30 lignes**), `middleware` (**36 lignes**), `handleSPA` (**45 lignes**).
- **`internal/httpapi/ws.go`** : `handleWS` (**58 lignes**), `readLoop` (**48 lignes**).
- **Frontend `.svelte`** : fonctions `boot`, `submitAdd`, `applyState`, `createPlayer`, `scheduleBlockedCheck` dépassant ou frôlant 40 lignes, en plus des blocs `<script>` et `<style>` massifs.

### 2.3 Défauts vis-à-vis du DDD et de l'Injection de Dépendances
1. **Absence de Ports Secondaires explicites** :
   - La persistance en mémoire (`map[string]*domain.Room`, `map[string]*domain.DeviceInfo`) est enfouie dans `app.Manager` au lieu d'être abstraite derrière des interfaces `RoomRepository` et `DeviceRepository`.
   - L'horloge (`time.Now`) et la génération d'identifiants/codes/PIN (`crypto/rand`) sont appelées directement ou via des champs ad-hoc au lieu de ports `Clock` et `IdentityGenerator` injectés.
   - La notification temps-réel utilise un setter mutatif `mgr.SetCallbacks(s.hub.Broadcast, s.hub.Ended)` après construction (dépendance circulaire implicite entre `Server`, `Hub` et `Manager`) au lieu d'un port `EventPublisher` injecté par constructeur.
2. **Modèle de Domaine anémique et couplé au transport** :
   - `internal/domain/types.go` mélange les entités du domaine (`Room`, `QueueEntry`, `NowPlaying`, `SkipVote`) et les DTOs de sérialisation JSON (`Snapshot`, `RoomView`, `NowPlayingView`, tags `json:"..."`).
   - Les règles métier pures (tri de file par score puis séquence d'insertion, calcul du quorum de skip vote `len/2 + 1`, transition countdown → playing, validation des durées vidéo 60s–600s) sont dispersées dans `internal/app`.
3. **Couplage dans l'adapteur HTTP** :
   - `httpapi.Server` dépend de la struct concrète `*app.Manager` au lieu d'un port primaire (interface).
   - `NewLimiter()`, la génération de QR code (`encodeQR`) et la signature HMAC des cookies sont instanciés ou appelés en dur dans `httpapi` sans injection par interface.
4. **Absence d'architecture hexagonale côté Frontend** :
   - Les composants `.svelte` importent directement les singletons `api` (`fetch`), `connectRoom` (`WebSocket`) et manipulent `window.YT.Player` sans inversion de dépendance.

---

## 3. Règles Architecturales et Métriques Strictes

### 3.1 Limites de taille (vérifiées automatiquement)

| Règle | Seuil Max | Définition précise | Périmètre |
| :--- | :---: | :--- | :--- |
| **Lignes par fichier (Brut)** | **<= 300 lignes** | Nombre total de lignes (`\n`), incluant déclarations de package, imports, commentaires, lignes vides, `<script>`, markup HTML et `<style>`. | Tous les fichiers `.go`, `.ts`, `.svelte`, `.css` (hors `node_modules`, `dist`). |
| **Lignes par fichier (SLOC)** | **<= 300 lignes** | Lignes non vides et non purement commentées. | Idem. |
| **Lignes par fonction (Brut)** | **<= 40 lignes** | De la ligne de signature (`func`, `function`, méthode ou fléchée assignée) jusqu'à la ligne de l'accolade fermante `}` incluse. | Toutes les fonctions/méthodes en `.go`, `.ts` et dans les blocs `<script>` des `.svelte`. |
| **Lignes par fonction (SLOC)** | **<= 40 lignes** | Lignes de code effectives à l'intérieur de la fonction (hors lignes vides et commentaires). | Idem. |

### 3.2 Matrice de dépendances autorisées (Isolation Hexagonale)

#### Backend Go (`internal/`)
1. **`internal/domain`** :
   - **Autorisé** : Bibliothèque standard Go pure uniquement (`time`, `strings`, `sort`, `fmt`, `unicode/utf8`, `math`).
   - **Interdit** : Tout import de `internal/ports`, `internal/usecase`, `internal/adapters`, `net/http`, `sync`, `crypto/rand`, `encoding/json`, ou librairie externe.
2. **`internal/ports`** :
   - **Autorisé** : Standard library (`context`, `time`, `net/http` uniquement si nécessaire pour types standards, préférer des types primitifs/domaine) + `internal/domain`.
   - **Interdit** : `internal/usecase`, `internal/adapters`.
3. **`internal/usecase`** :
   - **Autorisé** : Standard library + `internal/domain` + `internal/ports`.
   - **Interdit** : `internal/adapters`, `net/http`, `nhooyr.io/websocket`, `rsc.io/qr`.
4. **`internal/adapters/primary/*` et `internal/adapters/secondary/*`** :
   - **Autorisé** : Standard library + `internal/domain` + `internal/ports` + librairies d'infrastructure spécifiques à l'adapteur.
   - **Interdit** : Importer `internal/usecase` (doit passer uniquement par les interfaces de `internal/ports/input`) ou importer un autre package voisin d'`adapters`.
5. **`cmd/server`** (Composition Root) :
   - **Autorisé** : Seul package autorisé à importer simultanément `ports`, `usecase`, `adapters/primary/*` et `adapters/secondary/*` pour construire le graphe d'objets.

#### Frontend TypeScript / Svelte (`web/src/`)
1. **`web/src/domain`** :
   - **Autorisé** : TypeScript pur uniquement (aucun import de `svelte`, `window`, `document`, `fetch`, `WebSocket`, ni des couches `ports`, `application`, `adapters`, `ui`).
2. **`web/src/ports`** :
   - **Autorisé** : `web/src/domain` uniquement. Définit les contrats (`RoomApiPort`, `RoomSocketPort`, `VideoPlayerPort`, `ClipboardPort`, `StoragePort`).
3. **`web/src/application`** :
   - **Autorisé** : `web/src/domain`, `web/src/ports`, `svelte/store`.
   - **Interdit** : `web/src/adapters`, `web/src/ui`, appels directs à `fetch`, `WebSocket` ou `window.YT`.
4. **`web/src/adapters`** :
   - **Autorisé** : `web/src/domain`, `web/src/ports`, APIs navigateur (`fetch`, `WebSocket`, `window.YT`, `localStorage`, `navigator.clipboard`).
   - **Interdit** : `web/src/ui`, `web/src/application`.
5. **`web/src/ui`** :
   - **Autorisé** : `web/src/domain`, `web/src/ports`, `web/src/application`, `svelte`. Récupère ses dépendances via le contexte DI Svelte (`getAppContext()`).
   - **Interdit** : Importer directement `web/src/adapters`.

---

## 4. Architecture Cible Backend (Go)

### 4.1 Arborescence des fichiers Backend (tous <= 300 lignes)

```text
cmd/
  server/
    main.go                                  (~55L) Point d'entrée minimal : charge config, appelle wire, gère signaux
    config.go                                (~95L) Parsing des variables d'environnement vers Config typée
    wire.go                                  (~90L) Assemblage DI par constructeurs de tous les adapteurs et use cases

internal/
  domain/                                    # Cœur fonctionnel pur (0 dépendance externe, 0 mutex, 0 I/O)
    room.go                                  (~110L) Types Room, DeviceInfo, fonctions pures de création et transitions Room
    queue.go                                 (~110L) Type QueueEntry, fonctions pures Enqueue, Remove, ToggleVote, Sort, Reorder
    playback.go                              (~110L) Type NowPlaying, Status, fonctions pures StartNext, Play, Pause, Cancel, Elapsed
    voting.go                                (~115L) Type SkipVote, fonctions pures EligibleVoters, StartVote, CastVote, ExpireVote
    video.go                                 (~90L) Value Objects VideoID, VideoMeta, SearchResult, ParseYouTubeURL, ValidateMeta
    notice.go                                (~55L) Type Notice, PrependNotice (max 20), formatage de clé
    snapshot.go                              (~120L) Types Snapshot/Views et fonction pure ProjectSnapshot(room, devices, now, viewer)
    errors.go                                (~65L) Type AppError et erreurs sentinelles du domaine

  ports/
    input.go                                 (~95L) Driving Ports : KaraokeFacade + interfaces RoomUseCase, QueueUseCase, PlaybackUseCase, VoteUseCase, CatalogUseCase
    output.go                                (~110L) Driven Ports : RoomRepository, DeviceRepository, EventPublisher, VideoCatalog, EmbedProber, Clock, IDGenerator, QRGenerator, SessionSigner, RateLimiter

  usecase/                                   # Orchestration applicative, coordonne les ports secondaires et le domaine pur
    facade.go                                (~110L) Struct Façade unique implémentant ports.KaraokeFacade en déléguant aux sous-services
    room_service.go                          (~120L) CreateRoom, ResolveCode, GetSnapshot, RotatePIN, EndRoom, HostSeen
    queue_service.go                         (~130L) AddSong, RemoveSong, Reorder, ClearQueue, LockQueue, ToggleVote
    playback_service.go                      (~120L) Playback(play/pause/next/cancelCountdown), TVEnded, TVError, TVProgress
    vote_service.go                          (~110L) StartSkipVote, CastSkipVote, CancelSkipVote
    lifecycle_service.go                     (~110L) Start(ctx), Tick() : avancement countdowns, expiration votes/cooldowns, nettoyage rooms
    catalog_service.go                       (~80L) Search et Lookup orchestrant cache, catalogue et validation domaine

  adapters/
    primary/
      http/
        server.go                            (~110L) Struct Server, constructeur New(...) avec DI, montage du routeur
        routes.go                            (~65L) Enregistrement déclaratif des routes sur http.ServeMux
        middleware.go                        (~95L) Headers sécurité (CSP, Referrer-Policy), vérification same-origin, recovery, logging
        handlers_room.go                     (~125L) handleHealth, handleConfig, handleCreateRoom, handleResolveRoom, handleTVSession, handleSnapshot, handleQR, handleEndRoom
        handlers_queue.go                    (~120L) handleSearch, handleAddSong, handleVote, handleRemoveSong, handleReorder, handleClearQueue, handleLockQueue
        handlers_playback.go                 (~115L) handlePlayback, handleStartSkipVote, handleCastSkipVote, handleCancelSkipVote, handleHostUnlock, handleRotatePIN
        handlers_spa.go                      (~95L) handleSPA (/tv/{token}, /r/{code}), handleAsset, handleFavicon
        context_helpers.go                   (~95L) Lecture/écriture cookies via SessionSigner, clientIP, réponse JSON et mapping erreurs
      ws/
        hub.go                               (~140L) Implémente ports.EventPublisher (Broadcast, Ended) et gère le registre de clients WS
        handler.go                           (~125L) Upgrade WebSocket, readLoop (tvEnded, tvError, tvProgress, ping), heartbeat

    secondary/
      memory/
        room_repo.go                         (~145L) Implémente ports.RoomRepository (thread-safe avec sync.RWMutex, opérations atomiques UpdateRoom)
        device_repo.go                       (~105L) Implémente ports.DeviceRepository (suivi présence, comptage invités actifs)
        rate_limiter.go                      (~70L) Implémente ports.RateLimiter (token bucket en mémoire par clé)
      invidious/
        client.go                            (~155L) Implémente ports.VideoCatalog (requêtes HTTP Invidious, rotation instances, scoring karaoké)
        parser.go                            (~110L) Parsing JSON Invidious et normalisation des requêtes/résultats
        oembed_prober.go                     (~95L) Implémente ports.EmbedProber (vérification YouTube oEmbed + fallback métadonnées)
        cached_catalog.go                    (~160L) Décorateur implémentant ports.VideoCatalog avec cache frais/stale et cache de vidéos bloquées
      crypto/
        id_generator.go                      (~95L) Implémente ports.IDGenerator (NewID, NewRoomCode, NewPIN, NewToken)
        words.go                             (~35L) Listes d'adjectifs et noms pour les codes de salle
        hmac_signer.go                       (~95L) Implémente ports.SessionSigner (Sign, Verify, génération cookies Hôte/Device/TV)
      qr/
        rsc_generator.go                     (~65L) Implémente ports.QRGenerator (génération SVG via rsc.io/qr)
      clock/
        system_clock.go                      (~25L) Implémente ports.Clock (time.Now) et FakeClock pour les tests

  architecture/
    architecture_test.go                     (~220L) Test automatisé vérifiant <=300L/fichier, <=40L/fonction et règles d'imports DDD (Go + Front)
```

### 4.2 Détail du Domaine Fonctionnel Pur (`internal/domain`)

Toutes les fonctions du package `domain` sont **pures** : elles prennent un état en entrée (et l'instant `now time.Time` ou un ID pré-généré), ne font aucun appel I/O ni verrouillage, et retournent le nouvel état (ou une erreur du domaine) :

```go
// internal/domain/room.go
func NewRoom(code, hostToken, pin string, now time.Time) Room
func TouchHost(room Room, now time.Time) Room
func IsExpired(room Room, now time.Time, roomTTL time.Duration) bool
func NormalizeCompactCode(input string) string

// internal/domain/queue.go
func EnqueueSong(room Room, entry QueueEntry, now time.Time, maxQueue int) (Room, error)
func RemoveEntry(room Room, entryID, noticeID string, now time.Time) (Room, error)
func ToggleEntryVote(room Room, entryID, deviceID string) (Room, error)
func ReorderQueue(room Room, orderedIDs []string, noticeID string, now time.Time) (Room, error)
func ClearEntries(room Room, noticeID string, now time.Time) Room
func SetQueueLock(room Room, locked bool, noticeID string, now time.Time) Room
func SortQueue(entries []QueueEntry, manualOrder bool) []QueueEntry

// internal/domain/playback.go
func ApplyPlaybackAction(room Room, action string, noticeID string, now time.Time) (Room, error)
func AdvanceOnEnded(room Room, entryID string, generation uint64, now time.Time) (Room, bool)
func AdvanceOnError(room Room, entryID, reason, noticeID string, generation uint64, now time.Time) (Room, bool)
func UpdateProgress(room Room, entryID string, generation uint64, pos, dur float64, now time.Time) (Room, bool)
func AdvanceCountdownIfDue(room Room, now time.Time) (Room, bool)

// internal/domain/voting.go
func StartSkipVote(room Room, initiatorID, noticeID string, activeDevices []string, now time.Time) (Room, error)
func CastSkipVote(room Room, voterID, noticeID string, now time.Time) (Room, error)
func CancelSkipVote(room Room) Room
func ExpireSkipVoteIfDue(room Room, noticeID string, now time.Time) (Room, bool)
func ClearCooldownIfDue(room Room, now time.Time) (Room, bool)

// internal/domain/video.go
func ParseYouTubeURL(rawURL string) (string, error)
func ValidateVideoID(id string) bool
func ValidateVideoMeta(meta VideoMeta) error
```

### 4.3 Détail des Ports Primaires et Secondaires (`internal/ports`)

#### Ports Primaires (`internal/ports/input.go`)
Conformément au choix d'une **Façade unique déléguant à des sous-services** :

```go
type RoomCommands interface {
	CreateRoom(ctx context.Context) (code, token, pin string, err error)
	ResolveCode(input string) (string, bool)
	RoomByHostToken(token string) (string, bool)
	PIN(code string) (string, bool)
	VerifyPIN(code, candidate string) bool
	RotatePIN(code string) (string, error)
	EndRoom(code, deviceID, role string) error
	Snapshot(code, deviceID, role string) (*domain.Snapshot, error)
	HostSeen(code string)
	TouchDevice(deviceID string, host bool)
	SetConnected(deviceID string, connected bool)
}

type QueueCommands interface {
	AddSong(ctx context.Context, code, deviceID, role, nickname, videoID, rawURL string) (*domain.Snapshot, error)
	ToggleVote(code, deviceID, role, entryID string) (*domain.Snapshot, error)
	RemoveSong(code, entryID, deviceID, role string) (*domain.Snapshot, error)
	Reorder(code string, ids []string, deviceID, role string) (*domain.Snapshot, error)
	ClearQueue(code, deviceID, role string) (*domain.Snapshot, error)
	LockQueue(code string, locked bool, deviceID, role string) (*domain.Snapshot, error)
}

type PlaybackCommands interface {
	Playback(code, action, deviceID, role string) (*domain.Snapshot, error)
	TVEnded(code, entryID string, generation uint64, deviceID string)
	TVError(code, entryID, reason string, generation uint64)
	TVProgress(code, entryID string, generation uint64, position, duration float64)
}

type VoteCommands interface {
	StartSkipVote(code, deviceID, role string) (*domain.Snapshot, error)
	CastSkipVote(code, deviceID, role string) (*domain.Snapshot, error)
	CancelSkipVote(code, role, deviceID string) (*domain.Snapshot, error)
}

type CatalogQueries interface {
	Search(ctx context.Context, query string) ([]domain.SearchResult, bool, error)
}

// KaraokeService est la façade unique injectée dans les adapteurs HTTP et WebSocket.
type KaraokeService interface {
	RoomCommands
	QueueCommands
	PlaybackCommands
	VoteCommands
	CatalogQueries
}
```

#### Ports Secondaires (`internal/ports/output.go`)

```go
type RoomRepository interface {
	Count() int
	SaveNew(room domain.Room) error
	Get(code string) (domain.Room, bool)
	FindByCompactCode(compact string) (string, bool)
	FindByHostToken(token string) (string, bool)
	Update(code string, fn func(domain.Room) (domain.Room, error)) (domain.Room, error)
	Delete(code string) bool
	ListCodes() []string
}

type DeviceRepository interface {
	Touch(deviceID string, host bool, now time.Time)
	SetConnected(deviceID string, connected bool, now time.Time)
	Get(deviceID string) (domain.DeviceInfo, bool)
	CountActiveGuests(excludeID string, now time.Time, window time.Duration) int
	FilterActive(deviceIDs []string, now time.Time, window time.Duration) []string
	All() map[string]domain.DeviceInfo
}

type EventPublisher interface {
	PublishRoomUpdated(code string)
	PublishRoomEnded(code string)
}

type VideoCatalog interface {
	Search(ctx context.Context, query string) ([]domain.SearchResult, bool, error)
	Lookup(ctx context.Context, videoID string) (domain.VideoMeta, error)
}

type EmbedProber interface {
	ProbeOEmbed(ctx context.Context, videoID string) (domain.VideoMeta, error)
}

type Clock interface {
	Now() time.Time
}

type IDGenerator interface {
	NewID(prefix string) string
	NewRoomCode() string
	NewHostToken() string
	NewPIN() (string, error)
}

type QRGenerator interface {
	EncodeSVG(url string) ([]byte, error)
}

type SessionSigner interface {
	Sign(value string) string
	Verify(signed string) (string, bool)
}

type RateLimiter interface {
	Allow(key string, limit int, window time.Duration) (ok bool, retryAfter int)
}
```

### 4.4 Injection de Dépendances dans `cmd/server/wire.go`

Plus aucun `SetCallbacks` après coup :
1. On instancie les adapteurs secondaires sans état (`clock.NewSystemClock()`, `crypto.NewIDGenerator()`, `crypto.NewHMACSigner(secret)`, `qr.NewRSCGenerator()`, `memory.NewRateLimiter()`).
2. On instancie `memory.NewRoomRepository(maxRooms)` et `memory.NewDeviceRepository()`.
3. On instancie `prober := invidious.NewOEmbedProber(httpClient, oembedURL)`, puis `rawCatalog := invidious.NewClient(opts, prober)`, puis le décorateur `catalog := invidious.NewCachedCatalog(rawCatalog, prober, cacheOpts, clk)`.
4. On instancie `wsHub := ws.NewHub(log)` qui implémente `ports.EventPublisher`.
5. On instancie les sous-services `roomSvc`, `queueSvc`, `playbackSvc`, `voteSvc`, `lifecycleSvc`, `catalogSvc` en leur injectant `(roomRepo, deviceRepo, wsHub, catalog, clk, ids, cfg)`, puis la façade `facade := usecase.NewFacade(...)`.
6. On passe `facade` au `wsHub.BindService(facade)` (ou au `ws.NewHandler(facade, wsHub, signer, log)`) et à `http.NewServer(facade, wsHandler, qrGen, signer, limiter, staticFS, httpCfg, log)`.

---

## 5. Architecture Cible Frontend (Svelte 5 / TypeScript / CSS)

### 5.1 Arborescence des fichiers Frontend (tous <= 300 lignes, fonctions <= 40 lignes)

```text
web/src/
  main.ts                                    (~20L) Instancie le conteneur DI de production et monte App.svelte
  App.svelte                                 (~65L) Injecte le conteneur DI dans le Context Svelte et gère le routeur

  styles/                                    # Découpage de app.css (480L) en 3 fichiers <= 200L
    tokens.css                               (~80L) Variables CSS (:root couleurs, ombres, rayons, polices) + animations @keyframes
    base.css                                 (~140L) Reset, body, typographie, utilitaires de layout (.row, .stack, .spread, .clamp-2)
    components.css                           (~190L) Boutons (.btn), inputs (.input), cartes (.card), badges (.badge), toasts, modales

  domain/                                    # Logique métier et types purs (0 dépendance Svelte ou DOM)
    models.ts                                (~105L) Types Snapshot, RoomView, QueueView, NowPlayingView, SkipVoteView, SearchResult
    playback.ts                              (~70L) Calculs purs : computeElapsedNow, computeProgressPercent, secondsUntil
    room_code.ts                             (~45L) Normalisation de code de salle, validation d'URL YouTube/surnom côté client
    player_policy.ts                         (~65L) Constantes et règles pures de décision d'autoplay / fallback muet

  ports/                                     # Interfaces TypeScript (Driving & Driven Ports)
    api_port.ts                              (~55L) Interface RoomApiPort (createRoom, getRoom, search, addSong, playback, votes...)
    socket_port.ts                           (~35L) Interface RoomSocketPort & RoomSocketHandlers
    player_port.ts                           (~45L) Interface VideoPlayerPort & PlayerEvents
    browser_port.ts                          (~35L) Interfaces StoragePort, ClipboardPort, FullscreenPort
    context.ts                               (~45L) Interface AppDependencies + helpers setAppContext / getAppContext (Svelte Context DI)

  application/                               # Stores et contrôleurs de cas d'usage injectés avec les Ports
    i18n/
      index.ts                               (~95L) Store lang, store dérivé tr, helpers errorText et noticeText
      en.ts                                  (~215L) Dictionnaire anglais (< 300L)
      fr.ts                                  (~215L) Dictionnaire français (< 300L)
    welcome_controller.ts                    (~95L) État et actions de la page d'accueil (création salle, résolution code, navigation)
    guest_controller.ts                      (~185L) Store et actions de la vue invité (recherche, ajout, vote, skip-vote, PIN unlock)
    tv_controller.ts                         (~195L) Store et actions de la vue TV (session, télémétrie player, contrôles hôte, raccourcis)

  adapters/                                  # Implémentations concrètes des Ports secondaires
    http_api_adapter.ts                      (~130L) Implémente RoomApiPort via fetch + classe ApiError
    ws_socket_adapter.ts                     (~110L) Implémente RoomSocketPort avec reconnexion exponentielle et ping
    youtube_iframe_adapter.ts                (~185L) Implémente VideoPlayerPort autour de window.YT.Player
    browser_adapters.ts                      (~75L) Implémentations localStorage, navigator.clipboard, document.fullscreen

  ui/                                        # Composants Svelte déclaratifs (tous <= 300L, fonctions <= 40L)
    common/
      ConnectionDot.svelte                   (~30L) Indicateur d'état WebSocket
      LanguageToggle.svelte                  (~25L) Sélecteur EN / FR
      Modal.svelte                           (~65L) Modale accessible avec focus trap et touche Escape
    welcome/
      WelcomeView.svelte                     (~180L) Page d'accueil assemblant Hero, JoinCard et FeatureGrid
      WelcomeStagePreview.svelte             (~170L) Aperçu visuel 3D de la scène TV et carte QR
      WelcomeFeatures.svelte                 (~120L) Grille des 3 fonctionnalités + pied de page confidentialité
    guest/
      GuestRoomView.svelte                   (~185L) Coquille de la vue invité (orchestre Header, SearchTab, QueueTab, HostBar, Modals)
      GuestHeader.svelte                     (~155L) En-tête collant : marque, code salle, compteur invités, bouton cadenas et onglets
      GuestSearchTab.svelte                  (~195L) Champ de recherche, liste des résultats karaoké et formulaire de collage d'URL
      GuestNowPlayingCard.svelte             (~185L) Carte "En cours" avec progression et zone de vote pour passer (skip vote)
      GuestQueueList.svelte                  (~175L) Liste des titres à suivre avec boutons d'upvote et actions de réordonnancement hôte
      GuestHostBar.svelte                    (~120L) Barre inférieure des contrôles animateur déverrouillés + modale de gestion
      GuestModals.svelte                     (~180L) Modale de saisie du surnom (ajout) et modale de déverrouillage PIN
    tv/
      TvRoomView.svelte                      (~195L) Coquille TV (orchestre Stage, TopBar, Overlays, BottomBar, QueueDrawer)
      TvPlayer.svelte                        (~185L) Composant lecteur vidéo utilisant VideoPlayerPort, puce "Tap for sound" et overlay "Tap to resume"
      TvTopBar.svelte                        (~175L) Code de salle géant, panneau QR + PIN et boutons d'outils (langue, lien, vue invité)
      TvStageOverlays.svelte                 (~165L) Overlays centraux (Démarrer, En attente, Compte à rebours 5s, Bannière de vote skip)
      TvBottomBar.svelte                     (~195L) Résumé chanson en cours, boutons Play/Pause/Next/Mute/Fullscreen/Queue et aperçu "À suivre"
      TvQueueDrawer.svelte                   (~210L) Tiroir latéral de file d'attente avec drag-and-drop, verrouillage, rotation PIN et fin de salle
```

### 5.2 Injection de Dépendances côté Frontend (Svelte Context)

Dans `web/src/main.ts`, les adapteurs concrets sont instanciés une seule fois et passés en prop `deps` à `App.svelte`, qui appelle `setAppContext(deps)` :

```ts
// web/src/ports/context.ts
export interface AppDependencies {
  api: RoomApiPort;
  socket: RoomSocketPort;
  createPlayer: (container: HTMLElement, events: PlayerEvents) => VideoPlayerPort;
  clipboard: ClipboardPort;
  fullscreen: FullscreenPort;
}
```

Chaque vue (`WelcomeView.svelte`, `GuestRoomView.svelte`, `TvRoomView.svelte`) récupère `const deps = getAppContext()` et instancie son contrôleur applicatif (`createGuestController(code, deps)`), ce qui permet de tester toute la logique de vue avec des ports simulés en mémoire sans DOM ni réseau.

---

## 6. Harnais de Vérification Automatique (`internal/architecture/architecture_test.go`)

Le test `go test ./internal/architecture/...` (exécuté aussi par `go test ./...`) garantit mécaniquement le respect de toutes les règles :

1. **`TestFileLineLimits`** :
   - Parcourt tous les fichiers `.go`, `.ts`, `.svelte`, `.css` du dépôt (hors `web/node_modules`, `web/dist`).
   - Vérifie que le **nombre total de lignes brutes <= 300** et le **SLOC <= 300**.
   - Échoue avec le chemin exact et le nombre de lignes en cas de dépassement.
2. **`TestGoFunctionLineLimits`** :
   - Utilise `go/parser` et `go/ast` (`*ast.FuncDecl` et `*ast.FuncLit`) avec `token.FileSet`.
   - Calcule `endLine - startLine + 1` (brut) ainsi que le SLOC interne de chaque fonction/méthode Go.
   - Échoue si une fonction dépasse **40 lignes** (brut ou SLOC).
3. **`TestFrontendFunctionLineLimits`** :
   - Analyse tous les fichiers `.ts` et les sections `<script>` des fichiers `.svelte`.
   - Détecte chaque déclaration de fonction/méthode et vérifie par équilibrage d'accolades que sa longueur brute et son SLOC sont **<= 40 lignes**.
4. **`TestGoHexagonalDependencies`** :
   - Utilise `go/parser` pour inspecter les imports de chaque package sous `internal/` et `cmd/`.
   - Vérifie la matrice d'isolation (section 3.2) :
     - `internal/domain` n'importe aucun package hors stdlib pure.
     - `internal/ports` n'importe que `internal/domain` et stdlib.
     - `internal/usecase` n'importe que `internal/domain`, `internal/ports` et stdlib.
     - `internal/adapters/*` n'importe jamais `internal/usecase` ni un autre adapteur.
5. **`TestFrontendHexagonalDependencies`** :
   - Inspecte les clauses `import ... from ...` des fichiers sous `web/src/{domain,ports,application,adapters,ui}`.
   - Vérifie que `domain` est pur, que `application` et `ui` n'importent jamais `adapters`, et que `adapters` n'importe jamais `ui` ni `application`.

---

## 7. Plan de Migration et Critères d'Acceptation

### 7.1 Ordre de réalisation recommandé
1. **Étape 1 — Domaine Pur & Ports Go** : Extraire `internal/domain/*` (fonctions pures <= 40L) et `internal/ports/{input,output}.go` avec leurs tests unitaires.
2. **Étape 2 — Adapteurs Secondaires Go** : Implémenter `memory`, `invidious` (client + prober + cache decorator), `crypto`, `qr`, `clock` sous `internal/adapters/secondary/`.
3. **Étape 3 — Use Cases & Façade Go** : Implémenter `internal/usecase/*` et la façade unique `KaraokeService`.
4. **Étape 4 — Adapteurs Primaires HTTP/WS & Composition Root** : Découper `internal/adapters/primary/http/*`, `internal/adapters/primary/ws/*` et câbler `cmd/server/{main,config,wire}.go`.
5. **Étape 5 — Styles, i18n & Architecture Hexagonale Frontend** : Scinder `styles/*.css` et `application/i18n/{index,en,fr}.ts`, créer `domain/`, `ports/`, `adapters/`, `application/` en TypeScript.
6. **Étape 6 — Découpage des Composants Svelte** : Remplacer les 3 vues monolithiques et `Player.svelte` par les sous-composants sous `web/src/ui/{common,welcome,guest,tv}/` (chacun <= 300L, fonctions <= 40L).
7. **Étape 7 — Harnais d'Architecture & Validation E2E** : Ajouter `internal/architecture/architecture_test.go`, builder le binaire et valider tous les tests unitaires, `svelte-check` et les 4 suites Playwright.

### 7.2 Critères d'acceptation finaux
- [ ] Aucun fichier `.go`, `.ts`, `.svelte`, `.css` ne dépasse **300 lignes** (brut et SLOC).
- [ ] Aucune fonction/méthode Go ou TypeScript/Svelte ne dépasse **40 lignes** (brut et SLOC).
- [ ] `internal/architecture/architecture_test.go` passe à 100 % (tailles + isolation DDD Go & TS).
- [ ] `go test -race ./...`, `go vet ./...` et `gofmt -l .` passent sans erreur ni avertissement.
- [ ] `cd web && npm run check && npm run build` passe avec 0 erreur et 0 warning.
- [ ] Les 4 suites E2E Playwright (`flow.mjs`, `autoplay.mjs`, `audit.mjs`, `voting.mjs`) passent intégralement (37/37 vérifications).
