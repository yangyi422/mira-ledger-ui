<script setup>
import { computed } from 'vue'
import { getIconSvg, ICON_DEFAULTS } from '../constants/icons.js'

const props = defineProps({
  name: { type: String, required: true },
  size: { type: Number, default: ICON_DEFAULTS.size },
})

const svgContent = computed(() => getIconSvg(props.name))
</script>

<template>
  <span
    class="app-icon"
    :style="{ width: size + 'px', height: size + 'px' }"
    v-html="svgContent"
  />
</template>
