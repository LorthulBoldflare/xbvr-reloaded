<template>
  <div class="filters-panel">
    <section class="filter-section">
      <h3 class="filter-heading">
        <b-icon pack="mdi" icon="bookmark-multiple-outline" size="is-small" aria-hidden="true"/>
        <span>{{ $t('Saved searches') }}</span>
      </h3>
      <SavedSearch mode="scenes"/>
    </section>

    <section class="filter-section">
      <h3 class="filter-heading">
        <b-icon pack="mdi" icon="playlist-check" size="is-small" aria-hidden="true"/>
        <span>{{ $t('Properties') }}</span>
      </h3>
      <div class="prop-grid">
        <b-checkbox-button v-model="lists" native-value="watchlist" type="is-primary">
          <b-icon pack="mdi" icon="calendar-check"/>
          <span>{{ $t('Watchlist') }}</span>
        </b-checkbox-button>
        <b-checkbox-button v-model="lists" native-value="favourite" type="is-danger">
          <b-icon pack="mdi" icon="heart"/>
          <span>{{ $t('Favourite') }}</span>
        </b-checkbox-button>
        <b-checkbox-button v-model="lists" native-value="wishlist" type="is-info">
          <b-icon pack="mdi" icon="oil-lamp"/>
          <span>{{ $t('Wishlist') }}</span>
        </b-checkbox-button>
        <b-checkbox-button v-model="lists" native-value="scripted" type="is-info">
          <b-icon pack="mdi" icon="pulse"/>
          <span>{{ $t('Scripted') }}</span>
        </b-checkbox-button>
      </div>
    </section>

    <section class="filter-section">
      <h3 class="filter-heading">
        <b-icon pack="mdi" icon="sort" size="is-small" aria-hidden="true"/>
        <span>{{ $t('Sorting & status') }}</span>
      </h3>

      <div class="filter-field">
        <label class="filter-label" for="filter-sort">{{ $t('Sort by') }}</label>
        <div class="select is-fullwidth">
          <select id="filter-sort" v-model="sort">
            <option value="release_desc">↓ {{ $t("Release date") }}</option>
            <option value="release_asc">↑ {{ $t("Release date") }}</option>
            <option value="added_desc">↓ {{ $t("File added date") }}</option>
            <option value="added_asc">↑ {{ $t("File added date") }}</option>
            <option value="title_desc">↓ {{ $t("Title") }}</option>
            <option value="title_asc">↑ {{ $t("Title") }}</option>
            <option value="total_file_size_desc">↓ {{ $t("File size") }}</option>
            <option value="total_file_size_asc">↑ {{ $t("File size") }}</option>
            <option value="rating_desc">↓ {{ $t("Rating") }}</option>
            <option value="rating_asc">↑ {{ $t("Rating") }}</option>
            <option value="total_watch_time_desc">↓ {{ $t("Watch time") }}</option>
            <option value="total_watch_time_asc">↑ {{ $t("Watch time") }}</option>
            <option value="duration_desc">↓ {{ $t("Duration") }}</option>
            <option value="duration_asc">↑ {{ $t("Duration") }}</option>
            <option value="scene_added_desc">↓ {{ $t("Scene added date") }}</option>
            <option value="scene_updated_desc">↓ {{ $t("Scene updated date") }}</option>
            <option value="last_opened_desc">↓ {{ $t("Last viewed date") }}</option>
            <option value="last_opened_asc">↑ {{ $t("Last viewed date") }}</option>
            <option value="script_published_desc">↓ {{ $t("Published Script Added") }}</option>
            <option value="scene_id_desc">↓ {{ $t("Scene Id") }}</option>
            <option value="site_asc">↑ {{ $t("Site") }}</option>
            <option value="alt_src_desc">↓ {{ $t("Linked to Alternate Sites") }}</option>
            <option value="random">↯ {{ $t("Random") }}</option>
          </select>
        </div>
      </div>

      <div class="filter-field">
        <label class="filter-label" for="filter-watched">{{ $t('Watched') }}</label>
        <div class="select is-fullwidth">
          <select id="filter-watched" v-model="isWatched">
            <option :value="null">{{ $t('Everything') }}</option>
            <option :value="true">{{ $t('Watched') }}</option>
            <option :value="false">{{ $t('Unwatched') }}</option>
          </select>
        </div>
      </div>

      <div class="filter-field">
        <label class="filter-label" for="filter-release-month">{{ $t('Release month') }}</label>
        <div class="field has-addons">
          <div class="control is-expanded">
            <div class="select is-fullwidth">
              <select id="filter-release-month" v-model="releaseMonth">
                <option></option>
                <option v-for="t in filters.release_month" :key="t">{{ t }}</option>
              </select>
            </div>
          </div>
          <div class="control">
            <button type="button" class="button is-light" @click="clearReleaseMonth" :aria-label="$t('Clear release month')">
              <b-icon pack="fas" icon="times" size="is-small"></b-icon>
            </button>
          </div>
        </div>
      </div>

      <div class="filter-field">
        <label class="filter-label" for="filter-folder">{{ $t('Folder') }}</label>
        <div class="field has-addons">
          <div class="control is-expanded">
            <div class="select is-fullwidth">
              <select id="filter-folder" v-model="volume">
                <option :value="0"></option>
                <option v-for="t in filters.volumes" :key="t.id" :value="t.id">{{ t.path }}</option>
              </select>
            </div>
          </div>
          <div class="control">
            <button type="button" class="button is-light" @click="clearVolume" :aria-label="$t('Clear folder')">
              <b-icon pack="fas" icon="times" size="is-small"></b-icon>
            </button>
          </div>
        </div>
      </div>
    </section>

    <section class="filter-section" v-if="Object.keys(filters).length !== 0">
      <h3 class="filter-heading">
        <b-icon pack="mdi" icon="filter-outline" size="is-small" aria-hidden="true"/>
        <span>{{ $t('Filters') }}</span>
      </h3>
      <p class="filter-hint">{{ $t('Click a chip to cycle: include → must have → exclude') }}</p>

      <div class="filter-field">
        <label class="filter-label">{{ $t('Cast') }}</label>
        <b-taginput v-model="cast" autocomplete :data="filteredCast" @typing="getFilteredCast">
          <template slot-scope="props">{{ props.option }}</template>
          <template slot="empty">No matching cast</template>
          <template #selected="props">
              <b-tag v-for="(tag, index) in props.tags"
                :type="tag.charAt(0)=='!' ? 'is-danger': (tag.charAt(0)=='&' ? 'is-success' : '')"
                :key="tag+index" :tabstop="false" closable  @close="cast=cast.filter(e => e !== tag)" @click="toggle3way(tag,index,'cast')">              
                <b-tooltip position="is-right" :delay="200"
                  :label="tag.charAt(0)=='!' ? 'Exclude ' + removeConditionPrefix(tag) : tag.charAt(0)=='&' ? 'Must Have ' + removeConditionPrefix(tag) : 'Include ' + removeConditionPrefix(tag)">
                  <b-icon pack="mdi" v-if="tag.charAt(0)=='!'" icon="minus-circle-outline" size="is-small" class="tagicon"></b-icon>
                  <b-icon pack="mdi" v-if="tag.charAt(0)=='&'" icon="plus-circle-outline" size="is-small" class="tagicon"></b-icon>
                  {{removeConditionPrefix(tag)}}
                </b-tooltip>
              </b-tag>
          </template>
        </b-taginput>
      </div>

      <div class="filter-field">
        <label class="filter-label">{{ $t('Site') }}</label>
        <b-taginput v-model="sites" autocomplete :data="filteredSites" @typing="getFilteredSites">
          <template slot-scope="props">{{ props.option }}</template>
          <template slot="empty">No matching sites</template>
          <template #selected="props">
            <b-tag v-for="(tag, index) in props.tags"
              :type="tag.charAt(0)=='!' ? 'is-danger': (tag.charAt(0)=='&' ? 'is-success' : '')"
              :key="tag+index" :tabstop="false" closable  @close="sites=sites.filter(e => e !== tag)" @click="toggle2Way(tag,index,'sites')">
                <b-tooltip position="is-right" :delay="200"
                  :label="tag.charAt(0)=='!' ? 'Exclude ' + removeConditionPrefix(tag) : 'Include ' + removeConditionPrefix(tag)">
                  <b-icon pack="mdi" v-if="tag.charAt(0)=='!'" icon="minus-circle-outline" size="is-small" class="tagicon"></b-icon>
                  {{removeConditionPrefix(tag)}}
                </b-tooltip>
            </b-tag>
          </template>
        </b-taginput>
      </div>

      <div class="filter-field">
        <label class="filter-label">{{ $t('Tags') }}</label>
        <b-taginput v-model="tags" autocomplete :data="filteredTags" @typing="getFilteredTags">
          <template slot-scope="props">{{ props.option }}</template>
          <template slot="empty">No matching tags</template>
          <template #selected="props">
            <b-tag v-for="(tag, index) in props.tags"
              :type="tag.charAt(0)=='!' ? 'is-danger': (tag.charAt(0)=='&' ? 'is-success' : '')"
              :key="tag+index" :tabstop="false" closable  @close="tags=tags.filter(e => e !== tag)" @click="toggle3way(tag,index,'tags')"> 
              <b-tooltip position="is-right" :delay="200"
                  :label="tag.charAt(0)=='!' ? 'Exclude ' + removeConditionPrefix(tag) : tag.charAt(0)=='&' ? 'Must Have ' + removeConditionPrefix(tag) : 'Include ' + removeConditionPrefix(tag)">
                <b-icon pack="mdi" v-if="tag.charAt(0)=='!'" icon="minus-circle-outline" size="is-small" class="tagicon"></b-icon>
                <b-icon pack="mdi" v-if="tag.charAt(0)=='&'" icon="plus-circle-outline" size="is-small" class="tagicon"></b-icon>
                {{removeConditionPrefix(tag)}}
              </b-tooltip>
            </b-tag>
          </template>
        </b-taginput>
      </div>

      <div class="filter-field">
        <label class="filter-label">{{ $t('Cuepoint') }}</label>
        <b-taginput v-model="cuepoint" autocomplete :data="filteredCuepoints" @typing="getFilteredCuepoints">
          <template slot-scope="props">{{ props.option }}</template>
          <template slot="empty">No matching cuepoints</template>
          <template #selected="props">
            <b-tag v-for="(tag, index) in props.tags"
              :type="tag.charAt(0)=='!' ? 'is-danger': (tag.charAt(0)=='&' ? 'is-success' : '')"
              :key="tag+index" :tabstop="false" closable  @close="cuepoint=cuepoint.filter(e => e !== tag)" @click="toggle3way(tag,index,'cuepoints')"> 
              <b-tooltip position="is-right" :delay="200"
                  :label="tag.charAt(0)=='!' ? 'Exclude ' + removeConditionPrefix(tag) : tag.charAt(0)=='&' ? 'Must Have ' + removeConditionPrefix(tag) : 'Include ' + removeConditionPrefix(tag)">
                <b-icon pack="mdi" v-if="tag.charAt(0)=='!'" icon="minus-circle-outline" size="is-small" class="tagicon"></b-icon>
                <b-icon pack="mdi" v-if="tag.charAt(0)=='&'" icon="plus-circle-outline" size="is-small" class="tagicon"></b-icon>
                {{removeConditionPrefix(tag)}}
              </b-tooltip>
            </b-tag>
          </template>
        </b-taginput>
      </div>

      <div class="filter-field">
        <b-tooltip position="is-top" label="Allows searching a variety of attributes such as: scenes in Watchlists, Favourites, Has Video, Scripts or HSP Files, Subscriptions, Ratings, Cuepoint Types, Number of Cast, FOV, Projection, Resolution, Frame Rate and Codecs" multilined :delay="1000">
          <label class="filter-label">
            <span>{{ $t('Attributes') }}</span>
            <b-icon pack="mdi" icon="help-circle-outline" size="is-small" aria-hidden="true"/>
          </label>
        </b-tooltip>
        <b-taginput v-model="attributes" autocomplete :data="filteredAttributes" @typing="getFilteredAttributes">
            <template slot-scope="props">{{ props.option }}</template>
            <template slot="empty">No matching attributes</template>
            <template #selected="props">
              <b-tag v-for="(tag, index) in props.tags"
                :type="tag.charAt(0)=='!' ? 'is-danger': (tag.charAt(0)=='&' ? 'is-success' : '')"
                :key="tag+index" :tabstop="false" closable  @close="attributes=attributes.filter(e => e !== tag)" @click="toggle3way(tag,index,'attributes')"> 
                  <b-icon pack="mdi" v-if="tag.charAt(0)=='!'" icon="minus-circle-outline" size="is-small" class="tagicon"></b-icon>
                  <b-icon pack="mdi" v-if="tag.charAt(0)=='&'" icon="plus-circle-outline" size="is-small" class="tagicon"></b-icon>
                  {{removeConditionPrefix(tag)}}
              </b-tag>
            </template>          
          </b-taginput>
      </div>
    </section>

    <section class="filter-section">
      <button type="button" class="filter-heading is-collapsible" @click="showAka = !showAka" :aria-expanded="showAka">
        <b-icon pack="mdi" icon="account-multiple-outline" size="is-small" aria-hidden="true"/>
        <span>{{ $t('Actor groups (AKA)') }}</span>
        <b-icon pack="mdi" icon="chevron-down" size="is-small" class="heading-chevron" :class="{ open: showAka }" aria-hidden="true"/>
      </button>
      <div v-show="showAka" class="btn-grid">
        <b-tooltip position="is-right" label="New Aka Group. Select 2 or more actors in the Cast filter" multilined :delay="200">
          <button class="button is-small is-outlined" @click="createAkaGroup" :disabled="disableNewAkaGroup">
            <b-icon pack="mdi" icon="account-multiple-plus-outline"></b-icon>
            <span>{{ $t('New') }}</span>
          </button>
        </b-tooltip>
        <b-tooltip position="is-right" label="Select the Aka Group to delete in the Cast Filter" multilined :delay="200">
          <button class="button is-small is-outlined" @click="deleteAkaGroup" :disabled="disableDeleteAkaGroup">
            <b-icon pack="mdi" icon="delete-outline"></b-icon>
            <span>{{ $t('Delete') }}</span>
          </button>
        </b-tooltip>
        <b-tooltip position="is-bottom" label="Add Cast to Aka Group. Select the Aka group and Actors to add in the Cast Filter" multilined :delay="200">
          <button class="button is-small is-outlined" @click="addToAkaGroup" :disabled="disableAddToAkaGroup">
            <b-icon pack="mdi" icon="account-plus-outline"></b-icon>
            <span>{{ $t('Add cast') }}</span>
          </button>
        </b-tooltip>
        <b-tooltip position="is-bottom" label="Remove Cast from Aka Group. Select the Aka group and Actors to remove in the Cast Filter" multilined :delay="200">
          <button class="button is-small is-outlined" @click="removeFromAkaGroup" :disabled="disableRemoveFromAkaGroup">
            <b-icon pack="mdi" icon="account-minus-outline"></b-icon>
            <span>{{ $t('Remove cast') }}</span>
          </button>
        </b-tooltip>
      </div>
    </section>

    <section class="filter-section">
      <button type="button" class="filter-heading is-collapsible" @click="showTagGroups = !showTagGroups" :aria-expanded="showTagGroups">
        <b-icon pack="mdi" icon="tag-multiple-outline" size="is-small" aria-hidden="true"/>
        <span>{{ $t('Tag groups') }}</span>
        <b-icon pack="mdi" icon="chevron-down" size="is-small" class="heading-chevron" :class="{ open: showTagGroups }" aria-hidden="true"/>
      </button>
      <div v-show="showTagGroups" class="btn-grid">
        <b-tooltip position="is-right" label="New Tag Group. Select 2 or more tags in the Tag filter" multilined :delay="200">
          <button class="button is-small is-outlined" @click="showGroupTagNameDialog('create')" :disabled="disableNewTagGroup">
            <b-icon pack="mdi" icon="tag-multiple-outline"></b-icon>
            <span>{{ $t('New') }}</span>
          </button>
        </b-tooltip>
        <b-tooltip position="is-right" label="Select the Tag Group to delete in the Tag Filter" multilined :delay="200">
          <button class="button is-small is-outlined" @click="deleteTagGroup" :disabled="disableDeleteRenameTagGroup">
            <b-icon pack="mdi" icon="delete-outline"></b-icon>
            <span>{{ $t('Delete') }}</span>
          </button>
        </b-tooltip>
        <b-tooltip position="is-bottom" label="Add Tag to Tag Group. Select the Tag  group and Tag to add in the Tag Filter" multilined :delay="200">
          <button class="button is-small is-outlined" @click="addToTagGroup" :disabled="disableAddToTagGroup">
            <b-icon pack="mdi" icon="tag-plus-outline"></b-icon>
            <span>{{ $t('Add tag') }}</span>
          </button>
        </b-tooltip>
        <b-tooltip position="is-bottom" label="Remove Tag from Tag Group. Select the Tag group and Tags to remove in the Tag Filter" multilined :delay="200">
          <button class="button is-small is-outlined" @click="removeFromTagGroup" :disabled="disableRemoveFromTagGroup">
            <b-icon pack="mdi" icon="tag-minus-outline"></b-icon>
            <span>{{ $t('Remove tag') }}</span>
          </button>
        </b-tooltip>
        <b-tooltip position="is-bottom" label="Rename Tag Group. Select the Tag group in the Tag Filter" multilined :delay="200">
          <button class="button is-small is-outlined" @click="showGroupTagNameDialog('rename')" :disabled="disableDeleteRenameTagGroup">
            <b-icon pack="mdi" icon="rename-outline"></b-icon>
            <span>{{ $t('Rename') }}</span>
          </button>
        </b-tooltip>
        <b-tooltip position="is-bottom" label="List Tags in Group" multilined :delay="200">
          <button class="button is-small is-outlined" @click="getTagGroup" :disabled="disableGetTagGroup">
            <b-icon pack="mdi" icon="tag-search-outline"></b-icon>
            <span>{{ $t('List tags') }}</span>
          </button>
        </b-tooltip>
      </div>
    </section>

    <b-modal :active.sync="isGroupTagNameModalActive"
             has-modal-card
             trap-focus
             aria-role="dialog"
             aria-modal>
      <div class="modal-card" style="width: auto">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ $t('Tag group') }}</p>
        </header>
        <section class="modal-card-body">
          <b-field :label="$t('Name')">
            <b-input
              type="name"
              v-model="tagGroupName"
              required>
            </b-input>
          </b-field>
        </section>
        <footer class="modal-card-foot">
          <button class="button is-primary" :disabled="tagGroupName===''" @click="tagGroupModalClicked()">{{ $t('Save') }}
          </button>
        </footer>
      </div>
    </b-modal>
  </div>
