import { useQueryClient } from '@tanstack/react-query'
import type { InfiniteData } from '@tanstack/react-query'
import type { ResponseSceneList, Scene } from '../api/types'

// Apply a patch to a scene wherever it appears in cached scene lists
// (infinite queries keyed ['sceneList', ...]) — optimistic toggles.
export function patchSceneInCaches(queryClient: ReturnType<typeof useQueryClient>, sceneId: number, patch: Partial<Scene>) {
  queryClient.setQueriesData<InfiniteData<ResponseSceneList>>({ queryKey: ['sceneList'] }, (data) => {
    if (!data) return data
    let changed = false
    const pages = data.pages.map((page) => {
      const index = page.scenes.findIndex((scene) => scene.id === sceneId)
      if (index === -1) return page
      changed = true
      const scenes = page.scenes.slice()
      scenes[index] = { ...scenes[index], ...patch }
      return { ...page, scenes }
    })
    if (!changed) return data
    return {
      ...data,
      pages
    }
  })
  queryClient.setQueryData<Scene>(['scene', sceneId], (old) => (old ? { ...old, ...patch } : old))
}
