<template>
  <div>
    <div class="content">
      <header class="options-page-head">
        <h1 class="options-title">{{ $t("Funscripts") }}</h1>
        <p class="options-desc">{{ $t('Export scripts for players and control funscript scraping.') }}</p>
      </header>

      <div class="settings-card">
        <div class="settings-card-title">
          <b-icon pack="mdi" icon="information-outline" size="is-small"/>
          {{ $t("Export funscripts") }}
        </div>
        <p class="card-text">
          {{$t('Here you can download a ZIP file containing a funscript for each scripted scene. The file names include scene title and scene id, as expected by DeoVR. If a scene has multiple scripts you can choose a preferred script in the scene details view. Otherwise, the most recently added script is chosen.')}}
        </p>
        <p class="card-text">
          {{ $t("Note that the filenames are not compatible with DLNA.") }}
        </p>
        <p class="card-text">
          {{
            $t(
              "To use this export with DeoVR: Unzip and put the files in the Interactive folder on your device."
            )
          }}
        </p>
        <p class="card-text">
          {{
            $t(
              "To use this export with ScriptPlayer: Unzip and put the files in a folder of your choice. In the ScriptPlayer settings, add this folder in the Paths section, then connect to DeoVR."
            )
          }}
        </p>
      </div>

      <div class="settings-card">
        <div class="settings-card-title">
          <b-icon pack="mdi" icon="download-outline" size="is-small"/>
          {{ $t("Download funscripts for DeoVR") }}
        </div>
        <div class="button-row">
          <b-button
            type="is-primary"
            @click="exportAllFunscripts"
            :disabled="countTotal === 0"
            icon-left="download"
            >{{ $t("Download all funscripts") }} ({{ countTotal }})</b-button
          >
          <b-button
            type="is-primary"
            @click="exportNewFunscripts"
            :disabled="countUpdated === 0"
            icon-left="download"
            >{{ $t("Download changes since last export") }} ({{
              countUpdated
            }})</b-button
          >
        </div>
      </div>

      <div class="settings-card">
        <div class="settings-card-title">
          <b-icon pack="mdi" icon="cog-outline" size="is-small"/>
          {{ $t("Scraping") }}
        </div>
        <b-field>
          <b-switch v-model="scrapeFunscripts" type="is-default">
            <strong>Scrape for Available Funscripts</strong>
          </b-switch>
        </b-field>
        <div class="card-actions">
          <b-button type="is-primary" @click="save">Save</b-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import ky from "ky";

export default {
  name: "Funscripts",
  mounted() {
    this.$store.dispatch("optionsFunscripts/load");
  },
  methods: {
    exportAllFunscripts() {
      const link = document.createElement("a");
      link.href = "/api/task/funscript/export-all";
      link.click();
    },
    exportNewFunscripts() {
      const link = document.createElement("a");
      link.href = "/api/task/funscript/export-new";
      link.click();
    },
    save () {
      this.$store.dispatch('optionsFunscripts/save')
    },
  },
  computed: {
    countTotal: function () {
      return this.$store.state.optionsFunscripts.countTotal;
    },
    countUpdated: function () {
      return this.$store.state.optionsFunscripts.countUpdated;
    },
    scrapeFunscripts: {
      get () {
        return this.$store.state.optionsFunscripts.optionsFunscripts.scrapeFunscripts
      },
      set (value) {
        this.$store.state.optionsFunscripts.optionsFunscripts.scrapeFunscripts = value
      },
    },
  },
};
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

.card-text {
  color: var(--xbvr-text-muted, #64708a);
  font-size: 0.9rem;
  margin-bottom: 0.6rem;
}

.card-text:last-child {
  margin-bottom: 0;
}

.button-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem;
}

.card-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
  margin-top: 1rem;
}
</style>
