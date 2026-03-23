<script setup lang="ts">
import { computed } from 'vue'

import { Avatar, AvatarFallback } from '@/components/ui/avatar'

const props = defineProps<{
  email: string
  class?: string
}>()

const initials = computed(() => {
  const parts = props.email.split('@')[0].split(/[._-]/)
  return parts
    .slice(0, 2)
    .map((p) => p[0]?.toUpperCase() ?? '')
    .join('')
})
</script>

<template>
  <Avatar :class="props.class">
    <AvatarFallback>{{ initials }}</AvatarFallback>
  </Avatar>
</template>
