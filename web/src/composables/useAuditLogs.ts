import { computed, onMounted, ref } from 'vue'

import { auditApi } from '@/api'

import type { AuditLog } from '@/types'

export function useAuditLogs() {
  const allLogs = ref<AuditLog[]>([])
  const loading = ref(false)
  const initialized = ref(false)
  const error = ref<string | null>(null)

  const userFilter = ref<string>('all')
  const actionFilter = ref<string>('all')
  const periodFilter = ref<string>('7d')
  const page = ref(1)
  const pageSize = 20

  const filteredLogs = computed(() => {
    let logs = allLogs.value

    if (userFilter.value && userFilter.value !== 'all') {
      logs = logs.filter((l) => l.user?.id === userFilter.value)
    }

    if (actionFilter.value && actionFilter.value !== 'all') {
      logs = logs.filter((l) => l.action === actionFilter.value)
    }

    if (periodFilter.value) {
      const now = Date.now()
      const periodMs: Record<string, number> = {
        '24h': 24 * 60 * 60 * 1000,
        '7d': 7 * 24 * 60 * 60 * 1000,
        '30d': 30 * 24 * 60 * 60 * 1000,
      }
      const cutoff = now - (periodMs[periodFilter.value] ?? periodMs['7d'])
      logs = logs.filter((l) => new Date(l.createdAt).getTime() >= cutoff)
    }

    return logs
  })

  const totalPages = computed(() => Math.max(1, Math.ceil(filteredLogs.value.length / pageSize)))

  const paginatedLogs = computed(() => {
    const start = (page.value - 1) * pageSize
    return filteredLogs.value.slice(start, start + pageSize)
  })

  const uniqueUsers = computed(() => {
    const users = new Map<string, string>()
    for (const log of allLogs.value) {
      if (log.user) users.set(log.user.id, log.user.email)
    }
    return [...users.entries()].map(([id, email]) => ({ id, email }))
  })

  const uniqueActions = computed(() => [...new Set(allLogs.value.map((l) => l.action))])

  async function fetch() {
    if (!initialized.value) loading.value = true
    error.value = null
    try {
      allLogs.value = await auditApi.list(500)
      initialized.value = true
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : 'Fehler beim Laden'
    } finally {
      loading.value = false
    }
  }

  function exportCsv() {
    const headers = ['Zeitpunkt', 'Benutzer', 'Aktion', 'Details', 'IP']
    const rows = filteredLogs.value.map((l) => [
      l.createdAt,
      l.user?.email ?? l.user?.id ?? '',
      l.action,
      JSON.stringify(l.details),
      l.ipAddress,
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

  onMounted(() => {
    void fetch()
  })

  return {
    logs: paginatedLogs,
    allLogs: filteredLogs,
    loading,
    error,
    page,
    totalPages,
    pageSize,
    userFilter,
    actionFilter,
    periodFilter,
    uniqueUsers,
    uniqueActions,
    refresh: fetch,
    exportCsv,
  }
}
