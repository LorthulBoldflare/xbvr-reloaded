import Vue from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import i18n from './i18n'
import { initTheme } from './util/theme'

import vueDebounce from 'vue-debounce'

import Buefy from 'buefy'
import 'buefy/dist/buefy.css'
// XBVR design system — loaded last so it overrides Bulma/Buefy defaults
import './assets/theme.css'

import 'video.js/dist/video-js.css'
import 'videojs-vr/dist/videojs-vr.css'
// FontAwesome CSS + webfonts instead of the ~1MB js/all SVG-inliner
import '@fortawesome/fontawesome-free/css/all.css'
import '@mdi/font/css/materialdesignicons.css'

// apply the saved/system theme before first paint to avoid a flash
initTheme()

Vue.config.productionTip = false
Vue.config.keyCodes = {
  arrowLeft: 37,
  arrowRight: 39,
  questionMark: 63
}
Vue.use(Buefy)
Vue.use(vueDebounce)

new Vue({
  router,
  store,
  i18n,
  render: h => h(App)
}).$mount('#app')
