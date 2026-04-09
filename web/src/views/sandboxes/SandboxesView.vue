<script setup lang="ts">
import {
  Camera,
  Clock,
  ExternalLink,
  MoreHorizontal,
  Pencil,
  Plus,
  Square,
  Trash2,
} from 'lucide-vue-next'
import { storeToRefs } from 'pinia'
import { computed, ref } from 'vue'
import { toast } from 'vue-sonner'

import ConfirmDialog from '@/components/modals/ConfirmDialog.vue'
import EditSandboxDialog from '@/components/modals/EditSandboxDialog.vue'
import ExtendTtlDialog from '@/components/modals/ExtendTtlDialog.vue'
import NewSandboxDialog from '@/components/modals/NewSandboxDialog.vue'
import { SandboxDetailDialog } from '@/components/modals/sandbox-detail'
import SnapshotDialog from '@/components/modals/SnapshotDialog.vue'
import TtlChip from '@/components/sandboxes/TtlChip.vue'
import DataTablePagination from '@/components/shared/DataTablePagination.vue'
import PageHeader from '@/components/shared/PageHeader.vue'
import StatusBadge from '@/components/shared/StatusBadge.vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
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
import { useImages } from '@/composables/useImages'
import { usePagination } from '@/composables/usePagination'
import { useSandboxes } from '@/composables/useSandboxes'
import { useAuthStore } from '@/stores/auth.store'
import { getApiErrorMessage } from '@/utils/error'
import { formatDateTime } from '@/utils/formatters'

import type { MetadataItem, Sandbox, SandboxStatus } from '@/types'

const authStore = useAuthStore()
const { isAdmin } = storeToRefs(authStore)

const {
  sandboxes,
  activeSandboxes,
  healthBySandboxId,
  recentSandboxes,
  allSandboxes,
  loading,
  busyIds,
  createSandbox,
  deleteSandbox,
  updateSandbox,
  snapshotSandbox,
  fetchAdminSandboxes,
} = useSandboxes()
const { images, uploadThumbnail, trackPendingImage } = useImages('all')

if (isAdmin.value) void fetchAdminSandboxes()

const adminStatusFilter = ref<string>('all')

const filteredAllSandboxes = computed(() => {
  const all = allSandboxes.value
  if (adminStatusFilter.value === 'all') return all
  const activeStatuses: SandboxStatus[] = ['running', 'starting', 'paused', 'stopping']
  const inactiveStatuses: SandboxStatus[] = ['stopped', 'expired', 'deleted', 'failed']
  if (adminStatusFilter.value === 'active')
    return all.filter((s) => activeStatuses.includes(s.status))
  if (adminStatusFilter.value === 'inactive')
    return all.filter((s) => inactiveStatuses.includes(s.status))
  return all
})

const PAGE_SIZE = 5

const {
  page: activePage,
  pageSize: activePageSize,
  paginatedItems: paginatedActiveSandboxes,
} = usePagination({ pageSize: PAGE_SIZE, source: activeSandboxes })

const {
  page: recentPage,
  pageSize: recentPageSize,
  paginatedItems: paginatedRecentSandboxes,
} = usePagination({ pageSize: PAGE_SIZE, source: recentSandboxes })

const {
  page: allPage,
  pageSize: allPageSize,
  paginatedItems: paginatedAllSandboxes,
} = usePagination({
  pageSize: PAGE_SIZE,
  source: filteredAllSandboxes,
  watchResetSources: [adminStatusFilter],
})

const showNewSandbox = ref(false)
const showDetail = ref(false)
const showEdit = ref(false)
const showExtend = ref(false)
const showSnapshot = ref(false)
const showConfirmDelete = ref(false)
const selectedSandboxId = ref<string | null>(null)

const selectedSandbox = computed(() => {
  if (!selectedSandboxId.value) return null
  return (
    sandboxes.value.find((s) => s.id === selectedSandboxId.value) ??
    allSandboxes.value.find((s) => s.id === selectedSandboxId.value) ??
    null
  )
})

