<template>
  <div>
    <b-loading :is-full-page="false" :active.sync="isLoading"></b-loading>
    <div class="content">
      <header class="options-page-head">
        <h1 class="options-title">{{$t("Task Schedules")}}</h1>
        <p class="options-desc">{{ $t('Automate recurring scrapes, rescans and preview generation.') }}</p>
      </header>
      <b-tabs v-model="activeTab" size="medium" type="is-boxed" id="importexporttab" class="options-tabs">
            <b-tab-item label="Rescrape"/>
            <b-tab-item label="Rescan"/>
            <b-tab-item label="Preview Generation"/>
            <b-tab-item label="Actor Rescrape"/>
            <b-tab-item label="Stashdb Rescrape"/>
            <b-tab-item :label="$t('Link Scenes')"/>
      </b-tabs>
      <div class="columns">
        <div class="column">
          <section class="settings-card">
            <div v-if="activeTab == 0">
              <div class="settings-card-title">
                <b-icon pack="mdi" icon="cloud-refresh" size="is-small"/>
                {{$t("Scrape Sites")}}
                <span class="schedule-chip" :class="{ 'is-on': rescrapeEnabled }">{{ rescrapeEnabled ? $t('Enabled') : $t('Disabled') }}</span>
              </div>
              <b-field>
                <b-switch v-model="rescrapeEnabled">Enable schedule</b-switch>
              </b-field>
              <b-field v-if="rescrapeEnabled">
                <b-slider v-model="rescrapeHourInterval" :min="1" :max="23" :step="1" ></b-slider>
                <div class="column is-one-third slider-value">{{`Run every ${this.rescrapeHourInterval} hour${this.rescrapeHourInterval > 1 ? 's': ''}`}}</div>
              </b-field>
              <b-field>
                <b-switch v-if="rescrapeEnabled" v-model="useRescrapeTimeRange">Limit time of day</b-switch>
              </b-field>
              <div v-if="useRescrapeTimeRange && rescrapeEnabled">
                <b-field>
                  <b-slider v-model="rescrapeTimeRange" :min="0" :max="48" :step="1" :custom-formatter="val => timeRange[val]" @input="restrictRescrapTo24Hours">
                    <b-slider-tick :value="0">00:00</b-slider-tick>
                    <b-slider-tick :value="6">06:00</b-slider-tick>
                    <b-slider-tick :value="12">12:00</b-slider-tick>
                    <b-slider-tick :value="18">18:00</b-slider-tick>
                    <b-slider-tick :value="24">Midnight</b-slider-tick>
                    <b-slider-tick :value="30">06:00</b-slider-tick>
                    <b-slider-tick :value="36">12:00</b-slider-tick>
                    <b-slider-tick :value="42">18:00</b-slider-tick>
                    <b-slider-tick :value="48">00:00</b-slider-tick>
                  </b-slider>
                  <div class="column is-one-third slider-value">{{`${this.timeRange[this.rescrapeTimeRange[0]]} - ${this.timeRange[this.rescrapeTimeRange[1]]}`}}</div>
                </b-field>
                <b-field>
                  <b-slider v-model="rescrapeMinuteStart" :min="0" :max="60" :step="1" ></b-slider>
                  <div class="column is-one-third slider-value">{{ minutesStartMsg(rescrapeMinuteStart) }}</div>
                </b-field>
              </div>
              
              <b-field label="Startup" class="startup-field">
                  <b-slider v-model="rescrapeStartDelay" :min="0" :max="60" :step="1" ></b-slider>
                  <div class="column is-one-third slider-value">{{ delayStartMsg(rescrapeStartDelay) }}</div>
              </b-field>
            </div>
            <div v-if="activeTab == 1">
              <div class="settings-card-title">
                <b-icon pack="mdi" icon="folder-refresh-outline" size="is-small"/>
                {{$t("Rescan Folders")}}
                <span class="schedule-chip" :class="{ 'is-on': rescanEnabled }">{{ rescanEnabled ? $t('Enabled') : $t('Disabled') }}</span>
              </div>
              <b-field>
                <b-switch v-model="rescanEnabled">Enable schedule</b-switch>
              </b-field>
              <b-field v-if="rescanEnabled">
                <b-slider v-model="rescanHourInterval" :min="1" :max="23" :step="1" ></b-slider>
                <div class="column is-one-third slider-value">{{`Run every ${this.rescanHourInterval} hour${this.rescanHourInterval > 1 ? 's': ''}`}}</div>
              </b-field>
              <b-field>
                <b-switch v-if="rescanEnabled" v-model="useRescanTimeRange">Limit time of day</b-switch>
              </b-field>
              <div v-if="useRescanTimeRange && rescanEnabled">
                <b-field>
                  <b-slider v-model="rescanTimeRange" :min="0" :max="48" :step="1" :custom-formatter="val => timeRange[val]" @input="restrictRescanTo24Hours">
                    <b-slider-tick :value="0">00:00</b-slider-tick>
                    <b-slider-tick :value="6">06:00</b-slider-tick>
                    <b-slider-tick :value="12">12:00</b-slider-tick>
                    <b-slider-tick :value="18">18:00</b-slider-tick>
                    <b-slider-tick :value="24">Midnight</b-slider-tick>
                    <b-slider-tick :value="30">06:00</b-slider-tick>
                    <b-slider-tick :value="36">12:00</b-slider-tick>
                    <b-slider-tick :value="42">18:00</b-slider-tick>
                    <b-slider-tick :value="48">00:00</b-slider-tick>
                  </b-slider>
                  <div class="column is-one-third slider-value">{{`${this.timeRange[this.rescanTimeRange[0]]} - ${this.timeRange[this.rescanTimeRange[1]]}`}}</div>
                </b-field>
                <b-field>
                  <b-slider v-model="rescanMinuteStart" :min="0" :max="60" :step="1" ></b-slider>
                  <div class="column is-one-third slider-value">{{ minutesStartMsg(rescanMinuteStart) }}</div>
                </b-field>
              </div>
              
              <b-field label="Startup" class="startup-field">
                  <b-slider v-model="rescanStartDelay" :min="0" :max="60" :step="1" ></b-slider>
                  <div class="column is-one-third slider-value">{{ delayStartMsg(rescanStartDelay) }}</div>
              </b-field>
            </div>
            <div v-if="activeTab == 2">
              <div class="settings-card-title">
                <b-icon pack="mdi" icon="motion-play-outline" size="is-small"/>
                {{$t("Preview Generation")}}
                <span class="schedule-chip" :class="{ 'is-on': previewEnabled }">{{ previewEnabled ? $t('Enabled') : $t('Disabled') }}</span>
              </div>
              <b-field>
                <b-switch v-model="previewEnabled">Enable schedule</b-switch>
              </b-field>
              <b-field v-if="previewEnabled">
                <b-slider v-model="previewHourInterval" :min="1" :max="23" :step="1" ></b-slider>
                <div class="column is-one-third slider-value">{{`Run every ${this.previewHourInterval} hour${this.previewHourInterval > 1 ? 's': ''}`}}</div>
              </b-field>
              <b-field>
                <b-switch v-if="previewEnabled" v-model="usePreviewTimeRange">Limit time of day</b-switch>
              </b-field>
              <div v-if="usePreviewTimeRange && previewEnabled">
                <b-field>
                  <b-slider v-model="previewTimeRange" :min="0" :max="48" :step="1" :custom-formatter="val => timeRange[val]" @input="restrictPreviewTo24Hours">
                    <b-slider-tick :value="0">00:00</b-slider-tick>
                    <b-slider-tick :value="6">06:00</b-slider-tick>
                    <b-slider-tick :value="12">12:00</b-slider-tick>
                    <b-slider-tick :value="18">18:00</b-slider-tick>
                    <b-slider-tick :value="24">Midnight</b-slider-tick>
                    <b-slider-tick :value="30">06:00</b-slider-tick>
                    <b-slider-tick :value="36">12:00</b-slider-tick>
                    <b-slider-tick :value="42">18:00</b-slider-tick>
                    <b-slider-tick :value="48">00:00</b-slider-tick>
                  </b-slider>
                  <div class="column is-one-third slider-value">{{`${this.timeRange[this.previewTimeRange[0]]} - ${this.timeRange[this.previewTimeRange[1]]}`}}</div>
                </b-field>
                <b-field>
                  <b-slider v-model="previewMinuteStart" :min="0" :max="60" :step="1" ></b-slider>
                  <div class="column is-one-third slider-value">{{ minutesStartMsg(previewMinuteStart) }}</div>
                </b-field>
                <p class="muted">
                  Preview Generation of a scene will not start after the Time Window Ends
                </p>
              </div>
              
              <b-field label="Startup" class="startup-field">
                  <b-slider v-model="previewStartDelay" :min="0" :max="60" :step="1" ></b-slider>
                  <div class="column is-one-third slider-value">{{ delayStartMsg(previewStartDelay) }}</div>
              </b-field>
              <p class="muted">
                BETA NOTE: Please note this is CPU-heavy process, if approriate limit the Time of Day the task runs
              </p>                  
            </div>
            <div v-if="activeTab == 3">
              <div class="settings-card-title">
                <b-icon pack="mdi" icon="account-refresh-outline" size="is-small"/>
                {{$t("Actor Rescrape")}}
                <span class="schedule-chip" :class="{ 'is-on': actorRescrapeEnabled }">{{ actorRescrapeEnabled ? $t('Enabled') : $t('Disabled') }}</span>
              </div>
              <b-field>
                <b-switch v-model="actorRescrapeEnabled">Enable schedule</b-switch>
              </b-field>
              <b-field v-if="actorRescrapeEnabled">
                <b-slider v-model="actorRescrapeHourInterval" :min="1" :max="23" :step="1" ></b-slider>
                <div class="column is-one-third slider-value">{{`Run every ${this.actorRescrapeHourInterval} hour${this.actorRescrapeHourInterval > 1 ? 's': ''}`}}</div>
              </b-field>
              <b-field>
                <b-switch v-if="actorRescrapeEnabled" v-model="useActorRescrapeTimeRange">Limit time of day</b-switch>
              </b-field>
              <div v-if="useActorRescrapeTimeRange && actorRescrapeEnabled">
                <b-field>
                  <b-slider v-model="actorRescrapeTimeRange" :min="0" :max="48" :step="1" :custom-formatter="val => timeRange[val]" @input="restrictActorRescrapeTo24Hours">
                    <b-slider-tick :value="0">00:00</b-slider-tick>
                    <b-slider-tick :value="6">06:00</b-slider-tick>
                    <b-slider-tick :value="12">12:00</b-slider-tick>
                    <b-slider-tick :value="18">18:00</b-slider-tick>
                    <b-slider-tick :value="24">Midnight</b-slider-tick>
                    <b-slider-tick :value="30">06:00</b-slider-tick>
                    <b-slider-tick :value="36">12:00</b-slider-tick>
                    <b-slider-tick :value="42">18:00</b-slider-tick>
                    <b-slider-tick :value="48">00:00</b-slider-tick>
                  </b-slider>
                  <div class="column is-one-third slider-value">{{`${this.timeRange[this.actorRescrapeTimeRange[0]]} - ${this.timeRange[this.actorRescrapeTimeRange[1]]}`}}</div>
                </b-field>
                <b-field>
                  <b-slider v-model="actorRescrapeMinuteStart" :min="0" :max="60" :step="1" ></b-slider>
                  <div class="column is-one-third slider-value">{{ minutesStartMsg(actorRescrapeMinuteStart) }}</div>
                </b-field>
              </div>
              
              <b-field label="Startup" class="startup-field">
                  <b-slider v-model="actorRescrapeStartDelay" :min="0" :max="60" :step="1" ></b-slider>
                  <div class="column is-one-third slider-value">{{ delayStartMsg(actorRescrapeStartDelay) }}</div>
              </b-field>
            </div>
            <div v-if="activeTab == 4">
              <div class="settings-card-title">
                <b-icon pack="mdi" icon="database-refresh-outline" size="is-small"/>
                {{$t("StashDB Rescrape")}}
                <span class="schedule-chip" :class="{ 'is-on': stashdbRescrapeEnabled }">{{ stashdbRescrapeEnabled ? $t('Enabled') : $t('Disabled') }}</span>
              </div>
              <b-field>
                <b-tooltip :active="stashApiKey==''" :label="$t('Enter a StashApi key to enable')" >
                  <b-switch v-model="stashdbRescrapeEnabled" :disabled="stashApiKey==''">Enable schedule</b-switch>
                </b-tooltip>
              </b-field>
              <b-field v-if="stashdbRescrapeEnabled">
                <b-slider v-model="stashdbRescrapeHourInterval" :min="1" :max="23" :step="1" ></b-slider>
                <div class="column is-one-third slider-value">{{`Run every ${this.stashdbRescrapeHourInterval} hour${this.stashdbRescrapeHourInterval > 1 ? 's': ''}`}}</div>
              </b-field>
              <b-field>
                <b-switch v-if="stashdbRescrapeEnabled" v-model="useStashdbRescrapeTimeRange">Limit time of day</b-switch>
              </b-field>
              <div v-if="useStashdbRescrapeTimeRange && stashdbRescrapeEnabled">
                <b-field>
                  <b-slider v-model="stashdbRescrapeTimeRange" :min="0" :max="48" :step="1" :custom-formatter="val => timeRange[val]" @input="restrictStashdbRescrapeTo24Hours">
                    <b-slider-tick :value="0">00:00</b-slider-tick>
                    <b-slider-tick :value="6">06:00</b-slider-tick>
                    <b-slider-tick :value="12">12:00</b-slider-tick>
                    <b-slider-tick :value="18">18:00</b-slider-tick>
                    <b-slider-tick :value="24">Midnight</b-slider-tick>
                    <b-slider-tick :value="30">06:00</b-slider-tick>
                    <b-slider-tick :value="36">12:00</b-slider-tick>
                    <b-slider-tick :value="42">18:00</b-slider-tick>
                    <b-slider-tick :value="48">00:00</b-slider-tick>
                  </b-slider>
                  <div class="column is-one-third slider-value">{{`${this.timeRange[this.stashdbRescrapeTimeRange[0]]} - ${this.timeRange[this.stashdbRescrapeTimeRange[1]]}`}}</div>
                </b-field>
                <b-field>
                  <b-slider v-model="stashdbRescrapeMinuteStart" :min="0" :max="60" :step="1" ></b-slider>
                  <div class="column is-one-third slider-value">{{ minutesStartMsg(stashdbRescrapeMinuteStart) }}</div>
                </b-field>
              </div>
              
              <b-field label="Startup" class="startup-field">
                  <b-slider v-model="stashdbRescrapeStartDelay" :min="0" :max="60" :step="1" ></b-slider>
                  <div class="column is-one-third slider-value">{{ delayStartMsg(stashdbRescrapeStartDelay) }}</div>
              </b-field>
            </div>
            <div v-if="activeTab == 5">
            <div class="settings-card-title">
              <b-icon pack="mdi" icon="link-variant" size="is-small"/>
              {{$t("Link Scenes")}}
              <span class="schedule-chip" :class="{ 'is-on': linkScenesEnabled }">{{ linkScenesEnabled ? $t('Enabled') : $t('Disabled') }}</span>
            </div>
            <b-field>
              <b-switch v-model="linkScenesEnabled">Enable schedule</b-switch>
            </b-field>
            <b-field v-if="linkScenesEnabled">
              <b-slider v-model="linkScenesHourInterval" :min="1" :max="23" :step="1" ></b-slider>
              <div class="column is-one-third slider-value">{{`Run every ${this.linkScenesHourInterval} hour${this.linkScenesHourInterval > 1 ? 's': ''}`}}</div>
            </b-field>
            <b-field>
              <b-switch v-if="linkScenesEnabled" v-model="useLinkScenesTimeRange">Limit time of day</b-switch>
            </b-field>
            <div v-if="useLinkScenesTimeRange && linkScenesEnabled">
              <b-field>
                <b-slider v-model="linkScenesTimeRange" :min="0" :max="48" :step="1" :custom-formatter="val => timeRange[val]" @input="restrictLinkScenesTo24Hours">
                  <b-slider-tick :value="0">00:00</b-slider-tick>
                  <b-slider-tick :value="6">06:00</b-slider-tick>
                  <b-slider-tick :value="12">12:00</b-slider-tick>
                  <b-slider-tick :value="18">18:00</b-slider-tick>
                  <b-slider-tick :value="24">Midnight</b-slider-tick>
                  <b-slider-tick :value="30">06:00</b-slider-tick>
                  <b-slider-tick :value="36">12:00</b-slider-tick>
                  <b-slider-tick :value="42">18:00</b-slider-tick>
                  <b-slider-tick :value="48">00:00</b-slider-tick>
                </b-slider>
                <div class="column is-one-third slider-value">{{`${this.timeRange[this.linkScenesTimeRange[0]]} - ${this.timeRange[this.linkScenesTimeRange[1]]}`}}</div>
              </b-field>
              <b-field>
                <b-slider v-model="linkScenesMinuteStart" :min="0" :max="60" :step="1" ></b-slider>
                <div class="column is-one-third slider-value">{{ minutesStartMsg(linkScenesMinuteStart) }}</div>
              </b-field>
              <p class="muted">
                Linking Scenes will not start after the Time Window Ends
              </p>
            </div>
            
            <b-field label="Startup" class="startup-field">
                <b-slider v-model="linkScenesStartDelay" :min="0" :max="60" :step="1" ></b-slider>
                <div class="column is-one-third slider-value">{{ delayStartMsg(linkScenesStartDelay) }}</div>
            </b-field>
          </div>
            <hr class="card-divider"/>
              <b-field grouped>
                <b-button type="is-primary" @click="saveSettings">Save settings</b-button>
              </b-field>
            <p class="muted restart-note">
              {{ $t('Restart XBVR to use new schedule settings') }}
            </p>
          </section>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import api from '../../../api'
