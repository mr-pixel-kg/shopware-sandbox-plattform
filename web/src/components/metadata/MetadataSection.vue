<script setup lang="ts">
import { computed } from 'vue'

import MetadataActionRenderer from '@/components/metadata/MetadataActionRenderer.vue'
import MetadataDisplayRenderer from '@/components/metadata/MetadataDisplayRenderer.vue'
import MetadataFieldRenderer from '@/components/metadata/MetadataFieldRenderer.vue'
import { Label } from '@/components/ui/label'
import {
  evaluateDependsOn,
  fieldAsDisplay,
  groupItems,
  isFieldItem,
  itemsForContext,
} from '@/utils/metadata'

import type {
  ActionItem,
  DisplayItem,
  FieldItem,
  MetadataContext,
  MetadataGroup,
  MetadataItem,
} from '@/types'

type View = 'form' | 'card'

const props = defineProps<{
  metadata: MetadataItem[] | null | undefined
  context: MetadataContext
  groups?: MetadataGroup[]
  modelValue?: Record<string, string>
  view?: View
  disabled?: boolean
  title?: string
  showHeading?: boolean
  actionsDisabled?: boolean
  hideActions?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [Record<string, string>]
}>()

const view = computed<View>(() => props.view ?? 'form')
const values = computed<Record<string, string>>(() => props.modelValue ?? {})

const visible = computed(() => itemsForContext(props.metadata, props.context))
const fields = computed(() =>
  visible.value.filter(isFieldItem).filter((i) => evaluateDependsOn(i, values.value)),
)
const actions = computed(() => visible.value.filter((i): i is ActionItem => i.type === 'action'))
const displays = computed(() => visible.value.filter((i): i is DisplayItem => i.type === 'display'))

const fieldLikeItems = computed<Array<FieldItem | DisplayItem>>(() => [
  ...displays.value,
  ...fields.value,
])

const bucketed = computed(() => groupItems(fieldLikeItems.value, props.groups))

function setValue(key: string, value: string) {
  emit('update:modelValue', { ...values.value, [key]: value })
}

function asDisplay(item: FieldItem): DisplayItem {
  return fieldAsDisplay(item, view.value === 'form') as DisplayItem
}
</script>

<template>
  <div v-if="bucketed.length || (!hideActions && actions.length)" class="space-y-4">
    <div v-if="showHeading" class="flex items-center gap-3">
      <Label class="text-sm font-medium">{{ title ?? 'Metadaten' }}</Label>
      <div class="bg-border h-px flex-1" />
    </div>

    <div
      v-for="bucket in bucketed"
      :key="bucket.group?.key ?? '_default'"
      :class="
        view === 'card'
          ? 'bg-muted/50 space-y-2 rounded-md border px-3 py-2'
          : 'bg-muted/30 space-y-4 rounded-lg border p-4'
      "
    >
      <div
        v-if="bucket.group"
        :class="
          view === 'card'
            ? 'text-muted-foreground/70 text-[11px] font-medium tracking-wider uppercase'
            : 'space-y-0.5 border-b pb-2'
        "
      >
        <template v-if="view === 'card'">{{ bucket.group.label }}</template>
        <template v-else>
          <p class="text-sm font-medium">{{ bucket.group.label }}</p>
          <p v-if="bucket.group.description" class="text-muted-foreground text-xs">
            {{ bucket.group.description }}
          </p>
        </template>
      </div>
      <div :class="view === 'card' ? '' : 'space-y-4'">
        <template v-for="item in bucket.items" :key="item.key">
          <template v-if="isFieldItem(item)">
            <MetadataFieldRenderer
              v-if="view === 'form'"
              :item="item"
              :model-value="values[item.key] ?? item.field.default ?? ''"
              :disabled="disabled"
              @update:model-value="(v) => setValue(item.key, v)"
            />
            <MetadataDisplayRenderer v-else :item="asDisplay(item)" compact />
          </template>
          <MetadataDisplayRenderer v-else :item="item as DisplayItem" :compact="view === 'card'" />
        </template>
      </div>
    </div>

    <div v-if="!hideActions && actions.length" class="flex flex-wrap gap-2">
      <MetadataActionRenderer
        v-for="item in actions"
        :key="item.key"
        :item="item"
        :disabled="actionsDisabled"
      />
    </div>
  </div>
</template>
