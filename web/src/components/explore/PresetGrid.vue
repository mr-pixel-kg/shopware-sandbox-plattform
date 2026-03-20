<script setup lang="ts">
import type { Image } from '@/types'
import PresetCard from './PresetCard.vue'
import EmptyState from '@/components/shared/EmptyState.vue'

defineProps<{
  images: Image[]
}>()

const emit = defineEmits<{
  start: [imageId: string]
}>()
</script>

<template>
  <div
    v-if="images.length > 0"
    class="grid grid-cols-[repeat(auto-fill,minmax(240px,1fr))] gap-4"
  >
    <PresetCard
      v-for="image in images"
      :key="image.id"
      :image="image"
      @start="emit('start', $event)"
    />
  </div>
  <EmptyState
    v-else
    title="Keine Vorlagen gefunden"
    description="Es sind keine Vorlagen für diesen Filter verfügbar."
  />
</template>
