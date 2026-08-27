<template>
    <b-modal :active="isModalActive"           
           has-modal-card
           trap-focus
           aria-role="dialog"
           @cancel="close"
           aria-modal
           can-cancel>
    
    <GlobalEvents
      :filter="e => !['INPUT', 'TEXTAREA'].includes(e.target.tagName)"
      @keyup.esc="close"
      @keydown.k="handleLeftArrow"
      @keydown.l="handleRightArrow"
    />

    <div class="modal-card search-card">
      <header class="modal-card-head">
        <p class="modal-card-title">Search Stashdb Actors</p>
        <button class="delete" @click="close" aria-label="close"></button>
      </header>

      <div class="modal-card-body">
        <div class="search-box">
          <b-field label="Find actor...">
            <b-input v-model="queryString" placeholder="Find actor..." @input="debouncedSearch" :loading="isFetching" custom-class="is-large"/>
          </b-field>
        </div>

        <b-table :data="searchResults" @click="onRowSelected" class="results-table">
          <b-table-column field="Name" >
            <template slot-scope="props">
              <div class="media result-row">
                <div class="media-left">
                  <b-carousel 
                    :ref="'actorcarousel' + props.index"
                    :id="'actorcarousel' + props.index"
                    :autoplay="false" :indicator="false" icon-size="is-small" width="120"
                  >
                    <b-carousel-item v-for="(image, index) in props.row.ImageUrl" :key="index">
                          <vue-load-image height="50px">
                            <img slot="image" :src="image && image.length ? getImageURL(image) : '/ui/images/blank_female_profile.png'" width="100" class="result-thumb" @mouseover="setShowTooltipImage(image, props.row.Id)" @mouseout="setShowTooltipImage('','')"/>
                            <img slot="preloader" src="/ui/images/blank.png" width="100" class="result-thumb"/>
                            <img slot="error" src="/ui/images/blank.png" width="100" class="result-thumb"/>
                          </vue-load-image>
                    </b-carousel-item>
                  </b-carousel>
                    <div v-if="tooltipImage!='' && tooltipID==props.row.Id" class="tooltipimg" @mouseout="setShowTooltipImage('','')">                      
                          <vue-load-image width="300">
                            <img slot="image" :src="tooltipImage" width="300"  />
                            <img slot="preloader" src="/ui/images/blank.png" width="300" />
                            <img slot="error" src="/ui/images/blank.png" width="300" />
                          </vue-load-image>
                    </div>
                  <div v-if="props.row.DOB">
                    <span class="smaller-text">
                      <strong>Birth Date:</strong>
                    </span>
                  </div>
                  <div v-if="props.row.DOB">
                    <span class="smaller-text">{{ format(parseISO(props.row.DOB), "yyyy-MM-dd") }}</span>
                  </div>
                  <div>
                    <span class="smaller-text">
                      <strong>Score:</strong> {{ props.row.Weight }}
                    </span>
                  </div>
                  <div>
                    <a class="button is-primary is-small link-button" @click="linktoStashdb(props.row)" :title="'Link Actor with stashdb'">
                      <b-icon pack="mdi" :icon="'link-variant-plus'" size="is-small" />
                    </a>
                  </div>
                </div>
                <div class="media-content">
                  <div class="truncate">
                    <strong>
                      <a :href="props.row.Url" target="_blank">{{ props.row.Name }} - {{ props.row.Disambiguation }}</a>
                    </strong>
                  </div>
                  <div>
                    <strong>Aliases:</strong>
                    <b-tag v-for="alias in props.row.Aliases" :key="alias.Alias" :class="{ 'is-primary': alias.Matched }" class="result-tag"> {{ alias.Alias }}</b-tag>
                  </div>
                  <div>                    
                    <b-tag v-for="link in props.row.Studios" :key="link.url" class="result-tag"><a :href="link.Url" :class="{ 'bold-tag': link.Matched }" target="_blank">{{ link.Name }}({{ link.SceneCount }})</a></b-tag>
                  </div>
                </div>
              </div>
            </template>
          </b-table-column>
        </b-table>
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
  name: 'SearchStashdbActors',
  components: {  GlobalEvents, VueLoadImage },
  data () {
    return {
        isModalActive: true,
        stashdbUrl: "",
        searchResults: [],
        queryString: "",
        isFetching: false,
        tooltipImage: "",
        tooltipID: "",
        actor: "",
        selectedRow: undefined,
        }        
  },
  created() {
    this.debouncedSearch = debounce(this.searchStashdb, 750); // 750ms delay
  },
  mounted () {
    const item = Object.assign({}, this.$store.state.overlay.searchStashDbActors.actor)    
    this.actor = item
    this.openDialog(item)
    this.queryString=this.actor.name
  },
  methods: {
    format,
    parseISO,
    close () {
      this.$store.commit('overlay/hideSearchStashdbActors')
    },
    searchStashdb() {
      this.$buefy.toast.open({message: `Searching Actors`, type: 'is-primary', duration: 5000})
        api.get('/extref/stashdb/searchactor/' + this.actor.id + "?q=" + this.queryString, {timeout: 6e6}).json().then(data => {
            this.searchResults = Object.values(data.Results).sort((a, b) => b.Weight - a.Weight)
            this.isModalActive = true
            if (data.Status!='') {
              this.$buefy.toast.open({message: `Warning:  ${data.Status}`, type: 'is-warning', duration: 5000})
            }
        })
    },
    selectActor(option) {
        this.stashdbUrl=option.Url.replace("https://stashdb.org/performers/","")
        this.$nextTick(() => {
            if (this.$refs.autocompleteInput) {
                this.$refs.autocompleteInput.focus();
            }
        });
    },    
    linktoStashdb(option) {
        this.stashdbUrl=option.Url.replace("https://stashdb.org/performers/","")
        api.get('/extref/stashdb/link2actor/' + this.actor.id +'/'+this.stashdbUrl ).json().then(data => {          
          // this.$store.commit('sceneList/updateScene', data)
           this.$store.commit('overlay/showActorDetails', { actor: data })
          this.close()
        })
    },    
    getImageURL (u) {
      // stashdb performers have no local actor id — unattributed context
      return getImageURLUtil(u, '120x', 'act-0')
    },
    openDialog(actor) {
        this.isModalActive = true
        this.searchStashdb()
        this.$nextTick(() => {
            if (this.$refs.autocompleteInput) {
                this.$refs.autocompleteInput.focus();
            }
        });
        this.actor = actor
    },
    setShowTooltipImage(val, id){
      this.tooltipImage=val
      this.tooltipID=id
    },
    handleLeftArrow() {
      let idx=0
      if (this.selectedRow!=undefined && this.searchResults.length){
        idx=this.searchResults.findIndex(element => element.Id==this.selectedRow.Id) 
      }
      let selectedCarousel = this.$refs['actorcarousel' + idx]
      selectedCarousel.prev()
},
    handleRightArrow() {
      let idx=0
      if (this.selectedRow!=undefined && this.searchResults.length){
        idx=this.searchResults.findIndex(element => element.Id==this.selectedRow.Id) 
      }
      let selectedCarousel = this.$refs['actorcarousel' + idx]
      selectedCarousel.next()
    },
    onRowSelected(row) {
      this.selectedRow= row      
    },
},
  computed: {

}
}
</script>

