import { useEffect, useRef } from 'react'
import { useSupervisorStore } from '../store/supervisor'

function formatMs(ms: number): string {
  if (ms < 1000) return `${ms}ms`
  const s = ms / 1000
  if (s < 60) return `${s.toFixed(1)}s`
  const m = Math.floor(s / 60)
  const sec = Math.round(s % 60)
  return `${m}m ${sec}s`
}

export function DetailPanel() {
  const events = useSupervisorStore(s => s.events)
  const selectedEventId = useSupervisorStore(s => s.selectedEventId)
  const selectedAgentId = useSupervisorStore(s => s.selectedAgentId)
  const selectEvent = useSupervisorStore(s => s.selectEvent)
  const agents = useSupervisorStore(s => s.agents)

  const selectedEvent = selectedEventId ? events.find(e => e.id === selectedEventId) : null
  const selectedAgent = selectedAgentId ? agents.get(selectedAgentId) : null

  // If an agent is selected but no event, show agent info
  if (!selectedEvent && selectedAgent) {
    const agentEvents = events.filter(e => e.agent_id === selectedAgent.id)
    // Find the agent_spawn event to get description/prompt
    const spawnId = selectedAgent.id.replace('agent-', '')
    const spawnEvent = events.find(e =>
      e.event_type === 'agent_spawn' && e.event_subtype === 'start' && e.tool_use_id?.startsWith(spawnId)
    )
    const spawnInput = spawnEvent?.tool_input as Record<string, unknown> | undefined
    const description = spawnInput?.description as string | undefined
    const prompt = spawnInput?.prompt as string | undefined

    return (
      <div style={{ width: 420, flexShrink: 0, borderLeft: '1px solid #222', padding: 12, overflowY: 'auto', fontSize: 12 }}>
        <h3 style={{ fontSize: 14, margin: '0 0 12px', color: '#e0e0e0' }}>{selectedAgent.name}</h3>
        <Field label="Type" value={selectedAgent.type} />
        <Field label="State" value={selectedAgent.state} />
        {description && <Field label="Task" value={description} />}
        <Field label="Tools used" value={String(selectedAgent.toolCount)} />
        {selectedAgent.currentTool && <Field label="Current" value={selectedAgent.currentTool} />}
        <h4 style={{ fontSize: 12, margin: '12px 0 8px', color: '#888' }}>Actions</h4>
        <ActionsScroll>
          {(() => {
            // Group start/complete by tool_use_id, sorted by first appearance
            const grouped = new Map<string, { id: string; toolUseId: string; toolName: string; subtype: string; durationMs?: number; timestamp: number }>()
            const sorted = [...agentEvents].sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime())
            for (const e of sorted) {
              const key = e.tool_use_id || e.id
              const existing = grouped.get(key)
              if (existing) {
                if (e.event_subtype === 'complete') {
                  existing.subtype = 'complete'
                  existing.id = e.id
                  if (e.duration_ms != null) existing.durationMs = e.duration_ms
                }
              } else {
                grouped.set(key, { id: e.id, toolUseId: key, toolName: e.tool_name || e.event_type, subtype: e.event_subtype, durationMs: e.duration_ms ?? undefined, timestamp: new Date(e.timestamp).getTime() })
              }
            }
            // Orphan starts: if a start has no complete but later events exist, mark as complete
            const entries = Array.from(grouped.values())
            const lastTimestamp = sorted.length > 0 ? new Date(sorted[sorted.length - 1].timestamp).getTime() : 0
            for (const entry of entries) {
              if (entry.subtype === 'start' && entry.timestamp < lastTimestamp) {
                entry.subtype = 'complete'
              }
            }
            return entries
          })().map(a => (
            <div
              key={a.toolUseId}
              onClick={() => selectEvent(a.id)}
              style={{
                padding: '4px 6px',
                borderBottom: '1px solid #1a1a1a',
                color: '#888',
                fontSize: 11,
                cursor: 'pointer',
                borderRadius: 2,
              }}
              onMouseEnter={el => (el.currentTarget.style.background = '#1a1a2e')}
              onMouseLeave={el => (el.currentTarget.style.background = 'transparent')}
            >
              <span style={{ color: a.subtype === 'complete' ? '#22c55e' : a.toolName === 'notification' ? '#a855f7' : '#3b82f6', marginRight: 4 }}>
                {a.subtype === 'complete' ? '✓' : a.toolName === 'notification' ? '●' : '▶'}
              </span>
              {a.toolName}
              {a.durationMs != null && <span style={{ color: '#555' }}> ({formatMs(a.durationMs)})</span>}
            </div>
          ))}
        </ActionsScroll>
        {prompt && agentEvents.length === 0 && (
          <CollapsibleJson label="Prompt" data={prompt} />
        )}
      </div>
    )
  }

  if (!selectedEvent) {
    return (
      <div style={{
        width: 420,
        borderLeft: '1px solid #222',
        padding: 12,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        color: '#555',
        fontSize: 12,
        fontFamily: 'monospace',
      }}>
        Click an agent or timeline node to see details
      </div>
    )
  }

  return (
    <div style={{ width: 420, flexShrink: 0, borderLeft: '1px solid #222', padding: 12, overflowY: 'auto', fontSize: 12 }}>
      <h3 style={{ fontSize: 14, margin: '0 0 12px', color: '#e0e0e0' }}>
        {selectedEvent.tool_name || selectedEvent.event_type}
      </h3>

      <Field label="Type" value={`${selectedEvent.event_type} / ${selectedEvent.event_subtype}`} />
      <Field label="Agent" value={selectedEvent.agent_id} />
      <Field label="Time" value={new Date(selectedEvent.timestamp).toLocaleTimeString()} />
      {selectedEvent.duration_ms != null && (
        <Field label="Duration" value={formatMs(selectedEvent.duration_ms)} />
      )}
      {selectedEvent.error && (
        <Field label="Error" value={selectedEvent.error} color="#ef4444" />
      )}

      {selectedEvent.tool_input != null && (
        <CollapsibleJson label="Input" data={selectedEvent.tool_input as Record<string, unknown>} />
      )}
      {selectedEvent.tool_output != null && (
        <CollapsibleJson label="Output" data={selectedEvent.tool_output as Record<string, unknown>} />
      )}
    </div>
  )
}

