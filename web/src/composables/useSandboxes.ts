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
  UpdateSandboxRequest,
} from '@/types'

const TRANSITIONAL_STATUSES = new Set<string>(['starting', 'stopping', 'paused'])
const TERMINAL_STATUSES = new Set<string>(['running', 'stopped', 'failed', 'deleted', 'expired'])

const KNOWN_SSE_STATUSES: Record<string, SandboxStatus> = {
  probing: 'starting',
  ready: 'running',
  starting: 'starting',
  running: 'running',
  paused: 'paused',
  stopping: 'stopping',
  stopped: 'stopped',
  expired: 'expired',
  deleted: 'deleted',
  failed: 'failed',
}

interface SseEvent {
  id: string
  status: string
  stateReason?: string
}

let pollInterval: ReturnType<typeof setInterval> | null = null
let pollConsumers = 0
const sseConnections = new Map<string, AbortController>()

export function useSandboxes() {
  const store = useSandboxesStore()
  const authStore = useAuthStore()
  const {
    sandboxes,
    adminSandboxes,
    activeSandboxes,
    recentSandboxes,
    allSandboxes,
    loading,
    error,
  } = storeToRefs(store)

  const busyIds = ref(new Set<string>())
  const healthBySandboxId = ref<Record<string, SandboxHealthEvent>>({})

  function subscribeSse(id: string) {
    if (sseConnections.has(id)) return

    const abort = new AbortController()
    sseConnections.set(id, abort)

    const baseURL = import.meta.env.WEB_API_URL || ''
    const headers: Record<string, string> = {}
    const token = getToken()
    if (token) headers.Authorization = `Bearer ${token}`

    fetch(`${baseURL}/api/sandboxes/${id}/stream`, {
      headers,
      credentials: 'include',
      signal: abort.signal,
    })
      .then((res) => {
        if (!res.ok || !res.body) {
          closeSse(id)
          return
        }

        const reader = res.body.getReader()
        const decoder = new TextDecoder()
        let buffer = ''

        function read(): Promise<void> {
          return reader.read().then(({ done, value }) => {
            if (done) {
              closeSse(id)
              void fetchSandboxes()
              return
            }

            buffer += decoder.decode(value, { stream: true })
            const lines = buffer.split('\n')
            buffer = lines.pop() ?? ''

            for (const line of lines) {
              if (!line.startsWith('data: ')) continue
              try {
                const data: SseEvent = JSON.parse(line.slice(6))
                const mapped = KNOWN_SSE_STATUSES[data.status] ?? 'failed'
                const sandbox = sandboxes.value.find((s) => s.id === id)
                if (sandbox) {
                  sandbox.status = mapped
                  sandbox.stateReason = data.stateReason ?? undefined
                }

                if (TERMINAL_STATUSES.has(mapped)) {
                  closeSse(id)
                  void fetchSandboxes()
                  return
                }
              } catch {
                continue
              }
            }

            return read()
          })
        }

        return read()
      })
      .catch(() => {
        closeSse(id)
      })
  }

  function closeSse(id: string) {
    const abort = sseConnections.get(id)
    if (abort) {
      abort.abort()
      sseConnections.delete(id)
    }
  }

  function closeAllSse() {
    for (const [, abort] of sseConnections) abort.abort()
    sseConnections.clear()
  }

  function syncSseSubscriptions() {
    for (const sandbox of sandboxes.value) {
      if (TRANSITIONAL_STATUSES.has(sandbox.status)) {
        subscribeSse(sandbox.id)
      }
    }
    for (const [id] of sseConnections) {
      const sandbox = sandboxes.value.find((s) => s.id === id)
      if (!sandbox || !TRANSITIONAL_STATUSES.has(sandbox.status)) {
        closeSse(id)
      }
    }
  }

  async function fetchSandboxes() {
    if (!store.initialized) loading.value = true
    error.value = null
    try {
      const response = await sandboxesApi.list({ limit: 500 })
      sandboxes.value = response.data
      syncSseSubscriptions()
      void fetchHealth()
      store.initialized = true
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : 'Fehler beim Laden'
    } finally {
      loading.value = false
    }
  }

  async function fetchAdminSandboxes() {
    if (!authStore.isAuthenticated || !authStore.isAdmin) return
    try {
      const response = await sandboxesApi.list({ scope: 'all', limit: 500 })
      adminSandboxes.value = response.data
    } catch {
      adminSandboxes.value = []
    }
  }

  async function fetchHealthForSandbox(id: string): Promise<SandboxHealthEvent | null> {
    const baseURL = import.meta.env.WEB_API_URL || ''
    const abort = new AbortController()
    const timeout = setTimeout(() => abort.abort(), 10_000)

    const headers: Record<string, string> = {}
    const token = getToken()
    if (token) headers.Authorization = `Bearer ${token}`

    try {
      const res = await fetch(`${baseURL}/api/sandboxes/${id}/health`, {
        headers,
        credentials: 'include',
        signal: abort.signal,
      })
      if (!res.ok || !res.body) return null

      const reader = res.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() ?? ''

        for (const line of lines) {
          if (!line.startsWith('data: ')) continue
          try {
            const event: SandboxHealthEvent = JSON.parse(line.slice(6))
            reader.cancel()
            return event
          } catch {
            continue
          }
        }
      }
    } catch {
      return null
    } finally {
      clearTimeout(timeout)
    }

    return null
  }

  async function fetchHealth() {
    const running = sandboxes.value.filter((s) => s.status === 'running' || s.status === 'starting')
    const results = await Promise.allSettled(running.map((s) => fetchHealthForSandbox(s.id)))
    const updated: Record<string, SandboxHealthEvent> = {}
    for (const result of results) {
      if (result.status === 'fulfilled' && result.value) {
        updated[result.value.sandboxId] = result.value
      }
    }
    healthBySandboxId.value = updated
  }

  function startPolling() {
    pollConsumers++
    if (pollInterval) return
    pollInterval = setInterval(fetchSandboxes, 5_000)
  }

  function stopPolling() {
    pollConsumers--
    if (pollConsumers <= 0) {
      pollConsumers = 0
      if (pollInterval) {
        clearInterval(pollInterval)
        pollInterval = null
      }
    }
  }

  async function createSandbox(req: CreateSandboxRequest): Promise<Sandbox> {
    const sandbox = await sandboxesApi.create(req)
    sandboxes.value.unshift(sandbox)
    void fetchSandboxes()
    return sandbox
  }

  async function updateSandbox(id: string, req: UpdateSandboxRequest): Promise<Sandbox> {
    const updated = await sandboxesApi.update(id, req)
    const idx = sandboxes.value.findIndex((s) => s.id === id)
    if (idx !== -1) sandboxes.value[idx] = updated
    void fetchSandboxes()
    return updated
  }

  async function deleteSandbox(id: string, { skipRemove = false } = {}) {
    closeSse(id)
    await sandboxesApi.remove(id)
    if (!skipRemove) {
      sandboxes.value = sandboxes.value.filter((s) => s.id !== id)
    }
    void fetchSandboxes()
  }

  function removeSandbox(id: string) {
    sandboxes.value = sandboxes.value.filter((s) => s.id !== id)
  }

  async function snapshotSandbox(id: string, req: CreateSnapshotRequest): Promise<Image> {
    const image = await sandboxesApi.snapshot(id, req)
    void fetchSandboxes()
    return image
  }

  onMounted(() => {
    void fetchSandboxes()
    startPolling()
  })

  onUnmounted(() => {
    stopPolling()
    if (pollConsumers <= 0) closeAllSse()
  })

  return {
    sandboxes,
    activeSandboxes,
    recentSandboxes,
    allSandboxes,
    loading,
    error,
    busyIds,
    healthBySandboxId,
    refresh: fetchSandboxes,
    fetchAdminSandboxes,
    createSandbox,
    updateSandbox,
    deleteSandbox,
    removeSandbox,
    snapshotSandbox,
  }
}
