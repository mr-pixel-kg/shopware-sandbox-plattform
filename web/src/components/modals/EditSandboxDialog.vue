<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'
import { ref, watch } from 'vue'

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

import type { Sandbox } from '@/types'

const props = defineProps<{
  open: boolean
  sandbox: Sandbox | null
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  submit: [payload: { id: string; displayName: string }, done: (success: boolean) => void]
}>()

const displayName = ref('')
const busy = ref(false)

watch(
  () => props.open,
  (open) => {
    if (open && props.sandbox) {
      displayName.value = props.sandbox.displayName || ''
      busy.value = false
    }
  },
)

function handleSubmit() {
  if (!props.sandbox) return
  busy.value = true
  emit('submit', { id: props.sandbox.id, displayName: displayName.value.trim() }, (success) => {
    busy.value = false
    if (success) {
      emit('update:open', false)
    }
  })
}
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="sm:max-w-[440px]">
      <DialogHeader>
        <DialogTitle>Sandbox bearbeiten</DialogTitle>
        <DialogDescription>Passe den Anzeigenamen dieser Sandbox an.</DialogDescription>
      </DialogHeader>
      <form class="grid gap-4 py-4" @submit.prevent="handleSubmit">
        <div class="grid gap-2">
          <Label for="edit-sandbox-name">Anzeigename</Label>
          <Input
            id="edit-sandbox-name"
            v-model="displayName"
            placeholder="z.B. Checkout-Test"
            :disabled="busy"
          />
        </div>
        <DialogFooter class="pt-2">
          <Button
            type="button"
            variant="outline"
            :disabled="busy"
            @click="emit('update:open', false)"
          >
            Abbrechen
          </Button>
          <Button type="submit" :disabled="busy">
            <Loader2 v-if="busy" class="mr-1 h-4 w-4 animate-spin" />
            {{ busy ? 'Wird gespeichert...' : 'Speichern' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
