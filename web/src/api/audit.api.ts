import { apiClient } from './client'
import type { AuditLog } from '@/types'

export const auditApi = {
  async list(limit?: number): Promise<AuditLog[]> {
    const { data } = await apiClient.get<AuditLog[]>('/api/audit-logs', {
      params: limit != null ? { limit } : undefined,
    })
    return data
  },
}
