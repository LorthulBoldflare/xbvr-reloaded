<template>
  <div class="modal is-active" role="dialog" aria-modal="true">
    <GlobalEvents
      :filter="e => !['INPUT', 'TEXTAREA'].includes(e.target.tagName)"
      @keyup.esc="close"
      @keydown.left="handleLeftArrow"
      @keydown.right="handleRightArrow"
      @keydown.o="prevActor"
      @keydown.p="nextActor"
      @keydown.f="$store.commit('actorList/toggleActorList', {actor_id: actor.id, list: 'favourite'})"
      @keydown.exact.w="$store.commit('actorList/toggleActorList', {actor_id: actor.id, list: 'watchlist'})"
      @keydown.e="$store.commit('overlay/editActorDetails', {actor: actor})"
      @keydown.s="$store.commit('overlay/showSearchStashdbActors', {actor: item})"
      @keydown.48="setRating(0)"
    />

    <div class="modal-background"></div>

    <div class="modal-card details-card">
      <section class="modal-card-body details-body">
        <div class="details-grid">

          <div class="media-pane">
            <b-tabs v-model="activeMedia" position="is-centered" :animated="false" class="media-tabs">
              <b-tab-item :label="$t('Gallery')">
                <b-carousel v-model="carouselSlide" @change="scrollToActiveIndicator" :autoplay="false" :indicator-inside="false">
                  <b-carousel-item v-for="(carousel, i) in images" :key="i">
                    <div class="image is-1by1 is-full"
                         v-bind:style="{backgroundImage: `url(${getImageURL(carousel, '700,fit', 'act-' + actor.id)})`, backgroundSize: 'contain', backgroundPosition: 'center', backgroundRepeat: 'no-repeat'}"></div>
                  </b-carousel-item>
                  <template slot="indicators" slot-scope="props">
                      <span class="al image indicator-thumb">
                        <vue-load-image>
                          <img slot="image" :src="getIndicatorURL(props.i)" class="indicator-img"/>
                          <img slot="preloader" :src="'/ui/images/blank.png'" class="indicator-img is-placeholder"/>
                          <img slot="error" src="/ui/images/blank_female_profile.png" class="indicator-img"/>
                        </vue-load-image>
                      </span>
                  </template>
                </b-carousel>
                <div class="media-actions">
                  <b-button class="button is-primary is-small" v-on:click="setActorImage()">{{$t('Set Main Image')}}</b-button>
                  <b-button v-if="images.length != 0" class="button is-danger is-outlined is-small" v-on:click="deleteActorImage()">{{$t('Delete Image')}}</b-button>
                </div>
              </b-tab-item>
            </b-tabs>
          </div>

          <div class="info-pane">

            <header class="details-header">
              <h2 class="details-title">
                <span>{{ actor.name }}</span>
                <span class="title-actions">
                  <b-tooltip position="is-right" :label="$t('Delete Aka Group')" multilined :delay="200" v-if="actor.name.startsWith('aka:')">
                    <button class="button is-small is-outlined" @click="deleteAkaGroup" >
                      <b-icon pack="mdi" icon="delete-outline"></b-icon>
                    </button>
                  </b-tooltip>
                  <b-tooltip v-if="enableNewAkaGroup()" position="is-right" :label="$t('Create a new Aka Group')" multilined :delay="200">
                    <button class="button is-small is-outlined" @click="createAkaGroup">
                      <b-icon pack="mdi" icon="account-multiple-plus-outline"></b-icon>
                    </button>
                  </b-tooltip>
                </span>
              </h2>
              <div class="meta-chips">
                <span v-if="actor.birth_date != '0001-01-01T00:00:00Z'" class="meta-chip">
                  <b-icon pack="mdi" icon="calendar-outline" size="is-small" aria-hidden="true"/><span>{{ format(parseISO(actor.birth_date), "yyyy-MM-dd") }}</span>
                </span>
                <span v-if="actor.birth_date != '0001-01-01T00:00:00Z'" class="meta-chip">
                  <b-icon pack="mdi" icon="cake-variant-outline" size="is-small" aria-hidden="true"/><span>{{ $t('Age') }}: {{ calcAge(actor.birth_date) }}</span>
                </span>
                <span v-if="actor.start_year + actor.end_year  != 0" class="meta-chip">
                  <b-icon pack="mdi" icon="calendar-range-outline" size="is-small" aria-hidden="true"/><span>{{ $t('Active') }}: {{ getYearsActive() }}</span>
                </span>
                <span class="meta-chip">
                  <b-icon pack="mdi" icon="filmstrip" size="is-small" aria-hidden="true"/><span>{{ $t('Scenes') }}: {{ actor.scenes.length }}</span>
                </span>
              </div>
            </header>

            <div class="details-toolbar">
              <b-field class="rating-block">
                <span class="rating-label">{{ $t('Your Rating') }}</span>
                <star-rating :key="actor.id" v-model="actor.star_rating" :rating="actor.star_rating" @rating-selected="setRating"
                             :increment="0.5" :star-size="20" :show-rating="true" />
                <b-tooltip :label="$t('Reset Rating')" position="is-right" :delay="250">
                  <b-icon pack="mdi" icon="autorenew" size="is-small" @click.native="setRating(0)" class="rating-reset"/>
                </b-tooltip>
              </b-field>
              <b-field class="rating-block">
                <span class="rating-label">{{ $t('Scene Average') }}</span>
                <star-rating :key="actor.id" :rating="Math.round(actor.scene_rating_average * 4) / 4" read-only :increment="0.25" :star-size="20" :show-rating="true" active-color="var(--xbvr-primary, #7957d5)"/>
              </b-field>
              <div class="action-row">
                <actor-favourite-button :actor="actor"/>
                <actor-watchlist-button :actor="actor"/>
                <actor-edit-button :actor="actor"/>
                <link-stashdb-button :item="actor" objectType="actor" />
              </div>
            </div>

            <div class="block-opts block">
              <b-tabs v-model="activeTab" :animated="false">
                <b-tab-item :label="$t('Details')">
                  <div class="attribute-container">
                    <b-field v-if="actor.birth_date != '0001-01-01T00:00:00Z'">
                      <strong class="attribute-heading">{{ $t('Age') }}:</strong><span class="attribute-data">{{ calcAge(actor.birth_date) }}</span>
                    </b-field>
                    <b-field v-if="actor.start_year + actor.end_year  != 0">
                      <strong class="attribute-heading">{{ $t('Active') }}:</strong><span class="attribute-data">{{ getYearsActive() }}</span>
                    </b-field>
                    <b-field v-if="actor.nationality">
                      <strong class="attribute-heading">{{ $t('Nationality') }}:</strong>
                      <b-field grouped class="attribute-data">
                        <vue-load-image>
                             <img slot="image" :src="getImageURL(this.getCountryFlag(actor.nationality), '700x', 'icon-' + String(actor.nationality).toLowerCase())" class="flag-img"/>
                        </vue-load-image>
                        <small>{{ this.getCountryName(actor.nationality) }}</small>
                      </b-field>
                    </b-field>
                    <b-field v-if="actor.ethnicity">
                      <strong class="attribute-heading">{{ $t('Ethnicity') }}:</strong><small  class="attribute-data">{{ actor.ethnicity }}</small>
                    </b-field>
                    <b-field v-if="actor.hair_color">
                      <strong class="attribute-heading">{{ $t('Hair Color') }}:</strong> <small class="attribute-data">{{ actor.hair_color }}</small>
                    </b-field>
                    <b-field v-if="actor.eye_color">
                      <strong class="attribute-heading">{{ $t('Eye Color') }}:</strong> <small class="attribute-data">{{ actor.eye_color }}</small>
                    </b-field>
                    <b-field v-if="actor.height">
                      <strong class="attribute-heading">{{ $t('Height') }}:</strong> <small class="attribute-data">{{ getHeight(actor.height) }}</small>
                    </b-field>
                    <b-field v-if="actor.weight">
                      <strong class="attribute-heading">{{ $t('Weight') }}:</strong> <small class="attribute-data">{{ getWeight(actor.weight) }}</small>
                    </b-field>
                    <b-field v-if="measurements() != ''">
                      <strong class="attribute-heading">{{ $t('Measurements') }}:</strong> <small class="attribute-data">{{ measurements() }}</small>
                    </b-field>
                    <b-field v-if="actor.breast_type != ''">
                      <strong class="attribute-heading">{{ $t('Breast Type') }}:</strong> <small class="attribute-data">{{ actor.breast_type }}</small>
                    </b-field>
                    <b-field v-if="actor.aliases != '' && actor.aliases != '[]'">
                      <strong class="attribute-heading">{{ $t('Aliases') }}:</strong> <small class="attribute-long-data">{{ joinArray(actor.aliases) }}</small>
                    </b-field>
                    <b-field v-if="actor.tattoos != '' && actor.tattoos != '[]'">
                      <strong class="attribute-heading">{{ $t('Tattoos') }}:</strong> <small class="attribute-long-data">{{ joinArray(actor.tattoos) }}</small>
                    </b-field>
                    <b-field v-if="actor.piercings != '' && actor.piercings != '[]'">
                      <strong class="attribute-heading">{{ $t('Piercings') }}:</strong> <small class="attribute-long-data">{{ joinArray(actor.piercings) }}</small>
                    </b-field>
                  </div>
                  <b-message  v-if="actor.biography != ''">
                      {{ actor.biography }}
                    </b-message>
                </b-tab-item>
                <b-tab-item>
                  <template #header>
                    Scenes ({{ actor.scenes.length }}) <a v-if="showOpenInNewWindow" :href='getCastScenesUrl([actor.name])' target="_blank" class="tab-link-icon"><b-icon pack="mdi" icon="open-in-new" size="is-small"></b-icon></a>
                  </template>
                  <div v-show="activeTab == 1" :class="['columns', 'is-multiline', actor.scenes.length > 6 ? 'scroll' : '']">
                    <div :class="['column', 'is-multiline', 'is-one-third']"
                      v-for="(scene, idx) in actor.scenes" :key="idx" class="image-wrapper">
                      <SceneCard :item="scene" :reRead=true />
                    </div>
                  </div>
                </b-tab-item>
                <b-tab-item :label="$t('Akas')" :visible="akas.aka_groups != null || akas.actors != null || akas.possible_akas != null">
                  <div v-show="activeTab == 2">
                    <b-field :label="$t('Aka Groups')" v-if="akas.aka_groups != null &&  akas.aka_groups.length!=0">
                      <div  class="columns is-multiline">
                        <div :class="['column', 'is-multiline', 'is-one-third']"
                          v-for="(akaactor, idx) in akas.aka_groups" :key="idx" class="image-wrapper">
                          <ActorCard :actor="akaactor"/>
                        </div>
                      </div>
                    </b-field>
                    <b-field :label="$t('Other Actors In Groups')" v-if="akas.actors != null &&  akas.actors.length!=0">
                      <div  class="columns is-multiline">
                        <div :class="['column', 'is-multiline', 'is-one-third']"
                          v-for="(akaactor, idx) in akas.actors" :key="idx" class="image-wrapper">
                          <ActorCard :actor="akaactor"/>
                          <b-tooltip position="is-bottom" :label="$t('Remove Cast from Aka Group. Select the Aka group and Actors to remove in the Cast Filter')" multilined :delay="200">
                            <button class="button is-small is-outlined" @click="removeFromAkaGroup(akaactor.name)" v-if="actor.name.startsWith('aka:')">
                              <b-icon pack="mdi" icon="account-minus-outline"></b-icon>
                            </button>
                          </b-tooltip>
                        </div>
                      </div>
                    </b-field>
                    <b-field :label="$t('Possible Matches')" v-if="akas.possible_akas != null &&  akas.possible_akas.length!=0">
                      <div class="columns is-multiline">
                        <div :class="['column', 'is-multiline', 'is-one-third']"
                          v-for="(akaactor, idx) in akas.possible_akas" :key="idx" class="image-wrapper">
                          <ActorCard :actor="akaactor"/>
                          <b-tooltip position="is-bottom" :label="$t('Add Cast to Aka Group. Select the Aka group and Actors to add in the Cast Filter')" multilined :delay="200">
                            <button class="button is-small is-outlined" @click="addToAkaGroup(akaactor.name)" v-if='actor.name.startsWith("aka:")'>
                              <b-icon pack="mdi" icon="account-plus-outline"></b-icon>
                            </button>
                          </b-tooltip>
                        </div>
                      </div>
                    </b-field>
                  </div>
                </b-tab-item>
                <b-tab-item :visible="colleagues.length != 0" :label="`Colleagues (${colleagues.length})`">
                  <div v-show="activeTab == 3" class="columns is-multiline scroll">
                    <div :class="['column', 'is-multiline', 'is-one-third']"
                      v-for="(colleague, idx) in colleagues" :key="idx" class="image-wrapper">
                      <ActorCard :actor="colleague" :colleague="actor.name" />
                    </div>
                  </div>
                </b-tab-item>
                <b-tab-item :label="`Links (${getActorUrls().length})`" v-show="getActorUrls().length !=0">
                  <div v-show="activeTab == 4">
                    <div >
                      <b-field :label="$t('Links')" >
                        <div class="link-list">
                          <div
                            v-for="(urllink, idx) in getActorUrls()" :key="idx">
                            <a class="tag is-info link-tag" :href="urllink.url" target="_blank" rel="noreferrer">{{urllink.url}}</a>
                          </div>
                        </div>
                      </b-field>
                    </div>
                  </div>
                </b-tab-item>
                <b-tab-item  :label="`Scrapers (${extrefs.length})`" v-show="extrefs.length !=0">
                  <div v-show="activeTab == 5">
                    <div >
                      <b-field :label="$t('Actor Scrapers')" >
                        <div class="link-list">
                          <div v-for="(extref, idx) in extrefs" :key="idx" class="scraper-row">
                              <a @click="refreshScraper(extref.external_reference.external_url)" :title="'Rescrape Actor Details now'" class="refresh-icon">
                                <b-icon pack="mdi" icon="refresh" size="is-small"/>
                              </a>
                            <a class="tag is-info link-tag" :href="extref.external_reference.external_url" target="_blank" rel="noreferrer">{{extref.external_source}} - Updated: {{format(parseISO(extref.external_reference.external_date), "yyyy-MM-dd") }}</a>
                          </div>
                        </div>
                      </b-field>
                    </div>
                  </div>
                </b-tab-item>
              </b-tabs>
            </div>

          </div>
        </div>
      </section>
    </div>
    <button class="modal-close is-large" aria-label="close" @click="close()"></button>
    <button type="button" class="actor-nav prev" @click="prevActor"
            title="Keyboard shortcut: O" aria-label="Previous actor">
      <b-icon pack="mdi" icon="chevron-left" size="is-medium" aria-hidden="true"/>
    </button>
    <button type="button" class="actor-nav next" @click="nextActor"
            title="Keyboard shortcut: P" aria-label="Next actor">
      <b-icon pack="mdi" icon="chevron-right" size="is-medium" aria-hidden="true"/>
    </button>
  </div>
