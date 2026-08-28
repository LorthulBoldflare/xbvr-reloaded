// TypeScript mirrors of the server JSON models (pkg/models, pkg/api).
// Field names match the JSON tags exactly.

export interface Scene {
  id: number
  created_at: string
  updated_at: string
  scene_id: string
  title: string
  scene_type: string
  scraper_id: string
  studio: string
  site: string
  tags: Tag[]
  cast: Actor[]
  filenames_arr: string
  images: string
  file: File[] | null
  duration: number
  synopsis: string
  release_date: string
  release_date_text: string
  cover_url: string
  scene_url: string
  members_url: string
  is_multipart: boolean
  star_rating: number
  favourite: boolean
  watchlist: boolean
  wishlist: boolean
  is_available: boolean
  is_accessible: boolean
  is_watched: boolean
  is_scripted: boolean
  cuepoints: SceneCuepoint[] | null
  history: History[] | null
  added_date: string
  last_opened: string
  total_file_size: number
  total_watch_time: number
  has_preview: boolean
  needs_update: boolean
  edits_applied: boolean
  trailer_type: string
  trailer_source: string
  passthrough: string
  trailerlist: boolean
  is_subscribed: boolean
  is_hidden: boolean
  legacy_scene_id: string
  script_published: string
  ai_script: boolean
  human_script: boolean
  description: string
  _score?: number
  alternate_source?: ExternalReferenceLink[]
}

export interface SceneImage {
  url: string
  type: string // 'cover' | 'gallery'
  orientation?: string
}

export interface SceneCuepoint {
  id: number
  name: string
  time_start: number
  time_end?: number
  track?: number | null
  rating?: number
}

export interface History {
  id: number
  time_start: string
  time_end: string
  duration: number
}

export interface Tag {
  id: number
  name: string
  count?: number
  clean?: string
}

export interface Actor {
  id: number
  name: string
  scenes?: Scene[]
  count: number
  avail_count: number
  image_url: string
  image_arr: string
  star_rating: number
  favourite: boolean
  watchlist: boolean
  birth_date: string
  nationality: string
  ethnicity: string
  eye_color: string
  hair_color: string
  height: number
  weight: number
  cup_size: string
  band_size: number
  waist_size: number
  hip_size: number
  breast_type: string
  start_year: number
  end_year: number
  tattoos: string
  piercings: string
  biography: string
  aliases: string
  gender: string
  urls: string
  scene_rating_average?: string
  aka_groups?: Aka[]
}

export interface Aka {
  id: number
  name: string
  actors?: Actor[]
}

export interface ActorLink {
  url: string
  type: string
}

export interface File {
  id: number
  created_at: string
  updated_at: string
  volume_id: number
  path: string
  filename: string
  size: number
  oshash: string
  created_time: string
  updated_time: string
  type: string // 'video' | 'script' | 'subtitles' | 'hsp'
  scene_id: number
  video_width: number
  video_height: number
  video_bitrate: number
  video_avgfps_val: number
  video_codec_name: string
  duration: number
  projection: string
  has_alpha: boolean
  has_heatmap: boolean
  is_selected_script: boolean
  is_exported: boolean
}

export interface Volume {
  id: number
  type: string // 'local' | 'putio'
  path: string
  last_scan: string
  is_available: boolean
  file_count: number
  unmatched_count: number
  total_size: number
}

export interface Playlist {
  id: number
  name: string
  ordering: number
  is_system: boolean
  is_deo_enabled: boolean
  is_smart: boolean
  playlist_type: string
  search_params: string
}

export interface Site {
  id: string
  name: string
  avatar_url: string
  is_builtin: boolean
  is_enabled: boolean
  last_update: string
  subscribed: boolean
  has_scraper: boolean
  limit_scraping: boolean
  master_site_id: string
  matching_params: string
  scrape_stash: boolean
  scene_count: number
}

export interface Scraper {
  id: string
  name: string
  avatar_url: string
}

