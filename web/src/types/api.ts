export interface ApiErrorResponse {
  error: {
    code: string;
    message: string;
    details?: unknown;
  };
}

export interface User {
  id: string;
  email: string;
  createdAt: string;
  updatedAt: string;
}

export interface ImageRecord {
  id: string;
  name: string;
  tag: string;
  title?: string | null;
  description?: string | null;
  thumbnailUrl?: string | null;
  isPublic: boolean;
  createdByUserId?: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface SandboxRecord {
  id: string;
  imageId: string;
  createdByUserId?: string | null;
  guestSessionId?: string | null;
  status: string;
  containerId: string;
  containerName: string;
  url: string;
  clientIp: string;
  expiresAt?: string | null;
  createdAt?: string;
  updatedAt?: string;
}

export interface AuditLogRecord {
  id: string;
  userId?: string | null;
  action: string;
  ipAddress?: string | null;
  details?: Record<string, unknown>;
  createdAt: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface CreateImagePayload {
  name: string;
  tag: string;
  title?: string | null;
  description?: string | null;
  thumbnailUrl?: string | null;
  isPublic: boolean;
}

export interface CreateSandboxPayload {
  imageId: string;
  ttlMinutes?: number | null;
}
