<script setup lang="ts">
import { Plus } from 'lucide-vue-next'
import { computed } from 'vue'

import MetadataItemEditor from '@/components/metadata/MetadataItemEditor.vue'
import { Button } from '@/components/ui/button'

import type { MetadataItem, MetadataSchema } from '@/types'

const props = defineProps<{
  modelValue: MetadataItem[]
  registrySchema?: MetadataSchema | null
  disabled?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [MetadataItem[]]
}>()

const items = computed(() => props.modelValue ?? [])
const lockedKeys = computed(() => new Set((props.registrySchema?.items ?? []).map((it) => it.key)))

function update(index: number, item: MetadataItem) {
  const next = [...items.value]
  next[index] = item
  emit('update:modelValue', next)
}

function remove(index: number) {
  const next = [...items.value]
  next.splice(index, 1)
  emit('update:modelValue', next)
}

function add() {
  emit('update:modelValue', [
    ...items.value,
    {
      key: '',
      label: '',
      type: 'action',
      action: { url: '', target: '_blank' },
      visibility: { contexts: ['sandbox.card', 'sandbox.details'] },
    },
  ])
}
</script>

<template>
  <div class="space-y-3">
    <MetadataItemEditor
      v-for="(item, idx) in items"
      :key="idx"
      :model-value="item"
      :index="idx"
      :disabled="disabled"
      :locked="lockedKeys.has(item.key)"
      @update:model-value="(v) => update(idx, v)"
      @remove="remove(idx)"
    />
    <Button type="button" variant="outline" size="sm" :disabled="disabled" @click="add">
      <Plus class="mr-1 h-4 w-4" />
      Item hinzufügen
    </Button>
  </div>
</template>
