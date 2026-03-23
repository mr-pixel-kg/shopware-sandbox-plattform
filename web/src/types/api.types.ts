interface BaseModel {
  createdAt: string
  updatedAt: string
  deletedAt?: string
}

export interface User extends BaseModel {
  id: string
  email: string
}

export type ImageStatus = 'pulling' | 'ready' | 'failed'

export interface Image extends BaseModel {
  id: string
  name: string
  tag: string
  title?: string
  description?: string
  thumbnailUrl?: string
  isPublic: boolean
  status: ImageStatus
  error?: string
  createdByUserId?: string
}

export interface PendingPull {
  id: string
  name: string
  tag: string
  title?: string
  percent: number
  status: string
}

export type SandboxStatus = 'starting' | 'running' | 'stopped' | 'expired' | 'deleted' | 'failed'

export interface Sandbox extends BaseModel {
  id: string
  imageId: string
  createdByUserId?: string
  guestSessionId?: string
  status: SandboxStatus
  containerId: string
  containerName: string
  url: string
  port?: number
  clientIp: string
  expiresAt?: string
  lastSeenAt?: string
}

export interface AuditLog {
  id: string
  userId?: string
  action: string
  ipAddress: string
  details: Record<string, unknown> | unknown[]
  createdAt: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
}

export interface CreateSandboxRequest {
  imageId: string
  ttlMinutes?: number
}

export interface CreateImageRequest {
  name: string
  tag: string
  title?: string
  description?: string
  isPublic: boolean
}

export interface UpdateImageRequest {
  title?: string | null
  description?: string | null
  isPublic: boolean
}

export type CreateSnapshotRequest = CreateImageRequest

export interface LoginResponse {
  token: string
  user: User
}

export interface ApiError {
  code: string
  message: string
  details?: unknown
}

export interface ApiErrorResponse {
  error: ApiError
}
