<script setup lang="ts">
import { Loader2, Lock, Plus, Trash2, Upload } from 'lucide-vue-next'
import { ref, watch } from 'vue'
import { toast } from 'vue-sonner'

import { imagesApi } from '@/api'
import IconPicker from '@/components/shared/IconPicker.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { getApiErrorMessage } from '@/utils/error'
import { resolveAssetUrl } from '@/utils/formatters'
import {
  collectMetadata,
  metadataItemToRow,
  type MetadataRow,
  newMetadataRow,
} from '@/utils/metadata'

import type { Image, MetadataItem } from '@/types'

const props = defineProps<{
  open: boolean
  image: Image | null
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  saved: []
}>()

const title = ref('')
const description = ref('')
const isPublic = ref(true)
const thumbnailFile = ref<File | null>(null)
const thumbnailPreview = ref<string | undefined>()
const removeThumbnail = ref(false)
const busy = ref(false)
const fileInputRef = ref<HTMLInputElement | null>(null)

const fieldRows = ref<MetadataRow[]>([])
const actionRows = ref<MetadataRow[]>([])
const registryKeys = ref<Set<string>>(new Set())

function revokeBlobPreview() {
  if (thumbnailPreview.value?.startsWith('blob:')) {
    URL.revokeObjectURL(thumbnailPreview.value)
  }
}

watch(
  () => props.open,
  async (open) => {
    if (!open || !props.image) return

    revokeBlobPreview()
    title.value = props.image.title ?? ''
    description.value = props.image.description ?? ''
    isPublic.value = props.image.isPublic
    thumbnailFile.value = null
    thumbnailPreview.value = resolveAssetUrl(props.image.thumbnailUrl)
    removeThumbnail.value = false
    busy.value = false
    if (fileInputRef.value) fileInputRef.value.value = ''

    const regMeta = await imagesApi
      .lookupRegistry(props.image.registryRef ?? props.image.name)
      .catch(() => [] as MetadataItem[])

    const regKeySet = new Set(regMeta.map((m) => m.key))
    registryKeys.value = regKeySet

    const allMeta = props.image.metadata ?? []
    const imgValueMap = new Map(allMeta.map((m) => [m.key, m.value]))
    const customOnly = allMeta.filter((m) => !regKeySet.has(m.key))

    function toRow(m: MetadataItem, fromReg: boolean): MetadataRow {
      const row = metadataItemToRow(m, fromReg)
      if (fromReg) {
        row.value = imgValueMap.get(m.key) ?? m.value ?? ''
      }
      return row
    }

    fieldRows.value = [
      ...regMeta
        .filter((m) => m.type === 'field' || m.type === 'setting')
        .map((m) => toRow(m, true)),
      ...customOnly
        .filter((m) => m.type === 'field' || m.type === 'setting')
        .map((m) => toRow(m, false)),
    ]

    actionRows.value = [
      ...regMeta.filter((m) => m.type === 'action').map((m) => toRow(m, true)),
      ...customOnly.filter((m) => m.type === 'action').map((m) => toRow(m, false)),
    ]
  },
)

function handleFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  revokeBlobPreview()
  thumbnailFile.value = file
  removeThumbnail.value = false
  thumbnailPreview.value = URL.createObjectURL(file)
}

function handleRemoveThumbnail() {
  revokeBlobPreview()
  thumbnailFile.value = null
  thumbnailPreview.value = undefined
  removeThumbnail.value = true
  if (fileInputRef.value) fileInputRef.value.value = ''
}

function addField() {
  fieldRows.value.push(newMetadataRow())
}

function removeField(index: number) {
  fieldRows.value.splice(index, 1)
}

function addAction() {
  actionRows.value.push(newMetadataRow())
}

function removeAction(index: number) {
  actionRows.value.splice(index, 1)
}