export interface ExternalReferenceLink {
  id: number
  external_source: string
  external_id: string
  internal_table: string
  internal_db_id: number
  internal_name_id: string
  match_type: number
  url: string
  site_icon: string
  external_data: string
}

// ---- API request/response shapes ----

export interface SceneFilters {
  dlState: 'any' | 'available' | 'downloaded' | 'missing' | 'hidden'
  isAvailable: boolean | null
  isAccessible: boolean | null
  isHidden: boolean
  isWatched: boolean | null
  lists: string[]
  cast: string[]
  sites: string[]
  tags: string[]
  cuepoint: string[]
  attributes: string[]
  volume: number
  releaseMonth: string
  sort: string
  // client-only: never sent to the server
  showUnmatched?: boolean
}

export interface RequestSceneList extends Omit<SceneFilters, 'showUnmatched'> {
  offset: number
  limit: number
}

export interface ResponseSceneList {
  results: number
  scenes: Scene[]
  count_any: number
  count_available: number
  count_downloaded: number
  count_not_downloaded: number
  count_hidden: number
}

export interface ResponseGetFilters {
  cast: string[]
  tags: string[]
  sites: string[]
  release_month: string[]
  volumes: Volume[]
  attributes: string[]
  cuepoints: string[]
}

export interface ResponseGetScenes {
  results: number
  scenes: Scene[]
}

export interface ResponseActorList {
  results: number
  actors: Actor[]
  count_any: number
  count_available: number
  count_downloaded: number
  count_not_downloaded: number
  count_hidden: number
  offset: number
}

export interface ResponseGetActorFilters {
  cast: string[]
  sites: string[]
  attributes: string[]
}

export interface ActorFilters {
  dlState: 'any' | 'available' | 'downloaded' | 'missing'
  isAvailable: boolean | null
  isAccessible: boolean | null
  lists: string[]
  cast: string[]
  sites: string[]
  tags: string[]
  attributes: string[]
  jumpTo: string
  min_age: number
  max_age: number
  min_height: number
  max_height: number
  min_weight: number
  max_weight: number
  min_count: number
  max_count: number
  min_avail: number
  max_avail: number
  min_rating: number
  max_rating: number
  min_scene_rating: number
  max_scene_rating: number
  sort: string
}

export interface FileFilters {
  state: 'all' | 'matched' | 'unmatched'
  filename: string
  createdDate: string[]
  resolutions: string[]
  framerates: string[]
  bitrates: string[]
  sort: string
}

export interface AkaResponse {
  aka_groups: Aka[]
  actors: Actor[]
  possible_akas: Actor[]
}

export interface CountryDetails {
  name: string
  code: string
}

export interface SceneSearchField {
  fieldName: string
  fieldValue: string
}

export interface AlternateSource {
  url: string
  site_icon: string
  external_source: string
  external_id: string
  external_data: string
}

export interface StashdbSceneResult {
  Url: string
  ImageUrl: string
  Performers: string[]
  Title: string
  Studio: string
  Duration: number
  Description: string
  Weight: number
  Date: string
  Id: string
}

export interface StashdbSceneSearchResponse {
  Status: string
  Results: StashdbSceneResult[]
}

export interface StashdbPerformerResult {
  Url: string
  Name: string
  Disambiguation: string
  Aliases: string[]
  Id: string
  ImageUrl: string[]
  DOB: string
  Weight: number
  Studios: { Name: string; SceneCount: number }[]
}

export interface StashdbPerformerSearchResponse {
  Status: string
  Results: StashdbPerformerResult[]
}

// ---- options / config ----

export interface WebhookConfig {
  method: string
  url: string
  headers: string
}

export interface WebhooksConfig {
  trigger_external_import: WebhookConfig
  refresh_external_import: WebhookConfig
}

export interface CronSchedule {
  enabled: boolean
  hourInterval: number
  useRange: boolean
  minuteStart: number
  hourStart: number
  hourEnd: number
  runAtStartDelay: number
}

