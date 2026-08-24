import api from '../api'

const state = {
  tpdb: {
    apiToken: '',
  },
  scrapers: [],
}

const mutations = {}

const actions = {
  async load({ state }, params) {
    api.get('/options/state')
      .json()
      .then(data => {
        state.tpdb.apiToken = data.config.vendor.tpdb.apiToken
        state.scrapers = data.scrapers        
      })
  },
}

export default {
  namespaced: true,
  state,
  mutations,
  actions
}
