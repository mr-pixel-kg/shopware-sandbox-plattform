import { storeToRefs } from 'pinia'
import { onMounted, onUnmounted } from 'vue'

import { useAuthStore } from '@/stores/auth.store'
import { useSandboxesStore } from '@/stores/sandboxes.store'

export function useSandboxes(mode: 'mine' | 'all' = 'mine') {
  const store = useSandboxesStore()
  const authStore = useAuthStore()
  const { sandboxes, activeSandboxes, recentSandboxes, loading, error } = storeToRefs(store)

  let pollInterval: ReturnType<typeof setInterval> | null = null

  function fetch() {
    if (mode === 'all') {
      return store.fetchAllSandboxes()
    }
    if (authStore.isAuthenticated) {
      return store.fetchMySandboxes()
    }
    return store.fetchGuestSandboxes()
  }

  onMounted(() => {
    store.$reset()
    fetch()
    pollInterval = setInterval(fetch, 10_000)
  })

  onUnmounted(() => {
    if (pollInterval) clearInterval(pollInterval)
    store.closeAllHealthStreams()
  })

  return {
    sandboxes,
    activeSandboxes,
    recentSandboxes,
    loading,
    error,
    refresh: fetch,
    createSandbox: store.createSandbox,
    createPublicDemo: store.createPublicDemo,
    extendTTL: store.extendTTL,
    deleteSandbox: store.deleteSandbox,
    snapshotSandbox: store.snapshotSandbox,
    healthBySandboxId: store.healthBySandboxId,
  }
}
