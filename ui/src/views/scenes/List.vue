<template>
  <div ref="scrollContainer">
    <b-loading :is-full-page="true" :active.sync="isLoading"></b-loading>

    <div class="list-toolbar">
      <strong class="results-count">{{total}} results</strong>
      <div class="toolbar-actions">
        <b-button size="is-small" icon-pack="mdi" icon-left="import"
                  :disabled="!webhookConfigured('trigger_external_import')"
                  :title="webhookConfigured('trigger_external_import') ? $t('Trigger External Import') : $t('Configure in Options → Storage → Webhooks')"
                  @click="triggerWebhook('trigger-import')">
          {{$t("Import")}}
        </b-button>
        <b-button size="is-small" icon-pack="mdi" icon-left="refresh"
                  :disabled="!webhookConfigured('refresh_external_import')"
                  :title="webhookConfigured('refresh_external_import') ? $t('Refresh External Import') : $t('Configure in Options → Storage → Webhooks')"
                  @click="triggerWebhook('refresh-import')">
          {{$t("Refresh")}}
        </b-button>
      </div>
    </div>

    <div class="list-controls">
      <b-field class="control-field availability-field">
        <b-radio-button v-model="dlState" native-value="any" size="is-small">
          {{$t("Any")}} ({{counts.any}})
        </b-radio-button>
        <b-radio-button v-model="dlState" native-value="available" size="is-small">
          {{$t("Available right now")}} ({{counts.available}})
        </b-radio-button>
        <b-radio-button v-model="dlState" native-value="downloaded" size="is-small">
          {{$t("Downloaded")}} ({{counts.downloaded}})
        </b-radio-button>
        <b-radio-button v-model="dlState" native-value="missing" size="is-small">
          {{$t("Not downloaded")}} ({{counts.not_downloaded}})
        </b-radio-button>
        <b-radio-button v-model="dlState" native-value="hidden" size="is-small">
          {{$t("Hidden")}} ({{counts.hidden}})
        </b-radio-button>
      </b-field>
      <b-field class="control-field card-size-field">
        <span class="list-header-label">{{$t('Card size')}}</span>
        <b-radio-button v-model="cardSize" native-value="1" size="is-small">
          XS
        </b-radio-button>
        <b-radio-button v-model="cardSize" native-value="2" size="is-small">
          S
        </b-radio-button>
        <b-radio-button v-model="cardSize" native-value="3" size="is-small">
          M
        </b-radio-button>
        <b-radio-button v-model="cardSize" native-value="4" size="is-small">
          L
        </b-radio-button>
      </b-field>
    </div>

    <div class="columns is-multiline">
      <div :class="['column', 'is-multiline', cardSizeClass]"
           v-for="item in items" :key="item.id">
        <SceneCard :item="item"/>
      </div>
    </div>

    <div class="column is-full" v-if="isLoadingMore">
      <b-loading :is-full-page="false" :active="true"></b-loading>
    </div>
    <div class="column is-full" v-if="!infiniteScrollEnabled && items.length < total">
      <b-button type="is-primary" @click="loadMore" :loading="isLoadingMore" expanded>{{$t('Load More')}}</b-button>
    </div>
  </div>
</template>

<script>
import SceneCard from './SceneCard'
import api from '../../api'

