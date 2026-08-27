<template>
  <b-modal :active.sync="isActive"
           :destroy-on-hide="false"
           has-modal-card
           trap-focus
           aria-role="dialog"
           aria-modal
           can-cancel>
    <div class="modal-card quickfind-card">
      <b-field class="quickfind-field">
        <b-autocomplete
          ref="autocompleteInput"
          :data="data"
          placeholder="Find scene..."
          field="query"
          :loading="isFetching"
          v-model="queryString"
          @typing="getAsyncData"
          @select="option => showSceneDetails(option)"
          :open-on-focus="true"
          custom-class="is-large"
          max-height="450">

          <template slot-scope="props">
            <div class="media quickfind-result">
              <div class="media-left">
                <vue-load-image>
                  <img slot="image" :src="getImageURL(props.option)" width="80" class="quickfind-thumb"/>
                  <img slot="preloader" src="/ui/images/blank.png" width="80" class="quickfind-thumb"/>
                  <img slot="error" src="/ui/images/blank.png" width="80" class="quickfind-thumb"/>
                </vue-load-image>
              </div>
              <div class="media-content">
                <span class="quickfind-site">
                  {{ props.option.site}}
                  <b-icon v-if="props.option.is_hidden" pack="mdi" icon="eye-off-outline" size="is-small"/>
                </span>
                <div class="truncate"><strong>{{ props.option.title }}</strong></div>
                <div class="quickfind-cast">
                  <small>
                    <span v-for="(c, idx) in props.option.cast" :key="'cast' + idx">
                      {{c.name}}<span v-if="idx < props.option.cast.length-1">, </span>
                    </span>
                  </small>
                </div>
                <star-rating v-if="props.option.star_rating != 0" :read-only="true" :rating="props.option.star_rating" :increment="0.5" :show-rating="false" :star-size="10"/>
              </div>
              <div class="media-right quickfind-date">
                {{format(parseISO(props.option.release_date), "yyyy-MM-dd")}}
              </div>
            </div>
          </template>
        </b-autocomplete>
      </b-field>

      <div class="quickfind-helpers">
        <span class="quickfind-helpers-label">{{$t('Search Fields')}}</span>
        <b-tooltip :label="$t('Optional: select one or more words to target searching to a specific field')" :delay="500" position="is-top">
          <b-button @click='searchPrefix("+title:")' class="hint-chip">title:</b-button>
          <b-button @click='searchPrefix("cast:")' class="hint-chip">cast:</b-button>
          <b-button @click='searchPrefix("+site:")' class="hint-chip">site:</b-button>
          <b-button @click='searchPrefix("+id:")' class="hint-chip">id:</b-button>
        </b-tooltip>
        <b-tooltip :label="$t('Add file duration to search')" :delay="500" position="is-top">
          <b-button @click='searchDurationPrefix("duration:")' class="hint-chip">duration:</b-button>
        </b-tooltip>
        <b-tooltip :label="$t('Defaults date range to the last week. Note:must match yyyy-mm-dd, include leading zeros')" :delay="500" position="is-top">
          <b-button @click='searchDatePrefix("released:")' class="hint-chip">released:</b-button>
          <b-button @click='searchDatePrefix("added:")' class="hint-chip">added:</b-button>
        </b-tooltip>
      </div>

      <div class="quickfind-kbds">
        <span class="kbd-hint"><kbd class="qf-kbd">↑</kbd><kbd class="qf-kbd">↓</kbd> {{$t('to navigate')}}</span>
        <span class="kbd-hint"><kbd class="qf-kbd">↵</kbd> {{$t('to open')}}</span>
        <span class="kbd-hint"><kbd class="qf-kbd">esc</kbd> {{$t('to close')}}</span>
      </div>
    </div>
  </b-modal>
</template>

<script>
import api from './api'
import { getImageURL as getImageURLUtil } from './util/image'
import VueLoadImage from 'vue-load-image'
import GlobalEvents from 'vue-global-events'
import { format, parseISO } from 'date-fns'
import StarRating from 'vue-star-rating'

