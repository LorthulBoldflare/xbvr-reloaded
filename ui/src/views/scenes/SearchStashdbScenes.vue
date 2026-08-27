<template>
  <b-modal :active="isModalActive"           
           has-modal-card
           trap-focus
           aria-role="dialog"
           @cancel="close"
           aria-modal
           can-cancel>
    

    <div class="modal-card stashdb-card" :style="getOverlayPosition()">
      <header class="modal-card-head">
        <p class="modal-card-title">Search Stashdb Scenes</p>
        <button class="delete" @click="close" aria-label="close"></button>
      </header>

      <div class="modal-card-body">
        <div class="stashdb-body">
          <b-field label="Find scene..." class="stashdb-search">
            <b-input v-model="queryString" placeholder="Find scene..." @input="debouncedSearch" :loading="isFetching" custom-class="is-large"/>
          </b-field>
    
          <b-table :data="searchResults" class="stashdb-table">
            <b-table-column field="ImageUrl" >
              <template slot-scope="props">
                <div class="media result-card">
                  <div class="media-left">
                      <vue-load-image>
                          <img slot="image" :src="getImageURL(props.row.ImageUrl, props.row.Id)" width="150" class="result-thumb" @mouseover="showTooltipImage(props.row.ImageUrl)" @mouseout="showTooltipImage('')" />
                          <img slot="preloader" src="/ui/images/blank.png" height="150"/>
                          <img slot="error" src="/ui/images/blank.png" height="150"/>
                      </vue-load-image>
                      <div v-if="tooltipImage!='' && tooltipImage==props.row.ImageUrl" class="tooltipimg">
                        <img :src="tooltipImage" alt="Tooltip Image" width="400px" />
                      </div>
                    <div v-if="props.row.Date!=''" class="result-meta"><small><strong>Released:</strong> {{format(parseISO(props.row.Date), "yyyy-MM-dd")}}</small></div>
                    <div v-if="props.row.Duration!=''" class="result-meta"><small><strong>Durn:</strong> {{ props.row.Duration }}</small></div>
                    <div class="result-meta"><small><strong>Score:</strong> {{ props.row.Weight }}</small></div>
                    <div class="result-link">
                      <a class="button is-primary is-small" @click="linktoStashdb(props.row)" :title="'Link scene with stashdb'">
                        <b-icon pack="mdi" :icon="'link-variant-plus'" size="is-small"/>
                      </a>
                    </div>
                  </div>
                  <div class="media-content">
                    <div class="truncate"><strong><a :href="props.row.Url"  target="_blank">{{ props.row.Studio }} - {{ props.row.Title }}</a></strong></div>
                    <div class="result-desc"><small>{{props.row.Description}}</small></div>
                    <div class="result-cast">
                      <small>
                        <span v-for="(c, idx) in props.row.Performers" :key="'Performers' + idx">{{c.Name}}<span v-if="idx < props.row.Performers.length-1">, </span></span>
                      </small>
                    </div>
                  </div>            
                </div>
              </template>
            </b-table-column>
          </b-table>
        </div>
      </div>
      <footer class="modal-card-foot">
      </footer>
    </div>
  </b-modal>
</template>

<script>
import GlobalEvents from 'vue-global-events'
import api from '../../api'
import { getImageURL as getImageURLUtil } from '../../util/image'
import VueLoadImage from 'vue-load-image'
import { format, parseISO } from 'date-fns'

function debounce(func, wait) {
  let timeout;
  return function(...args) {
    const context = this;
    clearTimeout(timeout);
    timeout = setTimeout(() => func.apply(context, args), wait);
  };
}

