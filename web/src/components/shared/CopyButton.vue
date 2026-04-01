<script setup lang="ts">
import { Check, Copy } from 'lucide-vue-next'
import { ref } from 'vue'

import { Button } from '@/components/ui/button'

const props = defineProps<{ value: string }>()

const copied = ref(false)
let timer: ReturnType<typeof setTimeout> | null = null

async function handleCopy() {
  await navigator.clipboard.writeText(props.value)
  copied.value = true
  if (timer) clearTimeout(timer)
  timer = setTimeout(() => (copied.value = false), 2000)
}
</script>

<template>
  <Button variant="ghost" size="icon-sm" class="h-6 w-6 shrink-0" @click.stop="handleCopy">
    <Check v-if="copied" class="h-3 w-3 text-green-600" />
    <Copy v-else class="text-muted-foreground h-3 w-3" />
  </Button>
</template>
