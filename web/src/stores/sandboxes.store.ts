import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { sandboxesApi } from '@/api'

import type { CreateSandboxRequest, CreateSnapshotRequest, Image, Sandbox } from '@/types'

export const useSandboxesStore = defineStore('sandboxes', () => {
  const sandboxes = ref<Sandbox[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

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

  function $reset() {
    sandboxes.value = []
    loading.value = true
    error.value = null
  }

  async function fetchMySandboxes() {
    error.value = null
    try {
      sandboxes.value = await sandboxesApi.listMine()
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
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : 'Fehler beim Laden'
    } finally {
      loading.value = false
    }
  }

  async function createSandbox(req: CreateSandboxRequest): Promise<Sandbox> {
    const sandbox = await sandboxesApi.create(req)
    sandboxes.value.unshift(sandbox)
    return sandbox
  }

  async function createPublicDemo(req: CreateSandboxRequest): Promise<Sandbox> {
    const sandbox = await sandboxesApi.createPublicDemo(req)
    sandboxes.value.unshift(sandbox)
    return sandbox
  }

  async function extendTTL(id: string, ttlMinutes: number): Promise<Sandbox> {
    const updated = await sandboxesApi.extendTTL(id, ttlMinutes)
    const idx = sandboxes.value.findIndex((s) => s.id === id)
    if (idx !== -1) sandboxes.value[idx] = updated
    return updated
  }

  async function deleteSandbox(id: string) {
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
  }
})
