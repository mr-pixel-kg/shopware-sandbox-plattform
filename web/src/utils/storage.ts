const TOKEN_KEY = 'auth_token'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY)
}

function decodePayload(token: string): Record<string, unknown> | null {
  try {
    const base64 = token.split('.')[1]
    if (!base64) return null
    const json = atob(base64.replace(/-/g, '+').replace(/_/g, '/'))
    return JSON.parse(json)
  } catch {
    return null
  }
}

export function isTokenExpired(token: string): boolean {
  const payload = decodePayload(token)
  if (!payload || typeof payload.exp !== 'number') return true
  return Date.now() >= payload.exp * 1000
}

export function getTokenExpiryMs(token: string): number {
  const payload = decodePayload(token)
  if (!payload || typeof payload.exp !== 'number') return 0
  return Math.max(0, payload.exp * 1000 - Date.now())
}

export function setupStorageListener(onLogout: () => void): () => void {
  const handler = (e: StorageEvent) => {
    if (e.key === TOKEN_KEY && e.newValue === null) {
      onLogout()
    }
  }
  window.addEventListener('storage', handler)
  return () => window.removeEventListener('storage', handler)
}
