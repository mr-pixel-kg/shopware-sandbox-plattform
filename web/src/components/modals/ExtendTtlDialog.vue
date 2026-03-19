<script setup lang="ts">
import { ref } from 'vue'
import { toast } from 'vue-sonner'
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

defineProps<{
  open: boolean
  sandboxName: string
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
}>()

const ttlMinutes = ref('60')

const options = [
  { value: '5', label: '+5 Min' },
  { value: '30', label: '+30 Min' },
  { value: '60', label: '+1 Std' },
  { value: '120', label: '+2 Std' },
  { value: '240', label: '+4 Std' },
]

function handleSubmit() {
  // TODO: Call extend API when available
  toast.info('Laufzeit verlängern ist noch nicht verfügbar')
  emit('update:open', false)
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
        <Button @click="handleSubmit">Verlängern</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
