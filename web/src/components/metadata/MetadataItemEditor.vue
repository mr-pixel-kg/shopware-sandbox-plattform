<script setup lang="ts">
import { Trash2 } from 'lucide-vue-next'
import { computed } from 'vue'

import MetadataActionRenderer from '@/components/metadata/MetadataActionRenderer.vue'
import MetadataDisplayRenderer from '@/components/metadata/MetadataDisplayRenderer.vue'
import MetadataFieldRenderer from '@/components/metadata/MetadataFieldRenderer.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'

import type {
  ActionSpec,
  ActionTarget,
  ActionVariant,
  DisplayFormat,
  DisplaySpec,
  FieldInput,
  FieldSpec,
  MetadataContext,
  MetadataItem,
} from '@/types'

const props = defineProps<{
  modelValue: MetadataItem
  index: number
  disabled?: boolean
  locked?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [MetadataItem]
  remove: []
}>()

const item = computed(() => props.modelValue)

function patch(partial: Partial<MetadataItem>) {
  emit('update:modelValue', { ...item.value, ...partial } as MetadataItem)
}

function patchField(partial: Partial<FieldSpec>) {
  if (item.value.type !== 'field') return
  emit('update:modelValue', {
    ...item.value,
    field: { ...item.value.field, ...partial },
  })
}

function patchAction(partial: Partial<ActionSpec>) {
  if (item.value.type !== 'action') return
  emit('update:modelValue', {
    ...item.value,
    action: { ...item.value.action, ...partial },
  })
}

function patchDisplay(partial: Partial<DisplaySpec>) {
  if (item.value.type !== 'display') return
  emit('update:modelValue', {
    ...item.value,
    display: { ...item.value.display, ...partial },
  })
}

const ALL_CONTEXTS: MetadataContext[] = [
  'image.create',
  'image.edit',
  'image.card',
  'sandbox.create',
  'sandbox.card',
  'sandbox.details',
]
const FIELD_INPUTS: FieldInput[] = [
  'text',
  'password',
  'number',
  'email',
  'url',
  'select',
  'multiselect',
  'toggle',
  'textarea',
]
const ACTION_VARIANTS: ActionVariant[] = ['default', 'outline', 'destructive']
const ACTION_TARGETS: ActionTarget[] = ['_blank', '_self']
const DISPLAY_FORMATS: DisplayFormat[] = ['text', 'code', 'badge', 'link', 'password']

const contexts = computed(() => item.value.visibility?.contexts ?? [])
function toggleContext(ctx: MetadataContext) {
  const next = new Set(contexts.value)
  if (next.has(ctx)) next.delete(ctx)
  else next.add(ctx)
  patch({
    visibility: {
      ...(item.value.visibility ?? {}),
      contexts: Array.from(next),
    },
  })
}

function changeType(next: 'field' | 'action' | 'display') {
  if (next === item.value.type) return
  const base = {
    key: item.value.key,
    label: item.value.label,
    visibility: item.value.visibility,
  }
  if (next === 'field') {
    emit('update:modelValue', {
      ...base,
      type: 'field',
      field: { input: 'text' },
    })
  } else if (next === 'action') {
    emit('update:modelValue', {
      ...base,
      type: 'action',
      action: { url: '' },
    })
  } else {
    emit('update:modelValue', {
      ...base,
      type: 'display',
      display: { value: '' },
    })
  }
}

function setOptions(raw: string) {
  if (item.value.type !== 'field') return
  const lines = raw
    .split('\n')
    .map((l) => l.trim())
    .filter(Boolean)
  const options = lines.map((line) => {
    const [value, label] = line.split('=').map((s) => s.trim())
    return { value, label: label || value }
  })
  patchField({ options })
}

const optionsAsText = computed(() => {
  if (item.value.type !== 'field' || !item.value.field.options) return ''
  return item.value.field.options.map((o) => `${o.value}=${o.label}`).join('\n')
})
</script>

