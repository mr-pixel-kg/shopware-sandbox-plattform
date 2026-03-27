import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import type { Sandbox, SandboxHealthEvent } from '@/types'

export const useSandboxesStore = defineStore('sandboxes', () => {
  const sandboxes = ref<Sandbox[]>([])
  const loading = ref(false)
  const initialized = ref(false)
  const error = ref<string | null>(null)
  const healthBySandboxId = ref<Record<string, SandboxHealthEvent>>({})

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
    loading.value = false
    initialized.value = false
    error.value = null
    healthBySandboxId.value = {}
  }

  return {
    sandboxes,
    loading,
    initialized,
    error,
    healthBySandboxId,
    activeSandboxes,
    recentSandboxes,
    $reset,
  }
})
