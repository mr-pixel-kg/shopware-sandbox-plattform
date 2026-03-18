import type { ClassValue } from "clsx"
import { clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatDateTime(value?: string | null) {
  if (!value) return "No expiry"

  return new Intl.DateTimeFormat("de-DE", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value))
}

export function relativeRemaining(value?: string | null) {
  if (!value) return "No expiry"

  const diffMs = new Date(value).getTime() - Date.now()

  if (diffMs <= 0) return "Expired"

  const totalMinutes = Math.floor(diffMs / 60000)
  const hours = Math.floor(totalMinutes / 60)
  const minutes = totalMinutes % 60

  if (hours === 0) return `${minutes}m left`
  if (minutes === 0) return `${hours}h left`

  return `${hours}h ${minutes}m left`
}
