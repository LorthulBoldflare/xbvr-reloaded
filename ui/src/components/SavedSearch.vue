<template>
  <div class="saved-search">
    <div class="saved-search-row">
      <b-select size="is-small" @input="setPlaylist" expanded v-model="currentPlaylist" class="saved-search-select">
        <optgroup label="Web">
          <option v-for="(obj, idx) in playlistsWeb" :value="obj.id" :key="idx">
            {{ obj.name }}
          </option>
        </optgroup>
        <optgroup label="VR Players">
          <option v-for="(obj, idx) in playlistsDeo" :value="obj.id" :key="idx">
            {{ obj.name }}
          </option>
        </optgroup>
      </b-select>

      <div class="saved-search-actions">
        <b-tooltip position="is-bottom" label="Save as new" :delay="200">
          <button class="button is-small is-outlined" @click="showNewDialog" :aria-label="$t('Save as new')">
            <b-icon pack="mdi" icon="content-save-outline"></b-icon>
          </button>
        </b-tooltip>
        <b-tooltip position="is-bottom" label="Edit" :delay="200">
          <button class="button is-small is-outlined" @click="showEditDialog" :disabled="disableEditDelete" :aria-label="$t('Edit saved search')">
            <b-icon pack="mdi" icon="square-edit-outline"></b-icon>
          </button>
        </b-tooltip>
        <b-tooltip position="is-bottom" label="Delete" :delay="200">
          <button class="button is-small is-outlined" @click="removePlaylist" :disabled="disableEditDelete" :aria-label="$t('Delete saved search')">
            <b-icon pack="mdi" icon="delete-outline"></b-icon>
          </button>
        </b-tooltip>
      </div>
    </div>

    <b-modal :active.sync="isPlaylistModalActive"
             has-modal-card
             trap-focus
             aria-role="dialog"
             aria-modal>
      <div class="modal-card saved-search-modal">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ modalTitle }}</p>
        </header>
        <section class="modal-card-body">
          <div class="filter-field">
            <label class="filter-label">{{ $t('Name') }}</label>
            <b-input
              type="name"
              v-model="playlistName"
              required>
            </b-input>
          </div>
          <b-checkbox v-if="!isActorMode" v-model="playlistDeoEnabled">Use as DeoVR list</b-checkbox>
        </section>
        <footer class="modal-card-foot">
          <button class="button is-primary" :disabled="playlistName===''" @click="savePlaylist(modalAction)">Save
          </button>
        </footer>
      </div>
    </b-modal>
  </div>
</template>

<script>
import api from '../api'

// Shared saved-search picker for the scenes and actors filter panels.
// `mode` selects the store module, target route and playlist flavor.
export default {
  name: 'SavedSearch',
  props: {
    mode: {
      type: String,
      default: 'scenes',
      validator: v => ['scenes', 'actors'].includes(v)
    }
  },
  mounted () {
    this.$store.dispatch(this.storeModule + '/filters')
  },
  data () {
    return {
      currentPlaylistObj: null,

      isPlaylistModalActive: false,
      modalTitle: '',
      modalAction: 'create',
      playlistName: '',
      playlistDeoEnabled: false
    }
  },
  methods: {
    showNewDialog () {
      this.modalTitle = 'Create new saved search'
      this.modalAction = 'create'
      this.playlistName = ''
      this.playlistDeoEnabled = false

      this.isPlaylistModalActive = true
    },
    showEditDialog () {
      if (this.currentPlaylistObj !== null) {
        this.modalTitle = 'Edit saved search'
        this.modalAction = 'update'
        this.playlistName = this.currentPlaylistObj.name
        this.playlistDeoEnabled = this.currentPlaylistObj.is_deo_enabled

        this.isPlaylistModalActive = true
      }
    },
    setPlaylist (val) {
      const obj = this.playlists.find(item => item.id === val)

      this.$router.push({
        name: this.routeName,
        query: {
          q: this.$store.getters[this.storeModule + '/getQueryParamsFromObject'](obj.search_params)
        }
      })

      this.$store.dispatch(this.storeModule + '/load', { offset: 0 })
    },
    async savePlaylist (action) {
      const payload = {
        name: this.playlistName,
        is_deo_enabled: this.isActorMode ? false : this.playlistDeoEnabled,
        is_smart: true,
        search_params: JSON.stringify(this.$store.state[this.storeModule].filters)
      }
      if (this.isActorMode) {
        payload.playlist_type = 'actor'
      }

      let p
      if (action === 'create') {
        p = await api.post('/playlist', { json: payload }).json()
      } else {
        p = await api.put(`/playlist/${this.currentPlaylistObj.id}`, { json: payload }).json()
      }

      await this.$store.dispatch(this.storeModule + '/filters')
      this.currentPlaylist = p.id
      this.isPlaylistModalActive = false
    },
    removePlaylist () {
      this.$buefy.dialog.confirm({
        title: 'Delete saved search',
        message: `Do you want to delete saved search <strong>${this.currentPlaylistObj.name}</strong>?`,
        type: 'is-danger',
        hasIcon: true,
        confirmText: 'Delete',
        onConfirm: () => {
          api.delete(`/playlist/${this.currentPlaylistObj.id}`).then(() => {
            this.$store.dispatch(this.storeModule + '/filters')
            this.currentPlaylist = null
          })
        }
      })
    }

  },
  computed: {
    isActorMode () {
      return this.mode === 'actors'
    },
    storeModule () {
      return this.isActorMode ? 'actorList' : 'sceneList'
    },
    routeName () {
      return this.isActorMode ? 'actors' : 'scenes'
    },
    playlists () {
      return this.$store.state[this.storeModule].playlists
    },
    playlistsWeb () {
      return this.playlists.filter((obj) => {
        return obj.is_deo_enabled === false
      })
    },
    playlistsDeo () {
      return this.playlists.filter((obj) => {
        return obj.is_deo_enabled === true
      })
    },
    currentPlaylist: {
      get () {
        if (this.currentPlaylistObj !== null) {
          return this.currentPlaylistObj.id
        }
        return null
      },
      set (val) {
        if (val === null) {
          this.currentPlaylistObj = null
          return null
        }
        this.currentPlaylistObj = this.playlists.find(item => item.id === val)
        return this.currentPlaylistObj.id
      }
    },
    disableEditDelete () {
      if (this.currentPlaylistObj === null || this.currentPlaylistObj.is_system === true) {
        return true
      }
      return false
    }
  }
}
</script>

<style lang="scss" scoped>
.saved-search-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.saved-search-select {
  flex: 1 1 auto;
  min-width: 0;
  margin-bottom: 0;
}

.saved-search-actions {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  flex: 0 0 auto;
}

.saved-search-actions .button {
  margin: 0;
  border-radius: var(--xbvr-radius-sm, 8px);
}

.saved-search-modal {
  width: min(420px, 92vw);
}

/* reuse the filter-panel field language inside the modal */
.filter-field {
  margin-bottom: 0.8rem;
}

.filter-label {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--xbvr-text-muted, #64708a);
  margin-bottom: 0.3rem;
}
</style>
