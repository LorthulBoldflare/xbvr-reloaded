# XBVR — Agent Guide

XBVR is a VR video library manager: Go backend (`main.go` → `pkg/`) serving two embedded web UIs plus player/DLNA APIs.

## Repository layout

- `pkg/` — Go backend. `pkg/api` REST resources (go-restful), `pkg/server` static/media serving + auth, `pkg/models` GORM models + query builders, `pkg/config` config/state (envconfig defaults, KV store), `pkg/tasks` background jobs/cron, `pkg/scrape` scrapers.
- `ui/` — **old UI** (Vue 2 + Buefy + vue-cli). Feature-complete; kept for reference/back-compat. Do not extend unless asked; do not break it.
- `web/` — **new UI** (React 18 + TypeScript strict + Vite + Tailwind v4, `@tanstack/react-query`, `zustand`, `wampy`), served at `/web/` via `web/fs.go` (embeds `web/dist`, SPA history-mode fallback). This is where new UI work happens.

## Build & test

- `mise run web` — install + build the new UI (`npm --prefix web install && npm --prefix web run build`).
- `mise run ui` / `mise run build-local` — old UI / full local binary (embeds both UIs; `go build` requires both `ui/dist` and `web/dist` to exist). Tasks `cd` to the invocation directory (`MISE_ORIGINAL_CWD`), so running from a worktree builds the worktree; `build-local` injects `main.version/commit/branch/date` from git via ldflags (version = current short commit, not the `CURRENT` default).
- `cd web && npm run build` runs `tsc --noEmit` + `vite build` — must stay clean.
- `go build -tags=json1 ./...`, `go vet ./...`, `go test ./pkg/...` — must stay clean. Add Go tests for new server behavior (see `pkg/models/model_scene_filesort_test.go` for the in-memory sqlite pattern).

## New UI (`web/`) styleguide

### Design language

Cinematic, dark-first, art-first. Purple accent, green = ok/positive, red = danger/favourite-heart. Never reintroduce the old Bulma look.

- **Tokens**: semantic CSS custom properties in `web/src/index.css`, switched via `[data-theme]` on `<html>`; exposed to Tailwind through `@theme inline` as `bg-page/surface/surface-2/surface-3`, `border-line/line-strong`, `text-fg/muted`, `*-accent/accent-strong/accent-soft/accent-fg`, `*-ok/warn/danger`. **Always use these semantic utilities — never raw color classes** (`bg-white`, `text-gray-500`, …), so light/dark both work.
- **Theme**: persisted in `localStorage` key `xbvr-theme`, applied pre-paint by the inline script in `web/index.html` (`web/src/theme.ts`).
- **Gradients**: `.brand-gradient` (purple→green) for the brand mark; `.btn-gradient` (purple) for primary buttons; subtle fixed radial purple/green tints on the page background; `.scrim` (black gradient) for text-over-artwork.
- **Shape**: tiles/cards `rounded-xl`, cards/sections `rounded-2xl`, buttons/chips `rounded-full` (pill), inputs `rounded-lg`. Hover on media tiles: `ring-2 ring-accent` + lift + deep shadow.
- **Icons**: hand-rolled inline SVGs in `web/src/components/icons.tsx` (Feather-style, 24×24, `stroke="currentColor"`). No emoji as icons — extend `icons.tsx`.

### Layout & structure

- Global chrome is a **collapsible left sidebar** (`web/src/components/Sidebar.tsx`; state persisted in `localStorage` `xbvr-sidebar`, labels hidden when collapsed). No top navbar.
- Routes (`web/src/router.tsx`, history mode under `basename="/web/"`): `/` scene grid, `/scenes/:id` scene page (string `scene_id` in URLs), `/files/:id` file page, `/actors`, `/options/:section`.
- Scene details are a **page, not a modal**: read-only first, `Edit` button toggles edit mode with management controls. Actor details stay a modal.
- The scene grid is a single card grid (`repeat(auto-fill,minmax(230px,1fr))`) — **no card-size controls, no list view**. Availability is a filter (default `available` = "Available right now"), ported exactly as `applyDlState` in `web/src/store/sceneFilters.ts`.
- Scene tiles are 16:9: media fills width, keeps natural proportions, vertically centered (overflow cropped). Tiles overlay: duration + cuepoint count top-left; top-right cluster (left→right) info badges → watched eye → rating digit → watchlist bookmark toggle (green, filled/outline) → favourite heart toggle (red, filled/outline). Unavailable scenes get a centered "Not downloaded" pill; offline-volume scenes "Storage offline".
- Unmatched files can mix into the grid as `FileCard`s behind the "Show unmatched files" filter switch; `/files/:id` shows player + metadata + "Match to scene" (no scene info, no Edit).
- Options is grouped sections (Files, Storage, …); shared chrome in `web/src/routes/options/common.tsx` (`SectionCard`, `Field`, `btnCls`, `btnPrimaryCls` (gradient), `inputCls`). Credentials (player auth + MCP token) live in **Authentication**; webhook triggers live in the sidebar **Actions** menu.

