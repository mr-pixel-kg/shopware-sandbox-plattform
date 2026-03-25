import { computed, onMounted, ref } from 'vue'

import { usersApi, whitelistApi } from '@/api'

import type { CreateUserRequest, ManagedUser, UpdateUserRequest, User } from '@/types'

export function useUsers() {
  const users = ref<User[]>([])
  const pendingUsers = ref<User[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const pendingIds = computed(() => new Set(pendingUsers.value.map((user) => user.id)))

  const activeUsers = computed<ManagedUser[]>(() =>
    users.value
      .filter((user) => !pendingIds.value.has(user.id))
      .map((user) => ({ ...user, pending: false })),
  )

  const invitedUsers = computed<ManagedUser[]>(() =>
    pendingUsers.value.map((user) => ({ ...user, pending: true })),
  )

  async function fetch() {
    loading.value = true
    error.value = null
    try {
      const [allUsers, pending] = await Promise.all([usersApi.list(), whitelistApi.list()])
      users.value = allUsers
      pendingUsers.value = pending
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : 'Fehler beim Laden'
    } finally {
      loading.value = false
    }
  }

  async function createUser(req: CreateUserRequest): Promise<User> {
    const user = await usersApi.create(req)
    users.value.unshift(user)
    return user
  }

  async function inviteUser(req: CreateUserRequest): Promise<User> {
    const user = await whitelistApi.add(req)
    users.value.unshift(user)
    pendingUsers.value.unshift(user)
    return user
  }

  async function updateUser(id: string, req: UpdateUserRequest): Promise<User> {
    const user = await usersApi.update(id, req)
    users.value = users.value.map((existing) => (existing.id === id ? user : existing))
    pendingUsers.value = pendingUsers.value.filter((existing) => existing.id !== id)
    return user
  }

  async function deleteUser(id: string): Promise<void> {
    await usersApi.remove(id)
    users.value = users.value.filter((user) => user.id !== id)
    pendingUsers.value = pendingUsers.value.filter((user) => user.id !== id)
  }

  async function deleteInvite(id: string): Promise<void> {
    await whitelistApi.remove(id)
    users.value = users.value.filter((user) => user.id !== id)
    pendingUsers.value = pendingUsers.value.filter((user) => user.id !== id)
  }

  onMounted(() => {
    void fetch()
  })

  return {
    users,
    activeUsers,
    invitedUsers,
    loading,
    error,
    createUser,
    inviteUser,
    updateUser,
    deleteUser,
    deleteInvite,
    refresh: fetch,
  }
}
