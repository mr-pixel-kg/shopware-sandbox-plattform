import { apiClient } from './client'

import type { CreateUserRequest, UpdateUserRequest, User } from '@/types'

export const usersApi = {
  async list(): Promise<User[]> {
    const { data } = await apiClient.get<User[]>('/api/admin/users')
    return data
  },

  async create(req: CreateUserRequest): Promise<User> {
    const { data } = await apiClient.post<User>('/api/admin/users', req)
    return data
  },

  async update(id: string, req: UpdateUserRequest): Promise<User> {
    const { data } = await apiClient.put<User>(`/api/admin/users/${id}`, req)
    return data
  },

  async remove(id: string): Promise<void> {
    await apiClient.delete(`/api/admin/users/${id}`)
  },
}
