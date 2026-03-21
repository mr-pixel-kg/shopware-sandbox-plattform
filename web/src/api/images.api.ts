import { apiClient } from './client'
import type { CreateImageRequest, Image, PendingPull } from '@/types'

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

  async remove(id: string): Promise<void> {
    await apiClient.delete(`/api/images/${id}`)
  },
}
