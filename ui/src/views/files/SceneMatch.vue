<template>
  <div class="modal is-active" role="dialog" aria-modal="true">
    <GlobalEvents
      :filter="e => !['INPUT', 'TEXTAREA'].includes(e.target.tagName)"
      @keyup.esc="close"
      @keydown.left="handleLeftArrow"
      @keydown.right="handleRightArrow"
      @keydown.o="prevFile"
      @keydown.p="nextFile"
    />
    <div class="modal-background"></div>
    <div class="modal-card match-card">
      <header class="modal-card-head">
        <p class="modal-card-title">{{ $t("Match file to scene") }}</p>
        <button class="delete" @click="close" aria-label="close"></button>
      </header>
      <section class="modal-card-body match-body">
        <header class="match-header">
          <h2 class="match-title">{{ file.filename }}</h2>
          <p class="pathDetails">{{ file.path }}</p>
          <div class="meta-chips">
            <span class="meta-chip">{{ prettyBytes(file.size) }}</span>
            <span v-if="file.type == 'video'" class="meta-chip">{{ file.video_width }}x{{ file.video_height }}</span>
            <span v-if="file.duration > 0" class="meta-chip">{{ Math.floor(file.duration / 60) }} min</span>
            <span class="meta-chip">{{ format(parseISO(file.created_time), "yyyy-MM-dd") }}</span>
          </div>
        </header>

        <div class="search-chips">
          <span class="search-chips-label">{{$t('Search Fields')}}</span>
          <b-tooltip :label="$t('Optional: select one or more words to target searching to a specific field')" :delay="500" position="is-top">
            <b-button @click='searchPrefix("+title:")' class="tag is-info is-small is-light">title:</b-button>
            <b-button @click='searchPrefix("cast:")' class="tag is-info is-small is-light">cast:</b-button>
            <b-button @click='searchPrefix("+site:")' class="tag is-info is-small is-light">site:</b-button>
            <b-button @click='searchPrefix("+id:")' class="tag is-info is-small is-light">id:</b-button>
          </b-tooltip>
          <b-tooltip :label="$t('Add file duration to search')" :delay="500" position="is-top">
            <b-button @click='searchDurationPrefix("duration:")' class="tag is-info is-small is-light">duration:</b-button>
          </b-tooltip>
          <b-tooltip :label="$t('Defaults date range to the last week. Note:must match yyyy-mm-dd, include leading zeros')" :delay="500" position="is-top">
            <b-button @click='searchDatePrefix("released:")' class="tag is-info is-small is-light">released:</b-button>
            <b-button @click='searchDatePrefix("added:")' class="tag is-info is-small is-light">added:</b-button>
          </b-tooltip>
        </div>

        <b-field :label="$t('Search')" label-position="on-border" class="search-field">
          <div class="control">
            <input class="input" type="text" v-model='queryString' v-debounce:200ms="loadData" autofocus ref="searchInput">
          </div>
        </b-field>

        <b-table :data="data" ref="table" paginated :current-page.sync="currentPage" per-page="5" class="match-table">
          <b-table-column field="cover_url" :label="$t('Image')" width="120" v-slot="props">
            <vue-load-image>
              <img slot="image" :src="getImageURL(props.row.cover_url)" class="cover-thumb"/>
              <img slot="preloader" src="/ui/images/blank.png" class="cover-thumb"/>
              <img slot="error" src="/ui/images/blank.png" class="cover-thumb"/>
            </vue-load-image>
          </b-table-column>
          <b-table-column field="site" :label="$t('Site')" sortable v-slot="props">
            <a class="site-link" :href="safeHref(props.row.scene_url)" target="_blank" rel="noreferrer">{{ props.row.site }}</a>
            <div class="site-tags">
              <b-tooltip v-if="props.row.is_hidden" label="Flagged as Hidden"  :delay="250" >
                <b-tag type="is-info is-light" >
                  <b-icon pack="mdi" icon="eye-off-outline" size="is-small" class="tag-icon"/>
                </b-tag>
              </b-tooltip>
              <b-tag type="is-info is-light" v-if="videoFilesCount(props.row)">
                <b-icon pack="mdi" icon="file" size="is-small" class="tag-icon"/>
                {{videoFilesCount(props.row)}}
              </b-tag>
              <b-tag type="is-info is-light" v-if="props.row.is_scripted">
                <b-icon pack="mdi" icon="pulse" size="is-small"/>
                <span v-if="scriptFilesCount(props.row) > 1">{{scriptFilesCount(props.row)}}</span>
              </b-tag>
              <b-tag type="is-info is-light" v-if="subtitlesFilesCount(props.row)">
                <b-icon pack="mdi" icon="subtitles" size="is-small" class="tag-icon"/>
                {{subtitlesFilesCount(props.row)}}
              </b-tag>
            </div>
          </b-table-column>
          <b-table-column field="title" :label="$t('Title')" sortable v-slot="props">
            <p v-if="props.row.title" class="scene-title">{{ props.row.title }}</p>
            <small class="cast-tags">
              <b-tag rounded v-for="i in props.row.cast" :key="i.id">{{ i.name }}</b-tag>
            </small>
          </b-table-column>
          <b-table-column field="release_date" :label="$t('Release date')" sortable nowrap v-slot="props">
            <span class="cell-meta">{{ format(parseISO(props.row.release_date), "yyyy-MM-dd") }}</span>
          </b-table-column>
          <b-table-column field="duration" :label="$t('Duration')" sortable nowrap v-slot="props">
            <span class="cell-meta">{{ props.row.duration > 0 ? props.row.duration + " min" : ""}}</span>
          </b-table-column>
          <b-table-column field="scene_id" :label="$t('ID')" sortable nowrap v-slot="props">
            <span class="cell-meta">{{ props.row.scene_id }}</span>
          </b-table-column>
          <b-table-column field="_score" :label="$t('Score')" sortable v-slot="props">
            <b-progress show-value :value="props.row._score * 100"></b-progress>
          </b-table-column>
          <b-table-column field="_assign" v-slot="props">
            <div class="row-actions">
              <button class="button is-small is-primary is-outlined" @click="assign(props.row.scene_id)">{{ $t("Assign") }}</button>
            </div>
          </b-table-column>
        </b-table>
      </section>
    </div>
    <a class="scene-nav prev" @click="prevFile" title="Keyboard shortcut: O">&#10094;</a>
    <a class="scene-nav next" @click="nextFile" title="Keyboard shortcut: P">&#10095;</a>
  </div>
