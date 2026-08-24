const state = {
  isPreviewReady: false,
  generatingPreview: false,
  previewFn: '',
  previewTs: 0,
  queue: {
    running: false,
    stopping: false,
    total: 0,
    completed: 0,
    remaining: 0,
    currentScene: ''
  }
}

const mutations = {
  hidePreview (state) {
    state.isPreviewReady = false
    state.generatingPreview = true
    state.previewFn = ''
  },
  showPreview (state, payload) {
    state.isPreviewReady = true
    state.generatingPreview = false
    state.previewFn = payload.previewFn
    state.previewTs = Date.now()
  },
  setQueue (state, payload) {
    state.queue = {
      running: !!payload.running,
      stopping: !!payload.stopping,
      total: payload.total || 0,
      completed: payload.completed || 0,
      remaining: payload.remaining || 0,
      currentScene: payload.currentScene || ''
    }
  }
}

export default {
  namespaced: true,
  state,
  mutations
}
