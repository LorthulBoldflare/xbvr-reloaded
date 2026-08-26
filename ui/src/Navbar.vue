<template>
  <b-navbar :fixed-top="true" class="app-navbar">
    <template slot="brand">
      <b-navbar-item tag="router-link" :to="{ path: './' }" class="brand">
        <span class="brand-mark" aria-hidden="true">
          <b-icon pack="mdi" icon="virtual-reality" size="is-small"/>
        </span>
        <span class="brand-name" translate="no">XBVR</span>
        <span v-if="currentVersion" class="brand-version">{{currentVersion}}</span>
      </b-navbar-item>
    </template>
    <template slot="start">
      <b-navbar-item tag="router-link" :to="{ path: './' }" class="nav-link">
        <b-icon pack="mdi" icon="movie-open-outline" size="is-small" aria-hidden="true"/>
        <span>{{$t('Scenes')}}</span>
      </b-navbar-item>
      <b-navbar-item tag="router-link" :to="{ path: './actors' }" class="nav-link">
        <b-icon pack="mdi" icon="account-multiple-outline" size="is-small" aria-hidden="true"/>
        <span>{{$t('Actors')}}</span>
      </b-navbar-item>
      <b-navbar-item tag="router-link" :to="{ path: './files' }" class="nav-link">
        <b-icon pack="mdi" icon="folder-outline" size="is-small" aria-hidden="true"/>
        <span>{{$t('Files')}}</span>
      </b-navbar-item>
      <b-navbar-item tag="router-link" :to="{ path: './options' }" class="nav-link">
        <b-icon pack="mdi" icon="cog-outline" size="is-small" aria-hidden="true"/>
        <span>{{$t('Options')}}</span>
      </b-navbar-item>
      <b-navbar-item @click="$store.commit('overlay/showQuickFind')" class="nav-link" role="button">
        <b-icon pack="mdi" icon="lightning-bolt-outline" size="is-small" aria-hidden="true"/>
        <span>{{$t('Quick find')}}</span>
        <kbd class="nav-kbd" aria-hidden="true">?</kbd>
      </b-navbar-item>
    </template>
    <template slot="end">
      <b-navbar-item class="theme-toggle-item">
        <button type="button" class="theme-toggle" @click="onToggleTheme"
                :aria-label="theme === 'dark' ? $t('Switch to light mode') : $t('Switch to dark mode')"
                :title="theme === 'dark' ? $t('Switch to light mode') : $t('Switch to dark mode')">
          <b-icon pack="mdi" :icon="theme === 'dark' ? 'white-balance-sunny' : 'weather-night'" size="is-small" aria-hidden="true"/>
        </button>
      </b-navbar-item>
      <b-navbar-item v-if="hasStatus" class="status-area" aria-live="polite">
        <div class="status-chips">
          <div v-if="Object.keys(lastRescanMessage).length !== 0" class="status-chip">
            <span class="status-dot" :class="{ 'is-active': lockRescan }" aria-hidden="true"></span>
            <strong>{{$t('Files')}}</strong>
            <span class="status-msg">{{lastRescanMessage.message}}</span>
          </div>
          <div v-if="Object.keys(lastScrapeMessage).length !== 0" class="status-chip">
            <span class="status-dot" :class="{ 'is-active': lockScrape }" aria-hidden="true"></span>
            <strong>{{$t('Data')}}</strong>
            <span class="status-msg">{{lastScrapeMessage.message}}</span>
          </div>
        </div>
      </b-navbar-item>
    </template>
  </b-navbar>
</template>

<script>
import api from './api'
import { currentTheme, toggleTheme } from './util/theme'

