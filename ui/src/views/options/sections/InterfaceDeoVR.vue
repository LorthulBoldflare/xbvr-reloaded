<template>
  <div>
    <b-loading :is-full-page="false" :active.sync="isLoading"></b-loading>
    <header class="options-page-head">
      <h1 class="options-title">{{ $t('Players') }}</h1>
      <p class="options-desc">{{ $t('Configure the DeoVR and Heresphere player interfaces.') }}</p>
    </header>
    <b-tabs v-model="activeTab" size="medium" type="is-boxed" id="playertab" class="options-tabs">
      <b-tab-item label="Shared Settings"/>
      <b-tab-item label="DeoVR"/>
      <b-tab-item label="Heresphere"/>
    </b-tabs>
    <div class="content" v-if="activeTab == 0">
      <div class="columns">
        <div class="column">
          <section class="settings-card">
            <div class="settings-card-title">
              <b-icon pack="mdi" icon="play-circle-outline" size="is-small"/>
              {{ $t('Shared Player Options') }}
            </div>
            <b-field label="Player integration">
              <b-switch v-model="enabled">
                Enabled
              </b-switch>
            </b-field>
            <div v-if="enabled">
              <hr class="card-divider"/>
              <b-field label="Authentication">
                <b-switch v-model="authEnabled">
                  Enabled
                </b-switch>
              </b-field>
              <div class="block">
                <b-field grouped>
                  <b-field label="Username" class="cred-field">
                    <b-input v-model="username" :disabled="authEnabled === false"></b-input>
                  </b-field>
                  <b-field label="Password" class="cred-field">
                    <b-input v-model="password" :disabled="authEnabled === false" type="password"></b-input>
                  </b-field>
                </b-field>
              </div>
              <hr class="card-divider"/>
              <div class="block">
                <b-field label="Funscript heatmaps">
                  <b-switch v-model="renderHeatmaps">
                    Enabled
                  </b-switch>
                </b-field>
                <p class="muted">
                  If you are using funscripts, you can add a heatmap to the thumbnails of scripted scenes in the Player interface.
                </p>
              </div>
              <hr class="card-divider"/>
              <div class="block">
                <b-field label="Watch time tracking">
                  <b-switch v-model="watchTimeTrackingEnabled">
                    Enabled
                  </b-switch>
                </b-field>
              </div>
              <hr class="card-divider"/>
              <div class="block">
                <b-tooltip label="Specify fields if you wish to control the sequence of the scene's video files" multilined :delay="750" >
                  <b-field label="Video File Sorting">
                    <b-input v-model="videoSequence" disabled></b-input>
                  </b-field>
                  <b-field>
                    <b-button label="Add Field" @click="addVideoField('video')" />
                    <b-button label="Clear Fields" @click="videoSequence=''" />                  
                    <b-dropdown v-model="selectedVideoField" aria-role="list" :scrollable=true max-height="200">
                        <template #trigger>
                            <b-button :label="selectedVideoField" icon-right="menu-down"/>
                        </template>
                        <b-dropdown-item aria-role="listitem" value='Filename'>Filename</b-dropdown-item>
                        <b-dropdown-item aria-role="listitem" value='Added'>Added</b-dropdown-item>
                        <b-dropdown-item aria-role="listitem" value='Updated'>Updated</b-dropdown-item>
                        <b-dropdown-item aria-role="listitem" value='Resolution'>Resolution</b-dropdown-item>
                        <b-dropdown-item aria-role="listitem" value='Size'>Size</b-dropdown-item>                      
                        <b-dropdown-item aria-role="listitem" value='Bitrate'>Bitrate</b-dropdown-item>
                        <b-dropdown-item aria-role="listitem" value='Frame Rate'>Frame Rate</b-dropdown-item>
                        <b-dropdown-item aria-role="listitem" value='Codec'>Codec</b-dropdown-item>
                        <b-dropdown-item aria-role="listitem" value='Duration'>Duration</b-dropdown-item>
                    </b-dropdown>
                    <b-dropdown v-model="selectedVideoSequence" aria-role="list">
                        <template #trigger>
                            <b-button :label="selectedVideoSequence" icon-right="menu-down" />
                        </template>
                        <b-dropdown-item aria-role="listitem" value='Ascending'>Ascending</b-dropdown-item>
                        <b-dropdown-item aria-role="listitem" value='Descending'>Descending</b-dropdown-item>
                    </b-dropdown>
                  </b-field>
                </b-tooltip>
                <b-tooltip label="Specify fields if you wish to control the sequence of the scene's script files" multilined :delay="750" >
                  <b-field label="Script File Sorting">
                    <b-input v-model="scriptSequence" disabled></b-input>
                  </b-field>
                  <b-field>
                    <b-button label="Add Field" @click="addVideoField('script')" />
                    <b-button label="Clear Fields" @click="scriptSequence=''" />                  
                    <b-dropdown v-model="selectedScriptField" aria-role="list">
                        <template #trigger>
                            <b-button :label="selectedScriptField" icon-right="menu-down" />
                        </template>
                        <b-dropdown-item aria-role="listitem" value='Filename'>Filename</b-dropdown-item>
                        <b-dropdown-item aria-role="listitem" value='Added'>Added</b-dropdown-item>
                        <b-dropdown-item aria-role="listitem" value='Updated'>Updated</b-dropdown-item>
                        <b-dropdown-item aria-role="listitem" value='Selected'>Selected</b-dropdown-item>                      
                    </b-dropdown>
                    <b-dropdown v-model="selectedScriptSequence" aria-role="list">
                        <template #trigger>
                            <b-button :label="selectedScriptSequence" icon-right="menu-down" />
                        </template>
                        <b-dropdown-item aria-role="listitem" value='Ascending'>Ascending</b-dropdown-item>
                        <b-dropdown-item aria-role="listitem" value='Descending'>Descending</b-dropdown-item>
                    </b-dropdown>
                  </b-field>
                </b-tooltip>
                <b-tooltip label="Specify fields if you wish to control the sequence of the scene's subtitle files" multilined :delay="750" >
                  <b-field label="Subtitle File Sorting">
                    <b-input v-model="subtitleSequence" disabled></b-input>
                  </b-field>
                  <b-field>
                    <b-button label="Add Field" @click="addVideoField('subtitle')" />
                    <b-button label="Clear Fields" @click="subtitleSequence=''" />
                    <b-dropdown v-model="selectedSubtitleField" aria-role="list">
                        <template #trigger>
                            <b-button :label="selectedSubtitleField" icon-right="menu-down" />
                        </template>
                        <b-dropdown-item aria-role="listitem" value='Filename'>Filename</b-dropdown-item>
                        <b-dropdown-item aria-role="listitem" value='Added'>Added</b-dropdown-item>
                        <b-dropdown-item aria-role="listitem" value='Updated'>Updated</b-dropdown-item>
                        <b-dropdown-item aria-role="listitem" value='Selected'>Selected</b-dropdown-item>
                    </b-dropdown>
                    <b-dropdown v-model="selectedSubtitleSequence" aria-role="list">
                        <template #trigger>
                            <b-button :label="selectedSubtitleSequence" icon-right="menu-down" />
                        </template>
                        <b-dropdown-item aria-role="listitem" value='Ascending'>Ascending</b-dropdown-item>
                        <b-dropdown-item aria-role="listitem" value='Descending'>Descending</b-dropdown-item>
                    </b-dropdown>
                  </b-field>
                </b-tooltip>
              </div>
            </div>
          </section>
        </div>
        <div class="column content">
          <div class="settings-card">
            <div class="settings-card-title">
              <b-icon pack="mdi" icon="link-variant" size="is-small"/>
              {{ $t('Player URLs') }}
            </div>
            <p class="muted">
              {{ $t("Player interface is available at following URLs:") }}
            </p>
            <div class="endpoint-list">
              <code v-for="(addr, idx) in boundIp" :key="'ip' + idx">{{ addr }}</code>
            </div>
            <hr class="card-divider"/>
            <p class="muted">
              NOTE: make sure DeoVR is using <strong>http://</strong> not <strong>https://</strong>.<br/>
              To toggle used protocol, click on it in DeoVR's URL bar.
            </p>
          </div>
        </div>
      </div>
    </div>
    <div class="content" v-if="activeTab == 1">
      <div class="settings-card">
        <div class="settings-card-title">
          <b-icon pack="mdi" icon="virtual-reality" size="is-small"/>
          {{ $t('DeoVR interface') }}
        </div>
        <div class="block">
          <b-field label="Remote mode">
            <b-switch v-model="remoteEnabled" :disabled="watchTimeTrackingEnabled === false">
              Enabled
            </b-switch>
          </b-field>
          <p class="muted">
            Requires: Watch time tracking
          </p>
          <p class="muted">
            To use remote mode, which enables more precise watch time tracking, you need to turn it on in DeoVR
            settings too - see <a href="https://deovr.com/doc#remote-control" target="_blank" rel="noreferrer">
            instructions in DeoVR documentation</a>.
          </p>
        </div>
      </div>
    </div>
    <div class="content" v-if="activeTab == 2">
      <div class="settings-card">
        <div class="settings-card-title">
          <b-icon pack="mdi" icon="sphere" size="is-small"/>
          {{ $t('Heresphere interface') }}
        </div>
          <b-tooltip
            label="WANRING: File deletion from Heresphere is PERMANENT! ALL files associated with a scene will be deleted."
            size="is-large" type="is-danger" multilined :delay="250" >
            <b-field label="Allow File Deletion">
              <b-switch v-model="allowFileDeletions">
                Enabled
              </b-switch>
            </b-field>
          </b-tooltip>
          <b-field label="Allow Ratings Updates">
            <b-switch v-model="allowRatingUpdates">
              Enabled
            </b-switch>
          </b-field>
          <b-field label="Allow Favorite Updates">
            <b-switch v-model="allowFavouriteUpdates">
              Enabled
            </b-switch>
          </b-field>
          <b-field label="Allow Tag Updates">
            <b-switch v-model="allowTagUpdates">
              Enabled
            </b-switch>
          </b-field>
          <b-field label="Allow Cuepoint Updates">
            <b-switch v-model="allowCuepointUpdates">
              Enabled
            </b-switch>
          </b-field>
          <b-tooltip
            label="Add or delete the Feature:watchlist tag to toggle the Watchlist flag in XBVR"
            size="is-large" type="is-primary" multilined :delay="250" >
            <b-field label="Allow Watchlist Updates">
              <b-switch v-model="allowWatchlistUpdates">
                Enabled
              </b-switch>
            </b-field>
          </b-tooltip>
          <b-field label="Allow Saving Hsp Files">
            <b-switch v-model="allowHspData">
              Enabled
            </b-switch>
          </b-field>
          <div class="columns">
            <div class="column is-one-thrid">
              <b-tooltip
                label="This option will split Cuepoints into multiple tracks, eg Standing-Doggy will split into 2 tracks in Heresphere"
                size="is-large" type="is-primary" multilined :delay="250" >
                <b-field label="Use Multi-Track Cuepoints">
                  <b-switch v-model="multiTrackCuepoints">
                    Enabled
                  </b-switch>
                </b-field>
              </b-tooltip>
            </div>
            <div class="column is-one-thrid">
              <b-tooltip
                label="This option will split Cuepoints matching the Actors Name into seperate tracks in Heresphere"
                size="is-large" type="is-primary" multilined :delay="250" >
                <b-field label="Use Multi-Track Cast Cuepoints">
                  <b-switch v-model="multiTrackCastCuepoints">
                    Enabled
                  </b-switch>
                </b-field>
              </b-tooltip>
            </div>
            <div class="column is-one-thrid"> 
              <b-tooltip
                label="This option controls whether you use wish to keep existing non-HSP Cuepoints when you sync cuepoints changes. Syncing changes to Cuepoints in HSP will be saved with extended fields"
                size="is-large" type="is-primary" multilined :delay="250" >
                <b-field label="Retain Non-HSP Cuepoints">
                  <b-switch v-model="retainNonHSPCuepoints">
                    Enabled
                  </b-switch>
                </b-field>
              </b-tooltip>
            </div>
          </div>
        </div>
    </div>
    <b-field class="save-row">
      <b-button type="is-primary" @click="save">Save and apply changes</b-button>
    </b-field>
  </div>
