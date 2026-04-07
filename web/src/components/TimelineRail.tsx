import { useEffect, useRef, useState, useCallback, useMemo } from 'react'
import { useSupervisorStore } from '../store/supervisor'
import type { SupervisorEvent } from '../types/events'

const NODE_SIZE = 14
const NODE_GAP = 4
const NODE_STEP = NODE_SIZE + NODE_GAP // 18px per node
const OVERSCAN = 20 // extra nodes rendered outside viewport

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

function TimelineNode({
  evt,
  isSelected,
  dimmed,
  left,
  onSelect,
}: {
  evt: SupervisorEvent
  isSelected: boolean
  dimmed: boolean
  left: number
  onSelect: (id: string) => void
}) {
  const style = getNodeStyle(evt)

  return (
    <div
      title={`${evt.tool_name || evt.event_type} ${evt.duration_ms != null ? `(${evt.duration_ms}ms)` : ''}`}
      onClick={() => onSelect(evt.id)}
      style={{
        position: 'absolute',
        left,
        top: 0,
        width: NODE_SIZE,
        height: NODE_SIZE,
        borderRadius: '50%',
        background: style.bg,
        border: `2px solid ${style.border}`,
        cursor: 'pointer',
        opacity: dimmed ? 0.15 : 1,
        boxShadow: isSelected ? `0 0 8px ${style.border}` : 'none',
        transform: isSelected ? 'scale(1.3)' : 'scale(1)',
        transition: 'transform 0.15s, box-shadow 0.15s, opacity 0.15s',
      }}
    />
  )
}

export function TimelineRail() {
  const events = useSupervisorStore(s => s.events)
  const selectedEventId = useSupervisorStore(s => s.selectedEventId)
  const selectedAgentId = useSupervisorStore(s => s.selectedAgentId)
  const selectEvent = useSupervisorStore(s => s.selectEvent)
  const scrollRef = useRef<HTMLDivElement>(null)
  const [scrollLeft, setScrollLeft] = useState(0)
  const [containerWidth, setContainerWidth] = useState(0)

  // Auto-scroll to latest
  useEffect(() => {
    const el = scrollRef.current
    if (!el) return
    requestAnimationFrame(() => {
      el.scrollLeft = el.scrollWidth
    })
  }, [events.length])

  // Track scroll position
  const handleScroll = useCallback(() => {
    const el = scrollRef.current
    if (el) setScrollLeft(el.scrollLeft)
  }, [])

  // Track container width
  useEffect(() => {
    const el = scrollRef.current
    if (!el) return
    const observer = new ResizeObserver(entries => {
      for (const entry of entries) {
        setContainerWidth(entry.contentRect.width)
      }
    })
    observer.observe(el)
    setContainerWidth(el.clientWidth)
    return () => observer.disconnect()
  }, [])

  // Compute visible range (virtualization) — spread nodes to fill container when few events
  const { startIdx, endIdx, totalWidth, nodeStep } = useMemo(() => {
    const naturalWidth = events.length * NODE_STEP
    const step = events.length > 0 && naturalWidth < containerWidth
      ? Math.max(NODE_STEP, containerWidth / events.length)
      : NODE_STEP
    const total = Math.max(events.length * step, containerWidth)
    const start = Math.max(0, Math.floor(scrollLeft / step) - OVERSCAN)
    const visibleCount = Math.ceil(containerWidth / step) + OVERSCAN * 2
    const end = Math.min(events.length, start + visibleCount)
    return { startIdx: start, endIdx: end, totalWidth: total, nodeStep: step }
  }, [events.length, scrollLeft, containerWidth])

  if (events.length === 0) {
    return <div ref={scrollRef} />
  }

  return (
    <div style={{
      borderTop: '1px solid #222',
      padding: '12px 16px',
      background: '#0d0d0d',
    }}>
      <div
        ref={scrollRef}
        onScroll={handleScroll}
        className="timeline-scroll"
        style={{
          overflowX: 'auto',
          scrollBehavior: 'smooth',
        }}
      >
        <div style={{ position: 'relative', width: totalWidth, height: NODE_SIZE }}>
          {events.slice(startIdx, endIdx).map((evt, i) => (
            <TimelineNode
              key={evt.id}
              evt={evt}
              isSelected={evt.id === selectedEventId}
              dimmed={selectedAgentId != null && evt.agent_id !== selectedAgentId}
              left={(startIdx + i) * nodeStep}
              onSelect={selectEvent}
            />
          ))}
        </div>
      </div>
    </div>
  )
}
