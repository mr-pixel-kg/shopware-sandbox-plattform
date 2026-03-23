<script setup lang="ts">
import { CircleCheck, CircleX, Loader2, Plus, Trash2 } from 'lucide-vue-next'
import { ref } from 'vue'
import { toast } from 'vue-sonner'

import AddImageDialog from '@/components/modals/AddImageDialog.vue'
import ConfirmDialog from '@/components/modals/ConfirmDialog.vue'
import PageHeader from '@/components/shared/PageHeader.vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { DonutProgress } from '@/components/ui/donut-progress'
import { Skeleton } from '@/components/ui/skeleton'
import { Switch } from '@/components/ui/switch'
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
import { getApiErrorMessage } from '@/utils/error'

const { images, pendingPulls, loading, createImage, deleteImage } = useImages('all')

const showAddImage = ref(false)
const showConfirmDelete = ref(false)
const selectedImageId = ref<string | null>(null)

function requestDelete(id: string) {
  selectedImageId.value = id
  showConfirmDelete.value = true
}

async function handleCreateImage(
  payload: { name: string; tag: string; title: string; description: string; isPublic: boolean },
  done: (success: boolean) => void,
) {
  try {
    const image = await createImage(payload)
    if (image.status === 'ready') {
      toast.success('Vorlage wurde hinzugefügt')
    } else {
      toast.success('Image wird heruntergeladen...')
    }
    done(true)
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Fehler beim Hinzufügen'))
    done(false)
  }
}

async function handleConfirmDelete() {
  if (!selectedImageId.value) return
  try {
    await deleteImage(selectedImageId.value)
    toast.success('Vorlage wurde gelöscht')
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Fehler beim Löschen'))
  }
}

function handleToggleVisibility() {
  // TODO: Call update API when available
  toast.info('Sichtbarkeit ändern ist noch nicht verfügbar')
}
</script>

<template>
  <div>
    <PageHeader title="Vorlagen" subtitle="Docker-Images als Sandbox-Vorlagen verwalten.">
      <template #actions>
        <Button @click="showAddImage = true">
          <Plus class="mr-1 h-4 w-4" />
          Vorlage hinzufügen
        </Button>
      </template>
    </PageHeader>

    <div v-if="pendingPulls.length > 0" class="mb-4 space-y-2">
      <div
        v-for="pull in pendingPulls"
        :key="pull.id"
        class="bg-muted/50 flex items-center gap-3 rounded-md border p-3"
      >
        <DonutProgress :model-value="pull.percent" class="h-5 w-5" />
        <div class="min-w-0 flex-1">
          <span class="text-sm font-medium">{{ pull.title || pull.name }}</span>
          <Badge variant="secondary" class="ml-2 text-xs">{{ pull.name }}:{{ pull.tag }}</Badge>
        </div>
        <span class="text-muted-foreground text-sm tabular-nums">{{ pull.percent }}%</span>
      </div>
    </div>

    <div class="rounded-md border">
      <Table class="table-fixed">
        <TableHeader>
          <TableRow>
            <TableHead class="w-[35%]">Vorlage</TableHead>
            <TableHead class="w-[25%]">Image</TableHead>
            <TableHead class="w-[15%]">Status</TableHead>
            <TableHead class="w-[15%]">Öffentlich</TableHead>
            <TableHead class="w-[10%] text-right">Aktionen</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <template v-if="loading">
            <TableRow v-for="i in 3" :key="i" class="h-13">
              <TableCell><Skeleton class="h-4 w-32" /></TableCell>
              <TableCell><Skeleton class="h-5 w-28 rounded-full" /></TableCell>
              <TableCell><Skeleton class="h-4 w-16" /></TableCell>
              <TableCell><Skeleton class="h-4 w-8 rounded-full" /></TableCell>
              <TableCell class="text-right"><Skeleton class="ml-auto h-7 w-7" /></TableCell>
            </TableRow>
          </template>
          <TableEmpty v-else-if="images.length === 0" :colspan="5">
            Keine Vorlagen vorhanden
          </TableEmpty>
          <TableRow v-for="image in images" :key="image.id" class="h-13">
            <TableCell>
              <div>
                <div class="font-medium">{{ image.title || image.name }}</div>
                <div v-if="image.description" class="text-muted-foreground text-xs">
                  {{ image.description }}
                </div>
              </div>
            </TableCell>
            <TableCell>
              <Badge variant="secondary">{{ image.name }}:{{ image.tag }}</Badge>
            </TableCell>
            <TableCell>
              <div
                v-if="image.status === 'ready'"
                class="flex items-center gap-1.5 text-emerald-600"
              >
                <CircleCheck class="h-4 w-4" />
                <span class="text-sm">Bereit</span>
              </div>
              <div
                v-else-if="image.status === 'pulling'"
                class="flex items-center gap-1.5 text-blue-600"
              >
                <Loader2 class="h-4 w-4 animate-spin" />
                <span class="text-sm">Wird geladen</span>
              </div>
              <div
                v-else-if="image.status === 'failed'"
                class="text-destructive flex items-center gap-1.5"
              >
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger as-child>
                      <div class="flex items-center gap-1.5">
                        <CircleX class="h-4 w-4" />
                        <span class="text-sm">Fehlgeschlagen</span>
                      </div>
                    </TooltipTrigger>
                    <TooltipContent v-if="image.error">{{ image.error }}</TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </div>
            </TableCell>
            <TableCell>
              <Switch :model-value="image.isPublic" @update:model-value="handleToggleVisibility" />
            </TableCell>
            <TableCell class="text-right">
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger as-child>
                    <Button
                      variant="ghost"
                      size="icon-sm"
                      class="text-destructive hover:text-destructive"
                      @click="requestDelete(image.id)"
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

    <AddImageDialog v-model:open="showAddImage" @submit="handleCreateImage" />

    <ConfirmDialog
      v-model:open="showConfirmDelete"
      title="Vorlage löschen"
      description="Bist du sicher, dass du diese Vorlage löschen möchtest? Alle zugehörigen Sandboxes werden ebenfalls beendet. Diese Aktion kann nicht rückgängig gemacht werden."
      confirm-label="Löschen"
      @confirm="handleConfirmDelete"
    />
  </div>
</template>
