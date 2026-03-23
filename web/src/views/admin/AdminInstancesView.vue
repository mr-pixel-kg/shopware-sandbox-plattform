<script setup lang="ts">
import { Clock, Square } from 'lucide-vue-next'
import { computed, ref } from 'vue'
import { toast } from 'vue-sonner'

import ConfirmDialog from '@/components/modals/ConfirmDialog.vue'
import ExtendTtlDialog from '@/components/modals/ExtendTtlDialog.vue'
import PageHeader from '@/components/shared/PageHeader.vue'
import StatusBadge from '@/components/shared/StatusBadge.vue'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableEmpty,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { useImages } from '@/composables/useImages'
import { useSandboxes } from '@/composables/useSandboxes'
import { getApiErrorMessage } from '@/utils/error'
import { formatDateTime } from '@/utils/formatters'

import type { Sandbox, SandboxStatus } from '@/types'

const { sandboxes, deleteSandbox, loading } = useSandboxes('all')
const { images } = useImages('all')

const statusFilter = ref<string>('all')

const filteredSandboxes = computed(() => {
  if (statusFilter.value === 'all') return sandboxes.value
  const activeStatuses: SandboxStatus[] = ['running', 'starting']
  const inactiveStatuses: SandboxStatus[] = ['stopped', 'expired', 'deleted', 'failed']
  if (statusFilter.value === 'active')
    return sandboxes.value.filter((s) => activeStatuses.includes(s.status))
  if (statusFilter.value === 'inactive')
    return sandboxes.value.filter((s) => inactiveStatuses.includes(s.status))
  return sandboxes.value
})

const showExtend = ref(false)
const showConfirmDelete = ref(false)
const selectedSandbox = ref<Sandbox | null>(null)

function getImageName(imageId: string): string {
  const image = images.value.find((i) => i.id === imageId)
  return image?.title || image?.name || '—'
}

function handleExtend(sandbox: Sandbox) {
  selectedSandbox.value = sandbox
  showExtend.value = true
}

function handleDelete(sandbox: Sandbox) {
  selectedSandbox.value = sandbox
  showConfirmDelete.value = true
}

async function handleConfirmDelete() {
  if (!selectedSandbox.value) return
  try {
    await deleteSandbox(selectedSandbox.value.id)
    toast.success('Sandbox wurde beendet')
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Fehler beim Beenden'))
  }
}
</script>

<template>
  <div>
    <PageHeader title="Instanzen" subtitle="Alle Sandbox-Instanzen verwalten." />

    <div class="mb-4 flex items-center gap-3">
      <Select v-model="statusFilter">
        <SelectTrigger class="w-[160px]">
          <SelectValue placeholder="Alle Status" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">Alle Status</SelectItem>
          <SelectItem value="active">Aktiv</SelectItem>
          <SelectItem value="inactive">Abgelaufen</SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div class="rounded-md border">
      <Table class="table-fixed">
        <TableHeader>
          <TableRow>
            <TableHead class="w-[15%]">Status</TableHead>
            <TableHead class="w-[30%]">Vorlage</TableHead>
            <TableHead class="w-[20%]">Gestartet</TableHead>
            <TableHead class="w-[20%]">Läuft ab</TableHead>
            <TableHead class="w-[15%] text-right">Aktionen</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <template v-if="loading">
            <TableRow v-for="i in 3" :key="i" class="h-13">
              <TableCell><Skeleton class="h-5 w-16 rounded-full" /></TableCell>
              <TableCell><Skeleton class="h-4 w-28" /></TableCell>
              <TableCell><Skeleton class="h-4 w-24" /></TableCell>
              <TableCell><Skeleton class="h-4 w-24" /></TableCell>
              <TableCell class="text-right">
                <div class="flex items-center justify-end gap-1">
                  <Skeleton class="h-7 w-7" />
                  <Skeleton class="h-7 w-7" />
                </div>
              </TableCell>
            </TableRow>
          </template>
          <TableEmpty v-else-if="filteredSandboxes.length === 0" :colspan="5">
            Keine Instanzen gefunden
          </TableEmpty>
          <TableRow v-for="sandbox in filteredSandboxes" :key="sandbox.id" class="h-13">
            <TableCell>
              <StatusBadge :status="sandbox.status" />
            </TableCell>
            <TableCell class="font-medium">{{ getImageName(sandbox.imageId) }}</TableCell>
            <TableCell class="text-muted-foreground">{{
              formatDateTime(sandbox.createdAt)
            }}</TableCell>
            <TableCell class="text-muted-foreground">
              {{ sandbox.expiresAt ? formatDateTime(sandbox.expiresAt) : '—' }}
            </TableCell>
            <TableCell class="text-right">
              <TooltipProvider>
                <div class="flex items-center justify-end gap-1">
                  <Tooltip v-if="sandbox.status === 'running' || sandbox.status === 'starting'">
                    <TooltipTrigger as-child>
                      <Button variant="ghost" size="icon-sm" @click="handleExtend(sandbox)">
                        <Clock class="h-4 w-4" />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>Verlängern</TooltipContent>
                  </Tooltip>
                  <Tooltip>
                    <TooltipTrigger as-child>
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        class="text-destructive hover:text-destructive"
                        @click="handleDelete(sandbox)"
                      >
                        <Square class="h-4 w-4" />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>Beenden</TooltipContent>
                  </Tooltip>
                </div>
              </TooltipProvider>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>

    <ExtendTtlDialog
      v-model:open="showExtend"
      :sandbox-id="selectedSandbox?.id ?? ''"
      :sandbox-name="selectedSandbox?.containerName ?? ''"
    />

    <ConfirmDialog
      v-model:open="showConfirmDelete"
      title="Sandbox beenden"
      :description="`Bist du sicher, dass du ${selectedSandbox?.containerName ?? 'diese Sandbox'} beenden möchtest? Diese Aktion kann nicht rückgängig gemacht werden.`"
      confirm-label="Beenden"
      @confirm="handleConfirmDelete"
    />
  </div>
</template>
