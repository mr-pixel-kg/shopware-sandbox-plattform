<script setup lang="ts">
import { Package } from 'lucide-vue-next'
import { computed } from 'vue'

import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { resolveAssetUrl } from '@/utils/formatters'
import { resolveIcon } from '@/utils/icons'

import ActionButton from './ActionButton.vue'

import type { CardAction } from './ActionButton.vue'
import type { MetadataGroup } from './SandboxCard.vue'
import type { Image } from '@/types'

const props = defineProps<{
  image: Image
  actions: CardAction[]
  metadata?: MetadataGroup[]
}>()

const thumbnailSrc = computed(() => resolveAssetUrl(props.image.thumbnailUrl))

const primaryActions = computed(() => props.actions.filter((a) => a.variant !== 'destructive'))
const destructiveActions = computed(() => props.actions.filter((a) => a.variant === 'destructive'))
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
    <CardContent v-if="metadata?.length" class="space-y-3 pt-0">
      <div
        v-for="group in metadata"
        :key="group.title"
        class="bg-muted/50 space-y-1.5 rounded-md border px-3 py-2"
      >
        <p class="text-muted-foreground/70 text-[11px] font-medium tracking-wider uppercase">
          {{ group.title }}
        </p>
        <div
          v-for="field in group.fields"
          :key="field.label"
          class="flex items-center justify-between gap-2 text-xs"
        >
          <span class="text-muted-foreground flex items-center gap-1">
            <component :is="resolveIcon(field.icon)" v-if="field.icon" class="h-3 w-3" />
            {{ field.label }}
          </span>
          <span class="font-mono text-[11px]">{{ field.value }}</span>
        </div>
      </div>
    </CardContent>
    <CardFooter class="mt-auto flex items-center gap-2">
      <div class="flex min-w-0 flex-1 gap-2 [&>*]:flex-1">
        <ActionButton
          v-for="action in primaryActions"
          :key="action.label"
          :action="action"
          full-width
        />
      </div>
      <ActionButton
        v-for="action in destructiveActions"
        :key="action.label"
        :action="action"
        class="shrink-0"
      />
    </CardFooter>
  </Card>
</template>
