export interface SupervisorEvent {
  id: string
  session_id: string
  timestamp: string
  sequence: number
  event_type: 'tool_call' | 'agent_spawn' | 'notification' | string
  event_subtype: 'start' | 'complete' | string
  agent_id: string
  parent_agent_id?: string
  tool_name?: string
  tool_use_id?: string
  tool_input?: unknown
  tool_output?: unknown
  duration_ms?: number
  error?: string
}

export interface Session {
  id: string
  cwd: string
  status: 'running' | 'completed' | 'error'
  started_at: number
  finished_at?: number
  tool_count: number
  agent_count: number
  prompt?: string
  model?: string
  total_cost_usd?: number
  total_tokens?: number
}

export type BotType = 'root' | 'explore' | 'plan' | 'general'
export type BotState = 'idle' | 'working' | 'error' | 'done'

export interface Agent {
  id: string
  parentId?: string
  type: BotType
  name: string
  state: BotState
  currentTool?: string
  toolCount: number
}
