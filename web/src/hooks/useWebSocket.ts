import { useEffect, useRef } from 'react'
import { useSupervisorStore } from '../store/supervisor'
import type { SupervisorEvent } from '../types/events'

const API_URL = 'http://localhost:3001'
const WS_URL = 'ws://localhost:3001/ws'

export function useWebSocket() {
  const activeSessionId = useSupervisorStore(s => s.activeSessionId)
  const wsRef = useRef<WebSocket | null>(null)

  useEffect(() => {
    if (!activeSessionId) return

    const store = useSupervisorStore.getState()
    store.clearEvents()

    let cancelled = false
    const controller = new AbortController()

    // Step 1: Fetch all historical events via REST
    fetch(`${API_URL}/api/sessions/${activeSessionId}/events`, { signal: controller.signal })
      .then(r => r.json())
      .then((events: SupervisorEvent[]) => {
        if (cancelled) return

        useSupervisorStore.getState().loadEvents(events)

        // Step 2: Connect WebSocket for live updates
        const lastSeq = events.length > 0 ? events[events.length - 1].sequence : 0
        connectWs(activeSessionId, lastSeq)
      })
      .catch(() => {
        if (cancelled) return
        // REST failed — connect WS from scratch
        connectWs(activeSessionId, 0)
      })

    function connectWs(sessionId: string, afterSequence: number) {
      if (cancelled) return

      const ws = new WebSocket(WS_URL)
      wsRef.current = ws

      ws.onopen = () => {
        if (cancelled) { ws.close(); return }
        ws.send(JSON.stringify({ subscribe: sessionId, afterSequence }))
        useSupervisorStore.getState().setConnected(true)
      }

      ws.onmessage = (msg) => {
        if (cancelled) return
        const evt: SupervisorEvent = JSON.parse(msg.data)
        useSupervisorStore.getState().addEvent(evt)
      }

      ws.onclose = () => useSupervisorStore.getState().setConnected(false)
      ws.onerror = () => useSupervisorStore.getState().setConnected(false)
    }

    return () => {
      cancelled = true
      controller.abort()
      if (wsRef.current) {
        wsRef.current.close()
        wsRef.current = null
      }
    }
  }, [activeSessionId])
}
