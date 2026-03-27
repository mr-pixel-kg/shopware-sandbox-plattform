<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'
import { ref, watch } from 'vue'

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
import { buttonVariants } from '@/components/ui/button'

const props = withDefaults(
  defineProps<{
    open: boolean
    title: string
    description: string
    confirmLabel?: string
    variant?: 'destructive' | 'default'
  }>(),
  {
    confirmLabel: 'Bestätigen',
    variant: 'destructive',
  },
)

const emit = defineEmits<{
  'update:open': [value: boolean]
  confirm: [done: (success: boolean) => void]
}>()

const confirming = ref(false)

watch(
  () => props.open,
  (open) => {
    if (open) confirming.value = false
  },
)

function handleOpenChange(value: boolean) {
  if (confirming.value) return
  emit('update:open', value)
}

function handleConfirm() {
  if (confirming.value) return
  confirming.value = true
  emit('confirm', (success: boolean) => {
    confirming.value = false
    if (success) {
      emit('update:open', false)
    }
  })
}
</script>

<template>
  <AlertDialog :open="open" @update:open="handleOpenChange">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>{{ title }}</AlertDialogTitle>
        <AlertDialogDescription>{{ description }}</AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel :disabled="confirming" @click.prevent="handleOpenChange(false)">
          Abbrechen
        </AlertDialogCancel>
        <AlertDialogAction
          :class="buttonVariants({ variant })"
          :disabled="confirming"
          @click.prevent="handleConfirm"
        >
          <Loader2 v-if="confirming" class="mr-1 h-4 w-4 animate-spin" />
          {{ confirmLabel }}
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>
