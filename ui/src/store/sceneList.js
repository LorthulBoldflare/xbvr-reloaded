import api from '../api'
import Vue from 'vue'
import { encodeJsonBase64, decodeJsonBase64 } from '../util/base64'

function defaultValue (v, d) {
  if (v === undefined) {
    return d
  }
  return v
}

const defaultFilterState = {
  dlState: 'available',
  cardSize: '2',  // 1 is now XS and 2 is now S

  lists: [],
  isAvailable: true,
  isAccessible: true,
  isHidden: false,
  isWatched: null,
  releaseMonth: '',
  cast: [],
  sites: [],
  tags: [],
  cuepoint: [],
  attributes: [],
  volume: 0,
  sort: 'release_desc'
}

const state = {
  items: [],
  playlists: [],
  isLoading: false,
  offset: 0,
  total: 0,
  limit: 80,
  counts: {
    any: 0,
    available: 0,
    downloaded: 0,
    not_downloaded: 0,
    hidden: 0
  },
  show_scene_id: '',
  filterOpts: {
    cast: [],
    sites: [],
    tags: [],
    volumes: []
  },
  filters: defaultFilterState
}

const getters = {
  filterQueryParams: (state) => {
    const st = Object.assign({}, state.filters)
    delete st.cardSize

    return encodeJsonBase64(st)
  },
  getQueryParamsFromObject: (state) => (payload) => {
    const st = Object.assign({}, JSON.parse(payload))
    delete st.cardSize

    return encodeJsonBase64(st)
  },
  prevScene: (state) => (currentScene) => {
    const i = state.items.findIndex(item => item.scene_id === currentScene.scene_id)
    if (i === 0) {
      return null
    }
    return state.items[i - 1]
  },
  nextScene: (state) => (currentScene) => {
    const i = state.items.findIndex(item => item.scene_id === currentScene.scene_id)
    if (i === state.items.length - 1) {
      return null
    }
    return state.items[i + 1]
  }
}

const mutations = {
  setItems (state, payload) {
    state.items = payload
  },
  toggleSceneList (state, payload) {
    state.items = state.items.map(obj => {
      if (obj.scene_id === payload.scene_id) {
        if (payload.list === 'watchlist') {
          obj.watchlist = !obj.watchlist
        }
        if (payload.list === 'favourite') {
          obj.favourite = !obj.favourite
        }
        if (payload.list == 'watched') {
          obj.is_watched = !obj.is_watched
        }
        if (payload.list === 'trailerlist') {
          obj.trailerlist = !obj.trailerlist
        }
        if (payload.list === 'needs_update') {
          obj.needs_update = !obj.needs_update
        }
        if (payload.list === 'is_hidden') {
          obj.is_hidden = !obj.is_hidden
        }        
        if (payload.list === 'wishlist') {
          obj.wishlist = !obj.wishlist
        }
      }
      return obj
    })

    api.post('scene/toggle', {
      json: {
        scene_id: payload.scene_id,
        list: payload.list
      }
    })
  },
  updateScene (state, payload) {
    state.items = state.items.map(obj => {
      if (obj.scene_id === payload.scene_id) {
        obj = payload
      }
      return obj
    })
  },
  stateFromJSON (state, payload) {
    try {
      const obj = JSON.parse(payload)
      for (const [k, v] of Object.entries(obj)) {
        Vue.set(state.filters, k, v)
      }
    } catch (err) {
    }
  },
  setFilterValue (state, { key, value }) {
    Vue.set(state.filters, key, value)
  },
  // apply the availability/hidden filter preset for a download-state choice
  applyDlState (state, value) {
    state.filters.dlState = value
    switch (value) {
      case 'any':
        state.filters.isAvailable = null
        state.filters.isAccessible = null
        state.filters.isHidden = false
        break
      case 'available':
        state.filters.isAvailable = true
        state.filters.isAccessible = true
        state.filters.isHidden = false
        break
      case 'downloaded':
        state.filters.isAvailable = true
        state.filters.isAccessible = null
        state.filters.isHidden = false
        break
      case 'missing':
        state.filters.isAvailable = false
        state.filters.isAccessible = null
        state.filters.isHidden = false
        break
      case 'hidden':
        state.filters.isAvailable = null
        state.filters.isAccessible = null
        state.filters.isHidden = true
        break
    }
  },
  setCastFilterOnly (state, actor) {
    state.filters.cast = actor
    state.filters.sites = []
    state.filters.tags = []
    state.filters.attributes = []
  },
  clearShowSceneId (state) {
    state.show_scene_id = ''
  },
  stateFromQuery (state, payload) {
    try {
      state.show_scene_id=payload.scene_id
      const obj = decodeJsonBase64(payload.q)
      for (const [k, v] of Object.entries(obj)) {
        Vue.set(state.filters, k, v)
      }
    } catch (err) {
    }
  }
}

const actions = {
  async filters ({ state }) {
    // long timeouts: these endpoints used 5-min budgets before the shared
    // API layer introduced a 60s default
    state.playlists = await api.get('playlist', { timeout: 300000 }).json()
    state.filterOpts = await api.get('scene/filters', { timeout: 300000 }).json()

    // Reverse list of release months for display purposes
    state.filterOpts.release_month = state.filterOpts.release_month.reverse()
  },
  async load ({ state, getters, commit }, params) {
    const iOffset = params.offset || 0

    state.isLoading = true

    const q = Object.assign({}, state.filters)
    q.offset = iOffset
    q.limit = state.limit

    const data = await api
      .post('scene/list', {
        json: q,
        // heavy on large libraries / slow storage; keep the historical
        // generous budget instead of the 60s shared default
        timeout: 6e6
      })
      .json()

    state.isLoading = false

    if (iOffset === 0) {
      commit('setItems', [])
    }

    commit('setItems', state.items.concat(data.scenes))
    state.offset = iOffset + state.limit
    state.total = data.results

    state.counts.any = data.count_any
    state.counts.available = data.count_available
    state.counts.downloaded = data.count_downloaded
    state.counts.not_downloaded = data.count_not_downloaded
    state.counts.hidden = data.count_hidden
  }
}

export default {
  namespaced: true,
  state,
  getters,
  mutations,
  actions
}
