<script setup lang="ts">
import { Package } from 'lucide-vue-next'
import { computed } from 'vue'

import MetadataActionRenderer from '@/components/metadata/MetadataActionRenderer.vue'
import MetadataSection from '@/components/metadata/MetadataSection.vue'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { resolveAssetUrl } from '@/utils/formatters'
import { extractFieldValues, itemsForContext } from '@/utils/metadata'

import ActionButton from './ActionButton.vue'

import type { CardAction } from './ActionButton.vue'
import type { ActionItem, Image } from '@/types'

const props = defineProps<{
  image: Image
  extraActions?: CardAction[]
}>()

const thumbnailSrc = computed(() => resolveAssetUrl(props.image.thumbnailUrl))
const values = computed(() => extractFieldValues(props.image.metadata))
const schemaActions = computed(() =>
  itemsForContext(props.image.metadata, 'image.card').filter(
    (i): i is ActionItem => i.type === 'action',
  ),
)

const primaryExtras = computed(() =>
  (props.extraActions ?? []).filter((a) => a.variant !== 'destructive'),
)
const destructiveExtras = computed(() =>
  (props.extraActions ?? []).filter((a) => a.variant === 'destructive'),
)
</script>

<template>
  <Card class="overflow-hidden pt-0">
    <div class="bg-muted relative flex h-36 shrink-0 items-center justify-center">
      <img
        v-if="thumbnailSrc"
        :src="thumbnailSrc"
        :alt="image.title || image.name"
        class="h-full w-full object-contain"
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
    <CardContent class="space-y-3 pt-0">
      <MetadataSection
        :metadata="image.metadata"
        context="image.card"
        :model-value="values"
        view="card"
        hide-actions
      />
    </CardContent>
    <CardFooter class="mt-auto flex items-center gap-2">
      <div class="flex min-w-0 flex-1 gap-2 [&>*]:flex-1">
        <MetadataActionRenderer v-for="item in schemaActions" :key="item.key" :item="item" />
        <ActionButton
          v-for="action in primaryExtras"
          :key="action.label"
          :action="action"
          full-width
        />
      </div>
      <ActionButton
        v-for="action in destructiveExtras"
        :key="action.label"
        :action="action"
        class="shrink-0"
      />
    </CardFooter>
  </Card>
</template>
