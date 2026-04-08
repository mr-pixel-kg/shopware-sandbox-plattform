<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'

import { sandboxesApi } from '@/api'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { LOG_TERMINAL_BG, useLogStream } from '@/composables/useLogStream'

import '@xterm/xterm/css/xterm.css'

import type { LogSource } from '@/types'

const props = defineProps<{
  sandboxId: string
}>()

const sources = ref<LogSource[]>([])
const selectedKey = ref<string>('')
const loadingMeta = ref(true)
const containerRef = ref<HTMLElement | null>(null)

const { isStreaming, error, connect, dispose } = useLogStream(containerRef)

async function loadSources() {
  loadingMeta.value = true
  try {
    sources.value = await sandboxesApi.listLogSources(props.sandboxId)
    if (sources.value.length > 0 && !selectedKey.value) {
      selectedKey.value = sources.value[0].key
    }
  } catch {
    sources.value = []
  } finally {
    loadingMeta.value = false
  }
}

watch(selectedKey, (key) => {
  if (key) {
    connect(props.sandboxId, key)
  }
})

onMounted(loadSources)

defineExpose({ disconnect: dispose })
</script>

<template>
  <div class="flex h-full flex-col gap-3">
    <div class="flex shrink-0 items-center gap-2">
      <Select v-model="selectedKey" :disabled="loadingMeta || sources.length === 0">
        <SelectTrigger class="w-56">
          <SelectValue placeholder="Log-Quelle wählen" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="src in sources" :key="src.key" :value="src.key">
            {{ src.label }}
          </SelectItem>
        </SelectContent>
      </Select>

      <Button
        v-if="error && !isStreaming"
        variant="outline"
        size="sm"
        @click="connect(sandboxId, selectedKey)"
      >
        Erneut verbinden
      </Button>
    </div>

    <div class="relative min-h-0 flex-1">
      <div
        v-if="loadingMeta"
        class="absolute inset-0 z-10 flex items-center justify-center"
        :style="{ backgroundColor: LOG_TERMINAL_BG }"
      >
        <span class="text-sm text-neutral-400">Lade Log-Quellen…</span>
      </div>

      <div
        v-else-if="sources.length === 0"
        class="absolute inset-0 z-10 flex items-center justify-center"
        :style="{ backgroundColor: LOG_TERMINAL_BG }"
      >
        <span class="text-sm text-neutral-400">Keine Log-Quellen konfiguriert</span>
      </div>

      <div
        v-if="!isStreaming && !error && !loadingMeta && selectedKey"
        class="absolute inset-0 z-10 flex items-center justify-center"
        :style="{ backgroundColor: LOG_TERMINAL_BG }"
      >
        <span class="text-sm text-neutral-400">Verbindung wird hergestellt…</span>
      </div>

      <div
        v-if="error && !isStreaming"
        class="absolute inset-0 z-10 flex flex-col items-center justify-center gap-3"
        :style="{ backgroundColor: LOG_TERMINAL_BG }"
      >
        <span class="text-sm text-red-400">{{ error }}</span>
      </div>

      <div ref="containerRef" class="h-full" />
    </div>
  </div>
</template>
