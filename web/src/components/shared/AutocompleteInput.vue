<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'
import { nextTick, ref, watch } from 'vue'

import { Command, CommandGroup, CommandItem, CommandList } from '@/components/ui/command'
import { Input } from '@/components/ui/input'
import { Popover, PopoverAnchor, PopoverContent } from '@/components/ui/popover'

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

const localValue = ref(props.modelValue)
const open = ref(false)
const focused = ref(false)
const selecting = ref(false)

watch(
  () => props.modelValue,
  (val) => {
    if (val !== localValue.value) {
      localValue.value = val
    }
  },
)

watch(localValue, (val) => {
  emit('update:modelValue', val)
  updateOpen()
})

watch([() => props.suggestions, () => props.loading], () => {
  updateOpen()
})

function updateOpen() {
  if (!focused.value) return
  open.value =
    localValue.value.length >= props.minChars && (props.suggestions.length > 0 || props.loading)
}

function onFocus() {
  focused.value = true
  updateOpen()
}

function onBlur() {
  if (selecting.value) return
  focused.value = false
  open.value = false
}

function onSelect(value: string) {
  selecting.value = true
  localValue.value = value
  open.value = false
  nextTick(() => {
    selecting.value = false
    focused.value = false
  })
}
</script>

<template>
  <Popover :open="open">
    <PopoverAnchor as-child>
      <Input
        :id="id"
        v-model="localValue"
        :placeholder="placeholder"
        :disabled="disabled"
        autocomplete="off"
        @focus="onFocus"
        @blur="onBlur"
      />
    </PopoverAnchor>

    <PopoverContent
      class="p-0"
      :style="{ width: 'var(--reka-popover-trigger-width)' }"
      @open-auto-focus.prevent
      @close-auto-focus.prevent
    >
      <Command>
        <CommandList class="max-h-[240px]">
          <div
            v-if="loading && suggestions.length === 0"
            class="text-muted-foreground flex items-center justify-center gap-2 py-4 text-sm"
          >
            <Loader2 class="h-4 w-4 animate-spin" />
            Suche...
          </div>

          <div
            v-else-if="suggestions.length === 0"
            class="text-muted-foreground py-6 text-center text-sm"
          >
            Keine Ergebnisse
          </div>

          <CommandGroup v-if="suggestions.length > 0">
            <CommandItem
              v-for="item in suggestions"
              :key="item.value"
              :value="item.value"
              class="flex-col items-start"
              @pointerdown.prevent
              @select="onSelect(item.value)"
            >
              <span>{{ item.label }}</span>
              <span v-if="item.description" class="text-muted-foreground line-clamp-1 text-xs">{{
                item.description
              }}</span>
            </CommandItem>
          </CommandGroup>
        </CommandList>
      </Command>
    </PopoverContent>
  </Popover>
</template>
