import { create } from 'zustand'
import type { SupervisorEvent, Session, Agent, BotType, BotState } from '../types/events'

interface SupervisorState {
  // Connection
  connected: boolean
  setConnected: (v: boolean) => void

  // Sessions
  sessions: Session[]
  setSessions: (s: Session[]) => void
  removeSession: (id: string) => void
  activeSessionId: string | null
  setActiveSession: (id: string | null) => void

  // Events for active session
  events: SupervisorEvent[]
  addEvent: (evt: SupervisorEvent) => void
  clearEvents: () => void

  // Agents derived from events
  agents: Map<string, Agent>

  // Live mode
  liveMode: boolean
  toggleLiveMode: () => void

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

function extractAgentDescription(evt: SupervisorEvent): string | undefined {
  if (!evt.tool_input || typeof evt.tool_input !== 'object') return undefined
  const input = evt.tool_input as Record<string, unknown>
  return (input.description as string) || (input.prompt as string) || undefined
}

function extractSubagentType(evt: SupervisorEvent): BotType | undefined {
  if (!evt.tool_input || typeof evt.tool_input !== 'object') return undefined
  const input = evt.tool_input as Record<string, unknown>
  const subType = (input.subagent_type as string)?.toLowerCase()
  if (subType === 'explore') return 'explore'
  if (subType === 'plan') return 'plan'
  return undefined
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
  removeSession: (id) => set(state => ({
    sessions: state.sessions.filter(s => s.id !== id),
    ...(state.activeSessionId === id ? { activeSessionId: null, events: [], agents: new Map() } : {}),
  })),
  activeSessionId: null,
  setActiveSession: (id) => set({ activeSessionId: id, events: [], agents: new Map(), selectedEventId: null, selectedAgentId: null }),

  events: [],
  addEvent: (evt) => {
    const state = get()
    const newEvents = [...state.events, evt]
    const agents = new Map(state.agents)

    const agentId = evt.agent_id || 'root'

    // Ensure root agent exists on first event
    if (!agents.has('root')) {
      agents.set('root', {
        id: 'root',
        type: 'root',
        name: generateBotName('root', evt.session_id),
        state: 'idle',
        toolCount: 0,
      })
    }

    // Ensure this agent exists (auto-create for sub-agents)
    if (agentId !== 'root' && !agents.has(agentId)) {
      const description = extractAgentDescription(evt)
      const type = inferBotType(agentId, description)
      agents.set(agentId, {
        id: agentId,
        parentId: evt.parent_agent_id || 'root',
        type,
        name: generateBotName(type, agentId),
        state: 'idle',
        toolCount: 0,
      })
    }

    // Handle agent_spawn events (Agent tool calls)
    if (evt.event_type === 'agent_spawn' && evt.event_subtype === 'start') {
      // The tool_input contains the sub-agent's description/type
      const description = extractAgentDescription(evt)
      const subAgentType = extractSubagentType(evt)
      const type = subAgentType || inferBotType(agentId, description)
      // The spawned agent will appear when its own tool calls arrive
      // But we can create a placeholder from the Agent tool_input
      const spawnedId = evt.tool_use_id ? `agent-${evt.tool_use_id.slice(0, 16)}` : agentId
      if (!agents.has(spawnedId)) {
        agents.set(spawnedId, {
          id: spawnedId,
          parentId: agentId,
          type,
          name: generateBotName(type, spawnedId),
          state: 'working',
          toolCount: 0,
        })
      }
    }

    // Update agent state on tool calls
    if (evt.event_type === 'tool_call' || evt.event_type === 'agent_spawn') {
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
        agents.set(id, { ...agent, state: 'done', currentTool: undefined })
      }
    }

    set({
      events: newEvents,
      agents,
      ...(state.liveMode ? { selectedEventId: evt.id } : {}),
    })
  },
  clearEvents: () => set({ events: [], agents: new Map(), liveMode: true }),

  agents: new Map(),

  liveMode: true,
  toggleLiveMode: () => set(state => ({ liveMode: !state.liveMode })),

  selectedEventId: null,
  selectedAgentId: null,
  selectEvent: (id) => set({ selectedEventId: id, liveMode: false }),
  selectAgent: (id) => set({ selectedAgentId: id, selectedEventId: null, liveMode: false }),
}))
