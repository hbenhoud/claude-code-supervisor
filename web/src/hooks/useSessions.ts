import { useEffect } from 'react'
import { useSupervisorStore } from '../store/supervisor'

const API_URL = 'http://localhost:3001'

export function useSessions() {
  const setSessions = useSupervisorStore(s => s.setSessions)

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

}
