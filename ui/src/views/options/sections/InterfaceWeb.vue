<template>
  <div>
    <b-loading :is-full-page="false" :active.sync="isLoading" />
    <div class="content">
      <header class="options-page-head">
        <h1 class="options-title">{{ $t('Web UI') }}</h1>
        <p class="options-desc">{{ $t('Tune what the web interface shows and how scene cards look.') }}</p>
      </header>
      <div class="columns">
        <div class="column">
          <section class="settings-card">
            <div class="settings-card-title">
              <b-icon pack="mdi" icon="sort" size="is-small"/>
              {{ $t('General') }}
            </div>
            <b-field label="Tag Sort">
              <div class="block">
                <b-radio v-model="tagSort" name="tagSort" native-value="by-tag-count">
                  By Tag Count
                </b-radio>
                <b-radio v-model="tagSort" name="tagSort" native-value="alphabetically">
                  Alphabetically
                </b-radio>
              </div>
            </b-field>

            <b-field label="Automatically Check for Updates">
              <b-switch v-model="updateCheck">
                Enabled
              </b-switch>
            </b-field>
          </section>

          <section class="settings-card">
            <div class="settings-card-title">
              <b-icon pack="mdi" icon="button-cursor" size="is-small"/>
              {{ $t('Buttons in Scene List') }}
            </div>
            <div class="switch-grid">
            <b-field>
              <b-switch v-model="sceneHidden" type="is-danger">
                show Toggle Hidden Status button
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="sceneWatchlist" type="is-default">
                show Add/Remove from Watchlist button
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="sceneFavourite" type="is-danger">
                show Add/Remove from Favourites button
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="sceneWishlist" type="is-info">
                show Add/Remove from Wishlist button
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="sceneTrailerlist" type="is-default">
                show Add/Remove from Trailer list button
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="sceneWatched" type="is-dark">
                show Toggle Watched Status button
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="sceneEdit" type="is-dark">
                show Edit Scene button
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="sceneDuration" type="is-dark">
                show Duration button
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="sceneCuepoint" type="is-dark">
                show Cuepoints button
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="hspFile" type="is-dark">
                show Hsp File button
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="subtitlesFile" type="is-dark">
                show subtitles File button
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="ScriptHeatmap" type="is-dark">
                show Script Heatmap
              </b-switch>
            </b-field>
            <b-field v-if="ScriptHeatmap">
              <b-switch v-model="AllHeatmaps" type="is-dark">
                show All Heatmaps
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="openInNewWindow" type="is-dark">
                show Open Tag in New Window
              </b-switch>
            </b-field>
            </div>
          </section>

          <section class="settings-card">
            <div class="settings-card-title">
              <b-icon pack="mdi" icon="palette-outline" size="is-small"/>
              {{ $t('Appearance') }}
            </div>
            <b-field label="Opacity of unavailable scenes">
              <div class="columns">
                <div class="column is-two-thirds">
                  <b-slider :min="0" :max="100" :step="10" :tooltip="false" v-model="isAvailOpacity" opacity:isAvailOpacity></b-slider>
                </div>
              </div>
            </b-field>

            <b-field label="Scene card aspect ratio">
              <b-select placeholder="Select aspect ratio" v-model="sceneCardAspectRatio">
                <option>1:1</option>
                <option>3:2</option>
                <option>16:9</option>
              </b-select>
            </b-field>
            <b-field>
              <b-switch v-model="sceneCardScaleToFit" type="is-dark">
                Scale cover to fit
              </b-switch>
            </b-field>

            <b-field label="Actor card aspect ratio">
              <b-select placeholder="Select aspect ratio" v-model="actorCardAspectRatio">
                <option>1:1</option>
                <option>2:3</option>
                <option>9:16</option>
              </b-select>
            </b-field>
            <b-field>
              <b-switch v-model="actorCardScaleToFit" type="is-dark">
                Scale cover to fit
              </b-switch>
            </b-field>
          </section>

          <div class="save-row">
            <b-button type="is-primary" @click="save">Save</b-button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'InterfaceWeb',
  mounted () {
    this.$store.dispatch('optionsWeb/load')
  },
  methods: {
    save () {
      this.$store.dispatch('optionsWeb/save')
    }
  },
  computed: {
    tagSort: {
      get () {
        return this.$store.state.optionsWeb.web.tagSort
      },
      set (value) {
        this.$store.state.optionsWeb.web.tagSort = value
      }
    },
    sceneHidden: {
      get () {
        return this.$store.state.optionsWeb.web.sceneHidden
      },
      set (value) {
        this.$store.state.optionsWeb.web.sceneHidden = value
      }
    },
    sceneWatchlist: {
      get () {
        return this.$store.state.optionsWeb.web.sceneWatchlist
      },
      set (value) {
        this.$store.state.optionsWeb.web.sceneWatchlist = value
      }
    },
    sceneFavourite: {
      get () {
        return this.$store.state.optionsWeb.web.sceneFavourite
      },
      set (value) {
        this.$store.state.optionsWeb.web.sceneFavourite = value
      }
    },
    sceneWishlist: {
      get () {
        return this.$store.state.optionsWeb.web.sceneWishlist
      },
      set (value) {
        this.$store.state.optionsWeb.web.sceneWishlist = value
      }
    },
    sceneTrailerlist: {
      get () {
        return this.$store.state.optionsWeb.web.sceneTrailerlist
      },
      set (value) {
        this.$store.state.optionsWeb.web.sceneTrailerlist = value
      }
    },
    sceneWatched: {
      get () {
        return this.$store.state.optionsWeb.web.sceneWatched
      },
      set (value) {
        this.$store.state.optionsWeb.web.sceneWatched = value
      }
    },
    sceneEdit: {
      get () {
        return this.$store.state.optionsWeb.web.sceneEdit
      },
      set (value) {
        this.$store.state.optionsWeb.web.sceneEdit = value
      }
    },
    ScriptHeatmap: {
      get () {
        return this.$store.state.optionsWeb.web.showScriptHeatmap
      },
      set (value) {
        this.$store.state.optionsWeb.web.showScriptHeatmap = value
      }
    },
    AllHeatmaps: {
      get () {
        return this.$store.state.optionsWeb.web.showAllHeatmaps
      },
      set (value) {
        this.$store.state.optionsWeb.web.showAllHeatmaps = value
      }
    },
    updateCheck: {
      get () {
        return this.$store.state.optionsWeb.web.updateCheck
      },
      set (value) {
        this.$store.state.optionsWeb.web.updateCheck = value
      }
    },
    sceneDuration: {
      get () {
        return this.$store.state.optionsWeb.web.sceneDuration
      },
      set (value) {
        this.$store.state.optionsWeb.web.sceneDuration = value
      }
    },
    sceneCuepoint: {
      get () {
        return this.$store.state.optionsWeb.web.sceneCuepoint
      },
      set (value) {
        this.$store.state.optionsWeb.web.sceneCuepoint = value
      }
    },
    hspFile: {
      get () {
        return this.$store.state.optionsWeb.web.showHspFile
      },
      set (value) {
        this.$store.state.optionsWeb.web.showHspFile = value
      }
    },
    subtitlesFile: {
      get () {
        return this.$store.state.optionsWeb.web.showSubtitlesFile
      },
      set (value) {
        this.$store.state.optionsWeb.web.showSubtitlesFile = value
      }
    },
    openInNewWindow: {
      get () {
        return this.$store.state.optionsWeb.web.showOpenInNewWindow
      },
      set (value) {
        this.$store.state.optionsWeb.web.showOpenInNewWindow = value
      }
    },
    isAvailOpacity: {
      get () {
        if  (this.$store.state.optionsWeb.web.isAvailOpacity == undefined) {
          return 40
        }
        return this.$store.state.optionsWeb.web.isAvailOpacity
      },
      set (value) {
        this.$store.state.optionsWeb.web.isAvailOpacity = value
      }
    },
    sceneCardAspectRatio: {
      get () {
        return this.$store.state.optionsWeb.web.sceneCardAspectRatio
      },
      set (value) {
        this.$store.state.optionsWeb.web.sceneCardAspectRatio = value
      }
    },
    sceneCardScaleToFit: {
      get () {
        return this.$store.state.optionsWeb.web.sceneCardScaleToFit
      },
      set (value) {
        this.$store.state.optionsWeb.web.sceneCardScaleToFit = value
      }
    },
    actorCardAspectRatio: {
      get () {
        return this.$store.state.optionsWeb.web.actorCardAspectRatio
      },
      set (value) {
        this.$store.state.optionsWeb.web.actorCardAspectRatio = value
      }
    },
    actorCardScaleToFit: {
      get () {
        return this.$store.state.optionsWeb.web.actorCardScaleToFit
      },
      set (value) {
        this.$store.state.optionsWeb.web.actorCardScaleToFit = value
      }
    },
    isLoading: function () {
      return this.$store.state.optionsWeb.loading
    }
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
  padding: 1.25rem;
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
  margin-bottom: 0.9rem;
}

.settings-card-title .icon {
  color: var(--xbvr-text-faint, #7d88a1);
}

/* switches are full-row toggles inside cards */
.settings-card :deep(.field) {
  margin-bottom: 0.45rem;
}

.settings-card :deep(.field:last-child) {
  margin-bottom: 0;
}

/* long toggle lists flow into two columns instead of one tall stack */
.switch-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 0 1.5rem;
}

.switch-grid :deep(.field) {
  margin-bottom: 0.45rem;
}

.save-row {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-top: 1rem;
}
</style>
