import { apiClient } from './client'

import type {
  CreateSandboxRequest,
  CreateSnapshotRequest,
  Image,
  LogSource,
  PaginatedResponse,
  PaginationParams,
  Sandbox,
  UpdateSandboxRequest,
} from '@/types'

export const sandboxesApi = {
  async list(params?: PaginationParams & { scope?: 'all' }): Promise<PaginatedResponse<Sandbox>> {
    const { data } = await apiClient.get<PaginatedResponse<Sandbox>>('/api/sandboxes', { params })
    return data
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

  async listLogSources(id: string): Promise<LogSource[]> {
    const { data } = await apiClient.get<LogSource[]>(`/api/sandboxes/${id}/logs`)
    return data
  },
}