export default {
  name: 'List',
  components: { SceneCard },
  props: {
    infiniteScrollEnabled: {
      type: Boolean,
      default: true
    }
  },
  data() {
    return {
      isLoadingMore: false,
      scrollHandler: null,
      debounceTimeout: null
    }
  },
  computed: {
    cardSize: {
      get () {
        return this.$store.state.sceneList.filters.cardSize
      },
      set (value) {
        this.$store.commit('sceneList/setFilterValue', { key: 'cardSize', value })
      }
    },
    cardSizeClass () {
      switch (this.$store.state.sceneList.filters.cardSize) {
        case '1':
          return 'is-2'
        case '2':
          return 'is-one-fifth'
        case '3':
          return 'is-one-quarter'
        case '4':
          return 'is-one-third'
        default:
          return 'is-one-fifth'
      }
    },
    dlState: {
      get () {
        return this.$store.state.sceneList.filters.dlState
      },
      set (value) {
        this.$store.commit('sceneList/applyDlState', value)
        this.reloadList()
      }
    },
    isLoading () {
      return this.$store.state.sceneList.isLoading
    },
    items () {
      return this.$store.state.sceneList.items
    },
    total () {
      return this.$store.state.sceneList.total
    },
    counts () {
      return this.$store.state.sceneList.counts
    },
  },
  methods: {
    webhookConfigured (key) {
      const webhooks = this.$store.state.optionsStorage.options.webhooks
      return !!(webhooks && webhooks[key] && webhooks[key].url)
    },
    async triggerWebhook (name) {
      await api.get('/task/webhook/' + name)
      this.$buefy.toast.open({
        message: this.$t('Webhook triggered'),
        type: 'is-success'
      })
    },
    reloadList () {
      this.$router.push({
        name: 'scenes',
        query: {
          q: this.$store.getters['sceneList/filterQueryParams']
        }
      })
    },
    async loadMore () {
      if (this.isLoadingMore || this.items.length >= this.total) return
      this.isLoadingMore = true
      await this.$store.dispatch('sceneList/load', { offset: this.$store.state.sceneList.offset })
      this.isLoadingMore = false
    },
    handleScroll () {
      if (this.debounceTimeout) clearTimeout(this.debounceTimeout)
      this.debounceTimeout = setTimeout(() => {
        const scrollY = window.scrollY || window.pageYOffset
        const viewportHeight = window.innerHeight
        const fullHeight = document.documentElement.scrollHeight
        // If user is within 600px of the bottom, load more
        if (scrollY + viewportHeight + 600 >= fullHeight) {
          this.loadMore()
        }
      }, 100)
    }
  },
  mounted () {
    this.scrollHandler = this.handleScroll.bind(this)
    if (this.infiniteScrollEnabled) {
      window.addEventListener('scroll', this.scrollHandler)
    }
  },
  beforeDestroy () {
    if (this.infiniteScrollEnabled && this.scrollHandler) {
      window.removeEventListener('scroll', this.scrollHandler)
    }
    if (this.debounceTimeout) clearTimeout(this.debounceTimeout)
  },
  watch: {
    // open the details overlay when a scene id is pushed to the store —
    // replaces the hidden-span + side-effect computed hack
    '$store.state.sceneList.show_scene_id'(newVal) {
      if (newVal != undefined && newVal != '') {
        api.get('/scene/' + newVal).json().then(data => {
          if (data.id != 0) {
            this.$store.commit('overlay/showDetails', { scene: data })
          }
        })
        this.$store.commit('sceneList/clearShowSceneId')
      }
    },
    infiniteScrollEnabled(newVal) {
      if (newVal) {
        window.addEventListener('scroll', this.scrollHandler)
      } else {
        window.removeEventListener('scroll', this.scrollHandler)
      }
    }
  }
}
</script>

<style scoped>
  /* results count + webhook actions */
  .list-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-bottom: 0.75rem;
  }

  .results-count {
    font-size: 1.05rem;
    font-weight: 700;
    color: var(--xbvr-text, #1c2333);
    white-space: nowrap;
  }

  .toolbar-actions {
    display: flex;
    align-items: center;
    gap: 0.35rem;
  }

  /* availability + card-size segmented controls */
  .list-controls {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 0.75rem;
    margin-bottom: 1rem;
  }

  .control-field {
    margin-bottom: 0;
  }

  .card-size-field {
    align-items: center;
  }

  .list-header-label {
    padding-right: 0.75em;
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--xbvr-text-muted, #64708a);
    white-space: nowrap;
  }
</style>
