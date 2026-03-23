import { apiClient } from './client'

import type { CreateImageRequest, Image, PendingPull, UpdateImageRequest } from '@/types'

export const imagesApi = {
  async listPublic(): Promise<Image[]> {
    const { data } = await apiClient.get<Image[]>('/api/public/images')
    return data
  },

  async listAll(): Promise<Image[]> {
    const { data } = await apiClient.get<Image[]>('/api/images')
    return data
  },

  async listPulls(): Promise<PendingPull[]> {
    const { data } = await apiClient.get<PendingPull[]>('/api/images/pulls')
    return data
  },

  async create(req: CreateImageRequest): Promise<Image> {
    const { data } = await apiClient.post<Image>('/api/images', req)
    return data
  },

  async update(id: string, req: UpdateImageRequest): Promise<Image> {
    const { data } = await apiClient.put<Image>(`/api/images/${id}`, req)
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
}