const hasActive = computed(() => activeSandboxes.value.length > 0)
const hasRecent = computed(() => recentSandboxes.value.length > 0)

const isSelectedActive = computed(() => {
  const s = selectedSandbox.value?.status
  return s !== undefined && ['running', 'starting', 'paused', 'stopping'].includes(s)
})

function getImageName(imageId: string): string {
  const image = images.value.find((i) => i.id === imageId)
  return image?.title || image?.name || '—'
}

function getSandboxDisplayName(sandbox: Sandbox): string {
  return sandbox.displayName || getImageName(sandbox.imageId)
}

function getImageTag(imageId: string): string | undefined {
  return images.value.find((i) => i.id === imageId)?.tag
}

function getSandboxOwnerLabel(sandbox: Sandbox): string {
  if (sandbox.owner?.email) return sandbox.owner.email
  if (sandbox.clientId) return `Gast (${sandbox.clientId.slice(0, 8)}…)`
  return 'Gast'
}

function getLiveHealth(sandbox: Sandbox) {
  return healthBySandboxId.value[sandbox.id]
}

function isSandboxReadyForOpen(sandbox: Sandbox): boolean {
  const health = getLiveHealth(sandbox)
  if (sandbox.status !== 'running') return false
  if (!health) return true
  return health.ready
}

const selectedImage = computed(() => {
  if (!selectedSandbox.value) return undefined
  return images.value.find((i) => i.id === selectedSandbox.value!.imageId)
})

const selectedHealth = computed(() => {
  if (!selectedSandbox.value) return undefined
  return healthBySandboxId.value[selectedSandbox.value.id]
})

function handleRowClick(sandbox: Sandbox) {
  selectedSandboxId.value = sandbox.id
  showDetail.value = true
}

function handleOpen(sandbox: Sandbox) {
  if (!isSandboxReadyForOpen(sandbox)) return
  if (sandbox.url) window.open(sandbox.url, '_blank')
}

function handleEdit(sandbox: Sandbox) {
  selectedSandboxId.value = sandbox.id
  showEdit.value = true
}

function handleExtend(sandbox: Sandbox) {
  selectedSandboxId.value = sandbox.id
  showExtend.value = true
}

function handleSnapshot(sandbox: Sandbox) {
  selectedSandboxId.value = sandbox.id
  showSnapshot.value = true
}

function handleDelete(sandbox: Sandbox) {
  selectedSandboxId.value = sandbox.id
  showConfirmDelete.value = true
}

async function handleEditSandbox(
  payload: { id: string; displayName: string },
  done: (success: boolean) => void,
) {
  try {
    await updateSandbox(payload.id, { displayName: payload.displayName })
    toast.success('Sandbox wurde aktualisiert')
    done(true)
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Fehler beim Speichern'))
    done(false)
  }
}

async function handleExtendTtl(
  payload: { sandboxId: string; ttlMinutes: number },
  done: (success: boolean) => void,
) {
  try {
    await updateSandbox(payload.sandboxId, { ttlMinutes: payload.ttlMinutes })
    toast.success('Laufzeit wurde verlängert')
    done(true)
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Fehler beim Verlängern der Laufzeit'))
    done(false)
  }
}

async function handleCreateSandbox(
  payload: {
    imageId: string
    ttlMinutes: number
    displayName?: string
    metadata?: Record<string, string>
  },
  done: (success: boolean) => void,
) {
  try {
    await createSandbox(payload)
    done(true)
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Fehler beim Starten der Sandbox'))
    done(false)
  }
}

