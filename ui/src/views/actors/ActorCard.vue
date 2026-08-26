<template>
  <div class="card actor-card">
    <div class="card-image">
      <div class="bbox"
           v-bind:style="{backgroundImage: `url(${getImageURL(actor.image_url)})`, backgroundSize: actorCardScale, backgroundPosition: 'center', backgroundRepeat: 'no-repeat', opacity:isAvailable(actor) ? 1.0 : isAvailOpacity, aspectRatio: actorCardAspectRatio}"
           @click="showDetails(actor)"
           @mouseover="preview = true"
           @mouseleave="preview = false">
      </div>
        <div class="overlay align-bottom-left">
         </div>
    </div>

    <div style="padding-top:4px;">
      <div class="scene_title">{{actor.name}}</div>
      <a v-if="colleague!=undefined" class="button is-info is-outlined is-small"
        @click="showColleague(actor.name,colleague)"
        :title="'Show Scenes with ' + actor.name">
        <b-icon pack="mdi" :icon="'movie-outline'" size="is-small"/>
      </a>
      <actor-favourite-button :actor="actor" v-if="this.$store.state.optionsWeb.web.sceneFavourite"/>
      <actor-watchlist-button :actor="actor" v-if="this.$store.state.optionsWeb.web.sceneWatchlist"/>
      <actor-edit-button :actor="actor"/>&nbsp;
      <link-stashdb-button :item="actor" objectType="actor" />
      <b-tooltip :label="$t('Your rating')" :delay="500">
      <b-tag type="is-warning" v-if="actor.star_rating != 0 " size="is-small" style="height:30px;">
        <b-icon pack="mdi" icon="star" size="is-small"/>
        {{actor.star_rating}}
      </b-tag>
      </b-tooltip>
      <b-tooltip :label="$t('Average rating of scenes')" :delay="500">
      <b-tag type="is-primary" v-if="actor.scene_rating_average != 0 " style="height:30px;">
        <b-icon pack="mdi" icon="star" size="is-small"/>
        {{Math.round(actor.scene_rating_average * 4) / 4}}
      </b-tag>
      </b-tooltip>

      <span class="is-pulled-right card-meta">
          <span v-if="actor.birth_date != '0001-01-01T00:00:00Z'">{{format(parseISO(actor.birth_date), "yyyy-MM-dd")}}</span>
          <vue-load-image style="display:inline-block">
            <img slot="image" :src="getImageURL('https://flagcdn.com/' + actor.nationality.toLowerCase() +'.svg')" style="height:10px;border: 1px solid black;margin-left:0.5em" />
          </vue-load-image>
          <div>
          <span v-if="actor.scenes.length == 1">{{actor.scenes.length}} scene</span>
          <span v-if="actor.scenes.length > 1">{{actor.scenes.length}} scenes</span>
          <span v-if="actor.avail_count > 0">, {{actor.avail_count}} available</span>
          </div>
      </span>
    </div>
  </div>
</template>

<script>
import { format, parseISO } from 'date-fns'
import ActorFavouriteButton from '../../components/ActorFavouriteButton'
import ActorWatchlistButton from '../../components/ActorWatchlistButton'
import ActorEditButton from '../../components/ActorEditButton'
import LinkStashdbButton from '../../components/LinkStashdbButton'
import VueLoadImage from 'vue-load-image'
import { getImageURL as getImageURLUtil } from '../../util/image'
import { tr } from 'date-fns/locale'

export default {
  name: 'ActorCard',
  props: { actor: Object, colleague: String },
   components: {ActorFavouriteButton, ActorWatchlistButton, VueLoadImage, ActorEditButton, LinkStashdbButton},
  data () {
    return {
      preview: false,
      format,
      parseISO
    }
  },
  computed: {
    isAvailOpacity () {      
      if (this.$store.state.optionsWeb.web.isAvailOpacity == undefined) {
        return .4
      }
      return this.$store.state.optionsWeb.web.isAvailOpacity / 100
    },
    actorCardAspectRatio () {
      if (this.$store.state.optionsWeb.web.actorCardAspectRatio == "2:3") {
        return 2 / 3
      } else if (this.$store.state.optionsWeb.web.actorCardAspectRatio == "9:16") {
        return 9 / 16
      } else {
        return 1
      }
    },
    actorCardScale () {
      if (this.$store.state.optionsWeb.web.actorCardScaleToFit) {
        return "contain"
      } else {
        return "cover"
      }
    },
  },
  methods: {
    getImageURL (u) {
      return getImageURLUtil(u, '700x', '/ui/images/blank_female_profile.png')
    },
    showDetails (actor) {
      this.$store.commit('overlay/showActorDetails', { actor: actor })
    },
    isAvailable(actor) {
      let index = actor.scenes.findIndex(scene => scene.is_available == 1);
      if (index == -1) {
        return false
      }
      return true
    },
    showColleague (main_actor, colleague) {      
      this.$store.state.sceneList.filters.cast = ["&" + main_actor , "&"+ colleague]
      this.$store.state.sceneList.filters.sites = []
      this.$store.state.sceneList.filters.tags = []
      this.$store.state.sceneList.filters.attributes = []
      this.$store.state.actorList.filters.dlState = "Any"
      this.$router.push({
        name: 'scenes',
        query: { q: this.$store.getters['sceneList/filterQueryParams'] }
      })
      this.$store.commit('overlay/hideActorDetails')
    },

  },
}
</script>

<style scoped>
  .actor-card {
    border: none;
    background: transparent;
    box-shadow: none;
  }

  .actor-card:hover {
    transform: none;
    border-color: transparent;
  }

  .button {
    margin-right: 3px;
  }

  .card-image {
    border-radius: var(--xbvr-radius, 12px);
    overflow: hidden;
    box-shadow: var(--xbvr-shadow, 0 1px 3px rgba(16, 24, 40, 0.08));
    transition: box-shadow var(--xbvr-med, 220ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
      transform var(--xbvr-med, 220ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
  }

  .actor-card:hover .card-image {
    box-shadow: var(--xbvr-shadow-lg, 0 16px 40px rgba(16, 24, 40, 0.16));
    transform: translateY(-3px);
  }

  .bbox {
    flex: 1 0 calc(25%);
    position: relative;
    background: var(--xbvr-surface-sunken, #eef0f4);
    border-radius: var(--xbvr-radius, 12px);
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    padding: 0;
    line-height: 0;
    cursor: pointer;
  }

  .bbox:not(:hover) > video {
    display: none;
  }

  video {
    object-fit: cover;
    position: absolute;
    width: 100%;
    height: 100%;
  }

  .overlay {
   position: absolute;
  bottom: 0;
  right: 0;
  display: flex;
  padding: 5px;
  max-width: 5px;
  }

  .align-bottom-left {
    align-items: flex-end;
    justify-content: flex-end;
  }

  .bbox:after {
    content: '';
    display: block;
    padding-bottom: 100%;
  }

  .tag {
    margin-left: 0.1em;
    border-radius: 999px;
  }

  .scene_title {
    font-size: 0.85rem;
    font-weight: 600;
    line-height: 1.35;
    color: var(--xbvr-text, #1c2333);
    margin-top: 6px;
    overflow: hidden;
    text-overflow: ellipsis;
    overflow-wrap: break-word;
  }

  .card-meta {
    font-size: 0.7rem;
    line-height: 1.5;
    text-align: right;
    color: var(--xbvr-text-muted, #64708a);
  }
</style>
