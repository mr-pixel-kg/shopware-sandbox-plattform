<script setup lang="ts">
import { Eye, EyeOff } from 'lucide-vue-next'
import { ref } from 'vue'

import CopyButton from '@/components/shared/CopyButton.vue'
import { Button } from '@/components/ui/button'

import type { MetadataItem } from '@/types'

defineProps<{
  items: MetadataItem[]
}>()

const revealedKeys = ref(new Set<string>())

function isSecret(item: MetadataItem): boolean {
  return item.input === 'password' || item.key.toLowerCase().includes('password')
}

function toggleReveal(key: string) {
  if (revealedKeys.value.has(key)) {
    revealedKeys.value.delete(key)
  } else {
    revealedKeys.value.add(key)
  }
}

function displayValue(item: MetadataItem): string {
  const val = item.value || '—'
  if (!isSecret(item) || revealedKeys.value.has(item.key)) return val
  return '•'.repeat(Math.min(val.length, 12))
}

defineExpose({ resetRevealed: () => revealedKeys.value.clear() })
</script>

<template>
  <table class="w-full text-sm">
    <tbody class="divide-y">
      <tr v-for="item in items" :key="item.key">
        <td class="text-muted-foreground w-1/3 py-2 pr-4 text-xs">{{ item.label }}</td>
        <td class="py-2">
          <div class="flex items-center gap-1">
            <span>{{ displayValue(item) }}</span>
            <Button
              v-if="isSecret(item) && item.value"
              variant="ghost"
              size="icon-sm"
              class="h-6 w-6"
              @click.stop="toggleReveal(item.key)"
            >
              <EyeOff v-if="revealedKeys.has(item.key)" class="h-3 w-3" />
              <Eye v-else class="h-3 w-3" />
            </Button>
            <CopyButton v-if="item.value" :value="item.value" />
          </div>
        </td>
      </tr>
    </tbody>
  </table>
</template>
