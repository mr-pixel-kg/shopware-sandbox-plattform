<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>

<template>
  <div class="flex flex-col items-center justify-center py-12 text-center">
    <p class="text-sm font-medium text-muted-foreground">{{ title }}</p>
    <p v-if="description" class="text-sm text-muted-foreground/70 mt-1">{{ description }}</p>
    <div class="mt-4">
      <slot name="action" />
    </div>
  </div>
</template>
