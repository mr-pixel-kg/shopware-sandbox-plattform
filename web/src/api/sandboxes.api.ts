import { apiClient } from './client'

import type { CreateSandboxRequest, CreateSnapshotRequest, Image, Sandbox } from '@/types'

export const sandboxesApi = {
  async list(): Promise<Sandbox[]> {
    const { data } = await apiClient.get<Sandbox[]>('/api/sandboxes')
    return data
  },

  async listMine(): Promise<Sandbox[]> {
    const { data } = await apiClient.get<Sandbox[]>('/api/me/sandboxes')
    return data
  },

  async listGuest(): Promise<Sandbox[]> {
    const { data } = await apiClient.get<Sandbox[]>('/api/public/sandboxes')
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

  async createPublicDemo(req: CreateSandboxRequest): Promise<Sandbox> {
    const { data } = await apiClient.post<Sandbox>('/api/public/demos', req)
    return data
  },

  async extendTTL(id: string, ttlMinutes: number): Promise<Sandbox> {
    const { data } = await apiClient.patch<Sandbox>(`/api/sandboxes/${id}/ttl`, { ttlMinutes })
    return data
  },

  async remove(id: string): Promise<void> {
    await apiClient.delete(`/api/sandboxes/${id}`)
  },

  async removeGuest(id: string): Promise<void> {
    await apiClient.delete(`/api/public/sandboxes/${id}`)
  },

  async snapshot(id: string, req: CreateSnapshotRequest): Promise<Image> {
    const { data } = await apiClient.post<Image>(`/api/sandboxes/${id}/snapshot`, req, {
      timeout: 120_000,
    })
    return data
  },
}
