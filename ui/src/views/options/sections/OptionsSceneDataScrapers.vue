<template>
  <div class="content">
    <b-loading :is-full-page="true" :active.sync="isLoading"></b-loading>
    <header class="options-page-head options-head-row">
      <div>
        <h1 class="options-title">{{$t('Scrapers')}}</h1>
        <p class="options-desc">{{ $t('Enable studios and control how their scenes get scraped.') }}</p>
      </div>
      <div class="options-head-actions buttons">
        <b-dropdown aria-role="list" position="is-bottom-left">
          <template slot="trigger">
            <b-button icon-left="cog" />
          </template>
          <b-dropdown-item aria-role="listitem" custom>
            <div class="field">
              <b-checkbox v-model="$store.state.optionsAdvanced.advanced.autoLimitScraping" @input="saveAdvancedSettings">
                {{$t('Auto Limit Scraping')}}
              </b-checkbox>
            </div>
          </b-dropdown-item>
        </b-dropdown>
        <a class="button" :class="[showAllScrapers ? '' : 'is-info']" v-on:click="toggleEnabledFilter">
          {{showAllScrapers ? $t('Show enabled only') : $t('Show all scrapers')}}
        </a>
        <a class="button is-primary" v-on:click="taskScrape('_enabled')">{{$t('Run selected scrapers')}}</a>
      </div>
    </header>
    <div class="settings-card table-card">
    <b-table :data="scraperList" ref="scraperTable">
      <b-table-column field="is_enabled" :label="$t('Enabled')" v-slot="props" width="80" sortable>
          <span><b-switch v-model ="props.row.is_enabled" @input="$store.dispatch('optionsSites/toggleSite', {id: props.row.id})"/></span>
      </b-table-column>
      <b-table-column field="icon" width="50" v-slot="props" cell-class="narrow">
            <span class="image is-32x32">
              <vue-load-image>
                <img slot="image" :src="getImageURL(props.row)"/>
                <img slot="preloader" src="/ui/images/blank.png"/>
                <img slot="error" src="/ui/images/blank.png"/>
              </vue-load-image>
            </span>
      </b-table-column>
      <b-table-column field="sitename" :label="$t('Studio')" sortable searchable v-slot="props">
        <b-tooltip class="is-warning" :active="props.row.has_scraper == false" :label="$t('Scraper does not exist')"  :delay="250" >
          <a @click="navigateToStudio(props.row.name)" :class="[props.row.has_scraper ? 'has-text-link' : 'has-text-danger']" class="clickable">{{ props.row.sitename }}</a>
        </b-tooltip>
      </b-table-column>
      <b-table-column field="source" :label="$t('Source')" sortable searchable v-slot="props">
        {{ props.row.source }}
      </b-table-column>
      <b-table-column field="last_update" :label="$t('Last scrape')" sortable v-slot="props" cell-class="no-wrap">
            <span :class="[runningScrapers.includes(props.row.id) ? 'invisible' : '']">
              <span v-if="props.row.last_update !== '0001-01-01T00:00:00Z'">
                {{formatCompactTimeAgo(props.row.last_update)}}</span>
              <span v-else>-</span>
            </span>
            <span :class="[runningScrapers.includes(props.row.id) ? 'scraping-container' : 'invisible']">
              <span class="loading-dots">
                <span>.</span><span>.</span><span>.</span>
              </span>
            </span>
      </b-table-column>
      <b-table-column field="limit_scraping" :label="$t('Limit Scraping')" v-slot="props" width="60" sortable>
        <b-tooltip class="is-info" :label="$t('Limit scraping to newest scenes on the website. Turn off if you are missing scenes.')" :delay="250" >
          <span><b-switch v-model ="props.row.limit_scraping" @input="$store.dispatch('optionsSites/toggleLimitScraping', {id: props.row.id})"/></span>
        </b-tooltip>
      </b-table-column>
      <b-table-column field="subscribed" :label="$t('Subscribed')" v-slot="props" width="60" sortable>
        <b-tooltip class="is-info" :label="$t('Highlights this studio in the scene view and includes scenes in the &quot;Has subscription&quot; attribute filter')" :delay="250" >
          <span v-if="props.row.master_site_id==''"><b-switch v-model ="props.row.subscribed" @input="$store.dispatch('optionsSites/toggleSubscribed', {id: props.row.id})"/></span>
        </b-tooltip>
      </b-table-column>
      <b-table-column field="scrape_stash" :label="$t('Scrape Stash')" v-slot="props" width="60" sortable>
        <b-tooltip class="is-info" :label="$t('Enables scraping Stashdb for Actors')" :delay="250" >
          <span v-if="props.row.master_site_id==''"><b-switch v-model ="props.row.scrape_stash" @input="$store.dispatch('optionsSites/toggleScrapeStash', {id: props.row.id})"/></span>
        </b-tooltip>
      </b-table-column>
      <b-table-column field="scene_count" :label="$t('Scenes')" v-slot="props" width="40" sortable numeric>
        <a @click="navigateToStudio(props.row.name)" class="clickable">
          <span class="tag is-info is-light is-medium"><strong>{{ props.row.scene_count }}</strong></span>
        </a>
      </b-table-column>
      <b-table-column field="options" v-slot="props" width="30">
        <div class="menu">
          <b-dropdown aria-role="list" class="is-pulled-right" position="is-bottom-left">
            <template slot="trigger">
              <b-icon icon="dots-vertical mdi-18px"></b-icon>
            </template>
            <b-dropdown-item v-if="props.row.has_scraper" aria-role="listitem" @click="taskScrape(props.row.id)">
              {{$t('Run this scraper')}}
            </b-dropdown-item>
            <b-dropdown-item v-if="props.row.has_scraper && props.row.id != 'baberoticavr'" aria-role="listitem" @click="taskScrapeScene(props.row.id)">
              {{$t('Scrape Single Scene')}}
            </b-dropdown-item>
            <b-dropdown-item v-if="props.row.has_scraper && props.row.master_site_id==''" aria-role="listitem" @click="forceSiteUpdate(props.row.name, props.row.id)">
              {{$t('Force update scenes')}}
            </b-dropdown-item>
            <b-dropdown-item v-if="props.row.has_scraper && props.row.master_site_id!=''" aria-role="listitem" @click="removeSceneLinks(props.row, true)">
              {{$t('Remove Scene Links')}}
            </b-dropdown-item>
            <b-dropdown-item v-if="props.row.has_scraper && props.row.master_site_id!=''" aria-role="listitem" @click="removeSceneLinks(props.row, false)">
              {{$t('Remove Scene Links (Keep edits)')}}
            </b-dropdown-item>
            <b-dropdown-item aria-role="listitem" @click="deleteScenes(props.row)">
              {{$t('Delete scraped scenes')}}
            </b-dropdown-item>
            <b-dropdown-item aria-role="listitem" @click="scrapeActors(props.row.name, props.row.id)" v-if="props.row.master_site_id==''">
              {{$t('Scrape Actor Details from Site')}}
            </b-dropdown-item>
          </b-dropdown>
        </div>
      </b-table-column>
      <b-table-column field="master_site_id" :label="$t('Main Site')" v-slot="props" width="60" sortable>
        <span>
          <a @click="editMatchParams(props.row)" title="Edit Scene Matching Parameters" v-if="props.row.master_site_id != ''"> 
            <b-icon pack="mdi" icon="cog-outline" size="is-small"/>
          </a>
          {{getMasterSiteName(props.row.master_site_id)}}
        </span>
      </b-table-column>
    </b-table>
    </div>
    <div class="footer-actions buttons">
      <a class="button is-small" v-on:click="toggleAllLimitScraping()">{{$t('Toggle Limit Scraping of all visible sites')}}</a>
      <a class="button is-small" v-on:click="toggleAllSubscriptions()">{{$t('Toggle Subscriptions of all visible sites')}}</a>
    </div>

    <b-modal :active.sync="isSingleScrapeModalActive"
             has-modal-card
             trap-focus
             aria-role="dialog"
             aria-modal>
      <div class="modal-card modal-card-auto">
        <header class="modal-card-head">
          <p class="modal-card-title">{{$t('Additional Details Required')}}</p>
        </header>
        <section class="modal-card-body">
          <b-field v-if="additionalInfoIdx == 0 && this.scraperwarning != ''"><span>{{this.scraperwarning}}</span></b-field>
          <b-field v-if="additionalInfoIdx == 0 && this.scraperwarning2 != ''"><span>{{this.scraperwarning2}}</span></b-field>          
          <b-field :label=this.additionalInfo[additionalInfoIdx].fieldPrompt>
            <b-input v-if="additionalInfo[additionalInfoIdx].type != 'checkbox'"
              :type=additionalInfo[additionalInfoIdx].type
              v-model='additionalInfo[additionalInfoIdx].fieldValue'
              :required=additionalInfo[additionalInfoIdx].required
              :placeholder=additionalInfo[additionalInfoIdx].placeholder                            
              ref="additionInfoInput"
              >
            </b-input>
            <b-checkbox v-if="additionalInfo[additionalInfoIdx].type == 'checkbox'" v-model="additionalInfo[additionalInfoIdx].fieldValue">{{this.additionalInfo[additionalInfoIdx].fieldPrompt}}</b-checkbox>
          </b-field>
        </section>
        <footer class="modal-card-foot">
          <button class="button is-primary" :disabled="this.additionalInfo[additionalInfoIdx].required && this.additionalInfo[additionalInfoIdx].fieldValue == ''" @click="taskScrapeSceneInfoEntered()">Continue
          </button>
        </footer>
      </div>
    </b-modal>

  </div>

