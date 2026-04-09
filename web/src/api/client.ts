import axios from 'axios'

import { getToken } from '@/utils/storage'

import type { AxiosInstance } from 'axios'

export const AUTH_LOGOUT_EVENT = 'auth:session-expired'

export const apiClient: AxiosInstance = axios.create({
  baseURL: import.meta.env.WEB_API_URL,
  timeout: 30_000,
  headers: { 'Content-Type': 'application/json' },
  withCredentials: true,
})

apiClient.interceptors.request.use((config) => {
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      const url = error.config?.url ?? ''
      if (!url.includes('/auth/login') && !url.includes('/auth/register')) {
        window.dispatchEvent(new CustomEvent(AUTH_LOGOUT_EVENT))
      }
    }
    return Promise.reject(error)
  },
)