export interface WebOptions {
  tagSort: string
  sceneHidden: boolean
  sceneWatchlist: boolean
  sceneFavourite: boolean
  sceneWishlist: boolean
  sceneWatched: boolean
  sceneEdit: boolean
  sceneDuration: boolean
  sceneCuepoint: boolean
  showHspFile: boolean
  showSubtitlesFile: boolean
  sceneTrailerlist: boolean
  showScriptHeatmap: boolean
  showAllHeatmaps: boolean
  showOpenInNewWindow: boolean
  updateCheck: boolean
  isAvailOpacity: number
  sceneCardAspectRatio: string
  sceneCardScaleToFit: boolean
  actorCardAspectRatio: string
  actorCardScaleToFit: boolean
}

export interface ObjectConfig {
  server: { bindAddress: string; port: number }
  web: WebOptions
  advanced: {
    showInternalSceneId: boolean
    showHSPApiLink: boolean
    showSceneSearchField: boolean
    stashApiKey: string
    scraperProxy: string
    scrapeActorAfterScene: boolean
    useImperialEntry: boolean
    linkScenesAfterSceneScraping: boolean
    useAltSrcInFileMatching: boolean
    useAltSrcInScriptFilters: boolean
    autoLimitScraping: boolean
    ignoreReleasedBefore: string
  }
  funscripts: { scrapeFunscripts: boolean }
  vendor: { tpdb: { apiToken: string } }
  interfaces: {
    dlna: { enabled: boolean; serviceName: string; serviceImage: string; allowedIp: string[] }
    deovr: {
      enabled: boolean
      auth_enabled: boolean
      render_heatmaps: boolean
      track_watch_time: boolean
      remote_enabled: boolean
      username: string
      password: string
    }
    heresphere: {
      allow_file_deletes: boolean
      allow_rating_updates: boolean
      allow_favorite_updates: boolean
      allow_hsp_data: boolean
      allow_tag_updates: boolean
      allow_cuepoint_updates: boolean
      allow_watchlist_updates: boolean
      multitrack_cuepoints: boolean
      multitrack_cast_cuepoints: boolean
      retain_non_hsp_cuepoints: boolean
    }
    players: { video_sort_seq: string; script_sort_seq: string; subtitle_sort_seq: string }
  }
  library: {
    preview: {
      enabled: boolean
      startTime: number
      snippetLength: number
      snippetAmount: number
      resolution: number
      extraSnippet: boolean
    }
  }
  cron: Record<string, CronSchedule>
  storage: { match_ohash: boolean; video_ext: string[] }
  webhooks: WebhooksConfig
  scraper_settings: { javr: { javrScraper: string } }
}

export interface ObjectState {
  server: { bound_ip: string[] }
  migration: { is_running: boolean; current: string; total: number; progress: number; message: string }
  web: WebOptions
  dlna: { running: boolean; images: string[]; recentIp: string[] }
  cacheSize: { images: number; previews: number; searchIndex: number }
}

export interface GetStateResponse {
  currentState: ObjectState
  config: ObjectConfig
  scrapers: Scraper[]
}

export interface GetStorageResponse {
  volumes: Volume[]
  match_ohash: boolean
  video_ext: string[]
  forbidden_video_ext: string[]
  default_video_ext: string[]
  webhooks: WebhooksConfig
}

export interface VersionCheckResponse {
  current_version: string
  latest_version: string
  update_notify: boolean
}

export interface PreviewQueueStatus {
  running: boolean
  stopping: boolean
  total: number
  completed: number
  remaining: number
  currentScene: string
}

export interface CollectorConfig {
  domain_key: string
  headers: { name: string; value: string }[]
  cookies: { name: string; value: string; domain: string; path: string; host: string }[]
  body: string
  other?: { name: string; value: string }[]
}

export interface AltSrcMatchParams {
  delay_linking_days: number
  keep_relinking_days: number
  ignore_released_before: string
  [key: string]: unknown
}
