import { useEffect } from 'react'
import { useSupervisorStore } from '../store/supervisor'
import type { Session } from '../types/events'

const API_URL = 'http://localhost:3001'

export function useSessions() {
  const setSessions = useSupervisorStore(s => s.setSessions)
  const sessions = useSupervisorStore(s => s.sessions)
  const activeSessionId = useSupervisorStore(s => s.activeSessionId)
  const setActiveSession = useSupervisorStore(s => s.setActiveSession)

  // Poll sessions
  useEffect(() => {
    const fetchSessions = () => {
      fetch(`${API_URL}/api/sessions`)
        .then(r => r.json())
        .then((data: Session[]) => setSessions(data))
        .catch(() => {})
    }
    fetchSessions()
    const interval = setInterval(fetchSessions, 3000)
    return () => clearInterval(interval)
  }, [setSessions])

  // Auto-select first running session
  useEffect(() => {
    if (!activeSessionId && sessions.length > 0) {
      const running = sessions.find(s => s.status === 'running')
      setActiveSession(running?.id || sessions[0].id)
    }
  }, [sessions, activeSessionId, setActiveSession])
}
