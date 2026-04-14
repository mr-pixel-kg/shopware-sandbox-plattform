<script setup lang="ts">
import { Check, Copy, Eye, EyeOff } from 'lucide-vue-next'
import { computed, ref } from 'vue'
import { toast } from 'vue-sonner'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { resolveIcon } from '@/utils/icons'
import { maskSecret } from '@/utils/metadata'

import type { DisplayItem } from '@/types'

const props = defineProps<{
  item: DisplayItem
  compact?: boolean
}>()

const revealed = ref(false)
const justCopied = ref(false)

const isSecret = computed(() => props.item.display.format === 'password')
const renderedValue = computed(() => {
  if (isSecret.value && !revealed.value) return maskSecret(props.item.display.value)
  return props.item.display.value
})

async function copy() {
  try {
    await navigator.clipboard.writeText(props.item.display.value)
    justCopied.value = true
    setTimeout(() => (justCopied.value = false), 1500)
  } catch {
    toast.error('Kopieren fehlgeschlagen')
  }
}
</script>

<template>
  <div v-if="compact" class="flex items-center justify-between gap-2 text-xs">
    <span class="text-muted-foreground flex items-center gap-1">
      <component :is="resolveIcon(item.icon)" v-if="item.icon" class="h-3 w-3" />
      {{ item.label }}
    </span>
    <div class="flex items-center gap-1">
      <span class="font-mono text-[11px] break-all">{{ renderedValue }}</span>
      <Button
        v-if="isSecret"
        type="button"
        variant="ghost"
        size="icon"
        class="h-5 w-5"
        :aria-label="revealed ? 'Wert verbergen' : 'Wert anzeigen'"
        :aria-pressed="revealed"
        @click="revealed = !revealed"
      >
        <EyeOff v-if="revealed" class="h-3 w-3" />
        <Eye v-else class="h-3 w-3" />
      </Button>
      <Button
        v-if="item.display.copyable"
        type="button"
        variant="ghost"
        size="icon"
        class="h-5 w-5"
        aria-label="Kopieren"
        @click="copy"
      >
        <Check v-if="justCopied" class="h-3 w-3 text-emerald-500" />
        <Copy v-else class="h-3 w-3" />
      </Button>
    </div>
  </div>

  <div v-else class="flex items-start gap-2">
    <div class="min-w-0 flex-1">
      <p class="text-muted-foreground text-xs">{{ item.label }}</p>

      <Badge v-if="item.display.format === 'badge'" variant="secondary" class="mt-1">
        {{ renderedValue }}
      </Badge>

      <a
        v-else-if="item.display.format === 'link'"
        :href="item.display.value"
        target="_blank"
        rel="noopener noreferrer"
        class="text-primary text-sm break-all hover:underline"
      >
        {{ renderedValue }}
      </a>

      <code
        v-else-if="item.display.format === 'code'"
        class="bg-muted mt-1 block rounded px-2 py-1 font-mono text-sm break-all"
      >
        {{ renderedValue }}
      </code>

      <p
        v-else
        class="text-sm break-words"
        :class="isSecret && !revealed ? 'font-mono tracking-widest' : ''"
      >
        {{ renderedValue }}
      </p>
    </div>

    <div class="flex shrink-0 items-center gap-1">
      <Button
        v-if="isSecret"
        type="button"
        variant="ghost"
        size="icon"
        class="h-7 w-7"
        :aria-label="revealed ? 'Wert verbergen' : 'Wert anzeigen'"
        :aria-pressed="revealed"
        @click="revealed = !revealed"
      >
        <EyeOff v-if="revealed" class="h-4 w-4" />
        <Eye v-else class="h-4 w-4" />
      </Button>
      <Button
        v-if="item.display.copyable"
        type="button"
        variant="ghost"
        size="icon"
        class="h-7 w-7"
        aria-label="Kopieren"
        @click="copy"
      >
        <Check v-if="justCopied" class="h-4 w-4 text-emerald-500" />
        <Copy v-else class="h-4 w-4" />
      </Button>
    </div>
  </div>
</template>
