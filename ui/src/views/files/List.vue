<template>
  <div class="files-list">
    <b-loading :is-full-page="true" :active.sync="isLoading"></b-loading>
    <div v-if="items.length > 0 && !isLoading">
      <b-table :data="items" ref="table" backend-sorting :default-sort="[sortField, sortOrder]" @sort="onSort"
               :paginated="true" class="files-table">
        <b-table-column field="filename" :label="$t('File')" sortable v-slot="props">
          <div class="file-cell">
            <span class="file-name">{{ props.row.filename }}</span>
            <span class="file-path">{{ props.row.path }}</span>
          </div>
        </b-table-column>
        <b-table-column field="created_time" :label="$t('Created')" sortable v-slot="props">
          <span class="cell-meta">{{ format(parseISO(props.row.created_time), "yyyy-MM-dd HH:mm:ss") }}</span>
        </b-table-column>
        <b-table-column field="size" :label="$t('Size')" sortable v-slot="props">
          <span class="cell-meta">{{ prettyBytes(props.row.size) }}</span>
        </b-table-column>
        <b-table-column field="video_width" :label="$t('Width')" sortable v-slot="props">
          <span class="cell-meta" v-if="props.row.video_width !== 0">{{ props.row.video_width }}</span>
          <span class="cell-meta is-empty" v-else>-</span>
        </b-table-column>
        <b-table-column field="video_height" :label="$t('Height')" sortable v-slot="props">
          <span class="cell-meta" v-if="props.row.video_height !== 0">{{ props.row.video_height }}</span>
          <span class="cell-meta is-empty" v-else>-</span>
        </b-table-column>
        <b-table-column field="video_bitrate" :label="$t('Bitrate')" sortable v-slot="props">
          <span class="cell-meta" v-if="props.row.video_bitrate !== 0">{{ prettyBytes(props.row.video_bitrate, { bits: true }) }}/s</span>
          <span class="cell-meta is-empty" v-else>-</span>
        </b-table-column>
        <b-table-column field="duration" :label="$t('Duration')" sortable v-slot="props">
          <span class="cell-meta" v-if="props.row.duration !== 0">{{ humanizeSeconds(props.row.duration) }}</span>
          <span class="cell-meta is-empty" v-else>-</span>
        </b-table-column>
        <b-table-column field="video_avgfps_val" :label="$t('FPS')" sortable v-slot="props">
          <span class="cell-meta" v-if="props.row.video_avgfps_val !== 0">{{ props.row.video_avgfps_val }}</span>
          <span class="cell-meta is-empty" v-else>-</span>
        </b-table-column>
        <b-table-column v-slot="props">
          <div class="row-actions">
            <b-button size="is-small" @click="play(props.row)" v-if="props.row.type === 'video'">{{ $t('Play') }}</b-button>
            <b-button size="is-small" v-if="props.row.scene_id === 0" class="is-primary is-outlined" @click="match(props.row)">{{ $t('Match') }}</b-button>
            <b-button size="is-small" v-else class="is-outlined" @click="unmatch(props.row)">{{ $t('Unmatch') }}</b-button>
            <button class="button is-small is-success is-outlined" @click="createScene(props.row)" :title="$t('Create a custom scene for this file')">
              <b-icon pack="fas" icon="plus-square" size="is-small"></b-icon>
            </button>
            <button class="button is-small is-danger is-outlined" @click='removeFile(props.row)' title='Delete file from disk'>
              <b-icon pack="fas" icon="trash" size="is-small"></b-icon>
            </button>
          </div>
        </b-table-column>
      </b-table>
    </div>
    <div v-if="items.length === 0 && !isLoading" class="empty-state">
      <span class="icon empty-icon">
        <i class="far fa-check-circle is-superlarge"></i>
      </span>
      <p class="empty-text">{{ $t('No files matching your selection') }}</p>
    </div>
  </div>
</template>

<script>
import prettyBytes from 'pretty-bytes'
import { format, parseISO } from 'date-fns'
import api from '../../api'
import { humanizeSeconds } from '../../util/image'

