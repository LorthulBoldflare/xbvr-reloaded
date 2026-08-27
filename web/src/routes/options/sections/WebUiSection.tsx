import { useEffect, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../../../api/client'
import type { WebOptions } from '../../../api/types'
import { useOptionsState } from '../../../api/hooks'
import { useToastStore } from '../../../store/toasts'
import { Toggle } from '../../../components/Toggle'
import { SectionCard, Field, SaveButton, inputCls } from '../common'

// Web UI appearance options (scene-card button visibility, heatmaps, etc.).
export function WebUiSection() {
  const { data: state } = useOptionsState()
  const queryClient = useQueryClient()
  const toast = useToastStore.getState()
  const [form, setForm] = useState<WebOptions | null>(null)

  useEffect(() => {
    if (state?.config?.web) setForm(state.config.web)
  }, [state])

  const save = useMutation({
    mutationFn: () => api.put('/options/interface/web', form),
    onSuccess: () => {
      toast.success('Web UI settings saved')
      queryClient.invalidateQueries({ queryKey: ['optionsState'] })
    }
  })

  if (!form) return null
  const set = (k: keyof WebOptions, v: unknown) => setForm((f) => (f ? { ...f, [k]: v } : f))

  const switches: [keyof WebOptions, string][] = [
    ['sceneHidden', 'Hide button'],
    ['sceneWatchlist', 'Watchlist button'],
    ['sceneTrailerlist', 'Trailer list button'],
    ['sceneFavourite', 'Favourite button'],
    ['sceneWishlist', 'Wishlist button'],
    ['sceneWatched', 'Watched button'],
    ['sceneEdit', 'Edit button'],
    ['sceneDuration', 'Duration badge'],
    ['sceneCuepoint', 'Cuepoint badge'],
    ['showHspFile', 'HereSphere file badge'],
    ['showSubtitlesFile', 'Subtitles badge'],
    ['showScriptHeatmap', 'Funscript heatmap'],
    ['showAllHeatmaps', 'Show all heatmaps']
  ]

  return (
    <div className="space-y-4">
      <SectionCard title="Cards" actions={<SaveButton onClick={() => save.mutate()} pending={save.isPending} />}>
        <div className="grid max-w-3xl grid-cols-1 gap-2 md:grid-cols-3">
          {switches.map(([k, label]) => (
            <Toggle key={k} checked={!!form[k]} onChange={(v) => set(k, v)} label={label} />
          ))}
        </div>
        <div className="mt-4 grid max-w-3xl grid-cols-1 gap-3 md:grid-cols-2">
          <Field label={`Opacity of unavailable scenes: ${form.isAvailOpacity}%`}>
            <input
              type="range" min={0} max={100} value={form.isAvailOpacity}
              onChange={(e) => set('isAvailOpacity', Number(e.target.value))}
              className="w-full"
            />
          </Field>
          <Field label="Tag sorting">
            <select value={form.tagSort} onChange={(e) => set('tagSort', e.target.value)} className={inputCls}>
              <option value="alphabetically">Alphabetically</option>
              <option value="by-tag-count">By scene count</option>
            </select>
          </Field>
          <Field label="Scene card aspect ratio" hint="The new UI (/web/) always uses 16:9; this applies to the classic UI">
            <select value={form.sceneCardAspectRatio} onChange={(e) => set('sceneCardAspectRatio', e.target.value)} className={inputCls}>
              <option value="1:1">1:1</option>
              <option value="3:2">3:2</option>
              <option value="16:9">16:9</option>
            </select>
          </Field>
          <Field label="Actor card aspect ratio">
            <select value={form.actorCardAspectRatio} onChange={(e) => set('actorCardAspectRatio', e.target.value)} className={inputCls}>
              <option value="1:1">1:1</option>
              <option value="2:3">2:3</option>
              <option value="9:16">9:16</option>
            </select>
          </Field>
          <Toggle checked={form.sceneCardScaleToFit} onChange={(v) => set('sceneCardScaleToFit', v)} label="Scale scene cards to fit (classic UI)" />
          <Toggle checked={form.actorCardScaleToFit} onChange={(v) => set('actorCardScaleToFit', v)} label="Scale actor cards to fit" />
        </div>
      </SectionCard>

      <SectionCard title="Misc">
        <Toggle checked={form.showOpenInNewWindow} onChange={(v) => set('showOpenInNewWindow', v)} label="Tag links open in a new window" />
        <div className="mt-2">
          <Toggle checked={form.updateCheck} onChange={(v) => set('updateCheck', v)} label="Check for updates on startup" />
        </div>
        <div className="mt-3">
          <SaveButton onClick={() => save.mutate()} pending={save.isPending} />
        </div>
      </SectionCard>
    </div>
  )
}
