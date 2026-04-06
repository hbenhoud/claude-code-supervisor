import { create } from 'zustand'
import type { SupervisorEvent, Session, Agent, BotType, BotState } from '../types/events'

interface SupervisorState {
  // Connection
  connected: boolean
  setConnected: (v: boolean) => void

  // Sessions
  sessions: Session[]
  setSessions: (s: Session[]) => void
  activeSessionId: string | null
  setActiveSession: (id: string | null) => void

  // Events for active session
  events: SupervisorEvent[]
  addEvent: (evt: SupervisorEvent) => void
  clearEvents: () => void

  // Agents derived from events
  agents: Map<string, Agent>

  // Selection
  selectedEventId: string | null
  selectedAgentId: string | null
  selectEvent: (id: string | null) => void
  selectAgent: (id: string | null) => void
}

function inferBotType(agentId: string, description?: string): BotType {
  if (!description) return agentId === 'root' ? 'root' : 'general'
  const d = description.toLowerCase()
  if (d.includes('explore') || d.includes('search') || d.includes('find')) return 'explore'
  if (d.includes('plan') || d.includes('architect') || d.includes('design')) return 'plan'
  return 'general'
}

function generateBotName(type: BotType, id: string): string {
  const hash = id.slice(-4).replace(/[^a-z0-9]/gi, '')
  const prefixes: Record<BotType, string[]> = {
    root: ['Rex', 'Atlas', 'Prime', 'Core'],
    explore: ['Scout', 'Finder', 'Radar', 'Seek'],
    plan: ['Forge', 'Arch', 'Blueprint', 'Draft'],
    general: ['Bolt', 'Gear', 'Spark', 'Cog'],
  }
  const names = prefixes[type]
  const idx = parseInt(hash, 36) % names.length || 0
  return `${names[idx]}-${hash.slice(0, 2)}`
}

function inferBotState(evt: SupervisorEvent): BotState {
  if (evt.error) return 'error'
  if (evt.event_subtype === 'start') return 'working'
  if (evt.event_subtype === 'complete') return 'idle'
  return 'idle'
}

export const useSupervisorStore = create<SupervisorState>((set, get) => ({
  connected: false,
  setConnected: (v) => set({ connected: v }),

  sessions: [],
  setSessions: (s) => set({ sessions: s }),
  activeSessionId: null,
  setActiveSession: (id) => set({ activeSessionId: id, events: [], agents: new Map(), selectedEventId: null, selectedAgentId: null }),

  events: [],
  addEvent: (evt) => {
    const state = get()
    const newEvents = [...state.events, evt]
    const agents = new Map(state.agents)

    // Handle agent spawn
    if (evt.event_type === 'agent_spawn') {
      const type = inferBotType(evt.agent_id)
      agents.set(evt.agent_id, {
        id: evt.agent_id,
        parentId: evt.parent_agent_id,
        type,
        name: generateBotName(type, evt.agent_id),
        state: 'idle',
        toolCount: 0,
      })
    }

    // Ensure root agent exists
    if (!agents.has('root') && evt.agent_id === 'root') {
      agents.set('root', {
        id: 'root',
        type: 'root',
        name: generateBotName('root', evt.session_id),
        state: 'idle',
        toolCount: 0,
      })
    }

    // Update agent state on tool call
    if (evt.event_type === 'tool_call') {
      const agentId = evt.agent_id || 'root'
      const agent = agents.get(agentId)
      if (agent) {
        agents.set(agentId, {
          ...agent,
          state: inferBotState(evt),
          currentTool: evt.event_subtype === 'start' ? evt.tool_name : undefined,
          toolCount: evt.event_subtype === 'complete' ? agent.toolCount + 1 : agent.toolCount,
        })
      }
    }

    // Handle notification (session end)
    if (evt.event_type === 'notification') {
      for (const [id, agent] of agents) {
        agents.set(id, { ...agent, state: 'done' })
      }
    }

    set({ events: newEvents, agents })
  },
  clearEvents: () => set({ events: [], agents: new Map() }),

  agents: new Map(),

  selectedEventId: null,
  selectedAgentId: null,
  selectEvent: (id) => set({ selectedEventId: id }),
  selectAgent: (id) => set({ selectedAgentId: id }),
}))
