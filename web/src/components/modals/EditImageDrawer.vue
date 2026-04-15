<script setup lang="ts">
import { Loader2, Trash2, Upload } from 'lucide-vue-next'
import { ref, watch } from 'vue'
import { toast } from 'vue-sonner'

import { imagesApi } from '@/api'
import MetadataEditor from '@/components/metadata/MetadataEditor.vue'
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
import { useImagesStore } from '@/stores/images.store'
import { getApiErrorMessage } from '@/utils/error'
import { resolveAssetUrl } from '@/utils/formatters'

import type { Image, MetadataItem, MetadataSchema } from '@/types'

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

const metadata = ref<MetadataItem[]>([])
const registrySchema = ref<MetadataSchema | null>(null)
const imagesStore = useImagesStore()

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

    metadata.value = [...(props.image.metadata ?? [])]
    registrySchema.value = await imagesApi.lookupRegistry(
      props.image.registryRef ?? props.image.name,
    )
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

async function handleSubmit() {
  if (!props.image) return
  busy.value = true
  try {
    let updated = await imagesApi.update(props.image.id, {
      title: title.value || null,
      description: description.value || null,
      isPublic: isPublic.value,
      metadata: metadata.value,
    })

    if (thumbnailFile.value) {
      updated = await imagesApi.uploadThumbnail(props.image.id, thumbnailFile.value)
    } else if (removeThumbnail.value && props.image.thumbnailUrl) {
      await imagesApi.deleteThumbnail(props.image.id)
      updated = { ...updated, thumbnailUrl: undefined }
    }

    imagesStore.upsertImage(updated)

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
        <div class="flex items-center gap-3 pt-2">
          <Label class="text-sm font-medium">Metadaten</Label>
          <div class="bg-border h-px flex-1" />
        </div>
        <MetadataEditor v-model="metadata" :registry-schema="registrySchema" :disabled="busy" />
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
