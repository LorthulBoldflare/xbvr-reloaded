// WAMP-over-websocket connection to the server (proxied at /ws/, realm
// "default"). Auth rides on the same credentials as the rest of the UI
// (browser-cached Basic auth or the player-session cookie).

import { Wampy } from 'wampy'
import { useMessagesStore } from '../store/messages'
import { useRemoteStore } from '../store/remote'
import { usePreviewsStore } from '../store/previews'
import type { GetStateResponse, ObjectState } from '../api/types'
import { queryClient } from '../queryClient'

type ArgsDict = Record<string, any>

let started = false

export function startSocket() {
  if (started) return
  started = true

  const ws = new Wampy('/ws/', { realm: 'default' })

  ws.subscribe('service.log', (dataArr: any) => {
    const args: ArgsDict = dataArr.argsDict ?? {}
    const msg: string = args.message ?? ''
    if (args.level === 'error') console.error(msg)
    else if (import.meta.env.DEV && args.level === 'debug') console.debug(msg)
    else if (import.meta.env.DEV && args.level === 'info') console.info(msg)

    const task: string | undefined = args.data?.task
    if (task === 'scrape') {
      useMessagesStore.getState().setLastScrapeMessage(msg)
    } else if (task === 'scraperProgress') {
      if (msg === 'DONE') {
        // old UI clears the whole list
        for (const id of useMessagesStore.getState().runningScrapers) {
          useMessagesStore.getState().removeRunningScraper(id)
        }
      }
      if (args.data?.started) useMessagesStore.getState().addRunningScraper(args.data.scraperID)
      if (args.data?.completed) useMessagesStore.getState().removeRunningScraper(args.data.scraperID)
    } else if (task === 'rescan') {
      useMessagesStore.getState().setLastRescanMessage(msg)
    }
  })

  ws.subscribe('lock.change', (dataArr: any) => {
    const args: ArgsDict = dataArr.argsDict ?? {}
    useMessagesStore.getState().setLock(args.name, !!args.locked)
  })

  ws.subscribe('state.change.optionsStorage', () => {
    queryClient.invalidateQueries({ queryKey: ['optionsStorage'] })
  })

  ws.subscribe('state.change.migration', (dataArr: any) => {
    const migration = (dataArr.argsDict ?? {}) as ObjectState['migration']
    queryClient.setQueryData<GetStateResponse>(['optionsState'], (state) =>
      state
        ? { ...state, currentState: { ...state.currentState, migration } }
        : state
    )
  })

  ws.subscribe('options.previews.previewReady', (dataArr: any) => {
    const args: ArgsDict = dataArr.argsDict ?? {}
    usePreviewsStore.getState().setPreviewFn(args.previewFn)
  })

  ws.subscribe('options.previews.queue', (dataArr: any) => {
    usePreviewsStore.getState().setQueue(dataArr.argsDict ?? {})
  })

  ws.subscribe('remote.state', (dataArr: any) => {
    useRemoteStore.getState().process(dataArr.argsDict ?? {})
  })
}
