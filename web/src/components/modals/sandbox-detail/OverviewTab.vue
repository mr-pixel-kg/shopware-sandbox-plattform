<script setup lang="ts">
import TtlChip from '@/components/sandboxes/TtlChip.vue'
import CopyButton from '@/components/shared/CopyButton.vue'
import { Badge } from '@/components/ui/badge'
import { formatDateTime } from '@/utils/formatters'

import type { Image, Sandbox } from '@/types'

defineProps<{
  sandbox: Sandbox
  image?: Image
  isActive: boolean
}>()
</script>

<template>
  <table class="w-full text-sm">
    <tbody class="divide-y">
      <tr>
        <td class="text-muted-foreground w-1/3 py-2 pr-4 text-xs">ID</td>
        <td class="py-2">
          <div class="flex items-center gap-1">
            <span class="min-w-0 truncate">{{ sandbox.id }}</span>
            <CopyButton :value="sandbox.id" />
          </div>
        </td>
      </tr>

      <tr v-if="sandbox.url">
        <td class="text-muted-foreground w-1/3 py-2 pr-4 text-xs">URL</td>
        <td class="py-2">
          <div class="flex items-center gap-1">
            <a
              :href="sandbox.url"
              target="_blank"
              class="min-w-0 truncate text-blue-600 hover:underline dark:text-blue-400"
              @click.stop
            >
              {{ sandbox.url }}
            </a>
            <CopyButton :value="sandbox.url" />
          </div>
        </td>
      </tr>

      <tr v-if="sandbox.displayName">
        <td class="text-muted-foreground w-1/3 py-2 pr-4 text-xs">Name</td>
        <td class="py-2">{{ sandbox.displayName }}</td>
      </tr>

      <tr>
        <td class="text-muted-foreground w-1/3 py-2 pr-4 text-xs">Vorlage</td>
        <td class="py-2">
          {{ image?.title || image?.name || '—' }}
          <Badge v-if="image?.tag" variant="secondary" class="ml-2 text-xs">
            {{ image.tag }}
          </Badge>
        </td>
      </tr>

      <tr>
        <td class="text-muted-foreground w-1/3 py-2 pr-4 text-xs">Besitzer</td>
        <td class="py-2">{{ sandbox.owner?.email ?? 'Gast' }}</td>
      </tr>

      <tr v-if="sandbox.clientId">
        <td class="text-muted-foreground w-1/3 py-2 pr-4 text-xs">Client ID</td>
        <td class="py-2 font-mono text-xs">{{ sandbox.clientId }}</td>
      </tr>

      <tr v-if="sandbox.port">
        <td class="text-muted-foreground w-1/3 py-2 pr-4 text-xs">Port</td>
        <td class="py-2">{{ sandbox.port }}</td>
      </tr>

      <tr>
        <td class="text-muted-foreground w-1/3 py-2 pr-4 text-xs">Erstellt</td>
        <td class="py-2">{{ formatDateTime(sandbox.createdAt) }}</td>
      </tr>

      <tr>
        <td class="text-muted-foreground w-1/3 py-2 pr-4 text-xs">Läuft ab</td>
        <td class="py-2">
          {{ sandbox.expiresAt ? formatDateTime(sandbox.expiresAt) : 'Unbegrenzt' }}
        </td>
      </tr>

      <tr v-if="isActive">
        <td class="text-muted-foreground w-1/3 py-2 pr-4 text-xs">Verbleibend</td>
        <td class="py-2">
          <TtlChip :expires-at="sandbox.expiresAt" :created-at="sandbox.createdAt" />
        </td>
      </tr>

      <tr v-if="sandbox.lastSeenAt">
        <td class="text-muted-foreground w-1/3 py-2 pr-4 text-xs">Zuletzt gesehen</td>
        <td class="py-2">{{ formatDateTime(sandbox.lastSeenAt) }}</td>
      </tr>

      <tr v-if="sandbox.containerId">
        <td class="text-muted-foreground w-1/3 py-2 pr-4 text-xs">Container</td>
        <td class="py-2">
          <div class="flex items-center gap-1">
            <span class="min-w-0 truncate">{{ sandbox.containerId.slice(0, 12) }}</span>
            <CopyButton :value="sandbox.containerId" />
          </div>
        </td>
      </tr>
    </tbody>
  </table>
</template>
