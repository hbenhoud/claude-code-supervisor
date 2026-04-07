import { useSupervisorStore } from '../store/supervisor'
import { Bot } from './Bot'
import type { BotState } from '../types/events'

interface SessionListProps {
  onSelect: () => void
}

function sessionState(status: string): BotState {
  if (status === 'running') return 'working'
  if (status === 'error') return 'error'
  return 'done'
}

const API_URL = 'http://localhost:3001'

export function SessionList({ onSelect }: SessionListProps) {
  const sessions = useSupervisorStore(s => s.sessions)
  const setActiveSession = useSupervisorStore(s => s.setActiveSession)
  const removeSession = useSupervisorStore(s => s.removeSession)

  const handleSelect = (id: string) => {
    setActiveSession(id)
    onSelect()
  }

  const handleDelete = (e: React.MouseEvent, id: string) => {
    e.stopPropagation()
    if (!confirm('Delete this session and all its data?')) return
    fetch(`${API_URL}/api/sessions/${id}`, { method: 'DELETE' })
      .then(() => removeSession(id))
      .catch(() => {})
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
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <div style={{ flexShrink: 0 }}>
              <Bot type="root" name="" state={sessionState(s.status)} size={36} />
            </div>
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
            <span style={{ color: '#888' }}>{s.agent_count} agents</span>
            <span style={{ color: '#666' }}>
              {new Date(s.started_at * 1000).toLocaleTimeString()}
            </span>
            <button
              onClick={(e) => handleDelete(e, s.id)}
              title="Delete session"
              style={{
                background: 'transparent',
                border: 'none',
                color: '#666',
                cursor: 'pointer',
                fontSize: 14,
                padding: '2px 6px',
                borderRadius: 4,
                fontFamily: 'monospace',
              }}
              onMouseEnter={e => (e.currentTarget.style.color = '#ef4444')}
              onMouseLeave={e => (e.currentTarget.style.color = '#666')}
            >
              x
            </button>
          </div>
        </div>
      ))}
    </div>
  )
}
