import { storeToRefs } from 'pinia'
import { onMounted, onUnmounted, ref } from 'vue'

import { sandboxesApi } from '@/api'
import { useAuthStore } from '@/stores/auth.store'
import { useSandboxesStore } from '@/stores/sandboxes.store'
import { getToken } from '@/utils/storage'

import type {
  CreateSandboxRequest,
  CreateSnapshotRequest,
  Image,
  Sandbox,
  SandboxHealthEvent,
  SandboxStatus,
} from '@/types'

export function useSandboxes() {
  const store = useSandboxesStore()
  const authStore = useAuthStore()
  const { sandboxes, activeSandboxes, recentSandboxes, loading, error, healthBySandboxId } =
    storeToRefs(store)

  const busyIds = ref(new Set<string>())
  const healthConnections = new Map<string, EventSource>()
  let pollInterval: ReturnType<typeof setInterval> | null = null

  function buildHealthStreamUrl(id: string): string {
    const base = import.meta.env.WEB_API_URL || window.location.origin
    const url = new URL(`/api/sandboxes/${id}/health`, base)
    const token = getToken()
    if (token) url.searchParams.set('access_token', token)
    return url.toString()
  }

  function closeHealthStream(id: string) {
    const es = healthConnections.get(id)
    if (es) {
      es.close()
      healthConnections.delete(id)
    }
  }

  function closeAllHealthStreams() {
    for (const [, es] of healthConnections) es.close()
    healthConnections.clear()
  }

  function updateSandboxStatus(id: string, status: SandboxStatus) {
    const sandbox = sandboxes.value.find((s) => s.id === id)
    if (sandbox) sandbox.status = status
  }

  function applyHealthEvent(event: SandboxHealthEvent) {
    healthBySandboxId.value = { ...healthBySandboxId.value, [event.sandboxId]: event }

    if (event.ready) {
      updateSandboxStatus(event.sandboxId, 'running')
      return
    }

    if (['deleted', 'expired', 'failed', 'stopped'].includes(event.status)) {
      updateSandboxStatus(event.sandboxId, event.status as SandboxStatus)
      closeHealthStream(event.sandboxId)
    }
  }

  function subscribeHealth(id: string) {
    if (healthConnections.has(id)) return

    const es = new EventSource(buildHealthStreamUrl(id), { withCredentials: true })
    healthConnections.set(id, es)

    es.onmessage = (message) => {
      let parsed: SandboxHealthEvent
      try {
        parsed = JSON.parse(message.data)
      } catch {
        return
      }
      applyHealthEvent(parsed)
    }

    es.onerror = () => {
      const sandbox = sandboxes.value.find((s) => s.id === id)
      if (!sandbox || (sandbox.status !== 'starting' && sandbox.status !== 'running')) {
        closeHealthStream(id)
      }
    }
  }

  function syncHealthSubscriptions() {
    const activeIds = new Set(
      sandboxes.value
        .filter((s) => s.status === 'starting' || s.status === 'running')
        .map((s) => s.id),
    )

    for (const id of activeIds) subscribeHealth(id)
    for (const [id] of healthConnections) {
      if (!activeIds.has(id)) closeHealthStream(id)
    }
  }

  async function fetch() {
    if (!store.initialized) loading.value = true
    error.value = null
    try {
      sandboxes.value = authStore.isAuthenticated
        ? await sandboxesApi.listMine()
        : await sandboxesApi.listGuest()
      syncHealthSubscriptions()
      store.initialized = true
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : 'Fehler beim Laden'
    } finally {
      loading.value = false
    }
  }

  async function createSandbox(req: CreateSandboxRequest): Promise<Sandbox> {
    const sandbox = await sandboxesApi.create(req)
    sandboxes.value.unshift(sandbox)
    syncHealthSubscriptions()
    return sandbox
  }

  async function createPublicDemo(req: CreateSandboxRequest): Promise<Sandbox> {
    const sandbox = await sandboxesApi.createPublicDemo(req)
    sandboxes.value.unshift(sandbox)
    syncHealthSubscriptions()
    return sandbox
  }

  async function updateSandbox(id: string, req: { displayName?: string }): Promise<Sandbox> {
    const updated = await sandboxesApi.update(id, req)
    const idx = sandboxes.value.findIndex((s) => s.id === id)
    if (idx !== -1) sandboxes.value[idx] = updated
    return updated
  }

  async function extendTTL(id: string, ttlMinutes: number): Promise<Sandbox> {
    const updated = await sandboxesApi.extendTTL(id, ttlMinutes)
    const idx = sandboxes.value.findIndex((s) => s.id === id)
    if (idx !== -1) sandboxes.value[idx] = updated
    return updated
  }

  async function deleteSandbox(id: string, guest = false) {
    closeHealthStream(id)
    if (guest) {
      await sandboxesApi.removeGuest(id)
    } else {
      await sandboxesApi.remove(id)
    }
    sandboxes.value = sandboxes.value.filter((s) => s.id !== id)
  }

  async function snapshotSandbox(id: string, req: CreateSnapshotRequest): Promise<Image> {
    return sandboxesApi.snapshot(id, req)
  }

  const allSandboxes = ref<Sandbox[]>([])
  const allLoading = ref(false)
  let allInitialized = false
  let adminPollInterval: ReturnType<typeof setInterval> | null = null

  async function fetchAllInstances() {
    if (!allInitialized) allLoading.value = true
    try {
      allSandboxes.value = await sandboxesApi.list()
      allInitialized = true
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : 'Fehler beim Laden'
    } finally {
      allLoading.value = false
    }
  }

  function startAdminPolling() {
    void fetchAllInstances()
    adminPollInterval = setInterval(fetchAllInstances, 10_000)
  }

  onMounted(() => {
    store.$reset()
    closeAllHealthStreams()
    void fetch()
    pollInterval = setInterval(fetch, 10_000)
  })

  onUnmounted(() => {
    if (pollInterval) clearInterval(pollInterval)
    if (adminPollInterval) clearInterval(adminPollInterval)
    closeAllHealthStreams()
  })

  return {
    sandboxes,
    activeSandboxes,
    recentSandboxes,
    loading,
    error,
    healthBySandboxId,
    busyIds,
    refresh: fetch,
    createSandbox,
    createPublicDemo,
    updateSandbox,
    extendTTL,
    deleteSandbox,
    snapshotSandbox,
    allSandboxes,
    allLoading,
    startAdminPolling,
  }
}
