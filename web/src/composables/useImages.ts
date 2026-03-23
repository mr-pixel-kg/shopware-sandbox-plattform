import { storeToRefs } from 'pinia'
import { onMounted, onUnmounted } from 'vue'

import { useAuthStore } from '@/stores/auth.store'
import { type FetchMode, useImagesStore } from '@/stores/images.store'

export function useImages(mode: FetchMode = 'public') {
  const store = useImagesStore()
  const authStore = useAuthStore()
  const { images, pendingPulls, publicImages, loading, error } = storeToRefs(store)

  const effectiveMode = mode === 'all' && authStore.isAuthenticated ? 'all' : 'public'

  onMounted(() => {
    store.fetchMode = effectiveMode
    store.fetchImages()
    if (authStore.isAuthenticated) {
      store.initPendingPulls()
    }
  })

  onUnmounted(() => {
    store.closeAllSse()
  })

  return {
    images,
    pendingPulls,
    publicImages,
    loading,
    error,
    refresh: () => store.fetchImages(),
    createImage: store.createImage,
    deleteImage: store.deleteImage,
  }
}
