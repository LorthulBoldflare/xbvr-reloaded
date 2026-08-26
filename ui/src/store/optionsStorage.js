import api from '../api'

const defaultWebhooks = () => ({
  trigger_external_import: { method: 'GET', url: '', headers: '' },
  refresh_external_import: { method: 'GET', url: '', headers: '' }
})

const state = {
  items: [],
  options: {
    match_ohash: false,
    forbidden_video_ext: [],
    video_ext: [],
    default_video_ext: [],
    webhooks: defaultWebhooks(),
  },
}

const mutations = {
  setOption (state, { key, value }) {
    state.options[key] = value
  }
}

const actions = {
  async load ({ state }, params) {
    await api.get('/options/storage').json()
    .then(data => {
      state.items = data.volumes
      state.options.match_ohash = data.match_ohash
      state.options.forbidden_video_ext = data.forbidden_video_ext
      state.options.video_ext = data.video_ext
      state.options.default_video_ext = data.default_video_ext
      state.options.webhooks = { ...defaultWebhooks(), ...(data.webhooks || {}) }
    })
  },
  async save ({ state }, enabled) {
    api.put('/options/storage', { json: { ...state.options } })
  },
}

export default {
  namespaced: true,
  state,
  mutations,
  actions
}
