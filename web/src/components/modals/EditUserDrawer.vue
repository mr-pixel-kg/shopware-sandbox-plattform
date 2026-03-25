<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'
import { ref, watch } from 'vue'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'

import type { ManagedUser } from '@/types'

const props = defineProps<{
  open: boolean
  user: ManagedUser | null
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  submit: [
    payload: { id: string; email: string; role: 'admin' | 'user'; password?: string },
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
    if (open && props.user) {
      email.value = props.user.email
      role.value = props.user.role as 'admin' | 'user'
      password.value = ''
      busy.value = false
    }
  },
)

function handleSubmit() {
  if (!props.user) return

  busy.value = true
  emit(
    'submit',
    {
      id: props.user.id,
      email: email.value,
      role: role.value,
      password: password.value || undefined,
    },
    (success) => {
      busy.value = false
      if (success) {
        emit('update:open', false)
      }
    },
  )
}
</script>

<template>
  <Sheet :open="open" @update:open="emit('update:open', $event)">
    <SheetContent side="right" class="overflow-y-auto">
      <SheetHeader>
        <SheetTitle>Benutzer bearbeiten</SheetTitle>
        <SheetDescription>
          Passe E-Mail, Rolle oder optional das Passwort des Benutzers an.
        </SheetDescription>
      </SheetHeader>
      <form id="edit-user-form" class="grid gap-4 px-4" @submit.prevent="handleSubmit">
        <div class="grid gap-2">
          <Label for="edit-user-email">E-Mail</Label>
          <Input id="edit-user-email" v-model="email" type="email" :disabled="busy" required />
        </div>
        <div class="grid gap-2">
          <Label for="edit-user-role">Rolle</Label>
          <Select v-model="role" :disabled="busy">
            <SelectTrigger id="edit-user-role">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="user">User</SelectItem>
              <SelectItem value="admin">Admin</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div class="grid gap-2">
          <Label for="edit-user-password">Neues Passwort</Label>
          <Input
            id="edit-user-password"
            v-model="password"
            type="password"
            :disabled="busy"
            placeholder="Leer lassen, um es unverändert zu lassen"
          />
          <p class="text-muted-foreground text-xs">
            Wenn du hier nichts einträgst, bleibt das bestehende Passwort erhalten.
          </p>
        </div>
      </form>
      <SheetFooter>
        <Button
          type="button"
          variant="outline"
          :disabled="busy"
          @click="emit('update:open', false)"
        >
          Abbrechen
        </Button>
        <Button type="submit" form="edit-user-form" :disabled="busy">
          <Loader2 v-if="busy" class="mr-1 h-4 w-4 animate-spin" />
          {{ busy ? 'Wird gespeichert...' : 'Speichern' }}
        </Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
</template>
