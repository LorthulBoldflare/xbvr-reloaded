<template>
  <div class="modal is-active">
    <GlobalEvents
      :filter="e => !['INPUT', 'TEXTAREA'].includes(e.target.tagName)"
      @keyup.esc="close"
    />
    <div class="modal-background"></div>
    <div class="modal-card match-params-card" v-if="site != null">
      <header class="modal-card-head match-params-head">
        <div class="match-params-heading">
          <p class="modal-card-title match-params-title">{{ $t("Matching parameters") }}</p>
          <div class="meta-chips">
            <span class="meta-chip">
              <b-icon pack="mdi" icon="web" size="is-small"/>
              {{ site.name }}
            </span>
          </div>
        </div>
        <button class="delete" @click="close" aria-label="close"></button>
      </header>
      <section class="modal-card-body match-params-body" v-if="params != null">

        <section class="mp-group">
          <h3 class="mp-group-title">
            <b-icon pack="mdi" icon="filter-variant" size="is-small"/>
            {{ $t('Selection Criteria') }}
          </h3>
          <div class="mp-grid">
            <b-tooltip :label="$t('Days to wait after the release date, before linking. Useful where the main site releases after SLR/VRPorn/POVR, eg LethalHardcore')"
              :delay="500" type="is-primary" multilined size="is-large" position="is-bottom">
              <b-field :label="$t('Delay linking(days)')">
                <b-numberinput v-model="params.delay_linking"></b-numberinput>
              </b-field>
            </b-tooltip>
            <b-tooltip :label="$t('Number of days to keep re-linking scenes after the release date')" :delay="500" type="is-primary" multilined>
              <b-field :label="$t('Keep Re-linking(days)')">
                <b-numberinput v-model="params.reprocess_links"></b-numberinput>
              </b-field>
            </b-tooltip>
            <b-tooltip :label="$t('Do not link scenes prior to the specified date.  The quality of metadata of older scenes is often poor and causes mismatches')"
              :delay="500" type="is-primary" multilined size="is-large" position="is-bottom">
              <b-field :label="$t('Ignore Scenes Released Prior To')">
                <b-datepicker v-model="ignoreReleasedBefore" :icon-right="ignoreReleasedBefore ? 'close-circle' : ''" icon-right-clickable @icon-right-click="clearDate">
                  <b-button
                      label="Today"
                      type="is-primary"
                      icon-left="calendar-today"
                      @click="ignoreReleasedBefore = new Date()" />

                  <b-button
                      label="Clear"
                      type="is-danger"
                      icon-left="close"
                      outlined
                      @click="ignoreReleasedBefore = null" />
                </b-datepicker>
              </b-field>
            </b-tooltip>
          </div>
        </section>

        <section class="mp-group">
          <h3 class="mp-group-title">
            <b-icon pack="mdi" icon="calendar-search" size="is-small"/>
            {{ $t('Release Date Searching') }}
          </h3>
          <div class="mp-grid">
            <b-field :label="$t('Match Type')">
              <b-select required v-model="params.released_match_type" expanded>
                <option value="should">Should match</option>
                <option value="must">Must</option>
                <option value="do not">Do not</option>
              </b-select>
            </b-field>
            <b-tooltip :label="$t('Weighting of Title matchs (vs Duration=1)')"
              :delay="500" type="is-primary" multilined size="is-large" position="is-bottom">
              <b-field :label="$t('Boost Value')">
                <b-numberinput v-model="params.boost_released" step=0.05></b-numberinput>
              </b-field>
            </b-tooltip>
            <b-tooltip :label="$t('The number of days prior to the release date to match, eg if the scene release date is 23/05/2023 and the days prior is 3, it will search >= 20/05/2023. If days prior and after are 0, the range is not used')"
              :delay="500" type="is-primary" multilined size="is-large" position="is-bottom">
              <b-field :label="$t('Days Prior')">
                <b-numberinput v-model="params.released_prior"></b-numberinput>
              </b-field>
            </b-tooltip>
            <b-tooltip :label="$t('The number of days after the release date to match, eg if the scene release date is 23/05/2023 and the days after is 3, it will search <= 23/05/2023. Usually set to 0. If days prior and after are 0, the range is not used')"
              :delay="500" type="is-primary" multilined size="is-large" position="is-bottom">
              <b-field :label="$t('Days After')">
                <b-numberinput v-model="params.released_after"></b-numberinput>
              </b-field>
            </b-tooltip>
          </div>
        </section>

        <section class="mp-group">
          <h3 class="mp-group-title">
            <b-icon pack="mdi" icon="format-title" size="is-small"/>
            {{ $t('Title Searching') }}
          </h3>
          <div class="mp-grid">
            <b-field :label="$t('Exact Match Boost Value')">
              <b-numberinput v-model="params.boost_title" step=0.05></b-numberinput>
            </b-field>
            <b-field :label="$t('Individual Word Match Boost Value')">
              <b-numberinput v-model="params.boost_title_any_words" step=0.05></b-numberinput>
            </b-field>
          </div>
        </section>

        <section class="mp-group">
          <h3 class="mp-group-title">
            <b-icon pack="mdi" icon="timer-outline" size="is-small"/>
            {{ $t('Duration Searching') }}
          </h3>
          <div class="mp-grid">
            <b-field :label="$t('Match Type')">
              <b-select required v-model="params.duration_match_type" expanded>
                <option value="should">Should match</option>
                <option value="must">Must</option>
                <option value="do not">Do not</option>
              </b-select>
            </b-field>
            <b-field :label="$t('Minimum Duration')">
              <b-numberinput v-model="params.duration_min"></b-numberinput>
            </b-field>
            <b-field :label="$t('Lower Search Range')">
              <b-numberinput v-model="params.duration_range_less"></b-numberinput>
            </b-field>
            <b-field :label="$t('Upper Search Range')">
              <b-numberinput v-model="params.duration_range_more"></b-numberinput>
            </b-field>
          </div>
        </section>

        <section class="mp-group">
          <h3 class="mp-group-title">
            <b-icon pack="mdi" icon="account-star-outline" size="is-small"/>
            {{ $t('Cast Searching') }}
          </h3>
          <div class="mp-grid">
            <b-field :label="$t('Match Type')">
              <b-select required v-model="params.cast_match_type" expanded>
                <option value="should">Should match</option>
                <option value="must">Must</option>
                <option value="do not">Do not</option>
              </b-select>
            </b-field>
            <b-field :label="$t('Exact Match Boost Value')">
              <b-numberinput v-model="params.boost_cast" step=0.05></b-numberinput>
            </b-field>
          </div>
        </section>

        <section class="mp-group">
          <h3 class="mp-group-title">
            <b-icon pack="mdi" icon="text-box-search-outline" size="is-small"/>
            {{ $t('Description Searching') }}
          </h3>
          <div class="mp-grid">
            <b-field :label="$t('Match Type')">
              <b-select required v-model="params.desc_match_type" expanded>
                <option value="should">Should match</option>
                <option value="must">Must</option>
                <option value="do not">Do not</option>
              </b-select>
            </b-field>
            <b-field :label="$t('Exact Match Boost Value')">
              <b-numberinput v-model="params.boost_description" step=0.05></b-numberinput>
            </b-field>
          </div>
        </section>
      </section>
      <footer class="modal-card-foot match-params-foot">
        <div class="action-row">
          <b-button type="is-primary" @click="saveSettings">Save settings</b-button>
        </div>
      </footer>
    </div>
  </div>
