import { storeToRefs } from 'pinia'
import { onMounted, onUnmounted, ref } from 'vue'

import { imagesApi } from '@/api'
import { useAuthStore } from '@/stores/auth.store'
import { type FetchMode, useImagesStore } from '@/stores/images.store'

import type { CreateImageRequest, Image, PendingPull, UpdateImageRequest } from '@/types'

export function useImages(mode: FetchMode = 'public') {
  const store = useImagesStore()
  const authStore = useAuthStore()
  const { images, pendingPulls, publicImages, loading, error } = storeToRefs(store)

  const busyIds = ref(new Set<string>())
  const sseConnections = new Map<string, EventSource>()
  const effectiveMode = mode === 'all' && authStore.isAuthenticated ? 'all' : 'public'

  function subscribePull(id: string) {
    if (sseConnections.has(id)) return

    const baseURL = import.meta.env.WEB_API_URL || ''
    const es = new EventSource(`${baseURL}/api/images/${id}/progress`)
    sseConnections.set(id, es)

    es.onmessage = (event) => {
      let data: { percent?: number; status?: string; error?: string }
      try {
        data = JSON.parse(event.data)
      } catch {
        return
      }

      if (data.status === 'ready' || data.status === 'complete') {
        closeSse(id)
        removePendingPull(id)
        const img = images.value.find((i) => i.id === id)
        if (img) {
          img.status = 'ready'
          img.error = undefined
        }
        void fetchImages()
        return
      }

      if (data.status === 'failed') {
        closeSse(id)
        removePendingPull(id)
        const img = images.value.find((i) => i.id === id)
        if (img) {
          img.status = 'failed'
          img.error = data.error
        }
        return
      }

      const pendingItem = pendingPulls.value.find((p) => p.id === id)
      if (pendingItem && data.percent !== undefined) {
        pendingItem.percent = Math.max(pendingItem.percent, data.percent)
      }
    }

    es.onerror = () => {
      closeSse(id)
      removePendingPull(id)
    }
  }

  function closeSse(id: string) {
    const es = sseConnections.get(id)
    if (es) {
      es.close()
      sseConnections.delete(id)
    }
  }

  function closeAllSse() {
    for (const [, es] of sseConnections) es.close()
    sseConnections.clear()
  }

  function removePendingPull(id: string) {
    pendingPulls.value = pendingPulls.value.filter((p) => p.id !== id)
  }

  async function fetchImages() {
    if (!store.initialized) loading.value = true
    error.value = null
    try {
      images.value =
        effectiveMode === 'all' ? await imagesApi.listAll() : await imagesApi.listPublic()
      store.initialized = true
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : 'Fehler beim Laden'
    } finally {
      loading.value = false
    }
  }

  async function initPendingPulls() {
    const pulls = await imagesApi.listPulls().catch(() => [] as PendingPull[])
    pendingPulls.value = pulls
    for (const pull of pulls) subscribePull(pull.id)
  }

  async function createImage(req: CreateImageRequest): Promise<Image> {
    const image = await imagesApi.create(req)
    images.value.unshift(image)

    if (image.status === 'pulling') {
      pendingPulls.value.unshift({
        id: image.id,
        name: image.name,
        tag: image.tag,
        title: image.title,
        percent: 0,
        status: 'pulling',
      } as PendingPull)
      subscribePull(image.id)
    }

    return image
  }

  async function updateImage(id: string, req: UpdateImageRequest): Promise<Image> {
    const updated = await imagesApi.update(id, req)
    const idx = images.value.findIndex((i) => i.id === id)
    if (idx !== -1) images.value[idx] = updated
    return updated
  }

  async function uploadThumbnail(id: string, file: File): Promise<Image> {
    const updated = await imagesApi.uploadThumbnail(id, file)
    const idx = images.value.findIndex((i) => i.id === id)
    if (idx !== -1) images.value[idx] = updated
    return updated
  }

  async function deleteThumbnail(id: string): Promise<void> {
    await imagesApi.deleteThumbnail(id)
    const idx = images.value.findIndex((i) => i.id === id)
    if (idx !== -1) images.value[idx] = { ...images.value[idx], thumbnailUrl: undefined }
  }

  async function deleteImage(id: string) {
    closeSse(id)
    await imagesApi.remove(id)
    images.value = images.value.filter((i) => i.id !== id)
    pendingPulls.value = pendingPulls.value.filter((p) => p.id !== id)
  }

  onMounted(() => {
    store.fetchMode = effectiveMode
    void fetchImages()
    if (authStore.isAuthenticated) void initPendingPulls()
  })

  onUnmounted(() => {
    closeAllSse()
  })

  return {
    images,
    pendingPulls,
    publicImages,
    loading,
    error,
    busyIds,
    refresh: fetchImages,
    createImage,
    updateImage,
    uploadThumbnail,
    deleteThumbnail,
    deleteImage,
  }
}
