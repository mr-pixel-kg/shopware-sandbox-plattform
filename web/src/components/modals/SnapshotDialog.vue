<script setup lang="ts">
import {Loader2, Trash2, Upload} from 'lucide-vue-next'
import {ref, watch} from 'vue'

import {Button} from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {Input} from '@/components/ui/input'
import {Label} from '@/components/ui/label'
import {Switch} from '@/components/ui/switch'
import {Textarea} from '@/components/ui/textarea'

const props = defineProps<{
  open: boolean
  sandboxName: string
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
    },
    done: (success: boolean) => void,
  ]
}>()

const name = ref('')
const tag = ref('')
const title = ref('')
const description = ref('')
const isPublic = ref(false)
const thumbnailFile = ref<File | null>(null)
const thumbnailPreview = ref<string | undefined>()
const busy = ref(false)
const fileInputRef = ref<HTMLInputElement | null>(null)

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
  isPublic.value = false
  revokeBlobPreview()
  thumbnailFile.value = null
  thumbnailPreview.value = undefined
  busy.value = false
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
    <DialogContent class="sm:max-w-125">
      <DialogHeader>
        <DialogTitle>Snapshot erstellen</DialogTitle>
        <DialogDescription>
          Erstelle ein Image aus der laufenden Sandbox "{{ sandboxName }}".
        </DialogDescription>
      </DialogHeader>
      <form class="grid gap-4 py-4" @submit.prevent="handleSubmit">
        <div class="grid gap-2">
          <Label for="snapshot-name">Image Name</Label>
          <Input id="snapshot-name" v-model="name" placeholder="my-shop-snapshot" required :disabled="busy"/>
        </div>
        <div class="grid gap-2">
          <Label for="snapshot-tag">Tag</Label>
          <Input id="snapshot-tag" v-model="tag" placeholder="v1.0" required :disabled="busy"/>
        </div>
        <div class="grid gap-2">
          <Label for="snapshot-title">Titel</Label>
          <Input id="snapshot-title" v-model="title" placeholder="Mein Shop Snapshot" :disabled="busy"/>
        </div>
        <div class="grid gap-2">
          <Label for="snapshot-description">Beschreibung</Label>
          <Textarea
id="snapshot-description" v-model="description" placeholder="Beschreibung des Snapshots..."
                    :disabled="busy"/>
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
              <Trash2 class="h-3.5 w-3.5"/>
            </Button>
          </div>
          <Label
              for="snapshot-thumbnail"
              class="flex cursor-pointer items-center gap-2 rounded-md border border-dashed p-3 text-sm text-muted-foreground hover:border-primary hover:text-foreground transition-colors"
              :class="{ 'pointer-events-none opacity-50': busy }"
          >
            <Upload class="h-4 w-4"/>
            {{ thumbnailPreview ? 'Thumbnail ersetzen' : 'Thumbnail hochladen' }}
          </Label>
          <input
              id="snapshot-thumbnail"
              ref="fileInputRef"
              type="file"
              accept="image/*"
              class="hidden"
              :disabled="busy"
              @change="handleFileChange"
          />
        </div>
        <div class="flex items-center justify-between">
          <Label for="snapshot-public">Öffentlich sichtbar</Label>
          <Switch id="snapshot-public" v-model="isPublic" :disabled="busy"/>
        </div>
        <DialogFooter class="pt-2">
          <Button type="button" variant="outline" :disabled="busy" @click="emit('update:open', false)">Abbrechen
          </Button>
          <Button type="submit" :disabled="!name || !tag || busy">
            <Loader2 v-if="busy" class="h-4 w-4 animate-spin mr-1"/>
            {{ busy ? 'Wird erstellt...' : 'Snapshot erstellen' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
