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

export interface ManagedUser extends User {
  pending: boolean
}

export interface UserSummary {
  id: string
  email: string
}

export type ImageStatus = 'pulling' | 'committing' | 'ready' | 'failed'

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
  owner?: UserSummary | null
}

export interface PendingImage {
  id: string
  name: string
  tag: string
  title?: string
  percent: number
  status: ImageStatus
}

export type SandboxStatus =
  | 'starting'
  | 'running'
  | 'paused'
  | 'stopping'
  | 'stopped'
  | 'expired'
  | 'deleted'
  | 'failed'

export interface SSHConnection {
  host: string
  port: number
  username: string
  password: string
  command: string
}

export interface Sandbox extends BaseModel {
  id: string
  imageId: string
  owner?: UserSummary | null
  guestSessionId?: string
  displayName: string
  status: SandboxStatus
  stateReason?: string
  containerId: string
  containerName: string
  url: string
  port?: number
  ssh?: SSHConnection
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
  user?: {
    id: string
    email: string
  } | null
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
  displayName?: string
  metadata?: Record<string, string>
}

export interface UpdateSandboxRequest {
  displayName?: string
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

export interface CreateUserRequest {
  email: string
  role: 'admin' | 'user'
  password?: string
}

export interface UpdateUserRequest {
  email: string
  role: 'admin' | 'user'
  password?: string
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
