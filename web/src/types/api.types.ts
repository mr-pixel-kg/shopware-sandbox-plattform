interface BaseModel {
  createdAt: string
  updatedAt: string
}

export interface User extends BaseModel {
  id: string
  email: string
  avatarUrl: string
  role: string
  isPending: boolean
}

export interface UserSummary {
  id: string
  email: string
  avatarUrl: string
}

export type ImageStatus = 'pulling' | 'committing' | 'ready' | 'failed'

export type MetadataContext =
  | 'image.create'
  | 'image.edit'
  | 'image.card'
  | 'sandbox.create'
  | 'sandbox.card'
  | 'sandbox.details'

export type FieldInput =
  | 'text'
  | 'password'
  | 'number'
  | 'email'
  | 'url'
  | 'select'
  | 'multiselect'
  | 'toggle'
  | 'textarea'

export type DisplayFormat = 'text' | 'code' | 'badge' | 'link' | 'password'

export type ActionVariant = 'default' | 'outline' | 'destructive'
export type ActionSize = 'default' | 'icon'
export type ActionTarget = '_blank' | '_self'

export interface SelectOption {
  value: string
  label: string
}

export interface FieldDependency {
  field: string
  value: string
}

export interface VisibilityRule {
  contexts?: MetadataContext[]
  condition?: string
  dependsOn?: FieldDependency
}

export interface FieldSpec {
  input: FieldInput
  default?: string
  placeholder?: string
  helpText?: string
  required?: boolean
  readOnly?: boolean
  options?: SelectOption[]
}

export interface ActionSpec {
  url: string
  variant?: ActionVariant
  size?: ActionSize
  target?: ActionTarget
  confirm?: string
}

export interface DisplaySpec {
  value: string
  format?: DisplayFormat
  copyable?: boolean
}

interface BaseMetadataItem {
  key: string
  label: string
  icon?: string
  group?: string
  visibility?: VisibilityRule
}

export interface FieldItem extends BaseMetadataItem {
  type: 'field'
  field: FieldSpec
}

export interface ActionItem extends BaseMetadataItem {
  type: 'action'
  action: ActionSpec
}

export interface DisplayItem extends BaseMetadataItem {
  type: 'display'
  display: DisplaySpec
}

export type MetadataItem = FieldItem | ActionItem | DisplayItem

export interface MetadataGroup {
  key: string
  label: string
  description?: string
}

export interface MetadataSchema {
  groups?: MetadataGroup[]
  items: MetadataItem[]
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
  metadata: MetadataItem[]
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
  clientId?: string
  displayName: string
  status: SandboxStatus
  stateReason?: string
  containerId: string
  containerName: string
  url: string
  port?: number
  ssh?: SSHConnection
  clientIp: string
  metadata: MetadataItem[]
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

export type LogSourceType = 'docker' | 'file' | 'lifecycle'

export interface LogSource {
  key: string
  label: string
  type: LogSourceType
}

export interface LogEvent {
  line: string
}

export interface AuditLog {
  id: string
  user?: {
    id: string
    email: string
  } | null
  action: string
  ipAddress?: string | null
  userAgent?: string | null
  clientId?: string | null
  resourceType?: string | null
  resourceId?: string | null
  details: Record<string, unknown> | unknown[]
  timestamp: string
}

export interface PaginationParams {
  limit?: number
  offset?: number
}

export interface PaginationMeta {
  limit: number
  offset: number
  count: number
  total: number
  hasMore: boolean
}

export interface PaginatedResponse<T> {
  data: T[]
  meta: {
    pagination: PaginationMeta
  }
}

export interface AuditLogListFilters {
  userId?: string | null
  action?: string | null
  resourceType?: string | null
  resourceId?: string | null
  clientId?: string | null
  from?: string | null
  to?: string | null
}

export interface AuditLogListMeta {
  pagination: PaginationMeta
  filters: AuditLogListFilters
}

export interface AuditLogListResponse {
  data: AuditLog[]
  meta: AuditLogListMeta
}

export interface AuditLogFacetsResponse {
  users: UserSummary[]
  actions: string[]
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
  ttlMinutes?: number
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
  // Legacy format
  error?: ApiError
  // RFC 7807 problem+json (Fuego format)
  title?: string
  status?: number
  detail?: string
}