</template>

<script>
import api from '../../../api'
import { format, parseISO } from 'date-fns'
import prettyBytes from 'pretty-bytes'
import GlobalEvents from 'vue-global-events'

export default {
  name: 'SceneMatchParams',
  components: { GlobalEvents },
  data () {
    return {
      site: null,
      params: null,
      ignoreReleasedBefore: null,
      format,
      parseISO
    }
  },
  computed: {
  },
  mounted () {    
    this.initView()
  },
  methods: {
    initView () {
      this.site=this.$store.state.overlay.sceneMatchParams.site
      api.get('/options/site/match_params/' + this.site.id).json().then(data => {
        this.params = data
        this.ignoreReleasedBefore = new Date(this.params.ignore_released_before);
      })
    },
    close () {
      this.$store.commit('overlay/hideSceneMatchParams')
    },
    clearDate() {
      this.ignoreReleasedBefore = null
    },
    saveSettings() {      
      this.params.ignore_released_before=this.ignoreReleasedBefore
      api.post(`/options/site/save_match_params`, { json: { site: this.site.id, match_params: this.params } })
      if (this.ignoreReleasedBefore != null) {        
        const formattedDate = this.ignoreReleasedBefore.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric',});
        this.$buefy.dialog.confirm({
          title: 'Clear existing links',
          message: `Do you also wish to clear links from <strong>${formattedDate}</strong>`,
          type: 'is-info is-wide',
          hasIcon: true,
          onConfirm: () => {
            api.delete(`/extref/delete_extref_source_links/keep_manual`, { json: {external_source: 'alternate scene ' + this.site.id, delete_date: this.ignoreReleasedBefore} });
          }
        })
        
      }
    },
    prettyBytes
  }
}
</script>

<style scoped>
.match-params-card {
  width: min(760px, 94vw);
}

.match-params-head {
  align-items: flex-start;
}

.match-params-heading {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
  min-width: 0;
}

.match-params-title {
  font-weight: 800;
  letter-spacing: -0.01em;
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
}

.match-params-body {
  padding: 1.25rem;
}

.mp-group {
  border: 1px solid var(--xbvr-border, #e3e6ec);
  border-radius: var(--xbvr-radius, 12px);
  background: var(--xbvr-surface-sunken, #eef0f4);
  padding: 1rem 1.25rem 1.25rem;
  margin-bottom: 1.25rem;
}

.mp-group:last-child {
  margin-bottom: 0;
}

.mp-group-title {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--xbvr-text-muted, #64708a);
  margin-bottom: 0.9rem;
}

.mp-group-title .icon {
  color: var(--xbvr-text-faint, #7d88a1);
}

.mp-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 0.75rem 1rem;
  align-items: start;
}

/* b-tooltip wrappers participate as grid items */
.mp-grid :deep(.b-tooltip) {
  display: block;
  min-width: 0;
}

.mp-grid .field {
  margin-bottom: 0;
}

.mp-grid :deep(.field .field) {
  margin-bottom: 0;
}

.match-params-foot {
  padding: 0.85rem 1.25rem;
}

.action-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  justify-content: flex-end;
}
</style>
