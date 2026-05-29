<script setup>
import { ref, computed } from 'vue'
import AppIcon from './AppIcon.vue'
import { iconGroups, getIconLabel } from '../constants/icons.js'

const props = defineProps({
  modelValue: { type: String, default: '' },
})

const emit = defineEmits(['update:modelValue'])

const open = ref(false)
const search = ref('')
const activeGroup = ref('')

const searchQuery = computed(() => search.value.trim().toLowerCase())

const filteredIcons = computed(() => {
  if (!searchQuery.value) return null
  const results = []
  for (const group of iconGroups) {
    for (const icon of group.icons) {
      const label = getIconLabel(icon).toLowerCase()
      if (icon.toLowerCase().includes(searchQuery.value) || label.includes(searchQuery.value)) {
        results.push(icon)
      }
    }
  }
  return results.slice(0, 48)
})

const displayGroups = computed(() => {
  if (searchQuery.value) return null
  return iconGroups
})

function selectIcon(iconKey) {
  emit('update:modelValue', iconKey)
  open.value = false
  search.value = ''
}

function toggle() {
  open.value = !open.value
  if (!open.value) {
    search.value = ''
    activeGroup.value = ''
  }
}

function clearIcon() {
  emit('update:modelValue', '')
  open.value = false
}
</script>

<template>
  <div class="icon-picker">
    <div class="icon-picker-current">
      <span class="icon-picker-preview">
        <AppIcon :name="modelValue || 'unknown'" :size="22" />
      </span>
      <span class="icon-picker-key">{{ getIconLabel(modelValue) }}</span>
      <span v-if="modelValue" class="icon-picker-key-en">{{ modelValue }}</span>
      <button type="button" class="btn-ghost icon-picker-btn" @click="toggle">
        {{ open ? '收起' : '选择图标' }}
      </button>
    </div>

    <div v-if="open" class="icon-picker-panel">
      <input
        v-model="search"
        type="search"
        placeholder="搜索图标（中/英文）..."
        class="icon-picker-search"
      />

      <div v-if="searchQuery && filteredIcons.length === 0" class="icon-picker-empty">
        未找到「{{ search }}」相关图标
      </div>

      <!-- Flat search results -->
      <div v-if="filteredIcons" class="icon-picker-grid">
        <button
          v-for="icon in filteredIcons"
          :key="icon"
          type="button"
          class="icon-picker-item"
          :class="{ selected: modelValue === icon }"
          @click="selectIcon(icon)"
          :title="icon"
        >
          <span class="icon-picker-item-icon"><AppIcon :name="icon" :size="22" /></span>
          <span class="icon-picker-item-label">{{ getIconLabel(icon) }}</span>
        </button>
      </div>

      <!-- Grouped display -->
      <template v-else-if="displayGroups">
        <div class="icon-picker-tabs">
          <button
            v-for="group in displayGroups"
            :key="group.key"
            type="button"
            class="icon-picker-tab"
            :class="{ active: activeGroup === group.key }"
            @click="activeGroup = activeGroup === group.key ? '' : group.key"
          >
            {{ group.label }}
          </button>
        </div>

        <template v-for="group in displayGroups" :key="group.key">
          <div v-if="!activeGroup || activeGroup === group.key">
            <div class="icon-picker-group-label">{{ group.label }}</div>
            <div class="icon-picker-grid">
              <button
                v-for="icon in group.icons"
                :key="icon"
                type="button"
                class="icon-picker-item"
                :class="{ selected: modelValue === icon }"
                @click="selectIcon(icon)"
                :title="icon"
              >
                <span class="icon-picker-item-icon"><AppIcon :name="icon" :size="22" /></span>
                <span class="icon-picker-item-label">{{ getIconLabel(icon) }}</span>
              </button>
            </div>
          </div>
        </template>
      </template>

      <div v-if="modelValue" class="icon-picker-clear-row">
        <button type="button" class="btn-ghost" @click="clearIcon">清除图标</button>
      </div>
    </div>
  </div>
</template>
