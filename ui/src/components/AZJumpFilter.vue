<template>
  <div class="columns is-gapless is-centered">
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
