<script setup lang="ts">
import { computed } from 'vue'
import type { SandboxStatus } from '@/types'

const props = defineProps<{
  status: SandboxStatus
}>()

const colorClass = computed(() => {
  switch (props.status) {
    case 'running':
      return 'bg-green-500'
    case 'starting':
      return 'bg-yellow-500'
    default:
      return 'bg-muted-foreground/40'
  }
})

const shouldPulse = computed(() => props.status === 'running' || props.status === 'starting')
</script>

<template>
  <span class="relative flex h-2.5 w-2.5">
    <span
      v-if="shouldPulse"
      class="absolute inline-flex h-full w-full animate-ping rounded-full opacity-75"
      :class="colorClass"
    />
    <span class="relative inline-flex h-2.5 w-2.5 rounded-full" :class="colorClass" />
  </span>
</template>
