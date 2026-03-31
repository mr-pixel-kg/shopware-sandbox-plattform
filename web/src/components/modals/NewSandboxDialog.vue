<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'
import { computed, ref, watch } from 'vue'

import { Badge } from '@/components/ui/badge'
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
import { ScrollArea } from '@/components/ui/scroll-area'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'

import type { Image } from '@/types'

const props = defineProps<{
  open: boolean
  images: Image[]
  preselectedImageId?: string
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  submit: [
    payload: {
      imageId: string
      ttlMinutes: number
      displayName?: string
      metadata?: Record<string, string>
    },
    done: (success: boolean) => void,
  ]
}>()

const selectedImageId = ref<string>('')
const displayName = ref('')
const ttlMinutes = ref('120')
const submitting = ref(false)
const metadataValues = ref<Record<string, string>>({})

const ttlOptions = [
  { value: '5', label: '5 Min' },
  { value: '30', label: '30 Min' },
  { value: '120', label: '2 Std' },
  { value: '240', label: '4 Std' },
  { value: '480', label: '8 Std' },
  { value: '1440', label: '24 Std' },
  { value: 'unlimited', label: 'Unbegrenzt' },
]

const selectedImage = computed(() => props.images.find((i) => i.id === selectedImageId.value))

const imageMeta = computed(() => selectedImage.value?.metadata ?? [])

const editableFields = computed(() => imageMeta.value.filter((m) => m.type === 'field'))
const readOnlyItems = computed(() =>
  imageMeta.value.filter((m) => m.type === 'setting' || m.type === 'info'),
)

function populateMetadataDefaults() {
  const defaults: Record<string, string> = {}
  for (const item of imageMeta.value) {
    if (item.type === 'field' || item.type === 'setting') {
      defaults[item.key] = item.value ?? ''
    }
  }
  metadataValues.value = defaults
}

watch(
  () => props.open,
  (open) => {
    if (open) {
      selectedImageId.value = props.preselectedImageId || props.images[0]?.id || ''
      displayName.value = ''
      ttlMinutes.value = '120'
      submitting.value = false
      populateMetadataDefaults()
    }
  },
)

watch(selectedImageId, populateMetadataDefaults)

function handleSubmit() {
  if (!selectedImageId.value) return
  submitting.value = true

  const metadata =
    Object.keys(metadataValues.value).length > 0 ? { ...metadataValues.value } : undefined

  const trimmedName = displayName.value.trim()
  emit(
    'submit',
    {
      imageId: selectedImageId.value,
      ttlMinutes: ttlMinutes.value === 'unlimited' ? 0 : Number(ttlMinutes.value),
      displayName: trimmedName || undefined,
      metadata,
    },
    (success: boolean) => {
      submitting.value = false
      if (success) {
        emit('update:open', false)
      }
    },
  )
}
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="gap-0 p-0 sm:max-w-170">
      <DialogHeader class="p-6 pb-4">
        <DialogTitle>Neue Sandbox</DialogTitle>
        <DialogDescription>Wähle eine Vorlage und konfiguriere die Laufzeit.</DialogDescription>
      </DialogHeader>

      <div class="flex min-h-85 overflow-hidden border-t">
        <div class="w-55 shrink-0 border-r">
          <ScrollArea class="h-85">
            <div class="p-2">
              <button
                v-for="image in images"
                :key="image.id"
                class="hover:bg-accent flex w-full items-start gap-2 rounded-md px-3 py-2 text-left text-sm transition-colors"
                :class="{ 'bg-accent': selectedImageId === image.id }"
                :disabled="submitting"
                @click="selectedImageId = image.id"
              >
                <div class="min-w-0 flex-1">
                  <div class="truncate font-medium">{{ image.title || image.name }}</div>
                  <div class="text-muted-foreground truncate text-xs">
                    {{ image.name }}:{{ image.tag }}
                  </div>
                </div>
              </button>
            </div>
          </ScrollArea>
        </div>

        <div class="flex-1 overflow-y-auto p-6">
          <div v-if="selectedImage" class="space-y-6">
            <div>
              <h3 class="text-sm font-medium">{{ selectedImage.title || selectedImage.name }}</h3>
              <p v-if="selectedImage.description" class="text-muted-foreground mt-1 text-sm">
                {{ selectedImage.description }}
              </p>
              <Badge variant="secondary" class="mt-2">
                {{ selectedImage.name }}:{{ selectedImage.tag }}
              </Badge>
            </div>

            <div class="grid gap-1.5">
              <Label for="sandbox-display-name">Name (optional)</Label>
              <Input
                id="sandbox-display-name"
                v-model="displayName"
                placeholder="z.B. Checkout-Test"
                :disabled="submitting"
              />
            </div>

            <div class="grid gap-2 overflow-hidden">
              <Label>Laufzeit</Label>
              <div class="overflow-x-auto">
                <ToggleGroup
                  v-model="ttlMinutes"
                  type="single"
                  variant="outline"
                  class="w-max"
                  :disabled="submitting"
                >
                  <ToggleGroupItem v-for="opt in ttlOptions" :key="opt.value" :value="opt.value">
                    {{ opt.label }}
                  </ToggleGroupItem>
                </ToggleGroup>
              </div>
            </div>

            <div v-if="editableFields.length > 0" class="space-y-3">
              <Label class="block">Konfiguration</Label>
              <div v-for="item in editableFields" :key="item.key" class="grid gap-1.5">
                <Label :for="'field-' + item.key" class="text-xs">
                  {{ item.label }}
                  <span v-if="item.required" class="text-destructive">*</span>
                </Label>
                <Input
                  :id="'field-' + item.key"
                  v-model="metadataValues[item.key]"
                  :type="item.input === 'password' ? 'password' : 'text'"
                  :placeholder="item.value"
                  :disabled="submitting"
                />
              </div>
            </div>

            <div
              v-if="readOnlyItems.length > 0"
              class="bg-muted/50 space-y-1.5 rounded-md border px-3 py-2"
            >
              <div
                v-for="item in readOnlyItems"
                :key="item.key"
                class="flex justify-between text-xs"
              >
                <span class="text-muted-foreground">{{ item.label }}</span>
                <span class="font-mono text-[11px]">{{
                  metadataValues[item.key] || item.value
                }}</span>
              </div>
            </div>
          </div>

          <div v-else class="text-muted-foreground flex h-full items-center justify-center text-sm">
            Wähle eine Vorlage aus der Liste
          </div>
        </div>
      </div>

      <DialogFooter class="border-t p-6 pt-4">
        <Button variant="outline" :disabled="submitting" @click="emit('update:open', false)"
          >Abbrechen</Button
        >
        <Button :disabled="!selectedImageId || submitting" @click="handleSubmit">
          <Loader2 v-if="submitting" class="mr-1 h-4 w-4 animate-spin" />
          {{ submitting ? 'Wird gestartet...' : 'Sandbox starten' }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
