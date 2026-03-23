<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'

import { Button } from '@/components/ui/button'

import type { Component } from 'vue'

// TODO: Replace with dynamic schema from API
export interface CardAction {
  label: string
  href?: string
  onClick?: () => void
  variant?: 'default' | 'destructive' | 'outline' | 'ghost'
  icon?: Component
  loading?: boolean
  disabled?: boolean
}

defineProps<{
  action: CardAction
}>()
</script>

<template>
  <Button
    v-if="action.href"
    size="sm"
    :variant="action.variant ?? 'outline'"
    class="flex-1"
    :disabled="action.loading || action.disabled"
    as="a"
    :href="action.href"
    target="_blank"
  >
    <Loader2 v-if="action.loading" class="mr-1 h-4 w-4 animate-spin" />
    <component :is="action.icon" v-else-if="action.icon" class="mr-1 h-4 w-4" />
    {{ action.label }}
  </Button>
  <Button
    v-else
    size="sm"
    :variant="action.variant ?? 'outline'"
    class="flex-1"
    :disabled="action.loading || action.disabled"
    @click="action.onClick?.()"
  >
    <Loader2 v-if="action.loading" class="mr-1 h-4 w-4 animate-spin" />
    <component :is="action.icon" v-else-if="action.icon" class="mr-1 h-4 w-4" />
    {{ action.label }}
  </Button>
</template>
