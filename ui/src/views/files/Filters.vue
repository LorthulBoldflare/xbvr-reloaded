<template>
  <div class="filters-panel files-filters">
    <div class="filters-grid">
      <div class="filter-field">
        <label class="filter-label">{{$t("State")}}</label>
        <b-field>
          <b-radio-button v-model="fileState" native-value="all">
            <span>{{$t("All")}}</span>
          </b-radio-button>
          <b-radio-button v-model="fileState" native-value="matched">
            <span>{{$t("Matched")}}</span>
          </b-radio-button>
          <b-radio-button v-model="fileState" native-value="unmatched">
            <span>{{$t("Unmatched")}}</span>
          </b-radio-button>
        </b-field>
      </div>

      <div class="filter-field">
        <label class="filter-label" for="filter-filename">{{$t("Filename")}}</label>
        <div class="field has-addons">
          <div class="control is-expanded">
            <b-input id="filter-filename" v-model="fileName"></b-input>
          </div>
          <div class="control">
            <button type="button" class="button is-light" @click="clearFilename" :aria-label="$t('Clear filename')">
              <b-icon pack="fas" icon="times" size="is-small"></b-icon>
            </button>
          </div>
        </div>
      </div>

      <div class="filter-field">
        <label class="filter-label">{{$t("Created between")}}</label>
        <div class="field has-addons">
          <div class="control is-expanded">
            <b-datepicker v-model="fileCreation" editable range>
              <div class="buttons">
                <b-button size="is-small" @click="setRange(subDays(new Date(), 7), new Date())">
                  <span>{{$t("Last 7 days")}}</span>
                </b-button>
                <b-button size="is-small" @click="setRange(subDays(new Date(), 14), new Date())">
                  <span>{{$t("Last 14 days")}}</span>
                </b-button>
                <b-button size="is-small" @click="setRange(subDays(new Date(), 30), new Date())">
                  <span>{{$t("Last 30 days")}}</span>
                </b-button>
              </div>
            </b-datepicker>
          </div>
          <div class="control">
            <button type="button" class="button is-light" @click="clearRange" :aria-label="$t('Clear date range')">
              <b-icon pack="fas" icon="times" size="is-small"></b-icon>
            </button>
          </div>
        </div>
      </div>

      <div class="filter-field">
        <label class="filter-label">{{$t("Resolution")}}</label>
        <b-dropdown v-model="fileResolutions" multiple hoverable aria-role="list" class="filter-dropdown">
            <button class="button" type="button" slot="trigger">
                <span>{{$t("Selected")}} ({{fileResolutions.length}})</span>
                <b-icon icon="menu-down"></b-icon>
            </button>
            <b-dropdown-item value="below4k" aria-role="listitem">
                <span>{{$t("Below 4K")}}</span>
            </b-dropdown-item>
            <b-dropdown-item value="4k" aria-role="listitem">
                <span>4K</span>
            </b-dropdown-item>
            <b-dropdown-item value="5k" aria-role="listitem">
                <span>5K</span>
            </b-dropdown-item>
            <b-dropdown-item value="6k" aria-role="listitem">
                <span>6K</span>
            </b-dropdown-item>
            <b-dropdown-item value="above6k" aria-role="listitem">
                <span>{{$t("Above 6K")}}</span>
            </b-dropdown-item>
        </b-dropdown>
      </div>

      <div class="filter-field">
        <label class="filter-label">{{$t("Bitrate")}}</label>
        <b-dropdown v-model="fileBitrates" multiple hoverable aria-role="list" class="filter-dropdown">
            <button class="button" type="button" slot="trigger">
                <span>{{$t("Selected")}} ({{fileBitrates.length}})</span>
                <b-icon icon="menu-down"></b-icon>
            </button>
            <b-dropdown-item value="low" aria-role="listitem">
                <span>{{$t("Low (below 15 Mbps)")}}</span>
            </b-dropdown-item>
            <b-dropdown-item value="medium" aria-role="listitem">
                <span>{{$t("Medium (15 to 24 Mbps)")}}</span>
            </b-dropdown-item>
            <b-dropdown-item value="high" aria-role="listitem">
                <span>{{$t("High (25 to 35 Mbps)")}}</span>
            </b-dropdown-item>
            <b-dropdown-item value="ultra" aria-role="listitem">
                <span>{{$t("Ultra (above 35 Mbps)")}}</span>
            </b-dropdown-item>
        </b-dropdown>
      </div>

      <div class="filter-field">
        <label class="filter-label">{{$t("Framerate")}}</label>
        <b-dropdown v-model="fileFramerates" multiple hoverable aria-role="list" class="filter-dropdown">
            <button class="button" type="button" slot="trigger">
                <span>{{$t("Selected")}} ({{fileFramerates.length}})</span>
                <b-icon icon="menu-down"></b-icon>
            </button>
            <b-dropdown-item value="30fps" aria-role="listitem">
                <span>30</span>
            </b-dropdown-item>
            <b-dropdown-item value="60fps" aria-role="listitem">
                <span>60</span>
            </b-dropdown-item>
            <b-dropdown-item value="other" aria-role="listitem">
                <span>{{$t("Other")}}</span>
            </b-dropdown-item>
        </b-dropdown>
      </div>
    </div>
  </div>