<template>
  <div v-if="locked && item.type === 'field'" class="bg-muted/30 rounded-lg border p-4">
    <MetadataFieldRenderer
      :item="item"
      :model-value="item.field.default ?? ''"
      :disabled="disabled"
      @update:model-value="(v) => patchField({ default: v })"
    />
  </div>

  <div
    v-else-if="locked && item.type === 'action'"
    class="bg-muted/30 flex items-center gap-2 rounded-lg border p-4"
  >
    <Label class="text-xs font-medium">{{ item.label }}</Label>
    <MetadataActionRenderer :item="item" disabled />
  </div>

  <div v-else-if="locked && item.type === 'display'" class="bg-muted/30 rounded-lg border p-4">
    <MetadataDisplayRenderer :item="item" />
  </div>

  <div v-else class="bg-muted/30 space-y-3 rounded-lg border p-4">
    <div class="flex items-center justify-between gap-2">
      <Label class="text-sm font-medium">Item #{{ index + 1 }}</Label>
      <Button
        type="button"
        variant="ghost"
        size="icon"
        class="h-7 w-7"
        :disabled="disabled"
        @click="emit('remove')"
      >
        <Trash2 class="h-4 w-4" />
      </Button>
    </div>

    <div class="grid grid-cols-2 gap-3">
      <div class="grid gap-1.5">
        <Label class="text-xs">Typ</Label>
        <Select
          :model-value="item.type"
          :disabled="disabled"
          @update:model-value="(v) => changeType(v as 'field' | 'action' | 'display')"
        >
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="field">Feld</SelectItem>
            <SelectItem value="action">Aktion</SelectItem>
            <SelectItem value="display">Anzeige</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <div class="grid gap-1.5">
        <Label class="text-xs">Schlüssel</Label>
        <Input
          :model-value="item.key"
          placeholder="my_button"
          :disabled="disabled"
          @update:model-value="(v) => patch({ key: String(v) })"
        />
      </div>
      <div class="col-span-2 grid gap-1.5">
        <Label class="text-xs">Label</Label>
        <Input
          :model-value="item.label"
          placeholder="Sichtbarer Name"
          :disabled="disabled"
          @update:model-value="(v) => patch({ label: String(v) })"
        />
      </div>
    </div>

    <div class="grid gap-1.5">
      <Label class="text-xs">Sichtbarkeit</Label>
      <div class="flex flex-wrap gap-1.5">
        <Button
          v-for="ctx in ALL_CONTEXTS"
          :key="ctx"
          type="button"
          size="sm"
          :variant="contexts.includes(ctx) ? 'default' : 'outline'"
          :disabled="disabled"
          @click="toggleContext(ctx)"
        >
          {{ ctx }}
        </Button>
      </div>
    </div>

    <template v-if="item.type === 'field'">
      <div class="grid grid-cols-2 gap-3">
        <div class="grid gap-1.5">
          <Label class="text-xs">Input</Label>
          <Select
            :model-value="item.field.input"
            :disabled="disabled"
            @update:model-value="(v) => patchField({ input: v as FieldInput })"
          >
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem v-for="i in FIELD_INPUTS" :key="i" :value="i">{{ i }}</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div class="grid gap-1.5">
          <Label class="text-xs">Default</Label>
          <Input
            :model-value="item.field.default ?? ''"
            :disabled="disabled"
            @update:model-value="(v) => patchField({ default: String(v) })"
          />
        </div>
        <div class="col-span-2 grid gap-1.5">
          <Label class="text-xs">Hilfetext</Label>
          <Input
            :model-value="item.field.helpText ?? ''"
            :disabled="disabled"
            @update:model-value="(v) => patchField({ helpText: String(v) || undefined })"
          />
        </div>
        <div class="flex items-center gap-2">
          <Switch
            :model-value="!!item.field.required"
            :disabled="disabled"
            @update:model-value="(v) => patchField({ required: v })"
          />
          <Label class="text-xs">Pflichtfeld</Label>
        </div>
        <div
          v-if="['select', 'multiselect'].includes(item.field.input)"
          class="col-span-2 grid gap-1.5"
        >
          <Label class="text-xs">Optionen (eine pro Zeile, value=label)</Label>
          <Textarea
            :model-value="optionsAsText"
            placeholder="prod=Production"
            :disabled="disabled"
            @update:model-value="(v) => setOptions(String(v))"
          />
        </div>
      </div>
    </template>

    <template v-else-if="item.type === 'action'">
      <div class="grid gap-3">
        <div class="grid gap-1.5">
          <Label class="text-xs">URL (Go-Template)</Label>
          <Input
            :model-value="item.action.url"
            :disabled="disabled"
            @update:model-value="(v) => patchAction({ url: String(v) })"
          />
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div class="grid gap-1.5">
            <Label class="text-xs">Variant</Label>
            <Select
              :model-value="item.action.variant ?? 'default'"
              :disabled="disabled"
              @update:model-value="(v) => patchAction({ variant: v as ActionVariant })"
            >
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem v-for="v in ACTION_VARIANTS" :key="v" :value="v">{{ v }}</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div class="grid gap-1.5">
            <Label class="text-xs">Target</Label>
            <Select
              :model-value="item.action.target ?? '_blank'"
              :disabled="disabled"
              @update:model-value="(v) => patchAction({ target: v as ActionTarget })"
            >
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem v-for="t in ACTION_TARGETS" :key="t" :value="t">{{ t }}</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
        <div v-if="item.action.variant === 'destructive'" class="grid gap-1.5">
          <Label class="text-xs">Bestätigung (Pflicht für destructive)</Label>
          <Input
            :model-value="item.action.confirm ?? ''"
            :disabled="disabled"
            @update:model-value="(v) => patchAction({ confirm: String(v) })"
          />
        </div>
      </div>
    </template>

    <template v-else-if="item.type === 'display'">
      <div class="grid gap-3">
        <div class="grid gap-1.5">
          <Label class="text-xs">Wert (Go-Template oder Text)</Label>
          <Textarea
            :model-value="item.display.value"
            :disabled="disabled"
            @update:model-value="(v) => patchDisplay({ value: String(v) })"
          />
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div class="grid gap-1.5">
            <Label class="text-xs">Format</Label>
            <Select
              :model-value="item.display.format ?? 'text'"
              :disabled="disabled"
              @update:model-value="(v) => patchDisplay({ format: v as DisplayFormat })"
            >
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem v-for="f in DISPLAY_FORMATS" :key="f" :value="f">{{ f }}</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div class="flex items-center gap-2">
            <Switch
              :model-value="!!item.display.copyable"
              :disabled="disabled"
              @update:model-value="(v) => patchDisplay({ copyable: v })"
            />
            <Label class="text-xs">Kopierbar</Label>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
