import { useEffect, useState } from 'react'
import { useSupervisorStore } from '../store/supervisor'

function Tooltip({ children, text }: { children: React.ReactNode; text: string }) {
  const [show, setShow] = useState(false)

  return (
    <div
      style={{ position: 'relative', display: 'inline-flex', alignItems: 'center', gap: 4 }}
      onMouseEnter={() => setShow(true)}
      onMouseLeave={() => setShow(false)}
    >
      {children}
      <span style={{ fontSize: 10, color: '#555', cursor: 'help' }}>i</span>
      {show && (
        <div style={{
          position: 'absolute',
          top: '100%',
          left: '50%',
          transform: 'translateX(-50%)',
          marginTop: 6,
          padding: '8px 12px',
          background: '#1e1e2e',
          border: '1px solid #333',
          borderRadius: 6,
          fontSize: 11,
          color: '#ccc',
          whiteSpace: 'pre-line',
          width: 220,
          zIndex: 100,
          boxShadow: '0 4px 12px rgba(0,0,0,0.5)',
          lineHeight: 1.5,
        }}>
          {text}
        </div>
      )}
    </div>
  )
}

export function TopBar() {
  const connected = useSupervisorStore(s => s.connected)
  const events = useSupervisorStore(s => s.events)
  const agents = useSupervisorStore(s => s.agents)
  const activeSessionId = useSupervisorStore(s => s.activeSessionId)
  const [elapsed, setElapsed] = useState(0)

  const toolCount = events.filter(e => e.event_type === 'tool_call' && e.event_subtype === 'complete').length
  const botCount = agents.size
  const firstEvent = events[0]

  // Tick duration
  useEffect(() => {
    if (!firstEvent) { setElapsed(0); return }
    const start = new Date(firstEvent.timestamp).getTime()
    const tick = () => setElapsed(Math.floor((Date.now() - start) / 1000))
    tick()
    const interval = setInterval(tick, 1000)
    return () => clearInterval(interval)
  }, [firstEvent])

  const formatDuration = (s: number) => {
    const m = Math.floor(s / 60)
    const sec = s % 60
    return m > 0 ? `${m}m ${sec}s` : `${sec}s`
  }

  return (
    <div style={{
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      padding: '8px 16px',
      borderBottom: '1px solid #222',
      background: '#0d0d0d',
      fontFamily: 'monospace',
      fontSize: 12,
    }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
        <span style={{ fontWeight: 'bold', fontSize: 14, color: '#e0e0e0' }}>
          Supervisor
        </span>
        <span style={{
          width: 8, height: 8, borderRadius: '50%',
          background: connected ? '#4ade80' : '#ef4444',
        }} />
        {activeSessionId && (
          <span
            onClick={() => navigator.clipboard.writeText(activeSessionId)}
            title="Click to copy"
            style={{ color: '#666', cursor: 'pointer', userSelect: 'all' }}
          >
            {activeSessionId}
          </span>
        )}
      </div>

      <div style={{ display: 'flex', gap: 20, alignItems: 'center' }}>
        <Tooltip text="Total tool calls completed in this session.">
          <span style={{ color: '#888' }}>Tools</span>
          <span style={{ color: '#e0e0e0', fontWeight: 'bold' }}>{toolCount}</span>
        </Tooltip>

        <Tooltip text="Wall-clock time since first event.\nIncludes time between tool calls.">
          <span style={{ color: '#888' }}>Duration</span>
          <span style={{ color: '#e0e0e0', fontWeight: 'bold' }}>{formatDuration(elapsed)}</span>
        </Tooltip>

        <Tooltip text="Number of active agents.\nRoot + sub-agents currently in the session.">
          <span style={{ color: '#888' }}>Bots</span>
          <span style={{ color: '#e0e0e0', fontWeight: 'bold' }}>{botCount}</span>
        </Tooltip>

        <Tooltip text="Available after session ends.\nCost and token data comes from Claude Code logs.">
          <span style={{ color: '#555' }}>Cost</span>
          <span style={{ color: '#444' }}>--</span>
        </Tooltip>

        <Tooltip text="Available after session ends.\nToken count comes from Claude Code logs.">
          <span style={{ color: '#555' }}>Tokens</span>
          <span style={{ color: '#444' }}>--</span>
        </Tooltip>
      </div>
    </div>
  )
}