</template>

<script>
import { subDays } from 'date-fns'

export default {
  name: 'Filters',
  methods: {
    clearFilename () {
      this.fileName = ''
    },
    clearRange () {
      this.fileCreation = []
    },
    setRange (start, end) {
      this.fileCreation = [start, end]
    },
    subDays
  },
  computed: {
    fileName: {
      get () {
        return this.$store.state.files.filters.filename
      },
      set (value) {
        this.$store.state.files.filters.filename = value
        if (value.length > 3 || value.length == 0) {
          this.$store.dispatch('files/load')
        }
      }
    },
    fileBitrates: {
      get () {
        return this.$store.state.files.filters.bitrates
      },
      set (values) {
        this.$store.state.files.filters.bitrates = values
        this.$store.dispatch('files/load')
      }
    },
    fileFramerates: {
      get () {
        return this.$store.state.files.filters.framerates
      },
      set (values) {
        this.$store.state.files.filters.framerates = values
        this.$store.dispatch('files/load')
      }
    },
    fileResolutions: {
      get () {
        return this.$store.state.files.filters.resolutions
      },
      set (values) {
        this.$store.state.files.filters.resolutions = values
        this.$store.dispatch('files/load')
      }
    },
    fileState: {
      get () {
        return this.$store.state.files.filters.state
      },
      set (value) {
        this.$store.state.files.filters.state = value
        this.$store.dispatch('files/load')
      }
    },
    fileCreation: {
      get () {
        return this.$store.state.files.filters.createdDate
      },
      set (value) {
        this.$store.state.files.filters.createdDate = value
        this.$store.dispatch('files/load')
      }
    }
  }
}
</script>

<style lang="less" scoped>
/* ------------------------------------------------------------------
   Files filter bar — same design language as scenes/Filters.vue,
   laid out as a horizontal panel instead of a sidebar
   ------------------------------------------------------------------ */

.files-filters {
  background: var(--xbvr-surface, #ffffff);
  border: 1px solid var(--xbvr-border, #e3e6ec);
  border-radius: var(--xbvr-radius-lg, 16px);
  box-shadow: var(--xbvr-shadow-sm, 0 1px 2px rgba(16, 24, 40, 0.05));
  padding: 1rem 1.25rem;
  margin: 0.75rem 0 1.25rem;
}

.filters-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(210px, 1fr));
  gap: 0.85rem 1.25rem;
  align-items: end;
}

.filter-field {
  min-width: 0;
}

.filter-label {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--xbvr-text-muted, #64708a);
  margin-bottom: 0.3rem;
  cursor: pointer;
}

.filter-field .field,
.filter-field :deep(.field) {
  margin-bottom: 0;
}

/* full-width multi-select dropdown triggers */
.filter-dropdown,
.filter-dropdown :deep(.dropdown-trigger),
.filter-dropdown :deep(.dropdown-trigger .button) {
  width: 100%;
}

.filter-dropdown :deep(.dropdown-trigger .button) {
  justify-content: space-between;
}
</style>
