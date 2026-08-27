<template>
  <div class="card scene-card">
    <div class="card-image">
      <div class="bbox"
           v-bind:style='{backgroundImage: `url("${getImageURL(item.cover_url, '700x', item.scene_id)}")`, backgroundSize: this.sceneCardScale, backgroundPosition: "center", backgroundRepeat: "no-repeat", opacity:item.is_available ? 1.0 : this.isAvailOpactiy, aspectRatio: this.sceneCardAspectRatio}'
           @click="showDetails(item)"
           @mouseover="preview = true"
           @mouseleave="preview = false">
        <video v-if="preview && item.has_preview" :src="`/api/dms/preview/${item.scene_id}`"
               autoplay muted loop playsinline></video>
        <div class="overlay align-bottom-left">
          <div class="badge-row">
            <b-tag v-if="item.is_watched && !this.$store.state.optionsWeb.web.sceneWatched">
              <b-icon pack="mdi" icon="eye" size="is-small"/>
            </b-tag>
            <b-tag type="is-info" v-if="videoFilesCount > 1 && !item.is_multipart">
              <b-icon pack="mdi" icon="file" size="is-small" style="margin-right:0.1em"/>
              {{videoFilesCount}}
            </b-tag>
            <b-tag type="is-info" v-if="item.is_scripted">
              <b-icon pack="mdi" icon="pulse" size="is-small"/>
              <span v-if="scriptFilesCount > 1">{{scriptFilesCount}}</span>
            </b-tag>
            <b-tag type="is-info" v-if="hspFilesCount > 0 && this.$store.state.optionsWeb.web.showHspFile">
              <b-icon pack="mdi" icon="safety-goggles" size="is-small"/>
              <span v-if="hspFilesCount > 1">{{hspFilesCount}}</span>
            </b-tag>
            <b-tag type="is-info" v-if="subtitlesFilesCount > 0 && this.$store.state.optionsWeb.web.showSubtitlesFile">
              <b-icon pack="mdi" icon="subtitles" size="is-small"/>
              <span v-if="subtitlesFilesCount > 1">{{subtitlesFilesCount}}</span>
            </b-tag>
            <b-tag type="is-info" v-if="item.cuepoints != null && item.cuepoints.length > 0 && this.$store.state.optionsWeb.web.sceneCuepoint">
              <b-icon pack="mdi" icon="skip-next-outline" size="is-small"/>
              <span v-if="item.cuepoints != null && item.cuepoints.length > 1">{{item.cuepoints.length}}</span>
            </b-tag>
            <b-tag type="is-warning" v-if="item.star_rating > 0">
              <b-icon pack="mdi" icon="star" size="is-small"/>
              {{item.star_rating}}
            </b-tag>
            <b-tag type="is-info" v-if="item.duration > 0 && this.$store.state.optionsWeb.web.sceneDuration">
              <b-icon pack="mdi" icon="clock" size="is-small"/>
              {{item.duration}}m
            </b-tag>
          </div>
          <div v-if="this.$store.state.optionsWeb.web.showScriptHeatmap && (files = getFunscripts(this.$store.state.optionsWeb.web.showAllHeatmaps))" style="padding: 0px 5px 5px">
            <div v-if="files.length" class="heatmapFunscript">
              <img v-for="file in files" :src="getHeatmapURL(file.id)"/>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div style="padding-top:4px;">
      <div class="scene_title">{{item.title}}</div>

      <hidden-button :item="item" v-if="this.$store.state.optionsWeb.web.sceneHidden"/>
      <watchlist-button :item="item" v-if="this.$store.state.optionsWeb.web.sceneWatchlist"/>
      <trailerlist-button :item="item" v-if="this.$store.state.optionsWeb.web.sceneTrailerlist"/>
      <favourite-button :item="item" v-if="this.$store.state.optionsWeb.web.sceneFavourite"/>
      <wishlist-button v-if="this.$store.state.optionsWeb.web.sceneWishlist && !item.is_available" :item="item"/>
      <watched-button :item="item" v-if="this.$store.state.optionsWeb.web.sceneWatched"/>
      <edit-button :item="item" v-if="this.$store.state.optionsWeb.web.sceneEdit" />
      <link-stashdb-button :item="item" v-if="!this.stashLinkExists" objectType="scene"/>

      <span class="is-pulled-right card-meta">
        <a v-if="item.members_url != ''" :href="safeHref(item.members_url)" target="_blank" title="Members Link" rel="noreferrer"><b-icon pack="mdi" icon="link-lock" custom-size="mdi-18px" style="height:0.7rem"/></a>
        <a :href="safeHref(item.scene_url)" :class="{'has-text-white has-background-primary-dark': item.is_subscribed }" target="_blank" rel="noreferrer" style="padding:2px">{{item.site}}</a><br/>
        <span v-if="item.release_date !== '0001-01-01T00:00:00Z'">
          {{format(parseISO(item.release_date), "yyyy-MM-dd")}}
        </span>        
      </span>
      <div class="image-row" v-if="alternateSources.length != 0">
        <div v-for="(altsrc, idx) in this.alternateSources" :key="idx" class="altsrc-image-wrapper">
          <b-tooltip type="is-light" :label="altsrc.title" :delay="100">
            <a :href="safeHref(altsrc.url)" target="_blank">
              <vue-load-image>
                 <img slot="image" :src="getImageURL(altsrc.site_icon, '700x', altSourceIconContext(altsrc))" alt="Image" class="thumbnail" width="20" />
                <b-icon slot="error" pack="mdi" icon="link" size="is-small" />
              </vue-load-image>
            </a>
          </b-tooltip>
        </div>
      </div>    
    </div>
  </div>
