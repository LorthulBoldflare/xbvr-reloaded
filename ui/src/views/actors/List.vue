<template>
  <div>
    <GlobalEvents
      :filter="e => !['INPUT', 'TEXTAREA'].includes(e.target.tagName)"
      @keydown.left="prevpage"
      @keydown.right="nextpage"
      @keydown.o="prevpage"
      @keydown.p="nextpage"
    />
    <b-loading :is-full-page="true" :active.sync="isLoading"></b-loading>

    <div class="columns is-multiline is-full list-toolbar">
      <div class="column is-narrow results-col">
        <strong class="results-count">{{total}} results</strong>
      </div>
      <div class="column pagination-col">
        <b-tooltip :label="$t('Press o/left arrow to page back, p/right arrow to page forward')" :delay="500" position="is-top">
          <b-pagination
              :total="total"
              v-model="current"
              range-before=1
              range-after=3    
              size="is-small"                                           
              :per-page="limit"
              aria-next-label="Next page"
              aria-previous-label="Previous page"
              aria-page-label="Page"
              aria-current-label="Current page"
              :page-input=true
              @change="pageChanged"
              debounce-page-input="250"
              >
          </b-pagination>
        </b-tooltip>
        <span v-show="show_actor_id==='never show, just need the computed show_actor_id to trigger '">{{show_actor_id}}</span>
      </div>
      <div class="column is-narrow">
        <div class="toolbar-end">
          <b-field class="card-size-field">
            <span class="list-header-label">{{$t('Card size')}}</span>
            <b-radio-button v-model="cardSize" native-value="1" size="is-small">
              XS
            </b-radio-button>
            <b-radio-button v-model="cardSize" native-value="2" size="is-small">
              S
            </b-radio-button>
            <b-radio-button v-model="cardSize" native-value="3" size="is-small">
              M
            </b-radio-button>
            <b-radio-button v-model="cardSize" native-value="4" size="is-small">
              L
            </b-radio-button>
          </b-field>
        </div>
      </div>
    </div>
        <AZJumpFilter v-model="jumpTo" v-if="hideLetters" class="az-jump"/>

    <div class="columns is-multiline">
      <div :class="['column', 'is-multiline', cardSizeClass]"
           v-for="actor in actors" :key="actor.id">
        <ActorCard :actor="actor"/>
      </div>
    </div>
      <AZJumpFilter v-model="jumpTo" v-if="hideLetters" class="az-jump az-jump-bottom"/>
      <div class="columns is-gapless is-centered pagination-bottom">
        <b-tooltip :label="$t('Press k to page back, l to page forward')" :delay="500" position="is-top">
          <b-pagination
            :total="total"
            v-model="current"
            range-before=2
            range-after=3 
            size="is-small"                                              
            :per-page="limit"
            aria-next-label="Next page"
            aria-previous-label="Previous page"
            aria-page-label="Page"
            aria-current-label="Current page"
            :page-input=true
            @change="pageChanged"
            debounce-page-input="250"
            >
        </b-pagination>
      </b-tooltip>
      </div>
  </div>
</template>

<script>
import ActorCard from './ActorCard'
import AZJumpFilter from '../../components/AZJumpFilter'
import api from '../../api'
import GlobalEvents from 'vue-global-events'