export default {
  data () {
    return {
      currentVersion: '',
      latestVersion: '',
      theme: currentTheme()
    }
  },
  computed: {
    lockRescan () {
      return this.$store.state.messages.lockRescan
    },
    lastRescanMessage () {
      return this.$store.state.messages.lastRescanMessage
    },
    lockScrape () {
      return this.$store.state.messages.lockScrape
    },
    lastScrapeMessage () {
      return this.$store.state.messages.lastScrapeMessage
    },
    hasStatus () {
      return Object.keys(this.lastRescanMessage).length !== 0 ||
        Object.keys(this.lastScrapeMessage).length !== 0
    }
  },
  methods: {
    onToggleTheme () {
      this.theme = toggleTheme()
    }
  },
  mounted () {
    api.get('/options/version-check').json().then(data => {
      this.currentVersion = data.current_version
      this.latestVersion = data.latest_version

      if (data.update_notify && this.currentVersion !== 'CURRENT') {
        this.$buefy.snackbar.open({
          message: `Version ${this.latestVersion} available!`,
          type: 'is-warning',
          position: 'is-bottom-right',
          actionText: this.$t('Download now'),
          indefinite: true,
          onAction: () => {
            window.location = 'https://github.com/LorthulBoldflare/xbvr-reloaded/releases'
          }
        })
      }
    })
  }
}
</script>

<style scoped>
  .brand {
    display: flex;
    align-items: center;
    gap: 0.55rem;
  }

  .brand:hover {
    background: transparent !important;
  }

  .brand-mark {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 2rem;
    height: 2rem;
    border-radius: 9px;
    color: #fff;
    background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 55%, #d946ef 100%);
    box-shadow: 0 2px 8px rgba(99, 102, 241, 0.4);
  }

  .brand-name {
    font-size: 1.2rem;
    font-weight: 800;
    letter-spacing: -0.02em;
    color: var(--xbvr-text, #1c2333);
  }

  .brand-version {
    font-size: 0.68rem;
    font-weight: 600;
    color: var(--xbvr-text-faint, #98a1b6);
    background: var(--xbvr-surface-sunken, #eef0f4);
    border-radius: 999px;
    padding: 0.15em 0.6em;
    margin-left: 0.15rem;
    align-self: center;
  }

  .nav-link {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    margin: 0 0.1rem;
  }

  .nav-kbd {
    font-family: inherit;
    font-size: 0.68rem;
    font-weight: 600;
    line-height: 1;
    color: var(--xbvr-text-faint, #98a1b6);
    background: var(--xbvr-surface-sunken, #eef0f4);
    border: 1px solid var(--xbvr-border, #e3e6ec);
    border-radius: 5px;
    padding: 0.25em 0.45em;
  }

  .theme-toggle-item {
    display: flex;
    align-items: center;
  }

  .theme-toggle {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 2.25rem;
    height: 2.25rem;
    padding: 0;
    border: 1px solid var(--xbvr-border, #e3e6ec);
    border-radius: 999px;
    background: var(--xbvr-surface, #ffffff);
    color: var(--xbvr-text-muted, #64708a);
    cursor: pointer;
    transition: color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
      border-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
      background-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
  }

  .theme-toggle:hover {
    color: var(--xbvr-primary, #4f46e5);
    border-color: var(--xbvr-primary, #4f46e5);
    background: var(--xbvr-primary-soft, #eef0fe);
  }

  .status-chips {
    display: flex;
    gap: 0.5rem;
    font-size: 0.8rem;
  }

  .status-chip {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    max-width: 22rem;
    padding: 0.3rem 0.7rem;
    border-radius: 999px;
    background: var(--xbvr-surface-sunken, #eef0f4);
    color: var(--xbvr-text-muted, #64708a);
  }

  .status-msg {
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .status-dot {
    flex: none;
    width: 8px;
    height: 8px;
    border-radius: 999px;
    background: var(--xbvr-border-strong, #cdd2dc);
  }

  .status-dot.is-active {
    background: var(--xbvr-primary, #4f46e5);
    animation: status-pulse 1.2s ease-in-out infinite;
  }

  @keyframes status-pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.35; }
  }

  @media (prefers-reduced-motion: reduce) {
    .status-dot.is-active {
      animation: none;
    }
  }
</style>
