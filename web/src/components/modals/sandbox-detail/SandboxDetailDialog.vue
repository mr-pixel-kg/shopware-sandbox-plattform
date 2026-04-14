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

import AccessTab from './AccessTab.vue'
import ConfigTab from './ConfigTab.vue'
import FilesTab from './FilesTab.vue'
import HealthTab from './HealthTab.vue'
import LogsTab from './LogsTab.vue'
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
const logsTabRef = ref<InstanceType<typeof LogsTab> | null>(null)

watch(
  () => props.open,
  (open) => {
    if (!open) {
      configTabRef.value?.resetRevealed()
      terminalTabRef.value?.disconnect()
      logsTabRef.value?.disconnect()
    }
  },
)

const isActive = computed(() => {
  const s = props.sandbox?.status
  return s === 'running' || s === 'starting' || s === 'paused' || s === 'stopping'
})

const isOffline = computed(
  () => props.sandbox?.status === 'running' && props.health && !props.health.ready,
)

const visibleMetadata = computed(() => {
  if (!Array.isArray(props.sandbox?.metadata)) return []
  return props.sandbox.metadata.filter((m) => m.show !== 'template' && m.type !== 'action')
})

const hasConfigTab = computed(() => visibleMetadata.value.length > 0)
const hasAccessTab = computed(() => isActive.value && !!props.sandbox?.ssh)
const hasFilesTab = computed(() => isActive.value)
const hasTerminalTab = computed(() => isActive.value && !!props.sandbox?.owner)
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="flex h-[80vh] flex-col gap-0 overflow-hidden p-0 sm:max-w-[80vw]">
      <DialogHeader class="shrink-0 px-6 pt-6 pb-4">
        <div class="flex items-center gap-3">
          <DialogTitle class="truncate text-lg">
            {{ sandbox?.displayName || image?.title || image?.name || 'Sandbox' }}
          </DialogTitle>
          <StatusBadge
            v-if="sandbox"
            :status="sandbox.status"
            :state-reason="sandbox.stateReason"
          />
          <Badge v-if="isOffline" variant="destructive" class="text-xs">Offline</Badge>
        </div>
        <DialogDescription>
          <template v-if="image">{{ image.name }}:{{ image.tag }}</template>
        </DialogDescription>
      </DialogHeader>

      <Separator />

      <Tabs default-value="overview" class="flex min-h-0 flex-1 flex-col">
        <TabsList class="mx-6 mt-4 w-fit">
          <TabsTrigger value="overview">Übersicht</TabsTrigger>
          <TabsTrigger v-if="hasConfigTab" value="config">Konfiguration</TabsTrigger>
          <TabsTrigger v-if="hasAccessTab" value="access">Zugang</TabsTrigger>
          <TabsTrigger v-if="hasFilesTab" value="files">Dateien</TabsTrigger>
          <TabsTrigger v-if="isActive" value="health">Health</TabsTrigger>
          <TabsTrigger v-if="isActive" value="logs">Logs</TabsTrigger>
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

        <TabsContent
          v-if="hasAccessTab"
          value="access"
          class="flex-1 overflow-y-auto px-6 pt-4 pb-6"
        >
          <AccessTab :ssh="sandbox!.ssh!" />
        </TabsContent>

        <TabsContent
          v-if="hasFilesTab"
          value="files"
          class="flex min-h-0 flex-1 flex-col px-6 pt-4 pb-6"
        >
          <FilesTab />
        </TabsContent>

        <TabsContent v-if="isActive" value="health" class="flex-1 overflow-y-auto px-6 pt-4 pb-6">
          <HealthTab :health="health" />
        </TabsContent>

        <TabsContent
          v-if="isActive"
          value="logs"
          class="flex min-h-0 flex-1 flex-col px-6 pt-4 pb-6"
        >
          <LogsTab ref="logsTabRef" :sandbox-id="sandbox!.id" />
        </TabsContent>

        <TabsContent
          v-if="hasTerminalTab"
          value="terminal"
          class="flex min-h-0 flex-1 flex-col px-6 pt-4 pb-6"
        >
          <TerminalTab ref="terminalTabRef" :sandbox-id="sandbox!.id" />
        </TabsContent>
      </Tabs>
    </DialogContent>
  </Dialog>
</template>
