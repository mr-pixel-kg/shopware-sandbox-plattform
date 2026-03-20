export function formatDateTime(iso: string): string {
  const date = new Date(iso)
  return date.toLocaleDateString('de-DE', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function formatRelativeTime(iso: string): string {
  const now = Date.now()
  const diff = now - new Date(iso).getTime()
  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (days > 0) return `vor ${days}d`
  if (hours > 0) return `vor ${hours}h`
  if (minutes > 0) return `vor ${minutes}m`
  return 'gerade eben'
}

export function formatTtlRemaining(expiresAt: string): string {
  const remaining = new Date(expiresAt).getTime() - Date.now()
  if (remaining <= 0) return 'abgelaufen'

  const hours = Math.floor(remaining / (1000 * 60 * 60))
  const minutes = Math.floor((remaining % (1000 * 60 * 60)) / (1000 * 60))

  if (hours > 0) return `${hours}h ${minutes}m`
  return `${minutes}m`
}

export function ttlMinutesToLabel(minutes: number): string {
  if (minutes < 60) return `${minutes} Minuten`
  const hours = minutes / 60
  return `${hours} Stunde${hours > 1 ? 'n' : ''}`
}
