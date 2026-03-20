<script setup lang="ts">
import type { Image } from '@/types'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Package } from 'lucide-vue-next'
import ActionButton from './ActionButton.vue'
import type { CardAction } from './ActionButton.vue'

defineProps<{
  image: Image
  actions: CardAction[]
}>()
</script>

<template>
  <Card class="overflow-hidden pt-0">
    <div class="relative h-36 bg-muted flex items-center justify-center">
      <img
        v-if="image.thumbnailUrl"
        :src="image.thumbnailUrl"
        :alt="image.title || image.name"
        class="h-full w-full object-cover"
      />
      <Package v-else class="h-10 w-10 text-muted-foreground/40" />
    </div>
    <CardHeader class="flex-1">
      <div>
        <CardTitle class="text-sm">{{ image.title || image.name }}</CardTitle>
        <p class="text-xs text-muted-foreground font-mono mt-0.5">{{ image.name }}:{{ image.tag }}</p>
      </div>
      <CardDescription v-if="image.description" class="line-clamp-2">
        {{ image.description }}
      </CardDescription>
    </CardHeader>
    <!-- TODO: Replace with dynamic schema from API -->
    <CardFooter class="flex gap-2 mt-auto">
      <ActionButton
        v-for="action in actions"
        :key="action.label"
        :action="action"
      />
    </CardFooter>
  </Card>
</template>
