<script setup lang="ts">
import { Progress } from '@/components/ui/progress'
import { useTtlCountdown } from '@/composables/useTtlCountdown'

const props = defineProps<{
  expiresAt?: string
  createdAt: string
}>()

const { remainingFormatted, progressPercent, isExpired, isWarning } = useTtlCountdown(
  () => props.expiresAt,
  () => props.createdAt,
)
</script>

<template>
  <div v-if="expiresAt" class="flex min-w-[140px] items-center gap-2">
    <div class="flex flex-1 flex-col gap-1">
      <span
        class="font-mono text-xs"
        :class="{
          'text-muted-foreground': isExpired,
          'text-yellow-600': isWarning && !isExpired,
          'text-foreground': !isExpired && !isWarning,
        }"
      >
        {{ isExpired ? 'abgelaufen' : remainingFormatted }}
      </span>
      <Progress :model-value="isExpired ? 0 : progressPercent" class="h-1" />
    </div>
  </div>
</template>
