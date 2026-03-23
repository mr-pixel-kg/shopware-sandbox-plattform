import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { imagesApi } from '@/api'
import type { Image, CreateImageRequest, UpdateImageRequest, PendingPull } from '@/types'

export type FetchMode = 'public' | 'all'

export const useImagesStore = defineStore('images', () => {
  const images = ref<Image[]>([])
  const pendingPulls = ref<PendingPull[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const fetchMode = ref<FetchMode>('public')

  const sseConnections = new Map<string, EventSource>()

  const publicImages = computed(() => images.value.filter((i) => i.isPublic))

  async function fetchImages() {
    const isInitial = images.value.length === 0
    if (isInitial) loading.value = true
    error.value = null
    try {
      images.value =
        fetchMode.value === 'all' ? await imagesApi.listAll() : await imagesApi.listPublic()
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : 'Fehler beim Laden'
    } finally {
      loading.value = false
    }
  }

  function subscribePull(id: string) {
    if (sseConnections.has(id)) return

    const baseURL = import.meta.env.WEB_API_URL || ''
    const es = new EventSource(`${baseURL}/api/images/${id}/progress`)
    sseConnections.set(id, es)

    es.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as { percent?: number; status?: string; error?: string }

        if (data.status === 'ready' || data.status === 'complete') {
          closeSse(id)
          removePendingPull(id)
          const img = images.value.find((i) => i.id === id)
          if (img) {
            img.status = 'ready'
            img.error = undefined
          }
          fetchImages()
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

        const pending = pendingPulls.value.find((p) => p.id === id)
        if (pending && data.percent !== undefined) {
          pending.percent = Math.max(pending.percent, data.percent)
        }
      } catch {
        // ignore parse errors
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
    for (const [id, es] of sseConnections) {
      es.close()
      sseConnections.delete(id)
    }
  }

  function removePendingPull(id: string) {
    pendingPulls.value = pendingPulls.value.filter((p) => p.id !== id)
  }
  async function initPendingPulls() {
    try {
      const pulls = await imagesApi.listPulls()
      pendingPulls.value = pulls
      for (const pull of pulls) {
        subscribePull(pull.id)
      }
    } catch {
      // silently ignore
    }
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
      })
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

  function $reset() {
    closeAllSse()
    images.value = []
    pendingPulls.value = []
    loading.value = false
    error.value = null
  }

  return {
    images,
    pendingPulls,
    loading,
    error,
    fetchMode,
    publicImages,
    fetchImages,
    initPendingPulls,
    closeAllSse,
    createImage,
    updateImage,
    uploadThumbnail,
    deleteThumbnail,
    deleteImage,
    $reset,
  }
})
