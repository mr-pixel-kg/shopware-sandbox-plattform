<script setup lang="ts">
import { Eye, EyeOff } from 'lucide-vue-next'
import { ref } from 'vue'

import CopyButton from '@/components/shared/CopyButton.vue'
import { Button } from '@/components/ui/button'

import type { SSHConnection } from '@/types'

defineProps<{
  ssh: SSHConnection
}>()

const passwordRevealed = ref(false)

function maskedPassword(pw: string): string {
  return '•'.repeat(Math.min(pw.length, 12))
}
</script>

<template>
  <div>
    <h3 class="text-sm font-medium">SSH</h3>
    <p class="text-muted-foreground mb-3 text-xs">Shell-Zugriff auf die Sandbox</p>
    <table class="w-full table-fixed text-sm">
      <tbody class="divide-y">
        <tr>
          <td class="text-muted-foreground w-1/3 py-2 pr-4 text-xs">Befehl</td>
          <td class="max-w-0 py-2">
            <div class="flex items-center gap-1">
              <span class="min-w-0 truncate">{{ ssh.command }}</span>
              <CopyButton :value="ssh.command" />
            </div>
          </td>
        </tr>
        <tr>
          <td class="text-muted-foreground w-1/3 py-2 pr-4 text-xs">Host</td>
          <td class="max-w-0 py-2">
            <div class="flex items-center gap-1">
              <span class="min-w-0 truncate">{{ ssh.host }}</span>
              <CopyButton :value="ssh.host" />
            </div>
          </td>
        </tr>
        <tr>
          <td class="text-muted-foreground w-1/3 py-2 pr-4 text-xs">Port</td>
          <td class="max-w-0 py-2">
            <div class="flex items-center gap-1">
              <span class="min-w-0 truncate">{{ ssh.port }}</span>
              <CopyButton :value="String(ssh.port)" />
            </div>
          </td>
        </tr>
        <tr>
          <td class="text-muted-foreground w-1/3 py-2 pr-4 text-xs">Benutzer</td>
          <td class="max-w-0 py-2">
            <div class="flex items-center gap-1">
              <span class="min-w-0 truncate">{{ ssh.username }}</span>
              <CopyButton :value="ssh.username" />
            </div>
          </td>
        </tr>
        <tr>
          <td class="text-muted-foreground w-1/3 py-2 pr-4 text-xs">Passwort</td>
          <td class="max-w-0 py-2">
            <div class="flex items-center gap-1">
              <span class="min-w-0 truncate">{{
                passwordRevealed ? ssh.password : maskedPassword(ssh.password)
              }}</span>
              <Button
                variant="ghost"
                size="icon-sm"
                class="h-6 w-6 shrink-0"
                @click.stop="passwordRevealed = !passwordRevealed"
              >
                <EyeOff v-if="passwordRevealed" class="h-3 w-3" />
                <Eye v-else class="text-muted-foreground h-3 w-3" />
              </Button>
              <CopyButton :value="ssh.password" />
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
