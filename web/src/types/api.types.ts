interface BaseModel {
  createdAt: string
  updatedAt: string
  deletedAt?: string
}

export interface User extends BaseModel {
  id: string
  email: string
  role: string
}

export type ImageStatus = 'pulling' | 'ready' | 'failed'

export type MetadataType = 'field' | 'setting' | 'info' | 'action'

export interface MetadataItem {
  key: string
  label: string
  type: MetadataType
  value?: string
  input?: string
  required?: boolean
  options?: string[]
  variant?: string
  show?: 'sandbox' | 'template' | 'both'
  condition?: 'ready' | 'always'
  icon?: string
  size?: 'default' | 'icon'
}

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
  metadata?: MetadataItem[]
  registryRef?: string
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
  metadata?: MetadataItem[]
  expiresAt?: string
  lastSeenAt?: string
}

export interface SandboxHealthEvent {
  sandboxId: string
  status: string
  ready: boolean
  url: string
  httpStatus?: number
  latencyMs?: number
  failureReason?: string
  message?: string
  checkedAt: string
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
  metadata?: Record<string, string>
}

export interface CreateImageRequest {
  name: string
  tag: string
  title?: string
  description?: string
  isPublic: boolean
  metadata?: MetadataItem[]
}

export interface UpdateImageRequest {
  title?: string | null
  description?: string | null
  isPublic: boolean
  metadata?: MetadataItem[]
}

export interface AddWhitelistRequest {
  email: string
  role: 'admin' | 'user'
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
