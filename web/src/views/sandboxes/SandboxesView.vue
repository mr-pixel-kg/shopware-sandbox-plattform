<script setup lang="ts">
import { Clock, ExternalLink, Plus, Square, Trash2 } from 'lucide-vue-next'
import { computed, ref } from 'vue'
import { toast } from 'vue-sonner'

import ConfirmDialog from '@/components/modals/ConfirmDialog.vue'
import ExtendTtlDialog from '@/components/modals/ExtendTtlDialog.vue'
import NewSandboxDialog from '@/components/modals/NewSandboxDialog.vue'
import TtlChip from '@/components/sandboxes/TtlChip.vue'
import PageHeader from '@/components/shared/PageHeader.vue'
import StatusBadge from '@/components/shared/StatusBadge.vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
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

import type { Sandbox } from '@/types'

const {
  activeSandboxes,
  recentSandboxes,
  loading,
  createSandbox,
  deleteSandbox,
  refresh,
} = useSandboxes()
const { images } = useImages()

const showNewSandbox = ref(false)
const showExtend = ref(false)
const showConfirmDelete = ref(false)
const selectedSandbox = ref<Sandbox | null>(null)

const hasActive = computed(() => activeSandboxes.value.length > 0)
const hasRecent = computed(() => recentSandboxes.value.length > 0)

const isSelectedActive = computed(() => {
  const s = selectedSandbox.value?.status
  return s === 'running' || s === 'starting'
})

function getImageName(imageId: string): string {
  const image = images.value.find((i) => i.id === imageId)
  return image?.title || image?.name || '—'
}

function getImageTag(imageId: string): string | undefined {
  return images.value.find((i) => i.id === imageId)?.tag
}

function handleOpen(sandbox: Sandbox) {
  if (sandbox.url) window.open(sandbox.url, '_blank')
}

function handleExtend(sandbox: Sandbox) {
  selectedSandbox.value = sandbox
  showExtend.value = true
}

function handleDelete(sandbox: Sandbox) {
  selectedSandbox.value = sandbox
  showConfirmDelete.value = true
}

