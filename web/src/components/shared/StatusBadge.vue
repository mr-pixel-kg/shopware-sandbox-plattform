<script setup lang="ts">
import { computed } from 'vue'
import { Badge } from '@/components/ui/badge'
import type { SandboxStatus } from '@/types'

const props = defineProps<{
  status: SandboxStatus
}>()

const config = computed(() => {
  switch (props.status) {
    case 'running':
      return {
        label: 'Aktiv',
        variant: 'outline' as const,
        class: 'bg-green-500/15 text-green-700 border-green-500/25',
      }
    case 'starting':
      return {
        label: 'Startet',
        variant: 'outline' as const,
        class: 'bg-yellow-500/15 text-yellow-700 border-yellow-500/25',
      }
    case 'stopped':
      return { label: 'Gestoppt', variant: 'secondary' as const, class: '' }
    case 'expired':
      return { label: 'Abgelaufen', variant: 'secondary' as const, class: '' }
    case 'failed':
      return { label: 'Fehlgeschlagen', variant: 'destructive' as const, class: '' }
    case 'deleted':
      return { label: 'Gelöscht', variant: 'secondary' as const, class: '' }
    default:
      return { label: props.status, variant: 'secondary' as const, class: '' }
  }
})
</script>

<template>
  <Badge :variant="config.variant" :class="config.class">
    {{ config.label }}
  </Badge>
</template>
