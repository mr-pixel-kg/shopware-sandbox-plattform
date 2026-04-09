<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'
import { ref, watch } from 'vue'

import {
  Combobox,
  ComboboxAnchor,
  ComboboxEmpty,
  ComboboxGroup,
  ComboboxInputBase as ComboboxInput,
  ComboboxItem,
  ComboboxList,
  ComboboxViewport,
} from '@/components/ui/combobox'

export interface Suggestion {
  value: string
  label: string
  description?: string
}

const props = withDefaults(
  defineProps<{
    modelValue: string
    suggestions: Suggestion[]
    loading?: boolean
    placeholder?: string
    disabled?: boolean
    minChars?: number
    id?: string
  }>(),
  {
    loading: false,
    placeholder: '',
    disabled: false,
    minChars: 2,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const open = ref(false)
const isFocused = ref(false)

watch([() => props.modelValue, () => props.suggestions, () => props.loading], () => {
  updateOpen()
})

function updateOpen() {
  open.value =
    isFocused.value &&
    props.modelValue.length >= props.minChars &&
    (props.suggestions.length > 0 || props.loading)
}

function onFocus() {
  isFocused.value = true
  updateOpen()
}

function onBlur() {
  isFocused.value = false
  open.value = false
}

function onInput(event: Event) {
  emit('update:modelValue', (event.target as HTMLInputElement).value)
}

function onSelect(value: string) {
  emit('update:modelValue', value)
  open.value = false
}
</script>

<template>
  <Combobox
    v-model:open="open"
    :ignore-filter="true"
    :reset-search-term-on-blur="false"
    :reset-search-term-on-select="false"
  >
    <ComboboxAnchor class="w-full">
      <ComboboxInput
        :id="id"
        :value="props.modelValue"
        :placeholder="placeholder"
        :disabled="disabled"
        autocomplete="off"
        data-slot="input"
        class="file:text-foreground placeholder:text-muted-foreground selection:bg-primary selection:text-primary-foreground dark:bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive h-9 w-full min-w-0 rounded-md border bg-transparent px-3 py-1 text-base shadow-xs transition-[color,box-shadow] outline-none file:inline-flex file:h-7 file:border-0 file:bg-transparent file:text-sm file:font-medium focus-visible:ring-[3px] disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 md:text-sm"
        @input="onInput"
        @focus="onFocus"
        @blur="onBlur"
      />
    </ComboboxAnchor>

    <ComboboxList class="w-[var(--reka-combobox-trigger-width)]">
      <ComboboxViewport class="max-h-[240px]">
        <div
          v-if="loading && suggestions.length === 0"
          class="text-muted-foreground flex items-center justify-center gap-2 py-4 text-sm"
        >
          <Loader2 class="h-4 w-4 animate-spin" />
          Suche...
        </div>

        <ComboboxEmpty v-if="!loading">Keine Ergebnisse</ComboboxEmpty>

        <ComboboxGroup v-if="suggestions.length > 0">
          <ComboboxItem
            v-for="item in suggestions"
            :key="item.value"
            :value="item.value"
            class="flex-col items-start"
            @select.prevent="onSelect(item.value)"
          >
            <span>{{ item.label }}</span>
            <span v-if="item.description" class="text-muted-foreground line-clamp-1 text-xs">{{
              item.description
            }}</span>
          </ComboboxItem>
        </ComboboxGroup>
      </ComboboxViewport>
    </ComboboxList>
  </Combobox>
</template>
