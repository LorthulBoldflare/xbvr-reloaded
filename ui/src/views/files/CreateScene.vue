<template>
  <div class="modal is-active">
    <div class="modal-background"></div>
    <div class="modal-card create-card">
      <header class="modal-card-head">
        <p class="modal-card-title">{{ $t("Create Custom Scene") }}</p>
        <button class="delete" @click="close" aria-label="close"></button>
      </header>
      <section class="modal-card-body create-body">
        <header class="create-header">
          <h2 class="create-title">{{ file.filename }}</h2>
          <p class="pathDetails">{{ file.path }}</p>
          <div class="meta-chips">
            <span class="meta-chip">{{ prettyBytes(file.size) }}</span>
            <span class="meta-chip">{{ file.video_width }}x{{ file.video_height }}</span>
            <span class="meta-chip">{{ format(parseISO(file.created_time), "yyyy-MM-dd") }}</span>
          </div>
        </header>
        <div class="form-field">
          <label class="form-label" for="create-scene-id">{{ $t('Scene Id') }}</label>
          <b-field grouped>
            <b-tooltip label="If blank a Scene Id will be generated but cannot be changed later"  :delay="500" class="is-fullwidth">
              <b-input id="create-scene-id" v-model="sceneId" placeholder="Can be empty" ref="sceneIdInput"></b-input>
            </b-tooltip>
          </b-field>
        </div>
        <div class="form-field">
          <label class="form-label" for="create-scene-title">{{ $t('Title') }}</label>
          <b-field>
            <b-input id="create-scene-title" v-model='title'></b-input>
          </b-field>
        </div>
        <div class="action-row">
          <b-button class="button is-primary" v-on:click="addScene(false)">{{$t('Create')}}</b-button>
          <b-button class="button is-primary is-outlined" v-on:click="addScene(true)">{{$t('Create and Edit')}} </b-button>
        </div>
      </section>
    </div>
  </div>
</template>

<script>
import api from '../../api'
import { format, parseISO } from 'date-fns'
import prettyBytes from 'pretty-bytes'

export default {
  name: 'CreateScene',  
  data () {
    return {
      title: '',
      sceneId: '',
      format,
      parseISO
    }
  },
  computed: {
    file () {
      return this.$store.state.overlay.createScene.file
    }
  },
  mounted () {
    this.initView()
  },
  methods: {
    initView () {
      const commonWords = [
        '180', '180x180', '2880x1440', '3d', '3dh', '3dv', '30fps', '30m', '360',
        '3840x1920', '4k', '5k', '5400x2700', '60fps', '6k', '7k', '7680x3840',
        '8k', 'fb360', 'fisheye190', 'funscript', 'h264', 'h265', 'hevc', 'hq', 'hsp', 'lq', 'lr',
        'mkv', 'mkx200', 'mkx220', 'mono', 'mp4', 'oculus', 'oculus5k',
        'oculusrift', 'original', 'rf52', 'smartphone', 'srt', 'ssa', 'tb', 'uhq', 'vrca220', 'vp9'
      ]
      const isNotCommonWord = word => !commonWords.includes(word.toLowerCase()) && !/^[0-9]+p$/.test(word)

      this.title = (
        this.file.filename
          .replace(/\.|_|\+|-/g, ' ').replace(/\s+/g, ' ').trim()
          .split(' ').filter(isNotCommonWord).join(' ')
          .replace(/ s /g, '\'s '))
      this.$refs.sceneIdInput.focus()
    },
    close () {
      this.$store.commit('overlay/hideCreateCustomScene')
    },
    toInt (value, radix, defaultValue) {
      return parseInt(value, radix || 10) || defaultValue || 0
    },
    addScene(showEdit) {      
      api.post('/scene/create', { json: { title: this.title, id: this.sceneId, filename: this.file.filename } })
        .json()
        .then(scene => {          
          api.post('/files/match', { json: {file_id: this.file.id, scene_id: scene.scene_id}})          
          .then(data => {
            this.$store.dispatch('files/load')
            this.close()
            if (showEdit) {
              this.$store.commit('overlay/editDetails', { scene: scene })
            }
          })          
        })
    },
    prettyBytes
  }
}
</script>

<style lang="less" scoped>
/* ------------------------------------------------------------------
   Create-scene overlay — same modal language as scenes/Details.vue
   ------------------------------------------------------------------ */

.create-card {
  position: relative;
  margin: 3rem auto;
  width: min(640px, 94vw);
}

.create-body {
  padding: 1.25rem;
}

.create-header {
  margin-bottom: 1rem;
}

.create-title {
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

/* labeled form fields */
.form-field {
  margin-bottom: 0.9rem;
}

.form-field .field {
  margin-bottom: 0;
}

.form-label {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--xbvr-text-muted, #64708a);
  margin-bottom: 0.3rem;
  cursor: pointer;
}

.form-field :deep(.b-tooltip) {
  width: 100%;
}

.action-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 1.1rem;
}

.action-row .button {
  margin: 0;
}
</style>
