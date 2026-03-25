<script setup lang="ts">
import { Card, CardContent, CardFooter, CardHeader } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

withDefaults(
  defineProps<{
    count?: number
    variant?: 'sandbox' | 'preset'
  }>(),
  {
    count: 6,
    variant: 'preset',
  },
)
</script>

<template>
  <div class="grid grid-cols-[repeat(auto-fill,minmax(320px,1fr))] gap-4">
    <Card
      v-for="i in count"
      :key="i"
      class="flex flex-col overflow-hidden pt-0"
      :class="variant === 'sandbox' ? 'h-[460px]' : ''"
    >
      <div class="bg-muted flex h-36 shrink-0 items-center justify-center">
        <Skeleton class="h-10 w-10 rounded-md" />
      </div>

      <CardHeader>
        <div class="flex items-start justify-between gap-2">
          <Skeleton class="h-4 w-32" />
          <Skeleton v-if="variant === 'sandbox'" class="h-5 w-16 rounded-full" />
        </div>
        <Skeleton v-if="variant === 'preset'" class="mt-1 h-3 w-24" />
      </CardHeader>

      <CardContent class="flex-1 space-y-3">
        <template v-if="variant === 'sandbox'">
          <Skeleton class="h-5 w-full rounded-full" />
          <div class="space-y-1.5 rounded-md border p-2.5">
            <Skeleton class="h-3 w-12" />
            <div class="flex justify-between">
              <Skeleton class="h-3 w-20" />
              <Skeleton class="h-3 w-16" />
            </div>
            <div class="flex justify-between">
              <Skeleton class="h-3 w-16" />
              <Skeleton class="h-3 w-12" />
            </div>
          </div>
        </template>
        <template v-else>
          <Skeleton class="h-3 w-full" />
          <Skeleton class="h-3 w-3/4" />
        </template>
      </CardContent>

      <CardFooter class="flex gap-2">
        <Skeleton class="h-8 w-24 rounded-md" />
        <Skeleton class="h-8 w-8 rounded-md" />
      </CardFooter>
    </Card>
  </div>
</template>
