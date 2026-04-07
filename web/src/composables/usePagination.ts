import { computed, ref, watch } from 'vue'

import type { PaginationMeta, PaginationParams } from '@/types'
import type { ComputedRef, Ref, WatchSource } from 'vue'

export interface UsePaginationOptions<T = unknown> {
  pageSize?: number
  source?: Ref<T[]> | ComputedRef<T[]>
  watchResetSources?: WatchSource[]
}

export function usePagination<T = unknown>(options: UsePaginationOptions<T> = {}) {
  const pageSize = options.pageSize ?? 20
  const page = ref(1)
  const totalItems = ref(0)

  const totalPages = computed(() => Math.max(1, Math.ceil(totalItems.value / pageSize)))

  const paginationParams = computed<PaginationParams>(() => ({
    limit: pageSize,
    offset: (page.value - 1) * pageSize,
  }))

  const paginatedItems = options.source
    ? computed(() => {
        const start = (page.value - 1) * pageSize
        return options.source!.value.slice(start, start + pageSize)
      })
    : undefined

  function updateFromMeta(meta: PaginationMeta) {
    totalItems.value = meta.total
  }

  if (options.source) {
    watch(
      options.source,
      (arr) => {
        totalItems.value = arr.length
      },
      { immediate: true },
    )
  }

  if (options.watchResetSources?.length) {
    watch(options.watchResetSources, () => {
      page.value = 1
    })
  }

  watch(totalPages, (newTotal) => {
    if (page.value > newTotal) {
      page.value = newTotal
    }
  })

  return {
    page,
    pageSize,
    totalPages,
    totalItems,
    paginationParams,
    paginatedItems: paginatedItems as T extends unknown
      ? ComputedRef<T[]> | undefined
      : ComputedRef<T[]>,
    updateFromMeta,
  }
}
