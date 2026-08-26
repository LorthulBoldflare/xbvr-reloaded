<template>
  <div>
    <b-loading :is-full-page="false" :active.sync="isLoading"></b-loading>
    <div class="content">
      <header class="options-page-head">
        <h1 class="options-title">{{$t("Previews")}}</h1>
        <p class="options-desc">{{ $t('Control how hover video previews are generated.') }}</p>
      </header>
      <div class="columns">
        <div class="column">
          <section class="settings-card">
            <div class="settings-card-title">
              <b-icon pack="mdi" icon="tune" size="is-small"/>
              {{ $t('Preview settings') }}
            </div>
            <b-field label="Start time">
              <div class="columns slider-row">
                <div class="column is-two-thirds">
                  <b-slider :min="5" :max="60" :step="5" :tooltip="false" v-model="startTime"></b-slider>
                </div>
                <div class="column">
                  <div class="slider-value">{{startTime}}sec</div>
                </div>
              </div>
            </b-field>
            <b-field label="Snippet length">
              <div class="columns slider-row">
                <div class="column is-two-thirds">
                  <b-slider :min="0.2" :max="5" :step="0.2" :tooltip="false" v-model="snippetLength"></b-slider>
                </div>
                <div class="column">
                  <div class="slider-value">{{snippetLength}}sec</div>
                </div>
              </div>
            </b-field>
            <b-field label="Number of snippets">
              <div class="columns slider-row">
                <div class="column is-two-thirds">
                  <b-slider :min="2" :max="40" :step="1" :tooltip="false" v-model="snippetAmount"></b-slider>
                </div>
                <div class="column">
                  <div class="slider-value">{{snippetAmount}}</div>
                </div>
              </div>
            </b-field>
            <b-field>
              <b-checkbox v-model="extraSnippet">Grab extra snippet from the end of video</b-checkbox>
            </b-field>
            <b-field label="Preview resolution">
              <div class="columns slider-row">
                <div class="column is-two-thirds">
                  <b-slider :min="300" :max="800" :step="20" :tooltip="false" v-model="resolution"></b-slider>
                </div>
                <div class="column">
                  <div class="slider-value">{{resolution}}px</div>
                </div>
              </div>
            </b-field>
            <b-field grouped>
              <b-button type="is-primary" @click="saveSettings">Save settings</b-button>
              <b-button @click="testSettings" :disabled="generatingPreview">Test settings</b-button>
              <b-button @click="regenerateTestVideo" :disabled="generatingPreview">Regenerate test video</b-button>
            </b-field>
          </section>
          <section class="settings-card">
            <div class="settings-card-title">
              <b-icon pack="mdi" icon="motion-play-outline" size="is-small"/>
              {{ $t('Generation') }}
            </div>
            <p class="muted">
              Once you picked preview settings, you should start generating them.
            </p>
            <p class="muted">
              BETA NOTE: Please note this is CPU-heavy process.
            </p>
            <b-field grouped>
              <b-button type="is-primary" @click="startGenerating" :disabled="queue.running">Start generating previews</b-button>
              <b-button type="is-danger" @click="stopGenerating" :disabled="!queue.running || queue.stopping">
                {{ queue.stopping ? 'Stopping...' : 'Stop generating previews' }}
              </b-button>
            </b-field>
            <div v-if="queue.running" class="queue-progress">
              <b-progress :value="queue.completed" :max="queue.total" show-value format="percent"></b-progress>
              <p class="muted progress-text">
                {{ queue.completed }} / {{ queue.total }} previews generated
                <template v-if="queue.remaining > 0">({{ queue.remaining }} remaining)</template>
                <template v-if="queue.currentScene"> - currently rendering: {{ queue.currentScene }}</template>
              </p>
            </div>
          </section>
        </div>
        <div class="column">
          <video v-if="isPreviewReady" :src="`/api/dms/preview/${previewFn}?ts=${previewTs}`" autoplay loop></video>
          <div v-if="generatingPreview">
            <div class="preview-loading">
              <div class="bbox">
                <b-icon pack="fas" icon="sync-alt" custom-class="fa-spin"></b-icon>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import api from '../../../api'
