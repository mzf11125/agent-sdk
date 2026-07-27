export { AgentReputationClient } from './client.js'
export { computeWinRate } from './recompute.js'
export type { AgentReputationClientConfig, ReputationData } from './types.js'
export {
  WIN_RATE_BPS_V0_HASH,
  WIN_RATE_BPS_V0_SPEC,
  governingConventionHash,
  pinWinRateBps,
  verifyWinRate,
} from './conventions.js'
export type { ConventionStatus, PinnedWinRate, Verdict } from './conventions.js'
