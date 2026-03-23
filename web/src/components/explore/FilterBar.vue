<script setup lang="ts">
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'

defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const filters = [
  { value: 'all', label: 'Alle' },
  { value: 'public', label: 'Öffentlich' },
  { value: 'private', label: 'Privat' },
]
</script>

<template>
  <div class="flex items-center gap-3">
    <ToggleGroup
      type="single"
      variant="outline"
      :model-value="modelValue"
      @update:model-value="emit('update:modelValue', ($event as string) || 'all')"
    >
      <ToggleGroupItem v-for="f in filters" :key="f.value" :value="f.value" size="sm">
        {{ f.label }}
      </ToggleGroupItem>
    </ToggleGroup>
  </div>
</template>