</template>

<script>
export default {
  name: 'InterfaceDeoVR',
  mounted () {
    this.$store.dispatch('optionsDeoVR/load')
  },
  data () {
    return {
      activeTab: 0,
      selectedVideoField: 'Filename',
      selectedVideoSequence: 'Ascending',
      selectedScriptField: 'Filename',
      selectedScriptSequence: 'Ascending',
      selectedSubtitleField: 'Filename',
      selectedSubtitleSequence: 'Ascending'
    }
  },
  methods: {
    save () {
      this.$store.dispatch('optionsDeoVR/save')
    },
    addIP (value) {
      const tmp = [...this.allowedIp]
      tmp.push(value)

      if (!this.hasDuplicates(tmp)) {
        this.allowedIp = tmp
      }
    },
    hasDuplicates (array) {
      return (new Set(array)).size !== array.length
    },
    addVideoField(type) {      
      let dbfield=''
      let field=this.selectedVideoField            
      if (type=='script') {
        field=this.selectedScriptField
      } else if (type=='subtitle') {
        field=this.selectedSubtitleField
      }

      switch (field) {
        case 'Added':
          dbfield = 'created_time'
          break
        case 'Updated':
          dbfield = 'updated_time'
          break
        case 'Resolution':
          dbfield = 'video_height'
          break
        case 'Bitrate':
          dbfield = 'video_bit_rate'
          break
        case 'Frame Rate':
          dbfield = 'video_avg_frame_rate_val'
          break
        case 'Codec':
          dbfield = "case when video_codec_name in ('hevc', 'h265') then 0 when video_codec_name='h264' then 1 else 2 end"
          break
        case 'Duration':
          dbfield = 'video_direction'
          break
        case 'Selected':
          dbfield = 'is_selected_script'
          break
        default:
          dbfield = field.toLowerCase()
      }

      if (type=='video') {
        if (this.selectedVideoSequence=='Ascending') {
          this.videoSequence=[this.videoSequence, dbfield ].filter(Boolean).join(',')
        } else {
          this.videoSequence=[this.videoSequence, dbfield+' desc' ].filter(Boolean).join(',')
        }
      } else if (type=='script') {
        if (this.selectedScriptSequence=='Ascending') {
          this.scriptSequence=[this.scriptSequence, dbfield ].filter(Boolean).join(',')
        } else {
          this.scriptSequence=[this.scriptSequence, dbfield+' desc' ].filter(Boolean).join(',')
        }
      } else {
        if (this.selectedSubtitleSequence=='Ascending') {
          this.subtitleSequence=[this.subtitleSequence, dbfield ].filter(Boolean).join(',')
        } else {
          this.subtitleSequence=[this.subtitleSequence, dbfield+' desc' ].filter(Boolean).join(',')
        }
      }
    }
  },
  computed: {
    enabled: {
      get () {
        return this.$store.state.optionsDeoVR.deovr.enabled
      },
      set (value) {
        this.$store.state.optionsDeoVR.deovr.enabled = value
      }
    },
    authEnabled: {
      get () {
        return this.$store.state.optionsDeoVR.deovr.auth_enabled
      },
      set (value) {
        this.$store.state.optionsDeoVR.deovr.auth_enabled = value
      }
    },
    renderHeatmaps: {
      get () {
        return this.$store.state.optionsDeoVR.deovr.render_heatmaps
      },
      set (value) {
        this.$store.state.optionsDeoVR.deovr.render_heatmaps = value
      }
    },
    watchTimeTrackingEnabled: {
      get () {
        return this.$store.state.optionsDeoVR.deovr.track_watch_time
      },
      set (value) {
        this.$store.state.optionsDeoVR.deovr.track_watch_time = value
      }
    },
    remoteEnabled: {
      get () {
        return this.$store.state.optionsDeoVR.deovr.remote_enabled
      },
      set (value) {
        this.$store.state.optionsDeoVR.deovr.remote_enabled = value
      }
    },
    username: {
      get () {
        return this.$store.state.optionsDeoVR.deovr.username
      },
      set (value) {
        this.$store.state.optionsDeoVR.deovr.username = value
      }
    },
    password: {
      get () {
        return this.$store.state.optionsDeoVR.deovr.password
      },
      set (value) {
        this.$store.state.optionsDeoVR.deovr.password = value
      }
    },
    boundIp: {
      get () {
        return this.$store.state.optionsDeoVR.deovr.boundIp
      }
    },        
    allowFileDeletions: {
      get () {
        return this.$store.state.optionsDeoVR.heresphere.allow_file_deletes
      },
      set (value) {
        this.$store.state.optionsDeoVR.heresphere.allow_file_deletes= value
      }
    },
    allowRatingUpdates: {
      get () {
        return this.$store.state.optionsDeoVR.heresphere.allow_rating_updates
      },
      set (value) {
        this.$store.state.optionsDeoVR.heresphere.allow_rating_updates = value
      }
    },
    allowFavouriteUpdates: {
      get () {
        return this.$store.state.optionsDeoVR.heresphere.allow_favorite_updates
      },
      set (value) {
        this.$store.state.optionsDeoVR.heresphere.allow_favorite_updates = value
      }
    },
    allowTagUpdates: {
      get () {
        return this.$store.state.optionsDeoVR.heresphere.allow_tag_updates
      },
      set (value) {
        this.$store.state.optionsDeoVR.heresphere.allow_tag_updates = value
      }
    },
    allowCuepointUpdates: {
      get () {
        return this.$store.state.optionsDeoVR.heresphere.allow_cuepoint_updates
      },
      set (value) {
        this.$store.state.optionsDeoVR.heresphere.allow_cuepoint_updates = value
      }
    },
    allowWatchlistUpdates: {
      get () {
        return this.$store.state.optionsDeoVR.heresphere.allow_watchlist_updates
      },
      set (value) {
        this.$store.state.optionsDeoVR.heresphere.allow_watchlist_updates = value
      }
    },
    allowHspData: {
      get () {
        return this.$store.state.optionsDeoVR.heresphere.allow_hsp_data
      },
      set (value) {
        this.$store.state.optionsDeoVR.heresphere.allow_hsp_data= value
      }
    },    
    multiTrackCuepoints: {
      get () {
        return this.$store.state.optionsDeoVR.heresphere.multitrack_cuepoints
      },
      set (value) {
        this.$store.state.optionsDeoVR.heresphere.multitrack_cuepoints = value
      }
    },
    videoSequence: {
      get () {
        return this.$store.state.optionsDeoVR.players.video_sort_seq
      },
      set (value) {
        this.$store.state.optionsDeoVR.players.video_sort_seq = value
      },
    },
    scriptSequence: {
      get () {
        return this.$store.state.optionsDeoVR.players.script_sort_seq
      },
      set (value) {
        this.$store.state.optionsDeoVR.players.script_sort_seq = value
      },
    },
    subtitleSequence: {
      get () {
        return this.$store.state.optionsDeoVR.players.subtitle_sort_seq
      },
      set (value) {
        this.$store.state.optionsDeoVR.players.subtitle_sort_seq = value
      },
    },
    multiTrackCastCuepoints: {
      get () {        
        return this.$store.state.optionsDeoVR.heresphere.multitrack_cast_cuepoints
      },
      set (value) {
        this.$store.state.optionsDeoVR.heresphere.multitrack_cast_cuepoints = value
      }
    },
    retainNonHSPCuepoints: {
      get () {        
        return this.$store.state.optionsDeoVR.heresphere.retain_non_hsp_cuepoints
      },
      set (value) {
        this.$store.state.optionsDeoVR.heresphere.retain_non_hsp_cuepoints = value
      }
    },
    isLoading: function () {
      return this.$store.state.optionsDeoVR.loading
    },
    deoVROptions: function () {
      return this.$store.state.optionsDeoVR.deovr
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

.options-tabs {
  margin-bottom: 1rem;
}

.options-tabs :deep(ul[role="tablist"]) {
  margin-left: 0;
}

.options-tabs :deep(.tabs.is-boxed a) {
  border-radius: var(--xbvr-radius-sm, 8px) var(--xbvr-radius-sm, 8px) 0 0;
}

.options-tabs :deep(.tabs.is-boxed li.is-active a) {
  background: var(--xbvr-surface, #ffffff);
  border-color: var(--xbvr-border, #e3e6ec);
  border-bottom-color: transparent;
  color: var(--xbvr-primary, #4f46e5);
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

.card-divider {
  background-color: var(--xbvr-border, #e3e6ec);
  margin: 0.85rem 0;
}

.muted {
  color: var(--xbvr-text-muted, #64708a);
  font-size: 0.9rem;
}

.cred-field {
  max-width: 240px;
}

.endpoint-list {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.35rem;
  margin: 0.5rem 0;
}

.save-row {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-top: 1rem;
}
</style>