</template>

<script>
import api from '../../../api'
import { getImageURL as getImageURLUtil, iconSlug } from '../../../util/image'
import VueLoadImage from 'vue-load-image'
import { formatDistanceToNow, parseISO } from 'date-fns'

export default {
  name: 'OptionsSites',
  components: { VueLoadImage },
  data () {
    return {
      javrQuery: '',
      tpdbSceneUrl: '',
      isLoading: false,
      sceneUrl: '',
      isSingleScrapeModalActive: false,
      additionalInfo: [{fieldName: "scene_url", fieldPrompt: "Scene Url", placeholder: "eg https://www.mysite.com/scenes/my scene", fieldValue: '', required: true, type: 'url' }],
      additionalInfoIdx: 0,
      currentScraper: '',
      scraperwarning: '',
      scraperwarning2: '',
      showAllScrapers: true,
    }
  },
  mounted () {
    this.$store.dispatch('optionsSites/load')
    this.$store.dispatch('optionsAdvanced/load')
    this.$store.dispatch('optionsWeb/load')
  },
  methods: {
    getImageURL (row) {
      return getImageURLUtil(row.avatar_url ? row.avatar_url : '/ui/images/blank.png', '128x', 'icon-' + iconSlug(row.id))
    },
    taskScrape (scraper) {
      api.get(`/task/scrape?site=${scraper}`)
    },
    taskScrapeScene (scraper) {
      this.currentScraper=scraper      
      this.additionalInfo = [{fieldName: "scene_url", fieldPrompt: "Scene Url", placeholder: "Enter the url for a VR Scene", fieldValue: '', required: true, type: 'url'}]      
      this.scraperwarning = "Take care to only use scene urls for the " + scraper + " Scraper"
      this.scraperwarning2 = ""
      switch (scraper) {
        case 'wankzvr':
        case 'milfvr':
        case 'herpovr':
        case  'brasilvr':
        case 'tranzvr':
          this.scraperwarning = "Only use povr.com urls for the " + scraper + " Scraper"
          break
        case 'tonightsgirlfriend':
          this.scraperwarning2 = "Warning " + scraper + " also includes 2d scenes, only select scenes from their VR section"
        case 'naughtyamericavr':
          this.scraperwarning2 = "Warning The NaughtyAmerica site also includes 2d scenes, only select scenes from their VR section"
          break
    }
      this.additionalInfoIdx=0
      this.isSingleScrapeModalActive = true      
    },
    taskScrapeSceneInfoEntered () {      
      const inputElement = this.$refs.additionInfoInput
      if (!inputElement.isValid) {
        // get the field again
        this.isSingleScrapeModalActive = true
        return
      }

      this.isSingleScrapeModalActive = false      
      var fieldCheckMsg = ""
      if (this.additionalInfo[0].fieldValue.toLowerCase().includes('fuckpassvr.com')) {
        var fieldCheckMsg="Note: Video Previews are not available when scraping single scenes from FuckpassVR"
      }
      if (this.additionalInfo[0].fieldValue.toLowerCase().includes('lethalhardcorevr.com')) {
        var fieldCheckMsg=`Please check the Site if the scene was for WhorecraftVR. Please check the Release Date`
      }
      if (this.additionalInfo[0].fieldValue.toLowerCase().includes('littlecaprice-dreams.com')) {
        var fieldCheckMsg=`Please specify a URL for the cover image`
      }
      if (this.additionalInfo[0].fieldValue.toLowerCase().includes('sexbabesvr.com')) {
        var fieldCheckMsg="Please check the Release Date"
      }
      if (this.additionalInfo[0].fieldValue.toLowerCase().includes('stasyqvr.com')) {
        var fieldCheckMsg=`Please specify a Duration if required`
      }
      if (this.additionalInfo[0].fieldValue.toLowerCase().includes('tonightsgirlfriend.com')) {
        var fieldCheckMsg="Please check the Release Date"
      }
      if (this.additionalInfo[0].fieldValue.toLowerCase().includes('virtualporn.com')) {
        var fieldCheckMsg=`Please check the Release Date and specify a Duration if required`
      }
      if (this.additionalInfo[0].fieldValue.toLowerCase().includes('wetvr.com')) {        
        var fieldCheckMsg="Please check the Release Date"
      }

      if (this.additionalInfoIdx == 0) {
        if (this.additionalInfo[0].fieldValue.toLowerCase().includes('wetvr.com')) {
          this.additionalInfo.push({fieldName: "scene_id", fieldPrompt: "Scene Id", placeholder: "eg 69037 (excl site prefix)", fieldValue: '', required: true, type: 'number'})
        }
      }
      
      this.additionalInfo[this.additionalInfoIdx].fieldValue = this.additionalInfo[this.additionalInfoIdx].fieldValue.trim()
      if (this.additionalInfoIdx + 1 < this.additionalInfo.length) {          
        this.additionalInfoIdx = this.additionalInfoIdx +1
        this.isSingleScrapeModalActive = true      
      } else {
        if (fieldCheckMsg != "") {
          this.$buefy.toast.open({message: `Scene scraping in progress, please wait for the Scene Detail popup`, type: 'is-warning', duration: 5000})
        } else {
          this.$buefy.toast.open({message: `Scene scraping in progress`, type: 'is-warning', duration: 5000})
        }
        // scrapes can run long and the response is consumed below — keep the
        // historical unbounded timeout instead of the 60s shared default
        api.post(`/task/singlescrape`, {timeout: false, json: { site: this.currentScraper, sceneurl: this.additionalInfo[0].fieldValue, additionalinfo: this.additionalInfo.slice(1)}})
        .json()
        .then(data => { 
          if (data.status == 'OK') {          
            this.$store.commit('overlay/editDetails', { scene: data.scene })
            if (fieldCheckMsg != "") {
              this.$buefy.toast.open({message: fieldCheckMsg, type: 'is-warning', duration: 10000})
            }
          }
        })
      }
    },
    forceSiteUpdate (site, scraper) {
      api.post('/options/scraper/force-site-update', {
        json: { scraper_id: scraper }
      })
      this.$buefy.toast.open(`Scenes from ${site} will be updated on next scrape`)
    },
    deleteScenes (site) {
      const self = this
      this.$buefy.dialog.confirm({
        title: this.$t('Delete scraped scenes'),
        message: `You're about to delete scraped scenes for <strong>${site.name}</strong>.`,
        type: 'is-danger',
        hasIcon: true,
        onConfirm: function () {
          if (site.master_site_id==""){
            api.post('/options/scraper/delete-scenes', {
              json: { scraper_id: site.id }
            }).then(() => {
              self.$store.dispatch('optionsSites/load')
            })
          } else {
            const external_source = 'alternate scene ' + site.id
            api.delete(`/extref/delete_extref_source`, {
              json: {external_source: external_source}
            }).then(() => {
              self.$store.dispatch('optionsSites/load')
            });
          }
        }
      })
    },
    removeSceneLinks (site, all) {
      this.$buefy.dialog.confirm({
        title: this.$t('Remove Scene Links'),
        message: `You're about to remove links for scenes from <strong>${site.name}</strong>. Scenes will be relinked after the next scrape.`,
        type: 'is-warning',
        hasIcon: true,
        onConfirm: function () {
          const external_source = 'alternate scene ' + site.id          
          if (all) {
            api.delete(`/extref/delete_extref_source_links/all`, {
              json: {external_source: external_source}
            });
          } else {
            api.delete(`/extref/delete_extref_source_links/keep_manual`, {
              json: {external_source: external_source}
            });
          }
        }
      })
    },
    scrapeActors(site, scraper) {      
      api.get('/extref/generic/scrape_by_site/' + scraper)
      this.$buefy.toast.open(`Scraping Actor Details from ${site}`)
    },
    async toggleAllSubscriptions(){
      const table = this.$refs.scraperTable;
      this.isLoading=true
      // one bulk call instead of a PUT per site plus a reload each
      await api.put('/options/sites/toggle_field', { json: { field: 'Subscribed', ids: table.newData.map(s => s.id) } }).json()
      this.$store.dispatch('optionsSites/load')
      this.isLoading=false
    },
    async toggleAllLimitScraping(){
      const table = this.$refs.scraperTable;
      this.isLoading=true
      await api.put('/options/sites/toggle_field', { json: { field: 'LimitScraping', ids: table.newData.map(s => s.id) } }).json()
      this.$store.dispatch('optionsSites/load')
      this.isLoading=false
    },
    editMatchParams(site){
      this.$store.commit('overlay/showSceneMatchParams', { site: site })
    },
    getMasterSiteName(siteId){
      if (siteId=="") {
        return ""
      }
      return  this.scraperList.find(element => element.id === siteId).name;
    },
    navigateToStudio(studioName) {
      // Handle different scraper types:
      // - Built-in scrapers (e.g., "AstroDomina (SLR)"): scenes use just "AstroDomina"
      // - Custom scrapers (e.g., "LethalhardcoreVR (Custom SLR)"): scenes use "LethalhardcoreVR (SLR)"
      let siteName = studioName

      // Handle custom scrapers from any aggregator
      const customMatch = siteName.match(/^(.+) \(Custom ([A-Z]+)\)$/)
      if (customMatch) {
        // Custom scrapers: replace "(Custom XXX)" with "(XXX)"
        siteName = customMatch[1] + ' (' + customMatch[2] + ')'
      } else {
        // Handle built-in scrapers with aggregator suffix
        const builtinMatch = siteName.match(/^(.+) \(([A-Z]+)\)$/)
        if (builtinMatch) {
          // Built-in scrapers: remove aggregator suffix
          siteName = builtinMatch[1]
        }
      }

      // Set the site filter and navigate to scenes page
      this.$store.state.sceneList.filters.sites = [siteName]
      this.$store.state.sceneList.filters.tags = []
      this.$store.state.sceneList.filters.attributes = []
      this.$router.push({
        name: 'scenes',
        query: { q: this.$store.getters['sceneList/filterQueryParams'] }
      })
    },
    toggleEnabledFilter() {
      this.showAllScrapers = !this.showAllScrapers
    },
    saveAdvancedSettings() {
      this.$store.dispatch('optionsAdvanced/save')
    },
    formatCompactTime(isoString) {
      const date = parseISO(isoString)
      const month = date.getMonth() + 1
      const day = date.getDate()
      const year = date.getFullYear()
      return `${month}/${day}/${year}`
    },
    formatCompactTimeAgo(isoString) {
      const now = new Date()
      const date = parseISO(isoString)
      const seconds = Math.floor((now - date) / 1000)

      if (seconds < 60) return 'just now'
      const minutes = Math.floor(seconds / 60)
      if (minutes < 60) return `${minutes}min ago`
      const hours = Math.floor(minutes / 60)
      if (hours < 24) return `${hours}h ago`
      const days = Math.floor(hours / 24)
      if (days < 7) return `${days}d ago`
      const weeks = Math.floor(days / 7)
      if (weeks < 4) return `${weeks}w ago`
      const months = Math.floor(days / 30)
      if (months < 12) return `${months}mo ago`
      const years = Math.floor(days / 365)
      return `${years}y ago`
    },
    parseISO,
    formatDistanceToNow
  },
  computed: {
    scraperList() {
      var items = this.$store.state.optionsSites.items;
      let re = /(.*)\s+\((.+)\)$/;
      for (let i=0; i < items.length; i++) {
        items[i].sitename = items[i].name;
        items[i].source = "";

        var m = re.exec(items[i].name);
        if (m) {
          items[i].sitename = m[1];
          items[i].source = m[2];
        }
      }

      // Filter by enabled status if the filter is active
      if (!this.showAllScrapers) {
        items = items.filter(item => item.is_enabled === true);
      }

      return items;
    },
    items () {
      return this.$store.state.optionsSites.items
    },
    runningScrapers () {
      this.$store.dispatch('optionsSites/load')
      return this.$store.state.messages.runningScrapers
    }
  }
}
</script>

