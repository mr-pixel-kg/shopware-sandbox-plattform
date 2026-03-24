import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { authApi } from '@/api'
import { AUTH_LOGOUT_EVENT } from '@/api/client'
import {
  clearToken,
  getToken,
  isTokenExpired,
  setToken,
  setupStorageListener,
} from '@/utils/storage'

import type { User } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(null)
  const user = ref<User | null>(null)

  let removeStorageListener: (() => void) | null = null
  let removeVisibilityListener: (() => void) | null = null
  let removeSessionExpiredListener: (() => void) | null = null

  const isAuthenticated = computed(() => {
    return token.value !== null && !isTokenExpired(token.value)
  })

  const isAdmin = computed(() => isAuthenticated.value && user.value?.role === 'admin')

  async function initialize() {
    if (removeStorageListener) return

    const stored = getToken()
    if (stored && !isTokenExpired(stored)) {
      token.value = stored
      try {
        user.value = await authApi.me()
      } catch {
        clearAuthState()
        return
      }
    } else if (stored) {
      clearAuthState()
    }

    removeStorageListener = setupStorageListener(() => {
      token.value = null
      user.value = null
    })

    const onSessionExpired = () => clearAuthState()
    window.addEventListener(AUTH_LOGOUT_EVENT, onSessionExpired)
    removeSessionExpiredListener = () =>
      window.removeEventListener(AUTH_LOGOUT_EVENT, onSessionExpired)

    const onVisibilityChange = async () => {
      if (document.visibilityState !== 'visible') return
      if (!token.value || isTokenExpired(token.value)) {
        clearAuthState()
        return
      }
      try {
        user.value = await authApi.me()
      } catch {
        clearAuthState()
      }
    }
    document.addEventListener('visibilitychange', onVisibilityChange)
    removeVisibilityListener = () =>
      document.removeEventListener('visibilitychange', onVisibilityChange)
  }

  async function login(email: string, password: string) {
    const response = await authApi.login({ email, password })
    token.value = response.token
    user.value = response.user
    setToken(response.token)
  }

  async function register(email: string, password: string) {
    await authApi.register({ email, password })
    await login(email, password)
  }

  async function logout() {
    try {
      await authApi.logout()
    } finally {
      clearAuthState()
    }
  }

  async function fetchMe() {
    user.value = await authApi.me()
  }

  function clearAuthState() {
    token.value = null
    user.value = null
    clearToken()
  }

  function cleanup() {
    removeStorageListener?.()
    removeStorageListener = null
    removeVisibilityListener?.()
    removeVisibilityListener = null
    removeSessionExpiredListener?.()
    removeSessionExpiredListener = null
  }

  return {
    token,
    user,
    isAuthenticated,
    isAdmin,
    initialize,
    login,
    register,
    logout,
    fetchMe,
    clearAuthState,
    cleanup,
  }
})
