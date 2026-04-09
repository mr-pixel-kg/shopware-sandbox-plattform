import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import type { Sandbox } from '@/types'

const ACTIVE_STATUSES = new Set(['running', 'starting', 'paused', 'stopping'])
const RECENT_STATUSES = new Set(['stopped', 'expired', 'deleted', 'failed'])

export const useSandboxesStore = defineStore('sandboxes', () => {
  const sandboxes = ref<Sandbox[]>([])
  const adminSandboxes = ref<Sandbox[]>([])
  const loading = ref(false)
  const initialized = ref(false)
  const error = ref<string | null>(null)

  const activeSandboxes = computed(() =>
    sandboxes.value.filter((s) => ACTIVE_STATUSES.has(s.status)),
  )

  const recentSandboxes = computed(() =>
    sandboxes.value.filter((s) => RECENT_STATUSES.has(s.status)),
  )

  const allSandboxes = computed(() => adminSandboxes.value)

  function $reset() {
    sandboxes.value = []
    adminSandboxes.value = []
    loading.value = false
    initialized.value = false
    error.value = null
  }

  return {
    sandboxes,
    adminSandboxes,
    loading,
    initialized,
    error,
    activeSandboxes,
    recentSandboxes,
    allSandboxes,
    $reset,
  }
})
