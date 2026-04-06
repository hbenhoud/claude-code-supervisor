import { useSupervisorStore } from '../store/supervisor'

interface SessionListProps {
  onSelect: () => void
}

export function SessionList({ onSelect }: SessionListProps) {
  const sessions = useSupervisorStore(s => s.sessions)
  const setActiveSession = useSupervisorStore(s => s.setActiveSession)

  const handleSelect = (id: string) => {
    setActiveSession(id)
    onSelect()
  }

  if (sessions.length === 0) {
    return (
      <div style={{
        padding: 40,
        textAlign: 'center',
        color: '#555',
        fontSize: 13,
      }}>
        <p>No sessions yet.</p>
        <p style={{ marginTop: 8, fontSize: 11 }}>
          Start Claude Code in your terminal — sessions appear automatically.
        </p>
      </div>
    )
  }

  return (
    <div style={{ padding: 16 }}>
      {sessions.map(s => (
        <div
          key={s.id}
          onClick={() => handleSelect(s.id)}
          style={{
            padding: '12px 16px',
            background: '#111',
            border: '1px solid #222',
            borderRadius: 8,
            marginBottom: 8,
            cursor: 'pointer',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            transition: 'border-color 0.15s',
          }}
          onMouseEnter={e => (e.currentTarget.style.borderColor = '#3b82f6')}
          onMouseLeave={e => (e.currentTarget.style.borderColor = '#222')}
        >
          <div>
            <div
              style={{ fontSize: 13, color: '#e0e0e0', marginBottom: 4, userSelect: 'all', cursor: 'text' }}
              onClick={(e) => { e.stopPropagation(); navigator.clipboard.writeText(s.id) }}
              title="Click to copy"
            >
              {s.id}
            </div>
            <div style={{ fontSize: 11, color: '#666' }}>
              {s.cwd || 'unknown directory'}
            </div>
          </div>

          <div style={{ display: 'flex', gap: 16, alignItems: 'center', fontSize: 11 }}>
            <span style={{
              padding: '2px 8px',
              borderRadius: 4,
              background: s.status === 'running' ? '#1a3d2e' : '#1a1a1a',
              color: s.status === 'running' ? '#4ade80' : '#666',
              fontWeight: s.status === 'running' ? 'bold' : 'normal',
            }}>
              {s.status === 'running' && '● '}{s.status}
            </span>
            <span style={{ color: '#888' }}>{s.tool_count} tools</span>
            <span style={{ color: '#888' }}>{s.agent_count} bots</span>
            <span style={{ color: '#666' }}>
              {new Date(s.started_at * 1000).toLocaleTimeString()}
            </span>
          </div>
        </div>
      ))}
    </div>
  )
}