import prettyBytes from 'pretty-bytes'

export default {
  name: 'Schedules',
  data () {
    return {
      isLoading: true,
      activeTab: 0,
      rescrapeEnabled: true,
      rescanEnabled: true,
      rescrapeTimeRange: [0, 23],
      lastTimeRange: [0, 23],
      rescanTimeRange: [0, 23],
      lastrescanTimeRange: [0, 23],
      useRescrapeTimeRange: false,
      useRescanTimeRange: false,
      rescrapeHourInterval: 0,
      rescrapeMinuteStart: 0,
      rescanMinuteStart: 0,
      rescanHourInterval: 0,
      previewEnabled: false,
      previewTimeRange:[0,23],
      previewHourInterval: 0,
      previewMinuteStart: 0,
      lastPreviewTimeRange: [0,23],
      usePreviewTimeRange: false,      
      rescrapeStartDelay: 0,
      rescanStartDelay: 0,
      previewStartDelay: 0,
      actorRescrapeEnabled: false,
      actorRescrapeTimeRange:[0,23],
      actorRescrapeHourInterval: 0,
      actorRescrapeMinuteStart: 0,
      lastActorRescrapeTimeRange: [0,23],
      useActorRescrapeTimeRange: false,      
      actorRescrapeStartDelay: 0,
      stashdbRescrapeEnabled: false,
      stashdbRescrapeTimeRange:[0,23],
      stashdbRescrapeHourInterval: 0,
      stashdbRescrapeMinuteStart: 0,
      lastStashdbRescrapeTimeRange: [0,23],
      useStashdbRescrapeTimeRange: false,      
      stashdbRescrapeStartDelay: 0,
      linkScenesEnabled: false,
      linkScenesTimeRange:[0,23],
      linkScenesHourInterval: 0,
      linkScenesMinuteStart: 0,
      lastlinkScenesTimeRange: [0,23],
      useLinkScenesTimeRange: false,      
      linkScenesStartDelay: 0,
      timeRange: ['00:00', '01:00', '02:00', '03:00', '04:00', '05:00', '06:00', '07:00', '08:00', '09:00', '10:00', '11:00',
        '12:00', '13:00', '14:00', '15:00', '16:00', '17:00', '18:00', '19:00', '20:00', '21:00', '22:00', '23:00',
        '00:00', '01:00', '02:00', '03:00', '04:00', '05:00', '06:00', '07:00', '08:00', '09:00', '10:00', '11:00',
        '12:00', '13:00', '14:00', '15:00', '16:00', '17:00', '18:00', '19:00', '20:00', '21:00', '22:00', '23:00', '00:00']
    }
  },
  async mounted () {
    await this.loadState()
  },
  computed: {
    stashApiKey: {
      get () {        
        return this.$store.state.optionsAdvanced.advanced.stashApiKey
      },
      set (value) {
        this.$store.state.optionsAdvanced.advanced.stashApiKey = value

      }
    },
  },
  methods: {
    restrictRescrapTo24Hours () {
      this.rescrapeTimeRange = this.restrictTo24Hours(this.rescrapeTimeRange, this.lastTimeRange)
      this.lastTimeRange = this.rescrapeTimeRange
    },
    restrictRescanTo24Hours () {
      this.rescanTimeRange = this.restrictTo24Hours(this.rescanTimeRange, this.lastrescanTimeRange)
      this.lastrescanTimeRange = this.rescanTimeRange
    },
    restrictPreviewTo24Hours () {
      this.previewTimeRange = this.restrictTo24Hours(this.previewTimeRange, this.lastPreviewTimeRange)
      this.lastPreviewTimeRange = this.previewTimeRange
    },
    restrictActorRescrapeTo24Hours () {
      this.actorRescrapeTimeRange = this.restrictTo24Hours(this.actorRescrapeTimeRange, this.lastActorRescrapeTimeRange)
      this.lastActorRescrapeTimeRange = this.actorRescrapeTimeRange
    },
    restrictStashdbRescrapeTo24Hours () {
      this.stashdbRescrapeTimeRange = this.restrictTo24Hours(this.stashdbRescrapeTimeRange, this.lastStashdbRescrapeTimeRange)
      this.lastStashdbRescrapeTimeRange = this.stashdbRescrapeTimeRange
    },
    restrictLinkScenesTo24Hours () {
      this.linkScenesTimeRange = this.restrictTo24Hours(this.linkScenesTimeRange, this.lastLinkScenesTimeRange)
      this.lastLinkScenesTimeRange = this.LinkScenesTimeRange
    },
    restrictTo24Hours (timeRange, lastTimeRange) {
      // check the first time is not in the second 24 hours, no need, should be in the first 24 hours
      if (timeRange[0] > 23) {
        timeRange[0] = 23
        timeRange = [timeRange[0], timeRange[1]]
      }
      // check they are not trying to select more than a 24 hour range
      if ((timeRange[1] - timeRange[0]) > 23 ) {
        if (timeRange[0] === lastTimeRange[0] || timeRange[0] === lastTimeRange[1]) {
          timeRange = [timeRange[1] - 23, timeRange[1]]
        } else {
          timeRange = [timeRange[0], timeRange[0] + 23]
        }
      }
      return timeRange
    },
    async loadState () {
      this.isLoading = true
      await api.get('/options/state')
        .json()
        .then(data => {
          this.rescrapeEnabled = data.config.cron.rescrapeSchedule.enabled
          this.rescrapeHourInterval = data.config.cron.rescrapeSchedule.hourInterval
          this.useRescrapeTimeRange = data.config.cron.rescrapeSchedule.useRange
          this.rescrapeMinuteStart = data.config.cron.rescrapeSchedule.minuteStart
          this.rescanEnabled = data.config.cron.rescanSchedule.enabled
          this.rescanHourInterval = data.config.cron.rescanSchedule.hourInterval
          this.useRescanTimeRange = data.config.cron.rescanSchedule.useRange
          this.rescanMinuteStart = data.config.cron.rescanSchedule.minuteStart
          this.previewEnabled = data.config.cron.previewSchedule.enabled
          this.previewHourInterval = data.config.cron.previewSchedule.hourInterval
          this.usePreviewTimeRange = data.config.cron.previewSchedule.useRange
          this.previewMinuteStart = data.config.cron.previewSchedule.minuteStart
          this.actorRescrapeEnabled = data.config.cron.actorRescrapeSchedule.enabled
          this.actorRescrapeHourInterval = data.config.cron.actorRescrapeSchedule.hourInterval
          this.useActorRescrapeTimeRange = data.config.cron.actorRescrapeSchedule.useRange
          this.actorRescrapeMinuteStart = data.config.cron.actorRescrapeSchedule.minuteStart          
          this.stashdbRescrapeEnabled = data.config.cron.stashdbRescrapeSchedule.enabled
          this.stashdbRescrapeHourInterval = data.config.cron.stashdbRescrapeSchedule.hourInterval
          this.useStashdbRescrapeTimeRange = data.config.cron.stashdbRescrapeSchedule.useRange
          this.stashdbRescrapeMinuteStart = data.config.cron.stashdbRescrapeSchedule.minuteStart          
          this.linkScenesEnabled = data.config.cron.linkScenesSchedule.enabled
          this.linkScenesHourInterval = data.config.cron.linkScenesSchedule.hourInterval
          this.useLinkScenesTimeRange = data.config.cron.linkScenesSchedule.useRange
          this.linkScenesMinuteStart = data.config.cron.linkScenesSchedule.minuteStart          
          if (data.config.cron.rescrapeSchedule.hourStart > data.config.cron.rescrapeSchedule.hourEnd) {
            this.rescrapeTimeRange = [data.config.cron.rescrapeSchedule.hourStart, data.config.cron.rescrapeSchedule.hourEnd + 24]
          } else {
            this.rescrapeTimeRange = [data.config.cron.rescrapeSchedule.hourStart, data.config.cron.rescrapeSchedule.hourEnd]
          }
          if (data.config.cron.rescanSchedule.hourStart > data.config.cron.rescanSchedule.hourEnd) {
            this.rescanTimeRange = [data.config.cron.rescanSchedule.hourStart, data.config.cron.rescanSchedule.hourEnd + 24]
          } else {
            this.rescanTimeRange = [data.config.cron.rescanSchedule.hourStart, data.config.cron.rescanSchedule.hourEnd]
          }
          if (data.config.cron.previewSchedule.hourStart > data.config.cron.previewSchedule.hourEnd) {
            this.previewTimeRange = [data.config.cron.previewSchedule.hourStart, data.config.cron.previewSchedule.hourEnd + 24]
          } else {
            this.previewTimeRange = [data.config.cron.previewSchedule.hourStart, data.config.cron.previewSchedule.hourEnd]            
          }
          if (data.config.cron.actorRescrapeSchedule.hourStart > data.config.cron.actorRescrapeSchedule.hourEnd) {
            this.actorRescrapeTimeRange = [data.config.cron.actorRescrapeSchedule.hourStart, data.config.cron.actorRescrapeSchedule.hourEnd + 24]
          } else {
            this.actorRescrapeTimeRange = [data.config.cron.actorRescrapeSchedule.hourStart, data.config.cron.actorRescrapeSchedule.hourEnd]            
          }
          
          if (data.config.cron.stashdbRescrapeSchedule.hourStart > data.config.cron.stashdbRescrapeSchedule.hourEnd) {
            this.stashdbRescrapeTimeRange = [data.config.cron.stashdbRescrapeSchedule.hourStart, data.config.cron.stashdbRescrapeSchedule.hourEnd + 24]
          } else {
            this.stashdbRescrapeTimeRange = [data.config.cron.stashdbRescrapeSchedule.hourStart, data.config.cron.stashdbRescrapeSchedule.hourEnd]            
          }

          if (data.config.cron.linkScenesSchedule.hourStart > data.config.cron.linkScenesSchedule.hourEnd) {
            this.linkScenesTimeRange = [data.config.cron.linkScenesSchedule.hourStart, data.config.cron.linkScenesSchedule.hourEnd + 24]
          } else {
            this.linkScenesTimeRange = [data.config.cron.linkScenesSchedule.hourStart, data.config.cron.linkScenesSchedule.hourEnd]            
          }
          
          this.rescrapeStartDelay = data.config.cron.rescrapeSchedule.runAtStartDelay
          this.rescanStartDelay = data.config.cron.rescanSchedule.runAtStartDelay          
          this.previewStartDelay = data.config.cron.previewSchedule.runAtStartDelay
          this.actorRescrapeStartDelay = data.config.cron.actorRescrapeSchedule.runAtStartDelay          
          this.stashdbRescrapeStartDelay = data.config.cron.stashdbRescrapeSchedule.runAtStartDelay          
          this.linkScenesStartDelay = data.config.cron.linkScenesSchedule.runAtStartDelay          
          this.isLoading = false
        })
    },
    minutesStartMsg (start) {
      if (start === 0) {
        return 'Start on the hour'
      }
      if (start === 1) {
        return 'Start at 1 minute past the hour'
      }
      return `Start at ${start} minutes past the hour`
    },
    delayStartMsg (start) {
      if (start === 0) {
        return 'Do not run at statup'
      }else{
        if (start === 1) {
          return `Run at 1 minute after startup`
        }else{
          return `Run at ${start} minutes after startup`
        }
      }
    },
    async saveSettings () {
      this.isLoading = true
      await api.post('/options/task-schedule', {
        json: {
          rescrapeEnabled: this.rescrapeEnabled,
          rescrapeHourInterval: this.rescrapeHourInterval,
          rescrapeUseRange: this.useRescrapeTimeRange,
          rescrapeMinuteStart: this.rescrapeMinuteStart,
          rescrapeHourStart: this.rescrapeTimeRange[0],
          rescrapeHourEnd: this.rescrapeTimeRange[1],
          rescrapeStartDelay: this.rescrapeStartDelay,
          rescanEnabled: this.rescanEnabled,
          rescanHourInterval: this.rescanHourInterval,
          rescanUseRange: this.useRescanTimeRange,
          rescanMinuteStart: this.rescanMinuteStart,
          rescanHourStart: this.rescanTimeRange[0],
          rescanHourEnd: this.rescanTimeRange[1],
          rescanStartDelay: this.rescanStartDelay,
          previewEnabled: this.previewEnabled,
          previewHourInterval: this.previewHourInterval,
          previewUseRange: this.usePreviewTimeRange,
          previewMinuteStart: this.previewMinuteStart,
          previewHourStart: this.previewTimeRange[0],
          previewHourEnd: this.previewTimeRange[1],
          previewStartDelay:this.previewStartDelay,
          actorRescrapeEnabled: this.actorRescrapeEnabled,
          actorRescrapeHourInterval: this.actorRescrapeHourInterval,
          actorRescrapeUseRange: this.useActorRescrapeTimeRange,
          actorRescrapeMinuteStart: this.actorRescrapeMinuteStart,
          actorRescrapeHourStart: this.actorRescrapeTimeRange[0],
          actorRescrapeHourEnd: this.actorRescrapeTimeRange[1],
          actorRescrapeStartDelay:this.actorRescrapeStartDelay          ,
          stashdbRescrapeEnabled: this.stashdbRescrapeEnabled,
          stashdbRescrapeHourInterval: this.stashdbRescrapeHourInterval,
          stashdbRescrapeUseRange: this.useStashdbRescrapeTimeRange,
          stashdbRescrapeMinuteStart: this.stashdbRescrapeMinuteStart,
          stashdbRescrapeHourStart: this.stashdbRescrapeTimeRange[0],
          stashdbRescrapeHourEnd: this.stashdbRescrapeTimeRange[1],
          stashdbRescrapeStartDelay:this.stashdbRescrapeStartDelay,
          linkScenesEnabled: this.linkScenesEnabled,
          linkScenesHourInterval: this.linkScenesHourInterval,
          linkScenesUseRange: this.useLinkScenesTimeRange,
          linkScenesMinuteStart: this.linkScenesMinuteStart,
          linkScenesHourStart: this.linkScenesTimeRange[0],
          linkScenesHourEnd: this.linkScenesTimeRange[1],
          linkScenesStartDelay:this.linkScenesStartDelay          
        }
      })
        .json()
        .then(data => {
          this.isLoading = false
        })
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

.options-tabs {
  margin-bottom: 1rem;
}

.options-tabs :deep(ul[role="tablist"]) {
  margin-left: 0;
}

.options-tabs :deep(.tabs.is-boxed a) {
  border-radius: var(--xbvr-radius-sm, 8px) var(--xbvr-radius-sm, 8px) 0 0;
}

.options-tabs :deep(.tabs.is-boxed li.is-active a) {
  background: var(--xbvr-surface, #ffffff);
  border-color: var(--xbvr-border, #e3e6ec);
  border-bottom-color: transparent;
  color: var(--xbvr-primary, #4f46e5);
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

.schedule-chip {
  margin-left: auto;
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--xbvr-text-faint, #7d88a1);
  background: var(--xbvr-surface-sunken, #eef0f4);
  border: 1px solid var(--xbvr-border, #e3e6ec);
  border-radius: 999px;
  padding: 0.15em 0.7em;
}

.schedule-chip.is-on {
  color: var(--xbvr-primary-strong, #4338ca);
  background: var(--xbvr-primary-soft, #eef0fe);
  border-color: transparent;
}

.slider-value {
  margin-left: 0.75em;
  color: var(--xbvr-text-muted, #64708a);
  font-size: 0.85rem;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.card-divider {
  background-color: var(--xbvr-border, #e3e6ec);
  margin: 0.85rem 0;
}

.startup-field {
  margin-top: 1.25rem;
}

.muted {
  color: var(--xbvr-text-muted, #64708a);
  font-size: 0.85rem;
}

.restart-note {
  margin-top: 0.75rem;
  margin-bottom: 0;
}
</style>
