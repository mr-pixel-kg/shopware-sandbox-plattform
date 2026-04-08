<script setup lang="ts">
import { Avatar, AvatarImage } from '@/components/ui/avatar'

const props = defineProps<{
  src: string
  alt?: string
  class?: string
}>()
</script>

<template>
  <Avatar :class="props.class">
    <AvatarImage :src="src" :alt="alt" />
  </Avatar>
</template>
