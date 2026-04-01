import { apiClient } from './client'

import type { AuditLog } from '@/types'

export interface AuditLogListParams {
  limit?: number
  offset?: number
  userId?: string
  action?: string
  resourceType?: string
  resourceId?: string
  clientToken?: string
  from?: string
  to?: string
}

export const auditApi = {
  async list(params?: number | AuditLogListParams): Promise<AuditLog[]> {
    const query = typeof params === 'number' ? { limit: params } : params

    const { data } = await apiClient.get<AuditLog[]>('/api/audit-logs', {
      params: query,
    })
    return data
  },
}