async function handleCreateSnapshot(
  payload: {
    name: string
    tag: string
    title: string
    description: string
    isPublic: boolean
    thumbnailFile?: File
    metadata?: MetadataItem[]
  },
  done: (success: boolean) => void,
) {
  if (!selectedSandbox.value) return
  try {
    const { thumbnailFile: _, ...snapshotPayload } = payload
    const image = await snapshotSandbox(selectedSandbox.value.id, snapshotPayload)
    trackPendingImage(image)

    if (payload.thumbnailFile) {
      void uploadThumbnail(image.id, payload.thumbnailFile)
    }

    done(true)
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Fehler beim Erstellen des Snapshots'))
    done(false)
  }
}

async function handleConfirmDelete(done: (success: boolean) => void) {
  if (!selectedSandbox.value) return done(false)
  const id = selectedSandbox.value.id
  busyIds.value.add(id)
  try {
    await deleteSandbox(id)
    toast.success('Sandbox wurde beendet')
    done(true)
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Fehler beim Beenden'))
    done(false)
  } finally {
    busyIds.value.delete(id)
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
                <TableHead class="w-[14%]">Status</TableHead>
                <TableHead class="w-[22%]">Name</TableHead>
                <TableHead class="w-[22%]">Vorlage</TableHead>
                <TableHead class="w-[30%]">Verbleibend</TableHead>
                <TableHead class="w-[12%] text-right">Aktionen</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <template v-if="loading">
                <TableRow v-for="i in 2" :key="i" class="h-13">
                  <TableCell><Skeleton class="h-5 w-14 rounded-full" /></TableCell>
                  <TableCell><Skeleton class="h-4 w-24" /></TableCell>
                  <TableCell><Skeleton class="h-4 w-28" /></TableCell>
                  <TableCell><Skeleton class="h-4 w-20" /></TableCell>
                  <TableCell class="text-right"><Skeleton class="ml-auto h-7 w-7" /></TableCell>
                </TableRow>
              </template>
              <TableEmpty v-else-if="!hasActive" :colspan="5">Keine aktiven Sandboxes</TableEmpty>
              <TableRow
                v-for="sandbox in paginatedActiveSandboxes"
                :key="sandbox.id"
                class="h-13 cursor-pointer"
                @click="handleRowClick(sandbox)"
              >
                <TableCell>
                  <StatusBadge :status="sandbox.status" :state-reason="sandbox.stateReason" />
                </TableCell>
                <TableCell>
                  <span class="text-muted-foreground truncate text-sm">{{
                    sandbox.displayName || '—'
                  }}</span>
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
                <TableCell class="text-right" @click.stop>
                  <DropdownMenu>
                    <DropdownMenuTrigger as-child>
                      <Button variant="ghost" size="icon-sm" :disabled="busyIds.has(sandbox.id)">
                        <MoreHorizontal class="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem
                        :disabled="sandbox.status !== 'running'"
                        @click="handleOpen(sandbox)"
                      >
                        <ExternalLink class="mr-2 h-4 w-4" />
                        Öffnen
                      </DropdownMenuItem>
                      <DropdownMenuItem @click="handleEdit(sandbox)">
                        <Pencil class="mr-2 h-4 w-4" />
                        Bearbeiten
                      </DropdownMenuItem>
                      <DropdownMenuItem @click="handleExtend(sandbox)">
                        <Clock class="mr-2 h-4 w-4" />
                        Verlängern
                      </DropdownMenuItem>
                      <DropdownMenuItem @click="handleSnapshot(sandbox)">
                        <Camera class="mr-2 h-4 w-4" />
                        Snapshot
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem class="text-destructive" @click="handleDelete(sandbox)">
                        <Square class="mr-2 h-4 w-4" />
                        Beenden
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
        <DataTablePagination
          :page="activePage"
          :total-items="activeSandboxes.length"
          :page-size="activePageSize"
          @update:page="activePage = $event"
        />
      </section>

      <section v-if="hasRecent || loading">
        <h3 class="text-muted-foreground mb-3 text-sm font-medium">Zuletzt beendet</h3>
        <div class="rounded-md border">
          <Table class="table-fixed">
            <TableHeader>
              <TableRow>
                <TableHead class="w-[14%]">Status</TableHead>
                <TableHead class="w-[22%]">Name</TableHead>
                <TableHead class="w-[26%]">Vorlage</TableHead>
                <TableHead class="w-[26%]">Beendet</TableHead>
                <TableHead class="w-[12%] text-right">Aktionen</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <template v-if="loading">
                <TableRow v-for="i in 2" :key="i" class="h-13">
                  <TableCell><Skeleton class="h-5 w-16 rounded-full" /></TableCell>
                  <TableCell><Skeleton class="h-4 w-24" /></TableCell>
                  <TableCell><Skeleton class="h-4 w-28" /></TableCell>
                  <TableCell><Skeleton class="h-4 w-24" /></TableCell>
                  <TableCell class="text-right"><Skeleton class="ml-auto h-7 w-7" /></TableCell>
                </TableRow>
              </template>
              <TableRow
                v-for="sandbox in paginatedRecentSandboxes"
                :key="sandbox.id"
                class="h-13 cursor-pointer"
                @click="handleRowClick(sandbox)"
              >
                <TableCell>
                  <StatusBadge :status="sandbox.status" :state-reason="sandbox.stateReason" />
                </TableCell>
                <TableCell>
                  <span class="text-muted-foreground truncate text-sm">{{
                    sandbox.displayName || '—'
                  }}</span>
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
                <TableCell class="text-right" @click.stop>
                  <DropdownMenu>
                    <DropdownMenuTrigger as-child>
                      <Button variant="ghost" size="icon-sm" :disabled="busyIds.has(sandbox.id)">
                        <MoreHorizontal class="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem class="text-destructive" @click="handleDelete(sandbox)">
                        <Trash2 class="mr-2 h-4 w-4" />
                        Entfernen
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
        <DataTablePagination
          :page="recentPage"
          :total-items="recentSandboxes.length"
          :page-size="recentPageSize"
          @update:page="recentPage = $event"
        />
      </section>

      <section v-if="isAdmin">
        <h3 class="text-muted-foreground mb-3 text-sm font-medium">Alle Instanzen</h3>
        <div class="mb-4 flex items-center gap-3">
          <Select v-model="adminStatusFilter">
            <SelectTrigger class="w-40">
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
                <TableHead class="w-[14%]">Status</TableHead>
                <TableHead class="w-[14%]">Name</TableHead>
                <TableHead class="w-[16%]">Vorlage</TableHead>
                <TableHead class="w-[18%]">Besitzer</TableHead>
                <TableHead class="w-[14%]">Gestartet</TableHead>
                <TableHead class="w-[14%]">Läuft ab</TableHead>
                <TableHead class="w-[10%] text-right">Aktionen</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <template v-if="loading">
                <TableRow v-for="i in 3" :key="i" class="h-13">
                  <TableCell><Skeleton class="h-5 w-16 rounded-full" /></TableCell>
                  <TableCell><Skeleton class="h-4 w-24" /></TableCell>
                  <TableCell><Skeleton class="h-4 w-28" /></TableCell>
                  <TableCell><Skeleton class="h-4 w-28" /></TableCell>
                  <TableCell><Skeleton class="h-4 w-24" /></TableCell>
                  <TableCell><Skeleton class="h-4 w-24" /></TableCell>
                  <TableCell class="text-right"><Skeleton class="ml-auto h-7 w-7" /></TableCell>
                </TableRow>
              </template>
              <TableEmpty v-else-if="filteredAllSandboxes.length === 0" :colspan="7">
                Keine Instanzen gefunden
              </TableEmpty>
              <TableRow
                v-for="sandbox in paginatedAllSandboxes"
                :key="sandbox.id"
                class="h-13 cursor-pointer"
                @click="handleRowClick(sandbox)"
              >
                <TableCell>
                  <StatusBadge :status="sandbox.status" :state-reason="sandbox.stateReason" />
                </TableCell>
                <TableCell>
                  <span class="text-muted-foreground truncate text-sm">{{
                    sandbox.displayName || '—'
                  }}</span>
                </TableCell>
                <TableCell class="font-medium">{{ getImageName(sandbox.imageId) }}</TableCell>
                <TableCell class="text-muted-foreground text-sm">
                  {{ getSandboxOwnerLabel(sandbox) }}
                </TableCell>
                <TableCell class="text-muted-foreground">{{
                  formatDateTime(sandbox.createdAt)
                }}</TableCell>
                <TableCell class="text-muted-foreground">
                  {{ sandbox.expiresAt ? formatDateTime(sandbox.expiresAt) : '—' }}
                </TableCell>
                <TableCell class="text-right" @click.stop>
                  <DropdownMenu>
                    <DropdownMenuTrigger as-child>
                      <Button variant="ghost" size="icon-sm" :disabled="busyIds.has(sandbox.id)">
                        <MoreHorizontal class="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem
                        v-if="sandbox.status === 'running' || sandbox.status === 'starting'"
                        @click="handleExtend(sandbox)"
                      >
                        <Clock class="mr-2 h-4 w-4" />
                        Verlängern
                      </DropdownMenuItem>
                      <DropdownMenuSeparator
                        v-if="sandbox.status === 'running' || sandbox.status === 'starting'"
                      />
                      <DropdownMenuItem class="text-destructive" @click="handleDelete(sandbox)">
                        <template
                          v-if="
                            ['running', 'starting', 'paused', 'stopping'].includes(sandbox.status)
                          "
                        >
                          <Square class="mr-2 h-4 w-4" />
                          Beenden
                        </template>
                        <template v-else>
                          <Trash2 class="mr-2 h-4 w-4" />
                          Entfernen
                        </template>
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
        <DataTablePagination
          :page="allPage"
          :total-items="filteredAllSandboxes.length"
          :page-size="allPageSize"
          @update:page="allPage = $event"
        />
      </section>
    </div>

    <SandboxDetailDialog
      v-model:open="showDetail"
      :sandbox="selectedSandbox"
      :health="selectedHealth"
      :image="selectedImage"
    />

    <NewSandboxDialog
      v-model:open="showNewSandbox"
      :images="images"
      @submit="handleCreateSandbox"
    />

    <EditSandboxDialog
      v-model:open="showEdit"
      :sandbox="selectedSandbox"
      @submit="handleEditSandbox"
    />

    <ExtendTtlDialog
      v-model:open="showExtend"
      :sandbox-id="selectedSandbox?.id ?? ''"
      :sandbox-name="selectedSandbox ? getSandboxDisplayName(selectedSandbox) : ''"
      @submit="handleExtendTtl"
    />

    <SnapshotDialog
      v-model:open="showSnapshot"
      :sandbox-name="selectedSandbox ? getSandboxDisplayName(selectedSandbox) : ''"
      :source-image="selectedSandbox ? images.find((i) => i.id === selectedSandbox!.imageId) : null"
      :source-sandbox="selectedSandbox"
      @submit="handleCreateSnapshot"
    />

    <ConfirmDialog
      v-model:open="showConfirmDelete"
      :title="isSelectedActive ? 'Sandbox beenden' : 'Aus Verlauf entfernen'"
      :description="
        isSelectedActive
          ? `Bist du sicher, dass du ${selectedSandbox ? getSandboxDisplayName(selectedSandbox) : 'diese Sandbox'} beenden möchtest? Diese Aktion kann nicht rückgängig gemacht werden.`
          : `Bist du sicher, dass du ${selectedSandbox ? getSandboxDisplayName(selectedSandbox) : 'diese Sandbox'} endgültig aus dem Verlauf entfernen möchtest?`
      "
      :confirm-label="isSelectedActive ? 'Beenden' : 'Entfernen'"
      @confirm="handleConfirmDelete"
    />
  </div>
</template>
