<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth.store'
import { getApiErrorMessage } from '@/utils/error'
import { toast } from 'vue-sonner'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'

const authStore = useAuthStore()
const router = useRouter()

const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const submitting = ref(false)

async function handleSubmit() {
  if (password.value !== confirmPassword.value) {
    toast.error('Passwörter stimmen nicht überein')
    return
  }
  submitting.value = true
  try {
    await authStore.register(email.value, password.value)
    router.push('/sandboxes')
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Registrierung fehlgeschlagen'))
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <Card class="w-full max-w-sm">
    <CardHeader>
      <CardTitle class="text-xl">Registrieren</CardTitle>
      <CardDescription>Erstelle ein Konto, um loszulegen.</CardDescription>
    </CardHeader>
    <CardContent>
      <form @submit.prevent="handleSubmit" class="grid gap-4">
        <div class="grid gap-2">
          <Label for="email">E-Mail</Label>
          <Input
            id="email"
            v-model="email"
            type="email"
            placeholder="name@example.com"
            required
            autocomplete="email"
          />
        </div>
        <div class="grid gap-2">
          <Label for="password">Passwort</Label>
          <Input
            id="password"
            v-model="password"
            type="password"
            required
            autocomplete="new-password"
          />
        </div>
        <div class="grid gap-2">
          <Label for="confirm-password">Passwort bestätigen</Label>
          <Input
            id="confirm-password"
            v-model="confirmPassword"
            type="password"
            required
            autocomplete="new-password"
          />
        </div>
        <Button type="submit" class="w-full" :disabled="submitting">
          {{ submitting ? 'Wird registriert...' : 'Registrieren' }}
        </Button>
        <p class="text-muted-foreground text-center text-sm">
          Bereits ein Konto?
          <RouterLink to="/login" class="hover:text-primary underline underline-offset-4">
            Anmelden
          </RouterLink>
        </p>
      </form>
    </CardContent>
  </Card>
</template>
