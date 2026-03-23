<script setup lang="ts">
import { useAuditLogs } from '@/composables/useAuditLogs'
import { formatDateTime } from '@/utils/formatters'
import PageHeader from '@/components/shared/PageHeader.vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableEmpty,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Download, ChevronLeft, ChevronRight } from 'lucide-vue-next'
import { Skeleton } from '@/components/ui/skeleton'

const {
  logs,
  allLogs,
  loading,
  page,
  totalPages,
  userFilter,
  actionFilter,
  periodFilter,
  uniqueUsers,
  uniqueActions,
  exportCsv,
} = useAuditLogs()

function actionBadgeConfig(action: string): { label: string; class: string } {
  const map: Record<string, { label: string; class: string }> = {
    boot: { label: 'Gestartet', class: 'bg-green-500/15 text-green-700 border-green-500/25' },
    create: { label: 'Erstellt', class: 'bg-green-500/15 text-green-700 border-green-500/25' },
    extend: { label: 'Verlängert', class: 'bg-yellow-500/15 text-yellow-700 border-yellow-500/25' },
    stop: { label: 'Gestoppt', class: 'bg-red-500/15 text-red-700 border-red-500/25' },
    delete: { label: 'Gelöscht', class: 'bg-red-500/15 text-red-700 border-red-500/25' },
    login: { label: 'Angemeldet', class: 'bg-blue-500/15 text-blue-700 border-blue-500/25' },
    invite: { label: 'Eingeladen', class: 'bg-purple-500/15 text-purple-700 border-purple-500/25' },
  }
  return map[action] ?? { label: action, class: '' }
}

function formatDetails(details: Record<string, unknown> | unknown[]): string {
  if (Array.isArray(details)) return details.join(', ')
  return Object.entries(details)
    .map(([k, v]) => `${k}: ${v}`)
    .join(', ')
}
</script>

<template>
  <div>
    <PageHeader title="Protokoll" subtitle="Aktivitäten und Audit-Einträge einsehen.">
      <template #actions>
        <Button variant="outline" @click="exportCsv">
          <Download class="mr-1 h-4 w-4" />
          Exportieren
        </Button>
      </template>
    </PageHeader>

    <div class="mb-4 flex items-center gap-3">
      <Select v-model="userFilter">
        <SelectTrigger class="w-[160px]">
          <SelectValue placeholder="Alle Benutzer" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">Alle Benutzer</SelectItem>
          <SelectItem v-for="u in uniqueUsers" :key="u" :value="u">{{ u }}</SelectItem>
        </SelectContent>
      </Select>

      <Select v-model="actionFilter">
        <SelectTrigger class="w-[160px]">
          <SelectValue placeholder="Alle Aktionen" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">Alle Aktionen</SelectItem>
          <SelectItem v-for="a in uniqueActions" :key="a" :value="a">{{ a }}</SelectItem>
        </SelectContent>
      </Select>

      <Select v-model="periodFilter">
        <SelectTrigger class="w-[130px]">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="24h">24 Stunden</SelectItem>
          <SelectItem value="7d">7 Tage</SelectItem>
          <SelectItem value="30d">30 Tage</SelectItem>
        </SelectContent>
      </Select>

      <span class="text-muted-foreground ml-auto text-sm"> {{ allLogs.length }} Einträge </span>
    </div>

    <div class="rounded-md border">
      <Table class="table-fixed">
        <TableHeader>
          <TableRow>
            <TableHead class="w-[20%]">Zeitpunkt</TableHead>
            <TableHead class="w-[20%]">Benutzer</TableHead>
            <TableHead class="w-[15%]">Aktion</TableHead>
            <TableHead class="w-[30%]">Details</TableHead>
            <TableHead class="w-[15%]">IP</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <template v-if="loading">
            <TableRow v-for="i in 3" :key="i" class="h-13">
              <TableCell><Skeleton class="h-4 w-28" /></TableCell>
              <TableCell><Skeleton class="h-4 w-20" /></TableCell>
              <TableCell><Skeleton class="h-5 w-16 rounded-full" /></TableCell>
              <TableCell><Skeleton class="h-4 w-36" /></TableCell>
              <TableCell><Skeleton class="h-4 w-20" /></TableCell>
            </TableRow>
          </template>
          <TableEmpty v-else-if="logs.length === 0" :colspan="5">
            Keine Einträge gefunden
          </TableEmpty>
          <TableRow v-for="log in logs" :key="log.id" class="h-13">
            <TableCell class="text-muted-foreground whitespace-nowrap">
              {{ formatDateTime(log.createdAt) }}
            </TableCell>
            <TableCell>{{ log.userId ?? '—' }}</TableCell>
            <TableCell>
              <Badge variant="outline" :class="actionBadgeConfig(log.action).class">
                {{ actionBadgeConfig(log.action).label }}
              </Badge>
            </TableCell>
            <TableCell class="text-muted-foreground max-w-[300px] truncate">
              {{ formatDetails(log.details) }}
            </TableCell>
            <TableCell class="text-muted-foreground font-mono text-xs">
              {{ log.ipAddress }}
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>

    <div v-if="totalPages > 1" class="mt-4 flex items-center justify-between">
      <span class="text-muted-foreground text-sm"> Seite {{ page }} von {{ totalPages }} </span>
      <div class="flex items-center gap-2">
        <Button variant="outline" size="sm" :disabled="page <= 1" @click="page--">
          <ChevronLeft class="h-4 w-4" />
        </Button>
        <Button variant="outline" size="sm" :disabled="page >= totalPages" @click="page++">
          <ChevronRight class="h-4 w-4" />
        </Button>
      </div>
    </div>
  </div>
</template>