function ActionsScroll({ children }: { children: React.ReactNode }) {
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (ref.current) {
      ref.current.scrollTop = ref.current.scrollHeight
    }
  })
  return (
    <div ref={ref} style={{ maxHeight: 'calc(100vh - 350px)', overflowY: 'auto' }}>
      {children}
    </div>
  )
}

function Field({ label, value, color }: { label: string; value: string; color?: string }) {
  return (
    <div style={{ marginBottom: 8 }}>
      <span style={{ color: '#666', fontSize: 10 }}>{label}</span>
      <div style={{ color: color || '#ccc', fontFamily: 'monospace', wordBreak: 'break-all' }}>{value}</div>
    </div>
  )
}

function CollapsibleJson({ label, data }: { label: string; data: unknown }) {
  // If data is an object, render each key/value as a structured field
  if (data && typeof data === 'object' && !Array.isArray(data)) {
    const entries = Object.entries(data as Record<string, unknown>)
    if (entries.length > 0) {
      return (
        <div style={{ marginTop: 12 }}>
          <span style={{ color: '#666', fontSize: 10 }}>{label}</span>
          <div style={{ marginTop: 4, borderLeft: '2px solid #222', paddingLeft: 8 }}>
            {entries.map(([key, val]) => {
              const str = typeof val === 'string' ? val : JSON.stringify(val, null, 2)
              const isLong = str.length > 100
              return (
                <div key={key} style={{ marginBottom: 6 }}>
                  <span style={{ color: '#60a5fa', fontSize: 10 }}>{key}</span>
                  {isLong ? (
                    <pre style={{
                      background: '#0a0a0a',
                      border: '1px solid #1a1a1a',
                      borderRadius: 4,
                      padding: 6,
                      fontSize: 10,
                      color: '#aaa',
                      overflow: 'auto',
                      maxHeight: 150,
                      whiteSpace: 'pre-wrap',
                      wordBreak: 'break-all',
                      margin: '2px 0 0',
                    }}>{str}</pre>
                  ) : (
                    <div style={{ color: '#ccc', fontSize: 11, fontFamily: 'monospace', wordBreak: 'break-all' }}>{str}</div>
                  )}
                </div>
              )
            })}
          </div>
        </div>
      )
    }
  }

  // Fallback for strings/arrays/primitives
  const json = typeof data === 'string' ? data : JSON.stringify(data, null, 2)
  return (
    <div style={{ marginTop: 12 }}>
      <span style={{ color: '#666', fontSize: 10 }}>{label}</span>
      <pre style={{
        background: '#0a0a0a',
        border: '1px solid #222',
        borderRadius: 4,
        padding: 8,
        fontSize: 10,
        color: '#aaa',
        overflow: 'auto',
        maxHeight: 200,
        whiteSpace: 'pre-wrap',
        wordBreak: 'break-all',
        margin: '4px 0 0',
      }}>{json}</pre>
    </div>
  )
}
