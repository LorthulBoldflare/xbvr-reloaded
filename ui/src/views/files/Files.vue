<template>
  <div class="files-page">
    <div class="container is-fluid">
      <header class="page-header">
        <h1 class="page-title">{{$t('Files')}}</h1>
        <span class="page-count">{{$store.state.files.items.length}} {{$t('files')}}</span>
      </header>
      <Filters/>
      <List/>
    </div>
    <Player v-if="showPlayerOverlay"/>
    <SceneMatch v-if="showMatchOverlay"/>
    <CreateScene v-if="showCreateOverlay"/>
  </div>
</template>

<script>
import Filters from './Filters'
import List from './List'
import Player from './Player'
import SceneMatch from './SceneMatch'
import CreateScene from './CreateScene'

export default {
  name: 'Files',
  components: { Filters, List, Player, SceneMatch, CreateScene },
  computed: {
    showPlayerOverlay () {
      return this.$store.state.overlay.player.show
    },
    showMatchOverlay () {
      return this.$store.state.overlay.match.show
    },
    showCreateOverlay () {
      return this.$store.state.overlay.createScene.show
    }
  }
}
</script>

<style scoped>
.files-page {
  padding-bottom: 2rem;
}

.page-header {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-bottom: 1.25rem;
}

.page-title {
  font-size: 1.6rem;
  font-weight: 700;
  line-height: 1.2;
  color: var(--xbvr-text, #1c2333);
}

.page-count {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--xbvr-text-muted, #64708a);
  white-space: nowrap;
}
</style>