</template>

<script>
import api from '../../api'
import { getImageURL as getImageURLUtil } from '../../util/image'
import { safeHref } from '../../util/url'
import { format, parseISO } from 'date-fns'
import prettyBytes from 'pretty-bytes'
import VueLoadImage from 'vue-load-image'
import GlobalEvents from 'vue-global-events'

export default {
  name: 'SceneMatch',
  components: { VueLoadImage, GlobalEvents },
  data () {
    return {
      data: [],
      dataNumRequests: 0,
      dataNumResponses: 0,
      currentPage: 1,
      queryString: '',
      format,
      parseISO
    }
  },
  computed: {
    file () {
      return this.$store.state.overlay.match.file
    }
  },
  mounted () {
    this.initView()
  },
  methods: {
    safeHref,
    initView () {
      const commonWords = [
        '180', '180x180', '2880x1440', '3d', '3dh', '3dv', '30fps', '30m', '360',
        '3840x1920', '4k', '5k', '5400x2700', '60fps', '6k', '7k', '7680x3840',
        '8k', 'fb360', 'fisheye190', 'funscript', 'cmscript', 'h264', 'h265', 'hevc', 'hq', 'hsp', 'lq', 'lr',
        'mkv', 'mkx200', 'mkx220', 'mono', 'mp4', 'oculus', 'oculus5k',
        'oculusrift', 'original', 'rf52', 'smartphone', 'srt', 'ssa', 'tb', 'uhq', 'vrca220', 'vp9'
      ]
      const isNotCommonWord = word => !commonWords.includes(word.toLowerCase()) && !/^[0-9]+p$/.test(word)

      this.data = []
      this.queryString = (
        this.file.filename
          .replace(/[._+'’`-]/g, ' ').replace(/\s+/g, ' ').trim()
          .split(' ').filter(isNotCommonWord).join(' '))
      this.loadData()
    },
    loadData: async function loadData () {
      const requestIndex = this.dataNumRequests
      this.dataNumRequests = this.dataNumRequests + 1

      const resp = await api.get('/scene/search', {
        searchParams: {
          q: this.queryString
        },
        timeout: 60000
      }).json()

      if (requestIndex >= this.dataNumResponses) {
        this.dataNumResponses = requestIndex + 1

        if (resp.scenes !== null) {
          this.data = resp.scenes
        } else {
          this.data = []
        }
        this.currentPage = 1
      }
    },
    getImageURL (u) {
      return getImageURLUtil(u, '120x')
    },
    assign: async function assign (scene_id) {
      await api.post('/files/match', {
        json: {
          file_id: this.toInt(this.$store.state.overlay.match.file.id),
          scene_id: scene_id
        }
      })

      this.$store.dispatch('files/load')

      const data = this.$store.getters['files/nextFile'](this.file)
      if (data !== null) {
        this.nextFile()
      } else {
        this.close()
      }
    },
    nextFile () {
      const data = this.$store.getters['files/nextFile'](this.file)
      if (data !== null) {
        this.$store.commit('overlay/showMatch', { file: data })
        this.initView()
      }
    },
    prevFile () {
      const data = this.$store.getters['files/prevFile'](this.file)
      if (data !== null) {
        this.$store.commit('overlay/showMatch', { file: data })
        this.initView()
      }
    },
    close () {
      this.$store.commit('overlay/hideMatch')
    },
    toInt (value, radix, defaultValue) {
      return parseInt(value, radix || 10) || defaultValue || 0
    },
    videoFilesCount (scene) {
      let count = 0      
      scene.file.forEach(obj => {
        if (obj.type === 'video') {
          count = count + 1
        }
      })
      return count
    },
    scriptFilesCount (scene) {
      let count = 0
      scene.file.forEach(obj => {
        if (obj.type === 'script') {
          count = count + 1
        }
      })
      return count
    },
    subtitlesFilesCount (scene) {
      let count = 0
      scene.file.forEach(obj => {
        if (obj.type === 'subtitles') {
          count = count + 1
        }
      })
      return count
    },
    handleRightArrow () {
      if ((this.currentPage) * 5 < this.data.length) {
        this.currentPage = this.currentPage + 1
      } else {
        this.currentPage = 1
      }
    },
    handleLeftArrow () {
      if (this.currentPage === 1) {
        // dont assume last page is 5
        this.currentPage = ~~((this.data.length + 4) / 5)
      } else {
        this.currentPage = this.currentPage - 1
      }
    },
    searchPrefix(prefix) {
      let textbox = this.$refs.searchInput
      if (textbox.selectionStart != textbox.selectionEnd) {
        let selected = textbox.value.substring(textbox.selectionStart, textbox.selectionEnd)
        selected=selected.replace(/_/g," ").replace(/-/g," ").trim()
        if (selected.indexOf(' ') >= 0)
        {
          selected='"' + selected + '"'
        }        
        this.queryString = textbox.value.substring(0,textbox.selectionStart) + " " + prefix + selected + " " + textbox.value.substr(textbox.selectionEnd)
        this.loadData()
      }
      
    },
    searchDatePrefix(prefix) {      
        let today = new Date().toISOString().slice(0, 10)
        let weekago = new Date(Date.now() - 604800000).toISOString().slice(0, 10)        
          this.queryString = this.queryString.trim() + ' ' + prefix + '>="' + weekago + '" ' +  prefix + '<="' + today + '"'        
        this.loadData()
    },
    searchDurationPrefix(prefix) {        
        if (this.file.duration==0) {
          this.queryString = this.queryString.trim() + ' ' + prefix + '>=0 '
        } else {
          this.queryString = this.queryString.trim() + ' ' + prefix + '>=' + (Math.floor(this.file.duration / 60)-1) + ' ' +  prefix + '<=' + (Math.floor(this.file.duration / 60)+1) + ''        
        }
        this.loadData()
    },
    prettyBytes
  }
}
</script>

<style lang="less" scoped>
/* ------------------------------------------------------------------
   Scene-match overlay — same modal language as scenes/Details.vue
   ------------------------------------------------------------------ */

.match-card {
  position: relative;
  margin: 3rem auto;
  width: min(1200px, 94vw);
  max-height: calc(100vh - 6rem);
}

.match-body {
  padding: 1.25rem;
}

/* file header */
.match-header {
  margin-bottom: 0.9rem;
}

.match-title {
  font-size: 1.15rem;
  font-weight: 800;
  letter-spacing: -0.01em;
  line-height: 1.25;
  color: var(--xbvr-text, #1c2333);
  overflow-wrap: anywhere;
  margin-bottom: 0.25rem;
}

.pathDetails {
  font-size: 0.78rem;
  color: var(--xbvr-text-faint, #7d88a1);
  overflow-wrap: anywhere;
  margin-bottom: 0.55rem;
}

.meta-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.meta-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.32rem;
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--xbvr-text-muted, #64708a);
  background: var(--xbvr-surface-sunken, #eef0f4);
  border: 1px solid var(--xbvr-border, #e3e6ec);
  border-radius: 999px;
  padding: 0.25em 0.7em;
  font-variant-numeric: tabular-nums;
}

/* search-field prefix chips */
.search-chips {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.4rem;
  margin-bottom: 0.8rem;
}

.search-chips-label {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--xbvr-text-muted, #64708a);
  margin-right: 0.2rem;
}

.search-chips :deep(.tag) {
  border-radius: 999px;
  margin: 0;
}

.search-field {
  margin-bottom: 1rem;
}

/* ------------------------------------------------------------------
   Result table — rows read as rounded surface cards with hover state
   ------------------------------------------------------------------ */

.match-table :deep(.table) {
  background: transparent;
  border-collapse: separate;
  border-spacing: 0 0.55rem;
}

.match-table :deep(.table thead th) {
  background: transparent;
  border: none;
  padding: 0.2rem 0.85rem 0.4rem;
  white-space: nowrap;
}

.match-table :deep(.table thead th .th-wrap) {
  text-transform: uppercase;
  letter-spacing: 0.05em;
  font-size: 0.72rem;
  font-weight: 700;
  color: var(--xbvr-text-muted, #64708a);
}

.match-table :deep(.table tbody tr) {
  background: transparent;
}

.match-table :deep(.table tbody td) {
  background: var(--xbvr-surface, #ffffff);
  border-top: 1px solid var(--xbvr-border, #e3e6ec);
  border-bottom: 1px solid var(--xbvr-border, #e3e6ec);
  border-left: none;
  border-right: none;
  vertical-align: middle;
  padding: 0.6rem 0.85rem;
  transition: background-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
    border-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

.match-table :deep(.table tbody td:first-child) {
  border-left: 1px solid var(--xbvr-border, #e3e6ec);
  border-radius: var(--xbvr-radius, 12px) 0 0 var(--xbvr-radius, 12px);
}

.match-table :deep(.table tbody td:last-child) {
  border-right: 1px solid var(--xbvr-border, #e3e6ec);
  border-radius: 0 var(--xbvr-radius, 12px) var(--xbvr-radius, 12px) 0;
}

.match-table :deep(.table tbody tr:hover td),
.match-table :deep(.table tbody tr:hover td:first-child),
.match-table :deep(.table tbody tr:hover td:last-child) {
  background: var(--xbvr-hover-bg, #fafbfd);
  border-color: var(--xbvr-border-strong, #cdd2dc);
}

/* cell content */
.cover-thumb {
  border-radius: var(--xbvr-radius-sm, 8px);
  overflow: hidden;
}

.site-link {
  font-weight: 600;
}

.site-tags {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.3rem;
  margin-top: 0.3rem;
}

.site-tags :deep(.tag) {
  border-radius: 999px;
  margin: 0;
}

.tag-icon {
  margin-right: 0.2em;
}

.scene-title {
  font-weight: 600;
  color: var(--xbvr-text, #1c2333);
  overflow-wrap: anywhere;
}

.cast-tags :deep(.tag) {
  border-radius: 999px;
}

.cell-meta {
  font-size: 0.85rem;
  color: var(--xbvr-text-muted, #64708a);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.row-actions {
  display: flex;
  justify-content: flex-end;
}

.row-actions .button {
  margin: 0;
}

/* prev/next — circular glass controls over the overlay */
.scene-nav {
  cursor: pointer;
  position: fixed;
  top: 50%;
  transform: translateY(-50%);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 46px;
  height: 46px;
  padding: 0;
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: 999px;
  background: var(--xbvr-chip-bg, rgba(20, 24, 36, 0.55));
  color: #fff;
  font-size: 20px;
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  user-select: none;
  -webkit-user-select: none;
  transition: background-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
    transform var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

.scene-nav:hover {
  background: rgba(20, 24, 36, 0.8);
}

.scene-nav.prev {
  left: 14px;
}

.scene-nav.prev:hover {
  transform: translateY(-50%) translateX(-2px);
}

.scene-nav.next {
  right: 14px;
}

.scene-nav.next:hover {
  transform: translateY(-50%) translateX(2px);
}
</style>