<style scoped>
  .options-page-head {
    margin-bottom: 1.25rem;
  }

  .options-head-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 1rem;
    flex-wrap: wrap;
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

  .options-head-actions {
    justify-content: flex-end;
  }

  .settings-card {
    background: var(--xbvr-surface, #ffffff);
    border: 1px solid var(--xbvr-border, #e3e6ec);
    border-radius: var(--xbvr-radius, 12px);
    box-shadow: var(--xbvr-shadow-sm, 0 1px 2px rgba(16, 24, 40, 0.05));
    padding: 0.5rem 1rem 1rem;
    margin-bottom: 1.25rem;
  }

  .table-card :deep(.table) {
    border-radius: var(--xbvr-radius-sm, 8px);
    background: transparent;
  }

  .table-card :deep(.table tbody tr) {
    transition: background-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
  }

  .table-card :deep(.table tbody tr:hover) {
    background: var(--xbvr-hover-bg, #fafbfd);
  }

  .footer-actions {
    justify-content: flex-end;
  }

  .modal-card-auto {
    width: auto;
  }

  .running {
    opacity: 0.6;
    pointer-events: none;
  }

  .tag.is-medium {
    padding-left: 0.5em;
    padding-right: 0.5em;
    transition: background-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
      color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
  }

  a:hover .tag.is-medium {
    background-color: var(--xbvr-info, #3e8ed0) !important;
    color: var(--xbvr-on-primary, #ffffff) !important;
  }

  .clickable {
    cursor: pointer;
  }

  .invisible {
    display: none;
  }

  .scraping-container {
    display: inline-block;
    vertical-align: middle;
  }

  /* Animated ellipsis for scraping indicator */
  .loading-dots {
    display: inline-block;
    font-size: 1.2em;
    font-weight: bold;
    color: var(--xbvr-primary, #4f46e5);
    line-height: 1;
    vertical-align: middle;
    position: relative;
    top: -0.4em;
  }

  .loading-dots span {
    animation: blink 1.4s infinite both;
    display: inline-block;
  }

  .loading-dots span:nth-child(2) {
    animation-delay: 0.2s;
  }

  .loading-dots span:nth-child(3) {
    animation-delay: 0.4s;
  }

  @keyframes blink {
    0%, 80%, 100% {
      opacity: 0;
    }
    40% {
      opacity: 1;
    }
  }
</style>

<style>
  .content table td.narrow{
    padding-top: 5px;
    padding-bottom: 2px;
  }

  .content table th .icon {
    display: none;
  }

  .content table td.no-wrap {
    white-space: nowrap;
  }

  .content table td {
    vertical-align: middle;
  }
</style>
