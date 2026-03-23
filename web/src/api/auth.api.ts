import { apiClient } from './client'

import type { LoginRequest, LoginResponse, RegisterRequest, User } from '@/types'

export const authApi = {
  async register(req: RegisterRequest): Promise<User> {
    const { data } = await apiClient.post<User>('/api/auth/register', req)
    return data
  },

  async login(req: LoginRequest): Promise<LoginResponse> {
    const { data } = await apiClient.post<LoginResponse>('/api/auth/login', req)
    return data
  },

  async logout(): Promise<void> {
    await apiClient.post('/api/auth/logout')
  },

  async me(): Promise<User> {
    const { data } = await apiClient.get<User>('/api/me')
    return data
  },
}
