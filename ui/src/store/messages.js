const state = {
  lockScrape: false,
  lastScrapeMessage: '',
  lockRescan: false,
  lastRescanMessage: '',
  lastProgressMessage: '',
  runningScrapers: []
}

const mutations = {
  setLastScrapeMessage (state, msg) { state.lastScrapeMessage = msg },
  setLastRescanMessage (state, msg) { state.lastRescanMessage = msg },
  setRunningScrapers (state, list) { state.runningScrapers = list },
  addRunningScraper (state, id) { state.runningScrapers.push(id) },
  removeRunningScraper (state, id) {
    const i = state.runningScrapers.indexOf(id)
    if (i >= 0) {
      state.runningScrapers.splice(i, 1)
    }
  },
  setLockScrape (state, v) { state.lockScrape = v },
  setLockRescan (state, v) { state.lockRescan = v }
}

export default {
  namespaced: true,
  state,
  mutations
}
