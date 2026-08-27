<template>
  <div class="modal is-active" role="dialog" aria-modal="true">
    <GlobalEvents
      :filter="e => !['INPUT', 'TEXTAREA'].includes(e.target.tagName)"
      @keyup.esc="close"
      @keydown.left="handleLeftArrow"
      @keydown.right="handleRightArrow"
      @keydown.o="prevScene"
      @keydown.p="nextScene"
      @keydown.f="$store.commit('sceneList/toggleSceneList', {scene_id: item.scene_id, list: 'favourite'})"
      @keydown.exact.w="$store.commit('sceneList/toggleSceneList', {scene_id: item.scene_id, list: 'watchlist'})"
      @keydown.shift.w="$store.commit('sceneList/toggleSceneList', {scene_id: item.scene_id, list: 'watched'})"
      @keydown.t="$store.commit('sceneList/toggleSceneList', {scene_id: item.scene_id, list: 'trailerlist'})"
      @keydown.e="$store.commit('overlay/editDetails', {scene: item})"
      @keydown.s="$store.commit('overlay/showSearchStashdbScenes', {scene: item})"
      @keydown.g="toggleGallery"
      @keydown.48="setRating(0)"
    />

    <div class="modal-background"></div>

    <div class="modal-card details-card">
      <section class="modal-card-body details-body">
        <div class="details-grid">

          <div class="media-pane">
            <b-tabs v-model="activeMedia" position="is-centered" :animated="false" class="media-tabs">

              <b-tab-item label="Gallery">
                <b-carousel v-model="carouselSlide" @change="scrollToActiveIndicator" :autoplay="false" :indicator-inside="false">
                  <b-carousel-item v-for="(carousel, i) in images" :key="i">
                    <div class="image is-1by1 is-full"
                         v-bind:style='{backgroundImage: `url("${getImageURL(carousel.url, "700,fit", sceneContext(item.scene_id))}")`, backgroundSize: "contain", backgroundPosition: "center", backgroundRepeat: "no-repeat"}'></div>
                  </b-carousel-item>
                  <template slot="indicators" slot-scope="props">
                      <span class="al image" style="width:max-content;">
                        <vue-load-image>
                          <img slot="image" :src="getIndicatorURL(props.i)" style="height:40px;"/>
                          <img slot="preloader" src="/ui/images/blank.png" style="height:40px;"/>
                          <img slot="error" src="/ui/images/blank.png" style="height:40px;"/>
                        </vue-load-image>
                      </span>
                  </template>
                </b-carousel>
              </b-tab-item>

              <b-tab-item label="Player" v-if="!displayingAlternateSource">
                <video ref="player" class="video-js vjs-default-skin" controls playsinline preload="none"/>
                <b-field position="is-centered" class="skip-row">
                  <b-field>
                    <b-tooltip v-for="(skipBack, i) in skipBackIntervals" class="is-size-7" :key="i" :active="skipBack == lastSkipBackInterval ? true : false" :label="$t('Keyboard shortcut: Left Arrow')"
                        position="is-top" type="is-primary is-light" >
                    <b-button class="tag is-small is-outlined is-info is-light"  @click="playerStepBack(skipBack)">
                      <b-icon v-if="skipBack == lastSkipBackInterval" pack="mdi" icon="arrow-left-thin" size="is-small"></b-icon> {{ skipBack }}</b-button>
                    </b-tooltip>
                  </b-field>
                  <b-field style="margin-left:1em">
                    <b-tooltip v-for="(skipForward, i) in skipForwardIntervals" :key="i" :active="skipForward == lastSkipFowardInterval ? true : false" :label="$t('Keyboard shortcut: Right Arrow')"
                        position="is-top" type="is-primary is-light" >
                    <b-button class="tag is-small is-outlined is-info is-light" @click="playerStepForward(skipForward)">
                      <b-icon v-if="skipForward == lastSkipFowardInterval" pack="mdi" icon="arrow-right-thin" size="is-small"></b-icon> +{{ skipForward }}</b-button>
                    </b-tooltip>
                  </b-field>
                </b-field>
             </b-tab-item>

            </b-tabs>

          </div>

          <div class="info-pane">

            <header class="details-header">
              <h2 class="details-title">
                <span v-if="item.title">{{ item.title }}</span>
                <span v-else class="missing">(no title)</span>
              </h2>
              <div class="meta-chips">
                <a class="meta-chip is-link" :href="safeHref(item.scene_url)" target="_blank" rel="noreferrer">
                  <b-icon pack="mdi" icon="web" size="is-small" aria-hidden="true"/><span>{{ item.site }}</span>
                </a>
                <a v-if="item.members_url != ''" class="meta-chip is-link" :href="safeHref(item.members_url)" target="_blank" rel="noreferrer">
                  <b-icon pack="mdi" icon="link-lock" size="is-small" aria-hidden="true"/><span>{{$t('Members')}}</span>
                </a>
                <span class="meta-chip">
                  <b-icon pack="mdi" icon="calendar-outline" size="is-small" aria-hidden="true"/><span>{{ format(parseISO(item.release_date), "yyyy-MM-dd") }}</span>
                </span>
                <span v-if="item.duration" class="meta-chip">
                  <b-icon pack="mdi" icon="clock-outline" size="is-small" aria-hidden="true"/><span>{{ item.duration }} min</span>
                </span>
              </div>
            </header>

            <div class="details-toolbar">
              <b-field v-if="!displayingAlternateSource" class="rating-block">
                <star-rating :key="item.id" v-model="item.star_rating" :rating="item.star_rating" @rating-selected="setRating"
                             :increment="0.5" :star-size="20" :show-rating="false" />
                <b-tooltip :label="$t('Reset Rating')" position="is-right" :delay="250">
                  <b-icon pack="mdi" icon="autorenew" size="is-small" @click.native="setRating(0)" class="rating-reset"/>
                </b-tooltip>
              </b-field>
              <b-field v-if="displayingAlternateSource" class="rating-block">
                <strong>Linked scene, Not an XBVR Scene</strong>
              </b-field>
              <div class="action-row">
                <a class="button is-primary is-outlined is-small" @click="searchAlternateSourceScene()" title="Search for a different scene" v-if="displayingAlternateSource">
                  <b-icon pack="mdi" icon="movie-search-outline" size="is-small"/>
                </a>
                <a class="button is-primary is-outlined is-small" @click="scrapeScene()" title="Scrape and create an XBVR scene (not a link)" v-if="displayingAlternateSource">
                  <b-icon pack="mdi" icon="plus" size="is-medium"/>
                </a>
                <a class="button is-primary is-outlined is-small" @click="refreshExtRef()" title="Removes the scene.  Rescrape to refresh the scene data and relink" v-if="displayingAlternateSource">
                  <b-icon pack="mdi" icon="refresh" size="is-small"/>
                </a>
                <a class="button is-danger is-outlined is-small" @click="flagExtRefDeleted()" title="Unlinks the scene. It cannot be relinked to any scene. This cannot be undone" v-if="displayingAlternateSource">
                  <b-icon pack="mdi" icon="delete" size="is-small"/>
                </a>
                <hidden-button :item="item" v-if="!displayingAlternateSource"/>
                <watchlist-button :item="item" v-if="!displayingAlternateSource"/>
                <trailerlist-button :item="item" v-if="!displayingAlternateSource"/>
                <favourite-button :item="item" v-if="!displayingAlternateSource"/>
                <wishlist-button :item="item" v-if="!displayingAlternateSource"/>
                <watched-button :item="item" v-if="!displayingAlternateSource"/>
                <edit-button :item="item"/>
                <refresh-button :item="item" v-if="!displayingAlternateSource"/>
                <rescrape-button :item="item" v-if="!displayingAlternateSource"/>
                <b-tooltip :label="$t('Delete generated preview')" position="is-top" v-if="!displayingAlternateSource && item.has_preview">
                  <a class="button is-danger is-outlined is-small" @click="deletePreview()">
                    <b-icon pack="mdi" icon="video-off" size="is-small"/>
                  </a>
                </b-tooltip>
                <link-stashdb-button :item="item" objectType="scene" />
              </div>
            </div>

            <div class="image-row altsrc-row" v-if="alternateSources.length != 0">
              <div v-for="(altsrc, idx) in alternateSourcesWithTitles" :key="idx" class="altsrc-image-wrapper" @click="showExtRefScene(altsrc)">
                <b-tooltip type="is-light" :label="altsrc.title" :delay="100" append-to-body>
                  <vue-load-image>
                    <img slot="image" :src="getImageURL(altsrc.site_icon, '700x', altSourceIconContext(altsrc))" alt="Image" width="28px" />
                    <b-icon slot="error" pack="mdi" icon="link" size="is-small" />
                  </vue-load-image>
                </b-tooltip>
              </div>
            </div>

            <div class="image-row cast-strip" v-if="activeTab != 1 && !displayingAlternateSource">
              <div v-for="(image, idx) in castimages" :key="idx" class="image-wrapper">
                <b-tooltip  type="is-light" :label="image.actor_label"  :delay=100>
                  <vue-load-image>
                    <img slot="image" :src="getImageURL(image.src, '700x', 'act-' + image.actor_id)" alt="Image" class="thumbnail" @mouseover="showTooltip(idx)" @mouseout="hideTooltip(idx)" @click='showActorDetail([image.actor_id])' />
                    <img slot="preloader" :src="'/ui/images/blank.png'" style="height: 50px;display: block;margin-left:auto;margin-right: auto;" @click='showCastScenes([image.actor_name])' />
                    <img slot="error" src="/ui/images/blank_female_profile.png" width="80" @click='showActorDetail([image.actor_id])' />
                  </vue-load-image>
                </b-tooltip>

                <div v-if="image.visible" class="tooltip">
                  <img :src="getImageURL(image.src, '700x', 'act-' + image.actor_id)" alt="Tooltip Image" />
                </div>
              </div>
            </div>

            <div class="block-tags block chips" v-if="activeTab != 1">
              <b-taglist>
                <span v-for="(c, idx) in item.cast" :key="'cast' + idx" >
                  <a class="tag is-warning is-small" @click='showCastScenes([c.name])' :style="showOpenInNewWindow ? 'margin-right: 0;': 'margin-right: .5em;'" >{{ c.name }} ({{ c.avail_count }}/{{ c.count }})</a>
                  <a v-if="showOpenInNewWindow" class="tag is-warning is-small" :href='getCastScenesUrl([c.name])' target="_blank" style="margin-right: 0.5em;"><b-icon pack="mdi" icon="open-in-new" size="is-small"></b-icon></a>
                </span>
                <span>
                  <a @click='showSiteScenes([item.site])' class="tag is-primary is-small" :style="showOpenInNewWindow ? 'margin-right: 0;': 'margin-right: .5em;'">{{ item.site }}</a>
                  <a v-if="showOpenInNewWindow" class="tag is-primary is-small" :href='getSiteScenesUrl([item.site])' target="_blank" style="margin-right: 0.5em;"><b-icon pack="mdi" icon="open-in-new" size="is-small"></b-icon></a>
                </span>
                <span v-for="(tag, idx) in item.tags" :key="'tag' + idx">
                  <a  @click='showTagScenes([tag.name])' class="tag is-info is-small" :style="showOpenInNewWindow ? 'margin-right: 0;': 'margin-right: .5em;'">{{ tag.name }} ({{ tag.count }})</a>
                  <a v-if="showOpenInNewWindow" class="tag is-info is-small" :href='getTagScenesUrl([tag.name])' target="_blank" style="margin-right: 0.5em;"><b-icon pack="mdi" icon="open-in-new" size="is-small"></b-icon></a>
                </span>
              </b-taglist>              
            </div>

            <div class="block-tags block" v-if="activeTab == 1">
             <b-taglist>
              <b-tooltip  type="is-danger" :label="disableSaveMsg()" position="is-right" :delay=250 :active="disableSaveButtons()">
                <b-button @click="updateCuepoint(false)" class="tag is-info is-small is-warning" accesskey="a" :disabled="disableSaveButtons()" >
                  <u>A</u>dd New
                </b-button>
              </b-tooltip>
                <b-button @click="vidPosition = new Date(0,0,0,0,0, 0, player.currentTime() * 1000)" class="tag is-info is-small is-warning" accesskey="t">Current <u>T</u>ime</b-button>
              <b-tooltip type="is-danger" :label="$t(disableSaveMsg())" position="is-right" :delay=250 :active="disableSaveButtons()">
                <b-button v-if="currentCuepointId > 0" @click="updateCuepoint(true)" class="tag is-info is-small is-warning" accesskey="s"
                  :disabled="disableSaveButtons()" >
                  <u>S</u>ave Edit
                </b-button>
              </b-tooltip>
                <b-button v-if="cuepointName!=''" @click='cuepointName=""' class="tag is-info is-small is-warning" >Clear Cuepoint Name</b-button>
                <b-button v-if="tagAct!=''" @click='setCuepointName("")' class="tag is-info is-small is-warning" accesskey="c"><u>C</u>lear Action</b-button>
              </b-taglist>
            </div>

            <div class="is-divider" data-content="Cuepoint Positions" v-if="activeTab == 1"></div>
            <div class="block-tags block" v-if="activeTab == 1">
              <b-taglist>
                <b-button v-for="(c, idx) in cuepointPositionTags.slice(1)" :key="'pos' + idx" @click='setCuepointName([c])' class="tag is-info is-small">{{c}}</b-button>
              </b-taglist>
            </div>
            <div class="is-divider" data-content="Default Cuepoint Actions" v-if="activeTab == 1"></div>
            <div class="block-tags block" v-if="activeTab == 1">
              <b-taglist>
                <b-button v-for="(c, idx) in cuepointActTags.slice(1)" :key="'action' + idx" @click='setCuepointName([c])' class="tag is-info is-small">{{c}}</b-button>
              </b-taglist>
            </div>
            <div class="is-divider" data-content="Cast Cuepoints" v-if="activeTab == 1"></div>
            <div class="block-tags block" v-if="activeTab == 1">
              <b-taglist>
                <b-button v-for="(c, idx) in item.cast" :key="'cast' + idx" @click='setCuepointName([c.name])' class="tag is-info is-small">{{c.name}}</b-button>
              </b-taglist>
            </div>
            <div class="is-divider" data-content="Scene Tag Cuepoints" v-if="activeTab == 1"></div>
            <div class="block-tags block" v-if="activeTab == 1">
              <b-taglist>
                <b-button v-for="(tag, idx) in item.tags" :key="'tag' + idx" @click='setCuepointName([tag.name])'
                   class="tag is-info is-small">{{ tag.name }}</b-button>
              </b-taglist>
            </div>


            <div class="block-opts block">
              <b-tabs v-model="activeTab" :animated="false">

                <b-tab-item :label="`Files (${fileCount})`" v-if="!displayingAlternateSource">
                  <div class="block-tab-content block">
                    <div class="content media is-small" v-for="(f, idx) in filesByType" :key="idx">
                      <div class="media-left">
                        <button rounded class="button is-success is-small" @click='playFile(f)'
                                v-show="f.type === 'video'">
                          <b-icon pack="fas" icon="play" size="is-small"></b-icon>
                        </button>
                        <b-tooltip :label="$t('Select this script for export')" position="is-right">
                        <button rounded class="button is-info is-small is-outlined" @click='selectScript(f)'
                          v-show="f.type === 'script'" v-bind:class="{ 'is-success': f.is_selected_script, 'is-info' :!f.is_selected_script }">
                          <b-icon pack="mdi" icon="pulse"></b-icon>
                        </button>
                        </b-tooltip>
                        <button rounded class="button is-info is-small is-outlined" disabled
                                v-show="f.type === 'hsp'">
                          <b-icon pack="mdi" icon="safety-goggles"></b-icon>
                        </button>
                        <button rounded class="button is-info is-small is-outlined" disabled
                                v-show="f.type === 'subtitles'">
                          <b-icon pack="mdi" icon="subtitles"></b-icon>
                        </button>
                      </div>
                      <div class="media-content" style="overflow-wrap: break-word;">
                        <strong>{{ f.filename }}</strong><br/>
                        <small>
                          <span class="pathDetails">{{ f.path }}</span>
                          <br/>
                          {{ prettyBytes(f.size) }}<span v-if="f.type === 'video'"> ({{ prettyBytes(f.video_bitrate, { bits: true })  }}/s)</span>,
                          <span v-if="f.type === 'video'"><span class="videosize">{{ f.video_width }}x{{ f.video_height }} {{ f.video_codec_name }}</span>, {{ f.projection }},&nbsp;</span>
                          <span v-if="f.duration > 1">{{ humanizeSeconds(f.duration) }},</span>
                          {{ format(parseISO(f.created_time), "yyyy-MM-dd") }}
                        </small>
                        <div v-if="f.type === 'script' && f.has_heatmap" class="heatmapFunscript">
                          <img :src="getHeatmapURL(f.id)"/>
                        </div>
                      </div>
                      <div class="media-right">
                        <button class="button is-dark is-small is-outlined" title="Unmatch file from scene" @click='unmatchFile(f)'>
                          <b-icon pack="fas" icon="unlink" size="is-small"></b-icon>
                        </button>&nbsp;
                        <button class="button is-danger is-small is-outlined" title="Delete file from disk" @click='removeFile(f)'>
                          <b-icon pack="fas" icon="trash" size="is-small"></b-icon>
                        </button>
                      </div>
                    </div>
                  </div>
                </b-tab-item>

                <b-tab-item :label="`Cuepoints (${sortedCuepoints.length})`" v-if="!displayingAlternateSource">
                  <div class="block-tab-content block">
                    <div class="block" >
                      <div class="columns">
                        <div class="column is-2">
                        <b-field label="Track" width="7.25em" label-position="on-border">
                          <b-input v-model="track" width="7.25em"></b-input>
                        </b-field>
                        </div>
                        <div class="column">
                        <b-field label="Name" label-position="on-border">
                          <b-autocomplete v-model="cuepointName" :data="filteredCuepointPositionList" :open-on-focus="true"></b-autocomplete>
                        </b-field>
                        </div>
                        <div class="column is-2">
                        <b-field label="Start" label-position="on-border">
                          <b-timepicker v-model="vidPosition" rounded editable placeholder="Defaults to player position" hour-format="24" :enable-seconds="true" :max-time="maxTime" :time-formatter="timeFormatter" :time-parser="timeParser" >
                          <b-button
                            label="Current Time"
                            type="is-primary"
                            @click="vidPosition = new Date(0,0,0,0,0, 0, player.currentTime() * 1000)" />
                          </b-timepicker>
                        </b-field>
                        </div>
                        <div class="column is-2">
                          <b-field label="End" label-position="on-border">
                          <b-timepicker v-model="endTime" rounded editable placeholder="Defaults to player position" hour-format="24" :enable-seconds="true" :max-time="maxTime" :time-formatter="timeFormatter" :time-parser="timeParser" >
                          <b-button
                            label="Current Time"
                            type="is-primary"
                            @click="endTime = new Date(0,0,0,0,0, 0, player.currentTime() * 1000)" />
                          </b-timepicker>
                        </b-field>
                        </div>
                      </div>
                    </div>
                    <div>
                      <!-- :sort-multiple="sortMultiple" :sort-multiple-data="cuepointSorting" -->
                        <b-table :data="sortedCuepoints"  :narrowed=true :per-page=7 focusable striped sticky-header
                          @select="cuepointSelected">
                          <!-- paginated  pagination-position="top" :pagination-rounded=true pagination-size="is-small" -->
                          <b-table-column field="track" label="Track" width="7.25em" v-slot="props" >
                            {{ props.row.track ==null ? "" :  props.row.track }}
                          </b-table-column>
                          <b-table-column field="name" label="Name" v-slot="props"  is-small>
                            {{ props.row.name }}
                          </b-table-column>
                          <b-table-column field="time_start" label="Start" v-slot="props" width="6.5em"  >
                            {{ humanizeSeconds1DP(props.row.time_start) }}
                          </b-table-column>
                          <b-table-column field="time_end" label="End" v-slot="props" width="6.5em"  >
                            {{ props.row.time_end==null ? "" :  humanizeSeconds1DP(props.row.time_end) }}
                          </b-table-column>
                          <b-table-column field="rating" v-slot="props" width="7em"  >
                            <b-field v-if="props.row.track!=null">
                              <star-rating :key="props.row.id" v-model="props.row.rating" :rating="props.row.rating" @rating-selected="setCuepointRating(props.row)" :increment="0.5" :star-size="10" />
                              <b-icon v-if="props.row.rating>0" pack="mdi" icon="autorenew" size="is-small" @click.native="clearCuepointRating(props.row)" style="padding-left: .25em;padding-top: .5em;"/>
                            </b-field>
                          </b-table-column>
                          <b-table-column v-slot="props" width="1em" >
                            <button class="button is-danger is-outlined is-small" @click="deleteCuepoint(props.row.id)" title="Delete cuepoint">
                              <b-icon pack="fas" icon="trash" />
                            </button>
                          </b-table-column>
                        </b-table>
                    </div>
                  </div>
                </b-tab-item>

                <b-tab-item label="Watch history" v-if="!displayingAlternateSource">
                  <div class="block-tab-content block">
                    <div>
                      {{ historySessionsCount }} view sessions, total duration
                      {{ humanizeSeconds(historySessionsDuration) }}
                    </div>
                    <div class="content is-small">
                      <div class="block" v-for="(session, idx) in item.history" :key="idx">
                        <strong>{{ format(parseISO(session.time_start), "yyyy-MM-dd HH:mm:ss") }} -
                          {{ humanizeSeconds(session.duration) }}</strong>
                      </div>
                    </div>
                  </div>
                </b-tab-item>

                <b-tab-item label="Description">
                  <div class="block-tab-content block">
                    <b-message>
                      {{ item.synopsis }}
                    </b-message>
                  </div>
                </b-tab-item>
                <b-tab-item v-if="this.$store.state.optionsAdvanced.advanced.showSceneSearchField && !displayingAlternateSource" label="Search fields">
                  <div class="block-tab-content block">
                    <div class="content is-small">
                      <div class="block" v-for="(field, idx) in searchfields" :key="idx">
                        <strong>{{ field.fieldName }} - </strong> {{ field.fieldValue }}
                      </div>
                    </div>
                  </div>
                </b-tab-item>

              </b-tabs>
            </div>

          </div>
        </div>
      </section>
      <div class="scene-id">
        {{ item.scene_id }}
        <span  v-if="this.$store.state.optionsAdvanced.advanced.showInternalSceneId">{{ $t('Internal ID') }}: {{item.id}}</span>
        <a v-if="this.$store.state.optionsAdvanced.advanced.showHSPApiLink" :href="`/heresphere/${item.id}`" target="_blank" rel="noreferrer" style="margin-left:0.5em">
          <img src="/ui/icons/heresphere_24.png" style="height:15px;"/>
        </a>
      </div>
    </div>
    <button class="modal-close is-large" aria-label="close" @click="close()"></button>
    <button type="button" class="scene-nav prev" @click="prevScene" v-if="$store.getters['sceneList/prevScene'](item) !== null && !displayingAlternateSource"
            title="Keyboard shortcut: O" aria-label="Previous scene">
      <b-icon pack="mdi" icon="chevron-left" size="is-medium" aria-hidden="true"/>
    </button>
    <button type="button" class="scene-nav next" @click="nextScene" v-if="$store.getters['sceneList/nextScene'](item) !== null && !displayingAlternateSource"
            title="Keyboard shortcut: P" aria-label="Next scene">
      <b-icon pack="mdi" icon="chevron-right" size="is-medium" aria-hidden="true"/>
    </button>
  </div>
