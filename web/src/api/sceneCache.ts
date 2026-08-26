import { useQueryClient } from '@tanstack/react-query'
import type { InfiniteData } from '@tanstack/react-query'
import type { ResponseSceneList, Scene } from '../api/types'

// Apply a patch to a scene wherever it appears in cached scene lists
// (infinite queries keyed ['sceneList', ...]) — optimistic toggles.
export function patchSceneInCaches(queryClient: ReturnType<typeof useQueryClient>, sceneId: number, patch: Partial<Scene>) {
  queryClient.setQueriesData<InfiniteData<ResponseSceneList>>({ queryKey: ['sceneList'] }, (data) => {
    if (!data) return data
    return {
      ...data,
      pages: data.pages.map((page) => ({
        ...page,
        scenes: page.scenes.map((s) => (s.id === sceneId ? { ...s, ...patch } : s))
      }))
    }
  })
  queryClient.setQueryData<Scene>(['scene', sceneId], (old) => (old ? { ...old, ...patch } : old))
}
