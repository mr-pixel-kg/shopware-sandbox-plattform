import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { imagesApi } from '@/api'
import type { Image, CreateImageRequest, CreateImageResult, PendingPull } from '@/types'

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

  function subscribePull(pull: PendingPull) {
    if (sseConnections.has(pull.id)) return

    const baseURL = import.meta.env.WEB_API_URL || ''
    const es = new EventSource(`${baseURL}/api/images/${pull.id}/progress`)
    sseConnections.set(pull.id, es)

    es.onmessage = (event) => {
      // TODO i hate empty catch blocks its not a nice pattern please refactor me
      try {
        const data = JSON.parse(event.data) as { percent?: number; status?: string }
        const pending = pendingPulls.value.find((p) => p.id === pull.id)

        if (data.status === 'ready' || data.status === 'complete') {
          removePendingPull(pull.id)
          closeSse(pull.id)
          fetchImages()
          return
        }

        if (data.status === 'failed') {
          removePendingPull(pull.id)
          closeSse(pull.id)
          return
        }

        if (pending && data.percent !== undefined) {
          pending.percent = Math.max(pending.percent, data.percent)
        }
      } catch {
        // ignore parse errors
      }
    }

    es.onerror = () => {
      closeSse(pull.id)
      removePendingPull(pull.id)
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
  // TODO i hate empty catch blocks its not a nice pattern please refactor me
  async function initPendingPulls() {
    try {
      const pulls = await imagesApi.listPulls()
      pendingPulls.value = pulls
      for (const pull of pulls) {
        subscribePull(pull)
      }
    } catch {
      // silently ignore
    }
  }

  async function createImage(req: CreateImageRequest): Promise<CreateImageResult> {
    const result = await imagesApi.create(req)
    if (result.image) {
      images.value.unshift(result.image)
    }
    if (result.pendingPull) {
      pendingPulls.value.unshift(result.pendingPull)
      subscribePull(result.pendingPull)
    }
    return result
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
    deleteImage,
    $reset,
  }
})
