import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import type { Image, PendingImage } from '@/types'

export type FetchMode = 'public' | 'all'

export const useImagesStore = defineStore('images', () => {
  const images = ref<Image[]>([])
  const pendingImages = ref<PendingImage[]>([])
  const loading = ref(false)
  const initialized = ref(false)
  const error = ref<string | null>(null)
  const fetchMode = ref<FetchMode>('public')

  const publicImages = computed(() => images.value.filter((i) => i.isPublic))

  function $reset() {
    images.value = []
    pendingImages.value = []
    loading.value = false
    initialized.value = false
    error.value = null
  }

  function upsertImage(image: Image) {
    const idx = images.value.findIndex((i) => i.id === image.id)
    if (idx === -1) images.value.push(image)
    else images.value[idx] = image
  }

  return {
    images,
    pendingImages,
    loading,
    initialized,
    error,
    fetchMode,
    publicImages,
    $reset,
    upsertImage,
  }
})
