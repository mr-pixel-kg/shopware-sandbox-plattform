<script setup lang="ts">
import { Badge } from '@/components/ui/badge'
import { formatDateTime } from '@/utils/formatters'

import type { SandboxHealthEvent } from '@/types'

defineProps<{
  health?: SandboxHealthEvent
}>()
</script>

<template>
  <div>
    <table v-if="health" class="w-full text-sm">
      <tbody class="divide-y">
        <tr>
          <td class="text-muted-foreground w-1/3 py-2 pr-4 text-xs">Status</td>
          <td class="py-2">
            <Badge
              :variant="health.ready ? 'outline' : 'destructive'"
              :class="health.ready ? 'border-green-500/25 bg-green-500/15 text-green-700' : ''"
              class="text-xs"
            >
              {{ health.ready ? 'Erreichbar' : 'Nicht erreichbar' }}
            </Badge>
          </td>
        </tr>
        <tr v-if="health.httpStatus">
          <td class="text-muted-foreground w-1/3 py-2 pr-4 text-xs">HTTP Status</td>
          <td class="py-2">{{ health.httpStatus }}</td>
        </tr>
        <tr v-if="health.latencyMs">
          <td class="text-muted-foreground w-1/3 py-2 pr-4 text-xs">Latenz</td>
          <td class="py-2">{{ health.latencyMs }} ms</td>
        </tr>
        <tr v-if="health.url">
          <td class="text-muted-foreground w-1/3 py-2 pr-4 text-xs">Health URL</td>
          <td class="py-2">{{ health.url }}</td>
        </tr>
        <tr v-if="health.checkedAt">
          <td class="text-muted-foreground w-1/3 py-2 pr-4 text-xs">Zuletzt geprüft</td>
          <td class="py-2">{{ formatDateTime(health.checkedAt) }}</td>
        </tr>
      </tbody>
    </table>

    <div
      v-if="health?.failureReason"
      class="mt-4 rounded-md border border-red-200 bg-red-50 px-4 py-3 dark:border-red-900 dark:bg-red-950"
    >
      <p class="text-sm font-medium text-red-700 dark:text-red-400">
        {{ health.failureReason }}
      </p>
      <p v-if="health.message" class="text-muted-foreground mt-1 text-sm">
        {{ health.message }}
      </p>
    </div>

    <p v-if="!health" class="text-muted-foreground text-sm">Noch keine Health-Daten verfügbar.</p>
  </div>
</template>
