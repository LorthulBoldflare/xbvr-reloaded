<template>
  <div class="modal is-active player-modal">
    <div class="modal-background"></div>
    <div class="modal-content player-content">
      <div class="player-pane">
        <video ref="player"
               width="640" height="640" class="video-js vjs-default-skin"
               controls playsinline autoplay>
          <source :src="sourceUrl" type="video/mp4">
        </video>
        <div class="skip-row">
          <button type="button" class="skip-btn" @click="stepBack(10)">
            <b-icon pack="mdi" icon="rewind-10" size="is-small" aria-hidden="true"></b-icon>
            <span>10s</span>
          </button>
          <button type="button" class="skip-btn" @click="stepForward(10)">
            <b-icon pack="mdi" icon="fast-forward-10" size="is-small" aria-hidden="true"></b-icon>
            <span>10s</span>
          </button>
        </div>
      </div>
    </div>
    <button class="modal-close is-large" aria-label="close"
            @click="close()"></button>
  </div>
</template>

<script>
import videojs from 'video.js'
import vr from 'videojs-vr/dist/videojs-vr.min.js'
import hotkeys from 'videojs-hotkeys'

export default {
  name: 'Details',
  data () {
    return {
      player: {}
    }
  },
  computed: {
    sourceUrl () {
      if (this.$store.state.overlay.player.file) {
        return '/api/dms/file/' + this.$store.state.overlay.player.file.id + '?dnt=true'
      }
      return ''
    }
  },
  mounted () {
    this.player = videojs(this.$refs.player)
    const vr = this.player.vr({
      projection: this.$store.state.overlay.player.file.projection == 'flat' ? 'NONE' : '180',
      forceCardboard: false
    })

    this.player.hotkeys({
      alwaysCaptureHotkeys: true,
      volumeStep: 0.1,
      seekStep: 5,
      enableModifiersForNumbers: false,
      customKeys: {
        closeModal: {
          key: function (event) {
            return event.which === 27
          },
          handler: (player, options, event) => {
            this.player.dispose()
            this.$store.commit('overlay/hidePlayer')
          }
        }
      }
    })

    this.player.on('loadedmetadata', function () {
      vr.camera.position.set(-1, 0, -1)
    })
  },
  methods: {
    close () {
      this.player.dispose()
      this.$store.commit('overlay/hidePlayer')
    },
    stepBack (seconds) {
      if (this.player && typeof this.player.currentTime === 'function') {
        this.player.currentTime(Math.max(0, this.player.currentTime() - seconds))
      }
    },
    stepForward (seconds) {
      if (this.player && typeof this.player.currentTime === 'function') {
        this.player.currentTime(this.player.currentTime() + seconds)
      }
    }
  }
}
</script>

<style lang="less" scoped>
.player-content {
  width: min(1080px, 94vw);
}

/* media chrome — token-driven so it follows light/dark themes */
.player-pane {
  background: var(--xbvr-media-bg, #0d1017);
  border-radius: var(--xbvr-radius-lg, 16px);
  padding: 0.75rem 0.75rem 1rem;
  box-shadow: var(--xbvr-shadow-lg, 0 16px 40px rgba(16, 24, 40, 0.16));
}

.player-pane .video-js {
  margin: 0 auto;
  width: 100%;
  border-radius: var(--xbvr-radius-sm, 8px);
  overflow: hidden;
}

/* pill skip buttons under the player */
.skip-row {
  display: flex;
  justify-content: center;
  gap: 0.5rem;
  margin-top: 0.75rem;
}

.skip-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.3em 1.1em;
  border-radius: 999px;
  background: var(--xbvr-media-chip-bg, rgba(255, 255, 255, 0.08));
  border: 1px solid var(--xbvr-media-chip-border, rgba(255, 255, 255, 0.18));
  color: var(--xbvr-media-chip-text, #cdd6ea);
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: background-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
    color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
    border-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

.skip-btn:hover {
  background: var(--xbvr-media-tab-active-bg, rgba(255, 255, 255, 0.14));
  color: var(--xbvr-media-tab-active-text, #ffffff);
}

/* circular glass close control over the overlay */
.modal-close {
  background: var(--xbvr-chip-bg, rgba(20, 24, 36, 0.55));
  border-radius: 999px;
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
}

.modal-close::before,
.modal-close::after {
  background-color: #fff;
}
</style>