### Coding conventions

- **Server state → TanStack Query** (`api/hooks.ts`, `web/src/queryClient.ts`); **UI/overlay state → zustand** (`web/src/store/`). No prop-drilling for toasts/confirms/modals: `useToastStore`, `useUIStore` (`askConfirm`, `showActorDetails`, …).
- API calls go through `web/src/api/client.ts` (`api.get/post/put/delete`); it prefixes `/api`, JSON-encodes, toasts on error, throws `ApiError`. Use `{ toastOnError: false }` for expected-404/auxiliary calls. DELETE supports a JSON body (the Go API relies on it).
- Types mirror server JSON tags exactly (`web/src/api/types.ts`).
- Images always through `getImageURL(url, size)` (`web/src/lib/image.ts`) — it must stay byte-compatible with the old `ui/src/util/image.js` (decode-before-re-encode for pre-encoded URLs) or proxied thumbnails 404. Every `<img>` needs an `onError` fallback to the bundled placeholder.
- Filter state ↔ URL `?q=` uses the base64 codec in `web/src/lib/base64.ts` — **byte-compatible with the old UI** (`btoa(unescape(encodeURIComponent(JSON.stringify()))`), standard base64, not base64url). Saved searches (`/api/playlist`) share this payload shape across both UIs.
- Keyboard shortcuts exist on the scene grid/page and actor list — always suppress them while focus is in `input/textarea/select` or a modal/popover is open.

### Hard-won pitfalls (do not regress)

1. **`config.web` vs `currentState.web`**: read web UI options from `config.web` (envconfig defaults applied). `currentState.web` can be zero-valued → e.g. `isAvailOpacity: 0` makes covers invisible.
2. **"null" JSON strings**: the server stores some array columns as the literal string `"null"`. `JSON.parse("null")` succeeds → `null`. Every parse of a server JSON-string column must pass through an `Array.isArray` check (see `parseSceneImages`, `draftFromScene`, `EditActor`'s parse).
3. **Legacy payloads**: saved searches / shared links from the old UI may contain `null` arrays and unknown keys — always sanitize via `normalizeSceneFilters` / `normalizeActorFilters` before they enter a store.
4. **ID forms**: page URLs use the string `scene_id`; only `GET /api/scene/{id}` and `alternate_source` accept strings — path mutations (`rate`, `cuepoint`, `selectscript`, `edit`, `preview`) take the **numeric PK** (`scene.id`); `toggle` takes the string in its body. Same for actors (PK-based mutations).
5. **Long-running endpoints**: no client timeout for `POST /api/task/singlescrape`, bundle backup/restore, `GET /api/playlist`, `GET /api/scene/filters`; effectively none for `POST /api/scene/list`.
6. **Server quirks to call verbatim**: `GET /api/task/relink_alt_aource_scenes` (sic, misspelled), redacted secrets round-trip (`"***"` sentinel — the server keeps the stored value when it receives the sentinel; a missing/empty value clears it).
7. **New sort values**: `file_added_desc/asc` (default sort) are server-side additions in `pkg/models/model_scene.go` — keep them working through saved-search round-trips.
8. **Websocket**: `wampy` on `/ws/` (realm `default`); topics: `service.log`, `lock.change`, `state.change.optionsStorage`, `options.previews.previewReady`, `options.previews.queue`, `remote.state` (see `web/src/ws/socket.ts`).
9. **Actor availability filter**: `POST /api/actor/list` filters availability only via the `isAvailable`/`isAccessible` flag pair (`applyActorDlState` in `web/src/store/actorFilters.ts`, default `available`). `dlState` is deliberately ignored server-side for actors — the old UI always sends `dlState:"available"` with no control for it. "Available right now" is a live `scene_cast`+`scenes` subquery (`actorAccessibleSceneExists`), NOT `actors.avail_count`: that column excludes hidden scenes but is only refreshed on scrape/import/clean-tags (`CountActorTags`), so it backs only "Downloaded"/"Not downloaded" and the card badge.

## Server conventions (for changes touching Go)

- New API surface goes through the existing resources in `pkg/api/`; adding an endpoint requires justification (both UIs share it).
- `/web/` serving is `web/fs.go` (embed + index.html fallback); auth is the shared `authHandle`/`apiAuthFilter` — no UI-specific auth code.
- The root `/` redirect stays on `/ui/`; old UI remains functional at all times.
