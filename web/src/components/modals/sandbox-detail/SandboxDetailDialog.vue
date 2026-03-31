<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import StatusBadge from '@/components/shared/StatusBadge.vue'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Separator } from '@/components/ui/separator'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

import ConfigTab from './ConfigTab.vue'
import HealthTab from './HealthTab.vue'
import OverviewTab from './OverviewTab.vue'
import TerminalTab from './TerminalTab.vue'

import type { Image, Sandbox, SandboxHealthEvent } from '@/types'

const props = defineProps<{
  open: boolean
  sandbox: Sandbox | null
  health?: SandboxHealthEvent
  image?: Image
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
}>()

const configTabRef = ref<InstanceType<typeof ConfigTab> | null>(null)
const terminalTabRef = ref<InstanceType<typeof TerminalTab> | null>(null)

watch(
  () => props.open,
  (open) => {
    if (!open) {
      configTabRef.value?.resetRevealed()
      terminalTabRef.value?.disconnect()
    }
  },
)

const isActive = computed(() => {
  const s = props.sandbox?.status
  return s === 'running' || s === 'starting'
})

const isOffline = computed(
  () => props.sandbox?.status === 'running' && props.health && !props.health.ready,
)

const visibleMetadata = computed(() => {
  if (!props.sandbox?.metadata) return []
  return props.sandbox.metadata.filter((m) => m.show !== 'template' && m.type !== 'action')
})

const hasConfigTab = computed(() => visibleMetadata.value.length > 0)
const hasTerminalTab = computed(() => isActive.value && !!props.sandbox?.owner)
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="flex h-180 flex-col gap-0 overflow-hidden p-0 sm:max-w-4xl">
      <DialogHeader class="shrink-0 px-6 pt-6 pb-4">
        <div class="flex items-center gap-3">
          <DialogTitle class="truncate text-lg">
            {{ sandbox?.displayName || image?.title || image?.name || 'Sandbox' }}
          </DialogTitle>
          <StatusBadge v-if="sandbox" :status="sandbox.status" />
          <Badge v-if="isOffline" variant="destructive" class="text-xs">Offline</Badge>
        </div>
        <DialogDescription v-if="image"> {{ image.name }}:{{ image.tag }} </DialogDescription>
      </DialogHeader>

      <Separator />

      <Tabs default-value="overview" class="flex min-h-0 flex-1 flex-col">
        <TabsList class="mx-6 mt-4 w-fit">
          <TabsTrigger value="overview">Übersicht</TabsTrigger>
          <TabsTrigger v-if="hasConfigTab" value="config">Konfiguration</TabsTrigger>
          <TabsTrigger v-if="isActive" value="health">Health</TabsTrigger>
          <TabsTrigger v-if="hasTerminalTab" value="terminal">Terminal</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" class="flex-1 overflow-y-auto px-6 pt-4 pb-6">
          <OverviewTab v-if="sandbox" :sandbox="sandbox" :image="image" :is-active="isActive" />
        </TabsContent>

        <TabsContent
          v-if="hasConfigTab"
          value="config"
          class="flex-1 overflow-y-auto px-6 pt-4 pb-6"
        >
          <ConfigTab ref="configTabRef" :items="visibleMetadata" />
        </TabsContent>

        <TabsContent v-if="isActive" value="health" class="flex-1 overflow-y-auto px-6 pt-4 pb-6">
          <HealthTab :health="health" />
        </TabsContent>

        <TabsContent
          v-if="hasTerminalTab"
          value="terminal"
          class="min-h-0 flex-1 overflow-hidden px-6 pt-4 pb-6"
        >
          <TerminalTab ref="terminalTabRef" :sandbox-id="sandbox!.id" />
        </TabsContent>
      </Tabs>
    </DialogContent>
  </Dialog>
</template>
