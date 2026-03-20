<script setup lang="ts">
import { useTtlCountdown } from '@/composables/useTtlCountdown'
import { Progress } from '@/components/ui/progress'

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
  <div v-if="expiresAt" class="flex items-center gap-2 min-w-[140px]">
    <div class="flex flex-col gap-1 flex-1">
      <span
        class="text-xs font-mono"
        :class="{
          'text-muted-foreground': isExpired,
          'text-yellow-600': isWarning && !isExpired,
          'text-foreground': !isExpired && !isWarning,
        }"
      >
        {{ isExpired ? 'abgelaufen' : remainingFormatted }}
      </span>
      <Progress
        :model-value="isExpired ? 0 : progressPercent"
        class="h-1"
      />
    </div>
  </div>
</template>
