// Sort options for the scene list (ported from the old UI; file_added_* are
// new server-side values added for the new UI's default).
export const SCENE_SORTS: { value: string; label: string }[] = [
  { value: 'file_added_desc', label: '↓ Recently added files' },
  { value: 'file_added_asc', label: '↑ Recently added files' },
  { value: 'release_desc', label: '↓ Release date' },
  { value: 'release_asc', label: '↑ Release date' },
  { value: 'added_desc', label: '↓ File added date' },
  { value: 'added_asc', label: '↑ File added date' },
  { value: 'title_desc', label: '↓ Title' },
  { value: 'title_asc', label: '↑ Title' },
  { value: 'total_file_size_desc', label: '↓ File size' },
  { value: 'total_file_size_asc', label: '↑ File size' },
  { value: 'rating_desc', label: '↓ Rating' },
  { value: 'rating_asc', label: '↑ Rating' },
  { value: 'total_watch_time_desc', label: '↓ Watch time' },
  { value: 'total_watch_time_asc', label: '↑ Watch time' },
  { value: 'duration_desc', label: '↓ Duration' },
  { value: 'duration_asc', label: '↑ Duration' },
  { value: 'scene_added_desc', label: '↓ Scene added date' },
  { value: 'scene_updated_desc', label: '↓ Scene updated date' },
  { value: 'last_opened_desc', label: '↓ Last viewed date' },
  { value: 'last_opened_asc', label: '↑ Last viewed date' },
  { value: 'script_published_desc', label: '↓ Published script added' },
  { value: 'scene_id_desc', label: '↓ Scene Id' },
  { value: 'site_asc', label: '↑ Site' },
  { value: 'alt_src_desc', label: '↓ Linked to alternate sites' },
  { value: 'random', label: '↯ Random' }
]