</template>

<script>
import SavedSearch from '../../components/SavedSearch'
import api from '../../api'
import { confirmAndDeleteAkaGroup } from '../../util/akaGroups'

export default {
  name: 'Filters',
  components: { SavedSearch },
  mounted () {
    this.$store.dispatch('sceneList/filters')
    this.fetchFilters()
  },
  data () {
    return {
      filteredCast: [],
      filteredSites: [],
      filteredTags: [],
      filteredCuepoints: [],
      filteredAttributes: [],
      isGroupTagNameModalActive: false,
      tagGroupName: '',
      groupNameDialogAction: 'create',
      showAka: false,
      showTagGroups: false,
    }
  },
  methods: {
    reloadList () {
      this.$router.push({
        name: 'scenes',
        query: {
          q: this.$store.getters['sceneList/filterQueryParams']
        }
      })
    },
    getFilteredCast (text) {
      this.filteredCast = this.filters.cast.filter(option => (
        option.toString().toLowerCase().indexOf(text.toLowerCase()) >= 0 &&
        !this.cast.some(entry => this.removeConditionPrefix(entry.toString()) === option.toString())
      ))      
    },
    getFilteredSites (text) {
      this.filteredSites = this.filters.sites.filter(option => (
        option.toString().toLowerCase().indexOf(text.toLowerCase()) >= 0 &&
        !this.sites.some(entry => this.removeConditionPrefix(entry.toString()) === option.toString())
      ))
    },
    getFilteredTags (text) {
      this.filteredTags = this.filters.tags.filter(option => (
        option.toString().toLowerCase().indexOf(text.toLowerCase()) >= 0 &&
        !this.tags.some(entry => this.removeConditionPrefix(entry.toString()) === option.toString())
      ))
    },
    getFilteredCuepoints (text) {
      this.filteredCuepoints = this.filters.cuepoints.filter(option => (
        option.toString().toLowerCase().indexOf(text.toLowerCase()) >= 0 &&
        !this.cuepoint.some(entry => this.removeConditionPrefix(entry.toString()) === option.toString())
      ))
    },
    getFilteredAttributes (text) {
      this.filteredAttributes = this.filters.attributes.filter(option => (
        option.toString().toLowerCase().indexOf(text.toLowerCase()) >= 0 &&
        !this.tags.some(entry => this.removeConditionPrefix(entry.toString()) === option.toString())
      ))
    },
    clearReleaseMonth () {
      this.$store.state.sceneList.filters.releaseMonth = ''
      this.reloadList()
    },
    clearVolume () {
      this.$store.state.sceneList.filters.volume = 0
      this.$store.dispatch('sceneList/filters')
      this.reloadList()
    },
    createAkaGroup () {
      this.$store.state.sceneList.isLoading = true
      api.post('/aka/create', {json: {actorList: this.cast}}).json().then(data => {
        this.cast.push(data.akas.aka_actor.name)
        this.$store.dispatch('sceneList/filters')
        this.reloadList()
        if (data.status != '') {
          this.$buefy.toast.open({message: `Warning:  ${data.status}`, type: 'is-warning', duration: 5000})
        }
        this.$store.state.sceneList.isLoading = false
      })
    },
    deleteAkaGroup () {
      confirmAndDeleteAkaGroup(this, this.cast[0], 'sceneList', () => {
        this.cast = []
        this.reloadList()
      })
    },
    addToAkaGroup () {
      this.$store.state.sceneList.isLoading = true
      api.post('/aka/add', {json: {actorList: this.cast}}).json().then(data => {        
        // delete old aka & add new name
        this.cast = this.cast.filter(e => !e.startsWith("aka:")) 
        this.cast.push(data.akas.aka_actor.name) 
        this.$store.dispatch('sceneList/filters')       
        this.reloadList()        
        if (data.status != '') {
          this.$buefy.toast.open({message: `Warning:  ${data.status}`, type: 'is-warning', duration: 5000})
        }
        this.$store.state.sceneList.isLoading = false
      })
      
    },
    removeFromAkaGroup () {
      this.$store.state.sceneList.isLoading = true
      api.post('/aka/remove', {json: {actorList: this.cast}}).json().then(data => {        
        // delete old aka & add new name
        this.cast = this.cast.filter(e => !e.startsWith("aka:")) 
        this.cast.push(data.akas.aka_actor.name)
        this.$store.dispatch('sceneList/filters')
        this.reloadList()
        if (data.status != '') {
          this.$buefy.toast.open({message: `Warning:  ${data.status}`, type: 'is-warning', duration: 5000})
        }
        this.$store.state.sceneList.isLoading = false
      })
    },
    showGroupTagNameDialog (action) {
      this.groupTagName = ''
      this.isGroupTagNameModalActive = true
      this.groupNameDialogAction = action
    },
    tagGroupModalClicked () {
      if (this.groupNameDialogAction == 'create') {
        this.createTagGroup()
      } else {
        this.renameTagGroup()
      }
    },
    createTagGroup () {
      this.isGroupTagNameModalActive = false
      this.$store.state.sceneList.isLoading = true
      api.post('/tag_group/create', {json: {name: this.tagGroupName, tagList: this.tags}}).json().then(data => {
        if (data.tag_group.tag_group_tag.name != "") {
          this.tags.push(data.tag_group.tag_group_tag.name)
        }
        this.$store.dispatch('sceneList/filters')
        this.reloadList()
        if (data.status != '') {
          this.$buefy.toast.open({message: `Warning:  ${data.status}`, type: 'is-warning', duration: 5000})
        }
        this.$store.state.sceneList.isLoading = false
      })
    },
    deleteTagGroup () {
      this.$buefy.dialog.confirm({
        title: this.$t('Delete tag group'),
        message: `Do you want to delete the tag group for <strong>${this.tags[0]}</strong>?`,
        type: 'is-danger',
        hasIcon: true,
        confirmText: this.$t('Delete'),
        onConfirm: () => {
          this.$store.state.sceneList.isLoading = true
          api.post('/tag_group/delete', {json: {name: this.tags[0]}}).json().then(data => {
        this.tags = []
        this.$store.dispatch('sceneList/filters')
        this.reloadList()
        this.$store.state.sceneList.isLoading = false
      })
        }
      })
    },
    addToTagGroup () {      
      this.$store.state.sceneList.isLoading = true
      api.post('/tag_group/add', {json: {tagList: this.tags}}).json().then(data => {
        this.$store.dispatch('sceneList/filters')       
        this.reloadList()        
        if (data.status != '') {
          this.$buefy.toast.open({message: `Warning:  ${data.status}`, type: 'is-warning', duration: 5000})
        }
        this.$store.state.sceneList.isLoading = false
      })
      
    },
    removeFromTagGroup () {
      this.$store.state.sceneList.isLoading = true
      api.post('/tag_group/remove', {timeout: 60000, json: {tagList: this.tags}}).json().then(data => {        
        this.$store.dispatch('sceneList/filters')
        this.reloadList()
        if (data.status != '') {
          this.$buefy.toast.open({message: `Warning:  ${data.status}`, type: 'is-warning', duration: 5000})
        }
        this.$store.state.sceneList.isLoading = false
      })
    },
    renameTagGroup () {
      this.isGroupTagNameModalActive = false
      this.$store.state.sceneList.isLoading = true
      api.post('/tag_group/rename', {json: {name: this.tagGroupName, tagList: this.tags}}).json().then(data => {
        if (data.status != '') {
          this.$buefy.toast.open({message: `${data.status}`, type: 'is-danger', duration: 5000})
        } else {
          this.reloadList()
          this.tags = []
          this.tags.push(data.tag_group.tag_group_tag.name)
          this.$store.dispatch('sceneList/filters')
        }
        this.$store.state.sceneList.isLoading = false
        })
    },
    getTagGroup () {
      this.$store.state.sceneList.isLoading = true
      let name = ""
      for (var i = 0; i < this.tags.length; i++) {        
        if (this.tags[i].startsWith("tag group:")) {
           name = this.tags[i]
        }
      }

      api.get('/tag_group/' + name, {timeout: 60000}).json().then(data => {        
        if (data.status != '') {
          this.$buefy.toast.open({message: `${data.status}`, type: 'is-danger', duration: 5000})
        } else {
          let newTagList = []
          newTagList.push("tag group:" + data.tag_group.name)
          for (var i = 0; i < data.tag_group.tags.length; i++) {
            newTagList.push(data.tag_group.tags[i].name)
          }
          this.tags = newTagList
          this.$store.dispatch('sceneList/filters')
        }
      })
      this.$store.state.sceneList.isLoading = false
    },
    toggle3way (text, idx, list) {      
      let tags = []
      switch (list) {
        case 'cast':
          tags=this.cast 
          break
        case 'tags':
          tags=this.tags
          break
        case 'cuepoints':
          tags=this.cuepoint
          break
        case 'attributes':
          tags=this.attributes
          break
      }      
      switch(tags[idx].charAt(0)) {
        case '!':
          tags[idx]=this.removeConditionPrefix(tags[idx])
          break
        case '&':
          tags[idx]='!' + this.removeConditionPrefix(tags[idx])
        break
        default:
        tags[idx]='&'+text        
      }      
      switch (list) {
        case 'cast':
          this.cast=tags
          break
        case 'tags':
          this.tags=tags
          break
        case 'cuepoints':
          this.cuepoint=tags
          break
        case 'attributes':
          this.attributes=tags
          break
      }
    },    
    toggle2Way (text, idx, list) {      
      let tags = []
      switch (list) {
        case 'sites':
          tags=this.sites
      }      
      switch(tags[idx].charAt(0)) {
        case '!':
          tags[idx]=this.removeConditionPrefix(tags[idx])
          break
        default:
        tags[idx]='!'+text        
      }      
      switch (list) {
        case 'sites':
          this.sites=tags
          break
      }      
    },    
    removeConditionPrefix(txt) {
      if (txt.charAt(0)=='!' || txt.charAt(0)=='&') {
        return txt.substring(1) 
      }
      return txt
    },
    async fetchFilters() {
        this.filteredAttributes=['Loading attributes']
        // heavy endpoint on large libraries; keep the historical 5-min budget
        api.get('/scene/filters', { timeout: 300000 }).json().then(data => {
          this.filteredAttributes=data.attributes
      })
    }
  },
  computed: {
    filters () {
      return this.$store.state.sceneList.filterOpts
    },
    lists: {
      get () {
        return this.$store.state.sceneList.filters.lists
      },
      set (value) {
        this.$store.state.sceneList.filters.lists = value
        this.reloadList()
      }
    },
    releaseMonth: {
      get () {
        return this.$store.state.sceneList.filters.releaseMonth
      },
      set (value) {
        this.$store.state.sceneList.filters.releaseMonth = value
        this.reloadList()
      }
    },
    volume: {
      get () {
        return this.$store.state.sceneList.filters.volume
      },
      set (value) {
        this.$store.state.sceneList.filters.volume = value
        this.reloadList()
      }
    },
    cast: {
      get () {
        return this.$store.state.sceneList.filters.cast
      },
      set (value) {
        this.$store.state.sceneList.filters.cast = value
        this.reloadList()
      }
    },
    sites: {
      get () {
        return this.$store.state.sceneList.filters.sites
      },
      set (value) {
        this.$store.state.sceneList.filters.sites = value
        this.reloadList()
      }
    },
    tags: {
      get () {
        return this.$store.state.sceneList.filters.tags
      },
      set (value) {
        this.$store.state.sceneList.filters.tags = value
        this.reloadList()
      }
    },
    cuepoint: {
      get () {
        return this.$store.state.sceneList.filters.cuepoint
      },
      set (value) {
        this.$store.state.sceneList.filters.cuepoint = value
        this.reloadList()
      }
    },
    attributes: {
      get () {
        return this.$store.state.sceneList.filters.attributes
      },
      set (value) {
        this.$store.state.sceneList.filters.attributes = value
        this.reloadList()        
      }
    },
    sort: {
      get () {
        return this.$store.state.sceneList.filters.sort
      },
      set (value) {
        this.$store.state.sceneList.filters.sort = value
        this.reloadList()
      }
    },
    isWatched: {
      get () {
        return this.$store.state.sceneList.filters.isWatched
      },
      set (value) {
        this.$store.state.sceneList.filters.isWatched = value
        this.reloadList()
      }
    },
    disableNewAkaGroup() {
      let akaCastCnt = 0
      let actorCnt = 0       
 
      for (var i = 0; i < this.cast.length; i++) {        
        if (this.cast[i].startsWith("aka:")) {
          akaCastCnt++
        } else {
          actorCnt++
        }
      }
      // you can create a new group from a list of actors (more than one)
      return akaCastCnt == 0 && actorCnt > 1 ? false : true
    },
    disableDeleteAkaGroup() {
      let akaCastCnt = 0
      let actorCnt = 0       
 
      for (var i = 0; i < this.cast.length; i++) {        
        if (this.cast[i].startsWith("aka:")) {
          akaCastCnt++
        } else {
          actorCnt++
        }
      }

      // you can only delete a group when it is the only thing selected      
      return akaCastCnt == 1 && actorCnt == 0 > 1 ? false : true
    },
    disableAddToAkaGroup() {
      let akaCastCnt = 0
      let actorCnt = 0       
 
      for (var i = 0; i < this.cast.length; i++) {        
        if (this.cast[i].startsWith("aka:")) {
          akaCastCnt++
        } else {
          actorCnt++
        }
      }

      // you can add to a group if you select one group and one or more actors
      return akaCastCnt == 1 && actorCnt > 0 ? false : true
    },
    disableRemoveFromAkaGroup() {
      let akaCastCnt = 0
      let actorCnt = 0       
 
      for (var i = 0; i < this.cast.length; i++) {        
        if (this.cast[i].startsWith("aka:")) {
          akaCastCnt++
        } else {
          actorCnt++
        }
      }

      // you can remove from a group if you select one group and one or more actors
      return akaCastCnt == 1 && actorCnt > 0 ? false : true

    },
    disableNewTagGroup() {
      let akaTagCnt = 0
      let tagCnt = 0       
 
      for (var i = 0; i < this.tags.length; i++) {        
        if (this.tags[i].startsWith("tag group:")) {
          akaTagCnt++
        } else {
          tagCnt++
        }
      }
      // you can create a new group from a list of tags (more than one)
      return akaTagCnt == 0 && tagCnt > 1 ? false : true
    },
    disableDeleteRenameTagGroup() {
      let tagGroupCnt = 0
      let tagCnt = 0       
 
      for (var i = 0; i < this.tags.length; i++) {        
        if (this.tags[i].startsWith("tag group:")) {
          tagGroupCnt++
        } else {
          tagCnt++
        }
      }

      // you can only delete a group when it is the only thing selected      
      return tagGroupCnt == 1 && tagCnt == 0 > 1 ? false : true
    },
    disableAddToTagGroup() {
      let tagGroupCnt = 0
      let tagCnt = 0       
 
      for (var i = 0; i < this.tags.length; i++) {        
        if (this.tags[i].startsWith("tag group:")) {
          tagGroupCnt++
        } else {
          tagCnt++
        }
      }

      // you can add to a group if you select one group and one or more tags
      return tagGroupCnt == 1 && tagCnt > 0 ? false : true
    },
    disableRemoveFromTagGroup() {
      let tagGroupCnt = 0
      let tagCnt = 0       
 
      for (var i = 0; i < this.tags.length; i++) {        
        if (this.tags[i].startsWith("tag group:")) {
          tagGroupCnt++
        } else {
          tagCnt++
        }
      }

      // you can remove from a group if you select one group and one or more tag
      return tagGroupCnt == 1 && tagCnt > 0 ? false : true

    },
    disableGetTagGroup() {
      let tagGroupCnt = 0
 
      for (var i = 0; i < this.tags.length; i++) {        
        if (this.tags[i].startsWith("tag group:")) {
          tagGroupCnt++
        }
      }

      // you can list a group if you select one 
      return tagGroupCnt == 1 ? false : true

    },
}
}
</script>