</template>

<script>
import api from '../../api'
import { getImageURL as getImageURLUtil } from '../../util/image'
import { confirmAndDeleteAkaGroup } from '../../util/akaGroups'
import { encodeJsonBase64 } from '../../util/base64'
import videojs from 'video.js'
import 'videojs-vr/dist/videojs-vr.min.js'
import { format, parseISO } from 'date-fns'
import VueLoadImage from 'vue-load-image'
import GlobalEvents from 'vue-global-events'
import StarRating from 'vue-star-rating'
import ActorFavouriteButton from '../../components/ActorFavouriteButton'
import ActorWatchlistButton from '../../components/ActorWatchlistButton'
import ActorEditButton from '../../components/ActorEditButton'
import LinkStashdbButton from '../../components/LinkStashdbButton'
import SceneCard from '../scenes/SceneCard'
import ActorCard from './ActorCard'

export default {
  name: 'ActorDetails',
  components: { VueLoadImage, GlobalEvents, StarRating, ActorWatchlistButton, ActorFavouriteButton, SceneCard, ActorEditButton,  ActorCard, LinkStashdbButton },
  data () {
    return {
      index: 1,
      activeTab: 0,
      activeMedia: 0,
      carouselSlide: 0,
      sortMultiple: true,
      countries: [],
      akas: [],
      extrefs: [],
      colleagues: [],
    }
  },
  computed: {
    actor () {      
      const actor = this.$store.state.overlay.actordetails.actor
      api.get(`/actor/akas/${actor.id}`)
      .json()
      .then(list => {          
        this.akas = list
      })
      api.get(`/actor/colleagues/${actor.id}`)
      .json()
      .then(list => {          
        this.colleagues = list
      })
      api.get(`/actor/extrefs/${actor.id}`)
      .json()
      .then(list => {          
        this.extrefs = list          
      })
      return actor
    },
    // Properties for gallery
    images () {
      if (this.actor.image_arr==undefined || this.actor.image_arr=="") {
        return []
      }      
      return JSON.parse(this.actor.image_arr).filter(im => im != "")      
    },
    showEdit () {
      return this.$store.state.overlay.actoredit.show
    },
    showOpenInNewWindow () {
      return this.$store.state.optionsWeb.web.showOpenInNewWindow
    },
  },
  mounted () {    
      api.get('/actor/countrylist')
        .json()
        .then(list => {
          this.countries=list
        })
  },
  watch: {
    // when a file is selected, then this will fire the upload process
    activeTab: function (newval, oldval) {      
      }
    },  
    methods: {
    getImageURL (u, size, context = 'scene-0') {
      return getImageURLUtil(u, size, context)
    },
    getIndicatorURL (idx) {
      if (this.images[idx] !== undefined) {
        return this.getImageURL(this.images[idx], 'x85', 'act-' + this.actor.id)
      } else {
        return '/ui/images/blank_female_profile.png'
      }
    },
    close () {      
      this.$store.commit('overlay/hideActorDetails')
    },
    setRating (val) {
      api.post(`/actor/rate/${this.actor.id}`, { json: { rating: val } })
      const updatedActor = Object.assign({}, this.actor)
      updatedActor.star_rating = val
      this.actor.star_rating = val      
      this.$store.commit('actorList/updateActor', updatedActor)
    },
    async nextActor () {      
      const data = this.$store.getters['actorList/nextActor'](this.actor)
      if (data !== null) {
        this.$store.commit('overlay/showActorDetails', { actor: data })
        this.activeMedia = 0
        this.carouselSlide = 0        
      } else {
        // no actor, get the next page (note offset already points to it)
        let newoffset = this.$store.state.actorList.offset
        if (newoffset>this.$store.state.actorList.total)
        {
          // wrap back to the start
          newoffset = 0
        }
        await this.$store.dispatch('actorList/load', { offset: newoffset })
        const data = this.$store.getters['actorList/firstActor'](this.actor)
        if (data !== null) {
          this.$store.commit('overlay/showActorDetails', { actor: data })
          this.activeMedia = 0
          this.carouselSlide = 0
        }
      }
    },
    async prevActor () {
      const data = this.$store.getters['actorList/prevActor'](this.actor)
      if (data !== null) {
        this.$store.commit('overlay/showActorDetails', { actor: data })
        this.activeMedia = 0
        this.carouselSlide = 0        
      } else {
        // no actor, get the previous page
        let newoffset = this.$store.state.actorList.offset - (this.$store.state.actorList.limit * 2)
        if (newoffset < 0) {
          // wrap back to the last actor
          newoffset = Math.floor(this.$store.state.actorList.total / this.$store.state.actorList.limit) * this.$store.state.actorList.limit
        }
        await this.$store.dispatch('actorList/load', { offset: newoffset })
        const data = this.$store.getters['actorList/lastActor'](this.actor)
        if (data !== null) {
          this.$store.commit('overlay/showActorDetails', { actor: data })
          this.activeMedia = 0
          this.carouselSlide = 0
        }
      }
    },    
    handleLeftArrow () {      
        this.carouselSlide = this.carouselSlide - 1
    },
    handleRightArrow () {
        this.carouselSlide = this.carouselSlide + 1
    },
    scrollToActiveIndicator (value) {
      const indicators = document.querySelector('.carousel-indicator')
      const active = indicators.children[value]
      indicators.scrollTo({
        top: 0,
        left: active.offsetLeft + active.offsetWidth / 2 - indicators.offsetWidth / 2,
        behavior: 'smooth'
      })
    },
    calcAge(birthdate){       
      const birthdateObj = new Date(birthdate);
      const now = new Date();
      const diffInMs = now - birthdateObj;
      const msPerYear = 1000 * 60 * 60 * 24 * 365.25; // average milliseconds per year, accounting for leap years
      const age = Math.floor(diffInMs / msPerYear);      
      return age
    },
    getYearsActive(){      
      let active = ""
      if (this.actor.start_year > 0) {
        active = this.actor.start_year 
      }      
      active +=  "-"      
      if (this.actor.end_year > 0) {
        active += this.actor.end_year 
      }
      return active
    },
    measurements(){      
      let metric_measurements=""
      let imperial_measurements=""
      if (this.actor.band_size != 0) {
        metric_measurements=this.actor.band_size
        imperial_measurements=Math.round(this.actor.band_size / 2.54)
      }
      if (this.actor.cup_size != ''){
        metric_measurements +=  this.actor.cup_size        
        imperial_measurements += this.actor.cup_size        
      }
      if (this.actor.waist_size != 0) {
        if (metric_measurements!='') {
          metric_measurements += '-'
          imperial_measurements += '-'
        }
        metric_measurements += this.actor.waist_size
        imperial_measurements += Math.round(this.actor.waist_size  / 2.54)
      }
      if (this.actor.hip_size != 0) {
        if (metric_measurements!='') {
          metric_measurements += '-'
          imperial_measurements += '-'
        }
        metric_measurements += this.actor.hip_size
        imperial_measurements += Math.round(this.actor.hip_size / 2.54)
      } 
      if (metric_measurements==''){
        return ''
      }
      return imperial_measurements + " / " + metric_measurements
    },
    joinArray(jsonArr){
      const arr = JSON.parse(jsonArr);
      return  arr.join(", ");       
    },
    setActorImage (val) {
      api.post('/actor/setimage', {
      json: {
        actor_id: this.actor.id,
        url: this.images[this.carouselSlide]
      }}).json().then(data => {        
        this.$store.state.overlay.actordetails.actor = data
        this.carouselSlide=0
        this.$store.dispatch('actorList/load', { offset: this.$store.state.actorList.offset - this.$store.state.actorList.limit })
      })    
    },
    deleteActorImage (val) {
      api.delete('/actor/delimage', {
      json: {
        actor_id: this.actor.id,
        url: this.images[this.carouselSlide]
      }}).json().then(data => {
        this.$store.state.overlay.actordetails.actor = data
      })    
    },
    getCountryName(countryCode){
      const country = this.countries.find(c => c.code === countryCode)
      if (country == undefined) {
        return countryCode
      }
      return country.name
    },
    getCountryFlag(countryCode){
      const country = this.countries.find(c => c.code === countryCode)
      if (country == undefined) {
        return getImageURLUtil('https://flagcdn.com/' + countryCode.toLowerCase() + '.svg', '700x', 'icon-' + countryCode.toLowerCase())
      }
      return country.flag_url
    },
    getWeight(kg) {
        return kg + " kg - " + Math.round( kg * 2.20462) + " lbs"
    },
    getHeight(cm){
      const totalInches = Math.round(cm / 2.54)
      let feet = Math.floor(totalInches / 12)
      return cm + " cm - " + feet + "' " +  Math.round(totalInches - (feet*12)) + '"'
    },
    getActorUrls() {
      if (this.actor.urls=="")
      {
        return []
      }      
      let array = JSON.parse(this.actor.urls)      
      return array
    },
    createAkaGroup () {
      this.$store.state.actorList.isLoading = true
      let actorlist = [this.actor.name]
      for (let idx = 0; idx < this.akas.possible_akas.length; idx++) {
        actorlist.push(this.akas.possible_akas[idx].name)
      }
      api.post('/aka/create', {json: {actorList: actorlist}}).json().then(data => {
        if (data.status != '') {
          this.$buefy.toast.open({message: `Warning:  ${data.status}`, type: 'is-warning', duration: 5000})
        }
        api.get('/actor/'+this.actor.id).json().then(data => {
          if (data.id != 0){
            this.$store.state.overlay.actordetails.actor = data          
          }          
        })
        this.$store.state.actorList.isLoading = false
        this.$store.dispatch('actorList/load', { offset: this.$store.state.actorList.offset - this.$store.state.actorList.limit })

      })
    },
    deleteAkaGroup () {
      confirmAndDeleteAkaGroup(this, this.actor.name, 'actorList', () => {
        this.$store.dispatch('actorList/load', { offset: this.$store.state.actorList.offset - this.$store.state.actorList.limit })
        this.close()
      })
    },
    addToAkaGroup (newMember) {
      this.$store.state.actorList.isLoading = true
      api.post('/aka/add', {json: {actorList: [this.actor.name, newMember]}}).json().then(data => {        
        // delete old aka & add new name
        if (data.status != '') {
          this.$buefy.toast.open({message: `Warning:  ${data.status}`, type: 'is-warning', duration: 5000})
        }
        api.get('/actor/'+this.actor.id).json().then(data => {
          if (data.id != 0){
            this.$store.state.overlay.actordetails.actor = data          
          }          
        })
        this.$store.state.actorList.isLoading = false
      })
      
    },
    removeFromAkaGroup (memberToRemove) {
      this.$store.state.actorList.isLoading = true
      api.post('/aka/remove', {json: {actorList: [this.actor.name, memberToRemove]}}).json().then(data => {        
        // delete old aka & add new name
        if (data.status != '') {
          this.$buefy.toast.open({message: `Warning:  ${data.status}`, type: 'is-warning', duration: 5000})
        }
        api.get('/actor/'+this.actor.id).json().then(data => {          
          if (data.id != 0){
            this.$store.state.overlay.actordetails.actor = data          
          }          
        })
        this.$store.state.actorList.isLoading = false
      })
    },
    enableNewAkaGroup () {
      if (this.actor.name.startsWith("aka:")){
        return false
      }
      if (this.akas.aka_groups != null)
      {
        return false
      }
      if (this.akas.possible_akas == null)
      {
        return false
      }
      for (let idx = 0; idx < this.akas.possible_akas.length; idx++) {
        if (this.akas.possible_akas[idx].name.startsWith("aka:")) {
          return false
        }
      }      
      return true
    },
    getCastScenesUrl(actor) {
      let newfilters = Object.assign({}, this.$store.state.sceneList.filters);
      console.log(newfilters)
      newfilters.cast = actor;
      newfilters.dlState = "any"
      newfilters.isAvailable=null
      newfilters.isAccessible=null
      console.log(newfilters)
      return this.$router.resolve({
        name: 'scenes',
        query: { q: encodeJsonBase64(newfilters) }
      }).href
    },
    refreshScraper(url){
      if (url.includes('stashdb')) {
        this.$store.state.actorList.isLoading = true
        const lastSlashIndex = url.lastIndexOf('/');
        api.get('/extref/stashdb/refresh_performer/'+url.substring(lastSlashIndex + 1)).then(data => {
          api.get('/actor/'+this.actor.id).json().then(data => {          
            if (data.id != 0){
              this.$store.state.overlay.actordetails.actor = data
              this.$store.state.actorList.isLoading = false
              this.$store.dispatch('actorList/load', { offset: this.$store.state.actorList.offset - this.$store.state.actorList.limit })
            }
          })
        })
      } else {
        this.$store.state.actorList.isLoading = true
        api.post('/extref/generic/scrape_single', { json: {id: this.actor.id,url: url}})
          .then(data => {
            api.get('/actor/'+this.actor.id).json().then(data => {
              if (data.id != 0){
                this.$store.state.overlay.actordetails.actor = data
                this.$store.state.actorList.isLoading = false
                this.$store.dispatch('actorList/load', { offset: this.$store.state.actorList.offset - this.$store.state.actorList.limit })
              }
            })
          })
      }
      this.$store.state.actorList.isLoading = false
    },
    format,
    parseISO
  }
}
</script>

