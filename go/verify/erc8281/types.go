package erc8281

import "github.com/ethereum/go-ethereum/common"

// RecordedEvent is the IObservationCommitment.Recorded event (ERC-8281),
// emitted on every successful record(digest) call. The event log is the
// ledger — the interface exposes no on-chain getter.
type RecordedEvent struct {
	Digest    common.Hash    // The committed digest (topics[1]).
	Committer common.Address // The address that called record (topics[2]).
}
