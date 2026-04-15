<script setup lang="ts">
import { computed, ref } from 'vue'

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'

import type { ActionItem } from '@/types'

const props = defineProps<{
  item: ActionItem
  disabled?: boolean
}>()

const variant = computed(() => props.item.action.variant ?? 'default')
const size = computed(() => (props.item.action.size === 'icon' ? 'icon' : 'default'))
const target = computed(() => props.item.action.target ?? '_blank')
const isDestructive = computed(() => variant.value === 'destructive')

const confirmOpen = ref(false)

function proceed() {
  const { url } = props.item.action
  if (target.value === '_self') {
    window.location.href = url
  } else {
    window.open(url, '_blank', 'noopener,noreferrer')
  }
  confirmOpen.value = false
}

function onClick(e: MouseEvent) {
  if (isDestructive.value) {
    e.preventDefault()
    confirmOpen.value = true
  }
}
</script>

<template>
  <template v-if="isDestructive">
    <Button type="button" :variant="variant" :size="size" :disabled="disabled" @click="onClick">
      {{ item.label }}
    </Button>
    <AlertDialog v-model:open="confirmOpen">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{{ item.label }}</AlertDialogTitle>
          <AlertDialogDescription>
            {{ item.action.confirm }}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Abbrechen</AlertDialogCancel>
          <AlertDialogAction @click="proceed">Fortfahren</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </template>

  <Button
    v-else
    as="a"
    :href="item.action.url"
    :target="target"
    :rel="target === '_blank' ? 'noopener noreferrer' : undefined"
    :variant="variant"
    :size="size"
    :disabled="disabled"
  >
    {{ item.label }}
  </Button>
</template>
