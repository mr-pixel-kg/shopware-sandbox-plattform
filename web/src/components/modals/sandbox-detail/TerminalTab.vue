<script setup lang="ts">
import { onMounted, ref, toRef } from 'vue'

import { Button } from '@/components/ui/button'
import { TERMINAL_BG, useTerminal } from '@/composables/useTerminal'

import '@xterm/xterm/css/xterm.css'

const props = defineProps<{
  sandboxId: string
}>()

const containerRef = ref<HTMLElement | null>(null)

const { isConnected, error, connect, disconnect } = useTerminal(
  toRef(() => props.sandboxId),
  containerRef,
)

onMounted(connect)

defineExpose({ disconnect })
</script>

<template>
  <div class="relative flex h-full flex-col">
    <div
      v-if="!isConnected && !error"
      class="absolute inset-0 z-10 flex items-center justify-center"
      :style="{ backgroundColor: TERMINAL_BG }"
    >
      <span class="text-sm text-neutral-400">Verbindung wird hergestellt…</span>
    </div>

    <div
      v-if="error && !isConnected"
      class="absolute inset-0 z-10 flex flex-col items-center justify-center gap-3"
      :style="{ backgroundColor: TERMINAL_BG }"
    >
      <span class="text-sm text-red-400">{{ error }}</span>
      <Button variant="outline" size="sm" @click="connect">Erneut verbinden</Button>
    </div>

    <div ref="containerRef" class="min-h-0 flex-1" />
  </div>
</template>
