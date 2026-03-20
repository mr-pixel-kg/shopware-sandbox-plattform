<script setup lang="ts">
import type { Image } from '@/types'
import type { CardAction } from './ActionButton.vue'
import PresetCard from './PresetCard.vue'
import EmptyState from '@/components/shared/EmptyState.vue'

defineProps<{
  images: Image[]
  getActions: (image: Image) => CardAction[]
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
      :actions="getActions(image)"
    />
  </div>
  <EmptyState
    v-else
    title="Keine Vorlagen gefunden"
    description="Es sind keine Vorlagen verfügbar."
  />
</template>