<style scoped>
.search-card {
  height: 80vh;
  width: min(900px, 60vw);
}

@media (max-width: 1024px) {
  .search-card {
    width: 92vw;
  }
}

.search-card .modal-card-body {
  overflow-y: auto;
}

.search-box {
  margin-bottom: 1rem;
}

/* result rows */
.results-table :deep(tbody tr) {
  cursor: pointer;
  transition: background-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

.results-table :deep(tbody tr:hover) {
  background-color: var(--xbvr-hover-bg, #fafbfd);
}

.results-table :deep(tbody tr td) {
  border-color: var(--xbvr-border, #e3e6ec);
}

.result-row {
  align-items: flex-start;
}

.result-thumb {
  border-radius: var(--xbvr-radius-sm, 8px);
  border: 1px solid var(--xbvr-border, #e3e6ec);
  object-fit: cover;
}

.result-tag {
  margin-right: 4px;
  border-radius: 999px;
}

.link-button {
  border-radius: 8px;
  margin-top: 0.4rem;
}

.tab-item {
  height: 40vh;
}

/* hover preview popover */
.tooltipimg {
  position: absolute;
  z-index: 1;
  width: 350;
  background-color: var(--xbvr-surface, #ffffff);
  border: 1px solid var(--xbvr-border, #e3e6ec);
  border-radius: var(--xbvr-radius, 12px);
  box-shadow: var(--xbvr-shadow-lg, 0 16px 40px rgba(16, 24, 40, 0.16));
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

.smaller-text {
  font-size: 0.8em;
  color: var(--xbvr-text-muted, #64708a);
}

.bold-tag {
  font-weight: bold;
}
</style>
