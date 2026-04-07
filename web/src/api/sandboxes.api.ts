import { apiClient } from './client'

import type {
  CreateDemoRequest,
  CreateSandboxRequest,
  CreateSnapshotRequest,
  Image,
  PaginatedResponse,
  Sandbox,
  UpdateSandboxRequest,
} from '@/types'

export const sandboxesApi = {
  async list(): Promise<Sandbox[]> {
    const { data } = await apiClient.get<PaginatedResponse<Sandbox>>('/api/sandboxes', {
      params: { limit: 500 },
    })
    return data.data
  },

  async get(id: string): Promise<Sandbox> {
    const { data } = await apiClient.get<Sandbox>(`/api/sandboxes/${id}`)
    return data
  },

  async create(req: CreateSandboxRequest): Promise<Sandbox> {
    const { data } = await apiClient.post<Sandbox>('/api/sandboxes', req)
    return data
  },

  async update(id: string, req: UpdateSandboxRequest): Promise<Sandbox> {
    const { data } = await apiClient.patch<Sandbox>(`/api/sandboxes/${id}`, req)
    return data
  },

  async remove(id: string): Promise<void> {
    await apiClient.delete(`/api/sandboxes/${id}`)
  },

  async snapshot(id: string, req: CreateSnapshotRequest): Promise<Image> {
    const { data } = await apiClient.post<Image>(`/api/sandboxes/${id}/snapshots`, req)
    return data
  },

  async createDemo(req: CreateDemoRequest): Promise<Sandbox> {
    const { data } = await apiClient.post<Sandbox>('/api/demos', req)
    return data
  },

  async listDemos(clientId: string): Promise<Sandbox[]> {
    const { data } = await apiClient.get<Sandbox[]>('/api/demos', {
      params: { clientId },
    })
    return data
  },

  async removeDemo(id: string): Promise<void> {
    await apiClient.delete(`/api/demos/${id}`)
  },
}
