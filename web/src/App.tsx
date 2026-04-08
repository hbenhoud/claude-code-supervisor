import { useState, useCallback, useRef } from 'react'
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
  const [detailWidth, setDetailWidth] = useState(420)
  const dragging = useRef(false)

  const DEFAULT_WIDTH = 420
  const MIN_WIDTH = 300
  const MAX_RATIO = 0.7

  const handleMouseDown = useCallback(() => {
    dragging.current = true
    document.body.style.cursor = 'col-resize'
    document.body.style.userSelect = 'none'

    const onMouseMove = (e: MouseEvent) => {
      if (!dragging.current) return
      const newWidth = Math.max(MIN_WIDTH, Math.min(window.innerWidth * MAX_RATIO, window.innerWidth - e.clientX))
      setDetailWidth(newWidth)
    }

    const onMouseUp = () => {
      dragging.current = false
      document.body.style.cursor = ''
      document.body.style.userSelect = ''
      window.removeEventListener('mousemove', onMouseMove)
      window.removeEventListener('mouseup', onMouseUp)
    }

    window.addEventListener('mousemove', onMouseMove)
    window.addEventListener('mouseup', onMouseUp)
  }, [])

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
        <AgentCards onShowSessions={() => {
          useSupervisorStore.getState().setActiveSession(null)
          setShowSessionList(true)
        }} />

        <div style={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column' }}>
          <Floor />
          <TimelineRail />
        </div>

        <div
          onMouseDown={handleMouseDown}
          onDoubleClick={() => setDetailWidth(DEFAULT_WIDTH)}
          style={{
            width: 4,
            flexShrink: 0,
            cursor: 'col-resize',
            background: '#222',
          }}
          onMouseEnter={e => (e.currentTarget.style.background = '#3b82f6')}
          onMouseLeave={e => (e.currentTarget.style.background = '#222')}
        />
        <DetailPanel width={detailWidth} />
      </div>

    </div>
  )
}

export default App
