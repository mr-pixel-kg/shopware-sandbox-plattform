import { onMounted, onUnmounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useImagesStore, type FetchMode } from '@/stores/images.store'
import { useAuthStore } from '@/stores/auth.store'

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