import prettyBytes from 'pretty-bytes'

export default {
  name: 'Previews',
  data () {
    return {
      isLoading: true,
      startTime: 5,
      snippetLength: 0.2,
      snippetAmount: 2,
      resolution: 300,
      extraSnippet: false
    }
  },
  async mounted () {
    await this.loadState()
    await this.loadQueueStatus()
  },
  computed: {
    generatingPreview () {
      return this.$store.state.optionsPreviews.generatingPreview
    },
    isPreviewReady () {
      return this.$store.state.optionsPreviews.isPreviewReady
    },
    previewFn () {
      return this.$store.state.optionsPreviews.previewFn
    },
    previewTs () {
      return this.$store.state.optionsPreviews.previewTs
    },
    queue () {
      return this.$store.state.optionsPreviews.queue
    }
  },
  methods: {
    async loadState () {
      this.isLoading = true
      await api.get('/options/state')
        .json()
        .then(data => {
          this.startTime = data.config.library.preview.startTime
          this.snippetLength = data.config.library.preview.snippetLength
          this.snippetAmount = data.config.library.preview.snippetAmount
          this.resolution = data.config.library.preview.resolution
          this.extraSnippet = data.config.library.preview.extraSnippet
          this.isLoading = false
        })
    },
    async saveSettings () {
      this.isLoading = true
      await api.put('/options/previews', {
        json: {
          startTime: this.startTime,
          snippetLength: this.snippetLength,
          snippetAmount: this.snippetAmount,
          resolution: this.resolution,
          extraSnippet: this.extraSnippet
        }
      })
        .json()
        .then(data => {
          this.isLoading = false
        })
    },
    async loadQueueStatus () {
      await api.get('/task/preview/status')
        .json()
        .then(data => {
          this.$store.commit('optionsPreviews/setQueue', data)
        })
    },
    async testSettings () {
      this.$store.commit('optionsPreviews/hidePreview')
      await api.post('/options/previews/test', {
        json: {
          startTime: this.startTime,
          snippetLength: this.snippetLength,
          snippetAmount: this.snippetAmount,
          resolution: this.resolution,
          extraSnippet: this.extraSnippet
        }
      })
    },
    async regenerateTestVideo () {
      this.$store.commit('optionsPreviews/hidePreview')
      await api.post('/options/previews/test', {
        json: {
          startTime: this.startTime,
          snippetLength: this.snippetLength,
          snippetAmount: this.snippetAmount,
          resolution: this.resolution,
          extraSnippet: this.extraSnippet,
          regenerate: true
        }
      })
    },
    async startGenerating () {
      await api.get('/task/preview/generate')
    },
    async stopGenerating () {
      await api.get('/task/preview/stop')
    },
    prettyBytes
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

  .slider-row {
    align-items: center;
  }

  .slider-value {
    color: var(--xbvr-text-muted, #64708a);
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
  }

  .muted {
    color: var(--xbvr-text-muted, #64708a);
    font-size: 0.9rem;
  }

  .progress-text {
    margin-top: 0.5rem;
  }

  video {
    width: 100%;
    border-radius: var(--xbvr-radius, 12px);
    border: 1px solid var(--xbvr-border, #e3e6ec);
    box-shadow: var(--xbvr-shadow-sm, 0 1px 2px rgba(16, 24, 40, 0.05));
  }

  .preview-loading {
    display: flex;
    flex-wrap: wrap;
  }

  .bbox {
    flex: 1 0 calc(25% - 10px);
    margin: 5px;
    background: var(--xbvr-surface-sunken, #eef0f4);
    border: 1px solid var(--xbvr-border, #e3e6ec);
    border-radius: var(--xbvr-radius, 12px);
    color: var(--xbvr-text-faint, #7d88a1);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .bbox:after {
    content: '';
    display: block;
    padding-bottom: 100%;
  }

  .queue-progress {
    margin-top: 1em;
  }
</style>
