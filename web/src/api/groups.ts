import { api } from './client'

// AKA groups (actor aliases) and tag groups. Response shapes are loose on the
// server ({status, akas} / {status, tag_group}); mirror the old UI's usage.

export interface AkaOpResponse {
  status: string
  akas: { aka_actor: { name: string } }
}

export interface TagGroupOpResponse {
  status: string
  tag_group: { name: string; tag_group_tag: { name: string }; tags: { name: string }[] }
}

export const akaApi = {
  create: (actorList: string[]) => api.post<AkaOpResponse>('/aka/create', { actorList }),
  add: (actorList: string[]) => api.post<AkaOpResponse>('/aka/add', { actorList }),
  remove: (actorList: string[]) => api.post<AkaOpResponse>('/aka/remove', { actorList }),
  delete: (name: string) => api.post<AkaOpResponse>('/aka/delete', { name })
}

export const tagGroupApi = {
  create: (name: string, tagList: string[]) => api.post<TagGroupOpResponse>('/tag_group/create', { name, tagList }),
  add: (tagList: string[]) => api.post<TagGroupOpResponse>('/tag_group/add', { tagList }),
  remove: (tagList: string[]) => api.post<TagGroupOpResponse>('/tag_group/remove', { tagList }),
  rename: (name: string, tagList: string[]) => api.post<TagGroupOpResponse>('/tag_group/rename', { name, tagList }),
  delete: (name: string) => api.post<TagGroupOpResponse>('/tag_group/delete', { name }),
  get: (name: string) => api.get<TagGroupOpResponse>(`/tag_group/${encodeURIComponent(name)}`)
}
