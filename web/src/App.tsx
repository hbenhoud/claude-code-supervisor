import { useState } from 'react'
import './App.css'
import { TopBar } from './components/TopBar'
import { AgentCards } from './components/AgentCards'
import { Floor } from './components/Floor'
import { TimelineRail } from './components/TimelineRail'
import { DetailPanel } from './components/DetailPanel'
import { SessionList } from './components/SessionList'
import { useWebSocket } from './hooks/useWebSocket'
import { useSessions } from './hooks/useSessions'
import { useSupervisorStore } from './store/supervisor'

function App() {
  const activeSessionId = useSupervisorStore(s => s.activeSessionId)
  const [showSessionList, setShowSessionList] = useState(false)

  useSessions()
  useWebSocket()

  if (!activeSessionId || showSessionList) {
    return (
      <div style={{ background: '#0a0a0a', minHeight: '100vh', color: '#e0e0e0', fontFamily: 'monospace' }}>
        <div style={{ padding: '16px', borderBottom: '1px solid #222', display: 'flex', alignItems: 'center', gap: 12 }}>
          <span style={{ fontWeight: 'bold', fontSize: 16 }}>Supervisor</span>
          <span style={{ color: '#666', fontSize: 12 }}>Select a session</span>
        </div>
        <SessionList onSelect={() => setShowSessionList(false)} />
      </div>
    )
  }

  return (
    <div style={{
      background: '#0a0a0a',
      minHeight: '100vh',
      display: 'flex',
      flexDirection: 'column',
      fontFamily: 'monospace',
    }}>
      <TopBar />

      <div style={{ display: 'flex', flex: 1, overflow: 'hidden' }}>
        <AgentCards />

        <div style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
          <Floor />
          <TimelineRail />
        </div>

        <DetailPanel />
      </div>

      {/* Session switcher button */}
      <button
        onClick={() => setShowSessionList(true)}
        style={{
          position: 'fixed',
          bottom: 12,
          left: 12,
          padding: '6px 12px',
          background: '#1e1e1e',
          border: '1px solid #333',
          borderRadius: 6,
          color: '#888',
          fontSize: 11,
          cursor: 'pointer',
          fontFamily: 'monospace',
        }}
      >
        Sessions
      </button>
    </div>
  )
}

export default App
