<template>
  <div>
    <GlobalEvents
      :filter="e => !['INPUT', 'TEXTAREA'].includes(e.target.tagName)"
      @keypress.prevent.questionMark="$store.commit('overlay/showQuickFind')"
    />
    <Navbar/>
    <div class="navbar-pad">
      <router-view/>
    </div>

    <Details v-if="showOverlay"/>
    <EditScene v-if="showEdit" />
    <ActorDetails v-if="showActorDetails"/>
    <EditActor v-if="showActorEdit" />
    <SearchStashdbScenes v-if="showSearchStashdbScenes" />
    <SearchStashdbActors v-if="showSearchStashdbActors" />

    <QuickFind/>
    <MigrationOverlay/>

    <Socket/>
  </div>
</template>

<script>
import GlobalEvents from 'vue-global-events'

import Navbar from './Navbar.vue'
import Socket from './Socket.vue'
import QuickFind from './QuickFind'
import MigrationOverlay from './components/MigrationOverlay'

// overlays load on demand — they are behind v-if and rarely all used
const Details = () => import('./views/scenes/Details')
const EditScene = () => import('./views/scenes/EditScene')
const ActorDetails = () => import('./views/actors/ActorDetails')
const EditActor = () => import('./views/actors/EditActor')
const SearchStashdbScenes = () => import('./views/scenes/SearchStashdbScenes')
const SearchStashdbActors = () => import('./views/actors/SearchStashdbActors')

export default {
  components: { Navbar, Socket, QuickFind, GlobalEvents, Details, EditScene, ActorDetails, EditActor, SearchStashdbScenes, SearchStashdbActors, MigrationOverlay },
  computed: {
    showOverlay () {
      return this.$store.state.overlay.details.show
    },
    showEdit () {
      return this.$store.state.overlay.edit.show
    },
    showActorDetails() {
      return this.$store.state.overlay.actordetails.show
    },
    showActorEdit() {
      return this.$store.state.overlay.actoredit.show
    },
    showSearchStashdbScenes() {
      return this.$store.state.overlay.searchStashDbScenes.show
    },
    showSearchStashdbActors() {
      return this.$store.state.overlay.searchStashDbActors.show
    },
  }
}
</script>

<style>
  .navbar-pad {
    padding-top: 1.5rem;
    padding-bottom: 3rem;
    min-height: calc(100vh - 3.75rem);
  }
  /* cap page width on very wide screens + consistent side padding */
  .container {
    max-width: 1800px;
    padding-left: 1.25rem;
    padding-right: 1.25rem;
  }
  .modal-background {
    background-color: rgba(16, 20, 32, .45) !important;
    backdrop-filter: blur(3px);
    -webkit-backdrop-filter: blur(3px);
  }
</style>
