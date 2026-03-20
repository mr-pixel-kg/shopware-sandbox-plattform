import { ref, computed, watch, onUnmounted } from 'vue'
import { formatTtlRemaining } from '@/utils/formatters'

export function useTtlCountdown(expiresAt: () => string | undefined, createdAt: () => string) {
  const now = ref(Date.now())
  let interval: ReturnType<typeof setInterval> | null = null

  function start() {
    if (interval) return
    interval = setInterval(() => {
      now.value = Date.now()
    }, 1000)
  }

  function stop() {
    if (interval) {
      clearInterval(interval)
      interval = null
    }
  }

  const remainingMs = computed(() => {
    const exp = expiresAt()
    if (!exp) return 0
    return Math.max(0, new Date(exp).getTime() - now.value)
  })

  const remainingFormatted = computed(() => {
    const exp = expiresAt()
    if (!exp) return '—'
    return formatTtlRemaining(exp)
  })

  const progressPercent = computed(() => {
    const exp = expiresAt()
    const created = createdAt()
    if (!exp || !created) return 0
    const total = new Date(exp).getTime() - new Date(created).getTime()
    if (total <= 0) return 0
    const elapsed = now.value - new Date(created).getTime()
    return Math.min(100, Math.max(0, ((total - elapsed) / total) * 100))
  })

  const isExpired = computed(() => remainingMs.value <= 0)
  const isWarning = computed(() => remainingMs.value > 0 && remainingMs.value < 30 * 60 * 1000)

  watch(
    () => expiresAt(),
    (val) => {
      if (val && !isExpired.value) start()
      else stop()
    },
    { immediate: true },
  )

  onUnmounted(stop)

  return { remainingMs, remainingFormatted, progressPercent, isExpired, isWarning }
}
