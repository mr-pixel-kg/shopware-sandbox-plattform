import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { sandboxesApi } from '@/api'
import { getToken } from '@/utils/storage'

import type {
  CreateSandboxRequest,
  CreateSnapshotRequest,
  Image,
  Sandbox,
  SandboxHealthEvent,
  SandboxStatus,
} from '@/types'

export const useSandboxesStore = defineStore('sandboxes', () => {
  const sandboxes = ref<Sandbox[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const healthBySandboxId = ref<Record<string, SandboxHealthEvent>>({})

  const healthConnections = new Map<string, EventSource>()

  const activeSandboxes = computed(() =>
    sandboxes.value.filter((s) => s.status === 'running' || s.status === 'starting'),
  )

  const recentSandboxes = computed(() =>
    sandboxes.value.filter(
      (s) =>
        s.status === 'stopped' ||
        s.status === 'expired' ||
        s.status === 'failed' ||
        s.status === 'deleted',
    ),
  )

  function buildHealthStreamUrl(id: string): string {
    const base = import.meta.env.WEB_API_URL || window.location.origin
    const url = new URL(`/api/sandboxes/${id}/health`, base)
    const token = getToken()
    if (token) {
      url.searchParams.set('access_token', token)
    }
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
    for (const [id, es] of healthConnections) {
      es.close()
      healthConnections.delete(id)
    }
  }

  function updateSandboxStatus(id: string, status: SandboxStatus) {
    const sandbox = sandboxes.value.find((item) => item.id === id)
    if (sandbox) sandbox.status = status
  }

  function applyHealthEvent(event: SandboxHealthEvent) {
    healthBySandboxId.value = {
      ...healthBySandboxId.value,
      [event.sandboxId]: event,
    }

    if (event.ready) {
      updateSandboxStatus(event.sandboxId, 'running')
      return
    }

    if (
      event.status === 'deleted' ||
      event.status === 'expired' ||
      event.status === 'failed' ||
      event.status === 'stopped'
    ) {
      updateSandboxStatus(event.sandboxId, event.status)
      closeHealthStream(event.sandboxId)
    }
  }

  function subscribeHealth(id: string) {
    if (healthConnections.has(id)) return

    const es = new EventSource(buildHealthStreamUrl(id), { withCredentials: true })
    healthConnections.set(id, es)

    es.onmessage = (message) => {
      try {
        applyHealthEvent(JSON.parse(message.data) as SandboxHealthEvent)
      } catch {
        // ignore malformed SSE payloads
      }
    }

    es.onerror = () => {
      const sandbox = sandboxes.value.find((item) => item.id === id)
      if (!sandbox || (sandbox.status !== 'starting' && sandbox.status !== 'running')) {
        closeHealthStream(id)
      }
    }
  }

  function syncHealthSubscriptions() {
    const activeIds = new Set(
      sandboxes.value
        .filter((sandbox) => sandbox.status === 'starting' || sandbox.status === 'running')
        .map((sandbox) => sandbox.id),
    )

    for (const sandboxId of activeIds) {
      subscribeHealth(sandboxId)
    }

    for (const [sandboxId] of healthConnections) {
      if (!activeIds.has(sandboxId)) {
        closeHealthStream(sandboxId)
      }
    }
  }

  function $reset() {
    closeAllHealthStreams()
    sandboxes.value = []
    loading.value = true
    error.value = null
    healthBySandboxId.value = {}
  }

  async function fetchMySandboxes() {
    error.value = null
    try {
      sandboxes.value = await sandboxesApi.listMine()
      syncHealthSubscriptions()
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : 'Fehler beim Laden'
    } finally {
      loading.value = false
    }
  }

  async function fetchAllSandboxes() {
    error.value = null
    try {
      sandboxes.value = await sandboxesApi.list()
      syncHealthSubscriptions()
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : 'Fehler beim Laden'
    } finally {
      loading.value = false
    }
  }

  async function fetchGuestSandboxes() {
    error.value = null
    try {
      sandboxes.value = await sandboxesApi.listGuest()
      syncHealthSubscriptions()
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

  async function extendTTL(id: string, ttlMinutes: number): Promise<Sandbox> {
    const updated = await sandboxesApi.extendTTL(id, ttlMinutes)
    const idx = sandboxes.value.findIndex((s) => s.id === id)
    if (idx !== -1) sandboxes.value[idx] = updated
    return updated
  }

  async function deleteSandbox(id: string) {
    closeHealthStream(id)
    await sandboxesApi.remove(id)
    sandboxes.value = sandboxes.value.filter((s) => s.id !== id)
  }

  function removeSandboxFromList(id: string) {
    sandboxes.value = sandboxes.value.filter((s) => s.id !== id)
  }

  async function snapshotSandbox(id: string, req: CreateSnapshotRequest): Promise<Image> {
    return await sandboxesApi.snapshot(id, req)
  }

  return {
    sandboxes,
    loading,
    error,
    healthBySandboxId,
    activeSandboxes,
    recentSandboxes,
    $reset,
    fetchMySandboxes,
    fetchAllSandboxes,
    fetchGuestSandboxes,
    createSandbox,
    createPublicDemo,
    extendTTL,
    deleteSandbox,
    removeSandboxFromList,
    snapshotSandbox,
    closeAllHealthStreams,
  }
})