async function handleSubmit() {
  if (!props.image) return
  busy.value = true
  try {
    const metadata = collectMetadata(fieldRows.value, actionRows.value)

    await imagesApi.update(props.image.id, {
      title: title.value || null,
      description: description.value || null,
      isPublic: isPublic.value,
      metadata,
    })

    if (thumbnailFile.value) {
      await imagesApi.uploadThumbnail(props.image.id, thumbnailFile.value)
    } else if (removeThumbnail.value && props.image.thumbnailUrl) {
      await imagesApi.deleteThumbnail(props.image.id)
    }

    toast.success('Vorlage wurde aktualisiert')
    emit('saved')
    emit('update:open', false)
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Fehler beim Speichern'))
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <Sheet :open="open" @update:open="emit('update:open', $event)">
    <SheetContent side="right" class="overflow-y-auto">
      <SheetHeader>
        <SheetTitle>Vorlage bearbeiten</SheetTitle>
        <SheetDescription>Bearbeite die Details dieser Vorlage.</SheetDescription>
      </SheetHeader>
      <form id="edit-image-form" class="grid gap-4 px-4" @submit.prevent="handleSubmit">
        <div class="grid gap-2">
          <Label for="edit-title">Titel</Label>
          <Input
            id="edit-title"
            v-model="title"
            placeholder="Leere Installation"
            :disabled="busy"
          />
        </div>
        <div class="grid gap-2">
          <Label for="edit-description">Beschreibung</Label>
          <Textarea
            id="edit-description"
            v-model="description"
            placeholder="Beschreibung der Vorlage..."
            :disabled="busy"
          />
        </div>
        <div class="grid gap-2">
          <Label>Thumbnail</Label>
          <div v-if="thumbnailPreview" class="relative">
            <img
              :src="thumbnailPreview"
              alt="Thumbnail"
              class="h-32 w-full rounded-md border object-cover"
            />
            <Button
              type="button"
              variant="destructive"
              size="icon-sm"
              class="absolute top-2 right-2"
              :disabled="busy"
              @click="handleRemoveThumbnail"
            >
              <Trash2 class="h-3.5 w-3.5" />
            </Button>
          </div>
          <Label
            for="edit-thumbnail"
            class="text-muted-foreground hover:border-primary hover:text-foreground flex cursor-pointer items-center gap-2 rounded-md border border-dashed p-3 text-sm transition-colors"
            :class="{ 'pointer-events-none opacity-50': busy }"
          >
            <Upload class="h-4 w-4" />
            {{ thumbnailPreview ? 'Thumbnail ersetzen' : 'Thumbnail hochladen' }}
          </Label>
          <input
            id="edit-thumbnail"
            ref="fileInputRef"
            type="file"
            accept="image/*"
            class="hidden"
            :disabled="busy"
            @change="handleFileChange"
          />
        </div>
        <div class="flex items-center justify-between">
          <Label for="edit-public">Öffentlich sichtbar</Label>
          <Switch id="edit-public" v-model="isPublic" :disabled="busy" />
        </div>

        <div class="grid gap-2">
          <Label>Felder</Label>
          <div v-for="(row, index) in fieldRows" :key="index" class="space-y-1.5">
            <div class="flex items-center gap-2">
              <div class="relative flex-1">
                <Input
                  v-model="row.label"
                  placeholder="Label"
                  :disabled="busy || row.fromRegistry"
                  :class="{ 'pr-7': row.fromRegistry }"
                />
                <Lock
                  v-if="row.fromRegistry"
                  class="text-muted-foreground absolute top-1/2 right-2 h-3.5 w-3.5 -translate-y-1/2"
                />
              </div>
              <Input v-model="row.value" placeholder="Wert" class="flex-1" :disabled="busy" />
              <Button
                v-if="!row.fromRegistry"
                type="button"
                variant="ghost"
                size="icon"
                :disabled="busy"
                @click="removeField(index)"
              >
                <Trash2 class="h-4 w-4" />
              </Button>
              <div v-else class="w-9" />
            </div>
            <div v-if="!row.fromRegistry" class="flex gap-1.5 pl-0.5">
              <IconPicker v-model="row.icon" :disabled="busy" />
              <select
                v-model="row.show"
                class="border-input bg-background h-7 rounded-md border px-1.5 text-[11px]"
                :disabled="busy"
              >
                <option value="sandbox">Sandbox</option>
                <option value="template">Template</option>
                <option value="both">Beide</option>
              </select>
              <select
                v-model="row.condition"
                class="border-input bg-background h-7 rounded-md border px-1.5 text-[11px]"
                :disabled="busy"
              >
                <option value="always">Immer</option>
                <option value="ready">Wenn bereit</option>
              </select>
            </div>
          </div>
          <Button type="button" variant="outline" size="sm" :disabled="busy" @click="addField">
            <Plus class="mr-1 h-4 w-4" />
            Feld hinzufügen
          </Button>
        </div>

        <div class="grid gap-2">
          <Label>Aktionen</Label>
          <div v-for="(row, index) in actionRows" :key="index" class="space-y-1.5">
            <div class="flex items-center gap-2">
              <div class="relative flex-1">
                <Input
                  v-model="row.label"
                  placeholder="Label"
                  :disabled="busy || row.fromRegistry"
                  :class="{ 'pr-7': row.fromRegistry }"
                />
                <Lock
                  v-if="row.fromRegistry"
                  class="text-muted-foreground absolute top-1/2 right-2 h-3.5 w-3.5 -translate-y-1/2"
                />
              </div>
              <div class="relative flex-1">
                <Input
                  v-model="row.value"
                  placeholder="https://..."
                  :disabled="busy || row.fromRegistry"
                  :class="{ 'pr-7': row.fromRegistry }"
                />
                <Lock
                  v-if="row.fromRegistry"
                  class="text-muted-foreground absolute top-1/2 right-2 h-3.5 w-3.5 -translate-y-1/2"
                />
              </div>
              <Button
                v-if="!row.fromRegistry"
                type="button"
                variant="ghost"
                size="icon"
                :disabled="busy"
                @click="removeAction(index)"
              >
                <Trash2 class="h-4 w-4" />
              </Button>
              <div v-else class="w-9" />
            </div>
            <div v-if="!row.fromRegistry" class="flex gap-1.5 pl-0.5">
              <IconPicker v-model="row.icon" :disabled="busy" />
              <select
                v-model="row.show"
                class="border-input bg-background h-7 rounded-md border px-1.5 text-[11px]"
                :disabled="busy"
              >
                <option value="sandbox">Sandbox</option>
                <option value="template">Template</option>
                <option value="both">Beide</option>
              </select>
              <select
                v-model="row.condition"
                class="border-input bg-background h-7 rounded-md border px-1.5 text-[11px]"
                :disabled="busy"
              >
                <option value="always">Immer</option>
                <option value="ready">Wenn bereit</option>
              </select>
              <select
                v-model="row.size"
                class="border-input bg-background h-7 rounded-md border px-1.5 text-[11px]"
                :disabled="busy"
              >
                <option value="default">Normal</option>
                <option value="icon">Nur Icon</option>
              </select>
            </div>
          </div>
          <Button type="button" variant="outline" size="sm" :disabled="busy" @click="addAction">
            <Plus class="mr-1 h-4 w-4" />
            Aktion hinzufügen
          </Button>
        </div>
      </form>
      <SheetFooter>
        <Button type="button" variant="outline" :disabled="busy" @click="emit('update:open', false)"
          >Abbrechen</Button
        >
        <Button type="submit" form="edit-image-form" :disabled="busy">
          <Loader2 v-if="busy" class="mr-1 h-4 w-4 animate-spin" />
          {{ busy ? 'Wird gespeichert...' : 'Speichern' }}
        </Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
</template>
