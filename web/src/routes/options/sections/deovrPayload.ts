import type { ObjectConfig } from '../../../api/types'

// PUT /api/options/interface/deovr takes one merged payload covering
// deovr + heresphere + players config. Build it from the current config so
// individual sections only override their own fields.
export function buildDeovrPayload(config: ObjectConfig, overrides: Record<string, unknown> = {}) {
  const d = config.interfaces.deovr
  const h = config.interfaces.heresphere
  const p = config.interfaces.players
  return {
    enabled: d.enabled,
    auth_enabled: d.auth_enabled,
    username: d.username,
    password: d.password, // redacted sentinel "***" round-trips (server keeps stored hash)
    remote_enabled: d.remote_enabled,
    track_watch_time: d.track_watch_time,
    render_heatmaps: d.render_heatmaps,
    allow_file_deletes: h.allow_file_deletes,
    allow_rating_updates: h.allow_rating_updates,
    allow_favorite_updates: h.allow_favorite_updates,
    allow_hsp_data: h.allow_hsp_data,
    allow_tag_updates: h.allow_tag_updates,
    allow_cuepoint_updates: h.allow_cuepoint_updates,
    allow_watchlist_updates: h.allow_watchlist_updates,
    multitrack_cuepoints: h.multitrack_cuepoints,
    multitrack_cast_cuepoints: h.multitrack_cast_cuepoints,
    retain_non_hsp_cuepoints: h.retain_non_hsp_cuepoints,
    video_sort_seq: p.video_sort_seq,
    script_sort_seq: p.script_sort_seq,
    subtitle_sort_seq: p.subtitle_sort_seq,
    ...overrides
  }
}
