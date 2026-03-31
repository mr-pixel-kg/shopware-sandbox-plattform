<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'
import { computed } from 'vue'

import { Badge } from '@/components/ui/badge'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'

import type { SandboxStatus } from '@/types'

const props = defineProps<{
  status: SandboxStatus
  stateReason?: string
}>()

const config = computed(() => {
  switch (props.status) {
    case 'running':
      return {
        label: 'Aktiv',
        variant: 'outline' as const,
        class: 'bg-green-500/15 text-green-700 border-green-500/25',
        spinner: false,
      }
    case 'starting':
      return {
        label: 'Startet',
        variant: 'outline' as const,
        class: 'bg-yellow-500/15 text-yellow-700 border-yellow-500/25',
        spinner: true,
      }
    case 'paused':
      return {
        label: 'Pausiert',
        variant: 'outline' as const,
        class: 'bg-blue-500/15 text-blue-700 border-blue-500/25',
        spinner: true,
      }
    case 'stopping':
      return {
        label: 'Wird beendet',
        variant: 'outline' as const,
        class: 'bg-orange-500/15 text-orange-700 border-orange-500/25',
        spinner: true,
      }
    case 'stopped':
      return { label: 'Gestoppt', variant: 'secondary' as const, class: '', spinner: false }
    case 'expired':
      return { label: 'Abgelaufen', variant: 'secondary' as const, class: '', spinner: false }
    case 'failed':
      return {
        label: 'Fehlgeschlagen',
        variant: 'destructive' as const,
        class: '',
        spinner: false,
      }
    case 'deleted':
      return { label: 'Gelöscht', variant: 'secondary' as const, class: '', spinner: false }
    default:
      return { label: props.status, variant: 'secondary' as const, class: '', spinner: false }
  }
})
</script>

<template>
  <TooltipProvider v-if="stateReason">
    <Tooltip>
      <TooltipTrigger as-child>
        <Badge :variant="config.variant" :class="config.class" class="gap-1">
          <Loader2 v-if="config.spinner" class="h-3 w-3 animate-spin" />
          {{ config.label }}
        </Badge>
      </TooltipTrigger>
      <TooltipContent>{{ stateReason }}</TooltipContent>
    </Tooltip>
  </TooltipProvider>
  <Badge v-else :variant="config.variant" :class="config.class" class="gap-1">
    <Loader2 v-if="config.spinner" class="h-3 w-3 animate-spin" />
    {{ config.label }}
  </Badge>
</template>
