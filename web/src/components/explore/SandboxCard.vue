<script setup lang="ts">
import { Check, Copy } from 'lucide-vue-next'
import { ref } from 'vue'

import TtlChip from '@/components/sandboxes/TtlChip.vue'
import StatusBadge from '@/components/shared/StatusBadge.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'

import ActionButton from './ActionButton.vue'

import type { CardAction } from './ActionButton.vue'
import type { Sandbox } from '@/types'

// TODO: Replace with dynamic schema from API
export interface MetadataField {
  label: string
  value: string
  secret?: boolean
}

// TODO: Replace with dynamic schema from API
export interface MetadataGroup {
  title: string
  fields: MetadataField[]
}

defineProps<{
  sandbox: Sandbox
  title: string
  actions?: CardAction[]
  metadata?: MetadataGroup[]
}>()

const copiedKey = ref<string>()

async function copyToClipboard(field: MetadataField) {
  await navigator.clipboard.writeText(field.value)
  copiedKey.value = field.label
  setTimeout(() => {
    copiedKey.value = undefined
  }, 1500)
}
</script>

<template>
  <Card class="flex flex-col">
    <CardHeader>
      <div class="flex items-start justify-between gap-2">
        <CardTitle class="truncate text-sm">{{ title }}</CardTitle>
        <StatusBadge :status="sandbox.status" />
      </div>
    </CardHeader>
    <CardContent class="flex-1 space-y-3">
      <TtlChip
        v-if="sandbox.expiresAt"
        :expires-at="sandbox.expiresAt"
        :created-at="sandbox.createdAt"
      />
      <!-- TODO: Replace with dynamic schema from API -->
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
          <span class="text-muted-foreground">{{ field.label }}</span>
          <div class="flex items-center gap-0.5">
            <span class="font-mono text-[11px]">{{ field.value }}</span>
            <Button
              variant="ghost"
              size="icon"
              class="h-3 w-3 min-w-0 p-0"
              @click="copyToClipboard(field)"
            >
              <Check v-if="copiedKey === field.label" class="size-2.5 text-green-500" />
              <Copy v-else class="size-2.5" />
            </Button>
          </div>
        </div>
      </div>
    </CardContent>
    <!-- TODO: Replace with dynamic schema from API -->
    <CardFooter v-if="actions?.length" class="flex gap-2">
      <ActionButton v-for="action in actions" :key="action.label" :action="action" />
    </CardFooter>
  </Card>
</template>
