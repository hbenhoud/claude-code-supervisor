import { useEffect, useState, useRef } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Bot } from './Bot'
import { useSupervisorStore } from '../store/supervisor'

const FADE_DELAY = 3000 // ms before completed agents fade out

export function Floor() {
  const agents = useSupervisorStore(s => s.agents)
  const selectedAgentId = useSupervisorStore(s => s.selectedAgentId)
  const selectAgent = useSupervisorStore(s => s.selectAgent)
  const events = useSupervisorStore(s => s.events)

  // Track recently completed agents to delay their removal
  const [visibleIds, setVisibleIds] = useState<Set<string>>(new Set())
  const timers = useRef<Map<string, ReturnType<typeof setTimeout>>>(new Map())

  const agentList = Array.from(agents.values())
  const root = agentList.find(a => a.id === 'root')
  const subAgents = agentList.filter(a => a.id !== 'root')

  useEffect(() => {
    const nextVisible = new Set<string>()

    for (const agent of subAgents) {
      if (agent.state === 'working' || agent.state === 'error') {
        // Active — show immediately, cancel any pending fade
        nextVisible.add(agent.id)
        const existing = timers.current.get(agent.id)
        if (existing) {
          clearTimeout(existing)
          timers.current.delete(agent.id)
        }
      } else if (visibleIds.has(agent.id) && !timers.current.has(agent.id)) {
        // Just completed — keep visible, start fade timer
        nextVisible.add(agent.id)
        const timer = setTimeout(() => {
          setVisibleIds(prev => {
            const next = new Set(prev)
            next.delete(agent.id)
            return next
          })
          timers.current.delete(agent.id)
        }, FADE_DELAY)
        timers.current.set(agent.id, timer)
      } else if (timers.current.has(agent.id)) {
        // Timer still running — keep visible
        nextVisible.add(agent.id)
      }
    }

    setVisibleIds(nextVisible)

    return () => {
      const currentTimers = timers.current
      for (const t of currentTimers.values()) clearTimeout(t)
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [agents])

  const visibleSubAgents = subAgents.filter(a => visibleIds.has(a.id))

  if (agentList.length === 0) {
    return (
      <div style={{
        flex: 1,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        color: '#555',
        fontSize: 14,
        fontFamily: 'monospace',
      }}>
        {events.length === 0
          ? 'Waiting for Claude Code...'
          : 'Processing events...'
        }
      </div>
    )
  }

  // Compute radial positions for visible sub-agents
  const centerX = 50
  const centerY = 35
  const radius = 30

  return (
    <div style={{
      flex: 1,
      position: 'relative',
      overflow: 'hidden',
      minHeight: 300,
    }}>
      {/* Conduit lines */}
      <svg style={{ position: 'absolute', inset: 0, width: '100%', height: '100%', pointerEvents: 'none' }}>
        {visibleSubAgents.map((agent, i) => {
          const angle = ((2 * Math.PI) / Math.max(visibleSubAgents.length, 1)) * i - Math.PI / 2
          const targetX = centerX + radius * Math.cos(angle)
          const targetY = centerY + radius * Math.sin(angle) + 15

          return (
            <motion.line
              key={agent.id}
              x1={`${centerX}%`} y1={`${centerY + 8}%`}
              x2={`${targetX}%`} y2={`${targetY - 5}%`}
              stroke={agent.state === 'working' ? '#3b82f6' : '#333'}
              strokeWidth="2"
              strokeDasharray={agent.state === 'working' ? '6 3' : '4 4'}
              opacity={0.6}
              initial={{ pathLength: 0 }}
              animate={{ pathLength: 1 }}
              transition={{ duration: 0.5 }}
            />
          )
        })}
      </svg>

      {/* Root bot */}
      {root && (
        <div style={{
          position: 'absolute',
          left: `${centerX}%`,
          top: `${centerY}%`,
          transform: 'translate(-50%, -50%)',
        }}>
          <Bot
            type={root.type}
            name={root.name}
            state={root.state}
            currentAction={root.currentTool}
            size={80}
            selected={selectedAgentId === root.id}
            onClick={() => selectAgent(root.id)}
          />
        </div>
      )}

      {/* Sub-agent bots */}
      <AnimatePresence>
        {visibleSubAgents.map((agent, i) => {
          const angle = ((2 * Math.PI) / Math.max(visibleSubAgents.length, 1)) * i - Math.PI / 2
          const x = centerX + radius * Math.cos(angle)
          const y = centerY + radius * Math.sin(angle) + 15

          return (
            <motion.div
              key={agent.id}
              initial={{ opacity: 0, scale: 0.3 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.3 }}
              transition={{ duration: 0.4, ease: 'backOut' }}
              style={{
                position: 'absolute',
                left: `${x}%`,
                top: `${y}%`,
                transform: 'translate(-50%, -50%)',
              }}
            >
              <Bot
                type={agent.type}
                name={agent.name}
                state={agent.state}
                currentAction={agent.currentTool}
                size={56}
                selected={selectedAgentId === agent.id}
                onClick={() => selectAgent(agent.id)}
              />
            </motion.div>
          )
        })}
      </AnimatePresence>
    </div>
  )
}
