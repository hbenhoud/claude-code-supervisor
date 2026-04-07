import { useState } from 'react'
import { motion } from 'framer-motion'
import { Bot } from './Bot'
import { useSupervisorStore } from '../store/supervisor'
import type { BotState } from '../types/events'

const HEALTH_COLORS: Record<BotState, string> = {
  idle: '#888',
  working: '#4ade80',
  error: '#ef4444',
  done: '#888',
}

function isActive(state: BotState) {
  return state === 'working' || state === 'error'
}

export function AgentCards({ onShowSessions }: { onShowSessions?: () => void }) {
  const agents = useSupervisorStore(s => s.agents)
  const selectedAgentId = useSupervisorStore(s => s.selectedAgentId)
  const selectAgent = useSupervisorStore(s => s.selectAgent)
  const [showCompleted, setShowCompleted] = useState(false)

  const agentList = Array.from(agents.values())
  // Root is always shown in active section
  const active = agentList.filter(a => a.id === 'root' || isActive(a.state))
  const completed = agentList.filter(a => a.id !== 'root' && !isActive(a.state))

  if (agentList.length === 0) return null

  return (
    <div style={{
      width: 140,
      flexShrink: 0,
      borderRight: '1px solid #222',
      padding: 8,
      display: 'flex',
      flexDirection: 'column',
      gap: 8,
      overflowY: 'auto',
    }}>
      {active.map(agent => (
        <AgentCard
          key={agent.id}
          agent={agent}
          selected={selectedAgentId === agent.id}
          onSelect={() => selectAgent(agent.id)}
        />
      ))}

      {completed.length > 0 && (
        <>
          <div
            onClick={() => setShowCompleted(!showCompleted)}
            style={{
              padding: '6px 8px',
              fontSize: 10,
              color: '#666',
              cursor: 'pointer',
              textAlign: 'center',
              borderTop: '1px solid #222',
              marginTop: 4,
              userSelect: 'none',
            }}
          >
            {showCompleted ? '▾' : '▸'} Completed ({completed.length})
          </div>
          {showCompleted && completed.map(agent => (
            <AgentCard
              key={agent.id}
              agent={agent}
              selected={selectedAgentId === agent.id}
              onSelect={() => selectAgent(agent.id)}
              compact
            />
          ))}
        </>
      )}

      {onShowSessions && (
        <div
          onClick={onShowSessions}
          style={{
            marginTop: 'auto',
            padding: '8px 0',
            borderTop: '1px solid #222',
            textAlign: 'center',
            fontSize: 11,
            color: '#666',
            cursor: 'pointer',
            userSelect: 'none',
          }}
          onMouseEnter={e => (e.currentTarget.style.color = '#aaa')}
          onMouseLeave={e => (e.currentTarget.style.color = '#666')}
        >
          ← Sessions
        </div>
      )}
    </div>
  )
}

function AgentCard({ agent, selected, onSelect, compact }: {
  agent: { id: string; type: 'root' | 'explore' | 'plan' | 'general'; name: string; state: BotState; currentTool?: string; toolCount: number }
  selected: boolean
  onSelect: () => void
  compact?: boolean
}) {
  return (
    <motion.div
      onClick={onSelect}
      initial={{ opacity: 0, x: -20 }}
      animate={{ opacity: 1, x: 0 }}
      style={{
        padding: compact ? 6 : 8,
        borderRadius: 8,
        background: selected ? '#1e1e2e' : '#111',
        border: `1px solid ${selected ? '#3b82f6' : '#222'}`,
        cursor: 'pointer',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: compact ? 2 : 4,
        opacity: compact ? 0.6 : 1,
        boxShadow: agent.state === 'working' ? '0 0 8px rgba(59,130,246,0.2)' : 'none',
      }}
    >
      <Bot type={agent.type} name={agent.name} state={agent.state} size={compact ? 28 : 40} />
      {!compact && agent.currentTool && agent.state === 'working' && (
        <span style={{ fontSize: 9, color: '#60a5fa', fontFamily: 'monospace' }}>
          {agent.currentTool}
        </span>
      )}
      <div style={{ display: 'flex', alignItems: 'center', gap: 4, fontSize: 9, color: '#666' }}>
        <div style={{
          width: 6, height: 6, borderRadius: '50%',
          background: HEALTH_COLORS[agent.state],
        }} />
        {agent.toolCount} tools
      </div>
    </motion.div>
  )
}
