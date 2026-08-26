<template>
  <div class="container is-fluid">
    <div class="columns">

      <div class="column is-one-fifth">
         <Filters/> 

        <a id="toTop" aria-label="Back to top" role="button">
          <b-icon pack="mdi" icon="navigation" />
        </a>
      </div>

      <div class="column is-four-fifths">
        <List/>
      </div>

    </div>
    
  </div>
</template>

<script>
import Filters from './Filters'
import List from './List'

export default {
  name: 'Actors',  
  components: { Filters, List},
  created () {
    // named handlers so the listeners can be removed on destroy
    this._onScroll = () => {
      const toTop = document.getElementById('toTop')
      if (!toTop) {
        return
      }
      toTop.style.display = document.body.scrollTop > 20 || document.documentElement.scrollTop > 20
        ? 'flex'
        : 'none'
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
        vm.$store.commit('actorList/stateFromQuery', to.query)
      }
      vm.$store.dispatch('optionsWeb/load')
      vm.$store.dispatch('actorList/load', { offset: 0 })
      vm.$store.dispatch('optionsAdvanced/load')
    })
  },
  beforeRouteUpdate (to, from, next) {
    if (to.query !== undefined) {
      vm.$store.commit('actorList/stateFromQuery', to.query)
    }
    this.$store.dispatch('actorList/load', { offset: 0 })
    next()
  },
  computed: {
  }
}
</script>

<style scoped>
  #toTop {
    display: none;
    align-items: center;
    justify-content: center;
    position: fixed;
    bottom: 20px;
    left: 30px;
    width: 44px;
    height: 44px;
    padding: 0;
    background-color: var(--xbvr-surface, #ffffff);
    color: var(--xbvr-text-muted, #64708a);
    border: 1px solid var(--xbvr-border, #e3e6ec);
    border-radius: 999px;
    font-size: 18px;
    box-shadow: var(--xbvr-shadow-md, 0 4px 12px rgba(16, 24, 40, 0.10));
    transition: box-shadow var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
      transform var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
      color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
  }

  #toTop:hover {
    background-color: var(--xbvr-surface, #ffffff);
    color: var(--xbvr-primary, #4f46e5);
    transform: translateY(-2px);
    box-shadow: var(--xbvr-shadow-lg, 0 16px 40px rgba(16, 24, 40, 0.16));
  }
</style>
