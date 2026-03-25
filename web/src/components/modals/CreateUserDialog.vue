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
  submit: [
    payload: { email: string; role: 'admin' | 'user'; password: string },
    done: (success: boolean) => void,
  ]
}>()

const email = ref('')
const role = ref<'admin' | 'user'>('user')
const password = ref('')
const busy = ref(false)

watch(
  () => props.open,
  (open) => {
    if (open) {
      email.value = ''
      role.value = 'user'
      password.value = ''
      busy.value = false
    }
  },
)

function handleSubmit() {
  busy.value = true
  emit('submit', { email: email.value, role: role.value, password: password.value }, (success) => {
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
        <DialogTitle>Benutzer anlegen</DialogTitle>
        <DialogDescription>Erstelle direkt einen aktiven Benutzerzugang.</DialogDescription>
      </DialogHeader>
      <form class="grid gap-4 py-4" @submit.prevent="handleSubmit">
        <div class="grid gap-2">
          <Label for="create-user-email">E-Mail</Label>
          <Input
            id="create-user-email"
            v-model="email"
            type="email"
            placeholder="name@example.com"
            required
          />
        </div>
        <div class="grid gap-2">
          <Label for="create-user-role">Rolle</Label>
          <Select v-model="role">
            <SelectTrigger id="create-user-role">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="user">User</SelectItem>
              <SelectItem value="admin">Admin</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div class="grid gap-2">
          <Label for="create-user-password">Passwort</Label>
          <Input
            id="create-user-password"
            v-model="password"
            type="password"
            placeholder="Sicheres Passwort vergeben"
            required
          />
        </div>
        <DialogFooter class="pt-2">
          <Button type="button" variant="outline" @click="emit('update:open', false)">
            Abbrechen
          </Button>
          <Button type="submit" :disabled="!email || !password || busy">
            <Loader2 v-if="busy" class="mr-1 h-4 w-4 animate-spin" />
            {{ busy ? 'Wird angelegt...' : 'Benutzer anlegen' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
