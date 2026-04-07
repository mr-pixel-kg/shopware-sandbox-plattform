import { apiClient } from './client'

import type { CreateUserRequest, PaginatedResponse, UpdateUserRequest, User } from '@/types'

export const usersApi = {
  async list(): Promise<User[]> {
    const { data } = await apiClient.get<PaginatedResponse<User>>('/api/users', {
      params: { limit: 500 },
    })
    return data.data
  },

  async create(req: CreateUserRequest): Promise<User> {
    const { data } = await apiClient.post<User>('/api/users', req)
    return data
  },

  async update(id: string, req: UpdateUserRequest): Promise<User> {
    const { data } = await apiClient.patch<User>(`/api/users/${id}`, req)
    return data
  },

  async remove(id: string): Promise<void> {
    await apiClient.delete(`/api/users/${id}`)
  },
}
