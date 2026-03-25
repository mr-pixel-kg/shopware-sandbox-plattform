<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'
import { computed } from 'vue'

import { Button } from '@/components/ui/button'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'

import type { Component } from 'vue'

export interface CardAction {
  label: string
  href?: string
  onClick?: () => void
  variant?: 'default' | 'destructive' | 'outline' | 'ghost'
  icon?: Component
  loading?: boolean
  disabled?: boolean
  tooltip?: string
  size?: 'default' | 'icon'
}

const props = defineProps<{
  action: CardAction
}>()

const showTooltip = computed(() => !!(props.action.tooltip || props.action.size === 'icon'))
const tooltipText = computed(() => props.action.tooltip || props.action.label)
const isIconOnly = computed(() => props.action.size === 'icon')
const btnSize = computed(() => (isIconOnly.value ? ('icon-sm' as const) : ('sm' as const)))
const isLink = computed(
  () => !!(props.action.href && !props.action.loading && !props.action.disabled),
)
const isDisabled = computed(() => !!(props.action.loading || props.action.disabled))
</script>

<template>
  <TooltipProvider v-if="showTooltip" :delay-duration="200">
    <Tooltip>
      <TooltipTrigger as-child>
        <span class="inline-flex">
          <Button
            v-if="isLink"
            :size="btnSize"
            :variant="action.variant ?? 'outline'"
            as="a"
            :href="action.href"
            target="_blank"
          >
            <component
              :is="action.icon"
              v-if="action.icon"
              class="h-4 w-4"
              :class="{ 'mr-1': !isIconOnly }"
            />
            <span v-if="!isIconOnly">{{ action.label }}</span>
          </Button>
          <Button
            v-else
            :size="btnSize"
            :variant="action.variant ?? 'outline'"
            :disabled="isDisabled"
            @click="action.onClick?.()"
          >
            <Loader2
              v-if="action.loading"
              class="h-4 w-4 animate-spin"
              :class="{ 'mr-1': !isIconOnly }"
            />
            <component
              :is="action.icon"
              v-else-if="action.icon"
              class="h-4 w-4"
              :class="{ 'mr-1': !isIconOnly }"
            />
            <span v-if="!isIconOnly">{{ action.label }}</span>
          </Button>
        </span>
      </TooltipTrigger>
      <TooltipContent side="top">
        <p>{{ tooltipText }}</p>
      </TooltipContent>
    </Tooltip>
  </TooltipProvider>

  <Button
    v-else-if="isLink"
    size="sm"
    :variant="action.variant ?? 'outline'"
    as="a"
    :href="action.href"
    target="_blank"
  >
    <component :is="action.icon" v-if="action.icon" class="mr-1 h-4 w-4" />
    {{ action.label }}
  </Button>
  <Button
    v-else
    size="sm"
    :variant="action.variant ?? 'outline'"
    :disabled="isDisabled"
    @click="action.onClick?.()"
  >
    <Loader2 v-if="action.loading" class="mr-1 h-4 w-4 animate-spin" />
    <component :is="action.icon" v-else-if="action.icon" class="mr-1 h-4 w-4" />
    {{ action.label }}
  </Button>
</template>
