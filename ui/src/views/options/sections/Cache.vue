<template>
  <div>
    <b-loading :is-full-page="false" :active.sync="isLoading"></b-loading>
    <div class="content">
      <header class="options-page-head">
        <h1 class="options-title">{{$t("Cache")}}</h1>
        <p class="options-desc">{{ $t('Inspect and reset the local caches XBVR keeps for speed.') }}</p>
      </header>
      <div class="settings-card" v-if="!isLoading">
        <div class="settings-card-title">
          <b-icon pack="mdi" icon="database-outline" size="is-small"/>
          {{ $t('Local caches') }}
        </div>
        <table class="table is-fullwidth cache-table">
          <tbody>
            <tr>
              <td>
                <p class="cache-name">Images</p>
                <p class="cache-desc">
                  Cache of remote images that were requested at least once.
                </p>
              </td>
              <td nowrap class="cache-size">{{prettyBytes(sizes.images)}}</td>
              <td class="cache-actions">
                <b-button size="is-small" @click="resetCache('images')">Reset</b-button>
              </td>
            </tr>
            <tr>
              <td>
                <p class="cache-name">Video previews</p>
                <p class="cache-desc">
                  Generated on demand for local files. Remove when you want to generate previews using new settings.
                </p>
              </td>
              <td nowrap class="cache-size">{{prettyBytes(sizes.previews)}}</td>
              <td class="cache-actions">
                <b-button size="is-small" @click="resetCache('previews')">Reset</b-button>
              </td>
            </tr>
            <tr>
              <td>
                <p class="cache-name">Search index
                  <span class="cache-status" :class="{ 'is-active': searchInprogress }">
                    {{ searchInprogress ? 'Indexing in progress' : `${indexSceneCount} scenes indexed` }}
                  </span>
                </p>
                <p class="cache-desc">
                  Remove search index when facing issues with finding/matching files.
                </p>
              </td>
              <td nowrap class="cache-size">{{prettyBytes(sizes.searchIndex)}}</td>
              <td class="cache-actions">
                <b-field>
                  <b-button size="is-small" @click="resetCache('searchIndex')">Reset</b-button>
                  <b-button size="is-small" @click="indexRescan">Rescan</b-button>
                </b-field>
              </td>
            </tr>
            <tr>
              <td>
                <p class="cache-name">Scene status</p>
                <p class="cache-desc">
                  Refresh scene status when scenes are not marked "available" or "scripted" despite having such files assigned.
                </p>
              </td>
              <td nowrap class="cache-size"></td>
              <td class="cache-actions">
                <b-button size="is-small" @click="taskRefresh">Refresh Scenes</b-button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script>
import api from '../../../api'
import prettyBytes from 'pretty-bytes'

export default {
  name: 'Cache',
  data () {
    return {
      isLoading: true,
      sizes: {},
      indexSceneCount: 0,
      searchInprogress: false,
    }
  },
  async mounted () {
    await this.loadState()
    this.loadSearchState()
  },
  methods: {
    async loadState () {
      this.isLoading = true
      await api.get('/options/state')
        .json()
        .then(data => {
          this.sizes = data.currentState.cacheSize
          this.isLoading = false
        })
    },
    resetCache (kind) {
      this.$buefy.dialog.confirm({
        title: this.$t('Reset cache'),
        message: `Do you want to reset the <strong>${kind}</strong> cache?`,
        type: 'is-danger',
        hasIcon: true,
        confirmText: this.$t('Reset'),
        onConfirm: async () => {
          this.isLoading = true
          await api.delete(`/options/cache/reset/${kind}`, { timeout: 30000 })
          await this.loadState()
          await this.loadSearchState()
        }
      })
    },
    taskRefresh: function () {
      api.get('/task/scene-refresh')
    },
    async loadSearchState () {
      this.isLoading = true
      await api.get('/options/state/search')
        .json()
        .then(data => {
          this.indexSceneCount = data.documentCount
          this.searchInprogress = data.inProgress
          this.isLoading = false
        })
    },
    async indexRescan () {
      this.isLoading = true
      await api.get('/task/index')
      this.searchInprogress = true
      this.isLoading = false
    },
    prettyBytes
  }
}
</script>

<style scoped>
.options-page-head {
  margin-bottom: 1.25rem;
}

.options-title {
  font-size: 1.4rem;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--xbvr-text, #1c2333);
  margin-bottom: 0.15rem;
}

.options-desc {
  color: var(--xbvr-text-muted, #64708a);
  font-size: 0.9rem;
  margin: 0;
}

.settings-card {
  background: var(--xbvr-surface, #ffffff);
  border: 1px solid var(--xbvr-border, #e3e6ec);
  border-radius: var(--xbvr-radius, 12px);
  box-shadow: var(--xbvr-shadow-sm, 0 1px 2px rgba(16, 24, 40, 0.05));
  padding: 1rem 1.25rem 0.25rem;
  margin-bottom: 1.25rem;
}

.settings-card-title {
  display: flex;
  gap: 0.45rem;
  align-items: center;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--xbvr-text-muted, #64708a);
  margin-bottom: 0.35rem;
}

.settings-card-title .icon {
  color: var(--xbvr-text-faint, #7d88a1);
}

.cache-table {
  background: transparent;
}

.cache-table td {
  vertical-align: middle;
  padding: 0.85rem 0.75rem;
}

.cache-table tr:last-child td {
  border-bottom: none;
}

.cache-table tr {
  transition: background-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

.cache-table tbody tr:hover {
  background: var(--xbvr-hover-bg, #fafbfd);
}

.cache-name {
  font-weight: 700;
  color: var(--xbvr-text, #1c2333);
  margin-bottom: 0.15rem;
}

.cache-status {
  display: inline-flex;
  align-items: center;
  font-size: 0.72rem;
  font-weight: 600;
  color: var(--xbvr-text-muted, #64708a);
  background: var(--xbvr-surface-sunken, #eef0f4);
  border: 1px solid var(--xbvr-border, #e3e6ec);
  border-radius: 999px;
  padding: 0.1em 0.6em;
  margin-left: 0.5rem;
  vertical-align: middle;
}

.cache-status.is-active {
  background: var(--xbvr-primary-soft, #eef0fe);
  color: var(--xbvr-primary-strong, #4338ca);
  border-color: transparent;
}

.cache-desc {
  color: var(--xbvr-text-muted, #64708a);
  font-size: 0.85rem;
  margin: 0;
}

.cache-size {
  font-variant-numeric: tabular-nums;
  color: var(--xbvr-text-muted, #64708a);
  font-weight: 600;
}

.cache-actions {
  text-align: right;
  white-space: nowrap;
}

.cache-actions .field {
  justify-content: flex-end;
}
</style>
