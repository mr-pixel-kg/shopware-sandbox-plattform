import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import { Terminal } from '@xterm/xterm'
import { onUnmounted, ref, type Ref } from 'vue'

import { getToken } from '@/utils/storage'

export const TERMINAL_BG = '#1e1e1e'

const encoder = new TextEncoder()

interface ControlMessage {
  type?: string
  message?: string
}

function parseControlMessage(text: string): ControlMessage | null {
  try {
    return JSON.parse(text) as ControlMessage
  } catch {
    return null
  }
}

export function useTerminal(
  sandboxId: Ref<string | undefined>,
  containerRef: Ref<HTMLElement | null>,
) {
  const isConnected = ref(false)
  const error = ref<string | null>(null)

  let terminal: Terminal | null = null
  let fitAddon: FitAddon | null = null
  let ws: WebSocket | null = null
  let resizeObserver: ResizeObserver | null = null

  function buildWsUrl(id: string, cols: number, rows: number): string {
    const base = import.meta.env.WEB_API_URL as string
    const wsBase = base.replace(/^http/, 'ws')
    const token = getToken()
    return `${wsBase}/api/sandboxes/${id}/terminal?access_token=${encodeURIComponent(token ?? '')}&cols=${cols}&rows=${rows}`
  }

  function connect() {
    disconnect()
    error.value = null

    const id = sandboxId.value
    if (!id || !containerRef.value) return

    terminal = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: {
        background: TERMINAL_BG,
        foreground: '#d4d4d4',
        cursor: '#d4d4d4',
        selectionBackground: '#264f78',
      },
    })

    fitAddon = new FitAddon()
    terminal.loadAddon(fitAddon)
    terminal.loadAddon(new WebLinksAddon())

    terminal.open(containerRef.value)
    fitAddon.fit()

    const cols = terminal.cols
    const rows = terminal.rows

    ws = new WebSocket(buildWsUrl(id, cols, rows))
    ws.binaryType = 'arraybuffer'

    ws.onopen = () => {
      isConnected.value = true
      terminal?.focus()
    }

    ws.onmessage = (event: MessageEvent) => {
      if (event.data instanceof ArrayBuffer) {
        terminal?.write(new Uint8Array(event.data))
      } else if (typeof event.data === 'string') {
        const msg = parseControlMessage(event.data)
        if (msg?.type === 'error') {
          error.value = msg.message ?? 'Connection lost'
          terminal?.write(`\r\n\x1b[31m${msg.message ?? 'Connection lost'}\x1b[0m\r\n`)
        } else if (msg?.type === 'exit') {
          terminal?.write('\r\n\x1b[90mSession ended.\x1b[0m\r\n')
        }
      }
    }

    ws.onclose = (event: CloseEvent) => {
      isConnected.value = false
      if (!error.value && event.code !== 1000) {
        error.value = event.reason || 'Verbindung geschlossen'
      }
    }

    ws.onerror = () => {
      if (!error.value) {
        error.value = 'Verbindung zum Terminal fehlgeschlagen'
      }
      isConnected.value = false
    }

    terminal.onData((data: string) => {
      if (ws?.readyState === WebSocket.OPEN) {
        ws.send(encoder.encode(data))
      }
    })

    terminal.onResize(({ cols, rows }) => {
      if (ws?.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'resize', cols, rows }))
      }
    })

    resizeObserver = new ResizeObserver(() => {
      fitAddon?.fit()
    })
    resizeObserver.observe(containerRef.value)
  }

  function disconnect() {
    resizeObserver?.disconnect()
    resizeObserver = null

    if (ws) {
      ws.onclose = null
      ws.onerror = null
      ws.onmessage = null
      if (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING) {
        ws.close()
      }
      ws = null
    }

    terminal?.dispose()
    terminal = null
    fitAddon = null
    isConnected.value = false
  }

  onUnmounted(disconnect)

  return {
    isConnected,
    error,
    connect,
    disconnect,
  }
}
