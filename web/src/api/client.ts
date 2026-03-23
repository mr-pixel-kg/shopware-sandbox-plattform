import axios from 'axios'

import { getToken } from '@/utils/storage'

import type { AxiosInstance } from 'axios'

const PUBLIC_PATHS = ['/auth/login', '/auth/register', '/public/']

function isPublicRoute(url: string | undefined): boolean {
  if (!url) return false
  return PUBLIC_PATHS.some((path) => url.includes(path))
}

export const AUTH_LOGOUT_EVENT = 'auth:session-expired'

export const apiClient: AxiosInstance = axios.create({
  baseURL: import.meta.env.WEB_API_URL,
  timeout: 30_000,
  headers: { 'Content-Type': 'application/json' },
  withCredentials: true,
})

apiClient.interceptors.request.use((config) => {
  if (!isPublicRoute(config.url)) {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
  }
  return config
})

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (
      axios.isAxiosError(error) &&
      error.response?.status === 401 &&
      !isPublicRoute(error.config?.url)
    ) {
      window.dispatchEvent(new CustomEvent(AUTH_LOGOUT_EVENT))
    }
    return Promise.reject(error)
  },
)
