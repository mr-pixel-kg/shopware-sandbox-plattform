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
  submit: [payload: { email: string; role: 'admin' | 'user' }, done: (success: boolean) => void]
}>()

const email = ref('')
const role = ref<'admin' | 'user'>('user')
const busy = ref(false)

watch(
  () => props.open,
  (open) => {
    if (open) {
      email.value = ''
      role.value = 'user'
      busy.value = false
    }
  },
)

function handleSubmit() {
  busy.value = true
  emit('submit', { email: email.value, role: role.value }, (success) => {
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
            :disabled="busy"
          />
        </div>
        <div class="grid gap-2">
          <Label for="invite-role">Rolle</Label>
          <Select v-model="role">
            <SelectTrigger id="invite-role">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="user">User</SelectItem>
              <SelectItem value="admin">Admin</SelectItem>
            </SelectContent>
          </Select>
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
          <Button type="submit" :disabled="!email || busy">
            <Loader2 v-if="busy" class="mr-1 h-4 w-4 animate-spin" />
            {{ busy ? 'Wird eingeladen...' : 'Einladen' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
