<script setup lang="ts">
import { Skeleton } from '@/components/ui/skeleton'

withDefaults(
  defineProps<{
    count?: number
  }>(),
  {
    count: 6,
  },
)
</script>

<template>
  <div class="grid grid-cols-[repeat(auto-fill,minmax(240px,1fr))] gap-4">
    <div
      v-for="i in count"
      :key="i"
      class="bg-card flex h-[200px] flex-col justify-between rounded-xl border p-6 shadow-sm"
    >
      <div class="space-y-3">
        <Skeleton class="h-4 w-24" />
        <Skeleton class="h-5 w-28 rounded-full" />
      </div>
      <Skeleton class="h-4 w-full" />
      <Skeleton class="h-8 w-full rounded-md" />
    </div>
  </div>
</template>
