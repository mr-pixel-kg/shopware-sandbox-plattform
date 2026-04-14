<script setup lang="ts">
import { Package } from 'lucide-vue-next'
import { computed } from 'vue'

import MetadataActionRenderer from '@/components/metadata/MetadataActionRenderer.vue'
import MetadataSection from '@/components/metadata/MetadataSection.vue'
import TtlChip from '@/components/sandboxes/TtlChip.vue'
import StatusBadge from '@/components/shared/StatusBadge.vue'
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { extractFieldValues, itemsForContext } from '@/utils/metadata'

import ActionButton from './ActionButton.vue'

import type { CardAction } from './ActionButton.vue'
import type { ActionItem, MetadataContext, Sandbox } from '@/types'

const props = defineProps<{
  sandbox: Sandbox
  title: string
  thumbnailUrl?: string
  extraActions?: CardAction[]
  stateReason?: string
  context?: MetadataContext
}>()

const ctx = computed<MetadataContext>(() => props.context ?? 'sandbox.card')
const values = computed(() => extractFieldValues(props.sandbox.metadata))
const schemaActions = computed(() =>
  itemsForContext(props.sandbox.metadata, ctx.value).filter(
    (i): i is ActionItem => i.type === 'action',
  ),
)

const destructiveExtras = computed(() =>
  (props.extraActions ?? []).filter((a) => a.variant === 'destructive'),
)
const primaryExtras = computed(() =>
  (props.extraActions ?? []).filter((a) => a.variant !== 'destructive'),
)
</script>

<template>
  <Card class="flex h-full flex-col overflow-hidden pt-0">
    <div class="bg-muted relative flex h-36 shrink-0 items-center justify-center">
      <img
        v-if="thumbnailUrl"
        :src="thumbnailUrl"
        :alt="title"
        class="h-full w-full object-contain"
      />
      <Package v-else class="text-muted-foreground/40 h-8 w-8" />
    </div>
    <CardHeader>
      <div class="flex items-start justify-between gap-2">
        <CardTitle class="truncate text-sm">{{ title }}</CardTitle>
        <div class="flex items-center gap-2">
          <StatusBadge :status="sandbox.status" :state-reason="stateReason" />
        </div>
      </div>
    </CardHeader>
    <CardContent class="flex-1 space-y-3">
      <TtlChip
        v-if="sandbox.expiresAt"
        :expires-at="sandbox.expiresAt"
        :created-at="sandbox.createdAt"
      />
      <MetadataSection
        :metadata="sandbox.metadata"
        :context="ctx"
        :model-value="values"
        view="card"
        hide-actions
      />
    </CardContent>
    <CardFooter
      v-if="schemaActions.length || extraActions?.length"
      class="sandbox-actions flex items-center gap-2"
    >
      <MetadataActionRenderer
        v-for="item in schemaActions"
        :key="item.key"
        :item="item"
        :disabled="sandbox.status !== 'running'"
        class="min-w-0 flex-1"
      />
      <ActionButton
        v-for="action in primaryExtras"
        :key="action.label"
        :action="action"
        :full-width="action.size !== 'icon'"
        :class="action.size === 'icon' ? 'shrink-0' : 'min-w-0 flex-1'"
      />
      <ActionButton
        v-for="action in destructiveExtras"
        :key="action.label"
        :action="action"
        class="shrink-0"
      />
    </CardFooter>
  </Card>
</template>

<style scoped>
.sandbox-actions {
  scrollbar-width: none;
  -ms-overflow-style: none;
}
.sandbox-actions::-webkit-scrollbar {
  display: none;
}
</style>
