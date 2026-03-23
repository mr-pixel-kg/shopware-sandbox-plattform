<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>

<template>
  <div class="flex flex-col items-center justify-center py-12 text-center">
    <p class="text-muted-foreground text-sm font-medium">{{ title }}</p>
    <p v-if="description" class="text-muted-foreground/70 mt-1 text-sm">{{ description }}</p>
    <div class="mt-4">
      <slot name="action" />
    </div>
  </div>
</template>
