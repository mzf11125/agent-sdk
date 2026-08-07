package erc8312

import (
	"github.com/ethereum/go-ethereum/common"
)

// Status is the lifecycle status of an ERC-8312 envelope
// (IBoundedAgentAction.Status).
type Status uint8

const (
	StatusNone      Status = iota // 0: nonexistent / not registered
	StatusActive                  // 1: envelope live; cursors can advance
	StatusCompleted               // 2: terminal — mandate fulfilled
	StatusContested               // 3: disputed; no state transitions while contested
	StatusRevoked                 // 4: terminal — mandate withdrawn
	StatusExpired                 // 5: effective status once expiresAt is reached
)

// Envelope is a registered bounded mandate (IBoundedAgentAction.Envelope).
// Field names mirror the ABI component names (PascalCase of the Solidity
// names) so go-ethereum can unpack the getEnvelope tuple output directly.
type Envelope struct {
	Id             common.Hash    // Unique registry identifier.
	Principal      common.Address // The mandate's principal.
	CapabilityRoot common.Hash    // Commitment to the agreed authority; immutable.
	CursorRoot     common.Hash    // Commitment to aggregate consumption; changes only via advanceCursor.
	CreatedAt      uint64         // block.timestamp at registration.
	ExpiresAt      uint64         // Unix timestamp after which the effective status is Expired.
	Status         Status         // Effective lifecycle status.
}

// Bound is the configured budget bound for an envelope
// (IBudgetSubstrate.bound): the maximum cap of asset consumable under the
// envelope.
type Bound struct {
	Cap   uint64         // Maximum consumable amount.
	Asset common.Address // The bound asset (e.g. USDC).
}

// AdvanceResult is the result of an advanceCursor call, parsed from the
// EnvelopeAdvanced event: the previous and the new cursor commitments.
type AdvanceResult struct {
	PrevCursor common.Hash
	NewCursor  common.Hash
}

// ContestInfo is the result of a contest call, parsed from the
// EnvelopeContested event.
type ContestInfo struct {
	Id         common.Hash    // The contested envelope.
	Challenger common.Address // Who initiated the contest.
}

// ResolveInfo is the result of a resolve call, parsed from the
// EnvelopeResolved event.
type ResolveInfo struct {
	Id      common.Hash // The resolved envelope.
	Outcome Status      // Active or Revoked.
}
