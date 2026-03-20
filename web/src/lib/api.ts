import type {
  ApiErrorResponse,
  AuditLogRecord,
  CreateImagePayload,
  CreateSandboxPayload,
  ImageRecord,
  LoginResponse,
  SandboxRecord,
  User,
} from "@/types/api";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080";

class ApiError extends Error {
  code: string;
  details?: unknown;

  constructor(message: string, code = "UNKNOWN_ERROR", details?: unknown) {
    super(message);
    this.code = code;
    this.details = details;
  }
}

async function request<T>(path: string, init: RequestInit = {}, token?: string): Promise<T> {
  const headers = new Headers(init.headers ?? {});
  headers.set("Content-Type", "application/json");

  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers,
    credentials: "include",
  });

  if (response.status === 204) {
    return undefined as T;
  }

  const body = await response.json().catch(() => null);

  if (!response.ok) {
    const errorBody = body as ApiErrorResponse | null;
    throw new ApiError(
      errorBody?.error?.message ?? "Unexpected request error",
      errorBody?.error?.code ?? "UNKNOWN_ERROR",
      errorBody?.error?.details
    );
  }

  return body as T;
}

export const api = {
  error: ApiError,
  getPublicImages() {
    return request<ImageRecord[]>("/api/public/images");
  },
  getGuestSandboxes() {
    return request<SandboxRecord[]>("/api/public/sandboxes");
  },
  createDemo(payload: CreateSandboxPayload) {
    return request<SandboxRecord>("/api/public/demos", {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },
  deleteGuestSandbox(id: string) {
    return request<void>(`/api/public/sandboxes/${id}`, {
      method: "DELETE",
    });
  },
  login(email: string, password: string) {
    return request<LoginResponse>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
  },
  me(token: string) {
    return request<User>("/api/me", {}, token);
  },
  getImages(token: string) {
    return request<ImageRecord[]>("/api/images", {}, token);
  },
  createImage(token: string, payload: CreateImagePayload) {
    return request<ImageRecord>("/api/images", {
      method: "POST",
      body: JSON.stringify(payload),
    }, token);
  },
  deleteImage(token: string, id: string) {
    return request<void>(`/api/images/${id}`, {
      method: "DELETE",
    }, token);
  },
  getSandboxes(token: string) {
    return request<SandboxRecord[]>("/api/sandboxes", {}, token);
  },
  createSandbox(token: string, payload: CreateSandboxPayload) {
    return request<SandboxRecord>("/api/sandboxes", {
      method: "POST",
      body: JSON.stringify(payload),
    }, token);
  },
  deleteSandbox(token: string, id: string) {
    return request<void>(`/api/sandboxes/${id}`, {
      method: "DELETE",
    }, token);
  },
  snapshotSandbox(token: string, id: string, payload: CreateImagePayload) {
    return request<ImageRecord>(`/api/sandboxes/${id}/snapshot`, {
      method: "POST",
      body: JSON.stringify(payload),
    }, token);
  },
  getAuditLogs(token: string, limit = 100) {
    return request<AuditLogRecord[]>(`/api/audit-logs?limit=${limit}`, {}, token);
  },
};
