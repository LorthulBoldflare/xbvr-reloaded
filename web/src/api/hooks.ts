import { useQuery } from '@tanstack/react-query'
import { api } from '../api/client'
import type { GetStateResponse, GetStorageResponse, WebOptions } from '../api/types'

// Global server state (config + runtime state). Used by options pages,
// card-appearance settings, feature gates, etc.
export function useOptionsState() {
  return useQuery({
    queryKey: ['optionsState'],
    queryFn: ({ signal }) => api.get<GetStateResponse>('/options/state', { signal }),
    staleTime: 60_000,
    refetchInterval: 5 * 60_000,
    refetchIntervalInBackground: false
  })
}

export function useWebOptions(): WebOptions | undefined {
  return useOptionsState().data?.config?.web
}

export function useOptionsStorage() {
  return useQuery({
    queryKey: ['optionsStorage'],
    queryFn: () => api.get<GetStorageResponse>('/options/storage'),
    staleTime: 30_000
  })
}

export function useVersionCheck() {
  return useQuery({
    queryKey: ['versionCheck'],
    queryFn: () => api.get<{ current_version: string; latest_version: string; update_notify: boolean }>(
      '/options/version-check',
      { toastOnError: false }
    ),
    staleTime: 6 * 60 * 60_000
  })
}
