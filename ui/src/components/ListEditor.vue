<template>
  <section class="list-editor">
    <div class="list-row" v-for="(item, i) in list" :key="`item-${i}`">
      <div class="list-row-inputs">
        <b-input v-if = "columnCount > 1" v-for="fieldidx in columnCount" :key="fieldidx" :class="`list-editor-input list-editor-input-${type}-${i}`" :value="item[fieldidx-1]" @blur="blur(i)" 
          :placeholder=getPlaceholder(fieldidx-1) :style="getColumnStyle(fieldidx-1)" />
        <b-input v-if = "columnCount == undefined || columnCount == 1" :class="`list-editor-input list-editor-input-${type}-${i}`" :value="item" @blur="blur(i)" :placeholder=getPlaceholder(1) />
      </div>
      <div class="list-row-actions">
        <!--<b-button type="is-danger" @click="deleteRow(i)">Delete</b-button>-->
        <b-button type="is-light" @click="deleteRow(i)" icon-right="delete" />
        <a v-if="showUrl" class="button is-light" 
          :title="`Go to $(${item}`" :href="item" target="_blank" rel="noreferrer">
          <b-icon pack="mdi" icon="link" size="is-small" />
        </a>
      </div>
    </div>

    <b-field class="list-editor-add">
      <b-button class="control" type="is-info" icon-right="plus-circle-outline" @click="addRow">{{$t('Add item')}}</b-button>
    </b-field>
  </section>
</template>

<script>
export default {
  name: 'List2Editor',
  props: {
    list: Array,
    type: String,
    blurFn: Function,
    showUrl: Boolean,
    columnCount: Number,
    placeholders: Array,
    columnStyles: Array,
  },
  methods: {
    addRow () {
      this.list.push('')
    },
    deleteRow (i) {
      this.list.splice(i, 1)
    },
    blur (i) {
      this.list[i] = document.querySelector(`.list-editor-input-${this.type}-${i} input`).value
      this.blurFn.call(null)
    },
    getPlaceholder (i) {
      if (this.placeholders == undefined){
        return ""
      }
      if (i+1 > this.placeholders.length ) {
        return ""
      }
      return this.placeholders[i]
    },
    getColumnStyle (i) {
      if (this.columnStyles == undefined){
        return ""
      }
      if (i+1 > this.columnStyles.length ) {
        return ""
      }
      return this.columnStyles[i]      
    },
  }
}
</script>

<style scoped>
  .list-editor {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .list-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.45rem 0.6rem;
    background: var(--xbvr-surface, #ffffff);
    border: 1px solid var(--xbvr-border, #e3e6ec);
    border-radius: var(--xbvr-radius, 12px);
    box-shadow: var(--xbvr-shadow-sm, 0 1px 2px rgba(16, 24, 40, 0.05));
    transition: border-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
      box-shadow var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
  }

  .list-row:hover {
    border-color: var(--xbvr-border-strong, #cdd2dc);
    box-shadow: var(--xbvr-shadow, 0 1px 3px rgba(16, 24, 40, 0.08), 0 1px 2px rgba(16, 24, 40, 0.04));
  }

  .list-row-inputs {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex: 1 1 auto;
    min-width: 0;
  }

  .list-row-inputs .control {
    flex: 1 1 auto;
    min-width: 0;
  }

  .list-editor-input {
    width: 100%;
  }

  /* ordering/delete actions grouped right */
  .list-row-actions {
    display: flex;
    align-items: center;
    gap: 0.3rem;
    flex: 0 0 auto;
  }

  .list-row-actions .button {
    border-radius: 8px;
  }

  .list-editor-add {
    margin-top: 0.25rem;
  }
</style>
