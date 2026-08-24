# XBVR Reloaded
### The ultimate tool for managing your VR porn library.


## Features

- Automatically match title, tags, cast, cover image, and more to your videos
- Support for all the most popular VR sites: BadoinkVR, CzechVR Network, DDFNetworkVR, MilfVR, NaughtyAmericaVR, SexBabesVR, StasyQVR, TmwVRnet, VirtualRealPorn, VirtualTaboo, VRBangers, VRHush, VRLatina, WankzVR and many studios on SexLikeReal
- Directly supports DeoVR and HereSphere VR players via API
- Built-in DLNA streaming server compatible with popular VR players (Pigasus, Skybox, Mobile Station VR)
- Sleek and simple web UI
- Browse your content by cast, site, tags, and release date
- Available for Windows, macOS, Linux (including ARM builds for RaspberryPi)

## XBVR vs XBVR Reloaded

XBVR Reloaded is a fork of [XBVR](https://github.com/xbapps/xbvr), currently under active development. Legend: ✅ full support · 🟡 partial · ❌ not available.

| Feature | XBVR | XBVR Reloaded | Difference explained |
|---------|:----:|:-------------:|----------------------|
| Scene scrapers for major VR sites | ✅ | ✅ | No difference |
| DeoVR / HereSphere player API support | ✅ | ✅ | No difference |
| Built-in DLNA streaming server | ✅ | ✅ | No difference |
| Content bundle backup & restore | ✅ | ✅ | No difference |
| **Correct file creation dates on macOS** | ❌ | ✅ | XBVR trusted the filesystem birth time blindly and could store epoch dates (1 Jan 1970) on macOS; Reloaded validates the birth time, falls back to the modification time, and repairs previously stored bad timestamps on rescan |
| Video preview generation | 🟡 | ✅ | XBVR's pipeline is rudimentary; Reloaded adds a proper job queue with live progress and stop control, GPU hardware acceleration (NVENC, QSV, VAAPI, AMF, Vulkan; VideoToolbox opt-in), and deletion of individual scene previews |
| MCP (Model Context Protocol) endpoint | ❌ | ✅ | LLM/AI clients can trigger `rescan_storage`, `scrape_scene`, `generate_previews`, and `match_file` via `/mcp` (see below) |
| Encrypted bundle export | ❌ | ✅ | Credentials in exported bundles are encrypted with a user-supplied password |
| Bundle validation on restore | ❌ | ✅ | Malformed bundles are rejected with HTTP 400 before any data is imported |
| Scene metadata sync & smarter matching | 🟡 | ✅ | Duration sorting support in metadata sync and apostrophe-tolerant title matching |
| Scene sorting options | 🟡 | ✅ | Reloaded adds sorting by video duration (↑/↓, scenes without a known duration always sort last), and the "scene added date" sort now keys off the oldest file creation time instead of the newest, so ordering no longer shifts when files are re-added |
| SSRF protection on outbound URLs | ❌ | ✅ | User-supplied URLs are validated at config and request time; scraped outbound URLs are sanitized and secrets masked in the UI |
| Server & protocol hardening | ❌ | ✅ | HTTP server, session handling, DLNA/UPnP/SSDP services, and binary downloads hardened |
| Database & UI performance | ❌ | ✅ | Connection pooling, query batching, index/auth state caching, additional DB indexes, and route-level code splitting |
| Self-contained binary | ❌ | ✅ | Release migration data is bundled into the binary via `go:embed` |
| UI polish | ❌ | ✅ | Confirmation dialogs for destructive actions, accessibility improvements (dialog roles, responsive modals), and translatable dialog strings |
| Internal cleanup | 🟡 | ✅ | Superficial changes only: shared UI API layer, deduplicated components, dead-code removal, and lint tooling fixes — no user-visible behavior change |

## Download

Currently only source code is made available.
Please note that during the first run XBVR automatically installs `ffprobe` and `ffmpeg` codecs from [ffbinaries site](https://ffbinaries.com/downloads). If `ffmpeg`/`ffprobe` are already installed, XBVR uses those instead of downloading its own: they are looked up on the system `PATH` as well as in well-known locations such as `/opt/homebrew/bin`, `/usr/local/bin` and `${HOME}/.local/bin` (symlinks are followed). A system ffmpeg with hardware encoders (e.g. NVENC, QSV, VAAPI, AMF, Vulkan) enables GPU-accelerated preview generation. VideoToolbox on macOS is opt-in: set `XBVR_PREVIEW_VIDEOTOOLBOX=1` to enable it.

## Quick Start

Once launched, web UI is available at `http://127.0.0.1:9999`.

Before anything else, you must allow the app to scan sites and populate its scene metadata library. Click through to Options -> Scene Data and "Run scraper". This can take several minutes to complete. Wait for it to finish, and then go to Options -> Folders and add the folders where your video files are stored.

When it's all done, you should see your media not only in web UI, but also through DLNA server in your favourite VR player.

Enjoy!


## Development

Make sure you have following installed:

- Go 1.24
- Node.js 22.x
- Yarn 1.17.x
- air (run `go install github.com/cosmtrek/air@latest` outside project directory)

Once all of the above is installed, running `yarn dev` from project directory launches file-watchers providing livereload for both Go and JavaScript.

### How To

#### Add specific filter to DeoVR
* On the XBVR scenes page, create a filter (cast, site, tags, etc.) and sort order, then create a "saved search" (see top left) and check "use as DeoVR list". 
* Inside DeoVR you will now see your saved search listed
#### Keyboard Shortcuts
* Global
? - Quick Find  
* Details Pane  
o - previous scene  
p - next scene  
e - edit scene  
w - toggle watchlist  
f - toggle favourite  
W - toggle Watched status (Capital W)  
g - toggles gallery / video window  
esc - closes details pane  
left arrow - cycles backwards in gallery / skips backwards in video  
right arrow - cycles forward in gallery / skips forward in video
* File Match Pane  
o - previous file  
p - next file  
left arrow - next page of search results  
right arrow - previous  page of search results  
esc - closes matching pane  
* Actor List  
o or left arrow - previous page of actors  
p or right arrow - next page of actors  
* Actor Details  
o - previous actor  
p - next actor  
left arrow - cycles backwards in gallery  
right arrow - cycles forward in gallery  
esc - closes details pane

#### using Command Line Arguments/Environment Variables
| Command line parameter | Environment Variable | Type | Description |
|------------------------|--------------|------|-------------|
| `--enableLocalStorage` | | boolean | Use local folder to store application data|
|	`--app_dir` | XBVR_APPDIR | String | path to the application directory|
|	`--cache_dir` | XBVR_CACHEDIR | String | path to the temporary scraper cache directory|
|	`--imgproxy_dir` | XBVR_IMAGEPROXYDIR | String | path to the imageproxy directory|
|	`--search_dir` | XBVR_SEARCHDIR | String | path to the Search Index directory|
|	`--preview_dir` | XBVR_VIDEOPREVIEWDIR | String | path to the Scraper Cache directory|
|	`--scriptsheatmap_dir` | XBVR_SCRIPTHEATMAPDIR | String| path to the scripts_heatmap directory|
|	`--myfiles_dir` | XBVR_MYFILESDIR | String | path to the myfiles directory for serving users own content (eg images|
|	`--databaseurl` | DATABASE_URL | String | override default database path|
|	`--web_port` | XBVR_WEB_PORT | Int | override default Web Page port 9999|
|	`--ws_addr` | XBVR_WS_ADDR | String | override default Websocket address from the default 0.0.0.0:9998|
|	`--db_connection_pool_size` | DB_CONNECTION_POOL_SIZE | Int | sets the connection pool size for mariadb databases|
|	`--concurrent_scrapers` | CONCURRENT_SCRAPERS | Int | set the number of scrapers that run concurrently default 9999|
| | UI_USERNAME | String | set the username for UI authentication
| | UI_PASSWORD | String | set the password for UI authentications

#### MCP (Model Context Protocol) endpoint

XBVR exposes an MCP endpoint (Streamable HTTP transport) at `/mcp` on the web port, so LLM clients can trigger common maintenance tasks. Available tools:

| Tool | Description |
|------|-------------|
| `rescan_storage` | Rescan all storage folders (same as Options → Storage → "Rescan all folders") |
| `scrape_scene` | Scrape a single scene from its URL and store it (same as Options → Create/Import scene → "Scrape a scene", including confirming the scene detail popup). Arguments: `url` (required), `scene_id` (only for wetvr.com scenes). Returns the scene-id determined by the scraper; if the scene is already present, its scene-id is returned. If the scene could not be scraped, the call fails with an error describing why. Note: scraping a new scene can take tens of seconds. |
| `generate_previews` | Start generating video preview clips (same as Options → Previews) |
| `match_file` | Match a file on disk to a scraped scene (same as the assign action on the Files page, but without fuzzy matching). Arguments: `filename` (exact filename on disk), `scene_id` (as returned by `scrape_scene`). Both must match exactly one record; the file must exist on disk and must not be matched to a different scene. Typical workflow: `scrape_scene` → download the file → `rescan_storage` → `match_file`. |

**Authentication:** when UI authentication is enabled (`UI_USERNAME`/`UI_PASSWORD` set), requests must send an `Authorization: Bearer <token>` header where the token is the concatenation of the UI username and UI password. For example, with username `UserA` and password `Password123`, the token is `UserAPassword123`. When UI auth is disabled, the endpoint is open.

Example client configuration:

```json
{
  "mcpServers": {
    "xbvr": {
      "url": "http://localhost:9999/mcp",
      "headers": {
        "Authorization": "Bearer UserAPassword123"
      }
    }
  }
}
```
