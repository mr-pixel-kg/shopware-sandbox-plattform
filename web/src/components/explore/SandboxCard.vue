<script setup lang="ts">
import { Check, Copy, Eye, EyeOff, Package } from 'lucide-vue-next'
import { ref } from 'vue'

import TtlChip from '@/components/sandboxes/TtlChip.vue'
import StatusBadge from '@/components/shared/StatusBadge.vue'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { resolveIcon } from '@/utils/icons'

import ActionButton from './ActionButton.vue'

import type { CardAction } from './ActionButton.vue'
import type { Sandbox } from '@/types'

export interface MetadataField {
  label: string
  value: string
  secret?: boolean
  icon?: string
  loading?: boolean
}

export interface MetadataGroup {
  title: string
  fields: MetadataField[]
}

defineProps<{
  sandbox: Sandbox
  title: string
  thumbnailUrl?: string
  actions?: CardAction[]
  metadata?: MetadataGroup[]
  statusNote?: string
}>()

const copiedKey = ref<string>()
const revealedKeys = ref<Set<string>>(new Set())

async function copyToClipboard(field: MetadataField) {
  await navigator.clipboard.writeText(field.value)
  copiedKey.value = field.label
  setTimeout(() => {
    copiedKey.value = undefined
  }, 1500)
}
</script>

<template>
  <Card class="flex h-[460px] flex-col overflow-hidden pt-0">
    <div class="bg-muted relative flex h-36 shrink-0 items-center justify-center">
      <img
        v-if="thumbnailUrl"
        :src="thumbnailUrl"
        :alt="title"
        class="h-full w-full object-cover"
      />
      <Package v-else class="text-muted-foreground/40 h-8 w-8" />
    </div>
    <CardHeader>
      <div class="flex items-start justify-between gap-2">
        <CardTitle class="truncate text-sm">{{ title }}</CardTitle>
        <div class="flex items-center gap-2">
          <StatusBadge :status="sandbox.status" />
          <Badge v-if="statusNote" variant="destructive" class="text-xs">
            {{ statusNote }}
          </Badge>
        </div>
      </div>
    </CardHeader>
    <CardContent class="flex-1 space-y-3">
      <TtlChip
        v-if="sandbox.expiresAt"
        :expires-at="sandbox.expiresAt"
        :created-at="sandbox.createdAt"
      />
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
          <div class="flex items-center gap-1">
            <template v-if="field.loading">
              <Skeleton class="h-3 w-16 rounded" />
            </template>
            <template v-else>
              <span class="font-mono text-[11px]">
                {{ field.secret && !revealedKeys.has(field.label) ? '••••••••' : field.value }}
              </span>
              <button
                v-if="field.secret"
                class="text-muted-foreground hover:text-foreground inline-flex h-5 w-5 items-center justify-center rounded transition-colors"
                @click="
                  revealedKeys.has(field.label)
                    ? revealedKeys.delete(field.label)
                    : revealedKeys.add(field.label)
                "
              >
                <EyeOff v-if="revealedKeys.has(field.label)" class="h-3 w-3" />
                <Eye v-else class="h-3 w-3" />
              </button>
              <button
                class="text-muted-foreground hover:text-foreground inline-flex h-5 w-5 items-center justify-center rounded transition-colors"
                @click="copyToClipboard(field)"
              >
                <Check v-if="copiedKey === field.label" class="h-3 w-3 text-green-500" />
                <Copy v-else class="h-3 w-3" />
              </button>
            </template>
          </div>
        </div>
      </div>
    </CardContent>
    <CardFooter v-if="actions?.length" class="sandbox-actions flex gap-2 overflow-x-auto">
      <ActionButton v-for="action in actions" :key="action.label" :action="action" />
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
