<script setup lang="ts">
import { Loader2, Lock, Plus, Trash2, Upload } from 'lucide-vue-next'
import { computed, ref, watch } from 'vue'

import { imagesApi } from '@/api'
import IconPicker from '@/components/shared/IconPicker.vue'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import {
  collectMetadata,
  metadataItemToRow,
  type MetadataRow,
  newMetadataRow,
} from '@/utils/metadata'

import type { MetadataItem } from '@/types'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  submit: [
    payload: {
      name: string
      tag: string
      title: string
      description: string
      isPublic: boolean
      thumbnailFile?: File
      metadata: MetadataItem[]
    },
    done: (success: boolean) => void,
  ]
}>()

const name = ref('')
const tag = ref('')
const title = ref('')
const description = ref('')
const isPublic = ref(true)
const thumbnailFile = ref<File | null>(null)
const thumbnailPreview = ref<string | undefined>()
const busy = ref(false)
const fileInputRef = ref<HTMLInputElement | null>(null)

const registryMeta = ref<MetadataItem[]>([])
const fieldRows = ref<MetadataRow[]>([])
const actionRows = ref<MetadataRow[]>([])

let lookupTimer: ReturnType<typeof setTimeout> | undefined

const registryFields = computed(() =>
  registryMeta.value.filter((m) => m.type === 'field' || m.type === 'setting'),
)
const registryActions = computed(() => registryMeta.value.filter((m) => m.type === 'action'))

watch(name, (val) => {
  clearTimeout(lookupTimer)
  if (!val) {
    registryMeta.value = []
    rebuildRows()
    return
  }
  lookupTimer = setTimeout(async () => {
    try {
      registryMeta.value = await imagesApi.lookupRegistry(val)
    } catch {
      registryMeta.value = []
    }
    rebuildRows()
  }, 400)
})

function mapRegistryRow(m: MetadataItem): MetadataRow {
  return metadataItemToRow(m, true)
}

function rebuildRows() {
  const customFields = fieldRows.value.filter((r) => !r.fromRegistry)
  const customActions = actionRows.value.filter((r) => !r.fromRegistry)

  fieldRows.value = [...registryFields.value.map(mapRegistryRow), ...customFields]
  actionRows.value = [...registryActions.value.map(mapRegistryRow), ...customActions]
}

function revokeBlobPreview() {
  if (thumbnailPreview.value?.startsWith('blob:')) {
    URL.revokeObjectURL(thumbnailPreview.value)
  }
}

function resetState() {
  name.value = ''
  tag.value = ''
  title.value = ''
  description.value = ''
  isPublic.value = true
  revokeBlobPreview()
  thumbnailFile.value = null
  thumbnailPreview.value = undefined
  busy.value = false
  registryMeta.value = []
  fieldRows.value = []
  actionRows.value = []
  if (fileInputRef.value) fileInputRef.value.value = ''
}

watch(
  () => props.open,
  (open) => {
    if (open) resetState()
  },
)

function handleFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  revokeBlobPreview()
  thumbnailFile.value = file
  thumbnailPreview.value = URL.createObjectURL(file)
}

function handleRemoveThumbnail() {
  revokeBlobPreview()
  thumbnailFile.value = null
  thumbnailPreview.value = undefined
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

function handleSubmit() {
  if (!name.value || !tag.value) return
  busy.value = true

  const metadata = collectMetadata(fieldRows.value, actionRows.value)

  emit(
    'submit',
    {
      name: name.value,
      tag: tag.value,
      title: title.value,
      description: description.value,
      isPublic: isPublic.value,
      thumbnailFile: thumbnailFile.value ?? undefined,
      metadata,
    },
    (success: boolean) => {
      busy.value = false
      if (success) {
        emit('update:open', false)
      }
    },
  )
}
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="max-h-[90vh] overflow-y-auto sm:max-w-[500px]">
      <DialogHeader>
        <DialogTitle>Vorlage hinzufügen</DialogTitle>
        <DialogDescription
          >Füge ein neues Docker-Image als Sandbox-Vorlage hinzu.</DialogDescription
        >
      </DialogHeader>
      <form class="grid gap-4 py-4" @submit.prevent="handleSubmit">
        <div class="grid gap-2">
          <Label for="image-name">Image Name</Label>
          <Input
            id="image-name"
            v-model="name"
            placeholder="dockware/dev"
            required
            :disabled="busy"
          />
        </div>
        <div class="grid gap-2">
          <Label for="image-tag">Tag</Label>
          <Input id="image-tag" v-model="tag" placeholder="latest" required :disabled="busy" />
        </div>
        <div class="grid gap-2">
          <Label for="image-title">Titel</Label>
          <Input
            id="image-title"
            v-model="title"
            placeholder="Leere Installation"
            :disabled="busy"
          />
        </div>
        <div class="grid gap-2">
          <Label for="image-description">Beschreibung</Label>
          <Textarea
            id="image-description"
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
            for="image-thumbnail"
            class="text-muted-foreground hover:border-primary hover:text-foreground flex cursor-pointer items-center gap-2 rounded-md border border-dashed p-3 text-sm transition-colors"
            :class="{ 'pointer-events-none opacity-50': busy }"
          >
            <Upload class="h-4 w-4" />
            {{ thumbnailPreview ? 'Thumbnail ersetzen' : 'Thumbnail hochladen' }}
          </Label>
          <input
            id="image-thumbnail"
            ref="fileInputRef"
            type="file"
            accept="image/*"
            class="hidden"
            :disabled="busy"
            @change="handleFileChange"
          />
        </div>
        <div class="flex items-center justify-between">
          <Label for="image-public">Öffentlich sichtbar</Label>
          <Switch id="image-public" v-model="isPublic" :disabled="busy" />
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

        <DialogFooter class="pt-2">
          <Button
            type="button"
            variant="outline"
            :disabled="busy"
            @click="emit('update:open', false)"
            >Abbrechen</Button
          >
          <Button type="submit" :disabled="!name || !tag || busy">
            <Loader2 v-if="busy" class="mr-1 h-4 w-4 animate-spin" />
            {{ busy ? 'Wird hinzugefügt...' : 'Hinzufügen' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
