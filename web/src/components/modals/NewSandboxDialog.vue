<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { Image } from '@/types'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Loader2 } from 'lucide-vue-next'

const props = defineProps<{
  open: boolean
  images: Image[]
  preselectedImageId?: string
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  submit: [payload: { imageId: string; ttlMinutes: number }, done: (success: boolean) => void]
}>()

const selectedImageId = ref<string>('')
const ttlMinutes = ref('120')
const submitting = ref(false)

const ttlOptions = [
  { value: '5', label: '5 Min' },
  { value: '30', label: '30 Min' },
  { value: '120', label: '2 Std' },
  { value: '240', label: '4 Std' },
  { value: '480', label: '8 Std' },
]

const selectedImage = computed(() => props.images.find((i) => i.id === selectedImageId.value))

watch(
  () => props.open,
  (open) => {
    if (open) {
      selectedImageId.value = props.preselectedImageId || props.images[0]?.id || ''
      ttlMinutes.value = '120'
      submitting.value = false
    }
  },
)

function handleSubmit() {
  if (!selectedImageId.value) return
  submitting.value = true
  emit(
    'submit',
    { imageId: selectedImageId.value, ttlMinutes: Number(ttlMinutes.value) },
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
    <DialogContent class="sm:max-w-[680px] p-0 gap-0">
      <DialogHeader class="p-6 pb-4">
        <DialogTitle>Neue Sandbox</DialogTitle>
        <DialogDescription>Wähle eine Vorlage und konfiguriere die Laufzeit.</DialogDescription>
      </DialogHeader>

      <div class="flex border-t min-h-[340px]">
        <div class="w-[220px] border-r">
          <ScrollArea class="h-[340px]">
            <div class="p-2">
              <button
                v-for="image in images"
                :key="image.id"
                class="w-full flex items-start gap-2 rounded-md px-3 py-2 text-left text-sm transition-colors hover:bg-accent"
                :class="{ 'bg-accent': selectedImageId === image.id }"
                :disabled="submitting"
                @click="selectedImageId = image.id"
              >
                <div class="min-w-0 flex-1">
                  <div class="font-medium truncate">{{ image.title || image.name }}</div>
                  <div class="text-xs text-muted-foreground truncate">{{ image.name }}:{{ image.tag }}</div>
                </div>
              </button>
            </div>
          </ScrollArea>
        </div>

        <div class="flex-1 p-6">
          <div v-if="selectedImage" class="space-y-6">
            <div>
              <h3 class="text-sm font-medium">{{ selectedImage.title || selectedImage.name }}</h3>
              <p v-if="selectedImage.description" class="text-sm text-muted-foreground mt-1">
                {{ selectedImage.description }}
              </p>
              <Badge variant="secondary" class="mt-2">
                {{ selectedImage.name }}:{{ selectedImage.tag }}
              </Badge>
            </div>

            <div>
              <Label class="mb-2 block">Laufzeit</Label>
              <ToggleGroup v-model="ttlMinutes" type="single" variant="outline" class="justify-start" :disabled="submitting">
                <ToggleGroupItem
                  v-for="opt in ttlOptions"
                  :key="opt.value"
                  :value="opt.value"
                >
                  {{ opt.label }}
                </ToggleGroupItem>
              </ToggleGroup>
            </div>
          </div>

          <div v-else class="flex items-center justify-center h-full text-sm text-muted-foreground">
            Wähle eine Vorlage aus der Liste
          </div>
        </div>
      </div>

      <DialogFooter class="p-6 pt-4 border-t">
        <Button variant="outline" :disabled="submitting" @click="emit('update:open', false)">Abbrechen</Button>
        <Button :disabled="!selectedImageId || submitting" @click="handleSubmit">
          <Loader2 v-if="submitting" class="h-4 w-4 animate-spin mr-1" />
          {{ submitting ? 'Wird gestartet...' : 'Sandbox starten' }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
