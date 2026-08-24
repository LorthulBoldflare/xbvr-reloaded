import api from '../api'

const state = {
  countTotal: 0,
  countUpdated: 0,  
  optionsFunscripts: {
    scrapeFunscripts: false,
  }
}

const mutations = {}

const actions = {
  async load({ state }, params) {
    api.get('/options/funscripts/count')
      .json()
      .then(data => {
        state.countTotal = data.total
        state.countUpdated = data.updated
      })
    api.get('/options/state')
      .json()
      .then(data => {
        state.optionsFunscripts.scrapeFunscripts = data.config.funscripts.scrapeFunscripts
      })

  },
  async save ({ state }) {    
    api.put('/options/funscripts', { json: { ...state.optionsFunscripts } })
      .json()
      .then(data => {
        state.optionsFunscripts.scrapeFunscripts = data.scrapeFunscripts
      })
  },
}

export default {
  namespaced: true,
  state,
  mutations,
  actions
}
