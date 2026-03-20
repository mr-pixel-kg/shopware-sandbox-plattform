import { apiClient } from './client'
import type { CreateImageRequest, CreateImageResult, Image, PendingPull } from '@/types'

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

  async create(req: CreateImageRequest): Promise<CreateImageResult> {
    const response = await apiClient.post('/api/images', req)
    if (response.status === 202) {
      return { pendingPull: response.data as PendingPull }
    }
    return { image: response.data as Image }
  },

  async remove(id: string): Promise<void> {
    await apiClient.delete(`/api/images/${id}`)
  },
}
