<template>
  <div class="columns is-gapless is-centered az-jump">
    <b-radio-button :value="value" @input="$emit('input', $event)" native-value="" size="is-small"></b-radio-button>
    <b-radio-button :value="value" @input="$emit('input', $event)" native-value="A" size="is-small">A</b-radio-button>
    <b-radio-button :value="value" @input="$emit('input', $event)" native-value="B" size="is-small">B</b-radio-button>
    <b-radio-button :value="value" @input="$emit('input', $event)" native-value="C" size="is-small">C</b-radio-button>
    <b-radio-button :value="value" @input="$emit('input', $event)" native-value="D" size="is-small">D</b-radio-button>
    <b-radio-button :value="value" @input="$emit('input', $event)" native-value="E" size="is-small">E</b-radio-button>
    <b-radio-button :value="value" @input="$emit('input', $event)" native-value="F" size="is-small">F</b-radio-button>
    <b-radio-button :value="value" @input="$emit('input', $event)" native-value="G" size="is-small">G</b-radio-button>
    <b-radio-button :value="value" @input="$emit('input', $event)" native-value="H" size="is-small">H</b-radio-button>
    <b-radio-button :value="value" @input="$emit('input', $event)" native-value="I" size="is-small">I</b-radio-button>
    <b-radio-button :value="value" @input="$emit('input', $event)" native-value="J" size="is-small">J</b-radio-button>
    <b-radio-button :value="value" @input="$emit('input', $event)" native-value="K" size="is-small">K</b-radio-button>
    <b-radio-button :value="value" @input="$emit('input', $event)" native-value="L" size="is-small">L</b-radio-button>
    <b-radio-button :value="value" @input="$emit('input', $event)" native-value="M" size="is-small">M</b-radio-button>
    <b-radio-button :value="value" @input="$emit('input', $event)" native-value="N" size="is-small">N</b-radio-button>
    <b-radio-button :value="value" @input="$emit('input', $event)" native-value="O" size="is-small">O</b-radio-button>
    <b-radio-button :value="value" @input="$emit('input', $event)" native-value="P" size="is-small">P</b-radio-button>
    <b-radio-button :value="value" @input="$emit('input', $event)" native-value="Q" size="is-small">Q/R</b-radio-button>
    <b-radio-button :value="value" @input="$emit('input', $event)" native-value="S" size="is-small">S</b-radio-button>
    <b-radio-button :value="value" @input="$emit('input', $event)" native-value="T" size="is-small">T</b-radio-button>
    <b-radio-button :value="value" @input="$emit('input', $event)" native-value="U" size="is-small">U/V</b-radio-button>
    <b-radio-button :value="value" @input="$emit('input', $event)" native-value="W" size="is-small">W/X/Y/Z</b-radio-button>
  </div>
</template>

<script>
export default {
  name: 'AZJumpFilter',
  props: { value: String }
}
</script>

<style scoped>
.az-jump {
  flex-wrap: wrap;
  row-gap: 4px;
}

.az-jump :deep(.b-radio.button) {
  border-radius: 8px;
  min-width: 1.9rem;
  justify-content: center;
  font-weight: 600;
  color: var(--xbvr-text-muted, #64708a);
  border-color: var(--xbvr-border, #e3e6ec);
  background: var(--xbvr-surface, #ffffff);
  transition: color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
    border-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1)),
    background-color var(--xbvr-fast, 140ms) var(--xbvr-ease, cubic-bezier(0.2, 0, 0, 1));
}

.az-jump :deep(.b-radio.button:hover) {
  color: var(--xbvr-text, #1c2333);
  border-color: var(--xbvr-border-strong, #cdd2dc);
  background: var(--xbvr-hover-bg, #fafbfd);
}
</style>