async function handleCreateSandbox(
  payload: { imageId: string; ttlMinutes: number },
  done: (success: boolean) => void,
) {
  try {
    await createSandbox(payload)
    toast.success('Sandbox wird gestartet')
    refresh()
    done(true)
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Fehler beim Starten der Sandbox'))
    done(false)
  }
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
    <PageHeader title="Sandboxes" subtitle="Deine aktiven und kürzlich beendeten Sandboxes.">
      <template #actions>
        <Button @click="showNewSandbox = true">
          <Plus class="mr-1 h-4 w-4" />
          Neue Sandbox
        </Button>
      </template>
    </PageHeader>

    <div class="space-y-8">
      <section>
        <h3 class="text-muted-foreground mb-3 text-sm font-medium">Aktive Sandboxes</h3>
        <div class="rounded-md border">
          <Table class="table-fixed">
            <TableHeader>
              <TableRow>
                <TableHead class="w-[15%]">Status</TableHead>
                <TableHead class="w-[35%]">Vorlage</TableHead>
                <TableHead class="w-[25%]">Verbleibend</TableHead>
                <TableHead class="w-[25%] text-right">Aktionen</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <template v-if="loading">
                <TableRow v-for="i in 2" :key="i" class="h-13">
                  <TableCell><Skeleton class="h-5 w-14 rounded-full" /></TableCell>
                  <TableCell><Skeleton class="h-4 w-28" /></TableCell>
                  <TableCell><Skeleton class="h-4 w-20" /></TableCell>
                  <TableCell class="text-right">
                    <div class="flex items-center justify-end gap-1">
                      <Skeleton class="h-7 w-7" />
                      <Skeleton class="h-7 w-7" />
                      <Skeleton class="h-7 w-7" />
                    </div>
                  </TableCell>
                </TableRow>
              </template>
              <TableEmpty v-else-if="!hasActive" :colspan="4"> Keine aktiven Sandboxes </TableEmpty>
              <TableRow v-for="sandbox in activeSandboxes" :key="sandbox.id" class="h-13">
                <TableCell>
                  <StatusBadge :status="sandbox.status" />
                </TableCell>
                <TableCell>
                  <div class="flex items-center gap-2">
                    <span class="truncate text-sm font-medium">{{
                      getImageName(sandbox.imageId)
                    }}</span>
                    <Badge v-if="getImageTag(sandbox.imageId)" variant="secondary" class="text-xs">
                      {{ getImageTag(sandbox.imageId) }}
                    </Badge>
                  </div>
                </TableCell>
                <TableCell>
                  <TtlChip :expires-at="sandbox.expiresAt" :created-at="sandbox.createdAt" />
                </TableCell>
                <TableCell class="text-right">
                  <TooltipProvider>
                    <div class="flex items-center justify-end gap-1">
                      <Tooltip>
                        <TooltipTrigger as-child>
                          <Button variant="ghost" size="icon-sm" @click="handleOpen(sandbox)">
                            <ExternalLink class="h-4 w-4" />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>Öffnen</TooltipContent>
                      </Tooltip>
                      <Tooltip>
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
      </section>

      <section v-if="hasRecent || loading">
        <h3 class="text-muted-foreground mb-3 text-sm font-medium">Zuletzt beendet</h3>
        <div class="rounded-md border">
          <Table class="table-fixed">
            <TableHeader>
              <TableRow>
                <TableHead class="w-[15%]">Status</TableHead>
                <TableHead class="w-[40%]">Vorlage</TableHead>
                <TableHead class="w-[25%]">Beendet</TableHead>
                <TableHead class="w-[20%] text-right">Aktionen</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <template v-if="loading">
                <TableRow v-for="i in 2" :key="i" class="h-13">
                  <TableCell><Skeleton class="h-5 w-16 rounded-full" /></TableCell>
                  <TableCell><Skeleton class="h-4 w-28" /></TableCell>
                  <TableCell><Skeleton class="h-4 w-24" /></TableCell>
                  <TableCell class="text-right"><Skeleton class="ml-auto h-7 w-7" /></TableCell>
                </TableRow>
              </template>
              <TableRow v-for="sandbox in recentSandboxes" :key="sandbox.id" class="h-13">
                <TableCell>
                  <StatusBadge :status="sandbox.status" />
                </TableCell>
                <TableCell>
                  <div class="flex items-center gap-2">
                    <span class="truncate text-sm font-medium">{{
                      getImageName(sandbox.imageId)
                    }}</span>
                    <Badge v-if="getImageTag(sandbox.imageId)" variant="secondary" class="text-xs">
                      {{ getImageTag(sandbox.imageId) }}
                    </Badge>
                  </div>
                </TableCell>
                <TableCell class="text-muted-foreground">
                  {{ formatDateTime(sandbox.updatedAt) }}
                </TableCell>
                <TableCell class="text-right">
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger as-child>
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          class="text-destructive hover:text-destructive"
                          @click="handleDelete(sandbox)"
                        >
                          <Trash2 class="h-4 w-4" />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>Löschen</TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
      </section>
    </div>

    <NewSandboxDialog
      v-model:open="showNewSandbox"
      :images="images"
      @submit="handleCreateSandbox"
    />

    <ExtendTtlDialog
      v-model:open="showExtend"
      :sandbox-id="selectedSandbox?.id ?? ''"
      :sandbox-name="selectedSandbox?.containerName ?? ''"
    />

    <ConfirmDialog
      v-model:open="showConfirmDelete"
      :title="isSelectedActive ? 'Sandbox beenden' : 'Aus Verlauf entfernen'"
      :description="
        isSelectedActive
          ? `Bist du sicher, dass du ${selectedSandbox?.containerName ?? 'diese Sandbox'} beenden möchtest? Diese Aktion kann nicht rückgängig gemacht werden.`
          : `Bist du sicher, dass du ${selectedSandbox?.containerName ?? 'diese Sandbox'} endgültig aus dem Verlauf entfernen möchtest?`
      "
      :confirm-label="isSelectedActive ? 'Beenden' : 'Entfernen'"
      @confirm="handleConfirmDelete"
    />
  </div>
</template>
