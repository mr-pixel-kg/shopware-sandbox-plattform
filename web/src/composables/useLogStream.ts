import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import { Terminal } from '@xterm/xterm'
import { onUnmounted, ref, type Ref } from 'vue'

import { getToken } from '@/utils/storage'

import type { LogEvent } from '@/types'

export const LOG_TERMINAL_BG = '#1e1e1e'

function safeParse(json: string): LogEvent | null {
  try {
    return JSON.parse(json) as LogEvent
  } catch {
    return null
  }
}

export function useLogStream(containerRef: Ref<HTMLElement | null>) {
  const isStreaming = ref(false)
  const error = ref<string | null>(null)

  let terminal: Terminal | null = null
  let fitAddon: FitAddon | null = null
  let abortController: AbortController | null = null
  let resizeObserver: ResizeObserver | null = null

  function initTerminal() {
    if (terminal) return

    const el = containerRef.value
    if (!el) return

    terminal = new Terminal({
      cursorBlink: false,
      disableStdin: true,
      fontSize: 13,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      convertEol: true,
      theme: {
        background: LOG_TERMINAL_BG,
        foreground: '#d4d4d4',
        cursor: LOG_TERMINAL_BG,
        selectionBackground: '#264f78',
      },
    })

    fitAddon = new FitAddon()
    terminal.loadAddon(fitAddon)
    terminal.loadAddon(new WebLinksAddon())

    terminal.open(el)
    fitAddon.fit()

    resizeObserver = new ResizeObserver(() => {
      fitAddon?.fit()
    })
    resizeObserver.observe(el)
  }

  function connect(sandboxId: string, logKey: string) {
    disconnect()
    error.value = null

    if (!containerRef.value) return

    initTerminal()
    terminal?.clear()

    const baseURL = import.meta.env.WEB_API_URL as string
    const token = getToken()

    abortController = new AbortController()

    fetch(`${baseURL}/api/sandboxes/${sandboxId}/logs/${logKey}`, {
      headers: { Authorization: `Bearer ${token}` },
      signal: abortController.signal,
    })
      .then((res) => {
        if (!res.ok) {
          error.value = `HTTP ${res.status}`
          return
        }

        isStreaming.value = true
        const reader = res.body!.getReader()
        const decoder = new TextDecoder()
        let buffer = ''

        function read(): Promise<void> {
          return reader.read().then(({ done, value }) => {
            if (done) {
              isStreaming.value = false
              return
            }

            buffer += decoder.decode(value, { stream: true })
            const lines = buffer.split('\n')
            buffer = lines.pop() ?? ''

            for (const line of lines) {
              if (!line.startsWith('data: ')) continue
              const parsed = safeParse(line.slice(6))
              if (parsed) terminal?.writeln(parsed.line)
            }

            return read()
          })
        }

        return read()
      })
      .catch((err: Error) => {
        if (err.name === 'AbortError') return
        error.value = err.message || 'Verbindung fehlgeschlagen'
      })
      .finally(() => {
        isStreaming.value = false
      })
  }

  function disconnect() {
    abortController?.abort()
    abortController = null
    isStreaming.value = false
  }

  function dispose() {
    disconnect()

    resizeObserver?.disconnect()
    resizeObserver = null

    terminal?.dispose()
    terminal = null
    fitAddon = null
  }

  onUnmounted(dispose)

  return {
    isStreaming,
    error,
    connect,
    dispose,
  }
}