export default {
  name: 'QuickFind',
  props: {
    active: Boolean,
    sceneId: String
  },
  components: { VueLoadImage, GlobalEvents, StarRating },
  computed: {
    isActive: {
      get () {
        return this.$store.state.overlay.quickFind.show
      },
      set (values) {
        if (values) {
          this.$store.commit('overlay/showQuickFind')
        } else {
          this.$store.commit('overlay/hideQuickFind')
        }
      }
    }
  },
  watch: {
    // fetch suggestions when the query changes (was a computed side effect)
    queryString (newVal) {
      if (newVal != null && newVal != "") {
        this.getAsyncData(newVal)
      }
    },
    // focus the input and pick up any preset search string when opened
    '$store.state.overlay.quickFind.show' (shown) {
      if (shown === true) {
        this.$nextTick(() => {
          this.$refs.autocompleteInput.$refs.input.focus()
          if (this.$store.state.overlay.quickFind.searchString != null && this.$store.state.overlay.quickFind.searchString != "") {
            this.queryString = this.$store.state.overlay.quickFind.searchString
            this.$store.commit('overlay/clearQuickFindSearchString')
          }
        })
      }
    }
  },
  data () {
    return {
      data: [],
      dataNumRequests: 0,
      dataNumResponses: 0,
      selected: null,
      isFetching: false,
      queryString: ""
    }
  },
  methods: {
    getImageURL (opt) {
      return getImageURLUtil(opt.cover_url, '120x', opt.scene_id)
    },
    format,
    parseISO,
    getAsyncData: async function (query) {
      const requestIndex = this.dataNumRequests
      this.dataNumRequests = this.dataNumRequests + 1

      if (!query.length) {
        this.data = []
        this.dataNumResponses = requestIndex + 1
        this.isFetching = false
        return
      }

      this.isFetching = true

      const resp = await api.get('/scene/search', {
        searchParams: {
          q: query
        }
      }).json()

      if (requestIndex >= this.dataNumResponses) {
        this.dataNumResponses = requestIndex + 1
        if (this.dataNumResponses === this.dataNumRequests) {
          this.isFetching = false
        }

        if (resp.results > 0) {
          this.data = resp.scenes
        } else {
          this.data = []
        }
      }
    },
    showSceneDetails (scene) {
      this.$store.commit('overlay/hideQuickFind')
      if (this.$store.state.overlay.quickFind.displaySelectedScene) {
        if (this.$router.currentRoute.name !== 'scenes') {
            this.$router.push({ name: 'scenes' })
          }
          this.$store.commit('overlay/hideQuickFind')
          this.data = []
          this.$store.commit('overlay/showDetails', { scene })
        } else {
          // don't display the scene, just pass the selected scene back in the $store.state and close
          this.$store.commit('overlay/setQuickFindSelectedScene', scene)
          this.$store.commit('overlay/hideQuickFind')
          this.data = []
      }
    },
    searchPrefix(prefix) {      
      let textbox = this.$refs.autocompleteInput.$refs.input.$refs.input
      if (textbox.selectionStart != textbox.selectionEnd) {
        let selected = textbox.value.substring(textbox.selectionStart, textbox.selectionEnd)
        selected=selected.replace(/_/g," ").replace(/-/g," ").trim()
        if (selected.indexOf(' ') >= 0) {
          selected='"' + selected + '"'
        }
        this.queryString = textbox.value.substring(0,textbox.selectionStart) + " " + prefix + selected + " " + textbox.value.substr(textbox.selectionEnd)
        this.getAsyncData(this.queryString)
        this.$refs.autocompleteInput.focus()
      }
    },
    searchDatePrefix(prefix) {      
        let today = new Date().toISOString().slice(0, 10)
        let weekago = new Date(Date.now() - 604800000).toISOString().slice(0, 10)
        if (this.queryString == undefined) {
          this.queryString = prefix + '>="' + weekago + '" ' +  prefix + '<="' + today + '"'          
        } else {
          this.queryString = this.queryString.trim() + ' ' + prefix + '>="' + weekago + '" ' +  prefix + '<="' + today + '"'        
        }
        this.getAsyncData(this.queryString)
        this.$refs.autocompleteInput.focus()
    },
    searchDurationPrefix(prefix) {
      if (this.queryString == undefined) {
        this.queryString = prefix + '>=0'
      } else {
        this.queryString = this.queryString.trim() + ' ' + prefix + '>=0'
      }
      this.getAsyncData(this.queryString)
      this.$refs.autocompleteInput.focus()
    }
  }
}
</script>

