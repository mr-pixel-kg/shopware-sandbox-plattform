import { apiClient } from './client'

import type { AuditLogFacetsResponse, AuditLogListResponse, PaginationParams } from '@/types'

export interface AuditLogListParams extends PaginationParams {
  userId?: string
  action?: string
  resourceType?: string
  resourceId?: string
  clientId?: string
  from?: string
  to?: string
}

export interface AuditLogFacetParams {
  action?: string
  resourceType?: string
  resourceId?: string
  clientId?: string
  from?: string
  to?: string
}

export const auditApi = {
  async list(params?: AuditLogListParams): Promise<AuditLogListResponse> {
    const { data } = await apiClient.get<AuditLogListResponse>('/api/audit-logs', {
      params,
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