</template>

<script>
import api from '../../api'
import { encodeJsonBase64 } from '../../util/base64'
import { getImageURL as getImageURLUtil, altSourceIconContext, sceneContext, humanizeSeconds, humanizeSeconds1DP } from '../../util/image'
import { safeHref } from '../../util/url'
import videojs from 'video.js'
import 'videojs-hotkeys'
import 'videojs-vr/dist/videojs-vr.min.js'
import { format, formatDistance, parseISO } from 'date-fns'
import prettyBytes from 'pretty-bytes'
import VueLoadImage from 'vue-load-image'
import GlobalEvents from 'vue-global-events'
import StarRating from 'vue-star-rating'
import FavouriteButton from '../../components/FavouriteButton'
import LinkStashdbButton from '../../components/LinkStashdbButton'
import WatchlistButton from '../../components/WatchlistButton'
import WishlistButton from '../../components/WishlistButton'
import WatchedButton from '../../components/WatchedButton'
import EditButton from '../../components/EditButton'
import RefreshButton from '../../components/RefreshButton'
import RescrapeButton from '../../components/RescrapeButton'
import TrailerlistButton from '../../components/TrailerlistButton'
import HiddenButton from '../../components/HiddenButton'

export default {
  name: 'Details',
  components: { VueLoadImage, GlobalEvents, StarRating, WatchlistButton, FavouriteButton, LinkStashdbButton, WishlistButton, WatchedButton, EditButton, RefreshButton, RescrapeButton, TrailerlistButton, HiddenButton },
  data () {
    return {
      index: 1,
      activeTab: 0,
      activeMedia: 0,
      player: {},
      tagAct: '',
      cuepointName: '',
      cuepointRating: 0,
      cuepointPositionTags: ['', 'standing', 'sitting', 'laying', 'kneeling'],
      cuepointActTags: ['', 'handjob', 'blowjob', 'doggy', 'cowgirl', 'revcowgirl', 'missionary', 'titfuck', 'anal', 'cumshot', '69', 'facesit'],
      carouselSlide: 0,
      vidPosition: null,
      skipForwardIntervals: [5, 10, 30, 60, 120, 300],
      skipBackIntervals: [-300, -120, -60, -30, -10, -5],
      lastSkipFowardInterval: 5,
      lastSkipBackInterval: -5,
      currentCuepointId: 0,
      maxTime: new Date(0, 0, 0, 5, 0, 0),
      cuepointSorting: [{ field: "is_hsp", order: "asc" },{ field: "time_start", order: "desc" }, {field: "track", order: "desc"}, {field: "time_end", order: "desc"}],
      trackInput: '',
      track: null,
      endTime: null,
      sortMultiple: true,
      castimages: [],
      searchfields: [],
      alternateSources: [],
      waitingForQuickFind: false,
    }
  },
  computed: {
    item () {
      const item = this.$store.state.overlay.details.scene
      if (item == null) {
        return item
      }
      // don't sort the store's array in place; work on a shallow copy
      if (this.$store.state.optionsWeb.web.tagSort === 'alphabetically') {
        return { ...item, tags: [...item.tags].sort((a, b) => a.name < b.name ? -1 : 1) }
      }
      let releasedate = parseISO(item.release_date)
      let imgs = item.cast.map((actor) => {
        let birthdate = parseISO(actor.birth_date)
        let label = actor.name
        if (birthdate.getFullYear() > 0) {
          let age = releasedate.getFullYear() - birthdate.getFullYear()
          if ((releasedate.getMonth() < birthdate.getMonth()) || (releasedate.getMonth() == birthdate.getMonth() && releasedate.getDate() < birthdate.getDate())) {
            age -= 1
          }
          label += `, ${age} in scene`
        }
        let img = actor.image_url
        if (img == "" ){
          img = "blank"  // forces an error image to load, blank won't display an image
        }
        if (actor.name.startsWith("aka:")) {
          img = ""
        }
        return {src: img, visible: false, actor_name: actor.name, actor_label: label, actor_id: actor.id};
      });

      this.castimages = imgs.filter((img) => {
        return img.src !== '';
        });
      return item
    },
    // Properties for gallery
    images () {
      if (this.item.images=="null") {
        return "[]"
      }
      return JSON.parse(this.item.images).filter(im => im && im.url)
    },
    // Tab: cuepoints
    sortedCuepoints () {
      if (this.item.cuepoints !== null) {
        for (let i = 0; i < this.item.cuepoints.length; i++) {
          this.item.cuepoints[i].is_hsp = this.item.cuepoints[i].track == null ? 0 : 1
        }
        let x=this.item.cuepoints.slice().sort((a, b) => (a.time_start > b.time_start) ? 1 : -1 || (a.is_hsp >b.is_hsp) ? 1 : -1 )
        x=this.item.cuepoints.slice().sort((a,b) => {
          let compare = (a.is_hsp<b.is_hsp) ? -1 : (a.is_hsp>b.is_hsp) ? 1 : 0
          if (compare!=0) {
            return compare
          }
          compare = (a.time_start<b.time_start) ? -1 : (a.time_start>b.time_start) ? 1 : 0
          if (compare!=0) {
            return compare
          }
          compare = (a.track<b.track) ? -1 : (a.track>b.track) ? 1 : 0
          if (compare!=0) {
            return compare
          }
          return  (a.time_end<b.time_end) ? -1 : (a.time_end>b.time_end) ? 1 : 0
        })
        return x
      }
      return []
    },
    // Tab: files
    fileCount () {
      if (this.item.file !== null) {
        return this.item.file.length
      }
      return 0
    },
    filesByType () {
      if (this.item.file !== null) {
        return this.item.file.slice().sort((a, b) => (a.type === 'video') ? -1 : 1)
      }
      return []
    },
    // Tab: history
    historySessionsCount () {
      if (this.item.history !== null) {
        return this.item.history.length
      }
      return 0
    },
    historySessionsDuration () {
      if (this.item.history !== null) {
        let total = 0
        this.item.history.slice().map(i => {
          total = total + i.duration
          return 0
        })
        return total
      }
      return 0
    },
    showEdit () {
      return this.$store.state.overlay.edit.show
    },
    filteredCuepointPositionList () {
      // filter the list of positions based on what has been entered so far
      let list=this.cuepointActTags.concat(this.cuepointPositionTags)
      return list.filter((option) => {
        return option
          .toString()
          .toLowerCase()
          .trim()
          .indexOf(this.cuepointName.toString().toLowerCase()) >= 0
      })
    },
    displayingAlternateSource () {
      // displayingAlternateSource indicates we aren't displaying a real xbvr scene from the scenes table,
      //  so functions like watchlist, ratings, etc don't apply
      // we are displaying scene data serialized and saved in the external_references table
      if ( this.$store.state.overlay.details.altsrc != null) return true
      return false
    },
    changeDetailsTab() {      
      return this.$store.state.overlay.changeDetailsTab
    },
    quickFindOverlayState() {
      return this.$store.state.overlay.quickFind.show
    },
    showOpenInNewWindow () {
      return this.$store.state.optionsWeb.web.showOpenInNewWindow
    },
    alternateSourcesWithTitles() {
      return this.alternateSources.map(altsrc => {
        const extdata = JSON.parse(altsrc.external_data);
        return {
          ...altsrc,
          title: extdata.scene?.title || 'No Title'
        };
      });
    }
  },
  mounted () {
    this.setupPlayer()
    this.loadAlternateSources()
    // the 'item.id' watcher below is not immediate, so fetch search fields
    // for the initially displayed scene here — otherwise the Search fields
    // tab stays empty until the user navigates to another scene
    if (this.item && this.item.id) {
      this.getSearchFields(this.item.id)
    }

    // load default cuepoint actions & positions from kv entry in the db
    api.get('/options/cuepoints').json().then(data => {
      this.cuepointActTags = data.actions
      this.cuepointPositionTags = data.positions
      this.cuepointActTags.unshift("")
      this.cuepointPositionTags.unshift("")
      })    
},
watch:{
  // refetch alternate sources and search fields when the displayed scene changes
  'item.id': function (newVal, oldVal) {
    if (newVal && newVal !== oldVal) {
      this.loadAlternateSources()
      this.getSearchFields(newVal)
    }
  },
  quickFindOverlayState(newVal, oldVal){
    if (newVal == true) {
      return
    }
    if (this.waitingForQuickFind){
      this.waitingForQuickFind = false
      if (this.$store.state.overlay.quickFind.selectedScene != null && this.$store.state.overlay.quickFind.selectedScene.id > 0) {
        this.$buefy.dialog.confirm({
          title: 'Relink scene',
          message: `Do you wish to link this scene to <strong>${this.$store.state.overlay.quickFind.selectedScene.title}</strong>`,
          type: 'is-info is-wide',
          hasIcon: true,
          onConfirm: () => {
            this.handleRelinkExtRef()
          }
        })
      }
    }
  },
  changeDetailsTab(newVal, oldVal){
    if (newVal == -1 ) {
      return
    }
    this.activeTab = newVal
    this.$store.commit('overlay/changeDetailsTab', { tab: -1 })
  },
  activeMedia(newVal, oldVal) {
    // Auto-load first video when Player tab is opened (without auto-playing)
    // The webUI video player doesn't work for some users without autoloading
    if (newVal === 1 && !this.displayingAlternateSource) {
      const videoFiles = this.filesByType.filter(f => f.type === 'video')
      if (videoFiles.length > 0) {
        this.activeMedia = 1
        this.updatePlayer('/api/dms/file/' + videoFiles[0].id + '?dnt=true', (videoFiles[0].projection == 'flat' ? 'NONE' : '180'))
      }
    }
  }
},
  methods: {
    safeHref,
    // Fetched on mount and whenever the displayed scene changes. Previously
    // an async computed used as v-if — an always-truthy Promise that
    // refetched on every render.
    async loadAlternateSources () {
      this.alternateSources = [];
      if (this.displayingAlternateSource) return
      if (!this.item || !this.item.id) return
      try {
        const response = await api.get('/scene/alternate_source/' + this.item.id).json();
        if (response == null) {
          return
        }
        response.forEach(altsrc => {
          if (altsrc.external_source.startsWith("alternate scene ")) {
            this.alternateSources.push(altsrc)
          }
        });
      } catch (error) {
        // leave the view without alternate sources on error
      }
    },
    setupPlayer () {
      this.player = videojs(this.$refs.player, {
        aspectRatio: '1:1',
        fluid: true,
        loop: true
      })

      this.player.hotkeys({
        alwaysCaptureHotkeys: true,
        volumeStep: 0.1,
        seekStep: 5,
        enableModifiersForNumbers: false,
        enableVolumeScroll: false,
        customKeys: {
          closeModal: {
            key: function (event) {
              return event.which === 27
            },
            handler: (player, options, event) => {
              if (!this.displayingAlternateSource) this.player.dispose()
              this.$store.commit('overlay/hideDetails')
            }
          },
          zoomIn: {
            handler: (player, options, event) => {
              this.zoomHandler(true)
            }
          },
          zoomOut: {
            handler: (player, options, event) => {
              this.zoomHandler(false)
            }
          }
        }
      })

      const videoElement = this.player.el();
      videoElement.addEventListener('wheel', this.zoomHandlerWeb.bind(this))
    },

    zoomHandlerWeb(event) {
      event.preventDefault();
      this.zoomHandler(event.deltaY < 0)
    },

    zoomHandler(isZoomingIn) {
      const vr = this.player.vr()
      const minFov = 30
      const maxFov = 130
      let fov = vr.camera.fov + (isZoomingIn ? -1 : 1)

      if (fov < minFov) {
        fov = minFov
      }

      if (fov > maxFov) {
        fov = maxFov
      }

      vr.camera.fov = fov;
      vr.camera.updateProjectionMatrix()
    },
    updatePlayer (src, projection) {
      this.player.reset()
      /* const vr = */ this.player.vr({
        projection: projection,
        forceCardboard: false
      })

      this.player.on('loadedmetadata', function () {
        // vr.camera.position.set(-1, 0, 2);
      })

      if (src) {
        this.player.src({ src: src, type: 'video/mp4' })
      }
      this.player.poster(this.getImageURL(this.item.cover_url, 'raw', sceneContext(this.item.scene_id)))
    },
    showCastScenes (actor) {
      this.$store.commit('sceneList/setCastFilterOnly', actor)
      this.$router.push({
        name: 'scenes',
        query: { q: this.$store.getters['sceneList/filterQueryParams'] }
      })
      this.close()
    },
    getCastScenesUrl(actor) {
      let newfilters = Object.assign({}, this.$store.state.sceneList.filters);
      newfilters.cast = actor;       
      newfilters.sites = []
      newfilters.tags = []
      newfilters.attributes = []
      return this.$router.resolve({
        name: 'scenes',
        query: { q: encodeJsonBase64(newfilters) }
      }).href
    },
    showTagScenes (tag) {
      this.$store.state.sceneList.filters.cast = []
      this.$store.state.sceneList.filters.sites = []
      this.$store.state.sceneList.filters.tags = tag
      this.$store.state.sceneList.filters.attributes = []
      this.$router.push({
        name: 'scenes',
        query: { q: this.$store.getters['sceneList/filterQueryParams'] }
      })
      this.close()
    },
    getTagScenesUrl(tag) {
      let newfilters = Object.assign({}, this.$store.state.sceneList.filters);      
      newfilters.tags = tag;       
      newfilters.cast = []       
      newfilters.sites = []
      newfilters.attributes = []
      return this.$router.resolve({
        name: 'scenes',
        query: { q: encodeJsonBase64(newfilters) }
      }).href
    },
    showSiteScenes (site) {
      this.$store.state.sceneList.filters.cast = []
      this.$store.state.sceneList.filters.sites = site
      this.$store.state.sceneList.filters.tags = []
      this.$store.state.sceneList.filters.attributes = []
      this.$router.push({
        name: 'scenes',
        query: { q: this.$store.getters['sceneList/filterQueryParams'] }
      })
      this.close()
    },
    getSiteScenesUrl(site) {
      let newfilters = Object.assign({}, this.$store.state.sceneList.filters);
      newfilters.sites = site;       
      newfilters.cast = []       
      newfilters.tags = []
      newfilters.attributes = []
      return this.$router.resolve({
        name: 'scenes',
        query: { q: encodeJsonBase64(newfilters) }
      }).href
    },
    showActorDetail (actor_id) {
      api.get('/actor/'+actor_id).json().then(data => {
        if (data.id != 0){
          this.$store.commit('overlay/showActorDetails', { actor: data })
          this.close()
        }
      })
    },
    playPreview () {
      this.activeMedia = 1
      this.updatePlayer('/api/dms/preview/' + this.item.scene_id, 'NONE')
      this.player.play()
    },
    playFile (file) {
      this.activeMedia = 1
      this.updatePlayer('/api/dms/file/' + file.id + '?dnt=true', (file.projection == 'flat' ? 'NONE' : '180'))
      this.player.play()
    },
    unmatchFile (file) {
      this.$buefy.dialog.confirm({
        title: 'Unmatch file',
        message: `You're about to unmatch the file <strong>${file.filename}</strong> from this scene. Afterwards, it can be matched again to this or any other scene.`,
        type: 'is-info is-wide',
        hasIcon: true,
        onConfirm: () => {
          api.post(`/files/unmatch`, {json:{file_id: file.id}}).json().then(data => {
            this.$store.commit('overlay/showDetails', { scene: data })
          })
        }
      })
    },
    removeFile (file) {
      this.$buefy.dialog.confirm({
        title: 'Remove file',
        message: `You're about to remove file <strong>${file.filename}</strong> from <strong>disk</strong>.`,
        type: 'is-danger',
        hasIcon: true,
        onConfirm: () => {
          api.delete(`/files/file/${file.id}`).json().then(data => {
            this.$store.commit('overlay/showDetails', { scene: data })
          })
        }
      })
    },
    deletePreview () {
      this.$buefy.dialog.confirm({
        title: 'Delete preview',
        message: `You're about to delete the generated preview for this scene. It can be regenerated at any time.`,
        type: 'is-danger',
        hasIcon: true,
        onConfirm: () => {
          api.delete(`/scene/${this.item.id}/preview`).json().then(data => {
            this.$store.commit('sceneList/updateScene', data)
            this.$store.commit('overlay/showDetails', { scene: data })
          })
        }
      })
    },
    selectScript (file) {
      api.post(`/scene/selectscript/${this.item.id}`, {
        json: {
          file_id: file.id,
        }
      }).json().then(data => {
          this.$store.commit('overlay/showDetails', { scene: data })
      })
    },
    getImageURL (u, size, context = 'scene-0') {
      return getImageURLUtil(u, size, context)
    },
    altSourceIconContext,
    sceneContext,
    getIndicatorURL (idx) {
      if (this.images[idx] !== undefined) {
        return this.getImageURL(this.images[idx].url, 'x40', sceneContext(this.item.scene_id))
      } else {
        return '/ui/images/blank.png'
      }
    },
    getHeatmapURL (fileId) {
      return `/api/dms/heatmap/${fileId}`
    },
    playCuepoint (cuepoint) {
      // populate the cuepoint edit fields
      this.vidPosition = new Date(0, 0, 0, 0, 0, 0, cuepoint.time_start*1000)
      this.endTime = new Date(0, 0, 0, 0, 0, 0, cuepoint.time_end*1000)
      this.currentCuepointId = cuepoint.id
      this.cuepointRating = cuepoint.rating
      if (cuepoint.name.indexOf('-') > 0) {
        this.cuepointName = cuepoint.name.substr(0, cuepoint.name.indexOf('-'))
        this.tagAct = cuepoint.name.substr(cuepoint.name.indexOf('-') + 1)
      } else {
        this.tagAct = cuepoint.name
        this.cuepointName = ''
      }
      // now mow the player position
      this.player.currentTime(cuepoint.time_start)
      this.player.play()
    },
    updateCuepoint (editCuepoint) {
      if (this.disableSaveButtons()) return
      // if edit choosen, delete existing cuepoint before add
      if (editCuepoint && this.currentCuepointId > 0) {
        this.deleteCuepoint(this.currentCuepointId)
      }
      let name =  this.cuepointName
      let pos = this.player.currentTime()
      let endpos=null
      this.track=parseInt(this.track)
      if (this.vidPosition != null) {
        pos = (this.vidPosition.getMilliseconds() / 1000) + this.vidPosition.getSeconds() + (this.vidPosition.getMinutes() * 60) + (this.vidPosition.getHours() * 60 * 60)
      }
      if (this.endTime != null) {
        endpos = (this.endTime.getMilliseconds() / 1000) + this.endTime.getSeconds() + (this.endTime.getMinutes() * 60) + (this.endTime.getHours() * 60 * 60)
      }
      this.currentCuepointId = 0

      api.post(`/scene/${this.item.id}/cuepoint`, {
        json: {
          track: this.track,
          name: name,
          time_start: pos,
          time_end: endpos,
          rating: this.cuepointRating
        }
      }).json().then(data => {
        this.vidPosition = null
        this.endTime = null
        this.cuepointName=''
        this.track = null
        this.$store.commit('sceneList/updateScene', data)
        this.$store.commit('overlay/showDetails', { scene: data })
      })
    },
    deleteCuepoint (cuepointid) {
      api.delete(`/scene/${this.item.id}/cuepoint/${cuepointid}`)
        .json().then(data => {
          this.$store.commit('sceneList/updateScene', data)
          this.$store.commit('overlay/showDetails', { scene: data })
        })
    },
    close () {
      if (!this.displayingAlternateSource) this.player.dispose()
      this.$store.commit('overlay/hideDetails')
    },
    humanizeSeconds,
    humanizeSeconds1DP,
    setRating (val) {
      api.post(`/scene/rate/${this.item.id}`, { json: { rating: val } })

      const updatedScene = Object.assign({}, this.item)
      updatedScene.star_rating = val
      this.item.star_rating = val
      this.$store.commit('sceneList/updateScene', updatedScene)
    },
    nextScene () {
      const data = this.$store.getters['sceneList/nextScene'](this.item)
      if (data !== null && !this.displayingAlternateSource) {
        this.$store.commit('overlay/showDetails', { scene: data })
        this.activeMedia = 0
        this.carouselSlide = 0
        this.updatePlayer(undefined, '180')
      }
    },
    prevScene () {
      const data = this.$store.getters['sceneList/prevScene'](this.item)
      if (data !== null && !this.displayingAlternateSource) {
        this.$store.commit('overlay/showDetails', { scene: data })
        this.activeMedia = 0
        this.carouselSlide = 0
        this.updatePlayer(undefined, '180')
      }
    },
    playerStepBack (interval) {
      const wasPlaying = !this.player.paused()
      if (wasPlaying) {
        this.player.pause()
      }
      let seekTime = this.player.currentTime() + interval
      if (seekTime <= 0) {
        seekTime = 0
      }
      this.player.currentTime(seekTime)
      if (wasPlaying) {
        this.player.play()
      }
      this.lastSkipBackInterval = interval
    },
    playerStepForward (interval) {
      const duration = this.player.duration()
      const wasPlaying = !this.player.paused()
      if (wasPlaying) {
        this.player.pause()
      }
      let seekTime = this.player.currentTime() + interval
      if (seekTime >= duration) {
        seekTime = wasPlaying ? duration - 0.001 : duration
      }
      this.player.currentTime(seekTime)
      if (wasPlaying) {
        this.player.play()
      }
      this.lastSkipFowardInterval = interval
    },
    setCuepointName (param) {
      if (this.activeTab === 1) {
        if (this.cuepointName=='') {
          this.cuepointName = param.toString()
        }else{
          this.cuepointName = this.cuepointName+'-'+param.toString()
        }
      }
    },
    toggleGallery () {
      if (this.activeMedia == 0) {
        this.activeMedia = 1
      } else {
        this.activeMedia = 0
        }
    },
    handleLeftArrow () {
      if (this.activeMedia === 0)
      {
        this.carouselSlide = this.carouselSlide - 1
      } else {
        this.playerStepBack(this.lastSkipBackInterval)
      }
    },
    handleRightArrow () {
      if (this.activeMedia === 0)
      {
        this.carouselSlide = this.carouselSlide + 1
      } else {
        this.playerStepForward(this.lastSkipFowardInterval)
      }
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
    timeFormatter(time) {
       return new Intl.DateTimeFormat('en', { hourCycle: 'h23', hour: "2-digit", minute: "2-digit", second: "2-digit", fractionalSecondDigits: 1 }).format(time)
    },
    timeParser(inputString) {
      let items = inputString.split(":")
      return new Date(0, 0, 0, items[0],items[1], 0, items[2]*1000)
    },
    cuepointSelected(cuepoint) {
      // populate the cuepoint edit fields
      this.vidPosition = new Date(0, 0, 0, 0, 0, 0, cuepoint.time_start*1000)
      this.endTime = new Date(0, 0, 0, 0, 0, 0, cuepoint.time_end*1000)
      this.currentCuepointId = cuepoint.id
      this.cuepointName = cuepoint.name
      this.track=cuepoint.track
      this.cuepointRating=cuepoint.rating
      // now mow the player position
      this.player.currentTime(cuepoint.time_start)
      this.player.play()
    },
    disableSaveButtons() {
      if (this.track!=null && this.track!="" && (isNaN(this.endTime) || this.endTime==null)) return true
      if ((this.track==null || this.track==="") && !isNaN(this.endTime) && this.endTime!=null) return true
      return false
    },
    disableSaveMsg() {
      if (this.track!=null && this.track!="" && (isNaN(this.endTime) || this.endTime==null)) return "Specify a End Time"
      if ((this.track==null || this.track==="") && !isNaN(this.endTime) && this.endTime!=null) return "End Time is only valid for HSP Cuepoints"
      return ""
    },
    setCuepointRating (row) {
      this.cuepointSelected(row)
      this.updateCuepoint(true)
    },
    clearCuepointRating (row) {
      row.rating=0
      this.cuepointSelected(row)
      this.updateCuepoint(true)
    },
    showTooltip(idx) {
      this.castimages[idx].visible = true;
    },
    hideTooltip(idx) {
      this.castimages[idx].visible = false;
    },
    getSearchFields(id) {
      // load search fields
      this.searchfields = []      
      if (this.$store.state.optionsAdvanced.advanced.showSceneSearchField && !this.displayingAlternateSource) {
        api.get('/scene/searchfields', {
          searchParams: {
            q: id
          },
          }).json().then(data => {
            this.searchfields = data
          })
      }
    },
    showExtRefScene (altsrc) {      
      const extdata = JSON.parse(altsrc.external_data);      
      if (extdata.scene.cast == null) 
      {
        extdata.scene.cast = []
      }
      this.$store.commit('overlay/showDetails', { scene: extdata.scene, altsrc: altsrc, prevscene: this.item, query_for_altsrc: extdata.query })
      this.activeTab = 0      
    },
    searchAlternateSourceScene() {
      // search for a new scene to link to the alternate source scene
      const  q = this.$store.state.overlay.details.query_for_altsrc == "" ? this.item.title : this.$store.state.overlay.details.query_for_altsrc      
      this.$store.commit('overlay/showQuickFind', { searchString:  q, displaySelectedScene: false })
      this.waitingForQuickFind = true
    }, 
    async handleRelinkExtRef() {
      const response = await api.post(`/extref/edit_link`, {
        json: {
          external_source: this.$store.state.overlay.details.altsrc.external_source,
          external_id: this.$store.state.overlay.details.altsrc.external_id,
          internal_table: "scenes",
          internal_db_id: this.$store.state.overlay.quickFind.selectedScene.id,
          internal_name_id: this.$store.state.overlay.quickFind.selectedScene.scene_id,
          match_type: 99999
        }
      });
      if (response.status === 200) {
        this.$store.state.overlay.details.prevscene = this.$store.state.overlay.quickFind.selectedScene;
        this.$buefy.toast.open({ message: `The scene was sucessfully relinked to a new Scene`, type: 'is-primary', duration: 3000 });
      }
    },
    async scrapeScene() {
      this.$buefy.dialog.confirm({
        title: 'Scrape & Create Scene',
        message: `Do you wish to create a seperate XBVR scene from this linked scene <strong>${this.$store.state.overlay.details.altsrc.url}</strong>`,
        type: 'is-info is-wide',
        hasIcon: true,
        onConfirm: () => {
          const url = this.$store.state.overlay.details.altsrc.url
          this.$store.state.overlay.details.altsrc = null
          this.$store.commit('overlay/hideDetails')
          // call the options screen passing the url in state   
          this.$store.commit('optionsSceneCreate/setScrapeScene', url )
          this.$store.commit('optionsSceneCreate/showSceneCreate', true )
          this.$router.push({ path: '/options'})
        }
      })

    },
    async refreshExtRef() {
      this.$buefy.dialog.confirm({
        title: 'Continue?',
        message: `This will remove the scene, rescrape the site to relink it to an XBVR scene`,
        type: 'is-info is-wide',
        hasIcon: true,
        onConfirm: () => {          
          this.handleRefreshExtRef()
        }
      })
    },
    async handleRefreshExtRef() {
      const response = await api.delete(`/extref/delete_extref`, {
        json: {
          external_source: this.$store.state.overlay.details.altsrc.external_source,
          external_id: this.$store.state.overlay.details.altsrc.external_id,
        }
      });
      if (response.status === 200) {
        this.$store.state.overlay.details.prevscene = this.$store.state.overlay.quickFind.selectedScene;
        this.$buefy.toast.open({ message: `The scene was removed, ready to rescan`, type: 'is-primary', duration: 3000 });
      }
    },
    flagExtRefDeleted() {
      let confirmed = false
      this.$buefy.dialog.confirm({
        title: 'Continue?',
        message: `This will unlink the scene and prevent it from relinking to any scene. This cannot be undone`,
        type: 'is-danger is-wide',
        hasIcon: true,
        onConfirm: () => {          
          this.handleFlagExtRefDeleted()
        },
      })    
    },    
    async handleFlagExtRefDeleted() {
      const response = await api.post(`/extref/edit_link`, {
        json: {
          external_source: this.$store.state.overlay.details.altsrc.external_source,
          external_id: this.$store.state.overlay.details.altsrc.external_id,
          internal_table: "scenes",
          internal_db_id: 0,
          internal_name_id: "deleted",
          match_type: -1
        }
      });
      if (response.status === 200) {
        this.$store.state.overlay.details.prevscene = this.$store.state.overlay.quickFind.selectedScene;
        this.$buefy.toast.open({ message: `The scene was unlinked and will not be relinked to any scene`, type: 'is-primary', duration: 3000 });
      }
    },    
    format,
    parseISO,
    prettyBytes,
    formatDistance
  }
}
</script>

<style lang="less" scoped>
/* ------------------------------------------------------------------
   Scene details overlay — two-pane media/info layout
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
   Media pane — dark chrome in both themes
   ------------------------------------------------------------------ */

.media-pane {
  background: var(--xbvr-media-bg, #0d1017);
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
  background: var(--xbvr-media-tabbar, rgba(255, 255, 255, 0.07));
  border-radius: 999px;
}

.media-tabs :deep(.tabs a) {
  border-bottom: none;
  border-radius: 999px;
  padding: 0.3em 1.2em;
  color: var(--xbvr-media-muted, rgba(230, 233, 242, 0.6));
  font-weight: 600;
  transition: color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
    background-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

.media-tabs :deep(.tabs a:hover) {
  color: var(--xbvr-text, #fff);
  border-bottom: none;
}

.media-tabs :deep(.tabs li.is-active a) {
  background: var(--xbvr-media-tab-active-bg, rgba(255, 255, 255, 0.14));
  color: var(--xbvr-media-tab-active-text, #fff);
}

.media-pane :deep(.carousel-item .image) {
  border-radius: var(--xbvr-radius-sm, 8px);
}

.media-pane .video-js {
  margin: 0 auto;
  border-radius: var(--xbvr-radius-sm, 8px);
  overflow: hidden;
}

/* seek-step buttons under the player */
.skip-row :deep(.tag) {
  border-radius: 999px;
  background: var(--xbvr-media-chip-bg, rgba(255, 255, 255, 0.08));
  border-color: var(--xbvr-media-chip-border, rgba(255, 255, 255, 0.18));
  color: var(--xbvr-media-chip-text, #cdd6ea);
  font-weight: 600;
  box-shadow: none;
  transition: background-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
    color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

.skip-row :deep(.tag:hover) {
  background: var(--xbvr-media-tabbar, rgba(255, 255, 255, 0.16));
  color: var(--xbvr-text, #fff);
  border-color: var(--xbvr-border-strong, rgba(255, 255, 255, 0.3));
}

/* ------------------------------------------------------------------
   Info pane
   ------------------------------------------------------------------ */

.details-header {
  margin-bottom: 0.9rem;
}

.details-title {
  font-size: 1.35rem;
  font-weight: 800;
  letter-spacing: -0.01em;
  line-height: 1.25;
  color: var(--xbvr-text, #1c2333);
  text-wrap: balance;
  overflow-wrap: break-word;
  margin-bottom: 0.55rem;
}

.missing {
  opacity: 0.6;
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
  transition: color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
    border-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
    background-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

.meta-chip.is-link:hover {
  color: var(--xbvr-primary-strong, #4338ca);
  border-color: var(--xbvr-primary, #4f46e5);
  background: var(--xbvr-primary-soft, #eef0fe);
}

/* rating + action toolbar */
.details-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem;
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

.rating-reset {
  padding-left: 0.6em;
  color: var(--xbvr-text-faint, #98a1b6);
  cursor: pointer;
  transition: color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

.rating-reset:hover {
  color: var(--xbvr-text, #1c2333);
}

.action-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.3rem;
  margin-left: auto;
}

.action-row :deep(.button.is-small) {
  border-radius: 8px;
}

/* alternate source icons */
.altsrc-row {
  justify-content: flex-end;
  margin-bottom: 0.5rem;
}

/* cast portraits */
.cast-strip {
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 0.9rem;
}

.image-row {
  display: flex;
}

.image-wrapper {
  position: relative;
}

.thumbnail {
  height: 96px;
  border-radius: 10px;
  object-fit: cover;
  cursor: pointer;
  box-shadow: var(--xbvr-shadow-sm, 0 1px 2px rgba(16, 24, 40, 0.05));
  transition: transform var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
    box-shadow var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

.thumbnail:hover {
  transform: translateY(-2px);
  box-shadow: var(--xbvr-shadow-md, 0 4px 12px rgba(16, 24, 40, 0.10));
}

/* tag chips */
.chips :deep(.tag) {
  border-radius: 999px;
}

.block-tags {
  max-height: 200px;
  overflow: auto;
  scrollbar-width: none;
}

.block-tags::-webkit-scrollbar {
  display: none;
}

.block-tab-content {
  flex: 1 1 auto;
}

/* files tab rows */
.block-tab-content :deep(.media) {
  padding: 0.6rem 0.7rem;
  border: 1px solid var(--xbvr-border, #e3e6ec);
  border-radius: var(--xbvr-radius-sm, 8px);
  background: var(--xbvr-surface, #ffffff);
  transition: border-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
    background-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

.block-tab-content :deep(.media:hover) {
  border-color: var(--xbvr-border-strong, #cdd2dc);
  background: var(--xbvr-hover-bg, #fafbfd);
}

.block-tab-content :deep(.media + .media) {
  border-top: 1px solid var(--xbvr-border, #e3e6ec);
  margin-top: 0.5rem;
  padding-top: 0.6rem;
}

.vue-star-rating {
  line-height: 0;
}

.scene-id {
  position: absolute;
  right: 12px;
  bottom: 6px;
  font-size: 10px;
  letter-spacing: 0.04em;
  color: var(--xbvr-text-faint, #98a1b6);
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

.scene-nav {
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

.scene-nav:hover {
  background: rgba(20, 24, 36, 0.8);
}

.scene-nav.prev {
  left: 14px;
}

.scene-nav.prev:hover {
  transform: translateY(-50%) translateX(-2px);
}

.scene-nav.next {
  right: 14px;
}

.scene-nav.next:hover {
  transform: translateY(-50%) translateX(2px);
}

.is-1by1 {
  padding-top: calc(100% - 40px - 1em) !important;
}

:deep(.video-js .vjs-big-play-button) {
  left: 50% !important;
  top: 50% !important;
  transform: translate(-50%, -50%) !important;
}

.pathDetails {
  color: var(--xbvr-text-faint, #98a1b6);
  overflow-wrap: anywhere;
}

.heatmapFunscript {
  width: 100%;
  padding: 0;
  margin-top: 0.5em;
}

.heatmapFunscript img {
  border: 1px solid var(--xbvr-border-strong, #cdd2dc);
  border-radius: 999px;
  width: 100%;
  height: 20px;
  margin: 0;
  padding: 0;
}

.videosize {
  color: var(--xbvr-text-muted, #64708a);
  font-weight: 600;
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

.is-divider {
  margin: 0.8rem 0;
}

.tooltip {
  position: absolute;
  z-index: 1;
  top: 50px;
  right: 100%;
  width: 400px;
  height: 400px;
  background-color: var(--xbvr-surface, #ffffff);
  border: 1px solid var(--xbvr-border, #e3e6ec);
  border-radius: var(--xbvr-radius, 12px);
  box-shadow: var(--xbvr-shadow-lg, 0 16px 40px rgba(16, 24, 40, 0.16));
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 10px;
  transform: translateX(10px);
}

.tooltip img {
  max-width: 100%;
  max-height: 100%;
  border-radius: 8px;
}

.altsrc-image-wrapper {
  display: inline-block;
  margin-left: 5px;
  cursor: pointer;
  border-radius: 6px;
  transition: transform var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

.altsrc-image-wrapper:hover {
  transform: translateY(-2px);
}

.altsrc-image-wrapper img {
  border-radius: 6px;
}
</style>
