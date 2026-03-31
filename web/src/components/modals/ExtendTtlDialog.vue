<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'
import { ref } from 'vue'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'

const props = defineProps<{
  open: boolean
  sandboxId: string
  sandboxName: string
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  submit: [payload: { sandboxId: string; ttlMinutes: number }, done: (success: boolean) => void]
}>()

const ttlMinutes = ref('60')
const submitting = ref(false)

const options = [
  { value: '5', label: '+5 Min' },
  { value: '30', label: '+30 Min' },
  { value: '60', label: '+1 Std' },
  { value: '120', label: '+2 Std' },
  { value: '240', label: '+4 Std' },
  { value: '1440', label: '+24 Std' },
  { value: 'unlimited', label: 'Unbegrenzt' },
]

function handleSubmit() {
  if (!props.sandboxId || submitting.value) return
  submitting.value = true
  emit(
    'submit',
    {
      sandboxId: props.sandboxId,
      ttlMinutes: ttlMinutes.value === 'unlimited' ? 0 : Number(ttlMinutes.value),
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
    <DialogContent class="sm:max-w-100">
      <DialogHeader>
        <DialogTitle>Laufzeit verlängern</DialogTitle>
        <DialogDescription>
          Verlängere die Laufzeit von <strong>{{ sandboxName }}</strong
          >.
        </DialogDescription>
      </DialogHeader>
      <div class="grid gap-2 overflow-hidden py-4">
        <Label>Laufzeit</Label>
        <div class="overflow-x-auto">
          <ToggleGroup v-model="ttlMinutes" type="single" variant="outline" class="w-max">
            <ToggleGroupItem v-for="opt in options" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </ToggleGroupItem>
          </ToggleGroup>
        </div>
      </div>
      <DialogFooter>
        <Button variant="outline" :disabled="submitting" @click="emit('update:open', false)"
          >Abbrechen</Button
        >
        <Button :disabled="submitting" @click="handleSubmit">
          <Loader2 v-if="submitting" class="mr-1 h-4 w-4 animate-spin" />
          Verlängern
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