export default {
  name: 'SearchStashdbScenes',
  components: {  GlobalEvents, VueLoadImage },
  data () {
    return {
        isModalActive: true,
        stashdbUrl: "",
        searchResults: [],
        queryString: "",
        isFetching: false,
        tooltipImage: '',
        scene: "",
        }        
  },
  created() {
    this.debouncedSearch = debounce(this.searchStashdb, 750); // 750ms delay
  },
  mounted () {
    const item = Object.assign({}, this.$store.state.overlay.searchStashDbScenes.scene)    
    this.scene = item
    this.openDialog(item)    
  },
  methods: {
    format,
    parseISO,
    close () {
      this.$store.commit('overlay/hideSearchStashdbScenes')
    },
    searchStashdb() {
        this.$buefy.toast.open({message: `Searching scenes`, type: 'is-primary', duration: 5000})
        api.get('/extref/stashdb/search/' + this.scene.id + "?q=" + this.queryString, {timeout: 6e6}).json().then(data => {
            this.searchResults = Object.values(data.Results).sort((a, b) => b.Weight - a.Weight)
            this.isModalActive = true
            if (data.Status!='') {
              this.$buefy.toast.open({message: `Warning:  ${data.Status}`, type: 'is-warning', duration: 5000})
            }
        })
    },
    selectScene(option) {
        this.stashdbUrl=option.Url.replace("https://stashdb.org/scenes/","")
        this.$nextTick(() => {
            if (this.$refs.autocompleteInput) {
                this.$refs.autocompleteInput.focus();
            }
        });
    },    
    linktoStashdb(option) {
        this.stashdbUrl=option.Url.replace("https://stashdb.org/scenes/","")
        api.get('/extref/stashdb/link2scene/' + this.scene.id +'/'+this.stashdbUrl ).json().then(data => {          
          this.$store.commit('sceneList/updateScene', data)
          this.$store.commit('overlay/showDetails', { scene: data })
          this.close()
        })
    },    
    getImageURL (u, stashId) {
      // results carry the raw stashdb UUID; scenes saved from stashdb use
      // scene_id 'stash-<uuid>' — match that format for the proxy context
      return getImageURLUtil(u, '120x', 'stash-' + stashId)
    },
    openDialog(scene) {
        this.isModalActive = true
        this.searchStashdb()
        this.$store.commit('overlay/changeDetailsTab', { tab: 3 })
        this.$nextTick(() => {
            if (this.$refs.autocompleteInput) {
                this.$refs.autocompleteInput.focus();
            }
        });
        this.scene = scene
    },
    getOverlayPosition(){
      if (this.$store.state.overlay.searchStashDbScenes.scene.synopsis == "") {
        return "height: 65vh; width: 40vw; left: 20vw; top: 20vh;"
      } else {
        return "height: 65vh; width: 40vw; left: -20vw"
      }
    },
    showTooltipImage(val){
      this.tooltipImage=val
    },
  },
  computed: {

}
}
</script>

<style scoped>
.b-modal {
  left: -20%;
  width: 40%;
  height: 65%;
  overflow: auto;
}

.stashdb-card {
  overflow-y: auto;
}

.stashdb-body {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.stashdb-search :deep(.label) {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--xbvr-text-muted, #64708a);
}

.stashdb-search :deep(.input.is-large) {
  border-radius: var(--xbvr-radius, 12px);
}

/* result cards */
.stashdb-table :deep(.table) {
  background: transparent;
}

.stashdb-table :deep(.table td) {
  border: none;
  padding: 0.35rem 0;
}

.result-card {
  padding: 0.75rem;
  background: var(--xbvr-surface, #ffffff);
  border: 1px solid var(--xbvr-border, #e3e6ec);
  border-radius: var(--xbvr-radius, 12px);
  box-shadow: var(--xbvr-shadow-sm, 0 1px 2px rgba(16, 24, 40, 0.05));
  transition: box-shadow var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
    border-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
    background-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

.result-card:hover {
  background: var(--xbvr-hover-bg, #fafbfd);
  border-color: var(--xbvr-border-strong, #cdd2dc);
  box-shadow: var(--xbvr-shadow-md, 0 4px 12px rgba(16, 24, 40, 0.10), 0 2px 4px rgba(16, 24, 40, 0.05));
}

.result-thumb {
  border-radius: var(--xbvr-radius-sm, 8px);
}

.result-meta {
  color: var(--xbvr-text-muted, #64708a);
}

.result-link {
  margin-top: 0.4rem;
}

.result-link .button {
  border-radius: 8px;
}

.result-desc {
  color: var(--xbvr-text-muted, #64708a);
}

.result-desc small,
.result-cast small {
  white-space: normal;
  display: block;
}

.result-cast {
  margin-top: 0.5em;
  color: var(--xbvr-text-muted, #64708a);
}

.truncate {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tab-item {
  height: 40vh;
}

.tooltipimg {
  position: absolute;
  z-index: 5;
  width: 350px;
  background-color: var(--xbvr-surface, #ffffff);
  border: 1px solid var(--xbvr-border, #e3e6ec);
  border-radius: var(--xbvr-radius, 12px);
  box-shadow: var(--xbvr-shadow-lg, 0 16px 40px rgba(16, 24, 40, 0.16), 0 4px 10px rgba(16, 24, 40, 0.08));
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 10px;
  transform: translateX(60px) translateY(-50px);
}

.tooltipimg img {
  max-width: 100%;
  max-height: 100%;
  border-radius: var(--xbvr-radius-sm, 8px);
}
</style>
