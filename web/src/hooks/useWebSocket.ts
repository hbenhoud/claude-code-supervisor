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

    // Step 1: Fetch all historical events via REST
    fetch(`${API_URL}/api/sessions/${activeSessionId}/events`)
      .then(r => r.json())
      .then((events: SupervisorEvent[]) => {
        // Load all historical events at once
        for (const evt of events) {
          useSupervisorStore.getState().addEvent(evt)
        }

        // Step 2: Connect WebSocket for live updates only (after last known sequence)
        const lastSeq = events.length > 0 ? events[events.length - 1].sequence : 0

        const ws = new WebSocket(WS_URL)
        wsRef.current = ws

        ws.onopen = () => {
          ws.send(JSON.stringify({ subscribe: activeSessionId, afterSequence: lastSeq }))
          useSupervisorStore.getState().setConnected(true)
        }

        ws.onmessage = (msg) => {
          const evt: SupervisorEvent = JSON.parse(msg.data)
          useSupervisorStore.getState().addEvent(evt)
        }

        ws.onclose = () => useSupervisorStore.getState().setConnected(false)
        ws.onerror = () => useSupervisorStore.getState().setConnected(false)
      })
      .catch(() => {
        // If REST fails, fall back to WS-only
        const ws = new WebSocket(WS_URL)
        wsRef.current = ws

        ws.onopen = () => {
          ws.send(JSON.stringify({ subscribe: activeSessionId, afterSequence: 0 }))
          useSupervisorStore.getState().setConnected(true)
        }

        ws.onmessage = (msg) => {
          const evt: SupervisorEvent = JSON.parse(msg.data)
          useSupervisorStore.getState().addEvent(evt)
        }

        ws.onclose = () => useSupervisorStore.getState().setConnected(false)
        ws.onerror = () => useSupervisorStore.getState().setConnected(false)
      })

    return () => {
      if (wsRef.current) {
        wsRef.current.close()
        wsRef.current = null
      }
    }
  }, [activeSessionId])
}