</template>

<script>
import { format, parseISO } from 'date-fns'
import WatchlistButton from '../../components/WatchlistButton'
import FavouriteButton from '../../components/FavouriteButton'
import WishlistButton from '../../components/WishlistButton'
import WatchedButton from '../../components/WatchedButton'
import EditButton from '../../components/EditButton'
import LinkStashdbButton from '../../components/LinkStashdbButton'
import TrailerlistButton from '../../components/TrailerlistButton'
import HiddenButton from '../../components/HiddenButton'
import api from '../../api'
import VueLoadImage from 'vue-load-image'
import { getImageURL, altSourceIconContext } from '../../util/image'
import { safeHref } from '../../util/url'

export default {
  name: 'SceneCard',
  props: { item: Object, reRead: Boolean },
  components: { WatchlistButton, FavouriteButton, WishlistButton, WatchedButton, EditButton, LinkStashdbButton, TrailerlistButton, HiddenButton, VueLoadImage },
  data () {
    return {
      preview: false,
      format,
      parseISO,
      alternateSources: [],
      stashLinkExists: false,
    }
  },
  computed: {
    videoFilesCount () {
      if (this.item.file == null) { return 0 }
      let count = 0
      this.item.file.forEach(obj => {
        if (obj.type === 'video') {
          count = count + 1
        }
      })
      return count
    },
    scriptFilesCount () {
      let count = 0
      if (this.item.file == null) { return 0 }
      this.item.file.forEach(obj => {
        if (obj.type === 'script') {
          count = count + 1
        }
      })
      return count
    },
    hspFilesCount () {
      let count = 0
      if (this.item.file == null) { return 0 }
      this.item.file.forEach(obj => {
        if (obj.type === 'hsp') {
          count = count + 1
        }
      })
      return count
    },
    subtitlesFilesCount () {
      if (this.item.file == null) { return 0 }
      let count = 0
      this.item.file.forEach(obj => {
        if (obj.type === 'subtitles') {
          count = count + 1
        }
      })
      return count
    },
    isAvailOpactiy () {
      if (this.$store.state.optionsWeb.web.isAvailOpacity == undefined) {
        return .4
      }
      return this.$store.state.optionsWeb.web.isAvailOpacity / 100
    },
    sceneCardAspectRatio () {
      if (this.$store.state.optionsWeb.web.sceneCardAspectRatio == "3:2") {
        return 3 / 2
      } else if (this.$store.state.optionsWeb.web.sceneCardAspectRatio == "16:9") {
        return 16 / 9
      } else {
        return 1
      }
    },
    sceneCardScale () {
      if (this.$store.state.optionsWeb.web.sceneCardScaleToFit) {
        return "contain"
      } else {
        return "cover"
      }
    }
  },
  created () {
    this.loadAlternateSources()
  },
  methods: {
    getImageURL,
    altSourceIconContext,
    safeHref,
    // Fetched once when the card is created. Previously this was an async
    // computed used as v-if — an always-truthy Promise that refetched on
    // every render of every card.
    async loadAlternateSources () {
      this.stashLinkExists = false
      try {
        const response = await api.get('/scene/alternate_source/' + this.item.id).json();
        this.alternateSources = [];
        if (response == null) {
          return;
        }

        this.alternateSources = response
          .filter(altsrc => altsrc.external_source.startsWith("alternate scene ") || altsrc.external_source == "stashdb scene")
          .map(altsrc => {
            const extdata = JSON.parse(altsrc.external_data);
            let title;
            if (altsrc.external_source.startsWith("alternate scene ")) {
              title = extdata.scene?.title || 'No Title';
            } else if (altsrc.external_source == "stashdb scene") {
              title = extdata.title || 'No Title';
            }
            if (altsrc.external_source.includes('stashdb')) {
              this.stashLinkExists = true
            }
            return {
              ...altsrc,
              title: title
            };
          });
      } catch (error) {
        // leave the card without alternate sources on error
      }
    },
    showDetails (scene) {
      // reRead is required when the SceneCard is clicked from the ActorDetails
      // the Scenes associated Tables such as Tags, Cast arwon't be Preloaded and
      // will cause errors when the Details Overlay loads
      if (this.reRead) {
        api.get('/scene/'+scene.id).json().then(data => {
          if (data.id != 0){
            this.$store.commit('overlay/showDetails', { scene: data })
          }
        })
      } else {
        this.$store.commit('overlay/showDetails', { scene: scene })
      }
      this.$store.commit('overlay/hideActorDetails')
    },
    getHeatmapURL (fileId) {
      return `/api/dms/heatmap/${fileId}`
    },
    getFunscripts (showAll) {
      if (showAll) {
        return this.item.file !== null && this.item.file.filter(a => a.type === 'script' && a.has_heatmap);
      } else {
        if (this.item.file !== null) {
          let script;
          if (script = this.item.file.find((a) => a.type === 'script' && a.has_heatmap && a.is_selected_script)) {
            return [script]
          }
          if (script = this.item.file.find((a) => a.type === 'script' && a.has_heatmap)) {
            return [script]
          }
        }
        return false;
      }
    }
  }
}
</script>

