<template>
  <div class="gallery-editor">
    <div class="field">
      <div class="control">
        <input 
          class="input" 
          type="text" 
          v-model="newItem" 
          :placeholder="$t('Add URL or drag local image files here')" 
          @keyup.enter="addItem"
          @drop="handleFileDrop"
          @dragover.prevent
          @dragenter.prevent
          @dragleave.prevent>
      </div>
    </div>

    <!-- Lock Control -->
    <div class="lock-bar">
      <span class="dnd-hint">
        Drag images to reorder
      </span>
      <b-button 
        type="is-light" 
        size="is-small" 
        @click="toggleLock"
        :class="{ 'is-info': !isLocked, 'is-warning': isLocked }"
        icon-left="lock"
        class="lock-btn">
        {{ isLocked ? 'Unlock' : 'Lock' }} Delete
      </b-button>
    </div>

    <draggable :list="internalList" @end="onDragEnd" class="image-grid">
      <div v-for="(item, index) in internalList" :key="index" class="image-item">
        <img :src="getImageURL(item)" alt="Gallery image" class="gallery-image"/>
        <div class="image-controls">
          <b-tooltip :label="$t('Delete Image')" type="is-dark" position="is-top" :delay="500" append-to-body>
            <b-button 
              type="is-danger" 
              size="is-small" 
              @click="removeItem(index)" 
              icon-left="delete"
              :disabled="isLocked"
              :class="{ 'is-light': isLocked }">
            </b-button>
          </b-tooltip>
          <b-tooltip :label="$t('Set as Cover')" type="is-dark" position="is-top" :delay="500" append-to-body>
            <b-button
              type="is-primary"
              size="is-small"
              :class="{ 'is-light': item !== coverUrl }"
              @click="setCover(item)"
              icon-left="image">
            </b-button>
          </b-tooltip>
        </div>
      </div>
    </draggable>
  </div>
</template>

<script>
import draggable from 'vuedraggable'
import { getImageURL as getImageURLUtil } from '../util/image'

export default {
  name: 'GalleryEditor',
  components: {
    draggable
  },
  props: {
    list: {
      type: Array,
      required: true
    },
    coverUrl: {
      type: String,
      default: ''
    },
    blurFn: {
      type: Function,
      default: () => {}
    }
  },
  data () {
    return {
      internalList: [...this.list],
      newItem: '',
      isLocked: true // Default to locked
    }
  },
  watch: {
    list(newList) {
      this.internalList = [...newList];
    }
  },
  methods: {
    toggleLock() {
      this.isLocked = !this.isLocked;
    },
    addItem () {
      if (this.newItem.trim() !== '') {
        this.internalList.push(this.newItem.trim())
        this.newItem = ''
        this.updateList()
      }
    },
    removeItem (index) {
      if (!this.isLocked) {
        const itemToDelete = this.internalList[index]
        
        // Check if this is the cover image
        if (itemToDelete === this.coverUrl) {
          // Show confirmation dialog for deleting cover image
          this.$buefy.dialog.confirm({
            title: 'Delete Cover Image',
            message: 'You are about to delete the current cover image. This will clear the cover selection. Are you sure you want to continue?',
            confirmText: 'Delete',
            cancelText: 'Cancel',
            type: 'is-warning',
            hasIcon: true,
            onConfirm: () => {
              // Remove the image
              this.internalList.splice(index, 1)
              // Clear the cover_url since we're deleting the cover image
              this.$emit('setCover', '')
              this.updateList()
            }
          })
        } else {
          // Regular image deletion - just remove it without affecting cover_url
          this.internalList.splice(index, 1)
          this.updateList()
        }
      }
    },
    updateList () {
      this.$emit('update:list', this.internalList)
      this.blurFn()
    },
    onDragEnd (evt) {
      // 'draggable' updates the list automatically, so we just need to emit the update
      this.updateList()
    },
    setCover (url) {
      this.$emit('setCover', url)
    },
    getImageURL (url) {
      // proxy remote images via the shared util; keep the local-path
      // backslash normalization for Windows-style paths
      const out = getImageURLUtil(url, '200x')
      if (typeof out === 'string' && out.includes('\\')) {
        return out.replace(/\\/g, '/')
      }
      return out
    },
    handleFileDrop(event) {
      event.preventDefault();
      const files = event.dataTransfer.files;
      for (let i = 0; i < files.length; i++) {
        const file = files[i];
        const reader = new FileReader();
        reader.onload = (e) => {
          const url = e.target.result;
          this.newItem = url;
          this.addItem();
        };
        reader.readAsDataURL(file);
      }
    }
  }
}
</script>

<style scoped>
.gallery-editor .input {
  border-radius: var(--xbvr-radius, 12px);
}

/* lock bar above the grid */
.lock-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
  line-height: 1;
}

.dnd-hint {
  font-size: 0.7rem;
  color: var(--xbvr-text-faint, #7d88a1);
  line-height: 1;
}

.lock-btn {
  font-size: 0.7rem;
  padding: 0.25rem 0.6rem;
  line-height: 1;
  height: auto;
  border-radius: 999px;
}

.image-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 178px));
  grid-auto-rows: 120px;
  gap: 0.5rem;
  overflow-y: auto;
}

.image-item {
  position: relative;
  overflow: hidden;
  word-break: break-all;
  display: flex;
  align-items: stretch;
  justify-content: center;
  min-height: 120px;
  max-height: 120px;
  min-width: 120px;
  max-width: 178px;
  aspect-ratio: 16/9;
  background: var(--xbvr-surface-sunken, #eef0f4);
  border: 1px solid var(--xbvr-border, #e3e6ec);
  border-radius: var(--xbvr-radius-sm, 8px);
  box-shadow: var(--xbvr-shadow-sm, 0 1px 2px rgba(16, 24, 40, 0.05));
  transition: border-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
    box-shadow var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

.image-item:hover {
  border-color: var(--xbvr-border-strong, #cdd2dc);
  box-shadow: var(--xbvr-shadow, 0 1px 3px rgba(16, 24, 40, 0.08), 0 1px 2px rgba(16, 24, 40, 0.04));
}

.gallery-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.image-controls {
  position: absolute;
  bottom: 0.4rem;
  right: 0.4rem;
  display: flex;
  gap: 0.35rem;
  opacity: 0;
  transition: opacity var(--xbvr-med, 220ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

.image-item:hover .image-controls {
  opacity: 1;
}

.image-controls .button {
  height: 2rem;
  width: 2rem;
  padding: 0;
  border-radius: 8px;
}
</style> 