export default {
  name: 'List',
  components: { ActorCard, AZJumpFilter, GlobalEvents },
  data () {
    return {      
      current: 1,      
    }
  },
  computed: {
    cardSize: {
      get () {
        return this.$store.state.actorList.filters.cardSize
      },
      set (value) {
        this.$store.state.actorList.filters.cardSize = value
        switch (value){
          case "1":
            this.limit=36
            break
          case "2":
            this.limit=18
            break
          case "3":
            this.limit=10
            break
          case "4":
            this.limit=8
            break
            }            
        }      
    },
    limit: {
      get(){
        return this.$store.state.actorList.limit
      },
      set(newLimit){
        // find the position of the first actor
        let currentOffset = this.$store.state.actorList.offset - this.$store.state.actorList.limit + 1
        // what is the new page number, based on the new limit
        this.current = Math.floor(currentOffset / newLimit) + 1
        if (this.current<1)
          this.current=1
        this.$store.state.actorList.limit = newLimit
        // what is the the first actor based on the new page size
        this.$store.state.actorList.offset = (this.current -1) * this.$store.state.actorList.limit          
        this.$store.dispatch('actorList/load', { offset: this.$store.state.actorList.offset })
      }
    },
    jumpTo: {
      get () {
        return this.$store.state.actorList.filters.jumpTo
      },
      set (value) {
        this.$store.state.actorList.filters.jumpTo = value
        this.reloadList()
      }
    },
    cardSizeClass () {
      switch (this.$store.state.actorList.filters.cardSize) {
        case '1':
          return 'is-1'
        case '2':
          return 'is-2'
        case '3':
          return 'is-one-fifth'
        case '4':
          return 'is-one-quarter'
        default:
          return 'is-2'
      }
    },
    isLoading () {
      this.current = this.$store.state.actorList.offset / this.$store.state.actorList.limit
      return this.$store.state.actorList.isLoading
    },
    actors () {
      return this.$store.state.actorList.actors
    },
    total () {
      return this.$store.state.actorList.total
    },
    show_actor_id() {
      if (this.$store.state.actorList.show_actor_id != undefined && this.$store.state.actorList.show_actor_id !='')
      {
        api.get('/actor/'+this.$store.state.actorList.show_actor_id).json().then(data => {
          if (data.id != 0){
            this.$store.commit('overlay/showActorDetails', { actor: data })
          }          
        })
        this.$store.state.actorList.show_actor_id = ''
      }
      
      return this.$store.state.actorList.show_actor_id
    },
    hideLetters: {
      get () {        
        switch (this.$store.state.actorList.filters.sort) {
          case "":
            return true
          case "name_asc":
            return true
          case "name_desc":
            return true
        }
        return false
        },
    },
  },
  methods: {
    reloadList () {
      this.$router.push({
        name: 'actors',
        query: {
          q: this.$store.getters['actorList/filterQueryParams']
        }
      })
    },
    async pageChanged () {      
      this.$store.state.actorList.offset = (this.current -1) * this.$store.state.actorList.limit
      this.$store.dispatch('actorList/load', { offset: this.$store.state.actorList.offset })
    },
    nextpage () {
      if (this.$store.state.overlay.actordetails.show){
        return 
      }
      if (this.$store.state.overlay.details.show){
        return 
      }
      if (this.current * this.limit >= this.total) {
        this.current = 1
      } else {
        this.current += 1
      }      
      this.pageChanged()
    },
    prevpage () {      
      if (this.$store.state.overlay.actordetails.show){
        return 
      }
      if (this.$store.state.overlay.details.show){
        return 
      }
      if (this.current > 1) {
        this.current -= 1
      } else {
        this.current = Math.floor(this.total / this.limit) + 1        
      }      
      this.pageChanged()
    },
  }
}
</script>

<style scoped>
  .list-toolbar {
    align-items: center;
    margin-bottom: 0.5rem;
  }

  .results-col {
    padding-top: 0;
    padding-bottom: 0;
    display: flex;
    align-items: center;
  }

  .results-count {
    font-size: 1.05rem;
    font-weight: 700;
    color: var(--xbvr-text, #1c2333);
    white-space: nowrap;
  }

  .pagination-col {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .toolbar-end {
    display: flex;
    align-items: center;
    justify-content: flex-end;
  }

  .list-header-label {
    padding-right: 0.75em;
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--xbvr-text-muted, #64708a);
    white-space: nowrap;
  }

  .card-size-field {
    align-items: center;
    margin-bottom: 0;
  }

  .az-jump {
    margin: 0.5rem 0 0.75rem;
  }

  .az-jump-bottom {
    margin: 1rem 0 0;
  }

  .pagination-bottom {
    margin-top: 1rem;
    justify-content: center;
  }
</style>
