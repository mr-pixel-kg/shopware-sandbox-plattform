<script setup lang="ts">
import EmptyState from '@/components/shared/EmptyState.vue'

import PresetCard from './PresetCard.vue'

import type { CardAction } from './ActionButton.vue'
import type { MetadataGroup } from './SandboxCard.vue'
import type { Image } from '@/types'

defineProps<{
  images: Image[]
  getActions: (image: Image) => CardAction[]
  getMetadata?: (image: Image) => MetadataGroup[]
}>()
</script>

<template>
  <div v-if="images.length > 0" class="grid grid-cols-[repeat(auto-fill,minmax(320px,1fr))] gap-4">
    <PresetCard
      v-for="image in images"
      :key="image.id"
      :image="image"
      :actions="getActions(image)"
      :metadata="getMetadata?.(image)"
    />
  </div>
  <EmptyState
    v-else
    title="Keine Vorlagen gefunden"
    description="Es sind keine Vorlagen verfügbar."
  />
</template>
