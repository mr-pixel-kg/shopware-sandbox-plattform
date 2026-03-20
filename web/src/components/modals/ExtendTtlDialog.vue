<script setup lang="ts">
import { ref } from 'vue'
import { toast } from 'vue-sonner'
import { useSandboxesStore } from '@/stores/sandboxes.store'
import { getApiErrorMessage } from '@/utils/error'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'

const props = defineProps<{
  open: boolean
  sandboxId: string
  sandboxName: string
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
}>()

const store = useSandboxesStore()
const ttlMinutes = ref('60')
const submitting = ref(false)

const options = [
  { value: '5', label: '+5 Min' },
  { value: '30', label: '+30 Min' },
  { value: '60', label: '+1 Std' },
  { value: '120', label: '+2 Std' },
  { value: '240', label: '+4 Std' },
]

async function handleSubmit() {
  if (!props.sandboxId || submitting.value) return
  submitting.value = true
  try {
    await store.extendTTL(props.sandboxId, Number(ttlMinutes.value))
    toast.success('Laufzeit wurde verlängert')
    emit('update:open', false)
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Fehler beim Verlängern der Laufzeit'))
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="sm:max-w-[400px]">
      <DialogHeader>
        <DialogTitle>Laufzeit verlängern</DialogTitle>
        <DialogDescription>
          Verlängere die Laufzeit von <strong>{{ sandboxName }}</strong>.
        </DialogDescription>
      </DialogHeader>
      <div class="py-4 space-y-2">
        <Label>Laufzeit</Label>
        <ToggleGroup v-model="ttlMinutes" type="single" variant="outline" class="justify-start">
          <ToggleGroupItem
            v-for="opt in options"
            :key="opt.value"
            :value="opt.value"
          >
            {{ opt.label }}
          </ToggleGroupItem>
        </ToggleGroup>
      </div>
      <DialogFooter>
        <Button variant="outline" @click="emit('update:open', false)">Abbrechen</Button>
        <Button :disabled="submitting" @click="handleSubmit">Verlängern</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
