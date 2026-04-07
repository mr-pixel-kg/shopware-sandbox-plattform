<script setup lang="ts">
import { computed } from 'vue'

import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationNext,
  PaginationPrevious,
} from '@/components/ui/pagination'

const props = withDefaults(
  defineProps<{
    page: number
    totalItems: number
    pageSize?: number
    siblingCount?: number
  }>(),
  {
    pageSize: 20,
    siblingCount: 1,
  },
)

const emit = defineEmits<{
  'update:page': [value: number]
}>()

const totalPages = computed(() => Math.max(1, Math.ceil(props.totalItems / props.pageSize)))
</script>

<template>
  <div class="flex items-center justify-between py-4">
    <p class="text-muted-foreground text-sm tabular-nums">{{ totalItems }} Eintr&auml;ge</p>
    <Pagination
      v-slot="{ page: currentPage }"
      :page="page"
      :total="Math.max(totalItems, 1)"
      :items-per-page="pageSize"
      :sibling-count="siblingCount"
      :disabled="totalPages <= 1"
      class="mx-0 w-auto justify-end"
      @update:page="emit('update:page', $event)"
    >
      <PaginationContent v-slot="{ items }">
        <PaginationPrevious />
        <template v-for="(item, index) in items" :key="index">
          <PaginationItem
            v-if="item.type === 'page'"
            :value="item.value"
            :is-active="item.value === currentPage"
          >
            {{ item.value }}
          </PaginationItem>
          <PaginationEllipsis v-else :index="index" />
        </template>
        <PaginationNext />
      </PaginationContent>
    </Pagination>
  </div>
</template>
