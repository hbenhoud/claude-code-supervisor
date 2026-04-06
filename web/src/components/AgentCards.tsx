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

export function AgentCards() {
  const agents = useSupervisorStore(s => s.agents)
  const selectedAgentId = useSupervisorStore(s => s.selectedAgentId)
  const selectAgent = useSupervisorStore(s => s.selectAgent)

  const agentList = Array.from(agents.values())

  if (agentList.length === 0) return null

  return (
    <div style={{
      width: 140,
      borderRight: '1px solid #222',
      padding: 8,
      display: 'flex',
      flexDirection: 'column',
      gap: 8,
      overflowY: 'auto',
    }}>
      {agentList.map(agent => (
        <motion.div
          key={agent.id}
          onClick={() => selectAgent(agent.id)}
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          style={{
            padding: 8,
            borderRadius: 8,
            background: selectedAgentId === agent.id ? '#1e1e2e' : '#111',
            border: `1px solid ${selectedAgentId === agent.id ? '#3b82f6' : '#222'}`,
            cursor: 'pointer',
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            gap: 4,
            boxShadow: agent.state === 'working' ? '0 0 8px rgba(59,130,246,0.2)' : 'none',
          }}
        >
          <Bot type={agent.type} name={agent.name} state={agent.state} size={40} />
          {agent.currentTool && agent.state === 'working' && (
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
      ))}
    </div>
  )
}
