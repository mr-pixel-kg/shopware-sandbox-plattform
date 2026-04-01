import { apiClient } from './client'

import type { AuditLogFacetsResponse, AuditLogListResponse } from '@/types'

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

export interface AuditLogFacetParams {
  action?: string
  resourceType?: string
  resourceId?: string
  clientToken?: string
  from?: string
  to?: string
}

export const auditApi = {
  async list(params?: number | AuditLogListParams): Promise<AuditLogListResponse> {
    const query = typeof params === 'number' ? { limit: params } : params

    const { data } = await apiClient.get<AuditLogListResponse>('/api/audit-logs', {
      params: query,
    })
    return data
  },

  async facets(params?: AuditLogFacetParams): Promise<AuditLogFacetsResponse> {
    const { data } = await apiClient.get<AuditLogFacetsResponse>('/api/audit-logs/facets', {
      params,
    })
    return data
  },
}