<style lang="scss" scoped>
.filters-panel {
  display: flex;
  flex-direction: column;
  gap: 1.4rem;
  padding-top: 0.25rem;
}

.filter-heading {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  width: 100%;
  margin-bottom: 0.6rem;
  padding-bottom: 0.45rem;
  border: none;
  border-bottom: 1px solid var(--xbvr-border, #e3e6ec);
  background: none;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--xbvr-text-muted, #64708a);
  text-align: left;
}

button.filter-heading {
  cursor: pointer;
  transition: color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

button.filter-heading:hover {
  color: var(--xbvr-text, #1c2333);
}

.heading-chevron {
  margin-left: auto;
  transition: transform var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

.heading-chevron.open {
  transform: rotate(180deg);
}

.filter-hint {
  font-size: 0.72rem;
  color: var(--xbvr-text-faint, #98a1b6);
  margin: -0.25rem 0 0.7rem;
}

.filter-field {
  margin-bottom: 0.8rem;
}

.filter-field:last-child {
  margin-bottom: 0;
}

.filter-label {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--xbvr-text-muted, #64708a);
  margin-bottom: 0.3rem;
  cursor: pointer;
}

/* property toggles: 2×2 grid, full-width buttons */
.prop-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.4rem;
}

.prop-grid :deep(.b-checkbox.button) {
  width: 100%;
  justify-content: flex-start;
  margin: 0;
  box-shadow: none;
}

/* group-management buttons: icon + label, 2 per row */
.btn-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.4rem;
}

.btn-grid :deep(.b-tooltip) {
  width: 100%;
}

.btn-grid .button {
  width: 100%;
  justify-content: flex-start;
  margin: 0;
}

.tagicon {
  margin-right: -0.2em !important;
}
</style>
