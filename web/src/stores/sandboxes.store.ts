import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { useAuthStore } from '@/stores/auth.store'

import type { Sandbox } from '@/types'

const ACTIVE_STATUSES = new Set(['running', 'starting', 'paused', 'stopping'])
const RECENT_STATUSES = new Set(['stopped', 'expired', 'deleted', 'failed'])

export const useSandboxesStore = defineStore('sandboxes', () => {
  const authStore = useAuthStore()

  const sandboxes = ref<Sandbox[]>([])
  const loading = ref(false)
  const initialized = ref(false)
  const error = ref<string | null>(null)

  const mySandboxes = computed(() => {
    if (!authStore.isAdmin) return sandboxes.value
    const userId = authStore.user?.id
    if (!userId) return sandboxes.value
    return sandboxes.value.filter((s) => s.owner?.id === userId || (!s.owner && s.guestSessionId))
  })

  const activeSandboxes = computed(() =>
    mySandboxes.value.filter((s) => ACTIVE_STATUSES.has(s.status)),
  )

  const recentSandboxes = computed(() =>
    mySandboxes.value.filter((s) => RECENT_STATUSES.has(s.status)),
  )

  const allSandboxes = computed(() => sandboxes.value)

  function $reset() {
    sandboxes.value = []
    loading.value = false
    initialized.value = false
    error.value = null
  }

  return {
    sandboxes,
    loading,
    initialized,
    error,
    mySandboxes,
    activeSandboxes,
    recentSandboxes,
    allSandboxes,
    $reset,
  }
})
