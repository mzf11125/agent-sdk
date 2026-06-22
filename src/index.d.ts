// Type surface for @onchain-ai/agent-sdk (v0.2).

export interface VerifyResult {
  valid: boolean;
  checks: Record<string, boolean>;
  proof_payload?: any;
  error?: string;
  how_to_verify?: string;
}

export interface VerifyOpts {
  /** x-only hex of the issuer you require (authorship gate). Omit to skip. */
  expectPubkey?: string;
  /** require content.schema to start with this, e.g. "onchain-ai." */
  schemaPrefix?: string;
}

export interface CommitSpec {
  job_id?: string;
  [k: string]: unknown;
}

export interface BuiltCommit {
  /** Unsigned event (no `sig`) — caller signs with their own key management. */
  event: Record<string, any>;
  id: string;
  artifact_hash: string;
}

export interface FullFlowResult {
  /** valid && artifact_hash_matches && anchored. Escrow ALSO checks on-chain delivery. */
  ok: boolean;
  valid: boolean;
  artifact_hash_matches: boolean;
  anchored: boolean;
  verify: VerifyResult;
  note: string;
}

export interface CommitmentProof {
  mechanism: string;
  event_id: string;
  published_at: number;
  relays: string[];
  how_to_check: string;
}

export interface PublishResult {
  event_id: string;
  published_at: number;
  /** relays that returned OK true for this exact event id */
  relays: string[];
  relay_count: number;
  /** the OTS proof (or { error } / null if not stamped) */
  ots: any;
  /** relay copies exist; the Bitcoin-PoW leg confirms later via `ots verify -d <event_id>` */
  anchored: boolean;
  /** canonical shape mirrored by GET /ledger/{entry}/commitment */
  commitment_proof: CommitmentProof;
}

export function verifyProof(event: object, opts?: VerifyOpts): VerifyResult;
export function nostrEventId(ev: object): string;
export function artifactHash(spec: CommitSpec): string;
export function canonical(value: unknown): string;
/** Recursively lowercase EVM-address-shaped strings (0x+40hex). The single anti-drift point — run on both sides. */
export function normalizeSpec<T>(spec: T): T;
export function buildCommitEvent(p: {
  spec: CommitSpec;
  pubkey: string;
  judgmentType: string;
  schema?: string;
  createdAt?: number;
}): BuiltCommit;
export function verifyFullFlow(p: {
  proofEvent: object;
  expectArtifactHash: string;
  expectPubkey?: string;
  schemaPrefix?: string;
  relaySeen?: boolean;
  otsVerified?: boolean;
}): FullFlowResult;
export function recompute(events: object[], opts?: VerifyOpts): {
  total: number; valid: number; entries: Array<Record<string, any>>;
};
/** Anchor a SIGNED commit event to public sources (relay publication + OTS). Injected I/O; never signs. */
export function publishCommit(p: {
  event: Record<string, any>;
  relays?: string[];
  publishToRelay?: (relayUrl: string, event: object) => Promise<boolean>;
  otsStamp?: (eventId: string) => Promise<any>;
  requireSig?: boolean;
}): Promise<PublishResult>;
/** Default NIP-01 WebSocket relay publisher (Node>=21/Bun/browser). Inject your own on older runtimes. */
export function relayPublish(relayUrl: string, event: object, opts?: { timeoutMs?: number }): Promise<boolean>;
export const PROOF_KIND: number;
export const CORE: { name: string; verifies_like: string; proof_kind: number };
