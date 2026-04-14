import { apiClient } from './client'

import type {
  CreateImageRequest,
  Image,
  MetadataSchema,
  PaginatedResponse,
  PaginationParams,
  PendingImage,
  UpdateImageRequest,
} from '@/types'

export const imagesApi = {
  async list(
    params?: PaginationParams & { visibility?: 'public' },
  ): Promise<PaginatedResponse<Image>> {
    const { data } = await apiClient.get<PaginatedResponse<Image>>('/api/images', { params })
    return data
  },

  async listPending(): Promise<PendingImage[]> {
    const { data } = await apiClient.get<PendingImage[]>('/api/images/pending')
    return data
  },

  async create(req: CreateImageRequest): Promise<Image> {
    const { data } = await apiClient.post<Image>('/api/images', req)
    return data
  },

  async update(id: string, req: UpdateImageRequest): Promise<Image> {
    const { data } = await apiClient.patch<Image>(`/api/images/${id}`, req)
    return data
  },

  async uploadThumbnail(id: string, file: File): Promise<Image> {
    const form = new FormData()
    form.append('thumbnail', file)
    const { data } = await apiClient.post<Image>(`/api/images/${id}/thumbnail`, form, {
      headers: { 'Content-Type': undefined },
    })
    return data
  },

  async deleteThumbnail(id: string): Promise<void> {
    await apiClient.delete(`/api/images/${id}/thumbnail`)
  },

  async remove(id: string): Promise<void> {
    await apiClient.delete(`/api/images/${id}`)
  },

  async lookupRegistry(imageName: string): Promise<MetadataSchema | null> {
    try {
      const { data } = await apiClient.get<MetadataSchema>('/api/registry', {
        params: { name: imageName },
      })
      return data
    } catch {
      return null
    }
  },
}
