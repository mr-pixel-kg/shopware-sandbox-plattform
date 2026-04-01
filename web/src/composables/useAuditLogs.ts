import { computed, ref, watch } from 'vue'

import { auditApi } from '@/api'

import type { AuditLog, AuditLogListMeta } from '@/types'

export function useAuditLogs() {
  const logs = ref<AuditLog[]>([])
  const meta = ref<AuditLogListMeta | null>(null)
  const availableUsers = ref<Array<{ id: string; email: string }>>([])
  const availableActions = ref<string[]>([])
  const loading = ref(false)
  const initialized = ref(false)
  const error = ref<string | null>(null)

  const userFilter = ref<string>('all')
  const actionFilter = ref<string>('all')
  const periodFilter = ref<string>('7d')
  const page = ref(1)
  const pageSize = 20

  const totalPages = computed(() => {
    const total = meta.value?.pagination.total ?? logs.value.length
    return Math.max(1, Math.ceil(total / pageSize))
  })

  const queryParams = computed(() => {
    return {
      limit: pageSize,
      offset: (page.value - 1) * pageSize,
      userId: userFilter.value !== 'all' ? userFilter.value : undefined,
      action: actionFilter.value !== 'all' ? actionFilter.value : undefined,
      from: periodStart.value,
    }
  })

  const periodStart = computed(() => {
    const now = Date.now()
    const periodMs: Record<string, number> = {
      '24h': 24 * 60 * 60 * 1000,
      '7d': 7 * 24 * 60 * 60 * 1000,
      '30d': 30 * 24 * 60 * 60 * 1000,
    }
    const duration = periodMs[periodFilter.value] ?? periodMs['7d']
    return new Date(now - duration).toISOString()
  })

  async function fetch() {
    if (!initialized.value) loading.value = true
    error.value = null
    try {
      const response = await auditApi.list(queryParams.value)
      logs.value = response.data
      meta.value = response.meta
      initialized.value = true
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : 'Fehler beim Laden'
    } finally {
      loading.value = false
    }
  }

  async function fetchFacets() {
    try {
      const response = await auditApi.facets({ from: periodStart.value })
      availableUsers.value = response.users
      availableActions.value = response.actions
    } catch {
      availableUsers.value = []
      availableActions.value = []
    }
  }

  async function exportCsv() {
    const exportedLogs: AuditLog[] = []
    let offset = 0

    while (true) {
      const response = await auditApi.list({
        ...queryParams.value,
        limit: 500,
        offset,
      })
      exportedLogs.push(...response.data)

      if (!response.meta.pagination.hasMore) {
        break
      }
      offset += response.meta.pagination.count
    }

    const headers = [
      'Zeitpunkt',
      'Benutzer',
      'Aktion',
      'Ressource',
      'Ressource-ID',
      'Details',
      'IP',
      'User-Agent',
      'Client-Token',
    ]
    const rows = exportedLogs.map((l) => [
      l.timestamp,
      l.user?.email ?? l.user?.id ?? '',
      l.action,
      l.resourceType ?? '',
      l.resourceId ?? '',
      JSON.stringify(l.details),
      l.ipAddress ?? '',
      l.userAgent ?? '',
      l.clientToken ?? '',
    ])
    const csv = [headers, ...rows].map((r) => r.join(';')).join('\n')
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `audit-log-${new Date().toISOString().slice(0, 10)}.csv`
    a.click()
    URL.revokeObjectURL(url)
  }

  watch([userFilter, actionFilter, periodFilter], () => {
    page.value = 1
  })

  watch(
    periodStart,
    () => {
      void fetchFacets()
    },
    { immediate: true },
  )

  watch(
    queryParams,
    () => {
      void fetch()
    },
    { immediate: true },
  )

  return {
    logs,
    meta,
    loading,
    error,
    page,
    totalPages,
    pageSize,
    userFilter,
    actionFilter,
    periodFilter,
    availableUsers,
    availableActions,
    refresh: fetch,
    exportCsv,
  }
}