<style scoped>
  .scene-card {
    border: none;
    background: transparent;
    box-shadow: none;
  }

  .scene-card:hover {
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

  .scene-card:hover .card-image {
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
    /* fill the thumbnail's width, keep the video's natural aspect,
       and center vertically — overflow is cropped by the card frame */
    position: absolute;
    left: 0;
    top: 50%;
    width: 100%;
    height: auto;
    transform: translateY(-50%);
  }

  .overlay {
    flex: 1 0 calc(25%);
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    padding: 0;
    line-height: 0;
    position: absolute;
    left: 0;
    top: 0;
    right: 0;
    bottom: 0;
    pointer-events: none;
  }

  .align-bottom-left {
    align-items: flex-end;
    justify-content: flex-end;
    flex-wrap: wrap;
    flex-direction: column
  }

  .bbox:after {
    content: '';
    display: block;
    padding-bottom: 100%;
  }

  /* Glassmorphic pill badges over the cover */
  .badge-row {
    padding: 6px;
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 4px;
  }

  .badge-row .tag {
    margin-left: 0;
    border-radius: 999px;
    background: var(--xbvr-badge-bg, rgba(255, 255, 255, 0.82));
    color: var(--xbvr-badge-text, #1c2333);
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    font-weight: 600;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.18);
  }

  .badge-row .tag.is-warning {
    background: rgba(232, 161, 58, 0.92);
    color: #1c2333;
  }

  .scene_title {
    font-size: 0.8rem;
    font-weight: 600;
    line-height: 1.35;
    color: var(--xbvr-text, #1c2333);
    margin-top: 6px;
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow-wrap: break-word;
  }

.heatmapFunscript {
  width: auto;
}

.heatmapFunscript img {
  border: 1px solid var(--xbvr-border-strong, #cdd2dc);
  width: 100%;
  height: 15px;
  border-radius: 999px;
}

.altsrc-image-wrapper {
  display: inline-block;
  margin-right: 5px;
  margin-top: 3px;
}

.altsrc-image-wrapper .thumbnail {
  border-radius: 4px;
}

.card-meta {
  font-size: 0.7rem;
  line-height: 1.5;
  text-align: right;
  color: var(--xbvr-text-muted, #64708a);
}

.card-meta a {
  color: var(--xbvr-text-muted, #64708a);
  font-weight: 600;
  border-radius: 4px;
  padding: 1px 4px;
  transition: background-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
    color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

.card-meta a:hover {
  background: var(--xbvr-primary-soft, #eef0fe);
  color: var(--xbvr-primary-strong, #4338ca);
}

</style>
