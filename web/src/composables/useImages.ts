import { storeToRefs } from 'pinia'
import { onMounted, onUnmounted, ref } from 'vue'

import { imagesApi } from '@/api'
import { useAuthStore } from '@/stores/auth.store'
import { type FetchMode, useImagesStore } from '@/stores/images.store'

import type { CreateImageRequest, Image, PendingImage, UpdateImageRequest } from '@/types'

export function useImages(mode: FetchMode = 'public') {
  const store = useImagesStore()
  const authStore = useAuthStore()
  const { images, pendingImages, publicImages, loading, error } = storeToRefs(store)

  const busyIds = ref(new Set<string>())
  const sseConnections = new Map<string, EventSource>()
  const effectiveMode = mode === 'all' && authStore.isAuthenticated ? 'all' : 'public'
  let pendingPollInterval: ReturnType<typeof setInterval> | null = null

  function subscribeProgress(id: string) {
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
        removePendingImage(id)
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
        removePendingImage(id)
        const img = images.value.find((i) => i.id === id)
        if (img) {
          img.status = 'failed'
          img.error = data.error
        }
        return
      }

      const pendingItem = pendingImages.value.find((p) => p.id === id)
      if (pendingItem && data.percent !== undefined) {
        pendingItem.percent = Math.max(pendingItem.percent, data.percent)
      }
    }

    es.onerror = () => {
      closeSse(id)
      removePendingImage(id)
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

  function removePendingImage(id: string) {
    pendingImages.value = pendingImages.value.filter((p) => p.id !== id)
  }

  function addPendingImage(image: Image) {
    pendingImages.value.unshift({
      id: image.id,
      name: image.name,
      tag: image.tag,
      title: image.title,
      percent: 0,
      status: image.status,
    })
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

  async function initPendingImages() {
    const pending = await imagesApi.listPending().catch(() => [] as PendingImage[])
    pendingImages.value = pending
    for (const item of pending) {
      if (item.status === 'pulling') {
        subscribeProgress(item.id)
      }
    }
    if (pending.some((p) => p.status !== 'pulling')) {
      startPendingPoll()
    }
  }

  function startPendingPoll() {
    if (pendingPollInterval) return
    pendingPollInterval = setInterval(pollPendingImages, 5_000)
  }

  function stopPendingPoll() {
    if (pendingPollInterval) {
      clearInterval(pendingPollInterval)
      pendingPollInterval = null
    }
  }

  async function pollPendingImages() {
    const pending = await imagesApi.listPending().catch(() => null)
    if (!pending) return

    const pendingIds = new Set(pending.map((p) => p.id))
    const finished = pendingImages.value.filter(
      (p) => p.status !== 'pulling' && !pendingIds.has(p.id),
    )

    if (finished.length > 0) {
      for (const item of finished) removePendingImage(item.id)
      void fetchImages()
    }

    if (!pendingImages.value.some((p) => p.status !== 'pulling')) {
      stopPendingPoll()
    }
  }

  async function createImage(req: CreateImageRequest): Promise<Image> {
    const image = await imagesApi.create(req)
    images.value.unshift(image)

    if (image.status === 'pulling') {
      addPendingImage(image)
      subscribeProgress(image.id)
    }

    return image
  }

  function trackPendingImage(image: Image) {
    images.value.unshift(image)
    addPendingImage(image)
    startPendingPoll()
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
    pendingImages.value = pendingImages.value.filter((p) => p.id !== id)
  }

  onMounted(() => {
    store.fetchMode = effectiveMode
    void fetchImages()
    if (authStore.isAuthenticated) void initPendingImages()
  })

  onUnmounted(() => {
    closeAllSse()
    stopPendingPoll()
  })

  return {
    images,
    pendingImages,
    publicImages,
    loading,
    error,
    busyIds,
    refresh: fetchImages,
    createImage,
    trackPendingImage,
    updateImage,
    uploadThumbnail,
    deleteThumbnail,
    deleteImage,
  }
}
