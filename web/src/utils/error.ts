import axios from 'axios'
import type { ApiErrorResponse } from '@/types'

export function getApiErrorMessage(
  error: unknown,
  fallback = 'An unexpected error occurred',
): string {
  if (axios.isAxiosError(error)) {
    const data = error.response?.data as ApiErrorResponse | undefined
    if (data?.error?.message) return data.error.message
    if (error.message) return error.message
  }

  if (error instanceof Error) return error.message

  return fallback
}
