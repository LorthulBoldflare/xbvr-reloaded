<template>
  <div>
    <b-loading :is-full-page="false" :active.sync="isLoading"></b-loading>
    <div class="content">
      <header class="options-page-head">
        <h1 class="options-title">{{$t("DLNA interface")}}</h1>
        <p class="options-desc">{{ $t('Broadcast your library to DLNA-capable players on the local network.') }}</p>
      </header>
      <div class="columns">
        <div class="column">
          <section class="settings-card">
            <div class="settings-card-title">
              <b-icon pack="mdi" icon="cast" size="is-small"/>
              {{ $t('DLNA server') }}
            </div>
            <b-field label="DLNA server">
              <b-switch v-model="enabled">
                Enabled
              </b-switch>
            </b-field>

            <b-field label="Visible name" class="name-field">
              <b-input v-model="name"></b-input>
            </b-field>

            <b-field grouped>
              <b-field label="Icon">
                <b-select placeholder="Select image" v-model="image">
                  <option v-for="s in dlnaOptions.availableImages" :value="s" :key="s.id">
                    {{ s }}
                  </option>
                </b-select>
              </b-field>
              <b-field label=" ">
                <img :src="`/ui/dlna/${image}.png`" width="64" class="icon-preview" v-if="image"/>
              </b-field>
            </b-field>

            <b-field label="Allowed IP addresses">
              <b-taginput v-model="allowedIp" :allow-new="true" placeholder="Type in a IP address" class="is-half"></b-taginput>
            </b-field>

            <b-field>
              <p v-if="!isLoading" class="recent-ip">
                {{ $t('Recent IP addresses:') }}
                <span v-if="dlnaOptions.recentIp.length > 0">
                  <b-tag rounded v-for="s in dlnaOptions.recentIp" :value="s" :key="s.id" class="ip-tag" type="is-info"><span @click="addIP(s)">{{ s }}</span></b-tag>
                </span>
                <span v-else class="muted">none (connect to DLNA at least once to find out device's IP address)</span>
              </p>
            </b-field>

            <div class="card-actions">
              <b-button type="is-primary" @click="save">Save and apply changes</b-button>
            </div>

          </section>
        </div>
        <div class="column content side-notes">
          <p>
            {{$t("Standard protocol that works with players such as: Skybox, Pigasus, Mobile Station VR, and others.")}}
          </p>
          <p>
            {{$t("Since it is broadcasted accross the whole local network, you might want to restrict access to selected IP addresses or disable it completely.")}}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'InterfaceDLNA',
  mounted () {
    this.$store.dispatch('optionsDLNA/load')
  },
  methods: {
    save () {
      this.$store.dispatch('optionsDLNA/save')
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
    }
  },
  computed: {
    enabled: {
      get () {
        return this.$store.state.optionsDLNA.dlna.enabled
      },
      set (value) {
        this.$store.state.optionsDLNA.dlna.enabled = value
      }
    },
    name: {
      get () {
        return this.$store.state.optionsDLNA.dlna.name
      },
      set (value) {
        this.$store.state.optionsDLNA.dlna.name = value
      }
    },
    image: {
      get () {
        return this.$store.state.optionsDLNA.dlna.image
      },
      set (value) {
        this.$store.state.optionsDLNA.dlna.image = value
      }
    },
    allowedIp: {
      get () {
        return this.$store.state.optionsDLNA.dlna.allowedIp
      },
      set (value) {
        this.$store.state.optionsDLNA.dlna.allowedIp = value
      }
    },
    isLoading: function () {
      return this.$store.state.optionsDLNA.loading
    },
    dlnaOptions: function () {
      return this.$store.state.optionsDLNA.dlna
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

.card-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
  margin-top: 1rem;
}

.settings-card-title .icon {
  color: var(--xbvr-text-faint, #7d88a1);
}

.name-field {
  max-width: 260px;
}

.icon-preview {
  margin-left: 1em;
  border-radius: var(--xbvr-radius-sm, 8px);
  border: 1px solid var(--xbvr-border, #e3e6ec);
}

.recent-ip {
  color: var(--xbvr-text-muted, #64708a);
  font-size: 0.85rem;
}

.ip-tag {
  margin-right: 0.25em;
  cursor: pointer;
  text-decoration: underline;
  transition: background-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

.side-notes p {
  color: var(--xbvr-text-muted, #64708a);
  font-size: 0.9rem;
  margin-bottom: 0.6rem;
}

.muted {
  color: var(--xbvr-text-faint, #7d88a1);
}
</style>