<style lang="less" scoped>
/* ------------------------------------------------------------------
   Actor details overlay — two-pane media/info layout
   (mirrors scenes/Details.vue)
   ------------------------------------------------------------------ */

.details-card {
  width: min(1500px, 92vw);
}

@media (max-width: 768px) {
  .details-card {
    width: 98vw;
  }
}

.details-body {
  padding: 1.25rem;
}

.details-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 1.25rem;
  align-items: start;
}

@media (max-width: 1024px) {
  .details-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}

/* ------------------------------------------------------------------
   Media pane — follows the active theme via --xbvr-media-* tokens
   ------------------------------------------------------------------ */

.media-pane {
  background: var(--xbvr-media-bg, #eef0f4);
  border-radius: var(--xbvr-radius-lg, 16px);
  padding: 0.75rem 0.75rem 1rem;
  position: sticky;
  top: 0;
}

.media-tabs :deep(.tabs ul) {
  border-bottom: none;
  justify-content: center;
  gap: 4px;
  width: fit-content;
  margin: 0 auto 0.6rem;
  padding: 4px;
  background: var(--xbvr-media-tabbar, rgba(16, 24, 40, 0.06));
  border-radius: 999px;
}

.media-tabs :deep(.tabs a) {
  border-bottom: none;
  border-radius: 999px;
  padding: 0.3em 1.2em;
  color: var(--xbvr-media-muted, #64708a);
  font-weight: 600;
  transition: color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
    background-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

.media-tabs :deep(.tabs a:hover) {
  color: var(--xbvr-text, #1c2333);
  border-bottom: none;
}

.media-tabs :deep(.tabs li.is-active a) {
  background: var(--xbvr-media-tab-active-bg, #ffffff);
  color: var(--xbvr-media-tab-active-text, #4338ca);
}

.media-pane :deep(.carousel-item .image) {
  border-radius: var(--xbvr-radius-sm, 8px);
}

.media-actions {
  display: flex;
  justify-content: center;
  gap: 0.5rem;
  margin-top: 0.75rem;
}

.media-actions :deep(.button) {
  border-radius: 999px;
  box-shadow: none;
}

/* ------------------------------------------------------------------
   Info pane
   ------------------------------------------------------------------ */

.details-header {
  margin-bottom: 0.9rem;
}

.details-title {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.5rem;
  font-size: 1.35rem;
  font-weight: 800;
  letter-spacing: -0.01em;
  line-height: 1.25;
  color: var(--xbvr-text, #1c2333);
  overflow-wrap: break-word;
  margin-bottom: 0.55rem;
}

.title-actions {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
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

/* rating + action toolbar */
.details-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem 1.25rem;
  padding: 0.55rem 0.8rem;
  margin-bottom: 0.9rem;
  background: var(--xbvr-surface-sunken, #eef0f4);
  border: 1px solid var(--xbvr-border, #e3e6ec);
  border-radius: var(--xbvr-radius, 12px);
}

.rating-block {
  display: flex;
  align-items: center;
  margin-bottom: 0 !important;
}

.rating-label {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--xbvr-text-muted, #64708a);
  margin-right: 0.6em;
  white-space: nowrap;
}

.rating-reset {
  padding-left: 0.6em;
  color: var(--xbvr-text-faint, #7d88a1);
  cursor: pointer;
  transition: color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

.rating-reset:hover {
  color: var(--xbvr-text, #1c2333);
}

.action-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.3rem;
  margin-left: auto;
}

.action-row :deep(.button.is-small) {
  border-radius: 8px;
}

.vue-star-rating {
  line-height: 0;
}

/* details tab attribute grid */
.attribute-container {
  display: flex;
  flex-wrap: wrap;
  gap: 0.2rem 1.5rem;
  margin-bottom: 0.9rem;
}

.attribute-container :deep(.field) {
  margin-bottom: 0.35rem;
}

.attribute-heading {
  width: 120px;
  color: var(--xbvr-text-muted, #64708a);
  font-weight: 600;
}

.attribute-data {
  width: 200px;
}

.attribute-long-data {
  min-width: 320px;
}

.flag-img {
  height: 15px;
  border: 1px solid var(--xbvr-border-strong, #cdd2dc);
  border-radius: 2px;
  margin-right: 0.5em;
}

/* links & scrapers tabs */
.link-list {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.4rem;
}

.link-tag {
  border-radius: 999px;
}

.scraper-row {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.refresh-icon {
  color: var(--xbvr-text-muted, #64708a);
  transition: color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

.refresh-icon:hover {
  color: var(--xbvr-primary, #4f46e5);
}

.tab-link-icon {
  padding-left: 0.1em;
  border-bottom-style: none;
}

/* card grids inside tabs */
.image-wrapper {
  position: relative;
}

div.scroll {
  max-height: 60vh;
  overflow-x: hidden;
  overflow-y: auto;
}

.block-tab-content {
  flex: 1 1 auto;
}

.is-1by1 {
  padding-top: calc(100% - 40px - 1em) !important;
}

/* carousel indicators */
.indicator-thumb {
  width: max-content;
}

.indicator-img {
  height: 85px;
  border-radius: 6px;
}

.indicator-img.is-placeholder {
  height: 25px;
}

:deep(.carousel .carousel-indicator) {
  justify-content: flex-start;
  width: 100%;
  max-width: min-content;
  margin-left: auto;
  margin-right: auto;
  overflow: auto;
}

:deep(.carousel .carousel-indicator .indicator-item:not(.is-active)) {
  opacity: 0.45;
}

:deep(.carousel .carousel-indicator .indicator-item img) {
  border-radius: 6px;
}

:deep(.carousel .carousel-indicator .indicator-item.is-active img) {
  outline: 2px solid var(--xbvr-primary, #4f46e5);
  outline-offset: 1px;
}

/* close + prev/next — circular glass controls over the overlay */
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

.actor-nav {
  position: fixed;
  top: 50%;
  transform: translateY(-50%);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 46px;
  height: 46px;
  padding: 0;
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: 999px;
  background: var(--xbvr-chip-bg, rgba(20, 24, 36, 0.55));
  color: #fff;
  cursor: pointer;
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  user-select: none;
  -webkit-user-select: none;
  transition: background-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
    transform var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

.actor-nav:hover {
  background: rgba(20, 24, 36, 0.8);
}

.actor-nav.prev {
  left: 14px;
}

.actor-nav.prev:hover {
  transform: translateY(-50%) translateX(-2px);
}

.actor-nav.next {
  right: 14px;
}

.actor-nav.next:hover {
  transform: translateY(-50%) translateX(2px);
}
</style>
