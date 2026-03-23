<script setup lang="ts">
import { ref, watch } from 'vue'
import { toast } from 'vue-sonner'

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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
}>()

const email = ref('')
const role = ref('developer')

watch(
  () => props.open,
  (open) => {
    if (open) {
      email.value = ''
      role.value = 'developer'
    }
  },
)

function handleSubmit() {
  // TODO: Call invite API when available
  toast.info('Benutzer einladen ist noch nicht verfügbar')
  emit('update:open', false)
}
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="sm:max-w-[400px]">
      <DialogHeader>
        <DialogTitle>Benutzer einladen</DialogTitle>
        <DialogDescription>Lade einen neuen Benutzer per E-Mail ein.</DialogDescription>
      </DialogHeader>
      <form class="grid gap-4 py-4" @submit.prevent="handleSubmit">
        <div class="grid gap-2">
          <Label for="invite-email">E-Mail</Label>
          <Input
            id="invite-email"
            v-model="email"
            type="email"
            placeholder="name@example.com"
            required
          />
        </div>
        <div class="grid gap-2">
          <Label for="invite-role">Rolle</Label>
          <Select v-model="role">
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="developer">Developer</SelectItem>
              <SelectItem value="admin">Admin</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <DialogFooter class="pt-2">
          <Button type="button" variant="outline" @click="emit('update:open', false)"
            >Abbrechen</Button
          >
          <Button type="submit">Einladen</Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
