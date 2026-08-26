import { api } from '../api/client'
import type { Scene, Site } from '../api/types'
import { useToastStore } from '../store/toasts'

// Port of the old RescrapeButton: mark needs_update, then trigger the
// appropriate scraper for the scene's source site.
export async function rescrapeScene(scene: Scene, toggleNeedsUpdate: () => void) {
  const toast = useToastStore.getState()
  toggleNeedsUpdate()

  if (scene.scraper_id && scene.needs_update) {
    await new Promise((r) => setTimeout(r, 200))
    api.get(`/task/scrape?site=${encodeURIComponent(scene.scraper_id)}`, { toastOnError: true })
    return
  }

  const url = scene.scene_url.toLowerCase()
  if (url.includes('dmm.co.jp')) {
    api.post('/task/scrape-javr', { s: 'r18d', q: scene.scene_id })
    return
  }

  const sites = await api.get<Site[]>('/options/sites')
  let site = ''
  for (const element of sites) {
    if (url.includes(element.id)) site = element.id
  }
  if (url.includes('naughtyamerica.com')) site = 'naughtyamericavr'
  if (url.includes('sexlikereal.com')) site = 'slr-single_scene'
  if (url.includes('czechvrnetwork.com')) site = 'czechvr-single_scene'
  if (url.includes('povr.com')) site = 'povr-single_scene'
  if (url.includes('vrporn.com')) site = 'vrporn-single_scene'
  if (url.includes('realvr.com')) site = 'realvr-single_scene'
  if (url.includes('vrphub.com')) site = 'vrphub-single_scene'

  if (!site) {
    toast.error('No scrapers exist for this domain')
    return
  }
  // Single scrapes can run long — fire-and-forget (old UI used timeout:false).
  api.post('/task/singlescrape', { site, sceneurl: scene.scene_url, additionalinfo: [] })
  toast.info('Scraping scene…')
}
