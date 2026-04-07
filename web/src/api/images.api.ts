import { apiClient } from './client'

import type {
  CreateImageRequest,
  Image,
  MetadataItem,
  PaginatedResponse,
  PendingImage,
  UpdateImageRequest,
} from '@/types'

export const imagesApi = {
  async listPublic(): Promise<Image[]> {
    const { data } = await apiClient.get<PaginatedResponse<Image>>('/api/images/public', {
      params: { limit: 500 },
    })
    return data.data
  },

  async listAll(): Promise<Image[]> {
    const { data } = await apiClient.get<PaginatedResponse<Image>>('/api/images', {
      params: { limit: 500 },
    })
    return data.data
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

  async lookupRegistry(imageName: string): Promise<MetadataItem[]> {
    const { data } = await apiClient.get<MetadataItem[]>('/api/registry', {
      params: { name: imageName },
    })
    return data
  },
}
