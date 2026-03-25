<script setup lang="ts">
import { Check, ChevronsUpDown } from 'lucide-vue-next'
import { computed, ref } from 'vue'

import { Button } from '@/components/ui/button'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { getIconNames, resolveIcon } from '@/utils/icons'

defineProps<{
  disabled?: boolean
}>()

const model = defineModel<string>({ default: '' })

const open = ref(false)
const search = ref('')
const allIcons = getIconNames()

const filtered = computed(() => {
  const q = search.value.toLowerCase()
  if (!q) return allIcons.slice(0, 50)
  return allIcons.filter((n) => n.includes(q)).slice(0, 50)
})

const selectedIcon = computed(() => resolveIcon(model.value))

function select(name: string) {
  model.value = name === model.value ? '' : name
  open.value = false
  search.value = ''
}
</script>

<template>
  <Popover v-model:open="open">
    <PopoverTrigger as-child>
      <Button
        variant="outline"
        role="combobox"
        :aria-expanded="open"
        size="icon-sm"
        class="h-7 w-7 shrink-0"
        :disabled="disabled"
        type="button"
      >
        <component :is="selectedIcon" v-if="selectedIcon" class="h-3.5 w-3.5" />
        <ChevronsUpDown v-else class="h-3 w-3 opacity-50" />
      </Button>
    </PopoverTrigger>
    <PopoverContent class="w-56 p-0" align="start">
      <Command>
        <CommandInput v-model="search" placeholder="Icon suchen..." class="h-8 text-xs" />
        <CommandList>
          <CommandEmpty class="py-3 text-center text-xs">Kein Icon gefunden</CommandEmpty>
          <CommandGroup class="max-h-48 overflow-y-auto">
            <CommandItem
              v-for="name in filtered"
              :key="name"
              :value="name"
              class="text-[11px]"
              @select="select(name)"
            >
              <component :is="resolveIcon(name)" class="mr-2 h-3.5 w-3.5 shrink-0" />
              <span class="truncate">{{ name }}</span>
              <Check v-if="model === name" class="ml-auto h-3.5 w-3.5 shrink-0" />
            </CommandItem>
          </CommandGroup>
        </CommandList>
      </Command>
    </PopoverContent>
  </Popover>
</template>
