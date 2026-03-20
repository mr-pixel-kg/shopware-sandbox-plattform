<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth.store'
import { getApiErrorMessage } from '@/utils/error'
import { toast } from 'vue-sonner'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()

const email = ref('')
const password = ref('')
const submitting = ref(false)

async function handleSubmit() {
  submitting.value = true
  try {
    await authStore.login(email.value, password.value)
    const redirect = (route.query.redirect as string) || '/sandboxes'
    router.push(redirect)
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Anmeldung fehlgeschlagen'))
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <Card class="w-full max-w-sm">
    <CardHeader>
      <CardTitle class="text-xl">Anmelden</CardTitle>
      <CardDescription>Melde dich an, um Sandboxes zu verwalten.</CardDescription>
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
            autocomplete="current-password"
          />
        </div>
        <Button type="submit" class="w-full" :disabled="submitting">
          {{ submitting ? 'Wird angemeldet...' : 'Anmelden' }}
        </Button>
        <p class="text-center text-sm text-muted-foreground">
          Noch kein Konto?
          <RouterLink to="/register" class="underline underline-offset-4 hover:text-primary">
            Registrieren
          </RouterLink>
        </p>
      </form>
    </CardContent>
  </Card>
</template>
