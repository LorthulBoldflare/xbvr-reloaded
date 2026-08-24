import Vue from 'vue'
import Router from 'vue-router'

// route-level code splitting: each view loads in its own chunk
const Options = () => import('./views/options/Options')
const Scenes = () => import('./views/scenes/Scenes')
const Actors = () => import('./views/actors/Actors')
const Files = () => import('./views/files/Files')

Vue.use(Router)

export default new Router({
  mode: 'hash',
  base: process.env.BASE_URL,
  routes: [
    {
      path: '/',
      name: 'scenes',
      component: Scenes
    },
    {
      path: '/actors',
      name: 'actors',
      component: Actors
    },
    {
      path: '/files',
      name: 'files',
      component: Files
    },
    {
      path: '/options',
      name: 'options',
      component: Options
    }
  ]
})
