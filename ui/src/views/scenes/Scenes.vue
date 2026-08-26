<template>
  <div class="container is-fluid">
    <div class="columns">

      <div class="column is-one-fifth">
        <Filters/>

        <div id="scrollButtons">
          <a id="toTop" aria-label="Back to top" role="button">
            <b-icon pack="mdi" icon="navigation" />
          </a>
          <a id="toggleInfiniteScroll" role="button" :aria-label="infiniteScrollEnabled ? 'Disable Auto Load More' : 'Enable Auto Load More'" @click="toggleInfiniteScroll" :title="infiniteScrollEnabled ? 'Disable Auto Load More' : 'Enable Auto Load More'">
            <b-icon pack="mdi" :icon="infiniteScrollEnabled ? 'reload' : 'pause'" />
          </a>
        </div>
      </div>

      <div class="column is-four-fifths">
        <List :infinite-scroll-enabled="infiniteScrollEnabled"/>
      </div>

    </div>
  </div>
</template>

<script>
import Filters from './Filters'
import List from './List'

export default {
  name: 'Scenes',
  components: { Filters, List },
  data() {
    return {
      infiniteScrollEnabled: true
    }
  },
  methods: {
    toggleInfiniteScroll() {
      this.infiniteScrollEnabled = !this.infiniteScrollEnabled
    }
  },
  created () {
    // named handlers so the listeners can be removed on destroy —
    // previously anonymous listeners leaked on every route mount
    this._onScroll = () => {
      const toTop = document.getElementById('toTop')
      const toggleBtn = document.getElementById('toggleInfiniteScroll')
      if (!toTop || !toggleBtn) {
        return
      }
      const show = document.body.scrollTop > 20 || document.documentElement.scrollTop > 20
      toTop.style.display = show ? 'flex' : 'none'
      toggleBtn.style.display = show ? 'flex' : 'none'
    }
    this._scrollToTop = () => {
      const c = document.documentElement.scrollTop || document.body.scrollTop
      if (c > 0) {
        window.requestAnimationFrame(this._scrollToTop)
        window.scrollTo(0, c - c / 16)
      }
    }
    this._onToTopClick = () => {
      this._scrollToTop()
    }
  },
  mounted () {
    window.addEventListener('scroll', this._onScroll)
    document.getElementById('toTop').addEventListener('click', this._onToTopClick)
  },
  beforeDestroy () {
    window.removeEventListener('scroll', this._onScroll)
    const toTop = document.getElementById('toTop')
    if (toTop) {
      toTop.removeEventListener('click', this._onToTopClick)
    }
  },
  beforeRouteEnter (to, from, next) {
    next(vm => {
      if (to.query !== undefined) {
        vm.$store.commit('sceneList/stateFromQuery', to.query)
      }
      vm.$store.dispatch('optionsWeb/load')
      vm.$store.dispatch('sceneList/load', { offset: 0 })
      vm.$store.dispatch('optionsAdvanced/load')
      vm.$store.dispatch('optionsStorage/load')
    })
  },
  beforeRouteUpdate (to, from, next) {
    if (to.query !== undefined) {
      this.$store.commit('sceneList/stateFromQuery', to.query)
    }
    this.$store.dispatch('sceneList/load', { offset: 0 })
    next()
  },
}
</script>

<style scoped>
  #scrollButtons {
    display: flex;
    justify-content: space-between;
    position: fixed;
    bottom: 20px;
    left: 30px;
    width: 18.5%;
  }
  #toTop, #toggleInfiniteScroll {
    display: none;
    align-items: center;
    justify-content: center;
    width: 44px;
    height: 44px;
    padding: 0;
    background-color: var(--xbvr-surface, #ffffff);
    color: var(--xbvr-text-muted, #64708a);
    border: 1px solid var(--xbvr-border, #e3e6ec);
    border-radius: 999px;
    font-size: 18px;
    margin-right: 8px;
    box-shadow: var(--xbvr-shadow-md, 0 4px 12px rgba(16, 24, 40, 0.10));
    transition: box-shadow var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
      transform var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
      color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
  }
  #toTop:hover, #toggleInfiniteScroll:hover {
    background-color: var(--xbvr-surface, #ffffff);
    color: var(--xbvr-primary, #4f46e5);
    transform: translateY(-2px);
    box-shadow: var(--xbvr-shadow-lg, 0 16px 40px rgba(16, 24, 40, 0.16));
  }
</style>
