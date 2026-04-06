import { motion } from 'framer-motion'
import type { BotType, BotState } from '../types/events'

interface BotProps {
  type: BotType
  name: string
  state: BotState
  currentAction?: string
  size?: number
  selected?: boolean
  onClick?: () => void
}

const BOT_COLORS: Record<BotType, { body: string; accent: string }> = {
  root: { body: '#6366f1', accent: '#818cf8' },
  explore: { body: '#3b82f6', accent: '#60a5fa' },
  plan: { body: '#f97316', accent: '#fb923c' },
  general: { body: '#22c55e', accent: '#4ade80' },
}

const STATE_GLOW: Record<BotState, string> = {
  idle: 'rgba(255,255,255,0.05)',
  working: 'rgba(59,130,246,0.3)',
  error: 'rgba(239,68,68,0.4)',
  done: 'rgba(34,197,94,0.2)',
}

const STATE_ANIMATIONS: Record<BotState, import('framer-motion').TargetAndTransition> = {
  idle: { y: [0, -2, 0], transition: { repeat: Infinity, duration: 3, ease: 'easeInOut' } },
  working: { y: [0, -4, 0], transition: { repeat: Infinity, duration: 0.8, ease: 'easeInOut' } },
  error: { x: [-2, 2, -2, 0], transition: { repeat: 2, duration: 0.3 } },
  done: { scale: [1, 1.05, 1], transition: { duration: 0.5 } },
}

export function Bot({ type, name, state, currentAction, size = 64, selected, onClick }: BotProps) {
  const colors = BOT_COLORS[type]
  const s = size

  return (
    <motion.div
      onClick={onClick}
      animate={STATE_ANIMATIONS[state]}
      style={{
        cursor: onClick ? 'pointer' : 'default',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 4,
      }}
    >
      <svg width={s} height={s} viewBox="0 0 64 64" style={{
        filter: selected ? `drop-shadow(0 0 8px ${colors.accent})` : `drop-shadow(0 0 4px ${STATE_GLOW[state]})`,
      }}>
        {/* Body */}
        <rect x="16" y="20" width="32" height="28" rx="6" fill={colors.body} />

        {/* Head */}
        <rect x="20" y="8" width="24" height="18" rx="4" fill={colors.body} />

        {/* Eyes */}
        <circle cx="28" cy="16" r="3" fill="white" />
        <circle cx="36" cy="16" r="3" fill="white" />
        <motion.circle
          cx="28" cy="16" r="1.5" fill="#1a1a2e"
          animate={state === 'working' ? { cx: [27, 29, 27] } : {}}
          transition={{ repeat: Infinity, duration: 1.5 }}
        />
        <motion.circle
          cx="36" cy="16" r="1.5" fill="#1a1a2e"
          animate={state === 'working' ? { cx: [35, 37, 35] } : {}}
          transition={{ repeat: Infinity, duration: 1.5 }}
        />

        {/* Antenna */}
        <line x1="32" y1="8" x2="32" y2="2" stroke={colors.accent} strokeWidth="2" />
        <motion.circle
          cx="32" cy="2" r="2" fill={colors.accent}
          animate={state === 'working' ? { opacity: [1, 0.3, 1] } : { opacity: 0.5 }}
          transition={{ repeat: Infinity, duration: 0.8 }}
        />

        {/* Arms */}
        <rect x="8" y="24" width="8" height="4" rx="2" fill={colors.accent} />
        <rect x="48" y="24" width="8" height="4" rx="2" fill={colors.accent} />

        {/* Legs */}
        <rect x="22" y="48" width="6" height="8" rx="2" fill={colors.accent} />
        <rect x="36" y="48" width="6" height="8" rx="2" fill={colors.accent} />

        {/* Type icon on chest */}
        {type === 'explore' && (
          <circle cx="32" cy="33" r="5" fill="none" stroke="white" strokeWidth="1.5" opacity="0.7" />
        )}
        {type === 'plan' && (
          <rect x="27" y="28" width="10" height="10" rx="1" fill="none" stroke="white" strokeWidth="1.5" opacity="0.7" />
        )}
        {type === 'general' && (
          <path d="M32 28 L36 36 L28 36 Z" fill="none" stroke="white" strokeWidth="1.5" opacity="0.7" />
        )}
        {type === 'root' && (
          <path d="M28 33 L32 28 L36 33 M32 28 L32 38" fill="none" stroke="white" strokeWidth="1.5" opacity="0.7" />
        )}

        {/* Error indicator */}
        {state === 'error' && (
          <motion.text
            x="32" y="0" textAnchor="middle" fill="#ef4444" fontSize="14" fontWeight="bold"
            initial={{ opacity: 0, y: 5 }} animate={{ opacity: 1, y: 0 }}
          >!</motion.text>
        )}
      </svg>

      {/* Name label */}
      <span style={{
        fontSize: 10,
        color: selected ? colors.accent : '#888',
        fontWeight: selected ? 'bold' : 'normal',
        fontFamily: 'monospace',
      }}>
        {name}
      </span>

      {/* Action label */}
      {currentAction && state === 'working' && (
        <motion.span
          initial={{ opacity: 0 }} animate={{ opacity: 1 }}
          style={{ fontSize: 9, color: '#60a5fa', fontFamily: 'monospace' }}
        >
          {currentAction}
        </motion.span>
      )}
    </motion.div>
  )
}
