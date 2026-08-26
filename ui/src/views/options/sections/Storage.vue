<template>
  <div class="content">
    <header class="options-page-head options-head-row">
      <div>
        <h1 class="options-title">{{$t('Storage')}}</h1>
        <p class="options-desc">{{ $t('Folders and services XBVR scans for video content.') }}</p>
      </div>
      <div class="options-head-actions">
        <a class="button is-primary" v-on:click="taskRescan">{{ $t('Rescan all folders') }}</a>
      </div>
    </header>

    <div class="settings-card table-card" v-if="items.length > 0">
      <div class="settings-card-title">
        <b-icon pack="mdi" icon="folder-multiple-outline" size="is-small"/>
        {{ $t('Storage paths') }}
      </div>
      <b-table :data="items"
               ref="table" default-sort="is_available" default-sort-direction="desc">
        <b-table-column field="path" :label="$t('Path')" sortable v-slot="props">
          {{ props.row.path }}
        </b-table-column>
        <b-table-column field="type" :label="$t('Type')" sortable v-slot="props">
          <b-icon pack="mdi" icon="cloud-outline" size="is-small" v-if="props.row.type !== 'local'"/>
          <b-icon pack="mdi" icon="folder-outline" size="is-small" v-else/>
        </b-table-column>
        <b-table-column field="is_available" :label="$t('Avail')" sortable v-slot="props">
          <b-icon pack="fas" icon="check" size="is-small" v-if="props.row.is_available"></b-icon>
        </b-table-column>
        <b-table-column field="file_count" :label="$t('# of files')" sortable v-slot="props">
          {{ props.row.file_count }}
        </b-table-column>
        <b-table-column field="unmatched_count" :label="$t('Not matched')" sortable v-slot="props">
          {{ props.row.unmatched_count }}
        </b-table-column>
        <b-table-column field="total_size" :label="$t('Total size')" sortable v-slot="props">
          {{ prettyBytes(props.row.total_size) }}
        </b-table-column>
        <b-table-column field="last_scan" :label="$t('Last scan')" sortable v-slot="props">
            <span v-if="props.row.last_scan !== '0001-01-01T00:00:00Z'">
              {{ formatDistanceToNow(parseISO(props.row.last_scan)) }} ago
            </span>
          <span v-else>never</span>
        </b-table-column>
        <b-table-column field="actions" v-slot="props">
          <b-field grouped>
            <button class="button is-small is-outlined" v-on:click='rescanFolder(props.row)' :title="$t('rescan folder')">
              <b-icon pack="mdi" icon="folder-refresh-outline"></b-icon>
            </button>
            <button class="button is-danger is-small is-outlined" v-on:click='removeFolder(props.row)'
                    :disabled="items.length <= 1"
                    :title="items.length <= 1 ? $t('Cannot remove the last remaining storage location') : $t('remove folder')">
              <b-icon pack="mdi" icon="close-circle" size="is-small"></b-icon>
            </button>
          </b-field>
        </b-table-column>
        <template slot="footer">
          <td></td>
          <td></td>
          <td></td>
          <td>{{ total.files }}</td>
          <td>{{ total.unmatched }}</td>
          <td>{{ prettyBytes(total.size) }}</td>
          <td></td>
          <td></td>
        </template>
      </b-table>
    </div>
    <div class="settings-card empty-card" v-else>
      <div class="empty-state">
        <span class="icon empty-icon">
          <b-icon pack="mdi" icon="folder-outline" size="is-large"></b-icon>
        </span>
        <p class="empty-text">{{ $t('Add folders with VR videos') }}</p>
      </div>
    </div>

    <div class="columns">
      <div class="column">
        <div class="settings-card">
          <div class="settings-card-title">
            <b-icon pack="mdi" icon="folder-plus-outline" size="is-small"/>
            {{ $t('Add local folder') }}
          </div>
          <div class="field narrow-field">
            <label class="label">{{ $t('Path to folder with content') }}</label>
            <div class="control">
              <input class="input" type="text" v-model='newVolumePath'>
            </div>
          </div>
          <div class="control">
            <button class="button is-link" v-on:click='addFolder'>{{ $t('Add new folder') }}</button>
          </div>
        </div>
      </div>
      <div class="column">
        <div class="settings-card">
          <div class="settings-card-title">
            <b-icon pack="mdi" icon="cloud-plus-outline" size="is-small"/>
            {{ $t('Add cloud storage') }}
          </div>
          <b-field grouped>
            <b-field :label="$t('Service')">
              <b-select placeholder="Select one" v-model="serviceSelected">
                <option v-for="option in serviceOpts" :value="option.id" :key="option.id">
                  {{ option.name }}
                </option>
              </b-select>
            </b-field>
            <b-field :label="$t('Token')" expanded>
              <b-input v-model='serviceToken' type='password' password-reveal/>
            </b-field>
          </b-field>
          <div class="control">
            <button class="button is-link" v-on:click='addCloudStorage'
                    :disabled="serviceSelected === null || serviceToken === ''">{{ $t('Add service') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="settings-card">
      <div class="settings-card-title">
        <b-icon pack="mdi" icon="webhook" size="is-small"/>
        {{ $t('Webhooks') }}
      </div>
      <div class="columns">
        <div class="column webhook-col" v-for="wh in webhookDefs" :key="wh.key">
          <h4 class="webhook-title">{{ $t(wh.title) }}</h4>
          <p class="webhook-desc">{{ $t(wh.description) }}</p>
          <b-field :label="$t('HTTP Method')">
            <b-select v-model="webhooks[wh.key].method">
              <option value="GET">GET</option>
              <option value="POST">POST</option>
              <option value="PUT">PUT</option>
            </b-select>
          </b-field>
          <b-field :label="$t('URL')">
            <b-input v-model="webhooks[wh.key].url" placeholder="https://example.com/hook"/>
          </b-field>
          <b-field :label="$t('Headers')">
            <b-input type="textarea" v-model="webhooks[wh.key].headers"
                     :placeholder="$t('One header per line, e.g.') + '\nAuthorization: Bearer token'"/>
          </b-field>
        </div>
      </div>
      <div class="control">
        <button class="button is-link" v-on:click="saveWebhooks">{{ $t('Save webhooks') }}</button>
      </div>
    </div>

  <div class="settings-card">
    <div class="settings-card-title">
      <b-icon pack="mdi" icon="tune" size="is-small"/>
      {{ $t('Options') }}
    </div>
    <b-field>
      <b-switch v-model="match_ohash" type="is-default">
        Match StashDB Hashes
      </b-switch>
    </b-field>

    <b-field label="Video File Extensions">
      <b-tooltip label="Only add video file extensions!" position="is-top" class="full-width-tooltip">
        <b-taginput
            ref="videoExtInput"
            v-model="video_ext"
            :allow-new="true"
            @add="OnExtAdded"
            @remove="saveExtensions"
            placeholder="Add a video extension (e.g. .mp4)">
        </b-taginput>
      </b-tooltip>
    </b-field>
    <b-field>
      <b-tooltip label="This only resets the Video File Extensions list." position="is-right">
        <b-button type="is-warning" @click="resetToDefaults">Reset</b-button>
      </b-tooltip>
    </b-field>
  </div>
  </div>
</template>

<script>
import api from '../../../api'
import prettyBytes from 'pretty-bytes'
import { formatDistanceToNow, parseISO } from 'date-fns'

export default {
  name: 'Storage',
  data () {
    return {
      volumes: [],
      serviceOpts: [{ name: 'Put.io', id: 'putio' }],
      serviceToken: '',
      serviceSelected: null,
      newVolumePath: '',
      webhookDefs: [
        {
          key: 'trigger_external_import',
          title: 'Trigger External Import',
          description: 'This can call an external tool to import scenes from another service.'
        },
        {
          key: 'refresh_external_import',
          title: 'Refresh External Import',
          description: 'This can call an external tool to refresh imports - this can be used if the tool provides a more lightweight import option if there is preexisting state.'
        }
      ],
      prettyBytes,
      parseISO,
      formatDistanceToNow,
      lastAddedTag: null,
      lastAddedTime: 0
    }
  },
  mounted () {
    this.$store.dispatch('optionsStorage/load')
  },
  methods: {
    async taskRescan () {
      try {
        await this.$store.dispatch('optionsStorage/save');
        await api.get('/task/rescan');
      } catch (e) {
        this.$buefy.dialog.alert({
          title: 'Error',
          message: 'Failed to save options. Rescan was not started.',
          type: 'is-danger',
          hasIcon: true,
          ariaRole: 'alertdialog',
          ariaModal: true
        });
      }
    },
    addFolder: async function () {
      await api.post('/options/storage', { json: { path: this.newVolumePath, type: 'local' } })
    },
    addCloudStorage: async function () {
      await api.post('/options/storage', { json: { token: this.serviceToken, type: this.serviceSelected } })
    },
    removeFolder: function (folder) {
      if (this.items.length <= 1) {
        return
      }
      this.$buefy.dialog.confirm({
        title: this.$t('Remove folder'),
        message: `You're about to remove storage location <strong>${folder.path}</strong> and its files from local database - files will remain intact at the location.`,
        type: 'is-danger',
        hasIcon: true,
        onConfirm: function () {
          api.delete(`/options/storage/${folder.id}`)
        }
      })
    },
    rescanFolder: function (folder) {
      api.get(`/task/rescan/${folder.id}`)
    },
    saveExtensions () {
      this.$store.dispatch('optionsStorage/save')
    },
    saveWebhooks () {
      this.$store.dispatch('optionsStorage/save')
    },
    OnExtAdded(tag) {
      // Debounce the add event as it also triggers on blur
      const now = Date.now();
      if (this.lastAddedTag === tag && now - this.lastAddedTime < 200) {
        return;
      }
      this.lastAddedTag = tag;
      this.lastAddedTime = now;

      // Always remove the raw tag to sanitize in logic below
      const index = this.video_ext.indexOf(tag);
      if (index > -1) {
        this.video_ext.splice(index, 1);
      }

      // Trim whitespace and convert to lowercase
      let cleanTag = tag.trim().toLowerCase();

      // Reject if it contains whitespace in the middle
      if (/\s/.test(cleanTag)) {
        this.$buefy.dialog.alert({
          title: 'Invalid Extension',
          message: 'File extensions cannot contain whitespace.',
          type: 'is-danger',
          hasIcon: true,
          ariaRole: 'alertdialog',
          ariaModal: true
        });
        // Clear the taginput's input field
        if (this.$refs.videoExtInput) {
          this.$refs.videoExtInput.newTag = '';
        }
        return;
      }

      // Reject if it's empty or just a dot
      if (cleanTag.length === 0 || cleanTag === '.') {
        return;
      }

      // Prepend dot if missing
      if (!cleanTag.startsWith('.')) {
        cleanTag = '.' + cleanTag;
      }

      // Reject if it's a forbidden extension
      if (this.forbidden_video_ext && this.forbidden_video_ext.includes(cleanTag)) {
        this.$buefy.dialog.alert({
          title: 'Invalid Extension',
          message: `The file extension <strong>${cleanTag}</strong> is a reserved extension and cannot be added.`,
          type: 'is-danger',
          hasIcon: true,
          ariaRole: 'alertdialog',
          ariaModal: true
        });
        // Clear the taginput's input field
        if (this.$refs.videoExtInput) {
          this.$refs.videoExtInput.newTag = '';
        }
        return;
      }

      // Only allow extensions that are a dot followed by alphanumeric characters
      if (!/^\.[a-z0-9]+$/.test(cleanTag)) {
        this.$buefy.dialog.alert({
          title: 'Invalid Extension',
          message: 'File extensions must only contain letters and numbers after the dot.',
          type: 'is-danger',
          hasIcon: true,
          ariaRole: 'alertdialog',
          ariaModal: true
        });
        if (this.$refs.videoExtInput) {
          this.$refs.videoExtInput.newTag = '';
        }
        return;
      }

      // Add the sanitized tag back, if not a duplicate
      if (!this.video_ext.includes(cleanTag)) {
        this.video_ext.push(cleanTag);
        this.saveExtensions();
      }
    },
    resetToDefaults () {
      this.video_ext = this.default_video_ext.slice()
      this.saveExtensions()
    },
  },
  computed: {
    match_ohash: {
      get () {        
        return this.$store.state.optionsStorage.options.match_ohash
      },
      set (value) {
        this.$store.commit('optionsStorage/setOption', { key: 'match_ohash', value })
      },
    },
    total () {
      let files = 0; let unmatched = 0; let size = 0
      this.$store.state.optionsStorage.items.map(v => {
        files = files + v.file_count
        unmatched = unmatched + v.unmatched_count
        size = size + v.total_size
      })
      return { files, unmatched, size }
    },
    items () {
      return this.$store.state.optionsStorage.items
    },
    video_ext: {
      get () {return this.$store.state.optionsStorage.options.video_ext},
      set (value) { this.$store.commit('optionsStorage/setOption', { key: 'video_ext', value }) },
    },
    forbidden_video_ext: {
      get () {return this.$store.state.optionsStorage.options.forbidden_video_ext}
    },
    default_video_ext: {
      get () {return this.$store.state.optionsStorage.options.default_video_ext}
    },
    webhooks () {
      return this.$store.state.optionsStorage.options.webhooks
    },
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

.table-card {
  padding: 1.25rem 1.25rem 0.75rem;
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

.narrow-field {
  max-width: 420px;
}

.empty-card {
  padding: 2.5rem 1.25rem;
}

.empty-state {
  text-align: center;
}

.empty-icon {
  color: var(--xbvr-text-faint, #7d88a1);
}

.empty-text {
  color: var(--xbvr-text-muted, #64708a);
  font-size: 1rem;
  margin-top: 0.5rem;
}

.webhook-title {
  font-weight: 700;
  color: var(--xbvr-text, #1c2333);
  margin-bottom: 0.25rem;
}

.webhook-desc {
  color: var(--xbvr-text-muted, #64708a);
  font-size: 0.85rem;
  margin-bottom: 0.75rem;
}

.full-width-tooltip {
  width: 100%;
}

.columns .settings-card {
  height: 100%;
}
</style>