export default {
  name: 'List',
  data () {
    return {
      files: [],
      prettyBytes,
      format,
      parseISO,
      sortField: 'created_time',
      sortOrder: 'desc'
    }
  },
  computed: {
    isLoading () {
      return this.$store.state.files.isLoading
    },
    items () {
      return this.$store.state.files.items
    }
  },
  mounted () {
    this.$store.state.files.filters.sort = `${this.sortField}_${this.sortOrder}`
    this.$store.dispatch('files/load')
  },
  methods: {
    onSort (field, order) {
      this.sortField = field
      this.sortOrder = order
      this.$store.state.files.filters.sort = `${field}_${order}`
      this.$store.dispatch('files/load')
    },
    play (file) {
      this.$store.commit('overlay/showPlayer', { file: file })
    },
    match (file) {
      this.$store.commit('overlay/showMatch', { file: file })
    },
    unmatch (file) {
      api.post('/files/unmatch', {
        json: {
          file_id: file.id
        }
      }).then(data => {
            this.$store.dispatch('files/load')
      })
    },
    createScene (file) {
      this.$store.commit('overlay/createCustomScene', { file: file })
    },
    humanizeSeconds,
    removeFile (file) {
      this.$buefy.dialog.confirm({
        title: 'Remove file',
        message: `You're about to remove file <strong>${file.filename}</strong> from <strong>disk</strong>.`,
        type: 'is-danger',
        hasIcon: true,
        onConfirm: () => {
          api.delete(`/files/file/${file.id}`).json().then(data => {
            this.$store.dispatch('files/load')
          })
        }
      })
    }
  }
}
</script>

<style lang="less" scoped>
.files-list {
  min-height: 40vh;
  width: 100%;
}

.files-table {
  width: 100%;
}

/* ------------------------------------------------------------------
   File table — rows read as rounded surface cards
   ------------------------------------------------------------------ */

.files-table :deep(.table) {
  background: transparent;
  border-collapse: separate;
  border-spacing: 0 0.55rem;
}

.files-table :deep(.table thead th) {
  background: transparent;
  border: none;
  padding: 0.2rem 0.85rem 0.4rem;
  white-space: nowrap;
}

.files-table :deep(.table thead th .th-wrap) {
  text-transform: uppercase;
  letter-spacing: 0.05em;
  font-size: 0.72rem;
  font-weight: 700;
  color: var(--xbvr-text-muted, #64708a);
}

.files-table :deep(.table tbody tr) {
  background: transparent;
}

.files-table :deep(.table tbody td) {
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

.files-table :deep(.table tbody td:first-child) {
  border-left: 1px solid var(--xbvr-border, #e3e6ec);
  border-radius: var(--xbvr-radius, 12px) 0 0 var(--xbvr-radius, 12px);
}

.files-table :deep(.table tbody td:last-child) {
  border-right: 1px solid var(--xbvr-border, #e3e6ec);
  border-radius: 0 var(--xbvr-radius, 12px) var(--xbvr-radius, 12px) 0;
}

.files-table :deep(.table tbody tr:hover td),
.files-table :deep(.table tbody tr:hover td:first-child),
.files-table :deep(.table tbody tr:hover td:last-child) {
  background: var(--xbvr-hover-bg, #fafbfd);
  border-color: var(--xbvr-border-strong, #cdd2dc);
}

/* ------------------------------------------------------------------
   Cell content
   ------------------------------------------------------------------ */

.file-cell {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
  min-width: 16rem;
}

.file-name {
  font-weight: 600;
  font-size: 0.92rem;
  line-height: 1.3;
  color: var(--xbvr-text, #1c2333);
  overflow-wrap: anywhere;
}

.file-path {
  font-size: 0.76rem;
  color: var(--xbvr-text-faint, #7d88a1);
  overflow-wrap: anywhere;
}

.cell-meta {
  font-size: 0.85rem;
  color: var(--xbvr-text-muted, #64708a);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.cell-meta.is-empty {
  color: var(--xbvr-text-faint, #7d88a1);
}

/* action buttons grouped right */
.row-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.row-actions .button {
  margin: 0;
}

/* ------------------------------------------------------------------
   Empty state
   ------------------------------------------------------------------ */

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.9rem;
  padding: 5rem 1rem;
  text-align: center;
}

.empty-icon {
  color: var(--xbvr-text-faint, #7d88a1);
}

.empty-text {
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--xbvr-text-muted, #64708a);
}

.is-superlarge {
  height: 96px;
  max-height: 96px;
  max-width: 96px;
  min-height: 96px;
  min-width: 96px;
  width: 96px;
}
</style>
