<script setup lang="ts">
import { Loader2, Trash2, Upload } from 'lucide-vue-next'
import { ref, watch } from 'vue'

import { imagesApi, registrySearchApi } from '@/api'
import MetadataEditor from '@/components/metadata/MetadataEditor.vue'
import AutocompleteInput from '@/components/shared/AutocompleteInput.vue'
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'

import type { Suggestion } from '@/components/shared/AutocompleteInput.vue'
import type { MetadataItem, MetadataSchema } from '@/types'

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

const imageSuggestions = ref<Suggestion[]>([])
const tagSuggestions = ref<Suggestion[]>([])
const imageSearchLoading = ref(false)
const tagSearchLoading = ref(false)

const metadata = ref<MetadataItem[]>([])
const registrySchema = ref<MetadataSchema | null>(null)

let imageSearchTimer: ReturnType<typeof setTimeout> | undefined
let tagSearchTimer: ReturnType<typeof setTimeout> | undefined
let registryLookupTimer: ReturnType<typeof setTimeout> | undefined

watch(name, (val) => {
  clearTimeout(imageSearchTimer)
  clearTimeout(registryLookupTimer)
  tagSuggestions.value = []

  if (!val || val.length < 2) {
    imageSuggestions.value = []
    imageSearchLoading.value = false
    registrySchema.value = null
    metadata.value = []
    return
  }

  registryLookupTimer = setTimeout(async () => {
    const schema = await imagesApi.lookupRegistry(val)
    registrySchema.value = schema
    metadata.value = [...(schema?.items ?? [])]
  }, 400)

  imageSearchLoading.value = true
  imageSearchTimer = setTimeout(async () => {
    try {
      const results = await registrySearchApi.searchImages(val)
      imageSuggestions.value = results.map((r) => ({
        value: r.name,
        label: r.name,
        description: r.description || undefined,
      }))
    } catch {
      imageSuggestions.value = []
    } finally {
      imageSearchLoading.value = false
    }
  }, 400)
})

watch(tag, (val) => {
  clearTimeout(tagSearchTimer)
  if (!name.value) {
    tagSuggestions.value = []
    return
  }
  tagSearchLoading.value = true
  tagSearchTimer = setTimeout(async () => {
    try {
      const results = await registrySearchApi.searchTags(name.value, val)
      tagSuggestions.value = results.map((r) => ({
        value: r.name,
        label: r.name,
        description: r.lastUpdated
          ? new Date(r.lastUpdated).toLocaleDateString('de-DE')
          : undefined,
      }))
    } catch {
      tagSuggestions.value = []
    } finally {
      tagSearchLoading.value = false
    }
  }, 400)
})

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
  metadata.value = []
  registrySchema.value = null
  clearTimeout(imageSearchTimer)
  clearTimeout(tagSearchTimer)
  clearTimeout(registryLookupTimer)
  imageSuggestions.value = []
  tagSuggestions.value = []
  imageSearchLoading.value = false
  tagSearchLoading.value = false
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

function handleSubmit() {
  if (!name.value || !tag.value) return
  busy.value = true

  emit(
    'submit',
    {
      name: name.value,
      tag: tag.value,
      title: title.value,
      description: description.value,
      isPublic: isPublic.value,
      thumbnailFile: thumbnailFile.value ?? undefined,
      metadata: metadata.value,
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
    <DialogContent class="max-h-[90vh] overflow-y-auto sm:max-w-160">
      <DialogHeader>
        <DialogTitle>Vorlage hinzufügen</DialogTitle>
        <DialogDescription>
          Füge ein neues Docker-Image als Sandbox-Vorlage hinzu.
        </DialogDescription>
      </DialogHeader>
      <form @submit.prevent="handleSubmit">
        <Tabs default-value="general" class="mt-2">
          <TabsList class="w-full">
            <TabsTrigger value="general" class="flex-1">Allgemein</TabsTrigger>
            <TabsTrigger value="metadata" class="flex-1">Metadaten</TabsTrigger>
          </TabsList>

          <TabsContent value="general" class="mt-4 grid min-h-95 gap-4">
            <div class="grid gap-2">
              <Label for="image-name">Image Name</Label>
              <AutocompleteInput
                id="image-name"
                v-model="name"
                placeholder="dockware/dev"
                :suggestions="imageSuggestions"
                :loading="imageSearchLoading"
                :disabled="busy"
                :min-chars="2"
              />
            </div>
            <div class="grid gap-2">
              <Label for="image-tag">Tag</Label>
              <AutocompleteInput
                id="image-tag"
                v-model="tag"
                placeholder="latest"
                :suggestions="tagSuggestions"
                :loading="tagSearchLoading"
                :disabled="busy || !name"
                :min-chars="1"
              />
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
          </TabsContent>

          <TabsContent value="metadata" class="mt-4 min-h-95">
            <MetadataEditor v-model="metadata" :registry-schema="registrySchema" :disabled="busy" />
          </TabsContent>
        </Tabs>

        <DialogFooter class="pt-4">
          <Button
            type="button"
            variant="outline"
            :disabled="busy"
            @click="emit('update:open', false)"
          >
            Abbrechen
          </Button>
          <Button type="submit" :disabled="!name || !tag || busy">
            <Loader2 v-if="busy" class="mr-1 h-4 w-4 animate-spin" />
            {{ busy ? 'Wird hinzugefügt...' : 'Hinzufügen' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