<style scoped>
  /* command-palette placement: centered horizontally, near the top */
  .modal {
    justify-content: normal;
    padding-top: 10vh;
  }

  .quickfind-card {
    width: min(640px, 92vw);
    margin: 0 auto;
    padding: 1rem 1rem 0.8rem;
    display: flex;
    flex-direction: column;
    gap: 0.7rem;
  }

  .quickfind-field {
    margin-bottom: 0;
  }

  .quickfind-field :deep(.input.is-large) {
    border-radius: var(--xbvr-radius, 12px);
    font-weight: 500;
  }

  /* result rows inside the autocomplete dropdown */
  .quickfind-result {
    padding: 0.25rem 0.35rem;
    border-radius: var(--xbvr-radius-sm, 8px);
  }

  .quickfind-thumb {
    border-radius: var(--xbvr-radius-sm, 8px);
  }

  .quickfind-site {
    font-size: 0.72rem;
    font-weight: 700;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    color: var(--xbvr-text-faint, #7d88a1);
  }

  .quickfind-cast {
    margin-top: 0.35em;
    color: var(--xbvr-text-muted, #64708a);
  }

  .quickfind-date {
    font-size: 0.8rem;
    color: var(--xbvr-text-faint, #7d88a1);
    font-variant-numeric: tabular-nums;
  }

  /* field-prefix hint chips */
  .quickfind-helpers {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.35rem;
  }

  .quickfind-helpers-label {
    font-size: 0.7rem;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--xbvr-text-muted, #64708a);
    margin-right: 0.15rem;
  }

  .quickfind-helpers :deep(.b-tooltip) {
    display: inline-flex;
    gap: 0.3rem;
  }

  .hint-chip {
    height: auto;
    font-size: 0.72rem;
    font-weight: 600;
    line-height: 1;
    padding: 0.35em 0.6em;
    border-radius: 999px;
    border: 1px solid var(--xbvr-border, #e3e6ec);
    background: var(--xbvr-surface-sunken, #eef0f4);
    color: var(--xbvr-text-muted, #64708a);
    transition: color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
      background-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
      border-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
  }

  .hint-chip:hover {
    background: var(--xbvr-primary-soft, #eef0fe);
    border-color: var(--xbvr-primary, #4f46e5);
    color: var(--xbvr-primary-strong, #4338ca);
  }

  /* keyboard hints */
  .quickfind-kbds {
    display: flex;
    flex-wrap: wrap;
    gap: 0.9rem;
    padding-top: 0.55rem;
    border-top: 1px solid var(--xbvr-border, #e3e6ec);
  }

  .kbd-hint {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
    font-size: 0.72rem;
    color: var(--xbvr-text-faint, #7d88a1);
  }

  .qf-kbd {
    font-family: inherit;
    font-size: 0.68rem;
    font-weight: 600;
    line-height: 1;
    color: var(--xbvr-text-faint, #7d88a1);
    background: var(--xbvr-surface-sunken, #eef0f4);
    border: 1px solid var(--xbvr-border, #e3e6ec);
    border-radius: 5px;
    padding: 0.25em 0.45em;
  }

  .truncate {
    max-width: 320px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
</style>
