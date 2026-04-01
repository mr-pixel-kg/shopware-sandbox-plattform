<script setup lang="ts">
import { computed } from 'vue'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { formatDateTime, formatRelativeTime } from '@/utils/formatters'

import type { AuditLog } from '@/types'

const props = defineProps<{
  open: boolean
  log: AuditLog | null
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
}>()

const detailJson = computed(() => {
  if (!props.log) return ''
  return JSON.stringify(props.log.details, null, 2)
})

function actionBadgeConfig(action: string): { label: string; class: string } {
  const map: Record<string, { label: string; class: string }> = {
    'auth.logged_in': {
      label: 'Angemeldet',
      class: 'bg-blue-500/15 text-blue-700 border-blue-500/25',
    },
    'auth.logged_out': {
      label: 'Abgemeldet',
      class: 'bg-slate-500/15 text-slate-700 border-slate-500/25',
    },
    'user.registered': {
      label: 'Registriert',
      class: 'bg-blue-500/15 text-blue-700 border-blue-500/25',
    },
    'user.created': {
      label: 'Benutzer erstellt',
      class: 'bg-green-500/15 text-green-700 border-green-500/25',
    },
    'user.updated': {
      label: 'Benutzer geändert',
      class: 'bg-yellow-500/15 text-yellow-700 border-yellow-500/25',
    },
    'user.deleted': {
      label: 'Benutzer gelöscht',
      class: 'bg-red-500/15 text-red-700 border-red-500/25',
    },
    'user.whitelisted': {
      label: 'Whitelist hinzugefügt',
      class: 'bg-purple-500/15 text-purple-700 border-purple-500/25',
    },
    'user.whitelist_removed': {
      label: 'Whitelist entfernt',
      class: 'bg-slate-500/15 text-slate-700 border-slate-500/25',
    },
    'image.created': {
      label: 'Image erstellt',
      class: 'bg-green-500/15 text-green-700 border-green-500/25',
    },
    'image.updated': {
      label: 'Image geändert',
      class: 'bg-yellow-500/15 text-yellow-700 border-yellow-500/25',
    },
    'image.deleted': {
      label: 'Image gelöscht',
      class: 'bg-red-500/15 text-red-700 border-red-500/25',
    },
    'image.thumbnail_uploaded': {
      label: 'Thumbnail hochgeladen',
      class: 'bg-blue-500/15 text-blue-700 border-blue-500/25',
    },
    'image.thumbnail_deleted': {
      label: 'Thumbnail gelöscht',
      class: 'bg-slate-500/15 text-slate-700 border-slate-500/25',
    },
    'image.snapshot_created': {
      label: 'Snapshot erstellt',
      class: 'bg-purple-500/15 text-purple-700 border-purple-500/25',
    },
    'sandbox.created': {
      label: 'Sandbox erstellt',
      class: 'bg-green-500/15 text-green-700 border-green-500/25',
    },
    'sandbox.updated': {
      label: 'Sandbox geändert',
      class: 'bg-yellow-500/15 text-yellow-700 border-yellow-500/25',
    },
    'sandbox.ttl_updated': {
      label: 'TTL geändert',
      class: 'bg-yellow-500/15 text-yellow-700 border-yellow-500/25',
    },
    'sandbox.deleted': {
      label: 'Sandbox gelöscht',
      class: 'bg-red-500/15 text-red-700 border-red-500/25',
    },
  }
  return map[action] ?? { label: action, class: '' }
}
</script>

<template>
  <Sheet :open="open" @update:open="emit('update:open', $event)">
    <SheetContent side="right" class="overflow-y-auto sm:max-w-xl">
      <SheetHeader>
        <SheetTitle>Audit-Log-Details</SheetTitle>
        <SheetDescription v-if="log">
          {{ formatDateTime(log.timestamp) }} · {{ formatRelativeTime(log.timestamp) }}
        </SheetDescription>
      </SheetHeader>

      <div v-if="log" class="grid gap-6 px-4 pb-6">
        <section class="grid gap-3">
          <div class="flex flex-wrap items-center gap-2">
            <Badge variant="outline" :class="actionBadgeConfig(log.action).class">
              {{ actionBadgeConfig(log.action).label }}
            </Badge>
            <span class="text-muted-foreground font-mono text-xs">{{ log.action }}</span>
          </div>
          <div class="grid gap-1 rounded-lg border bg-slate-50/70 p-3">
            <span class="text-muted-foreground text-xs tracking-[0.18em] uppercase">Eintrag</span>
            <span class="font-mono text-xs break-all">{{ log.id }}</span>
          </div>
        </section>

        <section class="grid gap-3">
          <h3 class="text-sm font-semibold">Benutzer</h3>
          <div class="grid gap-3 rounded-lg border p-3">
            <div class="grid gap-1">
              <span class="text-muted-foreground text-xs tracking-[0.18em] uppercase">E-Mail</span>
              <span>{{ log.user?.email ?? 'System / unbekannt' }}</span>
            </div>
            <div class="grid gap-1">
              <span class="text-muted-foreground text-xs tracking-[0.18em] uppercase"
                >Benutzer-ID</span
              >
              <span class="font-mono text-xs break-all">{{ log.user?.id ?? '—' }}</span>
            </div>
          </div>
        </section>

        <section class="grid gap-3">
          <h3 class="text-sm font-semibold">Ressource</h3>
          <div class="grid gap-3 rounded-lg border p-3">
            <div class="grid gap-1">
              <span class="text-muted-foreground text-xs tracking-[0.18em] uppercase">Typ</span>
              <span>{{ log.resourceType ?? '—' }}</span>
            </div>
            <div class="grid gap-1">
              <span class="text-muted-foreground text-xs tracking-[0.18em] uppercase"
                >Ressourcen-ID</span
              >
              <span class="font-mono text-xs break-all">{{ log.resourceId ?? '—' }}</span>
            </div>
          </div>
        </section>

        <section class="grid gap-3">
          <h3 class="text-sm font-semibold">Client-Kontext</h3>
          <div class="grid gap-3 rounded-lg border p-3">
            <div class="grid gap-1">
              <span class="text-muted-foreground text-xs tracking-[0.18em] uppercase"
                >IP-Adresse</span
              >
              <span class="font-mono text-xs break-all">{{ log.ipAddress ?? '—' }}</span>
            </div>
            <div class="grid gap-1">
              <span class="text-muted-foreground text-xs tracking-[0.18em] uppercase"
                >Client-Token</span
              >
              <span class="font-mono text-xs break-all">{{ log.clientToken ?? '—' }}</span>
            </div>
            <div class="grid gap-1">
              <span class="text-muted-foreground text-xs tracking-[0.18em] uppercase"
                >User-Agent</span
              >
              <span class="font-mono text-xs break-all">{{ log.userAgent ?? '—' }}</span>
            </div>
          </div>
        </section>

        <section class="grid gap-3">
          <div class="flex items-center justify-between gap-3">
            <h3 class="text-sm font-semibold">Details</h3>
            <span class="text-muted-foreground text-xs">Formatiertes JSON</span>
          </div>
          <pre
            class="bg-muted overflow-x-auto rounded-lg border p-3 font-mono text-xs leading-5 break-words whitespace-pre-wrap"
            >{{ detailJson }}</pre
          >
        </section>

        <div class="flex justify-end">
          <Button type="button" variant="outline" @click="emit('update:open', false)"
            >Schließen</Button
          >
        </div>
      </div>
    </SheetContent>
  </Sheet>
</template>
