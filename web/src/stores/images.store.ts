import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import type { Image, PendingPull } from '@/types'

export type FetchMode = 'public' | 'all'

export const useImagesStore = defineStore('images', () => {
  const images = ref<Image[]>([])
  const pendingPulls = ref<PendingPull[]>([])
  const loading = ref(false)
  const initialized = ref(false)
  const error = ref<string | null>(null)
  const fetchMode = ref<FetchMode>('public')

  const publicImages = computed(() => images.value.filter((i) => i.isPublic))

  function $reset() {
    images.value = []
    pendingPulls.value = []
    loading.value = false
    initialized.value = false
    error.value = null
  }

  return {
    images,
    pendingPulls,
    loading,
    initialized,
    error,
    fetchMode,
    publicImages,
    $reset,
  }
})
