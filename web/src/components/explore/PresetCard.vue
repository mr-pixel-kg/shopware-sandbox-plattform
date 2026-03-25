<script setup lang="ts">
import { Package } from 'lucide-vue-next'
import { computed } from 'vue'

import { Card, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { resolveAssetUrl } from '@/utils/formatters'

import ActionButton from './ActionButton.vue'

import type { CardAction } from './ActionButton.vue'
import type { Image } from '@/types'

const props = defineProps<{
  image: Image
  actions: CardAction[]
}>()

const thumbnailSrc = computed(() => resolveAssetUrl(props.image.thumbnailUrl))
</script>

<template>
  <Card class="overflow-hidden pt-0">
    <div class="bg-muted relative flex h-36 items-center justify-center">
      <img
        v-if="thumbnailSrc"
        :src="thumbnailSrc"
        :alt="image.title || image.name"
        class="h-full w-full object-cover"
      />
      <Package v-else class="text-muted-foreground/40 h-10 w-10" />
    </div>
    <CardHeader class="flex-1">
      <div>
        <CardTitle class="text-sm">{{ image.title || image.name }}</CardTitle>
        <p class="text-muted-foreground mt-0.5 font-mono text-xs">
          {{ image.name }}:{{ image.tag }}
        </p>
      </div>
      <CardDescription v-if="image.description" class="line-clamp-2">
        {{ image.description }}
      </CardDescription>
    </CardHeader>
    <CardFooter class="mt-auto flex gap-2 overflow-x-auto">
      <ActionButton v-for="action in actions" :key="action.label" :action="action" />
    </CardFooter>
  </Card>
</template>
