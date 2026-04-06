import { useEffect, useRef } from 'react'
import { motion } from 'framer-motion'
import { useSupervisorStore } from '../store/supervisor'

const NODE_COLORS = {
  start: { bg: '#1e3a5f', border: '#3b82f6' },
  complete: { bg: '#1a3d2e', border: '#22c55e' },
  error: { bg: '#3d1a1a', border: '#ef4444' },
  agent_spawn: { bg: '#3d2e1a', border: '#f97316' },
  notification: { bg: '#3d1a3d', border: '#a855f7' },
}

function getNodeStyle(evt: { event_type: string; event_subtype: string; error?: string }) {
  if (evt.error) return NODE_COLORS.error
  if (evt.event_type === 'agent_spawn') return NODE_COLORS.agent_spawn
  if (evt.event_type === 'notification') return NODE_COLORS.notification
  if (evt.event_subtype === 'start') return NODE_COLORS.start
  return NODE_COLORS.complete
}

export function TimelineRail() {
  const events = useSupervisorStore(s => s.events)
  const selectedEventId = useSupervisorStore(s => s.selectedEventId)
  const selectEvent = useSupervisorStore(s => s.selectEvent)
  const scrollRef = useRef<HTMLDivElement>(null)

  // Auto-scroll to latest
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollLeft = scrollRef.current.scrollWidth
    }
  }, [events])

  if (events.length === 0) return null

  return (
    <div style={{
      borderTop: '1px solid #222',
      padding: '8px 12px',
      background: '#0d0d0d',
    }}>
      <div
        ref={scrollRef}
        style={{
          display: 'flex',
          gap: 4,
          overflowX: 'auto',
          scrollBehavior: 'smooth',
          paddingBottom: 4,
        }}
      >
        {events.map(evt => {
          const style = getNodeStyle(evt)
          const isSelected = evt.id === selectedEventId

          return (
            <motion.div
              key={evt.id}
              initial={{ scale: 0, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              title={`${evt.tool_name || evt.event_type} ${evt.duration_ms != null ? `(${evt.duration_ms}ms)` : ''}`}
              onClick={() => selectEvent(evt.id)}
              style={{
                width: 14,
                height: 14,
                borderRadius: '50%',
                background: style.bg,
                border: `2px solid ${style.border}`,
                cursor: 'pointer',
                flexShrink: 0,
                boxShadow: isSelected ? `0 0 8px ${style.border}` : 'none',
                transform: isSelected ? 'scale(1.3)' : 'scale(1)',
                transition: 'transform 0.15s, box-shadow 0.15s',
              }}
            />
          )
        })}
      </div>
    </div>
  )
}
