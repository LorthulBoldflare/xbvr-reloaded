import api from '../api'
import { encodeJsonBase64, decodeJsonBase64 } from '../util/base64'
import Vue from 'vue'

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
  cast: [],
  sites: [],
  tags: [],
  attributes: [],
  jumpTo: '',
  min_age: 0,
  max_age: 100,  
  min_height: 120,
  max_height: 220,  
  min_weight: 25,
  max_weight: 150,  
  min_count: 0,
  max_count: 150,  
  min_avail: 0,
  max_avail: 150,  
  min_rating: 0,
  max_rating: 5,
  min_scene_rating: 0,
  max_scene_rating: 5,
  sort: 'name_asc'
}

const state = {
  actors: [],
  playlists: [],
  isLoading: false,
  offset: 0,
  total: 0,
  limit: 18,
  show_actor_id: '',
  filterOpts: {
    cast: [],
    sites: [],
    tags: []
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
  prevActor: (state) => (currentActor) => {
    const i = state.actors.findIndex(actor => actor.id === currentActor.id)
    if (i === 0) {
      return null
    }
    return state.actors[i - 1]
  },
  nextActor: (state) => (currentActor) => {
    const i = state.actors.findIndex(actor => actor.id === currentActor.id)
    if (i === state.actors.length - 1) {
      return null
    }
    return state.actors[i + 1]
  },
  firstActor: (state) => () => {    
    return state.actors[0]
  },
  lastActor: (state) => () => {    
    return state.actors[state.actors.length-1]
  }
}

const mutations = {
  setActors (state, payload) {
    state.actors = payload
  },
  toggleActorList (state, payload) {
    state.actors = state.actors.map(obj => {
      if (obj.actor_id === payload.actor_id) {
        if (payload.list === 'watchlist') {
          obj.watchlist = !obj.watchlist
        }
        if (payload.list === 'favourite') {
          obj.favourite = !obj.favourite
        }
        if (payload.list === 'needs_update') {
          obj.needs_update = !obj.needs_update
        }
      }
      return obj
    })

    api.post('actor/toggle', {
      json: {
        actor_id: payload.actor_id,
        list: payload.list
      }
    })
  },
  updateActor (state, payload) {
    state.actors = state.actors.map(obj => {
      if (obj.id === payload.id) {
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
  stateFromQuery (state, payload) {
    try {
      state.show_actor_id=payload.actor_id
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
    state.playlists = await api.get('playlist/actor').json()
    state.filterOpts = await api.get('actor/filters').json()    
  },
  async load ({ state, getters, commit }, params) {    
    const iOffset = params.offset || 0

    state.isLoading = true

    const q = Object.assign({}, state.filters)
    q.offset = iOffset
    q.limit = state.limit
    
    const data = await api
      .post('actor/list', { json: q })
      .json()

    state.isLoading = false
    state.filters.jumpTo = ""  // clear the jumpto value now that we have the data

    if (iOffset === 0) {
      commit('setActors', [])
    }

    commit('setActors', state.actors=data.actors)
    state.offset = data.offset + state.limit
    state.total = data.results    
  }
}

export default {
  namespaced: true,
  state,
  getters,
  mutations,
  actions
}
