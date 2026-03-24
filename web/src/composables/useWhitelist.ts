import { onMounted, ref } from 'vue'

import { whitelistApi } from '@/api'

import type { AddWhitelistRequest, User } from '@/types'

export function useWhitelist() {
  const pendingUsers = ref<User[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetch() {
    loading.value = true
    error.value = null
    try {
      pendingUsers.value = await whitelistApi.list()
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : 'Fehler beim Laden'
    } finally {
      loading.value = false
    }
  }

  async function add(req: AddWhitelistRequest): Promise<User> {
    const user = await whitelistApi.add(req)
    pendingUsers.value.unshift(user)
    return user
  }

  async function remove(id: string): Promise<void> {
    await whitelistApi.remove(id)
    pendingUsers.value = pendingUsers.value.filter((u) => u.id !== id)
  }

  onMounted(() => {
    pendingUsers.value = []
    loading.value = true
    fetch()
  })

  return {
    pendingUsers,
    loading,
    error,
    add,
    remove,
    refresh: fetch,
  }
}
