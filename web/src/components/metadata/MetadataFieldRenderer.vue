<script setup lang="ts">
import { Check, Eye, EyeOff } from 'lucide-vue-next'
import { computed, ref } from 'vue'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'

import type { FieldItem } from '@/types'

const props = defineProps<{
  item: FieldItem
  modelValue: string
  disabled?: boolean
  error?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const helpId = computed(() => `field-help-${props.item.key}`)
const showSecret = ref(false)

const stringValue = computed<string>({
  get: () => props.modelValue ?? '',
  set: (v) => emit('update:modelValue', v ?? ''),
})

const toggleValue = computed<boolean>({
  get: () => props.modelValue === 'true',
  set: (v) => emit('update:modelValue', v ? 'true' : 'false'),
})

const multiValues = computed<string[]>(() =>
  props.modelValue
    ? props.modelValue
        .split(',')
        .map((s) => s.trim())
        .filter(Boolean)
    : [],
)

function setMulti(value: string, checked: boolean) {
  const next = new Set(multiValues.value)
  if (checked) next.add(value)
  else next.delete(value)
  emit('update:modelValue', Array.from(next).join(','))
}

function isMultiChecked(value: string) {
  return multiValues.value.includes(value)
}

const readOnly = computed(() => props.item.field.readOnly === true)
const effectiveDisabled = computed(() => props.disabled || readOnly.value)
</script>

<template>
  <div class="space-y-1.5">
    <Label :for="item.key" class="flex items-center gap-1">
      {{ item.label }}
      <span v-if="item.field.required" class="text-destructive">*</span>
    </Label>

    <Input
      v-if="['text', 'email', 'url', 'number'].includes(item.field.input)"
      :id="item.key"
      v-model="stringValue"
      :type="item.field.input === 'number' ? 'number' : item.field.input"
      :placeholder="item.field.placeholder"
      :disabled="effectiveDisabled"
      :aria-describedby="item.field.helpText ? helpId : undefined"
      :required="item.field.required"
    />

    <div v-else-if="item.field.input === 'password'" class="relative">
      <Input
        :id="item.key"
        v-model="stringValue"
        :type="showSecret ? 'text' : 'password'"
        :placeholder="item.field.placeholder"
        :disabled="effectiveDisabled"
        :aria-describedby="item.field.helpText ? helpId : undefined"
        :required="item.field.required"
        class="pr-10"
      />
      <Button
        v-if="stringValue"
        type="button"
        variant="ghost"
        size="icon"
        class="absolute top-1/2 right-1 h-7 w-7 -translate-y-1/2"
        :aria-label="showSecret ? 'Wert verbergen' : 'Wert anzeigen'"
        :aria-pressed="showSecret"
        @click="showSecret = !showSecret"
      >
        <EyeOff v-if="showSecret" class="h-4 w-4" />
        <Eye v-else class="h-4 w-4" />
      </Button>
    </div>

    <Textarea
      v-else-if="item.field.input === 'textarea'"
      :id="item.key"
      v-model="stringValue"
      :placeholder="item.field.placeholder"
      :disabled="effectiveDisabled"
      :aria-describedby="item.field.helpText ? helpId : undefined"
      :required="item.field.required"
    />

    <div v-else-if="item.field.input === 'toggle'" class="flex h-9 items-center gap-2">
      <Switch
        :id="item.key"
        v-model="toggleValue"
        :disabled="effectiveDisabled"
        :aria-describedby="item.field.helpText ? helpId : undefined"
      />
      <span class="text-muted-foreground text-sm">
        {{ toggleValue ? 'Aktiv' : 'Inaktiv' }}
      </span>
    </div>

    <Select
      v-else-if="item.field.input === 'select'"
      v-model="stringValue"
      :disabled="effectiveDisabled"
    >
      <SelectTrigger
        :id="item.key"
        class="w-full"
        :aria-describedby="item.field.helpText ? helpId : undefined"
      >
        <SelectValue :placeholder="item.field.placeholder ?? 'Bitte wählen'" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem v-for="opt in item.field.options ?? []" :key="opt.value" :value="opt.value">
          {{ opt.label }}
        </SelectItem>
      </SelectContent>
    </Select>

    <Popover v-else-if="item.field.input === 'multiselect'">
      <PopoverTrigger as-child>
        <Button
          :id="item.key"
          type="button"
          variant="outline"
          class="w-full justify-between font-normal"
          :disabled="effectiveDisabled"
          :aria-describedby="item.field.helpText ? helpId : undefined"
        >
          <span class="truncate">
            <template v-if="multiValues.length"> {{ multiValues.length }} ausgewählt </template>
            <template v-else>
              <span class="text-muted-foreground">
                {{ item.field.placeholder ?? 'Auswählen' }}
              </span>
            </template>
          </span>
        </Button>
      </PopoverTrigger>
      <PopoverContent class="w-64 p-2">
        <ul class="space-y-1">
          <li
            v-for="opt in item.field.options ?? []"
            :key="opt.value"
            class="hover:bg-muted flex cursor-pointer items-center gap-2 rounded px-2 py-1"
            @click="setMulti(opt.value, !isMultiChecked(opt.value))"
          >
            <span
              class="flex h-4 w-4 items-center justify-center rounded border"
              :class="
                isMultiChecked(opt.value)
                  ? 'bg-primary border-primary text-primary-foreground'
                  : 'border-input'
              "
            >
              <Check v-if="isMultiChecked(opt.value)" class="h-3 w-3" />
            </span>
            <span class="text-sm">{{ opt.label }}</span>
          </li>
        </ul>
      </PopoverContent>
    </Popover>

    <p v-if="item.field.helpText" :id="helpId" class="text-muted-foreground text-xs">
      {{ item.field.helpText }}
    </p>
    <p v-if="error" class="text-destructive text-xs">{{ error }}</p>
  </div>
</template>
