import { useSupervisorStore } from '../store/supervisor'

export function DetailPanel() {
  const events = useSupervisorStore(s => s.events)
  const selectedEventId = useSupervisorStore(s => s.selectedEventId)
  const selectedAgentId = useSupervisorStore(s => s.selectedAgentId)
  const agents = useSupervisorStore(s => s.agents)

  const selectedEvent = selectedEventId ? events.find(e => e.id === selectedEventId) : null
  const selectedAgent = selectedAgentId ? agents.get(selectedAgentId) : null

  // If an agent is selected but no event, show agent info
  if (!selectedEvent && selectedAgent) {
    const agentEvents = events.filter(e => e.agent_id === selectedAgent.id)
    return (
      <div style={{ width: 280, borderLeft: '1px solid #222', padding: 12, overflowY: 'auto', fontSize: 12 }}>
        <h3 style={{ fontSize: 14, margin: '0 0 12px', color: '#e0e0e0' }}>{selectedAgent.name}</h3>
        <Field label="Type" value={selectedAgent.type} />
        <Field label="State" value={selectedAgent.state} />
        <Field label="Tools used" value={String(selectedAgent.toolCount)} />
        {selectedAgent.currentTool && <Field label="Current" value={selectedAgent.currentTool} />}
        <h4 style={{ fontSize: 12, margin: '12px 0 8px', color: '#888' }}>Recent actions</h4>
        {agentEvents.slice(-5).map(e => (
          <div key={e.id} style={{ padding: '4px 0', borderBottom: '1px solid #1a1a1a', color: '#888', fontSize: 11 }}>
            {e.tool_name || e.event_type} {e.duration_ms != null && `(${e.duration_ms}ms)`}
          </div>
        ))}
      </div>
    )
  }

  if (!selectedEvent) {
    return (
      <div style={{
        width: 280,
        borderLeft: '1px solid #222',
        padding: 12,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        color: '#555',
        fontSize: 12,
        fontFamily: 'monospace',
      }}>
        Click a bot or timeline node to see details
      </div>
    )
  }

  return (
    <div style={{ width: 280, borderLeft: '1px solid #222', padding: 12, overflowY: 'auto', fontSize: 12 }}>
      <h3 style={{ fontSize: 14, margin: '0 0 12px', color: '#e0e0e0' }}>
        {selectedEvent.tool_name || selectedEvent.event_type}
      </h3>

      <Field label="Type" value={`${selectedEvent.event_type} / ${selectedEvent.event_subtype}`} />
      <Field label="Agent" value={selectedEvent.agent_id} />
      <Field label="Time" value={new Date(selectedEvent.timestamp).toLocaleTimeString()} />
      {selectedEvent.duration_ms != null && (
        <Field label="Duration" value={`${selectedEvent.duration_ms}ms`} />
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

function Field({ label, value, color }: { label: string; value: string; color?: string }) {
  return (
    <div style={{ marginBottom: 8 }}>
      <span style={{ color: '#666', fontSize: 10 }}>{label}</span>
      <div style={{ color: color || '#ccc', fontFamily: 'monospace', wordBreak: 'break-all' }}>{value}</div>
    </div>
  )
}

function CollapsibleJson({ label, data }: { label: string; data: unknown }) {
  const json = typeof data === 'string' ? data : JSON.stringify(data, null, 2)
  const truncated = json.length > 500

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
      }}>
        {truncated ? json.slice(0, 500) + '\n...' : json}
      </pre>
    </div>
  )
}
